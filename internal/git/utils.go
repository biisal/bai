package git

import (
	"os"
	"slices"
	"strings"
)

// GitIgnoreFolders returns the plain folder/file names ignored by the
// .gitignore at path. Patterns are processed in order, so a later negation
// ("!name") re-includes a name ignored earlier, matching git's
// "last matching pattern wins" rule. Patterns with wildcards or inner
// slashes are skipped, since callers only match on base names.
func GitIgnoreFolders(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var folders []string
	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSuffix(line, "\r")

		// Trailing spaces are dropped unless escaped with a backslash.
		for strings.HasSuffix(line, " ") && !strings.HasSuffix(line, "\\ ") {
			line = line[:len(line)-1]
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		negate := false
		switch {
		case strings.HasPrefix(line, "!"):
			negate = true
			line = line[1:]
		case strings.HasPrefix(line, `\#`), strings.HasPrefix(line, `\!`):
			line = line[1:] // escaped leading '#' or '!' is literal
		}
		if strings.HasSuffix(line, `\ `) {
			line = line[:len(line)-2] + " " // escaped trailing space is literal
		}

		line = strings.Trim(line, "/")
		if line == "" || strings.ContainsAny(line, "/*?[") {
			continue
		}

		idx := slices.Index(folders, line)
		switch {
		case negate && idx >= 0:
			folders = slices.Delete(folders, idx, idx+1)
		case !negate && idx < 0:
			folders = append(folders, line)
		}
	}
	if len(folders) == 0 {
		return nil, nil
	}
	return folders, nil
}
