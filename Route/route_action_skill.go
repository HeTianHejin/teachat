package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	dao "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/skill/new
func HandleNewSkill(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("cannot get s_u from session %v", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	switch r.Method {
	case http.MethodGet:
		SkillNewGet(s_u, w, r)
	case http.MethodPost:
		SkillNewPost(s_u, w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handler /v1/skill/detail
func HandleSkillDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("cannot get s_u from session %v", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	SkillDetailGet(s_u, w, r)
}

// Handler /v1/skills/user_list
func HandleSkillsUserList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("cannot get s_u from session %v", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	SkillsUserListGet(s_u, w, r)
}

// GET /v1/skill/new?user_id=123&team_id=0/456
func SkillNewGet(s_u dao.User, w http.ResponseWriter, r *http.Request) {
	// 获取user_id参数
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		report(w, s_u, "你好，缺少用户ID参数，请确认后再试。")
		return
	}
	intUserId, err := strconv.Atoi(userId)
	if err != nil || intUserId <= dao.UserId_None {
		report(w, s_u, "你好，无效的用户ID参数，请确认后再试。")
		return
	}
	if intUserId != s_u.Id {
		report(w, s_u, "你没有权限为其他用户创建技能记录。")
		return
	}

	// 获取team_id参数
	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		report(w, s_u, "你好，缺少团队ID参数，请确认后再试。")
		return
	}
	intTeamId, err := strconv.Atoi(teamIdStr)
	if err != nil || intTeamId < 0 || intTeamId == dao.TeamIdFreelancer {
		report(w, s_u, "你好，无效的团队ID参数，请确认后再试。")
		return
	}

	team := dao.Team{Id: 0} //默认声明是个人技能，与团队无关
	if intTeamId != dao.TeamIdNone {
		// 获取用户所在的团队
		team, err = dao.GetTeam(intTeamId)
		if err != nil {
			util.Error("cannot fetch team %d, error: %v", intTeamId, err)
			report(w, s_u, "你好，团队不存在或团队参数无效，请确认后再试。")
			return
		}
		// 检查是否目标团队的核心成员
		isCoreMember, err := team.IsCoreMember(s_u.Id)
		if err != nil {
			util.Error("cannot check userId %d isCoreMember team %d, error: %v", intUserId, intTeamId, err)
			report(w, s_u, "你好，茶博士找不到放大镜，未能帮忙查询团队资料，请先喝茶。")
			return
		}
		if !isCoreMember {
			report(w, s_u, "你好，权限不足，必须是核心成员才能为团队登记新技能。")
			return
		}
	}

	var skillData struct {
		SessUser dao.User
		Team     dao.Team
	}
	skillData.SessUser = s_u
	skillData.Team = team

	generateHTML(w, &skillData, "layout", "navbar.private", "skill.new")
}

// POST /v1/skill/new
func SkillNewPost(s_u dao.User, w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	userIdStr := r.PostFormValue("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if userIdStr == "" || err != nil || userId != s_u.Id {
		report(w, s_u, "你好，用户参数无效，请确认后再试。")
		return
	}

	teamIdStr := r.PostFormValue("team_id")
	if teamIdStr == "" {
		report(w, s_u, "你好，缺少团队ID参数，请确认后再试。")
		return
	}
	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil || teamId < dao.TeamIdNone || teamId == dao.TeamIdFreelancer {
		report(w, s_u, "你好，无效的团队ID参数，请确认后再试。")
		return
	}

	var team dao.Team
	if teamId != dao.TeamIdNone {
		team, err = dao.GetTeam(teamId)
		if err != nil {
			report(w, s_u, "你好，团队不存在或团队参数无效，请确认后再试。")
			return
		}
		isCoreMember, err := team.IsCoreMember(s_u.Id)
		if err != nil {
			util.Error("check core member failed, user %d team %d: %v", s_u.Id, teamId, err)
			report(w, s_u, "你好，开水房云雾缭绕，请先喝茶。")
			return
		}
		if !isCoreMember {
			report(w, s_u, "你好，权限不足，必须是核心成员才能登记团队新技能。")
			return
		}
	}

	// 验证必填字段
	name := strings.TrimSpace(r.PostFormValue("name"))
	description := strings.TrimSpace(r.PostFormValue("description"))

	if name == "" {
		report(w, s_u, "技能名称不能为空。")
		return
	}
	if description == "" {
		report(w, s_u, "技能描述不能为空。")
		return
	}

	// 解析表单数据
	category, _ := strconv.Atoi(r.PostFormValue("category"))
	if category < 1 || category > 2 {
		category = 2 // 默认通用硬技能
	}

	strengthLevel, _ := strconv.Atoi(r.PostFormValue("strength_level"))
	if strengthLevel < 1 || strengthLevel > 5 {
		strengthLevel = 3 // 默认中等
	}

	difficultyLevel, _ := strconv.Atoi(r.PostFormValue("difficulty_level"))
	if difficultyLevel < 1 || difficultyLevel > 5 {
		difficultyLevel = 3 // 默认中等
	}

	level, _ := strconv.Atoi(r.PostFormValue("level"))
	if level < 1 || level > 5 {
		level = 1 // 默认入门
	}

	skill := dao.Skill{
		UserId:          s_u.Id,
		Name:            name,
		Nickname:        strings.TrimSpace(r.PostFormValue("nickname")),
		Description:     description,
		StrengthLevel:   dao.StrengthLevel(strengthLevel),
		DifficultyLevel: dao.DifficultyLevel(difficultyLevel),
		Category:        dao.SkillCategory(category),
		Level:           level,
	}

	if err := skill.Create(r.Context()); err != nil {
		util.Debug("user %d cannot create  skill %v", s_u.Id, err)
		report(w, s_u, "创建技能记录失败，请重试。")
		return
	}

	addMine := r.PostForm.Has("add_to_my_skills")
	addTeam := r.PostForm.Has("add_to_team_skills")

	// 检查是否添加到个人技能列表
	//addToMySkills := r.PostFormValue("add_to_my_skills") == "1"
	if addMine {
		skillUser := dao.SkillUser{
			SkillId: skill.Id,
			UserId:  s_u.Id,
			Level:   1,                         // 默认等级1
			Status:  dao.NormalSkillUserStatus, // 默认中能状态
		}
		if err := skillUser.Create(r.Context()); err != nil {
			util.Debug("cannot create skill s_u %d record %v", s_u.Id, err)
			// 不阻止流程，仅记录错误
		}
	}

	// 检查是否添加到团队技能列表
	addToTeamSkills := r.PostFormValue("add_to_team_skills") == "1"
	if addTeam && teamId == dao.TeamIdNone {
		util.Info("user %d add team skill but no team target", s_u.Id)
		report(w, s_u, "你好，团队参数异常，请先喝茶。")
		return
	}
	if addTeam && addToTeamSkills && teamId != dao.TeamIdNone {
		// 创建团队技能记录
		skillTeam := dao.SkillTeam{
			SkillId: skill.Id,
			TeamId:  teamId,
			Level:   1,                         // 默认等级1
			Status:  dao.NormalSkillTeamStatus, // 默认正常状态
		}
		if err := skillTeam.Create(r.Context()); err != nil {
			util.Error("user %d cannot create skill team %d record: %v", s_u.Id, teamId, err)
			// 不阻止流程，仅记录错误
		}

		// 为用户返回团队技能列表页面
		teamURL := "/v1/skills/team_list?uuid=" + team.Uuid
		http.Redirect(w, r, teamURL, http.StatusFound)
		return
	}

	userURL := "/v1/skills/user_list?uuid=" + s_u.Uuid
	http.Redirect(w, r, userURL, http.StatusFound)
}

// GET /v1/skill/detail?id=123
func SkillDetailGet(s_u dao.User, w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")
	uuidStr := r.URL.Query().Get("uuid")
	if idStr == "" && uuidStr == "" {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var skill dao.Skill
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
		skill.Id = id
	} else {
		skill.Uuid = uuidStr
	}

	if err := skill.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get skill by id/uuid %v", err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var skillData struct {
		SessUser dao.User
		Skill    dao.Skill
	}
	skillData.SessUser = s_u
	skillData.Skill = skill

	generateHTML(w, &skillData, "layout", "navbar.private", "skill.detail")
}

// GET /v1/skills/user_list
func SkillsUserListGet(s_u dao.User, w http.ResponseWriter, r *http.Request) {

	// 确保用户拥有默认技能
	if err := dao.EnsureDefaultSkills(s_u.Id, r.Context()); err != nil {
		util.Debug("cannot ensure default skills for s_u: %v", err)
	}

	// 获取SkillUserBean
	skillUserBean, err := fetchSkillUserBean(s_u, r.Context())
	if err != nil {
		util.Debug("cannot fetch skill s_u bean: %v", err)
		report(w, s_u, "获取茶友技能列表失败，请重试。")
		return
	}

	// 创建技能与用户技能的映射
	skillUserMap := make(map[int]dao.SkillUser)
	for _, skillUser := range skillUserBean.SkillUsers {
		skillUserMap[skillUser.SkillId] = skillUser
	}

	// 创建包含技能和用户信息的结构
	type SkillWithUserInfo struct {
		Skill     dao.Skill
		SkillUser dao.SkillUser
	}

	// 按技能类型分组
	var hardSkills, softSkills []SkillWithUserInfo
	for _, skill := range skillUserBean.Skills {
		if skillUser, exists := skillUserMap[skill.Id]; exists {
			skillWithInfo := SkillWithUserInfo{
				Skill:     skill,
				SkillUser: skillUser,
			}
			switch skill.Category {
			case dao.GeneralHardSkill:
				hardSkills = append(hardSkills, skillWithInfo)
			case dao.GeneralSoftSkill:
				softSkills = append(softSkills, skillWithInfo)
			}
		}
	}

	var SkillDetailTemplateData struct {
		SessUser       dao.User
		SkillUserBean  dao.SkillUserBean
		HardSkills     []SkillWithUserInfo
		SoftSkills     []SkillWithUserInfo
		HardSkillCount int
		SoftSkillCount int
	}

	SkillDetailTemplateData.SessUser = s_u
	SkillDetailTemplateData.SkillUserBean = skillUserBean
	SkillDetailTemplateData.HardSkills = hardSkills
	SkillDetailTemplateData.SoftSkills = softSkills
	SkillDetailTemplateData.HardSkillCount = len(hardSkills)
	SkillDetailTemplateData.SoftSkillCount = len(softSkills)

	generateHTML(w, &SkillDetailTemplateData, "layout", "navbar.private", "skills.user_list", "component_user_skill_bean")
}

// Handler /v1/skill_user/edit
func HandleSkillUserEdit(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("cannot get s_u from session %v", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	switch r.Method {
	case http.MethodGet:
		SkillUserEditGet(s_u, w, r)
	case http.MethodPost:
		SkillUserEditPost(s_u, w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/skill_user/edit?id=123
func SkillUserEditGet(s_u dao.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		report(w, s_u, "缺少技能记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, s_u, "无效的技能记录ID。")
		return
	}

	// 获取技能用户记录
	var skillUser dao.SkillUser
	if err := skillUser.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill s_u by id %v", err)
		report(w, s_u, "技能记录不存在。")
		return
	}

	// 权限检查：只有同一家庭的parents成员可以编辑
	if skillUser.UserId != s_u.Id {
		// 获取目标用户的默认家庭
		targetUser, err := dao.GetUser(skillUser.UserId)
		if err != nil {
			report(w, s_u, "权限验证失败。")
			return
		}

		targetFamily, err := targetUser.GetLastDefaultFamily()
		if err != nil {
			report(w, s_u, "权限验证失败。")
			return
		}

		// 检查当前用户是否为该家庭的parent成员
		isParent, err := targetFamily.IsParentMember(s_u.Id)
		if err != nil || !isParent {
			report(w, s_u, "您没有权限编辑此技能记录。")
			return
		}
	}

	// 获取技能信息
	var skill dao.Skill
	skill.Id = skillUser.SkillId
	if err := skill.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get skill by id %v", err)
		report(w, s_u, "技能信息获取失败。")
		return
	}

	var editData struct {
		SessUser  dao.User
		SkillUser dao.SkillUser
		Skill     dao.Skill
		ReturnURL string
	}
	editData.SessUser = s_u
	editData.SkillUser = skillUser
	editData.Skill = skill
	editData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &editData, "layout", "navbar.private", "skill_user.edit")
}

// POST /v1/skill_user/edit
func SkillUserEditPost(s_u dao.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.PostFormValue("id")
	if idStr == "" {
		report(w, s_u, "缺少技能记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, s_u, "无效的技能记录ID。")
		return
	}

	// 获取原始技能用户记录
	var skillUser dao.SkillUser
	if err := skillUser.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill s_u by id %v", err)
		report(w, s_u, "技能记录不存在。")
		return
	}

	// 权限检查
	if skillUser.UserId != s_u.Id {
		targetUser, err := dao.GetUser(skillUser.UserId)
		if err != nil {
			report(w, s_u, "权限验证失败。")
			return
		}

		targetFamily, err := targetUser.GetLastDefaultFamily()
		if err != nil {
			report(w, s_u, "权限验证失败。")
			return
		}

		isParent, err := targetFamily.IsParentMember(s_u.Id)
		if err != nil || !isParent {
			report(w, s_u, "您没有权限编辑此技能记录。")
			return
		}
	}

	// 解析表单数据
	level, _ := strconv.Atoi(r.PostFormValue("level"))
	if level < 1 || level > 9 {
		report(w, s_u, "技能等级必须在1-9之间。")
		return
	}

	status, _ := strconv.Atoi(r.PostFormValue("status"))
	if status < 0 || status > 3 {
		report(w, s_u, "技能状态值无效。")
		return
	}

	// 更新技能用户记录
	skillUser.Level = level
	skillUser.Status = dao.SkillUserStatus(status)

	if err := skillUser.Update(); err != nil {
		util.Debug("cannot update skill s_u %v", err)
		report(w, s_u, "更新技能记录失败，请重试。")
		return
	}

	// 获取返回URL
	returnURL := r.PostFormValue("return_url")
	if returnURL == "" {
		returnURL = "/v1/skills/user_list"
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}

// Handler /v1/skills/team_list
func HandleSkillsTeamList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("cannot get s_u from session %v", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	SkillsTeamListGet(s_u, w, r)
}

// GET /v1/skills/team_list?uuid=xxx
func SkillsTeamListGet(s_u dao.User, w http.ResponseWriter, r *http.Request) {
	uuidStr := r.URL.Query().Get("uuid")
	if uuidStr == "" {
		report(w, s_u, "缺少团队UUID参数。")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeamByUUID(uuidStr)
	if err != nil {
		util.Debug("cannot get team by uuid %v", err)
		report(w, s_u, "团队不存在。")
		return
	}

	// 检查权限：只有团队成员可以查看
	// isMember, err := team.IsMember(s_u.Id)
	// if err != nil || !isMember {
	// 	report(w, s_u, "您没有权限查看此团队的技能列表。")
	// 	return
	// }

	// 获取SkillTeamBean
	skillTeamBean, err := fetchSkillTeamBean(team, r.Context())
	if err != nil {
		util.Debug("cannot fetch skill team bean: %v", err)
		report(w, s_u, "获取团队技能列表失败，请重试。")
		return
	}

	// 创建技能与团队技能的映射
	skillTeamMap := make(map[int]dao.SkillTeam)
	for _, skillTeam := range skillTeamBean.SkillTeams {
		skillTeamMap[skillTeam.SkillId] = skillTeam
	}

	// 创建包含技能和团队信息的结构
	type SkillWithTeamInfo struct {
		Skill     dao.Skill
		SkillTeam dao.SkillTeam
	}

	// 按技能类型分组
	var hardSkills, softSkills []SkillWithTeamInfo
	for _, skill := range skillTeamBean.Skills {
		if skillTeam, exists := skillTeamMap[skill.Id]; exists {
			skillWithInfo := SkillWithTeamInfo{
				Skill:     skill,
				SkillTeam: skillTeam,
			}
			switch skill.Category {
			case dao.GeneralHardSkill:
				hardSkills = append(hardSkills, skillWithInfo)
			case dao.GeneralSoftSkill:
				softSkills = append(softSkills, skillWithInfo)
			}
		}
	}

	var SkillDetailTemplateData struct {
		SessUser       dao.User
		Team           dao.Team
		SkillTeamBean  dao.SkillTeamBean
		HardSkills     []SkillWithTeamInfo
		SoftSkills     []SkillWithTeamInfo
		HardSkillCount int
		SoftSkillCount int
	}

	SkillDetailTemplateData.SessUser = s_u
	SkillDetailTemplateData.Team = team
	SkillDetailTemplateData.SkillTeamBean = skillTeamBean
	SkillDetailTemplateData.HardSkills = hardSkills
	SkillDetailTemplateData.SoftSkills = softSkills
	SkillDetailTemplateData.HardSkillCount = len(hardSkills)
	SkillDetailTemplateData.SoftSkillCount = len(softSkills)

	generateHTML(w, &SkillDetailTemplateData, "layout", "navbar.private", "skills.team_list", "component_team_skill_bean")
}

// Handler /v1/skill_team/edit
func HandleSkillTeamEdit(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("cannot get s_u from session %v", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	switch r.Method {
	case http.MethodGet:
		SkillTeamEditGet(s_u, w, r)
	case http.MethodPost:
		SkillTeamEditPost(s_u, w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/skill_team/edit?id=123
func SkillTeamEditGet(s_u dao.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		report(w, s_u, "缺少技能记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, s_u, "无效的技能记录ID。")
		return
	}

	// 获取团队技能记录
	var skillTeam dao.SkillTeam
	if err := skillTeam.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill team by id %v", err)
		report(w, s_u, "技能记录不存在。")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeam(skillTeam.TeamId)
	if err != nil {
		report(w, s_u, "团队信息获取失败。")
		return
	}

	// 权限检查：只有团队核心成员可以编辑
	isCoreMember, err := team.IsCoreMember(s_u.Id)
	if err != nil || !isCoreMember {
		report(w, s_u, "您没有权限编辑此技能记录。")
		return
	}

	// 获取技能信息
	var skill dao.Skill
	skill.Id = skillTeam.SkillId
	if err := skill.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get skill by id %v", err)
		report(w, s_u, "技能信息获取失败。")
		return
	}

	var editData struct {
		SessUser  dao.User
		Team      dao.Team
		SkillTeam dao.SkillTeam
		Skill     dao.Skill
		ReturnURL string
	}
	editData.SessUser = s_u
	editData.Team = team
	editData.SkillTeam = skillTeam
	editData.Skill = skill
	editData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &editData, "layout", "navbar.private", "skill_team.edit")
}

// POST /v1/skill_team/edit
func SkillTeamEditPost(s_u dao.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.PostFormValue("id")
	if idStr == "" {
		report(w, s_u, "缺少技能记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, s_u, "无效的技能记录ID。")
		return
	}

	// 获取原始团队技能记录
	var skillTeam dao.SkillTeam
	if err := skillTeam.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill team by id %v", err)
		report(w, s_u, "技能记录不存在。")
		return
	}

	// 获取团队信息并检查权限
	team, err := dao.GetTeam(skillTeam.TeamId)
	if err != nil {
		report(w, s_u, "团队信息获取失败。")
		return
	}

	isCoreMember, err := team.IsCoreMember(s_u.Id)
	if err != nil || !isCoreMember {
		report(w, s_u, "您没有权限编辑此技能记录。")
		return
	}

	// 解析表单数据
	level, _ := strconv.Atoi(r.PostFormValue("level"))
	if level < 1 || level > 9 {
		report(w, s_u, "技能等级必须在1-9之间。")
		return
	}

	status, _ := strconv.Atoi(r.PostFormValue("status"))
	if status < 0 || status > 3 {
		report(w, s_u, "技能状态值无效。")
		return
	}

	// 更新团队技能记录
	skillTeam.Level = level
	skillTeam.Status = dao.SkillTeamStatus(status)

	if err := skillTeam.Update(); err != nil {
		util.Debug("cannot update skill team %v", err)
		report(w, s_u, "更新技能记录失败，请重试。")
		return
	}

	// 获取返回URL
	returnURL := r.PostFormValue("return_url")
	if returnURL == "" {
		returnURL = fmt.Sprintf("/v1/skills/team_list?uuid=%s", team.Uuid)
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}
