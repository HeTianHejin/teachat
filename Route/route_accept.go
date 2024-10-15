package route

import (
	"database/sql"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
)

// Handler /v1/office/polite
// 通用友邻盲评页面，是否接受新茶语录
func Polite(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		PoliteGet(w, r)
	case "POST":
		PolitePost(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// Post /v1/office/polite
func PolitePost(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		util.Danger(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, _ := sess.User()
	err = r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "你好，(摸摸头想了又想),电脑去热带海岛度假了。")
		return
	}
	civilizer := r.PostFormValue("civilizer")
	care := r.PostFormValue("care")
	id_str := r.PostFormValue("id")
	// 把准备审核的对象茶语id_str转成id_int
	ao_id_int, err := strconv.Atoi(id_str)
	if err != nil {
		util.Danger(err, " Cannot get id")
		Report(w, r, "你好，(摸摸头想了又想), 你能否再给一次提示，这次该押阿根廷还是英格兰赢球？")
		return
	}

	//检查提交的参数是否合法
	switch civilizer {
	case "yes", "no":
		break
	default:
		util.Danger(err, " Cannot get form value civilizer")
		Report(w, r, "你好，(摸摸头想了又想),中文真是博大精深。")
		return
	}
	switch care {
	case "yes", "no":
		break
	default:
		util.Danger(err, " Cannot get form value care")
		Report(w, r, "你好，(摸摸头想了又想),请问这是火星文吗？")
		return
	}

	// 声明 1 友邻盲评 记录
	newAcceptance := data.Acceptance{
		AcceptObjectId: ao_id_int,
		XAccept:        false,
		XUserId:        s_u.Id,
		XAcceptedAt:    time.Now(),
		YAccept:        false,
		YUserId:        1,
		YAcceptedAt:    time.Now(),
	}

	// 权限检查。。。
	if !s_u.CheckHasReadAcceptMessage(ao_id_int) {
		Report(w, r, "你好，(摸摸头想了又想), 这里真的可以接受无票登陆星际飞船吗？")
		return
	}

	// 检查ao对象是否已经有记录，如果没有，说明是第一位友邻提交评判结果。如果有记录，说明是第二位友邻提交。
	oldAcceptance, err := newAcceptance.GetByAcceptObjectId()
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有记录
			if civilizer == "yes" && care == "yes" {
				// ok
				newAcceptance.XAccept = true
			} else {
				newAcceptance.XAccept = false
			}
			// 创建初记录
			if err = newAcceptance.Create(); err != nil {
				util.Danger(err, " Cannot create acceptance")
				Report(w, r, "你好，(摸摸头想了又想), 茴香豆的茴字真的有四种写法吗？")
				return
			}
			// 友邻盲评审茶首次记录完成
			Report(w, r, "好茶香护有缘人，感谢你出手维护文明秩序！")
			return
		} else {
			util.Danger(err, " Cannot check acceptance by ao_id")
			Report(w, r, "你好，(摸摸头想了又想)，去年今日此门中，人面桃花相映红。")
			return
		}
	}
	// err==nil说明已经有记录，说明是第二位审核官的提交
	// update旧记录
	if civilizer == "yes" && care == "yes" {
		// ok
		oldAcceptance.YAccept = true
	} else {
		oldAcceptance.YAccept = false
	}
	oldAcceptance.YUserId = s_u.Id
	oldAcceptance.YAcceptedAt = time.Now()
	if err = oldAcceptance.Update(); err != nil {
		util.Danger(err, " Cannot update acceptance")
		Report(w, r, "你好，(摸摸头想了又想),隔岸花分一脉香。")
		return
	}

	//以下是根据两个审核意见处理新茶语

	// 新声明一个审核对象，映射审核对象id
	ao := data.AcceptObject{
		Id: oldAcceptance.AcceptObjectId,
	}
	// 读取这个审核对象（根据审核对象id）
	if err = ao.Get(); err != nil {
		util.Danger(err, "Cannot get accept-object")
		Report(w, r, "你好，(摸摸头想了又想),得道多助，失道寡助。")
		return
	}
	// 检查新茶评审结果,如果任意一位友邻否定这是文明发言，就判断为不通过审核
	if !oldAcceptance.XAccept || !oldAcceptance.YAccept {

		//友邻盲评拒绝接纳这个茶语！Oh my...
		// 通知茶语主人友邻盲评结果为：婉拒！---没有通知
		// 根据对象类型处理
		switch ao.ObjectType {
		case 1:
			ob := data.Objective{
				Id: ao.ObjectId}
			if err = ob.GetById(); err != nil {
				util.Danger(err, "Cannot get objective")
				Report(w, r, "你好，茶博士失魂鱼，竟然说没有找到新茶评审的资料未必是怪事。")
				return
			}
			if ob.Class == 10 {
				ob.Class = 31
			} else if ob.Class == 20 {
				ob.Class = 32
			}
			// 更新茶话会，友邻盲评未通过！
			if err = ob.UpdateClass(); err != nil {
				util.Warning(err, "Cannot update ob class")
				Report(w, r, "你好，(摸摸头想了又想), 为什么踢足球的人都说临门一脚最麻烦呢？")
				return
			}
		case 2:
			pr := data.Project{
				Id: ao.ObjectId,
			}
			if err = pr.GetById(); err != nil {
				util.Danger(err, "Cannot get project")
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时找资料也是一种修养的过程。")
				return
			}
			if pr.Class == 10 {
				pr.Class = 31
			} else if pr.Class == 20 {
				pr.Class = 32
			}
			// 更新����，
			if err = pr.UpdateClass(); err != nil {
				util.Warning(err, "Cannot update pr class")
				Report(w, r, "你好，一畦春韭绿，十里稻花香。")
				return
			}
		case 3:
			dThread := data.DraftThread{
				Id: ao.ObjectId,
			}

			if err = dThread.GetById(); err != nil {
				util.Danger(err, "Cannot get dfart-thread")
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术,而是耐心。")
				return
			}
			// 更新������，友邻盲评 已拒绝公开发布
			if err = dThread.UpdateClass(2); err != nil {
				util.Warning(err, "Cannot update thread class")
				Report(w, r, "你好，睿藻仙才盈彩笔，自惭何敢再为辞。")
				return
			}
		case 4:
			dPost := data.DraftPost{
				Id: ao.ObjectId,
			}
			if err = dPost.GetById(); err != nil {
				util.Danger(err, "Cannot get draft-post")
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候 弄丢草稿的人不一定是诗人？")
				return
			}
			if err = dPost.UpdateDraftPost(2); err != nil {
				util.Danger(err, "Cannot update draft-post class")
				Report(w, r, "你好，宝鼎茶闲烟尚绿，幽窗棋罢指犹凉。")
				return
			}
		case 5:
			team, err := data.GetTeamById(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get team")
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
				return
			}
			if team.Class == 10 {
				team.Class = 31
			} else if team.Class == 20 {
				team.Class = 32
			}
			if err = team.UpdateClass(); err != nil {
				util.Danger(err, "Cannot update team class")
				Report(w, r, "你好，（摸摸头）考一考你，情中情因情感妹妹　错里错以错劝哥哥是什么茶品种？")
				return
			}
		}
		// 感谢友邻维护了茶棚的文明秩序
		Report(w, r, "好茶香护有缘人，感谢你出手维护文明品茶秩序！")
		return
	} else {
		// 两个审茶官都认为这是文明发言
		// 根据对象类型处理
		switch ao.ObjectType {
		case 1:
			ob := data.Objective{
				Id: ao.ObjectId}
			if err = ob.GetById(); err != nil {
				util.Danger(err, "Cannot get objective")
				Report(w, r, "你好，茶博士失魂鱼，竟然说没有找到新茶评审的资料未必是怪事。")
				return
			}
			if ob.Class == 10 {
				ob.Class = 1
			} else if ob.Class == 20 {
				ob.Class = 2
			}
			// 更新茶话会，友邻盲评已通过！
			if err = ob.UpdateClass(); err != nil {
				util.Warning(err, "Cannot update ob class")
				Report(w, r, "你好，(摸摸头想了又想), 为什么踢足球的人都说临门一脚最麻烦呢？")
				return
			}

		case 2:
			pr := data.Project{
				Id: ao.ObjectId,
			}
			if err = pr.GetById(); err != nil {
				util.Danger(err, "Cannot get project")
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时找资料也是一种修养的过程。")
				return
			}
			if pr.Class == 10 {
				pr.Class = 1
			} else if pr.Class == 20 {
				pr.Class = 2
			}
			// 更新����，友邻盲评已通过！
			if err = pr.UpdateClass(); err != nil {
				util.Warning(err, "Cannot update pr class")
				Report(w, r, "你好，一畦春韭绿，十里稻花香。")
				return
			}

		case 3:
			dThread := data.DraftThread{
				Id: ao.ObjectId,
			}

			if err = dThread.GetById(); err != nil {
				util.Danger(err, "Cannot get dfart-thread")
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术而是耐心。")
				return
			}
			// 更新������，友邻盲评已通过！
			if err = dThread.UpdateClass(1); err != nil {
				util.Warning(err, "Cannot update draft-thread class")
				Report(w, r, "你好，睿藻仙才盈彩笔，自惭何敢再为辞。")
				return
			}

			// 转为正式茶议稿,第一步，记录thread
			thread := data.Thread{
				Body:      dThread.Body,
				UserId:    dThread.UserId,
				Class:     1,
				Title:     dThread.Title,
				ProjectId: dThread.ProjectId,
				HitCount:  0,
				Type:      dThread.Type,
				PostId:    dThread.PostId,
				TeamId:    dThread.TeamId,
			}
			if err = thread.Create(); err != nil {
				util.Danger(err, "Cannot save thread")
				Report(w, r, "你好，吟成荳蔻才犹艳，睡足酴醾梦也香。")
				return
			}
			//转为正式茶议稿第二步，记录花费
			thread_cost := data.ThreadCost{
				UserId:    thread.UserId,
				ThreadId:  thread.Id,
				Cost:      dThread.Cost,
				Type:      0,
				CreatedAt: time.Now(),
				ProjectId: dThread.ProjectId,
			}
			if err = thread_cost.Create(); err != nil {
				util.Danger(err, thread.Id, "Cannot save thread_cost")
				Report(w, r, "你好，闪电考拉说，有时找资料快不一定好。")
				return
			}
			//记录计划用时（耗时）
			thread_time_slot := data.ThreadTimeSlot{
				UserId:    thread.UserId,
				ThreadId:  thread.Id,
				TimeSlot:  dThread.TimeSlot,
				IsConfirm: 0,
				CreatedAt: time.Now(),
				ProjectId: dThread.ProjectId,
			}
			if err = thread_time_slot.Create(); err != nil {
				util.Danger(err, thread.Id, "Cannot save thread_time_slot")
				Report(w, r, "你好，闪电考拉说，玉树则不大。")
				return
			}

		case 4:
			dPost := data.DraftPost{
				Id: ao.ObjectId,
			}
			if err = dPost.GetById(); err != nil {
				util.Danger(err, "Cannot get draft-post given acceptObject.object_id")
				Report(w, r, "你好，闪电考拉失魂鱼，竟然说有时候找资料的人不一定是外星人？")
				return
			}
			if err = dPost.UpdateDraftPost(1); err != nil {
				util.Danger(err, "Cannot update draft-post class")
				Report(w, r, "你好，宝鼎茶闲烟尚绿，幽窗棋罢指犹凉。")
				return
			}
			// 转为正式品味稿
			post_author, err := data.GetUserById(dPost.UserId)
			if err != nil {
				util.Danger(err, "Cannot get post_author given draftPost.user_id")
				Report(w, r, "你好，绕堤柳借三篙翠。")
				return
			}
			if _, err = post_author.CreatePost(dPost.ThreadId, dPost.TeamId, dPost.Attitude, dPost.Body); err != nil {
				util.Danger(err, "Cannot create post")
				Report(w, r, "你好，品茶是一种艺术，一杯为品，二杯为解渴。")
				return
			}

		case 5:
			team, err := data.GetTeamById(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get team")
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候临急抱佛脚，比刻苦奋斗更有用？")
				return
			}
			if team.Class == 10 {
				team.Class = 1
			} else if team.Class == 20 {
				team.Class = 2
			}
			// 更新��队，友邻盲评已通过！
			if err = team.UpdateClass(); err != nil {
				util.Danger(err, "Cannot update team class")
				Report(w, r, "你好，（摸摸头）考一考你，情中情因情感妹妹　错里错以错劝哥哥.是什么意思？")
				return
			}
			// 将team的Founder作为默认的CEO，teamMember.Role="CEO"
			teamMember := data.TeamMember{
				TeamId: team.Id,
				UserId: team.FounderId,
				Role:   "CEO",
				Class:  1,
			}
			if err = teamMember.Create(); err != nil {
				util.Danger(err, "Cannot create team-member")
				Report(w, r, "你好，花因喜洁难寻偶，人为悲秋易断魂。")
				return
			}
		case 6:
			group, err := data.GetGroup(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get group")
				Report(w, r, "你好，������失������，��然说有时�������������比��������更有用？")
				return
			}
			if group.Class == 10 {
				group.Class = 1
			} else if group.Class == 20 {
				group.Class = 2
			}
			// 更新，友评已通过！
			if err = group.Update(); err != nil {
				util.Danger(err, "Cannot update group class")
				Report(w, r, "你好，（����头）考一考你，情中情因情感������　错里错以错������是��么意思？")
				return
			}

		default:
			util.Danger(err, "Cannot get object")
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时候什么都不做,就能赢50%的竞争对手？")
			return
		}
		// 感谢友邻维护了茶棚的文明秩序
		Report(w, r, "好茶香护有缘人，感谢你出手维护文明品茶秩序！")
		return
	}
}

// Get /v1/office/polite
// 向用户返回“友邻盲评”审茶页面
func PoliteGet(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		util.Danger(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	su, _ := sess.User()

	var aopd data.AcceptObjectPageData

	// 读取提交的ID参数
	vals := r.URL.Query()
	ob_id_str := vals.Get("id")
	// 换成int
	ob_id, err := strconv.Atoi(ob_id_str)
	if err != nil {
		util.Warning(err, " Cannot get id")
		Report(w, r, "你好，换成int失败，评审的资料？")
		return
	}
	// 友邻盲评对象
	ao := data.AcceptObject{
		Id: ob_id,
	}
	if err = ao.Get(); err != nil {
		util.Danger(err, "Cannot get object")
		Report(w, r, "你好，茶博士都糊涂了，竟然唱问世间情为何物，直教人找不到对象？")
		return
	}

	//读取友邻盲评邀请函
	var acceptMessage data.AcceptMessage
	if err = acceptMessage.GetAccMesByUIdAndAOId(su.Id, ao.Id); err != nil {
		util.Warning(err, "Cannot get accept-message invitation")
		Report(w, r, "你好，茶博士莫名其妙，竟然说没有机票也可以登机有时候是合情合理的。")
		return
	}

	// 检查用户是否受邀请的新茶评审官
	if !su.CheckHasAcceptMessage(ao.Id) {
		util.Danger(err, "Cannot get accept new tea")
		Report(w, r, "你好，莫名其妙的茶博士竟然强词夺理说，外星人不能评估新茶～")
		return
	}
	// 根据对象类型处理
	switch ao.ObjectType {
	case 1:
		ob := data.Objective{
			Id: ao.ObjectId}
		if err = ob.GetById(); err != nil {
			util.Danger(err, "Cannot get objective")
			Report(w, r, "你好，有时候找不到新茶评审的资料未必是外星人闹事。")
			return
		}
		//aopd.Title = ob.Title
		aopd.Body = ob.Title + "." + ob.Body
		// 更新友邻盲评邀请函class为已读
		if err = acceptMessage.Update(su.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update ob accept-message class")
		}
	case 2:
		pr := data.Project{
			Id: ao.ObjectId,
		}
		if err = pr.GetById(); err != nil {
			util.Danger(err, "Cannot get project")
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时找资料也是一种修心养性的过程。")
			return
		}
		//aopd.Title = pr.Title
		aopd.Body = pr.Title + "." + pr.Body
		// 更新友邻盲评邀请函class为已读
		if err = acceptMessage.Update(su.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update pr accept-message class")
		}
	case 3:
		dThread := data.DraftThread{
			Id: ao.ObjectId,
		}
		if err = dThread.GetById(); err != nil {
			util.Danger(err, "Cannot get dfart-thread")
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术而是耐心。")
			return
		}

		aopd.Body = dThread.Title + "." + dThread.Body
		if err = acceptMessage.Update(su.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update draft-thread accept-message class")
		}
		// 更新友邻盲评邀请函class为已读
		acceptMessage.Update(su.Id, dThread.Id)
	case 4:
		dPost := data.DraftPost{
			Id: ao.ObjectId,
		}
		if err = dPost.GetById(); err != nil {
			util.Danger(err, "Cannot get post")
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料的人不一定是外星人？")
			return
		}
		aopd.Body = dPost.Body
		// 更新友邻盲评邀请函class为已读
		if err = acceptMessage.Update(su.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update po accept-message class")
		}
	case 5:
		team, err := data.GetTeamById(ao.ObjectId)
		if err != nil {
			util.Danger(err, "Cannot get team")
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
			return
		}
		//aopd.Title = team.Name
		aopd.Body = team.Name + "." + team.Mission
		// 更新友邻盲评邀请函class为已读
		if err = acceptMessage.Update(su.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update team accept-message class")
		}
	default:
		util.Danger(err, "Cannot get object")
		Report(w, r, "你好，茶博士失魂鱼，竟然说有时候什么都不做就能赢50%的竞争对手？")
		return
	}

	aopd.SessUser = su
	aopd.Id = ao.Id

	// 减少1新消息小黑板用户消息记录
	if err = data.SubtractUserMessageCount(su.Id); err != nil {
		util.Warning(err, "Cannot subtract 1 user message")
	}

	GenerateHTML(w, &aopd, "layout", "navbar.private", "watch_your_language")
}
