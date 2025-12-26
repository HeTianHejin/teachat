package dao

import "time"

// 个人加盟茶团申请书
type MemberApplication struct {
	Id        int
	Uuid      string
	TeamId    int    //拟加盟的茶团
	UserId    int    //申请人
	Content   string //申请书正文内容
	Status    int    //0: "待处理",1: "已查看",2: "已批准",3: "已婉拒",4: "已过期",
	CreatedAt time.Time
	UpdatedAt *time.Time
}

const (
	// 申请书状态
	MemberApplicationStatusPending  = iota //0: "待处理"
	MemberApplicationStatusViewed          // 1: "已查看"
	MemberApplicationStatusApproved        //2: "已批准"
	MemberApplicationStatusRejected        //3: "已婉拒"
	MemberApplicationStatusExpired         //4: "已过期"
)

// 申请书状态map
var MemberApplicationStatus = map[int]string{
	0: "待处理",
	1: "已查看",
	2: "已批准",
	3: "已婉拒",
	4: "已过期",
}

// 根据team_id,查询全部加盟申请书，[]MemberApplication，error
func GetMemberApplicationByTeamId(team_id int) (member_application_slice []MemberApplication, err error) {
	rows, err := DB.Query("SELECT * FROM member_applications WHERE team_id = $1", team_id)
	if err != nil {
		return
	}
	for rows.Next() {
		var memberApplication MemberApplication
		err = rows.Scan(&memberApplication.Id, &memberApplication.Uuid, &memberApplication.TeamId, &memberApplication.UserId, &memberApplication.Content, &memberApplication.Status, &memberApplication.CreatedAt, &memberApplication.UpdatedAt)
		if err != nil {
			return
		}
		member_application_slice = append(member_application_slice, memberApplication)
	}
	rows.Close()
	return
}

// 根据team_id,status <= 1, 查询全部加盟申请书中需要处理的申请书,[]MemberApplication，error
func GetMemberApplicationByTeamIdAndStatus(team_id int) (member_application_slice []MemberApplication, err error) {
	rows, err := DB.Query("SELECT * FROM member_applications WHERE team_id = $1 AND status <= 1", team_id)
	if err != nil {
		return
	}
	for rows.Next() {
		var memberApplication MemberApplication
		err = rows.Scan(&memberApplication.Id, &memberApplication.Uuid, &memberApplication.TeamId, &memberApplication.UserId, &memberApplication.Content, &memberApplication.Status, &memberApplication.CreatedAt, &memberApplication.UpdatedAt)
		if err != nil {
			return
		}
		member_application_slice = append(member_application_slice, memberApplication)
	}
	rows.Close()
	return
}

// 返回某个茶团待处理申请书数量
func GetMemberApplicationByTeamIdAndStatusCount(team_id int) (count int, err error) {
	err = DB.QueryRow("SELECT COUNT(*) FROM member_applications WHERE team_id = $1 AND status <= 1", team_id).Scan(&count)
	if err != nil {
		return
	}
	return
}

// 检测当前用户是否向指定茶团，已经提交过加盟申请？而且申请书状态为<=指定状态
func CheckMemberApplicationByTeamIdAndUserId(team_id int, user_id int, status int) (member_application MemberApplication, err error) {
	err = DB.QueryRow("SELECT * FROM member_applications WHERE team_id = $1 AND user_id = $2 AND status <= $3", team_id, user_id, status).Scan(&member_application.Id, &member_application.Uuid, &member_application.TeamId, &member_application.UserId, &member_application.Content, &member_application.Status, &member_application.CreatedAt, &member_application.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// 根据user_id，查询用户全部加盟申请书 []MemberApplication，error
func GetMemberApplies(user_id int) (member_application_slice []MemberApplication, err error) {
	rows, err := DB.Query("SELECT * FROM member_applications WHERE user_id = $1", user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		var memberApplication MemberApplication
		err = rows.Scan(&memberApplication.Id, &memberApplication.Uuid, &memberApplication.TeamId, &memberApplication.UserId, &memberApplication.Content, &memberApplication.Status, &memberApplication.CreatedAt, &memberApplication.UpdatedAt)
		if err != nil {
			return
		}
		member_application_slice = append(member_application_slice, memberApplication)
	}
	rows.Close()
	return
}

// 根据UserId,查询全部申请书Id，获取 []teamId
func GetApplyTeamIdsByUserId(user_id int) (teamIds []int, err error) {
	rows, err := DB.Query("SELECT team_id FROM member_applications WHERE user_id = $1", user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		var teamId int
		err = rows.Scan(&teamId)
		if err != nil {
			return
		}
		teamIds = append(teamIds, teamId)
	}
	rows.Close()
	return
}

// 加盟申请书答复
type MemberApplicationReply struct {
	Id                  int
	Uuid                string
	MemberApplicationId int
	TeamId              int
	UserId              int
	ReplyContent        string
	Status              int //0: "待处理",1: "已查看",2: "已批准",3: "已婉拒",4: "已过期",
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// 加盟申请书答复map
var MemberApplicationReplyStatus = map[int]string{
	0: "待处理",
	1: "已查看",
	2: "已批准",
	3: "已婉拒",
	4: "已过期",
}

// 根据加盟申请书的status值，返回其状态
func (memberApplication *MemberApplication) GetStatus() string {
	return MemberApplicationStatus[memberApplication.Status]
}

// MemberApplication.Check() 如果用户已经查阅了加盟申请书则返回false，表示无须再处理，否则返回true
func (memberApplication *MemberApplication) Check() bool {
	return memberApplication.Status <= 1
}

// MemberApplication.Reply() 查询申请书的审查反馈结果
func (memberApplication *MemberApplication) Reply() (memberApplicationReply MemberApplicationReply, err error) {
	memberApplicationReply = MemberApplicationReply{}
	err = DB.QueryRow("SELECT id, uuid, member_application_id, team_id, user_id, reply_content, status, created_at FROM member_application_replies WHERE member_application_id = $1", memberApplication.Id).
		Scan(&memberApplicationReply.Id, &memberApplicationReply.Uuid, &memberApplicationReply.MemberApplicationId, &memberApplicationReply.TeamId, &memberApplicationReply.UserId, &memberApplicationReply.ReplyContent, &memberApplicationReply.Status, &memberApplicationReply.CreatedAt)
	return
}

// MemberApplication.ReplyCreatedAtDate() 加盟申请书回复创建时间
func (memberApplication *MemberApplication) ReplyCreatedAtDate() string {
	memberApplicationReply, _ := memberApplication.Reply()
	return memberApplicationReply.CreatedAtDate()
}

// 根据加盟申请书的答复status返回其状态
func (memberApplicationReply *MemberApplicationReply) GetStatus() string {
	return MemberApplicationReplyStatus[memberApplicationReply.Status]
}

// MemberApplication.CreateAtDate()
func (memberApplication *MemberApplication) CreatedAtDate() string {
	return memberApplication.CreatedAt.Format("2006-01-02 15:04:05")
}

// MemberApplicationReply.CreateAtDate()
func (memberApplicationReply *MemberApplicationReply) CreatedAtDate() string {
	return memberApplicationReply.CreatedAt.Format("2006-01-02 15:04:05")
}

// 创建一个加盟申请书
// AWS CodeWhisperer assist in writing
func (memberApplication *MemberApplication) Create() (err error) {
	statement := `INSERT INTO member_applications (uuid, team_id, user_id, content, status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(Random_UUID(),
		memberApplication.TeamId,
		memberApplication.UserId,
		memberApplication.Content,
		memberApplication.Status,
		time.Now()).Scan(&memberApplication.Id, &memberApplication.Uuid)
	return
}

// 根据id获取一个加盟申请书
func (memberApplication *MemberApplication) Get() (err error) {
	statement := `SELECT * FROM member_applications WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(memberApplication.Id).Scan(
		&memberApplication.Id,
		&memberApplication.Uuid,
		&memberApplication.TeamId,
		&memberApplication.UserId,
		&memberApplication.Content,
		&memberApplication.Status,
		&memberApplication.CreatedAt,
		&memberApplication.UpdatedAt)
	return
}

// MemberApplication.GetByUuid()
func (memberApplication *MemberApplication) GetByUuid() (err error) {
	statement := `SELECT * FROM member_applications WHERE uuid = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(memberApplication.Uuid).Scan(
		&memberApplication.Id,
		&memberApplication.Uuid,
		&memberApplication.TeamId,
		&memberApplication.UserId,
		&memberApplication.Content,
		&memberApplication.Status,
		&memberApplication.CreatedAt,
		&memberApplication.UpdatedAt)
	return
}

// update status MemberApplication
// AWS CodeWhisperer assist in writing
func (memberApplication *MemberApplication) Update() (err error) {
	statement := `UPDATE member_applications SET status = $1, updated_at = $2 WHERE id = $3`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		memberApplication.Status,
		memberApplication.UpdatedAt,
		memberApplication.Id,
	)
	return
}

// 创建一个加盟申请书答复
// AWS CodeWhisperer assist in writing
func (memberApplicationReply *MemberApplicationReply) Create() (err error) {
	statement := `INSERT INTO member_application_replies (uuid, member_application_id, team_id, user_id, reply_content, status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		Random_UUID(),
		memberApplicationReply.MemberApplicationId,
		memberApplicationReply.TeamId,
		memberApplicationReply.UserId,
		memberApplicationReply.ReplyContent,
		memberApplicationReply.Status,
		time.Now())
	return
}

// GetById() MemberApplicationReply
func (memberApplicationReply *MemberApplicationReply) Get() (err error) {
	statement := `SELECT * FROM member_application_replies WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(memberApplicationReply.Id).Scan(
		&memberApplicationReply.Id,
		&memberApplicationReply.Uuid,
		&memberApplicationReply.TeamId,
		&memberApplicationReply.UserId,
		&memberApplicationReply.ReplyContent,
		&memberApplicationReply.Status,
		&memberApplicationReply.CreatedAt,
		&memberApplicationReply.UpdatedAt)
	return
}

// update MemberApplicationReply
// AWS CodeWhisperer assist in writing
func (memberApplicationReply *MemberApplicationReply) Update() (err error) {
	statement := `UPDATE member_application_replies SET status = $1, updated_at = $2 WHERE id = $3`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		memberApplicationReply.Status,
		memberApplicationReply.UpdatedAt,
		memberApplicationReply.Id)
	return
}
