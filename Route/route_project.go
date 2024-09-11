package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// 处理新建茶台的操作处理器
func HandleNewProject(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//请求表单
		GetCreateProjectPage(w, r)
	case "POST":
		//处理表单
		CreateProject(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// POST /v1/project/new
// 用户在某个指定茶话会新开一张茶台
func CreateProject(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, err := s.User()
	if err != nil {
		util.Danger(err, " Cannot get user from session")
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	//获取用户提交的表单数据
	title := r.PostFormValue("name")
	body := r.PostFormValue("description")
	uuid := r.PostFormValue("uuid")
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
	_, err = data.GetTeamMemberByTeamIdAndUserId(team_id, u.Id)
	if err != nil {
		util.Info(err, " Cannot get team member")
		Report(w, r, "您好，如果你不是团中人，就不能以该团成员身份入围开台呢，未能创建新茶台，请稍后再试。")
		return
	}

	// 检测一下name是否>2中文字，desc是否在17-456中文字，
	// 如果不是，返回错误信息
	if CnStrLen(title) < 2 || CnStrLen(title) > 36 {
		util.Info(err, "Project name is too short")
		Report(w, r, "您好，粗声粗气的茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if CnStrLen(body) < 17 || CnStrLen(body) > 456 {
		util.Info(err, " Project description is too long or too short")
		Report(w, r, "您好，茶博士迷糊了，竟然说字数太少或者太多记不住，请确认后再试。")
		return
	}

	//获取目标茶话会
	ob := data.Objective{
		Uuid: uuid}
	if err = ob.GetByUuid(); err != nil {
		util.Info(err, " Cannot get objective")
		Report(w, r, "您好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		return
	}

	var proj data.Project

	// 根据茶话会属性判断
	// 检查一下该茶话会是否草围（待盲评审核状态）
	switch ob.Class {
	case 10, 20:
		// 该茶话会是草围,尚未启用，不能新开茶台
		Report(w, r, "您好，这个茶话会尚未启用。")
		return

	case 1:
		// 该茶话会是开放式茶话会，可以新开茶台
		// 检查提交的class值是否有效，必须为10或者20
		if class == 10 {
			// 创建开放式草台
			proj, err = u.CreateProject(title, body, ob.Id, class, team_id)
			if err != nil {
				util.Warning(err, " Cannot create project")
				Report(w, r, "您好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}
		} else if class == 20 {
			tIds_str := r.PostFormValue("invite-team-ids")
			//用正则表达式检测一下s，是否符合“整数，整数，整数...”的格式
			if !VerifyTeamIdListFormat(tIds_str) {
				util.Info(err, " TeamId list format is wrong")
				Report(w, r, "您好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
				return
			}
			//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId
			team_ids_str := strings.Split(tIds_str, ",")
			// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
			if len(team_ids_str) > util.Config.MaxInviteTeams {
				util.Info(err, " Too many team ids")
				Report(w, r, "您好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，开水不够用，请确认后再试。")
				return
			}
			team_id_list := make([]int, 0, util.Config.MaxInviteTeams)
			for _, te_id_str := range team_ids_str {
				t_id_int, _ := strconv.Atoi(te_id_str)
				team_id_list = append(team_id_list, t_id_int)
			}

			//创建封闭式草台
			proj, err = u.CreateProject(title, body, ob.Id, class, team_id)
			if err != nil {
				util.Warning(err, " Cannot create project")
				Report(w, r, "您好，斜阳寒草带重门，苔翠盈铺雨后盆。")
				return
			}
			// 迭代team_id_list，尝试保存新封闭式茶台邀请的茶团
			for _, team_id := range team_id_list {
				poInviTeams := data.ProjectInvitedTeam{
					ProjectId: proj.Id,
					TeamId:    team_id,
				}
				if err = poInviTeams.Create(); err != nil {
					util.Warning(err, " Cannot save invited teams")
					Report(w, r, "您好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		} else {
			Report(w, r, "您好，茶博士摸摸头，说看不懂拟开新茶台是否封闭式，请确认。")
			return
		}

	case 2:
		// 封闭式茶话会
		// 检查用户是否可以在此茶话会下新开茶台
		ok, err := ob.IsInvitedMember(u.Id)
		if !ok {
			// 当前用户不是茶话会邀请团队成员，不能新开茶台
			util.Warning(err, " Cannot create project")
			Report(w, r, "您好，茶博士惊讶地说，不是此茶话会邀请团队成员不能开新茶台，请确认。")
			return
		}
		// 当前用户是茶话会邀请团队成员，可以新开茶台
		if class == 10 {
			Report(w, r, "您好，封闭式茶话会内不能开启开放式茶台，请确认后再试。")
			return
		}
		if class == 20 {
			tIds_str := r.PostFormValue("invite-team-ids")
			//用正则表达式检测一下s，是否符合“整数，整数，整数...”的格式
			if !VerifyTeamIdListFormat(tIds_str) {
				util.Info(err, " TeamId list format is wrong")
				Report(w, r, "您好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
				return
			}
			//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId
			team_ids_str := strings.Split(tIds_str, ",")
			// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
			if len(team_ids_str) > util.Config.MaxInviteTeams {
				util.Info(err, " Too many team ids")
				Report(w, r, "您好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，开水不够用，请确认后再试。")
				return
			}
			team_id_list := make([]int, 0, util.Config.MaxInviteTeams)
			for _, te_id_str := range team_ids_str {
				t_id_int, _ := strconv.Atoi(te_id_str)
				team_id_list = append(team_id_list, t_id_int)
			}

			//创建茶台
			proj, err := u.CreateProject(title, body, ob.Id, class, team_id)
			if err != nil {
				util.Warning(err, " Cannot create project")
				Report(w, r, "您好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
				return
			}
			// 迭代team_id_list，尝试保存新封闭式茶台邀请的茶团
			for _, team_id := range team_id_list {
				poInviTeams := data.ProjectInvitedTeam{
					ProjectId: proj.Id,
					TeamId:    team_id,
				}
				if err = poInviTeams.Create(); err != nil {
					util.Warning(err, " Cannot save invited teams")
					Report(w, r, "您好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		}

	default:
		// 该茶话会属性不合法
		util.Info(err, " Project class is not valid")
		Report(w, r, "您好，茶博士摸摸头，竟然说这个茶话会被外星人霸占了，请确认后再试。")
		return
	}
	// 创建一条友邻盲评,是否接纳 新茶的记录
	aO := data.AcceptObject{
		ObjectId:   proj.Id,
		ObjectType: 1,
	}
	if err = aO.Create(); err != nil {
		util.Warning(err, "Cannot create accept_object")
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	// 发送盲评请求消息给两个在线用户
	// 构造消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
		AcceptObjectId: aO.Id,
	}
	// 发送消息给两个在线用户
	err = AcceptMessageSendExceptUserId(u.Id, mess)
	if err != nil {
		util.Danger(err, " Cannot send message")
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}

	// 提示用户草台保存成功
	t := fmt.Sprintf("您好，新开茶话会 %s 已准备妥当，稍等有缘茶友评审通过之后，即可启用。", proj.Title)
	// 提示用户草稿保存成功
	Report(w, r, t)

}

// GET
// 渲染创建新茶台表单页面
func GetCreateProjectPage(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取提交的数据，确定是哪一个茶话会需求新开茶台
	vals := r.URL.Query()
	uuid := vals.Get("id")
	var oD data.ObjectiveDetail
	// 获取指定的目标茶话会
	o := data.Objective{
		Uuid: uuid}
	if err = o.GetByUuid(); err != nil {
		util.Danger(err, " Cannot read project")
		Report(w, r, "您好，茶博士失魂鱼，未能找到茶台，请稍后再试。")
		return
	}
	//根据会话从数据库中读取当前用户的信息
	s_u, s_default_team, s_survival_teams, err := FetchUserRelatedData(s)
	if err != nil {
		Report(w, r, "您好，�����������了，��没有��水未能加载��话会页面，请稍后再试。")
		return
	}
	// 填写页面数据
	// 填写页面会话用户资料
	oD.SessUser = s_u
	oD.SessUserDefaultTeam = s_default_team
	oD.SessUserSurvivalTeams = s_survival_teams
	oD.ObjectiveBean, err = GetObjectiveBean(o)
	if err != nil {
		Report(w, r, "您好，������失������，未能找到��台，请稍后再试。")
		return
	}

	// 检查当前用户是否可以在此茶话会下新开茶台
	// 首先检查茶话会属性，class=1开放式，class=2封闭式，
	// 如果是开放式，则可以在茶话会下新开茶台
	// 如果是封闭式，则需要看围主指定了那些茶团成员可以开新茶台，如果围主没有指定，则不能新开茶台
	switch o.Class {
	case 1:
		// 开放式茶话会，可以在茶话会下新开茶台
		// 向用户返回添加指定的茶台的表单页面
		GenerateHTML(w, &oD, "layout", "navbar.private", "project.new")
		return
	case 2:
		// 封闭式茶话会，需要看围主指定了那些茶团成员可以开新茶台，如果围主没有指定，则不能新开茶台
		//检查team_ids是否为空
		// 围主没有指定茶团成员，不能新开茶台
		// 当前用户是茶话会邀请团队成员，可以新开茶台
		ok, err := o.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Warning(err, " Cannot read project")
			Report(w, r, "您好，������失������，未能找到��台，请稍后再试。")
			return
		}
		if ok {
			GenerateHTML(w, &oD, "layout", "navbar.private", "project.new")
			return
		} else {
			// 当前用户不是茶话会邀请团队成员，不能新开茶台
			Report(w, r, "您好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
			return
		}

		// 非法茶话会属性，不能新开茶台
	default:
		Report(w, r, "您好，茶博士失魂鱼，竟然说受邀请品茶名单被外星人霸占了，请稍后再试。")
		return
	}

}

// GET /v1/project/detail
// 展示指定的UUID茶台详情
func ProjectDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	var pD data.ProjectDetail
	// 读取用户提交的查询参数
	vals := r.URL.Query()
	uuid := vals.Get("id")
	// 获取请求的茶台详情
	pD.Project, err = data.GetProjectByUuid(uuid)
	if err != nil {
		util.Warning(err, " Cannot read project")
		Report(w, r, "您好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
		return
	}
	//检查project.Class=1 or 2,否则属于未经 友邻盲评 通过的草稿，不允许查看
	if pD.Project.Class != 1 && pD.Project.Class != 2 {
		Report(w, r, "您好，荡昏寐，饮之以茶。请稍后再试。")
		return
	}

	pD.Master, err = pD.Project.User()
	if err != nil {
		util.Warning(err, " Cannot read project user")
		Report(w, r, "您好，霁月难逢，彩云易散。请稍后再试。")
		return
	}
	pD.MasterTeam, err = data.GetTeamById(pD.Project.TeamId)
	if err != nil {
		util.Warning(err, " Cannot read project team")
		Report(w, r, "您好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺。请稍后再试。")
		return
	}

	// 准备页面数据
	if pD.Project.Class == 1 {
		pD.Open = true
	} else {
		pD.Open = false
	}
	if pD.IsEdited {
		pD.IsEdited = true
	} else {
		pD.IsEdited = false
	}

	pD.QuoteObjective, err = pD.Project.Objective()
	if err != nil {
		util.Warning(err, " Cannot read objective")
		Report(w, r, "您好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺。")
		return
	}
	// 截短此引用的茶围内容以方便展示
	pD.QuoteObjective.Body = Substr(pD.QuoteObjective.Body, 66)
	pD.QuoteObjectiveAuthor, err = pD.QuoteObjective.User()
	if err != nil {
		util.Warning(err, " Cannot read objective author")
		Report(w, r, "您好，������失������，��然说指定的����名单��然保存失败，请确认后再试。")
		return
	}
	pD.QuoteObjectiveAuthorTeam, err = data.GetTeamById(pD.QuoteObjective.TeamId)
	if err != nil {
		util.Warning(err, " Cannot read objective author team")
		Report(w, r, "您好，������失������，��然说指定的����名单��然保存失败，请确认后再试。")
		return
	}

	var oabList []data.ThreadBean
	// 读取全部茶议资料
	threadlist, err := pD.Project.Threads()
	if err != nil {
		util.Warning(err, " Cannot read threads given project")
		Report(w, r, "您好，满头大汗的茶博士说，倦绣佳人幽梦长，金笼鹦鹉唤茶汤。")
		return
	}

	len := len(threadlist)
	pD.ThreadCount = len
	// 检测pageData.ThreadList数量是否超过一打dozen
	if len > 12 {
		pD.IsOverTwelve = true
	} else {
		//测试时都设为true显示效果 🐶🐶🐶
		pD.IsOverTwelve = true
	}
	// 获取茶议和作者相关资料荚
	oabList, err = GetThreadBeanList(threadlist)
	if err != nil {
		util.Warning(err, " Cannot read thread-bean list")
		Report(w, r, "您好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你忙碌中...")
		return
	}
	pD.ThreadBeanList = oabList

	// 获取会话session
	sess, err := Session(r)
	if err != nil {
		// 未登录，游客
		// 填写页面数据
		pD.Project.PageData.IsAuthor = false
		pD.IsInput = false
		pD.IsGuest = true
		pD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		// 返回给浏览者茶台详情页面
		GenerateHTML(w, &pD, "layout", "navbar.public", "project.detail")
		return
	}

	// 已登陆用户
	pD.IsGuest = false
	//从会话查获当前浏览用户资料荚
	s_u, s_default_team, s_survival_teams, err := FetchUserRelatedData(sess)
	if err != nil {
		util.Warning(err, " Cannot get user-related data from session")
		Report(w, r, "您好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	pD.SessUser = s_u
	pD.SessUserDefaultTeam = s_default_team
	pD.SessUserSurvivalTeams = s_survival_teams

	//如果这是class=2封闭式茶台，需要检查当前浏览用户是否可以创建新茶议
	if pD.Project.Class == 2 {
		// ���查当前用户是否是��话会��请��队成员
		ok, err := pD.Project.IsInvitedMember(s_u.Id)
		if err != nil {
			Report(w, r, "您好，������失������，未能读取专����台资料。")
			return
		}
		if ok {
			// 当前用户是��话会��请��队成员，可以新开茶议
			pD.IsInput = true
		} else {
			// 当前用户不是��话会��请��队成员，不能新开茶议
			pD.IsInput = false
		}
	} else {
		// 开放式茶议，任何人都可以新开茶议
		pD.IsInput = true
	}

	// 检查是否台主？
	pD.Project.PageData.IsAuthor = false
	if s_u.Id == pD.Project.UserId {
		pD.Project.PageData.IsAuthor = true
	}

	GenerateHTML(w, &pD, "layout", "navbar.private", "project.detail")
}
