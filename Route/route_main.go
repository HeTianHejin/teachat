package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET //First page
// 打开首页,展示最热门的茶议？
func Index(w http.ResponseWriter, r *http.Request) {
	var indexPD data.IndexPageData
	var tb_list []data.ThreadBean

	//读取最热的茶议2dozen?
	num := 24

	// 读取最热的茶议
	thread_list, err := data.ThreadsIndex(num)
	if err != nil {
		Report(w, r, "您好，茶博士摸摸头，竟然惊讶地说茶语本被狗叼进花园里去了，请稍后再试。")
		return
	}
	len := len(thread_list)

	if len == 0 {
		Report(w, r, "您好，茶博士摸摸头，说茶语本上落了片白茫茫大地真干净，请稍后再试。")
		return
	}

	tb_list, err = GetThreadBeanList(thread_list)
	if err != nil {
		util.Warning(err, " Cannot read thread and author list")
		Report(w, r, "您好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你忙碌中。")
		return
	}
	indexPD.ThreadBeanList = tb_list

	//是否登录
	s, err := Session(r)
	if err != nil {
		//游客
		indexPD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		for i := range thread_list {
			thread_list[i].PageData.IsAuthor = false
		}
		//展示游客首页
		GenerateHTML(w, &indexPD, "layout", "navbar.public", "index")
		return
	}
	//已登录
	sUser, err := s.User()
	if err != nil {
		util.Warning(err, " Cannot read user info from session")
		Report(w, r, "您好，茶博士摸摸头，说有眼不识泰山。")
		return
	}
	indexPD.SessUser = sUser
	for i := range thread_list {
		if sUser.Id == thread_list[i].UserId {
			thread_list[i].PageData.IsAuthor = true
		} else {
			thread_list[i].PageData.IsAuthor = false
		}
	}
	//展示茶客的首页
	GenerateHTML(w, &indexPD, "layout", "navbar.private", "index")

}

// GET
// show About page 区别是导航条不同
func About(w http.ResponseWriter, r *http.Request) {
	var userBPD data.UserBiography
	s, err := Session(r)
	if err != nil {
		userBPD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		GenerateHTML(w, &userBPD, "layout", "navbar.public", "about")

	}

	userBPD.SessUser, _ = s.User()
	//展示tea客的首页
	GenerateHTML(w, &userBPD, "layout", "navbar.private", "about")

}

// HandleSearch()
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		Search(w, r)
	case "POST":
		Fetch(w, r)
	}
}

// Post /v1/search
func Fetch(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "您好，茶博士摸摸头说，这个服务目前空缺，要不你来担当吧？")
}

// GET /v1/Search
// 打开查询页面
func Search(w http.ResponseWriter, r *http.Request) {
	var err error
	// 是否已登陆？
	_, err = Session(r)
	if err != nil {
		GenerateHTML(w, nil, "layout", "navbar.public", "search")
		return
	}
	// 查询页面
	GenerateHTML(w, nil, "layout", "navbar.private", "search")
}
