package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /pilot/office
// 返回飞行员列表页面
func OfficePilot(w http.ResponseWriter, r *http.Request) {
	//获取session
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//获取用户信息
	user, err := sess.User()
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//判断用户角色是否为飞行员或者船长，如果是，则显示飞行员列表
	if user.Role == "pilot" || user.Role == "captain" {

		pilots, err := data.GetAdministrators()
		if err != nil {
			util.Danger(util.LogError(err), " Cannot get pilots")
			http.Redirect(w, r, "/v1/", http.StatusFound)
		}
		RenderHTML(w, &pilots, "layout", "navbar.private", "pilot.office")
	}
	//如果不是，则显示错误信息
	Report(w, r, "你好，欢迎光临茶博士服务室！")
}

// GET /pilot/new
// 返回添加飞行员页面
func NewPilot(w http.ResponseWriter, r *http.Request) {
	RenderHTML(w, nil, "layout", "navbar.private", "pilot.new")
}

// POST /pilot/add
// 添加一个飞行员
func AddPilot(w http.ResponseWriter, r *http.Request) {
	//获取session
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//获取用户信息
	user, err := sess.User()
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//判断用户角色是否为飞行员或者船长，如果是，则添加飞行员
	if user.Role == "pilot" || user.Role == "captain" {

		err = r.ParseForm()
		if err != nil {
			util.Danger(util.LogError(err), " Cannot parse form")
			http.Redirect(w, r, "/v1/pilot/new", http.StatusFound)
			return
		}
		//获取新飞行员的用户信息，并添加到数据库中
		email := r.PostFormValue("email")
		newPilot, err := data.GetUserByEmail(email)
		if err != nil {
			util.Danger(util.LogError(err), " Cannot find user by email")
			http.Redirect(w, r, "/v1/pilot/new", http.StatusFound)
			return
		}
		passw := r.PostFormValue("password")
		pilot := data.Administrator{
			UserId:        newPilot.Id,
			Role:          "pilot",
			Password:      passw,
			InvalidReason: "?",
		}
		//创建飞行员，并返回错误信息
		if err := pilot.Create(); err != nil {
			util.Danger(util.LogError(err), " Cannot create pilot")
			http.Redirect(w, r, "/v1/pilot/new", http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "/v1/pilot/office", http.StatusFound)
}
