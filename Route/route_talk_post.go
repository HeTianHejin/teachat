package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	dao "teachat/DAO"
	util "teachat/Util"
	"time"
)

// GET /v1/post/detail?uuid=
// 品味的详情
func PostDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	var pD dao.PostDetail
	s_u := dao.UserUnknown
	t_post := dao.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail", err)
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	post_bean, err := fetchPostBean(t_post)
	if err != nil {
		util.Debug(" Cannot get post bean given post", err)
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	pD.PostBean = post_bean
	// 读取此品味引用的茶议
	quote_thread, err := t_post.Thread()
	if err != nil {
		util.Debug(" Cannot get thread given post", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取茶议资料。")
		return
	}
	pD.QuoteThreadBean, err = fetchThreadBean(quote_thread, r)
	if err != nil {
		util.Debug(" Cannot get thread given post", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取茶议资料。")
		return
	}
	// 截短此引用的茶议内容以方便展示
	pD.QuoteThreadBean.Thread.Body = subStr(pD.QuoteThreadBean.Thread.Body, 66)

	// 读取全部针对此品味的茶议
	thread_slice, err := t_post.Threads()
	if err != nil {
		util.Debug(" Cannot get thread_slice given t_post", err)
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	pD.ThreadBeanSlice, err = fetchThreadBeanSlice(thread_slice, r)
	if err != nil {
		util.Debug(" Cannot get thread_bean_slice given thread_slice", err)
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}

	// 读取此品味的引用茶议（源自）引用茶台
	quote_project, err := quote_thread.Project()
	if err != nil {
		util.Debug(quote_thread.Id, " Cannot get project given thread")
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	pD.QuoteProjectBean, err = fetchProjectBean(quote_project)
	if err != nil {
		util.Debug(quote_project.Id, " Cannot get project given project")
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}

	// 读取此品味的引用茶议（源自）引用茶台，引用的茶围
	quote_objective, err := quote_project.Objective()
	if err != nil {
		util.Debug(quote_project.Id, " Cannot get objective given project")
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	pD.QuoteObjectiveBean, err = fetchObjectiveBean(quote_objective)
	if err != nil {
		util.Debug(quote_objective.Id, " Cannot get objective given objective")
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}

	// 检测pageData.ThreadSlice数量是否超过一打dozen
	if len(thread_slice) > 12 {
		pD.IsOverTwelve = true
	} else {
		pD.IsOverTwelve = false
	}

	// 读取会话
	s, err := session(r)
	if err != nil {
		// 未登录，游客
		pD.IsGuest = true

		// 填写页面数据
		pD.SessUser = dao.User{
			Id:   dao.UserId_None,
			Name: "游客",
			// 陛下足迹
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		generateHTML(w, &pD, "layout", "navbar.public", "post.detail", "component_sess_capacity", "component_avatar_name_gender")
		return
	}
	// 读取已登陆陛下资料
	s_u, err = s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	pD.SessUser = s_u
	// 从会话查获当前浏览陛下资料荚
	s_u, s_default_family, s_all_families, s_default_team, s_survival_teams, s_default_place, s_places, err := fetchSessionUserRelatedData(s)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 陛下足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	pD.SessUser = s_u
	pD.IsGuest = false
	pD.IsInput = true

	pD.IsAdmin, err = checkObjectiveAdminPermission(&quote_objective, s_u.Id)
	if err != nil {
		util.Debug(" Cannot check objective admin permission", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	pD.IsMaster, err = checkProjectMasterPermission(&quote_project, s_u.Id)
	if err != nil {
		util.Debug(" Cannot check project master permission", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	if !pD.IsAdmin && !pD.IsMaster {
		veri_team := dao.Team{Id: dao.TeamIdVerifier}
		is_member, err := veri_team.IsMember(s_u.Id)
		if err != nil {
			util.Debug("Cannot check verifier team member", err)
			report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
			return
		}
		if is_member {
			pD.IsVerifier = true
		}
	}

	// 默认家庭
	pD.SessUserDefaultFamily = s_default_family
	// 全部家庭
	pD.SessUserSurvivalFamilies = s_all_families

	// 默认团队
	pD.SessUserDefaultTeam = s_default_team
	// 全部团队
	pD.SessUserSurvivalTeams = s_survival_teams

	// 默认地点
	pD.SessUserDefaultPlace = s_default_place
	// 全部绑定地点
	pD.SessUserBindPlaces = s_places

	// 当前会话陛下是否此品味作者？
	if s_u.Id == t_post.UserId {
		pD.IsAuthor = true
	} else {
		pD.IsAuthor = false
	}

	generateHTML(w, &pD, "layout", "navbar.private", "post.detail", "component_sess_capacity", "component_avatar_name_gender")

}

// POST /v1/post/draft
// Create the post 创建品味（跟帖/回复）草稿 new
func NewPostDraft(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，未能读取陛下会话信息。请重新登录或联系管理员。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说人工智能助理飞去热带海岛潜水度假了。")
		return
	}

	//读取针对的目标茶议
	thread_uuid := r.PostFormValue("uuid")
	//检查uuid是否有效
	t_thread, err := dao.GetThreadByUUID(thread_uuid)
	if err != nil {
		report(w, s_u, "你好，根据陛下的指示，却未能读取目标茶议。")
		return
	}
	ctx := r.Context()
	posted := dao.Post{UserId: s_u.Id, ThreadId: t_thread.Id}
	posted_exists, err := posted.HasUserPostedInThread(ctx)
	if err != nil {
		util.Debug("failed to check has-user-posted-thread", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说人工智能助理飞去热带海岛潜水度假了。")
		return
	}
	if posted_exists {
		report(w, s_u, "你好，陛下已在该话题下发表过品味了。还有话说？可以加深您的品味内涵噢。")
		return
	}

	//读取表态,立场是支持（true）或者反对(false)
	post_attitude := r.PostFormValue("attitude") == "true"

	body := r.PostFormValue("body")
	//检查body的长度，规则是不能少于刘姥姥评价老君眉的品味字数
	if cnStrLen(body) <= int(util.Config.PostMinWord) {
		report(w, s_u, "你好，戴着厚厚眼镜片的茶博士居然说，请不要用隐形墨水来写品味内容。")
		return
	} else if cnStrLen(body) > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "你好，彬彬有礼戴着厚厚眼镜片的茶博士居然说，内容太多，茶叶蛋壳都用光了也写不完呀。")
		return
	}

	te_id_str := r.PostFormValue("team_id")
	//change team_id to int
	team_id := 0 // Default value for invalid input
	if te_id_str != "" {
		team_id, err = strconv.Atoi(te_id_str)
		if err != nil {
			util.Debug(" Cannot change team_id to int", te_id_str, err)
			report(w, s_u, "一年三百六十日，风刀霜剑严相逼，请确认提交的团队编号。")
			return
		}
	}
	family_id_str := r.PostFormValue("family_id")
	//change family_id to int
	family_id := 0 // Default value for invalid input
	if family_id_str != "" {
		family_id, err = strconv.Atoi(family_id_str)
		if err != nil {
			util.Debug(" Cannot change family_id to int", family_id_str, err)
			report(w, s_u, "一年三百六十日，风刀霜剑严相逼，请确认提交的家庭编号。")
			return
		}
	}
	is_private := r.PostFormValue("is_private") == "true"

	// 茶议所在的茶台
	t_proj, err := t_thread.Project()
	if err != nil {
		util.Debug(" Cannot get project by project id", t_proj.Id, err)
		report(w, s_u, "你好，未能读取专属茶台资料。")
		return
	}

	switch t_proj.Class {
	case dao.PrClassOpen:
		// 开放式茶台，任何人可以品茶
		// 直接继续创建流程

	case dao.PrClassClose:
		// 封闭式茶台，检查邀请状态
		ok, err := t_proj.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug("Cannot check is invited member by project id", t_proj.Id, err)
			report(w, s_u, "你好，未能读取专属茶台资料。")
			return
		}
		if !ok {
			report(w, s_u, "你好，难以置信，陛下的大名竟然不在邀请品茶名单上！")
			return
		}

	default:
		report(w, s_u, "你好，茶博士说，这个茶台状态异常无法使用。")
		return
	}

	//确定品味发布者身份
	is_admin := false

	is_master, err := checkProjectMasterPermission(&t_proj, s_u.Id)
	if err != nil {
		util.Debug(" Cannot check project master permission", t_proj.Id, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}
	//所在的茶围
	t_obje, err := t_proj.Objective()
	if err != nil {
		util.Debug(" Cannot get objective by objective id", t_obje.Id, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}
	if !is_master {
		is_admin, err = checkObjectiveAdminPermission(&t_obje, s_u.Id)
		if err != nil {
			util.Debug(" Cannot check objective admin permission", t_obje.Id, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
			return
		}
	}
	if is_admin {
		is_private = t_obje.IsPrivate
		family_id = t_obje.FamilyId
		team_id = t_obje.TeamId
	} else if is_master {
		is_private = t_proj.IsPrivate
		family_id = t_proj.FamilyId
		team_id = t_proj.TeamId
	}

	//确定是哪一种级别发布
	dp_class := dao.PostClassNormal
	if is_admin || is_master {
		dp_class = dao.PostClassAdmin
	}

	new_draft_post := dao.DraftPost{
		UserId:    s_u.Id,
		ThreadId:  t_thread.Id,
		FamilyId:  family_id,
		TeamId:    team_id,
		Attitude:  post_attitude,
		IsPrivate: is_private,
		Body:      body,
		Class:     dp_class,
	}
	if err = new_draft_post.Create(); err != nil {
		util.Debug("Cannot create draft post", s_u.Email, err)
		report(w, s_u, "你好，茶博士摸摸头，嘀咕笔头宝珠掉了，记录您的品味失败。")
		return
	}

	if util.Config.PoliteMode {
		// 友邻蒙评模式
		if err := createAndSendAcceptNotification(new_draft_post.Id, dao.AcceptObjectTypePost, s_u.Id, r.Context()); err != nil {
			// 根据错误类型返回不同提示
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				report(w, s_u, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				report(w, s_u, "你好，茶博士迷路了，未能发送蒙评请求通知。")
			}
			return
		}
		t := fmt.Sprintf("你好，对“ %s ”发布的品味已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", t_thread.Title)
		// 提示草稿保存成功
		report(w, s_u, t)
		return

	} else {

		_, err := acceptNewDraftPost(new_draft_post.Id)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "获取品味草稿失败"):
				util.Debug("Cannot get draft-post", err)
				report(w, s_u, "你好，茶博士竟然说，有时候泡一壶好茶的关键，需要的不是技术而是耐心。")
			case strings.Contains(err.Error(), "创建新品味失败"):
				util.Debug("Cannot save post", err)
				report(w, s_u, "你好，吟成荳蔻才犹艳，睡足酴醾梦也香。")
			default:
				util.Debug("未知错误", err)
				report(w, s_u, "世事洞明皆学问，人情练达即文章。")
			}
			return
		}
		http.Redirect(w, r, "/v1/thread/detail?uuid="+thread_uuid, http.StatusFound)
		return
	}
}

// 修改post的处理器
func HandleSupplementPost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SupplementPostGet(w, r)
	case http.MethodPost:
		SupplementPostPost(w, r)
	case "PUT":
		//未允许的方法
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	case "DELETE":
		//未开放的窗口
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /v1/post/supplement
// 加水（补充）已经发布的post，规则是可以补充内容，不能覆载修改之前说的话，不能变更立场（从支持变反对）。
func SupplementPostPost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//从会话中读取陛下资料
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}

	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	t_post := dao.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail given uuid", uuid)
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	body_addi := r.PostFormValue("additional")
	if body_addi != "" {
		//检查补充内容是否有意义，字数>int(util.Config.ThreadMinWord),总的post字数<int(util.Config.ThreadMaxWord)
		if cnStrLen(body_addi) < int(util.Config.ThreadMinWord) {
			report(w, s_u, "你好， 粗鲁的茶博士竟然说，陛下下旨字数太少了？")
			return
		} else if cnStrLen(t_post.Body)+cnStrLen(body_addi) > int(util.Config.ThreadMaxWord) {
			//提示陛下总字数或者本次提交补充内容超出字数限制
			report(w, s_u, "你好， 粗鲁的茶博士竟然说字数满了，纸条写不下您的品味。")
			return
		}
	}

	// 检查是否有权加水当前帖子
	ok := false
	// 如果是作者本人，ok
	if t_post.UserId == s_u.Id {
		ok = true
	} else {
		//检查是否是品味发布者所在团队成员，或者家庭成员
		if t_post.IsPrivate {
			//检查是否是品味发布者所在家庭成员
			family := dao.Family{Id: t_post.FamilyId}
			if is_member, err := family.IsMember(s_u.Id); err != nil || !is_member {
				util.Debug(" Cannot check family member", err)
				report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
				return
			}
			//是发布者所在家庭成员，可以编辑
			ok = true

		} else {
			//检查是否是品味发布者所在团队成员
			team := dao.Team{Id: t_post.TeamId}
			if is_member, err := team.IsMember(s_u.Id); err != nil || !is_member {
				util.Debug(" Cannot check team member", err)
				report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
				return
			}
			//是发布者所在团队成员，可以编辑
			ok = true
		}

	}
	if ok {
		//可以补充自己的表态内容
		//当前“[中文时间字符 + 补充]” + body
		now := time.Now()
		timeStr := now.Format("2006年1月2日 15:04:05")
		name := s_u.Name
		// 追加内容（另起一行）// 注意开头的 \n
		t_post.Body += "\n[ " + timeStr + " " + name + " 补充 ]" + body_addi

		err = t_post.UpdateBody()
		if err != nil {
			util.Debug(" Cannot update post", err)
			report(w, s_u, "你好，茶博士失魂鱼，墨水中断未能补充品味。")
			return
		}
		// thread, err := dao.GetThreadById(t_post.ThreadId)
		// if err != nil {
		// 	util.Debug(" Cannot read thread", err)
		// 	Report(w, r, "身后有余忘缩手，眼前无路想回头，请稍后再试。")
		// 	return
		// }
		url := fmt.Sprint("/v1/post/detail?uuid=", t_post.Uuid)
		http.Redirect(w, r, url, http.StatusFound)
		return
	} else {
		//提示无权操作
		report(w, s_u, "你好，陛下英明，请勿往陌生人的茶杯加水呢？")
		return
	}
}

// GET /v1/post/supplement?uuid=xxx
// 陛下加水（补充）自己的表态内容,完善补漏
func SupplementPostGet(w http.ResponseWriter, r *http.Request) {
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
	uuid := vals.Get("uuid")
	t_post := dao.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail", err)
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	ok := false
	// 检查是否是品味发布者本人
	if t_post.UserId == s_u.Id {
		//是发布者本人，可以编辑
		ok = true
	} else {
		//检查是否是品味发布者所在团队成员，或者家庭成员
		if t_post.IsPrivate {
			//检查是否是品味发布者所在家庭成员
			family := dao.Family{Id: t_post.FamilyId}
			if is_member, err := family.IsMember(s_u.Id); err != nil || !is_member {
				util.Debug(" Cannot check family member", err)
				report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
				return
			}
			//是发布者所在家庭成员，可以编辑
			ok = true

		} else {
			//检查是否是品味发布者所在团队成员
			team := dao.Team{Id: t_post.TeamId}
			if is_member, err := team.IsMember(s_u.Id); err != nil || !is_member {
				util.Debug(" Cannot check team member", err)
				report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
				return
			}
			//是发布者所在团队成员，可以编辑
			ok = true
		}
	}
	var pD dao.PostDetail
	pD.SessUser = s_u
	pD.IsInput = true
	pD.PostBean, err = fetchPostBean(t_post)
	if err != nil {
		util.Debug(" Cannot fetch post bean", t_post.Id, err)
		report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
		return
	}
	quote_thread, err := t_post.Thread()
	if err != nil {
		util.Debug(" Cannot get thread given post", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取茶议资料。")
		return
	}
	pD.QuoteThreadBean, err = fetchThreadBean(quote_thread, r)
	if err != nil {
		util.Debug(" Cannot get thread given post", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能读取茶议资料。")
		return
	}
	project, err := quote_thread.Project()
	if err != nil {
		util.Debug(" Cannot read project given thread", err)
		report(w, s_u, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}
	pD.QuoteProjectBean, err = fetchProjectBean(project)
	if err != nil {
		util.Debug(" Cannot read project bean given project", project.Id, err)
		report(w, s_u, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}
	objective, err := project.Objective()
	if err != nil {
		util.Debug(" Cannot read objective given project", err)
		report(w, s_u, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}
	pD.QuoteObjectiveBean, err = fetchObjectiveBean(objective)
	if err != nil {
		util.Debug(" Cannot read objective given project", project.Id, err)
		report(w, s_u, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}
	if ok {
		//显示编辑页面
		generateHTML(w, &pD, "layout", "navbar.private", "post.supplement", "component_avatar_name_gender")
	} else {
		//提示陛下无权编辑
		report(w, s_u, "你好，茶博士扶起厚厚的眼镜，居然说陛下您没有权限加水呢。")
		return
	}

}
