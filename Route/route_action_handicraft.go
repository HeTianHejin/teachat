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

	// 获取收茶叶方登记的技能和法力
	// 读取茶围的约茶记录，确定收茶叶方团队
	project_appointment, err := data.GetAppointmentByProjectId(project.Id, r.Context())
	if err != nil {
		util.Debug("Cannot get project appointment by project id", project.Id, err)
		report(w, r, "获取约茶记录失败")
	}
	teamId := 0
	if project_appointment.PayeeTeamId > 0 {
		teamId = project_appointment.PayeeTeamId
	} else if project_appointment.PayeeFamilyId > 0 {
		family, err := data.GetFamily(project_appointment.PayeeFamilyId)
		if err != nil {
			util.Debug("Cannot get family by id", project_appointment.PayeeFamilyId, err)
			report(w, r, "获取家庭信息失败")
			return
		}
		// 注意：Family结构体没有DefaultTeamId字段，需要通过其他方式获取
		// 这里假设使用家庭创建者的默认团队
		founder_family, err := data.GetUser(family.AuthorId)
		if err != nil {
			util.Debug("Cannot get family founder", family.AuthorId, err)
			report(w, r, "获取家庭创建者信息失败")
			return
		}
		defaultTeam, err := founder_family.GetLastDefaultTeam()
		if err != nil {
			util.Debug("Cannot get founder default team", err)
			report(w, r, "获取默认团队信息失败")
			return
		}
		teamId = defaultTeam.Id
	}
	if teamId == 0 {
		report(w, r, "团队信息缺失")
		return
	}

	// 获取团队CEO作为默认策动人
	team := data.Team{Id: teamId}
	if err := team.Get(); err != nil {
		util.Debug("Cannot get team", teamId, err)
		report(w, r, "获取收茶叶方团队信息失败")
		return
	}
	teamCEO, err := team.MemberCEO()
	if err != nil || teamCEO.UserId <= 0 {
		util.Debug("Cannot get team CEO", teamId, err)
		report(w, r, "获取收茶叶方团队CEO失败，无法确定策动人")
		return
	}
	defaultInitiatorId := teamCEO.UserId

	// 获取团队技能
	skillTeams, err := data.GetTeamSkills(teamId, r.Context())
	if err != nil {
		util.Debug("Cannot get team skills", teamId, err)
		skillTeams = []data.SkillTeam{}
	}
	// 获取公开的技能详情
	var skills []data.Skill
	for _, st := range skillTeams {
		var skill data.Skill
		skill.Id = st.SkillId
		if err := skill.GetByIdOrUUID(r.Context()); err == nil {
			skills = append(skills, skill)
		}
	}
	// 获取团队公开登记的法力列表
	magicTeams, err := data.GetTeamMagics(teamId, r.Context())
	if err != nil {
		util.Debug("Cannot get team magics", teamId, err)
		magicTeams = []data.MagicTeam{}
	}
	// 获取公开的法力详情
	var magics []data.Magic
	for _, mt := range magicTeams {
		var magic data.Magic
		magic.Id = mt.MagicId
		if err := magic.GetByIdOrUUID(r.Context()); err == nil {
			magics = append(magics, magic)
		}
	}

	is_master, err := checkProjectMasterPermission(&project, user.Id)
	if err != nil {
		util.Debug(" Cannot check project master permission", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	templateData := struct {
		SessUser               data.User
		IsAdmin                bool
		IsMaster               bool
		IsVerifier             bool
		ProjectBean            data.ProjectBean
		QuoteObjectiveBean     data.ObjectiveBean
		Skills                 []data.Skill
		Magics                 []data.Magic
		DefaultInitiatorId     int
		EvidenceHandicraftBean []data.EvidenceHandicraftBean
	}{
		SessUser:               user,
		IsMaster:               is_master,
		IsVerifier:             isVerifier(user.Id),
		ProjectBean:            projectBean,
		QuoteObjectiveBean:     objectiveBean,
		Skills:                 skills,
		Magics:                 magics,
		DefaultInitiatorId:     defaultInitiatorId,
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
		report(w, r, "主理人ID不能为空。")
		return
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

	// 处理协助者关联
	contributorIds := r.Form["contributor_ids[]"]
	contributorRates := r.Form["contributor_rates[]"]
	contributorCount := 0
	for i, contributorIdStr := range contributorIds {
		if contributorIdStr == "" {
			continue
		}
		contributorId, err := strconv.Atoi(contributorIdStr)
		if err != nil || contributorId <= 0 {
			continue
		}
		contributionRate := 0
		if i < len(contributorRates) {
			contributionRate, _ = strconv.Atoi(contributorRates[i])
		}
		if contributionRate < 1 || contributionRate > 100 {
			contributionRate = 50 // 默认贡献值
		}
		contributor := data.HandicraftContributor{
			HandicraftId:     handicraft.Id,
			UserId:           contributorId,
			ContributionRate: contributionRate,
		}
		if err := contributor.Create(); err == nil {
			contributorCount++
		}
	}
	// 更新协助者计数
	if contributorCount > 0 {
		handicraft.ContributorCount = contributorCount
		handicraft.Update()
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
