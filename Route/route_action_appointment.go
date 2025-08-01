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
func NewAppointmentGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "uuid is required", http.StatusBadRequest)
		return
	}
	if !isVerifier(s_u.Id) {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	pr := data.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	pr_bean, err := fetchProjectBean(pr)
	if err != nil {
		util.Debug(" Cannot get project bean", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	master_family, err := data.GetFamily(pr.FamilyId)
	if err != nil {
		util.Debug(" Cannot get family", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	master_team, err := data.GetTeam(pr.TeamId)
	if err != nil {
		util.Debug(" Cannot get team", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
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
	admin_family, err := data.GetFamily(ob.FamilyId)
	if err != nil {
		util.Debug(" Cannot get family", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	admin_team, err := data.GetTeam(ob.TeamId)
	if err != nil {
		util.Debug(" Cannot get team", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	p_a := data.ProjectAppointmentBean{
		ProjectAppointment: data.ProjectAppointment{},
		Project:            pr,
		PayerFamily:        master_family,
		PayerTeam:          master_team,
		PayeeFamily:        admin_family,
		PayeeTeam:          admin_team,
		Verifier:           s_u,
		VerifierFamily:     data.FamilyUnknown,
		VerifierTeam:       data.TeamVerifier,
	}
	pAD := data.AppointmentPageData{
		SessUser:        s_u,
		IsVerifier:      true,
		ProjectBean:     pr_bean,
		ObjectiveBean:   ob_bean,
		AppointmentBean: p_a,
	}
	renderHTML(w, &pAD, "layout", "navbar.private", "appointment.new", "component_sess_capacity")
}

// POST /v1/appointment/new
func NewAppointmentPost(w http.ResponseWriter, r *http.Request) {
	p_a := data.ProjectAppointment{}
	fetchProjectAppointmentBean(p_a)

}

// Get /v1/appointment/detail?uuid=xXx
func AppointmentDetail(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
