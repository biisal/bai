package instruction

import (
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/biisal/bai/internal/agent/core/tools"
	"github.com/biisal/bai/internal/skills"
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

func internalToolSkills() string {
	internal := skills.InternalSkills()
	if len(internal) == 0 {
		return ""
	}
	ids := make([]string, len(internal))
	for i, skill := range internal {
		ids[i] = skill.Location
	}
	return fmt.Sprintf("Internal tools (can't read with bash, use %s to read):\n%s", tools.ReadFileName, strings.Join(ids, "\n"))
}

func BuildSystemPrompt(usersInstructions []string, skillList []skills.Skill, agentTools []fantasy.AgentTool) string {
	guidelines := ""

	addGuidelines := func(g string) {
		guidelines += fmt.Sprintf("%s\n", g)
	}
	tools := []string{}
	for _, tool := range agentTools {
		tools = append(tools, tool.Info().Name)
	}

	addGuidelines("Be concise in your responses")
	addGuidelines("Show file paths clearly when working with files")
	addGuidelines(skills.FormatSkills(skillList))

	prompt := fmt.Sprintf(`
You are an expert coding assistant operating inside bai, a coding agent harness. You help users by reading files, executing commands, editing code, and writing new files.

Available tools:
[%s]

In addition to the tools above, you may have access to other custom tools depending on the project.

%s

%s

Guidelines:
%s
	`, strings.Join(tools, ", "), formatUsersInstructions(usersInstructions), internalToolSkills(), guidelines)

	return prompt
}
