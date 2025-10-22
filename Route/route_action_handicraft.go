package route

import (
	"database/sql"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/handicraft/new
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

// GET /v1/handicraft/new?project_uuid=xxx
func HandicraftNewGet(w http.ResponseWriter, r *http.Request) {
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

	uuid := r.URL.Query().Get("project_uuid")
	if uuid == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	t_proj := data.Project{Uuid: uuid}
	if err := t_proj.GetByUuid(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	t_obje, err := t_proj.Objective()
	if err != nil {
		util.Debug("Cannot get objective given proj_id", t_proj.Id, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	projBean, err := fetchProjectBean(t_proj)
	if err != nil {
		util.Debug("Cannot get projBean", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	objeBean, err := fetchObjectiveBean(t_obje)
	if err != nil {
		util.Debug("Cannot get objeBean", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 检查是否已存在手工艺记录
	existingHandicraft, err := data.GetHandicraftsByProjectId(t_proj.Id, r.Context())
	if err == nil && len(existingHandicraft) > 0 {
		// 已存在记录，跳转到详情页
		http.Redirect(w, r, "/v1/handicraft/detail?uuid="+existingHandicraft[0].Uuid, http.StatusFound)
		return
	}

	// 获取收茶叶方登记的技能和法力
	// 读取茶围的约茶记录，确定收茶叶方团队
	project_appointment, err := data.GetAppointmentByProjectId(t_proj.Id, r.Context())
	if err != nil {
		util.Debug("Cannot get project appointment by project id", t_proj.Id, err)
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

	is_master, err := checkProjectMasterPermission(&t_proj, s_u.Id)
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
		SessUser:               s_u,
		IsMaster:               is_master,
		IsVerifier:             isVerifier(s_u.Id),
		ProjectBean:            projBean,
		QuoteObjectiveBean:     objeBean,
		Skills:                 skills,
		Magics:                 magics,
		DefaultInitiatorId:     defaultInitiatorId,
		EvidenceHandicraftBean: []data.EvidenceHandicraftBean{},
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "action.handicraft.new", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/handicraft/new
func HandicraftNewPost(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "表单数据解析失败")
		return
	}

	projectUuid := r.FormValue("project_uuid")
	if projectUuid == "" {
		report(w, r, "项目信息缺失")
		return
	}

	t_proj := data.Project{Uuid: projectUuid}
	if err := t_proj.GetByUuid(); err != nil {
		util.Debug("Cannot get project by uuid", projectUuid, err)
		report(w, r, "项目不存在")
		return
	}

	name := r.FormValue("name")
	nickname := r.FormValue("nickname")
	description := r.FormValue("description")
	categoryStr := r.FormValue("category")
	skillDifficultyStr := r.FormValue("skill_difficulty")
	magicDifficultyStr := r.FormValue("magic_difficulty")
	initiatorIdStr := r.FormValue("initiator_id")
	ownerIdStr := r.FormValue("owner_id")

	if name == "" || description == "" {
		report(w, r, "请填写完整的基本信息")
		return
	}

	category, _ := strconv.Atoi(categoryStr)
	skillDifficulty, _ := strconv.Atoi(skillDifficultyStr)
	magicDifficulty, _ := strconv.Atoi(magicDifficultyStr)
	initiatorId, _ := strconv.Atoi(initiatorIdStr)
	ownerId, _ := strconv.Atoi(ownerIdStr)

	if ownerId <= 0 {
		ownerId = s_u.Id
	}
	if initiatorId <= 0 {
		initiatorId = s_u.Id
	}
	if skillDifficulty <= 0 {
		skillDifficulty = 3
	}
	if magicDifficulty <= 0 {
		magicDifficulty = 3
	}

	handicraft := data.Handicraft{
		RecorderUserId:  s_u.Id,
		Name:            name,
		Nickname:        nickname,
		Description:     description,
		ProjectId:       t_proj.Id,
		InitiatorId:     initiatorId,
		OwnerId:         ownerId,
		Category:        data.HandicraftCategory(category),
		Status:          data.NotStarted,
		SkillDifficulty: skillDifficulty,
		MagicDifficulty: magicDifficulty,
	}

	if err := handicraft.Create(r.Context()); err != nil {
		util.Debug("Cannot create handicraft", err)
		report(w, r, "创建手工艺记录失败")
		return
	}

	// 处理协助者
	contributorIds := r.Form["contributor_ids[]"]
	contributorRates := r.Form["contributor_rates[]"]
	for i := 0; i < len(contributorIds) && i < len(contributorRates); i++ {
		contribId, err1 := strconv.Atoi(contributorIds[i])
		contribRate, err2 := strconv.Atoi(contributorRates[i])
		if err1 == nil && err2 == nil && contribId > 0 && contribRate > 0 {
			contributor := data.HandicraftContributor{
				HandicraftId:     handicraft.Id,
				UserId:           contribId,
				ContributionRate: contribRate,
			}
			if err := contributor.Create(); err != nil {
				util.Debug("Cannot create handicraft contributor", err)
			}
		}
	}

	// 处理技能关联
	skillIds := r.Form["skill_ids"]
	for _, skillIdStr := range skillIds {
		skillId, err := strconv.Atoi(skillIdStr)
		if err == nil && skillId > 0 {
			handicraftSkill := data.HandicraftSkill{
				SkillId:      skillId,
				HandicraftId: handicraft.Id,
			}
			if err := handicraftSkill.Create(); err != nil {
				util.Debug("Cannot create skill handicraft relation", err)
			}
		}
	}

	// 处理法力关联
	magicIds := r.Form["magic_ids"]
	for _, magicIdStr := range magicIds {
		magicId, err := strconv.Atoi(magicIdStr)
		if err == nil && magicId > 0 {
			handicraftMagic := data.HandicraftMagic{
				MagicId:      magicId,
				HandicraftId: handicraft.Id,
			}
			if err := handicraftMagic.Create(); err != nil {
				util.Debug("Cannot create magic handicraft relation", err)
			}
		}
	}

	http.Redirect(w, r, "/v1/handicraft/step2?uuid="+handicraft.Uuid, http.StatusFound)
}

// Handler /v1/handicraft/detail
func HandleHandicraftDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	HandicraftDetailGet(w, r)
}

// GET /v1/handicraft/detail?uuid=xxx
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
		if err == sql.ErrNoRows {
			report(w, r, "手工艺记录不存在")
			return
		}
		util.Debug("Cannot get handicraft by uuid", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

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

	templateData := struct {
		SessUser           data.User
		Handicraft         data.Handicraft
		ProjectBean        data.ProjectBean
		QuoteObjectiveBean data.ObjectiveBean
	}{
		SessUser:           user,
		Handicraft:         handicraft,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "action.handicraft.detail", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}
