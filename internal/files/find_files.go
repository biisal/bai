package files

import (
	"fmt"
	"log/slog"
	"regexp"
)

var findFileRegex = regexp.MustCompile(`@[^\s]*$`)

func IfFileFinding(text string) (mathched bool, query string) {
	if findFileRegex.MatchString(text) {
		mathched = true
		// TODO: check
		query = findFileRegex.FindString(text)[1:]
	}

	slog.Debug("file finding", "matched", mathched, "query", query)
	return
}

func ReplaceFileQuery(fullText, path string) string {
	return findFileRegex.ReplaceAllString(fullText, fmt.Sprintf("%q", path))
}
