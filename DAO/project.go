package data

import (
	"errors"
	"time"
)

// 茶台 teaTable，寓意为某个愿景下的分类项目，具体事件，活动节点，关卡。
// 规则1：如果class=1, 是普通的开放式茶台，全部登录用户可以入座品茶聆听，而且可以提出新茶议（主张/议题），属于全员可参与的自由开放式圆桌茶话会。游客可以查看主题和品味跟帖，但是不能品味（跟帖）。
// 规则2：如果class=2，是封闭式茶台，则需要检查访问用户是否台主（创建人）指定团体成员，非成员只能旁观但是不能参与品茶活动（不可以提议新主张/跟帖/表态）。
// 类似于某个公开但不是人人均可投票的议程，如奥运会高台跳水比赛，仅有评委成员可以评议，而观众只能观看不能表决；
// 又或者某个歌唱比赛，评委成员可以表态（票决），听众仅能旁听；又或者是某些服务评价案件，仅同行专业人士可以评议，其他人围观。
type Project struct {
	Id          int
	Uuid        string
	Title       string
	Body        string
	ObjectiveId int
	UserId      int
	CreatedAt   time.Time
	Class       int // 属性 0:  "追加待评草台",1:  "开放式茶台",2:  "封闭式茶台",10: "开放式草台",20: "封闭式草台",31: "已婉拒开台",32: "已婉拒封台",
	EditAt      time.Time
	Cover       string
	TeamId      int //作者支持团队id
	// 仅用于页面渲染，不保存到数据库
	PageData PublicPData
}

// 封闭式茶台限定茶团（团队）集合
type ProjectInvitedTeam struct {
	Id        int
	ProjectId int
	TeamId    int
	CreatedAt time.Time
}

// 茶台（事件）地点
type ProjectPlace struct {
	Id        int
	ProjectId int
	PlaceId   int
	CreatedAt time.Time
}

// project_place.Create()
func (projectPlace *ProjectPlace) Create() (err error) {
	statement := "INSERT INTO project_place (project_id, place_id) VALUES ($1, $2) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(projectPlace.ProjectId, projectPlace.PlaceId).Scan(&projectPlace.Id)
	return
}

// project_place.GetByProjectId()
func (projectPlace *ProjectPlace) GetByProjectId() (err error) {
	err = Db.QueryRow("SELECT id, project_id, place_id, created_at FROM project_place WHERE project_id = $1", projectPlace.ProjectId).Scan(&projectPlace.Id, &projectPlace.ProjectId, &projectPlace.PlaceId, &projectPlace.CreatedAt)
	return
}

// project.Place() 根据project_id，从project_place表中获取place_id,然后根据place_id,从places表中获取place对象
func (project *Project) Place() (place Place, err error) {
	projectPlace := ProjectPlace{ProjectId: project.Id}
	if err = projectPlace.GetByProjectId(); err != nil {
		return
	}
	place = Place{Id: projectPlace.PlaceId}
	if err = place.Get(); err != nil {
		return
	}
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

// IsEdited() 通过比较Objective.CreatedAt和EditAt时间是否相差一秒钟以上，来判断是否编辑过内容，是为true，返回 bool
func (project *Project) IsEdited() bool {
	return project.CreatedAt.Sub(project.EditAt) >= time.Second
}

// InvitedTeamIds() 获取一个封闭式茶台的全部受邀请茶团id
func (project *Project) InvitedTeamIds() (teamIdList []int, err error) {
	rows, err := Db.Query("SELECT team_id FROM project_invited_teams WHERE project_id = $1", project.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var team_id int
		if err = rows.Scan(&team_id); err != nil {
			return
		}
		teamIdList = append(teamIdList, team_id)
	}
	rows.Close()
	return
}

// GetMemberUserIdsByTeamId() 从TeamMember获取全部茶团成员Userid
func GetMemberUserIdsByTeamId(team_id int) (user_ids []int, err error) {
	rows, err := Db.Query("SELECT user_id FROM team_members WHERE team_id = $1", team_id)
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

// 封闭式茶台邀请的团队计数
func (project *Project) InvitedTeamsCount() (count int, err error) {
	err = Db.QueryRow("SELECT COUNT(*) FROM project_invited_teams WHERE project_id = $1", project.Id).Scan(&count)
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

// 用户在某个茶话会内创建新的茶台
func (user *User) CreateProject(title, body string, objectiveId int, class, team_id int) (project Project, err error) {
	statement := "INSERT INTO projects (uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), title, body, objectiveId, user.Id, time.Now(), class, time.Now(), "default-pr-cover", team_id).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId)
	return
}

// 根据project的id,从projects表查询获取一个茶台对象信息
// 返回一个茶台对象，如果查询失败，则返回err不为nil
func (project *Project) GetById() (err error) {
	err = Db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id FROM projects WHERE id = $1", project.Id).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId)
	return
}

// 根据project的uuid,从projects表查询获取一个茶台对象信息
// 返回一个茶台对象，如果查询失败，则返回err不为nil
func GetProjectByUuid(uuid string) (project Project, err error) {
	project = Project{}
	err = Db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id FROM projects WHERE uuid = $1", uuid).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId)
	return
}

// 获取茶议所属的茶台
func (t *Thread) Project() (project Project, err error) {
	project = Project{}
	err = Db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id FROM projects WHERE id = $1", t.ProjectId).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId)
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
	rows, err := Db.Query("SELECT count(*) FROM threads WHERE project_id = $1", project.Id)
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
	rows, err := Db.Query("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id FROM projects WHERE objective_id = $1", objective.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		project := Project{}
		if err = rows.Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId); err != nil {
			return
		}
		projects = append(projects, project)
	}
	rows.Close()
	return
}

// objective.GetPublicProjects() fetch project.Class=1 or 2,return projects
func (objective *Objective) GetPublicProjects() (projects []Project, err error) {
	rows, err := Db.Query("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id FROM projects WHERE objective_id = $1 AND class IN (1, 2)", objective.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		project := Project{}
		if err = rows.Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId); err != nil {
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
	err = Db.QueryRow("SELECT COUNT(*) FROM project_invited_teams WHERE project_id = $1 AND team_id = $2", project_invited_teams.ProjectId, project_invited_teams.TeamId).Scan(&count)
	if err != nil {
		return
	}
	if count > 0 {
		_, err = Db.Exec("DELETE FROM project_invited_teams WHERE project_id = $1 AND team_id = $2", project_invited_teams.ProjectId, project_invited_teams.TeamId)
		if err != nil {
			return
		}
	}
	// 受邀请茶团号保存到数据库
	statement := "INSERT INTO project_invited_teams (project_id, team_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := Db.Prepare(statement)
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
	_, err = Db.Exec("DELETE FROM project_invited_teams WHERE project_id = $1 AND team_id = $2", project_invited_teams.ProjectId, project_invited_teams.TeamId)
	return
}

// UpdateClass() 通过project的id,更新��台的class
func (project *Project) UpdateClass() (err error) {
	_, err = Db.Exec("UPDATE projects SET class = $1 WHERE id = $2", project.Class, project.Id)
	return
}

// 检查当前会话用户是否茶台邀请团队成员
func (proj *Project) IsInvitedMember(user_id int) (ok bool, err error) {
	count, err := proj.InvitedTeamsCount()
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, errors.New("this tea-table  host has not invited any teams to drink tea")
	}
	teamIDs, err := proj.InvitedTeamIds()
	if err != nil {
		return false, errors.New("cannot read project invited team ids")
	}
	for _, teamID := range teamIDs {
		userIDs, err := GetMemberUserIdsByTeamId(teamID)
		if err != nil {
			return false, err
		}

		for _, userID := range userIDs {
			if userID == user_id {
				return true, nil
			}
		}
	}
	return
}
