package data

import (
	"context"
	"errors"
	util "teachat/Util"
	"time"
)

// 约茶
type ProjectAppointment struct {
	Id        int
	ProjectId int
	Note      string //做备注，简明扼要地描述一下约会目的

	PayerTeamId   int //出茶叶团队Id
	PayerFamilyId int //出茶叶家庭Id
	PayeeTeamId   int //收茶叶团队Id
	PayeeFamilyId int //收茶叶家庭Id

	VerifierUserId   int
	VerifierFamilyId int
	VerifierTeamId   int

	PayerUserId int //出茶叶代表人Id
	PayeeUserId int //收茶叶代表人Id

	// 状态管理：
	// 0、"待确认"
	// 1、"已确认"
	// 2、"已拒绝"
	// 3、"已取消"
	Status      AppointmentStatus // 枚举类型
	ConfirmedAt *time.Time        // 明确记录确认时间
	RejectedAt  *time.Time        // 记录拒绝时间（可选）

	// 基础时间戳
	CreatedAt time.Time
	UpdatedAt time.Time // 始终记录最后更新时间
}

// ProjectAppointment.Isconfirmed() bool //已确认
func (t *ProjectAppointment) IsConfirmed() bool {
	return t.Status == StatusConfirmed
}

// ProjectAppointment.IsRejected() bool //已拒绝
func (t *ProjectAppointment) IsRejected() bool {
	return t.Status == StatusRejected
}

// ProjectAppointment.IsCancelled() bool //已取消
func (t *ProjectAppointment) IsCancelled() bool {
	return t.Status == StatusCancelled
}

// ProjectAppointment.IsPending() bool //待确认
func (t *ProjectAppointment) IsPending() bool {
	return t.Status == StatusPending
}

// ProjectAppointment.IsExpired() bool //24小时过期
func (t *ProjectAppointment) IsExpired() bool {
	return t.Status == StatusPending && t.CreatedAt.Add(24*time.Hour).Before(time.Now())
}

// ProjectAppointment.StatusString() string
func (t *ProjectAppointment) StatusString() string {
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

// ProjectAppointment.Create(ctx context.Context) (err error)
func (t *ProjectAppointment) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `INSERT INTO project_appointments (project_id, note, payer_team_id, payer_family_id, payee_team_id, payee_family_id, verifier_user_id,verifier_family_id,verifier_team_id, payer_user_id, payee_user_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`
	stmt, err := Db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.ProjectId, t.Note, t.PayerTeamId, t.PayerFamilyId, t.PayeeTeamId, t.PayeeFamilyId, t.VerifierUserId, t.VerifierFamilyId, t.VerifierTeamId, t.PayerUserId, t.PayeeUserId).Scan(&t.Id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out")
		}
		return
	}

	return
}
func (t *ProjectAppointment) Get(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, project_id, note,payer_team_id, payer_family_id, payee_team_id, payee_family_id, verifier_user_id, verifier_family_id, verifier_team_id, payer_user_id, payee_user_id, status, confirmed_at, rejected_at, created_at, updated_at FROM project_appointments WHERE id = $1`
	stmt, err := Db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.Id).Scan(&t.Id, &t.ProjectId, &t.Note, &t.PayerTeamId, &t.PayerFamilyId, &t.PayeeTeamId, &t.PayeeFamilyId, &t.VerifierUserId, &t.VerifierFamilyId, &t.VerifierTeamId, &t.PayerUserId, &t.PayeeUserId, &t.Status, &t.ConfirmedAt, &t.RejectedAt, &t.CreatedAt, &t.UpdatedAt)
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
	StatusPending   AppointmentStatus = iota // 0 待确认
	StatusConfirmed                          // 1 已确认
	StatusRejected                           // 2 已拒绝
	StatusCancelled                          // 3 已取消
)

// 约茶确认
func (a *ProjectAppointment) Confirm(confirmerID int) error {
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

// 取消预约
func (a *ProjectAppointment) Cancel(cancelerID int) error {
	if a.Status != StatusPending {
		return errors.New("只能取消待处理的预约")
	}
	a.Status = StatusCancelled
	now := time.Now()
	a.UpdatedAt = now
	return nil
}

// 拒绝预约
func (a *ProjectAppointment) Reject(rejecterID int) error {
	if a.Status != StatusPending {
		return errors.New("只能拒绝待处理的预约")
	}
	a.Status = StatusRejected
	now := time.Now()
	a.RejectedAt = &now
	a.UpdatedAt = now
	return nil
}

// project.AppointmentStatusString() string读取茶台预约状态
func (project *Project) AppointmentStatusString(ctx context.Context) string {
	projectAppointment := ProjectAppointment{ProjectId: project.Id}
	if err := projectAppointment.Get(ctx); err != nil {
		return "未知"
	}
	return projectAppointment.StatusString()

}

// projectAppointment.CreareAt() string读取茶台预约创建时间
func (projectAppointment *ProjectAppointment) CreareAt() string {
	return projectAppointment.CreatedAt.Format("2006-01-02 15:04:05")
}
