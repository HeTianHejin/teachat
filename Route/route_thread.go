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

// NewDraftThreadHandle()
func NewDraftThreadHandle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		NewDraftThreadGet(w, r)
	case "POST":
		NewDraftThreadPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/thread/new?id=
// GET /v1/thread/new?postid=
// 处理提交的新茶议草稿，索要表单请求
func NewDraftThreadGet(w http.ResponseWriter, r *http.Request) {
	//尝试从http请求中读取用户会话信息
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话读取当前用户的信息
	s_u, s_d_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		util.ScaldingTea(util.LogError(err), "cannot fetch s_u s_teams given session")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	//把系统默认家庭资料加入s_survival_families
	s_survival_families = append(s_survival_families, DefaultFamily)
	//把系统默认团队资料加入s_survival_teams
	s_survival_teams = append(s_survival_teams, FreelancerTeam)

	var tD data.ThreadDetail

	// 读取用户提交的茶台参数
	vals := r.URL.Query()
	uuid := vals.Get("id")

	if uuid == "" {
		uuid = vals.Get("postid")
		//读取品味资料
		post := data.Post{Uuid: uuid}
		if err = post.Get(); err != nil {
			util.ScaldingTea(util.LogError(err), uuid, " Cannot read post given uuid")
			Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
			return
		}
		tD.QuotePost = post

		tD.QuotePostAuthor, err = tD.QuotePost.User()
		if err != nil {
			util.ScaldingTea(util.LogError(err), tD.QuotePost.Id, " Cannot read post user")
			Report(w, r, "你好，茶博士失魂鱼，松影一庭见鹤，梨花满地不闻莺。请稍后再试。")
			return
		}

		tD.QuotePostAuthorTeam, err = data.GetTeam(tD.QuotePost.TeamId)
		if err != nil {
			util.ScaldingTea(util.LogError(err), tD.QuotePost.TeamId, " Cannot read post team")
			Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺。请稍后再试。")
			return
		}
		tD.ThreadBean.Thread.PostId = tD.QuotePost.Id

		//读取茶台资料
		tD.QuoteProject, err = tD.QuotePost.Project()
		if err != nil {
			util.ScaldingTea(util.LogError(err), uuid, " Cannot read project given uuid")
			Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
			return
		}
	} else {
		tD.ThreadBean.Thread.PostId = 0
		//读取茶台资料
		pr := data.Project{Uuid: uuid}
		if err = pr.GetByUuid(); err != nil {
			util.ScaldingTea(util.LogError(err), uuid, " Cannot read project given uuid")
			Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
			return
		}
		tD.QuoteProject = pr
	}
	//检查project.Class=1 or 2,否则属于未经 友邻蒙评 通过的草稿，不允许查看
	if tD.QuoteProject.Class != 1 && tD.QuoteProject.Class != 2 {
		util.ScaldingTea(util.LogError(err), s_u.Id, "欲查看未经友邻蒙评通过的茶台资料被阻止")
		Report(w, r, "你好，荡昏寐，饮之以茶。请稍后再试。")
		return
	}

	// 填写页面数据

	tD.QuoteProjectAuthor, err = tD.QuoteProject.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), tD.QuoteProject.Id, " Cannot read project user")
		Report(w, r, "你好，霁月难逢，彩云易散。请稍后再试。")
		return
	}
	tD.QuoteProjectAuthorFamily, err = GetFamilyByFamilyId(tD.QuoteProject.FamilyId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), tD.QuoteProject.Id, " Cannot read project user family")
		Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见，梨花满地不闻莺。请稍后再试。")
		return
	}
	tD.QuoteProjectAuthorTeam, err = data.GetTeam(tD.QuoteProject.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), tD.QuoteProject.TeamId, " Cannot read project team")
		Report(w, r, "你好，茶博士失魂鱼，松影一庭惟见鹤，梨花满地不闻莺。请稍后再试。")
		return
	}
	tD.SessUser = s_u
	tD.SessUserDefaultFamily = s_d_family
	tD.SessUserSurvivalFamilies = s_survival_families
	tD.SessUserDefaultTeam = s_default_team
	tD.SessUserSurvivalTeams = s_survival_teams
	tD.SessUserDefaultPlace = s_default_place
	tD.SessUserBindPlaces = s_places
	// 给请求用户返回新建完整版茶议表单页面
	RenderHTML(w, &tD, "layout", "navbar.private", "thread.new")
}

// POST /v1/thread/draft
// 处理提交的简化版新茶议草稿，待邻座蒙评后转为正式茶议
func NewDraftThreadPost(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot parse form")
		Report(w, r, "你好，闪电茶博士为你极速服务但是迷路了，未能找到你想要的资料。")
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), sess.Email, " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//读取表单数据
	ty, err := strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		util.ScaldingTea(util.LogError(err), ty, "Failed to convert type to int")
		Report(w, r, "你好，闺中女儿惜春暮，愁绪满怀无释处。")
		return
	}
	// 检查ty值是否 0、1
	switch ty {
	case 0, 1:
		break
	default:
		util.ScaldingTea("Invalid thread type value")
		Report(w, r, "你好，闺中女儿惜春暮，愁绪满怀无释处。")
		return
	}
	body := r.PostFormValue("topic")
	title := r.PostFormValue("title")
	project_id, err := strconv.Atoi(r.PostFormValue("project_id"))
	if err != nil {
		util.ScaldingTea(util.LogError(err), project_id, "Failed to convert project_id to int")
		Report(w, r, "你好，闪电茶博士极速查找茶台中，请确认后再试。")
		return
	}
	post_id, err := strconv.Atoi(r.PostFormValue("post_id"))
	if err != nil {
		util.ScaldingTea(util.LogError(err), project_id, "Failed to convert post_id to int")
		Report(w, r, "你好，闪电茶博士极速服务，任然无法识别提交的品味资料，请确认后再试。")
		return
	}
	/// check submit post_id is valid, if not 0 表示属于“议中议”
	post := data.Post{Id: post_id}
	proj := data.Project{Id: project_id}
	if post_id > 0 {
		if err = post.Get(); err != nil {
			util.ScaldingTea(util.LogError(err), post_id, " Cannot get post given id")
			Report(w, r, "你好，闪电茶博士极速服务，然而无法识别提交的品味资料，请确认后再试。")
			return
		}

		// 检查提及的post和project是否匹配
		t_proj, err := post.Project()
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot get project given post_id")
			Report(w, r, "你好，闪电茶博士极速服务后居然说这个茶台有一些问题，请确认后再试一次")
			return
		}
		if t_proj.Id != project_id {
			util.ScaldingTea(project_id, "post_id and project_id do not match")
			Report(w, r, "你好，闪电茶博士极速服务后居然说这个茶台有一点点问题，请确认后再试一次。")
			return
		}
	}
	//检查该茶台是否存在，而且状态不是草台状态
	if err = proj.Get(); err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get project")
		Report(w, r, "你好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	if proj.Class == 10 || proj.Class == 20 {
		util.ScaldingTea(s_u.Email, "试图访问未蒙评审核的茶台被阻止。")
		Report(w, r, "你好，茶博士竟然说该茶台尚未启用，请确认后再试一次。")
		return
	}

	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"

	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.ScaldingTea(util.LogError(err), "Failed to convert class to int")
		Report(w, r, "你好，此地无这个茶团，请确认后再试。")
		return
	}
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.ScaldingTea(util.LogError(err), "Failed to convert class to int")
		Report(w, r, "你好，此地无这个茶团，请确认后再试。")
		return
	}

	//提交的茶团id,是team.id
	if team_id != 0 {
		// check submit team_id is valid
		_, err = data.GetMemberByTeamIdUserId(team_id, s_u.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				//util.ScaldingTea(util.LogError(err), " Cannot get team member by team id and user id")
				Report(w, r, "你好，茶博士认为您不是这个茶团的成员，请确认后再试。")
				return
			}
			util.ScaldingTea(util.LogError(err), " Cannot get team member by team id and user id")
			Report(w, r, "你好，茶团成员资格检查失败，请确认后再试。")
			return
		}
	}
	//提交的茶团id,是family.id
	if family_id != 0 {
		// check submit family_id is valid
		family := data.Family{
			Id: family_id,
		}
		is_member, err := family.IsMember(s_u.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				//util.ScaldingTea(util.LogError(err), " Cannot get family member by family id and user id")
				Report(w, r, "你好，茶博士认为您不是这个茶团的成员，请确认后再试。")
				return
			}
			util.ScaldingTea(util.LogError(err), " Cannot get family member by family id and user id")
			Report(w, r, "你好，家庭成员资格检查失败，请确认后再试。")
			return
		}
		if !is_member {
			util.ScaldingTea(util.LogError(err), " Cannot get family member by family id and user id")
			Report(w, r, "你好，家庭成员资格检查失败，请确认后再试。")
			return
		}
	}

	// 如果茶台class=1，存为开放式茶议草稿，
	// 如果茶台class=2， 存为封闭式茶议草稿
	if proj.Class == 1 || proj.Class == 2 {
		//检测一下title是否不为空，而且中文字数<24,topic不为空，而且中文字数<456
		if CnStrLen(title) < 1 {
			Report(w, r, "你好，茶博士竟然说该茶议标题为空，请确认后再试一次。")
			return
		}
		if CnStrLen(title) > 36 {
			Report(w, r, "你好，茶博士竟然说该茶议标题过长，请确认后再试一次。")
			return
		}
		if CnStrLen(body) < 1 {
			Report(w, r, "你好，茶博士竟然说该茶议内容为空，请确认后再试一次。")
			return
		} else if CnStrLen(body) < 17 {
			Report(w, r, "你好，茶博士竟然说该茶议内容太短，请确认后再试一次。")
			return
		}
		if CnStrLen(body) > 456 {
			Report(w, r, "你好，茶博士小声说茶棚的小纸条只能写456字，请确认后再试一次。")
			return
		}

		//保存新茶议草稿
		draft_thread := data.DraftThread{
			UserId:    s_u.Id,
			ProjectId: project_id,
			Title:     title,
			Body:      body,
			Class:     proj.Class,
			Type:      ty,
			PostId:    post_id,
			TeamId:    team_id,
			IsPrivate: is_private,
			FamilyId:  family_id,
		}
		if err = draft_thread.Create(); err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot create thread draft")
			Report(w, r, "你好，茶博士没有墨水了，未能保存新茶议草稿。")
			return
		}
		// 创建一条友邻蒙评,是否接纳 新茶的记录
		aO := data.AcceptObject{
			ObjectId:   draft_thread.Id,
			ObjectType: 3,
		}
		if err = aO.Create(); err != nil {
			util.ScaldingTea(util.LogError(err), "Cannot create accept_object")
			Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
			return
		}
		// 发送蒙评请求消息给两个在线用户
		//构造消息
		mess := data.AcceptMessage{
			FromUserId:     1,
			Title:          "新茶语邻座评审邀请",
			Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
			AcceptObjectId: aO.Id,
		}
		//发送消息
		if err = TwoAcceptMessagesSendExceptUserId(s_u.Id, mess); err != nil {
			Report(w, r, "你好，早知日后闲争气，岂肯今朝错读书！未能发送蒙评请求消息。")
			return
		}

		// 提示用户草稿保存成功
		t := fmt.Sprintf("你好，你在“ %s ”茶台发布的茶议已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", proj.Title)
		// 提示用户草稿保存成功
		Report(w, r, t)
		return
	}
	//出现非法的class值
	Report(w, r, "你好，糊里糊涂的茶博士竟然说该茶台坐满了外星人，请确认后再试一次。")

}

// GET /v1/thread/detail?id=
// 显示需求uuid茶议（议题）的详细信息，包括品味（回复帖子）和记录品味的表格
func ThreadDetail(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	uuid := vals.Get("id")

	if uuid == "" {
		Report(w, r, "你好，茶博士看不透您提交的茶议id。")
		return
	}

	// 准备一个空白的表
	var tD data.ThreadDetail

	// 读取茶议内容以填空
	thread, err := data.ThreadByUUID(uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Report(w, r, "你好，茶博士竟然说该茶议不存在，请确认后再试一次。")
			return
		}
		util.ScaldingTea(util.LogError(err), " Cannot read thread given uuid", uuid)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议。")
		return
	}

	//读取茶台资料
	tD.QuoteProject, err = thread.Project()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			//	util.ScaldingTea(util.LogError(err), " Cannot read project given uuid", uuid)
			Report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个茶台不存在。")
			return
		}
		util.ScaldingTea(util.LogError(err), " Cannot read project")
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取茶台资料。")
		return
	}

	tD.QuoteProjectAuthor, err = tD.QuoteProject.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project author")
		Report(w, r, "你好，静夜不眠因酒渴，沉烟重拨索烹茶。未能读取茶台资料。")
		return
	}
	tD.QuoteProjectAuthorFamily, err = GetFamilyByFamilyId(tD.QuoteProject.FamilyId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project author family")
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取茶台资料。")
		return
	}
	tD.QuoteProjectAuthorTeam, err = data.GetTeam(tD.QuoteProject.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project author team")
		Report(w, r, "你好，绛芸轩里绝喧哗，桂魄流光浸茜纱。未能读取茶台资料。")
		return
	}

	//读取茶围资料
	tD.QuoteObjective, err = tD.QuoteProject.Objective()
	if err != nil {
		util.ScaldingTea(util.LogError(err), tD.QuoteProject.Id, " Cannot read objective given project")
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}

	//检查品味的类型
	if thread.PostId != 0 {
		// 说明这是一个附加类型的,针对某个post发表的茶议(chat-in-chat，讲开又讲，延伸话题)
		post := data.Post{Id: thread.PostId}
		if err = post.Get(); err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot read post given post_id")
			Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取品味资料。")
			return
		}
		tD.QuotePost = post

		// 截短body
		tD.QuotePost.Body = Substr(tD.QuotePost.Body, 66)
		tD.QuotePostAuthor, err = tD.QuotePost.User()
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot read post author")
			Report(w, r, "你好，呜咽一声犹未了，落花满地鸟惊飞。未能读取品味资料。")
			return
		}
		tD.QuotePostAuthorFamily, err = GetFamilyByFamilyId(tD.QuotePost.FamilyId)
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot read post author family")
			Report(w, r, "你好，呜咽一声犹未了，落花满地鸟惊飞。未能读取品味资料。")
			return
		}
		tD.QuotePostAuthorTeam, err = data.GetTeam(tD.QuotePost.TeamId)
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot read post author team")
			Report(w, r, "你好，花谢花飞飞满天，红消香断有谁怜？未能读取品味资料。")
			return
		}

	} else {
		// 是一个普通的茶议
		// 截短body
		tD.QuoteProject.Body = Substr(tD.QuoteProject.Body, 66)

	}

	// 读取茶议资料荚
	tD.ThreadBean, err = FetchThreadBean(thread)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read threadBean")
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议资料荚。")
		return
	}
	tD.NumSupport = thread.NumSupport()
	tD.NumOppose = thread.NumOppose()
	//品味中颔首的合计得分与总得分的比值，取整数，用于客户端页面进度条设置，正反双方进展形势对比
	// n1, err := tD.ThreadBean.Thread.PostsScoreSupport()
	// if n1 != 0 && err != nil {

	// 	util.PanicTea(util.LogError(err), " Cannot get posts score support")
	// 	Report(w, r, "你好，莫失莫忘，仙寿永昌，有些资料被黑风怪瓜州了。")
	// 	return

	// }
	// n2, err := tD.ThreadBean.Thread.PostsScore()
	// if n2 != 0 && err != nil {

	// 	util.PanicTea(util.LogError(err), " Cannot get posts score oppose")
	// 	Report(w, r, "你好，莫失莫忘，仙寿永昌，有些资料,被黑风怪瓜州了。")
	// 	return

	// }
	n1 := 60  //测试临时值
	n2 := 120 //测试临时值
	tD.ProgressSupport = ProgressRound(n1, n2)
	tD.ProgressOppose = 100 - tD.ProgressSupport
	// 读取全部回复帖子（品味）
	post_slice, err := tD.ThreadBean.Thread.Posts()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read posts")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	tD.PostBeanSlice, err = FetchPostBeanSlice(post_slice)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read posts")
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	// 读取会话
	s, err := Session(r)

	if err != nil {
		// 游客
		tD.IsAuthor = false
		// 检查茶议的级别状态
		if tD.ThreadBean.Thread.Class == 1 || tD.ThreadBean.Thread.Class == 2 {
			//记录茶议被点击数
			//tD.ThreadBean.Thread.AddHitCount()
			// 填写页面数据
			tD.ThreadBean.Thread.PageData.IsAuthor = false
			tD.IsGuest = true
			tD.IsInput = false
			tD.IsAdmin = false
			tD.IsMaster = false

			tD.SessUser = data.User{
				Id:   0,
				Name: "游客",
				// 用户足迹
				Footprint: r.URL.Path,
				Query:     r.URL.RawQuery,
			}
			//迭代postSlice,标记非品味作者
			for i := range tD.PostBeanSlice {
				tD.PostBeanSlice[i].Post.PageData.IsAuthor = false
			}

			//show the thread and the posts展示页面
			RenderHTML(w, &tD, "layout", "navbar.public", "thread.detail")
			return
		} else {
			//非法访问未开放的话题？
			util.ScaldingTea(util.LogError(err), " 试图访问未公开的thread", uuid)
			Report(w, r, "茶水温度太高了，不适合品味，请稍后再试。")
			return
		}
	} else {
		//用户是登录状态,可以访问1和2级茶议
		tD.IsGuest = false

		if tD.ThreadBean.Thread.Class == 1 || tD.ThreadBean.Thread.Class == 2 {
			//从会话查获当前浏览用户资料荚
			s_u, s_d_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchUserRelatedData(s)
			if err != nil {
				util.ScaldingTea(util.LogError(err), " Cannot get user-related data from session", s_u.Id)
				Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
				return
			}
			// 用户足迹
			s_u.Footprint = r.URL.Path
			s_u.Query = r.URL.RawQuery

			tD.SessUser = s_u

			tD.SessUserDefaultFamily = s_d_family
			tD.SessUserSurvivalFamilies = s_survival_families

			tD.SessUserDefaultTeam = s_default_team
			tD.SessUserSurvivalTeams = s_survival_teams

			tD.SessUserDefaultPlace = s_default_place
			tD.SessUserBindPlaces = s_places

			// 检测是否茶议作者
			if s_u.Id == tD.ThreadBean.Thread.UserId {
				// 是茶议作者！
				tD.IsAuthor = true
				ob_team, err := data.GetTeam(tD.QuoteProject.TeamId)
				if err != nil {
					util.ScaldingTea(util.LogError(err), " Cannot get team given team_id", tD.QuoteProject.TeamId)
					Report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个团队不存在。")
					return
				}
				tD.QuoteObjectiveAuthorTeam = ob_team
				is_admin, err := ob_team.IsMember(s_u.Id)
				if err != nil {
					util.ScaldingTea(util.LogError(err), " Cannot check team membership", tD.QuoteObjectiveAuthorTeam.Id)
					Report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个团队不存在。")
					return
				}
				tD.IsAdmin = is_admin

				pr_team, err := data.GetTeam(tD.QuoteProject.TeamId)
				if err != nil {
					util.ScaldingTea(util.LogError(err), " Cannot get team given team_id", tD.QuoteProject.TeamId)
					Report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个团队不存在。")
					return
				}
				tD.QuoteProjectAuthorTeam = pr_team
				is_master, err := pr_team.IsMember(s_u.Id)
				if err != nil {
					util.ScaldingTea(util.LogError(err), " Cannot check team membership", tD.QuoteProjectAuthorTeam.Id)
					Report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个团队不存在。")
					return
				}
				tD.IsMaster = is_master
				// 填写页面数据
				tD.ThreadBean.Thread.PageData.IsAuthor = true
				// 提议作者不能自评品味，王婆卖瓜也不行？！
				tD.IsInput = false
				//点击数+1
				//tD.ThreadBean.Thread.AddHitCount()
				//记录用户阅读该帖子一次
				data.SaveReadedUserId(tD.ThreadBean.Thread.Id, s_u.Id)
				//迭代PostSlice，把其PageData.IsAuthor设置为false，页面渲染时检测布局用
				for i := range tD.PostBeanSlice {
					tD.PostBeanSlice[i].Post.PageData.IsAuthor = false
				}

				//show the thread and the posts展示页面
				RenderHTML(w, &tD, "layout", "navbar.private", "thread.detail")
				return
			} else {
				//不是茶议作者
				//记录用户阅读该帖子一次
				data.SaveReadedUserId(tD.ThreadBean.Thread.Id, s_u.Id)
				//记录茶议被点击数
				//tD.ThreadBean.Thread.AddHitCount()
				// 填写页面数据

				tD.ThreadBean.Thread.PageData.IsAuthor = false

				//检查是否封闭式茶台
				if tD.QuoteProject.Class == 2 {
					//是封闭式茶台，需要检查当前用户身份是否受邀请茶团的成员，以决定是否允许发言
					ok, err := tD.QuoteProject.IsInvitedMember(s_u.Id)
					if err != nil {
						Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？你真的是受邀请茶团成员吗？")
						return
					}
					if ok {
						// 当前用户是��话会��请��队成员，可以新开茶议
						tD.IsInput = true
					} else {
						// 当前用户不是��话会��请��队成员，不能新开茶议
						tD.IsInput = false
					}
				} else {
					// 是开放式茶台，任何人都可以发布品味
					tD.IsInput = true
				}

				// 如果当前用户已经品味过了，则关闭撰写输入面板(每人仅可表态一次)
				// 用于页面判断是否显示品味POST（回复）撰写面板
				for i := range tD.PostBeanSlice {
					if tD.PostBeanSlice[i].Post.UserId == s_u.Id {
						tD.IsInput = false
						tD.IsPostExist = true
						break
					}
				}

				// 检测是否其中某一个Post品味作者
				for i := range tD.PostBeanSlice {
					if tD.PostBeanSlice[i].Post.UserId == s_u.Id {
						tD.PostBeanSlice[i].Post.PageData.IsAuthor = true
						break
					} else {
						tD.PostBeanSlice[i].Post.PageData.IsAuthor = false
					}
				}

				//展示茶议详情
				RenderHTML(w, &tD, "layout", "navbar.private", "thread.detail")
				return
			}
		} else if tD.ThreadBean.Thread.Class == 0 {
			//茶议的等级发生了变化，需要重新进行邻桌评估
			Report(w, r, "你好，茶议加水后出现了神迹！请耐心等待邻桌来推荐。")
			return
		} else {
			//访问未开放等级的话题？
			Report(w, r, "你好，外星人出没注意！请确认或者联系管理员确认稍后再试。")
			return
		}
	}

}

// POST /v1/thread/approve
// 茶台管理，决定采纳（认可）某个goodidea（thread），被采纳的thread被标识为“已采纳”approved
func ThreadApprove(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot parse form")
		Report(w, r, "你好，闪电茶博士为了提高服务速度而迷路了，未能找到你想要的茶台。")
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//读取表单数据
	uuid := r.PostFormValue("id")
	if uuid == "" {
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取茶议资料，请稍后再试。")
		return
	}

	//读取提及的茶议资料
	thread, err := data.ThreadByUUID(uuid)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read thread given uuid", uuid)
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取茶议资料，请稍后再试。")
		return
	}
	proj, err := thread.Project()
	if err != nil {
		util.ScaldingTea(util.LogError(err), thread.Id, " Cannot read project given thread_id")
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取茶台资料，请稍后再试。")
		return
	}

	//检查用户是否有权限处理这个请求
	team, err := data.GetTeam(proj.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), proj.TeamId, " Cannot get team given id")
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取团队资料，请稍后再试。")
		return
	}
	//读取支持team核心成员资料
	team_members, err := team.CoreMembers()
	if err != nil {
		util.ScaldingTea(util.LogError(err), team.Id, " Cannot get team core members given team")
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取团队管理成员资料，请稍后再试。")
		return
	}

	ok := false
	//检查s_u 是否是茶台作者
	if proj.UserId == s_u.Id {
		//是台主，可以处理请求
		ok = true
	} else {
		//不是台主，检查是否是team核心成员
		for _, tm := range team_members {
			if tm.UserId == s_u.Id {
				//是team核心成员，可以处理请求
				ok = true
				break
			}
		}
	}
	if ok {
		//处理采纳茶议请求
		thread_approved := data.ThreadApproved{
			ThreadId:  thread.Id,
			ProjectId: proj.Id,
			UserId:    s_u.Id,
		}
		if err = thread_approved.Create(); err != nil {
			util.ScaldingTea(util.LogError(err), thread.Id, " Cannot create thread approved")
			Report(w, r, "你好，闪电茶博士极速服务中，未能处理你的请求，请稍后再试。")
			return
		}

		//采纳（认可好主意）成功
		Report(w, r, "你好，闪电茶博士极速服务，采纳该主意操作成功，请返回，刷新页面查看。")
		return
	} else {
		//没有权限处理请求
		Report(w, r, "你好，闪电茶博士极速服务，火星保安竟然说你没有权限处理该请求。")
		return
	}

}

// GET /thread/edit
// 打开指定的茶议（议程）追加（补充内容）页面，
// 这是为了便利作者为自己的投票立场附加解释。
// 规则是只能补充剩余字数,
// 不能超过max_word，不能修改已记录的茶议内容，
// 不能修改茶议等级，
// 不能修改茶议标题，
// 不能修改茶议创建时间，
// 不能修改茶议创建者，
// 不能修改茶议点击数，
// 不能修改茶议回复数，
// 不能修改茶议支持数，
// 不能修改茶议反对数，
// 不能修改茶议是否开放式/封闭式，
// 不能修改茶议是否删除，
func EditThread(w http.ResponseWriter, r *http.Request) {
	var thDPD data.ThreadDetail
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	} else {
		// 读取当前访问用户资料
		sUser, err := sess.User()
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot get user from session")
			Report(w, r, "你好，茶博士失魂鱼，未能读取会话用户资料。")
			return
		}
		vals := r.URL.Query()
		uuid := vals.Get("id")
		thDPD.ThreadBean.Thread, err = data.ThreadByUUID(uuid)
		if err != nil {
			util.ScaldingTea("Cannot not read thread")
			Report(w, r, "茶博士失魂鱼，未能读取茶议资料，请稍后再试。")
			return
		}
		if thDPD.ThreadBean.Thread.UserId == sUser.Id {
			// 是作者，可以加水（补充内容）
			thDPD.ThreadBean.Thread.PageData.IsAuthor = true
			thDPD.SessUser = sUser
			RenderHTML(w, &thDPD, "layout", "navbar.private", "thread.edit")
			return
		}
		//不是作者，不能加水
		util.ScaldingTea("Cannot edit other user's thread")
		Report(w, r, "茶博士提示，目前仅能给自己的茶杯加水呢，补充说明自己的茶议貌似是合理的。")
		return

	}

}

// POST /v1/thread/update
// Update the thread 更新茶议内容
func UpdateThread(w http.ResponseWriter, r *http.Request) {
	// 测试时不启用追加方法？

	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	} else {
		err = r.ParseForm()
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot parse form")
			return
		}
		user, err := sess.User()
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot get user from session")
			Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶议。")
			return
		}
		uuid := r.PostFormValue("uuid")
		//title := r.PostFormValue("title")
		topi := r.PostFormValue("additional")
		//根据用户提供的uuid读取指定茶议
		thread, err := data.ThreadByUUID(uuid)
		if err != nil {
			util.ScaldingTea(util.LogError(err), " Cannot read thread by uuid")
			Report(w, r, "茶博士失魂鱼，未能读取专属茶议，请稍后再试。")
			return
		}
		//核对一下用户身份
		if thread.UserId == user.Id {
			//检查topi内容是否中文字数>17,并且thread.Topic总字数<456,如果是则可以补充内容
			if CnStrLen(topi) >= 17 && CnStrLen(thread.Body+topi) < 456 {
				thread.Body += topi
			} else {
				util.ScaldingTea("Cannot update thread")
				Report(w, r, "闪电茶博士居然说字太少或者超过456字的茶议，无法记录，请确认后再试。")
				return
			}
			// 修改过的茶议,重置class=0,表示草稿状态，
			thread.Class = 0
			//许可修改自己的茶议
			if err := thread.UpdateTopicAndClass(thread.Body, thread.Class); err != nil {
				util.ScaldingTea(util.LogError(err), " Cannot update thread")
				Report(w, r, "茶博士失魂鱼，未能更新专属茶议，请稍后再试。")
				return
			}
			url := fmt.Sprint("/v1/thread/detail?id=", uuid)
			http.Redirect(w, r, url, http.StatusFound)
			return
		} else {
			//阻止修改别人的茶议
			util.ScaldingTea("Cannot edit other user's thread")
			Report(w, r, "茶博士提示，粗鲁的茶博士竟然说，仅能对自己的茶杯加水（追加内容）。")
			return
		}
	}
}
