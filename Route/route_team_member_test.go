package route

import (
	"strings"
	"testing"

	dao "teachat/DAO"
)

func TestBuildTeamRoleAdjustedMessage(t *testing.T) {
	team := dao.Team{Uuid: "team-123"}
	user := dao.User{Name: "茶友A", FamilyName: "茶", GivenName: "友A", AliasName: "B"}

	msg := buildTeamRoleAdjustedMessage(team, user, dao.RoleCEO)

	if !strings.Contains(msg, "返回团队详情") {
		t.Fatalf("expected success message to include return link, got %q", msg)
	}

	if !strings.Contains(msg, "/v1/team/detail?uuid=team-123") {
		t.Fatalf("expected success message to include team detail link, got %q", msg)
	}

	if !strings.Contains(msg, "茶友A") {
		t.Fatalf("expected success message to mention member name, got %q", msg)
	}
}
