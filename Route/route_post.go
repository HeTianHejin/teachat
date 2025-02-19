package route

import (
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// 加水 ，修改回复post的处理器
func HandleEditPost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		EditPost(w, r)
	case "POST":
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
	post, err := data.GetPostByUuid(uuid)
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get post detail")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	post_bean, err := FetchPostBean(post)
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get post bean given post")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	pD.PostBean = post_bean
	// 读取此品味引用的茶议
	pD.QuoteThread, err = post.Thread()
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get thread given post")
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议资料。")
		return
	}
	// 截短此引用的茶议内容以方便展示
	pD.QuoteThread.Body = Substr(pD.QuoteThread.Body, 66)
	// 此品味针对的茶议作者资料
	pD.QuoteThreadAuthor, err = pD.QuoteThread.User()
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get thread author given post")
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议主人资料。")
		return
	}
	// 引用的茶议作者发帖时候选择的茶团
	pD.QuoteThreadAuthorTeam, err = data.GetTeam(pD.QuoteThread.TeamId)
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get quote-thread-author-default-team given post")
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议主人资料。")
		return
	}
	// 读取全部针对此品味的茶议
	thread_slice, err := post.Threads()
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get thread_slice given post")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	pD.ThreadBeanSlice, err = FetchThreadBeanSlice(thread_slice)
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get thread_bean_slice given thread_slice")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}

	pD.QuoteProject, err = pD.QuoteThread.Project()
	if err != nil {
		util.Warning(util.LogError(err), pD.QuoteThread.Id, " Cannot get project given thread")
		Report(w, r, "你好，������失������，未能读取专��资料。")
		return
	}
	pD.QuoteObjective, err = pD.QuoteProject.Objective()
	if err != nil {
		util.Warning(util.LogError(err), pD.QuoteProject.Id, " Cannot get objective given project")
		Report(w, r, "你好，������失������，未能读取专��资料。")
		return
	}

	// 读取会话
	// 检测pageData.ThreadSlice数量是否超过一打dozen
	if len(thread_slice) > 12 {
		pD.IsOverTwelve = true
	} else {
		pD.IsOverTwelve = false
	}
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
	s_u, _ := s.User()
	pD.SessUser = s_u
	// 从会话查获当前浏览用户资料荚
	s_u, _, _, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get user-related data from session")
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	pD.SessUser = s_u
	pD.IsGuest = false
	pD.IsInput = true
	pD.SessUserDefaultTeam = s_default_team
	pD.SessUserSurvivalTeams = s_survival_teams
	// 默认地点
	pD.SessUserDefaultPlace = s_default_place
	// 全部绑定地点
	pD.SessUserBindPlaces = s_places

	// 当前会话用户是否此品味作者？
	if s_u.Id == post.UserId {
		pD.IsAuthor = true
	} else {
		pD.IsAuthor = false
	}

	RenderHTML(w, &pD, "layout", "navbar.private", "post.detail")

}

// POST /v1/post/draft
// Create the post 创建品味（跟帖/回复）草稿
func NewPostDraft(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Warning(util.LogError(err), " Cannot parse form")
		Report(w, r, "你好，茶博士摸摸头，竟然说今天电脑去热带海岛潜水了。")
		return
	}

	s_u, err := s.User()
	if err != nil {
		util.Warning(util.LogError(err), " Cannot get user from session")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}

	//读取用户表态,立场是支持（true）或者反对(false)
	var attitude bool
	a := r.PostFormValue("attitude")
	switch a {
	case "true":
		attitude = true
	case "false":
		attitude = false
	default:
		Report(w, r, "你好，茶博士失魂鱼，未能读懂您的表态内容。")
		return
	}

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
	tid := r.PostFormValue("team_id")
	//change team_id to int
	team_id, err := strconv.Atoi(tid)
	if err != nil {
		Report(w, r, "一年三百六十日，风刀霜剑严相逼")
		return
	}
	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"

	if !is_private {
		//提交的茶团id,是team.id
		// check the given team_id is valid
		_, err = data.GetMemberByTeamIdUserId(team_id, s_u.Id)
		if err != nil {
			Report(w, r, "你好，眼前无路想回头，什么团成员？什么茶话会？请稍后再试。")
			return
		}
	} else {
		//提交的茶团id,是family.id
		// check submit family_id is valid
		family := data.Family{
			Id: team_id,
		}
		is_member, _ := family.IsMember(s_u.Id)
		if !is_member {
			util.Warning(util.LogError(err), " Cannot get family member by family id and user id")
			Report(w, r, "你好，家庭成员资格检查失败，请确认后再试。")
			return
		}
	}

	//  检查茶议所在的茶台属性，
	t_proj, err := thread.Project()
	if err != nil {
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
		return
	}
	var dPost data.DraftPost
	switch t_proj.Class {
	case 1:
		// class=1可以品茶，
		if dPost, err = s_u.CreateDraftPost(thread.Id, team_id, attitude, is_private, body); err != nil {
			Report(w, r, "你好，茶博士摸摸头，记录品味失败。")
			return
		}

	case 2:
		// 当前会话用户是否可以入席品茶？需要看台主指定了那些茶团成员可以品茶
		ok, err := t_proj.IsInvitedMember(s_u.Id)
		if err != nil {
			Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶台资料。")
			return
		}
		if !ok {
			// Cannot have tea
			Report(w, r, "你好，你的大名竟然不在邀请品茶名单上。")
			return
		}

		// Can have tea
		if dPost, err = s_u.CreateDraftPost(thread.Id, team_id, attitude, is_private, body); err != nil {
			Report(w, r, "你好，茶博士摸摸头，竟然说没有墨水，记录品味失败。")
			return
		}

	default:
		// 异常状态的茶台
		Report(w, r, "你好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
		return
	}

	// 创建一条友邻盲评,是否接纳 新茶的记录
	aO := data.AcceptObject{
		ObjectId:   dPost.Id,
		ObjectType: 4,
	}
	if err = aO.Create(); err != nil {
		util.Warning(util.LogError(err), "Cannot create accept_object")
		Report(w, r, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
		return
	}
	// 发送邻座盲评消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
		AcceptObjectId: aO.Id,
	}
	// 发送消息
	if err = TwoAcceptMessagesSendExceptUserId(s_u.Id, mess); err != nil {
		Report(w, r, "你好，茶博士迷路了，未能发送盲评请求消息。")
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
		util.Danger(util.LogError(err), " Cannot parse form")
	}
	//从会话中读取用户资料
	user, err := sess.User()
	if err != nil {
		util.Danger(util.LogError(err), " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	uuid := r.PostFormValue("uuid")
	post, err := data.GetPostByUuid(uuid)
	if err != nil {
		util.Danger(util.LogError(err), " Cannot read given post")
		Report(w, r, "茶博士失魂鱼，未能读取指定表态，请稍后再试。")
		return
	}
	if post.UserId != user.Id {
		util.Danger(util.LogError(err), " Cannot edit other user's post")
		Report(w, r, "茶博士提示，目前仅能补充自己的回复")
		return
	} else {
		//可以补充自己的表态内容
		body := r.PostFormValue("body")
		if body != "" {
			//检查补充内容是否有意义，rune 字数>17,总的post字数<456
			if CnStrLen(body) > 17 && CnStrLen(post.Body)+CnStrLen(body) < 456 {
				post.Body += body
			} else {
				//提示用户总字数或者本次提交补充内容超出字数限制
				Report(w, r, "你好， 粗鲁的茶博士竟然说字数满了，纸条写不下您的品味。")
				return
			}
			err = post.UpdateBody(body)
			if err != nil {
				util.Danger(util.LogError(err), " Cannot update post")
				Report(w, r, "茶博士失魂鱼，未能更新专属资料，请稍后再试。")
				return
			}
			thread, err := data.GetThreadById(post.ThreadId)
			if err != nil {
				util.Danger(util.LogError(err), " Cannot read thread")
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
	} else {

		user, err := sess.User()
		if err != nil {
			util.Danger(util.LogError(err), " Cannot get user from session")
			http.Redirect(w, r, "/v1/login", http.StatusFound)
			return
		}
		vals := r.URL.Query()
		uuid := vals.Get("id")
		post, err := data.GetPostByUuid(uuid)
		if err != nil {
			util.Danger(util.LogError(err), " Cannot read post")
			Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料，请稍后再试。")
			return
		}
		if post.UserId == user.Id {
			RenderHTML(w, &post, "layout", "navbar.private", "post.edit")
		} else {
			util.Danger(util.LogError(err), " Cannot edit other user's post")
			Report(w, r, "茶博士提示，目前仅能补充自己的回复")
		}

	}

}
