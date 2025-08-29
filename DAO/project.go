package data

import (
	"context"
	"fmt"
	"time"
)

// 茶台 teaTable，寓意为某个愿景下的分类项目，具体事件，活动节点，关卡。
// 规则1：如果class=1, 是普通的开放式茶台，任意用户可以入座品茶聆听，而且可以提出新茶议（主张/议题），属于全员可参与的自由开放式圆桌茶话会。游客可以查看主题和品味跟帖，但是不能品味（跟帖）。
// 规则2：如果class=2，是封闭式茶台，则需要检查访问用户是否台主（创建茶团）指定团体成员，非成员只能旁观但是不能参与品茶活动（不可以提议新主张/跟帖/表态）。
// 类似于某个公开但不是人人均可投票的议程，如奥运会高台跳水比赛，仅有评委成员可以评议，而观众只能观看不能表决；
// 又或者某个歌唱比赛，评委成员可以表态（票决），听众仅能旁听；又或者是某些服务评价案件，仅同行专业人士可以评议，其他人围观。
type Project struct {
	Id          int
	Uuid        string
	Title       string
	Body        string
	ObjectiveId int //所属茶围
	UserId      int //开台人，台主，作者
	CreatedAt   time.Time
	EditAt      *time.Time
	Cover       string //封面图片文件名
	IsPrivate   bool   //公私类型，代表&家庭（family）=true，代表$事业团队（team）=false。默认是false
	TeamId      int    //作者发帖时选择的成员所属茶团id（team）
	FamilyId    int    //作者发帖时选择的成员所属家庭id(family_id)

	// 0: 追加待评草台 (Pending append review)
	// 1: 开放式茶台 (Open tea table)
	// 2: 封闭式茶台 (Closed tea table)
	// 跳过 3-9
	// 10: 开放式草台 (Open straw table)
	// 跳过 11-19
	// 20: 封闭式草台 (Closed straw table)
	// 跳过 21-30
	// 31: 已婉拒开放式茶台 (Rejected open table)
	// 32: 已婉拒封闭式茶台 (Rejected close table)
	Class int

	// 仅用于页面渲染，不保存到数据库
	ActiveData PublicPData
}

const (
	PrClassPendingAppendReview int  = iota // 0: 追加待评草台 (Pending append review)
	PrClassOpen                            // 1: 开放式茶台 (Open tea table)
	PrClassClose                           // 2: 封闭式茶台 (Closed tea table)
	_                                      // 跳过 3-9
	PrClassOpenStraw           = 10        // 10: 开放式草台 (Open straw table)
	_                                      // 跳过 11-19
	PrClassCloseStraw          = 20        // 20: 封闭式草台 (Closed straw table)
	_                                      // 跳过 21-30
	PrClassRejectedOpen        = 31        // 31: 已婉拒开放式茶台 (Rejected open table)
	PrClassRejectedClose       = 32        // 32: 已婉拒封闭式茶台 (Rejected close table)
)

// 同意入围，许可某个茶台（项目）设立
type ProjectApproved struct {
	Id          int
	UserId      int       //批准者id
	ProjectId   int       //茶台id
	ObjectiveId int       //茶围id
	CreatedAt   time.Time //批准时间
}

// project_approved.Create()  ----【tencent AI协助】
func (projectApproved *ProjectApproved) Create() (err error) {
	statement := "INSERT INTO project_approved (user_id, project_id, objective_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(projectApproved.UserId, projectApproved.ProjectId, projectApproved.ObjectiveId, time.Now()).Scan(&projectApproved.Id)
	return
}

// project_approved.GetByProjectId() ----【tencent AI协助】
func (projectApproved *ProjectApproved) GetByObjectiveIdProjectId() (err error) {
	err = db.QueryRow("SELECT id, user_id, project_id, objective_id, created_at FROM project_approved WHERE objective_id=$1 and project_id = $2", projectApproved.ObjectiveId, projectApproved.ProjectId).Scan(&projectApproved.Id, &projectApproved.UserId, &projectApproved.ProjectId, &projectApproved.ObjectiveId, &projectApproved.CreatedAt)
	return
}

// project_approved.GetById() ----【tencent AI协助】
func (projectApproved *ProjectApproved) GetById() (err error) {
	err = db.QueryRow("SELECT id, user_id, project_id, objective_id, created_at FROM project_approved WHERE id = $1", projectApproved.Id).Scan(&projectApproved.Id, &projectApproved.UserId, &projectApproved.ProjectId, &projectApproved.ObjectiveId, &projectApproved.CreatedAt)
	return
}

// project_approved.CountByObjectiveId() 统计某个茶围下许可/批准入围的茶台数量 --【tencent AI协助】
func (projectApproved *ProjectApproved) CountByObjectiveId() (count int) {
	err := db.QueryRow("SELECT COUNT(*) FROM project_approved WHERE objective_id = $1", projectApproved.ObjectiveId).Scan(&count)
	if err != nil {
		return
	}
	return
}

// project *Project.IsApproved() 判断某个茶台是否被批准（立项） --【tencent AI协助】
func (project *Project) IsApproved() (approved bool, err error) {
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM project_approved WHERE project_id = $1)", project.Id).Scan(&approved)
	return
}

// CountProjectByTitleObjectiveId() 统计某个茶围下相同名称的茶台数量
func CountProjectByTitleObjectiveId(title string, objectiveId int) (count int, err error) {
	err = db.QueryRow("SELECT count(*) FROM projects WHERE title = $1 AND objective_id = $2", title, objectiveId).Scan(&count)
	return
}

// 封闭式茶台限定茶团（团队/家庭）集合
type ProjectInvitedTeam struct {
	Id        int
	ProjectId int
	TeamId    int
	CreatedAt time.Time
}

// 茶台（事件/项目）发生地方
type ProjectPlace struct {
	Id        int
	ProjectId int
	PlaceId   int
	CreatedAt time.Time
	UserId    int //创建人
}

// project_place.Create()
func (projectPlace *ProjectPlace) Create() (err error) {
	statement := "INSERT INTO project_place (project_id, place_id, user_id) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(projectPlace.ProjectId, projectPlace.PlaceId, projectPlace.UserId).Scan(&projectPlace.Id)
	return
}

// project_place.GetByProjectId()
func (projectPlace *ProjectPlace) GetByProjectId() (err error) {
	err = db.QueryRow("SELECT id, project_id, place_id, created_at, user_id FROM project_place WHERE project_id = $1 ORDER BY created_at DESC LIMIT 1", projectPlace.ProjectId).Scan(&projectPlace.Id, &projectPlace.ProjectId, &projectPlace.PlaceId, &projectPlace.CreatedAt, &projectPlace.UserId)
	return
}

var PrProperty = map[int]string{
	0:  "追加待评草台",
	1:  "开放式茶台",
	2:  "封闭式茶台",
	10: "开放式草台",
	20: "封闭式草台",
	31: "已婉拒开台",
	32: "已婉拒封台",
}

// IsEdited()
func (project *Project) IsEdited() bool {
	return project.EditAt != nil && !project.EditAt.Equal(project.CreatedAt)
}

// InvitedTeamIds() 获取一个封闭式茶台的全部受邀请茶团id
func (project *Project) InvitedTeamIds() (team_id_slice []int, err error) {
	if project.Class != PrClassClose {
		return nil, fmt.Errorf("project.Class != ClassClosedTeaTable")
	}
	rows, err := db.Query("SELECT team_id FROM project_invited_teams WHERE project_id = $1", project.Id)
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

// 封闭式茶台邀请的团队计数
func (project *Project) InvitedTeamsCount() (count int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM project_invited_teams WHERE project_id = $1", project.Id).Scan(&count)
	return
}

// 获取茶台的属性
func (project *Project) GetStatus() string {
	return PrProperty[project.Class]
}

// farmat the CreatedAt date to display nicely on the screen
// 格式化创建时间
func (project *Project) CreatedAtDate() string {
	return project.CreatedAt.Format(FMT_DATE_CN)
}

// format the EditAt date to display nicely on the screen
// 格式化修改时间
func (project *Project) EditAtDate() string {
	return project.EditAt.Format(FMT_DATE_CN)
}

// Project.Create()  编写postgreSQL语句，插入新纪录，return （err error）
func (project *Project) Create() (err error) {
	statement := "INSERT INTO projects (uuid, title, body, objective_id, user_id, created_at, class, cover, team_id, is_private, family_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), project.Title, project.Body, project.ObjectiveId, project.UserId, time.Now(), project.Class, project.Cover, project.TeamId, project.IsPrivate, project.FamilyId).
		Scan(&project.Id, &project.Uuid)
	return
}

// Project.Get()  编写postgreSQL语句，根据id查询纪录，return （err error）
func (project *Project) Get() (err error) {
	err = db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE id = $1", project.Id).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId, &project.IsPrivate, &project.FamilyId)
	return
}

// 根据project的uuid,从projects表查询获取一个茶台对象信息
// 返回一个茶台对象，如果查询失败，则返回err不为nil
func (project *Project) GetByUuid() (err error) {

	err = db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE uuid = $1", project.Uuid).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId, &project.IsPrivate, &project.FamilyId)
	return
}

// 获取茶议所属的茶台
func (t *Thread) Project() (project Project, err error) {
	project = Project{}
	err = db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE id = $1", t.ProjectId).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId, &project.IsPrivate, &project.FamilyId)
	return
}

// 获取某个post属于哪一个project,
func (post *Post) Project() (project Project, err error) {
	thread, err := post.Thread()
	if err != nil {
		return
	}
	project, err = thread.Project()
	if err != nil {
		return
	}
	return
}

// 获取茶台的茶议总数量
func (project *Project) NumReplies() (count int) {
	rows, err := db.Query("SELECT count(*) FROM threads WHERE project_id = $1", project.Id)
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

// 获取某个ID的茶话会下全部茶台
func (objective *Objective) Projects() (projects []Project, err error) {
	rows, err := db.Query("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE objective_id = $1", objective.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		project := Project{}
		if err = rows.Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId, &project.IsPrivate, &project.FamilyId); err != nil {
			return
		}
		projects = append(projects, project)
	}
	rows.Close()
	return
}

// objective.GetPublicProjects() fetch project.Class=1 or 2,return projects
func (objective *Objective) GetPublicProjects() (projects []Project, err error) {
	rows, err := db.Query("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE objective_id = $1 AND class IN (1, 2)", objective.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		project := Project{}
		if err = rows.Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId, &project.IsPrivate, &project.FamilyId); err != nil {
			return
		}
		projects = append(projects, project)
	}
	rows.Close()
	return
}

// 保存封闭式茶台的邀请茶团号，返回 error
func (project_invited_teams *ProjectInvitedTeam) Create() (err error) {
	// 茶团号是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM project_invited_teams WHERE project_id = $1 AND team_id = $2", project_invited_teams.ProjectId, project_invited_teams.TeamId).Scan(&count)
	if err != nil {
		return
	}
	if count > 0 {
		_, err = db.Exec("DELETE FROM project_invited_teams WHERE project_id = $1 AND team_id = $2", project_invited_teams.ProjectId, project_invited_teams.TeamId)
		if err != nil {
			return
		}
	}
	// 受邀请茶团号保存到数据库
	statement := "INSERT INTO project_invited_teams (project_id, team_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(project_invited_teams.ProjectId, project_invited_teams.TeamId, time.Now()).
		Scan(&project_invited_teams.Id)
	return
}

// 删除一个封闭式茶台的许可茶团号
func (project_invited_teams *ProjectInvitedTeam) Delete() (err error) {
	_, err = db.Exec("DELETE FROM project_invited_teams WHERE project_id = $1 AND team_id = $2", project_invited_teams.ProjectId, project_invited_teams.TeamId)
	return
}

// UpdateClass() 通过project的id,更新茶台的class
func (project *Project) UpdateClass() (err error) {
	_, err = db.Exec("UPDATE projects SET class = $1 WHERE id = $2", project.Class, project.Id)
	return
}

// IsInvitedMember 检查用户是否是封闭式茶台邀请的团队或家庭成员
func (proj *Project) IsInvitedMember(user_id int) (bool, error) {
	if proj.Class != PrClassClose {
		return false, fmt.Errorf("茶台类型不是封闭式茶台,不存在邀请名单")
	}
	team_ids, err := proj.InvitedTeamIds()
	if err != nil {
		return false, fmt.Errorf("读取邀请团队ID失败: %v", err)
	}
	if len(team_ids) == 0 {
		return false, nil // 没有邀请任何团队/家庭
	}

	if !proj.IsPrivate {
		return isUserInAnyTeam(user_id, team_ids)
	}
	return isUserInAnyFamily(user_id, team_ids)
}

// SearchProjectByTitle(keyword) 根据关键字搜索茶台,返回 []Project, error, 限制返回limit数量
func SearchProjectByTitle(keyword string, limit int, ctx context.Context) (projects []Project, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE title LIKE $1 LIMIT $2", "%"+keyword+"%", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		project := Project{}
		if err = rows.Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId, &project.IsPrivate, &project.FamilyId); err != nil {
			return
		}
		projects = append(projects, project)
	}
	rows.Close()
	return
}

// project.PlaceId() 获取茶台的地点place_id，最后一次更新place_id的记录
func (project *Project) PlaceId() (place_id int, err error) {
	err = db.QueryRow("SELECT place_id FROM project_place WHERE project_id = $1", project.Id).Scan(&place_id)
	return
}

// 检查项目的SeeSeek是否已完成
func (p *Project) IsSeeSeekCompleted(ctx context.Context) bool {
	seeSeek, err := GetSeeSeekByProjectId(p.Id, ctx)
	if err != nil {
		return false
	}
	return seeSeek.Status == SeeSeekStatusCompleted
}
