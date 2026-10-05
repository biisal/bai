package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSkill(t *testing.T, root, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadSkills(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "hello-world", "---\nname: hello-world\ndescription: Says hello\n---\n\nBody here")
	writeSkill(t, root, "Bad_Name", "---\nname: Bad_Name\ndescription: x\n---\nbody")
	writeSkill(t, root, "wrong-name", "---\nname: other\ndescription: x\n---\nbody")

	got := LoadSkills(root)
	if len(got) != 1 {
		t.Fatalf("got %d skills, want 1: %+v", len(got), got)
	}
	s := got[0]
	if s.Name != "hello-world" || s.Description != "Says hello" || s.Context != "Body here" {
		t.Errorf("skill = %+v", s)
	}
	if s.Location != filepath.Join(root, "hello-world", "SKILL.md") {
		t.Errorf("location = %q", s.Location)
	}
}

func TestParseSkillRejectsNonSkillContent(t *testing.T) {
	for _, content := range []string{"no frontmatter", "---\nname: other\ndescription: x\n---\nbody", "---\nname: x"} {
		if skill, ok := parseSkill("x", "", content); ok {
			t.Errorf("parseSkill(%q) = %+v, want rejection", content, skill)
		}
	}
}

func TestGetInternalTool(t *testing.T) {
	content, ok := GetInternalTool("internal:skills:tools_maker")
	if !ok || strings.TrimSpace(content) == "" {
		t.Errorf("GetInternalTool = %q, %v; want non-empty content", content, ok)
	}
	if _, ok := GetInternalTool("nope"); ok {
		t.Error("expected miss for unknown tool")
	}
	internal := InternalSkills()
	if len(internal) != 1 || internal[0].Location != "internal:skills:tools_maker" {
		t.Errorf("InternalSkills() = %+v, want tools_maker", internal)
	}
	if got := LoadSkills(); len(got) != 0 {
		t.Errorf("LoadSkills() = %+v, want no disk skills", got)
	}
}
