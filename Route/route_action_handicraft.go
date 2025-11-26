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
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	if !isVerifier(s_u.Id) {
		report(w, s_u, "你没有权限执行此操作")
		return
	}

	uuid := r.URL.Query().Get("project_uuid")
	if uuid == "" {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	t_proj := data.Project{Uuid: uuid}
	if err := t_proj.GetByUuid(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	t_obje, err := t_proj.Objective()
	if err != nil {
		util.Debug("Cannot get objective given proj_id", t_proj.Id, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	projBean, err := fetchProjectBean(t_proj)
	if err != nil {
		util.Debug("Cannot get projBean", err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	objeBean, err := fetchObjectiveBean(t_obje)
	if err != nil {
		util.Debug("Cannot get objeBean", err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
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
		report(w, s_u, "获取约茶记录失败")
	}
	teamId := 0
	if project_appointment.PayeeTeamId > 0 {
		teamId = project_appointment.PayeeTeamId
	} else if project_appointment.PayeeFamilyId > 0 {
		family, err := data.GetFamily(project_appointment.PayeeFamilyId)
		if err != nil {
			util.Debug("Cannot get family by id", project_appointment.PayeeFamilyId, err)
			report(w, s_u, "获取家庭信息失败")
			return
		}
		// 注意：Family结构体没有DefaultTeamId字段，需要通过其他方式获取
		// 这里假设使用家庭创建者的默认团队
		founder_family, err := data.GetUser(family.AuthorId)
		if err != nil {
			util.Debug("Cannot get family founder", family.AuthorId, err)
			report(w, s_u, "获取家庭创建者信息失败")
			return
		}
		defaultTeam, err := founder_family.GetLastDefaultTeam()
		if err != nil {
			util.Debug("Cannot get founder default team", err)
			report(w, s_u, "获取默认团队信息失败")
			return
		}
		teamId = defaultTeam.Id
	}
	if teamId == 0 {
		report(w, s_u, "团队信息缺失")
		return
	}

	// 获取团队CEO作为默认策动人
	team := data.Team{Id: teamId}
	if err := team.Get(); err != nil {
		util.Debug("Cannot get team", teamId, err)
		report(w, s_u, "获取收茶叶方团队信息失败")
		return
	}
	teamCEO, err := team.MemberCEO()
	if err != nil || teamCEO.UserId <= 0 {
		util.Debug("Cannot get team CEO", teamId, err)
		report(w, s_u, "获取收茶叶方团队CEO失败，无法确定策动人")
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
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
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
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	if !isVerifier(s_u.Id) {
		report(w, s_u, "你没有权限执行此操作")
		return
	}

	if err := r.ParseForm(); err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "表单数据解析失败")
		return
	}

	projectUuid := r.FormValue("project_uuid")
	if projectUuid == "" {
		report(w, s_u, "项目信息缺失")
		return
	}

	t_proj := data.Project{Uuid: projectUuid}
	if err := t_proj.GetByUuid(); err != nil {
		util.Debug("Cannot get project by uuid", projectUuid, err)
		report(w, s_u, "项目不存在")
		return
	}

	name := r.FormValue("name")
	nickname := r.FormValue("nickname")
	description := r.FormValue("description")
	typeStr := r.FormValue("type")
	categoryStr := r.FormValue("category")
	skillDifficultyStr := r.FormValue("skill_difficulty")
	magicDifficultyStr := r.FormValue("magic_difficulty")
	initiatorIdStr := r.FormValue("initiator_id")
	ownerIdStr := r.FormValue("owner_id")

	if name == "" || description == "" {
		report(w, s_u, "请填写完整的基本信息")
		return
	}

	handicraftType, _ := strconv.Atoi(typeStr)
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
		Type:            data.HandicraftType(handicraftType),
		Category:        category,
		Status:          data.NotStarted,
		SkillDifficulty: skillDifficulty,
		MagicDifficulty: magicDifficulty,
	}

	if err := handicraft.Create(r.Context()); err != nil {
		util.Debug("Cannot create handicraft", err)
		report(w, s_u, "创建手工艺记录失败")
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
				report(w, s_u, "创建协助者失败")
				return
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
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 先尝试作为手工艺UUID查询
	handicraft := data.Handicraft{Uuid: uuid}
	if err := handicraft.GetByIdOrUUID(r.Context()); err != nil {
		// 如果不是手工艺UUID，尝试作为项目UUID查询
		if err == sql.ErrNoRows {
			project := data.Project{Uuid: uuid}
			if err := project.GetByUuid(); err == nil {
				// 是项目UUID，重定向到项目的手工艺列表页面
				http.Redirect(w, r, "/v1/handicraft/list?project_uuid="+uuid, http.StatusFound)
				return
			}
			report(w, s_u, "手工艺记录或项目不存在")
			return
		}
		util.Debug("Cannot get handicraft by uuid", uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	project := data.Project{Id: handicraft.ProjectId}
	if err := project.Get(); err != nil {
		util.Debug("Cannot get project", err)
		report(w, s_u, "获取项目信息失败")
		return
	}

	objective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective", err)
		report(w, s_u, "获取目标信息失败")
		return
	}

	projectBean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, s_u, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, s_u, "获取目标详情失败")
		return
	}

	handicraft_bean, err := fetchHandicraftBean(handicraft)
	if err != nil {
		util.Debug("Cannot fetch handicraft bean", err)
		report(w, s_u, "获取手工艺详情失败")
		return
	}

	is_master, err := checkProjectMasterPermission(&project, s_u.Id)
	if err != nil {
		util.Debug("Cannot check project master permission", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	is_admin, err := checkObjectiveAdminPermission(&objective, s_u.Id)
	if err != nil {
		util.Debug("Cannot check objective admin permission", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	is_verifier := isVerifier(s_u.Id)
	is_invited := false
	if handicraft.Category == data.HandicraftCategorySecret {
		if !is_master && !is_admin && !is_invited {
			is_invited, err = objective.IsInvitedMember(s_u.Id)
			if err != nil {
				util.Debug("Cannot check objective invited member", err)
				report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
				return
			}
			if !is_invited {
				report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
				return
			}
		}
	}

	// 获取相关技能
	var skills []data.Skill
	if skillRelations, err := data.GetHandicraftSkills(handicraft.Id); err == nil {
		for _, sr := range skillRelations {
			skill := data.Skill{Id: sr.SkillId}
			if err := skill.GetByIdOrUUID(r.Context()); err == nil {
				skills = append(skills, skill)
			}
		}
	}

	// 获取相关法力
	var magics []data.Magic
	if magicRelations, err := data.GetHandicraftMagics(handicraft.Id); err == nil {
		for _, mr := range magicRelations {
			magic := data.Magic{Id: mr.MagicId}
			if err := magic.GetByIdOrUUID(r.Context()); err == nil {
				magics = append(magics, magic)
			}
		}
	}

	// 获取相关凭证
	var evidenceHandicraftBean []data.EvidenceHandicraftBean
	if evidences, err := data.GetEvidencesByHandicraftId(handicraft.Id); err == nil {
		evidenceHandicraftBean = append(evidenceHandicraftBean, data.EvidenceHandicraftBean{
			Evidences:  evidences,
			Handicraft: handicraft,
		})
	}

	templateData := data.HandicraftDetailTemplateData{
		SessUser:               s_u,
		IsMaster:               is_master,
		IsAdmin:                is_admin,
		IsVerifier:             is_verifier,
		IsInvited:              is_invited,
		HandicraftBean:         handicraft_bean,
		ProjectBean:            projectBean,
		QuoteObjectiveBean:     objectiveBean,
		Skills:                 skills,
		Magics:                 magics,
		EvidenceHandicraftBean: evidenceHandicraftBean,
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "action.handicraft.detail", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}

// Handler /v1/handicraft/list
func HandleHandicraftList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	HandicraftListGet(w, r)
}

// GET /v1/handicraft/list?project_uuid=xxx
func HandicraftListGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	projectUuid := r.URL.Query().Get("project_uuid")
	if projectUuid == "" {
		report(w, s_u, "项目信息缺失")
		return
	}

	project := data.Project{Uuid: projectUuid}
	if err := project.GetByUuid(); err != nil {
		util.Debug("Cannot get project by uuid", projectUuid, err)
		report(w, s_u, "项目不存在")
		return
	}

	objective, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective", err)
		report(w, s_u, "获取目标信息失败")
		return
	}

	projectBean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, s_u, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, s_u, "获取目标详情失败")
		return
	}

	// 获取项目的所有手工艺记录
	handicrafts, err := data.GetHandicraftsByProjectId(project.Id, r.Context())
	if err != nil {
		util.Debug("Cannot get handicrafts by project id", project.Id, err)
		handicrafts = []data.Handicraft{}
	}

	// 将 Handicraft 转换为 HandicraftBean
	var handicraftBeans []data.HandicraftBean
	for _, h := range handicrafts {
		bean, err := fetchHandicraftBean(h)
		if err != nil {
			util.Debug("Cannot fetch handicraft bean", err)
			continue
		}
		handicraftBeans = append(handicraftBeans, bean)
	}

	templateData := struct {
		SessUser           data.User
		HandicraftBeans    []data.HandicraftBean
		ProjectBean        data.ProjectBean
		QuoteObjectiveBean data.ObjectiveBean
	}{
		SessUser:           s_u,
		HandicraftBeans:    handicraftBeans,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "handicraft.list", "component_handicraft_bean", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}
