package route

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/team/new
// 显示创建新茶团的表单页面
func NewTeam(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var tpd data.TeamSquare
	tpd.SessUser = u
	GenerateHTML(w, &tpd, "layout", "navbar.private", "team.new")
}

// POST /v1/team/create
// 创建新茶团
func CreateTeam(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "您好，茶博士失魂鱼，未能开新茶团，请稍后再试。")
		return
	}
	sUser, err := s.User()
	if err != nil {
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	name := r.PostFormValue("name")
	// 茶团名称是否在4-24中文字符
	le := CnStrLen(name)
	if le < 4 || le > 24 {
		Report(w, r, "您好，茶博士摸摸头，竟然说茶团名字字数太多或者太少，未能创建新茶团。")
		return
	}
	abbr := r.PostFormValue("abbreviation")
	// 队名简称是否在4-6中文字符
	lenA := CnStrLen(abbr)
	if lenA < 4 || lenA > 6 {
		Report(w, r, "您好，茶博士摸摸头，竟然说队名简称字数太多或者太少，未能创建新茶团。")
		return
	}

	mission := r.PostFormValue("mission")
	// 检测mission是否在17-456中文字符
	lenM := CnStrLen(mission)
	if lenM < 17 || lenM > 456 {
		Report(w, r, "您好，茶博士摸摸头，竟然说愿景字数太多或者太少，未能创建新茶团。")
		return
	}
	class, err := strconv.Atoi(r.PostFormValue("class"))
	if err != nil {
		util.Info(err, " Cannot convert class to int")
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	group_id, err := strconv.Atoi(r.PostFormValue("group_id"))
	if err != nil {
		util.Info(err, " Cannot convert group_id to int")
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}

	//检测class是否合规
	switch class {
	case 10, 20:
		break
	default:
		Report(w, r, "您好，茶博士摸摸头，竟然说茶团类别太多或者太少，未能创建新茶团。")
		return
	}
	//检测同名的team是否已经存在，团队不允许同名
	_, err = data.GetTeamByName(name)
	if err == nil {
		Report(w, r, "您好，茶博士摸摸头，竟然说这个茶团名称已经被占用，请换一个更响亮的团名，再试一次。")
		return
	}

	_, err = data.GetTeamByAbbreviation(abbr)
	if err == nil {
		// 重复的简称
		Report(w, r, "您好，茶博士摸摸头，竟然说这个茶团简称已经被占用，请换一个更响亮的团名，再试一次。")
		return
	}

	// 将NewTeam草稿存入数据库，class=10/20
	logo := "teamLogo"
	team, err := sUser.CreateTeam(name, abbr, mission, logo, class, group_id)
	if err != nil {
		util.Info(err, " At create team")
		Report(w, r, "您好，茶博士失魂鱼，未能创建你的天团，请稍后再试。")
		return
	}

	// 创建一条友邻盲评,是否接纳 新茶团的记录
	aO := data.AcceptObject{
		ObjectId:   team.Id,
		ObjectType: 5,
	}
	if err = aO.Create(); err != nil {
		util.Warning(err, "Cannot create accept_object")
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
	}

	// 发送盲评请求消息给两个在线用户
	//构造消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您好，茶博士隆重宣布：您被茶棚选中为新茶语评审官啦，请及时处理。",
		AcceptObjectId: aO.Id,
	}
	//发送消息
	if err = AcceptMessageSendExceptUserId(sUser.Id, mess); err != nil {
		Report(w, r, "您好，茶博士迷路了，未能发送盲评请求消息。")
		return
	}
	t := fmt.Sprintf("您好，新开茶团 %s 已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", team.Name)
	// 提示用户草稿保存成功
	Report(w, r, t)

	// 跳转到team详情页
	// url := fmt.Sprint("/v1/team/detail?id=", team.Uuid)
	// http.Redirect(w, r, url, http.StatusFound)

}

// GET /v1/team/open
// 显示茶棚全部开放式茶团详细信息
func OpenTeams(w http.ResponseWriter, r *http.Request) {
	var err error
	var ts data.TeamSquare

	team_list, err := data.GetOpenTeams()
	if err != nil {
		util.Info(err, " Cannot get open teams")
		Report(w, r, "您好，茶博士失魂鱼，未能获取茶团详细信息，请稍后再试。")
		return
	}
	ts.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "您好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	// 用户是否已经登录?
	s, err := Session(r)
	if err != nil {
		ts.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		GenerateHTML(w, &ts, "layout", "navbar.public", "teams.open", "teams.public")
		return
	}
	sUser, _ := s.User()
	ts.SessUser = sUser
	GenerateHTML(w, &ts, "layout", "navbar.private", "teams.open", "teams.public")

}

// Get /v1/team/closed
// 显示茶棚全部封闭式茶团
func ClosedTeams(w http.ResponseWriter, r *http.Request) {
	var err error
	var ts data.TeamSquare

	team_list, err := data.GetClosedTeams()
	if err != nil {
		util.Info(err, " Cannot get closed teams")
		Report(w, r, "您好，茶博士失魂鱼，未能获取茶团详细信息，请稍后再试。")
		return
	}
	ts.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "您好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	s, _ := Session(r)
	u, err := s.User()
	if err != nil {
		ts.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		GenerateHTML(w, &ts, "layout", "navbar.public", "teams.closed", "teams.public")
		return
	}
	ts.SessUser = u
	GenerateHTML(w, &ts, "layout", "navbar.private", "teams.closed", "teams.public")

}

// GET /v1/team/hold
// 显示当前用户拥有或者加入的全部茶团
func HoldTeams(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	sUser, err := s.User()
	if err != nil {
		util.Info(err, " Cannot get user from session")
		return
	}
	var ts data.TeamSquare
	team_list, err := sUser.HoldTeams()
	if err != nil {
		util.Info(err, " Cannot get hold teams")
		return
	}
	ts.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "您好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	ts.SessUser = sUser
	GenerateHTML(w, &ts, "layout", "navbar.private", "teams.hold", "teams.public")
}

// GET /v1/team/joined
// 显示用户已经加入的全部茶团的表单页面
func JoinedTeam(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, _ := sess.User()

	var ts data.TeamSquare
	ts.SessUser = u
	team_list, err := data.GetSurvivalTeamsByUserId(u.Id)
	if err != nil {
		util.Info(err, " Cannot get joined teams")
		Report(w, r, "您好，茶博士未能帮忙查看茶团，请稍后再试。")
		return
	}
	ts.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "您好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}

	GenerateHTML(w, &ts, "layout", "navbar.private", "teams.joined", "teams.public")
}

// GET /v1/team/Employed
// 显示用户任职高管的全部茶团表单页面
func EmployedTeam(w http.ResponseWriter, r *http.Request) {
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
		Report(w, r, "您好，茶博士必须先找到自己的高度近视眼镜，再帮您查询资料。请稍后再试。")
		return
	}
	ts.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		util.Info(err, " Cannot get team bean list")
		Report(w, r, "您好，酒未敌腥还用菊，性防积冷定须姜。请稍后再试。")
		return
	}
	GenerateHTML(w, &ts, "layout", "navbar.private", "teams.employed", "teams.public")
}

// GET /v1/team/quit

// GET /v1/team/detail
// 显示茶团详细信息
func TeamDetail(w http.ResponseWriter, r *http.Request) {
	//获取茶团资料
	vals := r.URL.Query()
	uuid := vals.Get("id")
	te, err := data.GetTeamByUuid(uuid)
	if err != nil {
		util.Info(err, " Cannot get team")
		Report(w, r, "您好，茶博士未能帮忙查看茶团，请稍后再试。")
		return
	}

	var tPD data.TeamDetail
	tPD.Team = te
	tPD.CreatedAtDate = te.CreatedAtDate()
	fu, err := te.Founder()
	if err != nil {
		util.Info(err, " Cannot get team founder")
		Report(w, r, "您好，闪电考拉为你极速效劳中，请稍后再试。")
		return
	}
	tPD.Founder = fu
	founder_default_team, _ := fu.GetLastDefaultTeam()
	tPD.FounderTeam = founder_default_team

	tPD.TeamMemberCount = te.NumMembers()
	teamCoreMembers, err := te.CoreMembers()
	if err != nil {
		util.Info(err, " Cannot get team core member")
		Report(w, r, "您好，闪电考拉为你效劳中，请稍后再试。")
		return
	}
	teamNormalMembers, err := te.NormalMembers()
	if err != nil {
		util.Info(err, " Cannot get team normal member")
		Report(w, r, "您好，闪电考拉为你效劳中，请稍后再试。")
		return
	}
	// 准备页面数据
	var tc data.TeamMemberBean //核心成员资料夹
	var tn data.TeamMemberBean //普通成员资料夹
	var tcList []data.TeamMemberBean
	var tnList []data.TeamMemberBean
	//据teamMembers中的UserId获取User
	for _, member := range teamCoreMembers {
		user, err := data.GetUserById(member.UserId)
		if err != nil {
			util.Info(err, " Cannot get user")
			Report(w, r, "您好，闪电考拉为你效劳中，请稍后再试。")
			return
		}

		tc.User = user
		tc.AuthorTeam, err = user.GetLastDefaultTeam()
		if err != nil {
			util.Info(err, " Cannot get user's default team")
			Report(w, r, "您好，闪电考拉为你效劳中，请稍后再试。")
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
			Report(w, r, "您好，疯狂的闪电考拉为你效劳中，请稍后再试。")
			return
		}
		tn.User = user
		tn.AuthorTeam, err = user.GetLastDefaultTeam()
		if err != nil {
			util.Info(err, " Cannot get user's default team")
			Report(w, r, "您好，闪电考拉为你效劳中，请稍后再试。")
			return
		}
		tn.TeamMemberRole = member.Role
		tn.CreatedAtDate = member.CreatedAtDate()
		tnList = append(tnList, tn)
	}
	tPD.CoreMemberDataList = tcList
	tPD.NormalMemberDataList = tnList

	if tPD.Team.Class == 1 {
		tPD.Open = true
	} else {
		tPD.Open = false
	}
	s, err := Session(r)
	if err != nil {
		//游客
		tPD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		GenerateHTML(w, &tPD, "layout", "navbar.public", "team.detail")
		return
	}
	su, err := s.User()
	if err != nil {
		Report(w, r, "你好，茶博士说久仰大名，请问大名是谁？")
		return
	}

	tPD.SessUser = su
	GenerateHTML(w, &tPD, "layout", "navbar.private", "team.detail")

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
	Report(w, r, "您好，茶博士未能帮忙查看茶团，请稍后再试。")
}

// GET /v1/team/manage
// 显示用户管理茶团的表单页面
func HandleManageTeamGet(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, _ := sess.User()
	//一次管理一个茶团，根据提交的team.Uuid来确定
	//这个茶团是否存在？
	team, err := data.GetTeamByUuid(r.FormValue("id"))
	if err != nil {
		util.Info(err, " Cannot get this team")
		Report(w, r, "您好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}
	var tPD data.TeamDetail
	//检查一下当前用户是否有权管理这个茶团？即teamMember中Role为"ceo"或者founder
	//如果是创建人，那么就可以管理这个茶团
	fund, err := team.Founder()
	if err != nil {
		Report(w, r, "您好，茶博士未能找到此茶团发起人资料，请确认后再试。")
		return
	}
	if fund.Id == u.Id {
		//是创建人,可以管理这个茶团
		tPD.SessUser = u
		tPD.Team = team
		GenerateHTML(w, &tPD, "layout", "navbar.private", "team.manage")
		return
	}
	//如果不是创建人，那么就检查一下是否是茶团的ceo
	manager, err := team.CEO()
	if err != nil {
		//检查一下err是否是因为teamMember中没有这个team的ceo
		//如果是，那么就说明当前用户无权管理这个茶团
		//这是特殊情况？CEO空缺？例如唐僧被妖怪抓走了？
		if err == sql.ErrNoRows {
			Report(w, r, "您好，您无权管理此茶团。")
			return
		} else {
			//茶团已经设定了ceo，但是出现了其他错误
			util.Info(err, " Cannot get ceo of this team")
			Report(w, r, "您好，茶博士未能找到此茶团资料，请确认后再试。")
			return
		}
	}
	if manager.UserId != u.Id {
		Report(w, r, "您好，您无权管理此茶团。")
		return
	}
	//是茶团的ceo，可以管理这个茶团
	tPD.SessUser = u
	tPD.Team = team
	GenerateHTML(w, &tPD, "layout", "navbar.private", "team.manage")
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
		Report(w, r, "您好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}
	//这个茶团是否存在？
	team, err := data.GetTeamById(team_id)
	if err != nil {
		util.Info(err, " Cannot get ceo of this team")
		Report(w, r, "您好，茶博士未能找到此茶团资料，请确认后再试。")
		return
	}
	ceo, err := team.CEO()
	if err != nil {
		//检查一下err是否是因为teamMember中没有这个team的ceo
		//如果是，那么就说明当前用户无权管理这个茶团
		//这是特殊情况？CEO空缺？例如唐僧被妖怪抓走了？
		if err == sql.ErrNoRows {
			Report(w, r, "您好，您无权管理此茶团。")
			return
		} else {
			//茶团已经设定了ceo，但是出现了其他错误
			util.Info(err, " Cannot get ceo of this team")
			Report(w, r, "您好，茶博士未能找到此茶团资料，请确认后再试。")
			return
		}
	}
	if u.Id != ceo.UserId {
		Report(w, r, "您好，您无权管理此茶团。")
		return
	}

	//一次管理一个核心角色，根据提交的teamMember.id来确定
	//这个用户是否在这个茶团中？角色是否正确？
	member_id, _ := strconv.Atoi(r.FormValue("member_id"))
	role := r.FormValue("role")
	if role != "CTO" && role != "CMO" && role != "CFO" {
		Report(w, r, "您好，请选择正确的角色。")
		return
	}
	member, err := team.GetTeamMemberByRole(role)
	if err != nil {
		util.Info(err, " Cannot get this team member")
		Report(w, r, "您好，茶博士未能找到此茶团成员资料，请确认后再试。")
		return
	}
	if member.Id != member_id {
		Report(w, r, "您好，请选择正确的角色。")
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
			Report(w, r, "您好，保存角色管理操作失败，请稍后再试。")
			return
		}
	default:
		Report(w, r, "您好，请选择正确的管理动作。")
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
		Report(w, r, "您好，茶博士摸摸头，居然说找不到这个茶团相关资料。")
		return
	}

	// 检查team的创建者
	if u.Id == team.FounderId {
		//如果是创建者，那么就可以上传图标
		//读取上传的图标
		ProcessUploadAvatar(w, r, team.Uuid)
	}
	Report(w, r, "您好，上传茶团图标出现未知问题。")

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
		Report(w, r, "您好，茶博士摸摸头，居然说找不到这个茶团相关资料。")
		return
	}

	// 检查team的创建者
	if u.Id == team.FounderId {
		//如果是创建者，那么就可以上传图标

		GenerateHTML(w, &uuid, "layout", "navbar.private", "team_avatar.upload")
		return
	}
	Report(w, r, "您好，茶博士摸摸头，居然说只有团建人可以修改这个茶团相关资料。")
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
		Report(w, r, "您好，茶博士失魂鱼，未能找到这个茶团，请稍后再试。")
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
		Report(w, r, "您好，茶博士正在努力的查找茶团发送的邀请函，请稍后再试。")
		return
	}
	//检查查询结果集是否为空
	if len(is) == 0 {
		Report(w, r, "您好，该茶团还没有发送过任何邀请函。")
		return
	}
	//填写页面资料
	isPD.InvitationList = is

	// 检查用户是否可以查看，CEO，CTO，CFO，CMO核心成员可以
	coreMembers, err := team.CoreMembers()
	// ���查err内容
	if err != nil {
		util.Info(err, " Cannot get core members")
		Report(w, r, "您好，茶博士失魂鱼，居然说这个茶团是由外星人组织的，请确认后再试。")
		return
	}
	// ���查用户是否在����中
	for _, member := range coreMembers {
		if u.Id == member.UserId {
			//向用户返回接收邀请函的表单页面
			GenerateHTML(w, &isPD, "layout", "navbar.private", "team.invitations")
			return
		}
	}

	Report(w, r, "您好，蛮不讲理的茶博士竟然说，只有茶团核心成员才能查看邀请函发送记录。")
}
