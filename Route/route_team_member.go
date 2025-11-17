package route

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// HandleMemberResign() /v1/team_member/resign
// 处理某个茶团的某个成员退出茶团声明撰写和提交
func HandleMemberResign(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		MemberResignGet(w, r)
	case http.MethodPost:
		MemberResignPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// POST /v1/team_member/resign
// 接收1个团队成员提交“退出茶团声明”资料
func MemberResignPost(w http.ResponseWriter, r *http.Request) {
	// 获取session
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 解析表单内容，获取当前用户提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 提交的声明标题
	titl := r.PostFormValue("title")
	// 检查提交的声明标题字数是否>2 and <64
	lenTit := cnStrLen(titl)
	if lenTit < 2 || lenTit > 64 {
		report(w, r, "你好，茶博士认为标题字数太长或者太短，请确认后再试。")
		return
	}

	// 提交的声明内容
	cont := r.PostFormValue("content")
	// 检查提交的声明内容字数是否>3 and <int(util.Config.ThreadMaxWord)
	lenCont := cnStrLen(cont)
	if lenCont < 3 || lenCont > int(util.Config.ThreadMaxWord) {
		report(w, r, "你好，茶博士认为内容字数太长或者太短，请确认后再试。")
		return
	}
	// 检查提交的成员邮箱
	m_email := r.PostFormValue("m_email")
	if ok := isEmail(m_email); !ok {
		report(w, r, "你好，涨红了脸的茶博士，竟然强词夺理说，电子邮箱格式太复杂看不懂，请确认后再提交。")
		return
	}
	//读取声明退出的成员资料
	t_user, err := data.GetUserByEmail(m_email, r.Context())
	if err != nil {
		util.Debug("Cannot get user by email", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查提交人是否和会话用户一致
	if s_u.Id != t_user.Id {
		report(w, r, "你好，目前不能代替他人提交退出茶团声明，请确认后再试。")
		return
	}

	// 提交的目标茶团
	team_id_str := r.PostFormValue("team_id")
	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		util.Debug("Cannot convert team_id to int", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 读取目标茶团资料
	t_team, err := data.GetTeam(team_id)
	if err != nil {
		util.Debug(team_id, "Cannot get team by id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 检查提交人是否为茶团成员
	t_member, err := data.GetMemberByTeamIdUserId(t_team.Id, t_user.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，你不是茶团成员，不接受退出声明噢。")
			return
		} else {
			util.Debug(t_team.Id, t_user.Id, "Cannot get team member by team id and user id", err)
			report(w, r, "你好，未能获取拟退出的茶团资料，请稍后再试。")
			return
		}
	}

	// //查看成员角色，分类处理：1、CEO，2、核心成员：CTO、CFO、CMO，3、普通成员：taster
	// switch t_member.Role {
	// case "taster":
	// 	break
	// case "CTO", RoleCFO, RoleCMO:
	// 	report(w, r, "你好，请先联系CEO，将你目前角色核心成员调整为普通成员品茶师，然后再声明退出。")
	// 	return
	// case RoleCEO:
	// 	report(w, r, "你好，请先联系茶团创建人，将你目前角色调整为普通成员品茶师，然后再声明退出。")
	// 	return
	// default:
	// 	report(w, r, "你好，满头大汗的茶博士表示找不到这个茶友角色，请确认后再试。")
	// 	return
	// }

	// 检查是否为核心成员（非CEO），如果是则先降级为普通成员
	if t_member.Role != data.RoleCEO && t_member.Role != "taster" {
		// 是核心成员（CTO/CFO/CMO），先降级为普通成员
		t_member.Role = "taster"
		if err := t_member.UpdateRoleStatus(); err != nil {
			util.Debug("Cannot update member role to taster", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
	}

	//声明一份茶团成员退出声明书
	tmqD := data.TeamMemberResignation{
		TeamId:            t_team.Id,
		CeoUserId:         data.UserId_None,
		CoreMemberUserId:  data.UserId_None,
		MemberId:          t_member.Id,
		MemberUserId:      t_user.Id,
		MemberCurrentRole: t_member.Role,
		Title:             titl,
		Content:           cont,
		Status:            data.ResignationStatusUnread,
	}

	//尝试保存退出声明
	if err := tmqD.Create(); err != nil {
		util.Debug("Cannot create team member resignation", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//返回成功保存声明的报告
	if t_member.Role == "taster" && tmqD.MemberCurrentRole != "taster" {
		report(w, r, "你好，作为核心成员，你的角色已自动调整为普通成员，退出声明已提交，请等待管理员审批。")
	} else {
		report(w, r, "你好，茶博士已经保存了你的退出声明，请等待管理员答复。")
	}

	//返回茶团主页
	//http.Redirect(w, r, fmt.Sprintf("/v1/team?uuid=%s", t_team.Uuid), http.StatusFound)

}

// MemberResignGet() GET /v1/team_member/resign?uuid=XXX
// 取出一张空白茶团成员“退出茶团声明”撰写页面
func MemberResignGet(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取提交的查询参数
	vals := r.URL.Query()
	team_uuid := vals.Get("uuid")

	//读取目标茶团资料
	t_team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(team_uuid, "Cannot get team by uuid", err)
		report(w, r, "你好，满头大汗的茶博士表示找不到这个茶团，稍后再试。")
		return
	}

	//检查目标茶友是否茶团成员
	_, err = data.GetMemberByTeamIdUserId(t_team.Id, s_u.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，您不是本茶团的成员，稍后再试。")
			return
		} else {
			util.Debug(t_team.Id, " when GetMemberByTeamIdAndUserId() checking team_member", err)
			report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	}

	var tmqPD data.TeamMemberResign
	tmqPD.SessUser = s_u
	tmqPD.Team = t_team

	//渲染退出声明撰写页面
	generateHTML(w, &tmqPD, "layout", "navbar.private", "member.resign_new")

}

// GET /v1/team_member/role_changed?uuid=XXX
// MemberRoleChanged() 某个茶团的全部已发布成员角色调整声明列表页面
func MemberRoleChanged(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取查询参数
	vals := r.URL.Query()
	team_uuid := vals.Get("uuid")

	//读取目标茶团资料
	t_team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(team_uuid, "Cannot get team by uuid", err)
		report(w, r, "你好，满头大汗的茶博士表示找不到这个茶团，稍后再试。")
		return
	}

	//读取这支茶团已发布的，全部成员角色调整声明
	role_notices, err := data.GetMemberRoleNoticesByTeamId(t_team.Id)
	if err != nil {
		util.Debug(t_team.Id, "Cannot get team member role change notice by team id", err)
		report(w, r, "你好，茶博士正在忙碌中，请稍后再试。")
		return
	}
	tmrnBeanSlice, err := fetchTeamMemberRoleNoticeBeanSlice(role_notices)
	if err != nil {
		util.Debug("Cannot fetch team member role notice bean slice", err)
		report(w, r, "你好，茶博士正在忙碌中，请稍后再试。")
		return
	}

	var tmrcnpd data.TeamMemberRoleChangedNoticesPageData
	tmrcnpd.SessUser = s_u
	tmrcnpd.Team = t_team
	tmrcnpd.TeamMemberRoleNoticeBeanSlice = tmrnBeanSlice

	//渲染茶团成员角色调整通知页面
	generateHTML(w, &tmrcnpd, "layout", "navbar.private", "member.role_changed_notices")

}

// Handle() /v1/team_member/role
// 调整茶团成员角色管理窗口
func HandleMemberRole(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//返回调整角色（撰写角色调整公告）页面
		MemberRoleChange(w, r)
	case http.MethodPost:
		//设置角色
		MemberRoleReply(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/team_member/role?id=XXX&m_email=XXX
// 取出一张空白茶团成员角色任命书
func MemberRoleChange(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取提交的查询参数
	vals := r.URL.Query()
	team_id_str := vals.Get("id")
	//检查提交的id参数格式是否正常
	team_id_int, err := strconv.Atoi(team_id_str)
	if err != nil {
		util.Debug(team_id_str, "Cannot convert team_id to int", err)
		report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
		return
	}

	member_email := vals.Get("m_email")

	//检查提交的email参数格式是否正常
	if ok := isEmail(member_email); !ok {
		report(w, r, "你好，茶博士的眼镜被闪电破坏了，看不清提及的邮箱，请稍后再试。")
		return
	}

	//读取目标茶团资料
	t_team, err := data.GetTeam(team_id_int)
	if err != nil {
		util.Debug(team_id_str, "Cannot get team by id")
		report(w, r, "你好，满头大汗的茶博士表示找不到这个茶团，稍后再试。")
		return
	}
	//读取拟调整角色目标茶友资料
	t_member, err := data.GetUserByEmail(member_email, r.Context())
	if err != nil {
		util.Debug(member_email, "Cannot get user given email", err)
		report(w, r, "你好，满头大汗的茶博士表示找不到这个茶友，稍后再试。")
		return
	}

	//检查目标茶友是否茶团成员
	_, err = data.GetMemberByTeamIdUserId(t_team.Id, t_member.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，满头大汗的茶博士表示这个不是茶团的成员，稍后再试。")
			return
		} else {
			util.Debug(t_team.Id, " when GetMemberByTeamIdAndUserId() checking team_member", err)
			report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	}

	//检查提交者（当前用户）是否茶团CEO
	//先读取茶团CEO成员资料
	member_ceo, err := t_team.MemberCEO()
	if err != nil {
		util.Debug(t_team.Id, "Cannot get team ceo given team id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取目标茶团创建人的资料，Founder也可以调整成员角色！包括CEO！太疯狂了？（参考自观音菩萨可以决定西天取经团队的任何角色人选）
	t_founder, err := t_team.Founder()
	if err != nil {
		util.Debug(t_team.Id, "Cannot get team founder given team id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
	}
	// 准备资料
	var tmrcnP data.TeamMemberRoleChangeNoticePage
	is_manager := false

	if t_founder.Id == s_u.Id {
		is_manager = true
		tmrcnP.IsCEO = false
	} else if member_ceo.UserId == s_u.Id {
		// 如果会话用户是CEO，可以调整目标成员角色
		is_manager = true
		tmrcnP.IsCEO = true
		// 然后检查目标茶友和目标茶团CEO身份，CEO不能自己调整自己的角色，（WHY？）
		if member_ceo.UserId == t_member.Id {
			is_manager = false
		}
	}

	if !is_manager {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你不是这个茶团的管理者，无权调整角色噢。")
		return
	}

	//读取CEO茶友资料
	t_ceo, err := data.GetUser(member_ceo.UserId)
	if err != nil {
		util.Debug(member_ceo.UserId, "Cannot get user by id", err)
		report(w, r, "你好，茶博士正在忙碌中，请稍后再试。")
		return
	}

	m_c_role, err := data.GetTeamMemberRoleByTeamIdAndUserId(t_team.Id, t_member.Id)
	if err != nil {
		util.Debug(t_team.Id, "Cannot get team member role given team id", err)
		report(w, r, "你好，茶博士正在忙碌中，请稍后再试。")
		return
	}

	//填充资料
	tmrcnP.SessUser = s_u
	tmrcnP.Team = t_team
	tmrcnP.TeamMemberRoleNoticeBean.Team = t_team
	tmrcnP.TeamMemberRoleNoticeBean.Founder = t_founder
	tmrcnP.TeamMemberRoleNoticeBean.CEO = t_ceo
	tmrcnP.TeamMemberRoleNoticeBean.Member = t_member
	tmrcnP.TeamMemberRoleNoticeBean.TeamMemberRoleNotice.MemberCurrentRole = m_c_role

	//渲染茶团角色调整页面
	generateHTML(w, &tmrcnP, "layout", "navbar.private", "member.role_change_new")
}

// POST /v1/team_member/role
// 提交一个成员新的团队角色任命书答复
func MemberRoleReply(w http.ResponseWriter, r *http.Request) {
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
	//解析表单内容，获取当前用户提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		http.Redirect(w, r, "/v1/", http.StatusFound)
		return
	}
	//提交的成员邮箱
	m_email := r.PostFormValue("m_email")
	if ok := isEmail(m_email); !ok {
		report(w, r, "你好，涨红了脸的茶博士，竟然强词夺理说，电子邮箱格式太复杂看不懂，请确认后再提交。")
		return
	}
	//提交的目标茶团
	team_id_str := r.PostFormValue("team_id")
	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		util.Debug(team_id_str, "Cannot convert team_id to int", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取目标茶团资料
	t_team, err := data.GetTeam(team_id)
	if err != nil {
		util.Debug(team_id, "Cannot get team by id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//提交的成员新角色参数
	new_role := r.PostFormValue("role")
	//提交的成员新角色参数是否正常
	switch new_role {
	case "taster":
		break
	case RoleCEO, "CTO", RoleCFO, RoleCMO:
		//需要检查目标角色是否空缺
		_, err = t_team.GetTeamMemberByRole(new_role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				//目标角色空缺,可以调整
				break
			} else {
				util.Debug(t_team.Id, new_role, "Cannot get team member by role", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
		} else {
			report(w, r, "你好，茶博士摸摸头嘀咕说，你提交的角色已经有人担任了，请确认后再提交。")
			return
		}

	default:
		report(w, r, "你好，茶博士摸摸头嘀咕说，你提交的角色不在茶团角色列表中，请确认后再提交。")
		return
	}

	//目标茶友
	t_member, err := data.GetUserByEmail(m_email, r.Context())
	if err != nil {
		util.Debug(m_email, "Cannot get user by email")
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查目标茶友是否茶团成员
	member, err := data.GetMemberByTeamIdUserId(t_team.Id, t_member.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，茶博士摸摸头嘀咕说，这个茶友不是茶团成员，无法调整角色。")
			return
		} else {
			util.Debug(t_team.Id, " when GetMemberByTeamIdAndUserId() checking team_member", err)
			report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	}
	//检查提交者是否在尝试调整自己的角色，不合规
	if member.UserId == s_u.Id {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你不能调整自己的角色。")
		return
	}
	//Role no change
	if new_role == member.Role {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你没有调整角色，无需提交。")
		return
	}

	//提交的角色调整标题
	title := r.PostFormValue("title")
	//检查提交的标题是否正常，中文字数>6,<24
	if cnStrLen(title) < 6 || cnStrLen(title) > 24 {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你提交的标题太长或太短，请确认后再提交。")
		return
	}
	//提交的角色调整内容
	content := r.PostFormValue("content")
	//检查提交的内容是否正常，中文字数>2,<int(util.Config.ThreadMaxWord),
	if cnStrLen(content) < 2 || cnStrLen(content) > int(util.Config.ThreadMaxWord) {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你提交的内容太长或太短，请确认后再提交。")
		return
	}

	//检查提交者（当前用户）是否茶团CEO？如果不是CEO，再检查是否是茶团创建人
	//读取CEO成员资料
	m_ceo, err := t_team.MemberCEO()
	if err != nil {
		util.Debug(t_team.Id, "Cannot get team ceo given team id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取目标茶团创建人资料
	t_founder, err := t_team.Founder()
	if err != nil {
		util.Debug(t_team.Id, "Cannot get team founder given team id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	if m_ceo.UserId == s_u.Id && new_role != RoleCEO {
		//会话用户是CEO，可以调整非CEO成员角色
		//创建一个新的团队成员角色变动公告
		team_member_role_notice := data.TeamMemberRoleNotice{
			TeamId:            t_team.Id,
			CeoId:             m_ceo.UserId,
			MemberId:          member.Id,
			MemberCurrentRole: member.Role,
			NewRole:           new_role,
			Title:             title,
			Content:           content,
			Status:            0,
		}
		if err = team_member_role_notice.Create(); err != nil {
			util.Debug(team_member_role_notice, "Cannot create team_member_role_notice", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		//更新成员角色
		member.Role = new_role
		if err = member.UpdateRoleStatus(); err != nil {
			util.Debug(member, "Cannot update member", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}

	} else if t_founder.Id == s_u.Id {
		//会话用户是创建人，可以调整CEO和非CEO成员角色
		//创建一个新的团队成员角色变动公告
		team_member_role_notice := data.TeamMemberRoleNotice{
			TeamId:            t_team.Id,
			CeoId:             m_ceo.UserId,
			MemberId:          member.Id,
			MemberCurrentRole: member.Role,
			NewRole:           new_role,
			Title:             title,
			Content:           content,
			Status:            0,
		}
		if err = team_member_role_notice.Create(); err != nil {
			util.Debug(team_member_role_notice, "Cannot create team_member_role_notice", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		// 更新成员角色
		member.Role = new_role
		if err = member.UpdateRoleStatus(); err != nil {
			util.Debug(member, "Cannot update member", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
	} else {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你不是这个茶团的管理者，无权调整角色噢。")
		return
	}

	//报告调整角色成功消息
	report(w, r, "你好，茶博士摸摸头说，已经调整了 "+t_member.Name+" 的角色为 "+new_role+" 。")
}

// /v1/team_member/invite
// 邀请一个指定的新茶友加入茶团
func HandleInviteMember(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//返回邀请团队新成员，即邀请函填写页面
		InviteMemberGet(w, r)
	case http.MethodPost:
		//生成邀请函方法
		InviteMemberPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// /v1/team_member/invitation/read
// 收到团队邀请函成员视角，阅读&处理1份茶团邀请函
func HandleMemberInvitationRead(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//返回邀请函阅读&处理页面
		MemberInvitationRead(w, r)
	case http.MethodPost:
		//设置邀请函回复方法
		MemberInvitationReply(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// /v1/team_member/application/new
// 申请加入一个开放式茶团
func HandleNewMemberApplication(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		NewMemberApplicationForm(w, r)
	case http.MethodPost:

		NewMemberApplication(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// /v1/team_member/application/review
// 团队管理员审查，处理茶团一份加盟申请书
func HandleMemberApplicationReview(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		MemberApplicationReview(w, r)
	case http.MethodPost:
		MemberApplicationReply(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GET /v1/team_member/application/detail?uuid=xxx
// 查看1份加盟申请书详情
func MemberApplicationDetail(w http.ResponseWriter, r *http.Request) {
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

	vals := r.URL.Query()
	application_uuid := vals.Get("uuid")
	application := data.MemberApplication{Uuid: application_uuid}
	if err = application.GetByUuid(); err != nil {
		util.Debug(application_uuid, "cannot get application given uuid", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	applicant, err := data.GetUser(application.UserId)
	if err != nil {
		util.Debug(application.UserId, "Cannot get user given id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	team, err := data.GetTeam(application.TeamId)
	if err != nil {
		util.Debug(application.TeamId, "Cannot get team given id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//权限检查，用户可以查看自己的申请书详情，或者是团队管理员可以
	if applicant.Id != s_u.Id && !canManageTeam(&team, s_u.Id, w, r) {
		return
	}

	applicant_default_team, err := applicant.GetLastDefaultTeam()
	if err != nil {
		util.Debug(applicant.Email, "Cannot get default team given user", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	var aR data.ApplicationReview
	aR.SessUser = s_u
	aR.Application = application
	aR.Team = team
	aR.Applicant = applicant
	aR.ApplicantDefaultTeam = applicant_default_team

	generateHTML(w, &aR, "layout", "navbar.private", "application.detail")
}

// POST /v1/team_member/application/review
// 加盟茶团申请书审查人，提交处理（审查）决定
func MemberApplicationReply(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//解析表单内容，获取茶友提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Debug(s.Email, "Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	//读取提交的参数
	application_id_str := r.PostFormValue("application_id")
	application_id, err := strconv.Atoi(application_id_str)
	if err != nil {
		util.Debug(application_id_str, "Cannot convert application_id to int", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	application := data.MemberApplication{
		Id: application_id,
	}
	//读取加盟茶团申请书
	if err = application.Get(); err != nil {
		util.Debug(application_id, "Cannot get application given id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查申请书的状态是否正常，已查看
	switch application.Status {
	case data.MemberApplicationStatusPending:
		//未查看
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	case data.MemberApplicationStatusViewed:
		//已查看，但未处理
		break
	case data.MemberApplicationStatusApproved, data.MemberApplicationStatusRejected:
		// 已经处理完成
		report(w, r, "你好，这份申请书已经被处理，请确认后再试。")
		return
	case data.MemberApplicationStatusExpired:
		//已经过期或者失效
		report(w, r, "你好，这份申请书已经过期或者失效，请确认后再试。")
		return
	default:
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
	}

	//读取申请人申请加盟的茶团
	team, err := data.GetTeam(application.TeamId)
	if err != nil {
		util.Debug(application.TeamId, "Cannot get team given id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	is_core_member, err := team.IsCoreMember(s_u.Id)
	if err != nil {
		util.Debug(team.Id, "Cannot get core members of team", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 如果不是茶团的核心成员，返回错误
	if !is_core_member {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你不是茶团的核心成员，无权处理申请书噢。")
		return
	}

	//读取申请人资料
	applicant, err := data.GetUser(application.UserId)
	if err != nil {
		util.Debug(application.UserId, "Cannot get user given id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查申请人是否已经是茶团成员=这个茶团是否已经存在该茶友
	_, err = data.GetMemberByTeamIdUserId(team.Id, applicant.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Debug(applicant.Email, " when checking team_member", err)
			report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经在茶团中了噢？请确认后再试。")
		return
	}
	//检查当前会话茶友是否和审查足迹资料中审查人一致

	//读取提交的回复内容
	reply := r.PostFormValue("reply")
	if reply == "" {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你没有填写回复内容或者内容太复杂了。")
		return
	}

	//读取提交的审查结果参数
	approval_str := r.PostFormValue("approval")
	approval_int, err := strconv.Atoi(approval_str)
	if err != nil {
		util.Debug(approval_str, "Cannot convert approval to int", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//如果approval_int = 1，是批准加盟；如果 = 0 ，是婉拒
	switch approval_int {
	case 1:
		//批准加盟
		//创建一个新的茶团成员
		team_member := data.TeamMember{
			TeamId: team.Id,
			UserId: applicant.Id,
			Role:   "taster",
			Status: 1,
		}
		//将新的茶团成员写入数据库
		if err = team_member.Create(); err != nil {
			util.Debug(team_member, "Cannot create team_member", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		//检查茶友加入茶团计数，如果是1，从默认的自由人，改设当前茶团为默认茶团
		count, err := applicant.SurvivalTeamsCount()
		if err != nil {
			util.Debug(applicant.Email, " Cannot get survival teams count", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		if count == 0 {
			user_default_team := data.UserDefaultTeam{
				UserId: applicant.Id,
				TeamId: team.Id,
			}
			if err = user_default_team.Create(); err != nil {
				util.Debug(applicant.Email, " Cannot create user_default_team", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
		}

		//更新申请书的状态为已批准
		application.Status = data.MemberApplicationStatusApproved
		if err = application.Update(); err != nil {
			util.Debug(application, "Cannot update application", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		//创建批准加盟申请书答复
		application_reply := data.MemberApplicationReply{
			MemberApplicationId: application.Id,
			TeamId:              team.Id,
			UserId:              s_u.Id,
			ReplyContent:        reply,
			Status:              data.MemberApplicationStatusApproved,
		}
		if err = application_reply.Create(); err != nil {
			util.Debug(application_reply, "Cannot create application_reply", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}

		//报告批准加盟成功消息
		report(w, r, "你好，已经批准新成员 "+applicant.Name+" 加盟 "+team.Name+" 茶团。")
		return
	case 0:
		//婉拒加盟
		application.Status = data.MemberApplicationStatusRejected
		if err = application.Update(); err != nil {
			util.Debug(application, "Cannot update application", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		//创建婉拒加盟申请书答复
		application_reply := data.MemberApplicationReply{
			MemberApplicationId: application.Id,
			TeamId:              team.Id,
			UserId:              s_u.Id,
			ReplyContent:        reply,
			Status:              data.MemberApplicationStatusRejected,
		}
		if err = application_reply.Create(); err != nil {
			util.Debug(application_reply, "Cannot create application_reply", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		report(w, r, "你好，茶团成员 "+applicant.Email+" 已经婉拒加盟茶团 "+team.Name+"。")
		return
	default:
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

}

// GET /v1/team_member/application/review?uuid=xxx
// 团队管理员打开一份新加盟申请书，审查其内容
func MemberApplicationReview(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取提交的查询参数
	vals := r.URL.Query()
	application_uuid := vals.Get("uuid")
	application := data.MemberApplication{
		Uuid: application_uuid,
	}
	// 读取申请书
	if err = application.GetByUuid(); err != nil {
		util.Debug(application_uuid, "cannot get application given uuid", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 检查申请书状态，如果已经处理 status>1（MemberApplicationStatusViewed），跳转个人加盟团队申请书详情
	if application.Status > data.MemberApplicationStatusViewed {
		http.Redirect(w, r, "/v1/team_member/application/detail?uuid="+application_uuid, http.StatusFound)
		return
	}

	//读取申请人资料
	applicant, err := data.GetUser(application.UserId)
	if err != nil {
		util.Debug(application.UserId, "Cannot get user given id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取申请人申请加盟的茶团
	team, err := data.GetTeam(application.TeamId)
	if err != nil {
		util.Debug(application.TeamId, "Cannot get team given id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 检查这个茶团是否已经存在该茶友
	_, err = data.GetMemberByTeamIdUserId(team.Id, applicant.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Debug(applicant.Email, " when checking team_member", err)
			report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		report(w, r, "你好，这个申请人已经在茶团中了噢？请确认后再试。")
		return
	}
	// 读取申请人默认所在的茶团
	applicant_default_team, err := applicant.GetLastDefaultTeam()
	if err != nil {
		util.Debug(applicant.Email, "Cannot get default team given user", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//检查当前会话茶友身份，是否team的管理成员（核心成员）
	core_members, err := team.CoreMembers()
	if err != nil {
		util.Debug(team.Id, "Cannot get core members of team", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查当前茶友是否是茶团的核心成员
	is_core_member := false
	for _, core_member := range core_members {
		if core_member.UserId == s_u.Id {
			is_core_member = true
			break
		}
	}
	//如果不是茶团的核心成员，返回错误
	if !is_core_member {
		report(w, r, "你好，茶博士摸摸头嘀咕说，你不是茶团的核心成员，无权审查申请书噢。")
		return
	}

	//准备全部资料
	var aR data.ApplicationReview
	aR.SessUser = s_u
	//申请书信息
	aR.Application = application
	aR.Team = team
	aR.Applicant = applicant
	aR.ApplicantDefaultTeam = applicant_default_team

	//记录查看足迹
	// footprint := data.Footprint{
	// 	UserId:    s_u.Id,
	// 	TeamId:    team.Id,
	// 	TeamName:  team.Abbreviation,
	// 	Content:   fmt.Sprintf("查看了茶团 %s 的加盟申请书", team.Name),
	// 	ContentId: application.Id,
	// }
	// if err = footprint.Create(); err != nil {
	// 	util.Debug("Cannot create footprint", err)
	// }
	//修改申请书状态为已查看
	application.Status = data.MemberApplicationStatusViewed
	if err = application.Update(); err != nil {
		util.Debug(application.Id, "Cannot update application status", err)
	}

	//渲染页面
	generateHTML(w, &aR, "layout", "navbar.private", "team.application_review", "component_member_application_bean")
}

// POST /v1/team_member/application/new
// 个人递交一份新的，茶团加盟申请书，提交处理
func NewMemberApplication(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//读取表单数据
	team_uuid := r.FormValue("team_uuid")
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(team_uuid, "Cannot get team given uuid", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	user_uuid := r.FormValue("user_uuid")
	app_user, err := data.GetUserByUUID(user_uuid)
	if err != nil {
		util.Debug(user_uuid, "Cannot get user given uuid", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	content := r.FormValue("content")
	//check content length
	if len(content) > 666 || len(content) < 2 {
		report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你的申请书内容太长了噢？墨水瓶也怕抄单词呀！")
		return
	}

	// s_u.Id != user.Id，检查是否茶友本人提交申请，不允许代他人申请
	if s_u.Id != app_user.Id {
		report(w, r, "你好，身前有余忘缩手，眼前无路想回头，目前仅接受本人申请加入茶团噢。")
		return
	}

	//检查这个茶团是否已经存在该茶友
	_, err = data.GetMemberByTeamIdUserId(team.Id, app_user.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Debug(app_user.Email, " when checking team_member", err)
			report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经在茶团中了噢？请确认后再试。")
		return
	}

	//创建茶团新茶友加盟申请书
	ma := data.MemberApplication{
		TeamId:  team.Id,
		UserId:  app_user.Id,
		Content: content,
		Status:  0,
	}
	//保存申请书
	if err = ma.Create(); err != nil {
		util.Debug(team.Id, "Cannot create member-application", err)
		report(w, r, "你好，茶博士正在飞速处理所有的技术问题，请耐心等待。")
		return
	}
	//发送邮件通知茶团管理员（等待茶团管理员上班查看茶团详情即可见申请书，不另外通知）

	//返回成功页面
	t := fmt.Sprintf("你好，%s ，加盟 %s 申请书已经提交，请等待茶团管理员的回复。", s_u.Name, team.Abbreviation)
	report(w, r, t)

}

// GET /v1/team_member/application/new?uuid=
// 返回一份个人，空白的团队加盟申请表单
func NewMemberApplicationForm(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 读取提交的查询参数
	vals := r.URL.Query()
	team_uuid := vals.Get("uuid")
	// 读取茶团资料
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug("Cannot get team given uuid", team_uuid, err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//检测当前用户是否向指定茶团，已经提交过加盟申请？而且申请书状态为（Status<=status）
	_, err = data.CheckMemberApplicationByTeamIdAndUserId(team.Id, s_u.Id, data.MemberApplicationStatusViewed)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Debug(" when checking member_application", s_u.Email, err)
			report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经提交过申请书了噢？请确认后再试。")
		return
	}

	//检查这个茶团是否已经存在该茶友
	ok, err := team.IsMember(s_u.Id)
	if err != nil {
		util.Debug("Cannot check team member", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	if ok {
		report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经在茶团中了噢？请确认后再试。")
		return
	}

	var tD data.TeamDetail
	tD.SessUser = s_u
	tD.TeamBean.Team = team
	//渲染页面
	generateHTML(w, &tD, "layout", "navbar.private", "member.application_new")

}

// POST /v1/team_member/invitation/read
// 团队邀请对象，提交一份邀请函答复书
// 根据答复选择，对应处理加入或者拒绝
func MemberInvitationReply(w http.ResponseWriter, r *http.Request) {
	//解析表单内容，获取茶友提交的内容
	err := r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 根据会话信息读取茶友资料
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, r, "你好，满地梨花一片天，请稍后再试一次")
		return
	}
	//读取提交的参数
	invitation_id, err := strconv.Atoi(r.PostFormValue("invitation_id"))
	if err != nil {
		util.Debug("failed to convert invitation_id to int", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	user_id, err := strconv.Atoi(r.PostFormValue("user_id"))
	if err != nil {
		util.Debug("Failed to convert user_id to int", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查是否存在该茶友注册资料
	reply_user, err := data.GetUser(user_id)
	if err != nil {
		util.Debug("Cannot get user given id", err)
		report(w, r, "你好，茶博士报告茶友资料查询繁忙，稍后再试。")
		return
	}

	//检查一下提交的茶友和会话茶友Id是否一致
	if reply_user.Id != s_u.Id {
		util.Debug("Inconsistency between submitted user id and session id", err)
		report(w, r, "你好，请勿冒充八戒骗孙悟空的芭蕉扇哦，稍后再试。")
		return
	}
	//根据茶友提交的invitation_id，检查是否存在该邀请函
	invitation, err := data.GetInvitationById(invitation_id)
	if err != nil {
		util.Debug(" Cannot get invitation", err)
		report(w, r, "你好，秋阴捧出何方雪？雨渍添来隔宿痕。稍后再试。")
		return
	}
	//检查一下邀请函是否已经被回复
	if invitation.Status > data.InvitationStatusViewed {
		report(w, r, "你好，这个邀请函已经答复或者已过期。")
		return
	}
	invi_user, err := data.GetUserByEmail(invitation.InviteEmail, r.Context())
	if err != nil {
		util.Debug(" Cannot get invited user given invitation's email", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	if invi_user.Id != s_u.Id {
		report(w, r, "你好，茶博士摸摸头嘀咕说，这个邀请函不是你的，你无权回答。")
		return
	}

	reply_class_int, err := strconv.Atoi(r.PostFormValue("reply"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	reply_word := r.PostFormValue("invitation_reply")
	//检查一下茶友提交的string，即reply_word是否不为空，中文长度小于int(util.Config.ThreadMaxWord)字符之间
	if reply_word == "" || cnStrLen(reply_word) > int(util.Config.ThreadMaxWord) {
		util.Debug(s_u.Email, " Cannot process invitation", err)
		report(w, r, "你好，瞪大眼睛涨红了脸的茶博士，竟然强词夺理说，答复的话太长了或者太短，只有外星人才接受呀，请确认再试。")
		return
	}

	//读取目标茶团资料
	team, err := data.GetTeam(invitation.TeamId)
	if err != nil {
		util.Debug(" Cannot get team by id", err)
		report(w, r, "你好，丢了眼镜的茶博士忙到现在，还没有找到茶团登记本，请稍后再试。")
		return
	}

	// 检查这个茶团是否已经存在该茶友了
	is_member, err := team.IsMember(invi_user.Id)
	if err != nil {
		util.Debug(" when checking team_member", err)
		report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
		return
	}
	if is_member {
		report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经在茶团中了噢？请确认后再试。")
		return
	}

	if reply_class_int == 1 {
		//接受加盟邀请!

		//检查受邀请的茶友团队角色是否已经被占用
		//   - CTO/CMO/CFO: 唯一性检查，不能重复
		//   - CEO: 唯一性检查，但允许团队创始人重新指定
		//   - taster: 无限制
		//   - 其他角色: 拒绝
		switch invitation.Role {
		case RoleCTO, RoleCMO, RoleCFO:
			//检查teamMember.Role是否已经存在
			member, err := team.CheckTeamMemberByRole(invitation.Role)
			if err != nil {
				util.Debug(" Check team member given team_id and role failed", err)
				report(w, r, "你好，茶博士报告说，天气太热茶壶开锅了，请稍后再试。")
				return
			}
			if member == nil {
				// 角色可用
				break
			} else {
				// 角色已被占用
				report(w, r, "你好，该团队已经存在所选择的核心角色，请确认所选择的角色是否恰当。")
				return
			}
		case RoleCEO:
			//检查teamMember.Role中是否已经存在"CEO"
			member_existingCEO, err := team.MemberCEO()
			if err != nil {
				//由于$事业茶团CEO角色不能空缺，所以空记录也是系统错误
				util.Debug(" Get team-member CEO given team_id and role failed", err)
				report(w, r, "你好，该团队构建资料缺失，请稍后再试。")
				return
			}
			//是否团队发起人首次指定CEO？
			if member_existingCEO.UserId == team.FounderId {
				//是发起人首次指定CEO，允许更换新的CEO
				break
			}
			//不是发起人首次指定CEO，
			//报告角色已被占用
			report(w, r, "你好，该团队已经存在所选择的核心角色，请确认所选择的角色是否恰当。")
			return

		case "taster":
			// No additional validation needed for the "taster" role
			break
		default:
			report(w, r, "你好，请选择正确的团队角色。")
			return
		}

		//已经接受邀请，而且通过了角色冲突检查，则升级邀请函状态并保存答复话语和时间
		invitation.Status = data.InvitationStatusAccepted
		if err = invitation.UpdateStatus(); err != nil {
			util.Debug(invitation.InviteEmail, " Cannot update invitation status", err)
			report(w, r, "你好，茶博士报告开水太烫了，请稍后再试。")
			return
		}
		repl := data.InvitationReply{
			InvitationId: invitation_id,
			UserId:       user_id,
			ReplyWord:    reply_word,
		}
		if err = repl.Create(); err != nil {
			util.Debug(invitation.InviteEmail, " Cannot create invitation_reply", err)
			report(w, r, "你好，茶博士报告开水太烫了，请稍后再试。")
			return
		}
		// 准备将新成员添加进茶团所需的资料
		team_member := data.TeamMember{
			TeamId: invitation.TeamId,
			UserId: reply_user.Id,
			Role:   invitation.Role,
			Status: data.TeMemberStatusActive,
		}

		// 如果team_member.Role == "CEO",采取更换CEO方法
		if team_member.Role == RoleCEO {
			if err = team_member.UpdateFirstCEO(reply_user.Id); err != nil {
				util.Debug(s_u.Email, " Cannot update team_member CEO", err)
				report(w, r, "你好，幽情欲向嫦娥诉，无奈虚廊夜色昏。请稍后再试。")
				return
			}

		} else {
			// 其它角色
			if err = team_member.Create(); err != nil {
				util.Debug(s_u.Email, " Cannot create team_member", err)
				report(w, r, "你好，晕头晕脑的茶博士竟然忘记登记新成员了，请稍后再试。")
				return
			}
		}

		//检查茶友加入茶团计数，如果是0，从默认的自由人，改设当前茶团为默认茶团
		count, err := reply_user.SurvivalTeamsCount()
		if err != nil {
			util.Debug(reply_user.Email, " Cannot get survival teams count", err)
			report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		if count == 0 {
			user_default_team := data.UserDefaultTeam{
				UserId: reply_user.Id,
				TeamId: invitation.TeamId,
			}
			if err = user_default_team.Create(); err != nil {
				util.Debug(reply_user.Email, " Cannot create user_default_team", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
		}

		//返回此茶团页面给茶友，成员列表上有该茶友，示意已经加入该茶团成功
		http.Redirect(w, r, "/v1/team/detail?uuid="+(team.Uuid), http.StatusFound)
		return

	} else if reply_class_int == 0 {
		//拒绝邀请，则改写邀请函状态并保存答复话语和时间
		invitation.Status = data.InvitationStatusRejected
		if err = invitation.UpdateStatus(); err != nil {
			util.Debug(s_u.Email, " Cannot update invitation status", err)
			report(w, r, "你好，晕头晕脑的茶博士竟然把邀请答复处理搞混了，请稍后再试。")
			return
		}
		repl := data.InvitationReply{
			InvitationId: invitation_id,
			UserId:       user_id,
			ReplyWord:    reply_word,
		}
		if err = repl.Create(); err != nil {
			util.Debug(s_u.Email, " Cannot create invitation_reply", err)
			report(w, r, "你好，晕头晕脑的茶博士竟然把邀请答复搞丢了，请稍后再试。")
			return
		}
		//报告用户已经保存拒绝该邀请到答复记录
		t := fmt.Sprintf("你好，茶博士已经保存关于 %s 婉拒加盟答复。", team.Abbreviation)
		report(w, r, t)
		return

	} else {
		// 无效的reply 数值
		report(w, r, "你好，何幸邀恩宠，宫车过往频。稍后再试。")
		return
	}
}

// POST /v1/team_member/invite
// 团队管理员，提交一封新的邀请函内容。
func InviteMemberPost(w http.ResponseWriter, r *http.Request) {
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
	//解析表单内容，获取茶友提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		http.Redirect(w, r, "/v1/", http.StatusFound)
		return
	}
	email := r.PostFormValue("email")
	if ok := isEmail(email); !ok {
		report(w, r, "你好，涨红了脸的茶博士，竟然强词夺理说，电子邮箱格式太复杂看不懂，请确认后再提交。")
		return
	}

	i_word := r.PostFormValue("invite_word")
	//检查一下茶友提交的string，即i_word是否不为空，中文长度小于239字符之间
	if 1 > cnStrLen(i_word) || cnStrLen(i_word) > 239 {
		util.Debug(" Cannot process invitation", err)
		report(w, r, "你好，瞪大眼睛涨红了脸的茶博士，竟然强词夺理说，邀请的话太长了或者太短，只有外星人才接受呀，请确认再试。")
		return
	}
	role := r.PostFormValue("role")
	team_uuid := r.PostFormValue("team_uuid")
	if role == "" || team_uuid == "" {
		report(w, r, "你好，请选择团队角色和团队资料信息。")
		return
	}
	author_uuid := r.PostFormValue("author_uuid")
	if author_uuid != s_u.Uuid {
		report(w, r, "你好，请勿冒充八戒骗孙悟空的芭蕉扇哦，稍后再试。")
		return
	}
	author, err := data.GetUserByUUID(author_uuid)
	if err != nil {
		util.Debug(" Cannot get author user by uuid", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查邀请的茶友是否存在
	//根据茶友提交的Uuid，检查是否存在该User
	invite_user, err := data.GetUserByEmail(email, r.Context())
	if err != nil {
		util.Debug(" Cannot search user given email", err)
		report(w, r, "你好，满头大汗的茶博士未能茶棚里找到这个茶友，请确认后再试。")
		return
	}

	//检查茶友是否自己邀请自己？
	//也许是可以的?例如观音菩萨也可以加入自己创建的西天取经茶团喝茶？？
	// if s_u.Email == email {
	// 	report(w, r, "你好，请不要邀请自己加入茶团哈。")
	// 	return
	// }
	//根据茶友提交的teamId，检查是否存在该team
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(" Cannot search team given team_uuid", err)
		report(w, r, "你好，茶博士未能找到这个团队，请确认后再试。")
		return
	}
	//检查当前茶友是否团队的Ceo或者founder，是否有权限邀请新成员

	ceo_member, err := team.MemberCEO()
	if err != nil {
		util.Debug(" Cannot search team ceo", err)
		report(w, r, "你好，未能找到茶团CEO，请确认后再试。")
		return
	}
	ceo_user, err := ceo_member.User()
	if err != nil {
		util.Debug(" Cannot search ceo_user", err)
		report(w, r, "你好，未能找到茶团CEO，请确认后再试。")
		return
	}

	founder, err := team.Founder()
	if err != nil {
		util.Debug(" Cannot search team founder", err)
		report(w, r, "你好，未能找到这个茶团的发起人，请确认后再试。")
		return
	}
	ok := founder.Id == s_u.Id ||
		s_u.Id == ceo_user.Id
	if !ok {
		report(w, r, "你好，只有团队CEO或者创建人可以邀请请新成员加盟。")
		return
	}

	//检查受邀请的茶友团队核心角色是否已经被占用
	switch role {
	case RoleCTO, RoleCMO, RoleCFO:
		//检查teamMember.Role中是否已经存在
		_, err = team.GetTeamMemberByRole(role)
		if err == nil {
			report(w, r, "你好，该团队已经存在所选择的核心角色，请返回选择其他角色。")
			return
		} else if !errors.Is(err, sql.ErrNoRows) {
			util.Debug(" Cannot search team member given team_id and role", err)
			report(w, r, "你好，茶博士在这个团队角色事情迷糊了，请确认后再试。")
			return
		}

	case RoleCEO:
		if ceo_user.Id == founder.Id {
			//CEO是默认创建人担任首个CEO，这意味着首次更换CEO，ok。
			//例如,西天取经团队发起人观音菩萨（默认首个ceo），指定第一个成员唐僧取代自己为取经团队CEO
			break
		} else {
			report(w, r, "你好，请先邀请茶友加盟为普通茶友，然后再调整角色，请确认后再试。")
			return
		}

	case "taster":
		// No additional validation needed for the "taster" role
		break
	default:
		//其他非法角色，不允许
		report(w, r, "你好，请选择正确的角色。")
		return
	}

	//检查team中是否存在teamMember
	_, err = data.GetMemberByTeamIdUserId(team.Id, invite_user.Id)
	if err != nil {
		//如果err类型为空行，说明团队中还没有这个茶友，可以向其发送加盟邀请函
		if errors.Is(err, sql.ErrNoRows) {

			//创建一封邀请函
			invi := data.Invitation{
				TeamId:       team.Id,
				InviteEmail:  invite_user.Email,
				Role:         role,
				InviteWord:   i_word,
				Status:       data.InvitationStatusPending,
				AuthorUserId: author.Id,
			}
			//存储邀请函
			if err = invi.Create(); err != nil {
				util.Debug(" Cannot create invitation", err)
				report(w, r, "你好，茶博士未能创建邀请函，请稍后再试。")
				return
			}
			// 向受邀请的茶友新消息小黑板上加1
			if err = data.AddUserMessageCount(invite_user.Id); err != nil {
				util.Debug(" Cannot add user new-message count", err)
				return
			}

			// 报告发送者成功消息
			mes := fmt.Sprintf("你好，成功以 %s 茶团名义，向茶友 %s 发送了加盟邀请函，请等待回复。", team.Abbreviation, invite_user.Name)
			report(w, r, mes)
			return
		}
		//其他类型的error，打印出来分析错误
		util.Debug("error for Search teamMember given teamId and userId", err)
		return
	}
	//如果err为nil，说明茶友已经在茶团中，无需邀请
	report(w, r, "你好，该茶友已经在茶团中，无需邀请。")

}

// GET /v1/team_member/invite?uuid=(user.uuid)&team_uuid=(team.uuid)
// 团队管理员，需要一份新的邀请函表单，
// 用于邀请某个看中的茶友到teamId指定的团队
func InviteMemberGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("cannot fetch sess_user given session", err)
		report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	//根据茶友提交的Uuid，查询获取拟邀请加盟的茶友信息
	vals := r.URL.Query()
	user_uuid := vals.Get("uuid")
	if user_uuid == "" {
		report(w, r, "你好，桃李明年能再发，明年闺中知有谁？请确认后再试")
		return
	}
	invi_user, err := data.GetUserByUUID(user_uuid)
	if err != nil {
		util.Debug(" Cannot get user given uuid", err)
		report(w, r, "你好，桃李明年能再发，明年闺中知有谁？请确认后再试")
		return
	}
	team_uuid := vals.Get("team_uuid")
	if team_uuid == "" {
		report(w, r, "你好，桃李明年能再发，明年闺中知有谁？请确认后再试")
		return
	}
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(" Cannot get team given uuid", err)
		report(w, r, "你好，桃李明年能再发，明年闺中知有谁？请确认后再试")
		return
	}

	var iD data.InvitationDetail
	// 填写页面资料
	iD.SessUser = s_u

	iD.InvitationBean.Team = team
	iD.InvitationBean.InviteUser = invi_user

	//检查一下s_u茶友是否有权以某个茶团Team的名义发送邀请函

	//首先检查是否这个茶团founder或者CEO，则可以发送邀请函
	if team.FounderId == s_u.Id {
		founder, err := team.Founder()
		if err != nil {
			util.Debug("cannot get team's founder given team", err)
			report(w, r, "你好，桃李明年能再发，明年闺中知有谁？")
			return
		}
		iD.InvitationBean.Author = founder
		//向茶友返回指定的团队邀请函创建表单页面
		generateHTML(w, &iD, "layout", "navbar.private", "member.invite")
		return
	}

	// 检查s_u是否这个茶团的ceo
	m_ceo, err := team.MemberCEO()
	if err != nil {
		util.Debug("cannot get member ceo given team", err)
		report(w, r, "你好，桃李明年能再发，明年闺中知有谁？")
		return
	}

	if m_ceo.UserId == s_u.Id {
		ceo, err := data.GetUser(m_ceo.UserId)
		if err != nil {
			util.Debug("cannot get user given id", err)
			report(w, r, "你好，桃李明年能再发，明年闺中知有谁？")
			return
		}
		iD.InvitationBean.Author = ceo
		//向茶友返回指定的团队邀请函创建表单页面
		generateHTML(w, &iD, "layout", "navbar.private", "member.invite")
		return
	}

	//既不是某个茶团发起人，也不是CEO，无法代表茶团发出邀请函
	report(w, r, "你好，慢条斯理的茶博士竟然说，茶团CEO或者创建人，才能发送该团邀请函呢。")

}

// GET /v1/team_member/invitation/detail?uuid=
// 团队管理员查看邀请函详情
func MemberInvitationDetail(w http.ResponseWriter, r *http.Request) {
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

	vals := r.URL.Query()
	invi_uuid := vals.Get("uuid")
	invi, err := data.GetInvitationByUuid(invi_uuid)
	if err != nil {
		util.Debug(invi_uuid, "Cannot get invitation given uuid", err)
		report(w, r, "你好，茶博士正在努力的查找邀请函，请稍后再试。")
		return
	}

	// 读取目标茶团资料
	team, err := data.GetTeam(invi.TeamId)
	if err != nil {
		util.Debug(invi.TeamId, "Cannot get team by id", err)
		report(w, r, "你好，茶博士正在努力的查找邀请函资料，请稍后再试。")
		return
	}

	// 权限检查：团队管理员可以查看
	if !canManageTeam(&team, s_u.Id, w, r) {
		return
	}

	// 获取邀请函Bean
	i_b, err := fetchInvitationBean(invi)
	if err != nil {
		util.Debug(invi.Id, "Cannot fetch invitationBean given invitation", err)
		report(w, r, "你好，茶博士正在努力的查找邀请函资料，请稍后再试。")
		return
	}
	// 读取答复
	reply, err := invi.Reply()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		util.Debug(invi.Id, "Cannot get invitation reply", err)
		report(w, r, "你好，茶博士正在努力的查找邀请函答复资料，请稍后再试。")
		return
	}

	var iD data.InvitationDetail
	iD.SessUser = s_u
	iD.InvitationBean = i_b
	iD.Reply = reply
	generateHTML(w, &iD, "layout", "navbar.private", "team.invitation_detail")
}

// GET /v1/team_member/invitation/read?uuid=
// 用户查看收到的某封加盟邀请函详情及处理页面
func MemberInvitationRead(w http.ResponseWriter, r *http.Request) {
	//获取session
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var iD data.InvitationDetail

	//根据茶友提交的Uuid，查询邀请函信息
	vals := r.URL.Query()
	invi_uuid := vals.Get("uuid")
	invi, err := data.GetInvitationByUuid(invi_uuid)
	if err != nil {
		util.Debug(" Cannot get invitation given uuid", invi_uuid, err)
		report(w, r, "你好，茶博士正在努力的查找邀请函，请稍后再试。")
		return
	}

	//检查一下当前茶友是否有权查看此邀请函,仅本人可以查看
	// team, err := data.GetTeam(invi.TeamId)
	// if err != nil {
	// 	util.Debug(" Cannot get team by id", err)
	// 	report(w, r, "你好，茶博士正在努力的查找邀请函资料，请稍后再试。")
	// 	return
	// }

	//if invi.InviteEmail != s_u.Email && !canManageTeam(&team, s_u.Id, w, r) {
	if invi.InviteEmail != s_u.Email {
		report(w, r, "你好，该邀请函不属于您，无法查看。")
		return
	}

	i_b, err := fetchInvitationBean(invi)
	if err != nil {
		util.Debug(invi.Id, " Cannot fetch invitationBean given invitation", err)
		report(w, r, "你好，茶博士正在努力的查找邀请函资料，请稍后再试。")
		return
	}

	//如果邀请函目前是未读状态=0，则将邀请函的状态改为已读=1
	if invi.Status == data.InvitationStatusPending {
		invi.Status = data.InvitationStatusViewed
		err = invi.UpdateStatus()
		if err != nil {
			util.Debug(s_u.Email, " Cannot update invitation", err)
			report(w, r, "你好，茶博士正在努力的更新邀请函状态，请稍后再试。")
			return
		}
		// 减去茶友1小黑板新消息数
		if err = data.SubtractUserMessageCount(s_u.Id); err != nil {
			util.Debug(" Cannot subtract user message count", err)
			return
		}

	}

	//填写页面资料
	iD.SessUser = s_u
	iD.InvitationBean = i_b

	//向茶友返回该邀请函的详细信息
	generateHTML(w, &iD, "layout", "navbar.private", "member.invitation_read")
}

// GET /v1/team/new_applications/check?uuid=
// 团队管理员，处理本茶团的全部新的加盟申请书列表
func TeamNewApplicationsCheck(w http.ResponseWriter, r *http.Request) {
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
	//检查用户是否茶团核心成员，非核心成员不能查看新的加盟申请书

	if !canManageTeam(&t_team, s_u.Id, w, r) {
		report(w, r, "你好，茶博士竟然说，不是这个茶团核心成员不能查看，请确认。")
		return
	}

	//查询当前茶团未处理的加盟申请书，包含已查看但未处理的
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
	generateHTML(w, &mAL, "layout", "navbar.private", "team.applications", "component_member_application_bean")

}

// GET /v1/team_members/fired
// 显示被开除的成员的表单页面
func MemberFired(w http.ResponseWriter, r *http.Request) {
	report(w, r, "您好，茶博士正在忙碌建设这个功能中。。。")
}

// GET /v1/team_member/resigned?uuid=
// 团队最高管理员查看本茶团全部退出声明，支持分页
func TeamMemberResigned(w http.ResponseWriter, r *http.Request) {
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

	vals := r.URL.Query()
	team_uuid := vals.Get("uuid")
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Debug(team_uuid, "Cannot get team by uuid", err)
		report(w, r, "你好，茶博士失魂鱼，未能找到这个茶团，请稍后再试。")
		return
	}

	if !canManageTeam(&team, s_u.Id, w, r) {
		return
	}

	page := 1
	if pageStr := vals.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	resignations, err := data.GetResignationsByTeamId(team.Id)
	if err != nil {
		util.Debug(team.Id, "Cannot get resignations by team id", err)
		report(w, r, "你好，茶博士正在努力的查找退出声明，请稍后再试。")
		return
	}

	pageSize := 12
	totalCount := len(resignations)
	totalPages := (totalCount + pageSize - 1) / pageSize
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	var pageResignations []data.TeamMemberResignation
	if totalCount > 0 {
		pageResignations = resignations[start:end]
	}

	type PageData struct {
		SessUser         data.User
		Team             data.Team
		ResignationSlice []data.TeamMemberResignation
		CurrentPage      int
		TotalPages       int
		HasPrev          bool
		HasNext          bool
	}

	pageData := PageData{
		SessUser:         s_u,
		Team:             team,
		ResignationSlice: pageResignations,
		CurrentPage:      page,
		TotalPages:       totalPages,
		HasPrev:          page > 1,
		HasNext:          page < totalPages,
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "team.resignations")
}

// /v1/team_member/resignation/detail
// 团队管理员查看和处理退出声明
func TeamMemberResignationDetail(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		TeamMemberResignationDetailGet(w, r)
	case http.MethodPost:
		TeamMemberResignationProcess(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/team_member/resignation/detail?uuid=
// 团队管理员查看退出声明详情
func TeamMemberResignationDetailGet(w http.ResponseWriter, r *http.Request) {
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

	vals := r.URL.Query()
	resignation_uuid := vals.Get("uuid")
	resignation := data.TeamMemberResignation{Uuid: resignation_uuid}
	if err = resignation.GetByUuid(); err != nil {
		util.Debug(resignation_uuid, "Cannot get resignation by uuid", err)
		report(w, r, "你好，茶博士正在努力的查找退出声明，请稍后再试。")
		return
	}

	team, err := data.GetTeam(resignation.TeamId)
	if err != nil {
		util.Debug(resignation.TeamId, "Cannot get team by id", err)
		report(w, r, "你好，茶博士正在努力的查找退出声明资料，请稍后再试。")
		return
	}

	if !canManageTeam(&team, s_u.Id, w, r) {
		return
	}

	member, err := data.GetUser(resignation.MemberUserId)
	if err != nil {
		util.Debug(resignation.MemberUserId, "Cannot get user by id", err)
		report(w, r, "你好，茶博士正在努力的查找退出声明资料，请稍后再试。")
		return
	}

	// 检查团队成员数量
	memberCount := team.NumMembers()

	// 获取CEO
	ceoMember, err := team.MemberCEO()
	if err != nil {
		util.Debug("Cannot get CEO", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 获取核心成员
	coreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug("Cannot get core members", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 判断是否需要双重审批
	needDoubleApproval := memberCount >= 3 && len(coreMembers) > 1

	// 判断当前用户角色
	isCEO := ceoMember.UserId == s_u.Id
	isCoreMember := false
	for _, cm := range coreMembers {
		if cm.UserId == s_u.Id && cm.Role != data.RoleCEO {
			isCoreMember = true
			break
		}
	}

	type PageData struct {
		SessUser           data.User
		Team               data.Team
		Resignation        data.TeamMemberResignation
		Member             data.User
		NeedDoubleApproval bool
		IsCEO              bool
		IsCoreMember       bool
	}

	pageData := PageData{
		SessUser:           s_u,
		Team:               team,
		Resignation:        resignation,
		Member:             member,
		NeedDoubleApproval: needDoubleApproval,
		IsCEO:              isCEO,
		IsCoreMember:       isCoreMember,
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "team.resignation_detail")
}

// GET /v1/invitations/member
// 某个成员，查看自己收到的全部茶团邀请函列表
func InvitationsReceived(w http.ResponseWriter, r *http.Request) {
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

	// 查询用户收到的全部邀请函
	invitations, err := s_u.Invitations()
	if err != nil {
		util.Debug("Cannot get invitations by user email", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 获取邀请函Bean列表
	invitationBeans := make([]data.InvitationBean, 0)
	for _, inv := range invitations {
		bean, err := fetchInvitationBean(inv)
		if err != nil {
			util.Debug("Cannot fetch invitation bean", err)
			continue
		}
		invitationBeans = append(invitationBeans, bean)
	}

	type PageData struct {
		SessUser        data.User
		InvitationBeans []data.InvitationBean
		IsEmpty         bool
	}

	pageData := PageData{
		SessUser:        s_u,
		InvitationBeans: invitationBeans,
		IsEmpty:         len(invitationBeans) == 0,
	}

	// 渲染页面
	generateHTML(w, &pageData, "layout", "navbar.private", "invitations.member")
}

// POST /v1/team_member/resignation/detail
// 团队管理员处理退出声明（双重审批：核心成员同意 + CEO批准）
func TeamMemberResignationProcess(w http.ResponseWriter, r *http.Request) {
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

	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	resignation_uuid := r.PostFormValue("resignation_uuid")
	resignation := data.TeamMemberResignation{Uuid: resignation_uuid}
	if err = resignation.GetByUuid(); err != nil {
		util.Debug(resignation_uuid, "Cannot get resignation by uuid", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	team, err := data.GetTeam(resignation.TeamId)
	if err != nil {
		util.Debug(resignation.TeamId, "Cannot get team by id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 检查团队成员数量
	memberCount := team.NumMembers()

	// 获取CEO
	ceoMember, err := team.MemberCEO()
	if err != nil {
		util.Debug("Cannot get CEO", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 获取核心成员
	coreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug("Cannot get core members", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 判断是否需要双重审批：团队成员>=3 且 有CEO以外的核心成员
	needDoubleApproval := memberCount >= 3 && len(coreMembers) > 1

	action := r.PostFormValue("action")

	if needDoubleApproval {
		// 需要双重审批
		switch action {
		case "core_agree":
			// 核心成员同意
			if resignation.Status >= data.ResignationStatusCoreMemberAgree {
				report(w, r, "你好，该声明已经处理过了。")
				return
			}
			// 检查是否为核心成员（非CEO）
			isCoreMember := false
			for _, cm := range coreMembers {
				if cm.UserId == s_u.Id && cm.Role != data.RoleCEO {
					isCoreMember = true
					break
				}
			}
			if !isCoreMember {
				report(w, r, "你好，只有核心成员（非CEO）才能同意退出声明。")
				return
			}
			resignation.Status = data.ResignationStatusCoreMemberAgree
			resignation.CoreMemberUserId = s_u.Id
			if err = resignation.UpdateCeoUserIdCoreMemberUserIdStatus(); err != nil {
				util.Debug("Cannot update resignation status", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			report(w, r, "你好，已同意该成员退出，等待CEO批准。")

		case "ceo_approve":
			// CEO批准
			if ceoMember.UserId != s_u.Id {
				report(w, r, "你好，只有CEO才能批准退出声明。")
				return
			}
			if resignation.Status != data.ResignationStatusCoreMemberAgree {
				report(w, r, "你好，需要核心成员先同意后，CEO才能批准。")
				return
			}
			resignation.Status = data.ResignationStatusApproved
			resignation.CeoUserId = s_u.Id
			if err = resignation.UpdateCeoUserIdCoreMemberUserIdStatus(); err != nil {
				util.Debug("Cannot update resignation status", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			// 更新成员状态为已退出
			member, err := data.GetMemberByTeamIdUserId(team.Id, resignation.MemberUserId)
			if err != nil {
				util.Debug("Cannot get team member", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			member.Status = data.TeMemberStatusResigned
			if err = member.UpdateRoleStatus(); err != nil {
				util.Debug("Cannot update member status", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			report(w, r, "你好，CEO已批准该成员退出茶团。")

		case "retain":
			// 挡留
			if !canManageTeam(&team, s_u.Id, w, r) {
				return
			}
			if resignation.Status >= data.ResignationStatusApproved {
				report(w, r, "你好，该声明已经批准，无法挡留。")
				return
			}
			resignation.Status = data.ResignationStatusPending
			if err = resignation.UpdateCeoUserIdCoreMemberUserIdStatus(); err != nil {
				util.Debug("Cannot update resignation status", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			report(w, r, "你好，已标记为挡留中，请与成员沟通。")

		default:
			report(w, r, "你好，无效的操作。")
		}
	} else {
		// 不需要双重审批，管理员直接处理
		if !canManageTeam(&team, s_u.Id, w, r) {
			return
		}
		if resignation.Status >= data.ResignationStatusApproved {
			report(w, r, "你好，该退出声明已经处理过了。")
			return
		}

		switch action {
		case "approve", "core_agree", "ceo_approve":
			// 批准退出
			resignation.Status = data.ResignationStatusApproved
			resignation.CeoUserId = s_u.Id
			if err = resignation.UpdateCeoUserIdCoreMemberUserIdStatus(); err != nil {
				util.Debug("Cannot update resignation status", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			// 更新成员状态为已退出
			member, err := data.GetMemberByTeamIdUserId(team.Id, resignation.MemberUserId)
			if err != nil {
				util.Debug("Cannot get team member", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			member.Status = data.TeMemberStatusResigned
			if err = member.UpdateRoleStatus(); err != nil {
				util.Debug("Cannot update member status", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			report(w, r, "你好，已批准该成员退出茶团。")

		case "retain":
			// 挡留
			resignation.Status = data.ResignationStatusPending
			if err = resignation.UpdateCeoUserIdCoreMemberUserIdStatus(); err != nil {
				util.Debug("Cannot update resignation status", err)
				report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
			report(w, r, "你好，已标记为挡留中，请与成员沟通。")

		default:
			report(w, r, "你好，无效的操作。")
		}
	}
}

// GET /v1/resignations/member
// 某个成员，查看自己的全部退出茶团声明列表
func ResignationsReceived(w http.ResponseWriter, r *http.Request) {
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

	// 查询用户的全部退出声明
	resignations, err := data.GetResignationsByUserId(s_u.Id)
	if err != nil {
		util.Debug("Cannot get resignations by user id", err)
		report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	// 获取每个退出声明对应的茶团信息
	type ResignationBean struct {
		Resignation data.TeamMemberResignation
		Team        data.Team
	}

	resignationBeans := make([]ResignationBean, 0)
	for _, res := range resignations {
		team, err := data.GetTeam(res.TeamId)
		if err != nil {
			util.Debug("Cannot get team by id", err)
			continue
		}
		resignationBeans = append(resignationBeans, ResignationBean{
			Resignation: res,
			Team:        team,
		})
	}

	type PageData struct {
		SessUser         data.User
		ResignationBeans []ResignationBean
		IsEmpty          bool
	}

	pageData := PageData{
		SessUser:         s_u,
		ResignationBeans: resignationBeans,
		IsEmpty:          len(resignationBeans) == 0,
	}

	// 渲染页面
	generateHTML(w, &pageData, "layout", "navbar.private", "resignations.member")
}

// GET /v1/applications/member
// 某个成员，查看自己全部茶团加盟申请书列表
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
	generateHTML(w, &mAL, "layout", "navbar.private", "applications.team_member", "component_member_application_bean")

}
