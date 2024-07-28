package data

import "time"

// 友邻盲评消息
type AcceptMessage struct {
	Id             int
	FromUserId     int // 发送者
	ToUserId       int // 受邀请的用户id
	Title          string
	Content        string
	AcceptObjectId int // 盲评接纳对象id
	Class          int // 状态： 0未读，1已读,
	CreatedAt      time.Time
	UpdatedAt      time.Time

	//页面动态数据，不存储到数据库
	//PageData AcceptMessagePageData
}

var AcceptMessageStatus = map[int]string{
	0: "未处理",
	1: "已处理",
}

// 获取acceptedMessage状态
func (a *AcceptMessage) Status() string {
	return AcceptMessageStatus[a.Class]
}

// CreatedAtDate() ��式化时间
func (a *AcceptMessage) CreatedAtDate() string {
	return a.CreatedAt.Format(FMT_DATE_CN)
}

// Invitee() 友邻盲评 受邀请者
func (a *AcceptMessage) Invitee() (user User, err error) {
	user, err = GetUserById(a.ToUserId)
	return
}

// Create() 创建邻桌盲评消息
func (a *AcceptMessage) Create() (err error) {
	statement := "INSERT INTO accept_messages (from_user_id, to_user_id, title, content, accept_object_id, class, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(&a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, time.Now(), time.Now()).Scan(&a.Id)
	return
}

// GetAccMesByUIdAndAOId() 根据接收用户ID和接纳对象id，获取邻桌盲评消息
func (a *AcceptMessage) GetAccMesByUIdAndAOId(user_id, accept_object_id int) (err error) {
	err = Db.QueryRow("SELECT * FROM accept_messages WHERE to_user_id = $1 AND accept_object_id = $2", user_id, accept_object_id).Scan(&a.Id, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, &a.CreatedAt, &a.UpdatedAt)
	return
}

// 根据ToUserId，获取受邀请用户全部的acceptMessage
func (u *User) AcceptMessages() (acceptMessages []AcceptMessage, err error) {
	rows, err := Db.Query("SELECT * FROM accept_messages where to_user_id = $1", u.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var a AcceptMessage
		if err = rows.Scan(&a.Id, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return
		}
		acceptMessages = append(acceptMessages, a)
	}
	rows.Close()
	return
}

// 根据ToUserId,获取用户全部未读的class=0的acceptMessage
func (u *User) UnreadAcceptMessages() (acceptMessages []AcceptMessage, err error) {
	rows, err := Db.Query("SELECT * FROM accept_messages where to_user_id = $1 and class = 0", u.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var a AcceptMessage
		if err = rows.Scan(&a.Id, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return
		}
		acceptMessages = append(acceptMessages, a)
	}
	rows.Close()
	return
}

// 根据UserId,获取全部未读的class=0的acceptMessage的统计数量
func (u *User) UnreadAcceptMessagesCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM accept_messages where to_user_id = $1 and class = 0", u.Id)
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

// ��据ToUserId,获取全部已读的class=1的acceptMessage的统计数量
func (u *User) ReadAcceptMessagesCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM accept_messages where to_user_id = $1 and class = 1", u.Id)
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

// 根据ToUserId，统计全部acceptMessage的数量
func (u *User) AllAcceptMessageCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM accept_messages where to_user_id = $1", u.Id)
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

// ��据ToUserId,获取全部已读的class=1的acceptMessage
func (u *User) ReadAcceptMessages() (acceptMessages []AcceptMessage, err error) {
	rows, err := Db.Query("SELECT * FROM accept_messages where to_user_id = $1 and class = 1", u.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var a AcceptMessage
		if err = rows.Scan(&a.Id, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return
		}
		acceptMessages = append(acceptMessages, a)
	}
	rows.Close()
	return
}

// Update() 根据ToUserId和接纳对象id，更新用户的邻桌新茶盲评消息为已读
func (a *AcceptMessage) Update(to_user_id, accept_object_id int) (err error) {
	statement := "UPDATE accept_messages SET class = 1, updated_at = $3 WHERE to_user_id = $1 and accept_object_id = $2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(to_user_id, accept_object_id, time.Now())
	return
}

// Send(user_ids) 向用户ID队列发送新茶盲评请求消息
func (a *AcceptMessage) Send(user_ids []int) (err error) {
	for _, user_id := range user_ids {
		a.ToUserId = user_id
		if err = a.Create(); err != nil {
			return err
		}
	}
	return nil
}

// AcceptMessageCount() 获取用户 新茶评审消息总数
func (user *User) AcceptMessageCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM accept_messages where to_user_id = $1", user.Id)
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

// NewAcceptMessagesCount() 获取用户 新茶评审消息数
func (user *User) NewAcceptMessagesCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM accept_messages where to_user_id = $1 and class = 0", user.Id)
	if err != nil {
		return 0
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return 0
		}
	}
	rows.Close()
	return
}

// 用户是否有新茶评审未读消息？
func (user *User) HasNewAcceptMessage() bool {
	count := user.NewAcceptMessagesCount()
	return count > 0
}

// 检查当前用户sUserId是否受邀请审茶，class = 0
func (user *User) CheckHasAcceptMessage(accept_object_id int) bool {
	rows, err := Db.Query("SELECT count(*) FROM accept_messages where to_user_id = $1 and accept_object_id = $2 and class = 0", user.Id, accept_object_id)
	if err != nil {
		return false
	}
	for rows.Next() {
		var count int
		if err = rows.Scan(&count); err != nil {
			return false
		}
		if count > 0 {
			return true
		}
	}
	rows.Close()
	return false
}

// 检查当前用户sUserId是否曾经受邀请审茶，class = 1
func (user *User) CheckHasReadAcceptMessage(accept_object_id int) bool {
	rows, err := Db.Query("SELECT count(*) FROM accept_messages where to_user_id = $1 and accept_object_id = $2 and class = 1", user.Id, accept_object_id)
	if err != nil {
		return false
	}
	for rows.Next() {
		var count int
		if err = rows.Scan(&count); err != nil {
			return false
		}
		if count > 0 {
			return true
		}
	}
	rows.Close()
	return false
}
