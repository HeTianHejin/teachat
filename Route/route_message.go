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
	s, err := Session(r)
	if err != nil {
		util.Danger(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Danger(err, " Cannot get user")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var lbPD data.LetterboxPageData

	i_list, err := s_u.Invitations()
	if err != nil {
		util.Warning(err, s_u.Email, " Cannot get invitations")
		Report(w, r, "你好，茶博士在加倍努力查找您的邀请函中，请稍后再试。")
		return
	}

	//填写页面资料
	lbPD.SessUser = s_u
	lbPD.InvitationList = i_list

	//向用户返回接收邀请函的表单页面
	RenderHTML(w, &lbPD, "layout", "navbar.private", "message.letterbox")
}

// Get /v1/message/accetp
// read AcceptMessages page
func AcceptMessages(w http.ResponseWriter, r *http.Request) {
	//获取session
	sess, err := Session(r)
	if err != nil {
		util.Danger(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, _ := sess.User()
	var amPD data.AcceptMessagePageData
	//填写页面资料
	amPD.SessUser = u
	amPD.AcceptMessageList, err = u.UnreadAcceptMessages()
	if err != nil {
		util.Warning(err, u.Email, " Cannot get invitations")
		Report(w, r, "你好，������在加倍��力查找您的��请��中，请稍后再试。")
		return
	}

	//向用户返回接收��请��的表单页面
	RenderHTML(w, &amPD, "layout", "navbar.private", "message.accept")

}
