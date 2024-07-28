package route

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/thread/new
// 显示创建新茶议表单页面
func NewThread(w http.ResponseWriter, r *http.Request) {
	// 根据会话读取当前用户信息
	s, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话读取当前用户信息
	u, _ := s.User()

	// 读取提交的id号,看是哪一张茶台请求开启新茶议
	vals := r.URL.Query()
	proj_uuid := vals.Get("id")
	// 获取茶台信息
	var prDPD data.ProjectDetailPageData
	prDPD.Project, err = data.GetProjectByUuid(proj_uuid)
	if err != nil {
		util.Danger(err, " Cannot get project")
		util.Report(w, r, "您好，茶博士都糊涂了，未能找到你想要的茶台。")
		return
	}

	// 填写会话用户资料
	prDPD.SessUser = u

	// 检查茶台属性，class=1开放式，class=2封闭式，
	// 如果是封闭式，需要查看台主指定了那些茶团成员可以品茶，确认用户是否能在此茶台提出茶议（主张/提议），
	switch prDPD.Project.Class {
	case 1:
		// 开放式茶台，不需要查看台主指定了那些茶团成员可以品茶
		util.GenerateHTML(w, &prDPD, "layout", "navbar.private", "thread.new")
		return
	case 2:
		// 封闭式茶台，需要查看台主指定了那些茶团（团队）可以品茶
		team_ids, err := prDPD.Project.InvitedTeamIds()
		if err != nil {
			util.Info(err, " Cannot get invited team_ids")
			util.Report(w, r, "您好，茶博士都糊涂了，未能找到台主邀请品茶的茶团名单。")
			return
		}

		// 检测team_ids是否为空
		if len(team_ids) > 0 {
			// 迭代team_ids,用data.GetTeamMemberUserIdsByTeamId()获取全部user_ids；
			// 以UserId == u.Id检查当前用户是否是茶台的邀请团队成员
			for _, team_id := range team_ids {
				user_ids, err := data.GetMemberUserIdsByTeamId(team_id)
				if err != nil {
					util.Warning(err, " Cannot get team member user ids by team id")
					util.Report(w, r, "您好，茶博士迷路了，未能找到台主邀请的茶团成员列表。")
					return
				}
				for _, user_id := range user_ids {
					if user_id == u.Id {
						// 可以提出茶议
						util.GenerateHTML(w, &prDPD, "layout", "navbar.private", "thread.new")
						return
					}
				}
			}
		} else {
			// 不是受邀请团队成员，则不能提出茶议
			util.Report(w, r, "您好，茶博士彬彬有礼的说，大王你居然不在此茶台邀请品茶名单上！")
			return
		}

	default:
		util.Report(w, r, "您好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
		return
	}

}

// POST /v1/thread/draft
// 创建新茶议草稿，待邻座盲审后转为正式茶议
func DraftThread(w http.ResponseWriter, r *http.Request) {
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		util.Report(w, r, "您好，茶博士迷路了，未能找到你想要的茶台。")
		return
	}
	sUser, err := sess.User()
	if err != nil {
		util.Warning(err, " Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//读取表单数据
	ty, _ := strconv.Atoi(r.PostFormValue("type"))
	// 检查ty值是否 0、1、2
	switch ty {
	case 0, 1, 2:
		break
	default:
		util.Warning("Invalid thread type value")
		util.Report(w, r, "您好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	body := r.PostFormValue("topic")
	title := r.PostFormValue("title")
	project_id, _ := strconv.Atoi(r.PostFormValue("project_id"))

	//检查该茶台是否存在，而且状态不是草台状态
	project, err := data.GetProjectById(project_id)
	if err != nil {
		util.Warning(err, " Cannot get project")
		util.Report(w, r, "您好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	if project.Class == 10 || project.Class == 20 {
		util.Warning(sUser.Email, "试图访问未盲评审核的茶台被阻止。")
		util.Report(w, r, "您好，茶博士竟然说该茶台尚未启用，请确认后再试一次。")
		return
	}

	// 如果茶台class=1，存为开放式茶议草稿，
	// 如果茶台class=2， 存为封闭式茶议草稿
	if project.Class == 1 || project.Class == 2 {
		//检测一下title是否不为空，而且中文字数<24,topic不为空，而且中文字数<456
		if util.CnStrLen(title) < 1 {
			util.Report(w, r, "您好，茶博士竟然说该茶议标题为空，请确认后再试一次。")
			return
		}
		if util.CnStrLen(title) > 25 {
			util.Report(w, r, "您好，茶博士竟然说该茶议标题过长，请确认后再试一次。")
			return
		}
		if util.CnStrLen(body) < 1 {
			util.Report(w, r, "您好，茶博士竟然说该茶议内容为空，请确认后再试一次。")
			return
		} else if util.CnStrLen(body) < 17 {
			util.Report(w, r, "您好，茶博士竟然说该茶议内容太短，请确认后再试一次。")
			return
		}
		if util.CnStrLen(body) > 456 {
			util.Report(w, r, "您好，茶博士小声说茶棚的小纸条只能写456字，请确认后再试一次。")
			return
		}

		//保存新茶议草稿
		draft_thread := data.DraftThread{
			UserId:    sUser.Id,
			ProjectId: project_id,
			Title:     title,
			Body:      body,
			Class:     project.Class,
			Type:      ty,
		}
		if err = draft_thread.Create(); err != nil {
			util.Warning(err, " Cannot create thread draft")
			util.Report(w, r, "您好，茶博士没有墨水了，未能保存新茶议草稿。")
			return
		}
		// 创建一条友邻盲评,是否接纳 新茶的记录
		aO := data.AcceptObject{
			ObjectId:   draft_thread.Id,
			ObjectType: 3,
		}
		if err = aO.Create(); err != nil {
			util.Warning(err, "Cannot create accept_object")
			util.Report(w, r, "您好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
			return
		}
		// 发送盲评请求消息给两个在线用户
		//构造消息
		mess := data.AcceptMessage{
			FromUserId:     1,
			Title:          "新茶语邻座评审邀请",
			Content:        "茶博士隆重宣布：您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
			AcceptObjectId: aO.Id,
		}
		//发送消息
		if err = AcceptMessageSendExceptUserId(sUser.Id, mess); err != nil {
			util.Report(w, r, "您好，茶博士迷路了，未能发送盲评请求消息。")
			return
		}

		// 提示用户草稿保存成功
		t := fmt.Sprintf("您好，在 %s 茶台发布的茶议已准备妥当，稍等有缘茶友评审通过，即可昭告天下。", project.Title)
		// 提示用户草稿保存成功
		util.Report(w, r, t)
		return
	}
	util.Report(w, r, "您好，糊里糊涂的茶博士竟然说该茶台坐满了外星人，请确认后再试一次。")

}

// GET /v1/thread/detail
// 显示茶议（议题）的详细信息，包括品味（回复帖子）和记录品味的表格
func ThreadDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	vals := r.URL.Query()
	uuid := vals.Get("id")
	// 准备一个空白的表
	var thDPD data.ThreadDetailPageData
	// 读取茶议内容以填空
	thDPD.Thread, err = data.ThreadByUUID(uuid)
	if err != nil {
		util.Warning(err, " Cannot read thread")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取茶议。")
		return
	}
	//品味中颔首的合计得分与总得分的比值，取整数，用于客户端页面进度条设置，正反双方进展形势对比
	n1, err := thDPD.Thread.PostsScoreSupport()
	if err != nil {
		if err == sql.ErrNoRows {
			n1 = 0
		} else {
			util.Warning(err, " Cannot get posts score support")
			util.Report(w, r, "您好，莫失莫忘，仙寿永昌，有些资料被黑风怪瓜州了。")
			return
		}
	}
	n2, err := thDPD.Thread.PostsScore()
	if err != nil {
		if err == sql.ErrNoRows {
			n2 = 0
		} else {
			util.Warning(err, " Cannot get posts score oppose")
			util.Report(w, r, "您好，莫失莫忘，仙寿永昌，有些资料被黑风怪瓜州了。")
			return
		}
	}
	thDPD.ProgressSupport = data.ProgressRound(n1, n2)
	thDPD.ProgressOppose = 100 - thDPD.ProgressSupport
	// 读取全部回复帖子（品味）
	thDPD.PostList, err = thDPD.Thread.Posts()
	if err != nil {
		util.Warning(err, " Cannot read posts")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	// 读取会话
	sess, err := util.Session(r)
	if err != nil {
		// 游客
		// 检查茶议的级别状态
		if thDPD.Thread.Class == 1 || thDPD.Thread.Class == 2 {
			//记录茶议被点击数
			thDPD.Thread.AddHitCount()
			// 填写页面数据
			//tPD.IsGuest = true
			thDPD.Thread.PageData.IsAuthor = false
			thDPD.IsInput = true
			thDPD.SessUser = data.User{
				Id:   0,
				Name: "游客",
			}
			//迭代postList,标记非品味作者
			for i := range thDPD.PostList {
				thDPD.PostList[i].PageData.IsAuthor = false
			}

			//show the thread and the posts展示页面
			util.GenerateHTML(w, &thDPD, "layout", "navbar.public", "thread.detail")
			return
		} else {
			//非法访问未开放的话题？
			util.Warning(err, " 试图访问未公开的thread")
			util.Report(w, r, "茶水温度太高了，不适合品味，请稍后再试。")
			return
		}
	} else {
		//用户是登录状态,可以访问1和2级茶议
		if thDPD.Thread.Class == 1 || thDPD.Thread.Class == 2 {
			//从会话查获当前浏览用户资料
			sUser, err := sess.User()
			if err != nil {
				util.Warning(err, " Cannot get user from session")
				util.Report(w, r, "您好，茶博士失魂鱼，有眼不识泰山。")
				return
			}

			// 检测是否茶议作者
			if sUser.Id == thDPD.Thread.UserId {
				// 是茶议作者！
				// 填写页面数据
				thDPD.SessUser = sUser
				// tPD.IsGuest = false
				thDPD.Thread.PageData.IsAuthor = true
				// 提议作者不能自评品味，王婆卖瓜也不行？！
				thDPD.IsInput = false
				//点击数+1
				thDPD.Thread.AddHitCount()
				//记录用户阅读该帖子一次
				data.SaveReadedUserId(thDPD.Thread.Id, sUser.Id)
				//迭代PostList，把其PageData.IsAuthor设置为false，页面渲染时检测布局用
				for i := range thDPD.PostList {
					thDPD.PostList[i].PageData.IsAuthor = false
				}

				//show the thread and the posts展示页面
				util.GenerateHTML(w, &thDPD, "layout", "navbar.private", "thread.detail")
				return
			} else {
				//不是茶议作者
				//记录用户阅读该帖子一次
				data.SaveReadedUserId(thDPD.Thread.Id, sUser.Id)
				//记录茶议被点击数
				thDPD.Thread.AddHitCount()
				// 填写页面数据
				thDPD.SessUser = sUser
				//tPD.IsGuest = false
				thDPD.Thread.PageData.IsAuthor = false

				// 打开品味撰写
				thDPD.IsInput = true
				// 如果当前用户已经品味过了，则关闭撰写输入面板
				// 用于页面判断是否显示品味POST（回复）撰写面板
				for i := range thDPD.PostList {
					if thDPD.PostList[i].UserId == sUser.Id {
						thDPD.IsInput = false
						break
					}
				}

				// 检测是否其中某一个Post品味作者
				for i := range thDPD.PostList {
					if thDPD.PostList[i].UserId == sUser.Id {
						thDPD.PostList[i].PageData.IsAuthor = true
						break
					} else {
						thDPD.PostList[i].PageData.IsAuthor = false
					}
				}

				//展示茶议详情
				util.GenerateHTML(w, &thDPD, "layout", "navbar.private", "thread.detail")
				return
			}
		} else if thDPD.Thread.Class == 0 {
			//茶议的等级发生了变化，需要重新进行邻桌评估
			util.Report(w, r, "您好，茶议加水后出现了神迹！请耐心等待邻桌来推荐。")
			return
		} else {
			//访问未开放等级的话题？
			util.Report(w, r, "您好，外星人出没注意！请确认或者联系管理员确认稍后再试。")
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
	var thDPD data.ThreadDetailPageData
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	} else {
		// 读取当前访问用户资料
		sUser, err := sess.User()
		if err != nil {
			util.Warning(err, " Cannot get user from session")
			util.Report(w, r, "您好，茶博士失魂鱼，未能读取会话用户资料。")
			return
		}
		vals := r.URL.Query()
		uuid := vals.Get("id")
		thDPD.Thread, err = data.ThreadByUUID(uuid)
		if err != nil {
			util.Warning("Cannot not read thread")
			util.Report(w, r, "茶博士失魂鱼，未能读取茶议资料，请稍后再试。")
			return
		}
		if thDPD.Thread.UserId == sUser.Id {
			// 是作者，可以加水（补充内容）
			thDPD.Thread.PageData.IsAuthor = true
			thDPD.SessUser = sUser
			util.GenerateHTML(w, &thDPD, "layout", "navbar.private", "thread.edit")
			return
		}
		//不是作者，不能加水
		util.Danger("Cannot edit other user's thread")
		util.Report(w, r, "茶博士提示，目前仅能给自己的茶杯加水呢，补充说明自己的茶议貌似是合理的。")
		return

	}

}

// POST /v1/thread/update
// Update the thread 更新茶议内容
func UpdateThread(w http.ResponseWriter, r *http.Request) {
	// 测试时不启用追加方法？

	sess, err := util.Session(r)
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
			util.Report(w, r, "您好，茶博士失魂鱼，未能读取专属茶议。")
			return
		}
		uuid := r.PostFormValue("uuid")
		//title := r.PostFormValue("title")
		topi := r.PostFormValue("additional")
		//根据用户提供的uuid读取指定茶议
		thread, err := data.ThreadByUUID(uuid)
		if err != nil {
			util.Warning(err, " Cannot read thread by uuid")
			util.Report(w, r, "茶博士失魂鱼，未能读取专属茶议，请稍后再试。")
			return
		}
		//核对一下用户身份
		if thread.UserId == user.Id {
			//检查topi内容是否中文字数>17,并且thread.Topic总字数<456,如果是则可以补充内容
			if util.CnStrLen(topi) >= 17 && util.CnStrLen(thread.Body+topi) < 456 {
				thread.Body += topi
			} else {
				util.Info("Cannot update thread")
				util.Report(w, r, "猫是彬彬有礼的茶博士竟然声称纸太贵，字太少或者超过456字的茶议不收录！请确认后再试。")
				return
			}
			// 修改过的茶议,重置class=0,表示草稿状态，
			thread.Class = 0
			//许可修改自己的茶议
			if err := thread.UpdateTopicAndClass(thread.Body, thread.Class); err != nil {
				util.Danger(err, " Cannot update thread")
				util.Report(w, r, "茶博士失魂鱼，未能更新专属茶议，请稍后再试。")
				return
			}
			url := fmt.Sprint("/v1/thread/detail?id=", uuid)
			http.Redirect(w, r, url, http.StatusFound)
			return
		} else {
			//阻止修改别人的茶议
			util.Danger("Cannot edit other user's thread")
			util.Report(w, r, "茶博士提示，粗鲁的茶博士竟然说，仅能对自己的茶杯加水（追加内容）。")
			return
		}
	}
}

// GET /v1/thread/accept
// 展示邻桌盲评页面
func AcceptDraftThread(w http.ResponseWriter, r *http.Request) {
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 读取当前访问用户资料
	sUser, err := sess.User()
	if err != nil {
		util.Warning(err, " Cannot get user from session")
		util.Report(w, r, "您好，茶博士失魂鱼，未能读取会话用户资料。")
		return
	}
	//检查一下用户是否普通用户
	if sUser.Role != "traveller" {
		util.Danger("Cannot accept thread")
		util.Report(w, r, "茶博士提示，粗鲁的茶博士竟然强词夺理说，仅普通旅客可以查看新泡茶内容！")
		return
	}
	vals := r.URL.Query()
	id, err := strconv.Atoi(vals.Get("id"))
	if err != nil {
		util.Warning(err, " Cannot convert id to int")
		util.Report(w, r, "茶博士失魂鱼，未能读取新泡茶议资料，请稍后再试。")
		return
	}
	var dThread data.DThreadDetailPageData
	dThread.DraftThread = data.DraftThread{
		Id: id,
	}
	err = dThread.DraftThread.Get()
	if err != nil {
		util.Warning("Cannot not read thread draft")
		util.Report(w, r, "茶博士失魂鱼，未能找到新泡茶议资料，请稍后再试。")
		return
	}
	//如果是普通用户，并且是未评估的茶议草稿，就进入盲评页面
	if dThread.DraftThread.Class == 1 || dThread.DraftThread.Class == 2 {
		dThread.SessUser = sUser
		util.GenerateHTML(w, &dThread, "layout", "navbar.private", "thread.accept")
		return
	} else {
		//拒绝打开详情页面
		util.Danger("Cannot accept thread draft")
		util.Report(w, r, "茶博士提示，粗鲁的茶博士竟然强词夺理说，新泡茶内容已经被外星人拿走了。")
		return
	}

}

// POST /v1/thread/plus
// 附加型茶议，是针对某个品味（跟帖）发起的茶议草稿
func PlusThread(w http.ResponseWriter, r *http.Request) {
	sess, err := util.Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Report(w, r, "您好，未能找到你想要的茶台。")
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
	ty, _ := strconv.Atoi(r.PostFormValue("type"))
	// 检查ty值是否 0、1、2
	if ty != 0 && ty != 1 && ty != 2 {
		util.Report(w, r, "您好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	title := r.PostFormValue("title")
	body := r.PostFormValue("topic")
	post_uuid := r.PostFormValue("uuid")

	// 检查提交的参数是否合规
	//检测一下title是否不为空，而且中文字数<19,topic不为空，而且中文字数<456
	if util.CnStrLen(title) < 1 {
		util.Report(w, r, "您好，茶博士竟然说该茶议标题为空，请确认后再试一次。")
		return
	}
	if util.CnStrLen(title) > 19 {
		util.Report(w, r, "您好，茶博士竟然说该茶议标题过长，请确认后再试一次。")
		return
	}
	if util.CnStrLen(body) < 1 {
		util.Report(w, r, "您好，茶博士竟然说该茶议内容为空，请确认后再试一次。")
		return
	} else if util.CnStrLen(body) < 17 {
		util.Report(w, r, "您好，茶博士竟然说该茶议内容太短，请确认后再试一次。")
		return
	}
	if util.CnStrLen(body) > 456 {
		util.Report(w, r, "您好，茶博士小声说茶棚的小纸条只能写456字，请确认后再试一次。")
		return
	}
	// 检查是那一张茶台发生的事
	post, _ := data.GetPostByUuid(post_uuid)
	proj, err := post.Project()
	if err != nil {
		util.Warning(err, " Cannot get project")
		util.Report(w, r, "您好，鲁莽的茶博士竟然声称这个茶台被火星人顺走了。")
		return
	}
	switch proj.Class {
	case 10, 20:
		util.Report(w, r, "您好，该茶台尚未启用，请确认后再试。")
		return
	case 1:
		// Open teaTable, Can have tea :)
		// 保存新茶议草稿
		draft_thread := data.DraftThread{
			UserId:    sUser.Id,
			ProjectId: proj.Id,
			Title:     title,
			Body:      body,
			Class:     proj.Class,
			Type:      ty,
		}
		if err = draft_thread.Create(); err != nil {
			util.Warning(err, " Cannot create thread draft")
			util.Report(w, r, "您好，茶博士没有墨水了，未能保存新茶议草稿。")
			return
		}
	case 2:
		//Is the current session user allowed to join the tea tasting? It depends on the tea group members invited by the host.
		ok := isUserInvitedByProject(proj, sUser)
		if ok {
			draft_thread := data.DraftThread{
				UserId:    sUser.Id,
				ProjectId: proj.Id,
				Title:     title,
				Body:      body,
				Class:     proj.Class,
				Type:      ty,
			}
			if err = draft_thread.Create(); err != nil {
				util.Warning(err, " Cannot create thread draft")
				util.Report(w, r, "您好，茶博士的笔没有墨水了，未能保存新茶议草稿。")
				return
			}
		} else {
			util.Report(w, r, "您好，茶博士彬彬有礼的说，你的大名竟然不在邀请品茶名单上。")
			return
		}
		// 异常状态的茶台
	default:
		util.Report(w, r, "您好，茶博士满头大汗说，陛下你的大名竟然不在邀请品茶名单上。")
		return
	}

	// 发送邻座盲评消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶议邻座评审邀请",
		Content:        "您好，茶博士隆重宣布：您被茶棚选中为新茶议评审官啦，请及时处理。",
		AcceptObjectId: 3,
	}
	// 发送消息
	if err = AcceptMessageSendExceptUserId(sUser.Id, mess); err != nil {
		util.Report(w, r, "您好，茶博士迷路了，未能发送盲评请求消息。")
		return
	}
	// 提示用户草稿保存成功
	util.Report(w, r, "您好，您的新茶议已准备妥当，等有缘茶友品评之后，即可昭告宇宙。")

}
