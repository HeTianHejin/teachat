package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	dao "teachat/DAO"
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
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 如果有team_id参数，检查用户是否为该团队的CEO或创建人
	teamId := r.URL.Query().Get("team_id")
	if teamId != "" {
		team, err := dao.GetTeamByUUID(teamId)
		if err != nil {
			util.Debug("Cannot get team by uuid", err)
			report(w, s_u, "你好，未能找到指定的团队。")
			return
		}

		// 检查权限：必须是创建人或CEO
		isCEO := false
		if s_u.Id != team.FounderId {
			ceo, err := team.MemberCEO()
			if err == nil && ceo.UserId == s_u.Id {
				isCEO = true
			}
			if !isCEO {
				report(w, s_u, "你好，只有团队创建人或CEO才能代表团队创建集团。")
				return
			}
		}
	}

	// 获取用户的团队列表，用于选择最高管理团队
	teams, err := s_u.SurvivalTeams()
	if err != nil {
		util.Debug("Cannot get user teams", err)
	}

	var pageData struct {
		SessUser        dao.User
		Teams           []dao.Team
		PreSelectedTeam string // 预选的团队UUID
	}
	pageData.SessUser = s_u
	pageData.Teams = teams
	pageData.PreSelectedTeam = teamId

	generateHTML(w, &pageData, "layout", "navbar.private", "group.new")
}

// POST /v1/group/create
// 创建新集团
func CreateGroupPost(w http.ResponseWriter, r *http.Request) {

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}
	// 读取表单数据
	name := r.PostFormValue("name")
	abbreviation := r.PostFormValue("abbreviation")
	mission := r.PostFormValue("mission")
	firstTeamIdStr := r.PostFormValue("first_team_id")
	classStr := r.PostFormValue("class")

	// 转换团队ID
	firstTeamId, err := strconv.Atoi(firstTeamIdStr)
	if err != nil {
		report(w, s_u, "你好，请选择最高管理团队。")
		return
	}

	// 检查权限：用户必须是该团队的创建人或CEO
	team, err := dao.GetTeam(firstTeamId)
	if err != nil {
		util.Debug("Cannot get team", err)
		report(w, s_u, "你好，未能找到指定的团队。")
		return
	}

	isCEO := false
	if s_u.Id != team.FounderId {
		ceo, err := team.MemberCEO()
		if err == nil && ceo.UserId == s_u.Id {
			isCEO = true
		}
		if !isCEO {
			report(w, s_u, "你好，只有团队创建人或CEO才能代表团队创建集团。")
			return
		}
	}

	// 验证集团名称长度
	nameLen := cnStrLen(name)
	if nameLen < 4 || nameLen > 24 {
		report(w, s_u, "你好，集团名称应在4-24个中文字符之间。")
		return
	}

	// 验证简称长度
	abbrLen := cnStrLen(abbreviation)
	if abbrLen < 2 || abbrLen > 8 {
		report(w, s_u, "你好，集团简称应在2-8个中文字符之间。")
		return
	}

	// 验证使命长度
	missionLen := cnStrLen(mission)
	if missionLen < int(util.Config.ThreadMinWord) || missionLen > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "你好，集团使命字数不符合要求。")
		return
	}

	// 转换类型
	class, err := strconv.Atoi(classStr)
	if err != nil {
		util.Debug("Cannot convert class to int", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}

	// 检测class是否合规（草稿状态）
	switch class {
	case dao.GroupClassOpenDraft, dao.GroupClassCloseDraft:
		break
	default:
		report(w, s_u, "你好，茶博士摸摸头，竟然说集团类别不合适，未能创建新集团。")
		return
	}

	// 使用事务创建集团并添加第一团队为成员
	group := dao.Group{
		Name:         name,
		Abbreviation: abbreviation,
		Mission:      mission,
		FounderId:    s_u.Id,
		FirstTeamId:  firstTeamId,
		Class:        class,
		Logo:         "groupLogo",
		Tags:         r.PostFormValue("tags"),
	}

	if err := createGroupWithFirstMember(&group, firstTeamId, s_u.Id); err != nil {
		util.Debug("Cannot create group with first member", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}

	if util.Config.PoliteMode {
		//启用了友邻蒙评
		if err = createAndSendAcceptNotification(group.Id, dao.AcceptObjectTypeGroup, s_u.Id, r.Context()); err != nil {
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				report(w, s_u, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				report(w, s_u, "你好，茶博士迷路了，未能发送蒙评请求通知。")
			}
			return
		}

		// 提示用户新集团草稿保存成功
		text := ""
		if s_u.Gender == dao.User_Gender_Female {
			text = fmt.Sprintf("%s 女士，你好，登记 %s 集团草稿已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", s_u.Name, group.Name)
		} else {
			text = fmt.Sprintf("%s 先生，你好，登记 %s 集团草稿已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", s_u.Name, group.Name)
		}
		report(w, s_u, text)
	} else {
		switch group.Class {
		case dao.GroupClassOpenDraft:
			group.Class = dao.GroupClassOpen
		case dao.GroupClassCloseDraft:
			group.Class = dao.GroupClassClose
		}
		if err := group.Update(); err != nil {
			util.Debug("Cannot update group class", err)
			report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
			return
		}

		// 跳转到集团详情页面
		http.Redirect(w, r, "/v1/group/read?uuid="+group.Uuid, http.StatusFound)
	}
}

// GET /v1/groups
// 显示所有集团列表
func GroupsGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var pageData struct {
		SessUser dao.User
		Groups   []dao.Group
	}
	pageData.SessUser = s_u

	// TODO: 实现获取所有活跃集团的方法
	// pageData.Groups, err = dao.GetActiveGroups()

	generateHTML(w, &pageData, "layout", "navbar.private", "groups.list")
}

// GET /v1/group/read?uuid=xxx
// 显示集团详情
func GroupReadGet(w http.ResponseWriter, r *http.Request) {

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}
	uuid := r.FormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，缺少集团标识。")
		return
	}

	group, err := dao.GetGroupByUUID(uuid)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, s_u, "你好，未能找到该集团。")
		return
	}

	// 获取集团的所有团队
	teams, err := dao.GetTeamsByGroupId(group.Id)
	if err != nil {
		util.Debug("Cannot get teams by group id", err)
	}

	// 检查用户权限
	canManage, err := group.CanManage(s_u.Id)
	if err != nil {
		util.Debug("Cannot check manage permission", err)
		canManage = false
	}

	var pageData struct {
		SessUser  dao.User
		Group     dao.Group
		Teams     []dao.Team
		CanManage bool
		IsFounder bool
	}
	pageData.SessUser = s_u
	pageData.Group = group
	pageData.Teams = teams
	pageData.CanManage = canManage
	pageData.IsFounder = group.IsFounder(s_u.Id)

	generateHTML(w, &pageData, "layout", "navbar.private", "group.read")
}

// checkGroupPermission 检查集团权限的辅助函数
func checkGroupPermission(group *dao.Group, userId int, permissionType string) (bool, error) {
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

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, s_u, "你好，茶博士失魂鱼，未能添加团队。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}
	groupIdStr := r.PostFormValue("group_id")
	teamIdStr := r.PostFormValue("team_id")
	levelStr := r.PostFormValue("level")
	role := r.PostFormValue("role")

	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		report(w, s_u, "你好，集团ID无效。")
		return
	}

	// 检查权限
	group := dao.Group{Id: groupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, s_u, "你好，未找到该集团。")
		return
	}

	canAdd, err := checkGroupPermission(&group, s_u.Id, "add_team")
	if err != nil {
		util.Debug("Cannot check permission", err)
		report(w, s_u, "你好，权限检查失败。")
		return
	}
	if !canAdd {
		report(w, s_u, "你好，您没有权限添加团队到该集团。")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, s_u, "你好，团队ID无效。")
		return
	}

	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 {
		level = 2 // 默认为次级
	}

	if role == "" {
		role = "成员团队"
	}

	member := dao.GroupMember{
		GroupId: groupId,
		TeamId:  teamId,
		Level:   level,
		Role:    role,
		Status:  dao.GroupMemberStatusActive,
		UserId:  s_u.Id,
	}

	if err := member.Create(); err != nil {
		util.Debug("Cannot create group member", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能添加团队到集团。")
		return
	}

	report(w, s_u, "你好，团队已成功添加到集团！")
}

// GET /v1/group/edit?id=xxx
// 显示编辑集团信息表单
func EditGroupGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		report(w, s_u, "你好，缺少集团标识。")
		return
	}

	group, err := dao.GetGroupByUUID(id)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, s_u, "你好，未能找到该集团。")
		return
	}

	// 检查编辑权限
	canEdit, err := group.CanEdit(s_u.Id)
	if err != nil || !canEdit {
		report(w, s_u, "你好，您没有权限编辑该集团。")
		return
	}

	var pageData struct {
		SessUser dao.User
		Group    dao.Group
	}
	pageData.SessUser = s_u
	pageData.Group = group

	generateHTML(w, &pageData, "layout", "navbar.private", "group.edit")
}

// HandleEditGroup GET/POST /v1/group/edit
// 处理集团编辑
func HandleEditGroup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		EditGroupGet(w, r)
	case http.MethodPost:
		EditGroupPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// POST /v1/group/edit
// 编辑集团信息
func EditGroupPost(w http.ResponseWriter, r *http.Request) {

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, s_u, "你好，茶博士失魂鱼，未能编辑集团。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}
	groupIdStr := r.PostFormValue("group_id")
	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		report(w, s_u, "你好，集团ID无效。")
		return
	}

	// 获取集团并检查权限
	group := dao.Group{Id: groupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, s_u, "你好，未找到该集团。")
		return
	}

	canEdit, err := checkGroupPermission(&group, s_u.Id, "edit")
	if err != nil {
		util.Debug("Cannot check permission", err)
		report(w, s_u, "你好，权限检查失败。")
		return
	}
	if !canEdit {
		report(w, s_u, "你好，您没有权限编辑该集团。")
		return
	}

	// 更新集团信息
	group.Name = r.PostFormValue("name")
	group.Abbreviation = r.PostFormValue("abbreviation")
	group.Mission = r.PostFormValue("mission")

	if err := group.Update(); err != nil {
		util.Debug("Cannot update group", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能更新集团信息。")
		return
	}

	http.Redirect(w, r, "/v1/group/manage?id="+group.Uuid, http.StatusFound)
}

// GET /v1/group/detail?id=xxx
// 显示集团详情（根据团队UUID或集团UUID）
func GroupDetailGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		report(w, s_u, "你好，缺少标识参数。")
		return
	}

	// 先尝试作为团队UUID查询
	team, err := dao.GetTeamByUUID(id)
	var group *dao.Group

	if err == nil {
		// 是团队UUID，查询该团队所属的集团
		group, err = dao.GetGroupByTeamId(team.Id)
		if err != nil {
			// 团队未加入任何集团，跳转到创建集团页面
			http.Redirect(w, r, "/v1/group/new?team_id="+id, http.StatusFound)
			return
		}
	} else {
		// 尝试作为集团UUID查询
		groupData, err := dao.GetGroupByUUID(id)
		if err != nil {
			util.Debug("Cannot get group by uuid", err)
			report(w, s_u, "你好，未能找到该集团或团队。")
			return
		}
		group = &groupData
	}

	// 获取集团的所有团队
	teams, err := dao.GetTeamsByGroupId(group.Id)
	if err != nil {
		util.Debug("Cannot get teams by group id", err)
	}

	// 获取集团创建者
	founder, err := dao.GetUser(group.FounderId)
	if err != nil {
		util.Debug("Cannot get group founder", err)
	}

	// 检查用户权限
	canManage, err := group.CanManage(s_u.Id)
	if err != nil {
		util.Debug("Cannot check manage permission", err)
		canManage = false
	}

	// 获取创建者的默认团队
	founderTeam, err := founder.GetLastDefaultTeam()
	if err != nil {
		util.Debug("Cannot get founder default team", err)
		founderTeam = dao.Team{Id: dao.TeamIdNone}
	}

	// 准备页面数据
	var pageData dao.GroupDetail
	pageData.SessUser = s_u
	pageData.CanManage = canManage

	pageData.GroupBean = dao.GroupBean{
		Group:         *group,
		CreatedAtDate: group.CreatedAtDate(),
		Open:          group.Class == dao.GroupClassOpen,
		Founder:       founder,
		FounderTeam:   founderTeam,
		TeamsCount:    len(teams),
	}

	// 获取第一团队（最高管理团队）
	// if group.FirstTeamId > 0 {
	// 	firstTeam, err := dao.GetTeam(group.FirstTeamId)
	// 	if err == nil {
	// 		pageData.FirstTeamBean, err = fetchTeamBean(firstTeam)
	// 		if err != nil {
	// 			util.Debug("Cannot fetch first team bean", err)
	// 		}
	// 	}
	// }

	// 获取团队Bean列表（排除第一团队）
	teamBeans := make([]dao.TeamBean, 0, len(teams))
	for _, t := range teams {
		// if t.Id == group.FirstTeamId {
		// 	continue // 跳过第一团队，因为已经单独显示
		// }
		tb, err := fetchTeamBean(t)
		if err == nil {
			teamBeans = append(teamBeans, tb)
		}
	}
	pageData.TeamBeanSlice = teamBeans
	pageData.IsOverTwelve = len(teamBeans) > 12

	generateHTML(w, &pageData, "layout", "navbar.private", "group.detail", "component_team")
}

// GET /v1/group/manage?id=xxx
// 显示集团管理页面
func GroupManageGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		report(w, s_u, "你好，缺少集团标识。")
		return
	}

	group, err := dao.GetGroupByUUID(id)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, s_u, "你好，未能找到该集团。")
		return
	}

	// 检查管理权限
	canManage, err := group.CanManage(s_u.Id)
	if err != nil {
		util.Debug("Cannot check manage permission", err)
		canManage = false
	}
	if !canManage {
		report(w, s_u, "你好，您没有权限管理该集团。")
		return
	}

	// 获取集团的所有团队
	teams, err := dao.GetTeamsByGroupId(group.Id)
	if err != nil {
		util.Debug("Cannot get teams by group id", err)
	}

	// 准备页面数据
	var pageData struct {
		SessUser      dao.User
		GroupBean     dao.GroupBean
		TeamBeanSlice []dao.TeamBean
	}
	pageData.SessUser = s_u
	pageData.GroupBean = dao.GroupBean{
		Group:         group,
		CreatedAtDate: group.CreatedAtDate(),
		Open:          group.Class == dao.GroupClassOpen,
		TeamsCount:    len(teams),
	}

	// 获取团队Bean列表
	teamBeans := make([]dao.TeamBean, 0, len(teams))
	for _, t := range teams {
		tb, err := fetchTeamBean(t)
		if err == nil {
			teamBeans = append(teamBeans, tb)
		}
	}
	pageData.TeamBeanSlice = teamBeans

	generateHTML(w, &pageData, "layout", "navbar.private", "group.manage")
}

// GET /v1/group/invitations?id=xxx
// 显示集团发出的所有邀请函列表
func GroupInvitationsGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		report(w, s_u, "你好，缺少集团标识。")
		return
	}

	group, err := dao.GetGroupByUUID(id)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, s_u, "你好，未能找到该集团。")
		return
	}

	// 检查管理权限
	canManage, err := group.CanManage(s_u.Id)
	if err != nil || !canManage {
		report(w, s_u, "你好，只有集团管理者才能查看邀请函列表。")
		return
	}

	// 获取集团发出的所有邀请函
	invitations, err := dao.GetInvitationsByGroupId(group.Id)
	if err != nil {
		util.Debug("Cannot get group invitations", err)
	}

	// 构建邀请函Bean列表
	invitationBeans := make([]dao.GroupInvitationBean, 0)
	for _, inv := range invitations {
		team, err := dao.GetTeam(inv.TeamId)
		if err != nil {
			continue
		}

		author, err := dao.GetUser(inv.AuthorUserId)
		if err != nil {
			continue
		}

		bean := dao.GroupInvitationBean{
			Invitation: inv,
			Group:      group,
			Author:     author,
			InviteUser: dao.User{}, // 受邀请团队CEO，
			Team:       team,
			Status:     inv.GetStatus(),
		}

		// 获取团队CEO信息
		if ceo, err := team.MemberCEO(); err == nil {
			if ceoUser, err := dao.GetUser(ceo.UserId); err == nil {
				bean.InviteUser = ceoUser
			}
		}
		// 获取团队信息
		if team, err := dao.GetTeam(inv.TeamId); err == nil {
			bean.Team = team
		}

		invitationBeans = append(invitationBeans, bean)
	}

	var pageData struct {
		SessUser        dao.User
		Group           dao.Group
		InvitationBeans []dao.GroupInvitationBean
	}
	pageData.SessUser = s_u
	pageData.Group = group
	pageData.InvitationBeans = invitationBeans

	generateHTML(w, &pageData, "layout", "navbar.private", "group.invitations")
}

// POST /v1/group/delete
// 删除集团（软删除）
func DeleteGroupPost(w http.ResponseWriter, r *http.Request) {

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, s_u, "你好，茶博士失魂鱼，未能删除集团。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}
	groupIdStr := r.PostFormValue("group_id")
	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		report(w, s_u, "你好，集团ID无效。")
		return
	}

	// 获取集团并检查权限
	group := dao.Group{Id: groupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, s_u, "你好，未找到该集团。")
		return
	}

	canDelete, err := checkGroupPermission(&group, s_u.Id, "delete")
	if err != nil {
		util.Debug("Cannot check permission", err)
		report(w, s_u, "你好，权限检查失败。")
		return
	}
	if !canDelete {
		report(w, s_u, "你好，只有集团创建者才能删除集团。")
		return
	}

	if err := group.SoftDelete(); err != nil {
		util.Debug("Cannot delete group", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能删除集团。")
		return
	}

	report(w, s_u, "你好，集团已成功删除！")
}

// createGroupWithFirstMember 使用事务创建集团并将第一团队登记为成员
func createGroupWithFirstMember(group *dao.Group, firstTeamId int, userId int) error {
	tx, err := dao.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := group.CreateWithTx(tx); err != nil {
		return err
	}

	firstMember := dao.GroupMember{
		GroupId: group.Id,
		TeamId:  firstTeamId,
		Level:   1,
		Role:    "最高管理团队",
		Status:  dao.GroupMemberStatusActive,
		UserId:  userId,
	}
	if err := firstMember.CreateWithTx(tx); err != nil {
		return err
	}

	return tx.Commit()
}
