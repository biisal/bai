package files

import "testing"

func TestFindFile(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedMatched bool
		expectedQuery   string
	}{
		{
			name:            "should mathch on normal user imput with random prefix text",
			input:           "this is a test@",
			expectedMatched: true,
			expectedQuery:   "",
		},
		{
			name:            "should mathch on normal user imput with random prefix text and query after @",
			input:           "this is a test@fileName",
			expectedMatched: true,
			expectedQuery:   "fileName",
		},
		{
			name:            "should return matched fales when @ is prent but query having space",
			input:           "this is a test@file Name",
			expectedMatched: false,
			expectedQuery:   "",
		},
		{
			name:            "should match on single @",
			input:           "@",
			expectedMatched: true,
			expectedQuery:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matched, query := IfFileFinding(test.input)
			if matched != test.expectedMatched {
				t.Errorf("expected %v, got %v", test.expectedMatched, matched)
			}
			if query != test.expectedQuery {
				t.Errorf("expected %v, got %v", test.expectedQuery, query)
			}
		})
	}
}

func TestReplaceFileQuery(t *testing.T) {
	tests := []struct {
		name     string
		fullText string
		path     string
		want     string
	}{
		{
			name:     "replaces file query at end",
			fullText: "find @myfile.txt",
			path:     "/path/to/file.txt",
			want:     "find /path/to/file.txt",
		},
		{
			name:     "replaces multiple @ symbols",
			fullText: "find @src/index.ts",
			path:     "./index.ts",
			want:     "find ./index.ts",
		},
		{
			name:     "no file query to replace",
			fullText: "find something",
			path:     "/path/to/file.txt",
			want:     "find something",
		},
		{
			name:     "file query with no spaces",
			fullText: "@file.go",
			path:     "new_file.go",
			want:     "new_file.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceFileQuery(tt.fullText, tt.path)
			if got != tt.want {
				t.Errorf("ReplaceFileQuery(%q, %q) = %q, want %q", tt.fullText, tt.path, got, tt.want)
			}
		})
	}
}
