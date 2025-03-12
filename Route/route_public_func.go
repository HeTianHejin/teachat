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
   存放各个路由文件共享的一些方法
*/

// 默认家庭ID常量，默认家庭ID为0
// DeepSeek建议使用默认家庭ID
const DefaultFamilyId = 0
const DefaultFamilyUuid = "x"

// 默认未明确资料的家庭="四海为家",id=0
// 任何人均是来自某个家庭，但是单独的个体，即使成年，属于一个未来家庭的成员之一，不能视为一个家庭。
var DefaultFamily = data.Family{
	Id:           DefaultFamilyId,
	Uuid:         "x", //代表未知数
	Name:         "四海为家",
	AuthorId:     1, //表示系统预设的值
	Introduction: "存在但未明确资料的家庭",
}

// 默认的系统“自由人”$事业茶团
// 刚注册或者没有加入任何$团队的用户，属于基础$事业茶团
var FreelancerTeam = data.Team{
	Id:                2,
	Uuid:              "72c06442-2b60-418a-6493-a91bd03ae4k8",
	Name:              "特立独行的自由人",
	Mission:           "星际旅行特立独行的自由人，不属于任何私有$茶团。",
	FounderId:         1, //表示系统预设的值
	Class:             0,
	Abbreviation:      "自由人",
	Logo:              "teamLogo",
	SuperiorTeamId:    0,
	SubordinateTeamId: 0,
}

// 获取用户最后一次设定的“默认家庭”
// 如果用户没有设定默认家庭，则返回默认id=0，名称为“四海为家”家庭
func GetLastDefaultFamilyByUserId(user_id int) (family data.Family, err error) {
	family = data.Family{}
	user, err := data.GetUser(user_id)
	if err != nil {
		return
	}

	family, err = user.GetLastDefaultFamily()
	if err != nil {
		// 未设定默认家庭
		if errors.Is(err, sql.ErrNoRows) {
			return DefaultFamily, nil
		}
		return
	}
	return
}

// 根据茶语中登记的家庭ID，获取某个茶语(茶围、茶台、茶议和品味)发布时选择的家庭
func GetFamilyByFamilyId(family_id int) (family data.Family, err error) {
	if family_id == 0 {
		return DefaultFamily, nil
	}
	family = data.Family{Id: family_id}
	if err = family.Get(); err != nil {
		return
	}

	return
}

// 记录用户最后的查询路径和参数
func RecordLastQueryPath(sess_user_id int, path, raw_query string) (err error) {
	lq := data.LastQuery{
		UserId:  sess_user_id,
		Path:    path,
		Query:   raw_query,
		QueryAt: time.Now(),
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
	if err != nil && err != sql.ErrNoRows {
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
func FetchUserRelatedData(sess data.Session) (s_u data.User, family data.Family, families []data.Family, team data.Team, teams []data.Team, place data.Place, places []data.Place, err error) {
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

	default_place, err := s_u.GetLastDefaultPlace()
	if err != nil && err != sql.ErrNoRows {
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

// 根据给出的thread_slice参数，去获取对应的茶议（截短正文保留前168字符），附属品味计数，作者资料，作者发帖时候选择的茶团。然后按结构拼装返回
func FetchThreadBeanSlice(thread_slice []data.Thread) (ThreadBeanSlice []data.ThreadBean, err error) {
	var oabslice []data.ThreadBean
	// 截短ThreadSlice中thread.Body文字长度为168字符,
	// 展示时长度接近，排列比较整齐，最小惊讶原则？效果比较nice
	for i := range thread_slice {
		thread_slice[i].Body = Substr(thread_slice[i].Body, 168)
	}
	for _, thread := range thread_slice {
		ThreadBean, err := FetchThreadBean(thread)
		if err != nil {
			return nil, err
		}
		oabslice = append(oabslice, ThreadBean)
	}
	ThreadBeanSlice = oabslice
	return
}

// 根据给出的thread参数，去获取对应的茶议，附属品味计数，作者资料，作者发帖时候选择的茶团，费用和费时。
func FetchThreadBean(thread data.Thread) (ThreadBean data.ThreadBean, err error) {
	var tB data.ThreadBean
	tB.Thread = thread
	tB.Status = thread.Status()
	tB.Count = thread.NumReplies()
	tB.CreatedAtDate = thread.CreatedAtDate()
	//作者资料
	author, err := thread.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read thread author")
		return
	}
	tB.Author = author
	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）
	tB.AuthorFamily, err = GetFamilyByFamilyId(thread.FamilyId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read thread author family")
		return
	}

	tB.AuthorTeam, err = data.GetTeam(thread.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read thread author team")
		return
	}

	//idea是否被采纳
	tB.IsApproved = thread.IsApproved()

	return tB, nil
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
func FetchObjectiveBean(o data.Objective) (ObjectiveBean data.ObjectiveBean, err error) {
	var oB data.ObjectiveBean

	oB.Objective = o
	if o.Class == 1 {
		oB.Open = true
	} else {
		oB.Open = false
	}
	oB.Status = o.GetStatus()
	oB.Count = o.NumReplies()
	oB.CreatedAtDate = o.CreatedAtDate()
	user, err := o.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read objective author")
		return
	}
	oB.Author = user

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	oB.AuthorFamily, err = GetFamilyByFamilyId(o.FamilyId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read objective author family")
		return
	}

	oB.AuthorTeam, err = data.GetTeam(o.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read objective author team")
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
	pb.Count = project.NumReplies()
	pb.CreatedAtDate = project.CreatedAtDate()
	author, err := project.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project author")
		return
	}
	pb.Author = author

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	pb.AuthorFamily, err = GetFamilyByFamilyId(project.FamilyId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project author family")
		return
	}

	pb.AuthorTeam, err = data.GetTeam(project.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read project author team")
		return
	}

	pb.Place, err = project.Place()
	if err != nil {
		util.ScaldingTea(util.LogError(err), "cannot read project place")
		return pb, err
	}
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
	PostBean.Count = post.NumReplies()
	PostBean.CreatedAtDate = post.CreatedAtDate()
	author, err := post.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read post author")
		return
	}
	PostBean.Author = author

	//作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）

	family, err := GetFamilyByFamilyId(post.FamilyId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read post author family")
		return
	}
	PostBean.AuthorFamily = family

	team, err := data.GetTeam(post.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read post author team")
		return
	}
	PostBean.AuthorTeam = team

	return PostBean, nil
}

// 据给出的team参数，去获取对应的茶团资料，是否开放，成员计数，发起日期，发起人（Founder）及其默认团队，然后按结构拼装返回。
func FetchTeamBean(team data.Team) (TeamBean data.TeamBean, err error) {
	TeamBean.Team = team
	if team.Class == 1 {
		TeamBean.Open = true
	} else {
		TeamBean.Open = false
	}
	TeamBean.CreatedAtDate = team.CreatedAtDate()

	founder, err := team.Founder()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read team founder")
		return
	}
	TeamBean.Founder = founder

	founder_default_family, err := GetLastDefaultFamilyByUserId(founder.Id)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read team founder default family")
		return
	}
	TeamBean.FounderDefaultFamily = founder_default_family

	TeamBean.FounderTeam, err = founder.GetLastDefaultTeam()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read team founder default team")
		return
	}

	TeamBean.Count = team.NumMembers()

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
		util.ScaldingTea(util.LogError(err), family.AuthorId, " Cannot read family founder")
		return FamilyBean, err
	}
	FamilyBean.FounderTeam, err = FamilyBean.Founder.GetLastDefaultTeam()
	if err != nil {
		util.ScaldingTea(util.LogError(err), family.AuthorId, " Cannot read family founder default team")
		return FamilyBean, err
	}

	FamilyBean.Count, err = data.CountFamilyMembers(family.Id)
	if err != nil {
		util.ScaldingTea(util.LogError(err), family.AuthorId, " Cannot read family member count")
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
		util.ScaldingTea(util.LogError(err), " Cannot read user given FamilyMember")
		return FMB, err
	}
	FMB.Member = u
	default_team, err := u.GetLastDefaultTeam()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read user given FamilyMember")
		return FMB, err
	}
	FMB.MemberDefaultTeam = default_team

	f := data.Family{Id: fm.FamilyId}

	//读取茶团的parent_members
	family_parent_members, err := f.ParentMembers()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get family core member FetchFamilyMemberBean()")
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
		util.ScaldingTea(util.LogError(err), " Cannot get GetLastDefaultFamily FetchFamilyMemberBean()")
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
		util.ScaldingTea(util.LogError(err), " Cannot read family given FamilyMemberSignIn")
		return FMSIB, err
	}
	FMSIB.Family = family

	FMSIB.NewMember, err = data.GetUser(fmsi.UserId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read new member given FamilyMemberSignIn")
		return FMSIB, err
	}

	FMSIB.Author, err = data.GetUser(fmsi.AuthorUserId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read author given FamilyMemberSignIn")
		return FMSIB, err
	}

	place := data.Place{Id: fmsi.PlaceId}
	if err = place.Get(); err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read place given FamilyMemberSignIn")
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
		util.ScaldingTea(util.LogError(err), " Cannot read user given TeamMember")
		return TMB, err
	}
	TMB.Member = u

	team, err := data.GetTeam(tm.TeamId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read team given team member")
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
		util.ScaldingTea(util.LogError(err), team.Id, " Cannot get ceo of this team")
		return
	}
	if member_ceo.UserId == u.Id {
		TMB.IsCEO = true
	} else {
		TMB.IsCEO = false
	}

	teamCoreMembers, err := team.CoreMembers()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot get team core member FetchTeamMemberBean()")
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
		util.ScaldingTea(util.LogError(err), " Cannot get GetLastDefaultTeam FetchTeamMemberBean()")
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
		util.ScaldingTea(util.LogError(err), " Cannot read team given author")
		return MemberApplicationBean, err
	}
	MemberApplicationBean.Team = team

	MemberApplicationBean.Author, err = data.GetUser(ma.UserId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read member application author")
		return MemberApplicationBean, err
	}
	MemberApplicationBean.AuthorTeam, err = MemberApplicationBean.Author.GetLastDefaultTeam()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read member application author default team")
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
		util.ScaldingTea(util.LogError(err), " Cannot read invitation default team")
		return I_B, err
	}

	I_B.AuthorCEO, err = i.AuthorCEO()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot fetch team CEO given invitation")
		return I_B, err
	}

	I_B.InviteUser, err = i.ToUser()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read invitation invite user")
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
		util.ScaldingTea(util.LogError(err), " Cannot read team given team member role notice")
		return tmrnBean, err
	}

	tmrnBean.CEO, err = data.GetUser(tmrn.CeoId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read ceo given team member role notice")
		return tmrnBean, err
	}

	tm := data.TeamMember{Id: tmrn.MemberId}
	if err = tm.Get(); err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read team member given team member role notice")
		return tmrnBean, err
	}
	tmrnBean.Member, err = data.GetUser(tm.UserId)
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read member given team member role notice")
		return tmrnBean, err
	}
	tmrnBean.MemberDefaultTeam, err = tmrnBean.Member.GetLastDefaultTeam()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read member default team given team member role notice")
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

// 据给出的 group 参数，去获取对应的 group 资料，是否开放，下属茶团计数，发起日期，发起人（Founder）及其默认团队，第一团队，然后按结构拼装返回。
func FetchGroupBean(group data.Group) (GroupBean data.GroupBean, err error) {
	var gb data.GroupBean
	gb.Group = group
	if group.Class == 1 {
		gb.Open = true
	} else {
		gb.Open = false
	}
	gb.CreatedAtDate = group.CreatedAtDate()
	u, _ := data.GetUser(group.FounderId)
	gb.Founder = u
	gb.FounderTeam, err = u.GetLastDefaultTeam()
	if err != nil {
		util.ScaldingTea(util.LogError(err), " Cannot read team given founder")
		return gb, err
	}
	gb.TeamsCount = data.GetTeamsCountByGroupId(gb.Group.Id)
	gb.Count = group.NumMembers()
	return gb, nil
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
	if ext != ".jpeg" {
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
	newFilePath := data.ImageDir + uuid + data.ImageExt
	newFile, err := os.Create(newFilePath)
	if err != nil {
		util.ScaldingTea(util.LogError(err), "创建头像文件名失败")
		Report(w, r, "创建头像文件失败，请稍后再试。")
		return err
	}
	// 确保文件在函数执行完毕后关闭
	defer newFile.Close()

	// 通过缓存方法写入硬盘
	buff := bufio.NewWriter(newFile)
	buff.Write(fileBytes)
	err = buff.Flush()
	if err != nil {
		util.ScaldingTea(util.LogError(err), "fail to write avatar image")
		Report(w, r, "你好，茶博士居然说没有墨水了，写入头像文件不成功，请稍后再试。")
		return err
	}

	// _, err = newFile.Write(fileBytes)
	return nil
}

// 茶博士向茶客报告信息的方法，包括但不限于意外事件和通知、感谢等等提示。
// 茶博士——古时专指陆羽。陆羽著《茶经》，唐德宗李适曾当面称陆羽为“茶博士”。
// 茶博士-teaOffice，是古代中华传统文化对茶馆工作人员的昵称，如：富家宴会，犹有专供茶事之人，谓之茶博士。——唐代《西湖志馀》
// 现在多指精通茶艺的师傅，尤其是四川的长嘴壶茶艺，茶博士个个都是身怀绝技的“高手”。
func Report(w http.ResponseWriter, r *http.Request, msg string) {
	var userBPD data.UserBean
	userBPD.Message = msg
	s, err := Session(r)
	if err != nil {
		userBPD.SessUser = data.User{
			Id:   0,
			Name: "游客",
		}
		RenderHTML(w, &userBPD, "layout", "navbar.public", "feedback")
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.ScaldingTea(util.LogError(err), "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	userBPD.SessUser = s_u

	// 记录用户最后查询的资讯
	// if err = RecordLastQueryPath(s_u.Id, r.URL.Path, r.URL.RawQuery); err != nil {
	// 	util.PanicTea(util.LogError(err), s_u.Id, " Cannot record last query path")
	// }
	RenderHTML(w, &userBPD, "layout", "navbar.private", "feedback")
}

// Checks if the user is logged in and has a Session, if not err is not nil
func Session(r *http.Request) (sess data.Session, err error) {
	cookie, err := r.Cookie("_cookie")
	if err == nil {
		sess = data.Session{Uuid: cookie.Value}
		if ok, _ := sess.Check(); !ok {
			err = errors.New("invalid session")
		}
	}
	return

}

// parse HTML templates
// pass in a slice of file names, and get a template
func ParseTemplateFiles(filenames ...string) (t *template.Template) {
	var files []string
	t = template.New("layout")
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.go.html", file))
	}
	t = template.Must(t.ParseFiles(files...))
	return
}

// 处理器把页面模版和需求数据揉合后，由这个方法，将填写好的页面“制作“成HTML格式，调用http响应方法，发送给浏览器端客户
func RenderHTML(w http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.go.html", file))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(w, "layout", data)
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

// 验证team_id_slice:"2,19,87..."字符串格式是否正确，正确返回true，错误返回false。
func Verify_id_slice_Format(team_id_slice string) bool {
	if team_id_slice == "" {
		return false
	}
	pattern := `^[0-9]+(,[0-9]+)*$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(team_id_slice)
}

// 输入两个统计数（辩论的正方累积得分数，辩论总得分数）（整数），计算前者与后者比值，结果浮点数向上四舍五入取整,
// 返回百分数的分子整数
func ProgressRound(numerator, denominator int) int {
	if denominator == 0 {
		// 分母为0时，视作未有记录，即未进行表决状态，返回100
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
	return int(math.Floor(ratio + 0.5))
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

// 计算中文字符串长度
func CnStrLen(str string) int {
	return utf8.RuneCountInString(str)
}

// 对未经盲评的草稿进行“火星文”遮盖隐秘处理，即用星号替换50%或者指定更高比例文字
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

// 入参string，截取前面一段指定长度文字，返回string
// 注意，输入负数=最大值
// 参考https://blog.thinkeridea.com/201910/go/efficient_string_truncation.html
func Substr(s string, length int) string {
	//这是根据range的特性加的，如果不加，截取不到最后一个字（end+1=意外，因为1中文=3字节！）
	//str += "."
	var n, i int
	for i = range s {
		if n == length {
			break
		}
		n++
	}

	return s[:i]
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
