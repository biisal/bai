package providers

import (
	"encoding/json"
	"testing"

	"github.com/biisal/bai/internal/domain"
)

func TestToProviderParts(t *testing.T) {
	p := &ProviderOpenAI{}
	parts := []domain.Part{
		{Type: domain.PartTextType, Data: domain.TextPartData{Text: "hello"}},
		{Type: domain.PartReasoningType, Data: domain.ReasoningPartData{Thinking: "think"}},
		{Type: domain.PartToolCallType, Data: domain.ToolCallPartData{ID: "call_1"}},
		{Type: domain.PartFinishType},
	}

	got := p.ToProviderParts(parts)
	if len(got) != 1 {
		t.Fatalf("ToProviderParts() len = %d, want 1 (only text parts)", len(got))
	}
	if txt := got[0].GetText(); txt == nil || *txt != "hello" {
		t.Errorf("ToProviderParts() text = %v, want %q", txt, "hello")
	}
}

func TestToProviderToolCalls(t *testing.T) {
	p := &ProviderOpenAI{}
	parts := []domain.Part{
		{Type: domain.PartToolCallType, Data: domain.ToolCallPartData{ID: "call_1", Name: "read_file", Input: []byte(`{"path":"/tmp/x"}`)}},
		{Type: domain.PartTextType, Data: domain.TextPartData{Text: "ignored"}},
		{Type: domain.PartToolCallType, Data: domain.ToolCallPartData{ID: "call_2", Name: "write_file", Input: []byte(`{}`)}},
	}

	got := p.ToProviderToolCalls(parts)
	if len(got) != 2 {
		t.Fatalf("ToProviderToolCalls() len = %d, want 2", len(got))
	}

	f := got[0].GetFunction()
	if f == nil {
		t.Fatal("ToProviderToolCalls()[0] has no function variant")
	}
	if got[0].GetID() == nil || *got[0].GetID() != "call_1" {
		t.Errorf("ToProviderToolCalls()[0] id = %v, want %q", got[0].GetID(), "call_1")
	}
	if f.Name != "read_file" {
		t.Errorf("ToProviderToolCalls()[0] name = %q, want %q", f.Name, "read_file")
	}
	if f.Arguments != `{"path":"/tmp/x"}` {
		t.Errorf("ToProviderToolCalls()[0] arguments = %q, want %q", f.Arguments, `{"path":"/tmp/x"}`)
	}

	data, err := json.Marshal(got[0])
	if err != nil {
		t.Fatalf("marshal tool call: %v", err)
	}
	want := `{"id":"call_1","function":{"arguments":"{\"path\":\"/tmp/x\"}","name":"read_file"},"type":"function"}`
	if string(data) != want {
		t.Errorf("ToProviderToolCalls()[0] json = %s, want %s", data, want)
	}
}

func TestToProviderToolResults(t *testing.T) {
	p := &ProviderOpenAI{}
	parts := []domain.Part{
		{Type: domain.PartToolResultType, Data: domain.ToolResultPartData{ToolCallID: "call_1", Name: "read_file", Content: "file contents"}},
		{Type: domain.PartTextType, Data: domain.TextPartData{Text: "ignored"}},
		{Type: domain.PartToolResultType, Data: domain.ToolResultPartData{ToolCallID: "call_2", Name: "read_file", Content: "a", Data: "b"}},
	}

	got := p.ToProviderToolResults(parts)
	if len(got) != 2 {
		t.Fatalf("ToProviderToolResults() len = %d, want 2", len(got))
	}

	type toolUnion struct {
		Role       string `json:"role"`
		ToolCallID string `json:"tool_call_id"`
		Content    string `json:"content"`
	}
	var tool toolUnion
	data, err := json.Marshal(got[0])
	if err != nil {
		t.Fatalf("marshal tool message: %v", err)
	}
	if err := json.Unmarshal(data, &tool); err != nil {
		t.Fatalf("unmarshal tool message: %v", err)
	}
	if tool.Role != "tool" {
		t.Errorf("tool message role = %q, want %q", tool.Role, "tool")
	}
	if tool.ToolCallID != "call_1" {
		t.Errorf("tool message tool_call_id = %q, want %q", tool.ToolCallID, "call_1")
	}
	if tool.Content != "file contents" {
		t.Errorf("tool message content = %q, want %q", tool.Content, "file contents")
	}
}
