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

// 加水 ，修改回复post的处理器
func HandleEditPost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		EditPost(w, r)
	case http.MethodPost:
		UpdatePost(w, r)
	case "PUT":
		//未开放的窗口
		w.WriteHeader(http.StatusInternalServerError)
		return
	case "DELETE":
		//未开放的窗口
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// get /v1/post/detail?id=
// 品味的详情
func PostDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	vals := r.URL.Query()
	uuid := vals.Get("id")
	var pD data.PostDetail
	t_post := data.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	post_bean, err := FetchPostBean(t_post)
	if err != nil {
		util.Debug(" Cannot get post bean given post", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
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
	pD.QuoteThreadBean, err = FetchThreadBean(quote_thread)
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
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	pD.ThreadBeanSlice, err = FetchThreadBeanSlice(thread_slice)
	if err != nil {
		util.Debug(" Cannot get thread_bean_slice given thread_slice", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}

	// 读取此品味的引用茶议（源自）引用茶台
	quote_project, err := quote_thread.Project()
	if err != nil {
		util.Debug(quote_thread.Id, " Cannot get project given thread")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	pD.QuoteProjectBean, err = FetchProjectBean(quote_project)
	if err != nil {
		util.Debug(quote_project.Id, " Cannot get project given project")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}

	// 读取此品味的引用茶议（源自）引用茶台，引用的茶围
	quote_objective, err := quote_project.Objective()
	if err != nil {
		util.Debug(quote_project.Id, " Cannot get objective given project")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	pD.QuoteObjectiveBean, err = FetchObjectiveBean(quote_objective)
	if err != nil {
		util.Debug(quote_objective.Id, " Cannot get objective given objective")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
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
		pD.IsAuthor = false
		pD.IsInput = false
		// 填写页面数据
		pD.SessUser = data.User{
			Id:   0,
			Name: "游客",
			// 用户足迹
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		RenderHTML(w, &pD, "layout", "navbar.public", "post.detail")
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
	s_u, s_default_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	// 把系统默认家庭资料加入s_survival_families
	s_survival_families = append(s_survival_families, DefaultFamily)
	// 把系统默认团队资料加入s_survival_teams
	s_survival_teams = append(s_survival_teams, FreelancerTeam)

	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	pD.SessUser = s_u
	pD.IsGuest = false
	pD.IsInput = true

	// 默认家庭
	pD.SessUserDefaultFamily = s_default_family
	// 全部家庭
	pD.SessUserSurvivalFamilies = s_survival_families

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

	RenderHTML(w, &pD, "layout", "navbar.private", "post.detail")

}

// POST /v1/post/draft
// Create the post 创建品味（跟帖/回复）草稿 new
func NewPostDraft(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		Report(w, r, "你好，茶博士摸摸头，竟然说今天电脑去热带海岛潜水了。")
		return
	}

	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}

	//读取用户表态,立场是支持（true）或者反对(false)
	attitude := r.PostFormValue("attitude") == "true"

	body := r.PostFormValue("body")
	//检查body的长度，规则是不能少于刘姥姥评价老君眉的品味字数
	if CnStrLen(body) <= 17 {
		Report(w, r, "你好，戴着厚厚眼镜片的茶博士居然说，请不要用隐形墨水来写品味内容。")
		return
	} else if CnStrLen(body) > 456 {
		Report(w, r, "你好，彬彬有礼戴着厚厚眼镜片的茶博士居然说，内容太多，茶叶蛋壳都用光了也写不完呀。")
		return
	}
	uuid := r.PostFormValue("uuid")
	//检查uuid是否有效
	thread, err := data.ThreadByUUID(uuid)
	if err != nil {
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶议。")
		return
	}
	tid_string := r.PostFormValue("team_id")
	if tid_string == "" {
		Report(w, r, "一年三百六十日，风刀霜剑严相逼，请确认提交的团队编号。")
		return
	}
	//change team_id to int
	team_id, err := strconv.Atoi(tid_string)
	if err != nil {
		Report(w, r, "一年三百六十日，风刀霜剑严相逼，请确认提交的团队编号。")
		return
	}
	family_id_str := r.PostFormValue("family_id")
	if family_id_str == "" {
		Report(w, r, "一年三百六十日，风刀霜剑严相逼，请确认提交的家庭编号。")
		return
	}
	//change family_id to int
	family_id, err := strconv.Atoi(family_id_str)
	if err != nil {
		Report(w, r, "一年三百六十日，风刀霜剑严相逼，请确认提交的家庭编号。")
		return
	}
	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"

	//提交的茶团id,是team.id,team_id = 2是默认的初始自由人团队，无需检查
	// check submit team_id is valid
	if team_id != 2 {
		team := data.Team{Id: team_id}
		is_member, err := team.IsMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot get family member by family id and user id", err)
			Report(w, r, "你好，茶博士失魂鱼，未能读取茶团成员资格资料。")
			return
		}
		if !is_member {
			Report(w, r, "你好，茶团成员资格检查未通过，请确认后再试。")
			return
		}

	}

	//提交的茶团id,是family.id,检查提交者是否是家庭成员
	// check submit family_id is valid
	if family_id != 0 {

		family := data.Family{Id: family_id}
		is_member, err := family.IsMember(s_u.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				//util.PanicTea(util.LogError(err), " Cannot get family member by family id and user id")
				Report(w, r, "你好，茶博士认为您不是这个家庭成员，请确认后再试。")
				return
			}
			util.Debug(" Cannot get family member by family id and user id", family_id, s_u.Id)
			Report(w, r, "你好，茶博士失魂鱼，未能读取家庭成员资格资料。")
			return
		}
		if !is_member {
			//util.PanicTea(util.LogError(err), " Cannot get family member by family id and user id")
			Report(w, r, "你好，家庭成员资格检查失败，请确认后再试。")
			return
		}
	}

	// dp_class_str := r.PostFormValue("dp_class")
	// if dp_class_str == "" {
	// 	Report(w, r, "一年三百六十日，风刀霜剑严相逼，请确认提交的品味分类编号。")
	// 	return
	// }
	//change dp_class to int
	// dp_class, err := strconv.Atoi(dp_class_str)
	// if err != nil {
	// 	Report(w, r, "一年三百六十日，风刀霜剑严相逼，请确认提交的品味分类编号。")
	// 	return
	// }

	// 茶议所在的茶台，
	t_proj, err := thread.Project()
	if err != nil {
		util.Debug(" Cannot get project by project id", t_proj.Id)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}
	//所在的茶围
	t_obje, err := t_proj.Objective()
	if err != nil {
		util.Debug(" Cannot get objective by objective id", t_obje.Id)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}

	dp_class := 0

	ob_team := data.Team{Id: t_obje.TeamId}
	is_member, err := ob_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug(" Cannot get family member by family id and user id", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶团成员资格资料。")
		return
	}
	if is_member {
		dp_class = 1
	}

	new_draft_post := data.DraftPost{
		UserId:    s_u.Id,
		ThreadId:  thread.Id,
		FamilyId:  family_id,
		TeamId:    team_id,
		Attitude:  attitude,
		IsPrivate: is_private,
		Body:      body,
		Class:     dp_class,
	}

	switch t_proj.Class {
	case 1:
		// class=1可以品茶，

		if err = new_draft_post.Create(); err != nil {
			util.Debug(s_u.Email, " Cannot create draft post")
			Report(w, r, "你好，茶博士摸摸头，嘀咕笔头宝珠掉了，记录您的品味失败。")
			return
		}

	case 2:
		// 当前会话用户是否可以入席品茶？需要看台主指定了那些茶团成员可以品茶
		ok, err := t_proj.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot get project by project id", t_proj.Id)
			Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
			return
		}
		if !ok {
			// Cannot have tea
			Report(w, r, "你好，你的大名竟然不在邀请品茶名单上。")
			return
		}

		// Can have tea
		if err = new_draft_post.Create(); err != nil {
			util.Debug(s_u.Email, " Cannot create draft post")
			Report(w, r, "你好，茶博士摸摸头，嘀咕笔头宝珠掉了，记录您的品味失败。")
			return
		}

	default:
		// 异常状态的茶台
		Report(w, r, "你好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
		return
	}

	// 创建一条友邻蒙评,是否接纳 新茶的记录
	aO := data.AcceptObject{
		ObjectId:   new_draft_post.Id,
		ObjectType: 4,
	}
	if err = aO.Create(); err != nil {
		util.Debug("Cannot create accept_object given draft_post_id", new_draft_post.Id)
		Report(w, r, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
		return
	}
	// 发送邻座蒙评消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
		AcceptObjectId: aO.Id,
	}
	// 发送消息
	if err = TwoAcceptMessagesSendExceptUserId(s_u.Id, mess); err != nil {
		Report(w, r, "你好，茶博士迷路了，未能发送蒙评请求消息。")
		return
	}
	// 提示用户草稿保存成功
	t := fmt.Sprintf("你好，对“ %s ”发布的品味已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", thread.Title)
	// 提示用户草稿保存成功
	Report(w, r, t)
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
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
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
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	t_post := data.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Debug(" Cannot get post detail given uuid", uuid)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
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
			//检查补充内容是否有意义，rune 字数>17,总的post字数<456
			if CnStrLen(body) > 17 && CnStrLen(t_post.Body)+CnStrLen(body) < 456 {
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
				Report(w, r, "茶博士失魂鱼，未能读取专属资料，请稍后再试。")
			}
			url := fmt.Sprint("/v1/thread/detail?id=", thread.Uuid)
			http.Redirect(w, r, url, http.StatusFound)
		} else {
			//空白或者一个字被认为是无意义追加内容
			Report(w, r, "你好，请勿提供小于17个字的品味补充")
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
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	if t_post.UserId == user.Id {
		RenderHTML(w, &t_post, "layout", "navbar.private", "post.edit")
	} else {
		util.Debug(" Cannot edit other user's post", err)
		Report(w, r, "茶博士提示，目前仅能补充自己的回复")
	}

}
