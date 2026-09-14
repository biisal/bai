package instruction

import (
	"fmt"
	"strings"

	"github.com/biisal/bai/internal/agent/core/tools"
)

func ReadAgentMd() string {
	files := []string{"AGENTS.md", "agents.md", "agent.md", "AGENT.md"}
	for _, file := range files {
		content, err := tools.ReadFile(file, 0, 0)
		if err != nil || content == "" {
			continue
		}
		return content
	}

	return ""
}

func formatUsersInstructions(usersInstructions []string) string {
	if len(usersInstructions) == 0 {
		return ""
	}

	content := strings.Join(usersInstructions, "\n")
	if content == "" {
		return ""
	}
	return fmt.Sprintf("User instructions:\n%s", content)
}

func BuildSystemPrompt(usersInstructions []string, skills []Skill) string {
	guidelines := ""

	addGuidelines := func(g string) {
		guidelines += fmt.Sprintf("%s\n", g)
	}
	tools := []string{tools.ReadFileName, tools.WriteFileName, tools.BashName, tools.EditFileName}

	addGuidelines("Be concise in your responses")
	addGuidelines("Show file paths clearly when working with files")
	addGuidelines(formatSkills(skills))

	prompt := fmt.Sprintf(`
You are an expert coding assistant operating inside bai, a coding agent harness. You help users by reading files, executing commands, editing code, and writing new files.

Available tools:
[%s]

In addition to the tools above, you may have access to other custom tools depending on the project.

%s

Guidelines:
%s
	`, strings.Join(tools, ", "), formatUsersInstructions(usersInstructions), guidelines)

	return prompt
}
