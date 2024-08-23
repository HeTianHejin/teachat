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
	s, err := Session(r)
	if err != nil {
		//打开登录页
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	su, err := s.User()
	if err != nil {
		util.Warning(err, " 未能读取用户信息！")
		Report(w, r, "您好，茶博士失魂鱼，未能读取用户信息.")
		return
	}
	var uBP data.UserBiography

	vals := r.URL.Query()
	uuid := vals.Get("id")
	//有id参数，读取指定用户资料
	user, err := data.UserByUUID(uuid)
	if err != nil {
		Report(w, r, "报告，大王，未能找到大唐和尚的资料！")
		return
	}
	// 准备页面数据
	uBP.SessUser = su
	team, err := user.GetLastDefaultTeam()
	if err != nil {
		util.Warning(err, " 未能读取用户团队信息！")
		Report(w, r, "报告，大王，找不到西天取经团队资料！")
		return
	}
	uBP.DefaultTeam, err = GetTeamBean(team)
	if err != nil {
		util.Warning(err, " 根据team未能读取用户团队信息")
		Report(w, r, "报告，大王，找不到西天取经团队资料！")
		return
	}
	//teams, err := data.GetSurvivalTeamsByUserId(user.Id)
	//检查登录者是否简介所有者本人？是则打开编辑页
	if user.Id == s.UserId {
		//uBP.User = data.User{}
		uBP.IsAuthor = true
		GenerateHTML(w, &uBP, "layout", "navbar.private", "biography.private")
		return
	} else {
		uBP.User = user
		uBP.IsAuthor = false
		//登录者不是简介主人,打开公开介绍页
		GenerateHTML(w, &uBP, "layout", "navbar.private", "biography.public")
	}

}

// POST /user/edit
// 修改会员的花名和简介
func EditIntroAndName(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	} else {
		err = r.ParseForm()
		if err != nil {
			util.Warning(err, "解析表单错误！")
		}
		sUser, _ := sess.User()

		n := r.PostFormValue("name")
		len := CnStrLen(n)
		if len < 2 || len > 12 {
			Report(w, r, "茶博士彬彬有礼的说：过高人易妒，过洁世同嫌。名字也是噢。")
			return
		}
		sUser.Name = n
		b := r.PostFormValue("biography")
		len = CnStrLen(b)
		if len < 2 || len > 66 {
			Report(w, r, "茶博士彬彬有礼的说：简介不能为空, 云空未必空，欲洁何曾洁噢。")
			return
		}
		sUser.Biography = b
		err = data.UpdateUserNameAndBiography(sUser.Id, sUser.Name, sUser.Biography)
		if err != nil {
			util.Warning(err, " 更新用户信息错误！")
			Report(w, r, "茶博士失魂鱼，花名或者简介修改失败！")
			return
		}
		//http.Redirect(w, r, "/v1/user/biography?id="+user.Uuid, http.StatusFound)
		Report(w, r, "您好，茶博士低声说，花名或者简介更新成功啦。")

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
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, err := s.User()
	if err != nil {
		util.Warning(err, " 获取用户信息错误！")
		Report(w, r, "您好，茶博士失魂鱼，未能读取用户信息！")
		return
	}
	// 处理上传到图片
	if ok := ProcessUploadAvatar(w, r, u.Uuid); ok == nil {
		u.Avatar = u.Uuid
		u.UpdateAvatar()
		Report(w, r, "茶博士微笑说头像修改成功。")
	}

}

// Get v1/user/avatar
// 编辑用户头像
func UploadAvatar(w http.ResponseWriter, r *http.Request) {
	_, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	GenerateHTML(w, nil, "layout", "navbar.private", "avatar.upload")
}

// GET
// update Password /Forgot Password?
func Forgot(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "尚未提供修改密码服务！")
}

// GET
// Reset Password?
func Reset(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "尚未提供重置密码服务！")
}

// GET /v1/users/connection_follow
func Follow(w http.ResponseWriter, r *http.Request) {
	GenerateHTML(w, nil, "layout", "navbar.private", "connection.follow")
}

// GET /v1/users/connection_fans
func Fans(w http.ResponseWriter, r *http.Request) {
	GenerateHTML(w, nil, "layout", "navbar.private", "connection.fans")
}

// GET /v1/users/connection_friend
func Friend(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话读取当前用户信息
	u, _ := s.User()
	var cFPData data.ConnectionFriendPageData
	cFPData.SessUser = u

	GenerateHTML(w, &cFPData, "layout", "navbar.private", "connection.friend")
}
