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
	case http.MethodGet:
		//NewDraftThreadGet(w, r)
	case http.MethodPost:
		NewDraftThreadPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// POST /v1/thread/draft
// 处理提交的新茶议草稿，待邻座蒙评后转为正式茶议
func NewDraftThreadPost(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		Report(w, r, "你好，茶博士迷路了，未能找到你想要的资料。")
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", sess.Email, err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	//读取茶议表达
	thre_type, err := strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		util.Debug("Failed to convert type to int", thre_type, err)
		Report(w, r, "你好，闺中女儿惜春暮，愁绪满怀无释处。")
		return
	}

	// 检查ty值是否合法
	switch thre_type {
	case data.ThreadTypeIthink, data.ThreadTypeIdea:
		break
	default:
		util.Debug("Invalid thread type value", err)
		Report(w, r, "你好，闺中女儿惜春暮，愁绪满怀无释处。")
		return
	}
	body := r.PostFormValue("topic")
	title := r.PostFormValue("title")
	project_id, err := strconv.Atoi(r.PostFormValue("project_id"))
	if err != nil {
		util.Debug("Failed to convert project_id to int", project_id, err)
		Report(w, r, "你好，闪电茶博士极速查找茶台中，请确认后再试。")
		return
	}
	post_id, err := strconv.Atoi(r.PostFormValue("post_id"))
	if err != nil {
		util.Debug("Failed to convert post_id to int", project_id, err)
		Report(w, r, "你好，闪电茶博士极速服务，任然无法识别提交的品味资料，请确认后再试。")
		return
	}
	/// check submit post_id is valid, if not 0 表示属于“议中议”
	post := data.Post{Id: post_id}
	proj := data.Project{Id: project_id}
	//检查该茶台是否存在，而且状态不是待友邻蒙评审查草台状态
	if err = proj.Get(); err != nil {
		util.Debug(" Cannot get project", err)
		Report(w, r, "你好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	if proj.Class != data.ClassOpenTeaTable && proj.Class != data.ClassClosedTeaTable {
		//util.Debug("试图访问未蒙评审核的茶台被阻止。", s_u.Email, err)
		Report(w, r, "你好，茶博士竟然说该茶台尚未启用，请确认后再试一次。")
		return
	}
	if post_id > 0 {
		if err = post.Get(); err != nil {
			util.Debug(" Cannot get post given id", post_id, err)
			Report(w, r, "你好，闪电茶博士极速服务，然而无法识别提交的品味资料，请确认后再试。")
			return
		}
		test_proj, err := post.Project()
		if err != nil {
			util.Debug(" Cannot get post given id", post_id, err)
			Report(w, r, "你好，闪电茶博士极速服务，然而无法识别提交的品味资料，请确认后再试。")
			return
		}
		// 检查提及的post和project是否匹配
		if proj.Id != test_proj.Id {
			util.Debug(project_id, "post_id and project_id do not match")
			Report(w, r, "你好，茶博士居然说这个茶台有一点点问题，请确认后再试一次。")
			return
		}
	}

	// 检查茶议（thread）创建权限
	if ok := checkCreateThreadPermission(proj, s_u.Id, w, r); !ok {
		Report(w, r, "你好，茶博士居然说,陛下您的大名竟然不在邀请名单上，请确认后再试一次。")
		return
	}

	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"

	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		Report(w, r, "你好，此地无这个茶团，请确认后再试。")
		return
	}
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		Report(w, r, "你好，此地无这个茶团，请确认后再试。")
		return
	}

	valid, err := validateTeamAndFamilyParams(is_private, team_id, family_id, s_u.Id, w, r)
	if !valid && err == nil {
		return // 参数不合法，已经处理了错误
	}
	if err != nil {
		// 处理数据库错误
		util.Debug("验证提交的团队和家庭id出现数据库错误", team_id, family_id, err)
		Report(w, r, "你好，成员资格检查失败，请确认后再试。")
		return
	}

	// 如果茶台class=1，存为开放式茶议草稿，
	// 如果茶台class=2， 存为封闭式茶议草稿
	if proj.Class == data.ClassOpenTeaTable || proj.Class == data.ClassClosedTeaTable {
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
			Type:      thre_type,
			PostId:    post_id,
			TeamId:    team_id,
			IsPrivate: is_private,
			FamilyId:  family_id,
		}
		if post_id > 0 {
			draft_thread.Category = data.ThreadCategoryNested
		}
		if err = draft_thread.Create(); err != nil {
			util.Debug(" Cannot create thread draft", err)
			Report(w, r, "你好，茶博士没有墨水了，未能保存新茶议草稿。")
			return
		}
		// 创建一条友邻蒙评,是否接纳 新茶的记录
		aO := data.AcceptObject{
			ObjectId:   draft_thread.Id,
			ObjectType: data.AcceptObjectTypeTeaProposal,
		}
		if err = aO.Create(); err != nil {
			util.Debug("Cannot create accept_object", err)
			Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
			return
		}
		// 发送蒙评请求消息给两个在线用户
		//构造消息
		mess := data.AcceptMessage{
			FromUserId:     data.UserId_SpaceshipCaptain,
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

	// 读取茶议内容
	thread, err := data.GetThreadByUUID(uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Report(w, r, "你好，茶博士竟然说该茶议不存在，请确认后再试一次。")
			return
		}
		util.Debug(" Cannot read thread given uuid", uuid, err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议。")
		return
	}

	//读取所在的茶台资料
	project, err := thread.Project()
	if err != nil {
		util.Debug(" Cannot read project given thread", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶台资料。")
		return
	}
	tD.QuoteProjectBean, err = FetchProjectBean(project)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			//	util.Debug(" Cannot read project given uuid", uuid)
			Report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个茶台不存在。")
			return
		}
		util.Debug(" Cannot read project", err)
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取茶台资料。")
		return
	}

	//读取茶围资料
	objective, err := project.Objective()
	if err != nil {
		util.Debug(" Cannot read objective given project", err)
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取茶围资料。")
		return
	}
	tD.QuoteObjectiveBean, err = FetchObjectiveBean(objective)
	if err != nil {
		util.Debug(" Cannot read objective given project", project.Id, err)
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}

	//检查品味的类型
	if thread.PostId != 0 {
		// 说明这是一个附加类型的,针对某个post发表的茶议(chat-in-chat，讲开又讲，延伸话题)
		post := data.Post{Id: thread.PostId}
		if err = post.Get(); err != nil {
			util.Debug(" Cannot read post given post_id", err)
			Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取品味资料。")
			return
		}
		tD.QuotePostBean, err = FetchPostBean(post)
		if err != nil {
			util.Debug(" Cannot fetch postBean given post", post.Id, err)
			Report(w, r, "你好，茶博士失魂鱼，未能读取品味资料。")
			return
		}

		// 截短body
		tD.QuotePostBean.Post.Body = Substr(tD.QuotePostBean.Post.Body, 66)

	} else {
		// 是一个普通的茶议
		// 截短body
		tD.QuoteProjectBean.Project.Body = Substr(tD.QuoteProjectBean.Project.Body, 66)

	}

	// 读取茶议资料荚
	tD.ThreadBean, err = FetchThreadBean(thread)
	if err != nil {
		util.Debug(" Cannot read threadBean", err)
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

	post_admi_slice, err := tD.ThreadBean.Thread.PostsAdmin()
	if err != nil {
		util.Debug(" Cannot read admin posts", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	tD.PostBeanAdminSlice, err = FetchPostBeanSlice(post_admi_slice)
	if err != nil {
		util.Debug(" Cannot read admin postbean", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}

	// 读取全部普通回复帖子（品味）
	post_slice, err := tD.ThreadBean.Thread.Posts()
	if err != nil {
		util.Debug(" Cannot read posts", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	tD.PostBeanSlice, err = FetchPostBeanSlice(post_slice)
	if err != nil {
		util.Debug(" Cannot read posts", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	// 读取会话
	s, err := Session(r)

	if err != nil {
		// 游客
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
			tD.IsVerifier = false

			tD.SessUser = data.User{
				Id:   0,
				Name: "游客",
				// 用户足迹
				Footprint: r.URL.Path,
				Query:     r.URL.RawQuery,
			}
			//迭代postSlice,标记非品味撰写者
			for i := range tD.PostBeanSlice {
				tD.PostBeanSlice[i].Post.PageData.IsAuthor = false

			}

			//show the thread and the posts展示页面
			RenderHTML(w, &tD, "layout", "navbar.public", "thread.detail", "component_sess_capacity", "component_post_left", "component_post_right", "component_avatar_name_gender")
			return
		} else {
			//非法访问未开放的话题？
			util.Debug(" 试图访问未公开的thread", uuid)
			Report(w, r, "茶水温度太高了，不适合品味，请稍后再试。")
			return
		}
	} else {
		//用户是登录状态
		tD.IsGuest = false

		if tD.ThreadBean.Thread.Class == data.ThreadClassOpen || tD.ThreadBean.Thread.Class == data.ThreadClassClosed {
			//从会话查获当前浏览用户资料荚
			s_u, s_d_family, s_all_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchSessionUserRelatedData(s)
			if err != nil {
				util.Debug(" Cannot get user-related data from session", s_u.Id)
				Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
				return
			}
			// 用户足迹
			s_u.Footprint = r.URL.Path
			s_u.Query = r.URL.RawQuery

			tD.SessUser = s_u

			tD.SessUserDefaultFamily = s_d_family
			tD.SessUserSurvivalFamilies = s_all_families
			tD.SessUserDefaultTeam = s_default_team
			tD.SessUserSurvivalTeams = s_survival_teams
			tD.SessUserDefaultPlace = s_default_place
			tD.SessUserBindPlaces = s_places

			tD.IsAdmin, err = checkObjectiveAdminPermission(&objective, s_u.Id)
			if err != nil {
				util.Debug(" Cannot check objective admin permission", objective.Id, err)
				Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
				return
			}

			tD.IsMaster, err = checkProjectMasterPermission(&project, s_u.Id)
			if err != nil {
				util.Debug(" Cannot check project master permission", project.Id, err)
				Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
				return
			}

			if !tD.IsAdmin && !tD.IsMaster {
				// 检测当前会话茶友是否见证者
				verifier_team := data.Team{Id: data.TeamIdVerifier}
				is_member, err := verifier_team.IsMember(s_u.Id)
				if err != nil {
					util.Debug(" Cannot check team member", err)
					Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
					return
				}
				if is_member {
					tD.IsVerifier = true
				}
			}

			// 检测是否其中某一个Post品味撰写者
			for i := range tD.PostBeanSlice {
				if tD.PostBeanSlice[i].Post.UserId == s_u.Id {
					tD.PostBeanSlice[i].Post.PageData.IsAuthor = true
					break
				} else {
					tD.PostBeanSlice[i].Post.PageData.IsAuthor = false
				}
			}

			if s_u.Id == tD.ThreadBean.Thread.UserId {
				// 是茶议撰写者！
				tD.ThreadBean.Thread.PageData.IsAuthor = true
				tD.IsInput = true
				//点击数+1
				//tD.ThreadBean.Thread.AddHitCount()
				//记录用户阅读该帖子一次
				//data.SaveReadedUserId(tD.ThreadBean.Thread.Id, s_u.Id)

				//展示撰写者视角茶议详情页面
				RenderHTML(w, &tD, "layout", "navbar.private", "thread.detail", "component_sess_capacity", "component_post_left", "component_post_right", "component_avatar_name_gender")
				return
			} else {
				//不是茶议撰写者
				//记录茶议被点击数
				//tD.ThreadBean.Thread.AddHitCount()
				tD.ThreadBean.Thread.PageData.IsAuthor = false

				//检查是否封闭式茶台
				if tD.QuoteProjectBean.Project.Class == data.ClassClosedTeaTable {
					//是封闭式茶台，需要检查当前用户身份是否受邀请茶团的成员，以决定是否允许发言
					ok, err := tD.QuoteProjectBean.Project.IsInvitedMember(s_u.Id)
					if err != nil {
						util.Debug(" Cannot check project invited member", err)
						Report(w, r, "你好，桃李明年能再发，明年闺中知有谁？你真的是受邀请茶团成员吗？")
						return
					}
					if ok {
						// 当前用户是茶围邀请团队成员，可以新开茶议
						tD.IsInput = true
					} else {
						// 当前用户不是茶围邀请团队成员，不能新开茶议
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

				//展示非撰写者视角茶议详情页面
				RenderHTML(w, &tD, "layout", "navbar.private", "thread.detail", "component_sess_capacity", "component_post_left", "component_post_right", "component_avatar_name_gender")
				return
			}
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
		util.Debug(" Cannot parse form", err)
		Report(w, r, "你好，闪电茶博士为了提高服务速度而迷路了，未能找到你想要的茶台。")
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
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
	thread, err := data.GetThreadByUUID(uuid)
	if err != nil {
		util.Debug(" Cannot read thread given uuid", uuid)
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取茶议资料，请稍后再试。")
		return
	}
	proj, err := thread.Project()
	if err != nil {
		util.Debug(thread.Id, " Cannot read project given thread_id")
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取茶台资料，请稍后再试。")
		return
	}
	ob, err := proj.Objective()
	if err != nil {
		util.Debug(proj.Id, " Cannot read objective given project")
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取茶台资料，请稍后再试。")
		return
	}

	//检查用户是否有权限处理这个请求
	admin_team, err := data.GetTeam(ob.TeamId)
	if err != nil {
		util.Debug(proj.TeamId, " Cannot get team given id")
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取团队资料，请稍后再试。")
		return
	}
	//检查是否支持team成员
	is_admin, err := admin_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug(admin_team.Id, " Cannot check team membership")
		Report(w, r, "你好，闪电茶博士极速服务中，未能读取团队资料，请稍后再试。")
		return
	}

	if !is_admin {
		//没有权限处理请求
		Report(w, r, "你好，闪电茶博士极速服务，火星保安竟然说你没有权限处理该请求。")
		return
	}
	//处理采纳茶议请求
	new_thread_approved := data.ThreadApproved{
		ThreadId:  thread.Id,
		ProjectId: proj.Id,
		UserId:    s_u.Id,
	}
	//检查是否已经采纳
	if err = new_thread_approved.GetByThreadId(); err == nil {
		Report(w, r, "你好，闪电茶博士极速服务，该茶议已被采纳，请刷新页面查看。")
		return
	}
	if err = new_thread_approved.Create(); err != nil {
		util.Debug(thread.Id, " Cannot create thread approved")
		Report(w, r, "你好，闪电茶博士极速服务中，未能处理你的请求，请稍后再试。")
		return
	}

	//采纳（认可好主意）成功,跳转茶议详情页面
	http.Redirect(w, r, "/v1/thread/detail?id="+thread.Uuid, http.StatusFound)
	//Report(w, r, "你好，闪电茶博士极速服务，采纳该主意操作成功，请刷新页面查看。")

}

// HandleFunc ThreadSupplement(w http.ResponseWriter, r *http.Request)
func HandleThreadSupplement(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		threadSupplementGet(w, r)
	case http.MethodPost:
		threadSupplementPost(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/thread/supplement?uuid=xxx
// 打开指定的茶议（议程）追加（补充必需内容）页面，
func threadSupplementGet(w http.ResponseWriter, r *http.Request) {

	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", sess.Email, err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属茶议。")
		return
	}
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		Report(w, r, "茶博士失魂鱼，未能读取茶议编号，请确认后再试。")
		return
	}
	// 读取茶议内容
	thread, err := data.GetThreadByUUID(uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Report(w, r, "你好，茶博士竟然说该茶议不存在，请确认后再试一次。")
			return
		}
		util.Debug(" Cannot read thread given uuid", uuid, err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议。")
		return
	}
	switch thread.Category {
	case data.ThreadCategoryAppointment:
		break
	case data.ThreadCategorySeeSeek:
		break
	case data.ThreadCategorySuggestion:
		break
	case data.ThreadCategoryGoods:
		break
	case data.ThreadCategoryHandcraft:
		break
	default:
		Report(w, r, "你好，茶博士表示，陛下，普通茶议不能加水呢。")
		return

	}

	var tD data.ThreadDetail
	// 读取茶议资料荚
	tD.ThreadBean, err = FetchThreadBean(thread)
	if err != nil {
		util.Debug(" Cannot read threadBean", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议资料荚。")
		return
	}
	//读取所在的茶台资料
	project, err := thread.Project()
	if err != nil {
		util.Debug(" Cannot read project given thread", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶台资料。")
		return
	}
	tD.QuoteProjectBean, err = FetchProjectBean(project)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			//	util.Debug(" Cannot read project given uuid", uuid)
			Report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个茶台不存在。")
			return
		}
		util.Debug(" Cannot read project bean given project", project.Id, err)
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取茶台资料。")
		return
	}

	//读取茶围资料
	objective, err := project.Objective()
	if err != nil {
		util.Debug(" Cannot read objective given project", err)
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取茶围资料。")
		return
	}
	tD.QuoteObjectiveBean, err = FetchObjectiveBean(objective)
	if err != nil {
		util.Debug(" Cannot read objective given project", project.Id, err)
		Report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}

	//核对用户身份，是否具有完善操作权限
	verifier_team := data.Team{Id: data.TeamIdVerifier}
	ok, err := verifier_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug(" Cannot check team membership", verifier_team.Id, err)
		Report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您没有权限补充该茶议。")
		return
	}
	if !ok {
		Report(w, r, "茶博士惊讶，陛下你没有权限补充该茶议，请确认后再试。")
		return
	}

	tD.IsGuest = false
	tD.IsMaster = false
	tD.IsAdmin = true
	tD.IsInput = true
	//从会话查获当前浏览用户资料荚
	s_u, s_d_family, s_all_families, s_default_team, s_survival_teams, s_default_place, s_places, err := FetchSessionUserRelatedData(sess)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", s_u.Id)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	tD.SessUser = s_u
	tD.SessUserDefaultFamily = s_d_family
	tD.SessUserSurvivalFamilies = s_all_families
	tD.SessUserDefaultTeam = s_default_team
	tD.SessUserSurvivalTeams = s_survival_teams
	tD.SessUserDefaultPlace = s_default_place
	tD.SessUserBindPlaces = s_places

	RenderHTML(w, &tD, "layout", "navbar.private", "thread.supplement")

}

// POST /v1/thread/supplement
// 补充完整必备茶议5部曲内容
// 规则是只能补充剩余字数,
// 不能超过max_word，不能修改已记录的茶议内容，
// 不能修改茶议等级，
// 不能修改茶议标题，
// 不能修改茶议创建时间，
// 不能修改茶议是否开放式/封闭式，
func threadSupplementPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")

}
