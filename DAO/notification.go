package dao

import (
	"context"
	"errors"
	"strconv"
	"time"
)

// 友邻蒙评通知
type AcceptNotification struct {
	Id             int
	Uuid           string
	FromUserId     int // 发送者
	ToUserId       int // 受邀请的用户id
	Title          string
	Content        string
	AcceptObjectId int // 蒙评接纳对象id
	Class          int // 状态： 0未读，1已读,
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

const (
	NotificationStatusUnread = 0
	NotificationStatusRead   = 1
)

var AcceptNotificationStatus = map[int]string{
	NotificationStatusUnread: "未读",
	NotificationStatusRead:   "已读",
}

// 获取AcceptNotification状态
func (a *AcceptNotification) Status() string {
	return AcceptNotificationStatus[a.Class]
}

// CreatedAtDate() 格式化时间
func (a *AcceptNotification) CreatedAtDate() string {
	return a.CreatedAt.Format(FMT_DATE_CN)
}

// Invitee() 友邻蒙评 受邀请者
func (a *AcceptNotification) Invitee() (user User, err error) {
	user, err = GetUser(a.ToUserId)
	return
}

// Create() 创建邻桌蒙评消息
func (a *AcceptNotification) Create() (err error) {
	return a.CreateWithContext(context.Background())
}

// CreateWithContext 创建邻桌蒙评消息（带context版本）
func (a *AcceptNotification) CreateWithContext(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	statement := "INSERT INTO accept_notifications (from_user_id, to_user_id, title, content, accept_object_id, class, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, time.Now()).Scan(&a.Id)
	return
}

// GetAccNotiByUIdAndAOId() 根据接收用户ID和接纳对象id，获取邻桌蒙评消息
func (a *AcceptNotification) GetAccNotiByUIdAndAOId(user_id, accept_object_id int) (err error) {
	return a.GetAccNotiByUIdAndAOIdWithContext(user_id, accept_object_id, context.Background())
}

// GetAccNotiByUIdAndAOIdWithContext 根据接收用户ID和接纳对象id，获取邻桌蒙评消息（带context版本）
func (a *AcceptNotification) GetAccNotiByUIdAndAOIdWithContext(user_id, accept_object_id int, ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	err = DB.QueryRowContext(ctx, "SELECT * FROM accept_notifications WHERE to_user_id = $1 AND accept_object_id = $2", user_id, accept_object_id).Scan(&a.Id, &a.Uuid, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, &a.CreatedAt, &a.UpdatedAt)
	return
}

// 根据ToUserId，获取受邀请用户全部的acceptNotification
func (u *User) AcceptNotifications() (acceptNotifications []AcceptNotification, err error) {
	rows, err := DB.Query("SELECT * FROM accept_notifications where to_user_id = $1", u.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var a AcceptNotification
		if err = rows.Scan(&a.Id, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return
		}
		acceptNotifications = append(acceptNotifications, a)
	}
	rows.Close()
	return
}

// 根据ToUserId,获取用户全部未读的class=0的acceptNotification
func (u *User) UnreadAcceptNotifications() (acceptNotifications []AcceptNotification, err error) {
	rows, err := DB.Query("SELECT * FROM accept_notifications where to_user_id = $1 and class = $2", u.Id, NotificationStatusUnread)
	if err != nil {
		return
	}
	for rows.Next() {
		var a AcceptNotification
		if err = rows.Scan(&a.Id, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return
		}
		acceptNotifications = append(acceptNotifications, a)
	}
	rows.Close()
	return
}

// 根据UserId,获取全部未读的class=0的acceptNotification的统计数量
func (u *User) UnreadAcceptNotificationsCount() (count int) {
	return u.getAcceptNotificationCountWithContext("class = "+strconv.Itoa(NotificationStatusUnread), context.Background())
}

// 根据ToUserId,获取全部已读的class=1的acceptNotification的统计数量
func (u *User) ReadAcceptNotificationsCount() (count int) {
	return u.getAcceptNotificationCountWithContext("class = "+strconv.Itoa(NotificationStatusRead), context.Background())
}

// 根据ToUserId，统计全部acceptNotification的数量
func (u *User) AllAcceptNotificationCount() (count int) {
	return u.getAcceptNotificationCountWithContext("", context.Background())
}

// 根据ToUserId,获取全部已读的class=1的acceptNotification
func (u *User) ReadAcceptNotifications() (acceptNotifications []AcceptNotification, err error) {
	rows, err := DB.Query("SELECT * FROM accept_notifications where to_user_id = $1 and class = $2", u.Id, NotificationStatusRead)
	if err != nil {
		return
	}
	for rows.Next() {
		var a AcceptNotification
		if err = rows.Scan(&a.Id, &a.FromUserId, &a.ToUserId, &a.Title, &a.Content, &a.AcceptObjectId, &a.Class, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return
		}
		acceptNotifications = append(acceptNotifications, a)
	}
	rows.Close()
	return
}

// Update() 根据ToUserId和接纳对象id，更新用户的邻桌新茶蒙评消息为已读
func (a *AcceptNotification) Update(to_user_id, accept_object_id int) (err error) {
	return a.UpdateWithContext(to_user_id, accept_object_id, context.Background())
}

// UpdateWithContext 根据ToUserId和接纳对象id，更新用户的邻桌新茶蒙评消息为已读（带context版本）
func (a *AcceptNotification) UpdateWithContext(to_user_id, accept_object_id int, ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	statement := "UPDATE accept_notifications SET class = $4, updated_at = $3 WHERE to_user_id = $1 and accept_object_id = $2"
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, to_user_id, accept_object_id, time.Now(), NotificationStatusRead)
	return
}

// Send(user_ids) 向用户ID队列发送新茶蒙评请求通知
func (a *AcceptNotification) Send(user_ids []int) (err error) {
	return a.SendWithContext(user_ids, context.Background())
}

// SendWithContext 向用户ID队列发送新茶蒙评请求通知（带context版本）
func (a *AcceptNotification) SendWithContext(user_ids []int, ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	for _, user_id := range user_ids {
		a.ToUserId = user_id
		if err = a.CreateWithContext(ctx); err != nil {
			return err
		}
	}
	return nil
}

// AcceptNotificationCount() 获取用户 新茶评审通知总数
// 注意：此函数与 AllAcceptNotificationCount() 功能重复，建议使用 AllAcceptNotificationCount()
// func (user *User) AcceptNotificationCount() (count int) {
// 	return user.AllAcceptNotificationCount()
// }

// NewAcceptNotificationsCount() 获取用户 新茶评审通知数
func (user *User) NewAcceptNotificationsCount() (count int) {
	return user.UnreadAcceptNotificationsCount()
}

// 用户是否有新茶评审未读通知？
func (user *User) HasNewAcceptNotification() bool {
	count := user.NewAcceptNotificationsCount()
	return count > 0
}

// 检查当前用户sUserId是否受邀请审茶，class = 0
func (user *User) CheckHasAcceptNotification(accept_object_id int) (ok bool, err error) {
	return user.CheckHasAcceptNotificationWithContext(accept_object_id, context.Background())
}

// CheckHasAcceptNotificationWithContext 检查当前用户sUserId是否受邀请审茶，class = 0（带context版本）
func (user *User) CheckHasAcceptNotificationWithContext(accept_object_id int, ctx context.Context) (ok bool, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return false, errors.New("database connection is nil")
	}

	query := "SELECT COUNT(*) > 0 FROM accept_notifications WHERE to_user_id = $1 AND accept_object_id = $2 AND class = $3"
	row := DB.QueryRowContext(ctx, query, user.Id, accept_object_id, NotificationStatusUnread)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// 检查当前用户sUserId是否曾经受邀请审茶，class = 1
func (user *User) CheckHasReadAcceptNotification(accept_object_id int, ctx context.Context) (ok bool, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return false, errors.New("database connection is nil")
	}

	query := "SELECT COUNT(*) > 0 FROM accept_notifications WHERE to_user_id = $1 AND accept_object_id = $2 AND class = $3"
	row := DB.QueryRowContext(ctx, query, user.Id, accept_object_id, NotificationStatusRead)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// getAcceptNotificationCountWithContext 通用辅助方法，用于获取不同条件的通知数量（带context版本）
func (user *User) getAcceptNotificationCountWithContext(whereClause string, ctx context.Context) (count int) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return 0
	}

	query := "SELECT count(*) FROM accept_notifications where to_user_id = $1"
	if whereClause != "" {
		query += " AND " + whereClause
	}

	err := DB.QueryRowContext(ctx, query, user.Id).Scan(&count)
	if err != nil {
		return 0
	}
	return
}
