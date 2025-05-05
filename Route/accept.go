package route

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
)

// Handler /v1/office/polite
// 通用友邻蒙评页面，是否接受新茶语录
func Polite(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		PoliteGet(w, r)
	case http.MethodPost:
		PolitePost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// Post /v1/office/polite
func PolitePost(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		util.Debug("Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user given session", err)
		Report(w, r, "你好，(摸摸头想了又想), 你能否再给一次提示，这次该押阿根廷还是英格兰赢球？")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		Report(w, r, "你好，(摸摸头想了又想),电脑去热带海岛度假了。")
		return
	}
	civilizer := r.PostFormValue("civilizer")
	care := r.PostFormValue("care")
	id_str := r.PostFormValue("id")
	if id_str == "" {
		util.Debug(" Cannot get id", err)
		Report(w, r, "你好，(摸摸头想了又想), 你能否再给一次提示，这次该押阿根廷还是英格兰赢球？")
		return
	}
	// 把准备审核的对象茶语id_str转成id_int
	ao_id_int, err := strconv.Atoi(id_str)
	if err != nil {
		util.Debug(" Cannot get id", err)
		Report(w, r, "你好，(摸摸头想了又想), 你能否再给一次提示，这次该押阿根廷还是英格兰赢球？")
		return
	}

	//检查提交的参数是否合法
	switch civilizer {
	case "yes", "no":
		break
	default:
		util.Debug(" Cannot get form value civilizer", err)
		Report(w, r, "你好，(茶博士摸摸头想了又想),无为有处有还无，中文真是博大精深。")
		return
	}
	switch care {
	case "yes", "no":
		break
	default:
		util.Debug(" Cannot get form value care", err)
		Report(w, r, "你好，(摸摸头想了又想),贾不假，请问这是火星文吗？")
		return
	}

	// 声明 1 友邻蒙评 记录
	newAcceptance := data.Acceptance{
		AcceptObjectId: ao_id_int,
		XAccept:        false,
		XUserId:        s_u.Id,
		YAccept:        false,
		YUserId:        0,
	}

	// 权限检查。。。
	ok, err := s_u.CheckHasReadAcceptMessage(ao_id_int)
	if err != nil {
		util.Debug("Cannot check acceptance", err)
		Report(w, r, "你好，(茶博士摸摸头想了又想), 茴香豆的茴字真的有四种写法吗？")
		return
	}
	if !ok {
		util.Debug(" Cannot check acceptance by ao_id", err)
		Report(w, r, "你好，(茶博士摸摸头想了又想), 这里真的可以接受无票登陆星际飞船吗？")
		return
	}

	// 检查ao对象是否已经有记录，如果没有，说明是第一位友邻提交评判结果。如果有记录，说明是第二位友邻提交。
	oldAcceptance, err := newAcceptance.GetByAcceptObjectId()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 没有记录
			if civilizer == "yes" && care == "yes" {
				// ok
				newAcceptance.XAccept = true
			} else {
				newAcceptance.XAccept = false
			}
			// 创建初记录
			if err = newAcceptance.Create(); err != nil {
				util.Debug(" Cannot create acceptance first", err)
				Report(w, r, "你好，(摸摸头想了又想), 茴香豆的茴字真的有四种写法吗？")
				return
			}
			// 友邻蒙评审茶首次记录完成
			Report(w, r, "好茶香护有缘人，感谢你出手维护文明秩序！")
			return
		} else {
			util.Debug(" Cannot check acceptance by ao_id", err)
			Report(w, r, "你好，(摸摸头想了又想)，去年今日此门中，人面桃花相映红。")
			return
		}
	}
	// err==nil说明已经有记录，说明是第二位茶语审核官的提交
	// update旧记录
	if civilizer == "yes" && care == "yes" {
		// ok
		oldAcceptance.YAccept = true
	} else {
		oldAcceptance.YAccept = false
	}
	oldAcceptance.YUserId = s_u.Id

	now_time := time.Now() //DeepSeek教的方法
	oldAcceptance.YAcceptedAt = &now_time

	if err = oldAcceptance.Update(); err != nil {
		util.Debug(" Cannot update acceptance", err)
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
		util.Debug("Cannot get accept-object", err)
		Report(w, r, "你好，(茶博士摸摸头想了又想),居然说，得道多茶，失道寡茶。")
		return
	}
	// 检查新茶评审结果,如果任意一位友邻否定这是文明发言，就判断为不通过审核
	if !oldAcceptance.XAccept || !oldAcceptance.YAccept {

		//友邻蒙评拒绝接纳这个茶语！Oh my...
		// 通知茶语主人友邻蒙评结果为：婉拒！---没有通知
		// 根据对象类型处理
		switch ao.ObjectType {
		case 1:
			ob := data.Objective{
				Id: ao.ObjectId}
			if err = ob.Get(); err != nil {
				util.Debug("Cannot get objective", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说没有找到新茶评审的资料未必是怪事。")
				return
			}
			if ob.Class == 10 {
				ob.Class = 31
			} else if ob.Class == 20 {
				ob.Class = 32
			}
			// 更新茶话会，友邻蒙评未通过！
			if err = ob.UpdateClass(); err != nil {
				util.Debug("Cannot update ob class", err)
				Report(w, r, "你好，(摸摸头想了又想), 为什么踢足球的人都说临门一脚最麻烦呢？")
				return
			}
		case 2:
			pr := data.Project{
				Id: ao.ObjectId,
			}
			if err = pr.Get(); err != nil {
				util.Debug("Cannot get project", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时找资料也是一种修养的过程。")
				return
			}
			if pr.Class == 10 {
				pr.Class = 31
			} else if pr.Class == 20 {
				pr.Class = 32
			}
			// 更新茶台属性，
			if err = pr.UpdateClass(); err != nil {
				util.Debug("Cannot update pr class", err)
				Report(w, r, "你好，一畦春韭绿，十里稻花香。")
				return
			}
		case 3:
			dThread := data.DraftThread{
				Id: ao.ObjectId,
			}

			if err = dThread.Get(); err != nil {
				util.Debug("Cannot get dfart-thread", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术,而是耐心。")
				return
			}
			// 更新茶议属性，友邻蒙评 已拒绝公开发布
			if err = dThread.UpdateClass(2); err != nil {
				util.Debug("Cannot update thread class", err)
				Report(w, r, "你好，睿藻仙才盈彩笔，自惭何敢再为辞。")
				return
			}
		case 4:
			dPost := data.DraftPost{
				Id: ao.ObjectId,
			}
			if err = dPost.Get(); err != nil {
				util.Debug("Cannot get draft-post", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候 弄丢草稿的人不一定是诗人？")
				return
			}
			if err = dPost.UpdateClass(403); err != nil {
				util.Debug("Cannot update draft-post class", err)
				Report(w, r, "你好，宝鼎茶闲烟尚绿，幽窗棋罢指犹凉。")
				return
			}
		case 5:
			team, err := data.GetTeam(ao.ObjectId)
			if err != nil {
				util.Debug("Cannot get team", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
				return
			}
			if team.Class == 10 {
				team.Class = 31
			} else if team.Class == 20 {
				team.Class = 32
			}
			if err = team.UpdateClass(); err != nil {
				util.Debug("Cannot update team class", err)
				Report(w, r, "你好，（摸摸头）考一考你，错里错以错劝哥哥是什么茶品种？")
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
			//茶围
			ob := data.Objective{
				Id: ao.ObjectId}
			if err = ob.Get(); err != nil {
				util.Debug("Cannot get objective", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说没有找到新茶评审的资料未必是怪事。")
				return
			}
			if ob.Class == 10 {
				ob.Class = 1
			} else if ob.Class == 20 {
				ob.Class = 2
			} else {
				//非法class值
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时找资料也是一种修养的过程。")
				return
			}
			// 更新茶话会，友邻蒙评已通过！
			if err = ob.UpdateClass(); err != nil {
				util.Debug("Cannot update ob class", err)
				Report(w, r, "你好，(摸摸头想了又想), 为什么踢足球的人都说临门一脚最麻烦呢？")
				return
			}

		case 2:
			//茶台
			pr := data.Project{
				Id: ao.ObjectId,
			}
			if err = pr.Get(); err != nil {
				util.Debug("Cannot get project", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时找资料也是一种修养的过程。")
				return
			}
			if pr.Class == 10 {
				pr.Class = 1
			} else if pr.Class == 20 {
				pr.Class = 2
			} else {
				//非法class值
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时找资料也是一种修养的过程。")
				return
			}
			// 更新茶台，友邻蒙评已通过！
			if err = pr.UpdateClass(); err != nil {
				util.Debug("Cannot update pr class", err)
				Report(w, r, "你好，一畦春韭绿，十里稻花香。")
				return
			}

		case 3:
			//茶议草稿
			dThread := data.DraftThread{
				Id: ao.ObjectId,
			}

			if err = dThread.Get(); err != nil {
				util.Debug("Cannot get dfart-thread", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术而是耐心。")
				return
			}
			// 更新茶议，友邻蒙评已通过！
			if err = dThread.UpdateClass(1); err != nil {
				util.Debug("Cannot update draft-thread class", err)
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
				FamilyId:  dThread.FamilyId,
				Type:      dThread.Type,
				PostId:    dThread.PostId,
				TeamId:    dThread.TeamId,
				IsPrivate: dThread.IsPrivate,
			}
			if err = thread.Create(); err != nil {
				util.Debug("Cannot save thread", err)
				Report(w, r, "你好，吟成荳蔻才犹艳，睡足酴醾梦也香。")
				return
			}

		case 4:
			//品味草稿
			dPost := data.DraftPost{
				Id: ao.ObjectId,
			}
			if err = dPost.Get(); err != nil {
				util.Debug("Cannot get draft-post given acceptObject.object_id", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料的人不一定是外星人？")
				return
			}

			// 转为正式品味稿
			new_post := data.Post{
				Body:      dPost.Body,
				UserId:    dPost.UserId,
				FamilyId:  dPost.FamilyId,
				TeamId:    dPost.TeamId,
				ThreadId:  dPost.ThreadId,
				IsPrivate: dPost.IsPrivate,
				Attitude:  dPost.Attitude,
				Class:     dPost.Class,
			}
			if err = new_post.Create(); err != nil {
				util.Debug("Cannot create post", err)
				Report(w, r, "你好，品茶是一种艺术，一杯为品，二杯为解渴。")
				return
			}

		case 5:
			//把草团转为正式$事业茶团
			team, err := data.GetTeam(ao.ObjectId)
			if err != nil {
				util.Debug("Cannot get team", err)
				Report(w, r, "你好，茶博士失魂鱼，竟然说有时候临急抱佛脚，比刻苦奋斗有用？")
				return
			}
			if team.Class == 10 {
				team.Class = 1
			} else if team.Class == 20 {
				team.Class = 2
			}
			// 更新团队属性，友邻蒙评已通过！
			if err = team.UpdateClass(); err != nil {
				util.Debug("Cannot update team class", err)
				Report(w, r, "你好，（摸摸头）考一考你，情中情因情感妹妹　错里错以错劝哥哥.是什么意思？")
				return
			}
			// 将team的Founder作为默认的CEO成员，teamMember.Role=RoleCEO
			teamMember := data.TeamMember{
				TeamId: team.Id,
				UserId: team.FounderId,
				Role:   RoleCEO,
				Class:  1,
			}
			if err = teamMember.Create(); err != nil {
				util.Debug("Cannot create team-member", err)
				Report(w, r, "你好，花因喜洁难寻偶，人为悲秋易断魂。")
				return
			}
			//检查团队发起人是否设置了默认$茶团，
			t_founder, err := data.GetUser(team.FounderId)
			if err != nil {
				util.Debug("Cannot get team founder", err)
				Report(w, r, "你好，茶博士失魂鱼，未能完成记录的任务，请稍后再试。")
				return
			}
			//如果还没有设置，把这个新茶团设置为默认$茶团
			old_default_team, err := t_founder.GetLastDefaultTeam()
			if err != nil {
				util.Debug(t_founder.Email, "Cannot get last default team")
				Report(w, r, "你好，茶博士失魂鱼，暂未能创建你的天命使团，请稍后再试。")
				return
			}
			//茶团2是“自由人$”，是系统预设占位茶团，
			if old_default_team.Id == 2 {
				// 没有设置默认茶团
				// 设置默认茶团
				uDT := data.UserDefaultTeam{
					UserId: t_founder.Id,
					TeamId: team.Id,
				}
				if err := uDT.Create(); err != nil {
					util.Debug(t_founder.Email, team.Id, "Cannot create default team")
					Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
					return
				}

			}
		case 6:
			//集团
			group, err := data.GetGroup(ao.ObjectId)
			if err != nil {
				util.Debug("Cannot get group", err)
				Report(w, r, "你好，满头大汗的茶博士请教你，错里错以错劝哥哥，是什么意思？")
				return
			}
			if group.Class == 10 {
				group.Class = 1
			} else if group.Class == 20 {
				group.Class = 2
			}
			// 更新，友评已通过！
			if err = group.Update(); err != nil {
				util.Debug("Cannot update group class", err)
				Report(w, r, "你好，满头大汗的茶博士问，情中情因情感妹妹是么意思？")
				return
			}

		default:
			util.Debug("Cannot get object", err)
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时候什么都不做,就能赢50%的竞争对手？")
			return
		}

		// 感谢友邻维护了茶棚的文明秩序
		Report(w, r, "好茶香护有缘人，感谢你出手维护文明品茶秩序！")
		return
	}
}

// Get /v1/office/polite
// 向用户返回“友邻蒙评”审茶页面
func PoliteGet(w http.ResponseWriter, r *http.Request) {
	sess, err := Session(r)
	if err != nil {
		util.Debug(" Cannot get session given session id", sess.Id)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get session given session id", sess.Id)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var aopd data.AcceptObjectPageData

	// 读取提交的ID参数
	vals := r.URL.Query()
	ob_id_str := vals.Get("id")
	if ob_id_str == "" {
		util.Debug("ID parameter is missing")
		Report(w, r, "你好，缺少编号参数，茶博士找不到评审的资料")
		return
	}
	// 换成int
	ob_id, err := strconv.Atoi(ob_id_str)
	if err != nil {
		util.Debug("Cannot convert id to integer", err)
		Report(w, r, "你好，转换编号失败，茶博士找不到评审的资料")
		return
	}
	// 友邻蒙评对象
	ao := data.AcceptObject{
		Id: ob_id,
	}
	if err = ao.Get(); err != nil {
		util.Debug("Cannot get object", err)
		Report(w, r, "你好，茶博士都糊涂了，竟然唱问世间情为何物，直教人找不到对象？")
		return
	}

	//读取友邻蒙评邀请函
	var acceptMessage data.AcceptMessage
	if err = acceptMessage.GetAccMesByUIdAndAOId(s_u.Id, ao.Id); err != nil {
		util.Debug("Cannot get accept-message invitation", err)
		Report(w, r, "你好，茶博士莫名其妙，竟然说没有机票也可以登船有时候是合情合理的。")
		return
	}

	// 检查用户是否受邀请的新茶评审官
	ok := false
	ok, err = s_u.CheckHasAcceptMessage(ao.Id)
	if err != nil {
		util.Debug("CheckHasAcceptMessage failed given accept-object id", ao.Id)
		Report(w, r, "你好，茶博士莫名其妙，竟然说没有机票也可以登船有时候是合情合理的。")
		return
	}
	if !ok {
		Report(w, r, "你好，莫名其妙的茶博士竟然强词夺理说，外星人不能评估新茶～")
		return
	}
	// 根据对象类型处理
	switch ao.ObjectType {
	case 1:
		ob := data.Objective{
			Id: ao.ObjectId}
		if err = ob.Get(); err != nil {
			util.Debug("Cannot get objective", err)
			Report(w, r, "你好，有时候找不到新茶评审的资料未必是外星人闹事。")
			return
		}
		//aopd.Title = ob.Title
		aopd.Body = ob.Title + "." + ob.Body
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update ob accept-message class", err)
		}
	case 2:
		pr := data.Project{
			Id: ao.ObjectId,
		}
		if err = pr.Get(); err != nil {
			util.Debug("Cannot get project", err)
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时找资料也是一种修心养性的过程。")
			return
		}
		//aopd.Title = pr.Title
		aopd.Body = pr.Title + "." + pr.Body
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update pr accept-message class", err)
		}
	case 3:
		dThread := data.DraftThread{
			Id: ao.ObjectId,
		}
		if err = dThread.Get(); err != nil {
			util.Debug("Cannot get dfart-thread", err)
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料需要的不是技术而是耐心。")
			return
		}

		aopd.Body = dThread.Title + "." + dThread.Body
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update draft-thread accept-message class", err)
		}
		// 更新友邻蒙评邀请函class为已读
		acceptMessage.Update(s_u.Id, dThread.Id)
	case 4:
		dPost := data.DraftPost{
			Id: ao.ObjectId,
		}
		if err = dPost.Get(); err != nil {
			util.Debug("Cannot get post", err)
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时候找资料的人也会迷路。")
			return
		}
		aopd.Body = dPost.Body
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update po accept-message class", err)
		}
	case 5:
		team, err := data.GetTeam(ao.ObjectId)
		if err != nil {
			util.Debug("Cannot get team", err)
			Report(w, r, "你好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
			return
		}
		//aopd.Title = team.Name
		aopd.Body = team.Name + "." + team.Mission
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update team accept-message class", err)
		}
	default:
		util.Debug("Cannot get object", err)
		Report(w, r, "你好，茶博士失魂鱼，竟然说有时候什么都不做就能赢50%的竞争对手？")
		return
	}

	aopd.SessUser = s_u
	aopd.Id = ao.Id

	// 减少1新消息小黑板用户消息记录
	if err = data.SubtractUserMessageCount(s_u.Id); err != nil {
		util.Debug("Cannot subtract 1 user message", err)
	}

	RenderHTML(w, &aopd, "layout", "navbar.private", "watch_your_language")
}
