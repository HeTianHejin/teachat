package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

/*
在一个茶围objective目标里，某个项目project被选中“入围”之后，系统将询问围主是否需要启动线下作业服务以解决问题。
-否，意味无需线下行动，没有TeaOrder启动，继续线上讨论即可；
-是，启动tea_order实例，这是一个类似”大观园的诗社活动“，有见证方verify team（主持&裁判），需求方payer team（出题），解题方payee team（作答），一个解题的过程是一个handicraft（手工艺），
可能需要多个handicraft才能完成work，为了慎独，另外引入监护方（care team）与解题方共同承担责任风险；
--启动tea_order之后，系统会生成一个tea_order实体，记录该解题服务的相关信息；
--一个tea_order可以包含多个handicraft，每个handicraft对应一个具体的解题任务(见handicraft.go文件)；
(作为存档记录，tea_order涉及的茶围objective与项目project信息及参与方实体将固定（已经完成的order不可删除）以实现历史还原)
。。。
*/
type TeaOrder struct {
	Id           int
	Uuid         string
	ObjectiveId  int    // 茶围目标ID
	ProjectId    int    // 项目ID
	Status       string // tea_order状态：pending/active/pause/completed/cancelled
	VerifyTeamId int    // 见证方团队ID
	PayerTeamId  int    // 需求方（出题方）团队ID
	PayeeTeamId  int    // 解题方团队ID
	CareTeamId   int    // 监护方团队ID

	// 审批人（见证者）填写，必填
	// 审批人角色是类似大观园海棠诗社活动中的李纨社长角色，批准主题、主持活动及裁判“违规”情形，将阻止贾宝玉作西厢记类那种“男女礼教脱轨诗”或者禁止薛蟠那种酒色情诗；
	// 又或者是老师组织的多团队协作任务活动里的老师角色，不过在这茶会里不负责技术方面的审核，所以说“见证”记录事件发生的真实性、合规性。
	// 见证人也是活动进程主持人，类似教堂神父主持婚礼活动，发现不道德的欺瞒情况，例如新郎或者新娘竟然是重婚者之类不符合道德规范的活动将取消或者宣布无效。
	TeaTopic                string     // 茶会主题，默认值'-'。审批时即使不批准也应当根据茶围出题内容提炼，例如：热水器维修，宠物狗口腔护理,等。
	IsApproved              bool       // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          int        // 审批人ID，也是关联订单负责人，必须是见证者团队成员,如果是0代表待审批，
	ApprovalRejectionReason string     // 审批意见，如果拒绝，填写原因,默认值:'-'
	ApprovedAt              *time.Time // 审批时间

	FinalScore sql.NullInt64 // 根据另外的打分表计算后得到的最终的解题评分，NULL代表未评分，0代表得0分。
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time //软删除时间（未完成的tea_order可以被取消删除，已完成的tea_order不可删除）
}

const (
	TeaOrderStatusPending   = "pending"   // 待处理
	TeaOrderStatusActive    = "active"    // 活动中
	TeaOrderStatusPause     = "pause"     // 暂停
	TeaOrderStatusCompleted = "completed" // 完成
	TeaOrderStatusCancelled = "cancelled" // 已取消
)

// 见证日志
type WitnessLog struct {
	Id         int
	Uuid       string
	TeaOrderId int
	Action     string // "审批"/"暂停"/"恢复"/"终止"
	Reason     string
	EvidenceId int // 证据材料 ->Evidence{}
	WitnessAt  time.Time
}

const (
	WitnessActionApprove = "审批"
	WitnessActionPause   = "暂停"
	WitnessActionResume  = "恢复"
	WitnessActionCancel  = "终止"
)

// WitnessLog.Create() 创建见证日志
func (w *WitnessLog) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `INSERT INTO witness_logs (tea_order_id, action, reason, evidence_id, witness_at)
		VALUES ($1, $2, $3, $4, $5)`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, w.TeaOrderId, w.Action, w.Reason, w.EvidenceId, w.WitnessAt)
	return err
}

// WitnessLog.GetByTeaOrderId() 获取见证日志列表
func (w *WitnessLog) GetByTeaOrderId(ctx context.Context) ([]*WitnessLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, tea_order_id, action, reason, evidence_id, witness_at FROM witness_logs WHERE tea_order_id = $1 ORDER BY witness_at DESC`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, w.TeaOrderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	witnessLogs := make([]*WitnessLog, 0)
	for rows.Next() {
		witnessLog := &WitnessLog{}
		err := rows.Scan(&witnessLog.Id, &witnessLog.Uuid, &witnessLog.TeaOrderId, &witnessLog.Action, &witnessLog.Reason, &witnessLog.EvidenceId, &witnessLog.WitnessAt)
		if err != nil {
			return nil, err
		}
		witnessLogs = append(witnessLogs, witnessLog)
	}
	return witnessLogs, nil
}

// 根据状态获取茶订单列表
func GetTeaOrdersByStatus(ctx context.Context, status string, page int, pageSize int) ([]*TeaOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, objective_id, project_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, tea_topic, is_approved, approver_user_id, approval_rejection_reason, approved_at, final_score, created_at, updated_at, deleted_at FROM tea_orders WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, status, pageSize, page*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	teaOrders := make([]*TeaOrder, 0)
	for rows.Next() {
		teaOrder := &TeaOrder{}
		err := rows.Scan(&teaOrder.Id, &teaOrder.Uuid, &teaOrder.ObjectiveId, &teaOrder.ProjectId, &teaOrder.Status, &teaOrder.VerifyTeamId, &teaOrder.PayerTeamId, &teaOrder.PayeeTeamId, &teaOrder.CareTeamId, &teaOrder.TeaTopic, &teaOrder.IsApproved, &teaOrder.ApproverUserId, &teaOrder.ApprovalRejectionReason, &teaOrder.ApprovedAt, &teaOrder.FinalScore, &teaOrder.CreatedAt, &teaOrder.UpdatedAt, &teaOrder.DeletedAt)
		if err != nil {
			return nil, err
		}
		teaOrders = append(teaOrders, teaOrder)
	}
	return teaOrders, nil
}

// 根据状态获取茶订单数量
func GetTeaOrderCountByStatus(ctx context.Context, status string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM tea_orders WHERE status = $1`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRowContext(ctx, status).Scan(&count)
	return count, err
}

// 获取待审批订单数量（用于徽章提示）
func GetPendingTeaOrderCount(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM tea_orders WHERE status = $1`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRowContext(ctx, TeaOrderStatusPending).Scan(&count)
	return count, err
}

// Create 创建新的茶订单记录
func (t *TeaOrder) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO tea_orders (objective_id, project_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, t.ObjectiveId, t.ProjectId, t.Status, t.VerifyTeamId, t.PayerTeamId, t.PayeeTeamId, t.CareTeamId, t.TeaTopic, t.IsApproved, t.ApproverUserId, t.ApprovalRejectionReason)
	return err
}

// GetByIdOrUUID 根据ID或UUID获取茶订单记录
func (t *TeaOrder) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if t.Id <= 0 && t.Uuid == "" {
		return errors.New("invalid TeaOrder ID or UUID")
	}
	statement := `SELECT id, uuid, objective_id, project_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, tea_topic, is_approved, approver_user_id, approval_rejection_reason, approved_at, final_score, created_at, updated_at, deleted_at FROM tea_orders WHERE id = $1 OR uuid = $2`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.Id, t.Uuid).Scan(&t.Id, &t.Uuid, &t.ObjectiveId, &t.ProjectId, &t.Status, &t.VerifyTeamId, &t.PayerTeamId, &t.PayeeTeamId, &t.CareTeamId, &t.TeaTopic, &t.IsApproved, &t.ApproverUserId, &t.ApprovalRejectionReason, &t.ApprovedAt, &t.FinalScore, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	return err
}

// Update 更新茶订单记录
func (t *TeaOrder) Update() error {
	statement := `UPDATE tea_orders SET status = $2, final_score = $3, updated_at = $4 
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(t.Id, t.Status, t.FinalScore, time.Now())
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
