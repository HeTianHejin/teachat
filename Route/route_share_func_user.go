package route

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	dao "teachat/DAO"
	util "teachat/Util"
)

// 获取用户最后一次设定的“默认家庭”
// 如果用户没有设定默认家庭，则返回名称为“四海为家”(未知)家庭
// route/family.go
func getLastDefaultFamilyByUserId(userID int) (dao.Family, error) {
	user, err := dao.GetUser(userID)
	if err != nil {
		return dao.Family{}, fmt.Errorf("failed to get user: %w", err)
	}

	family, err := user.GetLastDefaultFamily()
	switch {
	case err == nil:
		return family, nil
	case errors.Is(err, sql.ErrNoRows):
		return dao.FamilyUnknown, nil
	default:
		return dao.Family{}, fmt.Errorf("failed to get default family: %w", err)
	}
}

// fetchUserDefaultDataBeanForBiography 为名片页面获取用户资料（轻量级）
func fetchUserDefaultDataBeanForBiography(user dao.User) (userbean dao.UserDefaultDataBean, err error) {
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
func fetchUserDefaultBean(user dao.User) (userbean dao.UserDefaultDataBean, err error) {

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
func fetchUserDefaultDataBeanSlice(user_slice []dao.User) (userbean_slice []dao.UserDefaultDataBean, err error) {
	for _, user := range user_slice {
		userbean, err := fetchUserDefaultBean(user)
		if err != nil {
			return nil, err
		}
		userbean_slice = append(userbean_slice, userbean)
	}
	return
}

// Fetch and process user-related data,从会话查获当前浏览用户资料荚,包括默认团队，全部已经加入的状态正常团队（排除茶台特殊家庭临时监护团队）
func fetchSessionUserRelatedData(sess dao.Session, ctx context.Context) (s_u dao.User, family dao.Family, families []dao.Family, team dao.Team, teams []dao.Team, place dao.Place, places []dao.Place, err error) {
	// 读取已登陆用户资料
	s_u, err = sess.User()
	if err != nil {
		if err.Error() == "session user_id is 0, invalid user id" {
			return s_u, family, families, team, teams, place, places, fmt.Errorf("invalid session: user_id is 0 for email %s", sess.Email)
		}
		return s_u, family, families, team, teams, place, places, fmt.Errorf("failed to get user from session for email %s: %w", sess.Email, err)
	}

	member_default_family, err := getLastDefaultFamilyByUserId(s_u.Id)
	if err != nil {
		return s_u, family, families, team, teams, place, places, fmt.Errorf("failed to get default family for user %s: %w", s_u.Email, err)
	}

	member_all_families, err := dao.GetAllFamilies(s_u.Id, ctx)
	if err != nil {
		return s_u, family, families, team, teams, place, places, fmt.Errorf("failed to get all families for user %s: %w", s_u.Email, err)
	}
	//remove member_default_family from member_all_families
	// for i, family := range member_all_families {
	// 	if family.Id == member_default_family.Id {
	// 		member_all_families = append(member_all_families[:i], member_all_families[i+1:]...)
	// 		break
	// 	}
	// }
	// 把系统默认的“未知”家庭资料加入families
	member_all_families = append(member_all_families, dao.FamilyUnknown)
	defaultTeam, err := s_u.GetLastDefaultTeam()
	if err != nil {
		return s_u, family, families, team, teams, place, places, fmt.Errorf("failed to get default team for user %s: %w", s_u.Email, err)
	}

	survivalTeams, err := dao.GetUserSurvivalTeams(s_u.Id, ctx)
	if err != nil {
		return s_u, family, families, team, teams, place, places, fmt.Errorf("failed to get survival teams for user %s: %w", s_u.Email, err)
	}
	// for i, team := range survivalTeams {
	// 	if team.Id == defaultTeam.Id {
	// 		survivalTeams = append(survivalTeams[:i], survivalTeams[i+1:]...)
	// 		break
	// 	}
	// }

	default_place, err := s_u.GetLastDefaultPlace()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return s_u, family, families, team, teams, place, places, fmt.Errorf("failed to get default place for user %s: %w", s_u.Email, err)
	}

	places, err = s_u.GetAllBindPlaces()
	if err != nil {
		return s_u, family, families, team, teams, place, places, fmt.Errorf("failed to get all bind places for user %s: %w", s_u.Email, err)
	}
	// if len(places) > 0 {
	// 	//移除默认地方
	// 	for i, place := range places {
	// 		if place.Id == default_place.Id {
	// 			places = append(places[:i], places[i+1:]...)
	// 			break
	// 		}
	// 	}
	// }

	return s_u, member_default_family, member_all_families, defaultTeam, survivalTeams, default_place, places, nil
}

// 准备用户相关数据
func prepareUserPageData(sess *dao.Session, ctx context.Context) (*dao.UserPageData, error) {
	user, defaultFamily, survivalFamilies, defaultTeam, survivalTeams, defaultPlace, places, err := fetchSessionUserRelatedData(*sess, ctx)
	if err != nil {
		return nil, err
	}

	return &dao.UserPageData{
		User:             user,
		DefaultFamily:    defaultFamily,
		SurvivalFamilies: survivalFamilies,
		DefaultTeam:      defaultTeam,
		SurvivalTeams:    survivalTeams,
		DefaultPlace:     defaultPlace,
		BindPlaces:       places,
	}, nil
}

// 根据给出的family参数，从数据库获取对应的家庭资料
func fetchFamilyBean(family dao.Family) (FamilyBean dao.FamilyBean, err error) {
	FamilyBean.Family = family
	//登记人资料
	FamilyBean.Founder, err = dao.GetUser(family.AuthorId)
	if err != nil {
		util.Error("Cannot read family founder for family %d: %v", family.AuthorId, err)
		return FamilyBean, err
	}
	FamilyBean.FounderTeam, err = FamilyBean.Founder.GetLastDefaultTeam()
	if err != nil {
		util.Error("Cannot read family founder default team for family %d: %v", family.AuthorId, err)
		return FamilyBean, err
	}

	FamilyBean.MemberCount, err = dao.CountFamilyMembers(family.Id)
	if err != nil {
		util.Error("Cannot read family member count for family %d: %v", family.Id, err)
		return FamilyBean, err
	}
	return
}

// 根据给出的家庭队列，查询，获取对应的家庭茶团资料集合
func fetchFamilyBeanSlice(family_slice []dao.Family) (FamilyBeanSlice []dao.FamilyBean, err error) {
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
func fetchFamilyMemberBean(fm dao.FamilyMember) (FMB dao.FamilyMemberBean, err error) {
	FMB.FamilyMember = fm

	u, err := dao.GetUser(fm.UserId)
	if err != nil {
		util.Debug(" Cannot read user given FamilyMember %v", err)
		return FMB, err
	}
	FMB.Member = u
	default_team, err := u.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read user given FamilyMember %v", err)
		return FMB, err
	}
	FMB.MemberDefaultTeam = default_team

	f := dao.Family{Id: fm.FamilyId}

	//读取茶团的parent_members
	family_parent_members, err := f.ParentMembers()
	if err != nil {
		util.Debug(" Cannot get family core member FetchFamilyMemberBean() %v", err)
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
		util.Debug(" Cannot get GetLastDefaultFamily FetchFamilyMemberBean() %v", err)
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
func fetchFamilyMemberBeanSlice(fm_slice []dao.FamilyMember) (FMB_slice []dao.FamilyMemberBean, err error) {
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
func fetchFamilyMemberSignInBean(fmsi dao.FamilyMemberSignIn) (FMSIB dao.FamilyMemberSignInBean, err error) {
	FMSIB.FamilyMemberSignIn = fmsi

	family := dao.Family{Id: fmsi.FamilyId}
	if err = family.Get(); err != nil {
		util.Debug(" Cannot read family given FamilyMemberSignIn %v", err)
		return FMSIB, err
	}
	FMSIB.Family = family

	FMSIB.NewMember, err = dao.GetUser(fmsi.UserId)
	if err != nil {
		util.Debug(" Cannot read new member given FamilyMemberSignIn %v", err)
		return FMSIB, err
	}

	FMSIB.Author, err = dao.GetUser(fmsi.AuthorUserId)
	if err != nil {
		util.Debug(" Cannot read author given FamilyMemberSignIn %v", err)
		return FMSIB, err
	}

	place := dao.Place{Id: fmsi.PlaceId}
	if err = place.Get(); err != nil {
		util.Debug(" Cannot read place given FamilyMemberSignIn %v", err)
		return FMSIB, err
	}
	FMSIB.Place = place

	return FMSIB, nil
}

// 根据给出的多个&家庭茶团增加成员声明书队列，获取资料夹队列
// func fetchFamilyMemberSignInBeanSlice(fmsi_slice []dao.FamilyMemberSignIn) (FMSIB_slice []dao.FamilyMemberSignInBean, err error) {
// 	for _, fmsi := range fmsi_slice {
// 		fmsiBean, err := fetchFamilyMemberSignInBean(fmsi)
// 		if err != nil {
// 			return nil, err
// 		}
// 		FMSIB_slice = append(FMSIB_slice, fmsiBean)
// 	}
// 	return
// }

// 检查并设置用户默认团队（非自由人占位团队）
func setUserDefaultTeam(s_u dao.User, newTeamID int, w http.ResponseWriter) bool {
	// 获取用户当前默认团队
	oldDefaultTeam, err := s_u.GetLastDefaultTeam()
	if err != nil {
		util.Debug(s_u.Email, "Cannot get last default team")
		report(w, s_u, "你好，茶博士失魂鱼，手滑未能创建你的天命使团，请稍后再试。")
		return false
	}

	// 检查是否为占位团队（自由人）
	if oldDefaultTeam.Id == dao.TeamIdFreelancer {
		uDT := dao.UserDefaultTeam{
			UserId: s_u.Id,
			TeamId: newTeamID,
		}
		if err := uDT.Create(); err != nil {
			util.Debug(s_u.Email, newTeamID, "Cannot create default team")
			report(w, s_u, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
			return false
		}
	}
	return true
}
