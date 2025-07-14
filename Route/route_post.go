package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/post/detail?uuid=
// 品味的详情
func PostDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	vals := r.URL.Query()
	uuid := vals.Get("uuid")
	var pD data.PostDetail
	t_post := data.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	post_bean, err := FetchPostBean(t_post)
	if err != nil {
		util.Debug(" Cannot get post bean given post", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	pD.PostBean = post_bean
	// 读取此品味引用的茶议
	quote_thread, err := t_post.Thread()
	if err != nil {
		util.Debug(" Cannot get thread given post", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议资料。")
		return
	}
	pD.QuoteThreadBean, err = FetchThreadBean(quote_thread, r)
	if err != nil {
		util.Debug(" Cannot get thread given post", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议资料。")
		return
	}
	// 截短此引用的茶议内容以方便展示
	pD.QuoteThreadBean.Thread.Body = Substr(pD.QuoteThreadBean.Thread.Body, 66)

	// 读取全部针对此品味的茶议
	thread_slice, err := t_post.Threads()
	if err != nil {
		util.Debug(" Cannot get thread_slice given t_post", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	pD.ThreadBeanSlice, err = FetchThreadBeanSlice(thread_slice, r)
	if err != nil {
		util.Debug(" Cannot get thread_bean_slice given thread_slice", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}

	// 读取此品味的引用茶议（源自）引用茶台
	quote_project, err := quote_thread.Project()
	if err != nil {
		util.Debug(quote_thread.Id, " Cannot get project given thread")
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	pD.QuoteProjectBean, err = FetchProjectBean(quote_project)
	if err != nil {
		util.Debug(quote_project.Id, " Cannot get project given project")
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}

	// 读取此品味的引用茶议（源自）引用茶台，引用的茶围
	quote_objective, err := quote_project.Objective()
	if err != nil {
		util.Debug(quote_project.Id, " Cannot get objective given project")
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	pD.QuoteObjectiveBean, err = FetchObjectiveBean(quote_objective)
	if err != nil {
		util.Debug(quote_objective.Id, " Cannot get objective given objective")
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}

	// 检测pageData.ThreadSlice数量是否超过一打dozen
	if len(thread_slice) > 12 {
		pD.IsOverTwelve = true
	} else {
		pD.IsOverTwelve = false
	}

	// 读取会话
	s, err := Session(r)
	if err != nil {
		// 未登录，游客
		pD.IsGuest = true

		// 填写页面数据
		pD.SessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
			// 用户足迹
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		RenderHTML(w, &pD, "layout", "navbar.public", "post.detail", "component_sess_capacity", "component_avatar_name_gender")
		return
	}
	// 读取已登陆用户资料
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	pD.SessUser = s_u
	// 从会话查获当前浏览用户资料荚
	s_u, s_default_family, s_all_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchSessionUserRelatedData(s)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	pD.SessUser = s_u
	pD.IsGuest = false
	pD.IsInput = true

	pD.IsAdmin, err = checkObjectiveAdminPermission(&quote_objective, s_u.Id)
	if err != nil {
		util.Debug(" Cannot check objective admin permission", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	pD.IsMaster, err = checkProjectMasterPermission(&quote_project, s_u.Id)
	if err != nil {
		util.Debug(" Cannot check project master permission", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	if !pD.IsAdmin && !pD.IsMaster {
		veri_team := data.Team{Id: data.TeamIdVerifier}
		is_member, err := veri_team.IsMember(s_u.Id)
		if err != nil {
			util.Debug("Cannot check verifier team member", err)
			Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
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

	// 当前会话用户是否此品味作者？
	if s_u.Id == t_post.UserId {
		pD.IsAuthor = true
	} else {
		pD.IsAuthor = false
	}

	RenderHTML(w, &pD, "layout", "navbar.private", "post.detail", "component_sess_capacity", "component_avatar_name_gender")

}

// POST /v1/post/draft
// Create the post 创建品味（跟帖/回复）草稿 new
func NewPostDraft(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取用户会话信息。请重新登录或联系管理员。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		Report(w, r, "你好，茶博士摸摸头，竟然说今天电脑飞去热带海岛潜水度假了。")
		return
	}

	//读取用户表态,立场是支持（true）或者反对(false)
	post_attitude := r.PostFormValue("attitude") == "true"

	body := r.PostFormValue("body")
	//检查body的长度，规则是不能少于刘姥姥评价老君眉的品味字数
	if CnStrLen(body) <= int(util.Config.PostMinWord) {
		Report(w, r, "你好，戴着厚厚眼镜片的茶博士居然说，请不要用隐形墨水来写品味内容。")
		return
	} else if CnStrLen(body) > int(util.Config.ThreadMaxWord) {
		Report(w, r, "你好，彬彬有礼戴着厚厚眼镜片的茶博士居然说，内容太多，茶叶蛋壳都用光了也写不完呀。")
		return
	}
	//读取针对的目标茶议
	thre_uuid := r.PostFormValue("uuid")
	//检查uuid是否有效
	t_thread, err := data.GetThreadByUUID(thre_uuid)
	if err != nil {
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶议。")
		return
	}

	te_id_str := r.PostFormValue("team_id")
	//change team_id to int
	team_id := 0 // Default value for invalid input
	if te_id_str != "" {
		team_id, err = strconv.Atoi(te_id_str)
		if err != nil {
			util.Debug(" Cannot change team_id to int", te_id_str, err)
			Report(w, r, "一年三百六十日，风刀霜剑严相逼，请确认提交的团队编号。")
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
			Report(w, r, "一年三百六十日，风刀霜剑严相逼，请确认提交的家庭编号。")
			return
		}
	}
	is_private := r.PostFormValue("is_private") == "true"

	// 茶议所在的茶台
	t_proj, err := t_thread.Project()
	if err != nil {
		util.Debug(" Cannot get project by project id", t_proj.Id, err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}

	switch t_proj.Class {
	case data.PrClassOpen:
		// 开放式茶台，任何人可以品茶
		// 直接继续创建流程

	case data.PrClassClose:
		// 封闭式茶台，检查邀请状态
		ok, err := t_proj.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug("Cannot check is invited member by project id", t_proj.Id, err)
			Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
			return
		}
		if !ok {
			Report(w, r, "你好，难以置信，陛下的大名竟然不在邀请品茶名单上！")
			return
		}

	default:
		Report(w, r, "你好，茶博士满头大汗说，这个茶台状态异常无法使用。")
		return
	}

	//确定品味发布者身份
	is_admin := false

	is_master, err := checkProjectMasterPermission(&t_proj, s_u.Id)
	if err != nil {
		util.Debug(" Cannot check project master permission", t_proj.Id, err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}
	//所在的茶围
	t_obje, err := t_proj.Objective()
	if err != nil {
		util.Debug(" Cannot get objective by objective id", t_obje.Id, err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}
	if !is_master {
		is_admin, err = checkObjectiveAdminPermission(&t_obje, s_u.Id)
		if err != nil {
			util.Debug(" Cannot check objective admin permission", t_obje.Id, err)
			Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
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
	dp_class := data.PostClassNormal
	if is_admin || is_master {
		dp_class = data.PostClassAdmin
	}

	new_draft_post := data.DraftPost{
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
		Report(w, r, "你好，茶博士摸摸头，嘀咕笔头宝珠掉了，记录您的品味失败。")
		return
	}

	if util.Config.PoliteMode {

		if err := CreateAndSendAcceptMessage(new_draft_post.Id, data.AcceptObjectTypePo, s_u.Id); err != nil {
			// 根据错误类型返回不同提示
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				Report(w, r, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				Report(w, r, "你好，茶博士迷路了，未能发送蒙评请求消息。")
			}
			return
		}
		t := fmt.Sprintf("你好，对“ %s ”发布的品味已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", t_thread.Title)
		// 提示用户草稿保存成功
		Report(w, r, t)
		return

	} else {
		post, err := AcceptNewDraftPost(new_draft_post.Id)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "获取品味草稿失败"):
				util.Debug("Cannot get draft-post", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候泡一壶好茶的关键，需要的不是技术而是耐心。")
			case strings.Contains(err.Error(), "创建新品味失败"):
				util.Debug("Cannot save post", err)
				Report(w, r, "你好，吟成荳蔻才犹艳，睡足酴醾梦也香。")
			default:
				util.Debug("未知错误", err)
				Report(w, r, "世事洞明皆学问，人情练达即文章。")
			}
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/v1/post/detail?uuid=%s", post.Uuid), http.StatusFound)
		return
	}
}

// 修改post的处理器
func HandleEditPost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		EditPost(w, r)
	case http.MethodPost:
		UpdatePost(w, r)
	case "PUT":
		//未允许的方法
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	case "DELETE":
		//未开放的窗口
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /post/edit
// update the post
// 更新用户的post，规则是可以补充内容，不能覆载修改之前说的话，不能变更立场（从支持变反对）。
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	//读取请求会话
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	//从会话中读取用户资料
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	t_post := data.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail given uuid", uuid)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}

	if t_post.UserId != s_u.Id {
		util.Debug(" Cannot edit other user's post", err)
		Report(w, r, "茶博士提示，目前仅能补充自己的回复")
		return
	} else {
		//可以补充自己的表态内容
		body := r.PostFormValue("body")
		if body != "" {
			//检查补充内容是否有意义，rune 字数>int(util.Config.ThreadMinWord),总的post字数<int(util.Config.ThreadMaxWord)
			if CnStrLen(body) > int(util.Config.ThreadMinWord) && CnStrLen(t_post.Body)+CnStrLen(body) < int(util.Config.ThreadMaxWord) {
				t_post.Body += body
			} else {
				//提示用户总字数或者本次提交补充内容超出字数限制
				Report(w, r, "你好， 粗鲁的茶博士竟然说字数满了，纸条写不下您的品味。")
				return
			}
			err = t_post.UpdateBody(body)
			if err != nil {
				util.Debug(" Cannot update post", err)
				Report(w, r, "茶博士失魂鱼，未能更新专属资料，请稍后再试。")
				return
			}
			thread, err := data.GetThreadById(t_post.ThreadId)
			if err != nil {
				util.Debug(" Cannot read thread", err)
				Report(w, r, "茶博士失魂鱼，嘀咕无为有处有还无？，请稍后再试。")
			}
			url := fmt.Sprint("/v1/thread/detail?uuid=", thread.Uuid)
			http.Redirect(w, r, url, http.StatusFound)
		} else {
			//空白或者一个字被认为是无意义追加内容
			Report(w, r, "你好，陛下英明，请勿提供小于17个字的品味补充")
			return
		}

	}

}

// GET /post/edit
// 用户补充自己的表态内容POST的界面
func EditPost(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	user, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	vals := r.URL.Query()
	uuid := vals.Get("id")
	t_post := data.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail", err)
		Report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	if t_post.UserId == user.Id {
		RenderHTML(w, &t_post, "layout", "navbar.private", "post.edit")
	} else {
		util.Debug(" Cannot edit other user's post", err)
		Report(w, r, "茶博士提示，目前仅能补充自己的回复")
	}

}
