package dao

import (
	"reflect"
	"testing"
)

func TestNormalizeAgencyScopes(t *testing.T) {
	got, err := NormalizeAgencyScopes([]string{"read", "POST", "read", " "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{AgentScopeRead, AgentScopePost}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	if _, err := NormalizeAgencyScopes([]string{"read", "admin"}); err == nil {
		t.Fatal("expected unknown scope to be rejected")
	}
	if _, err := NormalizeAgencyScopes(nil); err == nil {
		t.Fatal("expected empty scope list to be rejected")
	}
}

func TestAgentScopeReviewRegistered(t *testing.T) {
	got, err := NormalizeAgencyScopes([]string{"read", "review"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[1] != AgentScopeReview {
		t.Fatalf("expected review scope normalized, got %v", got)
	}
	if !profileHasScopeOption(AgentScopeReview) {
		t.Fatal("review scope must be present in AgentScopeOptions whitelist")
	}
}

func profileHasScopeOption(scope string) bool {
	for _, opt := range AgentScopeOptions {
		if opt.Key == scope {
			return true
		}
	}
	return false
}

func TestAgentProfileHasScope(t *testing.T) {
	profile := AgentProfile{AgencyScope: "read, post"}
	if !profile.HasScope(AgentScopePost) {
		t.Fatal("expected post scope")
	}
	if !profile.HasScope(" READ ") {
		t.Fatal("scope check should be whitespace/case insensitive")
	}
	if profile.HasScope(AgentScopeTrade) {
		t.Fatal("trade scope was not declared")
	}
}

// 人类用户（含历史 captain/traveller 角色）不进入 AGI 范围校验，
// 因此无需访问数据库即可放行。
func TestCheckAgentScopeAllowsHumans(t *testing.T) {
	for _, role := range []string{User_Role_Human, "captain", "traveller", ""} {
		allowed, err := CheckAgentScope(User{Id: 1, Role: role}, AgentScopeTrade)
		if err != nil {
			t.Fatalf("role %q: unexpected error: %v", role, err)
		}
		if !allowed {
			t.Fatalf("role %q should be allowed", role)
		}
	}
}
