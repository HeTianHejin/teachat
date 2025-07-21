package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// 处理新建茶话会的操作处理器
// 如果匹配到GET请求，检查用户是否已经登录，是就给打开objective.new页面
// 如果匹配到POST请求，检查用户是否已经登录，是就调用CreateObjective（）方法
// 如果匹配到其他方式请求，返回404错误
func HandleNewObjective(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		NewObjectiveGet(w, r)
	case http.MethodPost:
		NewObjectivePost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /objective/create
// 返回objective.new页面
func NewObjectiveGet(w http.ResponseWriter, r *http.Request) {
	//尝试从http请求中读取用户会话信息
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var oD data.ObjectiveDetail
	//根据会话读取当前用户的信息
	s_u, s_d_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchSessionUserRelatedData(s)
	if err != nil {
		util.Debug("cannot fetch s_u s_teams given session", err)
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}

	// 填写页面数据
	oD.SessUser = s_u

	oD.SessUserDefaultFamily = s_d_family
	oD.SessUserSurvivalFamilies = s_survival_families

	oD.SessUserDefaultTeam = s_default_team
	oD.SessUserSurvivalTeams = s_survival_teams

	oD.SessUserDefaultPlace = s_default_place
	oD.SessUserBindPlaces = s_places
	// 给请求用户返回新建茶话会页面
	RenderHTML(w, &oD, "layout", "navbar.private", "objective.new")
}

// POST /objective/create
// create the objective
// 创建新茶话会
func NewObjectivePost(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		Report(w, r, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 默认的茶话会封面图名
	cover := "default-ob-cover"
	// 读取http请求中form的数据
	title := r.PostFormValue("name")
	body := r.PostFormValue("description")
	class, err := strconv.Atoi(r.PostFormValue("class"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		Report(w, r, "茶话会类型参数格式错误")
		return
	}

	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.Debug("Failed to convert family_id to int", err)
		Report(w, r, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return
	}
	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		Report(w, r, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return
	}

	valid, err := validateTeamAndFamilyParams(is_private, team_id, family_id, s_u.Id, w, r)
	if !valid && err == nil {
		return // 参数不合法，已经处理了错误
	}
	if err != nil {
		// 处理数据库错误
		util.Debug("验证提交的团队和家庭id出现数据库错误", team_id, family_id, err)
		Report(w, r, "你好，成员资格检查失败，请确认后再试。")
		return
	}

	// 检查是否已经存在相同名字的茶话会
	obj := data.Objective{
		Title: title,
	}
	t_ob, err := obj.GetByTitle()
	if err != nil {
		util.Debug("查询茶话会失败", err)
		Report(w, r, "查询茶话会失败")
		return
	}
	if len(t_ob) > 0 {
		Report(w, r, "你好，编新不如述旧，刻古终胜雕今。茶话会名字重复哦，请确认后再试。")
		return
	}

	countObj := data.Objective{TeamId: team_id}
	count_team, err := countObj.CountByTeamId()
	if err != nil {
		util.Debug(" cannot get count given objective team_id", err)
		Report(w, r, "你好，游丝软系飘春榭，落絮轻沾扑绣帘。请确认后再试。")
		return
	}
	// 最大团队可以创建 茶话会 数量
	if count_team > int(util.Config.MaxInviteTeams) {
		Report(w, r, "你好，编新不如述旧，一个茶团最多可以开的茶话会数量是有限的，请确认后再试。")
		return
	}

	// 检测一下name是否>2中文字，body是否在min_word-int(util.Config.ThreadMaxWord)中文字，
	// 如果不是，返回错误信息
	if CnStrLen(title) < 2 || CnStrLen(title) > 36 {
		Report(w, r, "你好，茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if CnStrLen(body) < int(util.Config.ThreadMinWord) || CnStrLen(body) > int(util.Config.ThreadMaxWord) {
		Report(w, r, "你好，茶博士迷糊了，竟然说字数太少或者太多记不住，请确认后再试。")
		return
	}

	new_ob := data.Objective{
		Title:     title,
		UserId:    s_u.Id,
		Body:      body,
		Cover:     cover,
		Class:     class,
		FamilyId:  family_id,
		TeamId:    team_id,
		IsPrivate: is_private,
	}

	switch class {
	case data.ObClassOpenStraw:
		//如果class=10开放式茶话会草围
		//尝试保存新茶话会
		if err = new_ob.Create(); err != nil {
			// 记录错误，提示用户新开茶话会未成功
			util.Debug(" Cannot create objective", err)
			Report(w, r, "你好，偷来梨蕊三分白，借得梅花一缕魂。")
			return
		}

	case data.ObClassCloseStraw:
		//如果class=20封闭式茶话会(草围)，需要读取指定茶团号TeamIds列表
		tIds_str := r.PostFormValue("invite_ids")

		//用正则表达式检测茶团号TeamIds，是否符合“整数，整数，整数...”的格式
		if !VerifyIdSliceFormat(tIds_str) {
			util.Debug(" TeamId slice format is wrong", err)
			Report(w, r, "你好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
			return
		}
		//用户提交的t_id是以逗号分隔的字符串,需要分割后，转换成[]Id,以便处理
		t_ids_str := strings.Split(tIds_str, ",")
		// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
		if len(t_ids_str) > int(util.Config.MaxInviteTeams) {
			util.Debug(" Too many team ids", err)
			Report(w, r, "你好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，茶壶不够用，请确认后再试。")
			return
		}
		t_id_slice := make([]int, 0, util.Config.MaxInviteTeams)
		for _, t_id := range t_ids_str {
			te_id_int, _ := strconv.Atoi(t_id)
			t_id_slice = append(t_id_slice, te_id_int)
		}
		// 使用事务创建封闭式茶话会及其许可茶团
		if err = data.CreateObjectiveWithTeams(&new_ob, t_id_slice); err != nil {
			util.Debug("创建封闭式茶话会失败", err)
			Report(w, r, "你好，茶博士迷糊了，未能创建茶话会，请稍后再试。")
			return
		}
	default:
		// 非法的茶话会属性
		util.Debug(" Unknown objective class", err)
		Report(w, r, "你好，身前有余勿伸手，眼前无路请回头，请稍后再试。")
		return
	}

	if util.Config.PoliteMode {
		if err = CreateAndSendAcceptMessage(new_ob.Id, data.AcceptObjectTypeOb, s_u.Id); err != nil {
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				Report(w, r, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				Report(w, r, "你好，茶博士迷路了，未能发送蒙评请求消息。")
			}
			return
		}

		t := fmt.Sprintf("你好，新开茶话会 %s 已准备妥当，稍等有缘茶友评审通过之后，即可启用。", new_ob.Title)
		// 提示用户草稿保存成功
		Report(w, r, t)
		return
	} else {
		objective, err := AcceptNewObjective(new_ob.Id)
		if err != nil {
			Report(w, r, err.Error())
			return
		}
		//跳转到茶话会详情页
		http.Redirect(w, r, fmt.Sprintf("/v1/objective/detail?uuid=%s", objective.Uuid), http.StatusFound)
	}

}

// GET /objective/square
// show the random objectives
// 根据用户是否登录显示不同导航条的茶话会广场页面
func ObjectiveSquare(w http.ResponseWriter, r *http.Request) {
	var oSpD data.ObjectiveSquare

	// 如何排序茶话会，是个问题！按照圆桌会议平等约定，应该是轮流出现在茶话会广场才合适，
	// 而且，应该是按人头出现计数，而不是按热度或者建的茶话会数量。
	// 每次展示2打=24个茶话会
	// 用（随机选中）选取24个用户的24茶话会的模式

	// test获取所有茶话会
	objective_slice, err := data.GetPublicObjectives(24)
	if err != nil {
		util.Debug(" Cannot get objectives", err)
		Report(w, r, "你好，茶博士失魂鱼，未能获取缘分茶话会资料，请稍后再试。")
		return
	}
	len := len(objective_slice)
	if len == 0 {
		Report(w, r, "你好，山穷水尽疑无路，为何没有任何茶话会资料？请稍后再试。")
		return
	}

	oSpD.ObjectiveBeanSlice, err = FetchObjectiveBeanSlice(objective_slice)
	if err != nil {
		util.Debug(" Cannot read objective-bean slice", err)
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}

	//检查用户是否已经登录
	s, err := Session(r)
	if err != nil {
		//未登录！游客
		oSpD.SessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
		}
		//迭代茶话会队列，把作者属性设置为false
		for i := range oSpD.ObjectiveBeanSlice {
			oSpD.ObjectiveBeanSlice[i].Objective.ActiveData.IsAuthor = false
		}

		//返回页面
		RenderHTML(w, &oSpD, "layout", "navbar.public", "objectives.square", "component_objective_bean", "component_avatar_name_gender")
		return
	}
	//已登录，读取用户信息
	sUser, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		//跳转登录页面
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	oSpD.SessUser = sUser
	for i := range oSpD.ObjectiveBeanSlice {
		if oSpD.ObjectiveBeanSlice[i].Objective.UserId == sUser.Id {
			oSpD.ObjectiveBeanSlice[i].Objective.ActiveData.IsAuthor = true
		} else {
			oSpD.ObjectiveBeanSlice[i].Objective.ActiveData.IsAuthor = false
		}
	}
	RenderHTML(w, &oSpD, "layout", "navbar.private", "objectives.square", "component_objective_bean", "component_avatar_name_gender")

}

// GET /v1/objective/detail?uuid=
// show the details of the objective
// 读取指定的uuid茶话会详情
func ObjectiveDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	var oD data.ObjectiveDetail
	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	if uuid == "" {
		Report(w, r, "你好，茶博士看不懂陛下提交的UUID参数，请稍后再试。")
		return
	}
	// 根据uuid查询茶话会资料
	ob := data.Objective{
		Uuid: uuid}
	if err = ob.GetByUuid(); err != nil {
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}

	switch ob.Class {
	case data.ObClassOpen, data.ObClassClose:
		break
	case data.ObClassOpenStraw, data.ObClassCloseStraw:
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	default:
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}
	oD.ObjectiveBean, err = FetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot read objective-bean slice", err)
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}

	//fetch public projects
	project_slice, err := ob.GetPublicProjects()
	if err != nil {
		util.Debug(" Cannot read objective-bean slice", err)
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}
	oD.ProjectBeanSlice, err = FetchProjectBeanSlice(project_slice)
	if err != nil {
		util.Debug(" Cannot read objective-bean slice", err)
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}
	//检查用户是否已经登录
	s, err := Session(r)
	if err != nil {
		//未登录！
		oD.IsGuest = true
		oD.SessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
			// 用户足迹
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		//配置公开导航条的茶话会详情页面
		RenderHTML(w, &oD, "layout", "navbar.public", "objective.detail", "component_project_bean", "component_avatar_name_gender", "component_sess_capacity")
		return
	}

	//已经登录！
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 记录用户查询的资讯
	// if err = RecordLastQueryPath(s_u.Id, r.URL.Path, r.URL.RawQuery); err != nil {
	// 	util.Debug(s_u.Email, " Cannot record last query path")
	// }
	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery
	oD.SessUser = s_u

	// 如果这个茶话会是封闭式，检查当前用户是否属于受邀请团队成员
	if ob.Class == data.ObClassClose {
		ok, err := oD.ObjectiveBean.Objective.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot read objective-bean slice", err)
			Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌着。")
			return
		}
		oD.IsInvited = ok
	}

	//检测当前用户身份
	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("Admin permission check failed",
			"userId", s_u.Id,
			"objectiveId", ob.Id,
			"error", err,
		)
		Report(w, r, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	oD.IsAdmin = is_admin

	if !oD.IsAdmin {
		veri_team := data.Team{Id: data.TeamIdVerifier}
		is_member, err := veri_team.IsMember(s_u.Id)
		if err != nil {
			util.Debug("Cannot check verifier team member", err)
			Report(w, r, "你好，茶博士，有眼不识泰山。")
			return
		}
		if is_member {
			oD.IsVerifier = true
		}
	}

	//配置私有导航条的茶话会详情页面
	RenderHTML(w, &oD, "layout", "navbar.private", "objective.detail", "component_project_bean", "component_avatar_name_gender", "component_sess_capacity")

}
