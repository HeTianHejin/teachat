package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// 消息盒子
type MessageBox struct {
	Id        int
	Uuid      string
	Type      int // 类型： 1：家庭，2:团队,
	ObjectId  int // 绑定对象id（家庭id，团队id)
	Count     int // 存量消息数量，默认为0
	MaxCount  int // 存活最大消息数量，默认为199
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time // 软删除，当软删除后，该字段不为null
}

const (
	MessageBoxTypeFamily = 1
	MessageBoxTypeTeam   = 2
)

// 纸条
type Message struct {
	Id             int
	Uuid           string
	MessageBoxId   int // 消息盒子id
	SenderType     int // 发送者类型： 1：家庭成员，2：团队成员,
	SenderObjectId int // 发送者id
	ReceiverType   int // 0：全体，1:成员，
	ReceiverId     int // 接收者id,default = 0

	Content string // 消息内容

	IsRead    bool // 是否已读, default = false
	CreatedAt time.Time
	UpdatedAt *time.Time // 阅读时间
	DeletedAt *time.Time // 软删除，当软删除后，该字段不为null
}

const (
	MessageSenderTypeFamily = 1
	MessageSenderTypeTeam   = 2
)

const (
	MessageReceiverTypeAll    = 0
	MessageReceiverTypeMember = 1
)

var MessageReceiverTypeText = map[int]string{
	MessageReceiverTypeAll:    "全体",
	MessageReceiverTypeMember: "成员",
}

// 获取Message接收者类型
func (m *Message) ReceiverTypeText() string {
	return MessageReceiverTypeText[m.ReceiverType]
}

// CreatedAtDate() 格式化时间
func (m *Message) CreatedAtDate() string {
	return m.CreatedAt.Format(FMT_DATE_CN)
}

// Create() 创建消息
func (m *Message) Create() (err error) {
	return m.CreateWithContext(context.Background())
}

// CreateWithContext 创建消息（带context版本）
func (m *Message) CreateWithContext(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	statement := "INSERT INTO messages (uuid, message_box_id, sender_type, sender_object_id, receiver_type, receiver_id, content, is_read, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, &m.Uuid, &m.MessageBoxId, &m.SenderType, &m.SenderObjectId, &m.ReceiverType, &m.ReceiverId, &m.Content, &m.IsRead, time.Now()).Scan(&m.Id)
	return
}

// GetMessageById() 根据ID获取消息
func (m *Message) GetMessageById(id int) (err error) {
	return m.GetMessageByIdWithContext(id, context.Background())
}

// GetMessageByIdWithContext 根据ID获取消息（带context版本）
func (m *Message) GetMessageByIdWithContext(id int, ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	err = DB.QueryRowContext(ctx, "SELECT * FROM messages WHERE id = $1 AND deleted_at IS NULL", id).Scan(&m.Id, &m.Uuid, &m.MessageBoxId, &m.SenderType, &m.SenderObjectId, &m.ReceiverType, &m.ReceiverId, &m.Content, &m.IsRead, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
	return
}

// UpdateRead() 更新消息为已读
func (m *Message) UpdateRead() (err error) {
	return m.UpdateReadWithContext(context.Background())
}

// UpdateReadWithContext 更新消息为已读（带context版本）
func (m *Message) UpdateReadWithContext(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	statement := "UPDATE messages SET is_read = $2, updated_at = $3 WHERE id = $1"
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, m.Id, true, time.Now())
	return
}

// SoftDelete() 软删除消息
func (m *Message) SoftDelete() (err error) {
	return m.SoftDeleteWithContext(context.Background())
}

// SoftDeleteWithContext 软删除消息（带context版本）
func (m *Message) SoftDeleteWithContext(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	statement := "UPDATE messages SET deleted_at = $2 WHERE id = $1"
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, m.Id, time.Now())
	return
}

// CreatedAtDate() 格式化时间
func (mb *MessageBox) CreatedAtDate() string {
	return mb.CreatedAt.Format(FMT_DATE_CN)
}

// Create() 创建消息盒子
func (mb *MessageBox) Create() (err error) {
	return mb.CreateWithContext(context.Background())
}

// CreateWithContext 创建消息盒子（带context版本）
func (mb *MessageBox) CreateWithContext(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	statement := "INSERT INTO message_boxes (uuid, type, object_id, count, max_count, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, &mb.Uuid, &mb.Type, &mb.ObjectId, &mb.Count, &mb.MaxCount, time.Now()).Scan(&mb.Id)
	return
}

// GetOrCreateMessageBoxWithContext 安全地获取或创建消息盒子（带context版本）
func (mb *MessageBox) GetOrCreateMessageBoxWithContext(msg_type, object_id int, ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	// 首先尝试获取现有的消息盒子
	err = mb.GetMessageBoxByTypeAndObjectIdWithContext(msg_type, object_id, ctx)
	if err == nil {
		// 找到了现有的消息盒子
		return nil
	}

	// 如果是"未找到"错误，尝试创建新的消息盒子
	if err.Error() == "message box not found" {
		// 准备新消息盒子的数据
		mb.Uuid = Random_UUID()
		mb.Type = msg_type
		mb.ObjectId = object_id
		mb.Count = 0
		mb.MaxCount = 199

		// 尝试创建，如果因为并发导致重复，则重新获取
		err = mb.CreateWithContext(ctx)
		if err == nil {
			// 创建成功
			return nil
		}

		// 如果创建失败（可能是唯一约束冲突），再次尝试获取
		if err != nil {
			// 等待一小段时间让其他事务完成
			time.Sleep(10 * time.Millisecond)
			err = mb.GetMessageBoxByTypeAndObjectIdWithContext(msg_type, object_id, ctx)
			if err == nil {
				// 其他并发请求已经创建了消息盒子
				return nil
			}
		}
	}

	return err
}

// GetMessageBoxById() 根据ID获取消息盒子
func (mb *MessageBox) GetMessageBoxById(id int) (err error) {
	return mb.GetMessageBoxByIdWithContext(id, context.Background())
}

// GetMessageBoxByIdWithContext 根据ID获取消息盒子（带context版本）
func (mb *MessageBox) GetMessageBoxByIdWithContext(id int, ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	err = DB.QueryRowContext(ctx, "SELECT id, uuid, type, object_id, count, max_count, created_at, updated_at, deleted_at FROM message_boxes WHERE id = $1 AND deleted_at IS NULL", id).Scan(&mb.Id, &mb.Uuid, &mb.Type, &mb.ObjectId, &mb.Count, &mb.MaxCount, &mb.CreatedAt, &mb.UpdatedAt, &mb.DeletedAt)
	return
}

// GetMessageBoxByTypeAndObjectId() 根据类型和对象ID获取消息盒子
func (mb *MessageBox) GetMessageBoxByTypeAndObjectId(msg_type, object_id int) (err error) {
	return mb.GetMessageBoxByTypeAndObjectIdWithContext(msg_type, object_id, context.Background())
}

// GetMessageBoxByTypeAndObjectIdWithContext 根据类型和对象ID获取消息盒子（带context版本）
func (mb *MessageBox) GetMessageBoxByTypeAndObjectIdWithContext(msg_type, object_id int, ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	err = DB.QueryRowContext(ctx, "SELECT * FROM message_boxes WHERE type = $1 AND object_id = $2 AND deleted_at IS NULL", msg_type, object_id).Scan(&mb.Id, &mb.Uuid, &mb.Type, &mb.ObjectId, &mb.Count, &mb.MaxCount, &mb.CreatedAt, &mb.UpdatedAt, &mb.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 记录不存在时返回特定的错误，让调用者能够区分
			return errors.New("message box not found")
		}
		return err
	}
	return nil
}

// SoftDelete() 软删除消息盒子
func (mb *MessageBox) SoftDelete() (err error) {
	return mb.SoftDeleteWithContext(context.Background())
}

// SoftDeleteWithContext 软删除消息盒子（带context版本）
func (mb *MessageBox) SoftDeleteWithContext(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	statement := "UPDATE message_boxes SET deleted_at = $2 WHERE id = $1"
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, mb.Id, time.Now())
	return
}

// Messages() 获取消息盒子中的所有消息
func (mb *MessageBox) Messages() (messages []Message, err error) {
	return mb.MessagesWithContext(context.Background())
}

// MessagesWithContext 获取消息盒子中的所有消息（带context版本）
func (mb *MessageBox) MessagesWithContext(ctx context.Context) (messages []Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return nil, errors.New("database connection is nil")
	}

	rows, err := DB.QueryContext(ctx, "SELECT * FROM messages WHERE message_box_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC", mb.Id)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		if err = rows.Scan(&m.Id, &m.Uuid, &m.MessageBoxId, &m.SenderType, &m.SenderObjectId, &m.ReceiverType, &m.ReceiverId, &m.Content, &m.IsRead, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return
		}
		messages = append(messages, m)
	}
	return
}

// UnreadMessages() 获取消息盒子中的未读消息
func (mb *MessageBox) UnreadMessages() (messages []Message, err error) {
	return mb.UnreadMessagesWithContext(context.Background())
}

// UnreadMessagesWithContext 获取消息盒子中的未读消息（带context版本）
func (mb *MessageBox) UnreadMessagesWithContext(ctx context.Context) (messages []Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return nil, errors.New("database connection is nil")
	}

	rows, err := DB.QueryContext(ctx, "SELECT * FROM messages WHERE message_box_id = $1 AND is_read = $2 AND deleted_at IS NULL ORDER BY created_at DESC", mb.Id, false)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		if err = rows.Scan(&m.Id, &m.Uuid, &m.MessageBoxId, &m.SenderType, &m.SenderObjectId, &m.ReceiverType, &m.ReceiverId, &m.Content, &m.IsRead, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return
		}
		messages = append(messages, m)
	}
	return
}

// UnreadMessagesCount() 获取消息盒子中未读消息数量
func (mb *MessageBox) UnreadMessagesCount() (count int) {
	return mb.getMessageCountWithContext("is_read = false", context.Background())
}

// UnreadMessagesCountForUser() 根据用户权限获取消息盒子中未读消息数量
func (mb *MessageBox) UnreadMessagesCountForUser(userId int) (count int) {
	return mb.getMessageCountForUserWithContext("is_read = false", userId, context.Background())
}

// getMessageCountForUserWithContext 通用辅助方法，用于获取用户不同条件的消息数量（带context版本）
func (mb *MessageBox) getMessageCountForUserWithContext(whereClause string, userId int, ctx context.Context) (count int) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return 0
	}

	query := `
		SELECT COUNT(*) FROM messages 
		WHERE message_box_id = $1 AND deleted_at IS NULL 
		AND (receiver_type = $2 OR (receiver_type = $3 AND receiver_id = $4))
	`
	if whereClause != "" {
		query += " AND " + whereClause
	}

	err := DB.QueryRowContext(ctx, query, mb.Id, MessageReceiverTypeAll, MessageReceiverTypeMember, userId).Scan(&count)
	if err != nil {
		return 0
	}
	return
}

// AllMessagesCount() 获取消息盒子中所有消息数量
func (mb *MessageBox) AllMessagesCount() (count int) {
	return mb.getMessageCountWithContext("", context.Background())
}

// AllMessagesCountForUser() 根据用户权限获取消息盒子中所有消息数量
func (mb *MessageBox) AllMessagesCountForUser(userId int) (count int) {
	return mb.getMessageCountForUserWithContext("", userId, context.Background())
}

// getMessageCountWithContext 通用辅助方法，用于获取不同条件的消息数量（带context版本）
func (mb *MessageBox) getMessageCountWithContext(whereClause string, ctx context.Context) (count int) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return 0
	}

	query := "SELECT count(*) FROM messages WHERE message_box_id = $1 AND deleted_at IS NULL"
	if whereClause != "" {
		query += " AND " + whereClause
	}

	err := DB.QueryRowContext(ctx, query, mb.Id).Scan(&count)
	if err != nil {
		return 0
	}
	return
}

// 根据用户ID获取用户相关的所有消息盒子
func (u *User) MessageBoxes() (messageBoxes []MessageBox, err error) {
	return u.MessageBoxesWithContext(context.Background())
}

// MessageBoxesWithContext 根据用户ID获取用户相关的所有消息盒子（带context版本）
func (u *User) MessageBoxesWithContext(ctx context.Context) (messageBoxes []MessageBox, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return nil, errors.New("database connection is nil")
	}

	// 获取用户所属的家庭和团队的消息盒子
	query := `
		SELECT DISTINCT mb.id, mb.uuid, mb.type, mb.object_id, mb.count, mb.max_count, mb.created_at, mb.updated_at, mb.deleted_at FROM message_boxes mb 
		LEFT JOIN family_members fm ON (mb.type = $1 AND mb.object_id = fm.family_id AND fm.user_id = $2)
		LEFT JOIN team_members tm ON (mb.type = $3 AND mb.object_id = tm.team_id AND tm.user_id = $2)
		WHERE mb.deleted_at IS NULL AND (fm.user_id = $2 OR tm.user_id = $2)
	`
	rows, err := DB.QueryContext(ctx, query, MessageBoxTypeFamily, u.Id, MessageBoxTypeTeam, u.Id)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var mb MessageBox
		if err = rows.Scan(&mb.Id, &mb.Uuid, &mb.Type, &mb.ObjectId, &mb.Count, &mb.MaxCount, &mb.CreatedAt, &mb.UpdatedAt, &mb.DeletedAt); err != nil {
			return
		}
		messageBoxes = append(messageBoxes, mb)
	}
	return
}

// 根据消息盒子ID和接收者ID获取用户相关的消息
func (u *User) MessagesByMessageBox(messageBoxId int) (messages []Message, err error) {
	return u.MessagesByMessageBoxWithContext(messageBoxId, context.Background())
}

// MessagesByMessageBoxWithContext 根据消息盒子ID和接收者ID获取用户相关的消息（带context版本）
func (u *User) MessagesByMessageBoxWithContext(messageBoxId int, ctx context.Context) (messages []Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT m.* FROM messages m 
		WHERE m.message_box_id = $1 AND m.deleted_at IS NULL 
		AND (m.receiver_type = $2 OR (m.receiver_type = $3 AND m.receiver_id = $4))
		ORDER BY m.created_at DESC
	`
	rows, err := DB.QueryContext(ctx, query, messageBoxId, MessageReceiverTypeAll, MessageReceiverTypeMember, u.Id)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		if err = rows.Scan(&m.Id, &m.Uuid, &m.MessageBoxId, &m.SenderType, &m.SenderObjectId, &m.ReceiverType, &m.ReceiverId, &m.Content, &m.IsRead, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return
		}
		messages = append(messages, m)
	}
	return
}

// UnreadMessagesCount() 获取用户所有未读消息总数
func (u *User) UnreadMessagesCount() (count int) {
	return u.getMessagesCountWithContext("is_read = false", context.Background())
}

// AllMessagesCount() 获取用户所有消息总数
func (u *User) AllMessagesCount() (count int) {
	return u.getMessagesCountWithContext("", context.Background())
}

// getMessagesCountWithContext 通用辅助方法，用于获取用户不同条件的消息数量（带context版本）
func (u *User) getMessagesCountWithContext(whereClause string, ctx context.Context) (count int) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return 0
	}

	query := `
		SELECT COUNT(*) FROM messages m 
		INNER JOIN message_boxes mb ON m.message_box_id = mb.id
		LEFT JOIN family_members fm ON (mb.type = $1 AND mb.object_id = fm.family_id AND fm.user_id = $2)
		LEFT JOIN team_members tm ON (mb.type = $3 AND mb.object_id = tm.team_id AND tm.user_id = $2)
		WHERE m.deleted_at IS NULL AND mb.deleted_at IS NULL 
		AND (fm.user_id = $2 OR tm.user_id = $2)
		AND (m.receiver_type = $4 OR (m.receiver_type = $5 AND m.receiver_id = $2))
	`
	if whereClause != "" {
		query += " AND " + whereClause
	}

	err := DB.QueryRowContext(ctx, query, MessageBoxTypeFamily, u.Id, MessageBoxTypeTeam, u.Id, MessageReceiverTypeAll, MessageReceiverTypeMember, u.Id).Scan(&count)
	if err != nil {
		return 0
	}
	return
}

// Update() 更新消息盒子
func (mb *MessageBox) Update() (err error) {
	return mb.UpdateWithContext(context.Background())
}

// UpdateWithContext 更新消息盒子（带context版本）
func (mb *MessageBox) UpdateWithContext(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	statement := "UPDATE message_boxes SET count = $2, updated_at = $3 WHERE id = $1"
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, mb.Id, mb.Count, time.Now())
	return
}

// GetMessageBoxByUUID() 根据UUID获取消息盒子
func (mb *MessageBox) GetMessageBoxByUUID(uuid string) (err error) {
	return mb.GetMessageBoxByUUIDWithContext(uuid, context.Background())
}

// GetMessageBoxByUUIDWithContext 根据UUID获取消息盒子（带context版本）
func (mb *MessageBox) GetMessageBoxByUUIDWithContext(uuid string, ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return errors.New("database connection is nil")
	}

	err = DB.QueryRowContext(ctx, "SELECT id, uuid, type, object_id, count, max_count, created_at, updated_at, deleted_at FROM message_boxes WHERE uuid = $1 AND deleted_at IS NULL", uuid).Scan(&mb.Id, &mb.Uuid, &mb.Type, &mb.ObjectId, &mb.Count, &mb.MaxCount, &mb.CreatedAt, &mb.UpdatedAt, &mb.DeletedAt)
	return
}

// GetMessages() 获取消息盒子的所有消息
func (mb *MessageBox) GetMessages() (messages []Message, err error) {
	return mb.GetMessagesWithContext(context.Background())
}

// GetMessagesWithContext 获取消息盒子的所有消息（带context版本）
func (mb *MessageBox) GetMessagesWithContext(ctx context.Context) (messages []Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return nil, errors.New("database connection is nil")
	}

	query := "SELECT * FROM messages WHERE message_box_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC"
	rows, err := DB.QueryContext(ctx, query, mb.Id)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		err = rows.Scan(&m.Id, &m.Uuid, &m.MessageBoxId, &m.SenderType, &m.SenderObjectId, &m.ReceiverType, &m.ReceiverId, &m.Content, &m.IsRead, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
		if err != nil {
			return
		}
		messages = append(messages, m)
	}

	return
}

// GetMessagesForUser() 根据用户权限获取消息盒子的消息
func (mb *MessageBox) GetMessagesForUser(userId int) (messages []Message, err error) {
	return mb.GetMessagesForUserWithContext(userId, context.Background())
}

// GetMessagesForUserWithContext 根据用户权限获取消息盒子的消息（带context版本）
func (mb *MessageBox) GetMessagesForUserWithContext(userId int, ctx context.Context) (messages []Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if DB == nil {
		return nil, errors.New("database connection is nil")
	}

	// 权限控制：发送给"全体"的消息所有人可见，发送给"成员"的消息仅指定成员可见
	query := `
		SELECT * FROM messages 
		WHERE message_box_id = $1 AND deleted_at IS NULL 
		AND (receiver_type = $2 OR (receiver_type = $3 AND receiver_id = $4))
		ORDER BY created_at DESC
	`
	rows, err := DB.QueryContext(ctx, query, mb.Id, MessageReceiverTypeAll, MessageReceiverTypeMember, userId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		err = rows.Scan(&m.Id, &m.Uuid, &m.MessageBoxId, &m.SenderType, &m.SenderObjectId, &m.ReceiverType, &m.ReceiverId, &m.Content, &m.IsRead, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
		if err != nil {
			return
		}
		messages = append(messages, m)
	}

	return
}
