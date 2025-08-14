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
	//读取预设的4个通用场所环境,id为1,2,3,4
	environments, err := data.GetDefaultEnvironments(r.Context())
	if err != nil {
		util.Debug(" Cannot get default environments", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 检查是否有新创建的环境条件
	newEnvIdStr := r.URL.Query().Get("new_env_id")
	if newEnvIdStr != "" {
		if newEnvId, err := strconv.Atoi(newEnvIdStr); err == nil {
			newEnv := data.Environment{Id: newEnvId}
			if err := newEnv.GetByIdOrUUID(); err == nil {
				// 将新环境添加到列表前面
				environments = append([]data.Environment{newEnv}, environments...)
			}
		}
	}

	var sSDpD data.SeeSeekDetailPageData
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

	renderHTML(w, &sSDpD, "layout", "navbar.private", "project.see-seek.new", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}
