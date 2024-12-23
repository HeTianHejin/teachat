package route

import (
	"database/sql"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// HandleSearch() 查询窗口 /v1/search
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
	//检查keyword的文字长度是否>1 and <32
	keyword_len := len(keyword)
	if keyword_len < 1 || keyword_len > 32 {
		Report(w, r, "你好，茶博士摸摸头，说关键词太长了记不住呢，请确认后再试。")
		return
	}

	//根据查询类型操作
	var fPD data.SearchPageData

	//初始化获取结果为零记录
	fPD.IsEmpty = true

	switch class_int {
	case 0:
		//查找茶友，user

		//用户可能提交了一个电子邮箱地址，如果是，我们需要先通过电子邮箱地址查找用户
		//检查keyword是否是电子邮箱地址
		if ok := IsEmail(keyword); ok {
			user, err := data.GetUserByEmail(keyword)
			if err != nil {
				util.Warning(err, keyword, " Cannot search user by keyword")
			}
			//如果user是非空
			if user.Id > 0 {
				user_bean, err := FetchUserBean(user)
				if err != nil {
					util.Warning(err, "cannot get user-bean given user")
				} else {
					fPD.UserBeanList = append(fPD.UserBeanList, user_bean)
					fPD.IsEmpty = false
				}
			}
		} else {

			user_list, err := data.SearchUserByNameKeyword(keyword)
			if err != nil {
				util.Warning(err, keyword, " Cannot search user by keyword")
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
	var f data.SearchPageData
	f.SessUser = s_u

	// 打开查询页面
	RenderHTML(w, &f, "layout", "navbar.private", "search")
}
