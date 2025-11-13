package data

import "time"

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

const (
	InvitationStatusPending = iota
	InvitationStatusViewed
	InvitationStatusAccepted
	InvitationStatusRejected
	InvitationStatusExpired
)

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
	stmt, err := db.Prepare(statement)
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
		time.Now(),
		invitation.Status,
		invitation.AuthorUserId)
	return
}

// update Invitation
// AWS CodeWhisperer assist in writing
func (invitation *Invitation) UpdateStatus() (err error) {
	statement := `UPDATE invitations SET status = $1 WHERE id = $2`
	stmt, err := db.Prepare(statement)
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
	rows, err := db.Query("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id FROM invitations WHERE team_id = $1 ORDER BY created_at DESC", team.Id)
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
	rows, err := db.Query("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id FROM invitations WHERE invite_email = $1 ORDER BY created_at DESC", user.Email)
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
	rows, _ := db.Query("SELECT COUNT(*) FROM invitations WHERE team_id = $1", team.Id)
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
	err = db.QueryRow("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id FROM invitations WHERE uuid = $1", uuid).
		Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status, &invitation.AuthorUserId)
	return
}

// GetInvitationById(invitation_id)
func GetInvitationById(id int) (invitation Invitation, err error) {
	invitation = Invitation{}
	err = db.QueryRow("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status, author_user_id FROM invitations WHERE id = $1", id).
		Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status, &invitation.AuthorUserId)
	return
}

// 根据InvitationReply struct创建一条邀请函回复记录
// AWS CodeWhisperer assist in writing
func (invitationReply *InvitationReply) Create() (err error) {
	statement := `INSERT INTO invitation_replies (uuid, invitation_id, user_id, reply_word, created_at)
	VALUES ($1, $2, $3, $4, $5)`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		Random_UUID(),
		invitationReply.InvitationId,
		invitationReply.UserId,
		invitationReply.ReplyWord,
		time.Now())
	return
}

// Invitation.Check() 如果用户已经查阅了邀请函则返回false，表示无须再处理，否则返回true
func (invitation *Invitation) Check() bool {
	return invitation.Status <= 1
}

// Invitation.Reply() 查询邀请函的回复
// AWS CodeWhisperer assist in writing
func (invitation *Invitation) Reply() (invitationReply InvitationReply, err error) {
	invitationReply = InvitationReply{}
	err = db.QueryRow("SELECT id, uuid, invitation_id, user_id, reply_word, created_at FROM invitation_replies WHERE invitation_id = $1", invitation.Id).
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
	err = db.QueryRow("SELECT id, uuid, invitation_id, user_id, reply_word, created_at FROM invitation_replies WHERE user_id = $1", teamMember.UserId).
		Scan(&invitationReply.Id, &invitationReply.Uuid, &invitationReply.InvitationId, &invitationReply.UserId, &invitationReply.ReplyWord, &invitationReply.CreatedAt)
	return
}

// teamMember.InvitationReply().CreatedAtDate()
func (teamMember *TeamMember) InvitationReplyCreatedAtDate() string {
	invitationReply, _ := teamMember.InvitationReply()
	return invitationReply.CreatedAtDate()
}
