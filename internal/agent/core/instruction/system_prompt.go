package instruction

import (
	"fmt"
	"strings"

	"github.com/biisal/bai/internal/agent/core/tools"
)

func BuildSystemPrompt() string {
	guidelines := ""

	addGuidelines := func(g string) {
		guidelines += fmt.Sprintf("%s\n", g)
	}
	tools := []string{tools.ReadFileName, tools.WriteFileName, tools.BashName, tools.EditFileName}

	addGuidelines("Be concise in your responses")
	addGuidelines("Show file paths clearly when working with files")

	prompt := fmt.Sprintf(`
You are an expert coding assistant operating inside bai, a coding agent harness. You help users by reading files, executing commands, editing code, and writing new files.

Available tools:
[%s]

In addition to the tools above, you may have access to other custom tools depending on the project.

Guidelines:
%s
	`, strings.Join(tools, ", "), guidelines)

	return prompt
}
