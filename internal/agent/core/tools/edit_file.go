package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Edit struct {
	OldText string
	NewText string
}

type editSpan struct {
	start, end int
	newText    string
}

func EditFile(ctx context.Context, path string, edits []Edit) error {
	if len(edits) == 0 {
		return fmt.Errorf("edit tool input is invalid: edits must contain at least one replacement")
	}
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	resolved, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve %q: %w", path, err)
	}

	return withFileMutationQueue(resolved, func() error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("edit aborted for %q: %w", resolved, err)
		}

		info, err := os.Stat(resolved)
		if err != nil {
			return fmt.Errorf("could not edit file %q: %w", resolved, err)
		}
		if info.IsDir() {
			return fmt.Errorf("could not edit file %q: path is not a file", resolved)
		}

		raw, err := os.ReadFile(resolved)
		if err != nil {
			return fmt.Errorf("could not edit file %q: %w", resolved, err)
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("edit aborted for %q: %w", resolved, err)
		}

		var bom []byte
		content := raw
		if bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}) {
			bom = content[:3]
			content = content[3:]
		}

		hasCRLF := bytes.Contains(content, []byte("\r\n"))
		normalized := strings.ReplaceAll(string(content), "\r\n", "\n")

		spans := make([]editSpan, 0, len(edits))
		for i, e := range edits {
			if e.OldText == "" {
				return fmt.Errorf("edit %d is invalid: old_text must not be empty", i)
			}
			count := strings.Count(normalized, e.OldText)
			if count == 0 {
				return fmt.Errorf("edit %d not applied: old_text not found in %q", i, path)
			}
			if count > 1 {
				return fmt.Errorf("edit %d not applied: old_text is not unique in %q (%d matches)", i, path, count)
			}
			start := strings.Index(normalized, e.OldText)
			spans = append(spans, editSpan{start: start, end: start + len(e.OldText), newText: e.NewText})
		}

		sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
		for i := 1; i < len(spans); i++ {
			if spans[i].start < spans[i-1].end {
				return fmt.Errorf("edits overlap in %q: merge overlapping edits into one", path)
			}
		}

		var out strings.Builder
		cursor := 0
		for _, sp := range spans {
			out.WriteString(normalized[cursor:sp.start])
			out.WriteString(sp.newText)
			cursor = sp.end
		}
		out.WriteString(normalized[cursor:])
		newContent := out.String()

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("edit aborted for %q: %w", resolved, err)
		}

		if hasCRLF {
			newContent = strings.ReplaceAll(newContent, "\n", "\r\n")
		}
		final := make([]byte, 0, len(bom)+len(newContent))
		final = append(final, bom...)
		final = append(final, []byte(newContent)...)

		if err := os.WriteFile(resolved, final, 0o644); err != nil {
			return fmt.Errorf("could not edit file %q: %w", resolved, err)
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("edit aborted for %q: %w", resolved, err)
		}

		return nil
	})
}
