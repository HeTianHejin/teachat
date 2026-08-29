package route

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	dao "teachat/DAO"
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
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session %v", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form %v", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能理解你的话语，请稍后再试。")
		return
	}

	//读取查询参数
	class_str := r.PostFormValue("class")
	//转换class_str为int
	class_int, err := strconv.Atoi(class_str)
	if err != nil {
		util.Debug("Cannot convert class_str to int %v", err)
		report(w, s_u, "你好，茶博士摸摸头，说茶语本上落了片白茫茫大地真干净，请稍后再试。")
		return
	}

	keyword := r.PostFormValue("keyword")
	//检查keyword的文字长度是否>1 and <32
	keyword_len := len(keyword)
	if keyword_len < 1 || keyword_len > 32 {
		report(w, s_u, "你好，茶博士摸摸头，说关键词太长了记不住呢，请确认后再试。")
		return
	}

	var fPD dao.SearchPageData
	fPD.SessUser = s_u
	//初始化获取结果为零记录
	fPD.IsEmpty = true

	//根据查询类型操作
	switch class_int {
	case dao.SearchTypeUserNameOrEmail:
		//按花名或者邮箱查找茶友，user

		//用户可能提交了一个电子邮箱地址，如果是，我们需要先通过电子邮箱地址查找用户
		//检查keyword是否是电子邮箱地址
		if ok := isEmail(keyword); ok {
			user, err := dao.GetUserByEmail(keyword, r.Context())
			if err != nil {
				util.Debug(keyword, " Cannot search user by keyword", err)
				report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
			//如果user是非空
			if user.Id > 0 {
				user_bean, err := fetchUserDefaultBean(user)
				if err != nil {
					util.Debug("cannot get user-bean given user %v", err)
					report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
					return
				} else {
					fPD.UserDefaultDataBeanSlice = append(fPD.UserDefaultDataBeanSlice, user_bean)
					fPD.IsEmpty = false
				}
			}
		} else {
			user_slice, err := dao.SearchUserByNameKeyword(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
			if err != nil {
				util.Debug(" Cannot search user by keyword %v", err)
				report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}

			if len(user_slice) >= 1 {
				fPD.UserDefaultDataBeanSlice, err = fetchUserDefaultDataBeanSlice(user_slice)
				if err != nil {
					util.Debug(" Cannot fetch user bean slice given user_slice %v", err)
					report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
					return
				}
				fPD.IsEmpty = false
			}
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_avatar_name_gender")
		return

	case dao.SearchTypeUserId:
		//按user_id查询茶友
		// 验证关键词是否为正整数
		keyword_int, err := strconv.Atoi(keyword)
		if err != nil || keyword_int <= 0 {
			report(w, s_u, "茶友号必须是正整数")
			return
		}
		user, err := dao.GetUser(keyword_int)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fPD.IsEmpty = true
				generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_avatar_name_gender")
				return
			} else {
				util.Debug("failed to get user given user_id:  %v", err)
				report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
		}

		//如果user是非空
		if user.Id > 0 {
			userbean, err := fetchUserDefaultBean(user)
			if err != nil {
				util.Debug("cannot get user-bean given user %v", err)
				report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			} else {
				fPD.UserDefaultDataBeanSlice = append(fPD.UserDefaultDataBeanSlice, userbean)
				fPD.IsEmpty = false
			}
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_avatar_name_gender")
		return
	case dao.SearchTypeTeamAbbr:
		//查询，茶团简称，team.abbreviation
		team_slice, err := dao.SearchTeamByAbbreviation(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" Cannot search team by abbreviation %v", err)
			report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
			return
		}

		if len(team_slice) >= 1 {
			t_b_slice, err := fetchTeamBeanSlice(team_slice)
			if err != nil {
				util.Debug(" Cannot fetch team bean slice given team_slice %v", err)
				report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
			if len(t_b_slice) >= 1 {
				fPD.Count = len(t_b_slice)
				fPD.TeamBeanSlice = t_b_slice
				fPD.IsEmpty = false
			}
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_team", "component_avatar_name_gender")
		return

	case dao.SearchTypeThreadTitle:
		//查询，茶议标题，thread.title
		thread_slice, err := dao.SearchThreadByTitle(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" Cannot search thread by title %v", err)
			report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
			return
		}
		if len(thread_slice) >= 1 {
			thread_bean_slice, err := fetchThreadBeanSlice(thread_slice, r)
			if err != nil {
				util.Debug(" Cannot fetch thread bean slice given thread_slice %v", err)
				report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
			fPD.Count = len(thread_slice)
			fPD.ThreadBeanSlice = thread_bean_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_thread_bean", "component_avatar_name_gender")
		return

	case dao.SearchTypeObjectiveTitle:
		//查询，茶会标题，objective.title
		objective_slice, err := dao.SearchObjectiveByTitle(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" Cannot search objective by title %v", err)
			report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
			return
		}
		if len(objective_slice) >= 1 {
			objective_bean_slice, err := FetchObjectiveBeanSlice(objective_slice)
			if err != nil {
				util.Debug(" Cannot fetch objective bean slice given objective_slice %v", err)
				report(w, s_u, "你好，茶博士摸摸头，说搜索关键词无效，请确认后再试。")
				return
			}
			fPD.Count = len(objective_slice)
			fPD.ObjectiveBeanSlice = objective_bean_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_objective_bean", "component_avatar_name_gender")
		return

	case dao.SearchTypeProjectTitle:
		//按茶台标题查询
		project_slice, err := dao.SearchProjectByTitle(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search project by title %v", err)
			fPD.IsEmpty = true
			generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_project_bean", "component_avatar_name_gender")
			return
		} else {
			if len(project_slice) >= 1 {
				project_bean_slice, err := fetchProjectBeanSlice(project_slice)
				if err != nil {
					util.Debug(" Cannot fetch project bean slice given project_slice %v", err)
				}
				fPD.Count = len(project_slice)
				fPD.ProjectBeanSlice = project_bean_slice
				fPD.IsEmpty = false
			}
			generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_project_bean", "component_avatar_name_gender")
			return
		}
	case dao.SearchTypePlaceName:
		//查询品茶地点 place
		place_slice, err := dao.FindPlaceByName(keyword)
		if err != nil {
			util.Debug(" failed to search place by keyword %v", err)
		}
		if len(place_slice) >= 1 {
			fPD.Count = len(place_slice)
			fPD.PlaceSlice = place_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_place")
		return

	case dao.SearchTypeEnvironment:
		//查询环境条件 environment
		environment_slice, err := dao.SearchEnvironmentByName(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search environment by keyword %v", err)
		}
		if len(environment_slice) >= 1 {
			fPD.Count = len(environment_slice)
			fPD.EnvironmentSlice = environment_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search")
		return

	case dao.SearchTypeHazard:
		//查询隐患 hazard
		hazard_slice, err := dao.SearchHazardByName(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search hazard by keyword %v", err)
		}
		if len(hazard_slice) >= 1 {
			fPD.Count = len(hazard_slice)
			fPD.HazardSlice = hazard_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search")
		return

	case dao.SearchTypeGoods:
		// 查询物资（goods）
		goods_slice, err := dao.SearchGoodsByName(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search goods by keyword %v", err)
		}
		if len(goods_slice) >= 1 {
			fPD.Count = len(goods_slice)
			fPD.GoodsSlice = goods_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search")
		return

	case dao.SearchTypeRisk: // SearchTypeRisk
		//查询风险 risk
		risk_slice, err := dao.SearchRiskByName(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search risk by keyword %v", err)
		}
		if len(risk_slice) >= 1 {
			fPD.Count = len(risk_slice)
			fPD.RiskSlice = risk_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search")
		return

	case dao.SearchTypeSkill:
		//查询技能 skill
		skill_slice, err := dao.SearchSkillByName(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search skill by keyword:  %v", err)
		}
		if len(skill_slice) >= 1 {
			fPD.Count = len(skill_slice)
			fPD.SkillSlice = skill_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search")
		return

	case dao.SearchTypeMagic:
		//查询法力 magic
		magic_slice, err := dao.SearchMagicByName(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err != nil {
			util.Debug(" failed to search magic by keyword:  %v", err)
		}
		if len(magic_slice) >= 1 {
			fPD.Count = len(magic_slice)
			fPD.MagicSlice = magic_slice
			fPD.IsEmpty = false
		}
		generateHTML(w, &fPD, "layout", "navbar.private", "search")
		return

	case dao.SearchTypeFamilyId:
		familyID, err := strconv.Atoi(keyword)
		if err != nil || familyID <= 0 {
			report(w, s_u, "家庭号必须是正整数")
			return
		}

		family, err := dao.SearchFamilyById(familyID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fPD.IsEmpty = true
				generateHTML(w, &fPD, "layout", "navbar.private", "search")
				return
			}
			util.Debug("failed to search family by id:  %v", err)
			report(w, s_u, "你好，开水太烫不好泡茶，请稍后再试。")
			return
		}

		// 检查用户是否有权限查看该家庭
		if !canViewFamily(&family, &s_u, r.Context()) {
			fPD.IsEmpty = true
			generateHTML(w, &fPD, "layout", "navbar.private", "search")
			return
		}

		familyBean, err := fetchFamilyBean(family)
		if err != nil {
			util.Debug("failed to fetch family bean by id:  %v", err)
			report(w, s_u, "你好，开水太烫不好泡茶，请稍后再试。")
			return
		}

		fPD.FamilyBeanSlice = append(fPD.FamilyBeanSlice, familyBean)
		fPD.Count = 1
		fPD.IsEmpty = false
		generateHTML(w, &fPD, "layout", "navbar.private", "search", "component_family")
		return

	default:
		report(w, s_u, "你好，茶博士摸摸头，还没有开放这种类型的查询功能，请换个查询类型再试。")
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
		util.Debug("Cannot get user from session %v", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var f dao.SearchPageData
	f.SessUser = s_u

	// 打开查询页面
	generateHTML(w, &f, "layout", "navbar.private", "search")
}

// canViewFamily 判断用户是否有权限查看家庭信息
// 公开家庭：所有人可见
// 私密家庭：仅被声明为新成员的用户可见
func canViewFamily(family *dao.Family, user *dao.User, ctx context.Context) bool {
	if family == nil {
		return false
	}
	if family.IsOpen {
		return true
	}
	// 未登录用户无法查看私密家庭
	if user == nil || user.Id <= dao.UserId_None {
		return false
	}

	announcements, err := dao.FindFamilyAnnouncementByMemberId(user.Id, ctx)
	if err != nil {
		util.Debug("failed to find family announcement by member id:  %v", err)
		return false
	}

	for _, fa := range announcements {
		if fa.FamilyId == family.Id {
			return true
		}
	}
	return false
}
