package dao

import (
	"database/sql"
	"fmt"
)

// SearchTeamsByTag 按自由标签模糊搜索团队
// 用于业余团队：tags 为用户自由填写的逗号分隔文本
func SearchTeamsByTag(tag string) ([]Team, error) {
	query := `SELECT ` + teamSelectColumns + `
	          FROM teams
	          WHERE tags LIKE $1 AND deleted_at IS NULL
	          ORDER BY created_at DESC LIMIT 50`

	rows, err := DB.Query(query, "%"+tag+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]Team, 0)
	for rows.Next() {
		team, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teams, nil
}

// SearchTeamsByIndustryCode 按行业代码前缀搜索职业团队（支持上卷）
//
// 搜 "C"   命中门类 C 下的全部职业团队；
// 搜 "C13" 命中大类 C13 及其下中类/小类的全部职业团队。
// 依据 teams.industry_path（如 "C/C13"）做前缀匹配，配合 idx_teams_industry_path 索引。
func SearchTeamsByIndustryCode(code string, limit int) ([]Team, error) {
	if limit <= 0 {
		limit = 50
	}

	tag, err := GetIndustryTagByCode(code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("行业代码 %s 不存在", code)
		}
		return nil, err
	}

	// 前缀 = 该节点自身的代码路径，从而命中自身及其全部后代
	prefix, err := GetIndustryTagPath(tag.Id)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + teamSelectColumns + `
	          FROM teams
	          WHERE deleted_at IS NULL
	            AND (industry_path = $1 OR industry_path LIKE $1 || '/%')
	          ORDER BY created_at DESC LIMIT $2`

	rows, err := DB.Query(query, prefix, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]Team, 0)
	for rows.Next() {
		team, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teams, nil
}

// SearchGroupsByTag 根据标签搜索集团
func SearchGroupsByTag(tag string) ([]Group, error) {
	query := `SELECT id, uuid, name, abbreviation, mission, founder_id, 
          first_team_id, class, nature, logo, tags, created_at, updated_at 
          FROM groups 
          WHERE tags LIKE $1 AND deleted_at IS NULL 
          ORDER BY created_at DESC LIMIT 50`

	rows, err := DB.Query(query, "%"+tag+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]Group, 0)
	for rows.Next() {
		var group Group
		err = rows.Scan(&group.Id, &group.Uuid, &group.Name, &group.Abbreviation,
			&group.Mission, &group.FounderId, &group.FirstTeamId, &group.Class, &group.Nature,
			&group.Logo, &group.Tags, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, rows.Err()
}
