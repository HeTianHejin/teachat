package data

import (
	"errors"
	"time"
)

// 茶台 teaTable，寓意为某个愿景下的分类项目，具体事件，活动节点，关卡。
// 规则1：如果class=1, 是普通的开放式茶台，任意用户可以入座品茶聆听，而且可以提出新茶议（主张/议题），属于全员可参与的自由开放式圆桌茶话会。游客可以查看主题和品味跟帖，但是不能品味（跟帖）。
// 规则2：如果class=2，是封闭式茶台，则需要检查访问用户是否台主（创建人）指定团体成员，非成员只能旁观但是不能参与品茶活动（不可以提议新主张/跟帖/表态）。
// 类似于某个公开但不是人人均可投票的议程，如奥运会高台跳水比赛，仅有评委成员可以评议，而观众只能观看不能表决；
// 又或者某个歌唱比赛，评委成员可以表态（票决），听众仅能旁听；又或者是某些服务评价案件，仅同行专业人士可以评议，其他人围观。
type Project struct {
	Id          int
	Uuid        string
	Title       string
	Body        string
	ObjectiveId int //茶围
	UserId      int //开台人，台主，作者
	CreatedAt   time.Time
	Class       int // 属性 0:  "追加待评草台",1:  "开放式茶台",2:  "封闭式茶台",10: "开放式草台",20: "封闭式草台",31: "已婉拒开台",32: "已婉拒封台",
	EditAt      *time.Time
	Cover       string //封面图片文件名
	TeamId      int    //作者发帖时选择的成员所属茶团id（team）
	IsPrivate   bool   //类型，代表&家庭（family）=true，代表$团队（team）=false。默认是false
	FamilyId    int    //作者发帖时选择的成员所属家庭id(family_id)

	// 仅用于页面渲染，不保存到数据库
	PageData PublicPData
}

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
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(projectApproved.UserId, projectApproved.ProjectId, projectApproved.ObjectiveId, time.Now()).Scan(&projectApproved.Id)
	return
}

// project_approved.GetByProjectId() ----【tencent AI协助】
func (projectApproved *ProjectApproved) GetByProjectId() (err error) {
	err = Db.QueryRow("SELECT id, user_id, project_id, objective_id, created_at FROM project_approved WHERE project_id = $1", projectApproved.ProjectId).Scan(&projectApproved.Id, &projectApproved.UserId, &projectApproved.ProjectId, &projectApproved.ObjectiveId, &projectApproved.CreatedAt)
	return
}

// project_approved.GetById() ----【tencent AI协助】
func (projectApproved *ProjectApproved) GetById() (err error) {
	err = Db.QueryRow("SELECT id, user_id, project_id, objective_id, created_at FROM project_approved WHERE id = $1", projectApproved.Id).Scan(&projectApproved.Id, &projectApproved.UserId, &projectApproved.ProjectId, &projectApproved.ObjectiveId, &projectApproved.CreatedAt)
	return
}

// project_approved.CountByObjectiveId() 统计某个茶围下许可/批准入围的茶台数量 --【tencent AI协助】
func (projectApproved *ProjectApproved) CountByObjectiveId() (count int) {
	err := Db.QueryRow("SELECT COUNT(*) FROM project_approved WHERE objective_id = $1", projectApproved.ObjectiveId).Scan(&count)
	if err != nil {
		return
	}
	return
}

// project *Project.IsApproved() 判断某个茶台是否被批准（立项） --【tencent AI协助】
func (project *Project) IsApproved() (approved bool, err error) {
	err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM project_approved WHERE project_id = $1)", project.Id).Scan(&approved)
	return
}

// CountProjectByTitleObjectiveId() 统计某个茶围下相同名称的茶台数量
func CountProjectByTitleObjectiveId(title string, objectiveId int) (count int, err error) {
	err = Db.QueryRow("SELECT count(*) FROM projects WHERE title = $1 AND objective_id = $2", title, objectiveId).Scan(&count)
	return
}

// 封闭式茶台限定茶团（团队/家庭）集合
type ProjectInvitedTeam struct {
	Id        int
	ProjectId int
	TeamId    int
	CreatedAt time.Time
}

// 茶台（事件）地方
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

// IsEdited()
func (project *Project) IsEdited() bool {
	return project.EditAt != nil && !project.EditAt.Equal(project.CreatedAt)
}

// InvitedTeamIds() 获取一个封闭式茶台的全部受邀请茶团id
func (project *Project) InvitedTeamIds() (team_id_slice []int, err error) {
	rows, err := Db.Query("SELECT team_id FROM project_invited_teams WHERE project_id = $1", project.Id)
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

// Project.Create()  编写postgreSQL语句，插入新纪录，return （err error）
func (project *Project) Create() (err error) {
	statement := "INSERT INTO projects (uuid, title, body, objective_id, user_id, created_at, class, cover, team_id, is_private, family_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
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
	err = Db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE id = $1", project.Id).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId, &project.IsPrivate, &project.FamilyId)
	return
}

// 根据project的uuid,从projects表查询获取一个茶台对象信息
// 返回一个茶台对象，如果查询失败，则返回err不为nil
func (project *Project) GetByUuid() (err error) {

	err = Db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE uuid = $1", project.Uuid).
		Scan(&project.Id, &project.Uuid, &project.Title, &project.Body, &project.ObjectiveId, &project.UserId, &project.CreatedAt, &project.Class, &project.EditAt, &project.Cover, &project.TeamId, &project.IsPrivate, &project.FamilyId)
	return
}

// 获取茶议所属的茶台
func (t *Thread) Project() (project Project, err error) {
	project = Project{}
	err = Db.QueryRow("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE id = $1", t.ProjectId).
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
	rows, err := Db.Query("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE objective_id = $1", objective.Id)
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
	rows, err := Db.Query("SELECT id, uuid, title, body, objective_id, user_id, created_at, class, edit_at, cover, team_id, is_private, family_id FROM projects WHERE objective_id = $1 AND class IN (1, 2)", objective.Id)
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

// 通过id，检查当前用户是否是茶台邀请茶团（$team/&family）成员,
// 是成员的话，返回 true，nil
func (proj *Project) IsInvitedMember(user_id int) (ok bool, err error) {
	count, err := proj.InvitedTeamsCount()
	if err != nil {
		return false, errors.New("this tea-table lost invited any teams to drink tea")
	}
	if count == 0 {
		return false, errors.New("this tea-table  host has not invited any teams to drink")
	}
	team_ids, err := proj.InvitedTeamIds()
	if err != nil {
		return false, errors.New("cannot read project invited team ids")
	}
	if len(team_ids) < 1 {
		return false, errors.New("this objective host has not invited any teams to drink tea")
	}

	if !proj.IsPrivate {
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
		for _, family_id := range team_ids {
			// 迭代team_ids,读取每个家庭的全部成员id
			member_user_ids, err := GetAllMembersUserIdsByFamilyId(family_id)
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
