package route

import (
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/see-seek/step
func HandleSeeSeekStep(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SeeSeekStepGet(w, r)
	case http.MethodPost:
		SeeSeekStepPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 继续执行“看看”记录第n(1<n<6)步
// GET /v1/see-seek/step?uuid=xxx
func SeeSeekStepGet(w http.ResponseWriter, r *http.Request) {
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

	// 检测当前会话茶友是否见证者
	is_verifier := isVerifier(s_u.Id)
	if !is_verifier {
		util.Debug(" Current user is not a verifier", s_u.Id)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	vals := r.URL.Query()
	uuid := vals.Get("uuid")

	if uuid == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	// 检查SeeSeek记录
	see_seek := data.SeeSeek{Uuid: uuid}
	if err = see_seek.GetByIdOrUUID(r.Context()); err != nil {
		if err.Error() == "没有记录" {
			// 项目的“看看”第一步还没有做！提醒
			report(w, r, "你好，项目的“看看”第一步还没有记录，请检查项目详情？")
			return
		}
		//这是发生了数据库操作错误
		util.Debug("Cannot get SeeSeek by uuid", err)
		report(w, r, "处理“看看”记录时发生错误，请稍后再试")
		return
	}

	if see_seek.Id < 1 || see_seek.Status < data.SeeSeekStatusInProgress {
		report(w, r, "无效的“看看”记录。")
		return
	}
	// 获取项目信息
	project := data.Project{Id: see_seek.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	// 获取目标信息
	quoteObjective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective by project id", project.Id, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	project_bean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, r, "获取项目详情失败，请稍后再试")
		return
	}
	objective_bean, err := fetchObjectiveBean(quoteObjective)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, r, "获取目标详情失败，请稍后再试")
		return
	}

	see_seek_bean, err := fetchSeeSeekBean(see_seek)
	if err != nil {
		util.Debug("Cannot fetch SeeSeek bean", err)
		report(w, r, "获取“看看”记录详情失败，请稍后再试")
		return
	}
	completedSteps := see_seek.Step
	currentStep := completedSteps + 1
	seeSeekStepTitle := data.GetSeeSeekStepTitle(currentStep)
	// 准备页面数据
	sssTD := data.SeeSeekStepTemplateData{
		SessUser:           s_u,
		IsVerifier:         is_verifier,
		Verifier:           s_u,
		VerifierFamily:     data.FamilyUnknown,
		VerifierTeam:       data.TeamVerifier,
		ProjectBean:        project_bean,
		QuoteObjectiveBean: objective_bean,
		SeeSeekBean:        see_seek_bean,
		SeeSeekStepTitle:   seeSeekStepTitle,
		CompletedSteps:     completedSteps,
		CurrentStep:        currentStep,
	}

	renderHTML(w, &sssTD, "layout", "navbar.private", "action.see-seek.step", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/see-seek/step
func SeeSeekStepPost(w http.ResponseWriter, r *http.Request) {
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

	if !isVerifier(user.Id) {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	seeSeekUuid := r.PostFormValue("see_seek_uuid")
	stepStr := r.PostFormValue("step")

	step, _ := strconv.Atoi(stepStr)
	if step < 1 || step > 5 {
		report(w, r, "无效的步骤参数。")
		return
	}

	// 获取SeeSeek记录
	seeSeek := data.SeeSeek{Uuid: seeSeekUuid}
	if err := seeSeek.GetByIdOrUUID(r.Context()); err != nil {
		report(w, r, "看看记录不存在。")
		return
	}

	// 获取项目信息
	project := data.Project{Id: seeSeek.ProjectId}
	if err := project.Get(); err != nil {
		report(w, r, "项目不存在。")
		return
	}

	// 跳转到下一步或完成
	nextStep := step + 1
	if nextStep > 5 {
		// 完成所有步骤，跳转到项目详情页
		http.Redirect(w, r, "/v1/project/detail?uuid="+project.Uuid, http.StatusFound)
	} else {
		// 跳转到下一步
		http.Redirect(w, r, "/v1/see-seek/step?uuid="+seeSeekUuid, http.StatusFound)
	}
}
