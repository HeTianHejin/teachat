package route

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
)

// HandleMemberResign() /v1/team_member/resign
// 处理某个茶团的某个成员退出茶团声明撰写和提交
func HandleMemberResign(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		MemberResign(w, r)
	case "POST":
		MemberResignReply(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// POST /v1/team_member/resign
// 处理成员提交“退出茶团声明”事务
func MemberResignReply(w http.ResponseWriter, r *http.Request) {
	// 获取session
	s, err := Session(r)
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 解析表单内容，获取当前用户提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Danger(util.LogError(err), " Cannot parse form")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 提交的声明标题
	titl := r.PostFormValue("title")
	// 检查提交的声明标题字数是否>3 and <32
	lenTit := CnStrLen(titl)
	if lenTit < 3 || lenTit > 32 {
		Report(w, r, "你好，茶博士认为标题字数太长或者太短，请确认后再试。")
		return
	}

	// 提交的声明内容
	cont := r.PostFormValue("content")
	// 检查提交的声明内容字数是否>3 and <456
	lenCont := CnStrLen(cont)
	if lenCont < 3 || lenCont > 456 {
		Report(w, r, "你好，茶博士认为内容字数太长或者太短，请确认后再试。")
		return
	}
	// 检查提交的成员邮箱
	m_email := r.PostFormValue("m_email")
	if ok := IsEmail(m_email); !ok {
		Report(w, r, "你好，涨红了脸的茶博士，竟然强词夺理说，电子邮箱格式太复杂看不懂，请确认后再提交。")
		return
	}
	//读取声明退出的成员资料
	t_user, err := data.GetUserByEmail(m_email)
	if err != nil {
		util.Warning(util.LogError(err), m_email, "Cannot get user by email")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查提交人是否和会话用户一致
	if s_u.Id != t_user.Id {
		Report(w, r, "你好，目前不能代替他人提交退出茶团声明，请确认后再试。")
		return
	}

	// 提交的目标茶团
	team_id_str := r.PostFormValue("team_id")
	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		util.Danger(util.LogError(err), team_id_str, "Cannot convert team_id to int")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 读取目标茶团资料
	t_team, err := data.GetTeamById(team_id)
	if err != nil {
		util.Danger(util.LogError(err), team_id, "Cannot get team by id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 检查提交人是否为茶团成员
	t_member, err := data.GetMemberByTeamIdUserId(t_team.Id, t_user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			Report(w, r, "你好，茶博士嘀咕，你不是茶团成员，不接受退出声明噢。")
			return
		} else {
			util.Warning(util.LogError(err), t_team.Id, t_user.Id, "Cannot get team member by team id and user id")
			Report(w, r, "你好，茶博士失魂鱼，未能获取拟退出的茶团资料，请稍后再试。")
			return
		}
	}

	//查看成员角色，分类处理：1、CEO，2、核心成员：CTO、CFO、CMO，3、普通成员：taster
	switch t_member.Role {
	case "taster":
		break
	case "CTO", "CFO", "CMO":
		Report(w, r, "你好，请先联系CEO，将你目前角色核心成员调整为普通成员品茶师，然后再声明退出。")
		return
	case "CEO":
		Report(w, r, "你好，请先联系茶团创建人，将你目前角色调整为品茶师，然后再声明退出。")
		return
	default:
		Report(w, r, "你好，满头大汗的茶博士表示找不到这个茶友角色，请确认后再试。")
		return
	}

	//声明一份茶团成员退出声明书
	tmqD := data.TeamMemberResignation{
		Uuid:              data.Random_UUID(),
		TeamId:            t_team.Id,
		CeoUserId:         0,
		CoreMemberUserId:  0,
		MemberId:          t_member.Id,
		MemberUserId:      t_user.Id,
		MemberCurrentRole: t_member.Role,
		Title:             titl,
		Content:           cont,
		Status:            0,
		CreatedAt:         time.Now(),
	}

	//尝试保存退出声明
	if err := tmqD.Create(); err != nil {
		util.Warning(util.LogError(err), "Cannot create team member resignation")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//返回成功保存声明的报告
	Report(w, r, "你好，茶博士已经收到了你的退出声明，我们会尽快处理。")

	//返回茶团主页
	//http.Redirect(w, r, fmt.Sprintf("/v1/team?id=%s", t_team.Uuid), http.StatusFound)

}

// MemberResign() GET /v1/team_member/resign?id=XXX
// 取出一张空白茶团成员“退出茶团声明”撰写页面
func MemberResign(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取提交的查询参数
	vals := r.URL.Query()
	team_uuid := vals.Get("id")

	//读取目标茶团资料
	t_team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Warning(util.LogError(err), team_uuid, "Cannot get team by uuid")
		Report(w, r, "你好，满头大汗的茶博士表示找不到这个茶团，稍后再试。")
		return
	}

	//检查目标茶友是否茶团成员
	_, err = data.GetMemberByTeamIdUserId(t_team.Id, s_u.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			Report(w, r, "你好，满头大汗的茶博士表示这个不是茶团的成员，稍后再试。")
			return
		} else {
			util.Warning(util.LogError(err), t_team.Id, " when GetMemberByTeamIdAndUserId() checking team_member")
			Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	}

	var tmqPD data.TeamMemberResign
	tmqPD.SessUser = s_u
	tmqPD.Team = t_team

	//渲染退出声明撰写页面
	RenderHTML(w, &tmqPD, "layout", "navbar.private", "member.resign_new")

}

// GET /v1/team_member/role_changed?id=XXX
// MemberRoleChanged() 某个茶团的全部已发布成员角色调整声明列表页面
func MemberRoleChanged(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取查询参数
	vals := r.URL.Query()
	team_uuid := vals.Get("id")

	//读取目标茶团资料
	t_team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Warning(util.LogError(err), team_uuid, "Cannot get team by uuid")
		Report(w, r, "你好，满头大汗的茶博士表示找不到这个茶团，稍后再试。")
		return
	}

	//读取这支茶团已发布的，全部成员角色调整声明
	role_notices, err := data.GetMemberRoleNoticesByTeamId(t_team.Id)
	if err != nil {
		util.Warning(util.LogError(err), t_team.Id, "Cannot get team member role change notice by team id")
		Report(w, r, "你好，茶博士正在忙碌中，请稍后再试。")
		return
	}
	tmrnBeanList, err := FetchTeamMemberRoleNoticeBeanList(role_notices)
	if err != nil {
		util.Warning(util.LogError(err), "Cannot fetch team member role notice bean list")
		Report(w, r, "你好，茶博士正在忙碌中，请稍后再试。")
		return
	}

	var tmrcnpd data.TeamMemberRoleChangedNoticesPageData
	tmrcnpd.SessUser = s_u
	tmrcnpd.Team = t_team
	tmrcnpd.TeamMemberRoleNoticeBeanList = tmrnBeanList

	//渲染茶团成员角色调整通知页面
	RenderHTML(w, &tmrcnpd, "layout", "navbar.private", "member.role_changed_notices")

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

// MemberRoleChange() GET /v1/team_member/role?id=XXX&m_email=XXX
// 取出一张空白茶团成员角色任命书
func MemberRoleChange(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取提交的查询参数
	vals := r.URL.Query()
	team_id_str := vals.Get("id")
	//检查提交的id参数格式是否正常
	team_id_int, err := strconv.Atoi(team_id_str)
	if err != nil {
		util.Warning(util.LogError(err), team_id_str, "Cannot convert team_id to int")
		Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
		return
	}

	member_email := vals.Get("m_email")

	//检查提交的email参数格式是否正常
	if ok := IsEmail(member_email); !ok {
		Report(w, r, "你好，茶博士的眼镜被闪电破坏了，看不清提及的邮箱，请稍后再试。")
		return
	}

	//读取目标茶团资料
	t_team, err := data.GetTeamById(team_id_int)
	if err != nil {
		util.Warning(util.LogError(err), team_id_str, "Cannot get team by id")
		Report(w, r, "你好，满头大汗的茶博士表示找不到这个茶团，稍后再试。")
		return
	}
	//读取拟调整角色目标茶友资料
	t_member, err := data.GetUserByEmail(member_email)
	if err != nil {
		util.Warning(util.LogError(err), member_email, "Cannot get user given email")
		Report(w, r, "你好，满头大汗的茶博士表示找不到这个茶友，稍后再试。")
		return
	}

	//检查目标茶友是否茶团成员
	_, err = data.GetMemberByTeamIdUserId(t_team.Id, t_member.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			Report(w, r, "你好，满头大汗的茶博士表示这个不是茶团的成员，稍后再试。")
			return
		} else {
			util.Warning(util.LogError(err), t_team.Id, " when GetMemberByTeamIdAndUserId() checking team_member")
			Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	}

	//检查提交者（当前用户）是否茶团CEO
	//先读取茶团CEO成员资料
	member_ceo, err := t_team.MemberCEO()
	if err != nil {
		util.Warning(util.LogError(err), t_team.Id, "Cannot get team ceo given team id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取目标茶团创建人的资料，Founder也可以调整成员角色！包括CEO！太疯狂了？（参考自观音菩萨可以决定西天取经团队的任何角色人选）
	t_founder, err := t_team.Founder()
	if err != nil {
		util.Warning(util.LogError(err), t_team.Id, "Cannot get team founder given team id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
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
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你不是这个茶团的管理者，无权调整角色噢。")
		return
	}

	//读取CEO茶友资料
	t_ceo, err := data.GetUserById(member_ceo.UserId)
	if err != nil {
		util.Warning(util.LogError(err), member_ceo.UserId, "Cannot get user by id")
		Report(w, r, "你好，茶博士正在忙碌中，请稍后再试。")
		return
	}

	m_c_role, err := data.GetTeamMemberRoleByTeamIdAndUserId(t_team.Id, t_member.Id)
	if err != nil {
		util.Warning(util.LogError(err), t_team.Id, "Cannot get team member role given team id")
		Report(w, r, "你好，茶博士正在忙碌中，请稍后再试。")
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
	RenderHTML(w, &tmrcnP, "layout", "navbar.private", "member.role_change_new")
}

// POST /v1/team_member/role
// 提交一个成员新的团队角色任命书答复
func MemberRoleReply(w http.ResponseWriter, r *http.Request) {
	//获取session
	s, err := Session(r)
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//解析表单内容，获取当前用户提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Danger(util.LogError(err), " Cannot parse form")
		http.Redirect(w, r, "/v1/", http.StatusFound)
		return
	}
	//提交的成员邮箱
	m_email := r.PostFormValue("m_email")
	if ok := IsEmail(m_email); !ok {
		Report(w, r, "你好，涨红了脸的茶博士，竟然强词夺理说，电子邮箱格式太复杂看不懂，请确认后再提交。")
		return
	}
	//提交的目标茶团
	team_id_str := r.PostFormValue("team_id")
	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		util.Danger(util.LogError(err), team_id_str, "Cannot convert team_id to int")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取目标茶团资料
	t_team, err := data.GetTeamById(team_id)
	if err != nil {
		util.Danger(util.LogError(err), team_id, "Cannot get team by id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//提交的成员新角色参数
	new_role := r.PostFormValue("role")
	//提交的成员新角色参数是否正常
	switch new_role {
	case "taster":
		break
	case "CEO", "CTO", "CFO", "CMO":
		//需要检查目标角色是否空缺
		_, err = t_team.GetTeamMemberByRole(new_role)
		if err != nil {
			if err == sql.ErrNoRows {
				//目标角色空缺,可以调整
				break
			} else {
				util.Danger(util.LogError(err), t_team.Id, new_role, "Cannot get team member by role")
				Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
		} else {
			Report(w, r, "你好，茶博士摸摸头嘀咕说，你提交的角色已经有人担任了，请确认后再提交。")
			return
		}

	default:
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你提交的角色不在茶团角色列表中，请确认后再提交。")
		return
	}

	//目标茶友
	t_member, err := data.GetUserByEmail(m_email)
	if err != nil {
		util.Danger(util.LogError(err), m_email, "Cannot get user by email")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查目标茶友是否茶团成员
	member, err := data.GetMemberByTeamIdUserId(t_team.Id, t_member.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			Report(w, r, "你好，茶博士摸摸头嘀咕说，这个茶友不是茶团成员，无法调整角色。")
			return
		} else {
			util.Danger(util.LogError(err), t_team.Id, " when GetMemberByTeamIdAndUserId() checking team_member")
			Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	}
	//检查提交者是否在尝试调整自己的角色，不合规
	if member.UserId == s_u.Id {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你不能调整自己的角色。")
		return
	}
	//Role no change
	if new_role == member.Role {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你没有调整角色，无需提交。")
		return
	}

	//提交的角色调整标题
	title := r.PostFormValue("title")
	//检查提交的标题是否正常，中文字数>6,<24
	if CnStrLen(title) < 6 || CnStrLen(title) > 24 {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你提交的标题太长或太短，请确认后再提交。")
		return
	}
	//提交的角色调整内容
	content := r.PostFormValue("content")
	//检查提交的内容是否正常，中文字数>2,<456,
	if CnStrLen(content) < 2 || CnStrLen(content) > 456 {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你提交的内容太长或太短，请确认后再提交。")
		return
	}

	//检查提交者（当前用户）是否茶团CEO？如果不是CEO，再检查是否是茶团创建人
	//读取CEO成员资料
	m_ceo, err := t_team.MemberCEO()
	if err != nil {
		util.Danger(util.LogError(err), t_team.Id, "Cannot get team ceo given team id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取目标茶团创建人资料
	t_founder, err := t_team.Founder()
	if err != nil {
		util.Danger(util.LogError(err), t_team.Id, "Cannot get team founder given team id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	if m_ceo.UserId == s_u.Id && new_role != "CEO" {
		//会话用户是CEO，可以调整非CEO成员角色
		//创建一个新的团队成员角色变动公告
		team_member_role_notice := data.TeamMemberRoleNotice{
			Uuid:              data.Random_UUID(),
			TeamId:            t_team.Id,
			CeoId:             m_ceo.UserId,
			MemberId:          member.Id,
			MemberCurrentRole: member.Role,
			NewRole:           new_role,
			Title:             title,
			Content:           content,
			Status:            0,
			CreatedAt:         time.Now(),
		}
		if err = team_member_role_notice.Create(); err != nil {
			util.Warning(util.LogError(err), team_member_role_notice, "Cannot create team_member_role_notice")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		//更新成员角色
		member.Role = new_role
		if err = member.UpdateRoleClass(); err != nil {
			util.Warning(util.LogError(err), member, "Cannot update member")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}

	} else if t_founder.Id == s_u.Id {
		//会话用户是创建人，可以调整CEO和非CEO成员角色
		//创建一个新的团队成员角色变动公告
		team_member_role_notice := data.TeamMemberRoleNotice{
			Uuid:              data.Random_UUID(),
			TeamId:            t_team.Id,
			CeoId:             m_ceo.UserId,
			MemberId:          member.Id,
			MemberCurrentRole: member.Role,
			NewRole:           new_role,
			Title:             title,
			Content:           content,
			Status:            0,
			CreatedAt:         time.Now(),
		}
		if err = team_member_role_notice.Create(); err != nil {
			util.Warning(util.LogError(err), team_member_role_notice, "Cannot create team_member_role_notice")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		// 更新成员角色
		member.Role = new_role
		if err = member.UpdateRoleClass(); err != nil {
			util.Warning(util.LogError(err), member, "Cannot update member")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
	} else {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你不是这个茶团的管理者，无权调整角色噢。")
		return
	}

	//报告调整角色成功消息
	Report(w, r, "你好，茶博士摸摸头说，已经调整了 "+t_member.Name+" 的角色为 "+new_role+" 。")
}

// /v1/team_member/invite
// 邀请一个指定的新茶友加入封闭式茶团
func HandleInviteMember(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//返回邀请团队新成员，即邀请函填写页面
		InviteMemberNew(w, r)
	case "POST":
		//生成邀请函方法
		InviteMemberReply(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// /v1/team_member/invitation
// 处理茶团邀请新成员函
func HandleMemberInvitation(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//返回邀请函详情页面
		MemberInvitationRead(w, r)
	case "POST":
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
	case "GET":
		NewMemberApplicationForm(w, r)
	case "POST":

		NewMemberApplication(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// /v1/team_member/application/review
// 审查，处理茶团加盟申请书
func HandleMemberApplication(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		MemberApplicationReview(w, r)
	case "POST":
		MemberApplicationReply(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// POST /v1/team_member/application/review
// 接受 加盟茶团申请书审查人，提交处理（决定）结果
func MemberApplicationReply(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//解析表单内容，获取茶友提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Danger(util.LogError(err), " Cannot parse form")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	//读取提交的参数
	application_id_str := r.PostFormValue("application_id")
	application_id, err := strconv.Atoi(application_id_str)
	if err != nil {
		util.Danger(util.LogError(err), application_id_str, "Cannot convert application_id to int")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	application := data.MemberApplication{
		Id: application_id,
	}
	//读取加盟茶团申请书
	if err = application.Get(); err != nil {
		util.Danger(util.LogError(err), application_id, "Cannot get application given id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查申请书的状态是否正常，已查看
	switch application.Status {
	case 0:
		//未查看
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	case 1:
		//已查看，未处理
		break
	case 2, 3:
		// 已经处理过了
		Report(w, r, "你好，这份申请书已经被处理，请确认后再试。")
		return
	case 4:
		//已经过期或者失效
		Report(w, r, "你好，这份申请书已经过期或者失效，请确认后再试。")
		return
	default:
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
	}

	//读取申请人申请加盟的茶团
	team, err := data.GetTeamById(application.TeamId)
	if err != nil {
		util.Danger(util.LogError(err), application.TeamId, "Cannot get team given id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//检查s_u是否茶团的核心成员，非核心成员不能审核申请书
	core_members, err := team.CoreMembers()
	if err != nil {
		util.Danger(util.LogError(err), team.Id, "Cannot get core members of team")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 检查当前茶友是否是茶团的核心成员
	is_core_member := false
	for _, core_member := range core_members {
		if core_member.UserId == s_u.Id {
			is_core_member = true
			break
		}
	}
	// 如果不是茶团的核心成员，返回错误
	if !is_core_member {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你不是茶团的核心成员，无权处理申请书噢。")
		return
	}

	//读取申请人资料
	applicant, err := data.GetUserById(application.UserId)
	if err != nil {
		util.Danger(util.LogError(err), application.UserId, "Cannot get user given id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查申请人是否已经是茶团成员=这个茶团是否已经存在该茶友
	_, err = data.GetMemberByTeamIdUserId(team.Id, applicant.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Danger(util.LogError(err), applicant.Email, " when checking team_member")
			Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经在茶团中了噢？请确认后再试。")
		return
	}
	//检查当前会话茶友是否和审查足迹资料中审查人一致

	//读取提交的回复内容
	reply := r.PostFormValue("reply")
	// if reply == "" {
	// 	Report(w, r, "你好，茶博士摸摸头嘀咕说，你没有填写回复内容或者内容太复杂了。")
	// 	return
	// }

	//读取提交的审查结果参数
	approval_str := r.PostFormValue("approval")
	approval_int, err := strconv.Atoi(approval_str)
	if err != nil {
		util.Danger(util.LogError(err), approval_str, "Cannot convert approval to int")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//如果approval_int = 1，是批准加盟；如果 = 0 ，是婉拒
	switch approval_int {
	case 1:
		//批准加盟
		//创建一个新的茶团成员
		team_member := data.TeamMember{
			Uuid:      data.Random_UUID(),
			TeamId:    team.Id,
			UserId:    applicant.Id,
			Role:      "taster",
			CreatedAt: time.Now(),
			Class:     1,
			UpdatedAt: time.Now(),
		}
		//将新的茶团成员写入数据库
		if err = team_member.Create(); err != nil {
			util.Danger(util.LogError(err), team_member, "Cannot create team_member")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		//检查茶友加入茶团计数，如果是1，从默认的自由人，改设当前茶团为默认茶团
		count, err := applicant.SurvivalTeamsCount()
		if err != nil {
			util.Danger(util.LogError(err), applicant.Email, " Cannot get survival teams count")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		if count == 0 {
			user_default_team := data.UserDefaultTeam{
				UserId:    applicant.Id,
				TeamId:    team.Id,
				CreatedAt: time.Now(),
			}
			if err = user_default_team.Create(); err != nil {
				util.Danger(util.LogError(err), applicant.Email, " Cannot create user_default_team")
				Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
		}

		//更新申请书的状态为已批准
		application.Status = 2
		if err = application.Update(); err != nil {
			util.Danger(util.LogError(err), application, "Cannot update application")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		//创建批准加盟申请书答复
		application_reply := data.MemberApplicationReply{
			MemberApplicationId: application.Id,
			TeamId:              team.Id,
			UserId:              s_u.Id,
			ReplyContent:        reply,
			Status:              2,
			CreatedAt:           time.Now(),
		}
		if err = application_reply.Create(); err != nil {
			util.Danger(util.LogError(err), application_reply, "Cannot create application_reply")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}

		//报告批准加盟成功消息
		Report(w, r, "你好，已经批准新成员 "+applicant.Name+" 加盟 "+team.Name+" 茶团。")
		return
	case 0:
		//婉拒加盟
		application.Status = 3
		if err = application.Update(); err != nil {
			util.Danger(util.LogError(err), application, "Cannot update application")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		//创建婉拒加盟申请书答复
		application_reply := data.MemberApplicationReply{
			MemberApplicationId: application.Id,
			TeamId:              team.Id,
			UserId:              s_u.Id,
			ReplyContent:        reply,
			Status:              3,
			CreatedAt:           time.Now(),
		}
		if err = application_reply.Create(); err != nil {
			util.Danger(util.LogError(err), application_reply, "Cannot create application_reply")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		Report(w, r, "你好，茶团成员 "+applicant.Email+" 已经婉拒加盟茶团 "+team.Name+"。")
		return
	default:
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

}

// GET /v1/team_member/application/review?id=xxx
// 打开新加盟申请书，审查其内容
func MemberApplicationReview(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 读取提交的查询参数
	vals := r.URL.Query()
	application_uuid := vals.Get("id")
	application := data.MemberApplication{
		Uuid: application_uuid,
	}
	// 读取申请书
	if err = application.GetByUuid(); err != nil {
		util.Danger(util.LogError(err), application_uuid, "Cannot get application given uuid")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取申请人资料
	applicant, err := data.GetUserById(application.UserId)
	if err != nil {
		util.Danger(util.LogError(err), application.UserId, "Cannot get user given id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//读取申请人申请加盟的茶团
	team, err := data.GetTeamById(application.TeamId)
	if err != nil {
		util.Danger(util.LogError(err), application.TeamId, "Cannot get team given id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 检查这个茶团是否已经存在该茶友
	_, err = data.GetMemberByTeamIdUserId(team.Id, applicant.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Danger(util.LogError(err), applicant.Email, " when checking team_member")
			Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		Report(w, r, "你好，这个申请人已经在茶团中了噢？请确认后再试。")
		return
	}
	// 读取申请人默认所在的茶团
	applicant_default_team, err := applicant.GetLastDefaultTeam()
	if err != nil {
		util.Danger(util.LogError(err), applicant.Email, "Cannot get default team given user")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//检查当前会话茶友身份，是否team的管理成员（核心成员）
	core_members, err := team.CoreMembers()
	if err != nil {
		util.Danger(util.LogError(err), team.Id, "Cannot get core members of team")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
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
		Report(w, r, "你好，茶博士摸摸头嘀咕说，你不是茶团的核心成员，无权审查申请书噢。")
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
	footprint := data.Footprint{
		UserId:    s_u.Id,
		TeamId:    team.Id,
		TeamName:  team.Abbreviation,
		Content:   fmt.Sprintf("查看了茶团 %s 的加盟申请书", team.Name),
		ContentId: application.Id,
		CreatedAt: time.Now(),
	}
	if err = footprint.Create(); err != nil {
		util.Warning(util.LogError(err), "Cannot create footprint")
	}
	//修改申请书状态为已查看
	application.Status = 1
	if err = application.Update(); err != nil {
		util.Danger(util.LogError(err), application.Id, "Cannot update application status")
	}

	//渲染页面
	RenderHTML(w, &aR, "layout", "navbar.private", "team.application_review")
}

// POST /v1/team_member/application/new
// 递交 茶团加盟申请书，处理窗口
func NewMemberApplication(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//读取表单数据
	team_uuid := r.FormValue("team_uuid")
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Danger(util.LogError(err), team_uuid, "Cannot get team given uuid")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	user_uuid := r.FormValue("user_uuid")
	app_user, err := data.GetUserByUUID(user_uuid)
	if err != nil {
		util.Danger(util.LogError(err), user_uuid, "Cannot get user given uuid")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	content := r.FormValue("content")
	//check content length
	if len(content) > 666 || len(content) < 2 {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你的申请书内容太长了噢？墨水瓶也怕抄单词呀！")
		return
	}

	// s_u.Id != user.Id，检查是否茶友本人提交申请，不允许代他人申请
	if s_u.Id != app_user.Id {
		Report(w, r, "你好，身前有余忘缩手，眼前无路想回头，目前仅接受本人申请加入茶团噢。")
		return
	}

	//检查这个茶团是否已经存在该茶友
	_, err = data.GetMemberByTeamIdUserId(team.Id, app_user.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Danger(util.LogError(err), app_user.Email, " when checking team_member")
			Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经在茶团中了噢？请确认后再试。")
		return
	}

	//创建茶团新茶友加盟申请书
	ma := data.MemberApplication{
		TeamId:    team.Id,
		UserId:    app_user.Id,
		Content:   content,
		Status:    0,
		CreatedAt: time.Now(),
	}
	//保存申请书
	if err = ma.Create(); err != nil {
		util.Danger(util.LogError(err), team.Id, "Cannot create member-application")
		Report(w, r, "你好，闪电考拉正在飞速处理所有的技术问题，请耐心等待。")
		return
	}
	//发送邮件通知茶团管理员（等待茶团管理员上班查看茶团详情即可见申请书，不另外通知）

	//返回成功页面
	t := fmt.Sprintf("你好，%s ，加盟 %s 申请书已经提交，请等待茶团管理员的回复。", s_u.Name, team.Abbreviation)
	Report(w, r, t)

}

// GET /v1/team_member/application/new
// 返回 申请加入表单 页面
func NewMemberApplicationForm(w http.ResponseWriter, r *http.Request) {
	// 读取会话
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 读取提交的查询参数
	vals := r.URL.Query()
	team_uuid := vals.Get("id")
	// 读取茶团资料
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Danger(util.LogError(err), team_uuid, "Cannot get team given uuid")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//检测当前用户是否向指定茶团，已经提交过加盟申请？而且申请书状态为等待处理（Status<=1）
	_, err = data.CheckMemberApplicationByTeamIdAndUserId(team.Id, s_u.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Danger(util.LogError(err), s_u.Email, " when checking member_application")
			Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经提交过申请书了噢？请确认后再试。")
		return
	}

	//检查这个茶团是否已经存在该茶友
	_, err = data.GetMemberByTeamIdUserId(team.Id, s_u.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			break
		default:
			util.Danger(util.LogError(err), s_u.Email, " when checking team_member")
			Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
			return
		}
	} else {
		Report(w, r, "你好，茶博士摸摸头嘀咕说，茶友你已经在茶团中了噢？")
		return
	}

	var tD data.TeamDetail
	tD.SessUser = s_u
	tD.Team = team
	//渲染页面
	RenderHTML(w, &tD, "layout", "navbar.private", "member.application_new")

}

// POST /v1/team_member/invitation
// 邀请函处理（回复）方法
func MemberInvitationReply(w http.ResponseWriter, r *http.Request) {
	//解析表单内容，获取茶友提交的内容
	err := r.ParseForm()
	if err != nil {
		util.Danger(util.LogError(err), " Cannot parse form")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 根据会话信息读取茶友资料
	s, err := Session(r)
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), s.Email, "Cannot get user from session")
		Report(w, r, "你好，满地梨花一片天，请稍后再试一次")
		return
	}
	//读取提交的参数
	invitation_id, err := strconv.Atoi(r.PostFormValue("invitation_id"))
	if err != nil {
		util.Warning(util.LogError(err), invitation_id, "Failed to convert invitation_id to int")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	user_id, err := strconv.Atoi(r.PostFormValue("user_id"))
	if err != nil {
		util.Warning(util.LogError(err), user_id, "Failed to convert user_id to int")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	//检查是否存在该茶友注册资料
	reply_user, err := data.GetUserById(user_id)
	if err != nil {
		util.Danger(util.LogError(err), user_id, "Cannot get user given id")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}

	//检查一下提交的茶友和会话茶友Id是否一致
	if reply_user.Id != s_u.Id {
		util.Warning(util.LogError(err), s_u.Email, "Inconsistency between submitted user id and session id")
		Report(w, r, "你好，请先登录，稍后再试。")
		return
	}
	//根据茶友提交的invitation_id，检查是否存在该邀请函
	invitation, err := data.GetInvitationById(invitation_id)
	if err != nil {
		util.Danger(util.LogError(err), s_u.Email, " Cannot get invitation")
		Report(w, r, "你好，秋阴捧出何方雪？雨渍添来隔宿痕。稍后再试。")
		return
	}
	//检查一下邀请函是否已经被回复
	if invitation.Status > 1 {
		Report(w, r, "你好，这个邀请函已经答复或者已过期。")
		return
	}
	reply_class_int, err := strconv.Atoi(r.PostFormValue("reply"))
	if err != nil {
		util.Warning(util.LogError(err), "Failed to convert class to int")
		Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	reply_word := r.PostFormValue("invitation_reply")
	//检查一下茶友提交的string，即reply_word是否不为空，中文长度小于239字符之间

	if 1 > CnStrLen(reply_word) || CnStrLen(reply_word) > 239 {
		util.Warning(util.LogError(err), s_u.Email, " Cannot process invitation")
		Report(w, r, "你好，瞪大眼睛涨红了脸的茶博士，竟然强词夺理说，答复的话太长了或者太短，只有外星人才接受呀，请确认再试。")
		return
	}

	//读取目标茶团资料
	team, err := data.GetTeamById(invitation.TeamId)
	if err != nil {
		util.Danger(util.LogError(err), s_u.Email, " Cannot get team")
		Report(w, r, "你好，丢了眼镜的茶博士忙到现在，还没有找到茶团登记本，请稍后再试。")
		return
	}

	if reply_class_int == 1 {
		//接受加盟邀请!

		// 检查这个茶团是否已经存在该茶友了
		_, err = data.GetMemberByTeamIdUserId(team.Id, reply_user.Id)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				break
			default:
				util.Danger(util.LogError(err), s_u.Email, " when checking team_member")
				Report(w, r, "你好，茶博士的眼镜被闪电破坏了，请稍后再试。")
				return
			}
		} else {
			Report(w, r, "你好，茶博士摸摸头嘀咕说，这个茶友已经在茶团中了噢。")
			return
		}

		//检查受邀请的茶友团队核心角色是否已经被占用
		switch invitation.Role {
		case "CTO", "CMO", "CFO", "CEO":
			//检查teamMember.Role中是否已经存在
			_, err = team.GetTeamMemberByRole(invitation.Role)
			if err == nil {
				Report(w, r, "你好，该团队已经存在所选择的核心角色，请确认所选择的角色是否恰当。")
				return
			} else if err != sql.ErrNoRows {
				util.Danger(util.LogError(err), s_u.Email, " Cannot search team member given team_id and role")
				Report(w, r, "你好，茶博士在这个团队角色事情迷糊了，请确认后再试。")
				return
			}

		case "taster":
			// No additional validation needed for the "taster" role
			break
		default:
			Report(w, r, "你好，请选择正确的角色。")
			return
		}
		//接受邀请，则升级邀请函状态并保存答复话语和时间
		invitation.Status = 2
		invitation.Update()
		repl := data.InvitationReply{
			InvitationId: invitation_id,
			UserId:       user_id,
			ReplyWord:    reply_word,
			CreatedAt:    time.Now(),
		}
		if err = repl.Create(); err != nil {
			util.Warning(util.LogError(err), s_u.Email, " Cannot create invitation_reply")
			Report(w, r, "你好，晕头晕脑的茶博士竟然把邀请答复搞丢了，请稍后再试。")
			return
		}
		// 准备将新成员添加进茶团所需的资料
		team_member := data.TeamMember{
			Uuid:      data.Random_UUID(),
			TeamId:    invitation.TeamId,
			UserId:    reply_user.Id,
			Role:      invitation.Role,
			CreatedAt: time.Now(),
			Class:     1,
			UpdatedAt: time.Now(),
		}

		// 如果team_member.Role == "CEO",采取更换CEO方法
		if team_member.Role == "CEO" {
			if err = team_member.UpdateFirstCEO(s_u.Id); err != nil {
				util.Warning(util.LogError(err), s_u.Email, " Cannot update team_member")
				Report(w, r, "你好，幽情欲向嫦娥诉，无奈虚廊夜色昏。请稍后再试。")
				return
			}

		} else {
			// 其它角色
			if err = team_member.Create(); err != nil {
				util.Warning(util.LogError(err), s_u.Email, " Cannot create team_member")
				Report(w, r, "你好，晕头晕脑的茶博士竟然忘记登记新成员了，请稍后再试。")
				return
			}
		}

		//检查茶友加入茶团计数，如果是1，从默认的自由人，改设当前茶团为默认茶团
		count, err := reply_user.SurvivalTeamsCount()
		if err != nil {
			util.Danger(util.LogError(err), reply_user.Email, " Cannot get survival teams count")
			Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
			return
		}
		if count == 0 {
			user_default_team := data.UserDefaultTeam{
				UserId:    reply_user.Id,
				TeamId:    invitation.TeamId,
				CreatedAt: time.Now(),
			}
			if err = user_default_team.Create(); err != nil {
				util.Danger(util.LogError(err), reply_user.Email, " Cannot create user_default_team")
				Report(w, r, "你好，茶博士正在忙碌中，稍后再试。")
				return
			}
		}

		//返回此茶团页面给茶友，成员列表上有该茶友，示意已经加入该茶团成功
		http.Redirect(w, r, "/v1/team/detail?id="+(team.Uuid), http.StatusFound)
		return

	} else if reply_class_int == 0 {
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
			util.Warning(util.LogError(err), s_u.Email, " Cannot create invitation_reply")
			Report(w, r, "你好，晕头晕脑的茶博士竟然把邀请答复搞丢了，请稍后再试。")
			return
		}
		//报告用户已经保存拒绝该邀请到答复记录
		t := fmt.Sprintf("你好，茶博士已经保存关于 %s 婉拒加盟答复。", team.Abbreviation)
		Report(w, r, t)
		return

	} else {
		// 无效的reply 数值
		Report(w, r, "你好，何幸邀恩宠，宫车过往频。稍后再试。")
		return
	}
}

// POST /v1/team_member/invite
// 提交一封邀请函参数。处理邀请某个看中的茶友到teamId指定的团队事项
func InviteMemberReply(w http.ResponseWriter, r *http.Request) {
	//获取session
	s, err := Session(r)
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//解析表单内容，获取茶友提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Danger(util.LogError(err), " Cannot parse form")
		http.Redirect(w, r, "/v1/", http.StatusFound)
		return
	}
	email := r.PostFormValue("email")
	if ok := IsEmail(email); !ok {
		Report(w, r, "你好，涨红了脸的茶博士，竟然强词夺理说，电子邮箱格式太复杂看不懂，请确认后再提交。")
		return
	}

	i_word := r.PostFormValue("invite_word")
	//检查一下茶友提交的string，即i_word是否不为空，中文长度小于239字符之间
	if 1 > CnStrLen(i_word) || CnStrLen(i_word) > 239 {
		util.Warning(util.LogError(err), s_u.Email, " Cannot process invitation")
		Report(w, r, "你好，瞪大眼睛涨红了脸的茶博士，竟然强词夺理说，邀请的话太长了或者太短，只有外星人才接受呀，请确认再试。")
		return
	}
	role := r.PostFormValue("role")
	team_uuid := r.PostFormValue("team_uuid")

	//根据茶友提交的Uuid，检查是否存在该User
	invite_user, err := data.GetUserByEmail(email)
	if err != nil {
		util.Danger(util.LogError(err), email, " Cannot search user given email")
		Report(w, r, "你好，满头大汗的茶博士未能茶棚里找到这个茶友，请确认后再试。")
		return
	}

	//检查茶友是否自己邀请自己？也许是可以的，例如观音菩萨也可以加入自己创建的西天取经茶团喝茶？？
	/* if u.Email == i_email {
		util.Pop_message(w, r, "你好，请不要邀请自己加入茶团哈。")
		return
	} */
	//根据茶友提交的teamId，检查是否存在该team
	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Danger(util.LogError(err), s_u.Email, " Cannot search team given team_uuid")
		Report(w, r, "你好，茶博士未能找到这个团队，请确认后再试。")
		return
	}
	//检查当前茶友是否团队的Ceo或者founder，是否有权限邀请新成员
	ceo, err := team.MemberCEO()
	if err != nil {
		util.Danger(util.LogError(err), s_u.Email, " Cannot search team ceo")
		Report(w, r, "你好，未能找到茶团CEO，请确认后再试。")
		return
	}
	ceo_user, err := ceo.User()
	if err != nil {
		util.Danger(util.LogError(err), s_u.Email, " Cannot search ceo_user")
		Report(w, r, "你好，未能找到茶团CEO，请确认后再试。")
		return
	}

	founder, err := team.Founder()
	if err != nil {
		util.Danger(util.LogError(err), s_u.Email, " Cannot search team founder")
		Report(w, r, "你好，未能找到这个茶团的发起人，请确认后再试。")
		return
	}
	ok := s_u.Id == ceo_user.Id || s_u.Id == founder.Id
	if !ok {
		Report(w, r, "你好，机关算尽太聪明，反算了卿卿性命。只有团队CEO或者创始人能够邀请请新成员加盟。")
		return
	}

	//检查受邀请的茶友团队核心角色是否已经被占用
	switch role {
	case "CTO", "CMO", "CFO":
		//检查teamMember.Role中是否已经存在
		_, err = team.GetTeamMemberByRole(role)
		if err == nil {
			Report(w, r, "你好，该团队已经存在所选择的核心角色，请返回选择其他角色。")
			return
		} else if err != sql.ErrNoRows {
			util.Danger(util.LogError(err), s_u.Email, " Cannot search team member given team_id and role")
			Report(w, r, "你好，茶博士在这个团队角色事情迷糊了，请确认后再试。")
			return
		}

	case "CEO":
		if ceo_user.Id == founder.Id {
			//CEO是默认创建人担任首个CEO，这意味着首次更换CEO，例如观音菩萨指定唐僧为取经团队CEO，这是初始化团队操作
			break
		} else {
			Report(w, r, "你好，请先邀请茶友加盟为普通茶友，然后再调整角色，请确认后再试。")
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
	_, err = data.GetMemberByTeamIdUserId(team.Id, invite_user.Id)
	if err != nil {
		//如果err类型为空行，说明团队中还没有这个茶友，可以向其发送邀请函
		if err == sql.ErrNoRows {

			//创建一封邀请函
			invi := data.Invitation{
				TeamId:      team.Id,
				InviteEmail: invite_user.Email,
				Role:        role,
				InviteWord:  i_word,
				CreatedAt:   time.Now(),
				Status:      0,
			}
			//存储邀请函
			err = invi.Create()
			if err != nil {
				util.Danger(util.LogError(err), s_u.Email, " Cannot create invitation")
				Report(w, r, "你好，茶博士未能创建邀请函，请稍后再试。")
				return
			}
			// 向受邀请的茶友新消息小黑板上加1
			if err = data.AddUserMessageCount(invite_user.Id); err != nil {
				util.Danger(util.LogError(err), " Cannot add user new-message count")
				return
			}

			// 报告发送者成功消息
			mes := fmt.Sprintf("你好，成功以 %s 茶团名义，向茶友 %s 发送了加盟邀请函，请等待回复。", team.Abbreviation, invite_user.Name)
			Report(w, r, mes)
			return
		}
		//其他类型的error，打印出来分析错误
		util.Danger(util.LogError(err), s_u.Email, "error for Search teamMember given teamId and userId")
		return
	}
	//如果err为nil，说明茶友已经在茶团中，无需邀请
	Report(w, r, "你好，该茶友已经在茶团中，无需邀请。")

}

// GET /v1/team_member/invite?id=
// 编写对某个指定茶友的邀请函
func InviteMemberNew(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, s_d_family, s_all_families, s_d_team, s_survival_teams, s_d_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		util.Danger(util.LogError(err), "cannot fetch s_u s_teams given session")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	//根据茶友提交的Uuid，查询获取拟邀请加盟的茶友信息
	vals := r.URL.Query()
	user_uuid := vals.Get("id")
	invi_user, err := data.GetUserByUUID(user_uuid)
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get user given uuid")
		Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？请确认后再试")
		return
	}

	var iD data.InvitationDetail
	// 填写页面资料
	iD.SessUser = s_u
	iD.SessUserDefaultFamily = s_d_family
	iD.SessUserAllFamilies = s_all_families
	iD.SessUserDefaultTeam = s_d_team
	iD.SessUserSurvivalTeams = s_survival_teams
	iD.SessUserDefaultPlace = s_d_place
	iD.SessUserBindPlaces = s_places

	iD.InvitationBean.InviteUser = invi_user

	//检查一下s_u茶友是否有权以某个茶团Team的名义发送邀请函

	//首先检查是否某个茶团founder，则可以发送邀请函
	founder_teams, err := s_u.FounderTeams()
	if err != nil {
		util.Danger(util.LogError(err), "cannot get founder_teams given sessUser")
		Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？")
		return
	}
	for _, f_te := range founder_teams {
		if f_te.FounderId == s_u.Id {
			//向茶友返回指定的团队邀请函创建表单页面
			RenderHTML(w, &iD, "layout", "navbar.private", "member.invite")
			return
		}
	}

	// 检查s_u是否某个茶团的ceo
	teams, err := s_u.CeoTeams()
	if err != nil {
		util.Danger(util.LogError(err), "cannot get teams given sessUser")
		Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？")
		return
	}
	for _, te := range teams {
		ceo, err := te.MemberCEO()
		if err != nil {
			util.Danger(util.LogError(err), "cannot get ceo given team")
			Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？")
			return
		}
		if ceo.UserId == s_u.Id {
			//向茶友返回指定的团队邀请函创建表单页面
			RenderHTML(w, &iD, "layout", "navbar.private", "member.invite")
			return
		}
	}

	//既不是某个茶团发起人，也不是CEO，无法代表任何茶团发出邀请函
	Report(w, r, "你好，慢条斯理的茶博士竟然说，茶团CEO或者创建人，才能发送该团邀请函呢。")

}

// GET /v1/team_member/invitation?id=
// 用户查看收到的某封加盟邀请函详情及处理页面
func MemberInvitationRead(w http.ResponseWriter, r *http.Request) {
	//获取session
	s, err := Session(r)
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get user")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var iD data.InvitationDetail

	//根据茶友提交的Uuid，查询邀请函信息
	vals := r.URL.Query()
	invi_uuid := vals.Get("id")
	invi, err := data.GetInvitationByUuid(invi_uuid)
	if err != nil {
		//util.Warning(util.LogError(err), invi_uuid," Cannot get invitation given uuid")
		Report(w, r, "你好，茶博士正在努力的查找邀请函，请稍后再试。")
		return
	}

	//检查一下当前茶友是否有权查看此邀请函
	if invi.InviteEmail != s_u.Email {
		Report(w, r, "你好，该邀请函不属于您或者权限问题，无法查看。")
		return
	}

	i_b, err := FetchInvitationBean(invi)
	if err != nil {
		util.Danger(util.LogError(err), invi.Id, " Cannot fetch invitationBean given invitation")
		Report(w, r, "你好，茶博士正在努力的查找邀请函资料，请稍后再试。")
		return
	}

	//如果邀请函目前是未读状态=0，则将邀请函的状态改为已读=1
	if invi.Status == 0 {
		invi.Status = 1
		err = invi.Update()
		if err != nil {
			util.Danger(util.LogError(err), s_u.Email, " Cannot update invitation")
			Report(w, r, "你好，茶博士正在努力的更新邀请函状态，请稍后再试。")
			return
		}
		// 减去茶友1小黑板新消息数
		if err = data.SubtractUserMessageCount(s_u.Id); err != nil {
			util.Danger(util.LogError(err), " Cannot subtract user message count")
			return
		}

	}

	//填写页面资料
	iD.SessUser = s_u
	iD.InvitationBean = i_b

	//向茶友返回该邀请函的详细信息
	RenderHTML(w, &iD, "layout", "navbar.private", "member.invitation_read")
}
