package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
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
	renderHTML(w, &pAD, "layout", "navbar.private", "project.appointment.new", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/appointment/new
func NewAppointmentPost(w http.ResponseWriter, r *http.Request) {
	// 获取当前会话
	// sess, err := session(r)
	// if err != nil {
	// 	// 如果会话获取失败，则重定向到登录页面
	// 	http.Redirect(w, r, "/v1/login", http.StatusFound)
	// 	return
	// }
	// // 获取当前用户
	// s_u, err := sess.User()
	// if err != nil {
	// 	// 如果用户获取失败，则记录错误并返回错误信息
	// 	util.Debug(" Cannot get user from session", err)
	// 	report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
	// }
	p_a := data.ProjectAppointment{}
	fetchProjectAppointmentBean(p_a)

}

// Get /v1/appointment/detail?uuid=xXx
func AppointmentDetail(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
