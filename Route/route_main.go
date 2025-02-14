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

	//读取最热的茶议dozen?
	num := 12

	// 读取最热的茶议
	thread_list, err := data.HotThreads(num)
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
		util.Warning(util.LogError(err), " Cannot read thread and author list")
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
		util.Warning(util.LogError(err), " Cannot read user info from session")
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
