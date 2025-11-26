package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/notification/invitation_group
// 查看集团邀请函列表
func InvitationGroup(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		util.Debug("Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		util.Debug("Cannot get user", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 获取用户所在的担任CEO的团队收到的所有集团邀请函
	invitations, err := data.GetGroupInvitationsByUserId(sessUser.Id)
	if err != nil {
		util.Debug("Cannot get group invitations", err)
		report(w, r, "你好，茶博士在努力查找您的邀请函中，请稍后再试。")
		return
	}

	// 构建页面数据
	type GroupInvitationItem struct {
		Invitation data.GroupInvitation
		Group      data.Group
		Team       data.Team
	}

	var invitationItems []GroupInvitationItem
	for _, inv := range invitations {
		group := data.Group{Id: inv.GroupId}
		if err := group.Get(); err != nil {
			continue
		}
		team, err := data.GetTeam(inv.TeamId)
		if err != nil {
			continue
		}
		invitationItems = append(invitationItems, GroupInvitationItem{
			Invitation: inv,
			Group:      group,
			Team:       team,
		})
	}

	// 统计各状态数量
	unreadCount, _ := data.CountGroupInvitationsByUserIdAndStatus(sessUser.Id, 0)
	viewedCount, _ := data.CountGroupInvitationsByUserIdAndStatus(sessUser.Id, 1)
	acceptedCount, _ := data.CountGroupInvitationsByUserIdAndStatus(sessUser.Id, 2)
	rejectedCount, _ := data.CountGroupInvitationsByUserIdAndStatus(sessUser.Id, 3)

	var pageData struct {
		SessUser                     data.User
		GroupInvitationSlice         []GroupInvitationItem
		GroupInvitationUnreadCount   int
		GroupInvitationViewedCount   int
		GroupInvitationAcceptedCount int
		GroupInvitationRejectedCount int
		GroupInvitationTotalCount    int
	}
	pageData.SessUser = sessUser
	pageData.GroupInvitationSlice = invitationItems
	pageData.GroupInvitationUnreadCount = unreadCount
	pageData.GroupInvitationViewedCount = viewedCount
	pageData.GroupInvitationAcceptedCount = acceptedCount
	pageData.GroupInvitationRejectedCount = rejectedCount
	pageData.GroupInvitationTotalCount = len(invitations)

	generateHTML(w, &pageData, "layout", "navbar.private", "notification.invitation_group")
}

// GET /v1/notification/invitation_team
// 用户关于团队邀请函的通知
func InvitationsTeam(w http.ResponseWriter, r *http.Request) {
	//获取session
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var lbPD data.LetterboxPageData

	i_slice, err := s_u.Invitations()
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations")
		report(w, r, "你好，满头大汗的茶博士在努力查找您的邀请函中，请稍后再试。")
		return
	}
	i_b_slice, err := fetchInvitationBeanSlice(i_slice)
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations bean slice")
		report(w, r, "你好，茶博士在加倍努力查找您的邀请函中，请稍后再试。")
		return
	}

	//填写页面资料
	lbPD.SessUser = s_u
	lbPD.InvitationBeanSlice = i_b_slice

	//向用户返回接收邀请函的表单页面
	generateHTML(w, &lbPD, "layout", "navbar.private", "notification.invitation_team")
}

// Get /v1/notification/accetp
// read AcceptNotifications page
func AcceptNotifications(w http.ResponseWriter, r *http.Request) {
	//获取session
	sess, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user", err)
		report(w, r, "你好，满头大汗的茶博士在努力中，请稍后再试。")
		return
	}
	var amPD data.AcceptNotificationPageData
	//填写页面资料
	amPD.SessUser = s_u
	amPD.AcceptNotificationSlice, err = s_u.UnreadAcceptNotifications()
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations")
		report(w, r, "你好，满头大汗的茶博士在加倍努力查找您的资料中，请稍后再试。")
		return
	}

	// 查询集团邀请未读数量
	amPD.GroupInvitationUnreadCount, _ = data.CountGroupInvitationsByUserIdAndStatus(s_u.Id, 0)

	//向用户返回表单页面
	generateHTML(w, &amPD, "layout", "navbar.private", "notification.accept")

}
