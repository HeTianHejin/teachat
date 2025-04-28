package route

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// HandleSearch() 查询窗口 /v1/search
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SearchGet(w, r)
	case http.MethodPost:
		SearchPost(w, r)
	default:
		//其他方法，不允许
		// return error
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

// POST /v1/search
// 处理用户提交的查询（参数）方法
func SearchPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Error(" Cannot parse form", err)
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
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	//读取查询参数
	class_str := r.PostFormValue("class")
	//转换class_str为int
	class_int, err := strconv.Atoi(class_str)
	if err != nil {
		util.Error("Cannot convert class_str to int", err)
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
				util.Error(keyword, " Cannot search user by keyword")
			}
			//如果user是非空
			if user.Id > 0 {
				user_bean, err := FetchUserBean(user)
				if err != nil {
					util.Error("cannot get user-bean given user", err)
				} else {
					fPD.UserBeanSlice = append(fPD.UserBeanSlice, user_bean)
					fPD.IsEmpty = false
				}
			}
		} else {

			user_slice, err := data.SearchUserByNameKeyword(keyword)
			if err != nil {
				util.Error(keyword, " Cannot search user by keyword")
			}

			if len(user_slice) >= 1 {
				fPD.UserBeanSlice, err = FetchUserBeanSlice(user_slice)
				if err != nil {
					util.Error(" Cannot fetch user bean slice given user_slice", err)
				}
				if len(fPD.UserBeanSlice) >= 1 {
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
		user, err := data.GetUser(keyword_int)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				//				util.Error(" Cannot get user by keyword_int id", err)
			}
		}

		//如果user是非空
		if user.Id > 0 {
			userbean, err := FetchUserBean(user)
			if err != nil {
				util.Error("cannot get user-bean given user", err)
			} else {
				fPD.UserBeanSlice = append(fPD.UserBeanSlice, userbean)
				fPD.IsEmpty = false
			}
		}

	case 1:
		//查询，茶团简称，team.abbreviation
		team_slice, err := data.SearchTeamByAbbreviation(keyword)
		if err != nil {
			util.Error(" Cannot search team by abbreviation", err)
		}

		if len(team_slice) >= 1 {
			t_b_slice, err := FetchTeamBeanSlice(team_slice)
			if err != nil {
				util.Error(" Cannot fetch team bean slice given team_slice", err)
			}
			if len(t_b_slice) >= 1 {
				fPD.Count = len(t_b_slice)
				fPD.TeamBeanSlice = t_b_slice
				fPD.IsEmpty = false
			}
		}
	case 5:
		//查询品茶地点 place
		place_slice, err := data.FindPlaceByName(keyword)
		if err != nil {
			util.Error(" Cannot search place by keyword", err)
		}
		if len(place_slice) >= 1 {
			fPD.Count = len(place_slice)
			fPD.PlaceSlice = place_slice
			fPD.IsEmpty = false
		}

	default:
		Report(w, r, "你好，茶博士摸摸头，还没有开放这种类型的查询功能，请换个查询类型再试。")
		return
	}
	fPD.SessUser = s_u

	RenderHTML(w, &fPD, "layout", "navbar.private", "search")
}

// GET /v1/SearchGet
// 打开查询页面
func SearchGet(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var f data.SearchPageData
	f.SessUser = s_u

	// 打开查询页面
	RenderHTML(w, &f, "layout", "navbar.private", "search")
}
