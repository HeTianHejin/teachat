package route

import (
	"database/sql"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
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

// POST /v1/see-seek/new
func SeeSeekNewPost(w http.ResponseWriter, r *http.Request) {
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

	// 尝试读取是否已经存在seeseek记录
	existingSeeSeek, err := data.GetSeeSeekByProjectId(t_proj.Id, r.Context())
	if err != nil && err != sql.ErrNoRows {
		util.Debug(" failed to check existing see-seek by project_id", err)
		report(w, r, "查询已有看看记录失败")
		return
	}
	if err == nil && existingSeeSeek.Id > 0 {
		// 已存在看看记录
		if existingSeeSeek.Status == data.SeeSeekStatusCompleted {
			report(w, r, "该项目的看看记录已完成，不能重复创建")
			return
		}
	}
	// 获取表单字段
	name := r.FormValue("name")
	nickname := r.FormValue("nickname")
	description := r.FormValue("description")
	startTimeStr := r.FormValue("start_time")
	placeIdStr := r.FormValue("place_id")
	environmentIdStr := r.FormValue("environment_id")
	categoryStr := r.FormValue("category")

	// 验证必填字段
	if name == "" || description == "" || startTimeStr == "" {
		report(w, r, "请填写完整的基本信息")
		return
	}

	// 解析开始时间
	startTime, err := time.Parse("2006-01-02T15:04", startTimeStr)
	if err != nil {
		util.Debug(" Cannot parse start time", startTimeStr, err)
		report(w, r, "开始时间格式不正确")
		return
	}

	// 转换数值字段
	placeId, _ := strconv.Atoi(placeIdStr)
	environmentId, _ := strconv.Atoi(environmentIdStr)
	category, _ := strconv.Atoi(categoryStr)

	// 验证环境条件必须选择
	if environmentId <= 0 {
		report(w, r, "请务必选择环境条件")
		return
	}
	// 尝试查找环境条件，测试是否存在该id的环境条件记录
	env := data.Environment{Id: environmentId}
	if err := env.GetByIdOrUUID(); err != nil {
		util.Debug(" Cannot get environment by id", environmentId, err)
		report(w, r, "环境条件不存在，请确认")
		return
	}

	// 获取参与方ID
	payerUserId, _ := strconv.Atoi(r.FormValue("payer_id"))
	payerTeamId, _ := strconv.Atoi(r.FormValue("payer_team_id"))
	payerFamilyId, _ := strconv.Atoi(r.FormValue("payer_family_id"))
	payeeUserId, _ := strconv.Atoi(r.FormValue("payee_id"))
	payeeTeamId, _ := strconv.Atoi(r.FormValue("payee_team_id"))
	payeeFamilyId, _ := strconv.Atoi(r.FormValue("payee_family_id"))

	// 创建SeeSeek记录
	seeSeek := data.SeeSeek{
		Name:             name,
		Nickname:         nickname,
		Description:      description,
		ProjectId:        t_proj.Id,
		PlaceId:          placeId,
		PayerUserId:      payerUserId,
		PayerTeamId:      payerTeamId,
		PayerFamilyId:    payerFamilyId,
		PayeeUserId:      payeeUserId,
		PayeeTeamId:      payeeTeamId,
		PayeeFamilyId:    payeeFamilyId,
		VerifierUserId:   s_u.Id,
		VerifierFamilyId: data.FamilyIdUnknown,
		VerifierTeamId:   data.TeamIdVerifier,
		Category:         category,
		Status:           data.SeeSeekStatusInProgress, // 进行中
		Step:             data.SeeSeekStepEnvironment,  // 步骤1：环境条件
		StartTime:        startTime,
	}

	if err := seeSeek.Create(r.Context()); err != nil {
		util.Debug(" Cannot create see seek", err)
		report(w, r, "创建看看记录失败")
		return
	}

	// 创建环境关联记录
	seeSeekEnv := data.SeeSeekEnvironment{
		SeeSeekId:     seeSeek.Id,
		EnvironmentId: environmentId,
	}
	if err := seeSeekEnv.Create(); err != nil {
		util.Debug(" Cannot create see seek environment", err)
		report(w, r, "创建环境关联记录失败")
		return
	}

	// 重定向到步骤页面
	http.Redirect(w, r, "/v1/see-seek/step2?uuid="+seeSeek.Uuid, http.StatusFound)
}

// GET /v1/see-seek/new?uuid=xXx
// 新建“看看”记录，从第一步开始创建
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

	//读取茶台的“约茶”资料
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
	// 检查是否已存在当前project_id的see-seek记录
	existingSeeSeek, err := data.GetSeeSeekByProjectId(t_proj.Id, r.Context())
	if err != nil && err != sql.ErrNoRows {
		util.Debug(" Cannot get existing see-seek", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	if err == nil && existingSeeSeek.Id > 0 {
		// 已存在看看记录,跳转到相应步骤
		url := "/v1/see-seek/step" + strconv.Itoa((existingSeeSeek.Step)) + "?uuid=" + existingSeeSeek.Uuid
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	//读取预设的4个通用场所环境,id为1,2,3,4
	environments, err := data.GetDefaultEnvironments(r.Context())
	if err != nil {
		util.Debug(" Cannot get default environments", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 准备页面数据

	var sSDpD data.SeeSeekDetailTemplateData
	sSDpD.SessUser = s_u
	sSDpD.IsVerifier = is_verifier

	sSDpD.Verifier = s_u
	sSDpD.VerifierFamily = proj_appointment_bean.VerifierFamily
	sSDpD.VerifierTeam = proj_appointment_bean.VerifierTeam
	sSDpD.Payer = proj_appointment_bean.Payer
	sSDpD.PayerFamily = proj_appointment_bean.PayerFamily
	sSDpD.PayerTeam = proj_appointment_bean.PayerTeam
	sSDpD.Payee = proj_appointment_bean.Payee
	sSDpD.PayeeFamily = proj_appointment_bean.PayeeFamily
	sSDpD.PayeeTeam = proj_appointment_bean.PayeeTeam

	sSDpD.ProjectBean = projBean
	sSDpD.QuoteObjectiveBean = objeBean
	sSDpD.ProjectAppointment = proj_appointment_bean
	sSDpD.Environments = environments

	renderHTML(w, &sSDpD, "layout", "navbar.private", "action.see-seek.new", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// Handler /v1/see-seek/detail
func HandleSeeSeekDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	SeeSeekDetailGet(w, r)
}

// GET /v1/see-seek/detail?uuid=xxx
func SeeSeekDetailGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	user, err := sess.User()
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

	// 获取SeeSeek记录
	seeSeek := data.SeeSeek{Uuid: uuid}
	if err := seeSeek.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			//尝试project的uuid
			project := data.Project{Uuid: uuid}
			if err := project.GetByUuid(); err != nil {
				util.Debug("Cannot get project by uuid", uuid, err)
				report(w, r, "你好，假作真时真亦假，无为有处有还无？")
				return
			}
			seeSeek, err = data.GetSeeSeekByProjectId(project.Id, r.Context())
			if err != nil {
				if err == sql.ErrNoRows {
					report(w, r, "该项目还没有“看看”记录")
					return
				}
				util.Debug("Cannot get SeeSeek by project_id", project.Id, err)
				report(w, r, "该项目的“看看”记录似乎被水泡糊了")
				return
			}
		} else {
			util.Debug("Cannot get SeeSeek by uuid", uuid, err)
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
	}

	// 获取项目信息
	project := data.Project{Id: seeSeek.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project", err)
		report(w, r, "获取项目信息失败")
		return
	}

	// 获取目标信息
	objective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective", err)
		report(w, r, "获取目标信息失败")
		return
	}

	// 获取完整的SeeSeekBean
	seeSeekBean, err := fetchSeeSeekBean(seeSeek)
	if err != nil {
		util.Debug("Cannot fetch SeeSeek bean", err)
		report(w, r, "获取看看记录详情失败")
		return
	}

	projectBean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, r, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, r, "获取目标详情失败")
		return
	}

	// 准备页面数据
	templateData := data.SeeSeekDetailTemplateData{
		SessUser:           user,
		IsVerifier:         isVerifier(user.Id),
		SeeSeekBean:        seeSeekBean,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
	}

	renderHTML(w, &templateData, "layout", "navbar.private", "action.see-seek.detail", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}
