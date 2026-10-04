package route

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	dao "teachat/DAO"
)

func TestVerifierWorkspaceRendersPendingServiceOfferings(t *testing.T) {
	pageData := dao.VerifierWorkspagePageData{
		ActiveWorkspaceTab:          "services",
		PendingServiceOfferingCount: 1,
		PendingServiceOfferings: []*dao.TeamServiceOffering{{
			Uuid: "pending-offering", TeamName: "青松团队", Name: "茶艺咨询",
			Summary: "茶艺流程梳理", Status: dao.PendingTeamServiceOfferingStatus,
		}},
	}

	tmpl, err := template.ParseFiles("../templates/verifier.workspace.go.html")
	if err != nil {
		t.Fatalf("parse verifier workspace template: %v", err)
	}
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "content", pageData); err != nil {
		t.Fatalf("render verifier workspace template: %v", err)
	}
	for _, expected := range []string{"服务项目", "青松团队", "茶艺咨询", "查看详情并处理", "pending-offering"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("workspace output missing %q", expected)
		}
	}
}

func TestVerifierServiceOfferingDetailRendersReviewActions(t *testing.T) {
	pageData := struct {
		Offering *dao.TeamServiceOffering
		Team     dao.Team
		Skills   []struct {
			Name      string
			Level     int
			IsPrimary bool
		}
		Events    []*dao.ServiceOfferingEvent
		IsPending bool
	}{
		Offering: &dao.TeamServiceOffering{
			Uuid: "pending-offering", Name: "茶艺咨询", Summary: "茶艺流程梳理",
			Status: dao.PendingTeamServiceOfferingStatus,
		},
		Team:      dao.Team{Name: "青松团队"},
		IsPending: true,
	}

	tmpl, err := template.ParseFiles("../templates/verifier.service_offering.detail.go.html")
	if err != nil {
		t.Fatalf("parse service offering detail template: %v", err)
	}
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "content", pageData); err != nil {
		t.Fatalf("render service offering detail template: %v", err)
	}
	for _, expected := range []string{"审核决定", "通过并上架", "婉拒原因", "name=\"action\" value=\"reject\"", "pending-offering"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("detail output missing %q", expected)
		}
	}
}
