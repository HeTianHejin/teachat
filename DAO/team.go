package data

import (
	"errors"
	"time"
)

// 根据茶团team_id_list，获取全部申请加盟的茶团[]team
func GetTeamsByIds(team_id_list []int) (teams []Team, err error) {
	n := len(team_id_list)
	if n == 0 {
		return nil, errors.New("team_id_list is empty")
	}
	teams = make([]Team, n)
	rows, err := Db.Query("SELECT * FROM team WHERE id IN (?)", team_id_list)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// 茶团=团队,同一桌喝茶的人，
// 原则上来说team的人数上限是12人，1 dozen。
type Team struct {
	Id           int
	Uuid         string
	Name         string
	Mission      string
	FounderId    int
	CreatedAt    time.Time
	Class        int    //0:"系统茶团", 1: "开放式茶团",2: "封闭式茶团",10: "开放式草团",20: "封闭式草团"
	Abbreviation string // 队名简称
	Logo         string // 团队标志
	UpdatedAt    time.Time
	GroupId      int // 上级集合单位(集团)id
}

// 团队成员=当前茶团加入成员记录
type TeamMember struct {
	Id        int
	Uuid      string
	TeamId    int
	UserId    int
	Role      string // 角色，职务,分别为：CEO，CTO，CMO，CFO，taster
	CreatedAt time.Time
	Class     int // 状态指数， 默认正常=1,活跃=2?，失联=0?
}

// 团队成员离开记录
type TeamMemberLeave struct {
	Id        int
	Uuid      string
	TeamId    int
	UserId    int
	Reason    string //离队原因
	LeaveTime time.Time
}

// 用户的“默认团队”设置记录
type UserDefaultTeam struct {
	Id        int
	UserId    int
	TeamId    int
	CreatedAt time.Time
}

// 记录团队成员角色变动书
type TeamRole struct {
	Id                 int
	Uuid               string
	TeamId             int
	TeamCeoUserId      int
	TargetTeamMemberId int
	Role               string
	Word               string
	CreatedAt          time.Time
	CheckTeamMemberId  int
	CheckAt            time.Time
}

var TeamProperty = map[int]string{
	0:  "系统茶团",
	1:  "开放式茶团",
	2:  "封闭式茶团",
	10: "开放式草团",
	20: "封闭式草团",
	31: "已接纳开团",
	32: "已婉拒闭团",
}

// 根据给出的关键词（keyword），查询相似的team.Abbreviation，返回 []team,err
func SearchTeamByAbbreviation(keyword string) ([]Team, error) {
	teams := []Team{}
	rows, err := Db.Query("SELECT * FROM teams WHERE abbreviation LIKE $1", "%"+keyword+"%")
	if err != nil {
		return teams, err
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return teams, err
		}
		teams = append(teams, team)
	}
	rows.Close()
	return teams, nil
}

// Create() UserDefaultTeam{}创建用户设置默认团队的记录
func (udteam *UserDefaultTeam) Create() (err error) {
	statement := "INSERT INTO user_default_teams (user_id, team_id) VALUES ($1, $2) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(udteam.UserId, udteam.TeamId).Scan(&udteam.Id)
	return
}

// 根据user_id 获取最后1条默认团队设置记录，返回udteam UserDefaultTeam
func GetDefaultTeamRecordByUserId(user_id int) (udteam UserDefaultTeam, err error) {
	err = Db.QueryRow("SELECT * FROM user_default_teams WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1", user_id).Scan(&udteam.Id, &udteam.UserId, &udteam.TeamId, &udteam.CreatedAt)
	return
}

// GetLastDefaultTeam() 根据user.Id从UserDefaultTeams获取用户最后记录的1个team
func (user *User) GetLastDefaultTeam() (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.updated_at, teams.group_id FROM teams JOIN user_default_teams ON teams.id = user_default_teams.team_id WHERE user_default_teams.user_id = $1 ORDER BY user_default_teams.created_at DESC LIMIT 1", user.Id).Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// GetTeamMemberRoleByTeamId() 获取用户在给定团队中担任的角色
func GetTeamMemberRoleByTeamIdAndUserId(team_id, user_id int) (role string, err error) {
	err = Db.QueryRow("SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2", team_id, user_id).Scan(&role)
	return
}

// SurvivalTeams() 获取用户当前所在的状态正常的全部团队,team.class = 1 or 2, team_members.class = 1
func (user *User) SurvivalTeams() ([]Team, error) {
	query := `
        SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.updated_at, teams.group_id
        FROM teams
        JOIN team_members ON teams.id = team_members.team_id
        WHERE teams.class IN (1, 2) AND team_members.user_id = $1 AND team_members.class = 1`

	estimatedCapacity := 6 //设定用户最多创建3茶团+担任3ceo
	teams := make([]Team, 0, estimatedCapacity)

	rows, err := Db.Query(query, user.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// SurvivalTeamsCount() 获取用户当前所在的状态正常的全部团队计数(不包括系统预留的“自由人”茶团)
func (user *User) SurvivalTeamsCount() (count int, err error) {
	query := `
        SELECT COUNT(DISTINCT teams.id)
        FROM teams
        JOIN team_members ON teams.id = team_members.team_id
        WHERE teams.class IN (1, 2) AND team_members.user_id = $1 AND team_members.class = 1`

	err = Db.QueryRow(query, user.Id).Scan(&count)
	//减记自由人茶团计数
	count = count - 1
	return
}

// 获取全部封闭式团队的信息
func GetClosedTeams() (teams []Team, err error) {
	rows, err := Db.Query("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE class = 2")
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// 获取全部开放式团队对象
func GetOpenTeams() (teams []Team, err error) {
	rows, err := Db.Query("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE class = 1")
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// 根据用户的id获取全部加入的团队
// AWS CodeWhisperer assist in writing
// func (user *User) JoinedTeams() (teams []Team, err error) {
// 	rows, err := Db.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.updated_at, teams.group_id FROM teams, team_members WHERE team_members.user_id = $1 AND team_members.team_id = teams.id", user.Id)
// 	if err != nil {
// 		return
// 	}
// 	for rows.Next() {
// 		team := Team{}
// 		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
// 			return
// 		}
// 		teams = append(teams, team)
// 	}
// 	rows.Close()
// 	return
// }

// 获取用户创建的全部团队，FounderId = UserId
// AWS CodeWhisperer assist in writing
func (user *User) HoldTeams() (teams []Team, err error) {
	rows, err := Db.Query("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE founder_id = $1", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// 用户担任CEO的团队，team_member.role = "CEO"
// AWS CodeWhisperer assist in writing
func (user *User) CeoTeams() (teams []Team, err error) {
	rows, err := Db.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.updated_at, teams.group_id FROM teams, team_members WHERE team_members.user_id = $1 AND team_members.team_id = teams.id AND team_members.role = 'CEO'", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// user.FounderTeams() 用户创建的全部团队，team.FounderId = user.Id, return teams []team
func (usre *User) FounderTeams() (teams []Team, err error) {
	rows, err := Db.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.updated_at, teams.group_id FROM teams WHERE teams.founder_id = $1", usre.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return

}

// 用户担任核心高管成员的全部茶团，team_member.role = "CEO", "CTO", "CMO", "CFO"
// AWS CodeWhisperer assist in writing
func (user *User) CoreExecTeams() (teams []Team, err error) {
	rows, err := Db.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.updated_at, teams.group_id FROM teams, team_members WHERE team_members.user_id = $1 AND team_members.team_id = teams.id AND (team_members.role = 'CEO' or team_members.role = 'CTO' or team_members.role = 'CMO' or team_members.role = 'CFO')", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// 用户作为普通成员的全部茶团，team_member.role = "taster"
// AWS CodeWhisperer assist in writing
func (user *User) NormalExecTeams() (teams []Team, err error) {
	rows, err := Db.Query("SELECT teams.id, teams.uuid, teams.name, teams.mission, teams.founder_id, teams.created_at, teams.class, teams.abbreviation, teams.logo, teams.updated_at, teams.group_id FROM teams, team_members WHERE team_members.user_id = $1 AND team_members.team_id = teams.id AND team_members.role = 'taster'", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
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

// create a new team
// AWS CodeWhisperer assist in writing
func (user *User) CreateTeam(name, abbreviation, mission, logo string, class, group_id int) (team Team, err error) {

	statement := `INSERT INTO teams (uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	cmd := stmt.QueryRow(Random_UUID(), name, mission, user.Id, time.Now(), class, abbreviation, logo, time.Now(), group_id)
	err = cmd.Scan(
		&team.Id,
		&team.Uuid,
		&team.Name,
		&team.Mission,
		&team.FounderId,
		&team.CreatedAt,
		&team.Class,
		&team.Abbreviation,
		&team.Logo,
		&team.UpdatedAt,
		&team.GroupId,
	)

	if err != nil {
		return
	}
	return
}

// create a new team member。
// 注意，这里的userId，可能是某个拉人入会的团队成员，不一定是团队创建者！
// 注意，这里的role是团队成员的角色，不是用户的角色。
// AWS CodeWhisperer assist in writing 20240127
func AddTeamMember(teamId int, user_id int, role string) (teamMember TeamMember, err error) {
	statement := `INSERT INTO team_members (uuid, team_id, user_id, role, created_at) 
	VALUES ($1, $2, $3, $4, $5) RETURNING id, uuid, team_id, user_id, role, created_at`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	cmd := stmt.QueryRow(Random_UUID(), teamId, user_id, role, time.Now())
	err = cmd.Scan(
		&teamMember.Id,
		&teamMember.Uuid,
		&teamMember.TeamId,
		&teamMember.UserId,
		&teamMember.Role,
		&teamMember.CreatedAt)
	return
}

// 根据邀请函中的TeamId，查询一个茶团
// AWS CodeWhisperer assist in writing
func (invitation *Invitation) Team() (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE id = $1", invitation.TeamId).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// get the nember of teams
// AWS CodeWhisperer assist in writing
// 统计全部注册团队数量
func GetNumAllTeams() (count int) {
	rows, _ := Db.Query("SELECT COUNT(*) FROM teams")
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// 获取某个团队的成员数
// AWS CodeWhisperer assist in writing
func (team *Team) NumMembers() (count int) {
	rows, _ := Db.Query("SELECT COUNT(*) FROM team_members WHERE team_id = $1", team.Id)
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// user.CountTeamsByFounderId() 获取用户创建的团队数量值
func (user *User) CountTeamsByFounderId() (count int, err error) {
	rows, err := Db.Query("SELECT COUNT(*) FROM teams WHERE founder_id = $1", user.Id)
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

// 根据用户提交的当前Uuid获取一个团队详情
// AWS CodeWhisperer assist in writing
func GetTeamByUuid(uuid string) (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE uuid = $1", uuid).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// 根据用户提交的Id获取一个团队
// AWS CodeWhisperer assist in writing
func GetTeamById(id int) (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE id = $1", id).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// 获取团队全部普通成员role=“品茶师”的方法
// AWS CodeWhisperer assist in writing
func (team *Team) NormalMembers() (team_members []TeamMember, err error) {
	rows, err := Db.Query("SELECT id, uuid, team_id, user_id, role, created_at, class FROM team_members WHERE team_id = $1 AND role = $2", team.Id, "taster")
	if err != nil {
		return
	}
	for rows.Next() {
		member := TeamMember{}
		if err = rows.Scan(&member.Id, &member.Uuid, &member.TeamId, &member.UserId, &member.Role, &member.CreatedAt, &member.Class); err != nil {
			return
		}
		team_members = append(team_members, member)
	}
	rows.Close()
	return
}

// coreMember() 返回茶团核心成员,teamMember.Role = “CEO” and “CTO” and “CMO” and “CFO”
// AWS CodeWhisperer assist in writing
func (team *Team) CoreMembers() (teamMembers []TeamMember, err error) {
	rows, err := Db.Query("SELECT id, uuid, team_id, user_id, role, created_at, class FROM team_members WHERE team_id = $1 AND (role = $2 OR role = $3 OR role = $4 OR role = $5)", team.Id, "CEO", "CTO", "CMO", "CFO")
	if err != nil {
		return
	}
	for rows.Next() {
		teamMember := TeamMember{}
		if err = rows.Scan(&teamMember.Id, &teamMember.Uuid, &teamMember.TeamId, &teamMember.UserId, &teamMember.Role, &teamMember.CreatedAt, &teamMember.Class); err != nil {
			return
		}
		teamMembers = append(teamMembers, teamMember)
	}
	rows.Close()
	return
}

// 根据用户id，检查是否茶团成员；team中是否存在某个teamMember
func GetTeamMemberByTeamIdAndUserId(team_id, user_id int) (teamMember TeamMember, err error) {
	teamMember = TeamMember{}
	err = Db.QueryRow("SELECT id, uuid, team_id, user_id, role, created_at, class FROM team_members WHERE team_id = $1 AND user_id = $2", team_id, user_id).
		Scan(&teamMember.Id, &teamMember.Uuid, &teamMember.TeamId, &teamMember.UserId, &teamMember.Role, &teamMember.CreatedAt, &teamMember.Class)
	return
}

// 根据team_member struct 生成Create()方法
// AWS CodeWhisperer assist in writing
func (teamMember *TeamMember) Create() (err error) {
	statement := `INSERT INTO team_members (uuid, team_id, user_id, role, created_at, class)
	VALUES ($1, $2, $3, $4, $5, $6)`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		Random_UUID(),
		teamMember.TeamId,
		teamMember.UserId,
		teamMember.Role,
		time.Now(),
		teamMember.Class)
	return
}
func (teamMember *TeamMember) CreatedAtDate() string {
	return teamMember.CreatedAt.Format(FMT_DATE_CN)
}

// 查询一个茶团team的CEO，不是founder，是teamMember.Role = “CEO”，返回一个 teamMember TeamMember
// AWS CodeWhisperer assist in writing
func (team *Team) CEO() (teamMember TeamMember, err error) {
	teamMember = TeamMember{}
	err = Db.QueryRow("SELECT id, uuid, team_id, user_id, role, created_at, class FROM team_members WHERE team_id = $1 AND role = $2", team.Id, "CEO").
		Scan(&teamMember.Id, &teamMember.Uuid, &teamMember.TeamId, &teamMember.UserId, &teamMember.Role, &teamMember.CreatedAt, &teamMember.Class)
	return
}

// GetTeamMemberByRole() 根据角色查找茶团成员资料。用于检查茶团拟邀请的新成员角色是否已经被占用
// AWS CodeWhisperer assist in writing
func (team *Team) GetTeamMemberByRole(role string) (teamMember TeamMember, err error) {
	teamMember = TeamMember{}
	err = Db.QueryRow("SELECT id, uuid, team_id, user_id, role, created_at, class FROM team_members WHERE team_id = $1 AND role = $2", team.Id, role).
		Scan(&teamMember.Id, &teamMember.Uuid, &teamMember.TeamId, &teamMember.UserId, &teamMember.Role, &teamMember.CreatedAt, &teamMember.Class)
	return
}

// teamMemberUpdate() 更新茶团成员的角色和属性
func (teamMember *TeamMember) Update() (err error) {
	statement := `UPDATE team_members SET role = $1, class = $2 WHERE id = $3`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(teamMember.Role, teamMember.Class, teamMember.Id)
	return
}

// 更换茶团默认CEO的方法，Update team_members记录中role=CEO的行 user_id 为当前user_id
func (teamMember *TeamMember) UpdateFirstCEO(user_id int) (err error) {
	statement := `UPDATE team_members SET user_id = $1, created_at = $2 WHERE team_id = $3 AND role = $4`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(user_id, time.Now(), teamMember.TeamId, "CEO")
	return
}

// 根据teamMember.teamId获取Team()，返回成员所在team的信息
// AWS CodeWhisperer assist in writing
func (teamMember *TeamMember) Team() (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE id = $1", teamMember.TeamId).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// GetTeamByName()
// AWS CodeWhisperer assist in writing
func GetTeamByName(name string) (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE name = $1", name).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// InvitedTeams()
// 根据ProjectId从LicenceTeam获取[]TeamId,然后用teamId，获取对应的Team，最后返回[]team
// 获取一个封闭式茶台的全部受邀请茶团
func (project *Project) InvitedTeams() (teams []Team, err error) {
	rows, err := Db.Query("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE id IN (SELECT team_id FROM project_invited_teams WHERE project_id = $1)", project.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// GetTeamsByGroupId()
func GetTeamsByGroupId(group_id int) (teams []Team, err error) {
	rows, err := Db.Query("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE group_id = $1", group_id)
	if err != nil {
		return
	}
	for rows.Next() {
		team := Team{}
		if err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId); err != nil {
			return
		}
		teams = append(teams, team)
	}
	rows.Close()
	return
}

// GetTeamByAbbreviationAndGroupId()
func GetTeamByAbbreviationAndGroupId(abbreviation string, group_id int) (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE abbreviation = $1 AND group_id = $2", abbreviation, group_id).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// GetTeamByAbbreviationAndFounderId()
func GetTeamByAbbreviationAndFounderId(abbreviation string, founder_id int) (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE abbreviation = $1 AND founder_id = $2", abbreviation, founder_id).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// GetTeamByAbbreviation()
func GetTeamByAbbreviation(abbreviation string) (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE abbreviation = $1", abbreviation).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// GetGroupFirstTeam() 根据group.first_team_id获取team
func GetGroupFirstTeam(group_id int) (team Team, err error) {
	team = Team{}
	err = Db.QueryRow("SELECT id, uuid, name, mission, founder_id, created_at, class, abbreviation, logo, updated_at, group_id FROM teams WHERE id = (SELECT first_team_id FROM groups WHERE id = $1)", group_id).
		Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission, &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation, &team.Logo, &team.UpdatedAt, &team.GroupId)
	return
}

// 获取开放式团队的数量
func OpenTeamCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM teams WHERE class = 1")
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

// 获取封闭式团队数量
func ClosedTeamCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM teams WHERE class = 2")
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

// 获取团队的属性
// AWS CodeWhisperer assist in writing
func (team *Team) TeamProperty() string {
	return TeamProperty[team.Class]
}

// UpdateClass()
func (team *Team) UpdateClass() (err error) {
	statement := `UPDATE teams SET class = $1 WHERE id = $2`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Class, team.Id)
	return
}

// UpdateLogo()
func (team *Team) UpdateLogo() (err error) {
	statement := `UPDATE teams SET logo = $1 WHERE id = $2`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Logo, team.Id)
	return
}

// UpdateAbbreviation()
func (team *Team) UpdateAbbreviation() (err error) {
	statement := `UPDATE teams SET abbreviation = $1 WHERE id = $2`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Abbreviation, team.Id)
	return
}

// UpdateName()
func (team *Team) UpdateName() (err error) {
	statement := `UPDATE teams SET name = $1 WHERE id = $2`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Name, team.Id)
	return
}

// UpdateMission()
func (team *Team) UpdateMission() (err error) {
	statement := `UPDATE teams SET mission = $1 WHERE id = $2`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(team.Mission, team.Id)
	return
}
