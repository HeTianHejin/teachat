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
	StarCount int         //星标（被收藏）次数
	Cover     string      // 封面
	PageData  PublicPData // 仅用于页面渲染，不保存到数据库
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
type ObjectiveSquareAccessCount struct {
	Id        int
	UserId    int
	Count     int
	CreatedAt time.Time
}

// 记录某个用户访问某个茶话会的次数和访问时间
type ObjectiveAccess struct {
	Id          int
	ObjectiveId int
	UserId      int
	CreatedAt   time.Time
}

// IsEdited()通过比较Objective.CreatedAt和EditAt时间是否相差一秒钟以上，来判断是否编辑过内容为true，返回 bool
func (objective *Objective) IsEdited() bool {
	return objective.CreatedAt.Sub(objective.EditAt) >= time.Second
}

// 创建封闭式茶话会的许可茶团号
func (obLicenseTeam *ObjectiveInvitedTeam) Save() (err error) {
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

// 获取属性为1或者2的茶话会limit个，返回[]Objective,
func GetPublicObjectives(limit int) (objectives []Objective, err error) {
	rows, err := Db.Query("SELECT id, uuid, title, body created_at, user_id, class, edit_at, star_count, cover FROM objectives WHERE class = 1 OR class=2 ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		objective := Objective{}
		if err = rows.Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover); err != nil {
			return
		}
		objectives = append(objectives, objective)
	}
	rows.Close()
	return
}

// 获取objective的属性是开放式还是封闭式，返回string
func (objective *Objective) GetStatus() string {
	return ObStatus[objective.Class]
}

// create a new objective
// 创建一个新茶话会
func (user *User) CreateObjective(title, body, cover string, class int) (objective Objective, err error) {
	statement := "INSERT INTO objectives (uuid, title, body, created_at, user_id, class, edit_at, star_count, cover) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	// use QueryRow to return a row and scan the returned id into the objective struct
	err = stmt.QueryRow(CreateUUID(), title, body, time.Now(), user.Id, class, time.Now(), 0, cover).Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover)
	return
}

// 获取现有的茶话会总数
func GetObjectiveCount() (count int) {
	row := Db.QueryRow("SELECT COUNT(*) FROM objectives")
	row.Scan(&count)
	return
}

// 通过UUID获取一个愿景（茶话会）
func GetObjectiveByUuid(uuid string) (objective Objective, err error) {
	objective = Objective{}
	err = Db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover FROM objectives WHERE uuid = $1", uuid).
		Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover)
	return
}

// 通过id获取一个茶话会objective
func GetObjectiveById(id int) (objective Objective, err error) {
	objective = Objective{}
	err = Db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover FROM objectives WHERE id = $1", id).
		Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover)
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
	objective = Objective{}
	err = Db.QueryRow("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover FROM objectives WHERE id = $1", project.ObjectiveId).
		Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover)
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

// 测试时使用!
// GetAllObjectivesForTest() 获取全部茶话会（愿景），返回 []Objective。
func GetAllObjectivesForTest() (objectives []Objective, err error) {
	rows, err := Db.Query("SELECT id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover FROM objectives ORDER BY created_at DESC")
	if err != nil {
		return
	}
	for rows.Next() {
		objective := Objective{}
		if err = rows.Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover); err != nil {
			return
		}
		objectives = append(objectives, objective)
	}
	rows.Close()
	return
}

// GetObjectiveById() 通过ObjectiveId获取一个愿景（茶话会）

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
	rows, err := Db.Query("SELECT DISTINCT ON(user_id) id, uuid, title, body, created_at, user_id, class, edit_at, star_count, cover FROM objectives ORDER BY user_id, random() LIMIT $1", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		objective := Objective{}
		if err = rows.Scan(&objective.Id, &objective.Uuid, &objective.Title, &objective.Body, &objective.CreatedAt, &objective.UserId, &objective.Class, &objective.EditAt, &objective.StarCount, &objective.Cover); err != nil {
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
