package dao

import (
	"context"
	"errors"
	util "teachat/Util"
	"time"
)

// 约茶
type ProjectAppointment struct {
	Id        int
	Uuid      string
	ProjectId int
	Note      string //做备注，简明扼要地描述一下约会目的
	// 日期时间
	StartTime time.Time // 开始时间，默认为当前时间
	EndTime   time.Time // 结束时间，默认为开始时间+1小时

	PlaceId int // 约茶地方ID
	// 相关团队和家庭信息
	PayerUserId   int //出茶叶代表人Id
	PayerTeamId   int //出茶叶团队Id
	PayerFamilyId int //出茶叶家庭Id

	PayeeUserId   int //收茶叶代表人Id
	PayeeTeamId   int //收茶叶团队Id
	PayeeFamilyId int //收茶叶家庭Id

	VerifierUserId   int
	VerifierFamilyId int
	VerifierTeamId   int

	// 状态管理：
	// 0、"待约茶"
	// 1、已提交
	// 2、"已确认"
	// 3、"已拒绝"
	// 4、"已取消"
	Status      AppointmentStatus // 枚举类型
	ConfirmedAt *time.Time        // 明确记录确认时间
	RejectedAt  *time.Time        // 记录拒绝时间（可选）

	// 基础时间戳
	CreatedAt time.Time
	UpdatedAt time.Time // 始终记录最后更新时间
}

// ProjectAppointment.Isconfirmed() bool //已确认
func (t *ProjectAppointment) IsConfirmed() bool {
	return t.Status == AppointmentStatusConfirmed
}

// ProjectAppointment.IsRejected() bool //已拒绝
func (t *ProjectAppointment) IsRejected() bool {
	return t.Status == AppointmentStatusRejected
}

// ProjectAppointment.IsCancelled() bool //已取消
func (t *ProjectAppointment) IsCancelled() bool {
	return t.Status == AppointmentStatusCancelled
}

// ProjectAppointment.IsPending() bool //待约茶
func (t *ProjectAppointment) IsPending() bool {
	return t.Status == AppointmentStatusPending
}

// ProjectAppointment.IsExpired() bool //24小时过期
func (t *ProjectAppointment) IsExpired() bool {
	return t.Status == AppointmentStatusPending && t.CreatedAt.Add(24*time.Hour).Before(time.Now())
}

// ProjectAppointment.StatusString() string
func (t *ProjectAppointment) StatusString() string {
	switch t.Status {
	case AppointmentStatusPending:
		return "待约茶"
	case AppointmentStatusSubmitted:
		return "已提交"
	case AppointmentStatusConfirmed:
		return "已确认"
	case AppointmentStatusRejected:
		return "已拒绝"
	case AppointmentStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// ProjectAppointment.Create(ctx context.Context) (err error)
func (t *ProjectAppointment) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `INSERT INTO project_appointments (project_id, note, start_time, end_time, place_id, payer_team_id, payer_family_id, payee_team_id, payee_family_id, verifier_user_id, verifier_family_id, verifier_team_id, payer_user_id, payee_user_id, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id,created_at`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.ProjectId, t.Note, t.StartTime, t.EndTime, t.PlaceId, t.PayerTeamId, t.PayerFamilyId, t.PayeeTeamId, t.PayeeFamilyId, t.VerifierUserId, t.VerifierFamilyId, t.VerifierTeamId, t.PayerUserId, t.PayeeUserId, t.Status).Scan(&t.Id, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out", err)
		}
		return
	}

	return
}
func (t *ProjectAppointment) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, project_id, note, start_time, end_time, place_id, payer_team_id, payer_family_id, payee_team_id, payee_family_id, verifier_user_id, verifier_family_id, verifier_team_id, payer_user_id, payee_user_id, status, confirmed_at, rejected_at, created_at, updated_at FROM project_appointments WHERE id=$1 OR uuid=$2`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.Id, t.Uuid).Scan(&t.Id, &t.Uuid, &t.ProjectId, &t.Note, &t.StartTime, &t.EndTime, &t.PlaceId, &t.PayerTeamId, &t.PayerFamilyId, &t.PayeeTeamId, &t.PayeeFamilyId, &t.VerifierUserId, &t.VerifierFamilyId, &t.VerifierTeamId, &t.PayerUserId, &t.PayeeUserId, &t.Status, &t.ConfirmedAt, &t.RejectedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("query ProjectAppointment timed out", err)
		}
		return err
	}

	return nil
}

// projectAppointment.Update(ctx context.Context) (err error)
func (t *ProjectAppointment) Update(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `UPDATE project_appointments SET note=$1, start_time=$2, end_time=$3, place_id=$4, payer_team_id=$5, payer_family_id=$6, payee_team_id=$7, payee_family_id=$8, verifier_user_id=$9, verifier_family_id=$10, verifier_team_id=$11, payer_user_id=$12, payee_user_id=$13, status=$14, confirmed_at=$15, rejected_at=$16, updated_at=$17 WHERE id=$18`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, t.Note, t.StartTime, t.EndTime, t.PlaceId, t.PayerTeamId, t.PayerFamilyId, t.PayeeTeamId, t.PayeeFamilyId, t.VerifierUserId, t.VerifierFamilyId, t.VerifierTeamId, t.PayerUserId, t.PayeeUserId, t.Status, t.ConfirmedAt, t.RejectedAt, time.Now(), t.Id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out", err)
		}
		return
	}

	return
}

// 预约状态枚举
type AppointmentStatus int

const (
	AppointmentStatusPending   AppointmentStatus = iota // 0 待约茶
	AppointmentStatusSubmitted                          // 1 已提交
	AppointmentStatusConfirmed                          // 2 已确认
	AppointmentStatusRejected                           // 3 已拒绝
	AppointmentStatusCancelled                          // 4 已取消
)

// action.appoinmentStatusString() string读取茶台预约状态
func (project *Project) AppointmentStatusString(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	projectAppointment := ProjectAppointment{ProjectId: project.Id}
	err := db.QueryRowContext(ctx, `select status from project_appointments where project_id = $1`, project.Id).Scan(&projectAppointment.Status)
	if err != nil {
		return "未知"
	}

	return projectAppointment.StatusString()
}

// projectAppointment.CrearedAtString() string读取茶台预约创建时间
func (projectAppointment *ProjectAppointment) CrearedAtString() string {
	return projectAppointment.CreatedAt.Format("2006-01-02 15:04:05")
}

// projectAppointment.StartTime() string读取茶台预约开始时间
func (projectAppointment *ProjectAppointment) StartTimeString() string {
	return projectAppointment.StartTime.Format("2006-01-02 15:04:05")
}

// projectAppointment.EndTime() string读取茶台预约结束时间
func (projectAppointment *ProjectAppointment) EndTimeString() string {
	return projectAppointment.EndTime.Format("2006-01-02 15:04:05")
}

// GetAppointmentByProjectId() 读取茶台预约
func GetAppointmentByProjectId(project_id int, ctx context.Context) (p_a ProjectAppointment, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, project_id, note, start_time, end_time, place_id, payer_team_id, payer_family_id, payee_team_id, payee_family_id, verifier_user_id, verifier_family_id, verifier_team_id, payer_user_id, payee_user_id, status, confirmed_at, rejected_at, created_at, updated_at FROM project_appointments WHERE project_id = $1`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, project_id).Scan(&p_a.Id, &p_a.Uuid, &p_a.ProjectId, &p_a.Note, &p_a.StartTime, &p_a.EndTime, &p_a.PlaceId, &p_a.PayerTeamId, &p_a.PayerFamilyId, &p_a.PayeeTeamId, &p_a.PayeeFamilyId, &p_a.VerifierUserId, &p_a.VerifierFamilyId, &p_a.VerifierTeamId, &p_a.PayerUserId, &p_a.PayeeUserId, &p_a.Status, &p_a.ConfirmedAt, &p_a.RejectedAt, &p_a.CreatedAt, &p_a.UpdatedAt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			util.Debug("Query timed out", err)
		}
		return
	}

	return p_a, nil

}
