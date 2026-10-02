package route

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	dao "teachat/DAO"
)

func TestTeamServiceOfferingsTemplateRendersPublishedActions(t *testing.T) {
	pageData := struct {
		TeamBean   dao.TeamBean
		Offerings  []*dao.TeamServiceOffering
		Editing    *dao.TeamServiceOffering
		IsManager  bool
		IsVerifier bool
	}{
		TeamBean: dao.TeamBean{Team: dao.Team{Uuid: "team-uuid", Name: "测试团队"}},
		Offerings: []*dao.TeamServiceOffering{{
			Uuid: "offering-uuid", Name: "咨询服务", Status: dao.PublishedTeamServiceOfferingStatus,
			Availability: dao.AvailableTeamServiceOffering,
		}},
		IsManager: true,
	}

	tmpl, err := template.ParseFiles("../templates/team.service_offerings.go.html")
	if err != nil {
		t.Fatalf("parse management template: %v", err)
	}
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "content", pageData); err != nil {
		t.Fatalf("render management template: %v", err)
	}
	if !strings.Contains(output.String(), "暂停接单") || !strings.Contains(output.String(), "下架") {
		t.Fatalf("published offering actions missing from output: %s", output.String())
	}
}
