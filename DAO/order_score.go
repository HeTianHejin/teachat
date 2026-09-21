package dao

import (
	"errors"
	"sort"
)

// TaskScoreInput 是订单中一个可独立验收任务的结算输入。
type TaskScoreInput struct {
	HandicraftId int
	TaskWeight   int
	TaskScore    int
	Participants []ParticipantScoreInput
}

// ParticipantScoreInput 描述成员在单个任务上的事实责任与实际贡献。
type ParticipantScoreInput struct {
	UserId               int
	ResponsibilityWeight int
	ActualContribution   int
	PerformanceScore     int
	EvidenceScore        int
}

// CalculateOrderScores 将任务成绩按任务权重和成员有效贡献归因到个人。
// 团队总分是任务成绩的加权平均；个人贡献分之和等于团队总分。
func CalculateOrderScores(tasks []TaskScoreInput) (float64, []MemberOrderScore, error) {
	if len(tasks) == 0 {
		return 0, nil, errors.New("订单至少需要一个已评分任务")
	}

	totalTaskWeight := 0
	weightedTeamScore := 0
	byUser := make(map[int]*MemberOrderScore)
	for _, task := range tasks {
		if task.TaskWeight <= 0 || task.TaskScore < 0 || task.TaskScore > 100 {
			return 0, nil, errors.New("任务权重或任务分数无效")
		}
		if len(task.Participants) == 0 {
			return 0, nil, errors.New("已评分任务必须有参与成员")
		}
		totalTaskWeight += task.TaskWeight
		weightedTeamScore += task.TaskWeight * task.TaskScore

		totalEffectiveWeight := 0
		for _, participant := range task.Participants {
			if participant.UserId <= 0 || participant.ResponsibilityWeight <= 0 || participant.ActualContribution < 0 || participant.ActualContribution > 100 {
				return 0, nil, errors.New("成员责任权重或实际贡献无效")
			}
			if participant.PerformanceScore < 0 || participant.PerformanceScore > 100 || participant.EvidenceScore < 0 || participant.EvidenceScore > 100 {
				return 0, nil, errors.New("成员表现分或证据分无效")
			}
			totalEffectiveWeight += participant.ResponsibilityWeight * participant.ActualContribution
		}
		if totalEffectiveWeight == 0 {
			return 0, nil, errors.New("任务没有有效贡献")
		}

		for _, participant := range task.Participants {
			share := float64(participant.ResponsibilityWeight*participant.ActualContribution) / float64(totalEffectiveWeight)
			result := byUser[participant.UserId]
			if result == nil {
				result = &MemberOrderScore{UserId: participant.UserId, CalculationVersion: "v1"}
				byUser[participant.UserId] = result
			}
			result.ContributionScore += float64(task.TaskWeight*task.TaskScore) * share / 100.0
			result.PerformanceScore += float64(task.TaskWeight*participant.PerformanceScore) * share / 100.0
			result.ResponsibilityScore += float64(task.TaskWeight*participant.ResponsibilityWeight) * share / 100.0
			result.EvidenceScore += float64(task.TaskWeight*participant.EvidenceScore) * share / 100.0
		}
	}

	results := make([]MemberOrderScore, 0, len(byUser))
	for _, result := range byUser {
		result.ContributionScore = result.ContributionScore * 100 / float64(totalTaskWeight)
		result.PerformanceScore = result.PerformanceScore * 100 / float64(totalTaskWeight)
		result.ResponsibilityScore = result.ResponsibilityScore * 100 / float64(totalTaskWeight)
		result.EvidenceScore = result.EvidenceScore * 100 / float64(totalTaskWeight)
		results = append(results, *result)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].UserId < results[j].UserId })
	return float64(weightedTeamScore) / float64(totalTaskWeight), results, nil
}
