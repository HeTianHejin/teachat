package route

import (
	"net/http"
	dao "teachat/DAO"
)

// GET /v1/team/service_offerings/unpublished?uuid=
// 显示团队已下架的服务项目
func TeamUnpublishedServiceOfferings(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，请提交有效的茶团编号，请稍后再试。")
		return
	}
	team, err := dao.GetTeamByID(uuid)
	if err != nil {
		report(w, s_u, "你好，茶博士未能帮忙查看这个茶团资料，请稍后再试。")
		return
	}
	if team.IsPrivate {
		isMember, err := team.IsActiveMember(s_u.Id)
		if err != nil || !isMember {
			report(w, s_u, "你好，这个茶团是私有的，你不能查看。")
			return
		}
	}

	teamBean, err := fetchTeamBean(team)
	if err != nil {
		report(w, s_u, "你好，茶博士未能帮忙查看这个茶团资料，请稍后再试。")
		return
	}
	offerings, err := dao.GetTeamServiceOfferingsByTeamId(team.Id, false, r.Context())
	if err != nil {
		report(w, s_u, "你好，茶博士未能帮忙查看已下架服务项目，请稍后再试。")
		return
	}
	unpublished := make([]*dao.TeamServiceOffering, 0)
	for _, offering := range offerings {
		if offering.Status == dao.UnpublishedTeamServiceOfferingStatus {
			unpublished = append(unpublished, offering)
		}
	}

	pageData := struct {
		SessUser  dao.User
		TeamBean  dao.TeamBean
		Offerings []*dao.TeamServiceOffering
	}{SessUser: s_u, TeamBean: teamBean, Offerings: unpublished}
	generateHTML(w, &pageData, "layout", "navbar.private", "team.service_offerings_unpublished")
}
