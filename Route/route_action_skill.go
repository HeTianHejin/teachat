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
		util.Debug("cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
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
		util.Debug("cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
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
		util.Debug("cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	SkillsUserListGet(s_u, w, r)
}

// GET /v1/skill/new
func SkillNewGet(user data.User, w http.ResponseWriter, r *http.Request) {

	var skillData struct {
		SessUser  data.User
		ReturnURL string
	}
	skillData.SessUser = user
	skillData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &skillData, "layout", "navbar.private", "skill.new")
}

// POST /v1/skill/new
func SkillNewPost(user data.User, w http.ResponseWriter, r *http.Request) {

	// 验证必填字段
	name := strings.TrimSpace(r.PostFormValue("name"))
	description := strings.TrimSpace(r.PostFormValue("description"))

	if name == "" {
		report(w, r, "技能名称不能为空。")
		return
	}
	if description == "" {
		report(w, r, "技能描述不能为空。")
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
		UserId:          user.Id,
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
		report(w, r, "创建技能记录失败，请重试。")
		return
	}

	// 检查是否添加到个人技能列表
	addToMySkills := r.PostFormValue("add_to_my_skills") == "1"
	if addToMySkills {
		skillUser := data.SkillUser{
			SkillId: skill.Id,
			UserId:  user.Id,
			Level:   1,                          // 默认等级1
			Status:  data.NormalSkillUserStatus, // 默认中能状态
		}
		if err := skillUser.Create(r.Context()); err != nil {
			util.Debug("cannot create skill user record", err)
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
func SkillDetailGet(user data.User, w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")
	uuidStr := r.URL.Query().Get("uuid")
	if idStr == "" && uuidStr == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var skill data.Skill
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
		skill.Id = id
	} else {
		skill.Uuid = uuidStr
	}

	if err := skill.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get skill by id/uuid", skill.Id, skill.Uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var skillData struct {
		SessUser data.User
		Skill    data.Skill
	}
	skillData.SessUser = user
	skillData.Skill = skill

	generateHTML(w, &skillData, "layout", "navbar.private", "skill.detail")
}

// GET /v1/skills/user_list
func SkillsUserListGet(user data.User, w http.ResponseWriter, r *http.Request) {

	// 确保用户拥有默认技能
	if err := data.EnsureDefaultSkills(user.Id, r.Context()); err != nil {
		util.Debug("cannot ensure default skills for user:", user.Id, err)
	}

	// 获取SkillUserBean
	skillUserBean, err := fetchSkillUserBean(user, r.Context())
	if err != nil {
		util.Debug("cannot fetch skill user bean:", user.Id, err)
		report(w, r, "获取茶友技能列表失败，请重试。")
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

	SkillDetailTemplateData.SessUser = user
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
		util.Debug("cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
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
func SkillUserEditGet(user data.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		report(w, r, "缺少技能记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, r, "无效的技能记录ID。")
		return
	}

	// 获取技能用户记录
	var skillUser data.SkillUser
	if err := skillUser.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill user by id", id, err)
		report(w, r, "技能记录不存在。")
		return
	}

	// 权限检查：只有同一家庭的parents成员可以编辑
	if skillUser.UserId != user.Id {
		// 获取目标用户的默认家庭
		targetUser, err := data.GetUser(skillUser.UserId)
		if err != nil {
			report(w, r, "权限验证失败。")
			return
		}

		targetFamily, err := targetUser.GetLastDefaultFamily()
		if err != nil {
			report(w, r, "权限验证失败。")
			return
		}

		// 检查当前用户是否为该家庭的parent成员
		isParent, err := targetFamily.IsParentMember(user.Id)
		if err != nil || !isParent {
			report(w, r, "您没有权限编辑此技能记录。")
			return
		}
	}

	// 获取技能信息
	var skill data.Skill
	skill.Id = skillUser.SkillId
	if err := skill.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get skill by id", skillUser.SkillId, err)
		report(w, r, "技能信息获取失败。")
		return
	}

	var editData struct {
		SessUser  data.User
		SkillUser data.SkillUser
		Skill     data.Skill
		ReturnURL string
	}
	editData.SessUser = user
	editData.SkillUser = skillUser
	editData.Skill = skill
	editData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &editData, "layout", "navbar.private", "skill_user.edit")
}

// POST /v1/skill_user/edit
func SkillUserEditPost(user data.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.PostFormValue("id")
	if idStr == "" {
		report(w, r, "缺少技能记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, r, "无效的技能记录ID。")
		return
	}

	// 获取原始技能用户记录
	var skillUser data.SkillUser
	if err := skillUser.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get skill user by id", id, err)
		report(w, r, "技能记录不存在。")
		return
	}

	// 权限检查
	if skillUser.UserId != user.Id {
		targetUser, err := data.GetUser(skillUser.UserId)
		if err != nil {
			report(w, r, "权限验证失败。")
			return
		}

		targetFamily, err := targetUser.GetLastDefaultFamily()
		if err != nil {
			report(w, r, "权限验证失败。")
			return
		}

		isParent, err := targetFamily.IsParentMember(user.Id)
		if err != nil || !isParent {
			report(w, r, "您没有权限编辑此技能记录。")
			return
		}
	}

	// 解析表单数据
	level, _ := strconv.Atoi(r.PostFormValue("level"))
	if level < 1 || level > 9 {
		report(w, r, "技能等级必须在1-9之间。")
		return
	}

	status, _ := strconv.Atoi(r.PostFormValue("status"))
	if status < 0 || status > 3 {
		report(w, r, "技能状态值无效。")
		return
	}

	// 更新技能用户记录
	skillUser.Level = level
	skillUser.Status = data.SkillUserStatus(status)

	if err := skillUser.Update(); err != nil {
		util.Debug("cannot update skill user", err)
		report(w, r, "更新技能记录失败，请重试。")
		return
	}

	// 获取返回URL
	returnURL := r.PostFormValue("return_url")
	if returnURL == "" {
		returnURL = "/v1/skills/user_list"
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}
