package data

import (
	"time"
)

// Group 集团，n个茶团集合，构成一个大组织。 group = team set
// 用于支持多团队协作的复杂项目场景
type Group struct {
	Id           int
	Uuid         string
	Name         string
	Abbreviation string // 集团简称
	Mission      string // 集团使命/目标
	FounderId    int    // 集团创建者用户ID
	FirstTeamId  int    // 最高管理团队ID
	Class        int    // 集团类型：1-开放式，2-封闭式，10-开放式草集团，20-封闭式草集团
	Logo         string // 集团标志
	Tags         string // 分类标签，逗号分隔，如"诗词书法,文化艺术"
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time // 软删除时间戳，NULL表示未删除
}

// Group 类型常量
const (
	GroupClassOpen               = 1  // 开放式集团
	GroupClassClose              = 2  // 封闭式集团
	GroupClassOpenDraft          = 10 // 开放式草集团
	GroupClassCloseDraft         = 20 // 封闭式草集团
	GroupClassRejectedOpenDraft  = 31 // 已婉拒开放式草集团
	GroupClassRejectedCloseDraft = 32 // 已婉拒封闭式草集团
)

// GroupMemberStatus 集团成员状态类型
type GroupMemberStatus int

const (
	GroupMemberStatusBlacklisted GroupMemberStatus = iota // 黑名单（禁止参与）
	GroupMemberStatusActive                               // 正常（活跃成员）
	GroupMemberStatusSuspended                            // 暂停（临时限制）
	GroupMemberStatusResigned                             // 已退出（主动离开）
	GroupMemberStatusPending                              // 待审核（申请中）
)

// GroupMember 集团成员，1 team = 1 member
// 代表一个团队在集团中的成员资格
type GroupMember struct {
	Id        int
	Uuid      string
	GroupId   int               // 所属集团ID
	TeamId    int               // 团队ID
	Level     int               // 等级：1-最高级，2-次级，3-次次级...
	Role      string            // 角色描述
	Status    GroupMemberStatus // 成员状态
	UserId    int               // 登记操作的用户ID
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time // 软删除时间戳，NULL表示未删除
}

// CreatedAtDate 返回集团创建时间的格式化字符串
func (group *Group) CreatedAtDate() string {
	return group.CreatedAt.Format(FMT_DATE_CN)
}

// IsActive 检查集团是否处于活跃状态（未删除且为正式集团）
func (group *Group) IsActive() bool {
	return !group.IsDeleted() && (group.Class == GroupClassOpen || group.Class == GroupClassClose)
}

// Property 返回集团类型的中文描述
func (group *Group) Property() string {
	switch group.Class {
	case GroupClassOpen:
		return "开放式集团"
	case GroupClassClose:
		return "封闭式集团"
	case GroupClassOpenDraft:
		return "开放式草集团"
	case GroupClassCloseDraft:
		return "封闭式草集团"
	case GroupClassRejectedOpenDraft:
		return "已婉拒开放式草集团"
	case GroupClassRejectedCloseDraft:
		return "已婉拒封闭式草集团"
	default:
		return "未知"
	}
}

// GetStatus 返回集团成员状态的中文描述
func (gm *GroupMember) GetStatus() string {
	switch gm.Status {
	case GroupMemberStatusBlacklisted:
		return "黑名单"
	case GroupMemberStatusActive:
		return "正常"
	case GroupMemberStatusSuspended:
		return "暂停"
	case GroupMemberStatusResigned:
		return "已退出"
	case GroupMemberStatusPending:
		return "待审核"
	default:
		return "未知"
	}
}

// CreatedAtDate 返回集团成员创建时间的格式化字符串
func (gm *GroupMember) CreatedAtDate() string {
	return gm.CreatedAt.Format(FMT_DATE_CN)
}

// IsDeleted 检查集团是否已被软删除
func (group *Group) IsDeleted() bool {
	return group.DeletedAt != nil
}

// SoftDelete 软删除集团
func (group *Group) SoftDelete() error {
	now := time.Now()
	group.DeletedAt = &now
	statement := "UPDATE groups SET deleted_at = $1, updated_at = $2 WHERE id = $3"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(now, now, group.Id)
	return err
}

// Restore 恢复已软删除的集团
func (group *Group) Restore() error {
	group.DeletedAt = nil
	now := time.Now()
	statement := "UPDATE groups SET deleted_at = NULL, updated_at = $1 WHERE id = $2"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(now, group.Id)
	return err
}

// Create 创建集团
func (group *Group) Create() error {
	statement := `INSERT INTO groups (name, abbreviation, mission, founder_id, 
	              first_team_id, class, logo, tags, created_at) 
	              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
	              RETURNING id, uuid, created_at`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(group.Name, group.Abbreviation, group.Mission,
		group.FounderId, group.FirstTeamId, group.Class, group.Logo, group.Tags, time.Now()).Scan(
		&group.Id, &group.Uuid, &group.CreatedAt)
	return err
}

// Get 根据ID获取集团
func (group *Group) Get() error {
	statement := `SELECT id, uuid, name, abbreviation, mission, founder_id, 
	              first_team_id, class, logo, tags, created_at, updated_at, deleted_at 
	              FROM groups WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(group.Id).Scan(&group.Id, &group.Uuid, &group.Name,
		&group.Abbreviation, &group.Mission, &group.FounderId, &group.FirstTeamId,
		&group.Class, &group.Logo, &group.Tags, &group.CreatedAt, &group.UpdatedAt, &group.DeletedAt)
	return err
}

// GetByUUID 根据UUID获取集团
func GetGroupByUUID(uuid string) (Group, error) {
	var group Group
	statement := `SELECT id, uuid, name, abbreviation, mission, founder_id, 
	              first_team_id, class, logo, tags, created_at, updated_at, deleted_at 
	              FROM groups WHERE uuid = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return group, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(uuid).Scan(&group.Id, &group.Uuid, &group.Name,
		&group.Abbreviation, &group.Mission, &group.FounderId, &group.FirstTeamId,
		&group.Class, &group.Logo, &group.Tags, &group.CreatedAt, &group.UpdatedAt, &group.DeletedAt)
	return group, err
}

// Update 更新集团信息
func (group *Group) Update() error {
	now := time.Now()
	group.UpdatedAt = &now
	statement := `UPDATE groups SET name = $1, abbreviation = $2, mission = $3, 
	              first_team_id = $4, class = $5, logo = $6, updated_at = $7 
	              WHERE id = $8`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(group.Name, group.Abbreviation, group.Mission,
		group.FirstTeamId, group.Class, group.Logo, now, group.Id)
	return err
}

// Create 创建集团成员
func (gm *GroupMember) Create() error {
	statement := `INSERT INTO group_members (uuid, group_id, team_id, level, role, 
	              status, user_id, created_at) 
	              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
	              RETURNING id, uuid, created_at`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), gm.GroupId, gm.TeamId, gm.Level,
		gm.Role, gm.Status, gm.UserId, time.Now()).Scan(&gm.Id, &gm.Uuid, &gm.CreatedAt)
	return err
}

// Get 根据ID获取集团成员
func (gm *GroupMember) Get() error {
	statement := `SELECT id, uuid, group_id, team_id, level, role, status, 
	              user_id, created_at, updated_at, deleted_at 
	              FROM group_members WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(gm.Id).Scan(&gm.Id, &gm.Uuid, &gm.GroupId, &gm.TeamId,
		&gm.Level, &gm.Role, &gm.Status, &gm.UserId, &gm.CreatedAt, &gm.UpdatedAt,
		&gm.DeletedAt)
	return err
}

// GetMembersByGroupId 获取集团的所有成员团队
func GetMembersByGroupId(groupId int) ([]GroupMember, error) {
	query := `SELECT id, uuid, group_id, team_id, level, role, status, 
	          user_id, created_at, updated_at, deleted_at 
	          FROM group_members WHERE group_id = $1 AND deleted_at IS NULL 
	          ORDER BY level ASC, created_at ASC`
	rows, err := db.Query(query, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]GroupMember, 0)
	for rows.Next() {
		var gm GroupMember
		if err = rows.Scan(&gm.Id, &gm.Uuid, &gm.GroupId, &gm.TeamId, &gm.Level,
			&gm.Role, &gm.Status, &gm.UserId, &gm.CreatedAt, &gm.UpdatedAt,
			&gm.DeletedAt); err != nil {
			return nil, err
		}
		members = append(members, gm)
	}
	return members, rows.Err()
}

// GetTeamsByGroupId 获取集团的所有团队
func GetTeamsByGroupId(groupId int) ([]Team, error) {
	query := `SELECT t.id, t.uuid, t.name, t.mission, t.founder_id, t.created_at, 
	          t.class, t.abbreviation, t.logo, t.updated_at, t.deleted_at 
	          FROM teams t 
	          INNER JOIN group_members gm ON t.id = gm.team_id 
	          WHERE gm.group_id = $1 AND gm.deleted_at IS NULL AND t.deleted_at IS NULL 
	          ORDER BY gm.level ASC, gm.created_at ASC`
	rows, err := db.Query(query, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]Team, 0)
	for rows.Next() {
		var team Team
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission,
			&team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation,
			&team.Logo, &team.UpdatedAt, &team.DeletedAt); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

// GetGroupsByTeamId 获取团队所属的所有集团
func GetGroupsByTeamId(teamId int) ([]Group, error) {
	query := `SELECT g.id, g.uuid, g.name, g.abbreviation, g.mission, g.founder_id, 
	          g.first_team_id, g.class, g.logo, g.tags, g.created_at, g.updated_at, g.deleted_at 
	          FROM groups g 
	          INNER JOIN group_members gm ON g.id = gm.group_id 
	          WHERE gm.team_id = $1 AND gm.deleted_at IS NULL AND g.deleted_at IS NULL 
	          ORDER BY gm.created_at DESC`
	rows, err := db.Query(query, teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]Group, 0)
	for rows.Next() {
		var group Group
		if err = rows.Scan(&group.Id, &group.Uuid, &group.Name, &group.Abbreviation,
			&group.Mission, &group.FounderId, &group.FirstTeamId, &group.Class,
			&group.Logo, &group.Tags, &group.CreatedAt, &group.UpdatedAt, &group.DeletedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

// CountGroupMembers 统计集团成员数量
func (group *Group) CountGroupMembers() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM group_members 
	          WHERE group_id = $1 AND deleted_at IS NULL AND status = $2`
	err := db.QueryRow(query, group.Id, GroupMemberStatusActive).Scan(&count)
	return count, err
}

// IsFounder 检查用户是否为集团创建者
func (group *Group) IsFounder(userId int) bool {
	return group.FounderId == userId
}

// IsFirstTeamMember 检查用户是否为最高管理团队成员
func (group *Group) IsFirstTeamMember(userId int) (bool, error) {
	if group.FirstTeamId == 0 {
		return false, nil
	}

	var count int
	query := `SELECT COUNT(*) FROM team_members 
	          WHERE team_id = $1 AND user_id = $2 AND status = $3`
	err := db.QueryRow(query, group.FirstTeamId, userId, TeMemberStatusActive).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CanManage 检查用户是否有管理集团的权限（创建者或最高管理团队成员）
func (group *Group) CanManage(userId int) (bool, error) {
	// 创建者有权限
	if group.IsFounder(userId) {
		return true, nil
	}

	// 最高管理团队成员有权限
	return group.IsFirstTeamMember(userId)
}

// CanAddTeam 检查用户是否有添加团队的权限
func (group *Group) CanAddTeam(userId int) (bool, error) {
	return group.CanManage(userId)
}

// CanRemoveTeam 检查用户是否有移除团队的权限
func (group *Group) CanRemoveTeam(userId int) (bool, error) {
	return group.CanManage(userId)
}

// CanEdit 检查用户是否有编辑集团信息的权限
func (group *Group) CanEdit(userId int) (bool, error) {
	return group.CanManage(userId)
}

// CanDelete 检查用户是否有删除集团的权限（仅创建者）
func (group *Group) CanDelete(userId int) bool {
	return group.IsFounder(userId)
}

// GetGroupByTeamId 获取团队所属的第一个集团（如果有）
func GetGroupByTeamId(teamId int) (*Group, error) {
	query := `SELECT g.id, g.uuid, g.name, g.abbreviation, g.mission, g.founder_id, 
	          g.first_team_id, g.class, g.logo, g.created_at, g.updated_at, g.deleted_at 
	          FROM groups g 
	          INNER JOIN group_members gm ON g.id = gm.group_id 
	          WHERE gm.team_id = $1 AND gm.deleted_at IS NULL AND g.deleted_at IS NULL 
	          ORDER BY gm.created_at ASC LIMIT 1`

	var group Group
	err := db.QueryRow(query, teamId).Scan(
		&group.Id, &group.Uuid, &group.Name, &group.Abbreviation,
		&group.Mission, &group.FounderId, &group.FirstTeamId, &group.Class,
		&group.Logo, &group.CreatedAt, &group.UpdatedAt, &group.DeletedAt)

	if err != nil {
		return nil, err
	}
	return &group, nil
}
