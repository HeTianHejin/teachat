package data

import "time"

// 茶团邀请函
type Invitation struct {
	Id          int
	Uuid        string
	TeamId      int
	InviteEmail string
	Role        string
	InviteWord  string
	CreatedAt   time.Time
	Status      int //0: "待处理",1: "已查看",2: "已接受",3: "已拒绝",4: "已过期",
	//页面渲染数据,不入库保存
	PageData InvitationDetailPageData
}

// 茶团邀请函答复
type InvitationReply struct {
	Id           int
	Uuid         string
	InvitationId int
	UserId       int
	ReplyWord    string
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
	statement := `INSERT INTO invitations (uuid, team_id, invite_email, role, invite_word, created_at, status)
	VALUES ($1, $2, $3, $4, $5 ,$6 ,$7)`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		CreateUUID(),
		invitation.TeamId,
		invitation.InviteEmail,
		invitation.Role,
		invitation.InviteWord,
		invitation.CreatedAt,
		invitation.Status)
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
	rows, err := Db.Query("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status FROM invitations WHERE team_id = $1", team.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		invitation := Invitation{}
		if err = rows.Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status); err != nil {
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

// CreateAtDate
func (invitation *Invitation) CreatedAtDate() string {
	return invitation.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// GetInvitationByUuid
func GetInvitationByUuid(uuid string) (invitation Invitation, err error) {
	invitation = Invitation{}
	err = Db.QueryRow("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status FROM invitations WHERE uuid = $1", uuid).
		Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status)
	return
}

// GetInvitationById(invitation_id)
func GetInvitationById(id int) (invitation Invitation, err error) {
	invitation = Invitation{}
	err = Db.QueryRow("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status FROM invitations WHERE id = $1", id).
		Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status)
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
		CreateUUID(),
		invitationReply.InvitationId,
		invitationReply.UserId,
		invitationReply.ReplyWord,
		invitationReply.CreatedAt)
	return
}

// check() 如果用户已经查阅了邀请函则返回false，表示无须再处理，否则返回true
func (invitation *Invitation) Check() bool {
	return invitation.Status <= 1
}

// Reply() 查询邀请函的回复
// AWS CodeWhisperer assist in writing
func (invitation *Invitation) Reply() (invitationReply InvitationReply, err error) {
	invitationReply = InvitationReply{}
	err = Db.QueryRow("SELECT id, uuid, invitation_id, user_id, reply_word, created_at FROM invitation_replies WHERE invitation_id = $1", invitation.Id).
		Scan(&invitationReply.Id, &invitationReply.Uuid, &invitationReply.InvitationId, &invitationReply.UserId, &invitationReply.ReplyWord, &invitationReply.CreatedAt)
	return
}

// Invitation.InvitationReply.CreatedAtDate()
func (invitation *Invitation) ReplyCreatedAtDate() string {
	invitationReply, _ := invitation.Reply()
	return invitationReply.CreatedAtDate()
}

// InvitationReply.CreatedAtDate()
func (invitationReply *InvitationReply) CreatedAtDate() string {
	return invitationReply.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// teamMember.UserId.InvitationReply()
func (teamMember *TeamMember) InvitationReply() (invitationReply InvitationReply, err error) {
	invitationReply = InvitationReply{}
	err = Db.QueryRow("SELECT id, uuid, invitation_id, user_id, reply_word, created_at FROM invitation_replies WHERE user_id = $1", teamMember.UserId).
		Scan(&invitationReply.Id, &invitationReply.Uuid, &invitationReply.InvitationId, &invitationReply.UserId, &invitationReply.ReplyWord, &invitationReply.CreatedAt)
	return
}

// teamMember.UserId.InvitationReply().CreatedAtDate()
func (teamMember *TeamMember) InvitationReplyCreatedAtDate() string {
	invitationReply, _ := teamMember.InvitationReply()
	return invitationReply.CreatedAtDate()
}

// 根据invite_email查询一个User收到的全部邀请函
// AWS CodeWhisperer assist in writing
func (user *User) Invitations() (invitations []Invitation, err error) {
	rows, err := Db.Query("SELECT id, uuid, team_id, invite_email, role, invite_word, created_at, status FROM invitations WHERE invite_email = $1", user.Email)
	if err != nil {
		return
	}
	for rows.Next() {
		invitation := Invitation{}
		if err = rows.Scan(&invitation.Id, &invitation.Uuid, &invitation.TeamId, &invitation.InviteEmail, &invitation.Role, &invitation.InviteWord, &invitation.CreatedAt, &invitation.Status); err != nil {
			return
		}
		invitations = append(invitations, invitation)
	}
	rows.Close()
	return
}
