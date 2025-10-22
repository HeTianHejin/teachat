package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/magic/new
func HandleNewMagic(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		MagicNewGet(w, r)
	case http.MethodPost:
		MagicNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handler /v1/magic/detail
func HandleMagicDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	MagicDetailGet(w, r)
}

// Handler /v1/magic/list
func HandleMagicList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	MagicListGet(w, r)
}

// GET /v1/magic/new
func MagicNewGet(w http.ResponseWriter, r *http.Request) {
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

	// 获取用户所在的团队
	userTeams, err := user.SurvivalTeams()
	if err != nil {
		util.Debug("cannot get user teams", err)
		userTeams = []data.Team{} // 如果获取失败，使用空列表
	}

	var magicData struct {
		SessUser  data.User
		UserTeams []data.Team
		ReturnURL string
	}
	magicData.SessUser = user
	magicData.UserTeams = userTeams
	magicData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &magicData, "layout", "navbar.private", "magic.new")
}

// POST /v1/magic/new
func MagicNewPost(w http.ResponseWriter, r *http.Request) {
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

	if name == "" {
		report(w, r, "法力名称不能为空。")
		return
	}
	if description == "" {
		report(w, r, "法力描述不能为空。")
		return
	}

	// 解析表单数据
	category, _ := strconv.Atoi(r.PostFormValue("category"))
	if category < 1 || category > 2 {
		category = 1 // 默认理性
	}

	intelligenceLevel, _ := strconv.Atoi(r.PostFormValue("intelligence_level"))
	if intelligenceLevel < 1 || intelligenceLevel > 5 {
		intelligenceLevel = 3 // 默认中等
	}

	difficultyLevel, _ := strconv.Atoi(r.PostFormValue("difficulty_level"))
	if difficultyLevel < 1 || difficultyLevel > 5 {
		difficultyLevel = 3 // 默认中等
	}

	level, _ := strconv.Atoi(r.PostFormValue("level"))
	if level < 1 || level > 5 {
		level = 1 // 默认入门
	}

	magic := data.Magic{
		UserId:            user.Id,
		Name:              name,
		Nickname:          strings.TrimSpace(r.PostFormValue("nickname")),
		Description:       description,
		IntelligenceLevel: data.IntelligenceLevel(intelligenceLevel),
		DifficultyLevel:   data.DifficultyLevel(difficultyLevel),
		Category:          data.MagicCategory(category),
		Level:             level,
	}

	if err := magic.Create(r.Context()); err != nil {
		util.Debug("Cannot create magic", err)
		report(w, r, "创建法力记录失败，请重试。")
		return
	}

	// 检查是否添加到个人法力列表
	addToMyMagics := r.PostFormValue("add_to_my_magics") == "1"
	if addToMyMagics {
		magicUser := data.MagicUser{
			MagicId: magic.Id,
			UserId:  user.Id,
			Level:   1,           // 默认等级1
			Status:  data.Clear, // 默认清醒状态
		}
		if err := magicUser.Create(r.Context()); err != nil {
			util.Debug("cannot create magic user record", err)
			// 不阻止流程，仅记录错误
		}
	}

	// 检查是否添加到团队法力列表
	teamMagicIds := r.Form["add_to_team_magics"]
	for _, teamIdStr := range teamMagicIds {
		teamId, err := strconv.Atoi(teamIdStr)
		if err != nil || teamId <= 0 {
			continue
		}
		// 验证用户是否为该团队成员
		team, err := data.GetTeam(teamId)
		if err != nil {
			continue
		}
		isMember, err := team.IsMember(user.Id)
		if err != nil || !isMember {
			continue
		}
		// 创建团队法力记录
		magicTeam := data.MagicTeam{
			MagicId: magic.Id,
			TeamId:  teamId,
			Level:   1,                              // 默认等级1
			Status:  data.ClearMagicTeamStatus, // 默认清晰状态
		}
		if err := magicTeam.Create(r.Context()); err != nil {
			util.Debug("cannot create magic team record", err)
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
		// 如果有返回URL，添加新创建的法力ID参数
		if returnURL != "/v1/" {
			if strings.Contains(returnURL, "?") {
				returnURL += fmt.Sprintf("&new_magic_id=%d", magic.Id)
			} else {
				returnURL += fmt.Sprintf("?new_magic_id=%d", magic.Id)
			}
		}
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}

// GET /v1/magic/detail?id=123
func MagicDetailGet(w http.ResponseWriter, r *http.Request) {
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

	idStr := r.URL.Query().Get("id")
	uuidStr := r.URL.Query().Get("uuid")
	if idStr == "" && uuidStr == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var magic data.Magic
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
		magic.Id = id
	} else {
		magic.Uuid = uuidStr
	}

	if err := magic.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get magic by id/uuid", magic.Id, magic.Uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 创建MagicBean结构
	type MagicBean struct {
		Magic data.Magic
	}

	magicBean := MagicBean{
		Magic: magic,
	}

	var magicData struct {
		SessUser   data.User
		IsVerifier bool
		IsAdmin    bool
		IsMaster   bool
		IsInvited  bool
		MagicBean  MagicBean
	}
	magicData.SessUser = user
	magicData.MagicBean = magicBean

	generateHTML(w, &magicData, "layout", "navbar.private", "magic.detail")
}

// GET /v1/magic/list
func MagicListGet(w http.ResponseWriter, r *http.Request) {
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

	// 获取所有法力
	magics, err := data.GetAllMagics(r.Context())
	if err != nil {
		util.Debug("Cannot get magics", err)
		report(w, r, "获取法力列表失败，请重试。")
		return
	}

	// 创建法力Bean列表
	var magicBeans []struct {
		Magic data.Magic
	}
	for _, magic := range magics {
		magicBeans = append(magicBeans, struct {
			Magic data.Magic
		}{
			Magic: magic,
		})
	}

	var magicData struct {
		SessUser data.User
		Magics   []struct {
			Magic data.Magic
		}
	}
	magicData.SessUser = user
	magicData.Magics = magicBeans

	generateHTML(w, &magicData, "layout", "navbar.private", "magic.list")
}

// Handler /v1/magics/user_list
func HandleMagicsUserList(w http.ResponseWriter, r *http.Request) {
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
	MagicsUserListGet(s_u, w, r)
}

// Handler /v1/magic_user/edit
func HandleMagicUserEdit(w http.ResponseWriter, r *http.Request) {
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
		MagicUserEditGet(s_u, w, r)
	case http.MethodPost:
		MagicUserEditPost(s_u, w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/magics/user_list
func MagicsUserListGet(user data.User, w http.ResponseWriter, r *http.Request) {
	// 确保用户拥有默认法力
	if err := data.EnsureDefaultMagics(user.Id, r.Context()); err != nil {
		util.Debug("cannot ensure default magics for user:", user.Id, err)
	}

	// 获取MagicUserBean
	magicUserBean, err := fetchMagicUserBean(user, r.Context())
	if err != nil {
		util.Debug("cannot fetch magic user bean:", user.Id, err)
		report(w, r, "获取茶友法力列表失败，请重试。")
		return
	}

	// 创建法力与用户法力的映射
	magicUserMap := make(map[int]data.MagicUser)
	for _, magicUser := range magicUserBean.MagicUsers {
		magicUserMap[magicUser.MagicId] = magicUser
	}

	// 创建包含法力和用户信息的结构
	type MagicWithUserInfo struct {
		Magic     data.Magic
		MagicUser data.MagicUser
	}

	// 按法力类型分组
	var rationalMagics, sensualMagics []MagicWithUserInfo
	for _, magic := range magicUserBean.Magics {
		if magicUser, exists := magicUserMap[magic.Id]; exists {
			magicWithInfo := MagicWithUserInfo{
				Magic:     magic,
				MagicUser: magicUser,
			}
			switch magic.Category {
			case data.Rational:
				rationalMagics = append(rationalMagics, magicWithInfo)
			case data.Sensual:
				sensualMagics = append(sensualMagics, magicWithInfo)
			}
		}
	}

	var MagicDetailTemplateData struct {
		SessUser           data.User
		MagicUserBean      data.MagicUserBean
		RationalMagics     []MagicWithUserInfo
		SensualMagics      []MagicWithUserInfo
		RationalMagicCount int
		SensualMagicCount  int
	}

	MagicDetailTemplateData.SessUser = user
	MagicDetailTemplateData.MagicUserBean = magicUserBean
	MagicDetailTemplateData.RationalMagics = rationalMagics
	MagicDetailTemplateData.SensualMagics = sensualMagics
	MagicDetailTemplateData.RationalMagicCount = len(rationalMagics)
	MagicDetailTemplateData.SensualMagicCount = len(sensualMagics)

	generateHTML(w, &MagicDetailTemplateData, "layout", "navbar.private", "magics.user_list", "component_user_magic_bean")
}

// GET /v1/magic_user/edit?id=123
func MagicUserEditGet(user data.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		report(w, r, "缺少法力记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, r, "无效的法力记录ID。")
		return
	}

	// 获取法力用户记录
	var magicUser data.MagicUser
	if err := magicUser.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get magic user by id", id, err)
		report(w, r, "法力记录不存在。")
		return
	}

	// 权限检查：只有同一家庭的parents成员可以编辑
	if magicUser.UserId != user.Id {
		// 获取目标用户的默认家庭
		targetUser, err := data.GetUser(magicUser.UserId)
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
			report(w, r, "您没有权限编辑此法力记录。")
			return
		}
	}

	// 获取法力信息
	var magic data.Magic
	magic.Id = magicUser.MagicId
	if err := magic.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get magic by id", magicUser.MagicId, err)
		report(w, r, "法力信息获取失败。")
		return
	}

	var editData struct {
		SessUser  data.User
		MagicUser data.MagicUser
		Magic     data.Magic
		ReturnURL string
	}
	editData.SessUser = user
	editData.MagicUser = magicUser
	editData.Magic = magic
	editData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &editData, "layout", "navbar.private", "magic_user.edit")
}

// POST /v1/magic_user/edit
func MagicUserEditPost(user data.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.PostFormValue("id")
	if idStr == "" {
		report(w, r, "缺少法力记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, r, "无效的法力记录ID。")
		return
	}

	// 获取原始法力用户记录
	var magicUser data.MagicUser
	if err := magicUser.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get magic user by id", id, err)
		report(w, r, "法力记录不存在。")
		return
	}

	// 权限检查
	if magicUser.UserId != user.Id {
		targetUser, err := data.GetUser(magicUser.UserId)
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
			report(w, r, "您没有权限编辑此法力记录。")
			return
		}
	}

	// 解析表单数据
	level, _ := strconv.Atoi(r.PostFormValue("level"))
	if level < 1 || level > 9 {
		report(w, r, "法力等级必须在1-9之间。")
		return
	}

	status, _ := strconv.Atoi(r.PostFormValue("status"))
	if status < 0 || status > 3 {
		report(w, r, "法力状态值无效。")
		return
	}

	// 更新法力用户记录
	magicUser.Level = level
	magicUser.Status = data.MagicUserStatus(status)

	if err := magicUser.Update(); err != nil {
		util.Debug("cannot update magic user", err)
		report(w, r, "更新法力记录失败，请重试。")
		return
	}

	// 获取返回URL
	returnURL := r.PostFormValue("return_url")
	if returnURL == "" {
		returnURL = "/v1/magics/user_list"
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}

// Handler /v1/magics/team_list
func HandleMagicsTeamList(w http.ResponseWriter, r *http.Request) {
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
	MagicsTeamListGet(s_u, w, r)
}

// GET /v1/magics/team_list?uuid=xxx
func MagicsTeamListGet(user data.User, w http.ResponseWriter, r *http.Request) {
	uuidStr := r.URL.Query().Get("uuid")
	if uuidStr == "" {
		report(w, r, "缺少团队UUID参数。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeamByUUID(uuidStr)
	if err != nil {
		util.Debug("cannot get team by uuid", uuidStr, err)
		report(w, r, "团队不存在。")
		return
	}

	// 检查权限：只有团队成员可以查看
	// isMember, err := team.IsMember(user.Id)
	// if err != nil || !isMember {
	// 	report(w, r, "您没有权限查看此团队的法力列表。")
	// 	return
	// }

	// 获取MagicTeamBean
	magicTeamBean, err := fetchMagicTeamBean(team, r.Context())
	if err != nil {
		util.Debug("cannot fetch magic team bean:", team.Id, err)
		report(w, r, "获取团队法力列表失败，请重试。")
		return
	}

	// 创建法力与团队法力的映射
	magicTeamMap := make(map[int]data.MagicTeam)
	for _, magicTeam := range magicTeamBean.MagicTeams {
		magicTeamMap[magicTeam.MagicId] = magicTeam
	}

	// 创建包含法力和团队信息的结构
	type MagicWithTeamInfo struct {
		Magic     data.Magic
		MagicTeam data.MagicTeam
	}

	// 按法力类型分组
	var rationalMagics, sensualMagics []MagicWithTeamInfo
	for _, magic := range magicTeamBean.Magics {
		if magicTeam, exists := magicTeamMap[magic.Id]; exists {
			magicWithInfo := MagicWithTeamInfo{
				Magic:     magic,
				MagicTeam: magicTeam,
			}
			switch magic.Category {
			case data.Rational:
				rationalMagics = append(rationalMagics, magicWithInfo)
			case data.Sensual:
				sensualMagics = append(sensualMagics, magicWithInfo)
			}
		}
	}

	var MagicDetailTemplateData struct {
		SessUser           data.User
		Team               data.Team
		MagicTeamBean      data.MagicTeamBean
		RationalMagics     []MagicWithTeamInfo
		SensualMagics      []MagicWithTeamInfo
		RationalMagicCount int
		SensualMagicCount  int
	}

	MagicDetailTemplateData.SessUser = user
	MagicDetailTemplateData.Team = team
	MagicDetailTemplateData.MagicTeamBean = magicTeamBean
	MagicDetailTemplateData.RationalMagics = rationalMagics
	MagicDetailTemplateData.SensualMagics = sensualMagics
	MagicDetailTemplateData.RationalMagicCount = len(rationalMagics)
	MagicDetailTemplateData.SensualMagicCount = len(sensualMagics)

	generateHTML(w, &MagicDetailTemplateData, "layout", "navbar.private", "magics.team_list", "component_team_magic_bean")
}

// Handler /v1/magic_team/edit
func HandleMagicTeamEdit(w http.ResponseWriter, r *http.Request) {
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
		MagicTeamEditGet(s_u, w, r)
	case http.MethodPost:
		MagicTeamEditPost(s_u, w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/magic_team/edit?id=123
func MagicTeamEditGet(user data.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		report(w, r, "缺少法力记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, r, "无效的法力记录ID。")
		return
	}

	// 获取团队法力记录
	var magicTeam data.MagicTeam
	if err := magicTeam.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get magic team by id", id, err)
		report(w, r, "法力记录不存在。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeam(magicTeam.TeamId)
	if err != nil {
		report(w, r, "团队信息获取失败。")
		return
	}

	// 权限检查：只有团队核心成员可以编辑
	isCoreMember, err := team.IsCoreMember(user.Id)
	if err != nil || !isCoreMember {
		report(w, r, "您没有权限编辑此法力记录。")
		return
	}

	// 获取法力信息
	var magic data.Magic
	magic.Id = magicTeam.MagicId
	if err := magic.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("cannot get magic by id", magicTeam.MagicId, err)
		report(w, r, "法力信息获取失败。")
		return
	}

	var editData struct {
		SessUser  data.User
		Team      data.Team
		MagicTeam data.MagicTeam
		Magic     data.Magic
		ReturnURL string
	}
	editData.SessUser = user
	editData.Team = team
	editData.MagicTeam = magicTeam
	editData.Magic = magic
	editData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &editData, "layout", "navbar.private", "magic_team.edit")
}

// POST /v1/magic_team/edit
func MagicTeamEditPost(user data.User, w http.ResponseWriter, r *http.Request) {
	idStr := r.PostFormValue("id")
	if idStr == "" {
		report(w, r, "缺少法力记录ID参数。")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, r, "无效的法力记录ID。")
		return
	}

	// 获取原始团队法力记录
	var magicTeam data.MagicTeam
	if err := magicTeam.GetById(id, r.Context()); err != nil {
		util.Debug("cannot get magic team by id", id, err)
		report(w, r, "法力记录不存在。")
		return
	}

	// 获取团队信息并检查权限
	team, err := data.GetTeam(magicTeam.TeamId)
	if err != nil {
		report(w, r, "团队信息获取失败。")
		return
	}

	isCoreMember, err := team.IsCoreMember(user.Id)
	if err != nil || !isCoreMember {
		report(w, r, "您没有权限编辑此法力记录。")
		return
	}

	// 解析表单数据
	level, _ := strconv.Atoi(r.PostFormValue("level"))
	if level < 1 || level > 9 {
		report(w, r, "法力段位必须在1-9之间。")
		return
	}

	status, _ := strconv.Atoi(r.PostFormValue("status"))
	if status < 0 || status > 3 {
		report(w, r, "法力状态值无效。")
		return
	}

	// 更新团队法力记录
	magicTeam.Level = level
	magicTeam.Status = data.MagicTeamStatus(status)

	if err := magicTeam.Update(); err != nil {
		util.Debug("cannot update magic team", err)
		report(w, r, "更新法力记录失败，请重试。")
		return
	}

	// 获取返回URL
	returnURL := r.PostFormValue("return_url")
	if returnURL == "" {
		returnURL = fmt.Sprintf("/v1/magics/team_list?uuid=%s", team.Uuid)
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}
