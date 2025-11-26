package route

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// 获取用户最后一次设定的“默认家庭”
// 如果用户没有设定默认家庭，则返回名称为“四海为家”(未知)家庭
// route/family.go
func getLastDefaultFamilyByUserId(userID int) (data.Family, error) {
	user, err := data.GetUser(userID)
	if err != nil {
		return data.Family{}, fmt.Errorf("failed to get user: %w", err)
	}

	family, err := user.GetLastDefaultFamily()
	switch {
	case err == nil:
		return family, nil
	case errors.Is(err, sql.ErrNoRows):
		return data.FamilyUnknown, nil
	default:
		return data.Family{}, fmt.Errorf("failed to get default family: %w", err)
	}
}

// fetchUserDefaultDataBeanForBiography 为名片页面获取用户资料（轻量级）
func fetchUserDefaultDataBeanForBiography(user data.User) (userbean data.UserDefaultDataBean, err error) {
	userbean.User = user

	// 获取默认家庭
	default_family, err := getLastDefaultFamilyByUserId(user.Id)
	if err != nil {
		return userbean, err
	}

	userbean.DefaultFamily = default_family

	// 获取默认团队
	default_team, err := user.GetLastDefaultTeam()
	if err != nil {
		return
	}

	userbean.DefaultTeam = default_team

	return
}

// Fetch userbean given user 根据user参数，查询用户资料荚,包括默认的家庭，团队，地方，
func fetchUserDefaultBean(user data.User) (userbean data.UserDefaultDataBean, err error) {

	userbean.User = user

	default_family, err := getLastDefaultFamilyByUserId(user.Id)
	if err != nil {
		return userbean, err
	}

	userbean.DefaultFamily = default_family

	default_team, err := user.GetLastDefaultTeam()
	if err != nil {
		return
	}

	userbean.DefaultTeam = default_team
	default_place, err := user.GetLastDefaultPlace()
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return
	}
	userbean.DefaultPlace = default_place

	return
}

// fetch userbean_slice given []user
func fetchUserDefaultDataBeanSlice(user_slice []data.User) (userbean_slice []data.UserDefaultDataBean, err error) {
	for _, user := range user_slice {
		userbean, err := fetchUserDefaultBean(user)
		if err != nil {
			return nil, err
		}
		userbean_slice = append(userbean_slice, userbean)
	}
	return
}

// Fetch and process user-related data,从会话查获当前浏览用户资料荚,包括默认团队，全部已经加入的状态正常团队
func fetchSessionUserRelatedData(sess data.Session) (s_u data.User, family data.Family, families []data.Family, team data.Team, teams []data.Team, place data.Place, places []data.Place, err error) {
	// 读取已登陆用户资料
	s_u, err = sess.User()
	if err != nil {
		return
	}

	member_default_family, err := getLastDefaultFamilyByUserId(s_u.Id)
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
	member_all_families = append(member_all_families, data.FamilyUnknown)
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
	teamFreelancer, err := data.GetTeam(data.TeamIdFreelancer)
	if err != nil {
		util.Debug("cannot fetch team by id", data.TeamIdFreelancer, err)
		return
	}
	survivalTeams = append(survivalTeams, teamFreelancer)

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
	user, defaultFamily, survivalFamilies, defaultTeam, survivalTeams, defaultPlace, places, err := fetchSessionUserRelatedData(*sess)
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

// 据给出的team参数，去获取对应的茶团资料，是否开放，成员计数，发起日期，发起人（Founder）及其默认团队，然后按结构拼装返回。
func fetchTeamBean(team data.Team) (TeamBean data.TeamBean, err error) {
	if team.Id == data.TeamIdNone {
		return TeamBean, fmt.Errorf("team id is none")
	}

	TeamBean.Team = team
	TeamBean.CreatedAtDate = team.CreatedAtDate()

	founder, err := team.Founder()
	if err != nil {
		util.Debug(" Cannot read team founder", err)
		return
	}
	TeamBean.Founder = founder

	TeamBean.FounderDefaultFamily, err = getLastDefaultFamilyByUserId(founder.Id)
	if err != nil {
		util.Debug(" Cannot read team founder default family", err)
		return
	}

	TeamBean.FounderTeam, err = founder.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read team founder default team", err)
		return
	}

	TeamBean.MembersCount = team.NumMembers()

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
	TeamBean.CEODefaultFamily, err = getLastDefaultFamilyByUserId(ceo.Id)
	if err != nil {
		util.Debug(" Cannot read team ceo default family", ceo.Id, err)
		return
	}

	return TeamBean, nil
}

// 根据给出的茶团队列，查询，获取对应的茶团资料夹
func fetchTeamBeanSlice(team_slice []data.Team) (TeamBeanSlice []data.TeamBean, err error) {
	for _, tea := range team_slice {
		teamBean, err := fetchTeamBean(tea)
		if err != nil {
			return nil, err
		}
		TeamBeanSlice = append(TeamBeanSlice, teamBean)
	}
	return
}

// 根据给出的family参数，从数据库获取对应的家庭资料
func fetchFamilyBean(family data.Family) (FamilyBean data.FamilyBean, err error) {
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

	FamilyBean.MemberCount, err = data.CountFamilyMembers(family.Id)
	if err != nil {
		util.Debug(family.AuthorId, " Cannot read family member count")
		return FamilyBean, err
	}
	return
}

// 根据给出的家庭队列，查询，获取对应的家庭茶团资料集合
func fetchFamilyBeanSlice(family_slice []data.Family) (FamilyBeanSlice []data.FamilyBean, err error) {
	for _, fam := range family_slice {
		familyBean, err := fetchFamilyBean(fam)
		if err != nil {
			return nil, err
		}
		FamilyBeanSlice = append(FamilyBeanSlice, familyBean)
	}
	return
}

// fetchFamilyMemberBean() 根据给出的FamilyMember参数，去获取对应的家庭成员资料夹
func fetchFamilyMemberBean(fm data.FamilyMember) (FMB data.FamilyMemberBean, err error) {
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

	member_default_family, err := getLastDefaultFamilyByUserId(fm.UserId)
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

// fetchFamilyMemberBeanSlice() 根据给出的FamilyMember列表参数，去获取对应的家庭成员资料夹列表
func fetchFamilyMemberBeanSlice(fm_slice []data.FamilyMember) (FMB_slice []data.FamilyMemberBean, err error) {
	for _, fm := range fm_slice {
		fmBean, err := fetchFamilyMemberBean(fm)
		if err != nil {
			return nil, err
		}
		FMB_slice = append(FMB_slice, fmBean)
	}
	return
}

// 根据给出的某个&家庭茶团增加成员声明书，获取&家庭茶团增加成员声明书资料夹
func fetchFamilyMemberSignInBean(fmsi data.FamilyMemberSignIn) (FMSIB data.FamilyMemberSignInBean, err error) {
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
// func fetchFamilyMemberSignInBeanSlice(fmsi_slice []data.FamilyMemberSignIn) (FMSIB_slice []data.FamilyMemberSignInBean, err error) {
// 	for _, fmsi := range fmsi_slice {
// 		fmsiBean, err := fetchFamilyMemberSignInBean(fmsi)
// 		if err != nil {
// 			return nil, err
// 		}
// 		FMSIB_slice = append(FMSIB_slice, fmsiBean)
// 	}
// 	return
// }

// fetchTeamMemberBean() 根据给出的TeamMember参数，去获取对应的团队成员资料夹
func fetchTeamMemberBean(tm data.TeamMember) (TMB data.TeamMemberBean, err error) {
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
func fetchTeamMemberBeanSlice(tm_slice []data.TeamMember) (TMB_slice []data.TeamMemberBean, err error) {
	for _, tm := range tm_slice {
		tmBean, err := fetchTeamMemberBean(tm)
		if err != nil {
			return nil, err
		}
		TMB_slice = append(TMB_slice, tmBean)
	}
	return
}

// 根据给出的MemberApplication参数，去获取对应的加盟申请书资料夹
func fetchMemberApplicationBean(ma data.MemberApplication) (MemberApplicationBean data.MemberApplicationBean, err error) {
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
func fetchMemberApplicationBeanSlice(ma_slice []data.MemberApplication) (MemberApplicationBeanSlice []data.MemberApplicationBean, err error) {
	for _, ma := range ma_slice {
		maBean, err := fetchMemberApplicationBean(ma)
		if err != nil {
			return nil, err
		}
		MemberApplicationBeanSlice = append(MemberApplicationBeanSlice, maBean)
	}
	return
}

// fetchInvitationBean() 根据给出的Invitation参数，去获取对应的邀请书资料夹
func fetchInvitationBean(i data.Invitation) (I_B data.InvitationBean, err error) {
	I_B.Invitation = i

	I_B.Team, err = i.Team()
	if err != nil {
		util.Debug(" Cannot read invitation default team", err)
		return I_B, err
	}

	I_B.Author, err = i.Author()
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

// fetchInvitationBeanSlice() 根据给出的Invitation列表参数，去获取对应的邀请书资料夹列表
func fetchInvitationBeanSlice(i_slice []data.Invitation) (I_B_slice []data.InvitationBean, err error) {
	for _, i := range i_slice {
		iBean, err := fetchInvitationBean(i)
		if err != nil {
			return nil, err
		}
		I_B_slice = append(I_B_slice, iBean)
	}
	return
}

// fetchTeamMemberRoleNoticeBean() 根据给出的TeamMemberRoleNotice参数，去获取对应的团队成员角色通知资料夹
func fetchTeamMemberRoleNoticeBean(tmrn data.TeamMemberRoleNotice) (tmrnBean data.TeamMemberRoleNoticeBean, err error) {
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

// fetchTeamMemberRoleNoticeBeanSlice() 根据给出的TeamMemberRoleNotice列表参数，去获取对应的团队成员角色通知资料夹列表
func fetchTeamMemberRoleNoticeBeanSlice(tmrn_slice []data.TeamMemberRoleNotice) (tmrnBeanSlice []data.TeamMemberRoleNoticeBean, err error) {
	for _, tmrn := range tmrn_slice {
		tmrnBean, err := fetchTeamMemberRoleNoticeBean(tmrn)
		if err != nil {
			return nil, err
		}
		tmrnBeanSlice = append(tmrnBeanSlice, tmrnBean)
	}
	return
}

// 检查并设置用户默认团队（非自由人占位团队）
func setUserDefaultTeam(founder *data.User, newTeamID int, w http.ResponseWriter, r *http.Request) bool {
	// 获取用户当前默认团队
	oldDefaultTeam, err := founder.GetLastDefaultTeam()
	if err != nil {
		util.Debug(founder.Email, "Cannot get last default team")
		report(w, r, "你好，茶博士失魂鱼，手滑未能创建你的天命使团，请稍后再试。")
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
			report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
			return false
		}
	}
	return true
}

// isVerifier 检查用户是否为见证者
func isVerifier(userId int) bool {
	verifier_team := data.Team{Id: data.TeamIdVerifier}
	is_member, err := verifier_team.IsMember(userId)
	if err != nil {
		util.Debug(" Cannot check team member", err)
		return false
	}
	return is_member
}
