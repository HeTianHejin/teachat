package route

import (
	"fmt"
	"net/http"
	dao "teachat/DAO"
	util "teachat/Util"
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

// 检查茶围目标管理权限，即用户是否属于茶围归属团队有效成员
func checkObjectiveAdminPermission(ob *dao.Objective, userID int) (bool, error) {

	//家庭管理的
	if ob.IsPrivate {
		if ob.FamilyId == dao.FamilyIdUnknown {
			return false, fmt.Errorf("checkObjectiveAdminPermission()-> invalid family id %d", ob.FamilyId)
		}

		family := dao.Family{Id: ob.FamilyId}
		return family.IsParentMember(userID)
	}

	// 团队管理的茶围
	if ob.TeamId == dao.TeamIdNone || ob.TeamId == dao.TeamIdFreelancer || ob.TeamId == dao.TeamIdSpaceshipCrew {
		return false, fmt.Errorf("checkProjectMasterPermission()-> invalid team id %d", ob.TeamId)
	}

	team := dao.Team{Id: ob.TeamId}
	return team.IsMember(userID)
}

// 检查茶台项目管理权限，即用户是否茶台项目归属团队有权成员
func checkProjectMasterPermission(pr *dao.Project, user_id int) (bool, error) {

	if pr.IsPrivate {
		if pr.FamilyId == dao.FamilyIdUnknown {
			return false, fmt.Errorf("checkProjectMasterPermission()-> invalid family id %d", pr.FamilyId)
		}

		pr_family := dao.Family{Id: pr.FamilyId}
		return pr_family.IsParentMember(user_id)
	}

	// 团队管理的
	if pr.TeamId == dao.TeamIdNone || pr.TeamId == dao.TeamIdFreelancer || pr.TeamId == dao.TeamIdSpaceshipCrew {
		return false, fmt.Errorf("checkProjectMasterPermission()-> invalid team id %d", pr.TeamId)
	}

	pr_team := dao.Team{Id: pr.TeamId}
	return pr_team.IsMember(user_id)
}

// 检查茶台创建权限
func checkCreateProjectPermission(objective dao.Objective, s_u dao.User, w http.ResponseWriter) bool {
	switch objective.Class {
	case dao.ObClassOpen: // 开放式茶话会
		return true
	case dao.ObClassClose: // 封闭式茶话会
		isInvited, err := objective.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug("检查邀请名单失败", "error", err)
			report(w, s_u, "你好，茶博士满头大汗说，邀请品茶名单被狗叼进了花园，请稍候。")
			return false
		}
		if !isInvited {
			report(w, s_u, "你好，茶博士无比惊讶说，陛下你的大名竟然不在邀请品茶名单上。")
			return false
		}
		return true
	default:
		report(w, s_u, "你好，茶博士失魂鱼，竟然说受邀请品茶名单失踪了，请稍后再试。")
		return false
	}
}

// 检查茶议（thread）创建权限
func checkCreateThreadPermission(project dao.Project, s_u dao.User, w http.ResponseWriter) bool {
	switch project.Class {
	case dao.PrClassOpen: // 开放式茶台
		return true
	case dao.PrClassClose: // 封闭式茶台
		isInvited, err := project.IsInvitedMember(s_u.Id)
		if err != nil {
			util.Debug("检查邀请名单失败", "error", err)
			report(w, s_u, "你好，茶博士满头大汗说，邀请品茶名单被狗叼进了花园，请稍候。")
			return false
		}
		if !isInvited {
			report(w, s_u, "你好，茶博士无比惊讶说，陛下你的大名竟然不在邀请品茶名单上。")
			return false
		}
		return true
	default:
		report(w, s_u, "你好，茶博士失魂鱼，竟然说受邀请品茶名单失踪了，请稍后再试。")
		return false
	}
}

// 准备茶围页面数据
func prepareObjectivePageData(objective dao.Objective, userData *dao.UserPageData) (*dao.ObjectiveDetail, error) {
	objectiveBean, err := fetchObjectiveBean(objective)
	if err != nil {
		return nil, err
	}

	return &dao.ObjectiveDetail{
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
func fetchThreadBean(thread dao.Thread, r *http.Request) (tB dao.ThreadBean, err error) {
	tB.Thread = thread

	tB.PostCount = thread.NumReplies()
	//作者资料
	tB.Author, err = thread.Author()
	if err != nil {
		util.Debug(fmt.Sprintf("Failed to read thread author for thread ID %d: %v", thread.Id, err))
		return tB, fmt.Errorf("failed to read thread author: %w", err)
	}
	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。
	tB.AuthorFamily, err = dao.GetFamily(thread.FamilyId)
	if err != nil {
		util.Debug(" Cannot read thread author family", err)
		return
	}

	tB.AuthorTeam, err = dao.GetTeam(thread.TeamId)
	if err != nil {
		util.Debug(" Cannot read thread author team", err)
		return
	}

	tB.StatsSet.MemberCount = 1 //默认为1(作者本人)
	tB.StatsSet.FamilyCount = 0
	tB.StatsSet.TeamCount = 0

	if thread.IsPrivate {
		p_f_count, err := dao.CountFamilyParentAndChildMembers(thread.FamilyId, r.Context())
		if err != nil {
			util.Debug(fmt.Sprintf("Failed to count family members for family ID %d: %v", thread.FamilyId, err))
			return tB, fmt.Errorf("failed to count family members: %w", err)
		}
		tB.StatsSet.MemberCount = p_f_count
		tB.StatsSet.FamilyCount = 1
	} else {
		teamMembersCount := tB.AuthorTeam.NumMembers()
		tB.StatsSet.MemberCount = teamMembersCount
	}

	if tB.AuthorTeam.Id > dao.TeamIdFreelancer {
		tB.StatsSet.TeamCount = 1
	}

	//idea是否被采纳
	tB.IsApproved = thread.IsApproved()

	return tB, nil
}

// 根据给出的thread_slice参数，去获取对应的茶议（截短正文保留前168字符），附属品味计数，作者资料，作者发帖时候选择的茶团。然后按结构拼装返回
func fetchThreadBeanSlice(thread_slice []dao.Thread, r *http.Request) (ThreadBeanSlice []dao.ThreadBean, err error) {
	var beanSlice []dao.ThreadBean
	// 截短ThreadSlice中thread.Body文字长度为168字符,
	// 展示时长度接近，页面排列比较整齐，
	for i := range thread_slice {
		thread_slice[i].Body = subStr(thread_slice[i].Body, 168)
	}
	for _, thread := range thread_slice {
		ThreadBean, err := fetchThreadBean(thread, r)
		if err != nil {
			return nil, err
		}
		beanSlice = append(beanSlice, ThreadBean)
	}
	ThreadBeanSlice = beanSlice
	return
}

// 根据给出的objectiv_slice参数，去获取对应的茶话会（objective），截短正文保留前168字符，附属茶台计数，发起人资料，发帖时候选择的茶团。然后按结构填写返回资料荚。
func FetchObjectiveBeanSlice(objectiv_slice []dao.Objective) (ObjectiveBeanSlice []dao.ObjectiveBean, err error) {
	// 截短ObjectiveSlice中objective.Body文字长度为168字符,
	for i := range objectiv_slice {
		objectiv_slice[i].Body = subStr(objectiv_slice[i].Body, 168)
	}
	for _, obj := range objectiv_slice {
		ob, err := fetchObjectiveBean(obj)
		if err != nil {
			return nil, err
		}
		ObjectiveBeanSlice = append(ObjectiveBeanSlice, ob)
	}
	return
}

// 根据给出的objectiv参数，去获取对应的茶话会（objective），附属茶台计数，发起人资料，作者发贴时选择的茶团。然后按结构填写返回资料荚。
func fetchObjectiveBean(ob dao.Objective) (ObjectiveBean dao.ObjectiveBean, err error) {
	var oB dao.ObjectiveBean

	oB.Objective = ob
	if ob.Class == 1 {
		oB.Open = true
	} else {
		oB.Open = false
	}
	oB.Status = ob.GetStatus()
	oB.ProjectCount = ob.NumReplies()
	oB.CreatedAtDate = ob.CreatedAtDate()
	user, err := ob.Admin()
	if err != nil {
		util.Debug(" Cannot read objective author", err)
		return
	}
	oB.Author = user

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	oB.AuthorFamily, err = dao.GetFamily(ob.FamilyId)
	if err != nil {
		util.Debug(" Cannot read objective author family", err)
		return
	}

	oB.AuthorTeam, err = dao.GetTeam(ob.TeamId)
	if err != nil {
		util.Debug(" Cannot read objective author team", err)
		return
	}

	return oB, nil
}

// 据给出的project_slice参数，去获取对应的茶台（project），截短正文保留前168字符，附属茶议计数，发起人资料，作者发帖时候选择的茶团。然后按结构填写返回资料。
func fetchProjectBeanSlice(project_slice []dao.Project) (ProjectBeanSlice []dao.ProjectBean, err error) {
	// 截短ObjectiveSlice中objective.Body文字长度为168字符,
	for i := range project_slice {
		project_slice[i].Body = subStr(project_slice[i].Body, 168)
	}
	for _, pro := range project_slice {
		pb, err := fetchProjectBean(pro)
		if err != nil {
			return nil, err
		}
		ProjectBeanSlice = append(ProjectBeanSlice, pb)
	}
	return
}

// 据给出的project参数，去获取对应的茶台（project），附属茶议计数，发起人资料，作者发帖时候选择的茶团。然后按结构填写返回资料。
func fetchProjectBean(project dao.Project) (ProjectBean dao.ProjectBean, err error) {
	var pb dao.ProjectBean
	pb.Project = project
	if project.Class == 1 {
		pb.Open = true
	} else {
		pb.Open = false
	}
	pb.Status = project.ClassString()
	pb.ThreadCount = project.NumReplies()
	pb.CreatedAtDate = project.CreatedAtDate()
	author, err := project.Master()
	if err != nil {
		util.Debug(" Cannot read project author", err)
		return
	}
	pb.Author = author

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	pb.AuthorFamily, err = dao.GetFamily(project.FamilyId)
	if err != nil {
		util.Debug(" Cannot read project author family", err)
		return
	}

	pb.AuthorTeam, err = dao.GetTeam(project.TeamId)
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
func fetchPostBeanSlice(post_slice []dao.Post) (PostBeanSlice []dao.PostBean, err error) {
	for _, pos := range post_slice {
		postBean, err := fetchPostBean(pos)
		if err != nil {
			return nil, err
		}
		PostBeanSlice = append(PostBeanSlice, postBean)
	}
	return
}

// 据给出的post参数，去获取对应的品味（Post），附属茶议计数，作者资料，作者发帖时候选择的茶团。然后按结构拼装返回。
func fetchPostBean(post dao.Post) (PostBean dao.PostBean, err error) {
	PostBean.Post = post
	PostBean.Attitude = post.Atti()
	PostBean.ThreadCount = post.NumReplies()
	PostBean.CreatedAtDate = post.CreatedAtDate()
	author, err := post.Author()
	if err != nil {
		util.Debug(" Cannot read post author", err)
		return
	}
	PostBean.Author = author

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	family, err := dao.GetFamily(post.FamilyId)
	if err != nil {
		util.Debug(" Cannot read post author family", err)
		return
	}
	PostBean.AuthorFamily = family

	team, err := dao.GetTeam(post.TeamId)
	if err != nil {
		util.Debug(" Cannot read post author team", err)
		return
	}
	PostBean.AuthorTeam = team

	return PostBean, nil
}
