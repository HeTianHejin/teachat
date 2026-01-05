package route

import (
	"context"
	"net/http"
	dao "teachat/DAO"
	util "teachat/Util"
)

// GET /office/draftThread
// 重新激活茶语草稿进入邻桌蒙评流程
func ActivateDraftThread(w http.ResponseWriter, r *http.Request) {

}

// 向2个非当前用户发送蒙评审核通知
func TwoAcceptNotificationsSendExceptUserId(u_id int, mess dao.AcceptNotification, ctx context.Context) error {
	var user_ids []int
	var err error
	// if dao.UserCount() < 50 {

	// 	if user_ids, err = dao.Get2RandomUserId(); err != nil {
	// 		util.PanicTea(util.LogError(err), " Cannot get 2 random user id")
	// 		return err
	// 	}
	// } else {
	// 	if dao.SessionCount() > 12 && dao.SessionCount() < 250 {
	// 		// 在线不同性别随机
	// 		user_ids, err = dao.Get2GenderRandomSUserIdExceptId(u_id)
	// 		if err != nil {
	// 			util.PanicTea(util.LogError(err), " Cannot get 2 gender random sess_user id")
	// 			user_ids, err = dao.Get2GenderRandomUserIdExceptId(u_id)
	// 			if err != nil {
	// 				util.PanicTea(util.LogError(err), " Cannot get 2 gender random user id")
	// 				// 在线不分性别随机
	// 				user_ids, err = dao.Get2RandomSUserIdExceptId(u_id)
	// 				if err != nil {
	// 					util.PanicTea(util.LogError(err), " Cannot get 2 random sess_user id")
	// 					return err
	// 				}
	// 			}
	// 		}
	// 	} else {
	if user_ids, err = dao.Get2RandomUserId(); err != nil {
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
		if err = dao.AddUserNotificationCount(u_id); err != nil {
			util.Debug(" Cannot add random user new-notification-count", err)
			return err
		}
	}
	return nil
}

// 向当前用户发送友邻蒙评结果通知通知
func AcceptNotificationSend(u_id int, mess dao.AcceptNotification, ctx context.Context) error {

	// 发送友评邻蒙结果通知通知
	if err := mess.SendWithContext([]int{u_id}, ctx); err != nil {
		util.Debug(" Cannot send accept notification", err)
		return err
	}
	// 记录用户有1新通知
	if err := dao.AddUserNotificationCount(u_id); err != nil {
		util.Debug(" Cannot add user new-notification-count", err)
		return err
	}
	return nil
}
