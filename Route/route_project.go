package route

import (
	"database/sql"
	"errors"
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
		NewProjectGet(w, r)
	case "POST":
		//处理表单
		NewProjectPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// POST /v1/project/new
// 用户在某个指定茶话会新开一张茶台
func NewProjectPost(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get user from session")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	//获取用户提交的表单数据
	title := r.PostFormValue("name")
	body := r.PostFormValue("description")
	ob_uuid := r.PostFormValue("ob_uuid")
	class, err := strconv.Atoi(r.PostFormValue("class"))
	if err != nil {
		util.ScaldingTea(util.LogError(err), "Failed to convert class to int")
		return
	}
	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.ScaldingTea(util.LogError(err), team_id, "Failed to convert team_id to int")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.ScaldingTea(util.LogError(err), "Failed to convert family_id to int")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	//获取目标茶话会
	t_ob := data.Objective{
		Uuid: ob_uuid}
	if err = t_ob.GetByUuid(); err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get objective")
		Report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		return
	}
	// 检查在此茶围下是否已经存在相同名字的茶台
	count_title, err := data.CountProjectByTitleObjectiveId(title, t_ob.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		util.ScaldingTea(util.LogError(err), " Cannot get count of project by title and objective id")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	//如果已经存在相同名字的茶台，返回错误信息
	if count_title > 0 {
		Report(w, r, "你好，已经存在相同名字的茶台，请更换一个名称后再试。")
		return
	}

	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"

	//提交的茶团id,是team.id，检查是否成员
	if team_id != 0 {
		// check the given team_id is valid
		_, err = data.GetMemberByTeamIdUserId(team_id, s_u.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				//util.PanicTea(util.LogError(err), " Cannot get member by team id and user id")
				Report(w, r, "你好，如果你不是团中人，就不能以该团成员身份入围开台呢，未能创建新茶台，请稍后再试。")
				return
			} else {
				util.ScaldingTea(util.LogError(err), " Cannot get member by team id and user id")
				Report(w, r, "你好，茶博士眼镜失踪了，未能创建新茶台，请稍后再试。")
				return
			}
		}
	}
	//提交的茶团id,是family.id，检查是否家庭成员
	// check submit family_id is valid
	if family_id != 0 {
		family := data.Family{
			Id: family_id,
		}
		is_member, err := family.IsMember(s_u.Id)
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot get family member by family id and user id")
			Report(w, r, "你好，茶博士眼镜失踪，未能创建新茶台，请稍后再试。")
			return
		}
		if !is_member {
			util.ScaldingTea(util.LogError(err), " Cannot get family member by family id and user id")
			Report(w, r, "你好，家庭成员资格检查失败，请确认后再试。")
			return
		}
	}

	place_uuid := r.PostFormValue("place_uuid")
	place := data.Place{Uuid: place_uuid}
	if err = place.GetByUuid(); err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get place")
		Report(w, r, "你好，茶博士服务中，眼镜都模糊了，也未能找到你提交的活动地方资料，请确认后再试。")
		return
	}

	// 检测一下name是否>2中文字，desc是否在17-456中文字，
	// 如果不是，返回错误信息
	if CnStrLen(title) < 2 || CnStrLen(title) > 36 {
		util.ScaldingTea(util.LogError(err), "Project name is too short")
		Report(w, r, "你好，粗声粗气的茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if CnStrLen(body) < 17 || CnStrLen(body) > 456 {
		util.ScaldingTea(util.LogError(err), " Project description is too long or too short")
		Report(w, r, "你好，茶博士傻眼了，竟然说字数太少或者太多记不住，请确认后再试。")
		return
	}

	new_proj := data.Project{
		UserId:      s_u.Id,
		Title:       title,
		Body:        body,
		ObjectiveId: t_ob.Id,
		Class:       class,
		TeamId:      team_id,
		FamilyId:    family_id,
		IsPrivate:   is_private,
		Cover:       "default-pr-cover",
	}

	// 根据茶话会属性判断
	// 检查一下该茶话会是否草围（待蒙评审核状态）
	switch t_ob.Class {
	case 10, 20:
		// 该茶话会是草围,尚未启用，不能新开茶台
		Report(w, r, "你好，这个茶话会尚未启用。")
		return

	case 1:
		// 该茶话会是开放式茶话会，可以新开茶台
		// 检查提交的class值是否有效，必须为10或者20
		if class == 10 {
			// 创建开放式草台
			if err = new_proj.Create(); err != nil {
				util.ScaldingTea(util.LogError(err), " Cannot create open project")
				Report(w, r, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}

		} else if class == 20 {
			tIds_str := r.PostFormValue("invite_ids")
			//用正则表达式检测一下s，是否符合“整数，整数，整数...”的格式
			if !Verify_id_slice_Format(tIds_str) {
				util.ScaldingTea(util.LogError(err), " TeamId slice format is wrong")
				Report(w, r, "你好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
				return
			}
			//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId
			team_ids_str := strings.Split(tIds_str, ",")
			// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
			if len(team_ids_str) > int(util.Config.MaxInviteTeams) {
				util.ScaldingTea(util.LogError(err), " Too many team ids")
				Report(w, r, "你好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，请确认后再试。")
				return
			}
			team_id_slice := make([]int, 0, util.Config.MaxInviteTeams)
			for _, te_id_str := range team_ids_str {
				t_id_int, _ := strconv.Atoi(te_id_str)
				team_id_slice = append(team_id_slice, t_id_int)
			}

			//创建封闭式草台
			if err = new_proj.Create(); err != nil {
				util.ScaldingTea(util.LogError(err), " Cannot create close project")
				Report(w, r, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}
			// 迭代team_id_slice，尝试保存新封闭式茶台邀请的茶团
			for _, team_id := range team_id_slice {
				poInviTeams := data.ProjectInvitedTeam{
					ProjectId: new_proj.Id,
					TeamId:    team_id,
				}
				if err = poInviTeams.Create(); err != nil {
					util.ScaldingTea(util.LogError(err), " Cannot save invited teams")
					Report(w, r, "你好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		} else {
			Report(w, r, "你好，茶博士摸摸头，说看不懂拟开新茶台是否封闭式，请确认。")
			return
		}

	case 2:
		// 封闭式茶话会
		// 检查用户是否可以在此茶话会下新开茶台
		ok, err := t_ob.IsInvitedMember(s_u.Id)
		if !ok {
			// 当前用户不是茶话会邀请团队成员，不能新开茶台
			util.ScaldingTea(util.LogError(err), " Cannot create project")
			Report(w, r, "你好，茶博士惊讶地说，不是此茶话会邀请团队成员不能开新茶台，请确认。")
			return
		}
		// 当前用户是茶话会邀请团队成员，可以新开茶台
		if class == 10 {
			Report(w, r, "你好，封闭式茶话会内不能开启开放式茶台，请确认后再试。")
			return
		}
		if class == 20 {
			tIds_str := r.PostFormValue("invite_ids")
			//用正则表达式检测一下s，是否符合“整数，整数，整数...”的格式
			if !Verify_id_slice_Format(tIds_str) {
				util.ScaldingTea(util.LogError(err), " TeamId slice format is wrong")
				Report(w, r, "你好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
				return
			}
			//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId
			team_ids_str := strings.Split(tIds_str, ",")
			// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
			if len(team_ids_str) > int(util.Config.MaxInviteTeams) {
				util.ScaldingTea(util.LogError(err), " Too many team ids")
				Report(w, r, "你好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，开水不够用，请确认后再试。")
				return
			}
			team_id_slice := make([]int, 0, util.Config.MaxInviteTeams)
			for _, te_id_str := range team_ids_str {
				t_id_int, _ := strconv.Atoi(te_id_str)
				team_id_slice = append(team_id_slice, t_id_int)
			}

			//创建茶台
			if err = new_proj.Create(); err != nil {
				util.ScaldingTea(util.LogError(err), " Cannot create project")
				Report(w, r, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}
			// 迭代team_id_slice，尝试保存新封闭式茶台邀请的茶团
			for _, team_id := range team_id_slice {
				poInviTeams := data.ProjectInvitedTeam{
					ProjectId: new_proj.Id,
					TeamId:    team_id,
				}
				if err = poInviTeams.Create(); err != nil {
					util.ScaldingTea(util.LogError(err), " Cannot save invited teams")
					Report(w, r, "你好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		}

	default:
		// 该茶话会属性不合法
		util.ScaldingTea(util.LogError(err), " Project class is not valid")
		Report(w, r, "你好，茶博士摸摸头，竟然说这个茶话会被外星人霸占了，请确认后再试。")
		return
	}

	// 保存草台活动地方
	pp := data.ProjectPlace{
		ProjectId: new_proj.Id,
		PlaceId:   place.Id}

	if err = pp.Create(); err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot create project place")
		Report(w, r, "你好，闪电考拉抹了抹汗，竟然说茶台地方保存失败，请确认后再试。")
		return
	}

	// 创建一条友邻蒙评,是否接纳 新茶的记录
	accept_object := data.AcceptObject{
		ObjectId:   new_proj.Id,
		ObjectType: 2,
	}
	if err = accept_object.Create(); err != nil {
		util.ScaldingTea(util.LogError(err), "Cannot create accept_object")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	// 发送蒙评请求消息给两个在线用户
	// 构造消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
		AcceptObjectId: accept_object.Id,
	}
	// 发送消息给两个在线用户
	err = TwoAcceptMessagesSendExceptUserId(s_u.Id, mess)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot send message")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}

	// 提示用户草台保存成功
	t := fmt.Sprintf("你好，新开茶话会 %s 已准备妥当，稍等有缘茶友评审通过之后，即可启用。", new_proj.Title)
	// 提示用户草稿保存成功
	Report(w, r, t)

}

// GET /v1/project/new?uuid=xxx
// 渲染创建新茶台表单页面
func NewProjectGet(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取提交的数据，确定是哪一个茶话会需求新开茶台
	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	var oD data.ObjectiveDetail
	// 获取指定的目标茶话会
	o := data.Objective{
		Uuid: uuid}
	if err = o.GetByUuid(); err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project")
		Report(w, r, "你好，茶博士失魂鱼，未能找到茶台，请稍后再试。")
		return
	}
	//根据会话从数据库中读取当前用户的团队,地方信息，
	s_u, s_default_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		Report(w, r, "你好，三人行，必有大佬焉，请稍后再试。")
		return
	}
	//默认和常用地方

	// 填写页面数据
	// 填写页面会话用户资料
	oD.SessUser = s_u
	oD.SessUserDefaultFamily = s_default_family
	oD.SessUserSurvivalFamilies = s_survival_families
	oD.SessUserDefaultTeam = s_default_team
	oD.SessUserSurvivalTeams = s_survival_teams
	oD.SessUserDefaultPlace = s_default_place
	oD.SessUserBindPlaces = s_places
	oD.ObjectiveBean, err = FetchObjectiveBean(o)
	if err != nil {
		Report(w, r, "你好，茶博士失魂鱼，未能找到茶围资料，请稍后再试。")
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
		RenderHTML(w, &oD, "layout", "navbar.private", "project.new")
		return
	case 2:
		// 封闭式茶话会，需要看围主指定了那些茶团成员可以开新茶台，如果围主没有指定，则不能新开茶台
		//检查team_ids是否为空
		// 围主没有指定茶团成员，不能新开茶台
		// 当前用户是茶话会邀请团队成员，可以新开茶台
		ok, err := o.IsInvitedMember(s_u.Id)
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot read objective Invited-list")
			Report(w, r, "你好，茶博士满头大汗说，邀请品茶名单被狗叼进了花园，请稍候。")
			return
		}
		if ok {
			RenderHTML(w, &oD, "layout", "navbar.private", "project.new")
			return
		} else {
			// 当前用户不是茶话会邀请团队成员，不能新开茶台
			Report(w, r, "你好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
			return
		}

		// 非法茶话会属性，不能新开茶台
	default:
		Report(w, r, "你好，茶博士失魂鱼，竟然说受邀请品茶名单被外星人霸占了，请稍后再试。")
		return
	}

}

// GET /v1/project/detail?id=
// 展示指定UUID茶台详情
func ProjectDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	var pD data.ProjectDetail
	// 读取用户提交的查询参数
	vals := r.URL.Query()
	uuid := vals.Get("id")
	// 获取请求的茶台详情
	pr, err := data.GetProjectByUuid(uuid)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project")
		Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
		return
	}
	//检查project.Class=1 or 2,否则属于未经 友邻蒙评 通过的草稿，不允许查看
	if pr.Class != 1 && pr.Class != 2 {
		Report(w, r, "你好，荡昏寐，饮之以茶。请稍后再试。")
		return
	}

	pD.ProjectBean, err = FetchProjectBean(pr)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project", pr.Uuid)
		Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
		return
	}

	// 准备页面数据
	if pD.ProjectBean.Project.Class == 1 {
		pD.Open = true
	} else {
		pD.Open = false
	}

	ob, err := pD.ProjectBean.Project.Objective()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read objective")
		Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺。")
		return
	}
	pD.QuoteObjectiveBean, err = FetchObjectiveBean(ob)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read objective")
		Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺。")
		return
	}
	// 截短此引用的茶围内容以方便展示
	pD.QuoteObjectiveBean.Objective.Body = Substr(pD.QuoteObjectiveBean.Objective.Body, 168)

	var oabSlice []data.ThreadBean
	// 读取全部茶议资料
	thread_slice, err := pD.ProjectBean.Project.Threads()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read threads given project")
		Report(w, r, "你好，满头大汗的茶博士说，倦绣佳人幽梦长，金笼鹦鹉唤茶汤。")
		return
	}

	len := len(thread_slice)
	// .ThreadCount数量
	pD.ThreadCount = len
	// 检测pageData.ThreadSlice数量是否超过一打dozen
	if len > 12 {
		pD.IsOverTwelve = true
	} else {
		//测试时都设为true显示效果 🐶🐶🐶
		pD.IsOverTwelve = true
	}
	// .ThreadIsApprovedCount数量
	ta := data.ThreadApproved{
		ProjectId: pD.ProjectBean.Project.Id,
	}
	pD.ThreadIsApprovedCount = ta.CountByProjectId()

	// 获取茶议和作者相关资料荚
	oabSlice, err = FetchThreadBeanSlice(thread_slice)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read thread-bean slice")
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。闪电考拉为你忙碌中...")
		return
	}
	pD.ThreadBeanSlice = oabSlice

	// 获取茶台项目活动地方
	pD.Place, err = pD.ProjectBean.Project.Place()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project place")
		Report(w, r, "你好，满头大汗的茶博士唱，过高花已妒，请稍后再试。")
		return
	}

	// 获取会话session
	s, err := Session(r)
	if err != nil {
		// 未登录，游客
		// 填写页面数据
		pD.ProjectBean.Project.PageData.IsAuthor = false
		pD.IsInput = false
		pD.IsGuest = true
		pD.IsAdmin = false
		pD.IsMaster = false

		pD.SessUser = data.User{
			Id:        0,
			Name:      "游客",
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		// 返回给浏览者茶台详情页面
		RenderHTML(w, &pD, "layout", "navbar.public", "project.detail")
		return
	}

	// 已登陆用户
	pD.IsGuest = false

	//从会话查获当前浏览用户资料荚
	s_u, s_default_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get user-related data from session")
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	//把系统默认家庭资料加入s_survival_families
	s_survival_families = append(s_survival_families, DefaultFamily)
	//把系统默认团队资料加入s_survival_teams
	s_survival_teams = append(s_survival_teams, FreelancerTeam)

	pD.SessUser = s_u
	pD.SessUserDefaultFamily = s_default_family
	pD.SessUserSurvivalFamilies = s_survival_families
	pD.SessUserDefaultTeam = s_default_team
	pD.SessUserSurvivalTeams = s_survival_teams
	pD.SessUserDefaultPlace = s_default_place
	pD.SessUserBindPlaces = s_places

	//如果这是class=2封闭式茶台，需要检查当前浏览用户是否可以创建新茶议
	if pD.ProjectBean.Project.Class == 2 {
		// 是封闭式茶台，需要检查当前用户身份是否受邀请茶团的成员，以决定是否允许发言
		ok, err := pD.ProjectBean.Project.IsInvitedMember(s_u.Id)
		if err != nil {
			Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？你真的是受邀请茶团成员吗？")
			return
		}
		if ok {
			// 当前用户是本茶话会邀请$团队成员，可以新开茶议
			pD.IsInput = true
		} else {
			// 当前会话用户不是本茶话会邀请$团队成员，不能新开茶议
			pD.IsInput = false
		}
	} else {
		// 开放式茶议，任何人都可以新开茶议
		pD.IsInput = true
	}

	//会话用户是否是作者
	if pD.ProjectBean.Project.UserId == s_u.Id {
		// 是作者
		pD.ProjectBean.Project.PageData.IsAuthor = true
	} else {
		// 不是作者
		pD.ProjectBean.Project.PageData.IsAuthor = false
	}

	//读取茶台管理团队资料
	pr_team, err := data.GetTeam(pr.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get team")
		Report(w, r, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	// 检查是否茶台管理员，
	is_master, err := pr_team.IsMember(s_u.Id)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get team-core-members")
		Report(w, r, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	//标记为管理员
	pD.IsMaster = is_master

	//获取管理这个茶围的团队
	admin_team, err := data.GetTeam(ob.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get team")
		Report(w, r, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	// 查是否茶围管理员
	is_admin, err := admin_team.IsMember(s_u.Id)
	if err != nil {

		util.ScaldingTea(util.LogError(err), " Cannot get team-core-members")
		Report(w, r, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	pD.IsAdmin = is_admin

	// 用户足迹
	pD.SessUser.Footprint = r.URL.Path
	pD.SessUser.Query = r.URL.RawQuery

	RenderHTML(w, &pD, "layout", "navbar.private", "project.detail")
}
