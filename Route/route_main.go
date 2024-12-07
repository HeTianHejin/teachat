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

	tb_list, err = GetThreadBeanList(thread_list)
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

	//根据查询类型操作
	var fPD data.FetchPageData
	switch class_int {
	case 0:
		//茶友，user
		user_list, err := data.SearchUserByNameKeyword(keyword)
		if err != nil {
			util.Warning(err, " Cannot search user with SearchUser().")
			Report(w, r, "你好，茶博士摸摸头，没有找到类似的茶友名称，请换个关键词再试。")
			return
		}
		fPD.UserBeanList, err = FetchUserBeanList(user_list)
		if err != nil {
			util.Warning(err, " Cannot read user info from session")
			Report(w, r, "你好，茶博士摸摸头，说有眼不识泰山。")
			return
		}

	case 10:
		//按user_id查询茶友
		keyword_int, err := strconv.Atoi(keyword)
		if err != nil {
			util.Warning(err, " Cannot convert keyword to int")
			Report(w, r, "你好，茶博士摸摸头，没有找到类似的茶友名称，请换个关键词再试。")
			return
		}
		user, err := data.GetUserById(keyword_int)
		if err != nil {
			if err == sql.ErrNoRows {
				Report(w, r, "你好，茶博士摸摸头，找不到这个茶友，请换个茶友号再试。")
				return
			}
			util.Warning(err, " Cannot search user with SearchUserByUserId().")
			Report(w, r, "你好，茶博士摸摸头，没有找到类似的茶友名称，请换个关键词再试。")
			return
		}
		userbean, err := FetchUserBean(user)
		if err != nil {
			util.Warning(err, " Cannot read user info from session")
			Report(w, r, "你好，茶博士摸摸头，说有眼不识泰山。")
			return
		}
		fPD.UserBeanList = append(fPD.UserBeanList, userbean)

	case 1:
		//茶议，thread
		// thread_list, err := data.SearchThread(keyword)
		// if err != nil {
		// 	util.Warning(err, " Cannot search thread with SearchThread().")
		Report(w, r, "你好，非常抱歉！负责这个检索作业的茶博士还没有出现，尚未能提供此服务。")
		// 	return
		// }
	default:
		Report(w, r, "你好，茶博士摸摸头，没有找到类似的茶语记录，请换个关键词再试。")
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
