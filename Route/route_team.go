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

// GET /v1/team/default?uuid=
// 设置某个茶友的默认$事业茶团
func SetDefaultTeam(w http.ResponseWriter, r *http.Request) {
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

	//获取参数
	uuid := r.URL.Query().Get("uuid")

	//检查是否将特殊茶团作为默认茶团
	if uuid == data.TeamUUIDFreelancer || uuid == data.TeamUUIDSpaceshipCrew {
		report(w, r, "你好，茶博士竟然说，陛下你不能将特殊茶团作为默认茶团，请确认。")
		return
	}
	//查询目标茶团是否存在
	t_team, err := data.GetTeamByUUID(uuid)
	if err != nil {
		util.Debug("Cannot get team by given uuid", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取茶团，请稍后再试。")
		return
	}

	if t_team.Id == data.TeamIdFreelancer {
		report(w, r, "你好，茶博士竟然说，陛下你不能将特殊茶团作为默认茶团，请确认。")
		return
	}

	//检查是否重复设置默认事业茶团
	lastDefaultTeam, err := s_u.GetLastDefaultTeam()
	if err != nil {
		util.Debug("Cannot get last default team", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取茶团，请稍后再试。")
		return
	}
	if lastDefaultTeam.Id > data.TeamIdFreelancer {
		if lastDefaultTeam.Id == t_team.Id {
			report(w, r, "你好，茶博士竟然说，陛下你已经设置过这个默认$事业茶团了，请确认。")
			return
		}
	}

	//检查用户是否茶团成员，非成员不能设置默认茶团
	ok, err := t_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug("Cannot check user is member of team", t_team.Id, err)
		report(w, r, "你好，茶博士失魂鱼，未能获取茶团，请稍后再试。")
		return
	}
	if !ok {
		report(w, r, "你好，茶博士竟然说，陛下你似乎不是这个茶团成员，请确认。")
		return
	}

	//设置默认茶团
	new_default_team := data.UserDefaultTeam{
		UserId: s_u.Id,
		TeamId: t_team.Id,
	}
	if err = new_default_team.Create(); err != nil {
		util.Debug("Cannot set default team", err)
		report(w, r, "你好，茶博士失魂鱼，未能设置默认茶团，请稍后再试。")
		return
	}

	//跳转“已加盟茶团”页面
	http.Redirect(w, r, "/v1/teams/joined", http.StatusFound)
}

// GET /v1/team/members/fired
// 显示被开除的成员的表单页面
func MemberFired(w http.ResponseWriter, r *http.Request) {
	report(w, r, "您好，茶博士正在忙碌建设这个功能中。。。")
}

// GET /v1/team_member/application/check?uuid=
// 查看加盟某个茶团的全部新的加盟申请书
func MemberApplyCheck(w http.ResponseWriter, r *http.Request) {
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

	//获取参数
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，茶博士竟然说，陛下你没有给我茶团的uuid，请确认。")
		return
	}

	if uuid == data.TeamUUIDFreelancer {
		report(w, r, "你好，茶博士竟然说，陛下你不能查看特殊茶团的加盟申请，请确认。")
		return
	}

	//查询目标茶团
	t_team, err := data.GetTeamByUUID(uuid)
	if err != nil {
		util.Debug("Cannot get team by given uuid", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	//检查用户是否茶团成员，非成员不能查看加盟申请书
	ok, err := t_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug("Cannot check user is member of team", t_team.Id, err)
		report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	if !ok {
		report(w, r, "你好，茶博士竟然说，陛下你似乎不是这个茶团成员，请确认。")
		return
	}

	//查询茶团全部新的加盟申请书，包含已查看但未处理的
	applies, err := data.GetMemberApplicationByTeamIdAndStatus(t_team.Id)
	if err != nil {
		util.Debug("Cannot get applys by team id", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	apply_bean_slice, err := fetchMemberApplicationBeanSlice(applies)
	if err != nil {
		util.Debug("Cannot get apply bean slice", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}

	//截短MemberApplication.Content为66字，方便布局列表预览
	for _, bean := range apply_bean_slice {
		bean.MemberApplication.Content = subStr(bean.MemberApplication.Content, 66)
	}

	var mAL data.MemberApplicationSlice
	//填写页面数据
	mAL.SessUser = s_u
	mAL.Team = t_team
	mAL.MemberApplicationBeanSlice = apply_bean_slice

	// 渲染页面
	renderHTML(w, &mAL, "layout", "navbar.private", "team.applications")

}

// GET /v1/teams/application
// 显示申请加入的全部茶团的表单页面
func ApplyTeams(w http.ResponseWriter, r *http.Request) {
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
	//查询用户全部加盟申请书
	applies, err := data.GetMemberApplies(s_u.Id)
	if err != nil {
		util.Debug("Cannot get applys by user id", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	apply_bean_slice, err := fetchMemberApplicationBeanSlice(applies)
	if err != nil {
		util.Debug("Cannot get apply bean slice", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取申请茶团，请稍后再试。")
		return
	}
	//截短MemberApplication.Content为66字，方便布局列表预览
	for _, bean := range apply_bean_slice {
		bean.MemberApplication.Content = subStr(bean.MemberApplication.Content, 66)
	}

	var mAL data.MemberApplicationSlice
	//查询用户全部加盟申请书
	mAL.SessUser = s_u
	mAL.MemberApplicationBeanSlice = apply_bean_slice

	// 渲染页面
	renderHTML(w, &mAL, "layout", "navbar.private", "teams.application")

}

// GET /v1/team/new
// 显示创建新茶团的表单页面
func NewTeamGet(w http.ResponseWriter, r *http.Request) {
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
	var tSPD data.TeamSquare
	tSPD.SessUser = s_u
	renderHTML(w, &tSPD, "layout", "navbar.private", "team.new")
}

// POST /v1/team/create
// 创建新茶团
func CreateTeamPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能开新茶团，请稍后再试。")
		return
	}
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}

	//统计当前用户已创建的茶团数量，如果>最大许可值，则不能再创建
	count, err := s_u.CountTeamsByFounderId()
	if err != nil {
		util.Debug("connot count teams given founder_id", err)
		report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	if count > int(util.Config.MaxTeamsCount) {
		report(w, r, "你好，三月香巢已垒成，梁间燕子太无情。开团数量太多了，未能创建新茶团。")
		return
	}

	new_name := r.PostFormValue("name")
	// 茶团名称是否在4-24中文字符
	l := cnStrLen(new_name)
	if l < 4 || l > 24 {
		report(w, r, "你好，茶博士摸摸头，竟然说茶团名字字数太多或者太少，未能创建新茶团。")
		return
	}

	abbr := r.PostFormValue("abbreviation")
	// 队名简称是否在4-6中文字符
	lenA := cnStrLen(abbr)
	if lenA < 4 || lenA > 6 {
		report(w, r, "你好，茶博士摸摸头，竟然说队名简称字数太多或者太少，未能创建新茶团。")
		return
	}

	mission := r.PostFormValue("mission")
	// 检测mission是否在min-int(util.Config.ThreadMaxWord)中文字符
	lenM := cnStrLen(mission)
	if lenM < int(util.Config.ThreadMinWord) || lenM > int(util.Config.ThreadMaxWord) {
		report(w, r, "你好，茶博士摸摸头，竟然说愿景字数太多或者太少，未能创建新茶团。")
		return
	}
	class, err := strconv.Atoi(r.PostFormValue("class"))
	if err != nil {
		util.Debug(" Cannot convert class to int", err)
		report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	// group_id, err := strconv.Atoi(r.PostFormValue("group_id"))
	// if err != nil {
	// 	util.PanicTea(util.LogError(err), " Cannot convert group_id to int")
	// 	Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
	// 	return
	// }

	//检测class是否合规
	switch class {
	case data.TeamClassOpenStraw, data.TeamClassCloseStraw:
		break
	default:
		report(w, r, "你好，茶博士摸摸头，竟然说茶团类别太多或者太少，未能创建新茶团。")
		return
	}
	//检测同名的team是否已经存在，团队不允许同名
	old_team := data.Team{Name: new_name}

	if err = old_team.GetByName(); err == nil {
		report(w, r, "你好，茶博士摸摸头，竟然说这个茶团名称已经被占用，请换一个更响亮的团名，再试一次。")
		return
	}

	//检测新团队名字是否包含“自由人”这样的保留关键字，
	if strings.Contains(new_name, "自由人") || strings.Contains(new_name, "Freelancer") {
		//不允许，以免误导其他茶友
		report(w, r, "你好，茶团名称不能包含“自由人”这样的保留关键字，请换一个更响亮的团名，再试一次。")
		return
	}
	if err = old_team.GetByAbbreviation(); err == nil {
		// 重复的简称,不允许，以免误导其他茶友
		report(w, r, "你好，茶博士摸摸头，竟然说这个茶团简称已经被占用，请换一个更响亮的团名，再试一次。")
		return
	}
	//检测abbr中是否包含“自由人”或者“Freelancer”.
	if strings.Contains(abbr, "自由人") || strings.Contains(abbr, "Freelancer") {
		//不允许使用的保留关键字
		report(w, r, "你好，茶团简称不能包含“自由人”或者“Freelancer”这样的保留关键字，请换一个更响亮的团名简称，再试一次。")
		return
	}

	logo := "teamLogo"
	new_team := data.Team{
		Name:              new_name,
		Abbreviation:      abbr + "$",
		Mission:           mission,
		Logo:              logo,
		Class:             class,
		FounderId:         s_u.Id,
		SuperiorTeamId:    data.TeamIdSpaceshipCrew,
		SubordinateTeamId: data.TeamIdNone,
	}
	if err := new_team.Create(); err != nil {
		util.Debug(" At create team", err)
		report(w, r, "你好，茶博士失魂鱼，暂未能创建你的天命使团，请稍后再试。")
		return
	}

	if err = createAndSendAcceptMessage(new_team.Id, data.AcceptObjectTypeTe, s_u.Id); err != nil {
		if strings.Contains(err.Error(), "创建AcceptObject失败") {
			report(w, r, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
		} else {
			report(w, r, "你好，茶博士迷路了，未能发送蒙评请求消息。")
		}
		return
	}

	// 提示用户新茶团保存成功
	text := ""
	if s_u.Gender == 0 {
		text = fmt.Sprintf("%s 女士，你好，登记 %s 已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", s_u.Name, new_team.Name)
	} else {
		text = fmt.Sprintf("%s 先生，你好，登记 %s 已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", s_u.Name, new_team.Name)
	}
	report(w, r, text)

}

// GET /v1/teams/open
// 显示茶棚全部开放式茶团列表信息
func OpenTeams(w http.ResponseWriter, r *http.Request) {
	var tS data.TeamSquare

	team_slice, err := data.GetOpenTeams()
	if err != nil {
		util.Debug(" Cannot get open teams", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取茶团详细信息，请稍后再试。")
		return
	}
	tS.TeamBeanSlice, err = fetchTeamBeanSlice(team_slice)
	if err != nil {
		util.Debug(" Cannot get team bean slice", err)
		report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	// 用户是否已经登录?
	s, err := session(r)
	if err != nil {
		tS.SessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
		}
		renderHTML(w, &tS, "layout", "navbar.public", "teams.open", "component_teams_public", "component_team")
		return
	}
	sUser, _ := s.User()
	tS.SessUser = sUser
	renderHTML(w, &tS, "layout", "navbar.private", "teams.open", "component_teams_public", "component_team")

}

// Get /v1/teams/closed
// 显示茶棚全部封闭式茶团
func ClosedTeams(w http.ResponseWriter, r *http.Request) {
	var err error
	var tS data.TeamSquare

	team_slice, err := data.GetClosedTeams()
	if err != nil {
		util.Debug(" Cannot get closed teams", err)
		report(w, r, "你好，茶博士失魂鱼，未能获取茶团详细信息，请稍后再试。")
		return
	}
	tS.TeamBeanSlice, err = fetchTeamBeanSlice(team_slice)
	if err != nil {
		util.Debug(" Cannot get team bean slice", err)
		report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	s, _ := session(r)
	s_u, err := s.User()
	if err != nil {
		tS.SessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
		}
		renderHTML(w, &tS, "layout", "navbar.public", "teams.closed", "component_teams_public", "component_team")
		return
	}
	tS.SessUser = s_u

	renderHTML(w, &tS, "layout", "navbar.private", "teams.closed", "component_teams_public", "component_team")

}

// GET /v1/teams/hold
// 显示当前用户创建的全部茶团
func HoldTeams(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var ts data.TeamSquare
	team_slice, err := s_u.HoldTeams()
	if err != nil {
		util.Debug(" Cannot get hold teams", err)
		return
	}
	ts.TeamBeanSlice, err = fetchTeamBeanSlice(team_slice)
	if err != nil {
		util.Debug(" Cannot get team bean slice", err)
		report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	ts.SessUser = s_u
	renderHTML(w, &ts, "layout", "navbar.private", "teams.hold", "component_teams_public", "component_team")
}

// GET /v1/teams/joined
// 显示用户已经加入的全部茶团的表单页面
func JoinedTeams(w http.ResponseWriter, r *http.Request) {
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

	var tS data.TeamSquare

	tS.SessUser = s_u

	survival_team_slice, err := s_u.SurvivalTeams()
	if err != nil {
		util.Debug(" Cannot get joined teams", err)
		report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
		return
	}
	if len(survival_team_slice) == 0 {
		tS.IsEmpty = true
	} else {
		tS.IsEmpty = false
		teamBeanSlice, err := fetchTeamBeanSlice(survival_team_slice)
		if err != nil {
			util.Debug(" Cannot get team bean slice", err)
			report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
			return
		}
		if len(teamBeanSlice) > 1 {
			//置顶默认团队
			last_default_team, err := s_u.GetLastDefaultTeam()
			if err != nil {
				util.Debug(" Cannot get last default team", err)
				report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
				return
			}
			teamBeanSlice, err = moveDefaultTeamToFront(teamBeanSlice, last_default_team.Id)
			if err != nil {
				util.Debug(" Cannot move default team to front", err)
				report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
				return
			}
		}
		tS.TeamBeanSlice = teamBeanSlice
	}

	renderHTML(w, &tS, "layout", "navbar.private", "teams.joined", "component_teams_public", "component_team")
}

// GET /v1/teams/employed
// 显示用户担任核心成员（管理员）的全部茶团表单页面
func EmployedTeams(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, _ := sess.User()

	var ts data.TeamSquare
	ts.SessUser = u
	team_slice, err := u.CoreExecTeams()
	if err != nil {
		util.Debug(" Cannot get employed teams", err)
		report(w, r, "你好，茶博士必须先找到自己的高度近视眼镜，再帮您查询资料。请稍后再试。")
		return
	}
	ts.TeamBeanSlice, err = fetchTeamBeanSlice(team_slice)
	if err != nil {
		util.Debug(" Cannot get team bean slice", err)
		report(w, r, "你好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	renderHTML(w, &ts, "layout", "navbar.private", "teams.employed", "component_teams_public", "component_team")
}

// GET /v1/team/detail?id=
// 显示茶团详细信息
func TeamDetail(w http.ResponseWriter, r *http.Request) {
	//获取茶团资料
	vals := r.URL.Query()
	uuid := vals.Get("id")

	if uuid == "" {
		report(w, r, "你好，请提交有效的茶团编号，请稍后再试。")
		return
	}

	if uuid == data.TeamUUIDFreelancer {
		report(w, r, "转会自由人，星际旅行特立独行的自由职业者大集合，不属于任何$事业茶团。")
		return
	}

	team, err := data.GetTeamByUUID(uuid)
	if err != nil {
		util.Debug(uuid, " Cannot get team given uuid.")
		report(w, r, "你好，满头大汗的茶博士未能帮忙查看这个茶团资料，请稍后再试。")
		return
	}

	var tD data.TeamDetail

	teamBean, err := fetchTeamBean(team)
	if err != nil {
		util.Debug(" Cannot get team bean", err)
		report(w, r, "你好，茶博士未能帮忙查看这个茶团资料，请稍后再试。")
		return
	}
	tD.TeamBean = teamBean

	// 准备页面数据
	var tc data.TeamMemberBean //核心成员资料荚
	var tn data.TeamMemberBean //普通成员资料荚
	var tcSlice []data.TeamMemberBean
	var tnSlice []data.TeamMemberBean

	teamCoreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug(" Cannot get team core member", err)
		report(w, r, "你好，茶博士为你效劳中，请稍后再试。")
		return
	}
	teamNormalMembers, err := team.NormalMembers()
	if err != nil {
		util.Debug(" Cannot get team normal member given team uuid", err)
		report(w, r, "你好，茶博士为你效劳中，请稍后再试。")
		return
	}

	//据teamMembers中的UserId获取User
	for _, member := range teamCoreMembers {
		cm_user, err := data.GetUser(member.UserId)
		if err != nil {
			util.Debug(" Cannot get user", err)
			report(w, r, "你好，茶博士为你效劳中，请稍后再试。")
			return
		}

		tc.Member = cm_user

		tc.MemberDefaultFamily, err = getLastDefaultFamilyByUserId(cm_user.Id)
		if err != nil {
			util.Debug(" Cannot get user's default family", err)
			report(w, r, "你好，满头大汗的茶博士，开口唱蝶恋花，请稍后再试。")
			return
		}

		tc.MemberDefaultTeam, err = cm_user.GetLastDefaultTeam()
		if err != nil {
			util.Debug(" Cannot get user's default team", err)
			report(w, r, "你好，闪电茶博士为你效劳中，请稍后再试。")
			return
		}

		tc.TeamMember = member
		tc.CreatedAtDate = member.CreatedAtDate()
		tcSlice = append(tcSlice, tc)
	}
	for _, member := range teamNormalMembers {
		tn_user, err := data.GetUser(member.UserId)
		if err != nil {
			util.Debug(" Cannot get user", err)
			report(w, r, "你好，茶博士为你疯狂效劳中，请稍后再试。")
			return
		}
		tn.Member = tn_user
		tn.MemberDefaultFamily, err = getLastDefaultFamilyByUserId(tn_user.Id)
		if err != nil {
			util.Debug(" Cannot get user's default family", err)
			report(w, r, "你好，满头大汗的茶博士说小生这边有礼了，请稍后再试。")
			return
		}
		tn.MemberDefaultTeam, err = tn_user.GetLastDefaultTeam()
		if err != nil {
			util.Debug(" Cannot get user's default team", err)
			report(w, r, "你好，茶博士为你效劳，请稍后再试。")
			return
		}
		tn.TeamMember = member
		tn.CreatedAtDate = member.CreatedAtDate()
		tnSlice = append(tnSlice, tn)
	}
	tD.CoreMemberBeanSlice = tcSlice
	tD.NormalMemberBeanSlice = tnSlice

	if tD.TeamBean.Team.Class == 1 {
		tD.TeamBean.Open = true
	} else {
		tD.TeamBean.Open = false
	}

	tD.IsCoreMember = false
	tD.IsMember = false
	s, err := session(r)
	if err != nil {
		//游客
		tD.SessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
			// 用户足迹
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		renderHTML(w, &tD, "layout", "navbar.public", "team.detail", "component_avatar_name_gender")
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}
	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery
	tD.SessUser = s_u

	//检查当前用户是否核心成员
	for _, member_bean := range tD.CoreMemberBeanSlice {
		if member_bean.Member.Id == s_u.Id {
			tD.IsCoreMember = true
			tD.IsMember = true
			break
		}
	}
	if !tD.IsCoreMember {
		//茶团发起人也是核心成员
		if teamBean.Founder.Id == s_u.Id {
			tD.IsCoreMember = true
			tD.IsMember = true
		}
	}

	if !tD.IsMember {
		//检查当前用户是否普通成员
		for _, member_bean := range tD.NormalMemberBeanSlice {
			if member_bean.Member.Id == s_u.Id {
				tD.IsMember = true
				break
			}
		}
	}

	tD.HasApplication = false
	tD.IsCEO = false
	tD.IsFounder = false
	if tD.IsCoreMember {
		//检查是否CEO
		if s_u.Id == teamBean.CEO.Id {
			tD.IsCEO = true
		}
		if s_u.Id == teamBean.Founder.Id {
			tD.IsFounder = true
		}
		//检查茶团是否有待处理的加盟申请书
		count, err := data.GetMemberApplicationByTeamIdAndStatusCount(tD.TeamBean.Team.Id)
		if err != nil {
			util.Debug("Cannot get member application count", err)
			report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
			return
		}
		if count > 0 {
			tD.HasApplication = true
		}
	}

	renderHTML(w, &tD, "layout", "navbar.private", "team.detail", "component_avatar_name_gender")

}

// HandleManageTeam() /v1/team/manage
func HandleManageTeam(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ManageTeamGet(w, r)
	case http.MethodPost:
		ManageTeamPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

// POST /v1/team/manage
// 根据提交的参数，处理管理某支团队事务，例如：冰封团队，调整状态等
func ManageTeamPost(w http.ResponseWriter, r *http.Request) {
	report(w, r, "您好，茶博士正在忙碌建设这个功能中。。。")

}

// GET /v1/team/manage?id=
// 显示管理茶团的表单(首页-角色调整)页面
func ManageTeamGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	vals := r.URL.Query()
	t_uuid := vals.Get("id")
	//一次管理一个茶团，根据提交的team.Uuid来确定
	team, err := data.GetTeamByUUID(t_uuid)
	if err != nil {
		util.Debug(" Cannot get this team", err)
		report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}

	var tD data.TeamDetail
	//检查一下当前用户是否有权管理这个茶团？即teamMember中Role为"ceo".或者是茶团创建人
	is_manager := false
	//如果是创建人，那么就可以管理这个茶团
	if s_u.Id == team.FounderId {
		is_manager = true
		tD.IsCEO = false
		tD.IsFounder = true
	} else {
		//检查当前用户是否是ceo
		member_ceo, err := team.MemberCEO()
		if err != nil {
			util.Debug(" Cannot get ceo of this team", err)
			report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
			return
		}
		if member_ceo.UserId == s_u.Id {
			//是茶团的ceo，可以管理这个茶团
			is_manager = true
			tD.IsCEO = true
			tD.IsFounder = false
		}
	}

	if is_manager {
		//是茶团的ceo或者创建人，可以管理这个茶团
		tD.IsCoreMember = true
		tD.IsMember = true
	} else {
		report(w, r, "你好，茶博士认为，您今天无权管理此茶团。")
		return
	}

	// 填写页面数据
	tD.TeamBean, err = fetchTeamBean(team)
	if err != nil {
		util.Debug(" Cannot get team bean", err)
		report(w, r, "你好，茶博士未能帮忙查看茶团，请稍后再试。")
		return
	}

	teamCoreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug(" Cannot get team core member", err)
		report(w, r, "你好，茶博士为你效劳中，请稍后再试。")
		return
	}
	teamNormalMembers, err := team.NormalMembers()
	if err != nil {
		util.Debug(team.Id, " Cannot get team normal member")
		report(w, r, "你好，茶博士为你效劳中，请稍后再试。")
		return
	}

	var tCMBSlice []data.TeamMemberBean
	var tNMBSlice []data.TeamMemberBean

	tCMBSlice, err = fetchTeamMemberBeanSlice(teamCoreMembers)
	if err != nil {
		util.Debug(" Cannot get FetchTeamMemberBeanSlice()", err)
		report(w, r, "你好，茶博士为你疯狂效劳中，请稍后再试。")
		return
	}
	tNMBSlice, err = fetchTeamMemberBeanSlice(teamNormalMembers)
	if err != nil {
		util.Debug(" Cannot get FetchTeamMemberBeanSlice()", err)
		report(w, r, "你好，茶博士为你疯狂效劳中，请稍后再试。")
		return
	}

	if len(tCMBSlice) == 0 {
		report(w, r, "你好，茶博士未能找到此茶团核心成员资料，请确认后再试。")
		return
	}

	tD.CoreMemberBeanSlice = tCMBSlice
	tD.NormalMemberBeanSlice = tNMBSlice

	tD.SessUser = s_u

	renderHTML(w, &tD, "layout", "navbar.private", "team.manage")
}

// CoreManage() 处理用户管理团队核心成员角色事务（例如：表决事项）
func CoreManage(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	//检查当前用户是否具备管理此core角色的权限
	//检查一下是否是茶团的ceo，如果是，那么就可以管理核心角色
	team_id, err := strconv.Atoi(r.FormValue("team_id"))
	if err != nil {
		util.Debug(" Cannot strconv this team_id", err)
		report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}
	//这个茶团是否存在？
	team, err := data.GetTeam(team_id)
	if err != nil {
		util.Debug(" Cannot get ceo of this team", err)
		report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}
	ceo, err := team.MemberCEO()
	if err != nil {
		//检查一下err是否是因为teamMember中没有这个team的ceo
		//如果是，那么就说明当前用户无权管理这个茶团
		//这是特殊情况？CEO空缺？例如唐僧被妖怪抓走了？
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，您无权管理此茶团。")
			return
		} else {
			//茶团已经设定了ceo，但是出现了其他错误
			util.Debug(" Cannot get ceo of this team", err)
			report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
			return
		}
	}
	if s_u.Id != ceo.UserId {
		report(w, r, "你好，您无权管理此茶团。")
		return
	}

	//一次管理一个核心角色，根据提交的teamMember.id来确定
	//这个用户是否在这个茶团中？角色是否正确？
	member_id, err := strconv.Atoi(r.FormValue("member_id"))
	if err != nil {
		util.Debug(" Cannot convert member_id to int", err)
		report(w, r, "茶博士失魂鱼，未能读取新泡茶议资料，请稍后再试。")
		return
	}
	role := r.FormValue("role")
	if role != "CTO" && role != RoleCMO && role != RoleCFO {
		report(w, r, "你好，请选择正确的角色。")
		return
	}
	member, err := team.GetTeamMemberByRole(role)
	if err != nil {
		util.Debug(" Cannot get this team member", err)
		report(w, r, "你好，茶博士未能找到此茶团成员资料，请确认后再试。")
		return
	}
	if member.Id != member_id {
		report(w, r, "你好，请选择正确的角色。")
		return
	}
	//如果是，那么就可以管理核心角色
	//读取提交的管理动作
	action := r.FormValue("action")
	switch action {
	case "appoint", "discharge":
		member.Role = role
		err := member.UpdateRoleClass()
		if err != nil {
			report(w, r, "你好，保存角色管理操作失败，请稍后再试。")
			return
		}
	default:
		report(w, r, "你好，请选择正确的管理动作。")
		return
	}

}

// TeamAvatar() 处理茶团图标
func TeamAvatar(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		TeamAvatarGet(w, r)
	case http.MethodPost:
		TeamAvatarPost(w, r)
	}
}

// POST /v1/team/avatar
func TeamAvatarPost(w http.ResponseWriter, r *http.Request) {
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
	uuid := r.FormValue("team_uuid")
	team, err := data.GetTeamByUUID(uuid)
	if err != nil {
		report(w, r, "你好，茶博士摸摸头，居然说找不到这个茶团相关资料。")
		return
	}

	// 检查team的创建者
	if s_u.Id == team.FounderId {
		//如果是创建者，那么就可以上传图标
		//读取上传的图标
		processUploadAvatar(w, r, team.Uuid)
	}
	report(w, r, "你好，上传茶团图标出现未知问题。")

}

// GET /v1/team/avatar?id=
func TeamAvatarGet(w http.ResponseWriter, r *http.Request) {
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
	uuid := r.URL.Query().Get("id")
	team, err := data.GetTeamByUUID(uuid)
	if err != nil {
		report(w, r, "你好，茶博士摸摸头，居然说找不到这个茶团相关资料。")
		return
	}

	// 检查team的创建者
	if s_u.Id == team.FounderId {
		//如果是创建者，那么就可以上传图标

		renderHTML(w, &uuid, "layout", "navbar.private", "team_avatar.upload")
		return
	}
	report(w, r, "你好，茶博士摸摸头，居然说只有团建人可以修改这个茶团相关资料。")
}

// GET /v1/team//invitations?id=
// 查询根据Uuid指定茶团team发送的全部邀请函
func InvitationsBrowse(w http.ResponseWriter, r *http.Request) {
	//获取session
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//根据用户提交的Uuid，查询获取团队信息
	v := r.URL.Query()
	tUid := v.Get("id")
	team, err := data.GetTeamByUUID(tUid)
	if err != nil {
		util.Debug(" Cannot get team", err)
		report(w, r, "你好，茶博士失魂鱼，未能找到这个茶团，请稍后再试。")
		return
	}

	var isPD data.InvitationsPageData
	/// 填写页面资料
	isPD.SessUser = s_u
	isPD.Team = team

	// 根据用户提交的Uuid，查询该茶团发送的全部邀请函
	is, err := team.Invitations()
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations")
		report(w, r, "你好，茶博士正在努力的查找茶团发送的邀请函，请稍后再试。")
		return
	}

	//填写页面资料
	isPD.InvitationSlice = is

	// 检查用户是否可以查看，FOUNDER，CEO，CTO，CFO，CMO核心成员可以
	IsCoreMember := false

	coreMembers, err := team.CoreMembers()
	// 查err内容
	if err != nil {
		util.Debug(" Cannot get core members", err)
		report(w, r, "你好，茶博士失魂鱼，居然说这个茶团是由外星人组织的，请确认后再试。")
		return
	}
	// 检测用户是否茶团管理员
	for _, member := range coreMembers {
		if s_u.Id == member.UserId {
			IsCoreMember = true
			break
		}
	}
	if !IsCoreMember {
		//查询创建人资料
		founder, err := team.Founder()
		if err != nil {
			util.Debug(team.Id, " Cannot get team founder")
			report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
			return
		}
		if founder.Id == s_u.Id {
			IsCoreMember = true
		}
	}

	if IsCoreMember {
		//say yes,向用户返回接收邀请函的表单页面
		renderHTML(w, &isPD, "layout", "navbar.private", "team.invitations")
		return
	} else {
		// say no
		report(w, r, "你好，蛮不讲理的茶博士竟然说，只有茶团核心成员才能查看邀请函发送记录。")
		return
	}

}

// GET /v1/team//invitation?id=
// 管理员根据Uuid查看某个茶团已发出的邀请函
func InvitationView(w http.ResponseWriter, r *http.Request) {
	//获取session
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//根据用户提交的Uuid，查询获取邀请函信息
	v := r.URL.Query()
	i_uuid := v.Get("id")
	in, err := data.GetInvitationByUuid(i_uuid)
	if err != nil {
		util.Debug(" Cannot get invitation", err)
		report(w, r, "你好，茶博士失魂鱼，未能找到这个邀请函，请稍后再试。")
		return
	}
	//目标茶团
	team, err := in.Team()
	if err != nil {
		util.Debug(" Cannot get team given invitation", err)
		report(w, r, "你好，茶博士正在努力的查找邀请函，请稍后再试。")
		return
	}

	var iD data.InvitationDetail
	/// 填写页面资料
	iD.SessUser = s_u
	i_b, err := fetchInvitationBean(in)
	if err != nil {
		util.Debug(" Cannot get invitation bean", err)
		report(w, r, "你好，茶博士正在努力的查找邀请函资料，请稍后再试。")
		return
	}

	iD.InvitationBean = i_b

	// 检查用户是否可以查看，CEO，CTO，CFO，CMO核心成员可以
	//检查当前用户是否茶团核心（管理员）
	IsCoreMember := false

	coreMembers, err := team.CoreMembers()
	// 查err内容
	if err != nil {
		util.Debug(" Cannot get core members", err)
		report(w, r, "你好，茶博士失魂鱼，居然说这个茶团是由外星人组织的，请确认后再试。")
		return
	}
	// 检测用户是否茶团管理员
	for _, member := range coreMembers {
		if s_u.Id == member.UserId {
			IsCoreMember = true
			break
		}
	}
	if !IsCoreMember {
		//查询创建人资料
		founder, err := team.Founder()
		if err != nil {
			util.Debug(team.Id, " Cannot get team founder")
			report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
			return
		}
		if founder.Id == s_u.Id {
			IsCoreMember = true
		}
	}

	if IsCoreMember {
		//say yes,向用户返回接收邀请函的表单页面
		renderHTML(w, &iD, "layout", "navbar.private", "team.invitation_view")
		return
	} else {
		// say no
		report(w, r, "你好，蛮不讲理的茶博士竟然说，只有茶团创始人和核心成员才能查看邀请函发送记录。")
	}

}
