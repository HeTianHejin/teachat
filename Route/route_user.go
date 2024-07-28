package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /user/Biography
// 展示用户个人主页
func Biography(w http.ResponseWriter, r *http.Request) {
	//检查是否已经登录
	s, err := util.Session(r)
	if err != nil {
		//打开登录页
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var uBP data.UserBiographyPageData

	vals := r.URL.Query()
	uuid := vals.Get("id")
	su, err := s.User()
	if err != nil {
		util.Warning(err, " 未能读取用户信息！")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取用户信息.")
		return
	}
	// //检查是否有id参数
	// if uuid == "" {
	// 	//没有id参数，打开当前浏览用户资料

	// 	uBP.SessUser = su
	// 	//检查登录者是否简介所有者本人？是则打开编辑页
	// 	if su.Id == s.UserId {
	// 		uBP.IsAuthor = true
	// 	} else {
	// 		uBP.IsAuthor = false
	// 	}
	// 	util.GenerateHTML(w, &uBP, "layout", "navbar.private", "biography.private")
	// 	return
	// }
	//有id参数，打开指定用户资料
	user, err := data.UserByUUID(uuid)
	if err != nil {
		util.Report(w, r, "报告，大王，未能找到大唐和尚的资料！")
		return
	}
	//检查登录者是否简介所有者本人？是则打开编辑页
	if user.Id == s.UserId {
		uBP.SessUser = su
		uBP.IsAuthor = true
		util.GenerateHTML(w, &uBP, "layout", "navbar.private", "biography.private")
		return
	} else {
		uBP.SessUser = su
		uBP.User = user
		uBP.IsAuthor = false
		//登录者不是简介主人,打开公开介绍页
		util.GenerateHTML(w, &uBP, "layout", "navbar.private", "biography.public")
	}

}

// POST /user/edit
// 修改会员的花名和简介
func EditIntroAndName(w http.ResponseWriter, r *http.Request) {
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	} else {
		err = r.ParseForm()
		if err != nil {
			util.Warning(err, " 解析表单错误！")
		}
		user, _ := sess.User()

		user.Name = r.PostFormValue("name")
		if user.Name == "" {
			util.Report(w, r, "茶博士彬彬有礼的说：用户名不能为空噢。")
			return
		}
		user.Biography = r.PostFormValue("biography")
		if user.Biography == "" {
			util.Report(w, r, "茶博士彬彬有礼的说：简介不能为空噢。")
			return
		}
		err = data.UpdateUserNameAndBiography(user.Id, user.Name, user.Biography)
		if err != nil {
			util.Warning(err, " 更新用户信息错误！")
			util.Report(w, r, "茶博士失魂鱼，花名或者简介修改失败！")
			return
		}
		//http.Redirect(w, r, "/v1/user/biography?id="+user.Uuid, http.StatusFound)
		util.Report(w, r, "您好，茶博士低声说，花名或者简介更新成功啦。")

	}

}

// 处理用户头像相片
func UserAvatar(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		UploadAvatar(w, r)
	case "POST":
		ProcessAvatar(w, r)
	}
}

// POST v1/user/avatar
// 处理用户头像相片
func ProcessAvatar(w http.ResponseWriter, r *http.Request) {
	s, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, err := s.User()
	if err != nil {
		util.Warning(err, " 获取用户信息错误！")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取用户信息！")
		return
	}
	// 处理上传到图片
	if ok := util.ProcessUploadAvatar(w, r, u.Uuid); ok == nil {
		u.Avatar = u.Uuid
		u.UpdateAvatar()
		util.Report(w, r, "茶博士微笑说头像修改成功。")
	}

}

// Get v1/user/avatar
// 编辑用户头像
func UploadAvatar(w http.ResponseWriter, r *http.Request) {
	_, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	util.GenerateHTML(w, nil, "layout", "navbar.private", "avatar.upload")
}

// GET
// update Password /Forgot Password?
func Forgot(w http.ResponseWriter, r *http.Request) {
	util.Report(w, r, "尚未提供修改密码服务！")
}

// GET
// Reset Password?
func Reset(w http.ResponseWriter, r *http.Request) {
	util.Report(w, r, "尚未提供重置密码服务！")
}

// GET /v1/users/connection_follow
func Follow(w http.ResponseWriter, r *http.Request) {
	util.GenerateHTML(w, nil, "layout", "navbar.private", "connection.follow")
}

// GET /v1/users/connection_fans
func Fans(w http.ResponseWriter, r *http.Request) {
	util.GenerateHTML(w, nil, "layout", "navbar.private", "connection.fans")
}

// GET /v1/users/connection_friend
func Friend(w http.ResponseWriter, r *http.Request) {
	s, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话读取当前用户信息
	u, _ := s.User()
	var cFPData data.ConnectionFriendPageData
	cFPData.SessUser = u

	util.GenerateHTML(w, &cFPData, "layout", "navbar.private", "connection.friend")
}
