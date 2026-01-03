package dao

import (
	"context"
	"errors"
	"time"
)

/*
在一个茶围objective目标里，某个项目project被选中“入围”之后，系统将询问围主是否需要启动线下作业服务以解决问题。
-否，意味无需线下行动，没有TeaOrder启动，继续线上讨论即可；
-是，启动tea-order实例，这是一个类似”大观园的诗社活动“，有见证方verify team（主持&裁判），需求方payer team（出题），解题方payee team（作答），一个解题的过程是一个handicraft（手工艺），
可能需要多个handicraft才能完成work，为了慎独，另外引入监护方（care team）与解题方共同承担责任风险；
--启动tea-order之后，系统会生成一个tea-order实体，记录该解题服务的相关信息；
--一个tea-order可以包含多个handicraft，每个handicraft对应一个具体的解题任务(见handicraft.go文件)；
(作为存档记录，tea-order涉及的茶围objective与项目project信息及参与方实体将固定（已经完成的order不可删除）以实现历史还原)
。。。
*/
type TeaOrder struct {
	Id           int
	Uuid         string
	ObjectiveId  int    // 茶围目标ID
	ProjectId    int    // 项目ID
	Status       string // tea-order状态：pending/active/completed/cancelled
	VerifyTeamId int    // 见证方团队ID
	PayerTeamId  int    // 需求方（出题方）团队ID
	PayeeTeamId  int    // 解题方团队ID
	CareTeamId   int    // 监护方团队ID
	Score        int    // 解题评分
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time //软删除时间（未完成的tea-order可以被取消删除，已完成的tea-order不可删除）
}

const (
	TeaOrderStatusPending   = "pending"
	TeaOrderStatusActive    = "active"
	TeaOrderStatusCompleted = "completed"
	TeaOrderStatusCancelled = "cancelled"
)

// Create 创建新的茶订单记录
func (t *TeaOrder) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO tea_orders 
		(uuid, objective_id, project_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, score) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), t.ObjectiveId, t.ProjectId, t.Status, t.VerifyTeamId, t.PayerTeamId, t.PayeeTeamId, t.CareTeamId, t.Score).Scan(&t.Id, &t.Uuid)
	return err
}

// GetByIdOrUUID 根据ID或UUID获取茶订单记录
func (t *TeaOrder) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if t.Id <= 0 && t.Uuid == "" {
		return errors.New("invalid TeaOrder ID or UUID")
	}
	statement := `SELECT id, uuid, objective_id, project_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, score, created_at, updated_at, deleted_at
		FROM tea_orders WHERE (id=$1 OR uuid=$2) AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.Id, t.Uuid).Scan(&t.Id, &t.Uuid, &t.ObjectiveId, &t.ProjectId, &t.Status, &t.VerifyTeamId, &t.PayerTeamId, &t.PayeeTeamId, &t.CareTeamId, &t.Score, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	return err
}

// Update 更新茶订单记录
func (t *TeaOrder) Update() error {
	statement := `UPDATE tea_orders SET status = $2, score = $3, updated_at = $4 
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(t.Id, t.Status, t.Score, time.Now())
	return err
}

// Delete 软删除茶订单记录
func (t *TeaOrder) Delete() error {
	statement := `UPDATE tea_orders SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(t.Id, now)
	if err == nil {
		t.DeletedAt = &now
	}
	return err
}

// CreatedDateTime 格式化创建时间
func (t *TeaOrder) CreatedDateTime() string {
	return t.CreatedAt.Format(FMT_DATE_TIME_CN)
}
