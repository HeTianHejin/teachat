package route

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	dao "teachat/DAO"
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
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//检查会话用户身份是否见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取用户提交uuid参数
	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取目标茶台
	pr := dao.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//读取提交的place_id参数
	place_id := r.PostFormValue("place_id")
	if place_id == "" {
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//检查提交的place_id是否合法,是否正整数
	place_id_int, err := strconv.Atoi(place_id)
	if err != nil {
		util.Debug(" Cannot convert place_id to int", place_id, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//检查提交的place_id是否合法
	if place_id_int < 1 || place_id_int > 1000000000 {
		util.Debug(" Invalid place_id", place_id, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	old_place_id, err := pr.PlaceId()
	if err != nil {
		util.Debug(" Cannot get place_id", place_id, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	if place_id_int == old_place_id {
		report(w, s_u, "你好，陛下英明！但是茶台位置没有变化？请确认后再试。")
		return
	}

	//更新茶台地点
	pp := dao.ProjectPlace{
		ProjectId: pr.Id,
		PlaceId:   place_id_int,
		UserId:    s_u.Id,
	}
	if err = pp.Create(); err != nil {
		util.Debug(" Cannot update place_id", place_id, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
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
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//检查会话用户身份是否见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取用户查询参数
	uuid := r.FormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//获取目标茶台
	pr := dao.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//读取目标茶台地点
	place := dao.ProjectPlace{ProjectId: pr.Id}
	if err = place.GetByProjectId(); err != nil {
		util.Debug(" Cannot get project place", uuid, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	prBean, err := fetchProjectBean(pr)
	if err != nil {
		util.Debug(" Cannot get project bean", uuid, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//读取目标茶围
	ob, err := pr.Objective()
	if err != nil {
		util.Debug(" Cannot get objective", ob.Id, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	obBean, err := fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot get objective bean", uuid, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	var pD dao.ProjectDetail
	pD.SessUser = s_u
	pD.IsVerifier = true
	pD.ProjectBean = prBean
	pD.QuoteObjectiveBean = obBean

	//渲染页面
	generateHTML(w, &pD, "layout", "navbar.private", "project.place_update", "component_sess_capacity", "component_project_bean")
}

// POST /v1/project/approve/step1
// 茶话会(茶围)管理员选择某个茶台入围（入选/邀约），第一步，返回页面待确认监护方选择
func ProjectApproveStep1(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能记录入围茶台，请稍后再试。")
		return
	}
	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶台，请确认后再试。")
		return
	}

	//获取目标茶台
	pr := dao.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶台，请确认后再试。")
		return
	}
	//读取目标茶围
	ob, err := pr.Objective()
	if err != nil {
		util.Debug(" Cannot get objective", ob.Id, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		return
	}
	type pageData struct {
		SessUser    dao.User      `json:"sessUser"`
		Objetctive  dao.Objective `json:"objective"`
		Project     dao.Project   `json:"project"`
		AdminFamily dao.Family    `json:"admin_family"`
		AdminTeam   dao.Team      `json:"admin_team"`
	}
	pD := pageData{
		SessUser:   s_u,
		Objetctive: ob,
		Project:    pr,
	}
	//检查用户是否有权限处理这个请求
	is_admin := false
	if ob.IsPrivate {
		admin_family, err := dao.GetFamily(ob.FamilyId)
		if err != nil {
			util.Debug(" Cannot get family", ob.FamilyId, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到茶话会举办方，请确认后再试。")
			return
		}
		is_admin, err = admin_family.IsMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot get family member", ob.FamilyId, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到茶话会管理成员，请确认后再试。")
			return
		}
		pD.AdminFamily = admin_family
	} else {
		admin_team, err := dao.GetTeam(ob.TeamId)
		if err != nil {
			util.Debug(" Cannot get team", ob.TeamId, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
			return
		}
		is_admin, err = admin_team.IsMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot get team", ob.TeamId, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
			return
		}
		pD.AdminTeam = admin_team
	}

	if !is_admin {
		//不是茶围管理员，无权处理
		report(w, s_u, "你好，茶博士面无表情，说没有权限处理这个入围操作，请确认。")
		return
	}
	// 准备记录入围的茶台
	new_project_approved := dao.ProjectApproved{
		ObjectiveId: ob.Id,
		ProjectId:   pr.Id,
		UserId:      s_u.Id,
	}
	// 检查是否已经入围过了
	if err = new_project_approved.GetByObjectiveIdProjectId(); err == nil {
		report(w, s_u, "你好，茶博士微笑，已成功记录入围茶台，请勿重复操作。")
		return
	}

	// 返回入围所需要确认的监护方选择页面：project.approve_step1
	// 渲染选择页面，提供两个选项：
	// -自己监护：由当前用户所在家庭转为团队，担任监护方。
	// -茶棚分配：由系统自动匹配有相似解题技能的团队作为监护方。

	generateHTML(w, &pD, "layout", "navbar.private", "project.approve_step1", "component_avatar_name_gender")
}

// POST /v1/project/approve/step2
func ProjectApproveStep2(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能记录入围茶台，请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能记录入围茶台，请稍后再试。")
		return
	}
	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶台，请确认后再试。")
		return
	}
	guardian_type := r.PostFormValue("guardian_type")
	if guardian_type == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能记录入围茶台监护方，请确认后再试。")
		return
	}
	//获取目标茶台
	pr := dao.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		util.Debug(" Cannot get project", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶台，请确认后再试。")
		return
	}
	//读取目标茶围
	ob, err := pr.Objective()
	if err != nil {
		util.Debug(" Cannot get objective", ob.Id, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		return
	}
	//检查用户是否有权限处理这个请求
	is_admin := false
	if ob.IsPrivate {
		admin_family, err := dao.GetFamily(ob.FamilyId)
		if err != nil {
			util.Debug(" Cannot get family", ob.FamilyId, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到茶话会举办方，请确认后再试。")
			return
		}
		is_admin, err = admin_family.IsMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot get family member", ob.FamilyId, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到茶话会管理成员，请确认后再试。")
			return
		}
	} else {
		admin_team, err := dao.GetTeam(ob.TeamId)
		if err != nil {
			util.Debug(" Cannot get team", ob.TeamId, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
			return
		}
		is_admin, err = admin_team.IsMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot get team", ob.TeamId, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
			return
		}
	}
	if !is_admin {
		//不是茶围管理员，无权处理
		report(w, s_u, "你好，茶博士面无表情，说没有权限处理这个入围操作，请确认。")
		return
	}
	// 准备记录入围的茶台
	new_project_approved := dao.ProjectApproved{
		ObjectiveId: ob.Id,
		ProjectId:   pr.Id,
		UserId:      s_u.Id,
	}
	// 检查是否已经入围过了
	if err = new_project_approved.GetByObjectiveIdProjectId(); err == nil {
		report(w, s_u, "你好，茶博士微笑，已成功记录入围茶台，请勿重复操作。")
		return
	}
	// 根据监护方选择，处理入围茶台
	tea_order := dao.TeaOrder{}
	family_care_team_id := 0
	tea_order.ObjectiveId = ob.Id
	tea_order.ProjectId = pr.Id
	tea_order.Status = dao.TeaOrderStatusPending
	tea_order.VerifyTeamId = dao.TeamIdVerifier
	tea_order.PayeeTeamId = pr.TeamId
	// 以家庭父母成员担任CEO/CFO，创建监护团队，担任监护方团队。
	family_care_team_id, err = dao.ConvertFamilyToObCareTeam(ob.FamilyId, s_u, ob)
	if err != nil {
		util.Debug(" Cannot convert family to ob care team", ob.FamilyId, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到茶话会管理成员，请确认后再试。")
		return
	}
	tea_order.PayerTeamId = family_care_team_id
	switch strings.ToLower(guardian_type) {
	case "self":
		if ob.IsPrivate {
			// -如果归属管理的是家庭，
			tea_order.CareTeamId = family_care_team_id
		} else {
			// 直接登记为监护方团队
			tea_order.CareTeamId = ob.TeamId
		}
	case "system":
		tea_order.CareTeamId = dao.TeamIdNone //待后续选定
	default:
		// 提交了不能识别的参数
		report(w, s_u, "你好，茶博士看不懂你的书法，未能记录入围茶台监护方，请确认后再试。")
		return
	}

	// 记录tea_order
	// TODO：在见证团队页面，设置“待审批新茶定单”(tea_order.status=TeaOrderStatusPending)作为徽章通知-功能入口，
	// 等待见证者团队成员查看pending_tea_orders列表->审查（明确线下活动主题，道德合规。。。）
	if err = tea_order.Create(r.Context()); err != nil {
		util.Debug(" Cannot create tea order", ob.FamilyId, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建茶订单记录，请确认后再试。")
		return
	}

	// TODO：由系统自动匹配有相似解题技能的团队3个（如果有的话），让茶围归属管理方（出题方）选择1个作为监护方。
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
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
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
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	family_id, err := strconv.Atoi(r.PostFormValue("family_id"))
	if err != nil {
		util.Debug("Failed to convert family_id to int", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	//读取提交的is_private bool参数
	is_private := r.PostFormValue("is_private") == "true"

	valid, err := validateTeamAndFamilyParams(is_private, team_id, family_id, s_u, w)
	if !valid && err == nil {
		return // 参数不合法，已经处理了错误
	}
	if err != nil {
		// 处理数据库错误
		util.Debug("验证提交的团队和家庭id出现数据库错误", team_id, family_id, err)
		report(w, s_u, "你好，成员资格检查失败，请确认后再试。")
		return
	}
	//获取目标茶话会
	t_ob := dao.Objective{Uuid: ob_uuid}
	if err = t_ob.GetByUuid(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			util.Debug("茶话会不存在", ob_uuid, err)
			report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶话会，请确认后再试。")
		} else {
			util.Debug("获取茶话会失败", ob_uuid, err)
			report(w, s_u, "你好，茶博士失魂鱼，系统繁忙，请稍后再试。")
		}
		return
	}
	// 检查在此茶围下是否已经存在相同名字的茶台
	count_title, err := dao.CountProjectByTitleObjectiveId(title, t_ob.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		util.Debug(" Cannot get count of project by title and objective id", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	//如果已经存在相同名字的茶台，返回错误信息
	if count_title > 0 {
		report(w, s_u, "你好，已经存在相同名字的茶台，请更换一个名称后再试。")
		return
	}

	place_uuid := r.PostFormValue("place_uuid")
	place := dao.Place{Uuid: place_uuid}
	if err = place.GetByUuid(); err != nil {
		util.Debug(" Cannot get place", err)
		report(w, s_u, "你好，茶博士服务中，眼镜都模糊了，也未能找到你提交的喝茶地方资料，请确认后再试。")
		return
	}

	// 检测一下name是否>2中文字，desc是否在17-int(util.Config.ThreadMaxWord)中文字，
	// 如果不是，返回错误信息
	if cnStrLen(title) < 2 || cnStrLen(title) > 36 {
		util.Debug("Project name is too short", err)
		report(w, s_u, "你好，粗声粗气的茶博士竟然说字太少浪费纸张，请确认后再试。")
		return
	}
	if cnStrLen(body) < int(util.Config.ThreadMinWord) || cnStrLen(body) > int(util.Config.ThreadMaxWord) {
		util.Debug(" Project description is too long or too short", err)
		report(w, s_u, "你好，茶博士傻眼了，竟然说字数太少或者太多记不住，请确认后再试。")
		return
	}

	new_proj := dao.Project{
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
	case dao.ObClassOpenDraft, dao.ObClassCloseDraft:
		// 该茶话会是草围,尚未启用，不能新开茶台
		report(w, s_u, "你好，这个茶话会尚未启用。")
		return

	case dao.ObClassOpen:
		// 该茶话会是开放式茶话会，可以新开茶台
		// 检查提交的class值是否有效，必须为10或者20
		switch class {
		case dao.ObClassOpenDraft:
			// 创建开放式草台
			if err = new_proj.Create(); err != nil {
				util.Debug(" Cannot create open project", err)
				report(w, s_u, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}

		case dao.ObClassCloseDraft:
			tIds_str := r.PostFormValue("invite_ids")
			if tIds_str == "" {
				report(w, s_u, "你好，茶博士迷糊了，竟然说封闭式茶话会的茶团号不能省事不写，请确认后再试。")
				return
			}
			team_id_slice, err := parseIdSlice(tIds_str)
			if err != nil {
				report(w, s_u, "你好，陛下填写的茶团号格式看不懂，必需是不重复的自然数用英文逗号分隔。")
				return
			}

			//创建封闭式草台
			if err = new_proj.Create(); err != nil {
				util.Debug(" Cannot create close project", err)
				report(w, s_u, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}
			// 迭代team_id_slice，尝试保存新封闭式茶台邀请的茶团
			for _, team_id := range team_id_slice {
				poInviTeams := dao.ProjectInvitedTeam{
					ProjectId: new_proj.Id,
					TeamId:    team_id,
				}
				if err = poInviTeams.Create(); err != nil {
					util.Debug(" Cannot save invited teams", err)
					report(w, s_u, "你好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		default:
			report(w, s_u, "你好，茶博士摸摸头，说看不懂拟开新茶台是否封闭式，请确认。")
			return
		}

	case dao.ObClassClose:
		// 封闭式茶话会
		// 检查用户是否可以在此茶话会下新开茶台
		ok, err := t_ob.IsInvitedMember(s_u.Id)
		if !ok {
			// 当前用户不是茶话会邀请团队成员，不能新开茶台
			util.Debug(" Cannot create project", err)
			report(w, s_u, "你好，茶博士惊讶地说，不是此茶话会邀请团队成员不能开新茶台，请确认。")
			return
		}
		// 当前用户是茶话会邀请团队成员，可以新开茶台
		if class == dao.ObClassOpenDraft {
			report(w, s_u, "你好，封闭式茶话会内不能开启开放式茶台，请确认后再试。")
			return
		}
		if class == dao.ObClassCloseDraft {
			tIds_str := r.PostFormValue("invite_ids")
			if tIds_str == "" {
				report(w, s_u, "你好，茶博士迷糊了，竟然说封闭式茶话会的茶团号不能省事不写，请确认后再试。")
				return
			}
			team_id_slice, err := parseIdSlice(tIds_str)
			if err != nil {
				report(w, s_u, "你好，陛下填写的茶团号格式看不懂，必需是不重复的自然数用英文逗号分隔。")
				return
			}

			//创建茶台
			if err = new_proj.Create(); err != nil {
				util.Debug("Cannot create project", err)
				report(w, s_u, "你好，出浴太真冰作影，捧心西子玉为魂。")
				return
			}
			// 迭代team_id_slice，尝试保存新封闭式茶台邀请的茶团
			for _, team_id := range team_id_slice {
				poInviTeams := dao.ProjectInvitedTeam{
					ProjectId: new_proj.Id,
					TeamId:    team_id,
				}
				if err = poInviTeams.Create(); err != nil {
					util.Debug(" Cannot save invited teams", err)
					report(w, s_u, "你好，受邀请的茶团名单竟然保存失败，请确认后再试。")
					return
				}
			}
		}

	default:
		// 该茶话会属性不合法
		util.Debug(" Project class is not valid", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个茶话会被外星人霸占了，请确认后再试。")
		return
	}

	// 保存草台喝茶地方
	pp := dao.ProjectPlace{
		ProjectId: new_proj.Id,
		PlaceId:   place.Id,
		UserId:    s_u.Id,
	}
	if err = pp.Create(); err != nil {
		util.Debug(" Cannot create project place", err)
		report(w, s_u, "你好，茶博士抹了抹汗，竟然说茶台地方保存失败，请确认后再试。")
		return
	}

	if util.Config.PoliteMode {

		if err = createAndSendAcceptNotification(new_proj.Id, dao.AcceptObjectTypeProject, s_u.Id, r.Context()); err != nil {
			if strings.Contains(err.Error(), "创建AcceptObject失败") {
				report(w, s_u, "你好，胭脂洗出秋阶影，冰雪招来露砌魂。")
			} else {
				report(w, s_u, "你好，茶博士迷路了，未能发送蒙评请求通知。")
			}
			return
		}

		// 提示用户草台保存成功
		t := fmt.Sprintf("你好，新开茶话会 %s 已准备妥当，稍等有缘茶友评审通过之后，即可启用。", new_proj.Title)
		report(w, s_u, t)
		return
	} else {
		if err = acceptNewProject(new_proj.Id); err != nil {
			report(w, s_u, err.Error())
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
	s_u := dao.UserUnknown
	// 2. 获取并验证茶话会UUID
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，请指定要加入的茶话会。")
		return
	}

	// 3. 获取茶话会详情
	objective := dao.Objective{Uuid: uuid}
	if err := objective.GetByUuid(); err != nil {
		util.Debug("获取茶话会失败", "uuid", uuid, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			report(w, s_u, "你好，茶博士失魂鱼，未能找到您指定的茶话会。")
		} else {
			report(w, s_u, "你好，茶博士失魂鱼，系统繁忙，请稍后再试。")
		}
		return
	}

	// 4. 获取用户相关数据
	sessUserData, err := prepareUserPageData(&sess)
	if err != nil {
		util.Debug("准备用户数据失败", "error", err)
		report(w, s_u, "你好，三人行，必有大佬焉，请稍后再试。")
		return
	}

	// 5. 准备页面数据
	obPageData, err := prepareObjectivePageData(objective, sessUserData)
	if err != nil {
		util.Debug("准备页面数据失败", "error", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到茶围资料，请稍后再试。")
		return
	}
	// 6. 检查茶台创建权限
	if !checkCreateProjectPermission(objective, sessUserData.User, w) {
		return
	}

	// 7. 渲染创建表单
	generateHTML(w, &obPageData, "layout", "navbar.private", "project.new", "component_avatar_name_gender")
}

// GET /v1/project/detail?uuid=
// 展示指定UUID茶台详情
func ProjectDetail(w http.ResponseWriter, r *http.Request) {
	var err error
	var pD dao.ProjectDetail
	s_u := dao.UserUnknown
	// 读取用户提交的查询参数
	vals := r.URL.Query()
	uuid := vals.Get("uuid")

	pr := dao.Project{Uuid: uuid}
	if err = pr.GetByUuid(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			util.Debug("Project not found by uuid: ", uuid)
			report(w, s_u, "你好，荡昏寐，饮之以茶。请稍后再试。")
			return
		}
		util.Debug(" Cannot read project by uuid: ", uuid, ", error: ", err)
		report(w, s_u, "你好，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
		return
	}
	//检查project.Class=1 or 2,否则属于未经 友邻蒙评 通过的草稿，不允许查看
	if pr.Class != dao.PrClassOpen && pr.Class != dao.PrClassClose {
		report(w, s_u, "你好，荡昏寐，饮之以茶。请稍后再试。")
		return
	}

	pD.ProjectBean, err = fetchProjectBean(pr)
	if err != nil {
		util.Debug(" Cannot read projectbean by project:", pr.Uuid, err)
		report(w, s_u, "你好，松影一庭惟见鹤，梨花满地不闻莺，请稍后再试。")
		return
	}

	ob, err := pD.ProjectBean.Project.Objective()
	if err != nil {
		util.Debug(" Cannot read objective", err)
		report(w, s_u, "你好，松影一庭惟见鹤，梨花满地不闻莺。请稍后再试。")
		return
	}
	pD.QuoteObjectiveBean, err = fetchObjectiveBean(ob)
	if err != nil {
		util.Debug(" Cannot read objective", err)
		report(w, s_u, "你好，松影一庭惟见鹤，梨花满地不闻莺。请稍后再试。")
		return
	}
	// 截短此引用的茶围内容以方便展示
	pD.QuoteObjectiveBean.Objective.Body = subStr(pD.QuoteObjectiveBean.Objective.Body, 168)

	var tb_normal_slice []dao.ThreadBean
	ctx := r.Context()
	thread_normal_slice, err := pD.ProjectBean.Project.ThreadsNormal(ctx)
	if err != nil {
		util.Debug(" Cannot read threads given project", err)
		report(w, s_u, "你好，倦绣佳人幽梦长，金笼鹦鹉唤茶汤。请稍后再试。")
		return
	}

	pD.ThreadCount = pr.NumReplies()
	if pD.ThreadCount > 12 {
		pD.IsOverTwelve = true
	} else {
		pD.IsOverTwelve = false
	}
	ta := dao.ThreadApproved{
		ProjectId: pD.ProjectBean.Project.Id,
	}
	pD.ThreadIsApprovedCount = ta.CountByProjectId()

	tb_normal_slice, err = fetchThreadBeanSlice(thread_normal_slice, r)
	if err != nil {
		util.Debug(" Cannot read thread-bean slice", err)
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
		return
	}
	pD.ThreadBeanSlice = tb_normal_slice

	pD.IsApproved = pD.ProjectBean.IsApproved

	//如果入围，读取入围必备6threads
	if pD.IsApproved {

		thread_appo, err := pr.ThreadAppointment(ctx)
		if err != nil {
			util.Debug(" Cannot read thread appointment", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		thread_appo_bean, err := fetchThreadBean(thread_appo, r)
		if err != nil {
			util.Debug(" Cannot read thread appointment bean", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.Approved6Threads.ThreadBeanAppointment = thread_appo_bean

		thread_seeseek_slice, err := pr.ThreadsSeeSeek(ctx)
		if err != nil {
			util.Debug(" Cannot read thread see seek", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。，请稍后再试。")
			return
		}
		thread_seeseek_bean_slice, err := fetchThreadBeanSlice(thread_seeseek_slice, r)
		if err != nil {
			util.Debug(" Cannot read thread see seek bean slice", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.Approved6Threads.ThreadBeanSeeSeekSlice = thread_seeseek_bean_slice

		thread_brain_fire_slice, err := pr.ThreadsBrainFire(ctx)
		if err != nil {
			util.Debug(" Cannot read thread brain fire", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		thread_brainfire_bean_slice, err := fetchThreadBeanSlice(thread_brain_fire_slice, r)
		if err != nil {
			util.Debug(" Cannot read thread brain fire bean", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.Approved6Threads.ThreadBeanBrainFireSlice = thread_brainfire_bean_slice

		thread_suggestion_slice, err := pr.ThreadsSuggestion(ctx)
		if err != nil {
			util.Debug(" Cannot read thread suggestion", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		thread_suggestion_bean_slice, err := fetchThreadBeanSlice(thread_suggestion_slice, r)
		if err != nil {
			util.Debug(" Cannot read thread suggestion bean slice", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.Approved6Threads.ThreadBeanSuggestionSlice = thread_suggestion_bean_slice

		thread_goods_slice, err := pr.ThreadsGoods(ctx)
		if err != nil {
			util.Debug(" Cannot read thread goods", err)
			report(w, s_u, "疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		thread_goods_bean_slice, err := fetchThreadBeanSlice(thread_goods_slice, r)
		if err != nil {
			util.Debug(" Cannot read thread goods bean slice", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.Approved6Threads.ThreadBeanGoodsSlice = thread_goods_bean_slice

		thread_handicraft_slice, err := pr.ThreadsHandicraft(ctx)
		if err != nil {
			util.Debug(" Cannot read thread handcraft", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		thread_handicraft_bean_slice, err := fetchThreadBeanSlice(thread_handicraft_slice, r)
		if err != nil {
			util.Debug(" Cannot read thread handcraft bean slice", err)
			report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。请稍后再试。")
			return
		}
		pD.Approved6Threads.ThreadBeanHandicraftSlice = thread_handicraft_bean_slice

	}

	// 获取会话session
	s, err := session(r)
	if err != nil {
		// 未登录，游客
		pD.IsGuest = true
		pD.SessUser = dao.User{
			Id:        dao.UserId_None,
			Name:      "游客",
			Footprint: r.URL.Path,
			Query:     r.URL.RawQuery,
		}
		// 返回茶台详情游客页面
		generateHTML(w, &pD, "layout", "navbar.public", "project.detail", "component_thread_bean_approved", "component_thread_bean", "component_avatar_name_gender", "component_sess_capacity")
		return
	}

	// 已登陆用户

	//从会话查获当前浏览用户资料荚
	s_u, s_default_family, s_survival_families, s_default_team, s_survival_teams, s_default_place, s_places, err := fetchSessionUserRelatedData(s)
	if err != nil {
		util.Debug(" Cannot get user-related data from session", s.Email, err)
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
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
	if pD.ProjectBean.Project.Class == dao.PrClassClose {
		is_invited, err := pD.ProjectBean.Project.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug(" Cannot check invited member", err)
			report(w, s_u, "你好，桃李明年能再发，明年闺中知有谁？你真的是受邀请茶团成员吗？")
			return
		}
		pD.IsInvited = is_invited
		// 是封闭式茶台，需要检查当前用户身份是否受邀请茶团的成员，以决定是否允许发言
		if is_invited {
			// 当前用户是本茶话会邀请$团队成员，可以新开茶议
			pD.IsInput = true
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
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
		return
	}
	pD.IsMaster = is_master

	if !is_master {
		is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
		if err != nil {
			util.Debug("Admin permission check failed",
				"userId", s_u.Id,
				"objectiveId", ob.Id,
				"error", err,
			)
			report(w, s_u, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
			return
		}
		pD.IsAdmin = is_admin
	}

	if !pD.IsAdmin && !pD.IsMaster {

		pD.IsVerifier = dao.IsVerifier(s_u.Id)

	}

	// 检查Appointment是否完成
	pD.IsAppointmentCompleted = pr.IsAppointmentCompleted(r.Context())
	// 检查SeeSeek是否完成
	pD.IsSeeSeekCompleted = pr.IsSeeSeekCompleted(r.Context())
	// 检查BrainFire是否完成
	pD.IsBrainFireCompleted = pr.IsBrainFireCompleted(r.Context())
	// 检查Suggestion是否完成
	pD.IsSuggestionCompleted = pr.IsSuggestionCompleted(r.Context())
	// 检查Goods是否完成
	pD.IsGoodsReadinessCompleted = pr.IsGoodsReadinessCompleted(r.Context())
	// 检查Handicrafts是否完成
	all_done, err := dao.IsAllHandicraftsCompleted(pr.Id, r.Context())
	if err != nil && err != sql.ErrNoRows {
		util.Debug("Handicraft check failed", "error:", err)
		report(w, s_u, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
		return
	}
	pD.IsHandicraftsCompleted = all_done

	// 用户足迹
	pD.SessUser.Footprint = r.URL.Path
	pD.SessUser.Query = r.URL.RawQuery

	generateHTML(w, &pD, "layout", "navbar.private", "project.detail", "component_thread_bean_approved", "component_thread_bean", "component_avatar_name_gender", "component_sess_capacity")
}
