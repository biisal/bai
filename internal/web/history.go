package web

import (
	"encoding/json"
	"fmt"
	"strconv"

	"charm.land/fantasy"
	repo "github.com/biisal/bai/internal/db/sqlc"
	"github.com/biisal/bai/internal/domain"
)

type uiOutMessage struct {
	ID    string `json:"id"`
	Role  string `json:"role"`
	Parts []any  `json:"parts"`
}

type uiTextPart struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	State string `json:"state,omitempty"`
}

type uiStepPart struct {
	Type string `json:"type"`
}

type uiToolPart struct {
	Type       string `json:"type"`
	ToolCallID string `json:"toolCallId"`
	State      string `json:"state"`
	Input      any    `json:"input,omitempty"`
	Output     any    `json:"output,omitempty"`
	ErrorText  string `json:"errorText,omitempty"`
}

func toUIMessages(rows []repo.Message) []uiOutMessage {
	out := make([]uiOutMessage, 0, len(rows))
	var current *uiOutMessage
	toolParts := make(map[string]*uiToolPart)

	flush := func() {
		if current != nil {
			out = append(out, *current)
			current = nil
			clear(toolParts)
		}
	}

	for _, row := range rows {
		switch row.Role {
		case domain.RoleUser:
			flush()
			parts := make([]any, 0)
			for _, part := range partsOf(row.Parts) {
				if text, ok := fantasy.AsMessagePart[fantasy.TextPart](part); ok {
					parts = append(parts, uiTextPart{Type: "text", Text: text.Text, State: "done"})
				}
			}
			out = append(out, uiOutMessage{ID: strconv.FormatInt(row.ID, 10), Role: "user", Parts: parts})

		case domain.RoleAssistant:
			if current == nil {
				current = &uiOutMessage{ID: strconv.FormatInt(row.ID, 10), Role: "assistant"}
			}
			current.Parts = append(current.Parts, uiStepPart{Type: "step-start"})
			for _, part := range partsOf(row.Parts) {
				switch p := part.(type) {
				case fantasy.TextPart:
					current.Parts = append(current.Parts, uiTextPart{Type: "text", Text: p.Text, State: "done"})
				case fantasy.ReasoningPart:
					current.Parts = append(current.Parts, uiTextPart{Type: "reasoning", Text: p.Text, State: "done"})
				case fantasy.ToolCallPart:
					call := &uiToolPart{
						Type:       "tool-" + p.ToolName,
						ToolCallID: p.ToolCallID,
						State:      "input-available",
						Input:      embedJSON(p.Input),
					}
					current.Parts = append(current.Parts, call)
					toolParts[p.ToolCallID] = call
				}
			}

		case domain.RoleTool:
			if current == nil {
				current = &uiOutMessage{ID: strconv.FormatInt(row.ID, 10), Role: "assistant"}
			}
			for _, part := range partsOf(row.Parts) {
				result, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](part)
				if !ok {
					continue
				}
				call, ok := toolParts[result.ToolCallID]
				if !ok {
					continue
				}
				switch output := result.Output.(type) {
				case fantasy.ToolResultOutputContentText:
					call.State = "output-available"
					call.Output = output.Text
				case fantasy.ToolResultOutputContentError:
					call.State = "output-error"
					call.ErrorText = "tool execution failed"
					if output.Error != nil {
						call.ErrorText = output.Error.Error()
					}
				default:
					call.State = "output-available"
					b, err := json.Marshal(output)
					if err != nil {
						call.Output = fmt.Sprint(output)
					} else {
						call.Output = string(b)
					}
				}
			}
		}
	}
	flush()
	return out
}

func partsOf(raw string) []fantasy.MessagePart {
	var parts []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &parts); err != nil {
		return []fantasy.MessagePart{fantasy.TextPart{Text: raw}}
	}
	out := make([]fantasy.MessagePart, 0, len(parts))
	for _, part := range parts {
		parsed, err := fantasy.UnmarshalMessagePart(part)
		if err == nil {
			out = append(out, parsed)
		}
	}
	return out
}

func embedJSON(input string) any {
	if json.Valid([]byte(input)) {
		return json.RawMessage(input)
	}
	return input
}
