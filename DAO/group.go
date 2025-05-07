package data

import (
	"time"
)

// Group 集团，n个茶团集合，构成一个大组织。 group = team set
// 作为一个例子，team可以看成某个服务班组（一个机舱的全部空乘服务人员），group可以看成公司的服务部门（包括地勤清洁、保养和补给等等的班组）
type Group struct {
	Id           int
	Uuid         string
	Name         string
	Mission      string
	FounderId    int
	FirstTeamId  int // 第一团队，部门最高管理层团队。董事会？
	CreatedAt    time.Time
	Class        int    // 1: "开放式集团",2: "封闭式集团",10: "开放式草集团",20: "封闭式草集团"
	Abbreviation string // 集团简称
	Logo         string // 集团标志
	UpdatedAt    *time.Time
	MinistryId   int //（预留）默认值 0，上级单位（协会/部委）id
}

// Group.CreatedAtDate()
func (group *Group) CreatedAtDate() string {
	return group.CreatedAt.Format(FMT_DATE_CN)
}

// Group.Property()
func (group *Group) Property() string {
	switch group.Class {
	case 1:
		return "开放式集团"
	case 2:
		return "封闭式集团"
	case 10:
		return "开放式草集团"
	case 20:
		return "封闭式草集团"
	default:
		return "未知"
	}
}

// Create() 根据Group{},在数据库中创建1新的group集团记录
func (group *Group) Create() (err error) {
	err = Db.QueryRow(
		"INSERT INTO groups(uuid,name,mission,founder_id,first_team_id,created_at,class,abbreviation,logo,ministry_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id",
		group.Uuid, group.Name, group.Mission, group.FounderId, group.FirstTeamId, time.Now(), group.Class, group.Abbreviation, group.Logo, group.MinistryId).Scan(&group.Id)
	return err
}

// Delete()
func (group *Group) Delete() (err error) {
	_, err = Db.Exec("DELETE FROM groups WHERE id = $1", group.Id)
	return err
}

// Update()
func (group *Group) Update() (err error) {
	_, err = Db.Exec("UPDATE groups SET uuid = $2, name = $3, mission = $4, founder_id = $5, first_team_id = $6, updated_at = $7, class = $8, abbreviation = $9, logo = $10, ministry_id = $11 WHERE id = $1",
		group.Id, group.Uuid, group.Name, group.Mission, group.FounderId, group.FirstTeamId, time.Now(), group.Class, group.Abbreviation, group.Logo, group.MinistryId)
	return err
}

// GetGroup() 据id,从数据库中获取 1集团记录
func GetGroup(id int) (group Group, err error) {
	group = Group{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, first_team_id, created_at, class, abbreviation, logo, updated_at, ministry_id FROM groups WHERE id = $1", id).
		Scan(&group.Id, &group.Uuid, &group.Name, &group.Mission, &group.FounderId, &group.FirstTeamId, &group.CreatedAt, &group.Class, &group.Abbreviation, &group.Logo, &group.UpdatedAt, &group.MinistryId)
	return group, err
}

// GetGroups() 从数据库中获取所有集团记录
func GetGroups() (groups []Group, err error) {
	rows, err := Db.Query("SELECT id, uuid, name, mission, founder_id, first_team_id, created_at, class, abbreviation, logo, updated_at, ministry_id FROM groups")
	if err != nil {
		return groups, err
	}
	for rows.Next() {
		group := Group{}
		err = rows.Scan(&group.Id, &group.Uuid, &group.Name, &group.Mission, &group.FounderId, &group.FirstTeamId, &group.CreatedAt, &group.Class, &group.Abbreviation, &group.Logo, &group.UpdatedAt, &group.MinistryId)
		if err != nil {
			return groups, err
		}
		groups = append(groups, group)
	}
	rows.Close()
	return groups, err
}

// GetGroupByUuid()
func GetGroupByUuid(uuid string) (group Group, err error) {
	group = Group{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, first_team_id, created_at, class, abbreviation, logo, updated_at, ministry_id FROM groups WHERE uuid = $1", uuid).
		Scan(&group.Id, &group.Uuid, &group.Name, &group.Mission, &group.FounderId, &group.FirstTeamId, &group.CreatedAt, &group.Class, &group.Abbreviation, &group.Logo, &group.UpdatedAt, &group.MinistryId)
	return group, err
}

// GetGroupByFounderId()
func GetGroupByFounderId(founderId int) (group Group, err error) {
	group = Group{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, first_team_id, created_at, class, abbreviation, logo, updated_at, ministry_id FROM groups WHERE founder_id = $1", founderId).
		Scan(&group.Id, &group.Uuid, &group.Name, &group.Mission, &group.FounderId, &group.FirstTeamId, &group.CreatedAt, &group.Class, &group.Abbreviation, &group.Logo, &group.UpdatedAt, &group.MinistryId)
	return group, err
}

// GetGroupByFirstTeamId() 据first_team_id,从数据库中获取group记录
func GetGroupByFirstTeamId(firstTeamId int) (group Group, err error) {
	group = Group{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, first_team_id, created_at, class, abbreviation, logo, updated_at, ministry_id FROM groups WHERE first_team_id = $1", firstTeamId).
		Scan(&group.Id, &group.Uuid, &group.Name, &group.Mission, &group.FounderId, &group.FirstTeamId, &group.CreatedAt, &group.Class, &group.Abbreviation, &group.Logo, &group.UpdatedAt, &group.MinistryId)
	return group, err
}

// GetGroupByName() 据name,从数据库中获取group记录
func GetGroupByName(name string) (group Group, err error) {
	group = Group{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, first_team_id, created_at, class, abbreviation, logo, updated_at, ministry_id FROM groups WHERE name = $1", name).
		Scan(&group.Id, &group.Uuid, &group.Name, &group.Mission, &group.FounderId, &group.FirstTeamId, &group.CreatedAt, &group.Class, &group.Abbreviation, &group.Logo, &group.UpdatedAt, &group.MinistryId)
	return group, err
}

// GetGroupByAbbreviation() 据abbreviation,从数据库中获取group记录
func GetGroupByAbbreviation(abbreviation string) (group Group, err error) {
	group = Group{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, first_team_id, created_at, class, abbreviation, logo, updated_at, ministry_id FROM groups WHERE abbreviation = $1", abbreviation).
		Scan(&group.Id, &group.Uuid, &group.Name, &group.Mission, &group.FounderId, &group.FirstTeamId, &group.CreatedAt, &group.Class, &group.Abbreviation, &group.Logo, &group.UpdatedAt, &group.MinistryId)
	return group, err
}

// GetTeamsCountByGroupId()
func GetTeamsCountByGroupId(group_id int) (count int) {
	rows, _ := Db.Query("SELECT COUNT(*) FROM teams WHERE group_id = $1", group_id)
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// team.Group() 获取team所在的group
func (team *Team) Group() (group Group, err error) {
	// fetch group
	group, err = GetGroup(team.SuperiorTeamId)
	return group, err
}

// group.NumMembers() 集团的总人员计数，下属的全部茶团人数累计
func (group *Group) NumMembers() (count int) {
	query := `
			SELECT SUM(team_member_count) AS total_members
			FROM (
				SELECT t.id, COUNT(tm.id) AS team_member_count
				FROM teams t
				LEFT JOIN team_members tm ON t.id = tm.team_id
				WHERE t.group_id = $1
				GROUP BY t.id
			) AS team_counts;
		`

	err := Db.QueryRow(query, group.Id).Scan(&count)
	if err != nil {
		// Handle error
		return
	}

	return count

	// first, fetch []team_id
	// rows, err := Db.Query("SELECT id FROM teams WHERE group_id = $1", group.Id)
	// teams := []int{}
	// // error handling
	// if err != nil {
	// 	return
	// }
	// // fetch rows
	// for rows.Next() {
	// 	var team_id int
	// 	if err := rows.Scan(&team_id); err != nil {
	// 		return
	// 	}
	// 	teams = append(teams, team_id)
	// }
	// // close rows
	// rows.Close()
	// // count

	// var tCount int
	// for _, team_id := range teams {
	// 	// fetch team members count
	// 	rows, _ := Db.Query("SELECT COUNT(*) FROM team_members WHERE team_id = $1", team_id)
	// 	for rows.Next() {
	// 		if err := rows.Scan(&tCount); err != nil {
	// 			return
	// 		}
	// 	}
	// 	count += tCount
	// 	rows.Close()
	// }

	// return count
}
