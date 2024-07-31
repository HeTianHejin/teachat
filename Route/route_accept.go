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
	u, _ := sess.User()
	err = r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "您好，(摸摸头想了又想),电脑去热带海岛度假了。")
		return
	}
	civilizer := r.PostFormValue("civilizer")
	care := r.PostFormValue("care")
	id_str := r.PostFormValue("id")
	// ���成int
	ao_id_int, err := strconv.Atoi(id_str)
	if err != nil {
		util.Danger(err, " Cannot get id")
		Report(w, r, "您好，(摸摸头想了又想), 你能否再给一次提示，这次该押阿根廷还是英格兰赢球？")
		return
	}

	switch civilizer {
	case "yes", "no":
		break
	default:
		util.Danger(err, " Cannot get form value civilizer")
		Report(w, r, "您好，(摸摸头想了又想),中文真是博大精深。")
		return
	}
	switch care {
	case "yes", "no":
		break
	default:
		util.Danger(err, " Cannot get form value care")
		Report(w, r, "您好，(摸摸头想了又想),请问这是火星文吗？")
		return
	}

	// 声明 1 友邻盲评 记录
	newAcceptance := data.Acceptance{
		AcceptObjectId: ao_id_int,
		XAccept:        false,
		XUserId:        u.Id,
		XAcceptedAt:    time.Now(),
		YAccept:        false,
		YUserId:        1,
		YAcceptedAt:    time.Now(),
	}

	// 权限检查。。。
	if !u.CheckHasReadAcceptMessage(ao_id_int) {
		Report(w, r, "您好，(摸摸头想了又想), 这里真的可以接受无票登陆星际飞船吗？")
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
				Report(w, r, "您好，(摸摸头想了又想), 茴香豆的茴字真的有四种写法吗？")
				return
			}
			// 首次记录完成
			Report(w, r, "您好，谨代表茶棚礼仪委员会感谢你！出手维护茶棚的文明秩序。")
			return
		} else {
			util.Danger(err, " Cannot check acceptance by ao_id")
			Report(w, r, "您好，(摸摸头想了又想)，去年今日此门中，人面桃花相映红。")
			return
		}
	}
	// update旧记录
	if civilizer == "yes" && care == "yes" {
		// ok
		oldAcceptance.YAccept = true
	} else {
		oldAcceptance.YAccept = false
	}
	oldAcceptance.YUserId = u.Id
	oldAcceptance.YAcceptedAt = time.Now()
	if err = oldAcceptance.Update(); err != nil {
		util.Danger(err, " Cannot update acceptance")
		Report(w, r, "您好，(摸摸头想了又想),隔岸花分一脉香。")
		return
	}
	// 读取友邻盲评邀请函
	var acMess data.AcceptMessage
	if err = acMess.GetAccMesByUIdAndAOId(u.Id, oldAcceptance.AcceptObjectId); err != nil {
		util.Warning(err, "Cannot get accept-message invitation")
		Report(w, r, "您好，茶博士莫名其妙，竟然说没有票也可以潜水有时候是合情合理的。")
		return
	}
	ao := data.AcceptObject{
		Id: acMess.AcceptObjectId,
	}
	if err = ao.Get(); err != nil {
		util.Danger(err, "Cannot get accept-object")
		Report(w, r, "您好，(摸摸头想了又想),得道多助，失道寡助。")
		return
	}
	// 检查新茶评审结果
	if oldAcceptance.XAccept && oldAcceptance.YAccept {

		// 根据对象类型处理
		switch ao.ObjectType {
		case 1:
			ob, err := data.GetObjectiveById(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get objective")
				Report(w, r, "您好，茶博士失魂鱼，竟然说没有找到新茶评审的资料未必是怪事。")
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
				Report(w, r, "您好，(摸摸头想了又想), 为什么踢足球的人都说临门一脚最麻烦呢？")
				return
			}
			// 通知茶语作者友邻盲评已通过
			//构造消息
			// acMess.FromUserId = 1
			// acMess.ToUserId = ob.UserId
			// acMess.Title = "新茶语邻座评审结果通知"
			// acMess.Content = "您好，茶博士隆重宣布：您的新茶语已经通过了友邻评审，可以昭告宇宙啦。"
			// acMess.AcceptObjectId = ao.Id
			// acMess.Class = 0

			// //发送消息
			// if err = PilotAcceptMessageSend(ob.UserId, acMess); err != nil {
			// 	util.Pop_message(w, r, "您好，茶博士迷路了，未能发送盲评通过消息。")
			// 	return
			// }

		case 2:
			pr, err := data.GetProjectById(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get project")
				Report(w, r, "您好，茶博士失魂鱼，竟然说有时找资料也是一种修养的过程。")
				return
			}
			if pr.Class == 10 {
				pr.Class = 1
			} else if pr.Class == 20 {
				pr.Class = 2
			}
			// 更新����，友����评已通过！
			if err = pr.UpdateClass(); err != nil {
				util.Warning(err, "Cannot update pr class")
				Report(w, r, "您好，一畦春韭绿，十里稻花香。")
				return
			}
			// 通知茶语作者友邻盲评已通过
			//构造消息
			// acMess.FromUserId = 1
			// acMess.ToUserId = pr.UserId
			// acMess.Title = "新茶语邻座评审结果通知"
			// acMess.Content = "您好，茶博士隆重宣布：您的新茶语已经通过了友邻评审，可以昭告宇宙啦。"
			// acMess.AcceptObjectId = ao.Id
			// acMess.Class = 0

			// //发送消息
			// if err = PilotAcceptMessageSend(pr.UserId, acMess); err != nil {
			// 	util.Pop_message(w, r, "您好，茶博士迷路了，未能发送盲评请求消息。")
			// 	return
			// }

		case 3:
			dThread := data.DraftThread{
				Id: ao.ObjectId,
			}

			if err = dThread.Get(); err != nil {
				util.Danger(err, "Cannot get dfart-thread")
				Report(w, r, "您好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术而是耐心。")
				return
			}
			// 更新������，友����评已通过！
			if err = dThread.UpdateClass(1); err != nil {
				util.Warning(err, "Cannot update thread class")
				Report(w, r, "您好，睿藻仙才盈彩笔，自惭何敢再为辞。")
				return
			}
			// ���为正式品��稿
			thread := data.Thread{
				Body:      dThread.Body,
				UserId:    dThread.UserId,
				Class:     1,
				Title:     dThread.Title,
				ProjectId: dThread.ProjectId,
				HitCount:  0,
				Type:      dThread.Type,
				PostId:    dThread.PostId,
			}
			if err = thread.Save(); err != nil {
				util.Danger(err, "Cannot save thread")
				Report(w, r, "您好，吟成荳蔻才犹艳，睡足酴醾梦也香。")
				return
			}
			// 通知茶语作者友邻盲评已通过
			//构造消息
			// acMess.FromUserId = 1
			// acMess.ToUserId = thread.UserId
			// acMess.Title = "新茶语邻座评审结果通知"
			// acMess.Content = "您好，茶博士隆重宣布：您的新茶语已经通过了友邻评审，可以昭告宇宙啦。"
			// acMess.AcceptObjectId = ao.Id
			// acMess.Class = 0

			// //发送消息
			// if err = PilotAcceptMessageSend(thread.UserId, acMess); err != nil {
			// 	util.Pop_message(w, r, "您好，茶博士迷路了，未能发送盲评请求消息。")
			// 	return
			// }

		case 4:
			dPost, err := data.GetDraftPost(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get draft-post")
				Report(w, r, "您好，茶博士失魂鱼，竟然说有时候找资料的人不一定是外星人？")
				return
			}
			if err = dPost.UpdateDraftPost(1); err != nil {
				util.Danger(err, "Cannot update draft-post class")
				Report(w, r, "您好，宝鼎茶闲烟尚绿，幽窗棋罢指犹凉。")
				return
			}
			// 转为正式品味稿
			pUser, err := data.GetUserById(dPost.UserId)
			if err != nil {
				util.Danger(err, "Cannot get user")
				Report(w, r, "您好，绕堤柳借三篙翠。")
				return
			}
			if _, err = pUser.CreatePost(dPost.ThreadId, dPost.Attitude, dPost.Body); err != nil {
				util.Danger(err, "Cannot create post")
				Report(w, r, "您好，品茶是一种艺术，一杯为品，二杯为解渴。")
				return
			}

			// 通知茶语作者友邻盲评已通过
			//构造消息
			// acMess.FromUserId = 1
			// acMess.ToUserId = pUser.Id
			// acMess.Title = "新茶语邻座评审结果通知"
			// acMess.Content = "您好，茶博士隆重宣布：您的新茶语已经通过了友邻评审，可以昭告宇宙啦。"
			// acMess.AcceptObjectId = ao.Id
			// acMess.Class = 0

			// //发送消息
			// if err = PilotAcceptMessageSend(pUser.Id, acMess); err != nil {
			// 	util.Pop_message(w, r, "您好，茶博士迷路了，未能发送盲评请求消息。")
			// 	return
			// }

		case 5:
			team, err := data.GetTeamById(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get team")
				Report(w, r, "您好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
				return
			}
			if team.Class == 10 {
				team.Class = 1
			} else if team.Class == 20 {
				team.Class = 2
			}
			// 更新��队，友����评已通过！
			if err = team.UpdateClass(); err != nil {
				util.Danger(err, "Cannot update team class")
				Report(w, r, "您好，（摸摸头）考一考你，情中情因情感妹妹　错里错以错劝哥哥是什么意思？")
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
				Report(w, r, "您好，花因喜洁难寻偶，人为悲秋易断魂。")
				return
			}

			// 通知茶语作者友邻盲评已通过
			//构造消息
			// acMess.FromUserId = 1
			// acMess.ToUserId = team.FounderId
			// acMess.Title = "新茶语邻座评审结果通知"
			// acMess.Content = "您好，茶博士隆重宣布：您的新茶语已经通过了友邻评审，可以昭告宇宙啦。"
			// acMess.AcceptObjectId = ao.Id
			// acMess.Class = 0

			// //发送消息
			// if err = PilotAcceptMessageSend(team.FounderId, acMess); err != nil {
			// 	util.Pop_message(w, r, "您好，茶博士糊里糊涂，忘记通知书在大观园里了。")
			// 	return
			// }

		default:
			util.Danger(err, "Cannot get object")
			Report(w, r, "您好，茶博士失魂鱼，竟然说有时候什么都不做,就能赢50%的竞争对手？")
			return
		}
		// 感谢友邻维护了茶棚的文明秩序
		Report(w, r, "您好，好茶香护有缘人，感谢你出手维护了茶棚的文明秩序！")
		return
	} else {
		//友邻盲评拒绝接纳这个茶语！Oh my...
		// 通知茶语主人友邻盲评结果为：婉拒！---没有通知
		// 根据对象类型处理
		switch ao.ObjectType {
		case 1:
			ob, err := data.GetObjectiveById(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get objective")
				Report(w, r, "您好，茶博士失魂鱼，竟然说没有找到新茶评审的资料未必是怪事。")
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
				Report(w, r, "您好，(摸摸头想了又想), 为什么踢足球的人都说临门一脚最麻烦呢？")
				return
			}
		case 2:
			pr, err := data.GetProjectById(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get project")
				Report(w, r, "您好，茶博士失魂鱼，竟然说有时找资料也是一种修养的过程。")
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
				Report(w, r, "您好，一畦春韭绿，十里稻花香。")
				return
			}
		case 3:
			dThread := data.DraftThread{
				Id: ao.ObjectId,
			}

			if err = dThread.Get(); err != nil {
				util.Danger(err, "Cannot get dfart-thread")
				Report(w, r, "您好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术而是耐心。")
				return
			}
			// 更新������，友邻盲评 已退稿！
			if err = dThread.UpdateClass(2); err != nil {
				util.Warning(err, "Cannot update thread class")
				Report(w, r, "您好，睿藻仙才盈彩笔，自惭何敢再为辞。")
				return
			}
		case 4:
			dPost, err := data.GetDraftPost(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get draft-post")
				Report(w, r, "您好，茶博士失魂鱼，竟然说有时候 弄丢草稿的人不一定是诗人？")
				return
			}
			if err = dPost.UpdateDraftPost(2); err != nil {
				util.Danger(err, "Cannot update draft-post class")
				Report(w, r, "您好，宝鼎茶闲烟尚绿，幽窗棋罢指犹凉。")
				return
			}
		case 5:
			team, err := data.GetTeamById(ao.ObjectId)
			if err != nil {
				util.Danger(err, "Cannot get team")
				Report(w, r, "您好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
				return
			}
			if team.Class == 10 {
				team.Class = 31
			} else if team.Class == 20 {
				team.Class = 32
			}
			if err = team.UpdateClass(); err != nil {
				util.Danger(err, "Cannot update team class")
				Report(w, r, "您好，（摸摸头）考一考你，情中情因情感妹妹　错里错以错劝哥哥是什么茶品种？")
				return
			}
		}
		// 感谢友邻维护了茶棚的文明秩序
		Report(w, r, "您好，好茶香护有缘人，感谢你出手维护了茶棚的文明秩序！")
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
	u, _ := sess.User()
	// 读取提交的ID参数
	vals := r.URL.Query()
	ob_id_str := vals.Get("id")
	// 换成int
	ob_id, err := strconv.Atoi(ob_id_str)
	if err != nil {
		util.Warning(err, " Cannot get id")
		Report(w, r, "您好，换成int失败，评审的资料？")
		return
	}
	// 友邻盲评对象
	ao := data.AcceptObject{
		Id: ob_id,
	}
	if err = ao.Get(); err != nil {
		util.Danger(err, "Cannot get object")
		Report(w, r, "您好，茶博士都糊涂了，竟然唱问世间情为何物，直教人找不到对象？")
		return
	}

	//读取友邻盲评邀请函
	var acceptMessage data.AcceptMessage
	if err = acceptMessage.GetAccMesByUIdAndAOId(u.Id, ao.Id); err != nil {
		util.Warning(err, "Cannot get accept-message invitation")
		Report(w, r, "您好，茶博士莫名其妙，竟然说没有机票也可以登机有时候是合情合理的。")
		return
	}

	// 检查用户是否受邀请的新茶评审官
	if !u.CheckHasAcceptMessage(ao.Id) {
		util.Danger(err, "Cannot get accept new tea")
		Report(w, r, "您好，莫名其妙的茶博士竟然强词夺理说，外星人不能评估新茶～")
		return
	}
	// 根据对象类型处理
	switch ao.ObjectType {
	case 1:
		ob, err := data.GetObjectiveById(ao.ObjectId)
		if err != nil {
			util.Danger(err, "Cannot get objective")
			Report(w, r, "您好，有时候找不到新茶评审的资料未必是外星人闹事。")
			return
		}
		//ao.PageData.Title = ob.Title
		ao.PageData.Body = ob.Title + "." + ob.Body
		// 更新友邻盲评邀请函class为已读
		if err = acceptMessage.Update(u.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update ob accept-message class")
		}
	case 2:
		pr, err := data.GetProjectById(ao.ObjectId)
		if err != nil {
			util.Danger(err, "Cannot get project")
			Report(w, r, "您好，茶博士失魂鱼，竟然说有时找资料也是一种修心养性的过程。")
			return
		}
		//ao.PageData.Title = pr.Title
		ao.PageData.Body = pr.Title + "." + pr.Body
		// 更新友邻盲评邀请函class为已读
		if err = acceptMessage.Update(u.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update pr accept-message class")
		}
	case 3:
		thread, err := data.GetThreadById(ao.ObjectId)
		if err != nil {
			util.Danger(err, "Cannot get thread")
			Report(w, r, "您好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术而是耐心。")
			return
		}
		//ao.PageData.Title = thread.Title
		ao.PageData.Body = thread.Title + "." + thread.Body
		if err = acceptMessage.Update(u.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update th accept-message class")
		}
		// 更新友邻盲评邀请函class为已读
		acceptMessage.Update(u.Id, thread.Id)
	case 4:
		post, err := data.GetPostbyId(ao.ObjectId)
		if err != nil {
			util.Danger(err, "Cannot get post")
			Report(w, r, "您好，茶博士失魂鱼，竟然说有时候找资料的人不一定是外星人？")
			return
		}
		ao.PageData.Body = post.Body
		// 更新友����评��请��class为已读
		if err = acceptMessage.Update(u.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update po accept-message class")
		}
	case 5:
		team, err := data.GetTeamById(ao.ObjectId)
		if err != nil {
			util.Danger(err, "Cannot get team")
			Report(w, r, "您好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
			return
		}
		//ao.PageData.Title = team.Name
		ao.PageData.Body = team.Name + "." + team.Mission
		// 更新友����评��请��class为已读
		if err = acceptMessage.Update(u.Id, ao.Id); err != nil {
			util.Warning(err, "Cannot update team accept-message class")
		}
	default:
		util.Danger(err, "Cannot get object")
		Report(w, r, "您好，茶博士失魂鱼，竟然说有时候什么都不做就能赢50%的竞争对手？")
		return
	}

	ao.PageData.SessUser = u

	// 减少1新消息小黑板用户消息记录
	if err = data.SubtractUserMessageCount(u.Id); err != nil {
		util.Warning(err, "Cannot subtract 1 user message")
	}

	GenerateHTML(w, &ao, "layout", "navbar.private", "watch_your_language")
}
