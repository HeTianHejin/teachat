package dao

import (
	"context"
	"fmt"
)

// AreUsersPeers 判断两个用户是否同行（同专业）
// 同行判定规则：
//  1. 只考虑用户所在的职业团队（nature = TeamNatureProfessional），且团队状态正常；
//  2. 不考虑集团关系，只看个人所属团队；
//  3. 若两人职业团队的标签存在交集，则视为同行；
//  4. 业余团队、未知/系统团队、已删除团队不参与同行判断。
func AreUsersPeers(userAId, userBId int) (bool, error) {
	if userAId <= 0 || userBId <= 0 {
		return false, fmt.Errorf("invalid user id: %d, %d", userAId, userBId)
	}
	if userAId == userBId {
		// 同一用户不视为自己的同行
		return false, nil
	}

	ctx := context.Background()

	tagsA, err := getUserProfessionalTags(userAId, ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get professional tags for user %d: %w", userAId, err)
	}
	if len(tagsA) == 0 {
		return false, nil
	}

	tagsB, err := getUserProfessionalTags(userBId, ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get professional tags for user %d: %w", userBId, err)
	}
	if len(tagsB) == 0 {
		return false, nil
	}

	return hasIntersection(tagsA, tagsB), nil
}

// getUserProfessionalTags 获取用户所在职业团队的全部标签集合（去重）
func getUserProfessionalTags(userId int, ctx context.Context) (map[string]struct{}, error) {
	teams, err := GetUserSurvivalTeams(userId, ctx)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]struct{})
	for _, team := range teams {
		if !team.IsPeerCandidate() {
			continue
		}
		for _, tag := range team.GetTags() {
			if tag == "" {
				continue
			}
			tags[tag] = struct{}{}
		}
	}
	return tags, nil
}

// hasIntersection 判断两个集合是否有交集
func hasIntersection(a, b map[string]struct{}) bool {
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if _, ok := b[k]; ok {
			return true
		}
	}
	return false
}

// GetPeerTagsBetweenUsers 获取两个用户之间共同的职业团队标签
func GetPeerTagsBetweenUsers(userAId, userBId int) ([]string, error) {
	if userAId <= 0 || userBId <= 0 {
		return nil, fmt.Errorf("invalid user id: %d, %d", userAId, userBId)
	}

	ctx := context.Background()
	tagsA, err := getUserProfessionalTags(userAId, ctx)
	if err != nil {
		return nil, err
	}
	tagsB, err := getUserProfessionalTags(userBId, ctx)
	if err != nil {
		return nil, err
	}

	common := make([]string, 0)
	for tag := range tagsA {
		if _, ok := tagsB[tag]; ok {
			common = append(common, tag)
		}
	}
	return common, nil
}

// IsTeamEligibleForShortlist 判断团队是否具备入围资格（职业团队）
func IsTeamEligibleForShortlist(teamId int) (bool, error) {
	if teamId == TeamIdNone {
		return false, nil
	}
	team, err := GetTeam(teamId)
	if err != nil {
		return false, err
	}
	return team.IsProfessional() && !team.IsDeleted(), nil
}
