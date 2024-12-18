package route

import (
	"database/sql"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET //First page
// 打开首页,展示最热门的茶议？
func Index(w http.ResponseWriter, r *http.Request) {
	var indexPD data.IndexPageData
	var tb_list []data.ThreadBean

	//读取最热的茶议dozen?
	num := 12

	// 读取最热的茶议
	thread_list, err := data.ThreadsIndex(num)
	if err != nil {
		Report(w, r, "你好，茶博士摸摸头，竟然惊讶地说茶语本被狗叼进花园里去了，请稍后再试。")
		return
	}
	len := len(thread_list)

	if len == 0 {
		Report(w, r, "你好，茶博士摸摸头，说茶语本上落了片白茫茫大地真干净，请稍后再试。")
		return
	}

	tb_list, err = FetchThreadBeanList(thread_list)
	if err != nil {
		util.Warning(err, " Cannot read thread and author list")
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你忙碌中。")
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
		RenderHTML(w, &indexPD, "layout", "navbar.public", "index")
		return
	}
	//已登录
	sUser, err := s.User()
	if err != nil {
		util.Warning(err, " Cannot read user info from session")
		Report(w, r, "你好，茶博士摸摸头，说有眼不识泰山。")
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
	RenderHTML(w, &indexPD, "layout", "navbar.private", "index")

}

// GET
// show About page 区别是导航条不同
func About(w http.ResponseWriter, r *http.Request) {
	var uB data.UserBean
	sess, err := Session(r)
	if err != nil {
		//游客
		uB.IsAuthor = false
		uB.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		RenderHTML(w, &uB, "layout", "navbar.public", "about")
		return
	} else {
		//已登录
		uB.SessUser, _ = sess.User()
		//展示tea客的首页
		RenderHTML(w, &uB, "layout", "navbar.private", "about")
	}

}

// FUNC 查询窗口 /v1/search
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		Search(w, r)
	case "POST":
		Fetch(w, r)
	}
}

// POST /v1/search
// 处理用户提交的查询（参数）方法
func Fetch(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "你好，茶博士失魂鱼，未能理解你的话语，请稍后再试。")
		return
	}
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	//读取查询参数
	class_str := r.PostFormValue("class")
	//转换class_str为int
	class_int, err := strconv.Atoi(class_str)
	if err != nil {
		util.Info(err, "Cannot convert class_str to int")
		Report(w, r, "你好，茶博士摸摸头，说茶语本上落了片白茫茫大地真干净，请稍后再试。")
		return
	}

	keyword := r.PostFormValue("keyword")
	//检查keyword的文字长度是否>1 and <17
	keyword_len := len(keyword)
	if keyword_len < 1 || keyword_len > 17 {
		Report(w, r, "你好，茶博士摸摸头，说关键词太长了记不住呢，请确认后再试。")
		return
	}

	//根据查询类型操作
	var fPD data.FetchPageData
	fPD.IsEmpty = true

	switch class_int {
	case 0:
		//茶友，user
		user_list, err := data.SearchUserByNameKeyword(keyword)
		if err != nil {
			util.Warning(err, " Cannot search user by keyword")
		}

		if len(user_list) >= 1 {
			fPD.UserBeanList, err = FetchUserBeanList(user_list)
			if err != nil {
				util.Info(err, " Cannot fetch user bean list given user_list")
			}
			if len(fPD.UserBeanList) >= 1 {
				fPD.IsEmpty = false
			}
		}

	case 10:
		//按user_id查询茶友
		keyword_int, err := strconv.Atoi(keyword)
		if err != nil {
			Report(w, r, "你好，茶博士摸摸头，看不懂提交的茶友号，请换个关键词再试。")
			return
		}
		user, err := data.GetUserById(keyword_int)
		if err != nil {
			if err != sql.ErrNoRows {
				util.Warning(err, " Cannot get user by keyword_int id")
			}
		}

		//如果user是非空
		if user.Id > 0 {
			userbean, err := FetchUserBean(user)
			if err != nil {
				util.Warning(err, "cannot get user-bean given user")
			} else {
				fPD.UserBeanList = append(fPD.UserBeanList, userbean)
				fPD.IsEmpty = false
			}
		}

	case 1:
		//查询，茶团简称，team.abbreviation
		team_list, err := data.SearchTeamByAbbreviation(keyword)
		if err != nil {
			util.Warning(err, " Cannot search team by abbreviation")
		}

		if len(team_list) >= 1 {
			t_b_list, err := FetchTeamBeanList(team_list)
			if err != nil {
				util.Warning(err, " Cannot fetch team bean list given team_list")
			}
			if len(t_b_list) >= 1 {
				fPD.TeamBeanList = t_b_list
				fPD.IsEmpty = false
			}
		}

	default:
		Report(w, r, "你好，茶博士摸摸头，还没有开放这种类型的查询功能，请换个查询类型再试。")
		return
	}
	fPD.SessUser = s_u

	RenderHTML(w, &fPD, "layout", "navbar.private", "search")
}

// GET /v1/Search
// 打开查询页面
func Search(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var f data.FetchPageData
	f.SessUser = s_u

	// 打开查询页面
	RenderHTML(w, &f, "layout", "navbar.private", "search")
}
