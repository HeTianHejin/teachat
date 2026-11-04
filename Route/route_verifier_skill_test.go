package route

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestVerifierSkillAPIRejectsWrongMethod(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		handler http.HandlerFunc
	}{
		{"skills POST", http.MethodPost, HandleVerifierSkillsAPI},
		{"review GET", http.MethodGet, HandleVerifierSkillReviewAPI},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c.handler(rec, httptest.NewRequest(c.method, "/v1/api/verifier/skills", nil))
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want 405", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			var resp ApiResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
			if resp.Success {
				t.Error("405 response should have success=false")
			}
		})
	}
}

func TestVerifierSkillAPIRequiresLogin(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		handler http.HandlerFunc
	}{
		{"skills", http.MethodGet, HandleVerifierSkillsAPI},
		{"review", http.MethodPost, HandleVerifierSkillReviewAPI},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(c.method, "/v1/api/verifier/skills", strings.NewReader(`{"uuid":"skill-uuid","action":"approve"}`))
			rec := httptest.NewRecorder()
			c.handler(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
			var resp ApiResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
			if resp.Success || resp.Message != "请先登录" {
				t.Errorf("unexpected response: %+v", resp)
			}
		})
	}
}

func TestNewVerifierSkillItem(t *testing.T) {
	item := newVerifierSkillItem(dao.Skill{
		Uuid: "skill-uuid", Name: "家具修复", UserId: 12,
		Category: dao.GeneralHardSkill, StrengthLevel: dao.ModerateStrength,
		CreatedAt: time.Date(2026, 11, 4, 9, 30, 0, 0, time.UTC),
	})
	if item.ReviewStatus != int(dao.PendingSkillReviewStatus) {
		t.Errorf("review_status = %d, want pending", item.ReviewStatus)
	}
	if item.ReviewStatusLabel != dao.PendingSkillReviewStatus.String() {
		t.Errorf("review_status_label = %q", item.ReviewStatusLabel)
	}
	if item.CategoryLabel != "通用硬技能" {
		t.Errorf("category_label = %q", item.CategoryLabel)
	}
	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"uuid":"skill-uuid"`, `"user_id":12`, `"review_status":2`, `"created_at":"2026-11-04T09:30:00Z"`} {
		if !strings.Contains(string(encoded), want) {
			t.Errorf("json output missing %q: %s", want, encoded)
		}
	}
}
