package dao

import (
	"database/sql"
	"fmt"
	"time"
)

/*
团队茶叶账户转账流程：
1、发起方法：团队转出额度茶叶，无论接收方是团队还是用户（个人），都要求由1成员填写转账表单，第一步创建待审核转账表单，
2、审核方法：由任意1核心成员审批待审核表单，
2.1、如果审核批准，转账表单状态更新为已批准，执行第3步锁定转账额度，
2.2、如果审核否决，转账表单状态更新为已否决，记录审批操作记录，流程结束。
3、锁定方法：已批准的转出团队账户转出额度茶叶数量被锁定，防止重复发起转账；
4.1、接收方法：目标接受方，用户或者团队任意1状态正常成员，有效期内，操作接收，继续第5步结算双方账户，
4.2、拒收方法，目标接受方，用户或者团队任意1状态正常成员，有效期内，操作拒收，转账表单状态更新为已拒收，解锁被转出方锁定茶叶，记录拒收操作用户id及时间原因，流程结束。
5、结算方法：按锁定额度（接收额度）清算双方账户数额，创建出入流水记录；
6、超时处理：自动解锁转出用户账户被锁定额度茶叶，不创建交易流水明细记录。
*/

// // 团队茶叶账户状态常量
const (
	TeaTeamAccountStatus_Normal  = "normal"
	TeaTeamAccountStatus_Frozen  = "frozen"
	TeaTeamAccountStatus_Deleted = "deleted"
)

// 团队转账类型常量
const (
	TransferType_TeamInitiated        = "team_initiated"         // 团队发起转账（单人团队自动审批）
	TransferType_TeamApprovalRequired = "team_approval_required" // 团队转账（需要审批）
)

// 团队茶叶账户结构体
type TeaTeamAccount struct {
	Id                 int
	Uuid               string
	TeamId             int
	BalanceGrams       float64 // 茶叶数量(克)
	LockedBalanceGrams float64 // 被锁定的茶叶数量(克)
	Status             string  // normal, frozen
	FrozenReason       *string
	CreatedAt          time.Time
	UpdatedAt          *time.Time
}

// 团队茶叶转账结构体（（单人团队自动批准,>2需内部核心成员审批，不能自己审批自己）
// payer转出，发起流程
// 转帐发起操作记录（从转出方视角）
// 注意不能转出0/负数，不能转给自己和自由人团队id=TeamIdFreelancer(2)
type TeaTeamTransferOut struct {
	Id         int
	Uuid       string
	FromTeamId int // 转出方，户主，团队id

	// 系统默认提交用户是，转出方发起操作人id
	InitiatorUserId int // 必须是团队成员

	// 操作人填写，接收方（必填其一）
	ToUserId    *int    // 用户接收
	ToTeamId    *int    // 团队接收
	AmountGrams float64 // 转账额度数量（克），也是锁定额度
	Notes       string  // 转账备注

	// StatusPendingApproval  = "pending_approval"  // 待审批（团队转出）
	// StatusApproved         = "approved"          // 审批通过
	// StatusApprovalRejected = "approval_rejected" // 审批拒绝
	// StatusPendingReceipt   = "pending_receipt"   // 待接收
	// StatusCompleted        = "completed"         // 已完成
	// StatusRejected         = "rejected"          // 以拒收
	// StatusExpired          = "expired"           // 已超时
	Status string // 统一状态枚举，系统自动管理

	// 审批相关（团队转出时使用）
	TransferType string // 1、单人团队自动批准'team_initiated', 2、多人团队审批'team_approval_required'
	// 审批人填写，A/R必填二选一
	ApproverUserId          *int       // 审批人ID
	ApprovedAt              *time.Time // 审批时间
	ApprovalRejectionReason *string    // 审批拒绝原因
	RejectedBy              *int       // 拒绝人ID
	RejectedAt              *time.Time // 拒绝时间

	// 余额记录
	BalanceAfterTransfer float64 // 转出后团队账户余额

	// 时间管理
	CreatedAt   time.Time  // 创建，流程开始时间，也是锁定额度起始时间
	ExpiresAt   time.Time  // 过期时间，也是锁定额度截止时间
	PaymentTime *time.Time // 实际支付时间（已批准+已接收）
	UpdatedAt   *time.Time
}

// GetTeaTeamAccountByTeamId 根据团队ID获取茶叶账户
func GetTeaTeamAccountByTeamId(teamId int) (TeaTeamAccount, error) {
	// 自由人团队没有茶叶资产，返回特殊的冻结账户
	if teamId == TeamIdFreelancer {
		reason := "自由人团队不支持茶叶资产"
		account := TeaTeamAccount{
			TeamId:       TeamIdFreelancer,
			BalanceGrams: 0.0,
			Status:       TeaTeamAccountStatus_Frozen,
			FrozenReason: &reason,
		}
		return account, nil
	}

	account := TeaTeamAccount{}
	err := DB.QueryRow("SELECT id, uuid, team_id, balance_grams, locked_balance_grams, status, frozen_reason, created_at, updated_at FROM tea.team_accounts WHERE team_id = $1", teamId).
		Scan(&account.Id, &account.Uuid, &account.TeamId, &account.BalanceGrams, &account.LockedBalanceGrams, &account.Status, &account.FrozenReason, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return account, fmt.Errorf("团队茶叶账户不存在")
		}
		return account, fmt.Errorf("查询团队茶叶账户失败: %v", err)
	}
	return account, nil
}

// Create 创建团队茶叶账户
func (account *TeaTeamAccount) Create() error {
	statement := "INSERT INTO tea.team_accounts (team_id, balance_grams, status) VALUES ($1, $2, $3) RETURNING id, uuid"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(account.TeamId, account.BalanceGrams, account.Status).Scan(&account.Id, &account.Uuid)
	if err != nil {
		return fmt.Errorf("创建团队茶叶账户失败: %v", err)
	}
	return nil
}

// UpdateStatus 更新账户状态
func (account *TeaTeamAccount) UpdateStatus(status, reason string) error {
	statement := "UPDATE tea.team_accounts SET status = $2, frozen_reason = $3, updated_at = $4 WHERE id = $1"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(account.Id, status, reason, time.Now())
	if err != nil {
		return fmt.Errorf("更新账户状态失败: %v", err)
	}

	account.Status = status
	if reason != "" {
		account.FrozenReason = &reason
	} else {
		account.FrozenReason = nil
	}
	return nil
}

// EnsureTeaTeamAccountExists 确保团队有茶叶账户
func EnsureTeaTeamAccountExists(teamId int) error {
	// 自由人团队不应该有茶叶资产
	if teamId == TeamIdFreelancer {
		return nil
	}

	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM tea.team_accounts WHERE team_id = $1)", teamId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查团队账户存在性失败: %v", err)
	}

	if !exists {
		account := &TeaTeamAccount{
			TeamId:       teamId,
			BalanceGrams: 0.0,
			Status:       TeaTeamAccountStatus_Normal,
		}
		return account.Create()
	}

	return nil
}

// CheckTeamAccountFrozen 检查团队账户是否被冻结
func CheckTeamAccountFrozen(teamId int) (bool, string, error) {
	// 自由人团队没有茶叶资产，视为冻结状态
	if teamId == TeamIdFreelancer {
		return true, "自由人团队不支持茶叶资产", nil
	}

	var status string
	var frozenReason sql.NullString
	err := DB.QueryRow("SELECT status, frozen_reason FROM tea.team_accounts WHERE team_id = $1", teamId).
		Scan(&status, &frozenReason)
	if err != nil {
		return false, "", fmt.Errorf("查询团队账户状态失败: %v", err)
	}

	if status == TeaTeamAccountStatus_Frozen {
		return true, frozenReason.String, nil
	}

	return false, "", nil
}

// GetTeamTeaTransactions 获取团队交易历史（从转出表和转入表中查询）
func GetTeamTeaTransactions(teamId int, page, limit int) ([]map[string]interface{}, error) {
	offset := (page - 1) * limit

	// 查询团队作为转出方的交易历史
	rows, err := DB.Query(`
		SELECT 
			'outgoing' as transaction_type,
			uuid,
			from_team_id,
			initiator_user_id,
			to_user_id,
			to_team_id,
			amount_grams,
			status,
			notes,
			payment_time,
			created_at
		FROM tea.team_transfer_out 
		WHERE from_team_id = $1 AND status = 'completed'
		
		UNION ALL
		
		-- 查询团队作为接收方的交易历史（从团队转出表）
		SELECT 
			'incoming' as transaction_type,
			uuid,
			from_team_id,
			initiator_user_id,
			to_user_id,
			to_team_id,
			amount_grams,
			status,
			notes,
			payment_time,
			created_at
		FROM tea.team_transfer_out 
		WHERE to_team_id = $1 AND status = 'completed'
		
		UNION ALL
		
		-- 查询团队作为接收方的交易历史（从用户转出表 + 转入表）
		SELECT 
			'incoming' as transaction_type,
			uto.uuid,
			NULL as from_team_id,
			uto.from_user_id as initiator_user_id,
			uto.to_user_id,
			uto.to_team_id,
			uto.amount_grams,
			uto.status,
			uto.notes,
			uto.payment_time,
			uto.created_at
		FROM tea.user_transfer_out uto
		INNER JOIN tea.transfer_in ti ON uto.id = ti.user_transfer_out_id
		WHERE uto.to_team_id = $1 AND uto.status = 'completed' AND ti.status = 'completed'
		
		UNION ALL
		
		-- 查询团队作为接收方的交易历史（从团队转出表 + 转入表）
		SELECT 
			'incoming' as transaction_type,
			tto.uuid,
			tto.from_team_id,
			tto.initiator_user_id,
			tto.to_user_id,
			tto.to_team_id,
			tto.amount_grams,
			tto.status,
			tto.notes,
			tto.payment_time,
			tto.created_at
		FROM tea.team_transfer_out tto
		INNER JOIN tea.transfer_in ti ON tto.id = ti.team_transfer_out_id
		WHERE tto.to_team_id = $1 AND tto.status = 'completed' AND ti.status = 'completed'
		
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, teamId, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("查询团队交易历史失败: %v", err)
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var transactionType, uuid, status, notes string
		var fromTeamId, initiatorUserId sql.NullInt64
		var toUserId, toTeamId sql.NullInt64
		var amountGrams float64
		var paymentTime, createdAt sql.NullTime

		err = rows.Scan(&transactionType, &uuid, &fromTeamId, &initiatorUserId, &toUserId, &toTeamId,
			&amountGrams, &status, &notes, &paymentTime, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("扫描团队交易记录失败: %v", err)
		}

		transaction := map[string]interface{}{
			"transaction_type":  transactionType,
			"uuid":              uuid,
			"from_team_id":      getNullableInt64(fromTeamId),
			"initiator_user_id": getNullableInt64(initiatorUserId),
			"to_user_id":        getNullableInt64(toUserId),
			"to_team_id":        getNullableInt64(toTeamId),
			"amount_grams":      amountGrams,
			"status":            status,
			"notes":             notes,
			"payment_time":      getNullableTime(paymentTime),
			"created_at":        getNullableTime(createdAt),
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// CreateTeamTeaTransaction 创建团队交易流水记录（已废弃，不再使用交易流水表）
// func CreateTeamTeaTransaction(teamId int, transactionType string, amountGrams, balanceBefore, balanceAfter float64, description string, targetUserId, targetTeamId, operatorUserId, approverUserId *int, targetType string) error {
// 	// 注：交易流水表已废弃，不再记录交易流水
// 	return nil
// }

// CanUserManageTeamAccount 检查用户是否可以管理团队账户
func CanUserManageTeamAccount(userId, teamId int) (bool, error) {
	team := Team{Id: teamId}
	// 检查是否是团队核心成员
	isCoreMember, err := team.IsCoreMember(userId)
	if err != nil {
		return false, fmt.Errorf("检查核心成员身份失败: %v", err)
	}

	return isCoreMember, nil
}

// IsTeamMember 检查用户是否是团队成员
func IsTeamMember(userId, teamId int) (bool, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM team_members 
		WHERE team_id = $1 AND user_id = $2 AND status = $3
	`, teamId, userId, TeamMemberStatusActive).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserJoinedTeams 获取用户加入的团队列表
func GetUserJoinedTeams(userId int) ([]Team, error) {
	query := `
		SELECT t.id, t.uuid, t.name, t.mission, t.founder_id, 
		       t.created_at, t.class, t.abbreviation, t.logo, 
		       t.is_private, t.updated_at, t.deleted_at, t.tags
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1 
		  AND tm.status = $2 
		  AND t.deleted_at IS NULL
		  AND t.id != $3
		ORDER BY t.created_at DESC
	`

	rows, err := DB.Query(query, userId, TeamMemberStatusActive, TeamIdFreelancer)
	if err != nil {
		return nil, fmt.Errorf("查询用户团队失败: %v", err)
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		var team Team
		err := rows.Scan(
			&team.Id, &team.Uuid, &team.Name, &team.Mission,
			&team.FounderId, &team.CreatedAt, &team.Class,
			&team.Abbreviation, &team.Logo, &team.IsPrivate,
			&team.UpdatedAt, &team.DeletedAt, &team.Tags,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描团队数据失败: %v", err)
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历团队数据失败: %v", err)
	}

	return teams, nil
}

// CreateTeaTransferTeamToUser 创建团队向用户转账记录
func CreateTeaTransferTeamToUser(fromTeamId, initiatorUserId, toUserId int, amount float64, notes string, expireHours int) (TeaTeamTransferOut, error) {
	// 验证参数
	if amount <= 0 {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：转账金额必须大于0")
	}
	if fromTeamId == TeamIdFreelancer {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：自由人团队不能发起转账")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出团队账户余额和锁定金额
	var teamBalance, teamLockedBalance float64
	var teamStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", fromTeamId).
		Scan(&teamBalance, &teamLockedBalance, &teamStatus)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：查询转出团队账户失败 - %v", err)
	}

	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := teamBalance - teamLockedBalance
	if availableBalance < amount {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：团队可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", teamBalance, teamLockedBalance, availableBalance)
	}

	// 检查发起人是否是团队成员
	isMember, err := IsTeamMember(initiatorUserId, fromTeamId)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：只有团队成员才能发起团队转账")
	}

	// 确保接收方用户账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea.user_accounts WHERE user_id = $1", toUserId).Scan(&toAccountId)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：接收方用户账户不存在 - %v", err)
	}

	// 确定转账类型（单人团队自动批准，多人团队需要审批）
	transferType := TransferType_TeamApprovalRequired
	teamMemberCount, err := getTeamMemberCount(tx, fromTeamId)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("查询团队成员数量失败: %v", err)
	}
	if teamMemberCount == 1 {
		transferType = TransferType_TeamInitiated
	}

	// 创建团队操作记录
	operation := TeaTeamTransferOut{
		FromTeamId:      fromTeamId,
		InitiatorUserId: initiatorUserId,
		ToUserId:        &toUserId,
		ToTeamId:        nil,
		AmountGrams:     amount,
		Notes:           notes,
		Status:          StatusPendingApproval,
		TransferType:    transferType,
		ExpiresAt:       time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:       time.Now(),
	}

	// 如果是单人团队，自动批准
	if transferType == TransferType_TeamInitiated {
		operation.Status = StatusApproved
		operation.ApproverUserId = &initiatorUserId
		approvedTime := time.Now()
		operation.ApprovedAt = &approvedTime
	}

	err = tx.QueryRow(`INSERT INTO tea.team_transfer_out 
		(from_team_id, initiator_user_id, to_user_id, amount_grams, notes, status, transfer_type, approver_user_id, approved_at, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid`,
		fromTeamId, initiatorUserId, toUserId, amount, notes, operation.Status, transferType, operation.ApproverUserId, operation.ApprovedAt, operation.ExpiresAt).
		Scan(&operation.Id, &operation.Uuid)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：创建团队转出记录失败 - %v", err)
	}

	// 锁定团队账户的相应金额
	_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE team_id = $3",
		amount, time.Now(), fromTeamId)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：锁定团队账户金额失败 - %v", err)
	}

	// 如果是单人团队，直接创建待接收的转账记录
	if transferType == TransferType_TeamInitiated {
		err = UpdateTeamTransferOutRecordStatusPendingReceipt(tx, operation)
		if err != nil {
			return TeaTeamTransferOut{}, fmt.Errorf("创建团队转出记录失败: %v", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：提交事务失败 - %v", err)
	}

	return operation, nil
}

// CreateTeaTransferTeamToTeam 创建团队向团队转账记录
func CreateTeaTransferTeamToTeam(fromTeamId, initiatorUserId, toTeamId int, amount float64, notes string, expireHours int) (TeaTeamTransferOut, error) {
	// 验证参数
	if amount <= 0 {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：转账金额必须大于0")
	}
	if fromTeamId == TeamIdFreelancer || toTeamId == TeamIdFreelancer {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：自由人团队不能参与转账")
	}
	if fromTeamId == toTeamId {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：不能向自己团队转账")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出团队账户余额和锁定金额
	var teamBalance, teamLockedBalance float64
	var teamStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", fromTeamId).
		Scan(&teamBalance, &teamLockedBalance, &teamStatus)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：查询转出团队账户失败 - %v", err)
	}

	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := teamBalance - teamLockedBalance
	if availableBalance < amount {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：团队可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", teamBalance, teamLockedBalance, availableBalance)
	}

	// 检查发起人是否是团队成员
	isMember, err := IsTeamMember(initiatorUserId, fromTeamId)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：只有团队成员才能发起团队转账")
	}

	// 确保接收方团队账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea.team_accounts WHERE team_id = $1", toTeamId).Scan(&toAccountId)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：接收方团队账户不存在 - %v", err)
	}

	// 确定转账类型（单人团队自动批准，多人团队需要审批）
	transferType := TransferType_TeamApprovalRequired
	teamMemberCount, err := getTeamMemberCount(tx, fromTeamId)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("查询团队成员数量失败: %v", err)
	}
	if teamMemberCount == 1 {
		transferType = TransferType_TeamInitiated
	}

	// 创建团队操作记录
	operation := TeaTeamTransferOut{
		FromTeamId:      fromTeamId,
		InitiatorUserId: initiatorUserId,
		ToUserId:        nil,
		ToTeamId:        &toTeamId,
		AmountGrams:     amount,
		Notes:           notes,
		Status:          StatusPendingApproval,
		TransferType:    transferType,
		ExpiresAt:       time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:       time.Now(),
	}

	// 如果是单人团队，自动批准
	if transferType == TransferType_TeamInitiated {
		operation.Status = StatusApproved
		operation.ApproverUserId = &initiatorUserId
		approvedTime := time.Now()
		operation.ApprovedAt = &approvedTime
	}

	err = tx.QueryRow(`INSERT INTO tea.team_transfer_out 
		(from_team_id, initiator_user_id, to_team_id, amount_grams, notes, status, transfer_type, approver_user_id, approved_at, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid`,
		fromTeamId, initiatorUserId, toTeamId, amount, notes, operation.Status, transferType, operation.ApproverUserId, operation.ApprovedAt, operation.ExpiresAt).
		Scan(&operation.Id, &operation.Uuid)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：创建团队转出记录失败 - %v", err)
	}

	// 锁定团队账户的相应金额
	_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE team_id = $3",
		amount, time.Now(), fromTeamId)
	if err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：锁定团队账户金额失败 - %v", err)
	}

	// 如果是单人团队，直接创建待接收的转账记录
	if transferType == TransferType_TeamInitiated {
		err = UpdateTeamTransferOutRecordStatusPendingReceipt(tx, operation)
		if err != nil {
			return TeaTeamTransferOut{}, fmt.Errorf("创建团队转出记录失败: %v", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaTeamTransferOut{}, fmt.Errorf("转账失败：提交事务失败 - %v", err)
	}

	return operation, nil
}

// getTeamMemberCount 获取团队成员数量
func getTeamMemberCount(tx *sql.Tx, teamId int) (int, error) {
	var count int
	err := tx.QueryRow(`
		SELECT COUNT(*) FROM team_members 
		WHERE team_id = $1 AND status = $2
	`, teamId, TeamMemberStatusActive).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateTeamTransferOutRecordStatusPendingReceipt 更新团队转出记录状态为待接收（审批通过后使用）
func UpdateTeamTransferOutRecordStatusPendingReceipt(tx *sql.Tx, operation TeaTeamTransferOut) error {
	// 更新团队转出记录状态为待接收
	_, err := tx.Exec(`UPDATE tea.team_transfer_out SET 
		status = $1, updated_at = $2 
		WHERE id = $3`,
		StatusPendingReceipt, time.Now(), operation.Id)
	if err != nil {
		return fmt.Errorf("更新团队转出记录状态失败: %v", err)
	}

	return nil
}

// ApproveTeamTransfer 审批团队转账
func ApproveTeamTransfer(operationUuid string, approverUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("审批失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取操作记录
	var operation TeaTeamTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, initiator_user_id, to_user_id, to_team_id, amount_grams, status 
		FROM tea.team_transfer_out WHERE uuid = $1 FOR UPDATE`, operationUuid).
		Scan(&operation.Id, &operation.Uuid, &operation.FromTeamId, &operation.InitiatorUserId,
			&operation.ToUserId, &operation.ToTeamId, &operation.AmountGrams, &operation.Status)
	if err != nil {
		return fmt.Errorf("审批失败：操作记录不存在 - %v", err)
	}

	// 验证状态
	if operation.Status != StatusPendingApproval {
		return fmt.Errorf("审批失败：操作状态异常")
	}

	// 检查审批人是否是团队成员（不能自己审批自己）
	isMember, err := IsTeamMember(approverUserId, operation.FromTeamId)
	if err != nil {
		return fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return fmt.Errorf("审批失败：只有团队成员才能审批")
	}

	// 检查不能自己审批自己
	if approverUserId == operation.InitiatorUserId {
		return fmt.Errorf("审批失败：不能自己审批自己发起的操作")
	}

	// 更新操作状态为已批准
	approvedAt := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_transfer_out SET 
		status = $1, approver_user_id = $2, approved_at = $3, updated_at = $4 
		WHERE id = $5`,
		StatusApproved, approverUserId, approvedAt, approvedAt, operation.Id)
	if err != nil {
		return fmt.Errorf("审批失败：更新操作状态失败 - %v", err)
	}

	// 创建团队转出记录
	err = UpdateTeamTransferOutRecordStatusPendingReceipt(tx, operation)
	if err != nil {
		return fmt.Errorf("创建团队转出记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// RejectTeamTransfer 拒绝团队转账
func RejectTeamTransfer(operationUuid string, approverUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("拒绝失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取操作记录
	var operation TeaTeamTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, initiator_user_id, amount_grams, status 
		FROM tea.team_transfer_out WHERE uuid = $1 FOR UPDATE`, operationUuid).
		Scan(&operation.Id, &operation.Uuid, &operation.FromTeamId, &operation.InitiatorUserId,
			&operation.AmountGrams, &operation.Status)
	if err != nil {
		return fmt.Errorf("拒绝失败：操作记录不存在 - %v", err)
	}

	// 验证状态
	if operation.Status != StatusPendingApproval {
		return fmt.Errorf("拒绝失败：操作状态异常")
	}

	// 检查审批人是否是团队成员（不能自己审批自己）
	isMember, err := IsTeamMember(approverUserId, operation.FromTeamId)
	if err != nil {
		return fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return fmt.Errorf("拒绝失败：只有团队成员才能审批")
	}

	// 检查不能自己审批自己
	if approverUserId == operation.InitiatorUserId {
		return fmt.Errorf("拒绝失败：不能自己审批自己发起的操作")
	}

	// 获取团队账户锁定金额
	var teamLockedBalance float64
	err = tx.QueryRow("SELECT locked_balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", operation.FromTeamId).Scan(&teamLockedBalance)
	if err != nil {
		return fmt.Errorf("查询团队账户锁定金额失败: %v", err)
	}

	// 解锁团队账户的锁定金额
	newLockedBalance := teamLockedBalance - operation.AmountGrams
	if newLockedBalance < 0 {
		newLockedBalance = 0
	}

	_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_grams = $1, updated_at = $2 WHERE team_id = $3",
		newLockedBalance, time.Now(), operation.FromTeamId)
	if err != nil {
		return fmt.Errorf("解锁团队账户金额失败: %v", err)
	}

	// 更新操作状态为已拒绝
	rejectedAt := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_transfer_out SET 
		status = $1, approver_user_id = $2, approval_rejection_reason = $3, rejected_by = $4, rejected_at = $5, updated_at = $6 
		WHERE id = $7`,
		StatusApprovalRejected, approverUserId, reason, approverUserId, rejectedAt, rejectedAt, operation.Id)
	if err != nil {
		return fmt.Errorf("拒绝失败：更新操作状态失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetPendingTeamOperations 获取团队待审批操作列表
func GetPendingTeamOperations(teamId int, page, limit int) ([]TeaTeamTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`SELECT id, uuid, from_team_id, initiator_user_id, to_user_id, to_team_id, 
		amount_grams, status, notes, transfer_type, balance_after_transfer, expires_at, created_at 
		FROM tea.team_transfer_out 
		WHERE from_team_id = $1 AND status = $2 AND expires_at > NOW() 
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		teamId, StatusPendingApproval, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待审批操作失败: %v", err)
	}
	defer rows.Close()

	var operations []TeaTeamTransferOut
	for rows.Next() {
		var operation TeaTeamTransferOut
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.FromTeamId, &operation.InitiatorUserId,
			&operation.ToUserId, &operation.ToTeamId, &operation.AmountGrams, &operation.Status,
			&operation.Notes, &operation.TransferType, &operation.BalanceAfterTransfer, &operation.ExpiresAt, &operation.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// GetPendingTeamTransfers 获取团队待接收转账列表
func GetPendingTeamTransfers(teamId int, page, limit int) ([]TeaTeamTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`SELECT id, uuid, from_team_id, initiator_user_id, to_user_id, to_team_id, 
		amount_grams, status, notes, transfer_type, balance_after_transfer, expires_at, created_at 
		FROM tea.team_transfer_out 
		WHERE to_team_id = $1 AND status = $2 AND expires_at > NOW() 
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		teamId, StatusPendingReceipt, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待接收转账失败: %v", err)
	}
	defer rows.Close()

	var operations []TeaTeamTransferOut
	for rows.Next() {
		var operation TeaTeamTransferOut
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.FromTeamId, &operation.InitiatorUserId,
			&operation.ToUserId, &operation.ToTeamId, &operation.AmountGrams, &operation.Status,
			&operation.Notes, &operation.TransferType, &operation.BalanceAfterTransfer, &operation.ExpiresAt, &operation.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// CountPendingTeamTransfers 获取团队待确认接收操作数量
func CountPendingTeamTransfers(teamId int) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM tea.team_transfer_out 
		WHERE to_team_id = $1 AND status = $2 AND expires_at > NOW()`,
		teamId, StatusPendingReceipt).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询待确认接收操作数量失败: %v", err)
	}
	return count, nil
}

// ConfirmTeamTransfer 确认接收团队转账
func ConfirmTeamTransfer(transferUuid string, confirmUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("确认转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaTeamTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, to_team_id, to_user_id, amount_grams, status, expires_at 
		FROM tea.team_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.ToTeamId, &transfer.ToUserId,
			&transfer.AmountGrams, &transfer.Status, &transfer.ExpiresAt)
	if err != nil {
		return fmt.Errorf("确认转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != StatusPendingReceipt {
		return fmt.Errorf("确认转账失败：转账状态异常")
	}
	if time.Now().After(transfer.ExpiresAt) {
		// 转账已过期，更新状态
		_, _ = tx.Exec("UPDATE tea.team_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			StatusExpired, time.Now(), transfer.Id)
		return fmt.Errorf("确认转账失败：转账已过期")
	}

	// 检查确认权限
	if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
		// 团队转账：检查用户是否是团队成员
		isMember, err := IsTeamMember(confirmUserId, *transfer.ToTeamId)
		if err != nil {
			return fmt.Errorf("检查团队成员身份失败: %v", err)
		}
		if !isMember {
			return fmt.Errorf("只有团队成员才能确认团队转账")
		}
	} else if transfer.ToUserId != nil && *transfer.ToUserId > 0 {
		// 用户转账：检查接收用户ID
		if *transfer.ToUserId != confirmUserId {
			return fmt.Errorf("无权确认此转账")
		}
	}

	// 根据转账类型执行不同的确认逻辑
	if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
		err = confirmTeamToTeamTransfer(tx, transfer, confirmUserId)
	} else if transfer.ToUserId != nil && *transfer.ToUserId > 0 {
		err = confirmTeamToUserTransfer(tx, transfer, confirmUserId)
	} else {
		return fmt.Errorf("转账目标不明确")
	}

	if err != nil {
		return err
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// confirmTeamToTeamTransfer 确认团队向团队转账
func confirmTeamToTeamTransfer(tx *sql.Tx, transfer TeaTeamTransferOut, confirmUserId int) error {
	// 确保接收方团队账户存在
	var toTeamBalance float64
	err := tx.QueryRow("SELECT balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", *transfer.ToTeamId).Scan(&toTeamBalance)
	if err != nil {
		return fmt.Errorf("查询接收团队账户余额失败: %v", err)
	}

	// 获取转出团队账户信息
	var fromTeamBalance, fromTeamLockedBalance float64
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", transfer.FromTeamId).
		Scan(&fromTeamBalance, &fromTeamLockedBalance)
	if err != nil {
		return fmt.Errorf("查询转出团队账户信息失败: %v", err)
	}

	// 检查余额是否足够
	if fromTeamLockedBalance < transfer.AmountGrams {
		return fmt.Errorf("锁定余额不足，无法完成转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromTeamLockedBalance, transfer.AmountGrams)
	}
	if fromTeamBalance < transfer.AmountGrams {
		return fmt.Errorf("账户余额不足，无法完成转账。当前余额: %.3f克, 转账金额: %.3f克", fromTeamBalance, transfer.AmountGrams)
	}

	// 更新账户余额
	newFromTeamBalance := fromTeamBalance - transfer.AmountGrams
	newFromTeamLockedBalance := fromTeamLockedBalance - transfer.AmountGrams

	_, err = tx.Exec("UPDATE tea.team_accounts SET balance_grams = $1, locked_balance_grams = $2, updated_at = $3 WHERE team_id = $4",
		newFromTeamBalance, newFromTeamLockedBalance, time.Now(), transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("更新转出团队账户余额失败: %v", err)
	}

	_, err = tx.Exec("UPDATE tea.team_accounts SET balance_grams = $1, updated_at = $2 WHERE team_id = $3",
		toTeamBalance+transfer.AmountGrams, time.Now(), *transfer.ToTeamId)
	if err != nil {
		return fmt.Errorf("更新接收团队账户余额失败: %v", err)
	}

	// 更新转账状态，设置实际支付时间，记录转出后余额
	paymentTime := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_transfer_out SET 
		status = $1, 
		payment_time = $2, 
		balance_after_transfer = $3,
		updated_at = $4 
		WHERE id = $5`,
		StatusCompleted, paymentTime, newFromTeamBalance, paymentTime, transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建转入记录（团队转账确认时创建转入记录，关联到确认用户，记录接收后余额）
	_, err = tx.Exec(`INSERT INTO tea.transfer_in 
		(holder_id, team_transfer_out_id, status, balance_after_receipt, confirmed_by, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		confirmUserId, transfer.Id, StatusCompleted, toTeamBalance+transfer.AmountGrams, confirmUserId, paymentTime)
	if err != nil {
		return fmt.Errorf("创建转入记录失败: %v", err)
	}

	// 记录交易流水
	// err = recordTeamToTeamTransferTransactions(tx, transfer, fromTeamBalance, newFromTeamBalance, toTeamBalance)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// confirmTeamToUserTransfer 确认团队向用户转账
func confirmTeamToUserTransfer(tx *sql.Tx, transfer TeaTeamTransferOut, confirmUserId int) error {
	// 确保接收方用户账户存在
	var toUserBalance float64
	err := tx.QueryRow("SELECT balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", *transfer.ToUserId).Scan(&toUserBalance)
	if err != nil {
		return fmt.Errorf("查询接收用户账户余额失败: %v", err)
	}

	// 获取转出团队账户信息
	var fromTeamBalance, fromTeamLockedBalance float64
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", transfer.FromTeamId).
		Scan(&fromTeamBalance, &fromTeamLockedBalance)
	if err != nil {
		return fmt.Errorf("查询转出团队账户信息失败: %v", err)
	}

	// 检查余额是否足够
	if fromTeamLockedBalance < transfer.AmountGrams {
		return fmt.Errorf("锁定余额不足，无法完成转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromTeamLockedBalance, transfer.AmountGrams)
	}
	if fromTeamBalance < transfer.AmountGrams {
		return fmt.Errorf("账户余额不足，无法完成转账。当前余额: %.3f克, 转账金额: %.3f克", fromTeamBalance, transfer.AmountGrams)
	}

	// 更新账户余额
	newFromTeamBalance := fromTeamBalance - transfer.AmountGrams
	newFromTeamLockedBalance := fromTeamLockedBalance - transfer.AmountGrams

	_, err = tx.Exec("UPDATE tea.team_accounts SET balance_grams = $1, locked_balance_grams = $2, updated_at = $3 WHERE team_id = $4",
		newFromTeamBalance, newFromTeamLockedBalance, time.Now(), transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("更新转出团队账户余额失败: %v", err)
	}

	_, err = tx.Exec("UPDATE tea.user_accounts SET balance_grams = $1, updated_at = $2 WHERE user_id = $3",
		toUserBalance+transfer.AmountGrams, time.Now(), *transfer.ToUserId)
	if err != nil {
		return fmt.Errorf("更新接收用户账户余额失败: %v", err)
	}

	// 更新转账状态，设置实际支付时间，记录转出后余额
	paymentTime := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_transfer_out SET 
		status = $1, 
		payment_time = $2, 
		balance_after_transfer = $3,
		updated_at = $4 
		WHERE id = $5`,
		StatusCompleted, paymentTime, newFromTeamBalance, paymentTime, transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建转入记录（团队向用户转账确认时创建转入记录，关联到确认用户，记录接收后余额）
	_, err = tx.Exec(`INSERT INTO tea.transfer_in 
		(holder_id, team_transfer_out_id, status, balance_after_receipt, confirmed_by, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		*transfer.ToUserId, transfer.Id, StatusCompleted, toUserBalance+transfer.AmountGrams, confirmUserId, paymentTime)
	if err != nil {
		return fmt.Errorf("创建转入记录失败: %v", err)
	}

	return nil
}

// recordTeamToTeamTransferTransactions 记录团队向团队转账的交易流水（已废弃，不再使用交易流水表）
// func recordTeamToTeamTransferTransactions(transfer TeaTeamTransferOut, fromBalance, newFromBalance, toBalance float64) error {
// 	// 注：交易流水表已废弃，不再记录交易流水
// 	return nil
// }

// recordTeamToUserTransferTransactions 记录团队向用户转账的交易流水（已废弃，不再使用交易流水表）
// func recordTeamToUserTransferTransactions(tx *sql.Tx, transfer TeaTeamTransferOut, fromBalance, newFromBalance, toBalance float64) error {
// 	// 注：交易流水表已废弃，不再记录交易流水
// 	return nil
// }

// RejectTeamTransferReceipt 拒绝接收团队转账
func RejectTeamTransferReceipt(transferUuid string, rejectUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("拒绝转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaTeamTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, to_team_id, to_user_id, amount_grams, status 
		FROM tea.team_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.ToTeamId, &transfer.ToUserId,
			&transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("拒绝转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != StatusPendingReceipt {
		return fmt.Errorf("拒绝转账失败：转账状态异常")
	}

	// 检查拒绝权限
	if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
		// 团队转账：检查用户是否是团队成员
		isMember, err := IsTeamMember(rejectUserId, *transfer.ToTeamId)
		if err != nil {
			return fmt.Errorf("检查团队成员身份失败: %v", err)
		}
		if !isMember {
			return fmt.Errorf("只有团队成员才能拒绝团队转账")
		}
	} else if transfer.ToUserId != nil && *transfer.ToUserId > 0 {
		// 用户转账：检查接收用户ID
		if *transfer.ToUserId != rejectUserId {
			return fmt.Errorf("无权拒绝此转账")
		}
	}

	// 获取转出团队账户的锁定金额信息
	var fromTeamLockedBalance float64
	err = tx.QueryRow("SELECT locked_balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", transfer.FromTeamId).Scan(&fromTeamLockedBalance)
	if err != nil {
		return fmt.Errorf("查询转出团队账户锁定金额失败: %v", err)
	}

	// 解锁转出团队账户的锁定金额
	newFromTeamLockedBalance := fromTeamLockedBalance - transfer.AmountGrams

	// 检查锁定余额是否足够解锁
	if newFromTeamLockedBalance < 0 {
		return fmt.Errorf("锁定余额不足，无法拒绝转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromTeamLockedBalance, transfer.AmountGrams)
	}

	_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_grams = $1, updated_at = $2 WHERE team_id = $3",
		newFromTeamLockedBalance, time.Now(), transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("解锁转出团队账户金额失败: %v", err)
	}

	// 更新转账状态
	_, err = tx.Exec(`UPDATE tea.team_transfer_out SET 
		status = $1, 
		updated_at = $2 
		WHERE id = $3`,
		StatusRejected, time.Now(), transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// 作为接收方“确认接收”操作历史
// GetTeamTransferInOperations 获取团队所有转入记录（接收）
func GetTeamTransferInOperations(teamId int, page, limit int) ([]TeaTransferIn, error) {
	offset := (page - 1) * limit

	// 查询团队转入记录（从transfer_in表，结合team_transfer_out表获取完整信息）
	rows, err := DB.Query(`
		SELECT ti.id, ti.uuid, ti.holder_id, ti.team_transfer_out_id,
			   ti.status, ti.balance_after_receipt, ti.confirmed_by, 
			   ti.rejected_by, ti.reception_rejection_reason, ti.created_at,
			   tto.from_team_id, tto.initiator_user_id, tto.to_team_id,
			   tto.to_user_id, tto.amount_grams, tto.notes, tto.payment_time
		FROM tea.transfer_in ti
		INNER JOIN tea.team_transfer_out tto ON ti.team_transfer_out_id = tto.id
		WHERE ti.holder_id = $1
		ORDER BY ti.created_at DESC LIMIT $2 OFFSET $3`,
		teamId, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("查询团队转入记录失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaTransferIn
	for rows.Next() {
		var transfer TeaTransferIn
		var fromTeamId, initiatorUserId sql.NullInt64
		var toTeamId, toUserId sql.NullInt64
		var amountGrams float64
		var notes string
		var paymentTime sql.NullTime

		err = rows.Scan(
			&transfer.Id, &transfer.Uuid, &transfer.HolderId, &transfer.TeamTransferOutId,
			&transfer.Status, &transfer.BalanceAfterReceipt, &transfer.ConfirmedBy,
			&transfer.RejectedBy, &transfer.ReceptionRejectionReason, &transfer.CreatedAt,
			&fromTeamId, &initiatorUserId, &toTeamId, &toUserId, &amountGrams, &notes, &paymentTime)
		if err != nil {
			return nil, fmt.Errorf("扫描转入记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetTeamTransferOutOperations 获取团队所有转出操作历史（包括各种状态）
func GetTeamTransferOutOperations(teamId int, page, limit int) ([]TeaTeamTransferOut, error) {
	offset := (page - 1) * limit

	// 查询团队作为转出方所有操作历史（包括各种状态）
	rows, err := DB.Query(`SELECT id, uuid, from_team_id, initiator_user_id, to_user_id, to_team_id, 
		amount_grams, status, notes, transfer_type, approver_user_id, approved_at, approval_rejection_reason, 
		rejected_by, rejected_at, balance_after_transfer, expires_at, payment_time, created_at 
		FROM tea.team_transfer_out 
		WHERE from_team_id = $1 OR to_team_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		teamId, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("查询团队操作历史失败: %v", err)
	}
	defer rows.Close()

	var operations []TeaTeamTransferOut
	for rows.Next() {
		var operation TeaTeamTransferOut
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.FromTeamId, &operation.InitiatorUserId,
			&operation.ToUserId, &operation.ToTeamId, &operation.AmountGrams, &operation.Status,
			&operation.Notes, &operation.TransferType, &operation.ApproverUserId, &operation.ApprovedAt,
			&operation.ApprovalRejectionReason, &operation.RejectedBy, &operation.RejectedAt,
			&operation.BalanceAfterTransfer, &operation.ExpiresAt, &operation.PaymentTime, &operation.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}
