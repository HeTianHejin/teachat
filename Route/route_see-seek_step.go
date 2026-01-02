package route

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	dao "teachat/DAO"
	util "teachat/Util"
	"time"
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
		util.Debug("Cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 检测当前会话茶友是否见证者
	is_verifier := dao.IsVerifier(s_u.Id)
	if !is_verifier {
		util.Debug(" Current s_u is not a verifier", s_u.Id)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	vals := r.URL.Query()
	uuid := vals.Get("uuid")

	if uuid == "" {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	// 检查SeeSeek记录
	see_seek := dao.SeeSeek{Uuid: uuid}
	if err = see_seek.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			// 项目的“看看”第一步还没有做！提醒
			report(w, s_u, "你好，项目的“看看”第一步还没有记录，请检查项目详情？")
			return
		}
		//这是发生了数据库操作错误
		util.Debug("Cannot get SeeSeek by uuid", err)
		report(w, s_u, "处理“看看”记录时发生错误，请稍后再试")
		return
	}

	if see_seek.Id < 1 || see_seek.Status < dao.SeeSeekStatusInProgress {
		report(w, s_u, "无效的“看看”记录。")
		return
	}
	// 获取项目信息
	project := dao.Project{Id: see_seek.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	// 获取目标信息
	quoteObjective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective by project id", project.Id, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	project_bean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, s_u, "获取项目详情失败，请稍后再试")
		return
	}
	objective_bean, err := fetchObjectiveBean(quoteObjective)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, s_u, "获取目标详情失败，请稍后再试")
		return
	}

	see_seek_bean, err := fetchSeeSeekBean(see_seek)
	if err != nil {
		util.Debug("Cannot fetch SeeSeek bean", err)
		report(w, s_u, "获取“看看”记录详情失败，请稍后再试")
		return
	}
	completedSteps := see_seek.Step
	currentStep := dao.SeeSeekStepHazard
	seeSeekStepTitle := dao.GetSeeSeekStepTitle(currentStep)

	// 获取默认隐患列表（仅在第2步时需要）
	var defaultHazards []dao.Hazard

	if hazards, err := dao.GetDefaultHazards(r.Context()); err == nil {
		defaultHazards = hazards
	}

	// 获取验证者团队
	verifier_team, err := dao.GetTeam(dao.TeamIdVerifier)
	if err != nil {
		util.Debug("Cannot get verifier team", err)
		report(w, s_u, "获取验证者团队失败，请稍后再试")
		return
	}
	// 准备页面数据
	sssTD := dao.SeeSeekStepTemplateData{
		SessUser:           s_u,
		IsVerifier:         is_verifier,
		Verifier:           s_u,
		VerifierFamily:     dao.FamilyUnknown,
		VerifierTeam:       verifier_team,
		ProjectBean:        project_bean,
		QuoteObjectiveBean: objective_bean,
		SeeSeekBean:        see_seek_bean,
		SeeSeekStepTitle:   seeSeekStepTitle,
		CompletedSteps:     completedSteps,
		CurrentStep:        currentStep,
		DefaultHazards:     defaultHazards,
	}

	generateHTML(w, &sssTD, "layout", "navbar.private", "action.see-seek.step2", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/see-seek/step2
func SeeSeekStep2Post(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	seeSeekUuid := r.PostFormValue("see_seek_uuid")
	stepStr := r.PostFormValue("step")

	step, _ := strconv.Atoi(stepStr)
	if step != dao.SeeSeekStepHazard {
		report(w, s_u, "无效的步骤参数。")
		return
	}

	// 获取SeeSeek记录
	see_seek := dao.SeeSeek{Uuid: seeSeekUuid}
	if err := see_seek.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			report(w, s_u, "看看记录不存在。")
			return
		}
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	if see_seek.Id < 1 || see_seek.Status < dao.SeeSeekStatusInProgress {
		report(w, s_u, "无效的看看记录。")
		return
	}
	// 允许从后续步骤返回修改，不检查重复提交

	// 获取项目信息
	project := dao.Project{Id: see_seek.ProjectId}
	if err := project.Get(); err != nil {
		report(w, s_u, "项目不存在。")
		return
	}

	//保存提交的场所隐患数据hazard_ids
	hazardIdsStr := r.PostFormValue("hazard_ids")
	var hazardIds []int
	if hazardIdsStr != "" {
		hazardIds, err = parseIdSlice(hazardIdsStr)
		if err != nil {
			report(w, s_u, "隐患ID格式不正确")
			return
		}
	}
	// 获取已有的隐患ID列表
	existingHazards, err := see_seek.GetHazards()
	if err != nil {
		util.Debug("Cannot get existing hazards", err)
		report(w, s_u, "获取已有隐患记录失败")
		return
	}

	// 读取旧数据id
	var existingIds []int
	for _, h := range existingHazards {
		existingIds = append(existingIds, h.HazardId)
	}

	// 检查新旧数据是否相同
	dataSame := compareIdsSlice(hazardIds, existingIds)

	// 只有数据变化时才更新数据库
	if !dataSame {
		// 删除已有的隐患关联记录
		err = dao.DeleteSeeSeekHazardsBySeeSeekId(see_seek.Id)
		if err != nil {
			util.Debug("Cannot delete existing SeeSeekHazards", err)
			report(w, s_u, "更新隐患记录时发生错误，请稍后再试")
			return
		}

		// 验证隐患ID是否存在
		for _, hazardId := range hazardIds {
			exist, err := dao.IsHazardIdExists(hazardId)
			if err != nil {
				util.Debug("Cannot check hazard ID existence", err)
				report(w, s_u, "验证隐患ID时发生错误，请稍后再试")
				return
			}
			if !exist {
				util.Debug("Invalid hazard ID", hazardId)
				report(w, s_u, fmt.Sprintf("隐患ID %d 不存在，请检查后再试", hazardId))
				return
			}

		}

		// 保存新的隐患记录
		if len(hazardIds) > 0 {
			for _, hazardId := range hazardIds {
				see_seek_hazard := dao.SeeSeekHazard{
					SeeSeekId: see_seek.Id,
					HazardId:  hazardId,
				}
				if err := see_seek_hazard.Create(); err != nil {
					util.Debug("Cannot create SeeSeekHazard", err)
					report(w, s_u, "保存隐患记录时发生错误，请稍后再试")
					return
				}
			}
		}
	}

	// 保存步骤，即更新seeSeek
	see_seek.Step = dao.SeeSeekStepHazard
	if err := see_seek.Update(); err != nil {
		util.Debug("Cannot update SeeSeek step", see_seek.Uuid, err)
		report(w, s_u, "保存看看记录步骤时发生错误，请稍后再试")
		return
	}

	nextStep := dao.SeeSeekStepHazard + 1

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
		util.Debug("Cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	is_verifier := dao.IsVerifier(s_u.Id)
	if !is_verifier {
		util.Debug(" Current s_u is not a verifier", s_u.Id)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	vals := r.URL.Query()
	uuid := vals.Get("uuid")

	if uuid == "" {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	see_seek := dao.SeeSeek{Uuid: uuid}
	if err = see_seek.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			report(w, s_u, "你好，项目的看看记录不存在，请检查项目详情？")
			return
		}
		util.Debug("Cannot get SeeSeek by uuid", err)
		report(w, s_u, "处理看看记录时发生错误，请稍后再试")
		return
	}

	if see_seek.Id < 1 || see_seek.Status < dao.SeeSeekStatusInProgress {
		report(w, s_u, "无效的看看记录。")
		return
	}

	project := dao.Project{Id: see_seek.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	quoteObjective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective by project id", project.Id, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	project_bean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, s_u, "获取项目详情失败，请稍后再试")
		return
	}

	objective_bean, err := fetchObjectiveBean(quoteObjective)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, s_u, "获取目标详情失败，请稍后再试")
		return
	}

	see_seek_bean, err := fetchSeeSeekBean(see_seek)
	if err != nil {
		util.Debug("Cannot fetch SeeSeek bean", err)
		report(w, s_u, "获取看看记录详情失败，请稍后再试")
		return
	}

	completedSteps := see_seek.Step
	currentStep := dao.SeeSeekStepRisk
	// 获取当前步骤标题
	seeSeekStepTitle := dao.GetSeeSeekStepTitle(currentStep)

	// 获取默认风险列表（仅在第3步时需要）
	var defaultRisks []dao.Risk

	if risks, err := dao.GetDefaultRisks(r.Context()); err == nil {
		defaultRisks = risks
	}

	verifier_team, err := dao.GetTeam(dao.TeamIdVerifier)
	if err != nil {
		util.Debug("Cannot get verifier team", err)
		report(w, s_u, "获取验证者团队失败，请稍后再试")
		return
	}

	sssTD := dao.SeeSeekStepTemplateData{
		SessUser:           s_u,
		IsVerifier:         is_verifier,
		Verifier:           s_u,
		VerifierFamily:     dao.FamilyUnknown,
		VerifierTeam:       verifier_team,
		ProjectBean:        project_bean,
		QuoteObjectiveBean: objective_bean,
		SeeSeekBean:        see_seek_bean,
		SeeSeekStepTitle:   seeSeekStepTitle,
		// 状态管理相关字段
		CompletedSteps: completedSteps,
		CurrentStep:    currentStep,
		DefaultRisks:   defaultRisks,
	}

	generateHTML(w, &sssTD, "layout", "navbar.private", "action.see-seek.step3", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/see-seek/step3
func SeeSeekStep3Post(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	seeSeekUuid := r.PostFormValue("see_seek_uuid")
	stepStr := r.PostFormValue("step")

	step, _ := strconv.Atoi(stepStr)
	if step != dao.SeeSeekStepRisk {
		report(w, s_u, "无效的步骤参数。")
		return
	}

	seeSeek := dao.SeeSeek{Uuid: seeSeekUuid}
	if err := seeSeek.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			report(w, s_u, "看看记录不存在。")
			return
		}
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 允许从后续步骤返回修改

	project := dao.Project{Id: seeSeek.ProjectId}
	if err := project.Get(); err != nil {
		report(w, s_u, "项目不存在。")
		return
	}

	riskIdsStr := r.PostFormValue("risk_ids")
	var riskIds []int
	if riskIdsStr != "" {
		riskIds, err = parseIdSlice(riskIdsStr)
		if err != nil {
			report(w, s_u, "风险ID格式不正确")
			return
		}
	}

	// 获取已有的风险ID列表
	existingRisks, err := seeSeek.GetRisks()
	if err != nil {
		util.Debug("Cannot get existing risks", err)
		report(w, s_u, "获取已有风险记录失败")
		return
	}

	// 比较新旧数据是否一致
	var existingIds []int
	for _, r := range existingRisks {
		existingIds = append(existingIds, r.RiskId)
	}

	// 检查数据是否变化
	dataSame := compareIdsSlice(riskIds, existingIds)

	// 只有数据变化时才更新数据库
	if !dataSame {
		// 删除已有的风险关联记录
		err = dao.DeleteSeeSeekRisksBySeeSeekId(seeSeek.Id)
		if err != nil {
			util.Debug("Cannot delete existing SeeSeekRisks", err)
			report(w, s_u, "更新风险记录时发生错误，请稍后再试")
			return
		}

		if len(riskIds) > 0 {
			for _, riskId := range riskIds {
				see_seek_risk := dao.SeeSeekRisk{
					SeeSeekId: seeSeek.Id,
					RiskId:    riskId,
				}
				if err := see_seek_risk.Create(); err != nil {
					util.Debug("Cannot create SeeSeekRisk", err)
					report(w, s_u, "保存风险记录时发生错误，请稍后再试")
					return
				}
			}
		}
	}

	seeSeek.Step = dao.SeeSeekStepRisk
	if err := seeSeek.Update(); err != nil {
		util.Debug("Cannot update SeeSeek step", seeSeek.Uuid, err)
		report(w, s_u, "保存看看记录步骤时发生错误，请稍后再试")
		return
	}

	nextStep := dao.SeeSeekStepRisk + 1

	http.Redirect(w, r, "/v1/see-seek/step"+strconv.Itoa(nextStep)+"?uuid="+seeSeekUuid, http.StatusFound)
}

// Handler /v1/see-seek/step4
// （处理感官观察）
func HandleSeeSeekStep4(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SeeSeekStep4Get(w, r)
	case http.MethodPost:
		SeeSeekStep4Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/see-seek/step4?uuid=xxx
func SeeSeekStep4Get(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	is_verifier := dao.IsVerifier(s_u.Id)
	if !is_verifier {
		util.Debug(" Current s_u is not a verifier", s_u.Id)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	vals := r.URL.Query()
	uuid := vals.Get("uuid")

	if uuid == "" {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	see_seek := dao.SeeSeek{Uuid: uuid}
	if err = see_seek.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			report(w, s_u, "你好，项目的看看记录不存在，请检查项目详情？")
			return
		}
		util.Debug("Cannot get SeeSeek by uuid", err)
		report(w, s_u, "处理看看记录时发生错误，请稍后再试")
		return
	}

	if see_seek.Id < 1 || see_seek.Status < dao.SeeSeekStatusInProgress {
		report(w, s_u, "无效的看看记录。")
		return
	}

	project := dao.Project{Id: see_seek.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	quoteObjective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective by project id", project.Id, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	project_bean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, s_u, "获取项目详情失败，请稍后再试")
		return
	}

	objective_bean, err := fetchObjectiveBean(quoteObjective)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, s_u, "获取目标详情失败，请稍后再试")
		return
	}

	see_seek_bean, err := fetchSeeSeekBean(see_seek)
	if err != nil {
		util.Debug("Cannot fetch SeeSeek bean", err)
		report(w, s_u, "获取看看记录详情失败，请稍后再试")
		return
	}

	completedSteps := see_seek.Step
	currentStep := dao.SeeSeekStepObservation
	seeSeekStepTitle := dao.GetSeeSeekStepTitle(currentStep)

	verifier_team, err := dao.GetTeam(dao.TeamIdVerifier)
	if err != nil {
		util.Debug("Cannot get verifier team", err)
		report(w, s_u, "获取验证者团队失败，请稍后再试")
		return
	}

	sssTD := dao.SeeSeekStepTemplateData{
		SessUser:           s_u,
		IsVerifier:         is_verifier,
		Verifier:           s_u,
		VerifierFamily:     dao.FamilyUnknown,
		VerifierTeam:       verifier_team,
		ProjectBean:        project_bean,
		QuoteObjectiveBean: objective_bean,
		SeeSeekBean:        see_seek_bean,
		SeeSeekStepTitle:   seeSeekStepTitle,
		CompletedSteps:     completedSteps,
		CurrentStep:        currentStep,
	}

	generateHTML(w, &sssTD, "layout", "navbar.private", "action.see-seek.step4", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/see-seek/step4
func SeeSeekStep4Post(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	seeSeekUuid := r.PostFormValue("see_seek_uuid")
	stepStr := r.PostFormValue("step")

	step, _ := strconv.Atoi(stepStr)
	if step != dao.SeeSeekStepObservation {
		report(w, s_u, "无效的步骤参数。")
		return
	}

	seeSeek := dao.SeeSeek{Uuid: seeSeekUuid}
	if err := seeSeek.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			report(w, s_u, "看看记录不存在。")
			return
		}
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 允许从后续步骤返回修改

	// 获取已有的感官观察数据
	existingLooks, _ := seeSeek.GetLooks()
	existingListens, _ := seeSeek.GetListens()
	existingSmells, _ := seeSeek.GetSmells()
	existingTouches, _ := seeSeek.GetTouches()

	// 保存视觉观察数据
	lookOutline := strings.TrimSpace(r.PostFormValue("look_outline"))
	lookSkin := strings.TrimSpace(r.PostFormValue("look_skin"))
	lookColor := strings.TrimSpace(r.PostFormValue("look_color"))
	lookIsDeform := r.PostFormValue("look_is_deform") == "1"
	lookIsGraze := r.PostFormValue("look_is_graze") == "1"
	lookIsChange := r.PostFormValue("look_is_change") == "1"

	// 检查视觉数据是否变化
	lookChanged := true
	if len(existingLooks) > 0 {
		existing := existingLooks[0]
		lookChanged = existing.Outline != lookOutline || existing.Skin != lookSkin || existing.Color != lookColor || existing.IsDeform != lookIsDeform || existing.IsGraze != lookIsGraze || existing.IsChange != lookIsChange
	} else {
		lookChanged = lookOutline != "" || lookSkin != "" || lookColor != "" || lookIsDeform || lookIsGraze || lookIsChange
	}

	if lookChanged && (lookOutline != "" || lookSkin != "" || lookColor != "" || lookIsDeform || lookIsGraze || lookIsChange) {
		err = dao.DeleteSeeSeekLooksBySeeSeekId(seeSeek.Id)
		if err != nil {
			util.Debug("Cannot delete existing SeeSeekLooks", err)
		}
		seeSeekLook := dao.SeeSeekLook{
			SeeSeekId: seeSeek.Id,
			Classify:  0,
			Status:    1,
			Outline:   lookOutline,
			IsDeform:  lookIsDeform,
			Skin:      lookSkin,
			IsGraze:   lookIsGraze,
			Color:     lookColor,
			IsChange:  lookIsChange,
		}
		if err := seeSeekLook.Create(); err != nil {
			util.Debug("Cannot create SeeSeekLook", err)
			report(w, s_u, "保存视觉观察记录时发生错误，请稍后再试")
			return
		}
	}

	// 保存听觉观察数据
	listenSound := strings.TrimSpace(r.PostFormValue("listen_sound"))
	listenIsAbnormal := r.PostFormValue("listen_is_abnormal") == "1"

	listenChanged := true
	if len(existingListens) > 0 {
		existing := existingListens[0]
		listenChanged = existing.Sound != listenSound || existing.IsAbnormal != listenIsAbnormal
	} else {
		listenChanged = listenSound != "" || listenIsAbnormal
	}

	if listenChanged && (listenSound != "" || listenIsAbnormal) {
		err = dao.DeleteSeeSeekListensBySeeSeekId(seeSeek.Id)
		if err != nil {
			util.Debug("Cannot delete existing SeeSeekListens", err)
		}
		seeSeekListen := dao.SeeSeekListen{
			SeeSeekId:  seeSeek.Id,
			Classify:   0,
			Status:     1,
			Sound:      listenSound,
			IsAbnormal: listenIsAbnormal,
		}
		if err := seeSeekListen.Create(); err != nil {
			util.Debug("Cannot create SeeSeekListen", err)
			report(w, s_u, "保存听觉观察记录时发生错误，请稍后再试")
			return
		}
	}

	// 保存嗅觉观察数据
	smellOdour := strings.TrimSpace(r.PostFormValue("smell_odour"))
	smellIsFoulOdour := r.PostFormValue("smell_is_foul_odour") == "1"

	smellChanged := true
	if len(existingSmells) > 0 {
		existing := existingSmells[0]
		smellChanged = existing.Odour != smellOdour || existing.IsFoulOdour != smellIsFoulOdour
	} else {
		smellChanged = smellOdour != "" || smellIsFoulOdour
	}

	if smellChanged && (smellOdour != "" || smellIsFoulOdour) {
		err = dao.DeleteSeeSeekSmellsBySeeSeekId(seeSeek.Id)
		if err != nil {
			util.Debug("Cannot delete existing SeeSeekSmells", err)
		}
		seeSeekSmell := dao.SeeSeekSmell{
			SeeSeekId:   seeSeek.Id,
			Classify:    0,
			Status:      1,
			Odour:       smellOdour,
			IsFoulOdour: smellIsFoulOdour,
		}
		if err := seeSeekSmell.Create(); err != nil {
			util.Debug("Cannot create SeeSeekSmell", err)
			report(w, s_u, "保存嗅觉观察记录时发生错误，请稍后再试")
			return
		}
	}

	// 保存触觉观察数据
	touchTemperature := strings.TrimSpace(r.PostFormValue("touch_temperature"))
	touchStretch := strings.TrimSpace(r.PostFormValue("touch_stretch"))
	touchShake := strings.TrimSpace(r.PostFormValue("touch_shake"))
	touchIsFever := r.PostFormValue("touch_is_fever") == "1"
	touchIsStiff := r.PostFormValue("touch_is_stiff") == "1"
	touchIsShake := r.PostFormValue("touch_is_shake") == "1"

	touchChanged := true
	if len(existingTouches) > 0 {
		existing := existingTouches[0]
		touchChanged = existing.Temperature != touchTemperature || existing.Stretch != touchStretch || existing.Shake != touchShake || existing.IsFever != touchIsFever || existing.IsStiff != touchIsStiff || existing.IsShake != touchIsShake
	} else {
		touchChanged = touchTemperature != "" || touchStretch != "" || touchShake != "" || touchIsFever || touchIsStiff || touchIsShake
	}

	if touchChanged && (touchTemperature != "" || touchStretch != "" || touchShake != "" || touchIsFever || touchIsStiff || touchIsShake) {
		err = dao.DeleteSeeSeekTouchesBySeeSeekId(seeSeek.Id)
		if err != nil {
			util.Debug("Cannot delete existing SeeSeekTouches", err)
		}
		seeSeekTouch := dao.SeeSeekTouch{
			SeeSeekId:   seeSeek.Id,
			Classify:    0,
			Status:      1,
			Temperature: touchTemperature,
			IsFever:     touchIsFever,
			Stretch:     touchStretch,
			IsStiff:     touchIsStiff,
			Shake:       touchShake,
			IsShake:     touchIsShake,
		}
		if err := seeSeekTouch.Create(); err != nil {
			util.Debug("Cannot create SeeSeekTouch", err)
			report(w, s_u, "保存触觉观察记录时发生错误，请稍后再试")
			return
		}
	}

	// 更新步骤
	seeSeek.Step = dao.SeeSeekStepObservation
	if err := seeSeek.Update(); err != nil {
		util.Debug("Cannot update SeeSeek step", seeSeek.Uuid, err)
		report(w, s_u, "保存看看记录步骤时发生错误，请稍后再试")
		return
	}

	nextStep := dao.SeeSeekStepObservation + 1

	http.Redirect(w, r, "/v1/see-seek/step"+strconv.Itoa(nextStep)+"?uuid="+seeSeekUuid, http.StatusFound)
}

// Handler /v1/see-seek/step5
func HandleSeeSeekStep5(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SeeSeekStep5Get(w, r)
	case http.MethodPost:
		SeeSeekStep5Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/see-seek/step5?uuid=xxx
func SeeSeekStep5Get(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	is_verifier := dao.IsVerifier(s_u.Id)
	if !is_verifier {
		util.Debug(" Current s_u is not a verifier", s_u.Id)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	vals := r.URL.Query()
	uuid := vals.Get("uuid")

	if uuid == "" {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	see_seek := dao.SeeSeek{Uuid: uuid}
	if err = see_seek.GetByIdOrUUID(r.Context()); err != nil {
		if err.Error() == "没有记录" {
			report(w, s_u, "你好，项目的看看记录不存在，请检查项目详情？")
			return
		}
		util.Debug("Cannot get SeeSeek by uuid", err)
		report(w, s_u, "处理看看记录时发生错误，请稍后再试")
		return
	}

	if see_seek.Id < 1 || see_seek.Status < dao.SeeSeekStatusInProgress {
		report(w, s_u, "无效的看看记录。")
		return
	}

	project := dao.Project{Id: see_seek.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	quoteObjective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective by project id", project.Id, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	project_bean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, s_u, "获取项目详情失败，请稍后再试")
		return
	}

	objective_bean, err := fetchObjectiveBean(quoteObjective)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, s_u, "获取目标详情失败，请稍后再试")
		return
	}

	see_seek_bean, err := fetchSeeSeekBean(see_seek)
	if err != nil {
		util.Debug("Cannot fetch SeeSeek bean", err)
		report(w, s_u, "获取看看记录详情失败，请稍后再试")
		return
	}

	completedSteps := see_seek.Step
	currentStep := dao.SeeSeekStepReport
	seeSeekStepTitle := dao.GetSeeSeekStepTitle(currentStep)

	verifier_team, err := dao.GetTeam(dao.TeamIdVerifier)
	if err != nil {
		util.Debug("Cannot get verifier team", err)
		report(w, s_u, "获取验证者团队失败，请稍后再试")
		return
	}

	sssTD := dao.SeeSeekStepTemplateData{
		SessUser:           s_u,
		IsVerifier:         is_verifier,
		Verifier:           s_u,
		VerifierFamily:     dao.FamilyUnknown,
		VerifierTeam:       verifier_team,
		ProjectBean:        project_bean,
		QuoteObjectiveBean: objective_bean,
		SeeSeekBean:        see_seek_bean,
		SeeSeekStepTitle:   seeSeekStepTitle,
		CompletedSteps:     completedSteps,
		CurrentStep:        currentStep,
	}

	generateHTML(w, &sssTD, "layout", "navbar.private", "action.see-seek.step5", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/see-seek/step5
func SeeSeekStep5Post(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	seeSeekUuid := r.PostFormValue("see_seek_uuid")
	stepStr := r.PostFormValue("step")

	step, _ := strconv.Atoi(stepStr)
	if step != dao.SeeSeekStepReport {
		report(w, s_u, "无效的步骤参数。")
		return
	}

	seeSeek := dao.SeeSeek{Uuid: seeSeekUuid}
	if err := seeSeek.GetByIdOrUUID(r.Context()); err != nil {
		if err.Error() == "没有记录" {
			report(w, s_u, "看看记录不存在。")
			return
		}
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 允许从后续步骤返回修改

	project := dao.Project{Id: seeSeek.ProjectId}
	if err := project.Get(); err != nil {
		report(w, s_u, "项目不存在。")
		return
	}

	// 获取完成时间
	endTimeStr := r.PostFormValue("end_time")
	if endTimeStr == "" {
		report(w, s_u, "请选择完成时间")
		return
	}

	// 解析完成时间
	endTime, err := time.Parse("2006-01-02T15:04", endTimeStr)
	if err != nil {
		util.Debug(" Cannot parse end time", endTimeStr, err)
		report(w, s_u, "完成时间格式不正确")
		return
	}

	// 保存检测报告数据（可选）
	reportTitle := strings.TrimSpace(r.PostFormValue("report_title"))
	reportContent := strings.TrimSpace(r.PostFormValue("report_content"))
	sampleType := strings.TrimSpace(r.PostFormValue("sample_type"))
	instrumentGoodsIdStr := r.PostFormValue("instrument_goods_id")
	classifyStr := r.PostFormValue("classify")
	classify, _ := strconv.Atoi(classifyStr)
	if classify < 1 || classify > 4 {
		classify = 1 // 默认设备
	}

	// 获取检测项目数据
	itemNames := r.Form["item_name[]"]
	itemResults := r.Form["item_result[]"]
	itemUnits := r.Form["item_unit[]"]
	itemMethods := r.Form["item_method[]"]
	itemRemarks := r.Form["item_remark[]"]
	itemAbnormals := r.Form["item_abnormal[]"]

	// 检查是否有检测数据
	hasReportData := reportTitle != "" || reportContent != "" || len(itemNames) > 0

	if hasReportData {
		instrumentGoodsId, _ := strconv.Atoi(instrumentGoodsIdStr)
		examinationReport := dao.SeeSeekExaminationReport{
			SeeSeekID:         seeSeek.Id,
			Classify:          classify,
			Status:            1,
			Name:              reportTitle,
			Description:       reportContent,
			SampleType:        sampleType,
			InstrumentGoodsID: instrumentGoodsId,
			ReportTitle:       reportTitle,
			ReportContent:     reportContent,
			MasterUserId:      s_u.Id,
			ReportDate:        time.Now(),
		}
		if err := examinationReport.Create(); err != nil {
			util.Debug("Cannot create SeeSeekExaminationReport", err)
			report(w, s_u, "保存检测报告时发生错误，请稍后再试")
			return
		}

		// 保存检测项目
		for i, itemName := range itemNames {
			itemName = strings.TrimSpace(itemName)
			if itemName == "" {
				continue
			}

			var itemResult, itemUnit, itemMethod, itemRemark string
			var isAbnormal bool

			if i < len(itemResults) {
				itemResult = strings.TrimSpace(itemResults[i])
			}
			if i < len(itemUnits) {
				itemUnit = strings.TrimSpace(itemUnits[i])
			}
			if i < len(itemMethods) {
				itemMethod = strings.TrimSpace(itemMethods[i])
			}
			if i < len(itemRemarks) {
				itemRemark = strings.TrimSpace(itemRemarks[i])
			}

			// 检查是否有异常标记
			for _, abnormal := range itemAbnormals {
				if abnormal == "1" {
					isAbnormal = true
					break
				}
			}

			examinationItem := dao.SeeSeekExaminationItem{
				Classify:                   classify,
				SeeSeekExaminationReportID: examinationReport.ID,
				ItemName:                   itemName,
				Result:                     itemResult,
				ResultUnit:                 itemUnit,
				Remark:                     itemRemark,
				AbnormalFlag:               isAbnormal,
				Method:                     itemMethod,
				Status:                     1,
			}

			if err := examinationItem.Create(); err != nil {
				util.Debug("Cannot create SeeSeekExaminationItem", err)
				report(w, s_u, "保存检测项目时发生错误，请稍后再试")
				return
			}
		}
	}

	// 更新步骤和状态
	seeSeek.Step = dao.SeeSeekStepReport
	seeSeek.Status = dao.SeeSeekStatusCompleted
	seeSeek.EndTime = endTime
	if err := seeSeek.Update(); err != nil {
		util.Debug("Cannot update SeeSeek step", seeSeek.Uuid, err)
		report(w, s_u, "保存看看记录步骤时发生错误，请稍后再试")
		return
	}

	// 完成所有步骤，跳转到项目详情页
	http.Redirect(w, r, "/v1/project/detail?uuid="+project.Uuid, http.StatusFound)
}
