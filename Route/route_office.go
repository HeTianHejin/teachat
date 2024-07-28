package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /office/draftThread
// 激活新茶语草稿进入邻桌盲评流程
func ActivateDraftThread(w http.ResponseWriter, r *http.Request) {

}

// GET /pilot/InvitePage
// 打开获取邀请词页面
func InvitePage(w http.ResponseWriter, r *http.Request) {
	s, e := util.Session(r)
	if e != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, e := s.User()
	if e != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	if u.Role == "pilot" || u.Role == "captain" {
		util.GenerateHTML(w, nil, "layout", "navbar.private", "pilot.invite")
	}
}

// 向2个非当前用户发送盲评审核消息
func AcceptMessageSendExceptUserId(u_id int, mess data.AcceptMessage) error {
	var user_ids []int
	var err error
	if data.UserCount() < 50 {
		//test
		if user_ids, err = data.Get2RandomUserId(); err != nil {
			util.Info(err, " Cannot get 2 random user id")
			return err
		}
	} else {
		if data.SessionCount() > 12 && data.SessionCount() < 250 {
			// 在线不同性别随机
			user_ids, err = data.Get2GenderRandomSUserIdExceptId(u_id)
			if err != nil {
				util.Info(err, " Cannot get 2 gender random sess_user id")
				user_ids, err = data.Get2GenderRandomUserIdExceptId(u_id)
				if err != nil {
					util.Info(err, " Cannot get 2 gender random user id")
					// 在线不分性别随机
					user_ids, err = data.Get2RandomSUserIdExceptId(u_id)
					if err != nil {
						util.Info(err, " Cannot get 2 random sess_user id")
						return err
					}
				}
			}
		} else {
			if user_ids, err = data.Get2RandomUserId(); err != nil {
				util.Info(err, " Cannot get 2 random user id")
				return err
			}
		}
	}
	if len(user_ids) != 2 {
		util.Info(err, " Cannot get 2 random-user-ids")
		return err
	}
	// 发送“是否接纳”消息
	if err = mess.Send(user_ids); err != nil {
		util.Info(err, " Cannot send accept message")
		return err
	}
	// 记录用户有1新消息
	for _, id := range user_ids {
		if err = data.AddUserMessageCount(id); err != nil {
			util.Info(err, " Cannot add user new-message-count")
			return err
		}
	}
	return nil
}

// 向当前用户发送友邻盲评结果通知消息
func PilotAcceptMessageSend(u_id int, mess data.AcceptMessage) error {

	// 发送友����评结果通知消息
	if err := mess.Send([]int{u_id}); err != nil {
		util.Info(err, " Cannot send accept message")
		return err
	}
	// ��录用户有1新消息
	if err := data.AddUserMessageCount(u_id); err != nil {
		util.Info(err, " Cannot add user new-message-count")
		return err
	}
	return nil
}
