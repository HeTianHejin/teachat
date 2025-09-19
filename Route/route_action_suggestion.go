package route

import (
	"database/sql"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

func HandleNewSuggestion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SuggestionNewGet(w, r)
	case http.MethodPost:
		SuggestionNewPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/suggestion/new?uuid=xxx
// 新建"建议"记录页面
func SuggestionNewGet(w http.ResponseWriter, r *http.Request) {
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

	// 检查提交者身份是否见证者
	if !isVerifier(s_u.Id) {
		report(w, r, "只有见证者才可以创建建议记录")
		return
	}

	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	if uuid == "" {
		util.Debug(" No uuid provided in query", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	t_proj := data.Project{Uuid: uuid}
	if err := t_proj.GetByUuid(); err != nil {
		util.Debug(" Cannot get project by uuid", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 检查是否已存在当前project_id的suggestion记录
	existingSuggestion, err := data.GetSuggestionByProjectId(t_proj.Id, r.Context())
	if err != nil && err != sql.ErrNoRows {
		util.Debug(" Cannot get existing suggestion", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	if err == nil && existingSuggestion.Id > 0 {
		// 已存在建议记录,报告提示不要重复创建
		report(w, r, "该项目的建议记录已存在，请勿重复创建")
		return
	}

	t_obje, err := t_proj.Objective()
	if err != nil {
		util.Debug(" Cannot get objective given proj_id", t_proj.Id, err)
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

	// 准备页面数据
	var suggestionData data.SuggestionDetailTemplateData
	suggestionData.SessUser = s_u
	suggestionData.ProjectBean = projBean
	suggestionData.QuoteObjectiveBean = objeBean

	generateHTML(w, &suggestionData, "layout", "navbar.private", "action.suggestion.new", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/suggestion/new
func SuggestionNewPost(w http.ResponseWriter, r *http.Request) {
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
	// 检查提交者身份是否见证者
	if !isVerifier(s_u.Id) {
		report(w, r, "只有见证者才可以创建建议记录")
		return
	}
	// 解析表单数据
	if err := r.ParseForm(); err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "表单数据解析失败")
		return
	}

	// 获取项目信息
	projectUuid := r.FormValue("project_uuid")
	if projectUuid == "" {
		report(w, r, "项目信息缺失")
		return
	}

	t_proj := data.Project{Uuid: projectUuid}
	if err := t_proj.GetByUuid(); err != nil {
		util.Debug(" Cannot get project by uuid", projectUuid, err)
		report(w, r, "项目不存在")
		return
	}

	// 检查是否已存在当前project_id的suggestion记录
	existingSuggestion, err := data.GetSuggestionByProjectId(t_proj.Id, r.Context())
	if err != nil && err != sql.ErrNoRows {
		util.Debug(" Cannot get existing suggestion", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	if err == nil && existingSuggestion.Id > 0 {
		if existingSuggestion.Status == int(data.SuggestionStatusSubmitted) {
			report(w, r, "该项目的建议记录已提交，不能重复创建")
			return
		}
	}

	// 获取表单字段
	body := r.FormValue("body")
	resolutionStr := r.FormValue("resolution")
	categoryStr := r.FormValue("category")

	// 验证必填字段
	if body == "" {
		report(w, r, "请填写建议内容")
		return
	}

	// 转换数值字段
	resolution := resolutionStr == "true"
	category, _ := strconv.Atoi(categoryStr)

	// 创建Suggestion记录
	suggestion := data.Suggestion{
		UserId:     s_u.Id,
		ProjectId:  t_proj.Id,
		Resolution: resolution,
		Body:       body,
		Category:   category,
		Status:     int(data.SuggestionStatusSubmitted), // 已提交
	}

	if err := suggestion.Create(r.Context()); err != nil {
		util.Debug(" Cannot create suggestion", err)
		report(w, r, "创建建议记录失败")
		return
	}
	// 反馈结果到茶台项目状态，如果Resolution的值为false，搁置，更新项目状态
	if !resolution {
		t_proj.Status = int(data.ProjectStatusTeaCold)
		if err := t_proj.Update(); err != nil {
			util.Debug(" Cannot update project status to TeaCold", err)
			report(w, r, "更新项目状态失败")
			return
		}
	}

	// 重定向到详情页面
	http.Redirect(w, r, "/v1/suggestion/detail?uuid="+suggestion.Uuid, http.StatusFound)
}

// GET /v1/suggestion/detail?uuid=xxx
func SuggestionDetail(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 获取Suggestion记录
	suggestion := data.Suggestion{Uuid: uuid}
	if err := suggestion.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			// 尝试project的uuid
			project := data.Project{Uuid: uuid}
			if err := project.GetByUuid(); err != nil {
				util.Debug("Cannot get project by uuid", uuid, err)
				report(w, r, "你好，假作真时真亦假，无为有处有还无？")
				return
			}
			suggestion, err = data.GetSuggestionByProjectId(project.Id, r.Context())
			if err != nil {
				if err == sql.ErrNoRows {
					report(w, r, "该项目还没有建议记录")
					return
				}
				util.Debug("Cannot get Suggestion by project_id", project.Id, err)
				report(w, r, "该项目的建议记录似乎被茶水泡糊了")
				return
			}
		} else {
			util.Debug("Cannot get Suggestion by uuid", uuid, err)
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
	}

	// 获取项目信息
	pr := data.Project{Id: suggestion.ProjectId}
	if err := pr.Get(); err != nil {
		util.Debug("Cannot get project", err)
		report(w, r, "获取项目信息失败")
		return
	}

	// 获取目标信息
	ob, err := pr.Objective()
	if err != nil {
		util.Debug("Cannot get objective", err)
		report(w, r, "获取目标信息失败")
		return
	}

	// 获取完整的SuggestionBean
	suggestionBean, err := fetchSuggestionBean(suggestion)
	if err != nil {
		util.Debug("Cannot fetch Suggestion bean", err)
		report(w, r, "获取建议记录详情失败")
		return
	}

	projectBean, err := fetchProjectBean(pr)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, r, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(ob)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, r, "获取目标详情失败")
		return
	}

	// 准备页面数据
	templateData := data.SuggestionDetailTemplateData{
		SessUser:           s_u,
		SuggestionBean:     suggestionBean,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
	}

	// 权限检查
	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("Admin permission check failed", "userId", s_u.Id, "objectiveId", ob.Id, "error", err)
		report(w, r, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	templateData.IsAdmin = is_admin

	if !is_admin {
		is_master, err := checkProjectMasterPermission(&pr, s_u.Id)
		if err != nil {
			util.Debug("Permission check failed", "user_id:", s_u.Id, "error:", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
			return
		}
		templateData.IsMaster = is_master
	}

	if !is_admin && !templateData.IsMaster {
		is_verifier := isVerifier(s_u.Id)
		templateData.IsVerifier = is_verifier
	}

	if suggestion.Category == int(data.SuggestionCategoryPrivate) {

		is_invited, err := ob.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug("Cannot check if user is invited to objective", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
			return
		}
		templateData.IsInvited = is_invited
	}

	// 检测私密建议的访问权限
	if suggestion.Id > 0 && suggestion.Category == int(data.SuggestionCategoryPrivate) {
		if !is_admin && !templateData.IsMaster && !templateData.IsVerifier && !templateData.IsInvited {
			util.Debug("User has no access to this private suggestion", "user_id:", s_u.Id, "suggestion_id:", suggestion.Id)
			report(w, r, "你没有权限查看此私密建议记录")
			return
		}
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "action.suggestion.detail", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}
