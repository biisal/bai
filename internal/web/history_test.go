package web

import (
	"encoding/json"
	"testing"

	"charm.land/fantasy"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/domain"
)

func partsJSON(t *testing.T, parts ...fantasy.MessagePart) string {
	t.Helper()
	b, err := json.Marshal(parts)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestToUIMessages(t *testing.T) {
	rows := []repo.Message{
		{ID: 1, ConversationID: 7, Role: domain.RoleUser, Parts: partsJSON(t, fantasy.TextPart{Text: "hi"})},
		{ID: 2, ConversationID: 7, Role: domain.RoleAssistant, Parts: partsJSON(t,
			fantasy.ReasoningPart{Text: "think"},
			fantasy.TextPart{Text: "answer"},
			fantasy.ToolCallPart{ToolCallID: "c1", ToolName: "bash", Input: `{"cmd":"ls"}`},
		)},
		{ID: 3, ConversationID: 7, Role: domain.RoleTool, Parts: partsJSON(t,
			fantasy.ToolResultPart{ToolCallID: "c1", Output: fantasy.ToolResultOutputContentText{Text: "file.txt"}},
		)},
		{ID: 4, ConversationID: 7, Role: domain.RoleAssistant, Parts: partsJSON(t, fantasy.TextPart{Text: "done"})},
	}

	got, err := json.Marshal(toUIMessages(rows))
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"id":"1","role":"user","parts":[{"type":"text","text":"hi","state":"done"}]},{"id":"2","role":"assistant","parts":[{"type":"step-start"},{"type":"reasoning","text":"think","state":"done"},{"type":"text","text":"answer","state":"done"},{"type":"tool-bash","toolCallId":"c1","state":"output-available","input":{"cmd":"ls"},"output":"file.txt"},{"type":"step-start"},{"type":"text","text":"done","state":"done"}]}]`
	if string(got) != want {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

func TestToUIMessagesToolError(t *testing.T) {
	rows := []repo.Message{
		{ID: 1, ConversationID: 7, Role: domain.RoleAssistant, Parts: partsJSON(t,
			fantasy.ToolCallPart{ToolCallID: "c1", ToolName: "read", Input: `{"path":"x"}`},
		)},
		{ID: 2, ConversationID: 7, Role: domain.RoleTool, Parts: partsJSON(t,
			fantasy.ToolResultPart{ToolCallID: "c1", Output: fantasy.ToolResultOutputContentError{Error: errTool}},
		)},
	}

	got, err := json.Marshal(toUIMessages(rows))
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"id":"1","role":"assistant","parts":[{"type":"step-start"},{"type":"tool-read","toolCallId":"c1","state":"output-error","input":{"path":"x"},"errorText":"boom"}]}]`
	if string(got) != want {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

type testError string

func (e testError) Error() string { return string(e) }

const errTool = testError("boom")
