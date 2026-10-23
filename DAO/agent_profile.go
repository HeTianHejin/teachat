package dao

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// AGI 代理范围。
//
// 这些值必须与程序中真正执行的权限判断一一对应，否则“代理范围”就只是声明文字。
// 目前先作为申报与展示的规范词表，后续接入具体权限判断时应直接引用这些常量。
const (
	AgentScopeRead   = "read"   // 读取站内公开内容
	AgentScopePost   = "post"   // 发表品味与回复
	AgentScopeThread = "thread" // 发起新的茶议
	AgentScopeTrade  = "trade"  // 发起星茶相关操作
	AgentScopeSign   = "sign"   // 代表受委托团队签署声明
)

// AgentScopeOption 代理范围选项，供注册表单与页面展示使用。
type AgentScopeOption struct {
	Key   string
	Label string
	Desc  string
}

// AgentScopeOptions 代理范围白名单。
var AgentScopeOptions = []AgentScopeOption{
	{Key: AgentScopeRead, Label: "阅读", Desc: "读取站内公开内容"},
	{Key: AgentScopePost, Label: "发言", Desc: "发表品味与回复"},
	{Key: AgentScopeThread, Label: "发起茶议", Desc: "发起新的茶议主题"},
	{Key: AgentScopeTrade, Label: "星茶操作", Desc: "代表受委托团队发起星茶操作"},
	{Key: AgentScopeSign, Label: "代表签署", Desc: "代表受委托团队签署声明"},
}

// AGI 自主程度。
const (
	AgentAutonomyAssisted   = 1 // 辅助：每次输出需人类确认
	AgentAutonomySupervised = 2 // 受监督：可自主行动，人类可随时接管
	AgentAutonomyAutonomous = 3 // 自主：在代理范围内自主行动
)

// AgentAutonomyOption 自主程度选项。
type AgentAutonomyOption struct {
	Level int
	Label string
	Desc  string
}

// AgentAutonomyOptions 自主程度白名单。
var AgentAutonomyOptions = []AgentAutonomyOption{
	{Level: AgentAutonomyAssisted, Label: "辅助", Desc: "每次输出需人类确认"},
	{Level: AgentAutonomySupervised, Label: "受监督", Desc: "可自主行动，人类可随时接管"},
	{Level: AgentAutonomyAutonomous, Label: "自主", Desc: "在代理范围内自主行动"},
}

// AgentAutonomyLabel 返回自主程度的中文名称。
func AgentAutonomyLabel(level int) string {
	for _, option := range AgentAutonomyOptions {
		if option.Level == level {
			return option.Label
		}
	}
	return "未知"
}

// AgentScopeLabel 返回代理范围的中文名称。
func AgentScopeLabel(key string) string {
	for _, option := range AgentScopeOptions {
		if option.Key == key {
			return option.Label
		}
	}
	return key
}

// NormalizeAgencyScopes 校验、去重并归一化提交的代理范围。
func NormalizeAgencyScopes(raw []string) ([]string, error) {
	allowed := make(map[string]bool, len(AgentScopeOptions))
	for _, option := range AgentScopeOptions {
		allowed[option.Key] = true
	}

	scopes := make([]string, 0, len(raw))
	seen := make(map[string]bool, len(raw))
	for _, value := range raw {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if !allowed[key] {
			return nil, fmt.Errorf("未知的代理范围: %q", value)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		scopes = append(scopes, key)
	}
	if len(scopes) == 0 {
		return nil, fmt.Errorf("请至少选择一项代理范围")
	}
	return scopes, nil
}

// ParseAgencyScopes 把逗号分隔的存储值拆成范围列表。
func ParseAgencyScopes(raw string) []string {
	parts := strings.Split(raw, ",")
	scopes := make([]string, 0, len(parts))
	for _, part := range parts {
		if key := strings.TrimSpace(part); key != "" {
			scopes = append(scopes, key)
		}
	}
	return scopes
}

// AgentProfile AGI 代理账号的公开申报资料。
// 与 users 一对一：只有 role = AGI 的用户才会拥有记录。
type AgentProfile struct {
	Id            int
	Uuid          string
	UserId        int
	SponsorTeamId int
	SponsorUserId int
	ModelFamily   string
	ModelVersion  string
	Provider      string
	AgencyScope   string // 逗号分隔，取值受 NormalizeAgencyScopes 约束
	AutonomyLevel int
	DeclaredAt    time.Time
	UpdatedAt     *time.Time
	RevokedAt     *time.Time
	RevokedReason string

	// 以下字段不落库，仅用于页面展示。
	SponsorTeam Team
	SponsorUser User
}

// ScopeList 返回代理范围列表。
func (profile AgentProfile) ScopeList() []string {
	return ParseAgencyScopes(profile.AgencyScope)
}

// IsRevoked 赞助关系是否已被撤销。
func (profile AgentProfile) IsRevoked() bool {
	return profile.RevokedAt != nil
}

// HasScope 判断公开申报的代理范围中是否包含某一项。
func (profile AgentProfile) HasScope(scope string) bool {
	scope = strings.ToLower(strings.TrimSpace(scope))
	for _, item := range profile.ScopeList() {
		if item == scope {
			return true
		}
	}
	return false
}

// CheckAgentScope 统一的代理范围校验入口。
//
// 人类用户始终放行；申报为 AGI 的账号必须：
//  1. 存在 agent_profiles 申报记录；
//  2. 赞助关系未被撤销；
//  3. agency_scope 包含目标范围。
//
// 无申报资料的 AGI 一律拒绝（fail-closed），不允许退化成人类账号。
func CheckAgentScope(user User, scope string) (bool, error) {
	if !user.IsAGI() {
		return true, nil
	}

	profile, err := GetAgentProfileByUserId(user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if profile.IsRevoked() {
		return false, nil
	}
	return profile.HasScope(scope), nil
}

// IsSystemTeam 判断是否为系统预设团队。系统团队不能作为 AGI 的赞助团队。
func IsSystemTeam(teamId int) bool {
	switch teamId {
	case TeamIdNone, TeamIdFreelancer, TeamIdSpaceshipCrew, TeamIdVerifier, TeamIdEscrow, TeamIdPublicGovernance:
		return true
	default:
		return false
	}
}

// CanSponsorAgent 校验某个茶友是否有资格为 AGI 代理账号背书。
// 规则：赞助人必须是申报的人类；且必须是该团队的发起人或现任在职 CEO；
// 系统预设团队不允许作为赞助团队。
func CanSponsorAgent(teamId, userId int) (bool, error) {
	if teamId <= 0 || userId <= 0 {
		return false, nil
	}
	if IsSystemTeam(teamId) {
		return false, nil
	}

	sponsor, err := GetUser(userId)
	if err != nil {
		return false, err
	}
	if !sponsor.IsHuman() {
		return false, nil
	}

	team, err := GetTeam(teamId)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if team.IsDeleted() {
		return false, nil
	}
	if team.FounderId == userId {
		return true, nil
	}

	var role int
	err = DB.QueryRow(
		"SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2 AND status = $3",
		teamId, userId, TeamMemberStatusActive,
	).Scan(&role)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return role == RoleCEO, nil
}

// GetSponsorableTeams 返回某位茶友可以作为赞助人的团队列表：
// 由他发起、或他担任在职 CEO 的团队，排除系统预设团队。
func GetSponsorableTeams(userId int) ([]Team, error) {
	if userId <= 0 {
		return nil, fmt.Errorf("invalid user id: %d", userId)
	}

	rows, err := DB.Query(`
		SELECT DISTINCT t.id, t.uuid, t.name, t.mission, t.founder_id, t.created_at,
		       t.class, t.nature, t.abbreviation, t.logo, t.is_private, t.updated_at,
		       t.deleted_at, t.tags, t.primary_industry_tag_id, t.industry_path
		FROM teams t
		LEFT JOIN team_members tm
		       ON tm.team_id = t.id AND tm.user_id = $1 AND tm.status = $2
		WHERE t.deleted_at IS NULL
		  AND t.id NOT IN ($3, $4, $5, $6, $7)
		  AND (t.founder_id = $1 OR tm.role = $8)
		ORDER BY t.created_at DESC
	`, userId, TeamMemberStatusActive,
		TeamIdFreelancer, TeamIdSpaceshipCrew, TeamIdVerifier, TeamIdEscrow, TeamIdPublicGovernance,
		RoleCEO)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]Team, 0, 4)
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

// GetAgentProfileByUserId 读取某个 AGI 用户的申报资料。
func GetAgentProfileByUserId(userId int) (profile AgentProfile, err error) {
	err = DB.QueryRow(`
		SELECT id, uuid, user_id, sponsor_team_id, sponsor_user_id,
		       model_family, model_version, provider, agency_scope, autonomy_level,
		       declared_at, updated_at, revoked_at, COALESCE(revoked_reason, '')
		FROM agent_profiles WHERE user_id = $1
	`, userId).Scan(
		&profile.Id, &profile.Uuid, &profile.UserId, &profile.SponsorTeamId, &profile.SponsorUserId,
		&profile.ModelFamily, &profile.ModelVersion, &profile.Provider, &profile.AgencyScope, &profile.AutonomyLevel,
		&profile.DeclaredAt, &profile.UpdatedAt, &profile.RevokedAt, &profile.RevokedReason,
	)
	return profile, err
}

// FillDisplay 填充赞助团队与赞助人信息，供页面展示。
func (profile *AgentProfile) FillDisplay() error {
	if profile.SponsorTeamId > 0 {
		if team, err := GetTeam(profile.SponsorTeamId); err == nil {
			profile.SponsorTeam = team
		}
	}
	if profile.SponsorUserId > 0 {
		if user, err := GetUser(profile.SponsorUserId); err == nil {
			profile.SponsorUser = user
		}
	}
	return nil
}

// CreateAgentAccount 在一个事务中创建 AGI 用户、申报资料、团队成员记录与默认团队。
// 任一步失败即整体回滚，避免出现“无申报资料的 AGI”或“无团队归属的代理”。
func CreateAgentAccount(newUser *User, profile *AgentProfile) error {
	if newUser == nil || profile == nil {
		return fmt.Errorf("用户与申报资料不能为空")
	}
	if profile.SponsorTeamId <= 0 {
		return fmt.Errorf("必须绑定受委托团队")
	}

	newUser.NormalizeNameFields()

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	err = tx.QueryRow(`
		INSERT INTO users (uuid, name, family_name, given_name, alias_name, email, password,
		                   created_at, biography, role, gender, avatar)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, uuid
	`, Random_UUID(), newUser.Name, newUser.FamilyName, newUser.GivenName, newUser.AliasName,
		newUser.Email, newUser.Password, time.Now(), newUser.Biography, newUser.Role,
		newUser.Gender, newUser.Avatar).Scan(&newUser.Id, &newUser.Uuid)
	if err != nil {
		return fmt.Errorf("创建AGI用户失败: %w", err)
	}

	profile.UserId = newUser.Id
	err = tx.QueryRow(`
		INSERT INTO agent_profiles (uuid, user_id, sponsor_team_id, sponsor_user_id,
		                            model_family, model_version, provider, agency_scope,
		                            autonomy_level, declared_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, uuid, declared_at
	`, Random_UUID(), profile.UserId, profile.SponsorTeamId, profile.SponsorUserId,
		profile.ModelFamily, profile.ModelVersion, profile.Provider, profile.AgencyScope,
		profile.AutonomyLevel, time.Now()).Scan(&profile.Id, &profile.Uuid, &profile.DeclaredAt)
	if err != nil {
		return fmt.Errorf("创建AGI申报资料失败: %w", err)
	}

	member := TeamMember{
		TeamId: profile.SponsorTeamId,
		UserId: newUser.Id,
		Role:   RoleTaster,
		Status: TeamMemberStatusActive,
	}
	if err = member.CreateWithTx(tx); err != nil {
		return fmt.Errorf("登记AGI团队成员失败: %w", err)
	}

	if _, err = tx.Exec(
		"INSERT INTO user_default_teams (user_id, team_id, created_at) VALUES ($1, $2, $3)",
		newUser.Id, profile.SponsorTeamId, time.Now(),
	); err != nil {
		return fmt.Errorf("设置AGI默认团队失败: %w", err)
	}

	return tx.Commit()
}
