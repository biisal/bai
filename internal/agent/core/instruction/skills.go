package instruction

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Skill struct {
	Name        string
	Description string
	Location    string
	Context     string
}

// validSkillName checks that name matches the required skill naming convention:
// 1–64 chars, lowercase alphanumeric with single hyphen separators,
// no leading/trailing hyphens, no consecutive hyphens.
var validSkillName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

const maxSkillNameLen = 64

func LoadSkills(paths ...string) []Skill {
	var skills []Skill

	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dirName := entry.Name()

			// Validate directory name against skill naming rules.
			if len(dirName) == 0 || len(dirName) > maxSkillNameLen {
				continue
			}
			if !validSkillName.MatchString(dirName) {
				continue
			}

			skillPath := filepath.Join(dir, dirName, "SKILL.md")
			content, err := os.ReadFile(skillPath)
			if err != nil {
				continue
			}

			skill, ok := parseSkill(dirName, skillPath, string(content))
			if !ok {
				continue
			}

			skills = append(skills, skill)
		}
	}

	return skills
}

// parseSkill extracts name, description and context from SKILL.md content.
// The file must have YAML frontmatter delimited by --- lines:
//
//	---
//	name: skill-name
//	description: What the skill does
//	---
//
//	## Context content follows
func parseSkill(dirName, location, content string) (Skill, bool) {
	content = strings.TrimSpace(content)

	// Must start with the opening frontmatter delimiter.
	if !strings.HasPrefix(content, "---") {
		return Skill{}, false
	}

	// Find the closing delimiter.
	rest := content[3:]
	before, after, ok := strings.Cut(rest, "\n---")
	if !ok {
		return Skill{}, false
	}

	frontmatter := before
	body := strings.TrimSpace(after)

	name := ""
	description := ""

	for line := range strings.SplitSeq(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(line[5:])
		} else if strings.HasPrefix(line, "description:") {
			description = strings.TrimSpace(line[12:])
		}
	}

	// The name in frontmatter must match the directory name.
	if name != dirName {
		return Skill{}, false
	}

	return Skill{
		Name:        name,
		Description: description,
		Location:    location,
		Context:     body,
	}, true
}

func formatSkills(skills []Skill) string {
	if len(skills) == 0 {
		return ""
	}

	sb := strings.Builder{}

	sb.WriteString("\nThe following skills provide specialized instructions for specific tasks.\n")
	sb.WriteString("Read the full skill file when the task matches its description.\n\n")

	sb.WriteString("<available_skills>")
	for _, skill := range skills {
		sb.WriteString("<skill>")
		fmt.Fprintf(&sb, "\t\t<name>%s</name>\n\t\t<description>%s</description>\n\t\t<location>%s</location>\n", skill.Name, skill.Description, skill.Location)
	}
	sb.WriteString("</available_skills>")

	return sb.String()
}
