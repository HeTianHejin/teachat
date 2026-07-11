package dao

import (
	"fmt"
	"time"
)

// TeaOrderDeposit 代表茶订单中的预备金（星茶）托管记录
type TeaOrderDeposit struct {
	Id               int
	Uuid             string
	TeaOrderId       int                 // 茶围订单ID
	Type             TeaOrderDepositType // 款项类型：1：约茶，2：探茶，3：看看，4：验茶，5：脑火,6：手工艺，7：其他
	PayerTeamId      int                 // 支付团队ID，款项来源方
	BankTeamId       int                 // 托管团队ID，款项托管方
	PayeeTeamId      int                 // 解题团队ID，款项最终接收方
	AmountMilligrams int64               // 托管星茶数量，以 毫克（0.001克） 为单位

	// 关联的转账记录（用于追踪星茶流向）
	TransferOutId int //  支付方→托管的转出记录,REFERENCES tea.team_to_team_transfer_out(id)
	TransferInId  int //  托管方的接收记录,REFERENCES tea.team_from_team_transfer_in(id)

	Status     TeaOrderDepositStatus
	Notes      string // 备注说明
	HasDispute bool   // 是否存在争议（快速查询标识）

	//时间节点
	ExpiredAt *time.Time // 支付过期时间，（超时自动取消）
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time // 软删除时间
}

// TeaOrderDepositStatus 托管状态枚举
type TeaOrderDepositStatus int

const (
	DepositStatusPendingPayment  TeaOrderDepositStatus = iota // 待支付处理
	DepositStatusPendingDeposit                               // 待托管处理
	DepositStatusPaid                                         // 已支付,托管方已接收成功
	DepositStatusReleasedToPayee                              // 已释放给解题方
	DepositStatusRefundedToPayer                              // 已退款给需求方
	DepositStatusRefundedToPayee                              // 已退款给解题方(见证人对无恶意但超出预设讨论范围的约茶的处理，退款星茶原路退回双方团队)
	DepositStatusDisputed                                     // 争议中(解题方与需求方争议,需要仲裁)
	DepositStatusCancelled                                    // 已取消(订单审批被撤销)
	DepositStatusForfeited                                    // 已罚没(见证人对违规恶意/不道德行为的处罚，罚没星茶转入系统特殊团队“公共治理团队”)
)

type TeaOrderDepositType int

// TeaOrderDepositType 款项类型枚举
const (
	DepositTypeTeaAppointment = iota + 1 // 约茶
	DepositTypeTeaExplore                // 探茶
	DepositTypeTeaSeeSeek                // 看看
	DepositTypeTeaExamine                // 验茶
	DepositTypeBrainFire                 // 脑火
	DepositTypeHandicraft                // 手工艺
	DepositTypePreparation               // 预备金（入围阶段双方各托管的星茶，审批通过后归入下一流程结算）
	DepositTypeOther                     // 其他
)

// StatusString 返回托管状态的中文描述
func (tod *TeaOrderDeposit) StatusString() string {
	switch tod.Status {
	case DepositStatusPendingPayment:
		return "待支付处理"
	case DepositStatusPendingDeposit:
		return "待托管处理"
	case DepositStatusPaid:
		return "已支付托管中"
	case DepositStatusReleasedToPayee:
		return "已释放给解题方"
	case DepositStatusRefundedToPayer:
		return "已退款给需求方"
	case DepositStatusRefundedToPayee:
		return "已退款给解题方"
	case DepositStatusDisputed:
		return "争议中"
	case DepositStatusCancelled:
		return "已取消"
	case DepositStatusForfeited:
		return "已罚没"
	default:
		return "未知状态"
	}
}

// TypeString 返回款项类型的中文描述
func (tod *TeaOrderDeposit) TypeString() string {
	switch tod.Type {
	case DepositTypeTeaAppointment:
		return "约茶"
	case DepositTypeTeaExplore:
		return "探茶"
	case DepositTypeTeaSeeSeek:
		return "看看"
	case DepositTypeTeaExamine:
		return "验茶"
	case DepositTypeBrainFire:
		return "脑火"
	case DepositTypeHandicraft:
		return "手工艺"
	case DepositTypePreparation:
		return "预备金"
	case DepositTypeOther:
		return "其他"
	default:
		return "未知类型"
	}
}

// AmountGrams 返回托管星茶数量，以克为单位
func (tod *TeaOrderDeposit) AmountGrams() float64 {
	return float64(tod.AmountMilligrams) / 1000.0
}

// Create 创建新的预备金托管记录
func (tod *TeaOrderDeposit) Create() error {
	now := time.Now()
	statement := `INSERT INTO tea.tea_order_deposits
		(uuid, tea_order_id, type, payer_team_id, bank_team_id, payee_team_id, amount_milligrams,
		transfer_out_id, transfer_in_id, status, notes, has_dispute, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(
		tod.Uuid, tod.TeaOrderId, tod.Type, tod.PayerTeamId, tod.BankTeamId, tod.PayeeTeamId,
		tod.AmountMilligrams, tod.TransferOutId, tod.TransferInId, tod.Status,
		tod.Notes, tod.HasDispute, now, now,
	).Scan(&tod.Id, &tod.Uuid)
	if err != nil {
		return err
	}
	tod.CreatedAt = now
	return nil
}

// GetByTeaOrderId 根据茶订单ID获取所有预备金托管记录
func GetTeaOrderDepositsByTeaOrderId(teaOrderId int) ([]*TeaOrderDeposit, error) {
	statement := `SELECT id, uuid, tea_order_id, type, payer_team_id, bank_team_id, payee_team_id,
		amount_milligrams, transfer_out_id, transfer_in_id, status, notes, has_dispute,
		expired_at, created_at, updated_at, deleted_at
		FROM tea.tea_order_deposits WHERE tea_order_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := DB.Query(statement, teaOrderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deposits := make([]*TeaOrderDeposit, 0)
	for rows.Next() {
		d := &TeaOrderDeposit{}
		err := rows.Scan(
			&d.Id, &d.Uuid, &d.TeaOrderId, &d.Type, &d.PayerTeamId, &d.BankTeamId, &d.PayeeTeamId,
			&d.AmountMilligrams, &d.TransferOutId, &d.TransferInId, &d.Status,
			&d.Notes, &d.HasDispute, &d.ExpiredAt, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		deposits = append(deposits, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return deposits, nil
}

// GetById 根据ID获取预备金托管记录
func GetTeaOrderDepositById(id int) (*TeaOrderDeposit, error) {
	statement := `SELECT id, uuid, tea_order_id, type, payer_team_id, bank_team_id, payee_team_id,
		amount_milligrams, transfer_out_id, transfer_in_id, status, notes, has_dispute,
		expired_at, created_at, updated_at, deleted_at
		FROM tea.tea_order_deposits WHERE id = $1 AND deleted_at IS NULL`
	d := &TeaOrderDeposit{}
	err := DB.QueryRow(statement, id).Scan(
		&d.Id, &d.Uuid, &d.TeaOrderId, &d.Type, &d.PayerTeamId, &d.BankTeamId, &d.PayeeTeamId,
		&d.AmountMilligrams, &d.TransferOutId, &d.TransferInId, &d.Status,
		&d.Notes, &d.HasDispute, &d.ExpiredAt, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// UpdateStatus 更新预备金托管状态
func (tod *TeaOrderDeposit) UpdateStatus(status TeaOrderDepositStatus) error {
	now := time.Now()
	statement := `UPDATE tea.tea_order_deposits SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := DB.Exec(statement, tod.Id, status, now)
	if err != nil {
		return err
	}
	tod.Status = status
	tod.UpdatedAt = &now
	return nil
}

// UpdateTransferIds 更新关联的转账记录ID
func (tod *TeaOrderDeposit) UpdateTransferIds(transferOutId, transferInId int) error {
	now := time.Now()
	statement := `UPDATE tea.tea_order_deposits SET transfer_out_id = $2, transfer_in_id = $3, updated_at = $4 WHERE id = $1`
	_, err := DB.Exec(statement, tod.Id, transferOutId, transferInId, now)
	if err != nil {
		return err
	}
	tod.TransferOutId = transferOutId
	tod.TransferInId = transferInId
	tod.UpdatedAt = &now
	return nil
}

// Forfeit 罚没预备金：将托管的星茶从BankTeamId转入公共治理团队
func (tod *TeaOrderDeposit) Forfeit() error {
	if tod.Status != DepositStatusPaid {
		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 从托管团队账户扣除星茶
	_, err = tx.Exec(`
		UPDATE tea.team_accounts
		SET balance_milligrams = balance_milligrams - $1, updated_at = $2
		WHERE team_id = $3 AND balance_milligrams >= $1`,
		tod.AmountMilligrams, time.Now(), tod.BankTeamId)
	if err != nil {
		return err
	}

	// 转入公共治理团队账户
	_, err = tx.Exec(`
		UPDATE tea.team_accounts
		SET balance_milligrams = balance_milligrams + $1, updated_at = $2
		WHERE team_id = $3`,
		tod.AmountMilligrams, time.Now(), TeamIdPublicGovernance)
	if err != nil {
		return err
	}

	// 更新托管状态为已罚没
	_, err = tx.Exec(`
		UPDATE tea.tea_order_deposits SET status = $2, updated_at = $3 WHERE id = $1`,
		tod.Id, DepositStatusForfeited, time.Now())
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	tod.Status = DepositStatusForfeited
	return nil
}

// Refund 退回预备金：将托管的星茶退回给原支付方
// initiatorUserId：执行退款操作的用户ID（见证者），用于创建转账记录
// 注意：对于见证者拒绝场景，有两个 deposit 记录（需求方和解题方各一个），
// 每个 deposit 的 PayerTeamId 分别是其各自的团队 ID，因此两次 Refund 调用
// 会将星茶分别退回双方团队账户，并创建对称的转账记录。
func (tod *TeaOrderDeposit) Refund(initiatorUserId int) error {
	if tod.Status != DepositStatusPaid {
		return fmt.Errorf("退款失败：托管记录(id=%d)状态不是已支付，当前状态=%d", tod.Id, tod.Status)
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 获取托管团队（茶庄）和原支付方的团队名称，用于转账记录
	var bankTeamName, payerTeamName string
	err = tx.QueryRow(`SELECT name FROM teams WHERE id = $1`, tod.BankTeamId).Scan(&bankTeamName)
	if err != nil {
		return fmt.Errorf("退款失败：查询托管团队名称失败: %v", err)
	}
	err = tx.QueryRow(`SELECT name FROM teams WHERE id = $1`, tod.PayerTeamId).Scan(&payerTeamName)
	if err != nil {
		return fmt.Errorf("退款失败：查询原支付方团队名称失败: %v", err)
	}

	// 从托管团队账户扣除星茶，检查受影响行数确保余额充足
	result, err := tx.Exec(`
		UPDATE tea.team_accounts
		SET balance_milligrams = balance_milligrams - $1, updated_at = $2
		WHERE team_id = $3 AND balance_milligrams >= $1`,
		tod.AmountMilligrams, time.Now(), tod.BankTeamId)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("退款失败：查询托管团队扣款影响行数出错: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("退款失败：托管团队(team_id=%d)星茶余额不足，需要 %d 毫克", tod.BankTeamId, tod.AmountMilligrams)
	}

	// 查询托管团队扣款后的余额
	var bankBalanceAfter int64
	err = tx.QueryRow(`SELECT balance_milligrams FROM tea.team_accounts WHERE team_id = $1`, tod.BankTeamId).Scan(&bankBalanceAfter)
	if err != nil {
		return fmt.Errorf("退款失败：查询托管团队余额失败: %v", err)
	}

	// 退回给原支付方
	result, err = tx.Exec(`
		UPDATE tea.team_accounts
		SET balance_milligrams = balance_milligrams + $1, updated_at = $2
		WHERE team_id = $3`,
		tod.AmountMilligrams, time.Now(), tod.PayerTeamId)
	if err != nil {
		return err
	}
	rows, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("退款失败：查询原支付方加款影响行数出错: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("退款失败：原支付方团队(team_id=%d)账户不存在", tod.PayerTeamId)
	}

	// 查询原支付方加款后的余额
	var payerBalanceAfter int64
	err = tx.QueryRow(`SELECT balance_milligrams FROM tea.team_accounts WHERE team_id = $1`, tod.PayerTeamId).Scan(&payerBalanceAfter)
	if err != nil {
		return fmt.Errorf("退款失败：查询原支付方余额失败: %v", err)
	}

	// 创建托管团队→原支付方的转出记录（状态直接为 completed）
	refundNotes := fmt.Sprintf("退款：%s", tod.Notes)
	now := time.Now()
	var transferOutId int
	var transferOutUuid string
	err = tx.QueryRow(`
		INSERT INTO tea.team_to_team_transfer_out 
		(from_team_id, to_team_id, initiator_user_id, amount_milligrams, notes, 
		 from_team_name, to_team_name, status, balance_after_transfer, 
		 expires_at, is_approved, approver_user_id, approved_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true, $3, $11, $11)
		RETURNING id, uuid`,
		tod.BankTeamId, tod.PayerTeamId, initiatorUserId, tod.AmountMilligrams, refundNotes,
		bankTeamName, payerTeamName,
		TeaTransferStatusCompleted, bankBalanceAfter,
		now.Add(24*time.Hour), now).Scan(&transferOutId, &transferOutUuid)
	if err != nil {
		return fmt.Errorf("退款失败：创建托管团队转出记录失败: %v", err)
	}

	// 创建原支付方的接收记录（状态直接为 completed）
	_, err = tx.Exec(`
		INSERT INTO tea.team_from_team_transfer_in (
			team_to_team_transfer_out_id, to_team_id, to_team_name,
			from_team_id, from_team_name, amount_milligrams, notes,
			balance_after_receipt, status, is_confirmed, operational_user_id, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		transferOutId, tod.PayerTeamId, payerTeamName,
		tod.BankTeamId, bankTeamName,
		tod.AmountMilligrams, refundNotes,
		payerBalanceAfter, TeaTransferStatusCompleted,
		true, initiatorUserId, now, now.Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("退款失败：创建原支付方接收记录失败: %v", err)
	}

	// 更新托管状态为已退款给需求方
	_, err = tx.Exec(`
		UPDATE tea.tea_order_deposits SET status = $2, updated_at = $3 WHERE id = $1`,
		tod.Id, DepositStatusRefundedToPayer, now)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	tod.Status = DepositStatusRefundedToPayer
	return nil
}
