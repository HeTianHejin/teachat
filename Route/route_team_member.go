package route

import (
	"database/sql"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
)

// /v1/team/team_member/invite
// 邀请一个新成员加入封闭式茶团
func HandleInviteMember(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//返回邀请团队新成员，即邀请函填写页面
		InviteMemberPage(w, r)
	case "POST":
		//生成邀请函方法
		InviteMember(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// /v1/team/team_member/invitation
// 处理封闭式茶团邀请新成员函
func HandleInvitation(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//返回邀请函详情页面
		InvitationDetail(w, r)
	case "POST":
		//设置邀请函回复方法
		HandleInvitationReply(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// /v1/team/team_member/join
// 申请加入一个开放式茶团
func HandleJoinTeam(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//返回 申请加入 页面
		JoinTeamPage(w, r)
	case "POST":
		//申请加入 处理方法
		JoinTeam(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// POST /v1/team/team_member/join
func JoinTeam(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "你好，茶博士正在忙碌抛砖引玉中，稍后再试。")

}

// GET /v1/team/team_member/join
func JoinTeamPage(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "你好，茶博士正在忙碌抛砖引玉中，稍后再试。")

}

// POST /v1/team/team_member/invitation
// 邀请函回复方法
func HandleInvitationReply(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		util.Danger(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//解析表单内容，获取用户提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Danger(err, " Cannot parse form")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	s_u, _ := sess.User()
	invitation_id, err := strconv.Atoi(r.PostFormValue("invitation_id"))
	if err != nil {
		util.Warning(err, "Failed to convert class to int")
		return
	}
	user_id, err := strconv.Atoi(r.PostFormValue("user_id"))
	if err != nil {
		util.Warning(err, "Failed to convert class to int")
		return
	}
	//检查一下提交的用户和会话用户Id是否一致
	if user_id != s_u.Id {
		util.Warning(err, s_u.Email, "Inconsistency between submitted user id and session id")
		Report(w, r, "你好，请先登录，稍后再试。")
		return
	}
	//根据用户提交的invitation_id，检查是否存在该邀请函
	invitation, err := data.GetInvitationById(invitation_id)
	if err != nil {
		util.Danger(err, s_u.Email, " Cannot get invitation")
		Report(w, r, "你好，秋阴捧出何方雪？雨渍添来隔宿痕。稍后再试。")
		return
	}
	//检查一下邀请函是否已经被回复
	if invitation.Status > 1 {
		Report(w, r, "你好，这个邀请函已经答复或者已过期。")
		return
	}
	reply, err := strconv.Atoi(r.PostFormValue("reply"))
	if err != nil {
		util.Warning(err, "Failed to convert class to int")
		return
	}
	reply_word := r.PostFormValue("invitation_reply")
	//检查一下用户提交的string，即reply_word是否不为空，中文长度小于239字符之间

	if 1 > CnStrLen(reply_word) || CnStrLen(reply_word) > 239 {
		util.Warning(err, s_u.Email, " Cannot process invitation")
		Report(w, r, "你好，瞪大眼睛涨红了脸的茶博士，竟然强词夺理说，答复的话太长了或者太短，只有外星人才接受呀，请确认再试。")
		return
	}
	if reply == 1 {
		//接受邀请，则升级邀请函状态并保存答复话语和时间
		invitation.Status = 2
		invitation.Update()
		repl := data.InvitationReply{
			InvitationId: invitation_id,
			UserId:       user_id,
			ReplyWord:    reply_word,
			CreatedAt:    time.Now(),
		}
		err = repl.Create()
		if err != nil {
			util.Warning(err, s_u.Email, " Cannot create invitation_reply")
			Report(w, r, "你好，晕头晕脑的茶博士竟然把邀请答复搞丢了，请稍后再试。")
			return
		}
		// 准备将新成员添加进茶团
		team_member := data.TeamMember{
			TeamId:    invitation.TeamId,
			UserId:    user_id,
			Role:      invitation.Role,
			CreatedAt: time.Now(),
			Class:     1,
		}
		//检查一下是否已经这个茶团是否已经存在该用户了

		_, err := data.GetTeamMemberByTeamIdAndUserId(team_member.TeamId, team_member.UserId)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				break
			default:
				Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
				return
			}
		} else {
			Report(w, r, "你好，茶博士摸摸头嘀咕说，这个茶友已经在茶团中了噢。")
			return
		}

		// 如果team_member.Role == "CEO",采取更换CEO方法
		if team_member.Role == "CEO" {
			if err = team_member.UpdateFirstCEO(s_u.Id); err != nil {
				util.Warning(err, s_u.Email, " Cannot update team_member")
				Report(w, r, "你好，幽情欲向嫦娥诉，无奈虚廊夜色昏。请稍后再试。")
				return
			}

		} else {
			// 其它角色
			if err = team_member.Create(); err != nil {
				util.Warning(err, s_u.Email, " Cannot create team_member")
				Report(w, r, "你好，晕头晕脑的茶博士竟然忘记登记新成员了，请稍后再试。")
				return
			}
		}

		//返回此茶团页面给用户，成员列表上有该用户，示意已经加入该茶团
		team, err := data.GetTeamById(invitation.TeamId)
		if err != nil {
			util.Danger(err, s_u.Email, " Cannot get team")
			Report(w, r, "你好，丢了眼镜的茶博士忙到现在，还没有找到茶团登记本，请稍后再试。")
			return
		}
		http.Redirect(w, r, "/v1/team/detail?id="+(team.Uuid), http.StatusFound)
		return
	} else if reply == 0 {
		//拒绝邀请，则改写邀请函状态并保存答复话语和时间
		invitation.Status = 3
		invitation.Update()
		repl := data.InvitationReply{
			InvitationId: invitation_id,
			UserId:       user_id,
			ReplyWord:    reply_word,
			CreatedAt:    time.Now(),
		}
		err = repl.Create()
		if err != nil {
			util.Warning(err, s_u.Email, " Cannot create invitation_reply")
			Report(w, r, "你好，晕头晕脑的茶博士竟然把邀请答复搞丢了，请稍后再试。")
			return
		}
		//返回此茶团页面给用户，成员名单上没有该用户，示意已经拒绝该邀请
		team, err := data.GetTeamById(invitation.TeamId)
		if err != nil {
			util.Danger(err, s_u.Email, " Cannot get team")
			Report(w, r, "你好，粗心大意的茶博士还没有找到茶团登记本，请稍后再试。")
			return
		}
		http.Redirect(w, r, "/v1/team/detail?id="+(team.Uuid), http.StatusFound)
		return
	} else {
		// 无效的reply 数值
		Report(w, r, "你好，何幸邀恩宠，宫车过往频。稍后再试。")
		return
	}
}

// POST /v1/team/team_member/invite_page
// 创建一封邀请函。邀请新成员到teamId指定的团队
func InviteMember(w http.ResponseWriter, r *http.Request) {
	//获取session
	sess, err := Session(r)
	if err != nil {
		util.Danger(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	su, _ := sess.User()
	//解析表单内容，获取用户提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Danger(err, " Cannot parse form")
		http.Redirect(w, r, "/v1/", http.StatusFound)
		return
	}
	email := r.PostFormValue("email")
	i_word := r.PostFormValue("invite_word")
	role := r.PostFormValue("role")
	team_uuid := r.PostFormValue("team_uuid")

	//检查用户是否自己邀请自己？也许是可以的，例如观音菩萨也可以加入自己创建的西天取经茶团喝茶？？
	/* if u.Email == i_email {
		util.Pop_message(w, r, "你好，请不要邀请自己加入茶团哈。")
		return
	} */
	//根据用户提交的teamId，检查是否存在该team
	team, err := data.TeamByUuid(team_uuid)
	if err != nil {
		util.Danger(err, su.Email, " Cannot search team given team_uuid")
		Report(w, r, "你好，茶博士未能找到这个团队，请确认后再试。")
		return
	}
	//检查当前用户是否团队的Ceo或者founder，是否有权限邀请新成员
	ceo, err := team.CEO()
	if err != nil {
		util.Danger(err, su.Email, " Cannot search team ceo")
		Report(w, r, "你好，������未能找到��队CEO，请确认后再试。")
		return
	}
	founder, err := team.Founder()
	if err != nil {
		util.Danger(err, su.Email, " Cannot search team founder")
		Report(w, r, "你好，������未能找到这个��队的��始人，请确认后再试。")
		return
	}

	ok := su.Id == ceo.UserId || su.Id == founder.Id
	if !ok {
		Report(w, r, "你好，机关算尽太聪明，反算了卿卿性命。只有团队CEO或者创始人能够邀请请新成员加盟。")
		return
	}

	//根据用户提交的Uuid，检查是否存在该User
	new_member, err := data.UserByEmail(email)
	if err != nil {
		util.Danger(err, email, " Cannot search user given email")
		Report(w, r, "你好，满头大汗的茶博士未能茶棚里找到这个用户，请确认后再试。")
		return
	}
	//检查受邀请的用户团队核心角色是否已经被占用
	switch role {
	case "CTO", "CMO", "CFO":
		//检查teamMember.Role中是否已经存在
		_, err = team.GetTeamMemberByRole(role)
		if err == nil {
			Report(w, r, "你好，该团队已经存在所选择的核心角色，请返回选择其他角色。")
			return
		} else if err != sql.ErrNoRows {
			util.Danger(err, su.Email, " Cannot search team member given team_id and role")
			Report(w, r, "你好，茶博士在这个团队角色事情迷糊了，请确认后再试。")
			return
		}

	case "CEO":
		if ceo.Id == founder.Id {
			//CEO是默认创建人担任首个CEO，这意味着首次更换CEO，例如观音菩萨指定唐僧为取经团队CEO，这是初始化团队操作
			break
		} else {
			Report(w, r, "你好，请先邀请用户加盟为普通用户，然后再调整角色，请确认后再试。")
			return
		}

	case "taster":
		// No additional validation needed for the "taster" role
		break
	default:
		Report(w, r, "你好，请选择正确的角色。")
		return
	}

	//检查team中是否存在teamMember
	_, err = data.GetTeamMemberByTeamIdAndUserId(team.Id, new_member.Id)
	if err != nil {
		//如果err类型为空行，说明团队中还没有这个用户，可以向其发送邀请函
		if err == sql.ErrNoRows {

			//创建一封邀请函
			invi := data.Invitation{
				TeamId:      team.Id,
				InviteEmail: new_member.Email,
				Role:        role,
				InviteWord:  i_word,
				CreatedAt:   time.Now(),
				Status:      0,
			}
			//存储邀请函
			err = invi.Create()
			if err != nil {
				util.Danger(err, su.Email, " Cannot create invitation")
				Report(w, r, "你好，茶博士未能创建邀请函，请稍后再试。")
				return
			}
			// 向受邀请的用户新消息小黑板上加1
			if err = data.AddUserMessageCount(new_member.Id); err != nil {
				util.Danger(err, " Cannot add user new-message count")
				return
			}

			// 报告发送者成功消息
			Report(w, r, "你好，成功向用户发送了邀请函，请��心等待。")
			return
		}
		//其他类型的error，打印出来分析错误
		util.Danger(err, su.Email, "error for Search teamMember given teamId and userId")
		return
	}
	//如果err为nil，说明用户已经在茶团中，无需邀请
	Report(w, r, "你好，该用户已经在茶团中，无需邀请。")

}

// GET /v1/team/team_member/invite_page?id=
// 对某一个用户的邀请函编写页面
func InviteMemberPage(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchUserRelatedData(sess)
	if err != nil {
		util.Danger(err, "cannot fetch s_u s_teams given session")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	//根据用户提交的Uuid，查询获取拟邀请加盟的用户信息
	vals := r.URL.Query()
	user_uuid := vals.Get("id")
	user, err := data.UserByUuid(user_uuid)
	if err != nil {
		util.Danger(err, " Cannot get user given uuid")
		Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？请确认后再试")
		return
	}

	var iD data.InvitationDetail
	// 填写页面资料
	iD.SessUser = s_u
	iD.SessUserDefaultTeam = s_default_team
	iD.SessUserSurvivalTeams = s_survival_teams
	iD.SessUserDefaultPlace = s_default_place
	iD.SessUserBindPlaces = s_places
	iD.InviteUser = user
	//检查一下用户是否有权以该茶团Team的名义发送邀请函
	//如果是茶团founder，或者是team CEO ，则可以发送邀请函
	ceo, err := s_default_team.CEO()
	if err != nil {
		//分析err类型
		if err == sql.ErrNoRows {
			//此茶团还没有指定CEO
			//如果用户是茶团创建人，返回指定的团队邀请函创建表单页面
			if s_default_team.FounderId == s_u.Id {
				GenerateHTML(w, &iD, "layout", "navbar.private", "member.invite")
				return
			}
			//无权发送邀请函
			Report(w, r, "你好，茶团在没有CEO之前，只有茶团创建人能发送该团邀请函。")
			return
		}
		util.Danger(err, " Cannot get team CEO")
		Report(w, r, "你好，茶博士在查找茶团CEO时迷失了，请稍后再试。")
		return
	}
	if s_default_team.FounderId == s_u.Id || ceo.UserId == s_u.Id {
		//向用户返回指定的团队邀请函创建表单页面
		GenerateHTML(w, &iD, "layout", "navbar.private", "member.invite")
		return
	}

	// 检查s_u是否某个茶团的ceo
	teams, err := s_u.CeoTeams()
	if err != nil {
		util.Danger(err, "cannot get teams given sessUser")
		Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？")
		return
	}
	if len(teams) > 0 {
		//向用户返回指定的团队邀请函创建表单页面
		GenerateHTML(w, &iD, "layout", "navbar.private", "member.invite")
		return
	}

	count, err := s_u.CountTeamsByFounderId()
	if err != nil {
		util.Danger(err, "cannot get teams given sessUser")
		Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？请稍后再试")
		return
	}
	if count > 0 {
		//向用户返回指定的团队邀请函创建表单页面
		GenerateHTML(w, &iD, "layout", "navbar.private", "member.invite")
		return
	}

	Report(w, r, "你好，慢条斯理的茶博士竟然说，先成为茶团CEO或者找创建人，才能发送该团邀请函噢。")

}

// GET /v1/team/team_member/invitation?id=
func InvitationDetail(w http.ResponseWriter, r *http.Request) {
	//获取session
	sess, err := Session(r)
	if err != nil {
		util.Danger(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, _ := sess.User()

	var iD data.InvitationDetail

	//根据用户提交的Uuid，查询邀请函信息
	vals := r.URL.Query()
	invi_uuid := vals.Get("id")
	invitation, err := data.GetInvitationByUuid(invi_uuid)
	if err != nil {
		util.Danger(err, " Cannot get invitation")
		Report(w, r, "你好，茶博士正在努力的查找邀请函，请稍后再试。")
		return
	}

	//检查一下当前用户是否有权查看此邀请函
	if invitation.InviteEmail != s_u.Email {
		Report(w, r, "你好，该邀请函不属于您，无法查看。")
		return
	}

	//填写页面资料
	iD.SessUser = s_u
	iD.Invitation = invitation
	iD.Team, err = invitation.Team()
	if err != nil {
		util.Danger(err, " Cannot get team")
		Report(w, r, "你好，������正在��力的查找��请��，请稍后再试。")
		return
	}
	iD.InviteUser, err = invitation.ToUser()
	if err != nil {
		util.Danger(err, " Cannot get user")
		Report(w, r, "你好，������正在��力的查找��请��，请稍后再试。")
		return
	}

	//如果邀请函目前是未读状态=0，则将邀请函的状态改为已读=1
	if invitation.Status == 0 {
		invitation.Status = 1
		err = invitation.Update()
		if err != nil {
			util.Danger(err, s_u.Email, " Cannot update invitation")
			Report(w, r, "你好，茶博士正在努力的更新邀请函状态，请稍后再试。")
			return
		}
		// 减去用户1小黑板新消息数
		if err = data.SubtractUserMessageCount(s_u.Id); err != nil {
			util.Danger(err, " Cannot subtract user message count")
			return
		}

	}
	//向用户返回该邀请函的详细信息
	GenerateHTML(w, &iD, "layout", "navbar.private", "member.invitation")
}

// GET /v1/team/team_member/quit?id=
// 退出某个茶团
func HandleMemberQuit(w http.ResponseWriter, r *http.Request) {
	Report(w, r, "你好，目前满头大汗的茶博士正在忙于抛砖引玉之中，欢迎加入茶棚服务。。")
}
