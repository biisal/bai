package web

import (
	"encoding/json"
	"fmt"
	"time"

	"charm.land/fantasy"
)

const frameDone = "[DONE]"

type textFrame struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Delta string `json:"delta,omitempty"`
}

type startFrame struct {
	Type      string `json:"type"`
	MessageID string `json:"messageId"`
}

type conversationFrame struct {
	Type      string           `json:"type"`
	Data      conversationData `json:"data"`
	Transient bool             `json:"transient"`
}

type conversationData struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type stepFrame struct {
	Type string `json:"type"`
}

type toolStartFrame struct {
	Type       string `json:"type"`
	ToolCallID string `json:"toolCallId"`
	ToolName   string `json:"toolName"`
}

type toolDeltaFrame struct {
	Type           string `json:"type"`
	ToolCallID     string `json:"toolCallId"`
	InputTextDelta string `json:"inputTextDelta"`
}

type toolInputFrame struct {
	Type       string `json:"type"`
	ToolCallID string `json:"toolCallId"`
	ToolName   string `json:"toolName"`
	Input      any    `json:"input"`
}

type toolOutputFrame struct {
	Type       string `json:"type"`
	ToolCallID string `json:"toolCallId"`
	Output     string `json:"output"`
}

type toolErrorFrame struct {
	Type       string `json:"type"`
	ToolCallID string `json:"toolCallId"`
	ErrorText  string `json:"errorText"`
}

type usageFrame struct {
	Type      string    `json:"type"`
	Data      usageData `json:"data"`
	Transient bool      `json:"transient"`
}

type usageData struct {
	InputTokens         int64 `json:"inputTokens"`
	OutputTokens        int64 `json:"outputTokens"`
	TotalTokens         int64 `json:"totalTokens"`
	ReasoningTokens     int64 `json:"reasoningTokens"`
	CacheCreationTokens int64 `json:"cacheCreationTokens"`
	CacheReadTokens     int64 `json:"cacheReadTokens"`
	ElapsedMs           int64 `json:"elapsedMs"`
}

type errorFrame struct {
	Type      string `json:"type"`
	ErrorText string `json:"errorText"`
}

type abortFrame struct {
	Type   string `json:"type"`
	Reason string `json:"reason,omitempty"`
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func fStart(messageID string) []byte {
	return mustJSON(startFrame{Type: "start", MessageID: messageID})
}

func fConversation(id int64, title string) []byte {
	return mustJSON(conversationFrame{
		Type:      "data-conversation",
		Data:      conversationData{ID: id, Title: title},
		Transient: true,
	})
}

func fStartStep() []byte  { return mustJSON(stepFrame{Type: "start-step"}) }
func fFinishStep() []byte { return mustJSON(stepFrame{Type: "finish-step"}) }
func fResetStep() []byte  { return mustJSON(stepFrame{Type: "reset-step"}) }

func fTextStart(id string) []byte {
	return mustJSON(textFrame{Type: "text-start", ID: id})
}

func fTextDelta(id, delta string) []byte {
	return mustJSON(textFrame{Type: "text-delta", ID: id, Delta: delta})
}

func fTextEnd(id string) []byte {
	return mustJSON(textFrame{Type: "text-end", ID: id})
}

func fReasoningStart(id string) []byte {
	return mustJSON(textFrame{Type: "reasoning-start", ID: id})
}

func fReasoningDelta(id, delta string) []byte {
	return mustJSON(textFrame{Type: "reasoning-delta", ID: id, Delta: delta})
}

func fReasoningEnd(id string) []byte {
	return mustJSON(textFrame{Type: "reasoning-end", ID: id})
}

func fToolInputStart(toolCallID, toolName string) []byte {
	return mustJSON(toolStartFrame{Type: "tool-input-start", ToolCallID: toolCallID, ToolName: toolName})
}

func fToolInputDelta(toolCallID, delta string) []byte {
	return mustJSON(toolDeltaFrame{Type: "tool-input-delta", ToolCallID: toolCallID, InputTextDelta: delta})
}

func fToolInputAvailable(toolCallID, toolName, input string) []byte {
	var embedded any = input
	if json.Valid([]byte(input)) {
		embedded = json.RawMessage(input)
	}
	return mustJSON(toolInputFrame{
		Type:       "tool-input-available",
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Input:      embedded,
	})
}

func fToolOutputAvailable(toolCallID, output string) []byte {
	return mustJSON(toolOutputFrame{Type: "tool-output-available", ToolCallID: toolCallID, Output: output})
}

func fToolOutputError(toolCallID, errorText string) []byte {
	return mustJSON(toolErrorFrame{Type: "tool-output-error", ToolCallID: toolCallID, ErrorText: errorText})
}

func fUsage(usage fantasy.Usage, elapsed time.Duration) []byte {
	return mustJSON(usageFrame{
		Type:      "data-usage",
		Transient: true,
		Data: usageData{
			InputTokens:         usage.InputTokens,
			OutputTokens:        usage.OutputTokens,
			TotalTokens:         usage.TotalTokens,
			ReasoningTokens:     usage.ReasoningTokens,
			CacheCreationTokens: usage.CacheCreationTokens,
			CacheReadTokens:     usage.CacheReadTokens,
			ElapsedMs:           elapsed.Milliseconds(),
		},
	})
}

func fFinish() []byte { return mustJSON(stepFrame{Type: "finish"}) }

func fAbort(reason string) []byte {
	return mustJSON(abortFrame{Type: "abort", Reason: reason})
}

func fError(text string) []byte {
	return mustJSON(errorFrame{Type: "error", ErrorText: text})
}

func toolResultFrame(result fantasy.ToolResultContent) []byte {
	switch out := result.Result.(type) {
	case fantasy.ToolResultOutputContentText:
		return fToolOutputAvailable(result.ToolCallID, out.Text)
	case fantasy.ToolResultOutputContentError:
		text := "tool execution failed"
		if out.Error != nil {
			text = out.Error.Error()
		}
		return fToolOutputError(result.ToolCallID, text)
	case nil:
		return fToolOutputAvailable(result.ToolCallID, "")
	default:
		b, err := json.Marshal(out)
		if err != nil {
			return fToolOutputAvailable(result.ToolCallID, fmt.Sprint(out))
		}
		return fToolOutputAvailable(result.ToolCallID, string(b))
	}
}
