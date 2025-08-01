package route

import (
	"fmt"
	data "teachat/DAO"
)

// 根据给出的projectAppointment参数，去获取对应的projectAppointmentBean资料，然后按结构拼装返回。
func fetchProjectAppointmentBean(pA data.ProjectAppointment) (pABean data.ProjectAppointmentBean, err error) {
	pABean.ProjectAppointment = pA
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
