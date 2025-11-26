package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
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
		util.Debug("cannot get s_u from session", err)
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
		util.Debug("cannot get s_u from session", err)
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
		util.Debug("cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	SkillsUserListGet(s_u, w, r)
}

// GET /v1/skill/new
func SkillNewGet(s_u data.User, w http.ResponseWriter, r *http.Request) {
	// 获取用户所在的团队
	userTeams, err := s_u.SurvivalTeams()
	if err != nil {
		util.Debug("cannot get s_u teams", err)
		userTeams = []data.Team{} // 如果获取失败，使用空列表
	}

	var skillData struct {
		SessUser  data.User
		UserTeams []data.Team
		ReturnURL string
	}
	skillData.SessUser = s_u
	skillData.UserTeams = userTeams
	skillData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &skillData, "layout", "navbar.private", "skill.new")
}

// POST /v1/skill/new
func SkillNewPost(s_u data.User, w http.ResponseWriter, r *http.Request) {

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

	skill := data.Skill{
		UserId:          s_u.Id,
		Name:            name,
		Nickname:        strings.TrimSpace(r.PostFormValue("nickname")),
		Description:     description,
		StrengthLevel:   data.StrengthLevel(strengthLevel),
		DifficultyLevel: data.DifficultyLevel(difficultyLevel),
		Category:        data.SkillCategory(category),
		Level:           level,
	}

	if err := skill.Create(r.Context()); err != nil {
		util.Debug("cannot create skill", err)
		report(w, s_u, "创建技能记录失败，请重试。")
		return
	}

	// 检查是否添加到个人技能列表
	addToMySkills := r.PostFormValue("add_to_my_skills") == "1"
	if addToMySkills {
		skillUser := data.SkillUser{
			SkillId: skill.Id,
			UserId:  s_u.Id,
			Level:   1,                          // 默认等级1
			Status:  data.NormalSkillUserStatus, // 默认中能状态
		}
		if err := skillUser.Create(r.Context()); err != nil {
			util.Debug("cannot create skill s_u record", err)
			// 不阻止流程，仅记录错误
		}
	}

	// 检查是否添加到团队技能列表
	teamSkillIds := r.Form["add_to_team_skills"]
	for _, teamIdStr := range teamSkillIds {
		teamId, err := strconv.Atoi(teamIdStr)
		if err != nil || teamId <= 0 {
			continue
		}
		// 验证用户是否为该团队成员
		team, err := data.GetTeam(teamId)
		if err != nil {
			continue
		}
		isMember, err := team.IsMember(s_u.Id)
		if err != nil || !isMember {
			continue
		}
		// 创建团队技能记录
		skillTeam := data.SkillTeam{
			SkillId: skill.Id,
			TeamId:  teamId,
			Level:   1,                          // 默认等级1
			Status:  data.NormalSkillTeamStatus, // 默认正常状态
		}
		if err := skillTeam.Create(r.Context()); err != nil {
			util.Debug("cannot create skill team record", err)
			// 不阻止流程，仅记录错误
		}
	}

	// 获取返回URL参数
	returnURL := r.PostFormValue("return_url")
	if returnURL == "" {
		returnURL = r.URL.Query().Get("return_url")
	}
	if returnURL == "" {
		returnURL = "/v1/"
	} else {
		// 如果有返回URL，添加新创建的技能ID参数
		if returnURL != "/v1/" {
			if strings.Contains(returnURL, "?") {
				returnURL += fmt.Sprintf("&new_skill_id=%d", skill.Id)
			} else {
				returnURL += fmt.Sprintf("?new_skill_id=%d", skill.Id)
			}
		}
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}

// GET /v1/skill/detail?id=123
func SkillDetailGet(s_u data.User, w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")
	uuidStr := r.URL.Query().Get("uuid")
	if idStr == "" && uuidStr == "" {
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var skill data.Skill
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
		util.Debug("cannot get skill by id/uuid", skill.Id, skill.Uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var skillData struct {
		SessUser data.User
		Skill    data.Skill
	}
	skillData.SessUser = s_u
	skillData.Skill = skill

	generateHTML(w, &skillData, "layout", "navbar.private", "skill.detail")
}

// GET /v1/skills/user_list
func SkillsUserListGet(s_u data.User, w http.ResponseWriter, r *http.Request) {

	// 确保用户拥有默认技能
	if err := data.EnsureDefaultSkills(s_u.Id, r.Context()); err != nil {
		util.Debug("cannot ensure default skills for s_u:", s_u.Id, err)
	}

	// 获取SkillUserBean
	skillUserBean, err := fetchSkillUserBean(s_u, r.Context())
	if err != nil {
		util.Debug("cannot fetch skill s_u bean:", s_u.Id, err)
		report(w, s_u, "获取茶友技能列表失败，请重试。")
		return
	}

	// 创建技能与用户技能的映射
	skillUserMap := make(map[int]data.SkillUser)
	for _, skillUser := range skillUserBean.SkillUsers {
		skillUserMap[skillUser.SkillId] = skillUser
	}

	// 创建包含技能和用户信息的结构
	type SkillWithUserInfo struct {
		Skill     data.Skill
		SkillUser data.SkillUser
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
			case data.GeneralHardSkill:
				hardSkills = append(hardSkills, skillWithInfo)
			case data.GeneralSoftSkill:
				softSkills = append(softSkills, skillWithInfo)
			}
		}
	}

	var SkillDetailTemplateData struct {
		SessUser       data.User
		SkillUserBean  data.SkillUserBean
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
		util.Debug("cannot get s_u from session", err)
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
func SkillUserEditGet(s_u data.User, w http.ResponseWriter, r *http.Request) {
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
	var skillUser data.SkillUser
	if err := skillUser.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill s_u by id", id, err)
		report(w, s_u, "技能记录不存在。")
		return
	}

	// 权限检查：只有同一家庭的parents成员可以编辑
	if skillUser.UserId != s_u.Id {
		// 获取目标用户的默认家庭
		targetUser, err := data.GetUser(skillUser.UserId)
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
	var skill data.Skill
	skill.Id = skillUser.SkillId
	if err := skill.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get skill by id", skillUser.SkillId, err)
		report(w, s_u, "技能信息获取失败。")
		return
	}

	var editData struct {
		SessUser  data.User
		SkillUser data.SkillUser
		Skill     data.Skill
		ReturnURL string
	}
	editData.SessUser = s_u
	editData.SkillUser = skillUser
	editData.Skill = skill
	editData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &editData, "layout", "navbar.private", "skill_user.edit")
}

// POST /v1/skill_user/edit
func SkillUserEditPost(s_u data.User, w http.ResponseWriter, r *http.Request) {
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
	var skillUser data.SkillUser
	if err := skillUser.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill s_u by id", id, err)
		report(w, s_u, "技能记录不存在。")
		return
	}

	// 权限检查
	if skillUser.UserId != s_u.Id {
		targetUser, err := data.GetUser(skillUser.UserId)
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
	skillUser.Status = data.SkillUserStatus(status)

	if err := skillUser.Update(); err != nil {
		util.Debug("cannot update skill s_u", err)
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
		util.Debug("cannot get s_u from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	SkillsTeamListGet(s_u, w, r)
}

// GET /v1/skills/team_list?uuid=xxx
func SkillsTeamListGet(s_u data.User, w http.ResponseWriter, r *http.Request) {
	uuidStr := r.URL.Query().Get("uuid")
	if uuidStr == "" {
		report(w, s_u, "缺少团队UUID参数。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeamByUUID(uuidStr)
	if err != nil {
		util.Debug("cannot get team by uuid", uuidStr, err)
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
		util.Debug("cannot fetch skill team bean:", team.Id, err)
		report(w, s_u, "获取团队技能列表失败，请重试。")
		return
	}

	// 创建技能与团队技能的映射
	skillTeamMap := make(map[int]data.SkillTeam)
	for _, skillTeam := range skillTeamBean.SkillTeams {
		skillTeamMap[skillTeam.SkillId] = skillTeam
	}

	// 创建包含技能和团队信息的结构
	type SkillWithTeamInfo struct {
		Skill     data.Skill
		SkillTeam data.SkillTeam
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
			case data.GeneralHardSkill:
				hardSkills = append(hardSkills, skillWithInfo)
			case data.GeneralSoftSkill:
				softSkills = append(softSkills, skillWithInfo)
			}
		}
	}

	var SkillDetailTemplateData struct {
		SessUser       data.User
		Team           data.Team
		SkillTeamBean  data.SkillTeamBean
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
		util.Debug("cannot get s_u from session", err)
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
func SkillTeamEditGet(s_u data.User, w http.ResponseWriter, r *http.Request) {
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
	var skillTeam data.SkillTeam
	if err := skillTeam.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill team by id", id, err)
		report(w, s_u, "技能记录不存在。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeam(skillTeam.TeamId)
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
	var skill data.Skill
	skill.Id = skillTeam.SkillId
	if err := skill.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get skill by id", skillTeam.SkillId, err)
		report(w, s_u, "技能信息获取失败。")
		return
	}

	var editData struct {
		SessUser  data.User
		Team      data.Team
		SkillTeam data.SkillTeam
		Skill     data.Skill
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
func SkillTeamEditPost(s_u data.User, w http.ResponseWriter, r *http.Request) {
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
	var skillTeam data.SkillTeam
	if err := skillTeam.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill team by id", id, err)
		report(w, s_u, "技能记录不存在。")
		return
	}

	// 获取团队信息并检查权限
	team, err := data.GetTeam(skillTeam.TeamId)
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
	skillTeam.Status = data.SkillTeamStatus(status)

	if err := skillTeam.Update(); err != nil {
		util.Debug("cannot update skill team", err)
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
