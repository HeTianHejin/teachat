package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/message/letter_box
// 用户信箱
func Letterbox(w http.ResponseWriter, r *http.Request) {
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
	renderHTML(w, &lbPD, "layout", "navbar.private", "message.letterbox")
}

// Get /v1/message/accetp
// read AcceptMessages page
func AcceptMessages(w http.ResponseWriter, r *http.Request) {
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
	var amPD data.AcceptMessagePageData
	//填写页面资料
	amPD.SessUser = s_u
	amPD.AcceptMessageSlice, err = s_u.UnreadAcceptMessages()
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations")
		report(w, r, "你好，满头大汗的茶博士在加倍努力查找您的资料中，请稍后再试。")
		return
	}

	//向用户返回表单页面
	renderHTML(w, &amPD, "layout", "navbar.private", "message.accept")

}
