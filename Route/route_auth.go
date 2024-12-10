package route

import (
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /Login?footprint=xxx&query=xxx
// Show the Login page
// 打开登录页面
func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "你好，闪电考拉正在为你服务的路上极速行动，请稍安勿躁。")
		return
	}
	// 读取用户提交的，点击‘登机’时所在页面资料，以便checkin成功时回到原页面，改善体验
	footprint := r.FormValue("footprint")
	query := r.FormValue("query")
	var aopD data.AcceptObjectPageData
	aopD.SessUser.Footprint = footprint
	aopD.SessUser.Query = query
	// aopD.SessUser = data.User{
	// 	Id:   0,
	// 	Name: "游客",
	// }
	_, err = Session(r)
	if err != nil {
		// t := ParseTemplateFiles("layout", "navbar.public", "login")
		// t.Execute(w, nil)
		RenderHTML(w, &aopD, "layout", "navbar.public", "login")
		return
	}
	http.Redirect(w, r, "/v1/", http.StatusFound)
}

// GET /SignupForm
// 新用户注册页面
func SignupForm(w http.ResponseWriter, r *http.Request) {
	RenderHTML(w, nil, "layout", "navbar.public", "signup")
}

// POST /signup
// 注册新用户帐号
func SignupAccount(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
	}
	// 读取用户提交的资料
	name := r.PostFormValue("name")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	biography := r.PostFormValue("biography")
	gender, err := strconv.Atoi(r.PostFormValue("gender"))
	if err != nil {
		Report(w, r, "你好，请确认您的洗手间服务选择是否正确。")
		return
	}
	if gender != 0 && gender != 1 {
		Report(w, r, "你好，请确认您的洗手间服务选择是否正确。")
		return
	}

	// 根据用户提交的资料填写新用户表格
	newU := data.User{
		Name:      name,
		Email:     email,
		Password:  password,
		Biography: biography,
		Role:      "traveller",
		Gender:    gender,
		Avatar:    "teaSet",
	}
	// 用正则表达式匹配一下提交的邮箱格式是否正确
	if ok := VerifyEmailFormat(newU.Email); !ok {
		Report(w, r, "你好，请确认邮箱拼写是否正确。")
		return
	}
	// 检查提交的邮箱是否已经注册过了
	exist, _ := data.UserExistByEmail(newU.Email)
	if exist {
		util.Warning(newU.Email, "提交注册的邮箱地址已经注册。")
		Report(w, r, "你好，提交注册的邮箱地址已经注册,请确认后再试。")
		return
	}
	// 存储新用户（测试时不作邮箱有效性检查，直接激活账户）
	if err := newU.Create(); err != nil {
		util.Danger(err, " Cannot create user")
		Report(w, r, "你好，粗鲁的茶博士因找不到笔导致注册失败，请确认情况后重试。")
		return
	}
	// 将新成员添加进默认的自由人茶团
	team_member := data.TeamMember{
		TeamId: 2,
		UserId: newU.Id,
		Role:   "taster",
		Class:  1,
	}
	if err = team_member.Create(); err != nil {
		util.Danger(err, " Cannot create default_free team_member")
		Report(w, r, "你好，满头大汗的茶博士因找不到笔导致注册失败，请确认情况后重试。")
		return
	}
	//设置默认团队
	udt := data.UserDefaultTeam{
		UserId: newU.Id,
		TeamId: 2,
	}
	if err = udt.Create(); err != nil {
		util.Danger(err, " Cannot create default team")
		Report(w, r, "你好，满头大汗的茶博士因摸索不到近视眼镜，导致注册失败，请确认情况后重试。")
		return
	}

	//util.Info(newU.Email, "注册新账号ok")
	t := fmt.Sprintf("%s 你好，注册成功！祝愿你拥有美好品茶时光。", newU.Name)
	Report(w, r, t)

}

// POST /Authenticate
// Authenticate the user given the email and password
// 用户登录，并记录会话记录
func Authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "你好，闪电考拉正在为你服务的路上极速行动，请稍安勿躁。")
		return
	}
	// 读取用户提交的资料
	footprint := r.FormValue("footprint")
	query := r.FormValue("query")
	// 增加口令检查，提示用户这是茶话会
	watchword := r.PostFormValue("watchword")
	pw := r.PostFormValue("password")

	wordValid, err := data.CheckWatchword(watchword)
	if err != nil {
		Report(w, r, "茶博士嘀咕说，今天吃鸡？不是来喝茶吗？")
		return
	} else {
		if wordValid {
			// 口令正确
			var s_u data.User
			email := r.PostFormValue("email")
			// Check if the email parameter is a positive integer (user ID)
			if s_u_id, err := strconv.Atoi(email); err == nil && s_u_id > 0 {
				// Retrieve user by ID
				s_u, err = data.UserById(s_u_id)
				if err != nil {
					Report(w, r, "茶博士嘀咕说，请确认握笔姿势是否正确，身形健美。")
					return
				}
			} else if VerifyEmailFormat(email) {
				// Retrieve user by email
				s_u, err = data.UserByEmail(email)
				if err != nil {
					util.Warning(err, email, "cannot get user given email")
					Report(w, r, "(嘀咕说) 请确保输入账号正确，握笔姿态优雅。")
					return
				}
			} else {
				// Invalid email format
				Report(w, r, "茶博士嘀咕说，请确认握笔姿势正确,而且身形健美")
				return
			}

			if s_u.Password == data.Encrypt(pw) {
				//util.Info(user.Email, "密码匹配成功")

				//创建新的会话
				session, err := s_u.CreateSession()
				if err != nil {
					util.Warning(err, " Cannot create session")
					return
				}
				//设置cookie
				cookie := http.Cookie{
					Name:     "_cookie",
					Value:    session.Uuid,
					HttpOnly: true,
				}

				http.SetCookie(w, &cookie)

				//读取足迹，以决定返回那一个页面
				if footprint == "" {
					footprint = "/v1/"
				} else {

					footprint = footprint + "?" + query
				}
				http.Redirect(w, r, footprint, http.StatusFound)
				return
			} else {
				//密码和用户名不匹配?
				//如果连续输错密码，需要采取一些防暴力冲击措施！！！
				util.Warning(s_u.Email, "密码和用户名不匹配。")
				Report(w, r, "无所事事的茶博士嘀咕说，请确认输入时姿势是否正确，键盘大小写灯是否有亮光？")
				return
			}

		} else {
			//输入了错误的口令
			Report(w, r, "你好，这是星际茶棚，想喝茶需要闻香识味噢，请确认再试。")
			return
		}

	}

}

// GET /Logout
// Logs the user out
// 用户登出，清除会话记录，并返回首页，并记录用户登出成功。
func Logout(w http.ResponseWriter, r *http.Request) {
	//读取会话 绿豆饼
	cookie, err := r.Cookie("_cookie")
	if err != http.ErrNoCookie {
		// 根据绿豆饼中的关键馅料获取库中的预留的饼印
		sess := data.Session{Uuid: cookie.Value}
		//查询一下会话资料，就是核对一下饼和饼印是否ok（有预留一致而且没过期）
		ok, err := sess.Check()
		if ok {
			//记录一下登出的用户邮箱
			err = sess.Delete()
			if err != nil {
				util.Danger(err, sess.Email, "Failed to delete session")
			}
			// else {
			//会话清除后的提示信息
			//util.Info(sess.Email, "Session deleted")
			//}
		} else {
			util.Danger(err, sess.Email, " 登出时会话资料查询失败")
		}

	}
	http.Redirect(w, r, "/v1/", http.StatusFound)

}
