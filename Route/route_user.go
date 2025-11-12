package route

import (
	"net/http"
	data "teachat/DAO"
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
		report(w, r, "你好，茶博士失魂鱼，未能读取用户信息.")
		return
	}
	var uB data.UserBean

	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	//没有id参数，报告uuid没有提及
	if uuid == "" {
		report(w, r, "你好，请提供需要查找茶友的识别码。")
		return
	}
	//有uuid参数，读取指定用户资料
	user, err := data.GetUserByUUID(uuid)
	if err != nil {
		report(w, r, "报告，大王，未能找到茶友的资料！")
		return
	}
	uB, err = fetchUserBean(user)
	if err != nil {
		util.Debug("Cannot get user Bean given uuid", user.Uuid, err)
		report(w, r, "你好，茶博士失魂鱼，未能读取用户信息.")
		return
	}

	// 准备页面数据
	uB.SessUser = s_u

	//检查登录者是否简介所有者本人？
	if user.Id == s_u.Id {
		//是本人,则打开编辑页
		uB.IsAuthor = true
		generateHTML(w, &uB, "layout", "navbar.private", "biography.private")
		return
	} else {
		//不是简介主人,打开公开介绍页
		uB.IsAuthor = false
		generateHTML(w, &uB, "layout", "navbar.private", "biography.public")
	}

}

// POST /user/edit
// 修改会员的花名和简介
func EditIntroAndName(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	} else {
		err = r.ParseForm()
		if err != nil {
			util.Debug("解析表单错误！", err)
		}
		s_u, _ := s.User()

		name := r.PostFormValue("name")
		len_biog := cnStrLen(name)
		if len_biog < 2 || len_biog > 16 {
			report(w, r, "茶博士彬彬有礼的说：过高人易妒，过洁世同嫌。名字也是噢。")
			return
		}
		s_u.Name = name

		biog := r.PostFormValue("biography")
		len_biog = cnStrLen(biog)
		if len_biog < 2 || len_biog > int(util.Config.ThreadMaxWord) {
			report(w, r, "茶博士彬彬有礼的说：简介不能为空, 云空未必空，欲洁何曾洁噢。")
			return
		}
		s_u.Biography = biog

		err = data.UpdateUserNameAndBiography(s_u.Id, s_u.Name, s_u.Biography)
		if err != nil {
			util.Debug(" 更新用户信息错误！", err)
			report(w, r, "茶博士失魂鱼，花名或者简介修改失败！")
			return
		}

		http.Redirect(w, r, "/v1/user/biography?uuid="+s_u.Uuid, http.StatusFound)
		//Report(w, r, "你好，茶博士低声说，花名或者简介更新成功啦。")

	}

}

// 处理用户头像相片
func UserAvatar(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		UploadAvatar(w, r)
	case http.MethodPost:
		SaveAvatar(w, r)
	}
}

// POST v1/user/avatar
// 处理用户头像相片
func SaveAvatar(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" 获取用户信息错误！", err)
		report(w, r, "你好，茶博士失魂鱼，未能读取用户信息！")
		return
	}
	// 处理上传到图片
	if ok := processUploadAvatar(w, r, s_u.Uuid); ok == nil {
		s_u.Avatar = s_u.Uuid
		s_u.UpdateAvatar()
		report(w, r, "茶博士微笑说头像修改成功。")
	}

}

// Get v1/user/avatar
// 编辑用户头像
func UploadAvatar(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" 获取用户信息错误！", err)
		report(w, r, "你好，茶博士失魂鱼，未能读取用户信息！")
		return
	}
	var lB data.LetterboxPageData
	lB.SessUser = s_u

	generateHTML(w, &lB, "layout", "navbar.private", "avatar.upload")
}

// GET
// update Password /Forgot Password?
func Forgot(w http.ResponseWriter, r *http.Request) {
	report(w, r, "尚未提供修改密码服务！")
}

// GET
// Reset Password?
func Reset(w http.ResponseWriter, r *http.Request) {
	report(w, r, "尚未提供重置密码服务！")
}

// GET /v1/users/connection_follow
func Follow(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "layout", "navbar.private", "connection.follow")
}
