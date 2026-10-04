package route

import (
	"fmt"
	"net/http"
	"strings"

	dao "teachat/DAO"
)

type verifierOfferingSkill struct {
	Name      string
	Level     int
	IsPrimary bool
}

// HandleVerifierServiceOfferingDetail 展示待审核服务项目详情。
func HandleVerifierServiceOfferingDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	user, err := verifierUser(r)
	if err != nil {
		if user == nil {
			http.Redirect(w, r, "/v1/login", http.StatusFound)
		} else {
			report(w, *user, "你好，您没有权限查看服务项目审核详情。")
		}
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, *user, "你好，未能找到指定的服务项目。")
		return
	}
	offering := &dao.TeamServiceOffering{Uuid: uuid}
	if err = offering.GetByIdOrUUID(r.Context()); err != nil {
		report(w, *user, "你好，未能找到指定的服务项目。")
		return
	}
	team, err := dao.GetTeam(offering.TeamId)
	if err != nil {
		report(w, *user, "你好，未能获取服务项目所属团队。")
		return
	}

	var skills []verifierOfferingSkill
	teamSkills, err := dao.GetTeamServiceOfferingSkillsByOfferingId(offering.Id, r.Context())
	if err != nil {
		report(w, *user, "你好，未能获取服务项目的技能要求。")
		return
	}
	for _, teamSkill := range teamSkills {
		skill := &dao.Skill{}
		if err = skill.GetById(teamSkill.SkillId, r.Context()); err != nil {
			report(w, *user, "你好，未能获取服务项目的技能名称。")
			return
		}
		skills = append(skills, verifierOfferingSkill{
			Name: skill.Name, Level: teamSkill.RequiredLevel, IsPrimary: teamSkill.IsPrimary,
		})
	}
	events, err := dao.GetServiceOfferingEventsByOfferingId(offering.Id, r.Context())
	if err != nil {
		report(w, *user, "你好，未能获取服务项目的审核记录。")
		return
	}

	pageData := struct {
		SessUser  dao.User
		Offering  *dao.TeamServiceOffering
		Team      dao.Team
		Skills    []verifierOfferingSkill
		Events    []*dao.ServiceOfferingEvent
		IsPending bool
	}{
		SessUser: *user, Offering: offering, Team: team, Skills: skills, Events: events,
		IsPending: offering.Status == dao.PendingTeamServiceOfferingStatus,
	}
	generateHTML(w, &pageData, "layout", "navbar.private", "verifier.service_offering.detail")
}

// HandleVerifierServiceOfferingReview 接收见证者的通过或婉拒决定。
func HandleVerifierServiceOfferingReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	user, err := verifierUser(r)
	if err != nil {
		if user == nil {
			http.Redirect(w, r, "/v1/login", http.StatusFound)
		} else {
			report(w, *user, "你好，您没有权限审核服务项目。")
		}
		return
	}
	if err = r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	uuid := r.PostFormValue("uuid")
	action := r.PostFormValue("action")
	if uuid == "" {
		report(w, *user, "你好，未能找到指定的服务项目。")
		return
	}
	offering := &dao.TeamServiceOffering{Uuid: uuid}
	if err = offering.GetByIdOrUUID(r.Context()); err != nil {
		report(w, *user, "你好，未能找到指定的服务项目。")
		return
	}
	if offering.Status != dao.PendingTeamServiceOfferingStatus {
		report(w, *user, "该服务项目已被处理，未重复审批。请返回待审核列表查看最新状态。")
		return
	}

	switch action {
	case "approve":
		err = offering.Approve(r.Context(), user.Id, user.Id)
	case "reject":
		reason := strings.TrimSpace(r.PostFormValue("reason"))
		if reason == "" {
			report(w, *user, "请填写婉拒原因。")
			return
		}
		err = offering.Reject(r.Context(), user.Id, reason)
	default:
		report(w, *user, "不支持的服务项目审核操作。")
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "已被其他操作变更") {
			report(w, *user, "该服务项目已被其他见证者处理，未重复审批。请返回列表刷新。")
			return
		}
		report(w, *user, fmt.Sprintf("服务项目审核失败：%v", err))
		return
	}
	http.Redirect(w, r, "/v1/verifier/workspace?tab=services", http.StatusSeeOther)
}

func verifierUser(r *http.Request) (*dao.User, error) {
	sess, err := session(r)
	if err != nil {
		return nil, err
	}
	user, err := sess.User()
	if err != nil {
		return nil, err
	}
	if !dao.IsVerifier(user.Id) {
		return &user, fmt.Errorf("not a verifier")
	}
	return &user, nil
}
