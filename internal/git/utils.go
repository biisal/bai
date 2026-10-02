package git

import (
	"os"
	"strings"
)

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
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		line = strings.Trim(line, "/")
		if line == "" || strings.ContainsAny(line, "/*?[") {
			continue
		}
		folders = append(folders, line)
	}
	return folders, nil
}
