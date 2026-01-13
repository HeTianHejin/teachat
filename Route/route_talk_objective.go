package route

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	dao "teachat/DAO"
	util "teachat/Util"
	"time"
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
// 返回创建目标茶围页面
func NewObjectiveGet(w http.ResponseWriter, r *http.Request) {
	//尝试从http请求中读取用户会话信息
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var oD dao.ObjectiveDetail
	//根据会话读取当前用户的信息
	s_u, s_d_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := fetchSessionUserRelatedData(s, r.Context())
	if err != nil {
		util.Debug("cannot fetch s_u s_teams given session", err)
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
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
	generateHTML(w, &oD, "layout", "navbar.private", "objective.new")
}

// POST /objective/create
// create the objective
// 创建新茶话会
func NewObjectivePost(w http.ResponseWriter, r *http.Request) {
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
	err = r.ParseForm()
	if err != nil {
		report(w, s_u, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
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
		report(w, s_u, "茶话会类型参数格式错误")
		return
	}

	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.Debug("Failed to convert family_id to int", err)
		report(w, s_u, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return
	}
	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		report(w, s_u, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return
	}

	valid, err := validateTeamAndFamilyParams(is_private, team_id, family_id, s_u, w)
	if !valid && err == nil {
		return // 参数不合法，已经处理了错误
	}
	if err != nil {
		// 处理数据库错误
		util.Debug("验证提交的团队和家庭id出现数据库错误", team_id, family_id, err)
		report(w, s_u, "你好，成员资格检查失败，请确认后再试。")
		return
	}

	// 检查是否已经存在相同名字的茶话会
	obj := dao.Objective{
		Title: title,
	}
	t_ob, err := obj.GetByTitle()
	if err != nil {
		util.Debug("查询茶话会失败", err)
		report(w, s_u, "查询茶话会失败")
		return
	}
	if len(t_ob) > 0 {
		report(w, s_u, "你好，编新不如述旧，刻古终胜雕今。茶话会名字重复哦，请确认后再试。")
		return
	}

	countObj := dao.Objective{TeamId: team_id}
	count_team, err := countObj.CountByTeamId()
	if err != nil {
		util.Debug(" cannot get count given objective team_id", err)
		report(w, s_u, "你好，游丝软系飘春榭，落絮轻沾扑绣帘。请确认后再试。")
		return
	}
	// 最大团队可以创建 茶话会 数量
	if count_team > int(util.Config.MaxInviteTeams) {
		report(w, s_u, "你好，编新不如述旧，一个茶团最多可以开的茶话会数量是有限的，请确认后再试。")
		return
	}

	// 检测一下name是否>2中文字，body是否在min_word-int(util.Config.ThreadMaxWord)中文字，
	// 如果不是，返回错误信息
	if cnStrLen(title) < 2 || cnStrLen(title) > 36 {
		report(w, s_u, "你好，茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if cnStrLen(body) < int(util.Config.ThreadMinWord) || cnStrLen(body) > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "你好，茶博士迷糊了，竟然说字数太少或者太多记不住，请确认后再试。")
		return
	}

	new_ob := dao.Objective{
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
	case dao.ObClassOpenDraft:
		//如果class=10开放式茶话会草围
		//尝试保存新茶话会
		if err = new_ob.Create(); err != nil {
			// 记录错误，提示用户新开茶话会未成功
			util.Debug(" Cannot create objective", err)
			report(w, s_u, "你好，偷来梨蕊三分白，借得梅花一缕魂。")
			return
		}

	case dao.ObClassCloseDraft:
		//如果class=20封闭式茶话会(草围)，需要读取指定茶团号TeamIds列表
		tIds_str := r.PostFormValue("invite_ids")
		if tIds_str == "" {
			report(w, s_u, "你好，茶博士迷糊了，竟然说封闭式茶话会的茶团号不能省事不写，请确认后再试。")
			return
		}
		t_id_slice, err := parseIdSlice(tIds_str)
		if err != nil {
			util.Debug(" Cannot parse team ids", err)
			report(w, s_u, "你好，陛下填写的茶团号格式看不懂，必需是不重复的自然数用英文逗号分隔。")
			return
		}

		// 使用事务创建封闭式茶话会及其许可茶团
		if err = dao.CreateObjectiveWithTeams(&new_ob, t_id_slice); err != nil {
			util.Debug("创建封闭式茶话会失败", err)
			report(w, s_u, "你好，茶博士迷糊了，未能创建茶话会，请稍后再试。")
			return
		}
	default:
		// 非法的茶话会属性
		util.Debug(" Unknown objective class", err)
		report(w, s_u, "你好，身前有余勿伸手，眼前无路请回头，请稍后再试。")
		return
	}

	if util.Config.PoliteMode {
		if err = createAndSendAcceptNotification(new_ob.Id, dao.AcceptObjectTypeObjective, s_u.Id, r.Context()); err != nil {
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				report(w, s_u, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				report(w, s_u, "你好，茶博士迷路了，未能发送蒙评请求通知。")
			}
			return
		}

		t := fmt.Sprintf("你好，新开茶话会 %s 已准备妥当，稍等有缘茶友评审通过之后，即可启用。", new_ob.Title)
		// 提示用户草稿保存成功
		report(w, s_u, t)
		return
	} else {
		objective, err := acceptNewObjective(new_ob.Id)
		if err != nil {
			report(w, s_u, err.Error())
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
	var oSpD dao.ObjectiveSquare
	s_u := dao.UserUnknown

	// 每次展示2打=24个茶话会
	// 用（随机选中）选取24个用户的24茶话会的模式

	// test获取24茶话会
	objective_slice, err := dao.GetPublicObjectives(24)
	if err != nil {
		util.Debug(" Cannot get objectives", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取缘分茶话会资料，请稍后再试。")
		return
	}
	len := len(objective_slice)
	if len > 0 {
		oSpD.ObjectiveBeanSlice, err = FetchObjectiveBeanSlice(objective_slice)
		if err != nil {
			util.Debug(" Cannot read objective-bean slice", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
			return
		}
	} else {
		oSpD.ObjectiveBeanSlice = []dao.ObjectiveBean{}
	}

	//检查用户是否已经登录
	s, err := session(r)
	if err != nil {
		//未登录！游客
		oSpD.SessUser = dao.User{
			Id:   dao.UserId_None,
			Name: "游客",
		}
		//迭代茶话会队列，把作者属性设置为false
		for i := range oSpD.ObjectiveBeanSlice {
			oSpD.ObjectiveBeanSlice[i].Objective.ActiveData.IsAuthor = false
		}

		//返回页面
		generateHTML(w, &oSpD, "layout", "navbar.public", "objectives.square", "component_objective_bean", "component_avatar_name_gender")
		return
	}
	//已登录，读取用户信息
	s_u, err = s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		//跳转登录页面
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	oSpD.SessUser = s_u
	for i := range oSpD.ObjectiveBeanSlice {
		if oSpD.ObjectiveBeanSlice[i].Objective.UserId == s_u.Id {
			oSpD.ObjectiveBeanSlice[i].Objective.ActiveData.IsAuthor = true
		} else {
			oSpD.ObjectiveBeanSlice[i].Objective.ActiveData.IsAuthor = false
		}
	}
	generateHTML(w, &oSpD, "layout", "navbar.private", "objectives.square", "component_objective_bean", "component_avatar_name_gender")

}

// GET /v1/objective/detail?uuid=
// show the details of the objective
// 读取指定的uuid茶话会详情
func ObjectiveDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	var oD dao.ObjectiveDetail
	s_u := dao.UserUnknown
	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士看不懂陛下提交的UUID参数，请稍后再试。")
		return
	}
	// 根据uuid查询茶话会资料
	ob := dao.Objective{
		Uuid: uuid}
	if err = ob.GetByUuid(); err != nil {
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}

	switch ob.Class {
	case dao.ObClassOpen, dao.ObClassClose:
		break
	case dao.ObClassOpenDraft, dao.ObClassCloseDraft:
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	default:
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}
	oD.ObjectiveBean, err = fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot read objective-bean slice", err)
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}

	//fetch public projects
	project_slice, err := ob.GetPublicProjects()
	if err != nil {
		util.Debug(" Cannot read objective-bean slice", err)
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}
	oD.ProjectBeanSlice, err = fetchProjectBeanSlice(project_slice)
	if err != nil {
		util.Debug(" Cannot read objective-bean slice", err)
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌奋斗着。")
		return
	}
	//检查用户是否已经登录
	s, err := session(r)
	if err != nil {
		//未登录！
		oD.IsGuest = true
		oD.SessUser = dao.User{
			Id:   dao.UserId_None,
			Name: "游客",
			// 用户足迹
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		//配置公开导航条的茶话会详情页面
		generateHTML(w, &oD, "layout", "navbar.public", "objective.detail", "component_project_bean", "component_avatar_name_gender", "component_sess_capacity")
		return
	}

	//已经登录！
	s_u, err = s.User()
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
	if ob.Class == dao.ObClassClose {
		is_invited, err := oD.ObjectiveBean.Objective.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot read objective-bean slice", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你时刻忙碌着。")
			return
		}
		oD.IsInvited = is_invited
	}

	//检测当前用户身份
	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("Admin permission check failed",
			"userId", s_u.Id,
			"objectiveId", ob.Id,
			"error", err,
		)
		report(w, s_u, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	oD.IsAdmin = is_admin

	if !oD.IsAdmin {

		oD.IsVerifier = dao.IsVerifier(s_u.Id)

	}

	//配置私有导航条的茶话会详情页面
	generateHTML(w, &oD, "layout", "navbar.private", "objective.detail", "component_project_bean", "component_avatar_name_gender", "component_sess_capacity")

}

// 处理补充茶围目标的响应和数据提交保存,GET POST
func HandleObjectiveSupplement(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		objectiveSupplementGet(w, r)
	case http.MethodPost:
		objectiveSupplementPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/objective/supplement?uuid=xxx
// 打开指定的茶围目标追加（补充必需内容）页面
func objectiveSupplementGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", sess.Email, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "茶博士失魂鱼，未能读取茶围目标编号，请确认后再试。")
		return
	}

	var obSupp dao.ObjectiveSupplement
	// 读取茶围目标内容
	ob := dao.Objective{Uuid: uuid}
	if err = ob.GetByUuid(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, s_u, "你好，茶博士竟然说该茶围目标不存在，请确认后再试一次。")
			return
		}
		util.Debug(" Cannot read objective given uuid", uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	//核对用户身份，是否具有完善操作权限
	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("objective Admin permission check failed:", "userId", s_u.Id, "objectiveId", ob.Id, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	obSupp.IsAdmin = is_admin
	obSupp.IsVerifier = dao.IsVerifier(s_u.Id)

	// 读取茶围目标资料荚
	obSupp.ObjectiveBean, err = fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot read objectiveBean", err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 计算字数信息
	currentLength := cnStrLen(ob.Body)
	maxLength := int(util.Config.ThreadMaxWord)
	remainingLength := maxLength - currentLength
	if remainingLength < 0 {
		remainingLength = 0
	}

	// 添加字数统计信息到模板数据
	obSupp.CurrentLength = currentLength
	obSupp.MaxLength = maxLength
	obSupp.RemainingLength = remainingLength

	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	obSupp.SessUser = s_u

	generateHTML(w, &obSupp, "layout", "navbar.private", "objective.supplement", "component_sess_capacity", "component_avatar_name_gender")
}

// POST /v1/objective/supplement
// 补充完整茶围目标内容
func objectiveSupplementPost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", sess.Email, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	//获取post方法提交的表单
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	o_uuid := r.PostFormValue("uuid")
	if o_uuid == "" {
		report(w, s_u, "你好，茶博士扶起厚厚的眼镜，居然说您补充的茶围目标编号不存在。")
		return
	}
	//读取提交的additional
	additional := r.PostFormValue("additional")

	// 读取茶围目标内容
	ob := dao.Objective{Uuid: o_uuid}
	if err = ob.GetByUuid(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, s_u, "你好，茶博士竟然说该茶围目标不存在，请确认后再试一次。")
			return
		}
		util.Debug(" Cannot read objective given uuid", o_uuid, err)
		report(w, s_u, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("Admin permission check failed",
			"userId", s_u.Id,
			"objectiveId", ob.Id,
			"error", err,
		)
		report(w, s_u, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	if !is_admin {
		report(w, s_u, "茶博士惊讶，陛下你没有权限补充该茶围目标，请确认后再试。")
		return
	}

	//读取提交内容要求>int(util.Config.ThreadMinWord)中文字符，加上已有内容是否<=int(util.Config.ThreadMaxWord)
	if ok := submitAdditionalContent(w, s_u, ob.Body, additional); !ok {
		report(w, s_u, "你好，茶博士扶起厚厚的眼镜，居然说陛下您补充的茶围目标内容太多了！请确认后再试。")
		return
	}
	//当前"[中文时间字符 + 补充]" + body
	//获取当前时间，格式化成中文时间字符
	now := time.Now()
	timeStr := now.Format("2006年1月2日 15:04:05")
	name := s_u.Name
	// 追加内容（另起一行）
	t := "\n[" + timeStr + " " + name + " 补充] " + additional // 注意开头的 \n
	ob.Body += t
	//更新茶围目标内容
	if err = ob.Update(); err != nil {
		util.Debug(" Cannot update objective", err)
		report(w, s_u, "你好，茶博士失魂鱼，墨水中断未能补充茶围目标。")
		return
	}

	http.Redirect(w, r, "/v1/objective/detail?uuid="+o_uuid, http.StatusFound)
}
