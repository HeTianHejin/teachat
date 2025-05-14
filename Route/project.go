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

// 准备用户相关数据
func prepareUserData(sess *data.Session) (*data.UserData, error) {
	user, defaultFamily, survivalFamilies, defaultTeam, survivalTeams, defaultPlace, places, err := FetchSessionUserRelatedData(*sess)
	if err != nil {
		return nil, err
	}

	// 添加特殊选项
	survivalFamilies = append(survivalFamilies, data.UnknownFamily)
	survivalTeams = append(survivalTeams, FreelancerTeam)

	return &data.UserData{
		User:             user,
		DefaultFamily:    defaultFamily,
		SurvivalFamilies: survivalFamilies,
		DefaultTeam:      defaultTeam,
		SurvivalTeams:    survivalTeams,
		DefaultPlace:     defaultPlace,
		BindPlaces:       places,
	}, nil
}

// 准备茶围页面数据
func prepareObjectivePageData(objective data.Objective, userData *data.UserData) (*data.ObjectiveDetail, error) {
	objectiveBean, err := FetchObjectiveBean(objective)
	if err != nil {
		return nil, err
	}

	return &data.ObjectiveDetail{
		SessUser:                 userData.User,
		SessUserDefaultFamily:    userData.DefaultFamily,
		SessUserSurvivalFamilies: userData.SurvivalFamilies,
		SessUserDefaultTeam:      userData.DefaultTeam,
		SessUserSurvivalTeams:    userData.SurvivalTeams,
		SessUserDefaultPlace:     userData.DefaultPlace,
		SessUserBindPlaces:       userData.BindPlaces,
		ObjectiveBean:            objectiveBean,
	}, nil
}

// POST /v1/project/approve
// 茶围管理员选择某个茶台入围，记录它 --【Tencent ai 协助】
func ProjectApprove(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		Report(w, r, "你好，茶博士失魂鱼，未能记录入围茶台，请稍后再试。")
		return
	}
	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		Report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶台，请确认后再试。")
		return
	}

	//获取目标茶台
	pr := data.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid)
		Report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶台，请确认后再试。")
		return
	}
	//读取目标茶围
	ob, err := pr.Objective()
	if err != nil {
		util.Debug(" Cannot get objective", ob.Id)
		Report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		return
	}
	//检查用户是否有权限处理这个请求
	admin_team, err := data.GetTeam(ob.TeamId)
	if err != nil {
		util.Debug(" Cannot get team", ob.TeamId)
		Report(w, r, "你好，茶博士失魂鱼，未能找到指定的团队，请确认后再试。")
		return
	}
	is_admin, err := admin_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug(" Cannot get team", ob.TeamId)
		Report(w, r, "你好，茶博士失魂鱼，未能找到指定的团队，请确认后再试。")
		return
	}
	if !is_admin {
		//不是茶围管理员，无权处理
		Report(w, r, "你好，茶博士面无表情，说你没有权限处理这个入围操作，请确认。")
		return
	}

	//记录入围的茶台
	new_project_approved := data.ProjectApproved{
		ObjectiveId: ob.Id,
		ProjectId:   pr.Id,
		UserId:      s_u.Id,
	}
	if err = new_project_approved.Create(); err != nil {
		util.Debug(" Cannot create project approved", err)
		Report(w, r, "你好，茶博士失魂鱼，未能记录入围茶台，请稍后再试。")
		return
	}

	//返回成功
	Report(w, r, "你好，茶博士微笑，已成功记录入围茶台，请稍后刷新页面查看。")
}

// 处理新建茶台的操作处理器
func HandleNewProject(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//请求表单
		NewProjectGet(w, r)
	case http.MethodPost:
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
		util.Debug(" Cannot get user from session", err)
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
		util.Debug("Failed to convert class to int", err)
		return
	}
	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.Debug(team_id, "Failed to convert team_id to int")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.Debug("Failed to convert family_id to int", err)
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶台，请稍后再试。")
		return
	}
	valid, err := validateTeamAndFamilyParams(w, r, team_id, family_id, s_u.Id)
	if !valid && err == nil {
		return // 参数不合法，已经处理了错误
	}
	if err != nil {
		// 处理数据库错误
		util.Debug("验证提交的团队和家庭id出现数据库错误", team_id, family_id, err)
		Report(w, r, "你好，成员资格检查失败，请确认后再试。")
		return
	}
	//获取目标茶话会
	t_ob := data.Objective{Uuid: ob_uuid}
	if err = t_ob.GetByUuid(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			util.Debug("茶话会不存在", ob_uuid, err)
			Report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		} else {
			util.Debug("获取茶话会失败", ob_uuid, err)
			Report(w, r, "你好，茶博士失魂鱼，系统繁忙，请稍后再试。")
		}
		return
	}
	// 检查在此茶围下是否已经存在相同名字的茶台
	count_title, err := data.CountProjectByTitleObjectiveId(title, t_ob.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		util.Debug(" Cannot get count of project by title and objective id", err)
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

	place_uuid := r.PostFormValue("place_uuid")
	place := data.Place{Uuid: place_uuid}
	if err = place.GetByUuid(); err != nil {
		util.Debug(" Cannot get place", err)
		Report(w, r, "你好，茶博士服务中，眼镜都模糊了，也未能找到你提交的活动地方资料，请确认后再试。")
		return
	}

	// 检测一下name是否>2中文字，desc是否在17-456中文字，
	// 如果不是，返回错误信息
	if CnStrLen(title) < 2 || CnStrLen(title) > 36 {
		util.Debug("Project name is too short", err)
		Report(w, r, "你好，粗声粗气的茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if CnStrLen(body) < 17 || CnStrLen(body) > 456 {
		util.Debug(" Project description is too long or too short", err)
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
				util.Debug(" Cannot create open project", err)
				Report(w, r, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}

		} else if class == 20 {
			tIds_str := r.PostFormValue("invite_ids")
			//用正则表达式检测一下s，是否符合“整数，整数，整数...”的格式
			if !Verify_id_slice_Format(tIds_str) {
				util.Debug(" TeamId slice format is wrong", err)
				Report(w, r, "你好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
				return
			}
			//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId
			team_ids_str := strings.Split(tIds_str, ",")
			// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
			if len(team_ids_str) > int(util.Config.MaxInviteTeams) {
				util.Debug(" Too many team ids", err)
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
				util.Debug(" Cannot create close project", err)
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
					util.Debug(" Cannot save invited teams", err)
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
			util.Debug(" Cannot create project", err)
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
				util.Debug(" TeamId slice format is wrong", err)
				Report(w, r, "你好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
				return
			}
			//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId
			team_ids_str := strings.Split(tIds_str, ",")
			// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
			if len(team_ids_str) > int(util.Config.MaxInviteTeams) {
				util.Debug(" Too many team ids", err)
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
				util.Debug("Cannot create project", err)
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
					util.Debug(" Cannot save invited teams", err)
					Report(w, r, "你好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		}

	default:
		// 该茶话会属性不合法
		util.Debug(" Project class is not valid", err)
		Report(w, r, "你好，茶博士摸摸头，竟然说这个茶话会被外星人霸占了，请确认后再试。")
		return
	}

	// 保存草台活动地方
	pp := data.ProjectPlace{
		ProjectId: new_proj.Id,
		PlaceId:   place.Id}

	if err = pp.Create(); err != nil {
		util.Debug(" Cannot create project place", err)
		Report(w, r, "你好，茶博士抹了抹汗，竟然说茶台地方保存失败，请确认后再试。")
		return
	}

	// 创建一条友邻蒙评,是否接纳 新茶的记录
	accept_object := data.AcceptObject{
		ObjectId:   new_proj.Id,
		ObjectType: 2,
	}
	if err = accept_object.Create(); err != nil {
		util.Debug("Cannot create accept_object", err)
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
		util.Debug(" Cannot send message", err)
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
	// 1. 检查用户会话
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 2. 获取并验证茶话会UUID
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		Report(w, r, "你好，茶博士失魂鱼，请指定要加入的茶话会。")
		return
	}

	// 3. 获取茶话会详情
	objective := data.Objective{Uuid: uuid}
	if err := objective.GetByUuid(); err != nil {
		util.Debug("获取茶话会失败", "uuid", uuid, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			Report(w, r, "你好，茶博士失魂鱼，未能找到您指定的茶话会。")
		} else {
			Report(w, r, "你好，茶博士失魂鱼，系统繁忙，请稍后再试。")
		}
		return
	}

	// 4. 获取用户相关数据
	sessUserData, err := prepareUserData(&sess)
	if err != nil {
		util.Debug("准备用户数据失败", "error", err)
		Report(w, r, "你好，三人行，必有大佬焉，请稍后再试。")
		return
	}

	// 5. 准备页面数据
	pageData, err := prepareObjectivePageData(objective, sessUserData)
	if err != nil {
		util.Debug("准备页面数据失败", "error", err)
		Report(w, r, "你好，茶博士失魂鱼，未能找到茶围资料，请稍后再试。")
		return
	}

	// 6. 检查茶台创建权限
	if !checkCreateProjectPermission(objective, sessUserData.User.Id, w, r) {
		return
	}

	// 7. 渲染创建表单
	RenderHTML(w, &pageData, "layout", "navbar.private", "project.new")
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

	pr := data.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot read project", err)
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
		util.Debug(" Cannot read project", pr.Uuid, err)
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
		util.Debug(" Cannot read objective", err)
		Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺。")
		return
	}
	pD.QuoteObjectiveBean, err = FetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot read objective", err)
		Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺。")
		return
	}
	// 截短此引用的茶围内容以方便展示
	pD.QuoteObjectiveBean.Objective.Body = Substr(pD.QuoteObjectiveBean.Objective.Body, 168)

	var tb_slice []data.ThreadBean
	// 读取全部茶议资料
	thread_slice, err := pD.ProjectBean.Project.Threads()
	if err != nil {
		util.Debug(" Cannot read threads given project", err)
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
	tb_slice, err = FetchThreadBeanSlice(thread_slice)
	if err != nil {
		util.Debug(" Cannot read thread-bean slice", err)
		Report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。茶博士为你忙碌中...")
		return
	}
	pD.ThreadBeanSlice = tb_slice

	// 获取茶台项目活动地方
	pD.Place, err = pD.ProjectBean.Project.Place()
	if err != nil {
		util.Debug(" Cannot read project place", err)
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
	s_u, s_default_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchSessionUserRelatedData(s)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", s.Email, err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	//把系统默认家庭资料加入s_survival_families
	s_survival_families = append(s_survival_families, data.UnknownFamily)
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

	is_master, err := checkProjectMasterPermission(&pr, s_u.Id)
	if err != nil {
		util.Debug("Permission check failed", "user", s_u.Id, "error", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	pD.IsMaster = is_master

	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("Admin permission check failed",
			"userId", s_u.Id,
			"objectiveId", ob.Id,
			"error", err,
		)
		Report(w, r, "你好，茶博士说：玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	pD.IsAdmin = is_admin

	// 用户足迹
	pD.SessUser.Footprint = r.URL.Path
	pD.SessUser.Query = r.URL.RawQuery

	RenderHTML(w, &pD, "layout", "navbar.private", "project.detail")
}
