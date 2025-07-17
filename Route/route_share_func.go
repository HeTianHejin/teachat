package route

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"
	mrand "math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
	"text/template"
	"time"
	"unicode/utf8"
)

/*
   存放各个路由文件共享的一些方法,常量
*/
// $事业茶团角色
const (
	RoleCEO    = "CEO"
	RoleCTO    = "CTO"
	RoleCMO    = "CMO"
	RoleCFO    = "CFO"
	RoleTaster = "taster"
)

func FetchSeeSeekBean(ss data.SeeSeek) (data.SeeSeekBean, error) {
	ssb := data.SeeSeekBean{SeeSeek: ss}

	ssb.IsOpen = ss.Category == data.SeeSeekCategoryPublic

	verifier, err := data.GetUser(ss.VerifierId)
	if err != nil {
		return ssb, err
	}
	ssb.Verifier = verifier
	verifier_beneficial_family, err := data.GetFamily(ss.VerifierFamilyId)
	if err != nil {
		return ssb, err
	}
	ssb.VerifierBeneficialFamily = verifier_beneficial_family
	verifier_beneficial_team, err := data.GetTeam(ss.VerifierTeamId)
	if err != nil {
		return ssb, err
	}
	ssb.VerifierBeneficialTeam = verifier_beneficial_team

	requester, err := data.GetUser(ss.RequesterId)
	if err != nil {
		return ssb, err
	}
	ssb.Requester = requester
	requester_beneficial_family, err := data.GetFamily(ss.RequesterFamilyId)
	if err != nil {
		return ssb, err
	}
	ssb.RequesterBeneficialFamily = requester_beneficial_family
	requester_beneficial_team, err := data.GetTeam(ss.RequesterTeamId)
	if err != nil {
		return ssb, err
	}
	ssb.RequesterBeneficialTeam = requester_beneficial_team

	provider, err := data.GetUser(ss.ProviderId)
	if err != nil {
		return ssb, err
	}
	ssb.Provider = provider
	provider_beneficial_family, err := data.GetFamily(ss.ProviderFamilyId)
	if err != nil {
		return ssb, err
	}
	ssb.ProviderBeneficialFamily = provider_beneficial_family
	provider_beneficial_team, err := data.GetTeam(ss.ProviderTeamId)
	if err != nil {
		return ssb, err
	}
	ssb.ProviderBeneficialTeam = provider_beneficial_team

	place, err := data.GetPlace(ss.PlaceId)
	if err != nil {
		return ssb, err
	}
	ssb.Place = place

	environment, err := data.GetEnvironment(ss.EnvironmentId)
	if err != nil {
		return ssb, err
	}
	ssb.Environment = environment

	return ssb, nil
}

func moveDefaultTeamToFront(teamSlice []data.TeamBean, defaultTeamID int) ([]data.TeamBean, error) {
	newSlice := make([]data.TeamBean, 0, len(teamSlice))
	var defaultTeam *data.TeamBean

	// 分离默认团队和其他团队
	for _, tb := range teamSlice {
		if tb.Team.Id == defaultTeamID {
			defaultTeam = &tb
			continue
		}
		newSlice = append(newSlice, tb)
	}

	if defaultTeam == nil {
		return nil, fmt.Errorf("默认团队 %d 未找到", defaultTeamID)
	}

	// 合并结果（默认团队在前）
	return append([]data.TeamBean{*defaultTeam}, newSlice...), nil
}

// validateTeamAndFamilyParams 验证团队和家庭ID参数的合法性
// 返回: (是否有效, 错误) ---deepseek协助优化
func validateTeamAndFamilyParams(is_private bool, team_id int, family_id int, currentUserID int, w http.ResponseWriter, r *http.Request) (bool, error) {

	// 基本参数检查（这些检查不涉及数据库操作）
	//非法id组合
	if family_id == data.FamilyIdUnknown && team_id == data.TeamIdNone {
		Report(w, r, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
		return false, nil
	}
	if team_id == data.TeamIdNone || team_id == data.TeamIdSpaceshipCrew {
		Report(w, r, "指定的团队编号是保留编号，不能使用。")
		return false, nil
	}

	if team_id < 0 || family_id < 0 {
		Report(w, r, "团队ID不合法。")
		return false, nil
	}

	// 茶语管理权限归属是.IsPrivate 属性声明的，
	//所以可以同时指定两者,符合任何人必然有某个家庭，但不一定有事业团队背景的实际情况
	if is_private {
		// 管理权属于家庭
		if family_id == data.FamilyIdUnknown {
			Report(w, r, "你好，四海为家者今天不能发布新茶语，请明天再试。")
			return false, fmt.Errorf("unknown family #%d cannot do this", family_id)
		}
		family := data.Family{Id: family_id}
		// if err := family.Get(); err != nil {
		// 	return false, err // 数据库错误，返回error
		// }
		isOnlyOne, err := family.IsOnlyOneMember()
		if err != nil {
			util.Debug("Cannot count family member given id", family.Id, err)
			Report(w, r, "你好，茶博士迷糊了，笔没有墨水未能创建茶话会，请稍后再试。")
			return false, err
		}
		if isOnlyOne {
			Report(w, r, "根据“慎独”约定，单独成员家庭目前暂时不能品茶噢，请向船长抗议。")
			return false, fmt.Errorf("onlyone member family #%d cannot do this", family_id)
		}

		is_member, err := family.IsMember(currentUserID)
		if err != nil {
			return false, err // 数据库错误，返回error
		}
		if !is_member {
			Report(w, r, "你好，家庭成员资格检查失败，请确认后再试。")
			return false, fmt.Errorf(" team %d id_member check failed", team_id)
		}
	} else {
		// 管理权属于团队
		if team_id == data.TeamIdNone || team_id == data.TeamIdSpaceshipCrew {
			Report(w, r, "你好，特殊团队今天还不能创建茶话会，请稍后再试。")
			return false, fmt.Errorf("special team #%d cannot do this", team_id)
		}
		//声明是四海为家【与家庭背景（责任）无关】
		if team_id == data.TeamIdFreelancer {
			//既隐藏家庭背景，也不声明团队的“独狼”
			// 违背了“慎独”原则
			Report(w, r, "你好，茶博士查阅了天书黄页，四海为家的自由人，今天不适宜发表茶话。")
			return false, nil
		}

		team := data.Team{Id: team_id}
		// if err := team.Get(); err != nil {
		// 	return false, err // 数据库错误，返回error
		// }
		is_member, err := team.IsMember(currentUserID)
		if err != nil {
			return false, err // 数据库错误，返回error
		}
		if !is_member {
			Report(w, r, "你好，眼前无路想回头，您是什么团成员？什么茶话会？请稍后再试。")
			return false, nil
		}

	}

	return true, nil
}

// 检查茶围管理权限
func checkObjectiveAdminPermission(ob *data.Objective, userID int) (bool, error) {

	//家庭管理的
	if ob.IsPrivate {
		if ob.FamilyId == data.FamilyIdUnknown {
			return false, fmt.Errorf("checkObjectiveAdminPermission()-> invalid family id %d", ob.FamilyId)
		}
		// family, err := data.GetFamily(ob.FamilyId)
		// if err != nil {
		// 	return false, fmt.Errorf("failed to get family (ID: %d): %v", ob.FamilyId, err)
		// }
		family := data.Family{Id: ob.FamilyId}
		return family.IsParentMember(userID)
	}

	// 团队管理的茶围
	if ob.TeamId == data.TeamIdNone || ob.TeamId == data.TeamIdFreelancer || ob.TeamId == data.TeamIdSpaceshipCrew {
		return false, fmt.Errorf("checkProjectMasterPermission()-> invalid team id %d", ob.TeamId)
	}
	// team, err := data.GetTeam(ob.TeamId)
	// if err != nil {
	// 	return false, fmt.Errorf("failed to get team (ID: %d): %v", ob.TeamId, err)
	// }
	team := data.Team{Id: ob.TeamId}
	return team.IsMember(userID)
}

// 检查茶台管理权限
func checkProjectMasterPermission(pr *data.Project, user_id int) (bool, error) {

	if pr.IsPrivate {
		if pr.FamilyId == data.FamilyIdUnknown {
			return false, fmt.Errorf("checkProjectMasterPermission()-> invalid family id %d", pr.FamilyId)
		}
		// pr_family, err := data.GetFamily(pr.FamilyId)
		// if err != nil {
		// 	return false, fmt.Errorf("failed to get family %d: %v", pr.FamilyId, err)
		// }
		pr_family := data.Family{Id: pr.FamilyId}
		return pr_family.IsParentMember(user_id)
	}

	// 团队管理的
	if pr.TeamId == data.TeamIdNone || pr.TeamId == data.TeamIdFreelancer || pr.TeamId == data.TeamIdSpaceshipCrew {
		return false, fmt.Errorf("checkProjectMasterPermission()-> invalid team id %d", pr.TeamId)
	}
	// pr_team, err := data.GetTeam(pr.TeamId)
	// if err != nil {
	// 	return false, fmt.Errorf("failed to get team %d: %v", pr.TeamId, err)
	// }
	pr_team := data.Team{Id: pr.TeamId}
	return pr_team.IsMember(user_id)
}

// 检查茶台创建权限
func checkCreateProjectPermission(objective data.Objective, userId int, w http.ResponseWriter, r *http.Request) bool {
	switch objective.Class {
	case data.ObClassOpen: // 开放式茶话会
		return true
	case data.ObClassClose: // 封闭式茶话会
		isInvited, err := objective.IsInvitedMember(userId)
		if err != nil {
			util.Debug("检查邀请名单失败", "error", err)
			Report(w, r, "你好，茶博士满头大汗说，邀请品茶名单被狗叼进了花园，请稍候。")
			return false
		}
		if !isInvited {
			Report(w, r, "你好，茶博士无比惊讶说，陛下你的大名竟然不在邀请品茶名单上。")
			return false
		}
		return true
	default:
		Report(w, r, "你好，茶博士失魂鱼，竟然说受邀请品茶名单失踪了，请稍后再试。")
		return false
	}
}

// 检查茶议（thread）创建权限
func checkCreateThreadPermission(project data.Project, userId int, w http.ResponseWriter, r *http.Request) bool {
	switch project.Class {
	case data.PrClassOpen: // 开放式茶台
		return true
	case data.PrClassClose: // 封闭式茶台
		isInvited, err := project.IsInvitedMember(userId)
		if err != nil {
			util.Debug("检查邀请名单失败", "error", err)
			Report(w, r, "你好，茶博士满头大汗说，邀请品茶名单被狗叼进了花园，请稍候。")
			return false
		}
		if !isInvited {
			Report(w, r, "你好，茶博士无比惊讶说，陛下你的大名竟然不在邀请品茶名单上。")
			return false
		}
		return true
	default:
		Report(w, r, "你好，茶博士失魂鱼，竟然说受邀请品茶名单失踪了，请稍后再试。")
		return false
	}
}

// 获取用户最后一次设定的“默认家庭”
// 如果用户没有设定默认家庭，则返回名称为“四海为家”(未知)家庭
// route/family.go
func GetLastDefaultFamilyByUserId(userID int) (data.Family, error) {
	user, err := data.GetUser(userID)
	if err != nil {
		return data.Family{}, fmt.Errorf("failed to get user: %w", err)
	}

	family, err := user.GetLastDefaultFamily()
	switch {
	case err == nil:
		return family, nil
	case errors.Is(err, sql.ErrNoRows):
		return data.UnknownFamily, nil
	default:
		return data.Family{}, fmt.Errorf("failed to get default family: %w", err)
	}
}

// 记录用户最后的查询路径和参数
func RecordLastQueryPath(sess_user_id int, path, raw_query string) (err error) {
	lq := data.LastQuery{
		UserId: sess_user_id,
		Path:   path,
		Query:  raw_query,
	}
	if err = lq.Create(); err != nil {
		return err
	}
	return
}

// Fetch userbean given user 根据user参数，查询用户所得资料荚,包括默认团队，全部已经加入的状态正常团队,成为核心团队，
func FetchUserBean(user data.User) (userbean data.UserBean, err error) {

	userbean.User = user

	default_family, err := GetLastDefaultFamilyByUserId(user.Id)
	if err != nil {
		return userbean, err
	}

	familybean, err := FetchFamilyBean(default_family)
	if err != nil {
		return
	}
	userbean.DefaultFamilyBean = familybean

	family_slice_parent, err := data.ParentMemberFamilies(user.Id)
	if err != nil {
		return
	}
	userbean.ParentMemberFamilyBeanSlice, err = FetchFamilyBeanSlice(family_slice_parent)
	if err != nil {
		return
	}
	family_slice_child, err := data.ChildMemberFamilies(user.Id)
	if err != nil {
		return
	}
	userbean.ChildMemberFamilyBeanSlice, err = FetchFamilyBeanSlice(family_slice_child)
	if err != nil {
		return
	}
	family_slice_other, err := data.OtherMemberFamilies(user.Id)
	if err != nil {
		return
	}
	userbean.OtherMemberFamilyBeanSlice, err = FetchFamilyBeanSlice(family_slice_other)
	if err != nil {
		return
	}
	family_slice_resign, err := data.ResignMemberFamilies(user.Id)
	if err != nil {
		return
	}
	userbean.ResignMemberFamilyBeanSlice, err = FetchFamilyBeanSlice(family_slice_resign)
	if err != nil {
		return
	}

	default_team, err := user.GetLastDefaultTeam()
	if err != nil {
		return
	}
	teambean, err := FetchTeamBean(default_team)
	if err != nil {
		return
	}
	userbean.DefaultTeamBean = teambean

	team_slice_core, err := user.CoreExecTeams()
	if err != nil {
		return
	}
	userbean.ManageTeamBeanSlice, err = FetchTeamBeanSlice(team_slice_core)
	if err != nil {
		return
	}

	team_slice_normal, err := user.NormalExecTeams()
	if err != nil {
		return
	}
	userbean.JoinTeamBeanSlice, err = FetchTeamBeanSlice(team_slice_normal)
	if err != nil {
		return
	}

	default_place, err := user.GetLastDefaultPlace()
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return
	}
	userbean.DefaultPlace = default_place

	return
}

// fetch userbean_slice given []user
func FetchUserBeanSlice(user_slice []data.User) (userbean_slice []data.UserBean, err error) {
	for _, user := range user_slice {
		userbean, err := FetchUserBean(user)
		if err != nil {
			return nil, err
		}
		userbean_slice = append(userbean_slice, userbean)
	}
	return
}

// Fetch and process user-related data,从会话查获当前浏览用户资料荚,包括默认团队，全部已经加入的状态正常团队
func FetchSessionUserRelatedData(sess data.Session) (s_u data.User, family data.Family, families []data.Family, team data.Team, teams []data.Team, place data.Place, places []data.Place, err error) {
	// 读取已登陆用户资料
	s_u, err = sess.User()
	if err != nil {
		return
	}

	member_default_family, err := GetLastDefaultFamilyByUserId(s_u.Id)
	if err != nil {
		return
	}

	member_all_families, err := data.GetAllFamilies(s_u.Id)
	if err != nil {
		return
	}
	//remove member_default_family from member_all_families
	for i, family := range member_all_families {
		if family.Id == member_default_family.Id {
			member_all_families = append(member_all_families[:i], member_all_families[i+1:]...)
			break
		}
	}
	// 把系统默认的“自由人”家庭资料加入families
	member_all_families = append(member_all_families, data.UnknownFamily)
	defaultTeam, err := s_u.GetLastDefaultTeam()
	if err != nil {
		return
	}

	survivalTeams, err := s_u.SurvivalTeams()
	if err != nil {
		return
	}
	for i, team := range survivalTeams {
		if team.Id == defaultTeam.Id {
			survivalTeams = append(survivalTeams[:i], survivalTeams[i+1:]...)
			break
		}
	}
	// 把系统默认团队资料加入teams
	survivalTeams = append(survivalTeams, data.FreelancerTeam)

	default_place, err := s_u.GetLastDefaultPlace()
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return
	}

	places, err = s_u.GetAllBindPlaces()
	if err != nil {
		return
	}
	if len(places) > 0 {
		//移除默认地方
		for i, place := range places {
			if place.Id == default_place.Id {
				places = append(places[:i], places[i+1:]...)
				break
			}
		}
	}

	return s_u, member_default_family, member_all_families, defaultTeam, survivalTeams, default_place, places, nil
}

func fetchUserRelatedData(user data.User) (s_u data.User, family data.Family, families []data.Family, team data.Team, teams []data.Team, place data.Place, places []data.Place, err error) {
	// 读取用户资料
	s_u = user

	member_default_family, err := GetLastDefaultFamilyByUserId(s_u.Id)
	if err != nil {
		return
	}

	member_all_families, err := data.GetAllFamilies(s_u.Id)
	if err != nil {
		return
	}
	//remove member_default_family from member_all_families
	for i, family := range member_all_families {
		if family.Id == member_default_family.Id {
			member_all_families = append(member_all_families[:i], member_all_families[i+1:]...)
			break
		}
	}
	// 把系统默认的“自由人”家庭资料加入families
	member_all_families = append(member_all_families, data.UnknownFamily)
	defaultTeam, err := s_u.GetLastDefaultTeam()
	if err != nil {
		return
	}

	survivalTeams, err := s_u.SurvivalTeams()
	if err != nil {
		return
	}
	for i, team := range survivalTeams {
		if team.Id == defaultTeam.Id {
			survivalTeams = append(survivalTeams[:i], survivalTeams[i+1:]...)
			break
		}
	}
	// 把系统默认团队资料加入teams
	survivalTeams = append(survivalTeams, data.FreelancerTeam)

	default_place, err := s_u.GetLastDefaultPlace()
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return
	}

	places, err = s_u.GetAllBindPlaces()
	if err != nil {
		return
	}
	if len(places) > 0 {
		//移除默认地方
		for i, place := range places {
			if place.Id == default_place.Id {
				places = append(places[:i], places[i+1:]...)
				break
			}
		}
	}

	return s_u, member_default_family, member_all_families, defaultTeam, survivalTeams, default_place, places, nil
}

// 准备用户相关数据
func prepareUserPageData(sess *data.Session) (*data.UserPageData, error) {
	user, defaultFamily, survivalFamilies, defaultTeam, survivalTeams, defaultPlace, places, err := FetchSessionUserRelatedData(*sess)
	if err != nil {
		return nil, err
	}

	return &data.UserPageData{
		User:             user,
		DefaultFamily:    defaultFamily,
		SurvivalFamilies: survivalFamilies,
		DefaultTeam:      defaultTeam,
		SurvivalTeams:    survivalTeams,
		DefaultPlace:     defaultPlace,
		BindPlaces:       places,
	}, nil
}

// 准备茶围页面数据
func prepareObjectivePageData(objective data.Objective, userData *data.UserPageData) (*data.ObjectiveDetail, error) {
	objectiveBean, err := FetchObjectiveBean(objective)
	if err != nil {
		return nil, err
	}

	return &data.ObjectiveDetail{
		SessUser:                 userData.User,
		SessUserDefaultFamily:    userData.DefaultFamily,
		SessUserSurvivalFamilies: userData.SurvivalFamilies,
		SessUserDefaultTeam:      userData.DefaultTeam,
		SessUserSurvivalTeams:    userData.SurvivalTeams,
		SessUserDefaultPlace:     userData.DefaultPlace,
		SessUserBindPlaces:       userData.BindPlaces,
		ObjectiveBean:            objectiveBean,
	}, nil
}

// 根据给出的thread参数，去获取对应的茶议，附属品味计数，作者资料，作者发帖时候选择的茶团，费用和费时。
func FetchThreadBean(thread data.Thread, r *http.Request) (tB data.ThreadBean, err error) {
	tB.Thread = thread

	tB.PostCount = thread.NumReplies()
	//作者资料
	tB.Author, err = thread.User()
	if err != nil {
		util.Debug(fmt.Sprintf("Failed to read thread author for thread ID %d: %v", thread.Id, err))
		return tB, fmt.Errorf("failed to read thread author: %w", err)
	}
	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。
	tB.AuthorFamily, err = data.GetFamily(thread.FamilyId)
	if err != nil {
		util.Debug(" Cannot read thread author family", err)
		return
	}

	tB.AuthorTeam, err = data.GetTeam(thread.TeamId)
	if err != nil {
		util.Debug(" Cannot read thread author team", err)
		return
	}

	tB.StatsSet.PersonCount = 1 //默认为1(作者本人)
	tB.StatsSet.FamilyCount = 0
	tB.StatsSet.TeamCount = 0

	if thread.IsPrivate {
		p_f_count, err := data.CountFamilyParentAndChildMembers(thread.FamilyId, r.Context())
		if err != nil {
			util.Debug(fmt.Sprintf("Failed to count family members for family ID %d: %v", thread.FamilyId, err))
			return tB, fmt.Errorf("failed to count family members: %w", err)
		}
		tB.StatsSet.PersonCount = p_f_count
		tB.StatsSet.FamilyCount = 1
	} else {
		teamMemberCount := tB.AuthorTeam.NumMembers()
		tB.StatsSet.PersonCount = teamMemberCount
	}

	if tB.AuthorTeam.Id > data.TeamIdFreelancer {
		tB.StatsSet.TeamCount = 1
	}

	//idea是否被采纳
	tB.IsApproved = thread.IsApproved()

	return tB, nil
}

// 根据给出的thread_slice参数，去获取对应的茶议（截短正文保留前168字符），附属品味计数，作者资料，作者发帖时候选择的茶团。然后按结构拼装返回
func FetchThreadBeanSlice(thread_slice []data.Thread, r *http.Request) (ThreadBeanSlice []data.ThreadBean, err error) {
	var beanSlice []data.ThreadBean
	// 截短ThreadSlice中thread.Body文字长度为168字符,
	// 展示时长度接近，页面排列比较整齐，
	for i := range thread_slice {
		thread_slice[i].Body = Substr(thread_slice[i].Body, 168)
	}
	for _, thread := range thread_slice {
		ThreadBean, err := FetchThreadBean(thread, r)
		if err != nil {
			return nil, err
		}
		beanSlice = append(beanSlice, ThreadBean)
	}
	ThreadBeanSlice = beanSlice
	return
}

// 根据给出的objectiv_slice参数，去获取对应的茶话会（objective），截短正文保留前168字符，附属茶台计数，发起人资料，发帖时候选择的茶团。然后按结构填写返回资料荚。
func FetchObjectiveBeanSlice(objectiv_slice []data.Objective) (ObjectiveBeanSlice []data.ObjectiveBean, err error) {
	// 截短ObjectiveSlice中objective.Body文字长度为168字符,
	for i := range objectiv_slice {
		objectiv_slice[i].Body = Substr(objectiv_slice[i].Body, 168)
	}
	for _, obj := range objectiv_slice {
		ob, err := FetchObjectiveBean(obj)
		if err != nil {
			return nil, err
		}
		ObjectiveBeanSlice = append(ObjectiveBeanSlice, ob)
	}
	return
}

// 根据给出的objectiv参数，去获取对应的茶话会（objective），附属茶台计数，发起人资料，作者发贴时选择的茶团。然后按结构填写返回资料荚。
func FetchObjectiveBean(ob data.Objective) (ObjectiveBean data.ObjectiveBean, err error) {
	var oB data.ObjectiveBean

	oB.Objective = ob
	if ob.Class == 1 {
		oB.Open = true
	} else {
		oB.Open = false
	}
	oB.Status = ob.GetStatus()
	oB.ProjectCount = ob.NumReplies()
	oB.CreatedAtDate = ob.CreatedAtDate()
	user, err := ob.User()
	if err != nil {
		util.Debug(" Cannot read objective author", err)
		return
	}
	oB.Author = user

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	oB.AuthorFamily, err = data.GetFamily(ob.FamilyId)
	if err != nil {
		util.Debug(" Cannot read objective author family", err)
		return
	}

	oB.AuthorTeam, err = data.GetTeam(ob.TeamId)
	if err != nil {
		util.Debug(" Cannot read objective author team", err)
		return
	}

	return oB, nil
}

// 据给出的project_slice参数，去获取对应的茶台（project），截短正文保留前168字符，附属茶议计数，发起人资料，作者发帖时候选择的茶团。然后按结构填写返回资料。
func FetchProjectBeanSlice(project_slice []data.Project) (ProjectBeanSlice []data.ProjectBean, err error) {
	// 截短ObjectiveSlice中objective.Body文字长度为168字符,
	for i := range project_slice {
		project_slice[i].Body = Substr(project_slice[i].Body, 168)
	}
	for _, pro := range project_slice {
		pb, err := FetchProjectBean(pro)
		if err != nil {
			return nil, err
		}
		ProjectBeanSlice = append(ProjectBeanSlice, pb)
	}
	return
}

// 据给出的project参数，去获取对应的茶台（project），附属茶议计数，发起人资料，作者发帖时候选择的茶团。然后按结构填写返回资料。
func FetchProjectBean(project data.Project) (ProjectBean data.ProjectBean, err error) {
	var pb data.ProjectBean
	pb.Project = project
	if project.Class == 1 {
		pb.Open = true
	} else {
		pb.Open = false
	}
	pb.Status = project.GetStatus()
	pb.ThreadCount = project.NumReplies()
	pb.CreatedAtDate = project.CreatedAtDate()
	author, err := project.User()
	if err != nil {
		util.Debug(" Cannot read project author", err)
		return
	}
	pb.Author = author

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	pb.AuthorFamily, err = data.GetFamily(project.FamilyId)
	if err != nil {
		util.Debug(" Cannot read project author family", err)
		return
	}

	pb.AuthorTeam, err = data.GetTeam(project.TeamId)
	if err != nil {
		util.Debug(" Cannot read project author team", err)
		return
	}

	pb.Place, err = project.Place()
	if err != nil {
		util.Debug("cannot read project place", err)
		return pb, err
	}

	ok, err := project.IsApproved()
	if err != nil {
		util.Debug("cannot read project is approved", project.Id)
		return pb, err
	}
	pb.IsApproved = ok

	return pb, nil
}

// 据给出的post_slice参数，去获取对应的品味（Post），附属茶议计数，作者资料，作者发帖时候选择的茶团。然后按结构拼装返回。
func FetchPostBeanSlice(post_slice []data.Post) (PostBeanSlice []data.PostBean, err error) {
	for _, pos := range post_slice {
		postBean, err := FetchPostBean(pos)
		if err != nil {
			return nil, err
		}
		PostBeanSlice = append(PostBeanSlice, postBean)
	}
	return
}

// 据给出的post参数，去获取对应的品味（Post），附属茶议计数，作者资料，作者发帖时候选择的茶团。然后按结构拼装返回。
func FetchPostBean(post data.Post) (PostBean data.PostBean, err error) {
	PostBean.Post = post
	PostBean.Attitude = post.Atti()
	PostBean.ThreadCount = post.NumReplies()
	PostBean.CreatedAtDate = post.CreatedAtDate()
	author, err := post.User()
	if err != nil {
		util.Debug(" Cannot read post author", err)
		return
	}
	PostBean.Author = author

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	family, err := data.GetFamily(post.FamilyId)
	if err != nil {
		util.Debug(" Cannot read post author family", err)
		return
	}
	PostBean.AuthorFamily = family

	team, err := data.GetTeam(post.TeamId)
	if err != nil {
		util.Debug(" Cannot read post author team", err)
		return
	}
	PostBean.AuthorTeam = team

	return PostBean, nil
}

// 据给出的team参数，去获取对应的茶团资料，是否开放，成员计数，发起日期，发起人（Founder）及其默认团队，然后按结构拼装返回。
func FetchTeamBean(team data.Team) (TeamBean data.TeamBean, err error) {
	if team.Id == data.TeamIdNone {
		return TeamBean, fmt.Errorf("team id is none")
	}

	if team.Class == 1 || team.Class == 0 {
		TeamBean.Open = true
	} else {
		TeamBean.Open = false
	}
	TeamBean.Team = team
	TeamBean.CreatedAtDate = team.CreatedAtDate()

	founder, err := team.Founder()
	if err != nil {
		util.Debug(" Cannot read team founder", err)
		return
	}
	TeamBean.Founder = founder

	TeamBean.FounderDefaultFamily, err = GetLastDefaultFamilyByUserId(founder.Id)
	if err != nil {
		util.Debug(" Cannot read team founder default family", err)
		return
	}

	TeamBean.FounderTeam, err = founder.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read team founder default team", err)
		return
	}

	TeamBean.MemberCount = team.NumMembers()

	if team.Id == data.TeamIdFreelancer {
		//茶友的默认团队还是“自由人”的情况
		TeamBean.CEO = founder
		TeamBean.CEOTeam = TeamBean.FounderTeam
		TeamBean.CEODefaultFamily = TeamBean.FounderDefaultFamily
		return TeamBean, nil
	}

	member_ceo, err := team.MemberCEO()
	if err != nil {
		util.Debug(" Cannot read team member ceo given team_id: ", team.Id, err)
		return
	}
	ceo, err := data.GetUser(member_ceo.UserId)
	if err != nil {
		util.Debug(" Cannot read team ceo given team_id: ", team.Id, err)
		return
	}
	TeamBean.CEO = ceo
	TeamBean.CEOTeam, err = ceo.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read team ceo default team", ceo.Id, err)
		return
	}
	TeamBean.CEODefaultFamily, err = GetLastDefaultFamilyByUserId(ceo.Id)
	if err != nil {
		util.Debug(" Cannot read team ceo default family", ceo.Id, err)
		return
	}

	return TeamBean, nil
}

// 根据给出的茶团队列，查询，获取对应的茶团资料夹
func FetchTeamBeanSlice(team_slice []data.Team) (TeamBeanSlice []data.TeamBean, err error) {
	for _, tea := range team_slice {
		teamBean, err := FetchTeamBean(tea)
		if err != nil {
			return nil, err
		}
		TeamBeanSlice = append(TeamBeanSlice, teamBean)
	}
	return
}

// 根据给出的family参数，从数据库获取对应的家庭资料
func FetchFamilyBean(family data.Family) (FamilyBean data.FamilyBean, err error) {
	FamilyBean.Family = family
	//登记人资料
	FamilyBean.Founder, err = data.GetUser(family.AuthorId)
	if err != nil {
		util.Debug(family.AuthorId, " Cannot read family founder")
		return FamilyBean, err
	}
	FamilyBean.FounderTeam, err = FamilyBean.Founder.GetLastDefaultTeam()
	if err != nil {
		util.Debug(family.AuthorId, " Cannot read family founder default team")
		return FamilyBean, err
	}

	FamilyBean.PersonCount, err = data.CountFamilyMembers(family.Id)
	if err != nil {
		util.Debug(family.AuthorId, " Cannot read family member count")
		return FamilyBean, err
	}
	return
}

// 根据给出的家庭队列，查询，获取对应的家庭茶团资料集合
func FetchFamilyBeanSlice(family_slice []data.Family) (FamilyBeanSlice []data.FamilyBean, err error) {
	for _, fam := range family_slice {
		familyBean, err := FetchFamilyBean(fam)
		if err != nil {
			return nil, err
		}
		FamilyBeanSlice = append(FamilyBeanSlice, familyBean)
	}
	return
}

// FetchFamilyMemberBean() 根据给出的FamilyMember参数，去获取对应的家庭成员资料夹
func FetchFamilyMemberBean(fm data.FamilyMember) (FMB data.FamilyMemberBean, err error) {
	FMB.FamilyMember = fm

	u, err := data.GetUser(fm.UserId)
	if err != nil {
		util.Debug(" Cannot read user given FamilyMember", err)
		return FMB, err
	}
	FMB.Member = u
	default_team, err := u.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read user given FamilyMember", err)
		return FMB, err
	}
	FMB.MemberDefaultTeam = default_team

	f := data.Family{Id: fm.FamilyId}

	//读取茶团的parent_members
	family_parent_members, err := f.ParentMembers()
	if err != nil {
		util.Debug(" Cannot get family core member FetchFamilyMemberBean()", err)
		return
	}
	FMB.IsParent = false
	FMB.IsChild = true
	FMB.IsHusband = false
	FMB.IsWife = false
	for _, f_p_member := range family_parent_members {
		if f_p_member.UserId == u.Id {
			// Set parent flags in one block
			FMB.IsParent = true
			FMB.IsChild = false
			FMB.IsHusband = f_p_member.Role == 1
			FMB.IsWife = f_p_member.Role == 2
			break // Exit loop since we found the match
		}
	}

	member_default_family, err := GetLastDefaultFamilyByUserId(fm.UserId)
	if err != nil {
		util.Debug(" Cannot get GetLastDefaultFamily FetchFamilyMemberBean()", err)
		return
	}
	FMB.MemberDefaultFamily = member_default_family

	if member_default_family.AuthorId == u.Id {
		FMB.IsFounder = true
	} else {
		FMB.IsFounder = false
	}

	return FMB, nil
}

// FetchFamilyMemberBeanSlice() 根据给出的FamilyMember列表参数，去获取对应的家庭成员资料夹列表
func FetchFamilyMemberBeanSlice(fm_slice []data.FamilyMember) (FMB_slice []data.FamilyMemberBean, err error) {
	for _, fm := range fm_slice {
		fmBean, err := FetchFamilyMemberBean(fm)
		if err != nil {
			return nil, err
		}
		FMB_slice = append(FMB_slice, fmBean)
	}
	return
}

// 根据给出的某个&家庭茶团增加成员声明书，获取&家庭茶团增加成员声明书资料夹
func FetchFamilyMemberSignInBean(fmsi data.FamilyMemberSignIn) (FMSIB data.FamilyMemberSignInBean, err error) {
	FMSIB.FamilyMemberSignIn = fmsi

	family := data.Family{Id: fmsi.FamilyId}
	if err = family.Get(); err != nil {
		util.Debug(" Cannot read family given FamilyMemberSignIn", err)
		return FMSIB, err
	}
	FMSIB.Family = family

	FMSIB.NewMember, err = data.GetUser(fmsi.UserId)
	if err != nil {
		util.Debug(" Cannot read new member given FamilyMemberSignIn", err)
		return FMSIB, err
	}

	FMSIB.Author, err = data.GetUser(fmsi.AuthorUserId)
	if err != nil {
		util.Debug(" Cannot read author given FamilyMemberSignIn", err)
		return FMSIB, err
	}

	place := data.Place{Id: fmsi.PlaceId}
	if err = place.Get(); err != nil {
		util.Debug(" Cannot read place given FamilyMemberSignIn", err)
		return FMSIB, err
	}
	FMSIB.Place = place

	return FMSIB, nil
}

// 根据给出的多个&家庭茶团增加成员声明书队列，获取资料夹队列
func FetchFamilyMemberSignInBeanSlice(fmsi_slice []data.FamilyMemberSignIn) (FMSIB_slice []data.FamilyMemberSignInBean, err error) {
	for _, fmsi := range fmsi_slice {
		fmsiBean, err := FetchFamilyMemberSignInBean(fmsi)
		if err != nil {
			return nil, err
		}
		FMSIB_slice = append(FMSIB_slice, fmsiBean)
	}
	return
}

// FetchTeamMemberBean() 根据给出的TeamMember参数，去获取对应的团队成员资料夹
func FetchTeamMemberBean(tm data.TeamMember) (TMB data.TeamMemberBean, err error) {
	u, err := data.GetUser(tm.UserId)
	if err != nil {
		util.Debug(" Cannot read user given TeamMember", err)
		return TMB, err
	}
	TMB.Member = u

	team, err := data.GetTeam(tm.TeamId)
	if err != nil {
		util.Debug(" Cannot read team given team member", err)
		return TMB, err
	}

	if tm.UserId == team.FounderId {
		TMB.IsFounder = true
	} else {
		TMB.IsFounder = false
	}

	//读取茶团的member_ceo
	member_ceo, err := team.MemberCEO()
	if err != nil {
		//茶团已经设定了ceo，但是出现了其他错误
		util.Debug(team.Id, " Cannot get ceo of this team")
		return
	}
	if member_ceo.UserId == u.Id {
		TMB.IsCEO = true
	} else {
		TMB.IsCEO = false
	}

	teamCoreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug(" Cannot get team core member FetchTeamMemberBean()", err)
		return
	}
	for _, coreMember := range teamCoreMembers {
		if coreMember.UserId == u.Id {
			TMB.IsCoreMember = true
			break
		}
	}

	member_default_team, err := u.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot get GetLastDefaultTeam FetchTeamMemberBean()", err)
		return
	}
	TMB.MemberDefaultTeam = member_default_team

	TMB.TeamMember = tm

	TMB.CreatedAtDate = team.CreatedAtDate()

	return TMB, nil
}

// FtchTeamMemberBeanSlice() 根据给出的TeamMember列表参数，去获取对应的团队成员资料夹列表
func FetchTeamMemberBeanSlice(tm_slice []data.TeamMember) (TMB_slice []data.TeamMemberBean, err error) {
	for _, tm := range tm_slice {
		tmBean, err := FetchTeamMemberBean(tm)
		if err != nil {
			return nil, err
		}
		TMB_slice = append(TMB_slice, tmBean)
	}
	return
}

// 根据给出的MemberApplication参数，去获取对应的加盟申请书资料夹
func FetchMemberApplicationBean(ma data.MemberApplication) (MemberApplicationBean data.MemberApplicationBean, err error) {
	MemberApplicationBean.MemberApplication = ma
	MemberApplicationBean.Status = ma.GetStatus()

	team, err := data.GetTeam(ma.TeamId)
	if err != nil {
		util.Debug(" Cannot read team given author", err)
		return MemberApplicationBean, err
	}
	MemberApplicationBean.Team = team

	MemberApplicationBean.Author, err = data.GetUser(ma.UserId)
	if err != nil {
		util.Debug(" Cannot read member application author", err)
		return MemberApplicationBean, err
	}
	MemberApplicationBean.AuthorTeam, err = MemberApplicationBean.Author.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read member application author default team", err)
		return MemberApplicationBean, err
	}

	MemberApplicationBean.CreatedAtDate = ma.CreatedAtDate()
	return MemberApplicationBean, nil
}
func FetchMemberApplicationBeanSlice(ma_slice []data.MemberApplication) (MemberApplicationBeanSlice []data.MemberApplicationBean, err error) {
	for _, ma := range ma_slice {
		maBean, err := FetchMemberApplicationBean(ma)
		if err != nil {
			return nil, err
		}
		MemberApplicationBeanSlice = append(MemberApplicationBeanSlice, maBean)
	}
	return
}

// FetchInvitationBean() 根据给出的Invitation参数，去获取对应的邀请书资料夹
func FetchInvitationBean(i data.Invitation) (I_B data.InvitationBean, err error) {
	I_B.Invitation = i

	I_B.Team, err = i.Team()
	if err != nil {
		util.Debug(" Cannot read invitation default team", err)
		return I_B, err
	}

	I_B.AuthorCEO, err = i.AuthorCEO()
	if err != nil {
		util.Debug(" Cannot fetch team CEO given invitation", err)
		return I_B, err
	}

	I_B.InviteUser, err = i.ToUser()
	if err != nil {
		util.Debug(" Cannot read invitation invite user", err)
		return I_B, err
	}

	I_B.Status = i.GetStatus()
	return I_B, nil
}

// FetchInvitationBeanSlice() 根据给出的Invitation列表参数，去获取对应的邀请书资料夹列表
func FetchInvitationBeanSlice(i_slice []data.Invitation) (I_B_slice []data.InvitationBean, err error) {
	for _, i := range i_slice {
		iBean, err := FetchInvitationBean(i)
		if err != nil {
			return nil, err
		}
		I_B_slice = append(I_B_slice, iBean)
	}
	return
}

// FetchTeamMemberRoleNoticeBean() 根据给出的TeamMemberRoleNotice参数，去获取对应的团队成员角色通知资料夹
func FetchTeamMemberRoleNoticeBean(tmrn data.TeamMemberRoleNotice) (tmrnBean data.TeamMemberRoleNoticeBean, err error) {
	tmrnBean.TeamMemberRoleNotice = tmrn

	tmrnBean.Team, err = data.GetTeam(tmrn.TeamId)
	if err != nil {
		util.Debug(" Cannot read team given team member role notice", err)
		return tmrnBean, err
	}

	tmrnBean.CEO, err = data.GetUser(tmrn.CeoId)
	if err != nil {
		util.Debug(" Cannot read ceo given team member role notice", err)
		return tmrnBean, err
	}

	tm := data.TeamMember{Id: tmrn.MemberId}
	if err = tm.Get(); err != nil {
		util.Debug(" Cannot read team member given team member role notice", err)
		return tmrnBean, err
	}
	tmrnBean.Member, err = data.GetUser(tm.UserId)
	if err != nil {
		util.Debug(" Cannot read member given team member role notice", err)
		return tmrnBean, err
	}
	tmrnBean.MemberDefaultTeam, err = tmrnBean.Member.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read member default team given team member role notice", err)
		return tmrnBean, err
	}

	return tmrnBean, nil
}

// FetchTeamMemberRoleNoticeBeanSlice() 根据给出的TeamMemberRoleNotice列表参数，去获取对应的团队成员角色通知资料夹列表
func FetchTeamMemberRoleNoticeBeanSlice(tmrn_slice []data.TeamMemberRoleNotice) (tmrnBeanSlice []data.TeamMemberRoleNoticeBean, err error) {
	for _, tmrn := range tmrn_slice {
		tmrnBean, err := FetchTeamMemberRoleNoticeBean(tmrn)
		if err != nil {
			return nil, err
		}
		tmrnBeanSlice = append(tmrnBeanSlice, tmrnBean)
	}
	return
}

// 处理头像图片上传方法，图片要求为jpeg格式，size<30kb,宽高尺寸是64，32像素之间
func ProcessUploadAvatar(w http.ResponseWriter, r *http.Request, uuid string) error {
	// 从请求中解包出单个上传文件
	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		Report(w, r, "获取头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer file.Close()

	// 获取文件大小，注意：客户端提供的文件大小可能不准确
	size := fileHeader.Size
	if size > 30*1024 {
		Report(w, r, "文件大小超过30kb,茶博士接不住。")
		return errors.New("the file size over 30kb")
	}
	// 实际读取文件大小进行校验，以防止客户端伪造
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		Report(w, r, "读取头像文件失败，请稍后再试。")
		return err
	}
	if len(fileBytes) > 30*1024 {
		Report(w, r, "文件大小超过30kb,茶博士接不住。")
		return errors.New("the file size over 30kb")
	}

	// 获取文件名和检查文件后缀
	filename := fileHeader.Filename
	ext := strings.ToLower(path.Ext(filename))
	if ext != ".jpeg" && ext != ".jpg" {
		Report(w, r, "注意头像图片文件类型, 目前仅限jpeg格式图片上传。")
		return errors.New("the file type is not jpeg")
	}

	// 获取文件类型，注意：客户端提供的文件类型可能不准确
	fileType := http.DetectContentType(fileBytes)
	if fileType != "image/jpeg" {
		Report(w, r, "注意图片文件类型,目前仅限jpeg格式。")
		return errors.New("the file type is not jpeg")
	}

	// 检测图片尺寸宽高和图像格式,判断是否合适
	width, height, err := GetWidthHeightForJpeg(fileBytes)
	if err != nil {
		Report(w, r, "注意图片文件格式, 目前仅限jpeg格式。")
		return err
	}
	if width < 32 || width > 64 || height < 32 || height > 64 {
		Report(w, r, "注意图片尺寸, 宽高需要在32-64像素之间。")
		return errors.New("the image size is not between 32 and 64")
	}

	// 创建新文件，无需切换目录，直接使用完整路径，减少安全风险
	newFilePath := util.Config.ImageDir + uuid + util.Config.ImageExt
	newFile, err := os.Create(newFilePath)
	if err != nil {
		util.Debug("创建头像文件名失败", err)
		Report(w, r, "创建头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer newFile.Close()

	// 通过缓存方法写入硬盘
	buff := bufio.NewWriter(newFile)
	if _, err = buff.Write(fileBytes); err != nil {
		util.Debug("fail to write avatar image", err)
		Report(w, r, "你好，茶博士居然说没有墨水了， 未能写完头像文件，请稍后再试。")
		return err
	}
	if err = buff.Flush(); err != nil {
		util.Debug("fail to write avatar image", err)
		Report(w, r, "你好，茶博士居然说没有墨水了，写入头像文件不成功，请稍后再试。")
		return err
	}

	return nil
}

// 茶博士——古时专指陆羽。陆羽著《茶经》，唐德宗李适曾当面称陆羽为“茶博士”。
// 茶博士-teaOffice，是古代中华传统文化对茶馆工作人员的昵称，如：富家宴会，犹有专供茶事之人，谓之茶博士。——唐代《西湖志馀》
// 现在多指精通茶艺的师傅，尤其是四川的长嘴壶茶艺，茶博士个个都是身怀绝技的“高手”。
// 茶博士向茶客报告信息的方法，包括但不限于意外事件和通知、感谢等等提示。
func Report(w http.ResponseWriter, r *http.Request, msg ...any) {
	var userBPD data.UserBean
	var b strings.Builder
	for i, arg := range msg {
		if i > 0 {
			b.WriteByte(' ') // 参数间添加空格
		}
		fmt.Fprint(&b, arg)
	}
	userBPD.Message = b.String()

	s, err := Session(r)
	if err != nil {
		userBPD.SessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
		}
		RenderHTML(w, &userBPD, "layout", "navbar.public", "feedback")
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", s.Email, err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	userBPD.SessUser = s_u

	RenderHTML(w, &userBPD, "layout", "navbar.private", "feedback")
}

// Checks if the user is logged in and has a Session, if not err is not nil
func Session(r *http.Request) (data.Session, error) {
	cookie, err := r.Cookie("_cookie")
	if err != nil {
		return data.Session{}, fmt.Errorf("cookie not found: %w", err)
	}

	sess := data.Session{Uuid: cookie.Value}
	ok, checkErr := sess.Check()
	if checkErr != nil {
		return data.Session{}, fmt.Errorf("session check failed: %w", checkErr)
	}
	if !ok {
		return data.Session{}, errors.New("invalid or expired session")
	}

	return sess, nil
}

// parse HTML templates
// pass in a slice of file names, and get a template
func ParseTemplateFiles(filenames ...string) *template.Template {
	var files []string
	t := template.New("layout")
	for _, file := range filenames {
		// 使用 filepath.Join 安全拼接路径,unix+windows
		filePath := filepath.Join(util.Config.TemplateExt, file+util.Config.TemplateExt)
		files = append(files, filePath)
	}
	t = template.Must(t.ParseFiles(files...))
	return t
}

// 处理器把页面模版和需求数据揉合后，由这个方法，将填写好的页面“制作“成HTML格式，调用http响应方法，发送给浏览器端客户
func RenderHTML(w http.ResponseWriter, data any, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.go.html", file))
	}

	// 手动解析模板并处理错误
	templates, err := template.ParseFiles(files...)
	if err != nil {
		// 添加详细的错误日志和HTTP错误响应
		util.PrintStdout("模板解析错误: ", err)
		http.Error(w, "*** 茶博士: 茶壶不见了，无法烧水冲茶，陛下稍安勿躁 ***", http.StatusInternalServerError)
		return
	}

	// 安全增强：设置内容类型为HTML并添加XSS防护头
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// 执行模板渲染
	if err = templates.ExecuteTemplate(w, "layout", data); err != nil {
		// 添加详细的错误日志
		util.PrintStdout("模板渲染错误: ", err)
		// 避免在错误响应中泄露敏感信息
		http.Error(w, "*** 茶博士: 茶壶不见了，无法烧水冲茶，陛下稍安勿躁 ***", http.StatusInternalServerError)
	}
}

// 验证邮箱地址，格式是否正确，正确返回true，错误返回false。
func IsEmail(email string) bool {
	pattern := `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// 验证提交的string是否 1 正整数？
func VerifyPositiveIntegerFormat(str string) bool {
	if str == "" {
		return false
	}
	pattern := `^[1-9]\d*$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(str)
}

// 验证team_id_slice，必需是正整数的逗号分隔的"2,19,87..."字符串格式是否正确，正确返回true，错误返回false。
func VerifyIdSliceFormat(team_id_slice string) bool {
	if team_id_slice == "" {
		return false
	}
	// 使用双引号显式声明正则表达式，避免隐藏字符
	pattern := "^[0-9]+(,[0-9]+)*$"
	reg, err := regexp.Compile(pattern)
	if err != nil {
		// 实际生产环境应记录该错误
		return false
	}
	return reg.MatchString(team_id_slice)
}

// 输入两个统计数（辩论的正方累积得分数，辩论总得分数）（整数），计算前者与后者比值，结果浮点数向上四舍五入取整,
// 返回百分数的分子整数
func ProgressRound(numerator, denominator int) int {
	if denominator == 0 {
		// 分母为0时，视作未有记录，即未进行表决状态，返回默认值100
		return 100
	}
	if numerator == denominator {
		// 分子等于分母时，表示100%正方
		return 100
	}
	ratio := float64(numerator) / float64(denominator) * 100

	// if numerator > denominator {
	// 	// 分子大于分母时，表示统计数据输入错误，返回一个中间值
	// 	return 50
	// } else if ratio < 0 {
	// 	// 分子小于分母且比例为负数，表示统计数据输入错误，返回一个中间值
	// 	return 50
	// } else if ratio < 1 {
	// 	// 比例小于1时，返回最低限度值1
	// 	return 1
	// }

	// 其他情况，使用math.Floor确保向下取整，然后四舍五入
	//return int(math.Floor(ratio + 0.5))
	return int(math.Round(ratio))
}

/*
* 入参： JPG 图片文件的二进制数据
* 出参：JPG 图片的宽和高
* Author Mr.YF https://www.cnblogs.com/voipman
 */
func GetWidthHeightForJpeg(imgBytes []byte) (int, int, error) {
	var offset int
	imgByteLen := len(imgBytes)
	for i := 0; i < imgByteLen-1; i++ {
		if imgBytes[i] != 0xff {
			continue
		}
		if imgBytes[i+1] == 0xC0 || imgBytes[i+1] == 0xC1 || imgBytes[i+1] == 0xC2 {
			offset = i
			break
		}
	}
	offset += 5
	if offset >= imgByteLen {
		return 0, 0, errors.New("unknown format")
	}
	height := int(imgBytes[offset])<<8 + int(imgBytes[offset+1])
	width := int(imgBytes[offset+2])<<8 + int(imgBytes[offset+3])
	return width, height, nil
}

// RandomInt() 生成count个随机且不重复的整数，范围在[start, end)之间，按升序排列
func RandomInt(start, end, count int) []int {
	// 检查参数有效性
	if count <= 0 || start >= end {
		return nil
	}

	// 初始化包含所有可能随机数的切片
	nums := make([]int, end-start)
	for i := range nums {
		nums[i] = start + i
	}

	// 使用Fisher-Yates洗牌算法打乱切片顺序
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	for i := len(nums) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		nums[i], nums[j] = nums[j], nums[i]
	}

	// 切片只需要前count个元素
	nums = nums[:count]

	// 对切片进行排序
	sort.Ints(nums)

	return nums
}

// 生成“火星文”替换下标队列
func StaRepIntSlice(str_len, ratio int) (numSlice []int, err error) {

	half := str_len / 2
	substandard := str_len * ratio / 100
	// 存放结果的slice
	numSlice = make([]int, str_len)

	// 随机生成替换下标
	switch {
	case ratio < 50:
		numSlice = []int{}
		return numSlice, errors.New("ratio must be not less than 50")
	case ratio == 50:
		numSlice = RandomInt(0, str_len, half)
	case ratio > 50:
		numSlice = RandomInt(0, str_len, substandard)
	}

	return
}

// 1. 校验茶议已有内容是否不超限,false == 超限
func SubmitAdditionalContent(w http.ResponseWriter, r *http.Request, body, additional string) bool {
	if CnStrLen(body) >= int(util.Config.ThreadMaxWord) {
		Report(w, r, "已有内容已超过最大字数限制，无法补充。")
		return false
	}

	// 2. 校验补充内容字数
	min := int(util.Config.ThreadMinWord)
	max := int(util.Config.ThreadMaxWord) - CnStrLen(body)
	current := CnStrLen(additional)

	if current < min || current > max {
		errMsg := fmt.Sprintf(
			"茶博士提示：补充内容需满足：%d ≤ 字数 ≤ %d（当前：%d）。",
			min, max, current,
		)
		Report(w, r, errMsg)
		return false
	}
	// 3. 校验补充内容是否包含敏感词

	return true
}

// 计算中文字符串长度
func CnStrLen(str string) int {
	return utf8.RuneCountInString(str)
}

// 对未经蒙评的草稿进行“火星文”遮盖隐秘处理，即用星号替换50%或者指定更高比例文字
func MarsString(str string, ratio int) string {
	len := CnStrLen(str)
	// 获取替换字符的下标队列
	nslice, err := StaRepIntSlice(len, ratio)
	if err != nil {
		return str
	}
	// 把字符串转换为[]rune
	rstr := []rune(str)
	// 遍历替换字符的下标队列

	for _, n := range nslice {
		// 替换下标指定的字符为星号
		rstr[n] = '*'
	}

	// 将[]rune转换为字符串

	return string(rstr)
}

// 入参string，截取前面一段指定长度文字，返回string，作为预览文字
// CodeBuddy修改
func Substr(s string, length int) string {
	if length <= 0 {
		return ""
	}
	var count int //统计字符数（而非字节数）
	end := 0      //记录最后一个字符的起始字节位置
	for i := range s {
		if count == length {
			break
		}
		count++
		end = i
	}
	if count < length {
		return s
	}
	_, size := utf8.DecodeRuneInString(s[end:])
	return s[:end+size]
}

// 截取一段指定开始和结束位置的文字，用range迭代方法。入参string，返回string“...”
// 注意，输入负数=最大值
func Substr2(str string, start, end int) string {

	//str += "." //这是根据range的特性加的，如果不加，截取不到最后一个字（end+1=意外，因为1中文=3字节！）

	var cnt, s, e int
	for s = range str {
		if cnt == start {
			break
		}
		cnt++
	}
	cnt = 0
	for e = range str {
		if cnt == end {
			break
		}
		cnt++
	}
	return str[s:e]
}

// sanitizeRedirectPath 只允许站内路径（如 /v1/home），禁止外部域名
// --- DeeSeek
func sanitizeRedirectPath(inputPath string) string {
	if inputPath == "" {
		return "/v1/" // 默认路径
	}

	// 检查是否以 "/" 开头（相对路径）
	if len(inputPath) > 0 && inputPath[0] == '/' {
		// 可选：进一步校验路径格式（避免路径遍历攻击，如 /../）
		cleanedPath := path.Clean(inputPath)
		if !strings.HasPrefix(cleanedPath, "/v1/") {
			return "/v1/" // 强制限制到特定前缀
		}
		return cleanedPath
	}

	// 非相对路径（如http://）则返回默认路径
	return "/v1/"
}

// Helper function for validating string length
func validateCnStrLen(value string, min int, max int, fieldName string, w http.ResponseWriter, r *http.Request) bool {
	if CnStrLen(value) < min {
		Report(w, r, fmt.Sprintf("你好，茶博士竟然说该茶议%s为空或太短，请确认后再试一次。", fieldName))
		return false
	}
	if CnStrLen(value) > max {
		Report(w, r, fmt.Sprintf("你好，茶博士竟然说该茶议%s过长，请确认后再试一次。", fieldName))
		return false
	}
	return true
}

// 创建AcceptObject并发送邻座蒙评消息
func CreateAndSendAcceptMessage(objectId int, objectType int, excludeUserId int) error {
	// 创建AcceptObject
	aO := data.AcceptObject{
		ObjectId:   objectId,
		ObjectType: objectType,
	}
	if err := aO.Create(); err != nil {
		util.Debug("Cannot create accept_object given objectId", objectId)
		return fmt.Errorf("创建AcceptObject失败: %w", err)
	}

	// 创建消息
	mess := data.AcceptMessage{
		FromUserId:     data.UserId_SpaceshipCaptain,
		Title:          "新茶语邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时去审理。",
		AcceptObjectId: aO.Id,
	}

	// 发送消息
	if err := TwoAcceptMessagesSendExceptUserId(excludeUserId, mess); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	// 返回提示信息
	return nil
}

// 接纳文明新茶台
func AcceptNewProject(objectId int) error {
	pr := data.Project{
		Id: objectId,
	}
	if err := pr.Get(); err != nil {
		util.Debug("Cannot get project", objectId, err)
		return errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}

	switch pr.Class {
	case data.PrClassOpenStraw:
		pr.Class = data.PrClassOpen
	case data.PrClassCloseStraw:
		pr.Class = data.PrClassClose
	default:
		return errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}

	if err := pr.UpdateClass(); err != nil {
		util.Debug("Cannot update pr class", objectId, err)
		return errors.New("你好，一畦春韭绿，十里稻花香。")
	}
	return nil
}

// 接纳文明新茶围
func AcceptNewObjective(objectId int) (*data.Objective, error) {
	ob := data.Objective{
		Id: objectId,
	}
	if err := ob.Get(); err != nil {
		util.Debug("Cannot get objective", objectId, err)
		return nil, errors.New("你好，茶博士失魂鱼，竟然说没有找到新茶茶叶的资料未必是怪事。")
	}
	// 检查当前茶围的状态
	switch ob.Class {
	case data.ObClassOpenStraw:
		ob.Class = data.ObClassOpen
	case data.ObClassCloseStraw:
		ob.Class = data.ObClassClose
	default:
		return nil, errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}

	if err := ob.UpdateClass(); err != nil {
		util.Debug("Cannot update ob class", objectId, err)
		return nil, errors.New("你好，一畦春韭绿，十里稻花香。")
	}
	return &ob, nil
}

// 接纳文明新茶团
func AcceptNewTeam(objectId int) (*data.Team, error) {
	t := data.Team{Id: objectId}
	if err := t.Get(); err != nil {
		util.Debug("Cannot get team", objectId, err)
		return nil, errors.New("你好，茶博士失魂鱼，竟然说没有找到新茶茶叶的资料未必是怪事。")
	}
	switch t.Class {
	case data.TeamClassOpenStraw:
		t.Class = data.TeamClassOpen
	case data.TeamClassCloseStraw:
		t.Class = data.TeamClassClose
	default:
		return nil, errors.New("你好，茶博士失魂鱼，竟然说有时找茶叶也是一种修养的过程。")
	}
	if err := t.UpdateClass(); err != nil {
		util.Debug("Cannot update t class", objectId, err)
		return nil, errors.New("你好，一畦春韭绿，十里稻花香。")
	}
	return &t, nil
}

// 接纳文明新茶议
func AcceptNewDraftThread(objectId int) (*data.Thread, error) {
	dThread := data.DraftThread{Id: objectId}
	if err := dThread.Get(); err != nil {
		return nil, fmt.Errorf("获取茶议草稿失败: %w", err)
	}

	if err := dThread.UpdateStatus(data.DraftThreadStatusAccepted); err != nil {
		return nil, fmt.Errorf("更新茶议草稿状态失败: %w", err)
	}

	thread := data.Thread{
		Body:      dThread.Body,
		UserId:    dThread.UserId,
		Class:     dThread.Class,
		Title:     dThread.Title,
		ProjectId: dThread.ProjectId,
		FamilyId:  dThread.FamilyId,
		Type:      dThread.Type,
		PostId:    dThread.PostId,
		TeamId:    dThread.TeamId,
		IsPrivate: dThread.IsPrivate,
		Category:  dThread.Category,
	}

	if err := thread.Create(); err != nil {
		return nil, fmt.Errorf("创建新茶议失败: %w", err)
	}

	return &thread, nil
}

// 接纳文明新茶语之品味
func AcceptNewDraftPost(objectId int) (*data.Post, error) {
	dPost := data.DraftPost{Id: objectId}
	if err := dPost.Get(); err != nil {
		return nil, fmt.Errorf("获取品味草稿失败: %w", err)
	}
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
	if err := new_post.Create(); err != nil {
		return nil, fmt.Errorf("创建新品味失败: %w", err)
	}
	return &new_post, nil
}

// 检查并设置用户默认团队（非自由人占位团队）
func SetUserDefaultTeam(founder *data.User, newTeamID int, w http.ResponseWriter, r *http.Request) bool {
	// 获取用户当前默认团队
	oldDefaultTeam, err := founder.GetLastDefaultTeam()
	if err != nil {
		util.Debug(founder.Email, "Cannot get last default team")
		Report(w, r, "你好，茶博士失魂鱼，手滑未能创建你的天命使团，请稍后再试。")
		return false
	}

	// 检查是否为占位团队（自由人）
	if oldDefaultTeam.Id == data.TeamIdFreelancer {
		uDT := data.UserDefaultTeam{
			UserId: founder.Id,
			TeamId: newTeamID,
		}
		if err := uDT.Create(); err != nil {
			util.Debug(founder.Email, newTeamID, "Cannot create default team")
			Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
			return false
		}
	}
	return true
}
