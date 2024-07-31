package route

import (
	"net/http"
	data "teachat/DAO"
)

// GET //First page
// 打开首页,展示最热门的茶议？
func Index(w http.ResponseWriter, r *http.Request) {
	var err error
	var indexPD data.IndexPageData
	//读取最热的茶议2dozen?
	num := 24
	indexPD.ThreadList, err = data.ThreadsIndex(num)
	if err != nil {
		Report(w, r, "您好，茶博士摸摸头，竟然惊讶地说茶语本被狗叼进花园里去了，请稍后再试。")
		return
	}
	if len(indexPD.ThreadList) == 0 {
		Report(w, r, "您好，茶博士摸摸头，说茶语本上落了片白茫茫大地真干净，请稍后再试。")
		return
	}
	// 迭代ThreadList，把.Body截取缩短108字符
	for i := range indexPD.ThreadList {
		indexPD.ThreadList[i].Body = Substr(indexPD.ThreadList[i].Body, 108)
	}
	s, err := Session(r)
	if err != nil {
		//游客
		indexPD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		for i := range indexPD.ThreadList {
			indexPD.ThreadList[i].PageData.IsAuthor = false
		}
		//展示游客首页
		GenerateHTML(w, &indexPD, "layout", "navbar.public", "index")
		return
	}
	//已登录
	sUser, err := s.User()
	if err != nil {
		Report(w, r, "您好，茶博士摸摸头，说有眼不识泰山。")
		return
	}
	indexPD.SessUser = sUser
	for i := range indexPD.ThreadList {
		if sUser.Id == indexPD.ThreadList[i].UserId {
			indexPD.ThreadList[i].PageData.IsAuthor = true
		} else {
			indexPD.ThreadList[i].PageData.IsAuthor = false
		}
	}
	//展示茶客的首页
	GenerateHTML(w, &indexPD, "layout", "navbar.private", "index")

}

// GET
// show About page 区别是导航条不同
func About(w http.ResponseWriter, r *http.Request) {
	var userBPD data.UserBiographyPageData
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
