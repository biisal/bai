package web

import (
	"testing"
	"time"

	"charm.land/fantasy"
	test_utils "github.com/biisal/bai/utils/tests"
)

func TestFrames(t *testing.T) {
	tests := []struct {
		name string
		got  []byte
		want string
	}{
		{"start", fStart("m1"), `{"type":"start","messageId":"m1"}`},
		{"conversation", fConversation(7, "hello"), `{"type":"data-conversation","data":{"id":7,"title":"hello"},"transient":true}`},
		{"start-step", fStartStep(), `{"type":"start-step"}`},
		{"finish-step", fFinishStep(), `{"type":"finish-step"}`},
		{"reset-step", fResetStep(), `{"type":"reset-step"}`},
		{"text-start", fTextStart("t"), `{"type":"text-start","id":"t"}`},
		{"text-delta", fTextDelta("t", "hi\nthere"), `{"type":"text-delta","id":"t","delta":"hi\nthere"}`},
		{"text-end", fTextEnd("t"), `{"type":"text-end","id":"t"}`},
		{"reasoning-start", fReasoningStart("r"), `{"type":"reasoning-start","id":"r"}`},
		{"reasoning-delta", fReasoningDelta("r", "think"), `{"type":"reasoning-delta","id":"r","delta":"think"}`},
		{"reasoning-end", fReasoningEnd("r"), `{"type":"reasoning-end","id":"r"}`},
		{"tool-start", fToolInputStart("c1", "bash"), `{"type":"tool-input-start","toolCallId":"c1","toolName":"bash"}`},
		{"tool-delta", fToolInputDelta("c1", `{"cmd":`), `{"type":"tool-input-delta","toolCallId":"c1","inputTextDelta":"{\"cmd\":"}`},
		{"tool-available-json", fToolInputAvailable("c1", "bash", `{"cmd":"ls"}`), `{"type":"tool-input-available","toolCallId":"c1","toolName":"bash","input":{"cmd":"ls"}}`},
		{"tool-available-text", fToolInputAvailable("c1", "bash", "not json"), `{"type":"tool-input-available","toolCallId":"c1","toolName":"bash","input":"not json"}`},
		{"tool-output", fToolOutputAvailable("c1", "file.txt"), `{"type":"tool-output-available","toolCallId":"c1","output":"file.txt"}`},
		{"tool-error", fToolOutputError("c1", "boom"), `{"type":"tool-output-error","toolCallId":"c1","errorText":"boom"}`},
		{"finish", fFinish(), `{"type":"finish"}`},
		{"abort", fAbort("user cancelled"), `{"type":"abort","reason":"user cancelled"}`},
		{"error", fError("bad"), `{"type":"error","errorText":"bad"}`},
		{"usage", fUsage(fantasy.Usage{InputTokens: 3, OutputTokens: 2, TotalTokens: 5}, 1500*time.Millisecond), `{"type":"data-usage","data":{"inputTokens":3,"outputTokens":2,"totalTokens":5,"reasoningTokens":0,"cacheCreationTokens":0,"cacheReadTokens":0,"elapsedMs":1500},"transient":true}`},
		{"done", []byte(frameDone), "[DONE]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.got); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestTextAndReasoningSequence(t *testing.T) {
	got := []string{
		string(fStartStep()),
		string(fTextStart("a")),
		string(fTextDelta("a", "hello")),
		string(fTextEnd("a")),
		string(fReasoningStart("b")),
		string(fReasoningDelta("b", "why")),
		string(fReasoningEnd("b")),
		string(fFinishStep()),
	}
	want := []string{
		`{"type":"start-step"}`,
		`{"type":"text-start","id":"a"}`,
		`{"type":"text-delta","id":"a","delta":"hello"}`,
		`{"type":"text-end","id":"a"}`,
		`{"type":"reasoning-start","id":"b"}`,
		`{"type":"reasoning-delta","id":"b","delta":"why"}`,
		`{"type":"reasoning-end","id":"b"}`,
		`{"type":"finish-step"}`,
	}
	test_utils.AssertSliceEqual(t, got, want)
}
