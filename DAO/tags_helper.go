package dao

import (
	"strings"
	"time"
)

// GetTags 获取团队标签数组
func (team *Team) GetTags() []string {
	if team.Tags == "" {
		return []string{}
	}
	tags := strings.Split(team.Tags, ",")
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// SetTags 设置团队标签
func (team *Team) SetTags(tags []string) {
	team.Tags = strings.Join(tags, ",")
}

// HasTag 检查团队是否包含某个标签
func (team *Team) HasTag(tag string) bool {
	tags := team.GetTags()
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// UpdateTags 更新团队标签
func (team *Team) UpdateTags() error {
	statement := `UPDATE teams SET tags = $1, updated_at = $2 WHERE id = $3`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Tags, time.Now(), team.Id)
	return err
}

// GetTags 获取集团标签数组
func (group *Group) GetTags() []string {
	if group.Tags == "" {
		return []string{}
	}
	tags := strings.Split(group.Tags, ",")
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// SetTags 设置集团标签
func (group *Group) SetTags(tags []string) {
	group.Tags = strings.Join(tags, ",")
}

// HasTag 检查集团是否包含某个标签
func (group *Group) HasTag(tag string) bool {
	tags := group.GetTags()
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// UpdateTags 更新集团标签
func (group *Group) UpdateTags() error {
	statement := `UPDATE groups SET tags = $1, updated_at = $2 WHERE id = $3`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(group.Tags, time.Now(), group.Id)
	return err
}

// SearchTeamsByTag 根据标签搜索团队
func SearchTeamsByTag(tag string) ([]Team, error) {
	query := `SELECT id, uuid, name, mission, founder_id, created_at, class, 
	          abbreviation, logo, tags, updated_at 
	          FROM teams 
	          WHERE tags LIKE $1 AND deleted_at IS NULL 
	          ORDER BY created_at DESC LIMIT 50`

	rows, err := db.Query(query, "%"+tag+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]Team, 0)
	for rows.Next() {
		var team Team
		err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission,
			&team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation,
			&team.Logo, &team.Tags, &team.UpdatedAt)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

// SearchGroupsByTag 根据标签搜索集团
func SearchGroupsByTag(tag string) ([]Group, error) {
	query := `SELECT id, uuid, name, abbreviation, mission, founder_id, 
	          first_team_id, class, logo, tags, created_at, updated_at 
	          FROM groups 
	          WHERE tags LIKE $1 AND deleted_at IS NULL 
	          ORDER BY created_at DESC LIMIT 50`

	rows, err := db.Query(query, "%"+tag+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]Group, 0)
	for rows.Next() {
		var group Group
		err = rows.Scan(&group.Id, &group.Uuid, &group.Name, &group.Abbreviation,
			&group.Mission, &group.FounderId, &group.FirstTeamId, &group.Class,
			&group.Logo, &group.Tags, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}
