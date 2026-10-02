package files

import (
	"fmt"
	"regexp"
)

var findFileRegex = regexp.MustCompile(`@[^\s]*$`)

func IfFileFinding(text string) (mathched bool, query string) {
	if findFileRegex.MatchString(text) {
		mathched = true
		text = findFileRegex.FindString(text)
		if len(text) > 1 {
			query = text[1:]
		}
	}
	return
}

func ReplaceFileQuery(fullText, path string) string {
	return findFileRegex.ReplaceAllString(fullText, fmt.Sprintf("%q ", path))
}
