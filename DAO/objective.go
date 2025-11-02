package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// 茶话会，teaParty，活动愿景Vision,目标；
// 别名茶围，objective,声明围炉品茶讨论范围：一个目的/需求？
// 规则1：如果class=1为开放式，则下面的茶台可以是开放式class=1，也可以是封闭式class=2；
// 规则2：如果class=2为封闭式，则下面的茶台都是封闭式class=2，仅限茶话会创建者指定团队成员可以创建茶议，实际上，由于品味可以被旁观者引用成为拓展茶议，所以封闭式也是相对而言的封闭。
// 开放式茶台是任何注册用户都可以入座创建茶议，封闭式茶台是开台人（台主）指定团队成员可以创建茶议，
// 类似于某个公开但不是人人均可投票的议程，如奥运会高台跳水比赛，仅有评委成员可以直接评议，而观众只能旁观或者说是间接场外引用评议；
// 又或者某个歌唱比赛，评委成员可以表态（票决），听众仅能旁听，又或者是某些服务评价案件，仅同行专业人士可以评议，其他人围观，引用外围议论。
type Objective struct {
	Id        int
	Uuid      string
	Title     string //标题
	Body      string //内容，茶话会活动主题，讨论涉及范围说明
	UserId    int    // 茶围发起人，围主，创建人，作者
	Class     int    //属性 0:  "修改待评草围",1:  "开放式茶话会",2:  "封闭式茶话会",10: "开放式草围",20: "封闭式草围",31: "友邻婉拒开围",32: "友邻婉拒闭围",
	FamilyId  int    //作者发帖时选择的家庭id(family_id)
	TeamId    int    //作者创建茶围时选择的茶团id（team_id）,即是管理团队id
	IsPrivate bool   //私有或者公有类型，代表&家庭（family）=true，代表$团队（team）=false。默认是false
	Cover     string // 封面
	CreatedAt time.Time
	EditAt    *time.Time

	// 仅用于页面渲染，不保存到数据库
	ActiveData PublicPData
}

const (
	ObClassPendingModReview    int  = iota // 0: 修改待评草围 (Pending modification review)
	ObClassOpen                            // 1: 开放式茶话会 (Open tea talk)
	ObClassClose                           // 2: 封闭式茶话会 (Closed tea talk)
	_                                      // 跳过 3-9
	ObClassOpenStraw           = 10        // 10: 开放式草围 (Open straw ring)
	_                                      // 跳过 11-19
	ObClassCloseStraw          = 20        // 20: 封闭式草围 (Closed straw ring)
	_                                      // 跳过 21-30
	ObClassNeighborRejectOpen  = 31        // 31: 友邻婉拒开围 (Neighbor rejected opening)
	ObClassNeighborRejectClose = 32        // 32: 友邻婉拒闭围 (Neighbor rejected closing)
)

// 把数字等级属性转换为字符串以显示
var ObStatus = map[int]string{
	0:  "修改待评草围",
	1:  "开放式茶话会",
	2:  "封闭式茶话会",
	10: "开放式草围",
	20: "封闭式草围",
	31: "友邻婉拒开围",
	32: "友邻婉拒闭围",
}

// objective.Create() Create a new record based on the given objective struct{},return a new objective and error
func (objective *Objective) Create() (err error) {
	statement := "INSERT INTO objectives (uuid, title, body, created_at, user_id, class, family_id, cover, team_id, is_private) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id,uuid"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), objective.Title, objective.Body, time.Now(), objective.UserId, objective.Class, objective.FamilyId, objective.Cover, objective.TeamId, objective.IsPrivate).Scan(&objective.Id, &objective.Uuid)
	return
}

// CreateWithTx 使用事务创建茶话会
func (objective *Objective) CreateWithTx(tx *sql.Tx) (err error) {
	statement := "INSERT INTO objectives (uuid, title, body, created_at, user_id, class, family_id, cover, team_id, is_private) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id,uuid"
	err = tx.QueryRow(statement, Random_UUID(), objective.Title, objective.Body, time.Now(), objective.UserId, objective.Class, objective.FamilyId, objective.Cover, objective.TeamId, objective.IsPrivate).Scan(&objective.Id, &objective.Uuid)
	if err != nil {
		return fmt.Errorf("创建茶话会失败: %w", err)
	}
	return nil
}

// objective.Update() Update the given objective struct{}
func (objective *Objective) Update() (err error) {
	statement := "UPDATE objectives SET title = $1, body = $2, edit_at = $3, family_id = $4, cover = $5, team_id = $6, is_private = $7 WHERE id = $8"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(objective.Title, objective.Body, time.Now(), objective.FamilyId, objective.Cover, objective.TeamId, objective.IsPrivate, objective.Id)
	return
}

// objective.UpdateClass() Update the given objective class
func (objective *Objective) UpdateClass() (err error) {
	statement := "UPDATE objectives SET class = $1, edit_at = $2 WHERE id = $3"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(objective.Class, time.Now(), objective.Id)
	return
}

// objective.Delete() Delete the given objective struct{}
// func (objective *Objective) Delete() (err error) {
// 	statement := "DELETE FROM objectives WHERE id = $1"
// 	stmt, err := db.Prepare(statement)
// 	if err != nil {
// 		return
// 	}
// 	defer stmt.Close()
// 	_, err = stmt.Exec(objective.Id)
// 	return
// }

// objective.GetByUuid() Get the given objective by uuid
func (objective *Objective) GetByUuid() (err error) {
	err = db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, family_id, cover, team_id, is_private FROM objectives WHERE uuid = $1", objective.Uuid).
		Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.FamilyId, &objective.Cover, &objective.TeamId, &objective.IsPrivate)
	return
}

// objective.GetByUuid() Get the given objective by id
func (objective *Objective) Get() (err error) {
	err = db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, family_id, cover, team_id, is_private FROM objectives WHERE id = $1", objective.Id).
		Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.FamilyId, &objective.Cover, &objective.TeamId, &objective.IsPrivate)
	return
}

// project.Objective() Get the objective by .objective_id
func (pr *Project) Objective() (objective Objective, err error) {
	err = db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, family_id, cover, team_id, is_private FROM objectives WHERE id = $1", pr.ObjectiveId).
		Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.FamilyId, &objective.Cover, &objective.TeamId, &objective.IsPrivate)
	return
}

// objective.GetByUserId() Get the given objective by user_id
func (objective *Objective) GetByUserId() (objectives []Objective, err error) {
	rows, err := db.Query("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, family_id, cover, team_id, is_private FROM objectives WHERE user_id = $1", objective.UserId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var objective Objective
		err = rows.Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.FamilyId, &objective.Cover, &objective.TeamId, &objective.IsPrivate)
		if err != nil {
			return
		}
		objectives = append(objectives, objective)
	}
	return
}

// 获取objective的属性是开放式还是封闭式，返回string
func (objective *Objective) GetStatus() string {
	return ObStatus[objective.Class]
}

// format the CreatedAt date to display nicely on the screen
// 返回创建茶话会时间的更易于阅读的字符串格式
func (objective *Objective) CreatedAtDate() string {
	return objective.CreatedAt.Format(FMT_DATE_CN)
}

// format the EditAt date to display nicely on the screen
// 返回修改茶话会时间的更易于阅读的字符串格式
func (objective *Objective) EditAtDate() string {
	return objective.EditAt.Format(FMT_DATE_CN)
}

// get the number of projects for this objective
// 获取指定茶话会下的茶台数量
func (objective *Objective) NumReplies() (count int) {
	rows, err := db.Query("SELECT count(*) FROM projects WHERE objective_id = $1", objective.Id)
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

// objective.GetByTitle() Get the given objective by title
func (objective *Objective) GetByTitle() (objectives []Objective, err error) {
	rows, err := db.Query("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, family_id, cover, team_id, is_private FROM objectives WHERE title = $1", objective.Title)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var objective Objective
		err = rows.Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.FamilyId, &objective.Cover, &objective.TeamId, &objective.IsPrivate)
		if err != nil {
			return
		}
		objectives = append(objectives, objective)
	}
	return
}

// objective.CountByTeamId() Count the given objective by team_id
func (objective *Objective) CountByTeamId() (count int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM objectives WHERE team_id = $1", objective.TeamId).Scan(&count)
	return
}

// InvitedTeamsCount() 通过ObjectiveId获取茶话会邀请的茶团数量
func (objective *Objective) InvitedTeamsCount() (count int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM objective_invited_teams WHERE objective_id = $1", objective.Id).Scan(&count)
	return
}

// InvitedTeamIds() 通过ObjectiveId获取封闭式茶话会的邀请茶团号列表
func (objective *Objective) InvitedTeamIds() (team_id_slice []int, err error) {
	rows, err := db.Query("SELECT team_id FROM objective_invited_teams WHERE objective_id = $1", objective.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var team_id int
		if err = rows.Scan(&team_id); err != nil {
			return
		}
		team_id_slice = append(team_id_slice, team_id)
	}
	rows.Close()
	return
}

// 通过id，检查当前用户是否是茶话会邀请茶团（$team/&family）成员,
// 是成员的话，返回 true，nil
func (ob *Objective) IsInvitedMember(user_id int) (ok bool, err error) {
	count, err := ob.InvitedTeamsCount()
	if err != nil {
		return false, err
	}
	// 没有邀请任何team
	if count < 1 {
		return false, errors.New("this objective host has not invited any teams to drink tea")
	}

	team_ids, err := ob.InvitedTeamIds()
	if err != nil {
		return false, err
	}

	if len(team_ids) < 1 {
		return false, errors.New("this objective host has not invited any teams to drink tea")
	}

	if !ob.IsPrivate {
		// 被邀请的对象是$事业团队 []Team.Id
		// 迭代team_ids,用data.GetMemberUserIdsByTeamId()获取全部user_ids；
		// 以UserId == u.Id？检查当前用户是否是茶话会邀请团队成员
		for _, team_id := range team_ids {
			user_ids, _ := GetAllMemberUserIdsByTeamId(team_id)
			for _, u_id := range user_ids {
				if u_id == user_id {
					return true, nil
				}
			}
		}

	} else {
		// 被邀请的对象是&家庭 []Family.Id
		for _, team_id := range team_ids {
			// 迭代team_ids,读取每个家庭的全部成员id
			member_user_ids, err := GetAllMembersUserIdsByFamilyId(team_id)
			if err != nil {
				return false, err
			}
			for _, u_id := range member_user_ids {
				// 检查是否家庭成员
				if u_id == user_id {
					return true, nil
				}
			}
		}
	}

	return
}

// 封闭式茶话会限定可以品茶的茶团号列表
type ObjectiveInvitedTeam struct {
	Id          int
	ObjectiveId int
	TeamId      int
	CreatedAt   time.Time
}

// 记录某个用户打开茶话会广场页面的次数，以决定展示那些19个未展示过的茶话会用户

// IsEdited() .edit_at != nil && 通过比较Objective.CreatedAt和EditAt时间是否相同，来判断是否编辑过内容为true，返回 bool
func (objective *Objective) IsEdited() bool {
	if objective.EditAt == nil {
		return false
	}
	return objective.EditAt.Sub(objective.CreatedAt) > 1*time.Second
}

// 创建封闭式茶话会的许可茶团号
func (obLicenseTeam *ObjectiveInvitedTeam) Create() (err error) {
	statement := "INSERT INTO objective_invited_teams (objective_id, team_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(obLicenseTeam.ObjectiveId, obLicenseTeam.TeamId, time.Now()).Scan(&obLicenseTeam.Id)
	return
}

// CreateWithTx 使用事务创建封闭式茶话会的许可茶团号
func (obLicenseTeam *ObjectiveInvitedTeam) CreateWithTx(tx *sql.Tx) (err error) {
	statement := "INSERT INTO objective_invited_teams (objective_id, team_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	err = tx.QueryRow(statement, obLicenseTeam.ObjectiveId, obLicenseTeam.TeamId, time.Now()).Scan(&obLicenseTeam.Id)
	if err != nil {
		return fmt.Errorf("创建茶话会许可茶团失败: %w", err)
	}
	return nil
}

// delete一个封闭式茶话会的许可茶团号
func (obLicenseTeam *ObjectiveInvitedTeam) Delete() (err error) {
	statement := "DELETE FROM objective_invited_teams WHERE objective_id = $1 AND team_id = $2"
	_, err = db.Exec(statement, obLicenseTeam.ObjectiveId, obLicenseTeam.TeamId)
	return
}

// Get class=1 or class=2,limit ，return []Objective
func GetPublicObjectives(limit int) (objectives []Objective, err error) {
	rows, err := db.Query("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, family_id, cover, team_id, is_private FROM objectives WHERE class IN (1,2) ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var objective Objective
		err = rows.Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.FamilyId, &objective.Cover, &objective.TeamId, &objective.IsPrivate)
		if err != nil {
			return
		}
		objectives = append(objectives, objective)
	}
	return
}

// CreateObjectiveWithTeams 使用事务创建封闭式茶话会及其许可茶团
func CreateObjectiveWithTeams(objective *Objective, teamIDs []int) error {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 开始事务
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("事务启动失败: %w", err)
	}
	defer tx.Rollback()

	// 创建茶话会
	if err = objective.CreateWithTx(tx); err != nil {
		return err
	}

	// 创建许可茶团
	for _, teamID := range teamIDs {
		obInviTeam := ObjectiveInvitedTeam{
			ObjectiveId: objective.Id,
			TeamId:      teamID,
		}
		if err = obInviTeam.CreateWithTx(tx); err != nil {
			return err
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("事务提交失败: %w", err)
	}

	return nil
}

// SearchObjectiveByTitle() 通过标题关键词搜索茶话会
func SearchObjectiveByTitle(keyword string, limit int, ctx context.Context) (objectives []Objective, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT id, uuid, title, body, created_at, user_id, class, edit_at, family_id, cover, team_id, is_private FROM objectives WHERE title ILIKE $1 ORDER BY created_at DESC LIMIT $2", "%"+keyword+"%", limit)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var objective Objective
		err = rows.Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.FamilyId, &objective.Cover, &objective.TeamId, &objective.IsPrivate)
		if err != nil {
			return
		}
		objectives = append(objectives, objective)
	}
	return
}
