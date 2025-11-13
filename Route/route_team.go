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
	generateHTML(w, &tSPD, "layout", "navbar.private", "team.new")
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

	//检测class是否合规
	switch class {
	case data.TeamClassOpenDraft, data.TeamClassCloseDraft:
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
		Name:         new_name,
		Abbreviation: abbr + "$",
		Mission:      mission,
		Logo:         logo,
		Class:        class,
		FounderId:    s_u.Id,
		Tags:         r.PostFormValue("tags"),
	}
	// 使用事务创建团队并添加创建人为第一个成员
	if err := createTeamWithFounderMember(&new_team, s_u.Id); err != nil {
		util.Debug("cannot create team with founder member", err)
		report(w, r, "你好，茶博士失魂鱼，暂未能创建你的天命使团，请稍后再试。")
		return
	}

	if util.Config.PoliteMode {
		//启用了友邻蒙评
		if err = createAndSendAcceptMessage(new_team.Id, data.AcceptObjectTypeTeam, s_u.Id); err != nil {
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				report(w, r, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				report(w, r, "你好，茶博士迷路了，未能发送蒙评请求消息。")
			}
			return
		}
		// 提示用户新茶团草稿保存成功
		text := ""
		if s_u.Gender == data.User_Gender_Female {
			text = fmt.Sprintf("%s 女士，你好，登记 %s 草稿已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", s_u.Name, new_team.Name)
		} else {
			text = fmt.Sprintf("%s 先生，你好，登记 %s 草稿已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", s_u.Name, new_team.Name)
		}
		report(w, r, text)
	} else {
		switch new_team.Class {
		case data.TeamClassOpenDraft:
			new_team.Class = data.TeamClassOpen
		case data.TeamClassCloseDraft:
			new_team.Class = data.TeamClassClose
		}
		if err := new_team.Update(); err != nil {
			util.Debug("Cannot update team", err)
			report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
			return
		}
		//跳转到团队详情页面
		http.Redirect(w, r, fmt.Sprintf("/v1/team/detail?uuid=%s", new_team.Uuid), http.StatusFound)
	}
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
	generateHTML(w, &ts, "layout", "navbar.private", "teams.hold", "component_teams_public", "component_team")
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

	generateHTML(w, &tS, "layout", "navbar.private", "teams.joined", "component_teams_public", "component_team")
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
	generateHTML(w, &ts, "layout", "navbar.private", "teams.employed", "component_teams_public", "component_team")
}

// GET /v1/team/detail?uuid=
// 显示茶团详细信息
func TeamDetail(w http.ResponseWriter, r *http.Request) {
	//获取茶团资料
	vals := r.URL.Query()
	uuid := vals.Get("uuid")

	if uuid == "" {
		report(w, r, "你好，请提交有效的茶团编号，请稍后再试。")
		return
	}

	if uuid == data.TeamUUIDFreelancer {
		report(w, r, "转会自由人，星际旅行特立独行的散客大集合，不属于任何$事业茶团。")
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
		generateHTML(w, &tD, "layout", "navbar.public", "team.detail", "component_avatar_name_gender")
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

	// 查询团队所属集团
	group, err := data.GetGroupByTeamId(team.Id)
	if err == nil && group != nil {
		groupBean := data.GroupBean{
			Group:         *group,
			CreatedAtDate: group.CreatedAtDate(),
			Open:          group.Class == data.GroupClassOpen,
		}
		founder, err := data.GetUser(group.FounderId)
		if err == nil {
			groupBean.Founder = founder
		}
		tD.GroupBean = &groupBean
	}

	generateHTML(w, &tD, "layout", "navbar.private", "team.detail", "component_avatar_name_gender")

}

// HandleManageTeam() /v1/team/manage
func HandleManageTeam(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ManageTeamIndexGet(w, r)
	case http.MethodPost:
		ManageTeamPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

// POST /v1/team/manage
// 根据提交的参数，处理管理某支团队事务，例如：注销团队，调整状态等
func ManageTeamPost(w http.ResponseWriter, r *http.Request) {
	report(w, r, "您好，茶博士正在忙碌建设这个功能中。。。")

}

// GET /v1/team/manage?uuid=
// 显示管理茶团的表单(首页-角色调整)页面
func ManageTeamIndexGet(w http.ResponseWriter, r *http.Request) {
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
	t_uuid := vals.Get("uuid")
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

	// 查询团队所属集团
	group, err := data.GetGroupByTeamId(team.Id)
	if err == nil && group != nil {
		// 团队已加入集团
		groupBean := data.GroupBean{
			Group:         *group,
			CreatedAtDate: group.CreatedAtDate(),
			Open:          group.Class == data.GroupClassOpen,
		}
		// 获取集团创建者
		founder, err := data.GetUser(group.FounderId)
		if err == nil {
			groupBean.Founder = founder
		}
		tD.GroupBean = &groupBean
	}

	generateHTML(w, &tD, "layout", "navbar.private", "team.manage")
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

// GET /v1/team/avatar?uuid=
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
	uuid := r.URL.Query().Get("uuid")
	team, err := data.GetTeamByUUID(uuid)
	if err != nil {
		report(w, r, "你好，茶博士摸摸头，居然说找不到这个茶团相关资料。")
		return
	}

	// 检查team的创建者
	if s_u.Id == team.FounderId {
		//如果是创建者，那么就可以上传图标

		generateHTML(w, &uuid, "layout", "navbar.private", "team_avatar.upload")
		return
	}
	report(w, r, "你好，茶博士摸摸头，居然说只有团建人可以修改这个茶团相关资料。")
}

// GET /v1/team/applications?uuid=
// 查询根据Uuid指定茶团的全部加盟申请书，支持分页
func TeamApplications(w http.ResponseWriter, r *http.Request) {
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

	v := r.URL.Query()
	tUid := v.Get("uuid")
	team, err := data.GetTeamByUUID(tUid)
	if err != nil {
		util.Debug("Cannot get team", err)
		report(w, r, "你好，茶博士失魂鱼，未能找到这个茶团，请稍后再试。")
		return
	}

	// 检查用户是否可以查看(管理权限），核心成员可以
	if !canManageTeam(&team, s_u.Id, w, r) {
		return
	}

	// 获取分页参数
	page := 1
	if pageStr := v.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// 查询全部申请书
	applications, err := data.GetMemberApplicationByTeamId(team.Id)
	if err != nil {
		util.Debug("Cannot get applications", err)
		report(w, r, "你好，茶博士正在努力的查找申请书，请稍后再试。")
		return
	}

	// 分页处理
	pageSize := 12
	totalCount := len(applications)
	totalPages := (totalCount + pageSize - 1) / pageSize
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	var pageApplications []data.MemberApplication
	if totalCount > 0 {
		pageApplications = applications[start:end]
	}

	// 获取申请书Bean
	applicationBeans, err := fetchMemberApplicationBeanSlice(pageApplications)
	if err != nil {
		util.Debug("Cannot fetch application beans", err)
		report(w, r, "你好，茶博士正在努力的查找申请书，请稍后再试。")
		return
	}

	var mAL data.MemberApplicationSlice
	mAL.SessUser = s_u
	mAL.Team = team
	mAL.MemberApplicationBeanSlice = applicationBeans

	// 分页信息
	type PageData struct {
		data.MemberApplicationSlice
		CurrentPage int
		TotalPages  int
		HasPrev     bool
		HasNext     bool
	}

	pageData := PageData{
		MemberApplicationSlice: mAL,
		CurrentPage:            page,
		TotalPages:             totalPages,
		HasPrev:                page > 1,
		HasNext:                page < totalPages,
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "team.applications", "component_member_application_bean")
}

// GET /v1/team/invitations?uuid=&page=
// 查询根据Uuid指定茶团team发送的全部邀请函，支持分页
func TeamInvitations(w http.ResponseWriter, r *http.Request) {
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
	tUid := v.Get("uuid")
	team, err := data.GetTeamByUUID(tUid)
	if err != nil {
		util.Debug(" Cannot get team", err)
		report(w, r, "你好，茶博士失魂鱼，未能找到这个茶团，请稍后再试。")
		return
	}

	// 检查用户是否可以查看，FOUNDER，CEO，CTO，CFO，CMO核心成员可以
	if !canManageTeam(&team, s_u.Id, w, r) {
		return
	}

	// 获取分页参数
	page := 1
	if pageStr := v.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// 根据用户提交的Uuid，查询该茶团发送的全部邀请函
	t_invi_slice, err := team.Invitations()
	if err != nil {
		util.Debug(s_u.Email, " Cannot get invitations")
		report(w, r, "你好，茶博士正在努力的查找茶团发送的邀请函，请稍后再试。")
		return
	}

	// 分页处理
	pageSize := 12
	totalCount := len(t_invi_slice)
	totalPages := (totalCount + pageSize - 1) / pageSize
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	var pageInvitations []data.Invitation
	if totalCount > 0 {
		pageInvitations = t_invi_slice[start:end]
	}

	// 分页信息
	type PageData struct {
		data.InvitationsPageData
		CurrentPage int
		TotalPages  int
		HasPrev     bool
		HasNext     bool
	}

	pageData := PageData{
		InvitationsPageData: data.InvitationsPageData{
			SessUser:        s_u,
			Team:            team,
			InvitationSlice: pageInvitations,
		},
		CurrentPage: page,
		TotalPages:  totalPages,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "team.invitations")

}

// canManageTeam 检查用户是否有权限管理团队（核心成员或创建人）
func canManageTeam(team *data.Team, userId int, w http.ResponseWriter, r *http.Request) bool {
	isCoreMember, err := team.IsCoreMember(userId)
	if err != nil {
		util.Debug("Cannot get core members", err)
		report(w, r, "你好，茶博士失魂鱼，居然说这个茶团是由外星人组织的，请确认后再试。")
		return false
	}

	if !isCoreMember {
		founder, err := team.Founder()
		if err != nil {
			util.Debug(team.Id, "Cannot get team founder")
			report(w, r, "你好，茶博士未能找到此茶团资料，请确认后再试。")
			return false
		}
		if founder.Id == userId {
			isCoreMember = true
		}
	}

	if !isCoreMember {
		report(w, r, "你好，蛮不讲理的茶博士竟然说，只有茶团核心成员才能查看此内容。")
		return false
	}

	return true
}

// GET /v1/team/edit?uuid=
// 显示编辑茶团资料表单
func EditTeamGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，缺少茶团标识。")
		return
	}

	team, err := data.GetTeamByUUID(uuid)
	if err != nil {
		util.Debug("Cannot get team by uuid", err)
		report(w, r, "你好，未能找到该茶团。")
		return
	}

	// 检查编辑权限：必须是创建人或CEO
	canEdit := false
	if sessUser.Id == team.FounderId {
		canEdit = true
	} else {
		ceo, err := team.MemberCEO()
		if err == nil && ceo.UserId == sessUser.Id {
			canEdit = true
		}
	}

	if !canEdit {
		report(w, r, "你好，只有茶团创建人或CEO才能编辑茶团资料。")
		return
	}

	var pageData struct {
		SessUser data.User
		Team     data.Team
	}
	pageData.SessUser = sessUser
	pageData.Team = team

	generateHTML(w, &pageData, "layout", "navbar.private", "team.edit")
}

// HandleEditTeam GET/POST /v1/team/edit
// 处理茶团编辑
func HandleEditTeam(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		EditTeamGet(w, r)
	case http.MethodPost:
		EditTeamPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// POST /v1/team/edit
// 编辑茶团资料
func EditTeamPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能编辑茶团。")
		return
	}

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，未能编辑茶团。")
		return
	}

	teamIdStr := r.PostFormValue("team_id")
	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, r, "你好，茶团ID无效。")
		return
	}

	// 获取茶团并检查权限
	team, err := data.GetTeam(teamId)
	if err != nil {
		util.Debug("Cannot get team", err)
		report(w, r, "你好，未找到该茶团。")
		return
	}

	// 检查编辑权限
	canEdit := false
	if sessUser.Id == team.FounderId {
		canEdit = true
	} else {
		ceo, err := team.MemberCEO()
		if err == nil && ceo.UserId == sessUser.Id {
			canEdit = true
		}
	}

	if !canEdit {
		report(w, r, "你好，只有茶团创建人或CEO才能编辑茶团资料。")
		return
	}

	// 读取表单数据
	name := r.PostFormValue("name")
	abbreviation := r.PostFormValue("abbreviation")
	mission := r.PostFormValue("mission")

	// 验证茶团名称长度
	nameLen := cnStrLen(name)
	if nameLen < 4 || nameLen > 24 {
		report(w, r, "你好，茶团名称应在4-24个中文字符之间。")
		return
	}

	// 验证简称长度（去掉$符号后验证）
	abbrWithoutSymbol := strings.TrimSuffix(abbreviation, "$")
	abbrLen := cnStrLen(abbrWithoutSymbol)
	if abbrLen < 4 || abbrLen > 6 {
		report(w, r, "你好，队名简称应在4-6个中文字符之间。")
		return
	}

	// 验证使命长度
	missionLen := cnStrLen(mission)
	if missionLen < int(util.Config.ThreadMinWord) || missionLen > int(util.Config.ThreadMaxWord) {
		report(w, r, "你好，团队使命字数不符合要求。")
		return
	}

	// 检查名称是否与其他茶团重复（排除自己）
	if name != team.Name {
		existingTeam := data.Team{Name: name}
		if err := existingTeam.GetByName(); err == nil {
			report(w, r, "你好，这个茶团名称已经被占用，请换一个更响亮的团名。")
			return
		}
	}

	// 检查简称是否与其他茶团重复（排除自己）
	if !strings.HasSuffix(abbreviation, "$") {
		abbreviation = abbreviation + "$"
	}
	if abbreviation != team.Abbreviation {
		existingTeam := data.Team{Abbreviation: abbreviation}
		if err := existingTeam.GetByAbbreviation(); err == nil {
			report(w, r, "你好，这个茶团简称已经被占用，请换一个更响亮的团名简称。")
			return
		}
	}

	// 更新茶团信息
	team.Name = name
	team.Abbreviation = abbreviation
	team.Mission = mission

	if err := team.Update(); err != nil {
		util.Debug("Cannot update team", err)
		report(w, r, "你好，茶博士失魂鱼，未能更新茶团信息。")
		return
	}

	http.Redirect(w, r, "/v1/team/manage?uuid="+team.Uuid, http.StatusFound)
}

// GET /v1/team/member_add?uuid=
// 显示团队搜索用户页面
func TeamMemberAddGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，缺少茶团标识。")
		return
	}

	team, err := data.GetTeamByUUID(uuid)
	if err != nil {
		util.Debug("Cannot get team by uuid", err)
		report(w, r, "你好，未能找到该茶团。")
		return
	}

	// 检查权限：必须是CEO或创建人
	canManage := false
	if sessUser.Id == team.FounderId {
		canManage = true
	} else if ceo, err := team.MemberCEO(); err == nil && ceo.UserId == sessUser.Id {
		canManage = true
	}

	if !canManage {
		report(w, r, "你好，只有茶团创建人或CEO才能邀请新成员。")
		return
	}

	var pageData struct {
		SessUser data.User
		Team     data.Team
	}
	pageData.SessUser = sessUser
	pageData.Team = team

	generateHTML(w, &pageData, "layout", "navbar.private", "team.member_add")
}

// HandleTeamSearchUser POST /v1/team/search_user
// 处理团队搜索用户请求
func HandleTeamSearchUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能理解你的话语，请稍后再试。")
		return
	}

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sessUser, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	teamUuid := r.PostFormValue("team_uuid")
	searchType := r.PostFormValue("search_type")
	keyword := r.PostFormValue("keyword")

	// 验证关键词长度
	if len(keyword) < 1 || len(keyword) > 32 {
		report(w, r, "你好，茶博士摸摸头，说关键词太长了记不住呢，请确认后再试。")
		return
	}

	// 获取团队信息
	team, err := data.GetTeamByUUID(teamUuid)
	if err != nil {
		util.Debug("Cannot get team by uuid", err)
		report(w, r, "你好，未能找到该茶团。")
		return
	}

	// 检查权限
	canManage := false
	if sessUser.Id == team.FounderId {
		canManage = true
	} else if ceo, err := team.MemberCEO(); err == nil && ceo.UserId == sessUser.Id {
		canManage = true
	}

	if !canManage {
		report(w, r, "你好，只有茶团创建人或CEO才能邀请新成员。")
		return
	}

	var pageData struct {
		SessUser      data.User
		Team          data.Team
		UserBeanSlice []data.UserBean
		IsEmpty       bool
	}
	pageData.SessUser = sessUser
	pageData.Team = team
	pageData.IsEmpty = true

	switch searchType {
	case "user_id":
		// 按茶友号查询
		userId, err := strconv.Atoi(keyword)
		if err != nil || userId <= 0 {
			report(w, r, "茶友号必须是正整数")
			return
		}

		user, err := data.GetUser(userId)
		if err == nil && user.Id > 0 {
			userBean, err := fetchUserBean(user)
			if err == nil {
				pageData.UserBeanSlice = append(pageData.UserBeanSlice, userBean)
				pageData.IsEmpty = false
			}
		}

	case "user_email":
		// 按邮箱查询
		if !isEmail(keyword) {
			report(w, r, "你好，请输入有效的电子邮箱地址。")
			return
		}

		user, err := data.GetUserByEmail(keyword, r.Context())
		if err == nil && user.Id > 0 {
			userBean, err := fetchUserBean(user)
			if err == nil {
				pageData.UserBeanSlice = append(pageData.UserBeanSlice, userBean)
				pageData.IsEmpty = false
			}
		}

	case "user_name":
		// 按花名查询
		userSlice, err := data.SearchUserByNameKeyword(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err == nil && len(userSlice) >= 1 {
			userBeanSlice, err := fetchUserBeanSlice(userSlice)
			if err == nil && len(userBeanSlice) >= 1 {
				pageData.UserBeanSlice = userBeanSlice
				pageData.IsEmpty = false
			}
		}

	default:
		report(w, r, "你好，请选择正确的查询方式。")
		return
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "team.search_user_result", "component_avatar_name_gender")
}

// createTeamWithFounderMember 使用事务创建团队并将创建人登记为第一个成员（CEO）
func createTeamWithFounderMember(team *data.Team, founderUserId int) error {
	tx, err := data.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := team.CreateWithTx(tx); err != nil {
		return err
	}

	member := data.TeamMember{
		TeamId: team.Id,
		UserId: founderUserId,
		Role:   data.RoleCEO,
		Status: data.TeMemberStatusActive,
	}
	if err := member.CreateWithTx(tx); err != nil {
		return err
	}

	return tx.Commit()
}
