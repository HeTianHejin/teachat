package route

import (
	"net/http"
	"strconv"
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

// GET /v1/see-seek/new?id=xXx&admin_id=xXx&master_id=xXx
func SeeSeekNewGet(w http.ResponseWriter, r *http.Request) {

	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	vals := r.URL.Query()
	place_id_str := vals.Get("place_id")
	place_id_int, err := strconv.Atoi(place_id_str)
	if err != nil {
		Report(w, r, "你好，茶博士无法理解提交的id资料，请确认后再试。")
		return
	}
	admin_id_str := vals.Get("admin_id")
	admin_id_int, err := strconv.Atoi(admin_id_str)
	if err != nil {
		util.Debug(" Cannot convert admin_id to int", err)
		Report(w, r, "你好，茶博士无法理解提交的id资料，请确认后再试。")
		return
	}
	if admin_id_int == data.UserId_None || admin_id_int == data.UserId_SpaceshipCaptain {
		Report(w, r, "你好，茶博士表示惊讶，特殊成员资格检查未通过。")
		return
	}
	master_id_str := vals.Get("master_id")
	master_id_int, err := strconv.Atoi(master_id_str)
	if err != nil {
		util.Debug(" Cannot convert master_id to int", err)
		Report(w, r, "你好，茶博士无法理解提交的id资料，请确认后再试。")
		return
	}
	if master_id_int == data.UserId_None || master_id_int == data.UserId_SpaceshipCaptain {
		Report(w, r, "你好，茶博士表示惊讶，特殊成员资格检查未通过。")
		return
	}
	uuid := vals.Get("id")
	place := data.Place{Id: place_id_int}
	if err = place.Get(); err != nil {
		util.Debug(" Cannot get place given id ", place_id_int, err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	t_proj := data.Project{Uuid: uuid}
	if err := t_proj.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	if t_proj.TeamId == data.TeamIdFreelancer {
		Report(w, r, "你好，自由人特殊团队不提供看看服务。")
		return
	}

	master_team, err := data.GetTeam(t_proj.TeamId)
	if err != nil {
		util.Debug(" Cannot get master team", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	is_member, err := master_team.IsMember(master_id_int)
	if err != nil {
		util.Debug(" Cannot check master-team-member given team_id,s_u.Email", master_team.Id, s_u.Email, err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议团队会员资格资料。")
		return
	}
	// if not member, return
	if !is_member {
		Report(w, r, "你好，茶博士表示惊讶，管理成员资格检查未通过。")
		return
	}
	t_obje, err := t_proj.Objective()
	if err != nil {
		util.Debug(" Cannot get objective given proj_id", t_proj.Id, err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	projBean, err := FetchProjectBean(t_proj)
	if err != nil {
		util.Debug(" Cannot get projBean", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	objeBean, err := FetchObjectiveBean(t_obje)
	if err != nil {
		util.Debug(" Cannot get objeBean", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	admin_team, err := data.GetTeam(t_obje.TeamId)
	if err != nil {
		util.Debug(" Cannot get ob_admin team", t_obje.TeamId, err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	is_member, err = admin_team.IsMember(admin_id_int)
	if err != nil {
		util.Debug(" Cannot check admin-team-member given team_id, admin_id", admin_team.Id, admin_id_int, err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议管理团队会员资格资料。")
		return
	}
	// if not member, return
	if !is_member {
		Report(w, r, "你好，茶博士表示惊讶，茶话会管理成员资格检查未通过。")
		return
	}

	admin, err := data.GetUser(admin_id_int)
	if err != nil {
		util.Debug(" Cannot get ob_admin given userid", admin_id_int, err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	master, err := data.GetUser(master_id_int)
	if err != nil {
		util.Debug(" Cannot get master given userid", master_id_int, err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}

	admin, a_default_family, a_survival_families, a_default_team, a_survival_teams, _, _, err := fetchUserRelatedData(admin)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", admin.Email, err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	master, m_default_family, m_survival_families, m_default_team, m_survival_teams, _, _, err := fetchUserRelatedData(master)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", master.Email, err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	sessUserData, err := prepareUserPageData(&sess)
	if err != nil {
		util.Debug(" Cannot prepare user page data", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}

	var sSDpD data.SeeSeekDetailPageData
	sSDpD.SessUser = sessUserData.User
	sSDpD.IsAdmin = false
	sSDpD.IsMaster = true
	sSDpD.IsGuest = false
	sSDpD.IsInvited = false
	sSDpD.SessUserDefaultFamily = sessUserData.DefaultFamily
	sSDpD.SessUserSurvivalFamilies = sessUserData.SurvivalFamilies
	sSDpD.SessUserDefaultTeam = sessUserData.DefaultTeam
	sSDpD.SessUserSurvivalTeams = sessUserData.SurvivalTeams
	sSDpD.SessUserDefaultPlace = sessUserData.DefaultPlace
	sSDpD.SessUserBindPlaces = sessUserData.BindPlaces
	sSDpD.Admin = admin
	sSDpD.AdminDefaultFamily = a_default_family
	sSDpD.AdminSurvivalFamilies = a_survival_families
	sSDpD.AdminDefaultTeam = a_default_team
	sSDpD.AdminSurvivalTeams = a_survival_teams
	sSDpD.Master = master
	sSDpD.MasterDefaultFamily = m_default_family
	sSDpD.MasterSurvivalFamilies = m_survival_families
	sSDpD.MasterDefaultTeam = m_default_team
	sSDpD.MasterSurvivalTeams = m_survival_teams

	sSDpD.ProjectBean = projBean
	sSDpD.ObjectiveBean = objeBean

	RenderHTML(w, &sSDpD, "layout", "navbar.private", "see-seek.new")
}
