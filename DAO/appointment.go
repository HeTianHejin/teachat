package data

import (
	"context"
	"errors"
	util "teachat/Util"
	"time"
)

type ThreadAppointment struct {
	ID              int
	ProjectID       int
	ThreadID        int
	RequesterUserID int
	ProviderUserID  int

	// 状态管理
	Status      AppointmentStatus // 枚举类型
	ConfirmedAt *time.Time        // 明确记录确认时间
	RejectedAt  *time.Time        // 记录拒绝时间（可选）

	// 基础时间戳
	CreatedAt time.Time
	UpdatedAt time.Time // 始终记录最后更新时间
}

// ThreadAppointment.Isconfirmed() bool
func (t *ThreadAppointment) IsConfirmed() bool {
	return t.Status == StatusConfirmed
}

// ThreadAppointment.IsRejected() bool
func (t *ThreadAppointment) IsRejected() bool {
	return t.Status == StatusRejected
}

// ThreadAppointment.IsCancelled() bool
func (t *ThreadAppointment) IsCancelled() bool {
	return t.Status == StatusCancelled
}

// ThreadAppointment.IsPending() bool
func (t *ThreadAppointment) IsPending() bool {
	return t.Status == StatusPending
}

// ThreadAppointment.IsExpired() bool
func (t *ThreadAppointment) IsExpired() bool {
	return t.Status == StatusPending && t.CreatedAt.Add(24*time.Hour).Before(time.Now())
}

// ThreadAppointment.StatusString() string
func (t *ThreadAppointment) StatusString() string {
	switch t.Status {
	case StatusPending:
		return "待确认"
	case StatusConfirmed:
		return "已确认"
	case StatusRejected:
		return "已拒绝"
	case StatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// ThreadAppointment.Create(ctx context.Context) (err error)
func (t *ThreadAppointment) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO thread_appointment (project_id, thread_id, requester_user_id, provider_user_id, status) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	stmt, err := Db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.ProjectID, t.ThreadID, t.RequesterUserID, t.ProviderUserID, t.Status).Scan(&t.ID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out")
		}
		return
	}
	return
}
func (t *ThreadAppointment) Get(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, project_id, thread_id, requester_user_id, provider_user_id, status, created_at, updated_at FROM thread_appointment WHERE id = $1`
	stmt, err := Db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.ID).Scan(&t.ID, &t.ProjectID, &t.ThreadID, &t.RequesterUserID, &t.ProviderUserID, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out")
		}
		return
	}
	return
}

// 预约状态枚举
type AppointmentStatus int

const (
	StatusPending   AppointmentStatus = iota // 待确认
	StatusConfirmed                          // 已确认
	StatusRejected                           // 已拒绝
	StatusCancelled                          // 已取消
)

// 约茶确认
func (a *ThreadAppointment) Confirm(confirmerID int) error {
	if a.Status != StatusPending {
		return errors.New("只能确认待处理的预约")
	}

	// 验证操作者权限
	// if confirmerID != a.ProviderUserID && confirmerID != a.RequesterUserID {
	// 	return errors.New("无操作权限")
	// }

	a.Status = StatusConfirmed
	now := time.Now()
	a.ConfirmedAt = &now
	a.UpdatedAt = now

	return nil
}

// func (t *ThreadAppointment) Create(ctx context.Context) (err error) {
// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	statement := `INSERT INTO thread_appointment (project_id, thread_id) VALUES ($1, $2) RETURNING id`
// 	stmt, err := Db.PrepareContext(ctx, statement)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
// 	err = stmt.QueryRowContext(ctx, t.ProjectId, t.ThreadId).Scan(&t.Id)
// 	if err != nil {
// 		if errors.Is(err, context.DeadlineExceeded) {
// 			util.Debug("Query timed out")
// 		}
// 		return
// 	}
// 	return
// }
// func (t *ThreadAppointment) Get(ctx context.Context) (err error) {
// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	statement := `SELECT id, project_id, thread_id, created_at FROM thread_appointment WHERE id = $1`
// 	stmt, err := Db.PrepareContext(ctx, statement)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
// 	err = stmt.QueryRowContext(ctx, t.Id).Scan(&t.Id, &t.ProjectId, &t.ThreadId, &t.CreatedAt)
// 	if err != nil {
// 		if errors.Is(err, context.DeadlineExceeded) {
// 			util.Debug("Query timed out")
// 		}
// 		return
// 	}
// 	return
// }

// type AppointmentConfirmed struct {
// 	Id        int
// 	ProjectId int
// 	PostId    int
// 	CreatedAt time.Time
// }

// func (a *AppointmentConfirmed) Create(ctx context.Context) (err error) {
// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	statement := `INSERT INTO appointment_confirmed (project_id, post_id) VALUES ($1, $2) RETURNING id`
// 	stmt, err := Db.PrepareContext(ctx, statement)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
// 	err = stmt.QueryRowContext(ctx, a.ProjectId, a.PostId).Scan(&a.Id)
// 	if err != nil {
// 		if errors.Is(err, context.DeadlineExceeded) {
// 			util.Debug("Query timed out")
// 		}
// 		return
// 	}
// 	return
// }
// func (a *AppointmentConfirmed) Get(ctx context.Context) (err error) {
// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	statement := `SELECT id, project_id, post_id, created_at FROM appointment_confirmed WHERE id = $1`
// 	stmt, err := Db.PrepareContext(ctx, statement)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
// 	err = stmt.QueryRowContext(ctx, a.Id).Scan(&a.Id, &a.ProjectId, &a.PostId, &a.CreatedAt)
// 	if err != nil {
// 		if errors.Is(err, context.DeadlineExceeded) {
// 			util.Debug("Query timed out")
// 		}
// 		return
// 	}
// 	return
// }
