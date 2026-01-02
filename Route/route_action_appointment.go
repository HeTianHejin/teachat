package route

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	dao "teachat/DAO"
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
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
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
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 根据uuid获取项目
	pr := dao.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		// 如果项目获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get project", uuid, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目的bean
	pr_bean, err := fetchProjectBean(pr)
	if err != nil {
		// 如果项目bean获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get project bean", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	master, err := dao.GetUser(pr.UserId)
	if err != nil {
		// 如果用户获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get user", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目所属的家族
	master_family, err := dao.GetFamily(pr.FamilyId)
	if err != nil {
		// 如果家族获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get family", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目所属的团队
	master_team, err := dao.GetTeam(pr.TeamId)
	if err != nil {
		// 如果团队获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get team", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取项目的目标
	ob, err := pr.Objective()
	if err != nil {
		// 如果目标获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get objective", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取目标的bean
	ob_bean, err := fetchObjectiveBean(ob)
	if err != nil {
		// 如果目标bean获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get objective bean", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	admin, err := dao.GetUser(ob.UserId)
	if err != nil {
		// 如果用户获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get user", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取目标所属的家族
	admin_family, err := dao.GetFamily(ob.FamilyId)
	if err != nil {
		// 如果家族获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get family", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取目标所属的团队
	admin_team, err := dao.GetTeam(ob.TeamId)
	if err != nil {
		// 如果团队获取失败，则记录错误并返回错误信息
		util.Debug(" Cannot get team", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取验证者团队
	verifier_team, err := dao.GetTeam(dao.TeamIdVerifier)
	if err != nil {
		util.Debug(" Cannot get verifier team", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 创建项目预约的bean
	p_a := dao.ProjectAppointmentBean{
		Appointment:    dao.ProjectAppointment{},
		Project:        pr,
		Payer:          master,
		PayerFamily:    master_family,
		PayerTeam:      master_team,
		Payee:          admin,
		PayeeFamily:    admin_family,
		PayeeTeam:      admin_team,
		Verifier:       s_u,
		VerifierFamily: dao.FamilyUnknown,
		VerifierTeam:   verifier_team,
	}
	// 创建预约页面数据
	pAD := dao.AppointmentTemplateData{
		SessUser:           s_u,
		IsVerifier:         true,
		ProjectBean:        pr_bean,
		QuoteObjectiveBean: ob_bean,
		AppointmentBean:    p_a,
	}
	// 渲染HTML页面
	generateHTML(w, &pAD, "layout", "navbar.private", "action.appointment.new", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
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
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
	}
	// 如果当前用户不是验证者，则返回错误信息
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// 获取表单数据
	if err := r.ParseForm(); err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	place_id_string := r.PostFormValue("place_id")
	if place_id_string == "" {
		util.Debug(" Cannot get place_id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	place_id, err := strconv.Atoi(place_id_string)
	if err != nil {
		util.Debug(" Cannot convert place_id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	start_time_string := r.PostFormValue("start_time")
	start_time, err := time.Parse("2006-01-02T15:04", start_time_string)
	if err != nil {
		util.Debug(" Cannot parse start_time", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	end_time_string := r.PostFormValue("end_time")
	end_time, err := time.Parse("2006-01-02T15:04", end_time_string)
	if err != nil {
		util.Debug(" Cannot parse end_time", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payer_user_id_string := r.PostFormValue("payer_user_id")
	payer_user_id, err := strconv.Atoi(payer_user_id_string)
	if err != nil {
		util.Debug(" Cannot convert payer_user_id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payer_team_id_string := r.PostFormValue("payer_team_id")
	payer_team_id, err := strconv.Atoi(payer_team_id_string)
	if err != nil {
		util.Debug(" Cannot convert payer_team_id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payer_family_id_string := r.PostFormValue("payer_family_id")
	payer_family_id, err := strconv.Atoi(payer_family_id_string)
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payee_user_id_string := r.PostFormValue("payee_user_id")
	payee_user_id, err := strconv.Atoi(payee_user_id_string)
	if err != nil {
		util.Debug(" Cannot convert payee_user_id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payee_team_id_string := r.PostFormValue("payee_team_id")
	payee_team_id, err := strconv.Atoi(payee_team_id_string)
	if err != nil {
		util.Debug(" Cannot convert payee_team_id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	payee_family_id_string := r.PostFormValue("payee_family_id")
	payee_family_id, err := strconv.Atoi(payee_family_id_string)
	if err != nil {
		util.Debug(" Cannot convert payee_family_id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
	}
	note := r.PostFormValue("note")
	project_id_string := r.PostFormValue("project_id")
	project_id, err := strconv.Atoi(project_id_string)
	if err != nil {
		util.Debug(" Cannot convert project_id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	// // 获取项目
	// pr := dao.Project{Id: project_id}
	// if err = pr.Get(); err != nil {
	// 	util.Debug(" Cannot get project", err)
	// 	report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
	// 	return
	// }
	// if place_id != pr.PlaceId{
	// 	report(w, s_u, "请选择正确的地点")
	// 	return
	// }

	verifier_family_id := dao.FamilyIdUnknown
	verifier_team_id := dao.TeamIdVerifier
	now := time.Now()

	// 创建新的预约记录
	new_p_a := dao.ProjectAppointment{
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
		Status:           dao.AppointmentStatusSubmitted,
		ConfirmedAt:      &now,
		UpdatedAt:        now,
	}
	// 保存预约记录
	err = new_p_a.Create(r.Context())
	if err != nil {
		util.Debug(" Cannot save project appointment", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 跳转约茶详情页面
	http.Redirect(w, r, "/v1/appointment/detail?uuid="+new_p_a.Uuid, http.StatusSeeOther)

}

// GET /v1/appointment/detail?uuid=xxx
func AppointmentDetail(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士看不懂陕下提交的UUID参数，请稍后再试。")
		return
	}

	var pr dao.Project
	// 尝试直接获取预约记录
	pr_appointment := dao.ProjectAppointment{Uuid: uuid}
	if err = pr_appointment.GetByIdOrUUID(r.Context()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 如果找不到预约记录，尝试用project uuid查找
			pr = dao.Project{Uuid: uuid}
			if err = pr.GetByUuid(); err != nil {
				util.Debug(" Cannot get project by uuid", uuid, err)
				report(w, s_u, "你好，茶博士找不到指定的茶台或预约记录。")
				return
			}
			// 用project id查找预约记录
			pr_appointment, err = dao.GetAppointmentByProjectId(pr.Id, r.Context())
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					report(w, s_u, "这个茶台尚未约茶。")
					return
				}
				util.Debug(" Cannot get appointment by project id", pr.Id, err)
				report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
				return
			}
		} else {
			util.Debug(" Cannot get appointment by uuid", uuid, err)
			report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
			return
		}
	}

	if !(pr.Id > 1) {
		// 获取项目信息
		pr = dao.Project{Id: pr_appointment.ProjectId}
		if err = pr.Get(); err != nil {
			util.Debug(" Cannot get project", pr_appointment.ProjectId, err)
			report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
			return
		}
	}

	pr_bean, err := fetchProjectBean(pr)
	if err != nil {
		util.Debug(" Cannot get project bean", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	p_a_bean, err := fetchAppointmentBean(pr_appointment)
	if err != nil {
		util.Debug(" Cannot fetch appointment bean", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 获取目标
	ob, err := pr.Objective()
	if err != nil {
		util.Debug(" Cannot get objective", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	ob_bean, err := fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot get objective bean", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	aPD := dao.AppointmentTemplateData{
		SessUser:           s_u,
		IsVerifier:         dao.IsVerifier(s_u.Id),
		AppointmentBean:    p_a_bean,
		ProjectBean:        pr_bean,
		QuoteObjectiveBean: ob_bean,
	}
	generateHTML(w, &aPD, "layout", "navbar.private", "action.appointment.detail", "component_sess_capacity", "component_avatar_name_gender")
}

// GET /v1/appointment/accept?uuid=xxx
// 确认约茶功能 - 接受预约
func AppointmentAccept(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士看不懂陛下提交的UUID参数，请稍后再试。")
		return
	}

	// 获取预约记录
	pr_appointment := dao.ProjectAppointment{Uuid: uuid}
	if err = pr_appointment.GetByIdOrUUID(r.Context()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, s_u, "你好，茶博士找不到指定的预约记录。")
			return
		}
		util.Debug(" Cannot get appointment", uuid, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查预约状态
	if pr_appointment.Status != dao.AppointmentStatusSubmitted {
		report(w, s_u, "该预约已处理，无需重复操作。")
		return
	}

	// 检查权限 - 只有付费方或收费方可以确认
	// if s_u.Id != pr_appointment.PayerUserId && s_u.Id != pr_appointment.PayeeUserId {
	// 	report(w, s_u, "你好，茶博士说只有相关当事人才能确认约茶。")
	// 	return
	// }
	// 检查权限 - 只有见证人可以确认
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，茶博士说只有见证人才能确认约茶。")
		return
	}

	// 更新预约状态为已确认
	now := time.Now()
	pr_appointment.Status = dao.AppointmentStatusConfirmed
	pr_appointment.ConfirmedAt = &now
	pr_appointment.UpdatedAt = now

	if err = pr_appointment.Update(r.Context()); err != nil {
		util.Debug(" Cannot update appointment status", err)
		report(w, s_u, "你好，茶博士墨水不够，未能确认约茶。")
		return
	}
	// 更新项目状态为已约茶
	pr := dao.Project{Id: pr_appointment.ProjectId}
	if err = pr.Get(); err != nil {
		util.Debug(" Cannot get project", pr_appointment.ProjectId, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	} else {
		pr.Status = dao.ProjectStatusHotTea

		if err = pr.Update(); err != nil {
			util.Debug(" Cannot update project status", err)
			report(w, s_u, "你好，茶博士墨水不够，未能确认约茶。")
			return
		}
	}

	// 重定向到预约详情页面
	http.Redirect(w, r, "/v1/appointment/detail?uuid="+uuid, http.StatusFound)
}

// GET /v1/appointment/reject?uuid=xxx
// 拒绝约茶功能 - 拒绝预约
func AppointmentReject(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士看不懂陛下提交的UUID参数，请稍后再试。")
		return
	}

	// 获取预约记录
	pr_appointment := dao.ProjectAppointment{Uuid: uuid}
	if err = pr_appointment.GetByIdOrUUID(r.Context()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, s_u, "你好，茶博士找不到指定的预约记录。")
			return
		}
		util.Debug(" Cannot get appointment", uuid, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查预约状态
	if pr_appointment.Status != dao.AppointmentStatusPending {
		report(w, s_u, "该预约已处理，无需重复操作。")
		return
	}

	// 检查权限 - 只有见证人可以拒绝
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，茶博士说只有见证人才能拒绝约茶。")
		return
	}

	// 更新预约状态为已拒绝
	now := time.Now()
	pr_appointment.Status = dao.AppointmentStatusRejected
	pr_appointment.UpdatedAt = now

	if err = pr_appointment.Update(r.Context()); err != nil {
		util.Debug(" Cannot update appointment status", err)
		report(w, s_u, "你好，茶博士墨水不够，未能拒绝约茶。")
		return
	}

	// 重定向到预约详情页面
	http.Redirect(w, r, "/v1/appointment/detail?uuid="+uuid, http.StatusFound)
}
