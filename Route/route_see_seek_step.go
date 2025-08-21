package route

import (
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/see-seek/step2
// （处理发现+记录场所隐患）
func HandleSeeSeekStep2(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SeeSeekStep2Get(w, r)
	case http.MethodPost:
		SeeSeekStep2Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 继续/v1/see-seek/new后，执行“看看”记录第2步（处理发现记录场所隐患）
// GET /v1/see-seek/step2?uuid=xxx
func SeeSeekStep2Get(w http.ResponseWriter, r *http.Request) {
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

	// 获取默认隐患列表（仅在第2步时需要）
	var defaultHazards []data.Hazard
	if currentStep == 2 {
		if hazards, err := data.GetDefaultHazards(r.Context()); err == nil {
			defaultHazards = hazards
		}
	}

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
		DefaultHazards:     defaultHazards,
	}

	renderHTML(w, &sssTD, "layout", "navbar.private", "action.see-seek.step2", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/see-seek/step2
func SeeSeekStep2Post(w http.ResponseWriter, r *http.Request) {
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
	if step != data.SeeSeekStepHazard {
		report(w, r, "无效的步骤参数。")
		return
	}

	// 获取SeeSeek记录
	see_seek := data.SeeSeek{Uuid: seeSeekUuid}
	if err := see_seek.GetByIdOrUUID(r.Context()); err != nil {
		if err.Error() == "没有记录" {
			report(w, r, "看看记录不存在。")
			return
		}
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	if see_seek.Id < 1 || see_seek.Status < data.SeeSeekStatusInProgress {
		report(w, r, "无效的看看记录。")
		return
	}
	// 检查是否重复提交
	if see_seek.Step >= data.SeeSeekStepHazard {
		report(w, r, "该步骤已经完成，请勿重复提交。")
		return
	}

	// 获取项目信息
	project := data.Project{Id: see_seek.ProjectId}
	if err := project.Get(); err != nil {
		report(w, r, "项目不存在。")
		return
	}

	//保存提交的场所隐患数据hazard_ids
	hazardIdsStr := r.PostFormValue("hazard_ids")
	if hazardIdsStr != "" {
		//用正则表达式检测hazard_ids，是否符合“整数，整数，整数...”的格式
		if !verifyIdSliceFormat(hazardIdsStr) {
			util.Debug(" hazard_ids slice format is wrong", err)
			report(w, r, "你好，填写的场所安全隐患号格式看不懂，请确认后再试。")
			return
		}
	}
	var hazardIds []int
	hazardIdsStrSlice := strings.Split(hazardIdsStr, ",")
	for _, idStr := range hazardIdsStrSlice {
		if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
			hazardIds = append(hazardIds, id)
		}
	}
	// 如果hazardIds为空，表示没有选择隐患
	//var see_seek_hazards []data.SeeSeekHazard
	// 允许为空，可能是没有发现隐患
	if len(hazardIds) > 0 {
		// 保存隐患记录
		for _, hazardId := range hazardIds {
			see_seek_hazard := data.SeeSeekHazard{
				SeeSeekId: see_seek.Id,
				HazardId:  hazardId,
			}
			if err := see_seek_hazard.Create(); err != nil {
				util.Debug("Cannot create SeeSeekHazard", err)
				report(w, r, "保存隐患记录时发生错误，请稍后再试")
				return
			}
			//see_seek_hazards = append(see_seek_hazards, see_seek_hazard)
		}
	}

	// 保存步骤，即更新seeSeek
	see_seek.Step = data.SeeSeekStepHazard
	if err := see_seek.Update(); err != nil {
		util.Debug("Cannot update SeeSeek step", see_seek.Uuid, err)
		report(w, r, "保存看看记录步骤时发生错误，请稍后再试")
		return
	}

	nextStep := data.SeeSeekStepHazard + 1

	// 跳转到下一步
	http.Redirect(w, r, "/v1/see-seek/step"+strconv.Itoa(nextStep)+"?uuid="+seeSeekUuid, http.StatusFound)

}

// Handler /v1/see-seek/step3
// （处理风险评估）
func HandleSeeSeekStep3(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SeeSeekStep3Get(w, r)
	case http.MethodPost:
		SeeSeekStep3Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/see-seek/step3?uuid=xxx
func SeeSeekStep3Get(w http.ResponseWriter, r *http.Request) {
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

	see_seek := data.SeeSeek{Uuid: uuid}
	if err = see_seek.GetByIdOrUUID(r.Context()); err != nil {
		if err.Error() == "没有记录" {
			report(w, r, "你好，项目的看看记录不存在，请检查项目详情？")
			return
		}
		util.Debug("Cannot get SeeSeek by uuid", err)
		report(w, r, "处理看看记录时发生错误，请稍后再试")
		return
	}

	if see_seek.Id < 1 || see_seek.Status < data.SeeSeekStatusInProgress {
		report(w, r, "无效的看看记录。")
		return
	}

	project := data.Project{Id: see_seek.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

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
		report(w, r, "获取看看记录详情失败，请稍后再试")
		return
	}

	completedSteps := see_seek.Step
	currentStep := completedSteps + 1
	// 获取当前步骤标题
	seeSeekStepTitle := data.GetSeeSeekStepTitle(currentStep)

	// 获取默认风险列表（仅在第3步时需要）
	var defaultRisks []data.Risk
	if currentStep == 3 {
		if risks, err := data.GetDefaultRisks(r.Context()); err == nil {
			defaultRisks = risks
		}
	}

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
		// 状态管理相关字段
		CompletedSteps: completedSteps,
		CurrentStep:    currentStep,
		DefaultRisks:   defaultRisks,
	}

	renderHTML(w, &sssTD, "layout", "navbar.private", "action.see-seek.step3", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/see-seek/step3
func SeeSeekStep3Post(w http.ResponseWriter, r *http.Request) {
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
	if step != data.SeeSeekStepRisk {
		report(w, r, "无效的步骤参数。")
		return
	}

	seeSeek := data.SeeSeek{Uuid: seeSeekUuid}
	if err := seeSeek.GetByIdOrUUID(r.Context()); err != nil {
		if err.Error() == "没有记录" {
			report(w, r, "看看记录不存在。")
			return
		}
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	if seeSeek.Step >= step {
		report(w, r, "该步骤已经完成，请勿重复提交。")
		return
	}

	project := data.Project{Id: seeSeek.ProjectId}
	if err := project.Get(); err != nil {
		report(w, r, "项目不存在。")
		return
	}

	riskIdsStr := r.PostFormValue("risk_ids")
	if riskIdsStr != "" {
		if !verifyIdSliceFormat(riskIdsStr) {
			util.Debug(" risk_ids slice format is wrong", err)
			report(w, r, "你好，填写的风险评估号格式看不懂，请确认后再试。")
			return
		}
	}

	var riskIds []int
	riskIdsStrSlice := strings.Split(riskIdsStr, ",")
	for _, idStr := range riskIdsStrSlice {
		if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
			riskIds = append(riskIds, id)
		}
	}

	if len(riskIds) > 0 {
		for _, riskId := range riskIds {
			see_seek_risk := data.SeeSeekRisk{
				SeeSeekId: seeSeek.Id,
				RiskId:    riskId,
			}
			if err := see_seek_risk.Create(); err != nil {
				util.Debug("Cannot create SeeSeekRisk", err)
				report(w, r, "保存风险记录时发生错误，请稍后再试")
				return
			}
		}
	}

	seeSeek.Step = data.SeeSeekStepRisk
	if err := seeSeek.Update(); err != nil {
		util.Debug("Cannot update SeeSeek step", seeSeek.Uuid, err)
		report(w, r, "保存看看记录步骤时发生错误，请稍后再试")
		return
	}

	nextStep := data.SeeSeekStepRisk + 1

	http.Redirect(w, r, "/v1/see-seek/step"+strconv.Itoa(nextStep)+"?uuid="+seeSeekUuid, http.StatusFound)
}
