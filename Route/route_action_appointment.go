package route

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
)

func HandleNewAppointment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		NewAppointmentGet(w, r)
	case http.MethodPost:
		NewAppointmentPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/appointment/new?uuid=xXx
// NewAppointmentGet 函数用于获取新的预约
func NewAppointmentGet(w http.ResponseWriter, r *http.Request) {
	// 获取当前会话
	sess, err := session(r)
	if err != nil {
		// 如果会话获取失败，则重定向到登录页面
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 获取当前用户
	s_u, err := sess.User()
	if err != nil {
		// 如果用户获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取URL中的uuid参数
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		// 如果uuid参数为空，则返回错误信息
		http.Error(w, "uuid is required", http.StatusBadRequest)
		return
	}
	// 如果当前用户不是验证者，则返回错误信息
	if !isVerifier(s_u.Id) {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 根据uuid获取项目
	pr := data.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		// 如果项目获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get project", uuid, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目的bean
	pr_bean, err := fetchProjectBean(pr)
	if err != nil {
		// 如果项目bean获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get project bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	master, err := data.GetUser(pr.UserId)
	if err != nil {
		// 如果用户获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get user", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目所属的家族
	master_family, err := data.GetFamily(pr.FamilyId)
	if err != nil {
		// 如果家族获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get family", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目所属的团队
	master_team, err := data.GetTeam(pr.TeamId)
	if err != nil {
		// 如果团队获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get team", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目的目标
	ob, err := pr.Objective()
	if err != nil {
		// 如果目标获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get objective", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取目标的bean
	ob_bean, err := fetchObjectiveBean(ob)
	if err != nil {
		// 如果目标bean获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get objective bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	admin, err := data.GetUser(ob.UserId)
	if err != nil {
		// 如果用户获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get user", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取目标所属的家族
	admin_family, err := data.GetFamily(ob.FamilyId)
	if err != nil {
		// 如果家族获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get family", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取目标所属的团队
	admin_team, err := data.GetTeam(ob.TeamId)
	if err != nil {
		// 如果团队获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get team", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 创建项目预约的bean
	p_a := data.ProjectAppointmentBean{
		Appointment:    data.ProjectAppointment{},
		Project:        pr,
		Payer:          master,
		PayerFamily:    master_family,
		PayerTeam:      master_team,
		Payee:          admin,
		PayeeFamily:    admin_family,
		PayeeTeam:      admin_team,
		Verifier:       s_u,
		VerifierFamily: data.FamilyUnknown,
		VerifierTeam:   data.TeamVerifier,
	}
	// 创建预约页面数据
	pAD := data.AppointmentPageData{
		SessUser:           s_u,
		IsVerifier:         true,
		ProjectBean:        pr_bean,
		QuoteObjectiveBean: ob_bean,
		AppointmentBean:    p_a,
	}
	// 渲染HTML页面
	renderHTML(w, &pAD, "layout", "navbar.private", "project.appointment.new", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/appointment/new
// NewAppointmentPost 函数用于接收用户提交的参数，以创建新的预约记录
func NewAppointmentPost(w http.ResponseWriter, r *http.Request) {
	// 获取当前会话
	sess, err := session(r)
	if err != nil {
		// 如果会话获取失败，则重定向到登录页面
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 获取当前用户
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
	}
	// 如果当前用户不是验证者，则返回错误信息
	if !isVerifier(s_u.Id) {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取表单数据
	if err := r.ParseForm(); err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	place_id_string := r.PostFormValue("place_id")
	if place_id_string == "" {
		util.Debug(" Cannot get place_id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	place_id, err := strconv.Atoi(place_id_string)
	if err != nil {
		util.Debug(" Cannot convert place_id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	start_time_string := r.PostFormValue("start_time")
	start_time, err := time.Parse("2006-01-02T15:04", start_time_string)
	if err != nil {
		util.Debug(" Cannot parse start_time", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	end_time_string := r.PostFormValue("end_time")
	end_time, err := time.Parse("2006-01-02T15:04", end_time_string)
	if err != nil {
		util.Debug(" Cannot parse end_time", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payer_user_id_string := r.PostFormValue("payer_user_id")
	payer_user_id, err := strconv.Atoi(payer_user_id_string)
	if err != nil {
		util.Debug(" Cannot convert payer_user_id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payer_team_id_string := r.PostFormValue("payer_team_id")
	payer_team_id, err := strconv.Atoi(payer_team_id_string)
	if err != nil {
		util.Debug(" Cannot convert payer_team_id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payer_family_id_string := r.PostFormValue("payer_family_id")
	payer_family_id, err := strconv.Atoi(payer_family_id_string)
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payee_user_id_string := r.PostFormValue("payee_user_id")
	payee_user_id, err := strconv.Atoi(payee_user_id_string)
	if err != nil {
		util.Debug(" Cannot convert payee_user_id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payee_team_id_string := r.PostFormValue("payee_team_id")
	payee_team_id, err := strconv.Atoi(payee_team_id_string)
	if err != nil {
		util.Debug(" Cannot convert payee_team_id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payee_family_id_string := r.PostFormValue("payee_family_id")
	payee_family_id, err := strconv.Atoi(payee_family_id_string)
	if err != nil {
		util.Debug(" Cannot convert payee_family_id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
	}
	note := r.PostFormValue("note")
	project_id_string := r.PostFormValue("project_id")
	project_id, err := strconv.Atoi(project_id_string)
	if err != nil {
		util.Debug(" Cannot convert project_id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目
	pr := data.Project{Id: project_id}
	if err = pr.Get(); err != nil {
		util.Debug(" Cannot get project", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	pr_bean, err := fetchProjectBean(pr)
	if err != nil {
		util.Debug(" Cannot get project bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	ob := data.Objective{Id: pr.ObjectiveId}
	if err = ob.Get(); err != nil {
		util.Debug(" Cannot get objective", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	ob_bean, err := fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot get objective bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	verifier_family_id := data.FamilyIdUnknown
	verifier_team_id := data.TeamIdVerifier
	now := time.Now()

	// 创建新的预约记录
	new_p_a := data.ProjectAppointment{
		PayerUserId:      payer_user_id,
		PayerTeamId:      payer_team_id,
		PayerFamilyId:    payer_family_id,
		PayeeUserId:      payee_user_id,
		PayeeTeamId:      payee_team_id,
		PayeeFamilyId:    payee_family_id,
		VerifierUserId:   s_u.Id,
		VerifierTeamId:   verifier_team_id,
		VerifierFamilyId: verifier_family_id,
		Note:             note,
		ProjectId:        project_id,
		StartTime:        start_time,
		EndTime:          end_time,
		PlaceId:          place_id,
		Status:           data.AppointmentStatusPending,
		ConfirmedAt:      &now,
		UpdatedAt:        now,
	}
	// 保存预约记录
	err = new_p_a.Create(r.Context())
	if err != nil {
		util.Debug(" Cannot save project appointment", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取预约记录的bean
	p_a_bean, err := fetchAppointmentBean(new_p_a)
	if err != nil {
		util.Debug(" Cannot fetch project appointment bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	aPD := data.AppointmentPageData{
		SessUser:           s_u,
		IsVerifier:         true,
		ProjectBean:        pr_bean,
		QuoteObjectiveBean: ob_bean,
		AppointmentBean:    p_a_bean,
	}

	renderHTML(w, &aPD, "layout", "navbar.private", "project.appointment.detail", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// Get /v1/appointment/detail?uuid=xXx
func AppointmentDetail(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取提交的uuid
	pr_uuid_string := r.URL.Query().Get("uuid")
	if pr_uuid_string == "" {
		util.Debug(" Cannot get uuid", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目
	pr := data.Project{Uuid: pr_uuid_string}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", pr_uuid_string, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	pr_bean, err := fetchProjectBean(pr)
	if err != nil {
		util.Debug(" Cannot get project bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 获取预约记录
	p_a, err := data.GetAppointmentByProjectId(pr.Id, r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "这个茶台尚未约茶。")
			return
		}
		if err.Error() == "没有找到相关的茶台预约" {
			report(w, r, "没有找到相关的茶台预约")
			return
		}
		util.Debug(" Cannot get project appointment", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	p_a_bean, err := fetchAppointmentBean(p_a)
	if err != nil {
		util.Debug(" Cannot fetch project appointment bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 获取目标
	ob, err := pr.Objective()
	if err != nil {
		util.Debug(" Cannot get objective", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	ob_bean, err := fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot get objective bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	aPD := data.AppointmentPageData{
		SessUser:           s_u,                // TODO: 检查权限-是否为预约记录的相关茶团或审核者
		IsVerifier:         isVerifier(s_u.Id), // TODO: 检查权限-是否为预约记录的审核者
		AppointmentBean:    p_a_bean,
		ProjectBean:        pr_bean,
		QuoteObjectiveBean: ob_bean,
	}
	renderHTML(w, &aPD, "layout", "navbar.private", "project.appointment.detail", "component_sess_capacity", "component_avatar_name_gender")

}
