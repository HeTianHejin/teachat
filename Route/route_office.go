package route

import (
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
	s_u, s_d_family, s_all_families, s_d_team, s_survival_teams, s_d_place, s_places, err := fetchSessionUserRelatedData(s)
	if err != nil {
		util.Debug(" Cannot fetch user related data", err)
		report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	//读取被邀请用户的相关资料
	user_uuid := r.FormValue("id")
	invi_user, err := data.GetUserByUUID(user_uuid)
	if err != nil {
		util.Debug(" Cannot get user by uuid", err)
		report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	var iD data.InvitationDetail
	// 填写页面资料
	iD.SessUser = s_u
	iD.SessUserDefaultFamily = s_d_family
	iD.SessUserAllFamilies = s_all_families
	iD.SessUserDefaultTeam = s_d_team
	iD.SessUserSurvivalTeams = s_survival_teams
	iD.SessUserDefaultPlace = s_d_place
	iD.SessUserBindPlaces = s_places

	iD.InvitationBean.InviteUser = invi_user

	renderHTML(w, &iD, "layout", "navbar.private", "pilot.invite")

}

// 向2个非当前用户发送蒙评审核消息
func TwoAcceptMessagesSendExceptUserId(u_id int, mess data.AcceptMessage) error {
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
	// 发送“是否接纳”消息
	if err = mess.Send(user_ids); err != nil {
		util.Debug(" Cannot send accept message", err)
		return err
	}
	// 记录用户有1新消息
	for _, u_id := range user_ids {
		if err = data.AddUserMessageCount(u_id); err != nil {
			util.Debug(" Cannot add random user new-message-count", err)
			return err
		}
	}
	return nil
}

// 向当前用户发送友邻蒙评结果通知消息
func PilotAcceptMessageSend(u_id int, mess data.AcceptMessage) error {

	// 发送友评邻蒙结果通知消息
	if err := mess.Send([]int{u_id}); err != nil {
		util.Debug(" Cannot send accept message", err)
		return err
	}
	// 记录用户有1新消息
	if err := data.AddUserMessageCount(u_id); err != nil {
		util.Debug(" Cannot add user new-message-count", err)
		return err
	}
	return nil
}
