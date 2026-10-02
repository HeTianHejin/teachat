package route

import (
	"errors"
	"fmt"
	dao "teachat/DAO"
	util "teachat/Util"
)

// 据给出的team参数，去获取对应的茶团资料，是否开放，成员计数，发起日期，发起人（Founder）及其默认团队，然后按结构拼装返回。
func fetchTeamBean(team dao.Team) (TeamBean dao.TeamBean, err error) {
	if team.Id == dao.TeamIdNone {
		return TeamBean, fmt.Errorf("team id is none")
	}

	TeamBean.Team = team
	TeamBean.CreatedAtDate = team.CreatedAtDate()

	founder, err := team.Founder()
	if err != nil {
		util.Debug(" Cannot read team founder %v", err)
		return
	}
	TeamBean.Founder = founder

	TeamBean.FounderDefaultFamily, err = getLastDefaultFamilyByUserId(founder.Id)
	if err != nil {
		util.Debug(" Cannot read team founder default family %v", err)
		return
	}

	TeamBean.FounderTeam, err = founder.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read team founder default team %v", err)
		return
	}

	activeMemberCount, err := team.NumActiveMembers()
	if err != nil {
		util.Debug(" Cannot read team active member count %v", err)
		return
	}
	TeamBean.MembersCount = activeMemberCount

	if team.Id == dao.TeamIdFreelancer {
		//茶友的默认团队还是“自由人”的情况
		TeamBean.CEO = founder
		TeamBean.CEOTeam = TeamBean.FounderTeam
		TeamBean.CEODefaultFamily = TeamBean.FounderDefaultFamily
		return TeamBean, nil
	}

	member_ceo, err := team.MemberCEO()
	if err != nil {
		util.Debug(" Cannot read team member ceo given team_id:  %v", err)
		return
	}
	ceo, err := dao.GetUser(member_ceo.UserId)
	if err != nil {
		util.Debug(" Cannot read team ceo given team_id:  %v", err)
		return
	}
	TeamBean.CEO = ceo
	TeamBean.CEOTeam, err = ceo.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read team ceo default team %v", err)
		return
	}
	TeamBean.CEODefaultFamily, err = getLastDefaultFamilyByUserId(ceo.Id)
	if err != nil {
		util.Debug(" Cannot read team ceo default family %v", err)
		return
	}

	return TeamBean, nil
}

// 根据给出的茶团队列，查询，获取对应的茶团资料夹
func fetchTeamBeanSlice(team_slice []dao.Team) (TeamBeanSlice []dao.TeamBean, err error) {
	if len(team_slice) == 0 {
		return TeamBeanSlice, errors.New("team_slice is empty")
	}
	for _, tea := range team_slice {
		teamBean, err := fetchTeamBean(tea)
		if err != nil {
			return nil, err
		}
		TeamBeanSlice = append(TeamBeanSlice, teamBean)
	}
	return
}

// fetchTeamMemberBean() 根据给出的TeamMember参数，去获取对应的团队成员资料夹
func fetchTeamMemberBean(tm dao.TeamMember) (TMB dao.TeamMemberBean, err error) {
	u, err := dao.GetUser(tm.UserId)
	if err != nil {
		util.Debug(" Cannot read user given TeamMember %v", err)
		return TMB, err
	}
	TMB.Member = u

	team, err := dao.GetTeam(tm.TeamId)
	if err != nil {
		util.Debug(" Cannot read team given team member %v", err)
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
		util.Error("Cannot get ceo of team %d: %v", team.Id, err)
		return
	}
	if member_ceo.UserId == u.Id {
		TMB.IsCEO = true
	} else {
		TMB.IsCEO = false
	}

	teamCoreMembers, err := team.CoreMembers()
	if err != nil {
		util.Debug(" Cannot get team core member FetchTeamMemberBean() %v", err)
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
		util.Debug(" Cannot get GetLastDefaultTeam FetchTeamMemberBean() %v", err)
		return
	}
	TMB.MemberDefaultTeam = member_default_team

	TMB.TeamMember = tm

	TMB.CreatedAtDate = team.CreatedAtDate()

	return TMB, nil
}

// FtchTeamMemberBeanSlice() 根据给出的TeamMember列表参数，去获取对应的团队成员资料夹列表
func fetchTeamMemberBeanSlice(tm_slice []dao.TeamMember) (TMB_slice []dao.TeamMemberBean, err error) {
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
func fetchMemberApplicationBean(ma dao.MemberApplication) (MemberApplicationBean dao.MemberApplicationBean, err error) {
	MemberApplicationBean.MemberApplication = ma
	MemberApplicationBean.Status = ma.GetStatus()

	team, err := dao.GetTeam(ma.TeamId)
	if err != nil {
		util.Debug(" Cannot read team given author %v", err)
		return MemberApplicationBean, err
	}

	MemberApplicationBean.Team = team

	MemberApplicationBean.Author, err = dao.GetUser(ma.UserId)
	if err != nil {
		util.Debug(" Cannot read member application author %v", err)
		return MemberApplicationBean, err
	}
	MemberApplicationBean.AuthorTeam, err = MemberApplicationBean.Author.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read member application author default team %v", err)
		return MemberApplicationBean, err
	}

	MemberApplicationBean.CreatedAtDate = ma.CreatedAtDate()
	return MemberApplicationBean, nil
}
func fetchMemberApplicationBeanSlice(ma_slice []dao.MemberApplication) (MemberApplicationBeanSlice []dao.MemberApplicationBean, err error) {
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
func fetchInvitationBean(i dao.Invitation) (I_B dao.InvitationBean, err error) {
	I_B.Invitation = i

	I_B.Team, err = i.Team()
	if err != nil {
		util.Debug(" Cannot read invitation default team %v", err)
		return I_B, err
	}

	I_B.Author, err = i.Author()
	if err != nil {
		util.Debug(" Cannot fetch team CEO given invitation %v", err)
		return I_B, err
	}

	I_B.InviteUser, err = i.ToUser()
	if err != nil {
		util.Debug(" Cannot read invitation invite user %v", err)
		return I_B, err
	}

	I_B.Status = i.GetStatus()
	return I_B, nil
}

// fetchInvitationBeanSlice() 根据给出的Invitation列表参数，去获取对应的邀请书资料夹列表
func fetchInvitationBeanSlice(i_slice []dao.Invitation) (I_B_slice []dao.InvitationBean, err error) {
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
func fetchTeamMemberRoleNoticeBean(tmrn dao.TeamMemberRoleNotice) (tmrnBean dao.TeamMemberRoleNoticeBean, err error) {
	tmrnBean.TeamMemberRoleNotice = tmrn

	tmrnBean.Team, err = dao.GetTeam(tmrn.TeamId)
	if err != nil {
		util.Debug(" Cannot read team given team member role notice %v", err)
		return tmrnBean, err
	}

	tmrnBean.CEO, err = dao.GetUser(tmrn.CeoId)
	if err != nil {
		util.Debug(" Cannot read ceo given team member role notice %v", err)
		return tmrnBean, err
	}

	tm := dao.TeamMember{Id: tmrn.MemberId}
	if err = tm.Get(); err != nil {
		util.Debug(" Cannot read team member given team member role notice %v", err)
		return tmrnBean, err
	}
	tmrnBean.Member, err = dao.GetUser(tm.UserId)
	if err != nil {
		util.Debug(" Cannot read member given team member role notice %v", err)
		return tmrnBean, err
	}
	tmrnBean.MemberDefaultTeam, err = tmrnBean.Member.GetLastDefaultTeam()
	if err != nil {
		util.Debug(" Cannot read member default team given team member role notice %v", err)
		return tmrnBean, err
	}

	return tmrnBean, nil
}

// fetchTeamMemberRoleNoticeBeanSlice() 根据给出的TeamMemberRoleNotice列表参数，去获取对应的团队成员角色通知资料夹列表
func fetchTeamMemberRoleNoticeBeanSlice(tmrn_slice []dao.TeamMemberRoleNotice) (tmrnBeanSlice []dao.TeamMemberRoleNoticeBean, err error) {
	for _, tmrn := range tmrn_slice {
		tmrnBean, err := fetchTeamMemberRoleNoticeBean(tmrn)
		if err != nil {
			return nil, err
		}
		tmrnBeanSlice = append(tmrnBeanSlice, tmrnBean)
	}
	return
}
