package tools

import "testing"

func TestPurposeFromToolCall(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: WriteFileName, input: `{"path":"a.txt","purpose":"add hello world","content":"x"}`, want: "add hello world"},
		{name: EditFileName, input: `{"path":"a.txt","purpose":"rename var","edits":[{"old_text":"a","purpose":"nested","new_text":"b"}]}`, want: "rename var"},
		{name: BashName, input: `{"command":"go test ./...","purpose":"run tests"}`, want: "run tests"},
		{name: ReadFileName, input: `{"path":"a.txt"}`, want: ""},
		{name: BashName, input: `not json`, want: ""},
		{name: BashName, input: `{"command":"go build"}`, want: ""},
		{name: BashName, input: `{"command":"ls","purpose":"  spaced  "}`, want: "spaced"},
	}
	for _, tt := range tests {
		got := PurposeFromToolCall(tt.name, tt.input)
		if got != tt.want {
			t.Errorf("PurposeFromToolCall(%q, %q) = %q, want %q", tt.name, tt.input, got, tt.want)
		}
	}
}