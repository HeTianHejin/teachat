package route

import (
	"errors"
	"fmt"
	"net/http"
	data "teachat/DAO"
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
			report(w, r, "你好，茶博士满头大汗说，邀请品茶名单被狗叼进了花园，请稍候。")
			return false
		}
		if !isInvited {
			report(w, r, "你好，茶博士无比惊讶说，陛下你的大名竟然不在邀请品茶名单上。")
			return false
		}
		return true
	default:
		report(w, r, "你好，茶博士失魂鱼，竟然说受邀请品茶名单失踪了，请稍后再试。")
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
			report(w, r, "你好，茶博士满头大汗说，邀请品茶名单被狗叼进了花园，请稍候。")
			return false
		}
		if !isInvited {
			report(w, r, "你好，茶博士无比惊讶说，陛下你的大名竟然不在邀请品茶名单上。")
			return false
		}
		return true
	default:
		report(w, r, "你好，茶博士失魂鱼，竟然说受邀请品茶名单失踪了，请稍后再试。")
		return false
	}
}

// 准备茶围页面数据
func prepareObjectivePageData(objective data.Objective, userData *data.UserPageData) (*data.ObjectiveDetail, error) {
	objectiveBean, err := fetchObjectiveBean(objective)
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
func fetchThreadBean(thread data.Thread, r *http.Request) (tB data.ThreadBean, err error) {
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
func fetchThreadBeanSlice(thread_slice []data.Thread, r *http.Request) (ThreadBeanSlice []data.ThreadBean, err error) {
	var beanSlice []data.ThreadBean
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
func FetchObjectiveBeanSlice(objectiv_slice []data.Objective) (ObjectiveBeanSlice []data.ObjectiveBean, err error) {
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
func fetchObjectiveBean(ob data.Objective) (ObjectiveBean data.ObjectiveBean, err error) {
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
func fetchProjectBeanSlice(project_slice []data.Project) (ProjectBeanSlice []data.ProjectBean, err error) {
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
func fetchProjectBean(project data.Project) (ProjectBean data.ProjectBean, err error) {
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
func fetchPostBeanSlice(post_slice []data.Post) (PostBeanSlice []data.PostBean, err error) {
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
func fetchPostBean(post data.Post) (PostBean data.PostBean, err error) {
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

// 创建AcceptObject并发送邻座蒙评消息
func createAndSendAcceptMessage(objectId int, objectType int, excludeUserId int) error {
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
		FromUserId:     data.UserId_Captain_Spaceship,
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
func acceptNewProject(objectId int) error {
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
func acceptNewObjective(objectId int) (*data.Objective, error) {
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
func acceptNewTeam(objectId int) (*data.Team, error) {
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
func acceptNewDraftThread(objectId int) (*data.Thread, error) {
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
func acceptNewDraftPost(objectId int) (*data.Post, error) {
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
