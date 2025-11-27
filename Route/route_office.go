package route

import (
	"context"
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /office/draftThread
// 激活新茶语草稿进入邻桌蒙评流程
func ActivateDraftThread(w http.ResponseWriter, r *http.Request) {

}

// GET /v1/user/invite?id=xxx
// 打开选择邀请茶友成为管理员页面
func Invite(w http.ResponseWriter, r *http.Request) {
	//读取会话资料
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//读取当前用户的相关资料
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot fetch user related data", err)
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	//读取被邀请用户的相关资料
	user_uuid := r.FormValue("id")
	invi_user, err := data.GetUserByUUID(user_uuid)
	if err != nil {
		util.Debug(" Cannot get user by uuid", err)
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	var iD data.InvitationDetail
	// 填写页面资料
	iD.SessUser = s_u

	iD.InvitationBean.Author = s_u
	iD.InvitationBean.InviteUser = invi_user

	generateHTML(w, &iD, "layout", "navbar.private", "pilot.invite")

}

// 向2个非当前用户发送蒙评审核通知
func TwoAcceptNotificationsSendExceptUserId(u_id int, mess data.AcceptNotification, ctx context.Context) error {
	var user_ids []int
	var err error
	// if data.UserCount() < 50 {

	// 	if user_ids, err = data.Get2RandomUserId(); err != nil {
	// 		util.PanicTea(util.LogError(err), " Cannot get 2 random user id")
	// 		return err
	// 	}
	// } else {
	// 	if data.SessionCount() > 12 && data.SessionCount() < 250 {
	// 		// 在线不同性别随机
	// 		user_ids, err = data.Get2GenderRandomSUserIdExceptId(u_id)
	// 		if err != nil {
	// 			util.PanicTea(util.LogError(err), " Cannot get 2 gender random sess_user id")
	// 			user_ids, err = data.Get2GenderRandomUserIdExceptId(u_id)
	// 			if err != nil {
	// 				util.PanicTea(util.LogError(err), " Cannot get 2 gender random user id")
	// 				// 在线不分性别随机
	// 				user_ids, err = data.Get2RandomSUserIdExceptId(u_id)
	// 				if err != nil {
	// 					util.PanicTea(util.LogError(err), " Cannot get 2 random sess_user id")
	// 					return err
	// 				}
	// 			}
	// 		}
	// 	} else {
	if user_ids, err = data.Get2RandomUserId(); err != nil {
		//test status
		util.Debug(" Cannot get 2 random user id", err)
		return err
	}
	// 	}
	// }
	if len(user_ids) != 2 {
		util.Debug(" Cannot get 2 random-user-ids", err)
		return err
	}
	// 发送“是否接纳”通知
	if err = mess.SendWithContext(user_ids, ctx); err != nil {
		util.Debug(" Cannot send accept notification", err)
		return err
	}
	// 记录用户有1新通知
	for _, u_id := range user_ids {
		if err = data.AddUserNotificationCount(u_id); err != nil {
			util.Debug(" Cannot add random user new-notification-count", err)
			return err
		}
	}
	return nil
}

// 向当前用户发送友邻蒙评结果通知通知
func PilotAcceptNotificationSend(u_id int, mess data.AcceptNotification, ctx context.Context) error {

	// 发送友评邻蒙结果通知通知
	if err := mess.SendWithContext([]int{u_id}, ctx); err != nil {
		util.Debug(" Cannot send accept notification", err)
		return err
	}
	// 记录用户有1新通知
	if err := data.AddUserNotificationCount(u_id); err != nil {
		util.Debug(" Cannot add user new-notification-count", err)
		return err
	}
	return nil
}
