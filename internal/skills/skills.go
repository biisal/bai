package skills

import (
	_ "embed"
)

//go:embed files/tools_maker.md
var ToolsMaker string

//go:embed files/theme_maker.md
var ThemeMaker string

// internalSkills are skills shipped with bai instead of living on disk.
// Their Location is the id read uses to fetch them.
var internalSkills = buildInternalSkills()

func buildInternalSkills() []Skill {
	embedded := []struct {
		name    string
		id      string
		content string
	}{
		{"tools_maker", "internal:skills:tools_maker", ToolsMaker},
		{"theme_maker", "internal:skills:theme_maker", ThemeMaker},
	}

	skills := make([]Skill, 0, len(embedded))
	for _, e := range embedded {
		skill, ok := parseSkill(e.name, e.id, e.content)
		if !ok {
			continue
		}
		skills = append(skills, skill)
	}
	return skills
}

// InternalSkills returns the skills shipped with bai, kept separate from the
// ones loaded from disk: they are fetched by id through read.
func InternalSkills() []Skill {
	return internalSkills
}

func GetInternalTool(id string) (string, bool) {
	for _, skill := range internalSkills {
		if skill.Location == id {
			return skill.Context, true
		}
	}
	return "", false
}
