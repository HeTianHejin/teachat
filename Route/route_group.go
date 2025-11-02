package route

import (
	"fmt"
	"net/http"
	"strconv"

	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/group/new
// 显示创建新集团的表单页面
func NewGroupGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 如果有team_id参数，检查用户是否为该团队的CEO或创建人
	teamId := r.URL.Query().Get("team_id")
	if teamId != "" {
		team, err := data.GetTeamByUUID(teamId)
		if err != nil {
			util.Debug("Cannot get team by uuid", err)
			report(w, r, "你好，未能找到指定的团队。")
			return
		}

		// 检查权限：必须是创建人或CEO
		isCEO := false
		if sessUser.Id != team.FounderId {
			ceo, err := team.MemberCEO()
			if err == nil && ceo.UserId == sessUser.Id {
				isCEO = true
			}
			if !isCEO {
				report(w, r, "你好，只有团队创建人或CEO才能代表团队创建集团。")
				return
			}
		}
	}

	// 获取用户的团队列表，用于选择最高管理团队
	teams, err := sessUser.SurvivalTeams()
	if err != nil {
		util.Debug("Cannot get user teams", err)
	}

	var pageData struct {
		SessUser        data.User
		Teams           []data.Team
		PreSelectedTeam string // 预选的团队UUID
	}
	pageData.SessUser = sessUser
	pageData.Teams = teams
	pageData.PreSelectedTeam = teamId

	generateHTML(w, &pageData, "layout", "navbar.private", "group.new")
}

// POST /v1/group/create
// 创建新集团
func CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}

	// 读取表单数据
	name := r.PostFormValue("name")
	abbreviation := r.PostFormValue("abbreviation")
	mission := r.PostFormValue("mission")
	firstTeamIdStr := r.PostFormValue("first_team_id")
	classStr := r.PostFormValue("class")

	// 调试输出
	util.Debug("CreateGroupPost - class value:", classStr)

	// 转换团队ID
	firstTeamId, err := strconv.Atoi(firstTeamIdStr)
	if err != nil {
		report(w, r, "你好，请选择最高管理团队。")
		return
	}

	// 检查权限：用户必须是该团队的创建人或CEO
	team, err := data.GetTeam(firstTeamId)
	if err != nil {
		util.Debug("Cannot get team", err)
		report(w, r, "你好，未能找到指定的团队。")
		return
	}

	isCEO := false
	if sessUser.Id != team.FounderId {
		ceo, err := team.MemberCEO()
		if err == nil && ceo.UserId == sessUser.Id {
			isCEO = true
		}
		if !isCEO {
			report(w, r, "你好，只有团队创建人或CEO才能代表团队创建集团。")
			return
		}
	}

	// 验证集团名称长度
	nameLen := cnStrLen(name)
	if nameLen < 4 || nameLen > 24 {
		report(w, r, "你好，集团名称应在4-24个中文字符之间。")
		return
	}

	// 验证简称长度
	abbrLen := cnStrLen(abbreviation)
	if abbrLen < 2 || abbrLen > 8 {
		report(w, r, "你好，集团简称应在2-8个中文字符之间。")
		return
	}

	// 验证使命长度
	missionLen := cnStrLen(mission)
	if missionLen < int(util.Config.ThreadMinWord) || missionLen > int(util.Config.ThreadMaxWord) {
		report(w, r, "你好，集团使命字数不符合要求。")
		return
	}

	// 转换类型
	class, err := strconv.Atoi(classStr)
	if err != nil {
		util.Debug("Cannot convert class to int", err)
		report(w, r, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}

	// 检测class是否合规（草稿状态）
	switch class {
	case data.GroupClassOpenDraft, data.GroupClassCloseDraft:
		break
	default:
		report(w, r, "你好，茶博士摸摸头，竟然说集团类别不合适，未能创建新集团。")
		return
	}

	// 创建集团
	group := data.Group{
		Name:         name,
		Abbreviation: abbreviation,
		Mission:      mission,
		FounderId:    sessUser.Id,
		FirstTeamId:  firstTeamId,
		Class:        class,
		Logo:         "groupLogo",
		Tags:         r.PostFormValue("tags"),
	}

	if err := group.Create(); err != nil {
		util.Debug("Cannot create group", err)
		report(w, r, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}

	// 发送友邻蒙评请求
	if err = createAndSendAcceptMessage(group.Id, data.AcceptObjectTypeGroup, sessUser.Id); err != nil {
		util.Debug("Cannot create accept message for group", err)
		report(w, r, "你好，茶博士迷路了，未能发送蒙评请求消息。")
		return
	}

	// 提示用户新集团草稿保存成功
	text := ""
	if sessUser.Gender == data.User_Gender_Female {
		text = fmt.Sprintf("%s 女士，你好，登记 %s 集团草稿已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", sessUser.Name, group.Name)
	} else {
		text = fmt.Sprintf("%s 先生，你好，登记 %s 集团草稿已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", sessUser.Name, group.Name)
	}
	report(w, r, text)
}

// GET /v1/groups
// 显示所有集团列表
func GroupsGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var pageData struct {
		SessUser data.User
		Groups   []data.Group
	}
	pageData.SessUser = sessUser

	// TODO: 实现获取所有活跃集团的方法
	// pageData.Groups, err = data.GetActiveGroups()

	generateHTML(w, &pageData, "layout", "navbar.private", "groups.list")
}

// GET /v1/group/read?uuid=xxx
// 显示集团详情
func GroupReadGet(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取集团详情。")
		return
	}

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	uuid := r.FormValue("uuid")
	if uuid == "" {
		report(w, r, "你好，缺少集团标识。")
		return
	}

	group, err := data.GetGroupByUUID(uuid)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, r, "你好，未能找到该集团。")
		return
	}

	// 获取集团的所有团队
	teams, err := data.GetTeamsByGroupId(group.Id)
	if err != nil {
		util.Debug("Cannot get teams by group id", err)
	}

	// 检查用户权限
	canManage, err := group.CanManage(sessUser.Id)
	if err != nil {
		util.Debug("Cannot check manage permission", err)
		canManage = false
	}

	var pageData struct {
		SessUser  data.User
		Group     data.Group
		Teams     []data.Team
		CanManage bool
		IsFounder bool
	}
	pageData.SessUser = sessUser
	pageData.Group = group
	pageData.Teams = teams
	pageData.CanManage = canManage
	pageData.IsFounder = group.IsFounder(sessUser.Id)

	generateHTML(w, &pageData, "layout", "navbar.private", "group.read")
}

// checkGroupPermission 检查集团权限的辅助函数
func checkGroupPermission(group *data.Group, userId int, permissionType string) (bool, error) {
	switch permissionType {
	case "manage":
		return group.CanManage(userId)
	case "edit":
		return group.CanEdit(userId)
	case "delete":
		return group.CanDelete(userId), nil
	case "add_team":
		return group.CanAddTeam(userId)
	case "remove_team":
		return group.CanRemoveTeam(userId)
	default:
		return false, nil
	}
}

// POST /v1/group/add_team
// 添加团队到集团
func AddTeamToGroupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能添加团队。")
		return
	}

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，未能添加团队。")
		return
	}

	groupIdStr := r.PostFormValue("group_id")
	teamIdStr := r.PostFormValue("team_id")
	levelStr := r.PostFormValue("level")
	role := r.PostFormValue("role")

	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		report(w, r, "你好，集团ID无效。")
		return
	}

	// 检查权限
	group := data.Group{Id: groupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, r, "你好，未找到该集团。")
		return
	}

	canAdd, err := checkGroupPermission(&group, sessUser.Id, "add_team")
	if err != nil {
		util.Debug("Cannot check permission", err)
		report(w, r, "你好，权限检查失败。")
		return
	}
	if !canAdd {
		report(w, r, "你好，您没有权限添加团队到该集团。")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, r, "你好，团队ID无效。")
		return
	}

	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 {
		level = 2 // 默认为次级
	}

	if role == "" {
		role = "成员团队"
	}

	member := data.GroupMember{
		GroupId: groupId,
		TeamId:  teamId,
		Level:   level,
		Role:    role,
		Status:  data.GroupMemberStatusActive,
		UserId:  sessUser.Id,
	}

	if err := member.Create(); err != nil {
		util.Debug("Cannot create group member", err)
		report(w, r, "你好，茶博士失魂鱼，未能添加团队到集团。")
		return
	}

	report(w, r, "你好，团队已成功添加到集团！")
}

// POST /v1/group/edit
// 编辑集团信息
func EditGroupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能编辑集团。")
		return
	}

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，未能编辑集团。")
		return
	}

	groupIdStr := r.PostFormValue("group_id")
	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		report(w, r, "你好，集团ID无效。")
		return
	}

	// 获取集团并检查权限
	group := data.Group{Id: groupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, r, "你好，未找到该集团。")
		return
	}

	canEdit, err := checkGroupPermission(&group, sessUser.Id, "edit")
	if err != nil {
		util.Debug("Cannot check permission", err)
		report(w, r, "你好，权限检查失败。")
		return
	}
	if !canEdit {
		report(w, r, "你好，您没有权限编辑该集团。")
		return
	}

	// 更新集团信息
	group.Name = r.PostFormValue("name")
	group.Abbreviation = r.PostFormValue("abbreviation")
	group.Mission = r.PostFormValue("mission")

	if err := group.Update(); err != nil {
		util.Debug("Cannot update group", err)
		report(w, r, "你好，茶博士失魂鱼，未能更新集团信息。")
		return
	}

	report(w, r, "你好，集团信息已成功更新！")
}

// GET /v1/group/detail?id=xxx
// 显示集团详情（根据团队UUID或集团UUID）
func GroupDetailGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		report(w, r, "你好，缺少标识参数。")
		return
	}

	// 先尝试作为团队UUID查询
	team, err := data.GetTeamByUUID(id)
	var group *data.Group

	if err == nil {
		// 是团队UUID，查询该团队所属的集团
		group, err = data.GetGroupByTeamId(team.Id)
		if err != nil {
			// 团队未加入任何集团，跳转到创建集团页面
			http.Redirect(w, r, "/v1/group/new?team_id="+id, http.StatusFound)
			return
		}
	} else {
		// 尝试作为集团UUID查询
		groupData, err := data.GetGroupByUUID(id)
		if err != nil {
			util.Debug("Cannot get group by uuid", err)
			report(w, r, "你好，未能找到该集团或团队。")
			return
		}
		group = &groupData
	}

	// 获取集团的所有团队
	teams, err := data.GetTeamsByGroupId(group.Id)
	if err != nil {
		util.Debug("Cannot get teams by group id", err)
	}

	// 获取集团创建者
	founder, err := data.GetUser(group.FounderId)
	if err != nil {
		util.Debug("Cannot get group founder", err)
	}

	// 检查用户权限
	// canManage, err := group.CanManage(sessUser.Id)
	// if err != nil {
	// 	util.Debug("Cannot check manage permission", err)
	// 	canManage = false
	// }

	// 准备页面数据
	var pageData data.GroupDetail
	pageData.SessUser = sessUser
	pageData.GroupBean = data.GroupBean{
		Group:         *group,
		CreatedAtDate: group.CreatedAtDate(),
		Open:          group.Class == data.GroupClassOpen,
		Founder:       founder,
	}

	// 获取团队Bean列表
	teamBeans := make([]data.TeamBean, 0, len(teams))
	for _, t := range teams {
		tb, err := fetchTeamBean(t)
		if err == nil {
			teamBeans = append(teamBeans, tb)
		}
	}
	pageData.TeamBeanSlice = teamBeans

	generateHTML(w, &pageData, "layout", "navbar.private", "group.detail")
}

// POST /v1/group/delete
// 删除集团（软删除）
func DeleteGroupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能删除集团。")
		return
	}

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，未能删除集团。")
		return
	}

	groupIdStr := r.PostFormValue("group_id")
	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		report(w, r, "你好，集团ID无效。")
		return
	}

	// 获取集团并检查权限
	group := data.Group{Id: groupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, r, "你好，未找到该集团。")
		return
	}

	canDelete, err := checkGroupPermission(&group, sessUser.Id, "delete")
	if err != nil {
		util.Debug("Cannot check permission", err)
		report(w, r, "你好，权限检查失败。")
		return
	}
	if !canDelete {
		report(w, r, "你好，只有集团创建者才能删除集团。")
		return
	}

	if err := group.SoftDelete(); err != nil {
		util.Debug("Cannot delete group", err)
		report(w, r, "你好，茶博士失魂鱼，未能删除集团。")
		return
	}

	report(w, r, "你好，集团已成功删除！")
}
