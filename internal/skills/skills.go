package skills

import (
	_ "embed"
)

//go:embed files/tools_maker.md
var ToolsMaker string

// internalSkills are skills shipped with bai instead of living on disk.
// Their Location is the id read uses to fetch them.
var internalSkills = buildInternalSkills()

func buildInternalSkills() []Skill {
	const id = "internal:skills:tools_maker"
	skill, ok := parseSkill("tools_maker", id, ToolsMaker)
	if !ok {
		return nil
	}
	return []Skill{skill}
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
