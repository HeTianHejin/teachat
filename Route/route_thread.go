package route

import (
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// POST /v1/thread/draft
// 创建新茶议草稿，待邻座盲审后转为正式茶议
func DraftThread(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "您好，茶博士迷路了，未能找到你想要的茶台。")
		return
	}
	sUser, err := sess.User()
	if err != nil {
		util.Warning(err, " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//读取表单数据
	ty, err := strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		util.Warning(err, "Failed to convert class to int")
		Report(w, r, "您好，此地无银三百两，请确认后再试。")
		return
	}
	// 检查ty值是否 0、1、2
	switch ty {
	case 0, 1, 2:
		break
	default:
		util.Warning("Invalid thread type value")
		Report(w, r, "您好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	body := r.PostFormValue("topic")
	title := r.PostFormValue("title")
	project_id, err := strconv.Atoi(r.PostFormValue("project_id"))
	if err != nil {
		util.Warning(err, "Failed to convert class to int")
		Report(w, r, "您好，此地无银三百两，请确认后再试。")
		return
	}
	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.Warning(err, "Failed to convert class to int")
		Report(w, r, "您好，此地无银三百两，请确认后再试。")
		return
	}
	// check submit team_id is valid
	_, err = data.GetTeamMemberByTeamIdAndUserId(team_id, sUser.Id)
	if err != nil {
		util.Warning(err, " Cannot get team member by team id and user id")
		Report(w, r, "您好，此地无银三百两，请确认后再试。")
		return
	}

	//检查该茶台是否存在，而且状态不是草台状态
	proj := data.Project{
		Id: project_id,
	}
	if err = proj.GetById(); err != nil {
		util.Warning(err, " Cannot get project")
		Report(w, r, "您好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	if proj.Class == 10 || proj.Class == 20 {
		util.Warning(sUser.Email, "试图访问未盲评审核的茶台被阻止。")
		Report(w, r, "您好，茶博士竟然说该茶台尚未启用，请确认后再试一次。")
		return
	}

	// 如果茶台class=1，存为开放式茶议草稿，
	// 如果茶台class=2， 存为封闭式茶议草稿
	if proj.Class == 1 || proj.Class == 2 {
		//检测一下title是否不为空，而且中文字数<24,topic不为空，而且中文字数<456
		if CnStrLen(title) < 1 {
			Report(w, r, "您好，茶博士竟然说该茶议标题为空，请确认后再试一次。")
			return
		}
		if CnStrLen(title) > 36 {
			Report(w, r, "您好，茶博士竟然说该茶议标题过长，请确认后再试一次。")
			return
		}
		if CnStrLen(body) < 1 {
			Report(w, r, "您好，茶博士竟然说该茶议内容为空，请确认后再试一次。")
			return
		} else if CnStrLen(body) < 17 {
			Report(w, r, "您好，茶博士竟然说该茶议内容太短，请确认后再试一次。")
			return
		}
		if CnStrLen(body) > 456 {
			Report(w, r, "您好，茶博士小声说茶棚的小纸条只能写456字，请确认后再试一次。")
			return
		}

		//保存新茶议草稿
		draft_thread := data.DraftThread{
			UserId:    sUser.Id,
			ProjectId: project_id,
			Title:     title,
			Body:      body,
			Class:     proj.Class,
			Type:      ty,
			TeamId:    team_id,
		}
		if err = draft_thread.Create(); err != nil {
			util.Warning(err, " Cannot create thread draft")
			Report(w, r, "您好，茶博士没有墨水了，未能保存新茶议草稿。")
			return
		}
		// 创建一条友邻盲评,是否接纳 新茶的记录
		aO := data.AcceptObject{
			ObjectId:   draft_thread.Id,
			ObjectType: 3,
		}
		if err = aO.Create(); err != nil {
			util.Warning(err, "Cannot create accept_object")
			Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
			return
		}
		// 发送盲评请求消息给两个在线用户
		//构造消息
		mess := data.AcceptMessage{
			FromUserId:     1,
			Title:          "新茶语邻座评审邀请",
			Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
			AcceptObjectId: aO.Id,
		}
		//发送消息
		if err = AcceptMessageSendExceptUserId(sUser.Id, mess); err != nil {
			Report(w, r, "您好，早知日后闲争气，岂肯今朝错读书！未能发送盲评请求消息。")
			return
		}

		// 提示用户草稿保存成功
		t := fmt.Sprintf("你好，你在“ %s ”茶台发布的茶议已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", proj.Title)
		// 提示用户草稿保存成功
		Report(w, r, t)
		return
	}
	//出现非法的class值
	Report(w, r, "您好，糊里糊涂的茶博士竟然说该茶台坐满了外星人，请确认后再试一次。")

}

// GET /v1/thread/detail
// 显示茶议（议题）的详细信息，包括品味（回复帖子）和记录品味的表格
func ThreadDetail(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	uuid := vals.Get("id")

	// 准备一个空白的表
	var tD data.ThreadDetail
	// 读取茶议内容以填空
	thread, err := data.ThreadByUUID(uuid)
	if err != nil {
		util.Warning(err, " Cannot read thread")
		Report(w, r, "您好，茶博士失魂鱼，未能读取茶议。")
		return
	}

	if thread.PostId != 0 {
		// 说明这是一个附加类型的,针对某个post发表的茶议(chat-in-chat，讲开又讲，延伸话题)
		tD.QuotePost, err = data.GetPostbyId(thread.PostId)
		if err != nil {
			util.Warning(err, " Cannot read post given post_id")
			Report(w, r, "您好，枕上轻寒窗外雨，眼前春色梦中人。未能读取品味资料。")
			return
		}
		// 截短body
		tD.QuotePost.Body = Substr(tD.QuotePost.Body, 66)
		tD.QuotePostAuthor, err = tD.QuotePost.User()
		if err != nil {
			util.Warning(err, " Cannot read post author")
			Report(w, r, "您好，呜咽一声犹未了，落花满地鸟惊飞。未能读取品味资料。")
			return
		}
		tD.QuotePostAuthorTeam, err = data.GetTeamById(tD.QuotePost.TeamId)
		if err != nil {
			util.Warning(err, " Cannot read post author team")
			Report(w, r, "您好，花谢花飞飞满天，红消香断有谁怜？未能读取品味资料。")
			return
		}

	} else {
		// 是一个普通的茶议
		tD.QuoteProject, err = thread.Project()
		if err != nil {
			util.Warning(err, " Cannot read project")
			Report(w, r, "您好，枕上轻寒窗外雨，眼前春色梦中人。未能读取茶台资料。")
			return
		}
		// 截短body
		tD.QuoteProject.Body = Substr(tD.QuoteProject.Body, 66)
		tD.QuoteProjectAuthor, err = tD.QuoteProject.User()
		if err != nil {
			util.Warning(err, " Cannot read project author")
			Report(w, r, "您好，静夜不眠因酒渴，沉烟重拨索烹茶。未能读取茶台资料。")
			return
		}
		tD.QuoteProjectAuthorTeam, err = data.GetTeamById(tD.QuoteProject.TeamId)
		if err != nil {
			util.Warning(err, " Cannot read project author team")
			Report(w, r, "您好，绛芸轩里绝喧哗，桂魄流光浸茜纱。未能读取茶台资料。")
			return
		}
	}

	// 读取茶议资料荚
	tD.ThreadBean, err = GetThreadBean(thread)
	if err != nil {
		util.Warning(err, " Cannot read threadBean")
		Report(w, r, "您好，茶博士失魂鱼，未能读取茶议资料荚。")
		return
	}
	tD.NumSupport = thread.NumSupport()
	tD.NumOppose = thread.NumOppose()
	//品味中颔首的合计得分与总得分的比值，取整数，用于客户端页面进度条设置，正反双方进展形势对比
	n1, err := tD.ThreadBean.Thread.PostsScoreSupport()
	if n1 != 0 && err != nil {

		util.Warning(err, " Cannot get posts score support")
		Report(w, r, "您好，莫失莫忘，仙寿永昌，有些资料被黑风怪瓜州了。")
		return

	}
	n2, err := tD.ThreadBean.Thread.PostsScore()
	if n2 != 0 && err != nil {

		util.Warning(err, " Cannot get posts score oppose")
		Report(w, r, "您好，莫失莫忘，仙寿永昌，有些资料,被黑风怪瓜州了。")
		return

	}
	tD.ProgressSupport = ProgressRound(n1, n2)
	tD.ProgressOppose = 100 - tD.ProgressSupport
	// 读取全部回复帖子（品味）
	post_list, err := tD.ThreadBean.Thread.Posts()
	if err != nil {
		util.Warning(err, " Cannot read posts")
		Report(w, r, "您好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	tD.PostBeanList, err = GetPostBeanList(post_list)
	if err != nil {
		util.Warning(err, " Cannot read posts")
		Report(w, r, "您好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	// 读取会话
	sess, err := Session(r)

	if err != nil {
		// 游客
		// 检查茶议的级别状态
		if tD.ThreadBean.Thread.Class == 1 || tD.ThreadBean.Thread.Class == 2 {
			//记录茶议被点击数
			tD.ThreadBean.Thread.AddHitCount()
			// 填写页面数据
			tD.ThreadBean.Thread.PageData.IsAuthor = false
			tD.IsInput = false
			tD.SessUser = data.User{
				Id:   0,
				Name: "游客",
			}
			//迭代postList,标记非品味作者
			for i := range tD.PostBeanList {
				tD.PostBeanList[i].Post.PageData.IsAuthor = false
			}

			//show the thread and the posts展示页面
			GenerateHTML(w, &tD, "layout", "navbar.public", "thread.detail")
			return
		} else {
			//非法访问未开放的话题？
			util.Warning(err, " 试图访问未公开的thread")
			Report(w, r, "茶水温度太高了，不适合品味，请稍后再试。")
			return
		}
	} else {
		//用户是登录状态,可以访问1和2级茶议
		if tD.ThreadBean.Thread.Class == 1 || tD.ThreadBean.Thread.Class == 2 {
			//从会话查获当前浏览用户资料荚
			s_u, s_default_team, s_survival_teams, err := FetchUserRelatedData(sess)
			if err != nil {
				util.Warning(err, " Cannot get user-related data from session")
				Report(w, r, "您好，茶博士失魂鱼，有眼不识泰山。")
				return
			}
			tD.SessUser = s_u
			tD.SessUserDefaultTeam = s_default_team
			tD.SessUserSurvivalTeams = s_survival_teams

			// 检测是否茶议作者
			if s_u.Id == tD.ThreadBean.Thread.UserId {
				// 是茶议作者！
				// 填写页面数据
				tD.ThreadBean.Thread.PageData.IsAuthor = true
				// 提议作者不能自评品味，王婆卖瓜也不行？！
				tD.IsInput = false
				//点击数+1
				tD.ThreadBean.Thread.AddHitCount()
				//记录用户阅读该帖子一次
				data.SaveReadedUserId(tD.ThreadBean.Thread.Id, s_u.Id)
				//迭代PostList，把其PageData.IsAuthor设置为false，页面渲染时检测布局用
				for i := range tD.PostBeanList {
					tD.PostBeanList[i].Post.PageData.IsAuthor = false
				}

				//show the thread and the posts展示页面
				GenerateHTML(w, &tD, "layout", "navbar.private", "thread.detail")
				return
			} else {
				//不是茶议作者
				//记录用户阅读该帖子一次
				data.SaveReadedUserId(tD.ThreadBean.Thread.Id, s_u.Id)
				//记录茶议被点击数
				tD.ThreadBean.Thread.AddHitCount()
				// 填写页面数据
				tD.SessUser = s_u
				//tPD.IsGuest = false
				tD.ThreadBean.Thread.PageData.IsAuthor = false

				// 打开品味撰写
				tD.IsInput = true
				// 如果当前用户已经品味过了，则关闭撰写输入面板
				// 用于页面判断是否显示品味POST（回复）撰写面板
				for i := range tD.PostBeanList {
					if tD.PostBeanList[i].Post.UserId == s_u.Id {
						tD.IsInput = false
						break
					}
				}

				// 检测是否其中某一个Post品味作者
				for i := range tD.PostBeanList {
					if tD.PostBeanList[i].Post.UserId == s_u.Id {
						tD.PostBeanList[i].Post.PageData.IsAuthor = true
						break
					} else {
						tD.PostBeanList[i].Post.PageData.IsAuthor = false
					}
				}

				//展示茶议详情
				GenerateHTML(w, &tD, "layout", "navbar.private", "thread.detail")
				return
			}
		} else if tD.ThreadBean.Thread.Class == 0 {
			//茶议的等级发生了变化，需要重新进行邻桌评估
			Report(w, r, "您好，茶议加水后出现了神迹！请耐心等待邻桌来推荐。")
			return
		} else {
			//访问未开放等级的话题？
			Report(w, r, "您好，外星人出没注意！请确认或者联系管理员确认稍后再试。")
			return
		}
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
			util.Warning(err, " Cannot get user from session")
			Report(w, r, "您好，茶博士失魂鱼，未能读取会话用户资料。")
			return
		}
		vals := r.URL.Query()
		uuid := vals.Get("id")
		thDPD.ThreadBean.Thread, err = data.ThreadByUUID(uuid)
		if err != nil {
			util.Warning("Cannot not read thread")
			Report(w, r, "茶博士失魂鱼，未能读取茶议资料，请稍后再试。")
			return
		}
		if thDPD.ThreadBean.Thread.UserId == sUser.Id {
			// 是作者，可以加水（补充内容）
			thDPD.ThreadBean.Thread.PageData.IsAuthor = true
			thDPD.SessUser = sUser
			GenerateHTML(w, &thDPD, "layout", "navbar.private", "thread.edit")
			return
		}
		//不是作者，不能加水
		util.Danger("Cannot edit other user's thread")
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
			util.Danger(err, " Cannot parse form")
			return
		}
		user, err := sess.User()
		if err != nil {
			util.Danger(err, " Cannot get user from session")
			Report(w, r, "您好，茶博士失魂鱼，未能读取专属茶议。")
			return
		}
		uuid := r.PostFormValue("uuid")
		//title := r.PostFormValue("title")
		topi := r.PostFormValue("additional")
		//根据用户提供的uuid读取指定茶议
		thread, err := data.ThreadByUUID(uuid)
		if err != nil {
			util.Warning(err, " Cannot read thread by uuid")
			Report(w, r, "茶博士失魂鱼，未能读取专属茶议，请稍后再试。")
			return
		}
		//核对一下用户身份
		if thread.UserId == user.Id {
			//检查topi内容是否中文字数>17,并且thread.Topic总字数<456,如果是则可以补充内容
			if CnStrLen(topi) >= 17 && CnStrLen(thread.Body+topi) < 456 {
				thread.Body += topi
			} else {
				util.Info("Cannot update thread")
				Report(w, r, "猫是彬彬有礼的茶博士竟然声称纸太贵，字太少或者超过456字的茶议不收录！请确认后再试。")
				return
			}
			// 修改过的茶议,重置class=0,表示草稿状态，
			thread.Class = 0
			//许可修改自己的茶议
			if err := thread.UpdateTopicAndClass(thread.Body, thread.Class); err != nil {
				util.Danger(err, " Cannot update thread")
				Report(w, r, "茶博士失魂鱼，未能更新专属茶议，请稍后再试。")
				return
			}
			url := fmt.Sprint("/v1/thread/detail?id=", uuid)
			http.Redirect(w, r, url, http.StatusFound)
			return
		} else {
			//阻止修改别人的茶议
			util.Danger("Cannot edit other user's thread")
			Report(w, r, "茶博士提示，粗鲁的茶博士竟然说，仅能对自己的茶杯加水（追加内容）。")
			return
		}
	}
}

// POST /v1/thread/plus
// 附加型茶议，是针对某个品味（跟帖）发起的茶议草稿
func PlusThread(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		Report(w, r, "您好，未能找到你想要的茶台。")
		return
	}
	//根据会话读取当前浏览用户信息
	sUser, err := sess.User()
	if err != nil {
		util.Warning(err, " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 读取表单数据
	ty, err := strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		util.Warning(err, " Cannot convert ty to int")
		Report(w, r, "茶博士失魂鱼，未能读取新泡茶议资料，请稍后再试。")
		return
	}
	// 检查ty值是否 0、1、2
	if ty != 0 && ty != 1 && ty != 2 {
		Report(w, r, "您好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	body := r.PostFormValue("topic")
	title := r.PostFormValue("title")
	post_uuid := r.PostFormValue("uuid")

	// 检查提交的参数是否合规
	//检测一下title是否不为空，而且中文字数<19,topic不为空，而且中文字数<456
	if CnStrLen(title) < 1 {
		Report(w, r, "您好，茶博士竟然说该茶议标题为空，请确认后再试一次。")
		return
	}
	if CnStrLen(title) > 36 {
		Report(w, r, "您好，茶博士竟然说该茶议标题过长，请确认后再试一次。")
		return
	}
	if CnStrLen(body) < 1 {
		Report(w, r, "您好，茶博士竟然说该茶议内容为空，请确认后再试一次。")
		return
	} else if CnStrLen(body) < 17 {
		Report(w, r, "您好，茶博士竟然说该茶议内容太短，请确认后再试一次。")
		return
	}
	if CnStrLen(body) > 456 {
		Report(w, r, "您好，茶博士小声说茶棚的小纸条只能写456字，请确认后再试一次。")
		return
	}
	// 检查是那一张茶台发生的事
	post, err := data.GetPostByUuid(post_uuid)
	// ���查post是否存在
	if err != nil {
		util.Warning(err, " Cannot get post")
		Report(w, r, "您好，������的��������然声称这个��台��火星人����了。")
		return
	}
	proj, err := post.Project()
	if err != nil {
		util.Warning(err, " Cannot get project")
		Report(w, r, "您好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}

	var draft_thread data.DraftThread

	switch proj.Class {
	case 10, 20:
		Report(w, r, "您好，该茶台尚未启用，请确认后再试。")
		return
	case 1:
		// Open teaTable, Can have tea :)
		// 保存新茶议草稿
		draft_thread = data.DraftThread{
			UserId:    sUser.Id,
			ProjectId: proj.Id,
			Title:     title,
			Body:      body,
			Class:     proj.Class,
			Type:      ty,
			PostId:    post.Id,
		}
		if err = draft_thread.Create(); err != nil {
			util.Warning(err, " Cannot create thread draft")
			Report(w, r, "您好，茶博士的笔没有墨水了，未能保存新茶议草稿。")
			return
		}
	case 2:
		//Is the current session user allowed to join the tea tasting?
		//It depends on the tea group members invited by the host.
		ok, err := proj.IsInvitedMember(sUser.Id)
		// ���查用户是否����请到��台
		if err != nil {
			util.Warning(err, " Cannot check if user is invited")
			Report(w, r, "您好，������的��没有��水了，未能确认你是否����请到这个��台。")
			return
		}
		if ok {
			draft_thread = data.DraftThread{
				UserId:    sUser.Id,
				ProjectId: proj.Id,
				Title:     title,
				Body:      body,
				Class:     proj.Class,
				Type:      ty,
				PostId:    post.Id,
			}
			if err = draft_thread.Create(); err != nil {
				util.Warning(err, " Cannot create thread draft")
				Report(w, r, "您好，茶博士的笔没有墨水了，未能保存新茶议草稿。")
				return
			}
		} else {
			Report(w, r, "您好，茶博士彬彬有礼的说，你的大名竟然不在邀请品茶名单上。")
			return
		}
		// 异常状态的茶台
	default:
		Report(w, r, "您好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
		return
	}

	// 创建一条友邻盲评,是否接纳 新茶的记录
	ao := data.AcceptObject{
		ObjectId:   draft_thread.Id,
		ObjectType: 3,
	}
	if err = ao.Create(); err != nil {
		util.Warning(err, "Cannot create accept_object")
		Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	// 发送邻座盲评消息
	am := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶议邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
		AcceptObjectId: ao.Id,
	}
	// 发送消息
	if err = AcceptMessageSendExceptUserId(sUser.Id, am); err != nil {
		Report(w, r, "您好，茶博士迷路了，未能发送盲评请求消息。")
		return
	}
	// 提示用户草稿保存成功
	tips := fmt.Sprintf("你好，你在“ %s ”茶台发布的茶议已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", proj.Title)
	// 提示用户草稿保存成功
	Report(w, r, tips)

}
