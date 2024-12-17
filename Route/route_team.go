package route

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/team_member/application/check?id=
// 查看加盟某个茶团的全部新的加盟申请书
func MemberApplyCheck(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	//获取参数
	uuid := r.URL.Query().Get("id")

	//查询目标茶团
	team, err := data.GetTeamByUuid(uuid)
	if err != nil {
		util.Danger(err, uuid, "Cannot get team by given uuid")
		Report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	//检查用户是否茶团成员，非成员不能查看加盟申请书
	_, err = data.GetTeamMemberByTeamIdAndUserId(team.Id, s_u.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			Report(w, r, "你好，茶博士失魂鱼，你不是茶团成员，无法查看申请书。")
			return
		} else {
			util.Danger(err, "Cannot get team member by team id and user id")
			Report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
			return
		}
	}

	//查询茶团全部新的加盟申请书，包含已查看但未处理的
	applies, err := data.GetMemberApplicationByTeamIdAndStatus(team.Id)
	if err != nil {
		util.Danger(err, "Cannot get applys by team id")
		Report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	apply_bean_list, err := GetMemberApplicationBeanList(applies)
	if err != nil {
		util.Danger(err, "Cannot get apply bean list")
		Report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}

	//截短MemberApplication.Content为66字，方便布局列表预览
	for _, bean := range apply_bean_list {
		bean.MemberApplication.Content = Substr(bean.MemberApplication.Content, 66)
	}

	var mAL data.MemberApplicationList
	//填写页面数据
	mAL.SessUser = s_u
	mAL.Team = team
	mAL.MemberApplicationBeanList = apply_bean_list

	// 渲染页面
	RenderHTML(w, &mAL, "layout", "navbar.private", "team.applications_check")

}

// GET /v1/teams/application
// 显示申请加入的全部茶团的表单页面
func ApplyTeams(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//查询用户全部加盟申请书
	applies, err := data.GetMemberApplies(s_u.Id)
	if err != nil {
		util.Danger(err, "Cannot get applys by user id")
		Report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	apply_bean_list, err := GetMemberApplicationBeanList(applies)
	if err != nil {
		util.Danger(err, "Cannot get apply bean list")
		Report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	//截短MemberApplication.Content为66字，方便布局列表预览
	for _, bean := range apply_bean_list {
		bean.MemberApplication.Content = Substr(bean.MemberApplication.Content, 66)
	}

	var mAL data.MemberApplicationList
	//查询用户全部加盟申请书
	mAL.SessUser = s_u
	mAL.MemberApplicationBeanList = apply_bean_list

	// 渲染页面
	RenderHTML(w, &mAL, "layout", "navbar.private", "teams.application")

}

// GET /v1/team/new
// 显示创建新茶团的表单页面
func NewTeam(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var tpd data.TeamSquare
	tpd.SessUser = s_u
	RenderHTML(w, &tpd, "layout", "navbar.private", "team.new")
}

// POST /v1/team/create
// 创建新茶团
func CreateTeam(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "你好，茶博士失魂鱼，未能开新茶团，请稍后再试。")
		return
	}
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}

	count, err := s_u.CountTeamsByFounderId()
	if err != nil {
		util.Danger(err, "connot count teams given founder_id")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	if count > int(util.Config.MaxTeamsCount) {
		Report(w, r, "你好，三月香巢已垒成，梁间燕子太无情。开团数量太多了，未能创建新茶团。")
		return
	}

	n := r.PostFormValue("name")
	// 茶团名称是否在4-24中文字符
	l := CnStrLen(n)
	if l < 4 || l > 24 {
		Report(w, r, "你好，茶博士摸摸头，竟然说茶团名字字数太多或者太少，未能创建新茶团。")
		return
	}

	abbr := r.PostFormValue("abbreviation")
	// 队名简称是否在4-6中文字符
	lenA := CnStrLen(abbr)
	if lenA < 4 || lenA > 6 {
		Report(w, r, "你好，茶博士摸摸头，竟然说队名简称字数太多或者太少，未能创建新茶团。")
		return
	}

	mission := r.PostFormValue("mission")
	// 检测mission是否在17-456中文字符
	lenM := CnStrLen(mission)
	if lenM < 17 || lenM > 456 {
		Report(w, r, "你好，茶博士摸摸头，竟然说愿景字数太多或者太少，未能创建新茶团。")
		return
	}
	class, err := strconv.Atoi(r.PostFormValue("class"))
	if err != nil {
		util.Info(err, " Cannot convert class to int")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	// group_id, err := strconv.Atoi(r.PostFormValue("group_id"))
	// if err != nil {
	// 	util.Info(err, " Cannot convert group_id to int")
	// 	Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
	// 	return
	// }

	//检测class是否合规
	switch class {
	case 10, 20:
		break
	default:
		Report(w, r, "你好，茶博士摸摸头，竟然说茶团类别太多或者太少，未能创建新茶团。")
		return
	}
	//检测同名的team是否已经存在，团队不允许同名
	_, err = data.GetTeamByName(n)
	if err == nil {
		Report(w, r, "你好，茶博士摸摸头，竟然说这个茶团名称已经被占用，请换一个更响亮的团名，再试一次。")
		return
	}
	//检测是否包含“自由人”这样的保留关键字，不允许，以免误导其他茶友
	if strings.Contains(n, "自由人") || strings.Contains(n, "Freelancer") {
		Report(w, r, "你好，茶团名称不能包含“自由人”这样的保留关键字，请换一个更响亮的团名，再试一次。")
		return
	}
	_, err = data.GetTeamByAbbreviation(abbr)
	if err == nil {
		// 重复的简称
		Report(w, r, "你好，茶博士摸摸头，竟然说这个茶团简称已经被占用，请换一个更响亮的团名，再试一次。")
		return
	}
	//检测abbr中是否包含“自由人”或者“Freelancer”这样不允许使用的保留关键字
	if strings.Contains(abbr, "自由人") || strings.Contains(abbr, "Freelancer") {
		Report(w, r, "你好，茶团简称不能包含“自由人”或者“Freelancer”这样的保留关键字，请换一个更响亮的团名简称，再试一次。")
		return
	}

	// 将NewTeam草稿存入数据库，class=10/20
	logo := "teamLogo"
	team, err := s_u.CreateTeam(n, abbr, mission, logo, class, 1)
	if err != nil {
		util.Info(err, " At create team")
		Report(w, r, "你好，茶博士失魂鱼，暂未能创建你的天命使团，请稍后再试。")
		return
	}

	// 创建一条友邻盲评,是否接纳 新茶团的记录
	aO := data.AcceptObject{
		ObjectId:   team.Id,
		ObjectType: 5,
	}
	if err = aO.Create(); err != nil {
		util.Warning(err, "Cannot create accept_object")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
	}

	// 发送盲评请求消息给两个在线用户
	//构造消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
		AcceptObjectId: aO.Id,
	}
	//发送消息
	if err = AcceptMessageSendExceptUserId(s_u.Id, mess); err != nil {
		Report(w, r, "你好，茶博士迷路了，未能发送盲评请求消息。")
		return
	}
	t := fmt.Sprintf("你好，新开茶团 %s 已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", team.Name)
	// 提示用户草稿保存成功
	Report(w, r, t)

}

// GET /v1/teams/open
// 显示茶棚全部开放式茶团列表信息
func OpenTeams(w http.ResponseWriter, r *http.Request) {
	var err error
	var tS data.TeamSquare

	team_list, err := data.GetOpenTeams()
	if err != nil {
		util.Info(err, " Cannot get open teams")
		Report(w, r, "你好，茶博士失魂鱼，未能获取茶团详细信息，请稍后再试。")
		return
	}
	tS.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	// 用户是否已经登录?
	s, err := Session(r)
	if err != nil {
		tS.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		RenderHTML(w, &tS, "layout", "navbar.public", "teams.open", "teams.public")
		return
	}
	sUser, _ := s.User()
	tS.SessUser = sUser
	RenderHTML(w, &tS, "layout", "navbar.private", "teams.open", "teams.public")

}

// Get /v1/teams/closed
// 显示茶棚全部封闭式茶团
func ClosedTeams(w http.ResponseWriter, r *http.Request) {
	var err error
	var tS data.TeamSquare

	team_list, err := data.GetClosedTeams()
	if err != nil {
		util.Info(err, " Cannot get closed teams")
		Report(w, r, "你好，茶博士失魂鱼，未能获取茶团详细信息，请稍后再试。")
		return
	}
	tS.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	s, _ := Session(r)
	u, err := s.User()
	if err != nil {
		tS.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		RenderHTML(w, &tS, "layout", "navbar.public", "teams.closed", "teams.public")
		return
	}
	tS.SessUser = u

	RenderHTML(w, &tS, "layout", "navbar.private", "teams.closed", "teams.public")

}

// GET /v1/teams/hold
// 显示当前用户创建的全部茶团
func HoldTeams(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var ts data.TeamSquare
	team_list, err := s_u.HoldTeams()
	if err != nil {
		util.Info(err, " Cannot get hold teams")
		return
	}
	ts.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	ts.SessUser = s_u
	RenderHTML(w, &ts, "layout", "navbar.private", "teams.hold", "teams.public")
}

// GET /v1/teams/joined
// 显示用户已经加入的全部茶团的表单页面
func JoinedTeams(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var tS data.TeamSquare
	tS.SessUser = s_u
	team_list, err := s_u.SurvivalTeams()
	if err != nil {
		util.Info(err, " Cannot get joined teams")
		Report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
		return
	}
	tS.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}

	RenderHTML(w, &tS, "layout", "navbar.private", "teams.joined", "teams.public")
}

// GET /v1/teams/employed
// 显示用户任职高管的全部茶团表单页面
func EmployedTeams(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, _ := sess.User()

	var ts data.TeamSquare
	ts.SessUser = u
	team_list, err := u.CoreExecTeams()
	if err != nil {
		util.Info(err, " Cannot get employed teams")
		Report(w, r, "你好，茶博士必须先找到自己的高度近视眼镜，再帮您查询资料。请稍后再试。")
		return
	}
	ts.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	RenderHTML(w, &ts, "layout", "navbar.private", "teams.employed", "teams.public")
}

// GET /v1/team/quit

// GET /v1/team/detail?id=
// 显示茶团详细信息
func TeamDetail(w http.ResponseWriter, r *http.Request) {
	//获取茶团资料
	vals := r.URL.Query()
	uuid := vals.Get("id")
	te, err := data.GetTeamByUuid(uuid)
	if err != nil {
		util.Info(err, " Cannot get team")
		Report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
		return
	}

	var tD data.TeamDetail
	tD.Team = te
	tD.CreatedAtDate = te.CreatedAtDate()
	f_u, err := te.Founder()
	if err != nil {
		util.Info(err, " Cannot get team founder")
		Report(w, r, "你好，闪电考拉为你极速效劳中，请稍后再试。")
		return
	}
	tD.Founder = f_u
	founder_default_team, _ := f_u.GetLastDefaultTeam()
	tD.FounderTeam = founder_default_team

	tD.TeamMemberCount = te.NumMembers()
	teamCoreMembers, err := te.CoreMembers()
	if err != nil {
		util.Info(err, " Cannot get team core member")
		Report(w, r, "你好，闪电考拉为你效劳中，请稍后再试。")
		return
	}
	teamNormalMembers, err := te.NormalMembers()
	if err != nil {
		util.Info(err, " Cannot get team normal member")
		Report(w, r, "你好，闪电考拉为你效劳中，请稍后再试。")
		return
	}
	// 准备页面数据
	var tc data.TeamMemberBean //核心成员资料荚
	var tn data.TeamMemberBean //普通成员资料荚
	var tcList []data.TeamMemberBean
	var tnList []data.TeamMemberBean
	//据teamMembers中的UserId获取User
	for _, member := range teamCoreMembers {
		user, err := data.GetUserById(member.UserId)
		if err != nil {
			util.Info(err, " Cannot get user")
			Report(w, r, "你好，闪电考拉为你效劳中，请稍后再试。")
			return
		}

		tc.User = user
		tc.AuthorTeam, err = user.GetLastDefaultTeam()
		if err != nil {
			util.Info(err, " Cannot get user's default team")
			Report(w, r, "你好，闪电考拉为你效劳中，请稍后再试。")
			return
		}
		tc.TeamMemberRole = member.Role
		tc.CreatedAtDate = member.CreatedAtDate()
		tcList = append(tcList, tc)
	}
	for _, member := range teamNormalMembers {
		user, err := data.GetUserById(member.UserId)
		if err != nil {
			util.Info(err, " Cannot get user")
			Report(w, r, "你好，闪电考拉为你疯狂效劳中，请稍后再试。")
			return
		}
		tn.User = user
		tn.AuthorTeam, err = user.GetLastDefaultTeam()
		if err != nil {
			util.Info(err, " Cannot get user's default team")
			Report(w, r, "你好，闪电考拉为你效劳中，请稍后再试。")
			return
		}
		tn.TeamMemberRole = member.Role
		tn.CreatedAtDate = member.CreatedAtDate()
		tnList = append(tnList, tn)
	}
	tD.CoreMemberDataList = tcList
	tD.NormalMemberDataList = tnList

	if tD.Team.Class == 1 {
		tD.Open = true
	} else {
		tD.Open = false
	}

	tD.IsCoreMember = false
	tD.IsMember = false
	s, err := Session(r)
	if err != nil {
		//游客
		tD.SessUser = data.User{
			Id:   0,
			Name: "游客",
			// 用户足迹
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		RenderHTML(w, &tD, "layout", "navbar.public", "team.detail")
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, " Cannot get user from session")
		Report(w, r, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}
	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery
	tD.SessUser = s_u
	//检查当前用户是否核心成员
	for _, member := range tD.CoreMemberDataList {
		if member.User.Id == s_u.Id {
			tD.IsCoreMember = true
			tD.IsMember = true
			break
		}
	}
	if !tD.IsCoreMember {
		//茶团发起人也是核心成员？
		f_teams, err := s_u.FounderTeams()
		if err != nil {
			util.Info(err, " Cannot get founderTeams given s_u")
			Report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
			return
		}
		//检查当前用户是否核心成员
		for _, team := range f_teams {
			if team.Id == tD.Team.Id {
				tD.IsCoreMember = true
				tD.IsMember = true
				break
			}
		}
	}

	if !tD.IsMember {
		//检查当前用户是否普通成员
		for _, member := range tD.NormalMemberDataList {
			if member.User.Id == s_u.Id {
				tD.IsMember = true
				break
			}
		}
	}

	tD.HasApplication = false
	if tD.IsCoreMember {
		//检查茶团是否有待处理的加盟申请书
		count, err := data.GetMemberApplicationByTeamIdAndStatusCount(tD.Team.Id)
		if err != nil {
			util.Info(err, "Cannot get member application count")
			Report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
			return
		}
		if count > 0 {
			tD.HasApplication = true
		}
	}

	RenderHTML(w, &tD, "layout", "navbar.private", "team.detail")

}

// v1/team/HandleManageTeam
func HandleManageTeam(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		HandleManageTeamGet(w, r)
	case "POST":
		HandleManageTeamPost(w, r)
	}
}

// POST /v1/team/manage
// 处理用户管理团队事务
func HandleManageTeamPost(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
}

// GET /v1/team/manage
// 显示用户管理茶团的表单页面
func HandleManageTeamGet(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Info(err, " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//一次管理一个茶团，根据提交的team.Uuid来确定
	//这个茶团是否存在？
	team, err := data.GetTeamByUuid(r.FormValue("id"))
	if err != nil {
		util.Info(err, " Cannot get this team")
		Report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}
	var tPD data.TeamDetail
	//检查一下当前用户是否有权管理这个茶团？即teamMember中Role为"ceo"或者founder
	//如果是创建人，那么就可以管理这个茶团
	fund, err := team.Founder()
	if err != nil {
		Report(w, r, "你好，茶博士未能找到此茶团发起人资料，请确认后再试。")
		return
	}
	if fund.Id == s_u.Id {
		//是创建人,可以管理这个茶团
		tPD.SessUser = s_u
		tPD.Team = team
		RenderHTML(w, &tPD, "layout", "navbar.private", "team.manage")
		return
	}
	//如果不是创建人，那么就检查一下是否是茶团的ceo
	manager, err := team.MemberCEO()
	if err != nil {
		//检查一下err是否是因为teamMember中没有这个team的ceo
		//如果是，那么就说明当前用户无权管理这个茶团
		//这是特殊情况？CEO空缺？例如唐僧被妖怪抓走了？
		if err == sql.ErrNoRows {
			Report(w, r, "你好，您无权管理此茶团。")
			return
		} else {
			//茶团已经设定了ceo，但是出现了其他错误
			util.Info(err, " Cannot get ceo of this team")
			Report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
			return
		}
	}
	if manager.UserId != s_u.Id {
		Report(w, r, "你好，您无权管理此茶团。")
		return
	}
	//是茶团的ceo，可以管理这个茶团
	tPD.SessUser = s_u
	tPD.Team = team
	RenderHTML(w, &tPD, "layout", "navbar.private", "team.manage")
}

// CoreManage() 处理用户管理团队核心成员角色事务
func CoreManage(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
	}
	u, _ := sess.User()
	//检查当前用户是否具备管理此core角色的权限
	//检查一下是否是茶团的ceo，如果是，那么就可以管理核心角色
	team_id, err := strconv.Atoi(r.FormValue("team_id"))
	if err != nil {
		util.Warning(err, " Cannot strconv this team_id")
		Report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}
	//这个茶团是否存在？
	team, err := data.GetTeamById(team_id)
	if err != nil {
		util.Info(err, " Cannot get ceo of this team")
		Report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}
	ceo, err := team.MemberCEO()
	if err != nil {
		//检查一下err是否是因为teamMember中没有这个team的ceo
		//如果是，那么就说明当前用户无权管理这个茶团
		//这是特殊情况？CEO空缺？例如唐僧被妖怪抓走了？
		if err == sql.ErrNoRows {
			Report(w, r, "你好，您无权管理此茶团。")
			return
		} else {
			//茶团已经设定了ceo，但是出现了其他错误
			util.Info(err, " Cannot get ceo of this team")
			Report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
			return
		}
	}
	if u.Id != ceo.UserId {
		Report(w, r, "你好，您无权管理此茶团。")
		return
	}

	//一次管理一个核心角色，根据提交的teamMember.id来确定
	//这个用户是否在这个茶团中？角色是否正确？
	member_id, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil {
		util.Warning(err, " Cannot convert member_id to int")
		Report(w, r, "茶博士失魂鱼，未能读取新泡茶议资料，请稍后再试。")
		return
	}
	role := r.FormValue("role")
	if role != "CTO" && role != "CMO" && role != "CFO" {
		Report(w, r, "你好，请选择正确的角色。")
		return
	}
	member, err := team.GetTeamMemberByRole(role)
	if err != nil {
		util.Info(err, " Cannot get this team member")
		Report(w, r, "你好，茶博士未能找到此茶团成员资料，请确认后再试。")
		return
	}
	if member.Id != member_id {
		Report(w, r, "你好，请选择正确的角色。")
		return
	}
	//如果是，那么就可以管理核心角色
	//读取提交的管理动作
	action := r.FormValue("action")
	switch action {
	case "appoint", "discharge":
		member.Role = role
		err := member.Update()
		if err != nil {
			Report(w, r, "你好，保存角色管理操作失败，请稍后再试。")
			return
		}
	default:
		Report(w, r, "你好，请选择正确的管理动作。")
		return
	}

}

// TeamAvatar() 处理茶团图标
func TeamAvatar(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		TeamAvatarGet(w, r)
	case "POST":
		TeamAvatarPost(w, r)
	}
}

// POST /v1/team/avatar
func TeamAvatarPost(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, _ := s.User()
	uuid := r.FormValue("team_uuid")
	team, err := data.GetTeamByUuid(uuid)
	if err != nil {
		Report(w, r, "你好，茶博士摸摸头，居然说找不到这个茶团相关资料。")
		return
	}

	// 检查team的创建者
	if u.Id == team.FounderId {
		//如果是创建者，那么就可以上传图标
		//读取上传的图标
		ProcessUploadAvatar(w, r, team.Uuid)
	}
	Report(w, r, "你好，上传茶团图标出现未知问题。")

}

// GET /v1/team/avatar?id=
func TeamAvatarGet(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, _ := s.User()
	uuid := r.URL.Query().Get("id")
	team, err := data.GetTeamByUuid(uuid)
	if err != nil {
		Report(w, r, "你好，茶博士摸摸头，居然说找不到这个茶团相关资料。")
		return
	}

	// 检查team的创建者
	if u.Id == team.FounderId {
		//如果是创建者，那么就可以上传图标

		RenderHTML(w, &uuid, "layout", "navbar.private", "team_avatar.upload")
		return
	}
	Report(w, r, "你好，茶博士摸摸头，居然说只有团建人可以修改这个茶团相关资料。")
}

// GET /v1/team//Invitations?id=
// 查询根据Uuid指定茶团team发送的全部邀请函
func ReviewInvitations(w http.ResponseWriter, r *http.Request) {
	//获取session
	s, err := Session(r)
	if err != nil {
		util.Info(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, _ := s.User()
	//根据用户提交的Uuid，查询获取团队信息
	v := r.URL.Query()
	tUid := v.Get("id")
	team, err := data.GetTeamByUuid(tUid)
	if err != nil {
		util.Info(err, " Cannot get team")
		Report(w, r, "你好，茶博士失魂鱼，未能找到这个茶团，请稍后再试。")
		return
	}

	var isPD data.InvitationsPageData
	/// 填写页面资料
	isPD.SessUser = u
	isPD.Team = team

	// 根据用户提交的Uuid，查询该茶团发送的全部邀请函
	is, err := team.Invitations()
	if err != nil {
		util.Info(err, u.Email, " Cannot get invitations")
		Report(w, r, "你好，茶博士正在努力的查找茶团发送的邀请函，请稍后再试。")
		return
	}
	//检查查询结果集是否为空
	if len(is) == 0 {
		Report(w, r, "你好，该茶团还没有发送过任何邀请函。")
		return
	}
	//填写页面资料
	isPD.InvitationList = is

	// 检查用户是否可以查看，CEO，CTO，CFO，CMO核心成员可以
	coreMembers, err := team.CoreMembers()
	// ���查err内容
	if err != nil {
		util.Info(err, " Cannot get core members")
		Report(w, r, "你好，茶博士失魂鱼，居然说这个茶团是由外星人组织的，请确认后再试。")
		return
	}
	// ���查用户是否在����中
	for _, member := range coreMembers {
		if u.Id == member.UserId {
			//向用户返回接收邀请函的表单页面
			RenderHTML(w, &isPD, "layout", "navbar.private", "team.invitations")
			return
		}
	}

	Report(w, r, "你好，蛮不讲理的茶博士竟然说，只有茶团核心成员才能查看邀请函发送记录。")
}
