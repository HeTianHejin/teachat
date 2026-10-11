package route

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	dao "teachat/DAO"
)

func TestVerifierSkillsTemplateRendersPendingSkill(t *testing.T) {
	tmpl, err := template.ParseFiles("../templates/verifier.skills.go.html")
	if err != nil {
		t.Fatal(err)
	}
	data := struct {
		SessUser dao.User
		Skills   []dao.Skill
	}{Skills: []dao.Skill{{
		Uuid: "skill-uuid", Name: "家具修复", Description: "修复日常家具", UserId: 12,
		Category: dao.GeneralHardSkill,
	}}}
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "content", data); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"家具修复", "待审核", "/v1/verifier/skill/review", "婉拒展示"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("rendered template does not contain %q", want)
		}
	}
}

func TestSkillDetailTemplateRendersDraftActions(t *testing.T) {
	tmpl, err := template.New("skill.detail").Funcs(template.FuncMap{
		"seq": func(start, end int) []int {
			values := make([]int, 0, end-start+1)
			for value := start; value <= end; value++ {
				values = append(values, value)
			}
			return values
		},
	}).ParseFiles("../templates/skill.detail.go.html")
	if err != nil {
		t.Fatal(err)
	}
	data := struct {
		SessUser dao.User
		Skill    *dao.Skill
	}{
		SessUser: dao.User{Id: 12},
		Skill: &dao.Skill{
			Id: 3, UserId: 12, Name: "家具修复", Description: "修复日常家具",
			Category: dao.GeneralHardSkill, Level: 1,
			StrengthLevel: dao.ModerateStrength, DifficultyLevel: dao.ModerateDifficulty,
			ReviewStatus: dao.DraftSkillReviewStatus,
		},
	}
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "content", data); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"草稿", "/v1/skill/update", "/v1/skill/submit-review"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("rendered template does not contain %q", want)
		}
	}
}
