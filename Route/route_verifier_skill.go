package route

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

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

// verifierSkillItem 是待审核技能的 JSON 表示。
// GetPendingSkills 未查询 review_status，且 dao.Skill 没有 json tag，
// 因此这里显式给出稳定的 snake_case 字段与 pending 状态，避免客户端读到零值。
type verifierSkillItem struct {
	Id                int       `json:"id"`
	Uuid              string    `json:"uuid"`
	UserId            int       `json:"user_id"`
	Name              string    `json:"name"`
	Nickname          string    `json:"nickname"`
	Description       string    `json:"description"`
	Category          int       `json:"category"`
	CategoryLabel     string    `json:"category_label"`
	StrengthLevel     int       `json:"strength_level"`
	DifficultyLevel   int       `json:"difficulty_level"`
	Level             int       `json:"level"`
	ReviewStatus      int       `json:"review_status"`
	ReviewStatusLabel string    `json:"review_status_label"`
	ReviewReason      string    `json:"review_reason"`
	CreatedAt         time.Time `json:"created_at"`
}

func newVerifierSkillItem(s dao.Skill) verifierSkillItem {
	return verifierSkillItem{
		Id:                s.Id,
		Uuid:              s.Uuid,
		UserId:            s.UserId,
		Name:              s.Name,
		Nickname:          s.Nickname,
		Description:       s.Description,
		Category:          int(s.Category),
		CategoryLabel:     s.CategoryString(),
		StrengthLevel:     int(s.StrengthLevel),
		DifficultyLevel:   int(s.DifficultyLevel),
		Level:             s.Level,
		ReviewStatus:      int(dao.PendingSkillReviewStatus),
		ReviewStatusLabel: dao.PendingSkillReviewStatus.String(),
		ReviewReason:      s.ReviewReason,
		CreatedAt:         s.CreatedAt,
	}
}

type verifierSkillReviewRequest struct {
	Uuid   string `json:"uuid"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// verifierAPIUser 校验 JSON 端点的登录状态与见证者身份，失败时已输出 JSON 错误响应。
func verifierAPIUser(w http.ResponseWriter, r *http.Request) (*dao.User, bool) {
	user, err := verifierUser(r)
	if err != nil {
		if user == nil {
			respondWithError(w, http.StatusUnauthorized, "请先登录")
		} else {
			respondWithError(w, http.StatusForbidden, "您不是见证者团队成员，无法处理技能审核")
		}
		return nil, false
	}
	return user, true
}

// HandleVerifierSkillsAPI 以 JSON 返回待审核技能列表，供见证者团队的 AGI 登录后读取。
func HandleVerifierSkillsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}
	if _, ok := verifierAPIUser(w, r); !ok {
		return
	}

	skills, err := dao.GetPendingSkills(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待审核技能失败，请稍后重试")
		return
	}
	items := make([]verifierSkillItem, 0, len(skills))
	for i := range skills {
		items = append(items, newVerifierSkillItem(skills[i]))
	}
	respondWithSuccess(w, "获取待审核技能成功", map[string]any{
		"skills": items,
		"total":  len(items),
	})
}

// HandleVerifierSkillReviewAPI 以 JSON 接收审核决定，复用 dao.Skill.Review 完成判定。
// 请求体示例：{"uuid":"...","action":"approve"} 或 {"uuid":"...","action":"reject","reason":"..."}
func HandleVerifierSkillReviewAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}
	user, ok := verifierAPIUser(w, r)
	if !ok {
		return
	}

	var req verifierSkillReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	req.Uuid = strings.TrimSpace(req.Uuid)
	req.Action = strings.TrimSpace(req.Action)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Uuid == "" {
		respondWithError(w, http.StatusBadRequest, "技能UUID不能为空")
		return
	}

	var approved bool
	switch req.Action {
	case "approve":
		approved = true
	case "reject":
		if req.Reason == "" {
			respondWithError(w, http.StatusBadRequest, "婉拒展示时必须填写原因")
			return
		}
		if utf8.RuneCountInString(req.Reason) > 500 {
			respondWithError(w, http.StatusBadRequest, "婉拒原因不能超过500字")
			return
		}
	default:
		respondWithError(w, http.StatusBadRequest, "不支持的技能审核操作，action 只能为 approve 或 reject")
		return
	}

	skill := dao.Skill{Uuid: req.Uuid}
	if err := skill.GetByIdOrUUID(r.Context()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "未能找到指定的技能记录")
		} else {
			respondWithError(w, http.StatusInternalServerError, "读取技能记录失败，请稍后重试")
		}
		return
	}
	if skill.ReviewStatus != dao.PendingSkillReviewStatus {
		respondWithError(w, http.StatusConflict, "该技能已被处理或不在待审核状态，请刷新列表")
		return
	}
	if err := skill.Review(r.Context(), user.Id, approved, req.Reason); err != nil {
		respondWithError(w, http.StatusConflict, fmt.Sprintf("技能审核失败：%v", err))
		return
	}
	respondWithSuccess(w, "技能审核完成", map[string]any{
		"uuid":                skill.Uuid,
		"review_status":       int(skill.ReviewStatus),
		"review_status_label": skill.ReviewStatus.String(),
		"review_reason":       skill.ReviewReason,
	})
}
