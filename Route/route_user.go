package route

import (
	"net/http"
	dao "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/user/biography?uuid=
// 展示用户个人主页
func Biography(w http.ResponseWriter, r *http.Request) {
	//检查是否已经登录
	s, err := session(r)
	if err != nil {
		//打开登录页
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	s_u, err := s.User()
	if err != nil {
		util.Debug(" 根据会话未能读取用户信息", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取用户信息.")
		return
	}
	var uB dao.UserDefaultDataBean

	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	//没有id参数，报告uuid没有提及
	if uuid == "" {
		report(w, s_u, "你好，请提供茶友的识别码。")
		return
	}
	//有uuid参数，读取指定用户资料
	user, err := dao.GetUserByID(uuid)
	if err != nil {
		util.Debug("Cannot get user given uuid", uuid, err)
		report(w, s_u, "报告，大王，未能找到茶友的资料！")
		return
	}
	uB, err = fetchUserDefaultDataBeanForBiography(user)
	if err != nil {
		util.Debug("Cannot get user Bean given uuid", user.Uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取用户信息.")
		return
	}

	// 准备页面数据
	uB.SessUser = s_u

	//检查登录者是否简介所有者本人？
	if user.Id == s_u.Id {
		//是本人,则打开编辑页
		uB.IsAuthor = true
		generateHTML(w, &uB, "layout", "navbar.private", "user.biography.private")
		return
	} else {
		//不是简介主人,打开公开介绍页
		generateHTML(w, &uB, "layout", "navbar.private", "user.biography.public")
	}

}

// POST /user/edit
// 修改会员的花名和简介
func EditIntroAndName(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("fail to fetch user by session", err)
		report(w, s_u, "读取用户资料出现意外故事，请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("fail parse form", err)
		report(w, s_u, "茶博士失魂鱼，读取表单出现意外故事。")
	}

	//测试时期暂时不允许改名字
	// name := r.PostFormValue("name")
	// len_name := cnStrLen(name)
	// if len_name < 2 || len_name > 16 {
	// 	report(w, s_u, "茶博士彬彬有礼的说：名字太短不够帅，太长了墨水都不够用噢。")
	// 	return
	// }
	// if isValidUserName(name) {
	// 	report(w, s_u, "你好，请勿使用特殊字符作为名称呢，将来登机称帝，都不知道如何高呼陛下万岁。")
	// 	return
	// }
	// newName := name

	biog := r.PostFormValue("biography")
	len_biog := cnStrLen(biog)
	if len_biog < 2 || len_biog > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "茶博士彬彬有礼的说：简介不能为空, 云空未必空，欲洁何曾洁噢。")
		return
	}
	newBiography := biog

	err = dao.UserUpdateBiography(s_u.Id, newBiography)
	if err != nil {
		util.Debug(" 更新用户信息错误！", err)
		report(w, s_u, "茶博士失魂鱼，请问你刚刚说的花名或者简介是什么来着？")
		return
	}

	http.Redirect(w, r, "/v1/user/biography?uuid="+s_u.Uuid, http.StatusFound)
	//Report(w, r, "你好，茶博士低声说，花名或者简介更新成功啦。")

}

// AvatarUploadUser() v1/user/avatar
// 处理用户头像相片
func AvatarUploadUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		avatarUploadUserGet(w, r)
	case http.MethodPost:
		avatarUploadUserPost(w, r)
	}
}

// POST v1/user/avatar
// 处理用户头像相片
func avatarUploadUserPost(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" 获取用户信息错误！", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取用户信息！")
		return
	}
	// 处理上传到图片
	errAvatar := saveUploadAvatar(r, s_u.Uuid, "user")
	if errAvatar == nil {
		s_u.Avatar = s_u.Uuid
		if err = s_u.UpdateAvatar(); err != nil {
			util.Debug("fail to update user avatar", err)
			report(w, s_u, "您好，请问你刚刚说的喜欢什么类型的音乐，就为你播放？")
			return
		}
		report(w, s_u, "茶博士微笑说，头像修改成功。")
	} else {
		report(w, s_u, "图片上传失败：%s", errAvatar)
	}

}

// Get v1/user/avatar
// 编辑用户头像
func avatarUploadUserGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" 获取用户信息错误！", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取用户信息！")
		return
	}
	var lB dao.LetterboxPageData
	lB.SessUser = s_u

	generateHTML(w, &lB, "layout", "navbar.private", "avatar.upload")
}

// GET
// update Password /Forgot Password?
func Forgot(w http.ResponseWriter, r *http.Request) {
	report(w, dao.UserUnknown, "尚未提供修改密码服务！")
}

// GET
// Reset Password?
func Reset(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// GET /v1/users/connection_follow
func Follow(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "layout", "navbar.private", "connection.follow")
}
