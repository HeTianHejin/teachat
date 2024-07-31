package data

import (
	"errors"
	"time"
)

// AcceptObject 友邻盲评对象
type AcceptObject struct {
	Id         int
	ObjectType int // 1:茶话会，2:茶台， 3:茶议，4: 品味， 5:茶团
	ObjectId   int

	// 页面动态数据,不存储到数据库中
	PageData AcceptObjectPageData
}

// 用户新消息统计
type NewMessageCount struct {
	Id     int
	UserId int
	Count  int
}

// 记录友邻盲评，评判结果
// 慎独，茶议正式发布之前需要邻桌盲评是否符合社区文明之约？
type Acceptance struct {
	Id             int
	AcceptObjectId int
	XAccept        bool
	XUserId        int
	XAcceptedAt    time.Time
	YAccept        bool
	YUserId        int
	YAcceptedAt    time.Time
}

// Create() Acceptance新建一条 友邻盲评 记录
func (a *Acceptance) Create() (err error) {
	statement := `INSERT INTO acceptances (accept_object_id, x_accept, x_user_id, x_accepted_at, y_accept, y_user_id, y_accepted_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(a.AcceptObjectId, a.XAccept, a.XUserId, a.XAcceptedAt, a.YAccept, a.YUserId, a.YAcceptedAt).Scan(&a.Id)
	if err != nil {
		return err
	}
	return nil
}

// Update() 根据id更新一条友邻盲评 Y记录
func (a *Acceptance) Update() (err error) {
	statement := `UPDATE acceptances SET y_accept = $1, y_user_id = $2, y_accepted_at = $3 WHERE id = $4`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(a.YAccept, a.YUserId, a.YAcceptedAt, a.Id)
	if err != nil {
		return err
	}
	return nil
}

// Get() ���id获取一条友����评记录
func (a *Acceptance) Get() (err error) {
	err = Db.QueryRow(`SELECT id, accept_object_id, x_accept, x_user_id, x_accepted_at, y_accept, y_user_id, y_accepted_at FROM acceptances WHERE id = $1`, a.Id).Scan(&a.Id, &a.AcceptObjectId, &a.XAccept, &a.XUserId, &a.XAcceptedAt, &a.YAccept, &a.YUserId, &a.YAcceptedAt)
	if err != nil {
		return err
	}
	return nil
}

// Acceptance.GetByAcceptObjectId() 根据友邻盲评对象id获取友邻盲评 1记录
func (a *Acceptance) GetByAcceptObjectId() (acceptance Acceptance, err error) {
	err = Db.QueryRow(`SELECT id, accept_object_id, x_accept, x_user_id, x_accepted_at, y_accept, y_user_id, y_accepted_at FROM acceptances WHERE accept_object_id = $1`, a.AcceptObjectId).Scan(&acceptance.Id, &acceptance.AcceptObjectId, &acceptance.XAccept, &acceptance.XUserId, &acceptance.XAcceptedAt, &acceptance.YAccept, &acceptance.YUserId, &acceptance.YAcceptedAt)
	if err != nil {
		return acceptance, err
	}
	return acceptance, nil
}

// Get() ���id获取��评记录

// Create（） AcceptObject新建一条盲评接纳对象的记录
func (a *AcceptObject) Create() (err error) {
	statement := `INSERT INTO accept_objects (object_type, object_id) VALUES ($1, $2) RETURNING id`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(a.ObjectType, a.ObjectId).Scan(&a.Id)
	if err != nil {
		return err
	}
	return nil
}

// Get() 据id获取������评对象
func (a *AcceptObject) Get() (err error) {
	err = Db.QueryRow(`SELECT id, object_type, object_id FROM accept_objects WHERE id = $1`, a.Id).Scan(&a.Id, &a.ObjectType, &a.ObjectId)
	if err != nil {
		return err
	}
	return nil
}

// 根据acceptObjectId查询返回其“友邻盲评”对象
func (ao *AcceptObject) GetObjectByACId() (object interface{}, err error) {
	switch ao.ObjectType {
	case 1:
		ob, err := GetObjectiveById(ao.ObjectId)
		if err != nil {
			return nil, err
		}
		return ob, err
	case 2:
		pr, err := GetProjectById(ao.ObjectId)
		if err != nil {
			return nil, err
		}
		return pr, err
	case 3:
		dThread := DraftThread{
			Id: ao.ObjectId,
		}
		if err = dThread.Get(); err != nil {
			return nil, err
		}
		return dThread, err
	case 4:
		dPost, err := GetPostbyId(ao.ObjectId)
		if err != nil {
			return nil, err
		}
		return dPost, err
	case 5:
		team, err := GetTeamById(ao.ObjectId)
		if err != nil {
			return nil, err
		}
		return team, err
	default:
		return nil, errors.New("unknown ObjectType")
	}
}

// 创建一个消息计数
func (m *NewMessageCount) Save() error {
	statement := `INSERT INTO new_message_counts (user_id, count) VALUES ($1, $2)`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.UserId, m.Count)
	if err != nil {
		return err
	}
	return nil
}

// update(） 修改一个消息计数
func (m *NewMessageCount) Update() error {
	statement := `UPDATE new_message_counts SET count = $1 WHERE id = $2`
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.Count, m.Id)
	if err != nil {
		return err
	}
	return nil
}

// 根据.UserId获取用户新消息的消息计数
func (m *NewMessageCount) GetByUserId() error {
	err := Db.QueryRow(`SELECT id, count FROM new_message_counts WHERE user_id = $1`, m.UserId).Scan(&m.Id, &m.Count)
	if err != nil {
		return err
	}
	return nil
}

// 检查new_message_counts表里是否已经存在用户id，返回bool，true为存在
func (m *NewMessageCount) Check() (valid bool, err error) {
	err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM new_message_counts WHERE user_id = $1)", m.UserId).Scan(&valid)
	if err != nil {
		return false, err
	}
	return valid, nil
}

// GetUserMessage() 获取用户消息数
func (user *User) GetNewMessageCount() (count int, err error) {
	err = Db.QueryRow(`SELECT count FROM new_message_counts WHERE user_id = $1`, user.Id).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, err
}

// 根据count是否大于0来判断是否有未读消息
func (user *User) HasUnreadMessage() (has bool) {
	// 友邻盲评未读消息数量
	count1, _ := user.GetNewMessageCount()
	// 邀请函未读消息数量
	count2 := user.InvitationUnviewedCount()

	count := count1 + count2

	return count > 0
}

// AddNewUserMessages() 添加一条用户新信息数
// 首先检查new_message_counts记录里是否已经存在用户id，
// 如果没有，执行save()，如果有，执行update()
func AddUserMessageCount(user_id int) error {
	messageCount := NewMessageCount{
		UserId: user_id,
	}
	// 这里检查是否存在此用户记录
	exists, err := messageCount.Check()
	if err != nil {
		return err
	}

	if exists {
		// User record exists, update the count +1
		if err := messageCount.GetByUserId(); err != nil {
			return err
		}
		messageCount.Count += 1
		return messageCount.Update()
	} else {
		// User record doesn't exist, create a new one
		messageCount.Count = 1
		return messageCount.Save()
	}

}

// SubtractUserMessageCount() 减去通知小黑板上用户1消息数
func SubtractUserMessageCount(user_id int) error {
	mesC := NewMessageCount{
		UserId: user_id,
	}

	if ok, err := mesC.Check(); !ok {
		// 不存在，返回错误
		return err
	}
	// 存在，-1消息记录，执行update()
	if err := mesC.GetByUserId(); err != nil {
		return err
	}

	if mesC.Count <= 0 {
		return errors.New("error in the number of messages, The number of messages must not be negative")
	}

	mesC.Count -= 1

	return mesC.Update()

}
