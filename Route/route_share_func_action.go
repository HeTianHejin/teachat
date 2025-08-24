package route

import (
	"fmt"
	data "teachat/DAO"
)

// 根据给出的projectAppointment参数，去获取对应的projectAppointmentBean资料，然后按结构拼装返回。
func fetchAppointmentBean(pA data.ProjectAppointment) (pABean data.ProjectAppointmentBean, err error) {
	pABean.Appointment = pA
	pr := data.Project{Id: pA.ProjectId}
	if err := pr.Get(); err != nil {
		return pABean, fmt.Errorf("获取茶台(project)失败: %w", err)
	}
	pABean.Project = pr
	payer, err := data.GetUser(pA.PayerUserId)
	if err != nil {
		return pABean, fmt.Errorf("获取payer失败: %w", err)
	}
	pABean.Payer = payer
	payer_family, err := data.GetFamily(pA.PayerFamilyId)
	if err != nil {
		return pABean, fmt.Errorf("获取payer_family失败: %w", err)
	}
	pABean.PayerFamily = payer_family
	payer_team, err := data.GetTeam(pA.PayerTeamId)
	if err != nil {
		return pABean, fmt.Errorf("获取payer_team失败: %w", err)
	}
	pABean.PayerTeam = payer_team
	payee, err := data.GetUser(pA.PayeeUserId)
	if err != nil {
		return pABean, fmt.Errorf("获取payee失败: %w", err)
	}
	pABean.Payee = payee
	payee_family, err := data.GetFamily(pA.PayeeFamilyId)
	if err != nil {
		return pABean, fmt.Errorf("获取payee_family失败: %w", err)
	}
	pABean.PayeeFamily = payee_family
	payee_team, err := data.GetTeam(pA.PayeeTeamId)
	if err != nil {
		return pABean, fmt.Errorf("获取payee_team失败: %w", err)
	}
	pABean.PayeeTeam = payee_team
	verifier, err := data.GetUser(pA.VerifierUserId)
	if err != nil {
		return pABean, fmt.Errorf("获取verifier失败: %w", err)
	}
	pABean.Verifier = verifier
	verifier_family, err := data.GetFamily(pA.VerifierFamilyId)
	if err != nil {
		return pABean, fmt.Errorf("获取verifier_family失败: %w", err)
	}
	pABean.VerifierFamily = verifier_family
	verifier_team, err := data.GetTeam(pA.VerifierTeamId)
	if err != nil {
		return pABean, fmt.Errorf("获取verifier_team失败: %w", err)
	}
	pABean.VerifierTeam = verifier_team
	return pABean, nil
}

// fetchSeeSeekBean() 根据给出的SeeSeek参数，去获取对应的SeeSeekBean资料，然后按结构拼装返回。
func fetchSeeSeekBean(sS data.SeeSeek) (sSBean data.SeeSeekBean, err error) {
	sSBean.SeeSeek = sS
	sSBean.IsOpen = (sS.Category == data.SeeSeekCategoryPublic)

	// 获取项目信息
	pr := data.Project{Id: sS.ProjectId}
	if err := pr.Get(); err != nil {
		return sSBean, fmt.Errorf("获取项目(project)失败: %w", err)
	}
	sSBean.Project = pr

	// 获取地点信息
	if sS.PlaceId > 0 {
		place := data.Place{Id: sS.PlaceId}
		if err := place.Get(); err != nil {
			return sSBean, fmt.Errorf("获取地点(place)失败: %w", err)
		}
		sSBean.Place = place
	}

	// 获取环境信息
	if envs, err := sS.GetEnvironments(); err != nil {
		return sSBean, fmt.Errorf("获取环境关联失败: %w", err)
	} else if len(envs) > 0 {
		env := data.Environment{Id: envs[0].EnvironmentId}
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
			hazard := data.Hazard{Id: h.HazardId}
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
			risk := data.Risk{Id: r.RiskId}
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

	return sSBean, nil
}
