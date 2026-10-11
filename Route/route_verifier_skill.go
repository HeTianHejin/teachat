package route

import (
	"fmt"
	"net/http"
	"strings"

	dao "teachat/DAO"
)

func HandleVerifierSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	user, err := verifierUser(r)
	if err != nil {
		if user == nil {
			http.Redirect(w, r, "/v1/login", http.StatusFound)
		} else {
			report(w, *user, "你好，您没有权限查看技能审核列表。")
		}
		return
	}

	skills, err := dao.GetPendingSkills(r.Context())
	if err != nil {
		report(w, *user, "获取待审核技能失败，请稍后重试。")
		return
	}
	pageData := struct {
		SessUser dao.User
		Skills   []dao.Skill
	}{SessUser: *user, Skills: skills}
	generateHTML(w, &pageData, "layout", "navbar.private", "verifier.skills")
}

func HandleVerifierSkillReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	user, err := verifierUser(r)
	if err != nil {
		if user == nil {
			http.Redirect(w, r, "/v1/login", http.StatusFound)
		} else {
			report(w, *user, "你好，您没有权限审核技能。")
		}
		return
	}
	if err = r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	skill := dao.Skill{Uuid: strings.TrimSpace(r.PostFormValue("uuid"))}
	if skill.Uuid == "" || skill.GetByIdOrUUID(r.Context()) != nil {
		report(w, *user, "未能找到指定的技能记录。")
		return
	}
	if skill.ReviewStatus != dao.PendingSkillReviewStatus {
		report(w, *user, "该技能已被处理或不在待审核状态，请刷新列表。")
		return
	}

	var approved bool
	var reason string
	switch r.PostFormValue("action") {
	case "approve":
		approved = true
	case "reject":
		reason = strings.TrimSpace(r.PostFormValue("reason"))
		if reason == "" {
			report(w, *user, "婉拒展示时必须填写原因。")
			return
		}
	default:
		report(w, *user, "不支持的技能审核操作。")
		return
	}

	if err = skill.Review(r.Context(), user.Id, approved, reason); err != nil {
		report(w, *user, fmt.Sprintf("技能审核失败：%v", err))
		return
	}
	http.Redirect(w, r, "/v1/verifier/skills", http.StatusSeeOther)
}
