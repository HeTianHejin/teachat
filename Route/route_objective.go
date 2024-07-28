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
		CreateObjectivePage(w, r)
	case "POST":
		CreateObjective(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// GET /objective/create
// 返回objective.new页面
func CreateObjectivePage(w http.ResponseWriter, r *http.Request) {
	//尝试从http请求中读取用户会话信息
	s, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//根据会话从数据库中读取当前用户的信息
	u, _ := s.User()

	// 填写页面数据
	SessUser := u
	// 给请求用户返回新建茶话会页面
	util.GenerateHTML(w, &SessUser, "layout", "navbar.private", "objective.new")
}

// POST /objective/create
// create the objective
// 创建新茶话会
func CreateObjective(w http.ResponseWriter, r *http.Request) {
	s, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Report(w, r, "您好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return
	}
	u, _ := s.User()
	// 默认的茶话会封面图名
	cover := "default-ob-cover"
	// 读取http请求中form的数据
	title := r.PostFormValue("name")
	body := r.PostFormValue("description")
	class, _ := strconv.Atoi(r.PostFormValue("class"))

	// 检查用户是否已经创建过相同名字的茶话会
	//objective, err := u.ObjectiveByName(name)

	// 检测一下name是否>2中文字，body是否在17-456中文字，
	// 如果不是，返回错误信息
	if util.CnStrLen(title) < 2 {
		util.Report(w, r, "您好，粗声粗气的茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if util.CnStrLen(body) < 17 || util.CnStrLen(body) > 456 {
		util.Report(w, r, "您好，茶博士迷糊了，竟然说字数太少或者太多记不住，请确认后再试。")
		return
	}

	var ob data.Objective

	switch class {
	case 10:
		//如果class=10开放式茶话会草围
		//尝试保存新茶话会
		ob, err = u.CreateObjective(title, body, cover, class)
		if err != nil {
			// 撤回（删除）发送给两个用户的消息，测试未做 ～～～～～～～～～:P

			// 记录错误，提示用户新开茶话会未成功
			util.Warning(err, " Cannot create objective")
			util.Report(w, r, "您好，偷来梨蕊三分白，借得梅花一缕魂。")
			return
		}

	case 20:
		//如果class=20封闭式茶话会草围，需要读取指定茶团号TeamId列表

		tIds_str := r.PostFormValue("invite-team-ids")

		//用正则表达式检测茶团号TeamIds，是否符合“整数，整数，整数...”的格式
		if !data.VerifyTeamIdListFormat(tIds_str) {
			util.Warning(err, " TeamId list format is wrong")
			util.Report(w, r, "您好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
			return
		}
		//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId,以便处理
		te_ids_str := strings.Split(tIds_str, ",")
		// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
		if len(te_ids_str) > util.Config.MaxInviteTeams {
			util.Warning(err, " Too many team ids")
			util.Report(w, r, "您好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，茶壶不够用，请确认后再试。")
			return
		}
		team_id_list := make([]int, 0, util.Config.MaxInviteTeams)
		for _, t_id := range te_ids_str {
			te_id_int, _ := strconv.Atoi(t_id)
			team_id_list = append(team_id_list, te_id_int)
		}
		//尝试保存新茶话会草稿
		ob, err = u.CreateObjective(title, body, cover, class)
		if err != nil {
			// 撤回发送给两个用户的消息，测试未做 ～～～～～～～～～:P

			util.Warning(err, " Cannot create objective")
			util.Report(w, r, "您好，茶博士迷糊了，未能创建茶话会，请稍后再试。")
			return
		}

		// 迭代team_id_list，尝试保存新封闭式茶话会草围邀请的茶团
		for _, team_id := range team_id_list {
			obInviTeams := data.ObjectiveInvitedTeam{
				ObjectiveId: ob.Id,
				TeamId:      team_id,
			}
			if err = obInviTeams.Save(); err != nil {
				// 撤回发送给两个用户的消息，测试未做 ～～～～～～～～～:P

				util.Warning(err, " Cannot create objectiveLicenseTeam")
			}
		}
	default:
		// 未知的茶话会属性
		util.Warning(err, " Unknown objective class")
		util.Report(w, r, "您好，茶博士还在研究茶话会的围字是不是有四种写法，忘记创建茶话会了，请稍后再试。")
		return
	}

	// 创建一条友邻盲评,是否接纳 新茶的记录
	aO := data.AcceptObject{
		ObjectId:   ob.Id,
		ObjectType: 1,
	}
	if err = aO.Create(); err != nil {
		util.Warning(err, "Cannot create accept_object")
		util.Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	// 发送盲评请求消息给两个在线用户
	// 构造消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您好，茶博士隆重宣布：您被选中为新茶语评茶官啦，请及时处理。",
		AcceptObjectId: aO.Id,
	}
	// 发送消息给两个在线用户
	if err = AcceptMessageSendExceptUserId(u.Id, mess); err != nil {
		util.Warning(err, "Cannot send 2 acceptMessage")
		util.Report(w, r, "您好，茶博士失魂鱼，未能创建新茶，请稍后再试。")
		return
	}

	t := fmt.Sprintf("您好，新开茶话会 %s 已准备妥当，稍等有缘茶友评审通过之后，即可启用。", ob.Title)
	// 提示用户草稿保存成功
	util.Report(w, r, t)

}

// GET /objective/square
// show the random objectives
// 根据用户是否登录显示不同导航条的茶话会广场页面
func ObjectiveSquare(w http.ResponseWriter, r *http.Request) {
	var err error
	var oSpD data.ObjectiveSquarePData

	// 如何排序茶话会，是个问题！按照圆桌会议平等约定，应该是轮流出现在茶话会广场才合适，而且，应该是按人头出现计数，而不是按热度或者建的茶话会数量。

	// 每次展示一打=12个茶话会
	//limit := 12
	// 用缘分（随机选中）选取12个用户的12茶话会的模式

	//oSpD.ObjectiveList, err = data.GetRandomObjectives(limit)
	// 用热度（按照评论数量排序）选取12个用户的12茶话会的模式

	// test获取所有茶话会
	oSpD.ObjectiveList, err = data.GetAllObjectives()
	if err != nil {
		util.Info(err, " Cannot get objectives")
		util.Report(w, r, "您好，茶博士失魂鱼，未能获取缘分茶话会资料，请稍后再试。")
		return
	}
	// 缩短.Body
	for i := range oSpD.ObjectiveList {
		oSpD.ObjectiveList[i].Body = util.Substr(oSpD.ObjectiveList[i].Body, 86)
	}

	// 如果茶话会状态是草围（未经邻座盲评审核的草稿）,对其名称和描述内容局部进行随机遮盖处理。
	for i := range oSpD.ObjectiveList {
		if oSpD.ObjectiveList[i].Class == 10 || oSpD.ObjectiveList[i].Class == 20 {
			// 随机遮盖50%处理
			oSpD.ObjectiveList[i].Title = util.MarsString(oSpD.ObjectiveList[i].Title, 50)
			oSpD.ObjectiveList[i].Body = util.MarsString(oSpD.ObjectiveList[i].Body, 50)
		}
	}

	//检查用户是否已经登录
	s, err := util.Session(r)
	if err != nil {
		//未登录！游客
		//准备页面数据
		oSpD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		//迭代茶话会队列，把作者属性设置为false
		for i := range oSpD.ObjectiveList {
			oSpD.ObjectiveList[i].PageData.IsAuthor = false
		}

		//返回页面
		util.GenerateHTML(w, &oSpD, "layout", "navbar.public", "objectives.square")
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
	for i := range oSpD.ObjectiveList {
		if oSpD.ObjectiveList[i].UserId == sUser.Id {
			oSpD.ObjectiveList[i].PageData.IsAuthor = true
		} else {
			oSpD.ObjectiveList[i].PageData.IsAuthor = false
		}
	}
	util.GenerateHTML(w, &oSpD, "layout", "navbar.private", "objectives.square")

}

// GET /objective/detail?id=
// show the details of the objective
// 读取指定的uuid茶话会详情
func ObjectiveDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	var obDetailPD data.ObjectiveDetailPageData
	vals := r.URL.Query()
	uuid := vals.Get("id")
	if uuid == "" {
		util.Report(w, r, "您好，茶博士迷糊了，请稍后再试。")
		return
	}
	// 根据uuid查询茶话会资料
	obDetailPD.Objective, err = data.GetObjectiveByUuid(uuid)
	if err != nil {
		util.Report(w, r, "您好，茶博士摸摸满头大汗，居然自言自语说外星人把这个茶话会资料带走了。")
		return
	}

	switch obDetailPD.Objective.Class {
	case 1, 2:
		break
	case 10, 20:
		util.Report(w, r, "您好，这个茶话会需要等待友邻盲评通过之后才能启用呢。")
		return
	default:
		util.Report(w, r, "您好，这个茶话会主人据说因为很帅，资料似乎被外星人看中带走了。")
		return
	}

	// 准备页面数据
	obDetailPD.Objective.PageData.IsAuthor = false
	obDetailPD.SessUser = data.User{
		Id:   0,
		Name: "游客",
	}
	obDetailPD.ProjectList, _ = obDetailPD.Objective.Projects()
	//截短project.Body
	for i := range obDetailPD.ProjectList {
		obDetailPD.ProjectList[i].Body = util.Substr(obDetailPD.ProjectList[i].Body, 86)
	}

	//检查用户是否已经登录
	s, err := util.Session(r)
	if err != nil {
		//未登录！
		//配置公开导航条的茶话会详情页面
		util.GenerateHTML(w, &obDetailPD, "layout", "navbar.public", "objective.detail")
		return
	}
	//已经登录！
	sUser, _ := s.User()
	obDetailPD.SessUser = sUser
	//检测u.Id == o.UserId是否这个茶话会作者
	if sUser.Id == obDetailPD.Objective.UserId {
		//是作者
		//配置私有导航条的茶话会详情页面
		//准备页面数据PageData
		obDetailPD.Objective.PageData.IsAuthor = true
		util.GenerateHTML(w, &obDetailPD, "layout", "navbar.private", "objective.detail")
		return
	} else {
		//不是作者
		//配置私有导航条的茶话会详情页面
		util.GenerateHTML(w, &obDetailPD, "layout", "navbar.private", "objective.detail")
		return
	}

}
