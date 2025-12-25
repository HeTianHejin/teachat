package route

import (
	"net/http"
	dao "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/group/notification/invitation
// 查看集团邀请函列表
func InvitationGroup(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		util.Debug("Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 获取用户所在的担任CEO的团队收到的所有集团邀请函
	invitations, err := dao.GetGroupInvitationsByUserId(s_u.Id)
	if err != nil {
		util.Debug("Cannot get group invitations", err)
		report(w, s_u, "你好，茶博士在努力查找您的邀请函中，请稍后再试。")
		return
	}

	// 构建页面数据
	type GroupInvitationItem struct {
		Invitation dao.GroupInvitation
		Group      dao.Group
		Team       dao.Team
	}

	var invitationItems []GroupInvitationItem
	for _, inv := range invitations {
		group := dao.Group{Id: inv.GroupId}
		if err := group.Get(); err != nil {
			continue
		}
		team, err := dao.GetTeam(inv.TeamId)
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
	unreadCount, _ := dao.CountGroupInvitationsByUserIdAndStatus(s_u.Id, 0)
	viewedCount, _ := dao.CountGroupInvitationsByUserIdAndStatus(s_u.Id, 1)
	acceptedCount, _ := dao.CountGroupInvitationsByUserIdAndStatus(s_u.Id, 2)
	rejectedCount, _ := dao.CountGroupInvitationsByUserIdAndStatus(s_u.Id, 3)

	var pageData struct {
		SessUser                     dao.User
		GroupInvitationSlice         []GroupInvitationItem
		GroupInvitationUnreadCount   int
		GroupInvitationViewedCount   int
		GroupInvitationAcceptedCount int
		GroupInvitationRejectedCount int
		GroupInvitationTotalCount    int
	}
	pageData.SessUser = s_u
	pageData.GroupInvitationSlice = invitationItems
	pageData.GroupInvitationUnreadCount = unreadCount
	pageData.GroupInvitationViewedCount = viewedCount
	pageData.GroupInvitationAcceptedCount = acceptedCount
	pageData.GroupInvitationRejectedCount = rejectedCount
	pageData.GroupInvitationTotalCount = len(invitations)

	generateHTML(w, &pageData, "layout", "navbar.private", "group.notification.invitation")
}

// GET /v1/team/notification/invitation
// 用户关于团队邀请函的通知
func TeamNotificationInvitations(w http.ResponseWriter, r *http.Request) {
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

	var lbPD dao.LetterboxPageData

	i_slice, err := s_u.Invitations()
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations")
		report(w, s_u, "你好，满头大汗的茶博士在努力查找您的邀请函中，请稍后再试。")
		return
	}
	i_b_slice, err := fetchInvitationBeanSlice(i_slice)
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations bean slice")
		report(w, s_u, "你好，茶博士在加倍努力查找您的邀请函中，请稍后再试。")
		return
	}

	//填写页面资料
	lbPD.SessUser = s_u
	lbPD.InvitationBeanSlice = i_b_slice

	//向用户返回接收邀请函的表单页面
	generateHTML(w, &lbPD, "layout", "navbar.private", "user.notification.invitation")
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
		report(w, s_u, "你好，满头大汗的茶博士在努力中，请稍后再试。")
		return
	}
	var amPD dao.AcceptNotificationPageData
	//填写页面资料
	amPD.SessUser = s_u
	amPD.AcceptNotificationSlice, err = s_u.UnreadAcceptNotifications()
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations")
		report(w, s_u, "你好，满头大汗的茶博士在加倍努力查找您的资料中，请稍后再试。")
		return
	}

	// 查询集团邀请未读数量
	amPD.GroupInvitationUnreadCount, _ = dao.CountGroupInvitationsByUserIdAndStatus(s_u.Id, 0)

	//向用户返回表单页面
	generateHTML(w, &amPD, "layout", "navbar.private", "accept.notifications")

}
