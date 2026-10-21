package route

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	dao "teachat/DAO"
)

func TestTeamServiceOfferingsTemplateRendersPublishedActions(t *testing.T) {
	type offeringWithVersion struct {
		*dao.TeamServiceOffering
		VersionString string
	}
	pageData := struct {
		TeamBean        dao.TeamBean
		Offerings       []offeringWithVersion
		Editing         *dao.TeamServiceOffering
		TeamSkills      []dao.TeamSkillDetail
		EditingSkillIds map[int]bool
		IsManager       bool
		IsVerifier      bool
	}{
		TeamBean: dao.TeamBean{Team: dao.Team{Uuid: "team-uuid", Name: "测试团队"}},
		Offerings: []offeringWithVersion{{
			TeamServiceOffering: &dao.TeamServiceOffering{
				Uuid: "offering-uuid", Name: "咨询服务", Status: dao.PublishedTeamServiceOfferingStatus,
				Availability: dao.AvailableTeamServiceOffering,
			},
			VersionString: "1.0.0",
		}},
		TeamSkills: []dao.TeamSkillDetail{{
			SkillTeam: dao.SkillTeam{SkillId: 7, Level: 3, Status: dao.NormalSkillTeamStatus},
			SkillName: "茶艺冲泡",
		}},
		EditingSkillIds: map[int]bool{7: true},
		IsManager:       true,
	}

	tmpl, err := template.ParseFiles("../templates/team.service_offerings.go.html")
	if err != nil {
		t.Fatalf("parse management template: %v", err)
	}
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "content", pageData); err != nil {
		t.Fatalf("render management template: %v", err)
	}
	out := output.String()
	if !strings.Contains(out, "暂停接单") || !strings.Contains(out, "下架") {
		t.Fatalf("published offering actions missing from output: %s", out)
	}
	if !strings.Contains(out, "v1.0.0") {
		t.Fatalf("version label missing from output: %s", out)
	}
	if !strings.Contains(out, `name="skill_ids"`) || !strings.Contains(out, "茶艺冲泡") {
		t.Fatalf("team skill checkboxes missing from output: %s", out)
	}
}
