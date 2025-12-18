package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	util "teachat/Util"
	"time"
)

// $事业茶团角色
const (
	RoleCEO    = "CEO"
	RoleCTO    = "CTO"
	RoleCMO    = "CMO"
	RoleCFO    = "CFO"
	RoleTaster = "taster"
)
const (
	TeamIdNone          = 0 // 0
	TeamIdSpaceshipCrew = 1 // 1   飞船茶棚团队，系统保留
	TeamIdFreelancer    = 2 // 2  系统预设“自由人”$事业茶团团队，系统保留，这是所有注册用户的默认$事业茶团，没有茶叶资产

	TeamIdVerifier = 3 // 3 见证者团队，系统保留

)

var (
	TeamUUIDSpaceshipCrew = "dcbe3046-b192-44b6-7afb-bc55817c13a9"
	TeamUUIDFreelancer    = "72c06442-2b60-418a-6493-a91bd03ae4k8"
	TeamUUIDVerifier      = "38be3046-b192-44b6-7afb-bc55817c13c4"
)

// 从数据库查询获取团队
func GetTeam(teamID int) (Team, error) {
	const query = `SELECT id, uuid, name, mission, founder_id, 
                  created_at, class, abbreviation, logo, is_private, updated_at, deleted_at, tags 
                  FROM teams WHERE id = $1`

	var team Team
	err := DB.QueryRow(query, teamID).Scan(
		&team.Id, &team.Uuid, &team.Name, &team.Mission,
		&team.FounderId, &team.CreatedAt, &team.Class,
		&team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.DeletedAt, &team.Tags,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Team{}, fmt.Errorf("team not found with id: %d", teamID)
		}
		return Team{}, fmt.Errorf("failed to query team: %w", err)
	}

	return team, nil
}

// GetTeamsByIds 根据$事业茶团team_id_slice，获取全部申请加盟的$事业茶团[]team
func GetTeamsByIds(teamIDs []int) ([]Team, error) {
	if len(teamIDs) == 0 {
		return []Team{}, nil
	}

	query := `SELECT id, uuid, name, mission, founder_id, created_at, class, 
	          abbreviation, logo, is_private, updated_at, deleted_at, tags 
	          FROM teams WHERE id = ANY($1) AND deleted_at IS NULL`

	rows, err := DB.Query(query, teamIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]Team, 0, len(teamIDs))
	for rows.Next() {
		var team Team
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission,
			&team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation,
			&team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.DeletedAt, &team.Tags); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// Team $事业茶团=同事团队
// 拥有共同目标，或者兴趣爱好/信仰/利益的成员间非血缘关系团队
// 预算上来说team的人数上限是12人，1 dozen
// 层级关系通过Group结构管理
type Team struct {
	Id           int
	Uuid         string
	Name         string
	Mission      string // 团队使命/目标
	FounderId    int    // 团队发起人茶友id
	Class        int    // 团队类型：0-系统团队，1-开放式，2-封闭式，10-开放式草团，20-封闭式草团，31-已婉拒开放式，32-已婉拒封闭式
	Abbreviation string // 队名简称
	Logo         string // $事业茶团标志
	Tags         string // 分类标签，逗号分隔，如"诗词书法,文化艺术"
	IsPrivate    bool   // 是否私密：true-私密团队（不公开但可接收通知），false-公开团队
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time // 软删除时间戳，NULL表示未删除
}

const (
	TeamClassSpaceship          = 0  //飞船茶棚团队，系统保留
	TeamClassOpen               = 1  //开放式$事业茶团
	TeamClassClose              = 2  // 封闭式$事业茶团
	TeamClassOpenDraft          = 10 //开放式草团
	TeamClassCloseDraft         = 20 // 封闭式草团
	TeamClassRejectedOpenDraft  = 31 //已婉拒开放式草团
	TeamClassRejectedCloseDraft = 32 // 已婉拒封闭式草团
)

// TeamMemberStatus 团队成员状态类型
type TeamMemberStatus int

const (
	TeMemberStatusBlacklist TeamMemberStatus = 0 // 黑名单（禁止参与）
	TeMemberStatusActive    TeamMemberStatus = 1 // 正常（活跃成员）
	TeMemberStatusSuspended TeamMemberStatus = 2 // 暂停（临时限制）
	TeMemberStatusResigned  TeamMemberStatus = 3 // 已退出（主动离开）
	TeMemberStatusPending   TeamMemberStatus = 4 // 待审核（申请中）
)

// TeamMember 团队成员=当前$事业茶团加入成员记录
type TeamMember struct {
	Id        int
	Uuid      string
	TeamId    int
	UserId    int
	Role      string           // 角色，职务：CEO，CTO，CMO，CFO，taster
	Status    TeamMemberStatus // 成员状态
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time // 软删除时间戳，NULL表示未删除
}

// GetStatus 返回团队成员状态的中文描述
func (member *TeamMember) GetStatus() string {
	switch member.Status {
	case TeMemberStatusBlacklist:
		return "黑名单"
	case TeMemberStatusActive:
		return "正常品茶"
	case TeMemberStatusSuspended:
		return "暂停品茶"
	case TeMemberStatusResigned:
		return "退出茶团"
	case TeMemberStatusPending:
		return "待审核"
	default:
		return "未知"
	}
}

// IsDeleted 检查团队是否已被软删除
func (team *Team) IsDeleted() bool {
	return team.DeletedAt != nil
}

func (team *Team) IsOpen() bool {
	return team.Class == TeamClassOpen || team.Class == TeamClassOpenDraft
}

// SoftDelete 软删除团队
func (team *Team) SoftDelete() error {
	now := time.Now()
	team.DeletedAt = &now
	statement := "UPDATE teams SET deleted_at = $1, updated_at = $2 WHERE id = $3"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(now, now, team.Id)
	return err
}

// Restore 恢复已软删除的团队
func (team *Team) Restore() error {
	team.DeletedAt = nil
	now := time.Now()
	statement := "UPDATE teams SET deleted_at = NULL, updated_at = $1 WHERE id = $2"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(now, team.Id)
	return err
}

// IsActive 检查团队是否处于活跃状态（未删除且为正式团队）
func (team *Team) IsActive() bool {
	return !team.IsDeleted() && (team.Class == TeamClassOpen || team.Class == TeamClassClose)
}

// 成员“退出$事业茶团声明书”（相当于辞职信？）
type TeamMemberResignation struct {
	Id               int
	Uuid             string
	TeamId           int //“声明退出$事业茶团”所指向的$事业茶团id
	CeoUserId        int //时任$事业茶团CEO茶友id，
	CoreMemberUserId int //时任核心成员茶友id，要求双确认，如果有核心成员，也要同意退出

	MemberId          int    //声明退出$事业茶团的成员id(team_member.id)
	MemberUserId      int    //声明退出$事业茶团的茶友id
	MemberCurrentRole string //时任角色

	Title     string //标题
	Content   string //内容
	Status    int    //声明状态： 0、未读，1、已读，2、已核对，3、已批准，4、挽留中(未批准)，5、强行退出
	CreatedAt time.Time
	UpdatedAt *time.Time
}

const (
	ResignationStatusUnread          = iota // 0 未阅读
	ResignationStatusRead                   // 1 已阅读
	ResignationStatusCoreMemberAgree        // 2 核心成员同意
	ResignationStatusApproved               // 3 CEO批准
	ResignationStatusPending                // 4 挽留中
	ResignationStatusForced                 // 5 强行退出
)

// TeamMemberResignation.GetStatus()
func (resignation *TeamMemberResignation) GetStatus() string {
	switch resignation.Status {
	case 0:
		return "未阅读"
	case 1:
		return "已阅读"
	case 2:
		return "核心成员同意"
	case 3:
		return "CEO批准"
	case 4:
		return "挽留中"
	case 5:
		return "强行退出"
	}
	return ""
}

// TeamMemberResignation.Create()
func (resignation *TeamMemberResignation) Create() (err error) {
	statement := "INSERT INTO team_member_resignations (uuid, team_id, ceo_user_id, core_member_user_id, member_id, member_user_id, member_current_role, title, content, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), resignation.TeamId, resignation.CeoUserId, resignation.CoreMemberUserId, resignation.MemberId, resignation.MemberUserId, resignation.MemberCurrentRole, resignation.Title, resignation.Content, resignation.Status, time.Now()).Scan(&resignation.Id, &resignation.Uuid)
	return
}

// TeamMemberResignation.CreatedAtDate()
func (resignation *TeamMemberResignation) CreatedAtDate() string {
	return resignation.CreatedAt.Format("2006-01-02")
}

// TeamMemberResignation.Get()
func (resignation *TeamMemberResignation) Get() (err error) {
	statement := "SELECT * FROM team_member_resignations WHERE id = $1"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(resignation.Id).Scan(&resignation.Id, &resignation.Uuid, &resignation.TeamId, &resignation.CeoUserId, &resignation.CoreMemberUserId, &resignation.MemberId, &resignation.MemberUserId, &resignation.MemberCurrentRole, &resignation.Title, &resignation.Content, &resignation.Status, &resignation.CreatedAt, &resignation.UpdatedAt)
	return
}

// TeamMemberResignation.GetByUuid()
func (resignation *TeamMemberResignation) GetByUuid() (err error) {
	statement := "SELECT * FROM team_member_resignations WHERE uuid = $1"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(resignation.Uuid).Scan(&resignation.Id, &resignation.Uuid, &resignation.TeamId, &resignation.CeoUserId, &resignation.CoreMemberUserId, &resignation.MemberId, &resignation.MemberUserId, &resignation.MemberCurrentRole, &resignation.Title, &resignation.Content, &resignation.Status, &resignation.CreatedAt, &resignation.UpdatedAt)
	return
}

// TeamMemberResignation.UpdateCeoUserIdCoreMemberUserIdStatus()
func (resignation *TeamMemberResignation) UpdateCeoUserIdCoreMemberUserIdStatus() (err error) {
	statement := "UPDATE team_member_resignations SET ceo_user_id = $1, core_member_user_id = $2, status = $3, updated_at = $4 WHERE id = $5"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(resignation.CeoUserId, resignation.CoreMemberUserId, resignation.Status, resignation.UpdatedAt, resignation.Id)
	return
}

// TeamMemberResignations.GetByUserIdAndTeamId()  获取某个用户在某个$事业茶团的全部“退出$事业茶团声明书”
func GetResignationsByUserIdAndTeamId(user_id, team_id int) (resignations []TeamMemberResignation, err error) {
	rows, err := DB.Query("SELECT * FROM team_member_resignations WHERE member_user_id = $1 AND team_id = $2", user_id, team_id)
	if err != nil {
		return
	}
	for rows.Next() {
		resignation := TeamMemberResignation{}
		if err = rows.Scan(&resignation.Id, &resignation.Uuid, &resignation.TeamId, &resignation.CeoUserId, &resignation.CoreMemberUserId, &resignation.MemberId, &resignation.MemberUserId, &resignation.MemberCurrentRole, &resignation.Title, &resignation.Content, &resignation.Status, &resignation.CreatedAt, &resignation.UpdatedAt); err != nil {
			return
		}
		resignations = append(resignations, resignation)
	}
	rows.Close()
	return
}

// TeamMemberResignations.GetByTeamId() 获取某个$事业茶团的全部“退出$事业茶团声明书”
func GetResignationsByTeamId(team_id int) (resignations []TeamMemberResignation, err error) {
	rows, err := DB.Query("SELECT * FROM team_member_resignations WHERE team_id = $1 ORDER BY created_at DESC", team_id)
	if err != nil {
		return
	}
	for rows.Next() {
		resignation := TeamMemberResignation{}
		if err = rows.Scan(&resignation.Id, &resignation.Uuid, &resignation.TeamId, &resignation.CeoUserId, &resignation.CoreMemberUserId, &resignation.MemberId, &resignation.MemberUserId, &resignation.MemberCurrentRole, &resignation.Title, &resignation.Content, &resignation.Status, &resignation.CreatedAt, &resignation.UpdatedAt); err != nil {
			return
		}
		resignations = append(resignations, resignation)
	}
	rows.Close()
	return
}

// GetResignationsByUserId() 获取某个用户的全部"退出$事业茶团声明书"
func GetResignationsByUserId(user_id int) (resignations []TeamMemberResignation, err error) {
	rows, err := DB.Query("SELECT * FROM team_member_resignations WHERE member_user_id = $1 ORDER BY created_at DESC", user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		resignation := TeamMemberResignation{}
		if err = rows.Scan(&resignation.Id, &resignation.Uuid, &resignation.TeamId, &resignation.CeoUserId, &resignation.CoreMemberUserId, &resignation.MemberId, &resignation.MemberUserId, &resignation.MemberCurrentRole, &resignation.Title, &resignation.Content, &resignation.Status, &resignation.CreatedAt, &resignation.UpdatedAt); err != nil {
			return
		}
		resignations = append(resignations, resignation)
	}
	rows.Close()
	return
}

// 用户的“默认$事业茶团”设置记录
type UserDefaultTeam struct {
	Id        int
	UserId    int
	TeamId    int
	CreatedAt time.Time
}

// $事业茶团成员角色变动声明
type TeamMemberRoleNotice struct {
	Id                int
	Uuid              string
	TeamId            int    //声明$事业茶团
	CeoId             int    //时任$事业茶团CEO茶友id
	MemberId          int    //成员id(team_member.id)
	MemberCurrentRole string //当前角色
	NewRole           string //新角色
	Title             string //标题
	Content           string //内容
	Status            int    //声明状态 0:未读,1:已读,2:已处理
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}

// TeamMemberRoleNotice.CreatedAtDate()
func (notice *TeamMemberRoleNotice) CreatedAtDate() string {
	return notice.CreatedAt.Format("2006-01-02")
}

// TeamMemberRoleNotice.Create()
func (notice *TeamMemberRoleNotice) Create() (err error) {
	statement := "INSERT INTO team_member_role_notices (uuid, team_id, ceo_id, member_id, member_current_role, new_role, title, content, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), notice.TeamId, notice.CeoId, notice.MemberId, notice.MemberCurrentRole, notice.NewRole, notice.Title, notice.Content, notice.Status, time.Now()).Scan(&notice.Id, &notice.Uuid)
	return
}

// TeamMemberRoleNotice.Get()
func (notice *TeamMemberRoleNotice) Get() (err error) {
	statement := "SELECT * FROM team_member_role_notices WHERE id = $1"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(notice.Id).Scan(&notice.Id, &notice.Uuid, &notice.TeamId, &notice.CeoId, &notice.MemberId, &notice.MemberCurrentRole, &notice.NewRole, &notice.Title, &notice.Content, &notice.Status, &notice.CreatedAt, &notice.UpdatedAt)
	return
}

// TeamMemberRoleNotice.GetByTeamId()
func GetMemberRoleNoticesByTeamId(team_id int) (notices []TeamMemberRoleNotice, err error) {
	rows, err := DB.Query("SELECT * FROM team_member_role_notices WHERE team_id = $1", team_id)
	if err != nil {
		return
	}
	for rows.Next() {
		notice := TeamMemberRoleNotice{}
		if err = rows.Scan(&notice.Id, &notice.Uuid, &notice.TeamId, &notice.CeoId, &notice.MemberId, &notice.MemberCurrentRole, &notice.NewRole, &notice.Title, &notice.Content, &notice.Status, &notice.CreatedAt, &notice.UpdatedAt); err != nil {
			return
		}
		notices = append(notices, notice)
	}
	rows.Close()
	return
}

// Team.CountTeamMemberRoleNotices() 统计某个$事业茶团的角色调整声明数量
func (team *Team) CountTeamMemberRoleNotices() (count int, err error) {
	statement := "SELECT COUNT(*) FROM team_member_role_notices WHERE team_id = $1"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(team.Id).Scan(&count)
	return
}

// TeamMemberRoleNotice.UpdateStatus()
func (notice *TeamMemberRoleNotice) UpdateStatus() (err error) {
	statement := "UPDATE team_member_role_notices SET status = $1 WHERE id = $2"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(notice.Status, notice.Id)
	return
}

// TeamMemberRoleNotice.Update()
func (notice *TeamMemberRoleNotice) Update() (err error) {
	statement := "UPDATE team_member_role_notices SET team_id = $1, ceo_id = $2, member_id = $3, member_current_role = $4, new_role = $5, title = $6, content = $7, status = $8, updated_at = $9 WHERE id = $10"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(notice.TeamId, notice.CeoId, notice.MemberId, notice.MemberCurrentRole, notice.NewRole, notice.Title, notice.Content, notice.Status, time.Now(), notice.Id)
	return
}

var TeamProperty = map[int]string{
	0:  "系统$事业茶团",
	1:  "开放式$事业茶团",
	2:  "封闭式$事业茶团",
	10: "开放式草团",
	20: "封闭式草团",
	31: "已接纳开团",
	32: "已婉拒闭团",
}

// 根据给出的团队简称关键词（keyword），查询相似的team.Abbreviation，返回 []team,err，限制返回结果数量为limit
func SearchTeamByAbbreviation(keyword string, limit int, ctx context.Context) ([]Team, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	teams := []Team{}
	rows, err := DB.QueryContext(ctx, "SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, deleted_at, tags FROM teams WHERE abbreviation LIKE $1 AND deleted_at IS NULL AND is_private = false LIMIT $2", "%"+keyword+"%", limit)
	if err != nil {
		return teams, err
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.DeletedAt, &team.Tags); err != nil {
			return teams, err
		}
		teams = append(teams, team)
	}
	rows.Close()
	return teams, nil
}

// Create() UserDefaultTeam{}创建用户设置默认$事业茶团的记录
func (udteam *UserDefaultTeam) Create() (err error) {
	statement := "INSERT INTO user_default_teams (user_id, team_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(udteam.UserId, udteam.TeamId, time.Now()).Scan(&udteam.Id)
	return
}

// GetLastDefaultTeam() 根据user.Id从user_default_teams表和teams表，获取用户最后记录的1个team
func (user *User) GetLastDefaultTeam() (team Team, err error) {
	// 如果用户没有设置默认$事业茶团，则返回系统预设的“自由人”$事业茶团
	count, err := user.SurvivalTeamsCount()
	if err != nil {
		return Team{}, err
	}
	if count == 0 {
		return GetTeam(TeamIdFreelancer)
	}
	team = Team{}
	err = DB.QueryRow("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.is_private, teams.updated_at, teams.tags FROM teams JOIN user_default_teams ON teams.id = user_default_teams.team_id WHERE user_default_teams.user_id = $1 AND teams.deleted_at IS NULL ORDER BY user_default_teams.created_at DESC", user.Id).Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
	if errors.Is(err, sql.ErrNoRows) {
		return GetTeam(TeamIdFreelancer)
	}
	return
}

// GetTeamMemberRoleByTeamId() 获取用户在给定$事业茶团中担任的角色
func GetTeamMemberRoleByTeamIdAndUserId(team_id, user_id int) (role string, err error) {
	err = DB.QueryRow("SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2", team_id, user_id).Scan(&role)
	return
}

// SurvivalTeams() 获取用户当前所在的状态正常的全部$事业茶团,
// team.class = 0、1 or 2, team_members.status = 1
func (user *User) SurvivalTeams() ([]Team, error) {
	count, err := user.SurvivalTeamsCount()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return []Team{}, nil
	}

	query := `
        SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.is_private, teams.updated_at, teams.tags
        FROM teams
        JOIN team_members ON teams.id = team_members.team_id
        WHERE teams.class IN ($1, $2, $3) AND team_members.user_id = $4 AND team_members.status = $5 AND teams.deleted_at IS NULL`

	estimatedCapacity := util.Config.MaxSurvivalTeams //设定用户最大允许活跃$事业茶团数值
	teams := make([]Team, 0, estimatedCapacity)

	query += ` LIMIT $6` // 限制最大团队数
	rows, err := DB.Query(query, TeamClassSpaceship, TeamClassOpen, TeamClassClose, user.Id, TeMemberStatusActive, estimatedCapacity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// SurvivalTeamsCount() 获取用户当前所在的状态正常的全部$事业茶团计数(不包括系统预留的“自由人”$事业茶团)
func (user *User) SurvivalTeamsCount() (count int, err error) {
	query := `
        SELECT COUNT(DISTINCT teams.id)
        FROM teams
        JOIN team_members ON teams.id = team_members.team_id
        WHERE teams.class IN ($1, $2) AND team_members.user_id = $3 AND team_members.status = $4 AND teams.deleted_at IS NULL`

	err = DB.QueryRow(query, TeamClassOpen, TeamClassClose, user.Id, TeMemberStatusActive).Scan(&count)

	return
}

// 获取用户创建的全部$事业茶团，FounderId = UserId
// AWS CodeWhisperer assist in writing
func (user *User) HoldTeams() (teams []Team, err error) {
	rows, err := DB.Query("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE founder_id = $1 AND deleted_at IS NULL", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// 用户担任CEO的$事业茶团，team_member.role = "CEO"
// AWS CodeWhisperer assist in writing
func (user *User) CeoTeams() (teams []Team, err error) {
	rows, err := DB.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.is_private, teams.updated_at, teams.tags FROM teams, team_members WHERE team_members.user_id = $1 AND team_members.team_id = teams.id AND team_members.role = 'CEO' AND teams.deleted_at IS NULL", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// user.FounderTeams() 用户创建的全部$事业茶团，team.FounderId = user.Id, return teams []team
func (usre *User) FounderTeams() (teams []Team, err error) {
	rows, err := DB.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.is_private, teams.updated_at, teams.tags FROM teams WHERE teams.founder_id = $1 AND teams.deleted_at IS NULL", usre.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return

}

// 用户担任核心高管成员的全部$事业茶团，team_member.role = "CEO", "CTO", "CMO", "CFO"
// AWS CodeWhisperer assist in writing
func (user *User) CoreExecTeams() (teams []Team, err error) {
	rows, err := DB.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.is_private, teams.updated_at, teams.tags FROM teams, team_members WHERE team_members.user_id = $1 AND team_members.team_id = teams.id AND (team_members.role = 'CEO' or team_members.role = 'CTO' or team_members.role = 'CMO' or team_members.role = 'CFO') AND teams.deleted_at IS NULL", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// 用户作为普通成员的全部$事业茶团，team_member.role = "taster"
// AWS CodeWhisperer assist in writing
func (user *User) NormalExecTeams() (teams []Team, err error) {
	rows, err := DB.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.is_private, teams.updated_at, teams.tags FROM teams, team_members WHERE team_members.user_id = $1 AND team_members.team_id = teams.id AND team_members.role = 'taster' AND teams.deleted_at IS NULL", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// format the created time to display nicely on the screen
func (team *Team) CreatedAtDate() string {
	return team.CreatedAt.Format(FMT_DATE_CN)
}

// Team.Create()创建$事业茶团
func (team *Team) Create() (err error) {
	statement := "INSERT INTO teams (uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, tags) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), team.Name, team.Mission, team.FounderId, time.Now(), team.Class, team.Abbreviation, team.Logo, team.IsPrivate, team.Tags).Scan(&team.Id, &team.Uuid)
	return
}

// 根据邀请函中的TeamId，查询一个$事业茶团
// AWS CodeWhisperer assist in writing
func (invitation *Invitation) Team() (team Team, err error) {
	team = Team{}
	err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE id = $1 AND deleted_at IS NULL", invitation.TeamId).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
	return
}

// get the nember of teams
// AWS CodeWhisperer assist in writing
// 统计全部注册$事业茶团数量
func GetNumAllTeams() (count int) {
	rows, _ := DB.Query("SELECT COUNT(*) FROM teams WHERE deleted_at IS NULL")
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// 统计某个$事业茶团的成员数
// AWS CodeWhisperer assist in writing
func (team *Team) NumMembers() (count int) {
	rows, _ := DB.Query("SELECT COUNT(*) FROM team_members WHERE team_id = $1 AND status < $2", team.Id, TeMemberStatusResigned)
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// user.CountTeamsByFounderId() 获取用户创建的$事业茶团数量值
func (user *User) CountTeamsByFounderId() (count int, err error) {
	rows, err := DB.Query("SELECT COUNT(*) FROM teams WHERE founder_id = $1 AND deleted_at IS NULL", user.Id)
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// GetTeamByID 根据UUID或ID获取一个$事业茶团详情
func GetTeamByID(uuid string) (team Team, err error) {
	if uuid == "" {
		return team, fmt.Errorf("uuid is empty")
	}
	// 先以uuid查询，如果不存在，再以id查询
	team = Team{}
	err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE uuid = $1 AND deleted_at IS NULL", uuid).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
	if err != nil {
		if err == sql.ErrNoRows {
			err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE id = $1 AND deleted_at IS NULL", uuid).
				Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
		} else {
			return team, fmt.Errorf("查询团队失败:参数: %s, %v", uuid, err)
		}
	}
	return
}

// 根据用户提交的当前Uuid获取一个$事业茶团详情
func GetTeamByUUID(uuid string) (team Team, err error) {

	team = Team{}
	err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE uuid = $1 AND deleted_at IS NULL", uuid).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
	return
}

// Team.Get() 根据$事业茶团Id获取$事业茶团
func (team *Team) Get() (err error) {
	if team.Id == TeamIdNone {
		return fmt.Errorf("team not found with id: %d", team.Id)
	}
	err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE id = $1 AND deleted_at IS NULL", team.Id).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
	return
}

// 获取$事业茶团，查询普通成员，role=“品茶师”（taster）,status = TeMemberStatusActive的方法
func (team *Team) NormalMembers() (team_members []TeamMember, err error) {

	if team.Id == TeamIdNone {
		return nil, fmt.Errorf("team not found with id: %d", team.Id)
	}
	if team.Id == TeamIdFreelancer {
		return nil, fmt.Errorf("team member cannot find with id: %d", team.Id)
	}
	rows, err := DB.Query("SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at FROM team_members WHERE team_id = $1 AND role = $2 AND status = $3", team.Id, "taster", TeMemberStatusActive)
	if err != nil {
		return
	}
	for rows.Next() {
		teamMember := TeamMember{}
		if err = rows.Scan(&teamMember.Id, &teamMember.Uuid, &teamMember.TeamId, &teamMember.UserId, &teamMember.Role, &teamMember.CreatedAt, &teamMember.Status, &teamMember.UpdatedAt); err != nil {
			return
		}
		team_members = append(team_members, teamMember)
	}
	rows.Close()
	return
}

// coreMember() 返回$事业茶团核心成员,teamMember.Role = “CEO” and “CTO” and “CMO” and “CFO”,status = TeMemberStatusActive
func (team *Team) CoreMembers() (team_members []TeamMember, err error) {
	if team.Id == TeamIdNone {
		return nil, fmt.Errorf("team not found with id: %d", team.Id)
	}
	if team.Id == TeamIdFreelancer {
		return nil, fmt.Errorf("team member cannot find with id: %d", team.Id)
	}
	rows, err := DB.Query("SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at FROM team_members WHERE team_id = $1 AND (role = $2 OR role = $3 OR role = $4 OR role = $5) AND status = $6", team.Id, RoleCEO, "CTO", RoleCMO, RoleCFO, TeMemberStatusActive)
	if err != nil {
		return
	}
	for rows.Next() {
		teamMember := TeamMember{}
		if err = rows.Scan(&teamMember.Id, &teamMember.Uuid, &teamMember.TeamId, &teamMember.UserId, &teamMember.Role, &teamMember.CreatedAt, &teamMember.Status, &teamMember.UpdatedAt); err != nil {
			return
		}
		team_members = append(team_members, teamMember)
	}
	rows.Close()
	return
}

// GetAllMemberUserIdsByTeamId() 从TeamMember获取某个茶团，全部状态正常的成员User_ids，返回 []user_id, err
func GetAllMemberUserIdsByTeamId(team_id int) (user_ids []int, err error) {
	if team_id == TeamIdNone {
		return nil, fmt.Errorf("team not found with id: %d", team_id)
	}
	if team_id == TeamIdFreelancer {
		return nil, fmt.Errorf("team member cannot find with id: %d", team_id)
	}
	rows, err := DB.Query("SELECT user_id FROM team_members WHERE team_id = $1 AND status = 1", team_id)
	if err != nil {
		return
	}
	for rows.Next() {
		var user_id int
		if err = rows.Scan(&user_id); err != nil {
			return
		}
		user_ids = append(user_ids, user_id)
	}
	rows.Close()
	return
}

// 根据用户id，检查当前用户是否$事业茶团成员；team中是否存在某个teamMember
func GetMemberByTeamIdUserId(team_id, user_id int) (team_member TeamMember, err error) {
	team_member = TeamMember{}
	if team_id == TeamIdNone {
		return team_member, fmt.Errorf("team not found with id: %d", team_id)
	}
	if team_id == TeamIdFreelancer {
		return team_member, fmt.Errorf("team member cannot find with id: %d", team_id)
	}
	err = DB.QueryRow("SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at FROM team_members WHERE team_id = $1 AND user_id = $2", team_id, user_id).
		Scan(&team_member.Id, &team_member.Uuid, &team_member.TeamId, &team_member.UserId, &team_member.Role, &team_member.CreatedAt, &team_member.Status, &team_member.UpdatedAt)
	return
}

// team *Team.IsMember() 检查当前用户是否$事业茶团成员；team中是否存在某个teamMember,如果是返回true，否则返回false
func (team *Team) IsMember(user_id int) (is_member bool, err error) {
	if team.Id == TeamIdNone {
		return false, fmt.Errorf("team not found with id: %d", team.Id)
	}
	if team.Id == TeamIdFreelancer {
		return true, nil
	}
	team_member, err := GetMemberByTeamIdUserId(team.Id, user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 没有找到团队记录,不可能是成员 --[DeepSeek said]
			return false, nil
		} else {
			return false, err
		}
	}

	return team_member.UserId > 0, nil
}

// team *Team.IsCoreMember() 检查当前用户是否$事业茶团核心成员（CEO/CTO/CMO/CFO）
func (team *Team) IsCoreMember(user_id int) (bool, error) {
	if team.Id == TeamIdNone {
		return false, fmt.Errorf("team not found with id: %d", team.Id)
	}
	if team.Id == TeamIdFreelancer {
		return false, nil
	}
	team_member, err := GetMemberByTeamIdUserId(team.Id, user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	// 检查是否为核心成员角色
	return team_member.Role == RoleCEO || team_member.Role == "CTO" || team_member.Role == RoleCMO || team_member.Role == RoleCFO, nil
}

// 查询一个$事业茶团team的担任CEO的成员资料，不是founder，是teamMember.Role = “CEO”，返回 (team_member TeamMember,err error)
// AWS CodeWhisperer assist in writing
func (team *Team) MemberCEO() (team_member TeamMember, err error) {
	team_member = TeamMember{}
	err = DB.QueryRow("SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at FROM team_members WHERE team_id = $1 AND role = $2", team.Id, RoleCEO).
		Scan(&team_member.Id, &team_member.Uuid, &team_member.TeamId, &team_member.UserId, &team_member.Role, &team_member.CreatedAt, &team_member.Status, &team_member.UpdatedAt)
	return
}

// GetTeamMemberByRole() 根据角色查找$事业茶团成员资料。用于检查$事业茶团拟邀请的新成员角色是否已经被占用
func (team *Team) GetTeamMemberByRole(role string) (team_member TeamMember, err error) {
	team_member = TeamMember{}
	err = DB.QueryRow("SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at FROM team_members WHERE team_id = $1 AND role = $2", team.Id, role).
		Scan(&team_member.Id, &team_member.Uuid, &team_member.TeamId, &team_member.UserId, &team_member.Role, &team_member.CreatedAt, &team_member.Status, &team_member.UpdatedAt)
	return
}

// CheckTeamMemberByRole 根据角色查找团队成员
// 参数:
//   - role: 要查询的角色名称
//
// 返回:
//   - *TeamMember: 如果找到返回成员指针，否则返回nil
//   - error: 如果查询出错返回错误，未找到不视为错误
func (team *Team) CheckTeamMemberByRole(role string) (*TeamMember, error) {
	if team == nil || team.Id == TeamIdNone || team.Id == TeamIdFreelancer {
		return nil, fmt.Errorf("invalid team id %d", team.Id)
	}

	if role == "" {
		return nil, errors.New("team role cannot be empty")
	}

	teamMember := &TeamMember{}
	err := DB.QueryRow(
		"SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at "+
			"FROM team_members WHERE team_id = $1 AND role = $2",
		team.Id, role,
	).Scan(
		&teamMember.Id, &teamMember.Uuid, &teamMember.TeamId,
		&teamMember.UserId, &teamMember.Role, &teamMember.CreatedAt,
		&teamMember.Status, &teamMember.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 未找到记录不算错误
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query team member: %w", err)
	}

	return teamMember, nil
}

// 根据team_member struct 生成Create()方法
func (tM *TeamMember) Create() (err error) {
	statement := `INSERT INTO team_members (uuid, team_id, user_id, role, created_at, status)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id,uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), tM.TeamId, tM.UserId, tM.Role, time.Now(), tM.Status).Scan(&tM.Id, &tM.Uuid)

	return
}
func (teamMember *TeamMember) CreatedAtDate() string {
	return teamMember.CreatedAt.Format(FMT_DATE_CN)
}

// TeamMember.Get()
func (teamMember *TeamMember) Get() (err error) {
	err = DB.QueryRow("SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at FROM team_members WHERE id = $1", teamMember.Id).
		Scan(&teamMember.Id, &teamMember.Uuid, &teamMember.TeamId, &teamMember.UserId, &teamMember.Role, &teamMember.CreatedAt, &teamMember.Status, &teamMember.UpdatedAt)
	return
}

// teamMemberUpdate() 更新$事业茶团成员的角色和状态
func (teamMember *TeamMember) UpdateRoleStatus() (err error) {
	statement := `UPDATE team_members SET role = $2, updated_at = $3, status = $4 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(teamMember.Id, teamMember.Role, time.Now(), teamMember.Status)
	return
}

// 更换$事业茶团默认CEO的方法，Update team_members记录中role=CEO的行 user_id 为当前user_id
func (teamMember *TeamMember) UpdateFirstCEO(user_id int) (err error) {
	statement := `UPDATE team_members SET user_id = $1, updated_at = $2 WHERE team_id = $3 AND role = $4`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(user_id, time.Now(), teamMember.TeamId, RoleCEO)
	return
}

// 根据teamMember.teamId获取Team()，返回成员所在team的信息
func (teamMember *TeamMember) Team() (team Team, err error) {
	team = Team{}
	err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE id = $1 AND deleted_at IS NULL", teamMember.TeamId).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
	return
}

// GetTeamByName()
func (team *Team) GetByName() (err error) {
	err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE name = $1 AND deleted_at IS NULL", team.Name).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
	return
}

// InvitedTeams()
// 根据ProjectId从LicenceTeam获取[]TeamId,然后用teamId，获取对应的Team，最后返回[]team
// 获取一个封闭式茶台的全部受邀请$事业茶团
func (project *Project) InvitedTeams() (teams []Team, err error) {
	rows, err := DB.Query("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE id IN (SELECT team_id FROM project_invited_teams WHERE project_id = $1) AND deleted_at IS NULL", project.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// GetTeamByAbbreviation()
func (team *Team) GetByAbbreviation() (err error) {
	err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, tags FROM teams WHERE abbreviation = $1 AND deleted_at IS NULL", team.Abbreviation).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.Tags)
	return
}

// GetGroupFirstTeam 根据group.first_team_id获取team
func GetGroupFirstTeam(groupID int) (team Team, err error) {
	team = Team{}
	err = DB.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, is_private, updated_at, deleted_at, tags FROM teams WHERE id = (SELECT first_team_id FROM groups WHERE id = $1)", groupID).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.IsPrivate, &team.UpdatedAt, &team.DeletedAt, &team.Tags)
	return
}

// 获取开放式$事业茶团的数量
func OpenTeamCount() (count int) {
	rows, err := DB.Query("SELECT count(*) FROM teams WHERE class = 1 AND deleted_at IS NULL")
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// 获取封闭式$事业茶团数量
func ClosedTeamCount() (count int) {
	rows, err := DB.Query("SELECT count(*) FROM teams WHERE class = 2 AND deleted_at IS NULL")
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// 获取$事业茶团的属性
func (team *Team) TeamProperty() string {
	return TeamProperty[team.Class]
}

// Update 更新团队信息
func (team *Team) Update() error {
	now := time.Now()
	team.UpdatedAt = &now
	statement := `UPDATE teams SET name = $1, mission = $2, class = $3, 
	              abbreviation = $4, logo = $5, tags = $6, is_private = $7, updated_at = $8 
	              WHERE id = $9`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Name, team.Mission, team.Class, team.Abbreviation,
		team.Logo, team.Tags, team.IsPrivate, now, team.Id)
	return err
}

// UpdateClass 更新团队类型
func (team *Team) UpdateClass() (err error) {
	statement := `UPDATE teams SET class = $1, updated_at = $2 WHERE id = $3`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Class, time.Now(), team.Id)
	return
}

// UpdateLogo 更新团队标志
func (team *Team) UpdateLogo() (err error) {
	statement := `UPDATE teams SET logo = $1, updated_at = $2 WHERE id = $3`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Logo, time.Now(), team.Id)
	return
}

// UpdateAbbreviation 更新团队简称
func (team *Team) UpdateAbbreviation() (err error) {
	statement := `UPDATE teams SET abbreviation = $1, updated_at = $2 WHERE id = $3`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Abbreviation, time.Now(), team.Id)
	return
}

// UpdateName 更新团队名称
func (team *Team) UpdateName() (err error) {
	statement := `UPDATE teams SET name = $1, updated_at = $2 WHERE id = $3`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Name, time.Now(), team.Id)
	return
}

// UpdateMission 更新团队使命
func (team *Team) UpdateMission() (err error) {
	statement := `UPDATE teams SET mission = $1, updated_at = $2 WHERE id = $3`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Mission, time.Now(), team.Id)
	return
}

// UpdateIsPrivate 更新团队私密属性
func (team *Team) UpdateIsPrivate() (err error) {
	statement := `UPDATE teams SET is_private = $1, updated_at = $2 WHERE id = $3`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.IsPrivate, time.Now(), team.Id)
	return
}

// CreateWithTx 在事务中创建$事业茶团
func (team *Team) CreateWithTx(tx *sql.Tx) (err error) {
	err = tx.QueryRow("INSERT INTO teams (uuid, name, mission, founder_id, created_at, class, abbreviation, logo, tags, is_private) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid", Random_UUID(), team.Name, team.Mission, team.FounderId, time.Now(), team.Class, team.Abbreviation, team.Logo, team.Tags, team.IsPrivate).Scan(&team.Id, &team.Uuid)
	return
}

// CreateWithTx 在事务中创建团队成员
func (tM *TeamMember) CreateWithTx(tx *sql.Tx) (err error) {
	err = tx.QueryRow(`INSERT INTO team_members (uuid, team_id, user_id, role, created_at, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id,uuid`, Random_UUID(), tM.TeamId, tM.UserId, tM.Role, time.Now(), tM.Status).Scan(&tM.Id, &tM.Uuid)
	return
}

// GetResignedMembersByTeamId 获取已离开的成员列表
func GetResignedMembersByTeamId(teamId int) ([]TeamMember, error) {
	rows, err := DB.Query("SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at FROM team_members WHERE team_id = $1 AND status = $2 ORDER BY updated_at DESC", teamId, TeMemberStatusResigned)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []TeamMember
	for rows.Next() {
		var member TeamMember
		if err = rows.Scan(&member.Id, &member.Uuid, &member.TeamId, &member.UserId, &member.Role, &member.CreatedAt, &member.Status, &member.UpdatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}
