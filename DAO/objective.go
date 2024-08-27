package data

import (
	"time"
)

// 茶话会 teaParty，活动愿景Vision,最终目标；
// 规则1：如果class=1为开放式，则下面的茶台可以是开放式class=1，也可以是封闭式class=2；
// 规则2：如果class=2为封闭式，则下面的茶台都是封闭式class=2，仅限茶话会创建者指定团队成员可以创建茶议，实际上，由于品味可以被旁观者引用成为拓展茶议，所以封闭式也是相对而言的封闭。
// 开放式茶台是任何注册用户都可以入座创建茶议，封闭式茶台是开台人（台主）指定团队成员可以创建茶议，
// 类似于某个公开但不是人人均可投票的议程，如奥运会高台跳水比赛，仅有评委成员可以直接评议，而观众只能旁观或者说是间接场外引用评议；
// 又或者某个歌唱比赛，评委成员可以表态（票决），听众仅能旁听，又或者是某些服务评价案件，仅同行专业人士可以评议，其他人围观，引用外围议论。
type Objective struct {
	Id        int
	Uuid      string
	Title     string
	Body      string
	CreatedAt time.Time
	UserId    int
	Class     int //属性 0:  "修改待评草围",1:  "开放式茶话会",2:  "封闭式茶话会",10: "开放式草围",20: "封闭式草围",31: "友邻婉拒开围",32: "友邻婉拒闭围",
	EditAt    time.Time
	StarCount int    //星标（被收藏）次数
	Cover     string // 封面
	TeamId    int    //作者团队id

	PageData PublicPData // 仅用于页面渲染，不保存到数据库
}

// objective.Create() Create a new record based on the given objective struct{},return a new objective and error
func (objective *Objective) Create() (err error) {
	statement := "INSERT INTO objectives (uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id,uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), objective.Title, objective.Body, objective.CreatedAt, objective.UserId, objective.Class, objective.EditAt, objective.StarCount, objective.Cover, objective.TeamId).Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover, &objective.TeamId)
	return
}

// objective.Update() Update the given objective struct{} ,add team_id!
func (objective *Objective) Update() (err error) {
	statement := "UPDATE objectives SET title = $1, body = $2, edit_at = $3, star_count = $4, cover = $5, team_id = $6 WHERE id = $7"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(objective.Title, objective.Body, objective.EditAt, objective.StarCount, objective.Cover, objective.TeamId, objective.Id)
	return
}

// objective.Delete() Delete the given objective struct{}
func (objective *Objective) Delete() (err error) {
	statement := "DELETE FROM objectives WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(objective.Id)
	return
}

// objective.GetByUuid() Get the given objective by uuid
func (objective *Objective) GetByUuid() (err error) {
	err = Db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id FROM objectives WHERE uuid = $1", objective.Uuid).
		Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover, &objective.TeamId)
	return
}

// objective.GetByUuid() Get the given objective by id
func (objective *Objective) GetById() (err error) {
	err = Db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id FROM objectives WHERE id = $1", objective.Id).
		Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover, &objective.TeamId)
	return
}

// objective.GetByUserId() Get the given objective by user_id
func (objective *Objective) GetByUserId() (objectives []Objective, err error) {
	rows, err := Db.Query("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id FROM objectives WHERE user_id = $1", objective.UserId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var obj Objective
		err = rows.Scan(&obj.Id, &obj.Uuid, &obj.Title, &obj.Body, &obj.CreatedAt, &obj.UserId, &obj.Class, &obj.EditAt, &obj.StarCount, &obj.Cover, &obj.TeamId)
		if err != nil {
			return
		}
		objectives = append(objectives, obj)
	}
	return
}

// objective.GetByTeamId() Get the given objective by team_id
func GetObjectiveByTeamId(team_id int) (objectives []Objective, err error) {
	rows, err := Db.Query("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id FROM objectives WHERE team_id = $1", team_id)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var obj Objective
		err = rows.Scan(&obj.Id, &obj.Uuid, &obj.Title, &obj.Body, &obj.CreatedAt, &obj.UserId, &obj.Class, &obj.EditAt, &obj.StarCount, &obj.Cover, &obj.TeamId)
		if err != nil {
			return
		}
		objectives = append(objectives, obj)
	}
	return
}

// objective.GetByTitle() Get the given objective by title
func (objective *Objective) GetByTitle() (objectives []Objective, err error) {
	rows, err := Db.Query("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id FROM objectives WHERE title = $1", objective.Title)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var obj Objective
		err = rows.Scan(&obj.Id, &obj.Uuid, &obj.Title, &obj.Body, &obj.CreatedAt, &obj.UserId, &obj.Class, &obj.EditAt, &obj.StarCount, &obj.Cover, &obj.TeamId)
		if err != nil {
			return
		}
		objectives = append(objectives, obj)
	}
	return
}

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

// 封闭式茶话会限定可以品茶的茶团号列表
type ObjectiveInvitedTeam struct {
	Id          int
	ObjectiveId int
	TeamId      int
	CreatedAt   time.Time
}

// 记录某个用户打开茶话会广场页面的次数，以决定展示那些19个未展示过的茶话会用户

// IsEdited()通过比较Objective.CreatedAt和EditAt时间是否相差一秒钟以上，来判断是否编辑过内容为true，返回 bool
func (objective *Objective) IsEdited() bool {
	return objective.CreatedAt.Sub(objective.EditAt) >= time.Second
}

// 创建封闭式茶话会的许可茶团号
func (obLicenseTeam *ObjectiveInvitedTeam) Create() (err error) {
	// 茶团号是否已存在
	var count int
	err = Db.QueryRow("SELECT COUNT(*) FROM objective_invited_teams WHERE objective_id = $1 AND team_id = $2", obLicenseTeam.ObjectiveId, obLicenseTeam.TeamId).Scan(&count)
	if err != nil {
		return
	}
	if count > 0 {
		_, err = Db.Exec("DELETE FROM objective_invited_teams WHERE objective_id = $1 AND team_id = $2", obLicenseTeam.ObjectiveId, obLicenseTeam.TeamId)
		if err != nil {
			return
		}
	}
	// 保存入新的号
	statement := "INSERT INTO objective_invited_teams (objective_id, team_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(obLicenseTeam.ObjectiveId, obLicenseTeam.TeamId, time.Now()).Scan(&obLicenseTeam.Id)
	return
}

// delete一个封闭式茶话会的许可茶团号
func (obLicenseTeam *ObjectiveInvitedTeam) Delete() (err error) {
	statement := "DELETE FROM objective_invited_teams WHERE objective_id = $1 AND team_id = $2"
	_, err = Db.Exec(statement, obLicenseTeam.ObjectiveId, obLicenseTeam.TeamId)
	return
}

// Get class=1 or class=2,limit ，return []Objective,add team_id
func GetPublicObjectives(limit int) (objectives []Objective, err error) {
	rows, err := Db.Query("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id FROM objectives WHERE class = 1 OR class = 2 ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var obj Objective
		err = rows.Scan(&obj.Id, &obj.Uuid, &obj.Title, &obj.Body, &obj.CreatedAt, &obj.UserId, &obj.Class, &obj.EditAt, &obj.StarCount, &obj.Cover, &obj.TeamId)
		if err != nil {
			return
		}
		objectives = append(objectives, obj)
	}
	return
}

// 获取objective的属性是开放式还是封闭式，返回string
func (objective *Objective) GetStatus() string {
	return ObStatus[objective.Class]
}

// 获取现有的茶话会总数
func GetObjectiveCount() (count int) {
	row := Db.QueryRow("SELECT COUNT(*) FROM objectives")
	row.Scan(&count)
	return
}

// 获取茶议所属的茶话会
func (t *Thread) Objective() (objective Objective, err error) {
	proj, err := t.Project()
	if err != nil {
		return
	}
	objective, err = proj.Objective()
	if err != nil {
		return
	}
	return
}

// 获取post属于哪一个objective
func (post *Post) Objective() (objective Objective, err error) {
	thread, err := post.Thread()
	if err != nil {
		return
	}
	project, err := thread.Project()
	if err != nil {
		return
	}
	objective, err = project.Objective()
	if err != nil {
		return
	}
	return
}

// 获取一个茶台的上级目录茶话会
// 根据project的objectiveId,从objectives表查询获取一个茶话会对象信息
func (project *Project) Objective() (objective Objective, err error) {
	_ = Db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id FROM objectives WHERE id = $1", project.ObjectiveId).Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover, &objective.TeamId)
	return
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
	rows, err := Db.Query("SELECT count(*) FROM projects WHERE objective_id = $1", objective.Id)
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

// GetObjectiveIdByUserId（） 通过objectives表中User_id,获取愿景Id，排序的条件先是star_count，然后是created_at
func GetObjectiveIdByUserId(userId int) (objectiveId int) {
	_ = Db.QueryRow("SELECT id FROM objectives WHERE user_id = $1 ORDER BY star_count DESC, created_at DESC", userId).Scan(&objectiveId)
	return
}

// GetNumObjectiveAuthor() 获取全部茶话会（愿景）作者数，根据objectives表中User_id，如果user_id是重复的，只算一个作者
func GetNumObjectiveAuthor() (count int) {
	_ = Db.QueryRow("SELECT count(DISTINCT user_id) FROM objectives").Scan(&count)
	return
}

// GetRandomObjectives() 随机获取limit个用户的objective，过滤重复的user_id，,返回 []Objective, 限量limit个
func GetRandomObjectives(limit int) (objectives []Objective, err error) {
	rows, err := Db.Query("SELECT DISTINCT ON(user_id) id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover, team_id FROM objectives ORDER BY user_id, random() LIMIT $1", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		objective := Objective{}
		if err = rows.Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover, &objective.TeamId); err != nil {
			return
		}
		objectives = append(objectives, objective)
	}
	rows.Close()
	return
}

// InvitedTeamIds() 通过ObjectiveId获取封闭式茶话会的邀请茶团号列表
func (objective *Objective) InvitedTeamIds() (teamIdList []int, err error) {
	rows, err := Db.Query("SELECT team_id FROM objective_invited_teams WHERE objective_id = $1", objective.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var teamId []int
		if err = rows.Scan(&teamId); err != nil {
			return
		}
		teamIdList = append(teamIdList, teamId...)
	}
	rows.Close()
	return
}

// UpdateClass() 通过ObjectiveId更新����的class属性
func (ob *Objective) UpdateClass() (err error) {
	statement := "UPDATE objectives SET class = $1 WHERE id = $2"
	_, err = Db.Exec(statement, ob.Class, ob.Id)
	return
}
