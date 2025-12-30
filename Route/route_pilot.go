package route

import (
	"net/http"
	dao "teachat/DAO"
)

// GET /pilot/office
// 返回飞行员列表页面
func OfficePilot(w http.ResponseWriter, r *http.Request) {
	//获取session
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//获取用户信息
	s_u, err := sess.User()
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//判断用户角色是否为飞行员或者船长，如果是，则显示飞行员列表
	if s_u.Role == "pilot" || s_u.Role == "captain" {

		pilots := dao.User{}
		generateHTML(w, &pilots, "layout", "navbar.private", "pilot.office")
	}
	//如果不是，则显示错误信息
	report(w, s_u, "你好，欢迎光临茶博士服务室！")
}

// GET /pilot/new
// 返回添加飞行员页面
func NewPilot(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "layout", "navbar.private", "pilot.new")
}

// POST /pilot/add
// 添加一个飞行员
func AddPilot(w http.ResponseWriter, r *http.Request) {

	//判断用户角色是否为飞行员或者船长，如果是，则添加飞行员

	http.Redirect(w, r, "/v1/pilot/office", http.StatusFound)
}
