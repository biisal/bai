package domain

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

type PartType string

const (
	PartTextType       PartType = "text"
	PartReasoningType  PartType = "reasoning"
	PartToolCallType   PartType = "tool_call"
	PartToolResultType PartType = "tool_result"
	PartFinishType     PartType = "finish"
)

type Part struct {
	Type PartType
	Data any
}

type TextPartData struct {
	Text string
}

type ReasoningPartData struct {
	Thinking   string
	Signature  string
	StartedAt  time.Time
	FinishedAt time.Time
}

type ToolCallPartData struct {
	ID               string
	Name             string
	Input            json.RawMessage
	ProviderExecuted bool
	Finished         bool
}

type ToolResultPartData struct {
	ToolCallID string
	Name       string
	Content    string
	Data       string
	MIMEType   string
	Metadata   string
	IsError    bool
}

func UnmarshalParts(data []byte) ([]Part, error) {
	var rawParts []json.RawMessage
	if err := json.Unmarshal(data, &rawParts); err != nil {
		return nil, err
	}

	parts := make([]Part, 0, len(rawParts))
	for _, rawPart := range rawParts {
		var wrapper struct {
			Type PartType        `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(rawPart, &wrapper); err != nil {
			return nil, err
		}

		switch wrapper.Type {
		case PartReasoningType:
			part := ReasoningPartData{}
			if err := json.Unmarshal(wrapper.Data, &part); err != nil {
				return nil, err
			}
			parts = append(parts, Part{Type: PartReasoningType, Data: part})
		case PartTextType:
			part := TextPartData{}
			if err := json.Unmarshal(wrapper.Data, &part); err != nil {
				return nil, err
			}
			parts = append(parts, Part{Type: PartTextType, Data: part})
		case PartToolCallType:
			part := ToolCallPartData{}
			if err := json.Unmarshal(wrapper.Data, &part); err != nil {
				return nil, err
			}
			parts = append(parts, Part{Type: PartToolCallType, Data: part})
		case PartToolResultType:
			part := ToolResultPartData{}
			if err := json.Unmarshal(wrapper.Data, &part); err != nil {
				return nil, err
			}
			parts = append(parts, Part{Type: PartToolResultType, Data: part})
		case PartFinishType:
			parts = append(parts, Part{Type: PartFinishType})
		default:
			return nil, fmt.Errorf("unknown part type: %s", wrapper.Type)
		}
	}

	return parts, nil
}

func MarshalParts(parts []Part) ([]byte, error) {
	slog.Debug("MarshalParts() = ", "parts", fmt.Sprintf("%q", parts))

	type envelope struct {
		Type PartType `json:"type"`
		Data any      `json:"data,omitempty"`
	}

	envelopes := make([]envelope, 0, len(parts))
	for _, part := range parts {
		envelopes = append(envelopes, envelope(part))
	}
	return json.Marshal(envelopes)
}
