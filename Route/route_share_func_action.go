package route

import (
	"context"
	"fmt"
	dao "teachat/DAO"
)

// 根据给出的projectAppointment参数，去获取对应的projectAppointmentBean资料，然后按结构拼装返回。
func fetchAppointmentBean(pA dao.ProjectAppointment) (pABean dao.ProjectAppointmentBean, err error) {
	pABean.Appointment = pA
	pr := dao.Project{Id: pA.ProjectId}
	if err := pr.Get(); err != nil {
		return pABean, fmt.Errorf("获取茶台(project)失败: %w", err)
	}
	pABean.Project = pr
	payer, err := dao.GetUser(pA.PayerUserId)
	if err != nil {
		return pABean, fmt.Errorf("获取payer失败: %w", err)
	}
	pABean.Payer = payer
	payer_family, err := dao.GetFamily(pA.PayerFamilyId)
	if err != nil {
		return pABean, fmt.Errorf("获取payer_family失败: %w", err)
	}
	pABean.PayerFamily = payer_family
	payer_team, err := dao.GetTeam(pA.PayerTeamId)
	if err != nil {
		return pABean, fmt.Errorf("获取payer_team失败: %w", err)
	}
	pABean.PayerTeam = payer_team
	payee, err := dao.GetUser(pA.PayeeUserId)
	if err != nil {
		return pABean, fmt.Errorf("获取payee失败: %w", err)
	}
	pABean.Payee = payee
	payee_family, err := dao.GetFamily(pA.PayeeFamilyId)
	if err != nil {
		return pABean, fmt.Errorf("获取payee_family失败: %w", err)
	}
	pABean.PayeeFamily = payee_family
	payee_team, err := dao.GetTeam(pA.PayeeTeamId)
	if err != nil {
		return pABean, fmt.Errorf("获取payee_team失败: %w", err)
	}
	pABean.PayeeTeam = payee_team
	verifier, err := dao.GetUser(pA.VerifierUserId)
	if err != nil {
		return pABean, fmt.Errorf("获取verifier失败: %w", err)
	}
	pABean.Verifier = verifier
	verifier_family, err := dao.GetFamily(pA.VerifierFamilyId)
	if err != nil {
		return pABean, fmt.Errorf("获取verifier_family失败: %w", err)
	}
	pABean.VerifierFamily = verifier_family
	verifier_team, err := dao.GetTeam(pA.VerifierTeamId)
	if err != nil {
		return pABean, fmt.Errorf("获取verifier_team失败: %w", err)
	}
	pABean.VerifierTeam = verifier_team
	return pABean, nil
}

// fetchSeeSeekBean() 根据给出的SeeSeek参数，去获取对应的SeeSeekBean资料，然后按结构拼装返回。
func fetchSeeSeekBean(sS dao.SeeSeek) (sSBean dao.SeeSeekBean, err error) {
	sSBean.SeeSeek = sS
	sSBean.IsOpen = (sS.Category == dao.SeeSeekCategoryPublic)

	// 获取项目信息
	pr := dao.Project{Id: sS.ProjectId}
	if err := pr.Get(); err != nil {
		return sSBean, fmt.Errorf("获取项目(project)失败: %w", err)
	}
	sSBean.Project = pr

	// 获取地点信息
	if sS.PlaceId > 0 {
		place := dao.Place{Id: sS.PlaceId}
		if err := place.Get(); err != nil {
			return sSBean, fmt.Errorf("获取地点(place)失败: %w", err)
		}
		sSBean.Place = place
	}

	// 获取环境信息
	if envs, err := sS.GetEnvironments(); err != nil {
		return sSBean, fmt.Errorf("获取环境关联失败: %w", err)
	} else if len(envs) > 0 {
		env := dao.Environment{Id: envs[0].EnvironmentId}
		if err := env.GetByIdOrUUID(); err != nil {
			return sSBean, fmt.Errorf("获取环境(environment)失败: %w", err)
		}
		sSBean.Environment = env
	}

	// 获取隐患信息
	if hazards, err := sS.GetHazards(); err != nil {
		return sSBean, fmt.Errorf("获取隐患关联失败: %w", err)
	} else {
		for _, h := range hazards {
			hazard := dao.Hazard{Id: h.HazardId}
			if err := hazard.GetByIdOrUUID(); err != nil {
				return sSBean, fmt.Errorf("获取隐患(hazard)失败: %w", err)
			}
			sSBean.Hazard = append(sSBean.Hazard, hazard)
		}
	}

	// 获取风险信息
	if risks, err := sS.GetRisks(); err != nil {
		return sSBean, fmt.Errorf("获取风险关联失败: %w", err)
	} else {
		for _, r := range risks {
			risk := dao.Risk{Id: r.RiskId}
			if err := risk.GetByIdOrUUID(); err != nil {
				return sSBean, fmt.Errorf("获取风险(risk)失败: %w", err)
			}
			sSBean.Risk = append(sSBean.Risk, risk)
		}
	}

	// 获取感官观察数据（允许为空，不算错误）
	if looks, err := sS.GetLooks(); err == nil && len(looks) > 0 {
		sSBean.SeeSeekLook = looks[0] // 取第一条记录
	}

	if listens, err := sS.GetListens(); err == nil && len(listens) > 0 {
		sSBean.SeeSeekListen = listens[0] // 取第一条记录
	}

	if smells, err := sS.GetSmells(); err == nil && len(smells) > 0 {
		sSBean.SeeSeekSmell = smells[0] // 取第一条记录
	}

	if touches, err := sS.GetTouches(); err == nil && len(touches) > 0 {
		sSBean.SeeSeekTouch = touches[0] // 取第一条记录
	}

	// 获取检测报告数据（允许为空，不算错误）
	if reports, err := sS.GetExaminationReports(); err == nil {
		sSBean.SeeSeekExaminationReport = reports
	}

	return sSBean, nil
}

// fetchBrainFireBean 获取完整的BrainFireBean
func fetchBrainFireBean(brainFire dao.BrainFire) (dao.BrainFireBean, error) {
	var bean dao.BrainFireBean
	bean.BrainFire = brainFire

	// 获取环境信息
	if brainFire.EnvironmentId > 0 {
		env := dao.Environment{Id: brainFire.EnvironmentId}
		if err := env.GetByIdOrUUID(); err == nil {
			bean.Environment = env
		}
	}

	// 获取项目信息
	project := dao.Project{Id: brainFire.ProjectId}
	if err := project.Get(); err == nil {
		bean.Project = project
	}

	return bean, nil
}

// fetchSuggestionBean 获取完整的SuggestionBean
func fetchSuggestionBean(suggestion dao.Suggestion) (dao.SuggestionBean, error) {
	var bean dao.SuggestionBean
	bean.Suggestion = suggestion

	// 获取项目信息
	project := dao.Project{Id: suggestion.ProjectId}
	if err := project.Get(); err == nil {
		bean.Project = project
	}

	return bean, nil
}

// fetchSkillUserBean 获取完整的SkillUserBean
func fetchSkillUserBean(user dao.User, ctx context.Context) (dao.SkillUserBean, error) {
	var bean dao.SkillUserBean
	bean.User = user

	// 获取用户技能记录
	userSkills, err := dao.GetUserSkills(user.Id, ctx)
	if err != nil {
		return bean, err
	}
	bean.SkillUsers = userSkills

	// 获取对应的技能信息
	skills, err := dao.GetSkillsBySkillUsers(userSkills, ctx)
	if err != nil {
		return bean, err
	}
	bean.Skills = skills

	return bean, nil
}

// fetchMagicUserBean 获取完整的MagicUserBean
func fetchMagicUserBean(user dao.User, ctx context.Context) (dao.MagicUserBean, error) {
	var bean dao.MagicUserBean
	bean.User = user

	// 获取用户法力记录
	userMagics, err := dao.GetUserMagics(user.Id, ctx)
	if err != nil {
		return bean, err
	}
	bean.MagicUsers = userMagics

	// 获取对应的法力信息
	magics, err := dao.GetMagicsByMagicUsers(userMagics, ctx)
	if err != nil {
		return bean, err
	}
	bean.Magics = magics

	return bean, nil
}

// fetchSkillTeamBean 根据团队获取完整的SkillTeamBean
func fetchSkillTeamBean(team dao.Team, ctx context.Context) (dao.SkillTeamBean, error) {
	var bean dao.SkillTeamBean
	bean.Team = team

	// 获取团队技能记录
	teamSkills, err := dao.GetTeamSkills(team.Id, ctx)
	if err != nil {
		return bean, err
	}
	bean.SkillTeams = teamSkills

	// 获取对应的技能信息
	if len(teamSkills) > 0 {
		var skillUsers []dao.SkillUser
		for _, ts := range teamSkills {
			skillUsers = append(skillUsers, dao.SkillUser{SkillId: ts.SkillId})
		}
		skills, err := dao.GetSkillsBySkillUsers(skillUsers, ctx)
		if err != nil {
			return bean, err
		}
		bean.Skills = skills
	}

	return bean, nil
}

// fetchMagicTeamBean 根据团队获取完整的MagicTeamBean
func fetchMagicTeamBean(team dao.Team, ctx context.Context) (dao.MagicTeamBean, error) {
	var bean dao.MagicTeamBean
	bean.Team = team

	// 获取团队法力记录
	teamMagics, err := dao.GetTeamMagics(team.Id, ctx)
	if err != nil {
		return bean, err
	}
	bean.MagicTeams = teamMagics

	// 获取对应的法力信息
	if len(teamMagics) > 0 {
		var magicUsers []dao.MagicUser
		for _, tm := range teamMagics {
			magicUsers = append(magicUsers, dao.MagicUser{MagicId: tm.MagicId})
		}
		magics, err := dao.GetMagicsByMagicUsers(magicUsers, ctx)
		if err != nil {
			return bean, err
		}
		bean.Magics = magics
	}

	return bean, nil
}

// fetchHandicraftBean 根据handicraft，获取完整的HandicraftBean
func fetchHandicraftBean(handicraft dao.Handicraft) (dao.HandicraftBean, error) {
	var bean dao.HandicraftBean
	bean.Handicraft = handicraft
	bean.IsOpen = (handicraft.Category == dao.HandicraftCategoryPublic)

	// 获取项目信息
	project := dao.Project{Id: handicraft.ProjectId}
	if err := project.Get(); err == nil {
		bean.Project = project
	}

	// 获取协助者列表
	if contributors, err := handicraft.GetContributors(); err == nil {
		bean.Contributors = contributors
	}

	// 获取开工仪式（允许为空）
	if inaugurations, err := dao.GetInaugurationsByHandicraftId(handicraft.Id); err == nil && len(inaugurations) > 0 {
		bean.Inauguration = &inaugurations[0]
	}

	// 获取过程记录（允许为空）
	if processRecords, err := dao.GetProcessRecordsByHandicraftId(handicraft.Id); err == nil {
		bean.ProcessRecords = processRecords
	}

	// 获取结束仪式（允许为空）
	if endings, err := dao.GetEndingsByHandicraftId(handicraft.Id); err == nil && len(endings) > 0 {
		bean.Ending = &endings[0]
	}

	return bean, nil
}
