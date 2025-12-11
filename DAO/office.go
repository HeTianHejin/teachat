package dao

import (
	"context"
	"errors"
	"time"
)

// 记录当前用户最后浏览页面，用于便利“返回”链接 功能
type LastQuery struct {
	Id      int
	UserId  int
	Path    string
	Query   string
	QueryAt time.Time
}

// Create() 新建一条 LastQuery记录
func (lq *LastQuery) Create() (err error) {
	statement := `INSERT INTO last_queries (user_id, path, query, query_at) VALUES ($1, $2, $3, $4) RETURNING id`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(lq.UserId, lq.Path, lq.Query, time.Now()).Scan(&lq.Id)
	// if err != nil {
	// 	return err
	// }
	return
}

// Get() 读取一条 LastQuery记录
func (lq *LastQuery) Get() (err error) {
	statement := `SELECT id, user_id, path, query, query_at FROM last_queries WHERE user_id = $1 ORDER BY query_at DESC LIMIT 1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(lq.UserId).Scan(&lq.Id, &lq.UserId, &lq.Path, &lq.Query, &lq.QueryAt)
	if err != nil {
		return err
	}
	return nil
}

// AcceptObject 友邻蒙评对象
type AcceptObject struct {
	Id         int
	ObjectType int // 1:茶话会，2:茶台， 3:茶议，4: 品味， 5:茶团， 6:集团
	ObjectId   int
}

const (
	AcceptObjectTypeObjective = iota + 1 //objective （目标）
	AcceptObjectTypeProject              //project （项目）
	AcceptObjectTypeThread               //thread （茶议）
	AcceptObjectTypePost                 //post （品味）
	AcceptObjectTypeTeam                 //team （茶团）
	AcceptObjectTypeGroup                //group（集团）
)

// 用户新通知统计
type NewNotificationCount struct {
	Id     int
	UserId int
	Count  int
}

// 记录友邻蒙评，评判结果
// 慎独，茶议正式发布之前需要邻桌蒙评是否符合社区文明之约？
type Acceptance struct {
	Id             int
	AcceptObjectId int
	XAccept        bool
	XUserId        int
	XAcceptedAt    time.Time
	YAccept        bool
	YUserId        int
	YAcceptedAt    *time.Time
}

// Create() Acceptance新建一条 友邻蒙评 记录
func (a *Acceptance) Create() (err error) {
	statement := `INSERT INTO acceptances (accept_object_id, x_accept, x_user_id, x_accepted_at, y_accept, y_user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(a.AcceptObjectId, a.XAccept, a.XUserId, time.Now(), a.YAccept, a.YUserId).Scan(&a.Id)

	return
}

// Update() 根据id更新一条友邻蒙评 Y记录
func (a *Acceptance) Update() (err error) {
	statement := `UPDATE acceptances SET y_accept = $1, y_user_id = $2, y_accepted_at = $3 WHERE id = $4`
	stmt, err := db.Prepare(statement)
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

// Get() 获取一条友友邻蒙评记录
func (a *Acceptance) Get() (err error) {
	err = db.QueryRow(`SELECT id, accept_object_id, x_accept, x_user_id, x_accepted_at, y_accept, y_user_id, y_accepted_at FROM acceptances WHERE id = $1`, a.Id).Scan(&a.Id, &a.AcceptObjectId, &a.XAccept, &a.XUserId, &a.XAcceptedAt, &a.YAccept, &a.YUserId, &a.YAcceptedAt)
	if err != nil {
		return err
	}
	return nil
}

// Acceptance.GetByAcceptObjectId() 根据友邻蒙评对象id获取友邻蒙评 1记录
func (a *Acceptance) GetByAcceptObjectId() (Acceptance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, accept_object_id, x_accept, x_user_id, x_accepted_at, y_accept, y_user_id, y_accepted_at 
              FROM acceptances 
              WHERE accept_object_id = $1`

	row := db.QueryRowContext(ctx, query, a.AcceptObjectId)
	var acceptance Acceptance
	err := row.Scan(&acceptance.Id, &acceptance.AcceptObjectId, &acceptance.XAccept, &acceptance.XUserId, &acceptance.XAcceptedAt, &acceptance.YAccept, &acceptance.YUserId, &acceptance.YAcceptedAt)
	if err != nil {
		return Acceptance{}, err
	}
	return acceptance, nil
}

// Create（） AcceptObject新建一条蒙评接纳对象的记录
func (a *AcceptObject) Create() (err error) {
	statement := `INSERT INTO accept_objects (object_type, object_id) VALUES ($1, $2) RETURNING id`
	stmt, err := db.Prepare(statement)
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

// Get() 据id获取友邻蒙评对象
func (a *AcceptObject) Get() (err error) {
	err = db.QueryRow(`SELECT id, object_type, object_id FROM accept_objects WHERE id = $1`, a.Id).Scan(&a.Id, &a.ObjectType, &a.ObjectId)
	if err != nil {
		return err
	}
	return nil
}

// 根据acceptObjectId查询返回其“友邻蒙评”对象
func (ao *AcceptObject) GetObjectByACId() (object any, err error) {
	switch ao.ObjectType {
	case 1:
		ob := Objective{
			Id: ao.ObjectId}
		if err = ob.Get(); err != nil {
			return nil, err
		}
		return ob, err
	case 2:
		pr := Project{
			Id: ao.ObjectId,
		}
		if err = pr.Get(); err != nil {
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
		d_post := DraftPost{
			Id: ao.ObjectId,
		}
		if err = d_post.Get(); err != nil {
			return nil, err
		}

		return d_post, err
	case 5:
		team, err := GetTeam(ao.ObjectId)
		if err != nil {
			return nil, err
		}
		return team, err
	default:
		return nil, errors.New("unknown ObjectType")
	}
}

// 创建一个通知计数
func (m *NewNotificationCount) Save() error {
	statement := `INSERT INTO new_notification_counts (user_id, count) VALUES ($1, $2)`
	stmt, err := db.Prepare(statement)
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

// update(） 修改一个通知计数
func (m *NewNotificationCount) Update() error {
	statement := `UPDATE new_notification_counts SET count = $1 WHERE id = $2`
	stmt, err := db.Prepare(statement)
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

// 根据.UserId获取用户新通知的通知计数
func (m *NewNotificationCount) GetByUserId() error {
	err := db.QueryRow(`SELECT id, count FROM new_notification_counts WHERE user_id = $1`, m.UserId).Scan(&m.Id, &m.Count)
	if err != nil {
		return err
	}
	return nil
}

// 检查new_notification_counts表里是否已经存在用户id，返回bool，true为存在
func (m *NewNotificationCount) Check() (valid bool, err error) {
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM new_notification_counts WHERE user_id = $1)", m.UserId).Scan(&valid)
	if err != nil {
		return false, err
	}
	return valid, nil
}

// GetUserNotification() 获取用户通知数
func (user *User) GetNewNotificationCount() (count int, err error) {
	err = db.QueryRow(`SELECT count FROM new_notification_counts WHERE user_id = $1`, user.Id).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, err
}

// 根据count是否大于0来判断是否有未读通知
func (user *User) HasUnreadNotification() (has bool) {
	// 友邻蒙评未读通知数量
	count1, _ := user.GetNewNotificationCount()
	// 邀请函未读通知数量
	count2 := user.InvitationUnviewedCount()

	count := count1 + count2

	return count > 0
}

// AddNewUserNotifications() 添加一条用户新通知数
// 首先检查new_notification_counts记录里是否已经存在用户id，
// 如果没有，执行save()，如果有，执行update()
func AddUserNotificationCount(user_id int) error {
	notificationCount := NewNotificationCount{
		UserId: user_id,
	}
	// 这里检查是否存在此用户记录
	exists, err := notificationCount.Check()
	if err != nil {
		return err
	}

	if exists {
		// User record exists, update the count +1
		if err := notificationCount.GetByUserId(); err != nil {
			return err
		}
		notificationCount.Count += 1
		return notificationCount.Update()
	} else {
		// User record doesn't exist, create a new one
		notificationCount.Count = 1
		return notificationCount.Save()
	}

}

// SubtractUserNotificationCount() 减去通知小黑板上用户1通知数
func SubtractUserNotificationCount(user_id int) error {
	mesC := NewNotificationCount{
		UserId: user_id,
	}

	if ok, err := mesC.Check(); !ok {
		// 不存在，返回错误
		return err
	}
	// 存在，-1通知记录，执行update()
	if err := mesC.GetByUserId(); err != nil {
		return err
	}

	if mesC.Count <= 0 {
		return errors.New("error in the number of notifications, The number of notifications must not be negative")
	}

	mesC.Count -= 1

	return mesC.Update()

}

// 查询类型常量
const (
	SearchTypeUserNameOrEmail = 0  // 按用户名查询
	SearchTypeTeamAbbr        = 1  // 按团队简称查询
	SearchTypeThreadTitle     = 2  // 按茶议标题查询
	SearchTypeObjectiveTitle  = 3  // 按茶围名称查询
	SearchTypeProjectTitle    = 4  // 按茶台名称查询
	SearchTypePlaceName       = 5  // 按茶室地方名称查询
	SearchTypeEnvironment     = 6  // 按环境条件名称查询
	SearchTypeHazard          = 7  // 按隐患名称查询
	SearchTypeRisk            = 8  //按风险名称查询
	SearchTypeGoods           = 9  // 按物资名称查询
	SearchTypeUserId          = 10 // 按用户id查询
	SearchTypeSkill           = 11 // 按技能名称查询
	SearchTypeMagic           = 12 // 按法力名称查询
	SearchTypeFamilyName      = 13 // 按家庭名称查询
)
