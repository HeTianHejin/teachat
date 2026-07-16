package dao

import (
	"context"
	"fmt"
	"strconv"
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

// 判断两个人是否同事（yes/no）
// 同一职业集团或者职业团队，业余团队不算同事
// 同一团队，则不考虑集团
// 如果不是同事，返回"No"
// 如果是团队同事，返回“Team+id”
// 如果集团同事，返回集团“Group+id”
func IsUsersColleagues(userAId, userBId int, ctx context.Context) (string, error) {
	if userAId <= 0 || userBId <= 0 {
		return "No", fmt.Errorf("invalid user id: %d, %d", userAId, userBId)
	}
	if userAId == userBId {
		return "No", fmt.Errorf("same user id: %d, %d", userAId, userBId)
	}

	//是否在同一团队
	//获取用户A的所有职业团队ID
	teamIdsA, err := GetUserAllProfessionalTeamsId(userAId)
	if err != nil {
		return "No", err
	}
	//获取用户B的所有职业团队ID
	teamIdsB, err := GetUserAllProfessionalTeamsId(userBId)
	if err != nil {
		return "No", err
	}
	//检查是否存在交集
	for _, teamIdA := range teamIdsA {
		if contains(teamIdsB, teamIdA) {
			return "Team+" + strconv.Itoa(teamIdA), nil
		}
	}

	// 检查用户A、B是否在同一职业集团
	// 获取用户A的所有职业集团ID，因为一个团队限制只能加入一个集团，所以用户A的所有团队的集团ID是唯一的，如果所在团队没有加入集团，则不存在集团ID
	// 判断团队是否加入集团,如果有，添加到groupsAId切片，迭代获取全部groupsAid，然后同样获取用户B的所有集团ID，判断是否存在交集
	groupIdsA := []int{len(teamIdsA)}
	for _, teamIdA := range teamIdsA {
		if yes, err := HasEverJoinedGroup(teamIdA); err != nil {
			return "No", err
		} else if yes {
			groupIdA, err := GetGroupIdByTeamId(teamIdA)
			if err != nil {
				return "No", err
			}
			groupIdsA = append(groupIdsA, groupIdA)
		}
	}

	//根据用户teamsIds获取所有职业集团ID
	groupIdsB := []int{len(teamIdsB)}
	for _, teamIdB := range teamIdsB {
		if yes, err := HasEverJoinedGroup(teamIdB); err != nil {
			return "No", err
		} else if yes {
			groupIdB, err := GetGroupIdByTeamId(teamIdB)
			if err != nil {
				return "No", err
			}
			groupIdsB = append(groupIdsB, groupIdB)
		}
	}
	// 检查两个用户是否在同一职业集团
	for _, groupIdA := range groupIdsA {
		if contains(groupIdsB, groupIdA) {
			return "Group+" + strconv.Itoa(groupIdA), nil
		}
	}

	return "No", nil
}
