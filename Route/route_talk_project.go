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
)

func HandleProjectPlace(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ProjectPlaceGet(w, r)
	case http.MethodPost:
		ProjectPlacePost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// POST /v1/project/place_update
func ProjectPlacePost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//检查会话用户身份是否见证者
	if !isVerifier(s_u.Id) {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取用户提交uuid参数
	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取目标茶台
	pr := data.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//读取提交的place_id参数
	place_id := r.PostFormValue("place_id")
	if place_id == "" {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//检查提交的place_id是否合法,是否正整数
	place_id_int, err := strconv.Atoi(place_id)
	if err != nil {
		util.Debug(" Cannot convert place_id to int", place_id, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//检查提交的place_id是否合法
	if place_id_int < 1 || place_id_int > 1000000000 {
		util.Debug(" Invalid place_id", place_id, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	old_place_id, err := pr.PlaceId()
	if err != nil {
		util.Debug(" Cannot get place_id", place_id, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	if place_id_int == old_place_id {
		report(w, r, "你好，陛下英明！但是茶台位置没有变化？请确认后再试。")
		return
	}

	//更新茶台地点
	pp := data.ProjectPlace{
		ProjectId: pr.Id,
		PlaceId:   place_id_int,
		UserId:    s_u.Id,
	}
	if err = pp.Create(); err != nil {
		util.Debug(" Cannot update place_id", place_id, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	//跳转茶台详情
	http.Redirect(w, r, "/v1/project/detail?uuid="+uuid, http.StatusFound)

}

// GET /v1/project/place_update?uuid=xXx
func ProjectPlaceGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//检查会话用户身份是否见证者
	if !isVerifier(s_u.Id) {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取用户查询参数
	uuid := r.FormValue("uuid")
	if uuid == "" {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取目标茶台
	pr := data.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//读取目标茶台地点
	place := data.ProjectPlace{ProjectId: pr.Id}
	if err = place.GetByProjectId(); err != nil {
		util.Debug(" Cannot get project place", uuid, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	prBean, err := fetchProjectBean(pr)
	if err != nil {
		util.Debug(" Cannot get project bean", uuid, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//读取目标茶围
	ob, err := pr.Objective()
	if err != nil {
		util.Debug(" Cannot get objective", ob.Id, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	obBean, err := fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot get objective bean", uuid, err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	var pD data.ProjectDetail
	pD.SessUser = s_u
	pD.IsVerifier = true
	pD.ProjectBean = prBean
	pD.QuoteObjectiveBean = obBean

	//渲染页面
	renderHTML(w, &pD, "layout", "navbar.private", "project.place_update", "component_sess_capacity", "component_project_bean")
}

// POST /v1/project/approve
// 茶话会(茶围)管理员选择某个茶台入围（入选/中标），
func ProjectApprove(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能记录入围茶台，请稍后再试。")
		return
	}
	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶台，请确认后再试。")
		return
	}

	//获取目标茶台
	pr := data.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid, err)
		report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶台，请确认后再试。")
		return
	}
	//读取目标茶围
	ob, err := pr.Objective()
	if err != nil {
		util.Debug(" Cannot get objective", ob.Id, err)
		report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		return
	}
	//检查用户是否有权限处理这个请求
	is_admin := false
	if ob.IsPrivate {
		admin_family, err := data.GetFamily(ob.FamilyId)
		if err != nil {
			util.Debug(" Cannot get family", ob.FamilyId, err)
			report(w, r, "你好，茶博士失魂鱼，未能找到茶话会举办方，请确认后再试。")
			return
		}
		is_admin, err = admin_family.IsMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot get family member", ob.FamilyId, err)
			report(w, r, "你好，茶博士失魂鱼，未能找到茶话会管理成员，请确认后再试。")
			return
		}
	} else {
		admin_team, err := data.GetTeam(ob.TeamId)
		if err != nil {
			util.Debug(" Cannot get team", ob.TeamId, err)
			report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
			return
		}
		is_admin, err = admin_team.IsMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot get team", ob.TeamId, err)
			report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
			return
		}
	}

	if !is_admin {
		//不是茶围管理员，无权处理
		report(w, r, "你好，茶博士面无表情，说没有权限处理这个入围操作，请确认。")
		return
	}

	// 准备记录入围的茶台
	new_project_approved := data.ProjectApproved{
		ObjectiveId: ob.Id,
		ProjectId:   pr.Id,
		UserId:      s_u.Id,
	}
	// 检查是否已经入围过了
	if err = new_project_approved.GetByObjectiveIdProjectId(); err == nil {
		report(w, r, "你好，茶博士微笑，已成功记录入围茶台，请勿重复操作。")
		return
	}
	if err = new_project_approved.Create(); err != nil {
		util.Debug(" Cannot create project approved", err)
		report(w, r, "你好，茶博士失魂鱼，未能记录入围茶台，请稍后再试。")
		return
	}

	// 预填充约茶...5部曲
	if err = data.CreateRequiredThreads(&ob, &pr, data.UserId_Verifier, r.Context()); err != nil {
		util.Debug(" Cannot create required threads", err)
		report(w, r, "你好，茶博士失魂鱼，未能预填充约茶5部曲，请稍后再试。")
		return
	}

	//跳转入围的茶台详情页面
	http.Redirect(w, r, "/v1/project/detail?uuid="+uuid, http.StatusFound)
}

// 处理新建茶台的操作处理器
func HandleNewProject(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//请求表单
		NewProjectGet(w, r)
	case http.MethodPost:
		//处理表单
		NewProjectPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// POST /v1/project/new
// 用户在某个指定茶话会新开一张茶台
func NewProjectPost(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取用户提交的表单数据
	title := r.PostFormValue("name")
	body := r.PostFormValue("description")
	ob_uuid := r.PostFormValue("ob_uuid")
	class, err := strconv.Atoi(r.PostFormValue("class"))
	if err != nil {
		util.Debug("Failed to convert class to int", err)
		return
	}
	team_id, err := strconv.Atoi(r.PostFormValue("team_id"))
	if err != nil {
		util.Debug(team_id, "Failed to convert team_id to int")
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.Debug("Failed to convert family_id to int", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"

	valid, err := validateTeamAndFamilyParams(is_private, team_id, family_id, s_u.Id, w, r)
	if !valid && err == nil {
		return // 参数不合法，已经处理了错误
	}
	if err != nil {
		// 处理数据库错误
		util.Debug("验证提交的团队和家庭id出现数据库错误", team_id, family_id, err)
		report(w, r, "你好，成员资格检查失败，请确认后再试。")
		return
	}
	//获取目标茶话会
	t_ob := data.Objective{Uuid: ob_uuid}
	if err = t_ob.GetByUuid(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			util.Debug("茶话会不存在", ob_uuid, err)
			report(w, r, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		} else {
			util.Debug("获取茶话会失败", ob_uuid, err)
			report(w, r, "你好，茶博士失魂鱼，系统繁忙，请稍后再试。")
		}
		return
	}
	// 检查在此茶围下是否已经存在相同名字的茶台
	count_title, err := data.CountProjectByTitleObjectiveId(title, t_ob.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		util.Debug(" Cannot get count of project by title and objective id", err)
		report(w, r, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//如果已经存在相同名字的茶台，返回错误信息
	if count_title > 0 {
		report(w, r, "你好，已经存在相同名字的茶台，请更换一个名称后再试。")
		return
	}

	place_uuid := r.PostFormValue("place_uuid")
	place := data.Place{Uuid: place_uuid}
	if err = place.GetByUuid(); err != nil {
		util.Debug(" Cannot get place", err)
		report(w, r, "你好，茶博士服务中，眼镜都模糊了，也未能找到你提交的喝茶地方资料，请确认后再试。")
		return
	}

	// 检测一下name是否>2中文字，desc是否在17-int(util.Config.ThreadMaxWord)中文字，
	// 如果不是，返回错误信息
	if cnStrLen(title) < 2 || cnStrLen(title) > 36 {
		util.Debug("Project name is too short", err)
		report(w, r, "你好，粗声粗气的茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if cnStrLen(body) < int(util.Config.ThreadMinWord) || cnStrLen(body) > int(util.Config.ThreadMaxWord) {
		util.Debug(" Project description is too long or too short", err)
		report(w, r, "你好，茶博士傻眼了，竟然说字数太少或者太多记不住，请确认后再试。")
		return
	}

	new_proj := data.Project{
		UserId:      s_u.Id,
		Title:       title,
		Body:        body,
		ObjectiveId: t_ob.Id,
		Class:       class,
		TeamId:      team_id,
		FamilyId:    family_id,
		IsPrivate:   is_private,
		Cover:       "default-pr-cover",
	}

	// 根据茶话会属性判断
	// 检查一下该茶话会是否草围（待蒙评审核状态）
	switch t_ob.Class {
	case data.ObClassOpenStraw, data.ObClassCloseStraw:
		// 该茶话会是草围,尚未启用，不能新开茶台
		report(w, r, "你好，这个茶话会尚未启用。")
		return

	case data.ObClassOpen:
		// 该茶话会是开放式茶话会，可以新开茶台
		// 检查提交的class值是否有效，必须为10或者20
		switch class {
		case data.ObClassOpenStraw:
			// 创建开放式草台
			if err = new_proj.Create(); err != nil {
				util.Debug(" Cannot create open project", err)
				report(w, r, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}

		case data.ObClassCloseStraw:
			tIds_str := r.PostFormValue("invite_ids")
			//用正则表达式检测一下s，是否符合“整数，整数，整数...”的格式
			if !verifyIdSliceFormat(tIds_str) {
				util.Debug(" TeamId slice format is wrong", err)
				report(w, r, "你好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
				return
			}
			//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId
			team_ids_str := strings.Split(tIds_str, ",")
			// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
			if len(team_ids_str) > int(util.Config.MaxInviteTeams) {
				util.Debug(" Too many team ids", err)
				report(w, r, "你好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，请确认后再试。")
				return
			}
			team_id_slice := make([]int, 0, util.Config.MaxInviteTeams)
			for _, te_id_str := range team_ids_str {
				t_id_int, _ := strconv.Atoi(te_id_str)
				team_id_slice = append(team_id_slice, t_id_int)
			}

			//创建封闭式草台
			if err = new_proj.Create(); err != nil {
				util.Debug(" Cannot create close project", err)
				report(w, r, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}
			// 迭代team_id_slice，尝试保存新封闭式茶台邀请的茶团
			for _, team_id := range team_id_slice {
				poInviTeams := data.ProjectInvitedTeam{
					ProjectId: new_proj.Id,
					TeamId:    team_id,
				}
				if err = poInviTeams.Create(); err != nil {
					util.Debug(" Cannot save invited teams", err)
					report(w, r, "你好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		default:
			report(w, r, "你好，茶博士摸摸头，说看不懂拟开新茶台是否封闭式，请确认。")
			return
		}

	case data.ObClassClose:
		// 封闭式茶话会
		// 检查用户是否可以在此茶话会下新开茶台
		ok, err := t_ob.IsInvitedMember(s_u.Id)
		if !ok {
			// 当前用户不是茶话会邀请团队成员，不能新开茶台
			util.Debug(" Cannot create project", err)
			report(w, r, "你好，茶博士惊讶地说，不是此茶话会邀请团队成员不能开新茶台，请确认。")
			return
		}
		// 当前用户是茶话会邀请团队成员，可以新开茶台
		if class == data.ObClassOpenStraw {
			report(w, r, "你好，封闭式茶话会内不能开启开放式茶台，请确认后再试。")
			return
		}
		if class == data.ObClassCloseStraw {
			tIds_str := r.PostFormValue("invite_ids")
			//用正则表达式检测一下s，是否符合“整数，整数，整数...”的格式
			if !verifyIdSliceFormat(tIds_str) {
				util.Debug(" TeamId slice format is wrong", err)
				report(w, r, "你好，茶博士迷糊了，竟然说填写的茶团号格式看不懂，请确认后再试。")
				return
			}
			//用户提交的team_id是以逗号分隔的字符串,需要分割后，转换成[]TeamId
			team_ids_str := strings.Split(tIds_str, ",")
			// 测试时，受邀请茶团Id数最多为maxInviteTeams设置限制数
			if len(team_ids_str) > int(util.Config.MaxInviteTeams) {
				util.Debug(" Too many team ids", err)
				report(w, r, "你好，茶博士摸摸头，竟然说指定的茶团数超过了茶棚最大限制数，开水不够用，请确认后再试。")
				return
			}
			team_id_slice := make([]int, 0, util.Config.MaxInviteTeams)
			for _, te_id_str := range team_ids_str {
				t_id_int, _ := strconv.Atoi(te_id_str)
				team_id_slice = append(team_id_slice, t_id_int)
			}

			//创建茶台
			if err = new_proj.Create(); err != nil {
				util.Debug("Cannot create project", err)
				report(w, r, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}
			// 迭代team_id_slice，尝试保存新封闭式茶台邀请的茶团
			for _, team_id := range team_id_slice {
				poInviTeams := data.ProjectInvitedTeam{
					ProjectId: new_proj.Id,
					TeamId:    team_id,
				}
				if err = poInviTeams.Create(); err != nil {
					util.Debug(" Cannot save invited teams", err)
					report(w, r, "你好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		}

	default:
		// 该茶话会属性不合法
		util.Debug(" Project class is not valid", err)
		report(w, r, "你好，茶博士摸摸头，竟然说这个茶话会被外星人霸占了，请确认后再试。")
		return
	}

	// 保存草台喝茶地方
	pp := data.ProjectPlace{
		ProjectId: new_proj.Id,
		PlaceId:   place.Id,
		UserId:    s_u.Id,
	}
	if err = pp.Create(); err != nil {
		util.Debug(" Cannot create project place", err)
		report(w, r, "你好，茶博士抹了抹汗，竟然说茶台地方保存失败，请确认后再试。")
		return
	}

	if util.Config.PoliteMode {

		if err = createAndSendAcceptMessage(new_proj.Id, data.AcceptObjectTypePr, s_u.Id); err != nil {
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				report(w, r, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				report(w, r, "你好，茶博士迷路了，未能发送蒙评请求消息。")
			}
			return
		}

		// 提示用户草台保存成功
		t := fmt.Sprintf("你好，新开茶话会 %s 已准备妥当，稍等有缘茶友评审通过之后，即可启用。", new_proj.Title)
		report(w, r, t)
		return
	} else {
		if err = acceptNewProject(new_proj.Id); err != nil {
			report(w, r, err.Error())
			return
		}
		//跳转到新茶台页面
		http.Redirect(w, r, fmt.Sprintf("/v1/project/detail?uuid=%s", new_proj.Uuid), http.StatusFound)
	}
}

// GET /v1/project/new?uuid=xxx
// 渲染创建新茶台表单页面
func NewProjectGet(w http.ResponseWriter, r *http.Request) {
	// 1. 检查用户会话
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// 2. 获取并验证茶话会UUID
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，茶博士失魂鱼，请指定要加入的茶话会。")
		return
	}

	// 3. 获取茶话会详情
	objective := data.Objective{Uuid: uuid}
	if err := objective.GetByUuid(); err != nil {
		util.Debug("获取茶话会失败", "uuid", uuid, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			report(w, r, "你好，茶博士失魂鱼，未能找到您指定的茶话会。")
		} else {
			report(w, r, "你好，茶博士失魂鱼，系统繁忙，请稍后再试。")
		}
		return
	}

	// 4. 获取用户相关数据
	sessUserData, err := prepareUserPageData(&sess)
	if err != nil {
		util.Debug("准备用户数据失败", "error", err)
		report(w, r, "你好，三人行，必有大佬焉，请稍后再试。")
		return
	}

	// 5. 准备页面数据
	obPageData, err := prepareObjectivePageData(objective, sessUserData)
	if err != nil {
		util.Debug("准备页面数据失败", "error", err)
		report(w, r, "你好，茶博士失魂鱼，未能找到茶围资料，请稍后再试。")
		return
	}
	// 6. 检查茶台创建权限
	if !checkCreateProjectPermission(objective, sessUserData.User.Id, w, r) {
		return
	}

	// 7. 渲染创建表单
	renderHTML(w, &obPageData, "layout", "navbar.private", "project.new", "component_avatar_name_gender")
}

// GET /v1/project/detail?uuid=
// 展示指定UUID茶台详情
func ProjectDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	var pD data.ProjectDetail
	// 读取用户提交的查询参数
	vals := r.URL.Query()
	uuid := vals.Get("uuid")

	pr := data.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			util.Debug("Project not found by uuid: ", uuid)
			report(w, r, "你好，荡昏寐，饮之以茶。请稍后再试。")
			return
		}
		util.Debug(" Cannot read project by uuid: ", uuid, ", error: ", err)
		report(w, r, "你好，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
		return
	}
	//检查project.Class=1 or 2,否则属于未经 友邻蒙评 通过的草稿，不允许查看
	if pr.Class != data.PrClassOpen && pr.Class != data.PrClassClose {
		report(w, r, "你好，荡昏寐，饮之以茶。请稍后再试。")
		return
	}

	pD.ProjectBean, err = fetchProjectBean(pr)
	if err != nil {
		util.Debug(" Cannot read projectbean by project:", pr.Uuid, err)
		report(w, r, "你好，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
		return
	}

	ob, err := pD.ProjectBean.Project.Objective()
	if err != nil {
		util.Debug(" Cannot read objective", err)
		report(w, r, "你好，松影一庭惟见鹤，梨花满地不闻莺。请稍后再试。")
		return
	}
	pD.QuoteObjectiveBean, err = fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot read objective", err)
		report(w, r, "你好，松影一庭惟见鹤，梨花满地不闻莺。请稍后再试。")
		return
	}
	// 截短此引用的茶围内容以方便展示
	pD.QuoteObjectiveBean.Objective.Body = subStr(pD.QuoteObjectiveBean.Objective.Body, 168)

	var tb_normal_slice []data.ThreadBean
	ctx := r.Context()
	thread_normal_slice, err := pD.ProjectBean.Project.ThreadsNormal(ctx)
	if err != nil {
		util.Debug(" Cannot read threads given project", err)
		report(w, r, "你好，倦绣佳人幽梦长，金笼鹦鹉唤茶汤。请稍后再试。")
		return
	}

	pD.ThreadCount = pr.NumReplies()
	if pD.ThreadCount > 12 {
		pD.IsOverTwelve = true
	} else {
		pD.IsOverTwelve = false
	}
	ta := data.ThreadApproved{
		ProjectId: pD.ProjectBean.Project.Id,
	}
	pD.ThreadIsApprovedCount = ta.CountByProjectId()

	tb_normal_slice, err = fetchThreadBeanSlice(thread_normal_slice, r)
	if err != nil {
		util.Debug(" Cannot read thread-bean slice", err)
		report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
		return
	}
	pD.ThreadBeanSlice = tb_normal_slice

	pD.IsApproved = pD.ProjectBean.IsApproved

	//如果入围，读取入围必备5threads
	if pD.IsApproved {
		threadAppointment, err := pr.ThreadAppointment(ctx)
		if err != nil {
			util.Debug(" Cannot read thread appointment", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		threadAppointmentBean, err := fetchThreadBean(threadAppointment, r)
		if err != nil {
			util.Debug(" Cannot read thread appointment bean", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.ApprovedFiveThreads.ThreadBeanAppointment = threadAppointmentBean
		threadSeeSeek_slice, err := pr.ThreadSeeSeek(ctx)
		if err != nil {
			util.Debug(" Cannot read thread see seek", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。，请稍后再试。")
			return
		}
		threadSeeSeekBean_slice, err := fetchThreadBeanSlice(threadSeeSeek_slice, r)
		if err != nil {
			util.Debug(" Cannot read thread see seek bean slice", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.ApprovedFiveThreads.ThreadBeanSeeSeekSlice = threadSeeSeekBean_slice
		threadSuggestion, err := pr.ThreadSuggestion(ctx)
		if err != nil {
			util.Debug(" Cannot read thread suggestion", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		threadSuggestionBean_slice, err := fetchThreadBeanSlice(threadSuggestion, r)
		if err != nil {
			util.Debug(" Cannot read thread suggestion bean slice", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.ApprovedFiveThreads.ThreadBeanSuggestionSlice = threadSuggestionBean_slice
		threadGoods_slice, err := pr.ThreadGoods(ctx)
		if err != nil {
			util.Debug(" Cannot read thread goods", err)
			report(w, r, "疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		threadGoodsBean_slice, err := fetchThreadBeanSlice(threadGoods_slice, r)
		if err != nil {
			util.Debug(" Cannot read thread goods bean slice", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.ApprovedFiveThreads.ThreadBeanGoodsSlice = threadGoodsBean_slice
		threadHandcraft_slice, err := pr.ThreadHandcraft(ctx)
		if err != nil {
			util.Debug(" Cannot read thread handcraft", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		threadHandcraftBean_slice, err := fetchThreadBeanSlice(threadHandcraft_slice, r)
		if err != nil {
			util.Debug(" Cannot read thread handcraft bean slice", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.ApprovedFiveThreads.ThreadBeanHandcraftSlice = threadHandcraftBean_slice

	}

	// 获取会话session
	s, err := session(r)
	if err != nil {
		// 未登录，游客
		pD.IsGuest = true

		pD.SessUser = data.User{
			Id:        data.UserId_None,
			Name:      "游客",
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		// 返回茶台详情游客页面
		renderHTML(w, &pD, "layout", "navbar.public", "project.detail", "component_thread_bean_approved", "component_thread_bean", "component_avatar_name_gender", "component_sess_capacity")
		return
	}

	// 已登陆用户

	//从会话查获当前浏览用户资料荚
	s_u, s_default_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := fetchSessionUserRelatedData(s)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", s.Email, err)
		report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
		return
	}

	pD.SessUser = s_u
	pD.SessUserDefaultFamily = s_default_family
	pD.SessUserSurvivalFamilies = s_survival_families
	pD.SessUserDefaultTeam = s_default_team
	pD.SessUserSurvivalTeams = s_survival_teams
	pD.SessUserDefaultPlace = s_default_place
	pD.SessUserBindPlaces = s_places

	//如果这是class=2封闭式茶台，需要检查当前浏览用户是否可以创建新茶议
	if pD.ProjectBean.Project.Class == data.PrClassClose {
		// 是封闭式茶台，需要检查当前用户身份是否受邀请茶团的成员，以决定是否允许发言
		ok, err := pD.ProjectBean.Project.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot check invited member", err)
			report(w, r, "你好，桃李明年能再发，明年闺中知有谁？你真的是受邀请茶团成员吗？")
			return
		}
		if ok {
			// 当前用户是本茶话会邀请$团队成员，可以新开茶议
			pD.IsInput = true
		} else {
			pD.IsInput = false
		}
	} else {
		// 开放式茶议，任何人都可以新开茶议
		pD.IsInput = true
	}

	//会话用户是否是作者
	if pD.ProjectBean.Project.UserId == s_u.Id {
		// 是作者
		pD.ProjectBean.Project.ActiveData.IsAuthor = true
	} else {
		// 不是作者
		pD.ProjectBean.Project.ActiveData.IsAuthor = false
	}

	is_master, err := checkProjectMasterPermission(&pr, s_u.Id)
	if err != nil {
		util.Debug("Permission check failed", "user_id:", s_u.Id, "error:", err)
		report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
		return
	}
	pD.IsMaster = is_master

	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("Admin permission check failed",
			"userId", s_u.Id,
			"objectiveId", ob.Id,
			"error", err,
		)
		report(w, r, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	pD.IsAdmin = is_admin

	if !pD.IsAdmin && !pD.IsMaster {
		veri_team := data.Team{Id: data.TeamIdVerifier}
		is_member, err := veri_team.IsMember(s_u.Id)
		if err != nil {
			util.Debug("Cannot check verifier team member", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
			return
		}
		if is_member {
			pD.IsVerifier = true
		}
	}

	// 用户足迹
	pD.SessUser.Footprint = r.URL.Path
	pD.SessUser.Query = r.URL.RawQuery

	renderHTML(w, &pD, "layout", "navbar.private", "project.detail", "component_thread_bean_approved", "component_thread_bean", "component_avatar_name_gender", "component_sess_capacity")
}
