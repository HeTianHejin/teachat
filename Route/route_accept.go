package route

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
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
	sess, err := session(r)
	if err != nil {
		util.Debug("Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user given session", err)
		report(w, r, "你好，(摸摸头想了又想), 陛下能否再给一次提示，这次该押阿根廷还是英格兰赢球？")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，(摸摸头想了又想),电脑去热带海岛度假了。")
		return
	}
	civilizer := r.PostFormValue("civilizer")
	care := r.PostFormValue("care")
	id_str := r.PostFormValue("id")
	if id_str == "" {
		util.Debug(" Cannot get id", err)
		report(w, r, "你好，(摸摸头想了又想), 陛下能否再给一次提示，这次该押阿根廷还是英格兰赢球？")
		return
	}
	// 把准备审核的对象茶语id_str转成id_int
	ao_id_int, err := strconv.Atoi(id_str)
	if err != nil {
		util.Debug(" Cannot get id", err)
		report(w, r, "你好，(摸摸头想了又想), 陛下能否再给一次提示，这次该押阿根廷还是英格兰赢球？")
		return
	}

	//检查提交的参数是否合法
	switch civilizer {
	case "yes", "no":
		break
	default:
		util.Debug(" Cannot get form value civilizer", err)
		report(w, r, "你好，(茶博士摸摸头想了又想),喝茶文化真是博大精深。")
		return
	}
	switch care {
	case "yes", "no":
		break
	default:
		util.Debug(" Cannot get form value care", err)
		report(w, r, "你好，(摸摸头想了又想),陛下，请问这是火星文吗？")
		return
	}

	// 声明 1 友邻蒙评 记录
	newAcceptance := data.Acceptance{
		AcceptObjectId: ao_id_int,
		XAccept:        false,
		XUserId:        s_u.Id,
		YAccept:        false,
		YUserId:        data.UserId_None,
	}

	// 权限检查。。。
	ok, err := s_u.CheckHasReadAcceptMessage(ao_id_int)
	if err != nil {
		util.Debug("Cannot check acceptance", err)
		report(w, r, "你好，(茶博士摸摸头想了又想), 茴香豆的茴字真的有四种写法吗？")
		return
	}
	if !ok {
		util.Debug(" Cannot check acceptance by ao_id", err)
		report(w, r, "你好，(茶博士摸摸头想了又想), 这里真的可以接受无票喝茶吗？")
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
				report(w, r, "你好，(摸摸头想了又想), 茴香豆的茴字真的有四种写法吗？")
				return
			}
			// 友邻蒙评审茶首次记录完成
			report(w, r, "好茶香护有缘人，感谢你出手维护文明秩序！")
			return
		} else {
			util.Debug(" Cannot check acceptance by ao_id", err)
			report(w, r, "你好，(摸摸头想了又想)，去年今日此门中，人面桃花相映红。")
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
		report(w, r, "你好，(摸摸头想了又想),隔岸花分一脉香。")
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
		report(w, r, "你好，(茶博士摸摸头想了又想),居然说，得道多茶，失道寡茶。")
		return
	}
	// 检查新茶评审结果,如果任意一位友邻否定这是文明发言，就判断为不通过审核
	if !oldAcceptance.XAccept || !oldAcceptance.YAccept {
		//友邻蒙评拒绝接纳这个茶语！Oh my...
		// 通知茶语主人友邻蒙评结果为：婉拒！---没有通知
		// 根据对象类型处理
		switch ao.ObjectType {
		case data.AcceptObjectTypeObjective:
			ob := data.Objective{
				Id: ao.ObjectId}
			if err = ob.Get(); err != nil {
				util.Debug("Cannot get objective", err)
				report(w, r, "你好，茶博士失魂鱼，竟然说没有找到新茶茶叶的资料未必是怪事。")
				return
			}
			switch ob.Class {
			case data.ObClassOpenDraft:
				ob.Class = data.ObClassNeighborRejectOpen
			case data.ObClassCloseDraft:
				ob.Class = data.ObClassNeighborRejectClose
			}
			// 更新茶话会，友邻蒙评未通过！
			if err = ob.UpdateClass(); err != nil {
				util.Debug("Cannot update ob class", err)
				report(w, r, "你好，(摸摸头想了又想), 为什么踢足球的人都说临门一脚最麻烦呢？")
				return
			}
		case data.AcceptObjectTypeProject:
			pr := data.Project{
				Id: ao.ObjectId,
			}
			if err = pr.Get(); err != nil {
				util.Debug("Cannot get project", err)
				report(w, r, "你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
				return
			}
			switch pr.Class {
			case data.PrClassOpenDraft:
				pr.Class = data.PrClassRejectedOpen
			case data.PrClassCloseDraft:
				pr.Class = data.PrClassClose
			}
			// 更新茶台属性，
			if err = pr.UpdateClass(); err != nil {
				util.Debug("Cannot update pr class", err)
				report(w, r, "你好，一畦春韭绿，十里稻花香。")
				return
			}
		case data.AcceptObjectTypeThread:
			dThread := data.DraftThread{
				Id: ao.ObjectId,
			}

			if err = dThread.Get(); err != nil {
				util.Debug("Cannot get dfart-thread", err)
				report(w, r, "你好，茶博士失魂鱼，竟然说有时候找茶叶需要的不是技术,而是耐心。")
				return
			}
			// 更新茶议属性，友邻蒙评 已拒绝公开发布
			if err = dThread.UpdateStatus(data.DraftThreadStatusRejected); err != nil {
				util.Debug("Cannot update thread class", err)
				report(w, r, "你好，睿藻仙才盈彩笔，自惭何敢再为辞。")
				return
			}
		case data.AcceptObjectTypePost:
			dPost := data.DraftPost{
				Id: ao.ObjectId,
			}
			if err = dPost.Get(); err != nil {
				util.Debug("Cannot get draft-post", err)
				report(w, r, "你好，茶博士失魂鱼，竟然说有时候 弄丢草稿的人不一定是诗人？")
				return
			}
			if err = dPost.UpdateClass(data.DraftPostClassRejectedByNeighbor); err != nil {
				util.Debug("Cannot update draft-post class", err)
				report(w, r, "你好，宝鼎茶闲烟尚绿，幽窗棋罢指犹凉。")
				return
			}
		case data.AcceptObjectTypeTeam:
			team, err := data.GetTeam(ao.ObjectId)
			if err != nil {
				util.Debug("Cannot get team", err)
				report(w, r, "你好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
				return
			}
			switch team.Class {
			case data.TeamClassOpenDraft:
				team.Class = data.TeamClassRejectedOpenDraft
			case data.TeamClassCloseDraft:
				team.Class = data.TeamClassRejectedCloseDraft
			}
			if err = team.UpdateClass(); err != nil {
				util.Debug("Cannot update team class", err)
				report(w, r, "你好，（摸摸头）考一考你，错里错以错劝哥哥是什么茶品种？")
				return
			}
		}
		// 感谢友邻维护了茶棚的文明秩序
		report(w, r, "好茶香护有缘人，感谢你出手维护文明品茶秩序！")
		return
	} else {
		// 两个审茶官都认为这是文明发言
		// 根据对象类型处理
		switch ao.ObjectType {
		case data.AcceptObjectTypeObjective:
			if _, err = acceptNewObjective(ao.ObjectId); err != nil {
				report(w, r, err.Error())
				return
			}

		case data.AcceptObjectTypeProject:
			if err = acceptNewProject(ao.ObjectId); err != nil {
				report(w, r, err.Error())
				return
			}
		case data.AcceptObjectTypeThread:
			_, err := acceptNewDraftThread(ao.ObjectId)
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
		case data.AcceptObjectTypePost:
			_, err = acceptNewDraftPost(ao.ObjectId)
			if err != nil {
				switch {
				case strings.Contains(err.Error(), "获取品味草稿失败"):
					util.Debug("Cannot get draft-post", err)
					report(w, r, "你好，茶博士失魂鱼，竟然说有时候泡一壶好茶的关键，需要的不是技术而是耐心。")
				case strings.Contains(err.Error(), "创建新品味失败"):
					util.Debug("Cannot save post", err)
					report(w, r, "你好，吟成荳蔻才犹艳，睡足酴醾梦也香。")
				default:
					util.Debug("处理接纳新品味时发生未知错误", err)
					report(w, r, "世事洞明皆学问，人情练达即文章。")
				}
				return
			}

		case data.AcceptObjectTypeTeam:
			//把草团转为正式$事业茶团
			team, err := acceptNewTeam(ao.ObjectId)
			if err != nil {
				util.Debug("Cannot accept new team", err)
				report(w, r, "盛世无饥馑，何须耕织忙？不急不急。")
				return
			}

			// 将设立team的Founder作为默认的CEO角色成员，teamMember.Role=RoleCEO
			teamMember := data.TeamMember{
				TeamId: team.Id,
				UserId: team.FounderId,
				Role:   RoleCEO,
				Status: data.TeMemberStatusActive,
			}
			if err = teamMember.Create(); err != nil {
				util.Debug("Cannot create team-member", err)
				report(w, r, "你好，花因喜洁难寻偶，人为悲秋易断魂。")
				return
			}
			//检查团队发起人是否设置了（有效）非占位默认$茶团，
			//如果还没有，把这个新茶团设置为默认$茶团
			t_founder, err := data.GetUser(team.FounderId)
			if err != nil {
				util.Debug("Cannot get team founder", err)
				report(w, r, "你好，吟成荳蔻才犹艳，睡足酴醾梦也香。请稍后再试。")
				return
			}
			if !setUserDefaultTeam(&t_founder, team.Id, w, r) {
				return
			}

		case data.AcceptObjectTypeGroup:
			// 接纳新集团
			_, err := acceptNewGroup(ao.ObjectId)
			if err != nil {
				util.Debug("Cannot accept new team", err)
				report(w, r, "盛世无饥馑，何须耕织忙？不急不急。")
			}

		default:
			util.Debug("Cannot get object", err)
			report(w, r, "你好，茶博士失魂鱼，竟然说有时候喝茶比做傻事强？")
			return
		}

		// 感谢友邻维护了茶棚的文明秩序
		report(w, r, "好茶香护有缘人，感谢你出手维护文明品茶秩序！")
		return
	}
}

// Get /v1/office/polite?id=123456
// 向用户返回“友邻蒙评”审茶页面
func PoliteGet(w http.ResponseWriter, r *http.Request) {
	// 读取提交的ID参数
	ob_id_str := r.URL.Query().Get("id")
	if ob_id_str == "" {
		report(w, r, "你好，缺少编号参数，茶博士找不到茶叶的资料")
		return
	}
	// 换成int
	ob_id, err := strconv.Atoi(ob_id_str)
	if err != nil {
		util.Debug("Cannot convert id to integer", err)
		report(w, r, "你好，转换编号失败，茶博士找不到茶叶的资料")
		return
	}
	sess, err := session(r)
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

	// 友邻蒙评对象
	ao := data.AcceptObject{Id: ob_id}
	if err = ao.Get(); err != nil {
		util.Debug("Cannot get object", err)
		report(w, r, "你好，茶博士都糊涂了，竟然唱问世间情为何物，直教人找不到对象？")
		return
	}

	//读取友邻蒙评邀请函
	var acceptMessage data.AcceptMessage
	if err = acceptMessage.GetAccMesByUIdAndAOId(s_u.Id, ao.Id); err != nil {
		util.Debug("Cannot get accept-message invitation", err)
		report(w, r, "你好，茶博士莫名其妙，竟然说没有机票也登船有时候是合情合理的。")
		return
	}

	// 检查用户是否受邀请的新茶评审官
	ok := false
	ok, err = s_u.CheckHasAcceptMessage(ao.Id)
	if err != nil {
		util.Debug("CheckHasAcceptMessage failed given accept-object id", ao.Id)
		report(w, r, "你好，茶博士莫名其妙，竟然说没有机票也可以登船有时候是合情合理的。")
		return
	}
	if !ok {
		report(w, r, "你好，莫名其妙的茶博士竟然强词夺理说，外星人不能评估新茶～")
		return
	}
	// 根据对象类型处理
	switch ao.ObjectType {
	case data.AcceptObjectTypeObjective:
		ob := data.Objective{
			Id: ao.ObjectId}
		if err = ob.Get(); err != nil {
			util.Debug("Cannot get objective", err)
			report(w, r, "你好，有时候找不到新茶茶叶的资料未必是外星人闹事。")
			return
		}
		//aopd.Title = ob.Title
		aopd.Body = ob.Title + "." + ob.Body
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update ob accept-message class", err)
		}
	case data.AcceptObjectTypeProject:
		pr := data.Project{
			Id: ao.ObjectId,
		}
		if err = pr.Get(); err != nil {
			util.Debug("Cannot get project", err)
			report(w, r, "你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修心养性的过程。")
			return
		}
		//aopd.Title = pr.Title
		aopd.Body = pr.Title + "." + pr.Body
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update pr accept-message class", err)
		}
	case data.AcceptObjectTypeThread:
		dThread := data.DraftThread{
			Id: ao.ObjectId,
		}
		if err = dThread.Get(); err != nil {
			util.Debug("Cannot get dfart-thread", err)
			report(w, r, "你好，茶博士失魂鱼，竟然说有时候找茶叶需要的不是技术而是耐心。")
			return
		}

		aopd.Body = dThread.Title + "." + dThread.Body
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update draft-thread accept-message class", err)
		}
		// 更新友邻蒙评邀请函class为已读
		acceptMessage.Update(s_u.Id, dThread.Id)

	case data.AcceptObjectTypePost:
		dPost := data.DraftPost{
			Id: ao.ObjectId,
		}
		if err = dPost.Get(); err != nil {
			util.Debug("Cannot get post", err)
			report(w, r, "你好，茶博士失魂鱼，竟然说有时候找茶叶的人也会迷路。")
			return
		}
		aopd.Body = dPost.Body
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update po accept-message class", err)
		}

	case data.AcceptObjectTypeTeam:
		team, err := data.GetTeam(ao.ObjectId)
		if err != nil {
			util.Debug("Cannot get team", err)
			report(w, r, "你好，茶博士失魂鱼，竟然说有时候临急抱佛脚比刻苦奋斗更有用？")
			return
		}
		//aopd.Title = team.Name
		aopd.Body = team.Name + " " + team.Mission
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update team accept-message class", err)
			return
		}

	case data.AcceptObjectTypeGroup:
		group := data.Group{Id: ao.ObjectId}
		if err = group.Get(); err != nil {
			util.Debug("Cannot get group", err)
			report(w, r, "你好，满头大汗的茶博士请教你，乌龙茶是什么茶品种？")
			return
		}
		aopd.Body = group.Name + " " + group.Mission
		// 更新友邻蒙评邀请函class为已读
		if err = acceptMessage.Update(s_u.Id, ao.Id); err != nil {
			util.Debug("Cannot update group accept-message class", err)
			return
		}

	default:
		util.Debug("Cannot get object", err)
		report(w, r, "你好，茶博士失魂鱼，竟然说有时候喝茶比什么都不做好？")
		return
	}

	aopd.SessUser = s_u
	aopd.Id = ao.Id

	// 减少1新消息小黑板用户消息记录
	if err = data.SubtractUserMessageCount(s_u.Id); err != nil {
		util.Debug("Cannot subtract 1 user message", err)
	}

	generateHTML(w, &aopd, "layout", "navbar.private", "watch_your_language")
}
