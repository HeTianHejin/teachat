package route

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
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
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，闺中女儿惜春暮，愁绪满怀无释处。")
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
	if err != nil || (thre_type != data.ThreadTypeIthink && thre_type != data.ThreadTypeIdea) {
		util.Debug("Invalid thread type value", thre_type, err)
		report(w, r, "你好，闺中女儿惜春暮，愁绪满怀无释处。")
		return
	}

	// 检查ty值是否合法
	switch thre_type {
	case data.ThreadTypeIthink, data.ThreadTypeIdea:
		break
	default:
		util.Debug("Invalid thread type value", err)
		report(w, r, "你好，闺中女儿惜春暮，愁绪满怀无释处。")
		return
	}
	body := r.PostFormValue("topic")
	title := r.PostFormValue("title")
	project_id, err := strconv.Atoi(r.PostFormValue("project_id"))
	if err != nil {
		util.Debug("Failed to convert project_id to int", project_id, err)
		report(w, r, "你好，闪电茶博士极速查找茶台中，请确认后再试。")
		return
	}
	post_id, err := strconv.Atoi(r.PostFormValue("post_id"))
	if err != nil {
		util.Debug("Failed to convert post_id to int", project_id, err)
		report(w, r, "你好，闪电茶博士极速服务，任然无法识别提交的品味资料，请确认后再试。")
		return
	}
	/// check submit post_id is valid, if not 0 表示属于“议中议”
	post := data.Post{Id: post_id}
	proj := data.Project{Id: project_id}
	//检查该茶台是否存在，而且状态不是待友邻蒙评审查草台状态
	if err = proj.Get(); err != nil {
		util.Debug(" Cannot get project", err)
		report(w, r, "你好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	if proj.Class != data.PrClassOpen && proj.Class != data.PrClassClose {
		//util.Debug("试图访问未蒙评审核的茶台被阻止。", s_u.Email, err)
		report(w, r, "你好，茶博士竟然说该茶台尚未启用，请确认后再试一次。")
		return
	}
	if post_id > 0 {
		if err = post.Get(); err != nil {
			util.Debug(" Cannot get post given id", post_id, err)
			report(w, r, "你好，闪电茶博士极速服务，然而无法识别提交的品味资料，请确认后再试。")
			return
		}
		test_proj, err := post.Project()
		if err != nil {
			util.Debug(" Cannot get post given id", post_id, err)
			report(w, r, "你好，闪电茶博士极速服务，然而无法识别提交的品味资料，请确认后再试。")
			return
		}
		// 检查提及的post和project是否匹配
		if proj.Id != test_proj.Id {
			util.Debug(project_id, "post_id and project_id do not match")
			report(w, r, "你好，茶博士居然说这个茶台有一点点问题，请确认后再试一次。")
			return
		}
	}

	// 检查茶议（thread）创建权限
	if ok := checkCreateThreadPermission(proj, s_u.Id, w, r); !ok {
		report(w, r, "你好，茶博士居然说,陛下您的大名竟然不在邀请名单上，请确认后再试一次。")
		return
	}

	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"

	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		report(w, r, "你好，此地无这个茶团，请确认后再试。")
		return
	}
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		report(w, r, "你好，此地无这个茶团，请确认后再试。")
		return
	}

	valid, err := validateTeamAndFamilyParams(is_private, team_id, family_id, s_u.Id, w, r)
	if !valid && err == nil {
		return // 参数不合法，已经处理了错误
	}
	if err != nil {
		// 处理数据库错误
		util.Debug("验证提交的团队和家庭id出现数据库错误", team_id, family_id, err)
		report(w, r, "你好，茶团成员资格检查未通过，请确认后再试。")
		return
	}

	//检测一下title是否不为空，而且中文字数<24,topic不为空，而且中文字数<int(util.Config.ThreadMaxWord)
	if !validateCnStrLen(title, 1, 36, "标题", w, r) {
		return
	}
	if !validateCnStrLen(body, int(util.Config.ThreadMinWord), int(util.Config.ThreadMaxWord), "内容", w, r) {
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
		Status:    data.DraftThreadClassPending,
	}
	if post_id > 0 {
		draft_thread.Category = data.ThreadCategoryNested
	}
	if err = draft_thread.Create(); err != nil {
		util.Debug(" Cannot create thread draft", err)
		report(w, r, "你好，茶博士没有墨水了，未能保存新茶议草稿。")
		return
	}

	if util.Config.PoliteMode {
		if err = createAndSendAcceptMessage(draft_thread.Id, data.AcceptObjectTypeTh, s_u.Id); err != nil {
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				report(w, r, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				report(w, r, "你好，茶博士迷路了，未能发送蒙评请求消息。")
			}
			return
		}
		t := fmt.Sprintf("你好，你在“ %s ”茶台发布的茶议已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", proj.Title)
		// 提示用户草稿保存成功
		report(w, r, t)
		return

	} else {
		// 无需发送AcceptObject消息，直接创建新茶议
		thread, err := acceptNewDraftThread(draft_thread.Id)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "获取茶议草稿失败"):
				util.Debug("Cannot get draft-thread", err)
				report(w, r, "你好，茶博士失魂鱼，竟然说有时候泡一壶好茶的关键，需要的不是技术而是耐心。")
			case strings.Contains(err.Error(), "更新茶议草稿状态失败"):
				util.Debug("Cannot update draft-thread status", err)
				report(w, r, "你好，睿藻仙才盈彩笔，自惭何敢再为辞。")
			case strings.Contains(err.Error(), "创建新茶议失败"):
				util.Debug("Cannot save thread", err)
				report(w, r, "你好，吟成荳蔻才犹艳，睡足酴醾梦也香。")
			default:
				util.Debug("未知错误", err)
				report(w, r, "世事洞明皆学问，人情练达即文章。")
			}
			return
		}
		//跳转到新茶议详情页
		http.Redirect(w, r, fmt.Sprintf("/v1/thread/detail?uuid=%s", thread.Uuid), http.StatusFound)
		return
	}
}

// GET /v1/thread/detail?uuid=
// 显示需求uuid茶议（议题）的详细信息，包括品味（回复帖子）和记录品味的表格
func ThreadDetail(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，茶博士看不透您提交的茶议编号。")
		return
	}

	var tD data.ThreadDetail

	// 读取茶议内容
	thread, err := data.GetThreadByUUID(uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，茶博士竟然说该茶议不存在，请确认后再试一次。")
			return
		}
		util.Debug(" Cannot read thread given uuid", uuid, err)
		report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？")
		return
	}

	//读取所在的茶台资料
	project, err := thread.Project()
	if err != nil {
		util.Debug(" Cannot read project given thread", err)
		report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？")
		return
	}
	tD.QuoteProjectBean, err = fetchProjectBean(project)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			//	util.Debug(" Cannot read project given uuid", uuid)
			report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个茶台不存在。")
			return
		}
		util.Debug(" Cannot read project", err)
		report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人？")
		return
	}

	//读取茶围资料
	objective, err := project.Objective()
	if err != nil {
		util.Debug(" Cannot read objective given project", err)
		report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}
	tD.QuoteObjectiveBean, err = fetchObjectiveBean(objective)
	if err != nil {
		util.Debug(" Cannot read objective given project", project.Id, err)
		report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}

	//检查品味的类型
	if thread.PostId != 0 {
		// 说明这是一个附加类型的,针对某个post发表的茶议(comment-in-thread，讲开又讲，延伸话题)
		post := data.Post{Id: thread.PostId}
		if err = post.Get(); err != nil {
			util.Debug(" Cannot read post given post_id", err)
			report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。未能读取品味资料。")
			return
		}
		tD.QuotePostBean, err = fetchPostBean(post)
		if err != nil {
			util.Debug(" Cannot fetch postBean given post", post.Id, err)
			report(w, r, "你好，茶博士失魂鱼，未能读取品味资料。")
			return
		}

		// 截短body
		tD.QuotePostBean.Post.Body = subStr(tD.QuotePostBean.Post.Body, 66)

	} else {
		// 是一个普通的茶台
		// 截短body
		tD.QuoteProjectBean.Project.Body = subStr(tD.QuoteProjectBean.Project.Body, 66)

	}

	// 读取茶议资料荚
	tD.ThreadBean, err = fetchThreadBean(thread, r)
	if err != nil {
		util.Debug(" Cannot read threadBean", err)
		report(w, r, "你好，茶博士失魂鱼，未能读取茶议资料荚。")
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
	tD.ProgressSupport = progressRound(n1, n2)
	tD.ProgressOppose = 100 - tD.ProgressSupport

	post_admin_slice, err := tD.ThreadBean.Thread.PostsAdmin()
	if err != nil {
		util.Debug(" Cannot read admin posts", err)
		report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	//统计post_admin_slice[i].FamilyId数量 ，重复的family_id数量不计入
	//统计post_admin_slice[i].TeamId数量 ，重复的id数量不计入
	familyMap := make(map[int]struct{})
	teamMap := make(map[int]struct{})

	for _, post := range post_admin_slice {
		// 处理家庭ID
		if post.FamilyId != data.FamilyIdUnknown {
			if _, exists := familyMap[post.FamilyId]; !exists {
				familyMap[post.FamilyId] = struct{}{}
			}
		}

		// 处理团队ID
		if post.TeamId > data.TeamIdFreelancer && post.TeamId != data.TeamIdVerifier {
			if _, exists := teamMap[post.TeamId]; !exists {
				teamMap[post.TeamId] = struct{}{}
			}
		}
	}

	tD.PostBeanAdminSlice, err = fetchPostBeanSlice(post_admin_slice)
	if err != nil {
		util.Debug(" Cannot read admin postbean", err)
		report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}

	// 读取全部普通回复帖子（品味）
	post_n_slice, err := tD.ThreadBean.Thread.Posts()
	if err != nil {
		util.Debug(" Cannot read posts", err)
		report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	for _, post := range post_n_slice {
		if post.FamilyId != data.FamilyIdUnknown {
			if _, exists := familyMap[post.FamilyId]; !exists {
				familyMap[post.FamilyId] = struct{}{}
			}
		}

		if post.TeamId > data.TeamIdFreelancer && post.TeamId != data.TeamIdVerifier {
			if _, exists := teamMap[post.TeamId]; !exists {
				teamMap[post.TeamId] = struct{}{}
			}
		}
	}

	tD.StatsSet.FamilyCount = len(familyMap)
	tD.StatsSet.TeamCount = len(teamMap)
	//tD.StatsSet.PersonCount = ?

	tD.PostBeanSlice, err = fetchPostBeanSlice(post_n_slice)
	if err != nil {
		util.Debug(" Cannot read posts", err)
		report(w, r, "你好，茶博士失魂鱼，嘀咕无为有处有还无？。")
		return
	}
	// 读取会话
	s, err := session(r)

	if err != nil {
		// 游客
		// 检查茶议的级别状态
		if tD.ThreadBean.Thread.Class == data.ThreadClassOpen || tD.ThreadBean.Thread.Class == data.ThreadClassClosed {
			//记录茶议被点击数
			//tD.ThreadBean.Thread.AddHitCount()
			// 填写页面数据

			tD.IsGuest = true

			tD.SessUser = data.User{
				Id:   data.UserId_None,
				Name: "游客",
				// 用户足迹
				Footprint: r.URL.Path,
				Query:     r.URL.RawQuery,
			}

			//show the thread and the posts展示页面
			generateHTML(w, &tD, "layout", "navbar.public", "thread.detail", "component_sess_capacity", "component_post_left", "component_post_right", "component_avatar_name_gender")
			return
		} else {
			report(w, r, "茶水温度太高了，不适合品味，请稍后再试。")
			return
		}
	} else {
		//用户是登录状态

		// 检查当前是哪一种类型茶议
		switch thread.Category {
		case data.ThreadCategoryNested:
			//针对某个post的议中议，检查是否有权限访问
			if thread.PostId == 0 {
				util.Debug(" Invalid thread category and post_id not match", thread.Id, thread.Category, thread.PostId)
				report(w, r, "你好，茶博士说，这个茶台状态异常无法使用。")
				return
			}
			//检查当前用户是否有权限访问该议中议

		case data.ThreadCategoryAppointment:
			//检查是否已存在约茶记录
			pr_appointment, err := data.GetAppointmentByProjectId(project.Id, r.Context())
			if err != nil && err != sql.ErrNoRows {
				util.Debug(" Cannot read appointment given project", err)
				report(w, r, "你好，假作真时真亦假，无为有处有还无？")
				return
			}
			// 可能没有记录
			if pr_appointment.Id > 0 {
				tD.Appointment = pr_appointment
			}
		case data.ThreadCategorySeeSeek:
			//检查see-seek是否存在记录？
			see_seek, err := data.GetSeeSeekByProjectId(project.Id, r.Context())
			if err != nil && err != sql.ErrNoRows {
				util.Debug(" Cannot read see-seek given project", err)
				report(w, r, "你好，假作真时真亦假，无为有处有还无？")
				return
			}
			if see_seek.Id > 0 {
				tD.SeeSeek = see_seek
			}

		case data.ThreadCategoryBrainFire:
			// 检查BrainFire是否存在记录
			brain_fire, err := data.GetBrainFireByProjectId(project.Id, r.Context())
			if err != nil && err != sql.ErrNoRows {
				util.Debug(" Cannot read brain-fire given project_id:", project.Id, err)
				report(w, r, "你好，假作真时真亦假，无为有处有还无？")
				return
			}
			if brain_fire.Id > 0 {
				tD.BrainFire = brain_fire
			}
		case data.ThreadCategorySuggestion:
			// 检查Suggestion是否存在记录
			suggestion, err := data.GetSuggestionByProjectId(project.Id, r.Context())
			if err != nil && err != sql.ErrNoRows {
				util.Debug("cannot read suggestion given project_id:", project.Id, err)
				report(w, r, "你好，假作真时真亦假，无为有处有还无？")
				return
			}
			if suggestion.Id > 0 {
				tD.Suggestion = suggestion
			}
		case data.ThreadCategoryGoods:
			// 检查goods记录是否存在
			//goods, er := data.GetG
		case data.ThreadCategoryHandcraft:
			break
		default:
			report(w, r, "你好，茶博士表示，陛下，奇怪的茶不能喝。")
			return

		}

		if tD.ThreadBean.Thread.Class == data.ThreadClassOpen || tD.ThreadBean.Thread.Class == data.ThreadClassClosed {
			//从会话查获当前浏览用户资料荚
			s_u, s_d_family, s_all_families, s_default_team, s_survival_teams, s_default_place, s_places, err := fetchSessionUserRelatedData(s)
			if err != nil {
				util.Debug(" Cannot get user-related data from session", s_u.Id)
				report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
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
				report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
				return
			}

			tD.IsMaster, err = checkProjectMasterPermission(&project, s_u.Id)
			if err != nil {
				util.Debug(" Cannot check project master permission", project.Id, err)
				report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
				return
			}

			if !tD.IsAdmin && !tD.IsMaster {
				// 检测当前会话茶友是否见证者
				is_member := isVerifier(s_u.Id)
				if is_member {
					tD.IsVerifier = true
				}
			}

			// 检测是否Post品味撰写者
			for i := range tD.PostBeanAdminSlice {
				if tD.PostBeanAdminSlice[i].Post.UserId == s_u.Id {
					tD.PostBeanAdminSlice[i].Post.ActiveData.IsAuthor = true
				}
			}
			for i := range tD.PostBeanSlice {
				if tD.PostBeanSlice[i].Post.UserId == s_u.Id {
					tD.PostBeanSlice[i].Post.ActiveData.IsAuthor = true
				}
			}

			if s_u.Id == tD.ThreadBean.Thread.UserId {
				// 是茶议撰写者！
				tD.ThreadBean.Thread.ActiveData.IsAuthor = true
				tD.IsInput = true
				//点击数+1
				//tD.ThreadBean.Thread.AddHitCount()
				//记录用户阅读该帖子一次
				//data.SaveReadedUserId(tD.ThreadBean.Thread.Id, s_u.Id)

				//展示撰写者视角茶议详情页面
				generateHTML(w, &tD, "layout", "navbar.private", "thread.detail", "component_sess_capacity", "component_post_left", "component_post_right", "component_avatar_name_gender")
				return
			} else {
				//不是茶议撰写者
				//记录茶议被点击数
				//tD.ThreadBean.Thread.AddHitCount()

				//检查是否封闭式茶台
				if tD.QuoteProjectBean.Project.Class == data.PrClassClose {
					//是封闭式茶台，需要检查当前用户身份是否受邀请茶团的成员，以决定是否允许发言
					ok, err := tD.QuoteProjectBean.Project.IsInvitedMember(s_u.Id)
					if err != nil {
						util.Debug(" Cannot check project invited member", err)
						report(w, r, "你好，桃李明年能再发，明年闺中知有谁？你真的是受邀请茶团成员吗？")
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
				if !tD.IsAdmin && !tD.IsMaster && !tD.IsVerifier {
					for i := range tD.PostBeanSlice {
						if tD.PostBeanSlice[i].Post.UserId == s_u.Id {
							tD.IsInput = false
							tD.IsPostExist = true
							break
						}
					}
				}
				//展示非撰写者视角茶议详情页面
				generateHTML(w, &tD, "layout", "navbar.private", "thread.detail", "component_sess_capacity", "component_post_left", "component_post_right", "component_avatar_name_gender")
				return
			}
		} else {
			//访问未开放等级的话题？
			report(w, r, "你好，外星人出没注意！请确认或者联系管理员确认稍后再试。")
			return
		}
	}

}

// POST /v1/thread/approve
// 茶台管理，决定采纳（认可）某个goodidea（thread），被采纳的thread被标识为“已采纳”approved
func ThreadApprove(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，闪电茶博士为了提高服务速度而迷路了，未能找到你想要的茶台。")
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	uuid := r.PostFormValue("id")
	if uuid == "" {
		report(w, r, "你好，闪电茶博士极速服务中，未能读取茶议资料，请稍后再试。")
		return
	}

	thread, err := data.GetThreadByUUID(uuid)
	if err != nil {
		util.Debug(" Cannot read thread given uuid", uuid)
		report(w, r, "你好，闪电茶博士极速服务中，未能读取茶议资料，请稍后再试。")
		return
	}
	proj, err := thread.Project()
	if err != nil {
		util.Debug(thread.Id, " Cannot read project given thread_id")
		report(w, r, "你好，闪电茶博士极速服务中，未能读取茶台资料，请稍后再试。")
		return
	}
	ob, err := proj.Objective()
	if err != nil {
		util.Debug(proj.Id, " Cannot read objective given project")
		report(w, r, "你好，闪电茶博士极速服务中，未能读取茶台资料，请稍后再试。")
		return
	}

	//检查用户是否有权限处理这个请求
	admin_team, err := data.GetTeam(ob.TeamId)
	if err != nil {
		util.Debug(proj.TeamId, " Cannot get team given id")
		report(w, r, "你好，闪电茶博士极速服务中，未能读取团队资料，请稍后再试。")
		return
	}
	//检查是否支持team成员
	is_admin, err := admin_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug(admin_team.Id, " Cannot check team membership")
		report(w, r, "你好，闪电茶博士极速服务中，未能读取团队资料，请稍后再试。")
		return
	}

	if !is_admin {
		//没有权限处理请求
		report(w, r, "你好，闪电茶博士极速服务，火星保安竟然说你没有权限处理该请求。")
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
		report(w, r, "你好，闪电茶博士极速服务，该茶议已被采纳，请刷新页面查看。")
		return
	}
	if err = new_thread_approved.Create(); err != nil {
		util.Debug(thread.Id, " Cannot create thread approved")
		report(w, r, "你好，闪电茶博士极速服务中，未能处理你的请求，请稍后再试。")
		return
	}

	//采纳（认可好主意）成功,跳转茶议详情页面
	http.Redirect(w, r, "/v1/thread/detail?uuid="+thread.Uuid, http.StatusFound)
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

	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", sess.Email, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "茶博士失魂鱼，未能读取茶议编号，请确认后再试。")
		return
	}

	var threSupp data.ThreadSupplement
	ctx := r.Context()
	// 读取茶议内容
	thread, err := data.GetThreadByUUID(uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，茶博士竟然说该茶议不存在，请确认后再试一次。")
			return
		}
		util.Debug(" Cannot read thread given uuid", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	//核对用户身份，是否具有完善操作权限
	verifier_team := data.Team{Id: data.TeamIdVerifier}
	ok, err := verifier_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug(" Cannot check team membership", verifier_team.Id, err)
		report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说陛下您没有权限补充该茶议。")
		return
	}
	if !ok {
		report(w, r, "茶博士惊讶，陛下你没有权限补充该茶议，请确认后再试。")
		return
	}
	threSupp.IsVerifier = true

	// 读取茶议资料荚
	threSupp.ThreadBean, err = fetchThreadBean(thread, r)
	if err != nil {
		util.Debug(" Cannot read threadBean", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	//读取所在的茶台资料
	project, err := thread.Project()
	if err != nil {
		util.Debug(" Cannot read project given thread", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	threSupp.QuoteProjectBean, err = fetchProjectBean(project)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			//	util.Debug(" Cannot read project given uuid", uuid)
			report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您提及的这个茶台不存在。")
			return
		}
		util.Debug(" Cannot read project bean given project", project.Id, err)
		report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}

	//读取茶围资料
	objective, err := project.Objective()
	if err != nil {
		util.Debug(" Cannot read objective given project", err)
		report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}
	threSupp.QuoteObjectiveBean, err = fetchObjectiveBean(objective)
	if err != nil {
		util.Debug(" Cannot read objective given project", project.Id, err)
		report(w, r, "你好，枕上轻寒窗外雨，眼前春色梦中人。")
		return
	}
	post_admin_slice, err := threSupp.ThreadBean.Thread.PostsAdmin()
	if err != nil {
		util.Debug(" Cannot read admin posts", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	threSupp.PostBeanAdminSlice, err = fetchPostBeanSlice(post_admin_slice)
	if err != nil {
		util.Debug(" Cannot read admin postbean", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	switch thread.Category {
	case data.ThreadCategoryAppointment:
		//检查是否已存在约茶记录
		pr_appointment, err := data.GetAppointmentByProjectId(project.Id, ctx)
		if err != nil && err != sql.ErrNoRows {
			util.Debug(" Cannot read appointment given project", err)
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
		threSupp.Appointment = pr_appointment

	case data.ThreadCategorySeeSeek:
		//检查see-seek是否存在记录？
		see_seek, err := data.GetSeeSeekByProjectId(project.Id, ctx)
		if err != nil && err != sql.ErrNoRows {
			util.Debug(" Cannot read see-seek given project", err)
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
		threSupp.SeeSeek = see_seek

	case data.ThreadCategoryBrainFire:
		// 检查脑火是否已存在记录
		brain_fire, err := data.GetBrainFireByProjectId(project.Id, ctx)
		if err != nil && err != sql.ErrNoRows {
			util.Debug(" Cannot read brain fire given project", err)
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
		threSupp.BrainFire = brain_fire

	case data.ThreadCategorySuggestion:
		// 检查Suggestion是否存在记录
		suggestion, err := data.GetSuggestionByProjectId(project.Id, r.Context())
		if err != nil && err != sql.ErrNoRows {
			util.Debug("cannot read suggestion given project_id:", project.Id, err)
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
		if suggestion.Id > 0 {
			threSupp.Suggestion = suggestion
		}
	case data.ThreadCategoryGoods:
		break
	case data.ThreadCategoryHandcraft:
		break
	default:
		report(w, r, "你好，茶博士表示，陛下，普通茶议不能加水呢。")
		return

	}

	// 用户足迹
	s_u.Footprint = r.URL.Path
	s_u.Query = r.URL.RawQuery

	threSupp.SessUser = s_u

	generateHTML(w, &threSupp, "layout", "navbar.private", "thread.supplement", "component_post_left", "component_post_right", "component_sess_capacity", "component_avatar_name_gender")

}

// POST /v1/thread/supplement
// 补充完整必备茶议5部曲内容
// 规则是只能补充剩余字数,
// 不能超过max_word，不能修改已记录的茶议内容，
// 不能修改茶议等级，
// 不能修改茶议标题，
// 不能修改茶议是否开放式/封闭式，
func threadSupplementPost(w http.ResponseWriter, r *http.Request) {

	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", sess.Email, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	//检查用户身份，是否具有完善操作权限
	verifier_team := data.Team{Id: data.TeamIdVerifier}
	ok, err := verifier_team.IsMember(s_u.Id)
	if err != nil {
		util.Debug(" Cannot check team membership", verifier_team.Id, err)
		report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说陛下您没有权限补充该茶议。")
		return
	}
	if !ok {
		report(w, r, "茶博士惊讶，陛下你没有权限补充该茶议，请确认后再试。")
		return
	}
	//获取post方法提交的表单
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	t_uuid := r.PostFormValue("uuid")
	if t_uuid == "" {
		report(w, r, "你好，茶博士扶起厚厚的眼镜，居然说您补充的茶议编号不存在。")
		return
	}
	//读取提交的additional
	additional := r.PostFormValue("additional")

	// 读取茶议内容
	thread, err := data.GetThreadByUUID(t_uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，茶博士竟然说该茶议不存在，请确认后再试一次。")
			return
		}
		util.Debug(" Cannot read thread given uuid", t_uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}
	switch thread.Category {
	case data.ThreadCategoryAppointment:
		break
	case data.ThreadCategorySeeSeek:
		break
	case data.ThreadCategoryBrainFire:
		break
	case data.ThreadCategorySuggestion:
		break
	case data.ThreadCategoryGoods:
		break
	case data.ThreadCategoryHandcraft:
		break
	default:
		report(w, r, "你好，茶博士表示，陛下，普通茶议不能加水呢。")
		return
	}
	// log.Println(CnStrLen(thread.Body))
	// log.Println(CnStrLen(additional))
	//读取提交内容要求>int(util.Config.ThreadMinWord)中文字符，加上已有内容是否<=int(util.Config.ThreadMaxWord),
	if ok = submitAdditionalContent(w, r, thread.Body, additional); !ok {
		return
	}
	//当前“[中文时间字符 + 补充]” + body
	//获取当前时间，格式化成中文时间字符
	now := time.Now()
	timeStr := now.Format("2006年1月2日 15:04:05")
	name := s_u.Name
	// 追加内容（另起一行）
	t := "\n[" + timeStr + " " + name + " 补充] " + additional // 注意开头的 \n
	thread.Body += t
	//更新茶议内容
	if err = thread.UpdateBodyAndClass(thread.Body, thread.Class, r.Context()); err != nil {
		util.Debug(" Cannot update thread", err)
		report(w, r, "你好，茶博士失魂鱼，墨水中断未能补充茶议。")
		return
	}

	http.Redirect(w, r, "/v1/thread/detail?uuid="+t_uuid, http.StatusFound)

}
