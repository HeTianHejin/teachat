package route

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /LoginGet?footprint=xxx&query=xxx
// Show the LoginGet page
// 打开登录页面
func LoginGet(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
	}
	// 读取用户提交的，点击‘登船’时所在页面资料，以便checkin成功时回到原页面，改善体验
	footprint := r.FormValue("footprint")
	query := r.FormValue("query")
	var aopD data.AcceptObjectPageData
	aopD.SessUser.Footprint = footprint
	aopD.SessUser.Query = query

	_, err = session(r)
	if err != nil {
		generateHTML(w, &aopD, "layout", "navbar.public", "login")
		return
	}
	http.Redirect(w, r, "/v1/", http.StatusFound)
}

// GET /SignupGet
// 新用户注册页面
func SignupGet(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "layout", "navbar.public", "signup")
}

// POST /signup
// 注册新用户帐号
func SignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
	}
	s_u := data.UserUnknown
	// 读取用户提交的资料
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	biography := r.PostFormValue("biography")
	gender, err := strconv.Atoi(r.PostFormValue("gender"))
	if err != nil {
		report(w, s_u, "你好，请确认您的洗手间服务选择是否正确。")
		return
	}
	if gender != data.User_Gender_Female && gender != data.User_Gender_Male {
		report(w, s_u, "你好，请确认您的洗手间服务选择是否正确。")
		return
	}
	name := r.PostFormValue("name")
	//检查用户名称，不允许使用特殊字符，&,$,$,@...
	if ok_name := isValidUserName(name); !ok_name {
		report(w, s_u, "你好，请确认用户名是否只包含字母、数字、下划线或中文字符。")
		return
	}

	// 根据用户提交的资料填写新用户表格
	newU := data.User{
		Name:      name,
		Email:     email,
		Password:  data.Encrypt(password),
		Biography: biography,
		Role:      "traveller",
		Gender:    gender,
		Avatar:    "teaSet",
	}
	// 用正则表达式匹配一下提交的邮箱格式是否正确
	if ok_email := isEmail(newU.Email); !ok_email {
		report(w, s_u, "你好，请确认邮箱拼写是否正确。")
		return
	}
	// 检查提交的邮箱是否已经注册过了
	exist_email, err := data.UserExistByEmail(newU.Email)
	if err != nil {
		util.Debug((fmt.Errorf("检查邮箱存在性时出错: %v, 邮箱: %s", err, newU.Email)), "数据库查询错误")
		report(w, s_u, "你好，茶博士因找不到笔导致注册失败，请确认情况后重试。")
		return
	}
	if exist_email {
		util.Debug((fmt.Errorf("重复注册尝试: 邮箱 %s 已注册", newU.Email)), "重复注册")
		report(w, s_u, "你好，提交注册的邮箱地址已经注册,请确认后再试。")
		return
	}
	// 存储新用户（测试时不作邮箱有效性检查，直接激活账户）
	if err := newU.Create(); err != nil {
		util.Debug(" Cannot create user", err)
		report(w, s_u, "你好，粗鲁的茶博士因找不到笔导致注册失败，请确认情况后重试。")
		return
	}
	// 将新成员添加进默认的自由人茶团
	team_member := data.TeamMember{
		TeamId: data.TeamIdFreelancer,
		UserId: newU.Id,
		Role:   "taster",
		Status: data.TeMemberStatusActive,
	}
	if err = team_member.Create(); err != nil {
		util.Debug(" Cannot create default_free team_member", err)
		report(w, s_u, "你好，满头大汗的茶博士因找不到笔导致注册失败，请确认情况后重试。")
		return
	}
	//设置茶棚预设的默认团队（自由人）
	udt := data.UserDefaultTeam{
		UserId: newU.Id,
		TeamId: data.TeamIdFreelancer,
	}
	if err = udt.Create(); err != nil {
		util.Debug(" Cannot create default team", err)
		report(w, s_u, "你好，茶博士因摸不到超高度近视眼镜，导致注册失败，请确认情况后重试。")
		return
	}

	util.Debug(newU.Email, "注册新账号ok")

	t := ""
	if newU.Gender == data.User_Gender_Female {
		t = fmt.Sprintf("%s 女士，你好，注册成功！请登船，祝愿你拥有美好品茶时光。", newU.Name)
	} else {
		t = fmt.Sprintf("%s 先生，你好，注册成功！请登船，祝愿你拥有美好品茶时光。", newU.Name)
	}
	report(w, newU, t)

}

// POST /Authenticate
// Authenticate the user given the email and password
// 用户登录，并记录会话记录
func Authenticate(w http.ResponseWriter, r *http.Request) {
	s_u := data.UserUnknown
	err := r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，茶博士正在为你服务的路上努力，请稍安勿躁。")
		return
	}
	// 读取用户提交的资料
	watchword := r.PostFormValue("watchword")
	pw := r.PostFormValue("password")
	email := r.PostFormValue("email")

	s_u = data.User{}

	// 口令检查，提示用户这是茶话会
	wordValid := watchword == "闻香识茶" || watchword == "Recognizing Tea by Its Aroma"

	if wordValid {
		// 口令正确，获取用户信息
		if userID, convErr := strconv.Atoi(email); convErr == nil && userID > 0 {
			s_u, err = data.GetUser(userID)
			if err != nil {
				report(w, s_u, "茶博士嘀咕说，请确认握笔姿势是否正确，身形健美。")
				return
			}
		} else if isEmail(email) {
			s_u, err = data.GetUserByEmail(email, r.Context())
			if err != nil {
				report(w, s_u, "(嘀咕说) 请确保输入账号正确，握笔姿态优雅。")
				return
			}
		} else {
			report(w, s_u, "茶博士嘀咕说，请确认握笔姿势正确,而且身姿健美。")
			return
		}

		encryptedPw := data.Encrypt(pw)
		if subtle.ConstantTimeCompare([]byte(s_u.Password), []byte(encryptedPw)) == 1 {
			// 创建新的会话
			session, err := s_u.CreateSession()
			if err != nil {
				util.Debug(" Cannot create session", err)
				report(w, s_u, "你好，茶博士因找不到笔导致登船验证失败，请确认情况后重试。")
				return
			}
			// 设置cookie
			cookie := http.Cookie{
				Name:     "_cookie",
				Value:    session.Uuid,
				HttpOnly: true,
				MaxAge:   60 * 60 * 24 * 7, // 7 days
				//Secure:   true, // 生产环境应启用HTTPS
				SameSite: http.SameSiteLaxMode,
			}

			http.SetCookie(w, &cookie)

			// 安全重定向⚠️本站
			footprint := sanitizeRedirectPath(r.FormValue("footprint"))
			query := r.FormValue("query")

			// 安全构建URL，防止注入攻击
			redirectURL, err := url.Parse(footprint)
			if err != nil || query == "" {
				http.Redirect(w, r, footprint, http.StatusFound)
				return
			}
			redirectURL.RawQuery = query
			http.Redirect(w, r, redirectURL.String(), http.StatusFound)
			return

		} else {
			// 密码不匹配
			util.Debug(s_u.Email, "密码和用户名不匹配。")
			report(w, s_u, "无所事事的茶博士嘀咕说，请确认输入时姿势是否正确，键盘大小写灯是否有亮光？")
			return
		}

	} else {
		// 输入了错误的口令
		report(w, s_u, "你好，这是星际茶棚，想喝茶需要闻香识味噢，请确认再试。")
		return
	}

}

// GET /Logout
// Logs the user out by deleting server-side session and clearing client cookie
// 用户登出，清除服务端会话记录和客户端Cookie，并重定向到首页
func Logout(w http.ResponseWriter, r *http.Request) {
	const operation = "handler.Logout"

	// 1. 获取并验证Cookie
	cookie, err := r.Cookie("_cookie")
	if err != nil {
		if err != http.ErrNoCookie {
			util.Debug(operation, "获取Cookie失败", err)
		}
		// 无Cookie直接重定向
		http.Redirect(w, r, "/v1/", http.StatusFound)
		return
	}

	if cookie.Value == "" {
		util.Warning(operation, "空Cookie值")
		http.Redirect(w, r, "/v1/", http.StatusFound)
		return
	}

	// 2. 检查会话有效性
	sess := data.Session{Uuid: cookie.Value}
	valid, err := sess.Check()
	if err != nil {
		util.Debug(operation, "检查会话失败", "uuid", cookie.Value, "error", err)
		report(w, data.UserUnknown, "你好，茶博士因找不到资料导致登出失败，请确认情况后重试。")
		return
	}

	if !valid {
		util.Warning(operation, "无效会话", "uuid", cookie.Value)
		clearSessionCookie(w)
		http.Redirect(w, r, "/v1/", http.StatusFound)
		return
	}

	// 3. 删除会话
	if err := sess.Delete(); err != nil {
		util.Debug(operation, "删除会话失败", "uuid", cookie.Value, "error", err)
		report(w, data.UserUnknown, "你好，茶博士因找不到笔导致登出失败，请确认情况后重试。")
		return
	}

	// 4. 清除客户端Cookie并重定向
	clearSessionCookie(w)
	http.Redirect(w, r, "/v1/", http.StatusFound)
}

// clearSessionCookie 清除客户端会话Cookie
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "_cookie",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,   // 立即过期
		HttpOnly: true, // 防止XSS
		//Secure:   true, // 生产环境应启用HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}
