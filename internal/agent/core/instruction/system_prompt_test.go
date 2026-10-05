package instruction

import (
	"fmt"
	"strings"
	"testing"

	"github.com/biisal/bai/internal/agent/core/tools"
)

func TestBuildSystemPromptHasInternalToolInstruction(t *testing.T) {
	prompt := BuildSystemPrompt(nil, nil, nil)
	want := fmt.Sprintf("Internal tools (can't read with bash, use %s to read):\ninternal:skills:tools_maker", tools.ReadFileName)
	if !strings.Contains(prompt, want) {
		t.Errorf("prompt missing internal tool instruction, got:\n%s", prompt)
	}
}
