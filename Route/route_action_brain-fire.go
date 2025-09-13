package route

import (
	"database/sql"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
)

// Handler /v1/brain-fire/new
func HandleNewBrainFire(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		BrainFireNewGet(w, r)
	case http.MethodPost:
		BrainFireNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/brain-fire/new?uuid=xxx
// 新建"脑火"记录页面
func BrainFireNewGet(w http.ResponseWriter, r *http.Request) {
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
	if uuid == "" {
		util.Debug(" No uuid provided in query")
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	t_proj := data.Project{Uuid: uuid}
	if err := t_proj.GetByUuid(); err != nil {
		util.Debug(" Cannot get project by uuid", uuid, err)
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

	// 检查是否已存在当前project_id的brain-fire记录
	existingBrainFire, err := data.GetBrainFireByProjectId(t_proj.Id, r.Context())
	if err != nil && err != sql.ErrNoRows {
		util.Debug(" Cannot get existing brain-fire", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	if err == nil && existingBrainFire.Id > 0 {
		if existingBrainFire.Status == data.BrainFireStatusExtinguished {
			report(w, r, "该项目的脑火记录已完成，不能重复创建")
			return
		}

	}

	//读取茶台的"约茶"资料
	proj_appointment, err := data.GetAppointmentByProjectId(t_proj.Id, r.Context())
	if err != nil {
		util.Debug(" Cannot get project appointment", t_proj.Id, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	proj_appointment_bean, err := fetchAppointmentBean(proj_appointment)
	if err != nil {
		util.Debug(" Cannot get project appointment bean", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
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

	//读取预设的环境条件
	environments, err := data.GetDefaultEnvironments(r.Context())
	if err != nil {
		util.Debug(" Cannot get default environments", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 准备页面数据
	var bfDtD data.BrainFireDetailTemplateData
	bfDtD.SessUser = s_u
	bfDtD.IsVerifier = is_verifier

	bfDtD.Verifier = s_u
	bfDtD.VerifierFamily = proj_appointment_bean.VerifierFamily
	bfDtD.VerifierTeam = proj_appointment_bean.VerifierTeam
	bfDtD.Payer = proj_appointment_bean.Payer
	bfDtD.PayerFamily = proj_appointment_bean.PayerFamily
	bfDtD.PayerTeam = proj_appointment_bean.PayerTeam
	bfDtD.Payee = proj_appointment_bean.Payee
	bfDtD.PayeeFamily = proj_appointment_bean.PayeeFamily
	bfDtD.PayeeTeam = proj_appointment_bean.PayeeTeam

	bfDtD.ProjectBean = projBean
	bfDtD.QuoteObjectiveBean = objeBean
	bfDtD.ProjectAppointment = proj_appointment_bean
	bfDtD.Environments = environments

	renderHTML(w, &bfDtD, "layout", "navbar.private", "action.brain-fire.new", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/brain-fire/new
func BrainFireNewPost(w http.ResponseWriter, r *http.Request) {
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

	// 检测当前会话茶友是否见证者
	is_verifier := isVerifier(s_u.Id)
	if !is_verifier {
		util.Debug(" Current user is not a verifier", s_u.Id)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
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
	// 检查是否已存在当前project_id的brain-fire记录
	existingBrainFire, err := data.GetBrainFireByProjectId(t_proj.Id, r.Context())
	if err != nil && err != sql.ErrNoRows {
		util.Debug(" Cannot get existing brain-fire", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	if err == nil && existingBrainFire.Id > 0 {
		if existingBrainFire.Status >= data.BrainFireStatusBurning {
			report(w, r, "该项目的脑火记录已存在，不能重复创建")
			return
		}

	}

	// 获取表单字段
	title := r.FormValue("title")
	inference := r.FormValue("inference")
	diagnose := r.FormValue("diagnose")
	judgement := r.FormValue("judgement")
	startTimeStr := r.FormValue("start_time")
	endTimeStr := r.FormValue("end_time")
	environmentIdStr := r.FormValue("environment_id")
	brainFireClassStr := r.FormValue("brain_fire_class")
	brainFireTypeStr := r.FormValue("brain_fire_type")

	// 验证必填字段
	if title == "" || inference == "" || diagnose == "" || judgement == "" {
		report(w, r, "请填写完整的脑火内容")
		return
	}

	// 转换数值字段
	environmentId, _ := strconv.Atoi(environmentIdStr)
	brainFireClass, _ := strconv.Atoi(brainFireClassStr)
	brainFireType, _ := strconv.Atoi(brainFireTypeStr)

	// 解析时间
	startTime, err := time.Parse("2006-01-02T15:04", startTimeStr)
	if err != nil {
		report(w, r, "开始时间格式错误")
		return
	}
	endTime, err := time.Parse("2006-01-02T15:04", endTimeStr)
	if err != nil {
		report(w, r, "结束时间格式错误")
		return
	}

	// 获取参与方ID
	payerUserId, _ := strconv.Atoi(r.FormValue("payer_user_id"))
	payerTeamId, _ := strconv.Atoi(r.FormValue("payer_team_id"))
	payerFamilyId, _ := strconv.Atoi(r.FormValue("payer_family_id"))
	payeeUserId, _ := strconv.Atoi(r.FormValue("payee_user_id"))
	payeeTeamId, _ := strconv.Atoi(r.FormValue("payee_team_id"))
	payeeFamilyId, _ := strconv.Atoi(r.FormValue("payee_family_id"))

	// 创建BrainFire记录
	brainFire := data.BrainFire{
		ProjectId:        t_proj.Id,
		StartTime:        startTime,
		EndTime:          endTime,
		EnvironmentId:    environmentId,
		Title:            title,
		Inference:        inference,
		Diagnose:         diagnose,
		Judgement:        judgement,
		PayerUserId:      payerUserId,
		PayerTeamId:      payerTeamId,
		PayerFamilyId:    payerFamilyId,
		PayeeUserId:      payeeUserId,
		PayeeTeamId:      payeeTeamId,
		PayeeFamilyId:    payeeFamilyId,
		VerifierUserId:   s_u.Id,
		VerifierFamilyId: data.FamilyIdUnknown,
		VerifierTeamId:   data.TeamIdVerifier,
		Status:           data.BrainFireStatusBurning, // 燃烧中
		BrainFireClass:   brainFireClass,
		BrainFireType:    brainFireType,
	}

	if err := brainFire.Create(r.Context()); err != nil {
		util.Debug(" Cannot create brain fire", err)
		report(w, r, "创建脑火记录失败")
		return
	}

	// 重定向到详情页面
	http.Redirect(w, r, "/v1/brain-fire/detail?uuid="+brainFire.Uuid, http.StatusFound)
}

// Handler /v1/brain-fire/detail
func HandleBrainFireDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	BrainFireDetailGet(w, r)
}

// GET /v1/brain-fire/detail?uuid=xxx
func BrainFireDetailGet(w http.ResponseWriter, r *http.Request) {
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

	// 获取BrainFire记录
	brainFire := data.BrainFire{Uuid: uuid}
	if err := brainFire.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			// 尝试project的uuid
			project := data.Project{Uuid: uuid}
			if err := project.GetByUuid(); err != nil {
				util.Debug("Cannot get project by uuid", uuid, err)
				report(w, r, "你好，假作真时真亦假，无为有处有还无？")
				return
			}
			brainFire, err = data.GetBrainFireByProjectId(project.Id, r.Context())
			if err != nil {
				if err == sql.ErrNoRows {
					report(w, r, "该项目还没有脑火记录")
					return
				}
				util.Debug("Cannot get BrainFire by project_id", project.Id, err)
				report(w, r, "该项目脑火记录似乎被茶水泡糊了")
				return
			}
		} else {
			util.Debug("Cannot get BrainFire by uuid", uuid, err)
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
	}

	// 获取项目信息
	pr := data.Project{Id: brainFire.ProjectId}
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

	// 获取完整的BrainFireBean
	brainFireBean, err := fetchBrainFireBean(brainFire)
	if err != nil {
		util.Debug("Cannot fetch BrainFire bean", err)
		report(w, r, "获取脑火记录详情失败")
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
	templateData := data.BrainFireDetailTemplateData{
		SessUser:           s_u,
		BrainFireBean:      brainFireBean,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
	}
	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("Admin permission check failed",
			"userId", s_u.Id,
			"objectiveId", ob.Id,
			"error", err,
		)
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
	if ob.Class == data.ObClassClose {
		is_invited, err := ob.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug("Cannot check if user is invited to objective", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
			return
		}
		templateData.IsInvited = is_invited
	}
	// 检测私密脑火的访问权限
	if brainFire.Id > 0 && brainFire.BrainFireType == data.BrainFireTypePrivate {
		if !is_admin && !templateData.IsMaster && !templateData.IsVerifier && !templateData.IsInvited {
			util.Debug("User has no access to this private brain-fire", "user_id:", s_u.Id, "brain_fire_id:", brainFire.Id)
			report(w, r, "你没有权限查看此私密脑火记录")
			return
		}

	}

	renderHTML(w, &templateData, "layout", "navbar.private", "action.brain-fire.detail", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}
