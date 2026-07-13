package dao

// SearchTeamsByTag 根据标签搜索团队
func SearchTeamsByTag(tag string) ([]Team, error) {
	query := `SELECT id, uuid, name, mission, founder_id, created_at, class, nature,
          abbreviation, logo, tags, updated_at 
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
		var team Team
		err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission,
			&team.FounderId, &team.CreatedAt, &team.Class, &team.Nature, &team.Abbreviation,
			&team.Logo, &team.Tags, &team.UpdatedAt)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teams, rows.Err()
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
