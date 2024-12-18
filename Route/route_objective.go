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
	case "GET":
		NewObjectiveForm(w, r)
	case "POST":
		CreateObjective(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /objective/create
// 返回objective.new页面
func NewObjectiveForm(w http.ResponseWriter, r *http.Request) {
	//尝试从http请求中读取用户会话信息
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var oD data.ObjectiveDetail
	//根据会话读取当前用户的信息
	s_u, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		util.Danger(err, "cannot fetch s_u s_teams given session")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 填写页面数据
	oD.SessUser = s_u
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
func CreateObjective(w http.ResponseWriter, r *http.Request) {
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
		util.Info(err, "Cannot get user from session")
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
		util.Warning(err, "Failed to convert class to int")
		return
	}
	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.Warning(err, "Failed to convert class to int")
		return
	}
	// check the given team_id is valid
	_, err = data.GetTeamMemberByTeamIdAndUserId(team_id, s_u.Id)
	if err != nil {
		Report(w, r, "你好，眼前无路想回头，什么团成员？什么茶话会？请稍后再试。")
		return
	}

	// 检查是否已经存在相同名字的茶话会
	obj := data.Objective{
		Title: title,
	}
	err = obj.GetByTitleClass()
	if err == nil {
		// 已经存在相同名字且状态正常的茶话会
		Report(w, r, "你好，编新不如述旧，刻古终胜雕今。茶话会名字重复无法辨雌雄哦，请确认后再试。")
		return
	}

	count, err := obj.CountByTitle()
	if err != nil {
		util.Danger(err, " cannot get count given objective title")
		Report(w, r, "你好，游丝软系飘春榭，落絮轻沾扑绣帘。请确认后再试。")
		return
	}
	// 最大可创建 相同名称 茶话会数量
	if count >= 3 {
		Report(w, r, "你好，编新不如述旧，刻古终胜雕今。茶话会相同名称仅能使用3次，请确认后再试。")
		return
	} else if obj.TeamId != team_id {
		// 已经存在相同名字的茶话会,检查是否同一团队
		Report(w, r, "你好，这个茶话会名称已经被其它团队使用了，请确认后再试。")
		return
	}
	count, err = obj.CountByTeamId()
	if err != nil {
		util.Danger(err, " cannot get count given objective team_id")
		Report(w, r, "你好，游丝软系飘春榭，落絮轻沾扑绣帘。请确认后再试。")
		return
	}
	// 最大团队可以创建 茶话会 数量
	if count >= int(util.Config.MaxInviteTeams) {
		Report(w, r, "你好，编新不如述旧，刻古终胜雕今!一个茶团最多可以开12个茶话会呢，请确认后再试。")
		return
	}

	// 检测一下name是否>2中文字，body是否在17-456中文字，
	// 如果不是，返回错误信息
	if CnStrLen(title) < 2 || CnStrLen(title) > 36 {
		Report(w, r, "你好，粗声粗气的茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if CnStrLen(body) < 17 || CnStrLen(body) > 456 {
		Report(w, r, "你好，茶博士迷糊了，竟然说字数太少或者太多记不住，请确认后再试。")
		return
	}

	new_ob := data.Objective{
		Title:  title,
		Body:   body,
		Cover:  cover,
		Class:  class,
		UserId: s_u.Id,
		TeamId: team_id,
	}

	switch class {
	case 10:
		//如果class=10开放式茶话会草围
		//尝试保存新茶话会
		if err = new_ob.Create(); err != nil {
			// 撤回（删除）发送给两个用户的消息，测试未做 ～～～～～～～～～:P

			// 记录错误，提示用户新开茶话会未成功
			util.Warning(err, " Cannot create objective")
			Report(w, r, "你好，偷来梨蕊三分白，借得梅花一缕魂。")
			return
		}

	case 20:
		//如果class=20封闭式茶话会草围，需要读取指定茶团号TeamId列表

		tIds_str := r.PostFormValue("invite_team_ids")

		//用正则表达式检测茶团号TeamIds，是否符合“整数，整数，整数...”的格式
		if !VerifyTeamIdListFormat(tIds_str) {
			util.Warning(err, " TeamId list format is wrong")
			Report(w, r, "你好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
			return
		}
		//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId,以便处理
		te_ids_str := strings.Split(tIds_str, ",")
		// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
		if len(te_ids_str) > int(util.Config.MaxInviteTeams) {
			util.Warning(err, " Too many team ids")
			Report(w, r, "你好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，茶壶不够用，请确认后再试。")
			return
		}
		team_id_list := make([]int, 0, util.Config.MaxInviteTeams)
		for _, t_id := range te_ids_str {
			te_id_int, _ := strconv.Atoi(t_id)
			team_id_list = append(team_id_list, te_id_int)
		}
		//尝试保存新茶话会草稿
		if err = new_ob.Create(); err != nil {
			// 撤回发送给两个用户的消息，测试未做 ～～～～～～～～～:P

			util.Warning(err, " Cannot create objective")
			Report(w, r, "你好，茶博士迷糊了，未能创建茶话会，请稍后再试。")
			return
		}

		// 迭代team_id_list，尝试保存新封闭式茶话会草围邀请的茶团
		for _, team_id := range team_id_list {
			obInviTeams := data.ObjectiveInvitedTeam{
				ObjectiveId: new_ob.Id,
				TeamId:      team_id,
			}
			if err = obInviTeams.Create(); err != nil {
				// 撤回发送给两个用户的消息，测试未做 ～～～～～～～～～:P

				util.Warning(err, " Cannot create objectiveLicenseTeam")
			}
		}
	default:
		// 未知的茶话会属性
		util.Warning(err, " Unknown objective class")
		Report(w, r, "你好，茶博士还在研究茶话会的围字是不是有四种写法，忘记创建茶话会了，请稍后再试。")
		return
	}

	// 创建一条友邻盲评,是否接纳 新茶的记录
	aO := data.AcceptObject{
		ObjectId:   new_ob.Id,
		ObjectType: 1,
	}
	if err = aO.Create(); err != nil {
		util.Warning(err, "Cannot create accept_object")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	// 发送盲评请求消息给两个在线用户
	// 构造消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "你好，茶博士隆重宣布：您被选中为新茶语评茶官啦，请及时处理。",
		AcceptObjectId: aO.Id,
	}
	// 发送消息给两个在线用户
	if err = AcceptMessageSendExceptUserId(s_u.Id, mess); err != nil {
		util.Warning(err, "Cannot send 2 acceptMessage")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶，请稍后再试。")
		return
	}

	t := fmt.Sprintf("你好，新开茶话会 %s 已准备妥当，稍等有缘茶友评审通过之后，即可启用。", new_ob.Title)
	// 提示用户草稿保存成功
	Report(w, r, t)

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
	objective_list, err := data.GetPublicObjectives(24)
	if err != nil {
		util.Info(err, " Cannot get objectives")
		Report(w, r, "你好，茶博士失魂鱼，未能获取缘分茶话会资料，请稍后再试。")
		return
	}
	len := len(objective_list)
	if len == 0 {
		Report(w, r, "你好，山穷水尽疑无路，为何没有任何茶话会资料？请稍后再试。")
		return
	}

	// 如果茶话会状态是草围（未经邻座盲评审核的草稿）,对其名称和描述内容局部进行随机遮盖处理。
	// for i := range objective_list {
	// 	if objective_list[i].Class == 10 || objective_list[i].Class == 20 {
	// 		// 随机遮盖50%处理
	// 		objective_list[i].Title = MarsString(objective_list[i].Title, 50)
	// 		objective_list[i].Body = MarsString(objective_list[i].Body, 50)
	// 	}
	// }

	oSpD.ObjectiveBeanList, err = FetchObjectiveBeanList(objective_list)
	if err != nil {
		util.Warning(err, " Cannot read objective-bean list")
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你时刻忙碌奋斗着。")
		return
	}

	//检查用户是否已经登录
	s, err := Session(r)
	if err != nil {
		//未登录！游客
		//准备页面数据
		oSpD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		//迭代茶话会队列，把作者属性设置为false
		for i := range oSpD.ObjectiveBeanList {
			oSpD.ObjectiveBeanList[i].Objective.PageData.IsAuthor = false
		}

		//返回页面
		RenderHTML(w, &oSpD, "layout", "navbar.public", "objectives.square")
		return
	}
	//已登录
	sUser, err := s.User()
	if err != nil {
		util.Info(err, " Cannot get user from session")
		//跳转登录页面
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//已经登录！
	//准备页面数据
	oSpD.SessUser = sUser
	//检测u.Id == o.UserId是否这个茶话会作者
	for i := range oSpD.ObjectiveBeanList {
		if oSpD.ObjectiveBeanList[i].Objective.UserId == sUser.Id {
			oSpD.ObjectiveBeanList[i].Objective.PageData.IsAuthor = true
		} else {
			oSpD.ObjectiveBeanList[i].Objective.PageData.IsAuthor = false
		}
	}
	RenderHTML(w, &oSpD, "layout", "navbar.private", "objectives.square")

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
		Report(w, r, "你好，茶博士迷糊了，请稍后再试。")
		return
	}
	// 根据uuid查询茶话会资料
	ob := data.Objective{
		Uuid: uuid}
	if err = ob.GetByUuid(); err != nil {
		Report(w, r, "你好，茶博士摸摸满头大汗，居然自言自语说外星人把这个茶话会资料带走了。")
		return
	}

	switch ob.Class {
	case 1, 2:
		break
	case 10, 20:
		Report(w, r, "你好，这个茶话会需要等待友邻盲评通过之后才能启用呢。")
		return
	default:
		Report(w, r, "你好，这个茶话会主人据说因为很cool，资料似乎被外星人看中带走了。")
		return
	}
	oD.ObjectiveBean, err = FetchObjectiveBean(ob)
	if err != nil {
		util.Warning(err, " Cannot read objective-bean list")
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你时刻忙碌奋斗着。")
		return
	}
	//fetch public projects
	project_list, _ := oD.ObjectiveBean.Objective.GetPublicProjects()
	oD.ProjectBeanList, err = FetchProjectBeanList(project_list)
	if err != nil {
		util.Warning(err, " Cannot read objective-bean list")
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你时刻忙碌奋斗着。")
		return
	}
	//检查用户是否已经登录
	s, err := Session(r)
	if err != nil {
		//未登录！
		// 准备页面数据
		oD.ObjectiveBean.Objective.PageData.IsAuthor = false
		oD.SessUser = data.User{
			Id:   0,
			Name: "游客",
			// 用户足迹
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}

		oD.IsGuest = true
		// 是否受邀请团队成员？
		oD.IsInvited = false

		//配置公开导航条的茶话会详情页面
		RenderHTML(w, &oD, "layout", "navbar.public", "objective.detail")
		return
	}
	//已经登录！
	s_u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	oD.IsGuest = false
	// 记录用户查询的资讯
	if err = RecordLastQueryPath(s_u.Id, r.URL.Path, r.URL.RawQuery); err != nil {
		util.Warning(err, s_u.Email, " Cannot record last query path")
	}
	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery
	oD.SessUser = s_u

	// 如果这个茶话会是封闭式，检查当前用户是否属于受邀请团队成员
	if ob.Class == 2 {
		ok, err := oD.ObjectiveBean.Objective.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Warning(err, " Cannot read objective-bean list")
			Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你时刻忙碌着。")
			return
		}
		oD.IsInvited = ok
	}

	//检测u.Id == o.UserId是否这个茶话会主人（作者）
	if s_u.Id == oD.ObjectiveBean.Author.Id {
		//是作者
		//准备页面数据
		oD.ObjectiveBean.Objective.PageData.IsAuthor = true
		//配置私有导航条的茶话会详情页面
		RenderHTML(w, &oD, "layout", "navbar.private", "objective.detail")
		return
	} else {
		//不是作者
		oD.ObjectiveBean.Objective.PageData.IsAuthor = false

		//配置私有导航条的茶话会详情页面
		RenderHTML(w, &oD, "layout", "navbar.private", "objective.detail")
		return
	}

}
