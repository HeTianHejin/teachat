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
	ouid := r.PostFormValue("uuid")
	clas, _ := strconv.Atoi(r.PostFormValue("class"))

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
	obje, err := data.GetObjectiveByUuid(ouid)
	if err != nil {
		util.Info(err, " Cannot get objective")
		Report(w, r, "您好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		return
	}
	var proj data.Project
	// 	//检测一下用户是否有相同名字的茶台
	// 	if data.HasProjectName(n) {
	// 		util.Info(err, " Project name is already used")
	// 		util.Pop_message(w, r, "您好，茶博士迷糊了，竟然说字数太少或者太多记不住，请确认后再试。")
	// 		return
	// 	}

	// 根据茶话会属性判断
	// 检查一下该茶话会是否草围（待盲评审核状态）
	switch obje.Class {
	case 10, 20:
		// 该茶话会是草围,尚未启用，不能新开茶台
		Report(w, r, "您好，这个茶话会尚未启用。")
		return

	case 1:
		// 该茶话会是开放式茶话会，可以新开茶台
		// 检查提交的class值是否有效，必须为10或者20
		if clas == 10 {
			// 创建开放式草台
			proj, err = u.CreateProject(title, body, obje.Id, clas)
			if err != nil {
				util.Warning(err, " Cannot create project")
				Report(w, r, "您好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}
		} else if clas == 20 {
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
			proj, err = u.CreateProject(title, body, obje.Id, clas)
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
				if err = poInviTeams.Save(); err != nil {
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
		ok := isUserInvitedByObjective(obje, u)
		if !ok {
			// 当前用户不是茶话会邀请团队成员，不能新开茶台
			util.Warning(err, " Cannot create project")
			Report(w, r, "您好，茶博士惊讶地说，不是此茶话会邀请团队成员不能开新茶台，请确认。")
			return
		}
		// 当前用户是茶话会邀请团队成员，可以新开茶台
		if clas == 10 {
			Report(w, r, "您好，封闭式茶话会内不能开启开放式茶台，请确认后再试。")
			return
		}
		if clas == 20 {
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
			proj, err := u.CreateProject(title, body, obje.Id, clas)
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
				if err = poInviTeams.Save(); err != nil {
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
		Content:        "您好，茶博士隆重宣布：您被茶棚选中为新茶语评审官啦，请及时处理。",
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
	//获取用户资料
	u, _ := s.User()
	// 读取提交的数据，确定是哪一个茶话会需求新开茶台
	vals := r.URL.Query()
	uuid := vals.Get("id")
	var obD data.ObjectiveDetail
	// 获取指定的目标茶话会
	ob, err := data.GetObjectiveByUuid(uuid)
	if err != nil {
		util.Danger(err, " Cannot read project")
		Report(w, r, "您好，茶博士失魂鱼，未能找到茶台，请稍后再试。")
		return
	}
	// 填写页面会话用户资料
	obD.SessUser = u

	// 检查当前用户是否可以在此茶话会下新开茶台
	// 首先检查茶话会属性，class=1开放式，class=2封闭式，
	// 如果是开放式，则可以在茶话会下新开茶台
	// 如果是封闭式，则需要看围主指定了那些茶团成员可以开新茶台，如果围主没有指定，则不能新开茶台
	switch ob.Class {
	case 1:
		// 开放式茶话会，可以在茶话会下新开茶台
		// 向用户返回添加指定的茶台的表单页面
		GenerateHTML(w, &obD.ObjectiveBean, "layout", "navbar.private", "project.new")
		return
	case 2:
		// 封闭式茶话会，需要看围主指定了那些茶团成员可以开新茶台，如果围主没有指定，则不能新开茶台
		//检查team_ids是否为空
		// 围主没有指定茶团成员，不能新开茶台
		// 当前用户是茶话会邀请团队成员，可以新开茶台
		ok := isUserInvitedByObjective(ob, u)
		if ok {
			GenerateHTML(w, &obD, "layout", "navbar.private", "project.new")
			return
		}

		// 当前用户不是茶话会邀请团队成员，不能新开茶台
		Report(w, r, "您好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
		return

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
	var pd data.ProjectDetail
	// 读取用户提交的查询参数
	vals := r.URL.Query()
	uuid := vals.Get("id")
	// 获取请求的茶台详情
	pd.Project, err = data.GetProjectByUuid(uuid)
	if err != nil {
		util.Warning(err, " Cannot read project")
		Report(w, r, "您好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
		return
	}
	pd.Master, err = pd.Project.User()
	if err != nil {
		util.Warning(err, " Cannot read project user")
		Report(w, r, "您好，霁月难逢，彩云易散。请稍后再试。")
		return
	}
	pd.MasterTeam, _ = pd.Master.GetLastDefaultTeam()
	// 准备页面数据
	if pd.Project.Class == 1 {
		pd.Open = true
	} else {
		pd.Open = false
	}
	if pd.IsEdited {
		pd.IsEdited = true
	} else {
		pd.IsEdited = false
	}

	pd.QuoteObjective, err = pd.Project.Objective()
	if err != nil {
		util.Warning(err, " Cannot read objective")
		Report(w, r, "您好，������失������，��然说指定的����名单��然保存失败，请确认后再试。")
		return
	}
	// 截短此引用的茶围内容以方便展示
	pd.QuoteObjective.Body = Substr(pd.QuoteObjective.Body, 66)
	pd.QuoteObjectiveAuthor, err = pd.QuoteObjective.User()
	if err != nil {
		util.Warning(err, " Cannot read objective author")
		Report(w, r, "您好，������失������，��然说指定的����名单��然保存失败，请确认后再试。")
		return
	}
	pd.QuoteObjectiveAuthorTeam, err = pd.QuoteObjectiveAuthor.GetLastDefaultTeam()
	if err != nil {
		util.Warning(err, " Cannot read objective author team")
		Report(w, r, "您好，������失������，��然说指定的����名单��然保存失败，请确认后再试。")
		return
	}

	var oabList []data.ThreadBean
	// 读取全部茶议资料
	threadlist, err := pd.Project.Threads()
	if err != nil {
		util.Warning(err, " Cannot read threads given project")
		Report(w, r, "您好，满头大汗的茶博士说，倦绣佳人幽梦长，金笼鹦鹉唤茶汤。")
		return
	}

	len := len(threadlist)
	pd.ThreadCount = len
	// 检测pageData.ThreadList数量是否超过一打dozen
	if len > 12 {
		pd.IsOverTwelve = true
	} else {
		//测试时都设为true显示效果 🐶🐶🐶
		pd.IsOverTwelve = true
	}
	// 获取茶议和作者相关资料夹
	oabList, err = GetThreadBeanList(threadlist)
	if err != nil {
		util.Warning(err, " Cannot read thread-bean list")
		Report(w, r, "您好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你忙碌中。")
		return
	}
	pd.ThreadBeanList = oabList

	// 获取会话session
	s, err := Session(r)
	if err != nil {
		// 未登录，游客
		// 填写页面数据
		pd.Project.PageData.IsAuthor = false
		pd.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		// 返回给浏览者茶台详情页面
		GenerateHTML(w, &pd, "layout", "navbar.public", "project.detail")
		return
	}
	// 获取当前会话用户资料
	u, _ := s.User()
	pd.SessUser = u
	// 检查是否台主？
	pd.Project.PageData.IsAuthor = false
	if u.Id == pd.Project.UserId {
		pd.Project.PageData.IsAuthor = true
	}

	GenerateHTML(w, &pd, "layout", "navbar.private", "project.detail")
}
