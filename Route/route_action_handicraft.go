package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/handicraft/project_new
func HandleNewHandicraft(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandicraftNewGet(w, r)
	case http.MethodPost:
		HandicraftNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handler /v1/handicraft/project_detail
func HandleHandicraftDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	HandicraftDetailGet(w, r)
}

// Handler /v1/handicrafts/project_list
func HandleHandicraftList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	HandicraftListGet(w, r)
}

// GET /v1/handicraft/project_new?project_uuid=xxx
func HandicraftNewGet(w http.ResponseWriter, r *http.Request) {
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

	projectUuid := r.URL.Query().Get("project_uuid")
	if projectUuid == "" {
		report(w, r, "项目信息缺失")
		return
	}

	project := data.Project{Uuid: projectUuid}
	if err := project.GetByUuid(); err != nil {
		util.Debug("Cannot get project by uuid", projectUuid, err)
		report(w, r, "项目不存在")
		return
	}

	// 获取项目相关信息
	objective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective", err)
		report(w, r, "获取目标信息失败")
		return
	}

	projectBean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot get project bean", err)
		report(w, r, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		util.Debug("Cannot get objective bean", err)
		report(w, r, "获取目标详情失败")
		return
	}

	// 获取用户的技能和法力
	skills, err := data.GetSkillsByRecordUserId(user.Id, r.Context())
	if err != nil {
		util.Debug("Cannot get skills", err)
		skills = []data.Skill{}
	}

	magics, err := data.GetAllMagics(r.Context())
	if err != nil {
		util.Debug("Cannot get magics", err)
		magics = []data.Magic{}
	}

	templateData := data.HandicraftDetailTemplateData{
		SessUser:               user,
		IsVerifier:             isVerifier(user.Id),
		ProjectBean:            projectBean,
		QuoteObjectiveBean:     objectiveBean,
		Skills:                 skills,
		Magics:                 magics,
		EvidenceHandicraftBean: []data.EvidenceHandicraftBean{},
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "handicraft.new", "component_sess_capacity")
}

// POST /v1/handicraft/project_new
func HandicraftNewPost(w http.ResponseWriter, r *http.Request) {
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

	// 验证必填字段
	name := strings.TrimSpace(r.PostFormValue("name"))
	description := strings.TrimSpace(r.PostFormValue("description"))
	projectUuid := strings.TrimSpace(r.PostFormValue("project_uuid"))

	if name == "" {
		report(w, r, "手艺名称不能为空。")
		return
	}
	if description == "" {
		report(w, r, "手艺描述不能为空。")
		return
	}
	if projectUuid == "" {
		report(w, r, "项目信息缺失。")
		return
	}

	// 获取项目信息
	project := data.Project{Uuid: projectUuid}
	if err := project.GetByUuid(); err != nil {
		util.Debug("Cannot get project by uuid", projectUuid, err)
		report(w, r, "项目不存在")
		return
	}

	// 解析表单数据
	category, _ := strconv.Atoi(r.PostFormValue("category"))
	if category < 1 || category > 6 {
		category = 1 // 默认轻体力
	}

	skillDifficulty, _ := strconv.Atoi(r.PostFormValue("skill_difficulty"))
	if skillDifficulty < 1 || skillDifficulty > 5 {
		skillDifficulty = 3 // 默认中等
	}

	magicDifficulty, _ := strconv.Atoi(r.PostFormValue("magic_difficulty"))
	if magicDifficulty < 1 || magicDifficulty > 5 {
		magicDifficulty = 3 // 默认中等
	}

	initiatorId, _ := strconv.Atoi(r.PostFormValue("initiator_id"))
	if initiatorId <= 0 {
		initiatorId = user.Id
	}

	ownerId, _ := strconv.Atoi(r.PostFormValue("owner_id"))
	if ownerId <= 0 {
		ownerId = user.Id
	}

	handicraft := data.Handicraft{
		RecorderUserId:  user.Id,
		Name:            name,
		Nickname:        strings.TrimSpace(r.PostFormValue("nickname")),
		Description:     description,
		ProjectId:       project.Id,
		InitiatorId:     initiatorId,
		OwnerId:         ownerId,
		Category:        data.HandicraftCategory(category),
		Status:          data.NotStarted,
		SkillDifficulty: skillDifficulty,
		MagicDifficulty: magicDifficulty,
	}

	if err := handicraft.Create(r.Context()); err != nil {
		util.Debug("Cannot create handicraft", err)
		report(w, r, "创建手艺记录失败，请重试。")
		return
	}

	// 处理技能关联
	skillIds := r.Form["skill_ids"]
	for _, skillIdStr := range skillIds {
		if skillId, err := strconv.Atoi(skillIdStr); err == nil && skillId > 0 {
			handicraftSkill := data.HandicraftSkill{
				HandicraftId: handicraft.Id,
				SkillId:      skillId,
			}
			handicraftSkill.Create()
		}
	}

	// 处理法力关联
	magicIds := r.Form["magic_ids"]
	for _, magicIdStr := range magicIds {
		if magicId, err := strconv.Atoi(magicIdStr); err == nil && magicId > 0 {
			handicraftMagic := data.HandicraftMagic{
				HandicraftId: handicraft.Id,
				MagicId:      magicId,
			}
			handicraftMagic.Create()
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/v1/handicraft/project_detail?uuid=%s", handicraft.Uuid), http.StatusFound)
}

// GET /v1/handicraft/project_detail?uuid=xxx
func HandicraftDetailGet(w http.ResponseWriter, r *http.Request) {
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

	handicraft := data.Handicraft{Uuid: uuid}
	if err := handicraft.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get handicraft by uuid", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 获取项目信息
	project := data.Project{Id: handicraft.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project", err)
		report(w, r, "获取项目信息失败")
		return
	}

	objective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective", err)
		report(w, r, "获取目标信息失败")
		return
	}

	projectBean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot get project bean", err)
		report(w, r, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		util.Debug("Cannot get objective bean", err)
		report(w, r, "获取目标详情失败")
		return
	}

	handicraftBean := data.HandicraftBean{
		Handicraft: handicraft,
		IsOpen:     true,
		Project:    project,
	}

	templateData := data.HandicraftDetailTemplateData{
		SessUser:               user,
		IsVerifier:             isVerifier(user.Id),
		HandicraftBean:         handicraftBean,
		ProjectBean:            projectBean,
		QuoteObjectiveBean:     objectiveBean,
		Skills:                 []data.Skill{},
		Magics:                 []data.Magic{},
		EvidenceHandicraftBean: []data.EvidenceHandicraftBean{},
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "handicraft.detail")
}

// GET /v1/handicrafts/project_list
func HandicraftListGet(w http.ResponseWriter, r *http.Request) {
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

	var handicraftData struct {
		SessUser     data.User
		Handicrafts  []data.HandicraftBean
		ProjectBeans []data.ProjectBean
	}
	handicraftData.SessUser = user

	generateHTML(w, &handicraftData, "layout", "navbar.private", "handicraft.list")
}
