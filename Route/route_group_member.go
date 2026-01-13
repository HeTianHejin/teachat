package route

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	dao "teachat/DAO"
	util "teachat/Util"
)

// HandleGroupMemberInvite GET/POST /v1/group/member_invite
// 邀请团队加入集团
func HandleGroupMemberInvite(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GroupMemberInviteGet(w, r)
	case http.MethodPost:
		GroupMemberInvitePost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GET /v1/group/member_invite?uuid=xxx&team_uuid=xxx
// 显示邀请团队加入集团的表单
func GroupMemberInviteGet(w http.ResponseWriter, r *http.Request) {
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

	groupUuid := r.URL.Query().Get("uuid")
	if groupUuid == "" {
		report(w, s_u, "你好，缺少集团标识。")
		return
	}

	group, err := dao.GetGroupByUUID(groupUuid)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, s_u, "你好，未能找到该集团。")
		return
	}

	// 检查权限
	canManage, err := group.CanManage(s_u.Id)
	if err != nil || !canManage {
		report(w, s_u, "你好，只有集团管理者才能邀请团队加入。")
		return
	}

	// 如果提供了team_uuid，获取团队信息
	var inviteTeam *dao.Team
	teamUuid := r.URL.Query().Get("team_uuid")
	if teamUuid != "" {
		team, err := dao.GetTeamByUUID(teamUuid)
		if err != nil {
			util.Debug("Cannot get team by uuid", err)
		} else {
			inviteTeam = &team
		}
	}

	var pageData struct {
		SessUser   dao.User
		Group      dao.Group
		InviteTeam *dao.Team
	}
	pageData.SessUser = s_u
	pageData.Group = group
	pageData.InviteTeam = inviteTeam

	generateHTML(w, &pageData, "layout", "navbar.private", "group.member_invite")
}

// POST /v1/group/member_invite
// 处理邀请团队加入集团的请求
func GroupMemberInvitePost(w http.ResponseWriter, r *http.Request) {

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, s_u, "你好，茶博士失魂鱼，未能处理邀请请求。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能处理邀请请求。")
		return
	}
	groupIdStr := r.PostFormValue("group_id")
	teamIdStr := r.PostFormValue("team_id")
	inviteWord := r.PostFormValue("invite_word")
	roleStr := r.PostFormValue("role")
	levelStr := r.PostFormValue("level")

	// 验证邀请词长度
	if cnStrLen(inviteWord) < 2 || cnStrLen(inviteWord) > 239 {
		report(w, s_u, "你好，邀请词长度应在2-239个字符之间。")
		return
	}

	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		report(w, s_u, "你好，集团ID无效。")
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

	// 转换角色字符串为整数
	roleInt := dao.GroupRoleMember // 默认为成员团队
	if roleStr == "最高管理团队" {
		roleInt = dao.GroupRoleTopManagement
	}

	// 获取集团并检查权限
	group := dao.Group{Id: groupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, s_u, "你好，未找到该集团。")
		return
	}

	canManage, err := group.CanManage(s_u.Id)
	if err != nil || !canManage {
		report(w, s_u, "你好，只有集团管理者才能邀请团队加入。")
		return
	}

	// 检查团队是否已经是集团成员
	_, err = dao.GetGroupMemberByGroupIdAndTeamId(groupId, teamId)
	if err == nil {
		report(w, s_u, "你好，该团队已经是集团成员。")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		util.Debug("Cannot check group member", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 创建集团邀请函
	invitation := dao.GroupInvitation{
		GroupId:      groupId,
		TeamId:       teamId,
		InviteWord:   inviteWord,
		Role:         roleInt,
		Level:        level,
		Status:       0,
		AuthorUserId: s_u.Id,
	}

	if err := invitation.Create(); err != nil {
		util.Debug("Cannot create group invitation", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建邀请函。")
		return
	}

	// 向团队CEO发送通知通知
	team, _ := dao.GetTeam(teamId)
	if ceo, err := team.MemberCEO(); err == nil {
		if err = dao.AddUserNotificationCount(ceo.UserId); err != nil {
			util.Debug("Cannot add user notification count", err)
		}
	}

	report(w, s_u, fmt.Sprintf("你好，已成功向团队 %s 发送加入集团邀请函。", team.Name))
}

// HandleGroupMemberInvitation GET/POST /v1/group/member_invitation
// 查看和处理集团邀请函
func HandleGroupMemberInvitation(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GroupMemberInvitationRead(w, r)
	case http.MethodPost:
		GroupMemberInvitationReply(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GET /v1/group/member_invitation?id=xxx
// 查看集团邀请函详情
func GroupMemberInvitationRead(w http.ResponseWriter, r *http.Request) {
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

	invitationUuid := r.URL.Query().Get("id")
	invitation, err := dao.GetGroupInvitationByUuid(invitationUuid)
	if err != nil {
		util.Debug("Cannot get group invitation", err)
		report(w, s_u, "你好，未能找到该邀请函。")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeam(invitation.TeamId)
	if err != nil {
		util.Debug("Cannot get team", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 检查权限：必须是团队CEO或创建人
	canReply := false
	if team.FounderId == s_u.Id {
		canReply = true
	} else if member_ceo, err := team.MemberCEO(); err == nil && member_ceo.UserId == s_u.Id {
		canReply = true
	}

	if !canReply {
		report(w, s_u, "你好，只有团队创建人或CEO才能处理邀请函。")
		return
	}

	// 获取集团信息
	group := dao.Group{Id: invitation.GroupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	ceo, _ := dao.GetUser(invitation.AuthorUserId)

	// 如果是未读状态，更新为已读
	if invitation.Status == 0 {
		invitation.Status = 1
		if err := invitation.UpdateStatus(); err != nil {
			util.Debug("Cannot update invitation status", err)
		}
		// 减少通知计数
		if err = dao.SubtractUserNotificationCount(s_u.Id); err != nil {
			util.Debug("Cannot subtract user notification count", err)
		}
	}

	var pageData struct {
		SessUser   dao.User
		Invitation dao.GroupInvitation
		Group      dao.Group
		CEO        dao.User
		Team       dao.Team
	}
	pageData.SessUser = s_u
	pageData.Invitation = invitation
	pageData.Group = group
	pageData.CEO = ceo
	pageData.Team = team

	generateHTML(w, &pageData, "layout", "navbar.private", "group.member_invitation_read")
}

// POST /v1/group/member_invitation
// 处理集团邀请函回复
func GroupMemberInvitationReply(w http.ResponseWriter, r *http.Request) {

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	invitationIdStr := r.PostFormValue("invitation_id")
	replyWord := r.PostFormValue("reply_word")
	replyClassStr := r.PostFormValue("reply")

	invitationId, err := strconv.Atoi(invitationIdStr)
	if err != nil {
		report(w, s_u, "你好，邀请函ID无效。")
		return
	}

	replyClass, err := strconv.Atoi(replyClassStr)
	if err != nil {
		report(w, s_u, "你好，回复类型无效。")
		return
	}

	// 验证回复内容
	if cnStrLen(replyWord) < 2 || cnStrLen(replyWord) > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "你好，回复内容长度不符合要求。")
		return
	}

	// 获取邀请函
	invitation, err := dao.GetGroupInvitationById(invitationId)
	if err != nil {
		util.Debug("Cannot get group invitation", err)
		report(w, s_u, "你好，未能找到该邀请函。")
		return
	}

	// 检查邀请函状态
	if invitation.Status > 1 {
		report(w, s_u, "你好，该邀请函已经处理过了。")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeam(invitation.TeamId)
	if err != nil {
		util.Debug("Cannot get team", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 检查权限
	canReply := false
	if team.FounderId == s_u.Id {
		canReply = true
	} else if ceo, err := team.MemberCEO(); err == nil && ceo.UserId == s_u.Id {
		canReply = true
	}

	if !canReply {
		report(w, s_u, "你好，只有团队创建人或CEO才能处理邀请函。")
		return
	}

	// 获取集团信息
	group := dao.Group{Id: invitation.GroupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	switch replyClass {
	case 1:
		// 接受邀请
		// 检查团队是否已经是集团成员
		_, err = dao.GetGroupMemberByGroupIdAndTeamId(group.Id, team.Id)
		if err == nil {
			report(w, s_u, "你好，该团队已经是集团成员。")
			return
		} else if !errors.Is(err, sql.ErrNoRows) {
			util.Debug("Cannot check group member", err)
			report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}

		// 创建集团成员记录
		member := dao.GroupMember{
			GroupId: group.Id,
			TeamId:  team.Id,
			Level:   invitation.Level,
			Role:    invitation.Role,
			Status:  dao.GroupMemberStatusActive,
			UserId:  s_u.Id,
		}

		if err := member.Create(); err != nil {
			util.Debug("Cannot create group member", err)
			report(w, s_u, "你好，茶博士失魂鱼，未能加入集团。")
			return
		}

		// 更新邀请函状态为已接受
		invitation.Status = 2
		if err := invitation.UpdateStatus(); err != nil {
			util.Debug("Cannot update invitation status", err)
		}

		// 创建回复记录
		reply := dao.GroupInvitationReply{
			InvitationId: invitationId,
			UserId:       s_u.Id,
			ReplyWord:    replyWord,
		}
		if err := reply.Create(); err != nil {
			util.Debug("Cannot create invitation reply", err)
		}

		http.Redirect(w, r, "/v1/group/detail?id="+group.Uuid, http.StatusFound)

	case 0:
		// 拒绝邀请
		invitation.Status = 3
		if err := invitation.UpdateStatus(); err != nil {
			util.Debug("Cannot update invitation status", err)
		}

		// 创建回复记录
		reply := dao.GroupInvitationReply{
			InvitationId: invitationId,
			UserId:       s_u.Id,
			ReplyWord:    replyWord,
		}
		if err := reply.Create(); err != nil {
			util.Debug("Cannot create invitation reply", err)
		}

		report(w, s_u, fmt.Sprintf("你好，已婉拒加入 %s 集团的邀请。", group.Name))

	default:
		report(w, s_u, "你好，无效的回复类型。")
	}
}

// HandleGroupSearchTeam POST /v1/group/search_team
// 处理集团搜索团队请求
func HandleGroupSearchTeam(w http.ResponseWriter, r *http.Request) {

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
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	groupUuid := r.PostFormValue("group_uuid")
	searchType := r.PostFormValue("search_type")
	keyword := r.PostFormValue("keyword")

	// 验证关键词长度
	if len(keyword) < 1 || len(keyword) > 32 {
		report(w, s_u, "你好，茶博士摸摸头，说关键词太长了记不住呢，请确认后再试。")
		return
	}

	// 获取集团信息
	group, err := dao.GetGroupByUUID(groupUuid)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, s_u, "你好，未能找到该集团。")
		return
	}

	// 检查权限
	canManage, err := group.CanManage(s_u.Id)
	if err != nil || !canManage {
		report(w, s_u, "你好，只有集团管理者才能添加成员。")
		return
	}

	var pageData struct {
		SessUser      dao.User
		Group         dao.Group
		TeamBeanSlice []dao.TeamBean
		IsEmpty       bool
	}
	pageData.SessUser = s_u
	pageData.Group = group
	pageData.IsEmpty = true

	switch searchType {
	case "team_id":
		// 按团队编号查询
		teamId, err := strconv.Atoi(keyword)
		if err != nil || teamId <= 0 {
			report(w, s_u, "团队编号必须是正整数")
			return
		}

		team, err := dao.GetTeam(teamId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				pageData.IsEmpty = true
			} else {
				util.Debug("Cannot get team by id", err)
				report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
		} else {
			teamBean, err := fetchTeamBean(team)
			if err == nil {
				pageData.TeamBeanSlice = append(pageData.TeamBeanSlice, teamBean)
				pageData.IsEmpty = false
			}
		}

	case "team_abbr":
		// 按团队简称查询
		teamSlice, err := dao.SearchTeamByAbbreviation(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug("Cannot search team by abbreviation", err)
			report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
			return
		}

		if len(teamSlice) >= 1 {
			teamBeanSlice, err := fetchTeamBeanSlice(teamSlice)
			if err == nil && len(teamBeanSlice) >= 1 {
				pageData.TeamBeanSlice = teamBeanSlice
				pageData.IsEmpty = false
			}
		}

	default:
		report(w, s_u, "你好，请选择正确的查询方式。")
		return
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "group.search_team_result", "component_team", "component_avatar_name_gender")
}

// GET /v1/group/member_add?uuid=xxx
// 显示集团增加成员页面（搜索团队）
func GroupMemberAddGet(w http.ResponseWriter, r *http.Request) {
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

	groupUuid := r.URL.Query().Get("uuid")
	if groupUuid == "" {
		report(w, s_u, "你好，缺少集团标识。")
		return
	}

	group, err := dao.GetGroupByUUID(groupUuid)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, s_u, "你好，未能找到该集团。")
		return
	}

	// 检查权限
	canManage, err := group.CanManage(s_u.Id)
	if err != nil || !canManage {
		report(w, s_u, "你好，只有集团管理者才能添加成员。")
		return
	}

	var pageData struct {
		SessUser dao.User
		Group    dao.Group
	}
	pageData.SessUser = s_u
	pageData.Group = group

	generateHTML(w, &pageData, "layout", "navbar.private", "group.member_add")
}

// HandleGroupMemberRemove GET/POST /v1/group/member_remove
// 处理集团移除成员
func HandleGroupMemberRemove(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GroupMemberRemoveGet(w, r)
	case http.MethodPost:
		GroupMemberRemovePost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GET /v1/group/member_remove?uuid=xxx
// 显示集团移除成员页面
func GroupMemberRemoveGet(w http.ResponseWriter, r *http.Request) {
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

	groupUuid := r.URL.Query().Get("uuid")
	if groupUuid == "" {
		report(w, s_u, "你好，缺少集团标识。")
		return
	}

	group, err := dao.GetGroupByUUID(groupUuid)
	if err != nil {
		util.Debug("Cannot get group by uuid", err)
		report(w, s_u, "你好，未能找到该集团。")
		return
	}

	// 检查权限
	canManage, err := group.CanManage(s_u.Id)
	if err != nil || !canManage {
		report(w, s_u, "你好，只有集团管理者才能移除成员。")
		return
	}

	// 获取集团的所有成员团队
	members, err := dao.GetMembersByGroupId(group.Id)
	if err != nil {
		util.Debug("Cannot get group members", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 获取团队详细信息
	var teams []dao.Team
	for _, member := range members {
		if member.Status == dao.GroupMemberStatusActive {
			team, err := dao.GetTeam(member.TeamId)
			if err == nil {
				teams = append(teams, team)
			}
		}
	}

	var pageData struct {
		SessUser dao.User
		Group    dao.Group
		Teams    []dao.Team
	}
	pageData.SessUser = s_u
	pageData.Group = group
	pageData.Teams = teams

	generateHTML(w, &pageData, "layout", "navbar.private", "group.member_remove")
}

// POST /v1/group/member_remove
// 处理移除集团成员
func GroupMemberRemovePost(w http.ResponseWriter, r *http.Request) {

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, s_u, "你好，茶博士失魂鱼，未能处理移除请求。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	groupIdStr := r.PostFormValue("group_id")
	teamIdStr := r.PostFormValue("team_id")

	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		report(w, s_u, "你好，集团ID无效。")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, s_u, "你好，团队ID无效。")
		return
	}

	// 获取集团并检查权限
	group := dao.Group{Id: groupId}
	if err := group.Get(); err != nil {
		util.Debug("Cannot get group", err)
		report(w, s_u, "你好，未找到该集团。")
		return
	}

	canManage, err := group.CanManage(s_u.Id)
	if err != nil || !canManage {
		report(w, s_u, "你好，只有集团管理者才能移除成员。")
		return
	}

	// 不能移除最高管理团队
	if teamId == group.FirstTeamId {
		report(w, s_u, "你好，不能移除最高管理团队。")
		return
	}

	// 获取集团成员记录
	member, err := dao.GetGroupMemberByGroupIdAndTeamId(groupId, teamId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, s_u, "你好，该团队不是集团成员。")
		} else {
			util.Debug("Cannot get group member", err)
			report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		}
		return
	}

	// 软删除成员记录
	if err := member.SoftDelete(); err != nil {
		util.Debug("Cannot delete group member", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能移除成员。")
		return
	}

	team, _ := dao.GetTeam(teamId)
	report(w, s_u, fmt.Sprintf("你好，已成功将团队 %s 移出集团。", team.Name))
}
