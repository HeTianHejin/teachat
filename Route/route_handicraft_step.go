package route

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/handicraft/step2
func HandleHandicraftStep2(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandicraftStep2Get(w, r)
	case http.MethodPost:
		HandicraftStep2Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/handicraft/step2?uuid=xxx
func HandicraftStep2Get(w http.ResponseWriter, r *http.Request) {
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

	if !isVerifier(s_u.Id) {
		report(w, r, "你没有权限执行此操作")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	handicraft := data.Handicraft{Uuid: uuid}
	if err = handicraft.GetByIdOrUUID(r.Context()); err != nil {
		if err == sql.ErrNoRows {
			report(w, r, "手工艺记录不存在")
			return
		}
		util.Debug("Cannot get handicraft by uuid", err)
		report(w, r, "处理手工艺记录时发生错误")
		return
	}

	project := data.Project{Id: handicraft.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	objective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
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
	is_master, err := checkProjectMasterPermission(&project, s_u.Id)
	if err != nil {
		util.Debug(" Cannot check project master permission", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	templateData := struct {
		SessUser           data.User
		IsMaster           bool
		IsAdmin            bool
		IsVerifier         bool
		Handicraft         data.Handicraft
		ProjectBean        data.ProjectBean
		QuoteObjectiveBean data.ObjectiveBean
		CurrentStep        int
	}{
		SessUser:           s_u,
		IsMaster:           is_master,
		IsVerifier:         isVerifier(s_u.Id),
		Handicraft:         handicraft,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
		CurrentStep:        2,
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "action.handicraft.step2", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/handicraft/step2
func HandicraftStep2Post(w http.ResponseWriter, r *http.Request) {
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
	if !isVerifier(s_u.Id) {
		report(w, r, "你没有权限执行此操作")
		return
	}

	handicraftUuid := r.PostFormValue("handicraft_uuid")
	skillDifficultyStr := r.PostFormValue("skill_difficulty")
	magicDifficultyStr := r.PostFormValue("magic_difficulty")

	handicraft := data.Handicraft{Uuid: handicraftUuid}
	if err := handicraft.GetByIdOrUUID(r.Context()); err != nil {
		report(w, r, "手工艺记录不存在")
		return
	}

	skillDifficulty, _ := strconv.Atoi(skillDifficultyStr)
	magicDifficulty, _ := strconv.Atoi(magicDifficultyStr)

	if skillDifficulty < 1 || skillDifficulty > 5 {
		skillDifficulty = 1
	}
	if magicDifficulty < 1 || magicDifficulty > 5 {
		magicDifficulty = 1
	}

	handicraft.SkillDifficulty = skillDifficulty
	handicraft.MagicDifficulty = magicDifficulty

	if err := handicraft.Update(); err != nil {
		util.Debug("Cannot update handicraft", err)
		report(w, r, "保存难度信息失败")
		return
	}

	http.Redirect(w, r, "/v1/handicraft/step3?uuid="+handicraftUuid, http.StatusFound)
}

// Handler /v1/handicraft/step3
func HandleHandicraftStep3(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandicraftStep3Get(w, r)
	case http.MethodPost:
		HandicraftStep3Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/handicraft/step3?uuid=xxx
func HandicraftStep3Get(w http.ResponseWriter, r *http.Request) {
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

	handicraft := data.Handicraft{Uuid: uuid}
	if err = handicraft.GetByIdOrUUID(r.Context()); err != nil {
		report(w, r, "手工艺记录不存在")
		return
	}

	project := data.Project{Id: handicraft.ProjectId}
	if err := project.Get(); err != nil {
		report(w, r, "项目不存在")
		return
	}

	objective, err := project.Objective()
	if err != nil {
		report(w, r, "获取目标信息失败")
		return
	}

	projectBean, err := fetchProjectBean(project)
	if err != nil {
		report(w, r, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		report(w, r, "获取目标详情失败")
		return
	}

	is_master, err := checkProjectMasterPermission(&project, s_u.Id)
	if err != nil {
		util.Debug("Cannot check project master permission", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	templateData := struct {
		SessUser           data.User
		IsMaster           bool
		IsAdmin            bool
		IsVerifier         bool
		Handicraft         data.Handicraft
		ProjectBean        data.ProjectBean
		QuoteObjectiveBean data.ObjectiveBean
		CurrentStep        int
	}{
		SessUser:           s_u,
		IsMaster:           is_master,
		IsVerifier:         isVerifier(s_u.Id),
		Handicraft:         handicraft,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
		CurrentStep:        3,
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "action.handicraft.step3", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/handicraft/step3
func HandicraftStep3Post(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	handicraftUuid := r.PostFormValue("handicraft_uuid")
	inaugurationName := strings.TrimSpace(r.PostFormValue("inauguration_name"))
	inaugurationDesc := strings.TrimSpace(r.PostFormValue("inauguration_desc"))
	evidenceIdStr := r.PostFormValue("evidence_id")

	handicraft := data.Handicraft{Uuid: handicraftUuid}
	if err := handicraft.GetByIdOrUUID(r.Context()); err != nil {
		report(w, r, "手工艺记录不存在")
		return
	}

	if inaugurationName != "" {
		evidenceId, _ := strconv.Atoi(evidenceIdStr)
		inauguration := data.Inauguration{
			HandicraftId:   handicraft.Id,
			Name:           inaugurationName,
			Description:    inaugurationDesc,
			RecorderUserId: s_u.Id,
			EvidenceId:     evidenceId,
			Status:         1,
		}
		if err := inauguration.Create(); err != nil {
			util.Debug("Cannot create inauguration", err)
			report(w, r, "保存开工仪式记录失败")
			return
		}
	}

	handicraft.Status = data.InProgress
	if err := handicraft.Update(); err != nil {
		util.Debug("Cannot update handicraft status", err)
		report(w, r, "更新状态失败")
		return
	}

	http.Redirect(w, r, "/v1/handicraft/step4?uuid="+handicraftUuid, http.StatusFound)
}

// Handler /v1/handicraft/step4
func HandleHandicraftStep4(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandicraftStep4Get(w, r)
	case http.MethodPost:
		HandicraftStep4Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/handicraft/step4?uuid=xxx
func HandicraftStep4Get(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	handicraft := data.Handicraft{Uuid: uuid}
	if err = handicraft.GetByIdOrUUID(r.Context()); err != nil {
		report(w, r, "手工艺记录不存在")
		return
	}

	project := data.Project{Id: handicraft.ProjectId}
	if err := project.Get(); err != nil {
		report(w, r, "项目不存在")
		return
	}

	objective, err := project.Objective()
	if err != nil {
		report(w, r, "获取目标信息失败")
		return
	}

	projectBean, err := fetchProjectBean(project)
	if err != nil {
		report(w, r, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		report(w, r, "获取目标详情失败")
		return
	}

	is_master, err := checkProjectMasterPermission(&project, s_u.Id)
	if err != nil {
		util.Debug("Cannot check project master permission", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	templateData := struct {
		SessUser           data.User
		IsMaster           bool
		IsAdmin            bool
		IsVerifier         bool
		Handicraft         data.Handicraft
		ProjectBean        data.ProjectBean
		QuoteObjectiveBean data.ObjectiveBean
		CurrentStep        int
	}{
		SessUser:           s_u,
		IsMaster:           is_master,
		IsVerifier:         isVerifier(s_u.Id),
		Handicraft:         handicraft,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
		CurrentStep:        4,
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "action.handicraft.step4", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/handicraft/step4
func HandicraftStep4Post(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	handicraftUuid := r.PostFormValue("handicraft_uuid")
	processName := strings.TrimSpace(r.PostFormValue("process_name"))
	processDesc := strings.TrimSpace(r.PostFormValue("process_desc"))

	handicraft := data.Handicraft{Uuid: handicraftUuid}
	if err := handicraft.GetByIdOrUUID(r.Context()); err != nil {
		report(w, r, "手工艺记录不存在")
		return
	}

	if processName != "" {
		processRecord := data.ProcessRecord{
			HandicraftId:   handicraft.Id,
			Name:           processName,
			Description:    processDesc,
			RecorderUserId: s_u.Id,
			Status:         1,
		}
		if err := processRecord.Create(); err != nil {
			util.Debug("Cannot create process record", err)
			report(w, r, "保存过程记录失败")
			return
		}
	}

	http.Redirect(w, r, "/v1/handicraft/step5?uuid="+handicraftUuid, http.StatusFound)
}

// Handler /v1/handicraft/step5
func HandleHandicraftStep5(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandicraftStep5Get(w, r)
	case http.MethodPost:
		HandicraftStep5Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/handicraft/step5?uuid=xxx
func HandicraftStep5Get(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	handicraft := data.Handicraft{Uuid: uuid}
	if err = handicraft.GetByIdOrUUID(r.Context()); err != nil {
		report(w, r, "手工艺记录不存在")
		return
	}

	project := data.Project{Id: handicraft.ProjectId}
	if err := project.Get(); err != nil {
		report(w, r, "项目不存在")
		return
	}

	objective, err := project.Objective()
	if err != nil {
		report(w, r, "获取目标信息失败")
		return
	}

	projectBean, err := fetchProjectBean(project)
	if err != nil {
		report(w, r, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		report(w, r, "获取目标详情失败")
		return
	}

	templateData := struct {
		SessUser           data.User
		Handicraft         data.Handicraft
		ProjectBean        data.ProjectBean
		QuoteObjectiveBean data.ObjectiveBean
		CurrentStep        int
	}{
		SessUser:           s_u,
		Handicraft:         handicraft,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
		CurrentStep:        5,
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "action.handicraft.step5", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/handicraft/step5
func HandicraftStep5Post(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	handicraftUuid := r.PostFormValue("handicraft_uuid")
	endingName := strings.TrimSpace(r.PostFormValue("ending_name"))
	endingDesc := strings.TrimSpace(r.PostFormValue("ending_desc"))
	statusStr := r.PostFormValue("status")

	handicraft := data.Handicraft{Uuid: handicraftUuid}
	if err := handicraft.GetByIdOrUUID(r.Context()); err != nil {
		report(w, r, "手工艺记录不存在")
		return
	}

	if endingName != "" {
		ending := data.Ending{
			HandicraftId:   handicraft.Id,
			Name:           endingName,
			Description:    endingDesc,
			RecorderUserId: s_u.Id,
			Status:         1,
		}
		if err := ending.Create(); err != nil {
			util.Debug("Cannot create ending", err)
			report(w, r, "保存结束仪式记录失败")
			return
		}
	}

	status, _ := strconv.Atoi(statusStr)
	if status < 0 || status > 4 {
		status = int(data.Completed)
	}
	handicraft.Status = data.HandicraftStatus(status)

	if err := handicraft.Update(); err != nil {
		util.Debug("Cannot update handicraft", err)
		report(w, r, "更新手工艺状态失败")
		return
	}

	project := data.Project{Id: handicraft.ProjectId}
	if err := project.Get(); err == nil {
		http.Redirect(w, r, "/v1/project/detail?uuid="+project.Uuid, http.StatusFound)
	} else {
		http.Redirect(w, r, "/v1/handicraft/detail?uuid="+handicraftUuid, http.StatusFound)
	}
}
