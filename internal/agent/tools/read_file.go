package tools

import (
	"bufio"
	"fmt"
	"os"
)

const (
	DefaultMaxLines = 2000
	DefaultMaxBytes = 256 * 1024
)

func ReadFile(path string, offset int64, limit int64) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot access %q: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%q is a directory, not a file", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open %q: %w", path, err)
	}
	defer file.Close()

	if offset <= 0 {
		offset = 1
	}

	maxLines := DefaultMaxLines
	if limit > 0 && int64(maxLines) > limit {
		maxLines = int(limit)
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // allow long lines up to 1MB

	var lineNum int64
	var out []byte
	linesRead := 0
	truncatedByBytes := false

	for scanner.Scan() {
		lineNum++
		if lineNum < offset {
			continue
		}
		line := scanner.Bytes()

		if len(out)+len(line)+1 > DefaultMaxBytes {
			truncatedByBytes = true
			break
		}
		out = append(out, line...)
		out = append(out, '\n')
		linesRead++

		if linesRead >= maxLines {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading %q: %w", path, err)
	}

	if linesRead == 0 {
		return "", fmt.Errorf("offset %d is beyond end of file (%d lines total)", offset, lineNum)
	}

	hasMore := scanner.Scan() || truncatedByBytes
	if hasMore {
		nextOffset := offset + int64(linesRead)
		reason := "line limit"
		if truncatedByBytes {
			reason = "size limit"
		}
		out = append(out, fmt.Appendf(nil,
			"\n[Showing lines %d-%d. Truncated by %s — use offset=%d to continue.]\n",
			offset, offset+int64(linesRead)-1, reason, nextOffset)...)
	}

	return string(out), nil
}
