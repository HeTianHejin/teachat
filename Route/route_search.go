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
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能理解你的话语，请稍后再试。")
		return
	}
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	//读取查询参数
	class_str := r.PostFormValue("class")
	//转换class_str为int
	class_int, err := strconv.Atoi(class_str)
	if err != nil {
		util.Debug("Cannot convert class_str to int", err)
		report(w, r, "你好，茶博士摸摸头，说茶语本上落了片白茫茫大地真干净，请稍后再试。")
		return
	}

	keyword := r.PostFormValue("keyword")
	//检查keyword的文字长度是否>1 and <32
	keyword_len := len(keyword)
	if keyword_len < 1 || keyword_len > 32 {
		report(w, r, "你好，茶博士摸摸头，说关键词太长了记不住呢，请确认后再试。")
		return
	}

	var fPD data.SearchPageData
	fPD.SessUser = s_u
	//初始化获取结果为零记录
	fPD.IsEmpty = true

	//根据查询类型操作
	switch class_int {
	case data.SearchTypeUserNameOrEmail:
		//按花名或者邮箱查找茶友，user

		//用户可能提交了一个电子邮箱地址，如果是，我们需要先通过电子邮箱地址查找用户
		//检查keyword是否是电子邮箱地址
		if ok := isEmail(keyword); ok {
			user, err := data.GetUserByEmail(keyword, r.Context())
			if err != nil {
				util.Debug(keyword, " Cannot search user by keyword", err)
				report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
			//如果user是非空
			if user.Id > 0 {
				user_bean, err := fetchUserBean(user)
				if err != nil {
					util.Debug("cannot get user-bean given user", err)
					report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
					return
				} else {
					fPD.UserBeanSlice = append(fPD.UserBeanSlice, user_bean)
					fPD.IsEmpty = false
				}
			}
		} else {
			user_slice, err := data.SearchUserByNameKeyword(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
			if err != nil {
				util.Debug(" Cannot search user by keyword", err)
				report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}

			if len(user_slice) >= 1 {
				fPD.UserBeanSlice, err = fetchUserBeanSlice(user_slice)
				if err != nil {
					util.Debug(" Cannot fetch user bean slice given user_slice", err)
					report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
					return
				}
				fPD.IsEmpty = false
			}
		}
		renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_avatar_name_gender")
		return

	case data.SearchTypeUserId:
		//按user_id查询茶友
		// 验证关键词是否为自然数
		keyword_int, err := strconv.Atoi(keyword)
		if err != nil || keyword_int <= 0 {
			report(w, r, "茶友号必须是正整数")
			return
		}
		user, err := data.GetUser(keyword_int)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fPD.IsEmpty = true
				renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_avatar_name_gender")
				return
			} else {
				util.Debug("failed to get user given user_id: ", keyword_int, err)
				report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
		}

		//如果user是非空
		if user.Id > 0 {
			userbean, err := fetchUserBean(user)
			if err != nil {
				util.Debug("cannot get user-bean given user", err)
				report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			} else {
				fPD.UserBeanSlice = append(fPD.UserBeanSlice, userbean)
				fPD.IsEmpty = false
			}
		}
		renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_avatar_name_gender")
		return
	case data.SearchTypeTeamAbbr:
		//查询，茶团简称，team.abbreviation
		team_slice, err := data.SearchTeamByAbbreviation(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" Cannot search team by abbreviation", err)
			report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
			return
		}

		if len(team_slice) >= 1 {
			t_b_slice, err := fetchTeamBeanSlice(team_slice)
			if err != nil {
				util.Debug(" Cannot fetch team bean slice given team_slice", err)
				report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
			if len(t_b_slice) >= 1 {
				fPD.Count = len(t_b_slice)
				fPD.TeamBeanSlice = t_b_slice
				fPD.IsEmpty = false
			}
		}
		renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_team", "component_avatar_name_gender")
		return

	case data.SearchTypeThreadTitle:
		//查询，茶议标题，thread.title
		thread_slice, err := data.SearchThreadByTitle(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" Cannot search thread by title", err)
			report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
			return
		}
		if len(thread_slice) >= 1 {
			thread_bean_slice, err := fetchThreadBeanSlice(thread_slice, r)
			if err != nil {
				util.Debug(" Cannot fetch thread bean slice given thread_slice", err)
				report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
			fPD.Count = len(thread_slice)
			fPD.ThreadBeanSlice = thread_bean_slice
			fPD.IsEmpty = false
		}
		renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_thread_bean", "component_avatar_name_gender")
		return

	case data.SearchTypeObjectiveTitle:
		//查询，茶会标题，objective.title
		objective_slice, err := data.SearchObjectiveByTitle(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" Cannot search objective by title", err)
			report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
			return
		}
		if len(objective_slice) >= 1 {
			objective_bean_slice, err := FetchObjectiveBeanSlice(objective_slice)
			if err != nil {
				util.Debug(" Cannot fetch objective bean slice given objective_slice", err)
				report(w, r, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
			fPD.Count = len(objective_slice)
			fPD.ObjectiveBeanSlice = objective_bean_slice
			fPD.IsEmpty = false
		}
		renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_objective_bean", "component_avatar_name_gender")
		return

	case data.SearchTypeProjectTitle:
		//按茶台标题查询
		project_slice, err := data.SearchProjectByTitle(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search project by title", err)
			fPD.IsEmpty = true
			renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_project_bean", "component_avatar_name_gender")
			return
		} else {
			if len(project_slice) >= 1 {
				project_bean_slice, err := fetchProjectBeanSlice(project_slice)
				if err != nil {
					util.Debug(" Cannot fetch project bean slice given project_slice", err)
				}
				fPD.Count = len(project_slice)
				fPD.ProjectBeanSlice = project_bean_slice
				fPD.IsEmpty = false
			}
			renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_project_bean", "component_avatar_name_gender")
			return
		}
	case data.SearchTypePlaceName:
		//查询品茶地点 place
		place_slice, err := data.FindPlaceByName(keyword)
		if err != nil {
			util.Debug(" failed to search place by keyword", err)
		}
		if len(place_slice) >= 1 {
			fPD.Count = len(place_slice)
			fPD.PlaceSlice = place_slice
			fPD.IsEmpty = false
		}
		renderHTML(w, &fPD, "layout", "navbar.private", "search", "component_place")
		return

	case data.SearchTypeEnvironment:
		//查询环境条件 environment
		environment_slice, err := data.SearchEnvironmentByName(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search environment by keyword", err)
		}
		if len(environment_slice) >= 1 {
			fPD.Count = len(environment_slice)
			fPD.EnvironmentSlice = environment_slice
			fPD.IsEmpty = false
		}
		renderHTML(w, &fPD, "layout", "navbar.private", "search")
		return

	default:
		report(w, r, "你好，茶博士摸摸头，还没有开放这种类型的查询功能，请换个查询类型再试。")
		return
	}
}

// GET /v1/SearchGet
// 打开查询页面
func SearchGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var f data.SearchPageData
	f.SessUser = s_u

	// 打开查询页面
	renderHTML(w, &f, "layout", "navbar.private", "search")
}
