package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/see-seek/new
func HandleNewSeeSeek(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SeeSeekNewGet(w, r)
	case http.MethodPost:
		SeeSeekNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func SeeSeekNewPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// GET /v1/see-seek/new?uuid=xXx
func SeeSeekNewGet(w http.ResponseWriter, r *http.Request) {

	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	vals := r.URL.Query()

	uuid := vals.Get("uuid")

	t_thread, err := data.GetThreadByUUID(uuid)
	if err != nil {
		util.Debug(" Cannot get thread given uuid ", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	if t_thread.Category != data.ThreadCategorySeeSeek {
		util.Debug(" this Thread category is not a see-seek", t_thread.Id, t_thread.Category)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	// 检测当前会话茶友是否见证者
	is_verifier := isVerifier(s_u.Id)
	if !is_verifier {
		util.Debug(" Current user is not a verifier", s_u.Id)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	t_proj, err := t_thread.Project()
	if err != nil {
		util.Debug(" Cannot get project given proj_id", t_thread.Id, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	t_obje, err := t_proj.Objective()
	if err != nil {
		util.Debug(" Cannot get objective given proj_id", t_thread.Id, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	projBean, err := fetchProjectBean(t_proj)
	if err != nil {
		util.Debug(" Cannot get projBean", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	objeBean, err := fetchObjectiveBean(t_obje)
	if err != nil {
		util.Debug(" Cannot get objeBean", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// master_team, err := data.GetTeam(t_proj.TeamId)
	// if err != nil {
	// 	util.Debug(" Cannot get master team", err)
	// 	Report(w, r, "你好，假作真时真亦假，无为有处有还无？")
	// 	return
	// }

	// admin_team, err := data.GetTeam(t_obje.TeamId)
	// if err != nil {
	// 	util.Debug(" Cannot get ob_admin team", t_obje.TeamId, err)
	// 	Report(w, r, "你好，假作真时真亦假，无为有处有还无？")
	// 	return
	// }

	sessUserData, err := prepareUserPageData(&sess)
	if err != nil {
		util.Debug(" Cannot prepare user page data", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var sSDpD data.SeeSeekDetailPageData
	sSDpD.ProjectBean = projBean
	sSDpD.ObjectiveBean = objeBean
	sSDpD.SessUser = sessUserData.User
	sSDpD.IsMaster = true

	// sSDpD.Admin = admin
	// sSDpD.AdminDefaultFamily = a_default_family
	// sSDpD.AdminSurvivalFamilies = a_survival_families
	// sSDpD.AdminDefaultTeam = a_default_team
	// sSDpD.AdminSurvivalTeams = a_survival_teams
	// sSDpD.Master = master
	// sSDpD.MasterDefaultFamily = m_default_family
	// sSDpD.MasterSurvivalFamilies = m_survival_families
	// sSDpD.MasterDefaultTeam = m_default_team
	// sSDpD.MasterSurvivalTeams = m_survival_teams

	renderHTML(w, &sSDpD, "layout", "navbar.private", "see-seek.new")
}
