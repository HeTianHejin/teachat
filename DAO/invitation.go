package data

import "time"

// 茶团加盟申请书
type MemberApplication struct {
	Id        int
	Uuid      string
	TeamId    int    //拟加盟的茶团
	UserId    int    //申请人
	Content   string //申请书正文内容
	Status    int    //0: "待处理",1: "已查看",2: "已批准",3: "已婉拒",4: "已过期",
	CreatedAt time.Time
	UpdatedAt time.Time
}

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
	rows, err := Db.Query("SELECT * FROM member_applications WHERE team_id = $1", team_id)
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
	rows, err := Db.Query("SELECT * FROM member_applications WHERE team_id = $1 AND status <= 1", team_id)
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
	err = Db.QueryRow("SELECT COUNT(*) FROM member_applications WHERE team_id = $1 AND status <= 1", team_id).Scan(&count)
	if err != nil {
		return
	}
	return
}

// 检测当前用户是否向指定茶团，已经提交过加盟申请？而且申请书状态为等待处理（Status<=1）
func CheckMemberApplicationByTeamIdAndUserId(team_id int, user_id int) (member_application MemberApplication, err error) {
	err = Db.QueryRow("SELECT * FROM member_applications WHERE team_id = $1 AND user_id = $2 AND status <= 1", team_id, user_id).Scan(&member_application.Id, &member_application.Uuid, &member_application.TeamId, &member_application.UserId, &member_application.Content, &member_application.Status, &member_application.CreatedAt, &member_application.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// 根据user_id，查询用户全部加盟申请书 []MemberApplication，error
func GetMemberApplies(user_id int) (member_application_slice []MemberApplication, err error) {
	rows, err := Db.Query("SELECT * FROM member_applications WHERE user_id = $1", user_id)
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
	rows, err := Db.Query("SELECT team_id FROM member_applications WHERE user_id = $1", user_id)
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
	statement := `INSERT INTO member_applications (uuid, team_id, user_id, content, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		Random_UUID(),
		memberApplication.TeamId,
		memberApplication.UserId,
		memberApplication.Content,
		memberApplication.Status,
		memberApplication.CreatedAt,
		memberApplication.UpdatedAt)
	return
}

// 根据id获取一个加盟申请书
func (memberApplication *MemberApplication) Get() (err error) {
	statement := `SELECT * FROM member_applications WHERE id = $1`
	stmt, err := Db.Prepare(statement)
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
	stmt, err := Db.Prepare(statement)
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
	stmt, err := Db.Prepare(statement)
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
	statement := `INSERT INTO member_application_replies (uuid, member_application_id, team_id, user_id, reply_content, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	stmt, err := Db.Prepare(statement)
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
		memberApplicationReply.CreatedAt,
		memberApplicationReply.UpdatedAt)
	return
}

// GetById() MemberApplicationReply
func (memberApplicationReply *MemberApplicationReply) Get() (err error) {
	statement := `SELECT * FROM member_application_replies WHERE id = $1`
	stmt, err := Db.Prepare(statement)
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
	stmt, err := Db.Prepare(statement)
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

// 茶团邀请函
type Invitation struct {
	Id           int
	Uuid         string
	TeamId       int
	InviteEmail  string //邀请对象的邮箱
	Role         string //拟邀请担任角色
	InviteWord   string //邀请涵内容
	CreatedAt    time.Time
	Status       int //0: "待处理",1: "已查看",2: "已接受",3: "已拒绝",4: "已过期",
	AuthorUserId int //邀请函的撰写者茶友id，就是现任团队CEO
}

// 茶团邀请函答复
type InvitationReply struct {
	Id           int
	Uuid         string
	InvitationId int
	UserId       int    //答复人茶友ID
	ReplyWord    string //答复内容
	CreatedAt    time.Time
}

var InvitationStatus = map[int]string{
	0: "待处理",
	1: "已查看",
	2: "已接受",
	3: "已拒绝",
	4: "已过期",
}

// 根据邀请函的status返回其状态
func (invitation *Invitation) GetStatus() string {
	return InvitationStatus[invitation.Status]
}

// 创建一个邀请函
// AWS CodeWhisperer assist in writing
func (invitation *Invitation) Create() (err error) {
	statement := `INSERT INTO invitations (uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id)
	VALUES ($1, $2, $3, $4, $5 ,$6 ,$7, $8)`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		Random_UUID(),
		invitation.TeamId,
		invitation.InviteEmail,
		invitation.Role,
		invitation.InviteWord,
		invitation.CreatedAt,
		invitation.Status,
		invitation.AuthorUserId)
	return
}

// update Invitation
// AWS CodeWhisperer assist in writing
func (invitation *Invitation) Update() (err error) {
	statement := `UPDATE invitations SET status = $1 WHERE id = $2`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		invitation.Status,
		invitation.Id)
	return
}

// 茶团team发送的全部邀请函-资料
func (team *Team) Invitations() (invitations []Invitation, err error) {
	rows, err := Db.Query("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id FROM invitations WHERE team_id = $1 ORDER BY created_at DESC", team.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		invitation := Invitation{}
		if err = rows.Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status, &invitation.AuthorUserId); err != nil {
			return
		}
		invitations = append(invitations, invitation)
	}
	rows.Close()
	return
}

// 根据invite_email查询一个User收到的全部邀请函
// AWS CodeWhisperer assist in writing
func (user *User) Invitations() (invitations []Invitation, err error) {
	rows, err := Db.Query("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id FROM invitations WHERE invite_email = $1 ORDER BY created_at DESC", user.Email)
	if err != nil {
		return
	}
	for rows.Next() {
		invitation := Invitation{}
		if err = rows.Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status, &invitation.AuthorUserId); err != nil {
			return
		}
		invitations = append(invitations, invitation)
	}
	rows.Close()
	return
}

// NumInvitations() 茶团发出的全部邀请函-数量
func (team *Team) NumInvitations() (count int) {
	rows, _ := Db.Query("SELECT COUNT(*) FROM invitations WHERE team_id = $1", team.Id)
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// Invitation.CreateAtDate()
func (invitation *Invitation) CreatedAtDate() string {
	return invitation.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// GetInvitationByUuid
func GetInvitationByUuid(uuid string) (invitation Invitation, err error) {
	invitation = Invitation{}
	err = Db.QueryRow("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id FROM invitations WHERE uuid = $1", uuid).
		Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status, &invitation.AuthorUserId)
	return
}

// GetInvitationById(invitation_id)
func GetInvitationById(id int) (invitation Invitation, err error) {
	invitation = Invitation{}
	err = Db.QueryRow("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id FROM invitations WHERE id = $1", id).
		Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status, &invitation.AuthorUserId)
	return
}

// 根据InvitationReply struct创建一条邀请函回复记录
// AWS CodeWhisperer assist in writing
func (invitationReply *InvitationReply) Create() (err error) {
	statement := `INSERT INTO invitation_replies (uuid, invitation_id, user_id, reply_word, created_at)
	VALUES ($1, $2, $3, $4, $5)`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		Random_UUID(),
		invitationReply.InvitationId,
		invitationReply.UserId,
		invitationReply.ReplyWord,
		invitationReply.CreatedAt)
	return
}

// Invitation.Check() 如果用户已经查阅了邀请函则返回false，表示无须再处理，否则返回true
func (invitation *Invitation) Check() bool {
	return invitation.Status <= 1
}

// MemberApplication.Check() 如果用户已经查阅了加盟申请书则返回false，表示无须再处理，否则返回true
func (memberApplication *MemberApplication) Check() bool {
	return memberApplication.Status <= 1
}

// MemberApplication.Reply() 查询申请书的审查反馈结果
func (memberApplication *MemberApplication) Reply() (memberApplicationReply MemberApplicationReply, err error) {
	memberApplicationReply = MemberApplicationReply{}
	err = Db.QueryRow("SELECT id, uuid, member_application_id, team_id, user_id, reply_content, status, created_at FROM member_application_replies WHERE member_application_id = $1", memberApplication.Id).
		Scan(&memberApplicationReply.Id, &memberApplicationReply.Uuid, &memberApplicationReply.MemberApplicationId, &memberApplicationReply.TeamId, &memberApplicationReply.UserId, &memberApplicationReply.ReplyContent, &memberApplicationReply.Status, &memberApplicationReply.CreatedAt)
	return
}

// MemberApplication.ReplyCreatedAtDate() 加盟申请书回复创建时间
func (memberApplication *MemberApplication) ReplyCreatedAtDate() string {
	memberApplicationReply, _ := memberApplication.Reply()
	return memberApplicationReply.CreatedAtDate()
}

// Invitation.Reply() 查询邀请函的回复
// AWS CodeWhisperer assist in writing
func (invitation *Invitation) Reply() (invitationReply InvitationReply, err error) {
	invitationReply = InvitationReply{}
	err = Db.QueryRow("SELECT id, uuid, invitation_id, user_id, reply_word, created_at FROM invitation_replies WHERE invitation_id = $1", invitation.Id).
		Scan(&invitationReply.Id, &invitationReply.Uuid, &invitationReply.InvitationId, &invitationReply.UserId, &invitationReply.ReplyWord, &invitationReply.CreatedAt)
	return
}

// Invitation.ReplyCreatedAtDate() 邀请函回复创建时间
func (invitation *Invitation) ReplyCreatedAtDate() string {
	invitationReply, _ := invitation.Reply()
	return invitationReply.CreatedAtDate()
}

// InvitationReply.CreatedAtDate() 邀请函创建时间
func (invitationReply *InvitationReply) CreatedAtDate() string {
	return invitationReply.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// teamMember.UserId.InvitationReply() 根据茶团成员的用户id查询其邀请函回复时间
func (teamMember *TeamMember) InvitationReply() (invitationReply InvitationReply, err error) {
	invitationReply = InvitationReply{}
	err = Db.QueryRow("SELECT id, uuid, invitation_id, user_id, reply_word, created_at FROM invitation_replies WHERE user_id = $1", teamMember.UserId).
		Scan(&invitationReply.Id, &invitationReply.Uuid, &invitationReply.InvitationId, &invitationReply.UserId, &invitationReply.ReplyWord, &invitationReply.CreatedAt)
	return
}

// teamMember.InvitationReply().CreatedAtDate()
func (teamMember *TeamMember) InvitationReplyCreatedAtDate() string {
	invitationReply, _ := teamMember.InvitationReply()
	return invitationReply.CreatedAtDate()
}
