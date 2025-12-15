package dao

import (
	"database/sql"
	"fmt"
	"time"
)

// 团队茶叶账户状态常量
const (
	TeamTeaAccountStatus_Normal = "normal"
	TeamTeaAccountStatus_Frozen = "frozen"
)

// 团队操作状态常量
const (
	TeamOperationStatus_Pending  = "pending"
	TeamOperationStatus_Approved = "approved"
	TeamOperationStatus_Rejected = "rejected"
	TeamOperationStatus_Expired  = "expired"
)

// 团队操作类型常量
const (
	TeamOperationType_Deposit     = "deposit"
	TeamOperationType_Withdraw    = "withdraw"
	TeamOperationType_TransferOut = "transfer_out"
	TeamOperationType_TransferIn  = "transfer_in"
)

// 团队交易类型常量
const (
	TeamTransactionType_Deposit      = "deposit"
	TeamTransactionType_Withdraw     = "withdraw"
	TeamTransactionType_TransferOut  = "transfer_out"
	TeamTransactionType_TransferIn   = "transfer_in"
	TeamTransactionType_SystemGrant  = "system_grant"
	TeamTransactionType_SystemDeduct = "system_deduct"
)

// 团队茶叶账户结构体
type TeamTeaAccount struct {
	Id           int
	Uuid         string
	TeamId       int
	BalanceGrams float64 // 茶叶数量(克)
	Status       string  // normal, frozen
	FrozenReason string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// 团队茶叶操作结构体
type TeamTeaOperation struct {
	Id              int
	Uuid            string
	TeamId          int
	OperationType   string // deposit, withdraw, transfer_out, transfer_in
	AmountGrams     float64
	Status          string // pending, approved, rejected, expired
	OperatorUserId  int
	ApproverUserId  *int
	TargetTeamId    *int
	TargetUserId    *int
	Notes           string
	RejectionReason *string
	ExpiresAt       time.Time
	ApprovedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

// 团队茶叶交易流水结构体
type TeamTeaTransaction struct {
	Id              int
	Uuid            string
	TeamId          int
	OperationId     *string
	TransactionType string // deposit, withdraw, transfer_out, transfer_in, system_grant, system_deduct
	AmountGrams     float64
	BalanceBefore   float64
	BalanceAfter    float64
	Description     string
	RelatedTeamId   *int
	RelatedUserId   *int
	CreatedAt       time.Time
}

// GetTeamTeaAccountByTeamId 根据团队ID获取茶叶账户
func GetTeamTeaAccountByTeamId(teamId int) (TeamTeaAccount, error) {
	// 自由人团队没有茶叶资产，返回特殊的冻结账户
	if teamId == TeamIdFreelancer {
		account := TeamTeaAccount{
			TeamId:       TeamIdFreelancer,
			BalanceGrams: 0.0,
			Status:       TeamTeaAccountStatus_Frozen,
			FrozenReason: "自由人团队不支持茶叶资产",
		}
		return account, nil
	}

	account := TeamTeaAccount{}
	err := db.QueryRow("SELECT id, uuid, team_id, balance_grams, status, COALESCE(frozen_reason, '') as frozen_reason, created_at, updated_at FROM team_tea_accounts WHERE team_id = $1", teamId).
		Scan(&account.Id, &account.Uuid, &account.TeamId, &account.BalanceGrams, &account.Status, &account.FrozenReason, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return account, fmt.Errorf("团队茶叶账户不存在")
		}
		return account, fmt.Errorf("查询团队茶叶账户失败: %v", err)
	}
	return account, nil
}

// Create 创建团队茶叶账户
func (account *TeamTeaAccount) Create() error {
	statement := "INSERT INTO team_tea_accounts (team_id, balance_grams, status) VALUES ($1, $2, $3) RETURNING id, uuid"
	stmt, err := db.Prepare(statement)
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
func (account *TeamTeaAccount) UpdateStatus(status, reason string) error {
	statement := "UPDATE team_tea_accounts SET status = $2, frozen_reason = $3, updated_at = $4 WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(account.Id, status, reason, time.Now())
	if err != nil {
		return fmt.Errorf("更新账户状态失败: %v", err)
	}

	account.Status = status
	account.FrozenReason = reason
	return nil
}

// CreateTeamTeaOperation 创建团队茶叶操作
func CreateTeamTeaOperation(teamId, operatorUserId int, operationType string, amount float64, notes string, expireHours int, targetTeamId, targetUserId *int) (TeamTeaOperation, error) {
	// 验证参数
	if amount <= 0 {
		return TeamTeaOperation{}, fmt.Errorf("操作金额必须大于0")
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return TeamTeaOperation{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查团队账户是否存在
	var accountExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM team_tea_accounts WHERE team_id = $1)", teamId).Scan(&accountExists)
	if err != nil {
		return TeamTeaOperation{}, fmt.Errorf("检查团队账户失败: %v", err)
	}
	if !accountExists {
		return TeamTeaOperation{}, fmt.Errorf("团队茶叶账户不存在")
	}

	// 检查团队成员数量
	var memberCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM team_members WHERE team_id = $1 AND status < $2", teamId, TeMemberStatusResigned).Scan(&memberCount)
	if err != nil {
		return TeamTeaOperation{}, fmt.Errorf("检查团队成员数量失败: %v", err)
	}

	// 如果是提取或转出操作，检查余额是否足够
	if operationType == TeamOperationType_Withdraw || operationType == TeamOperationType_TransferOut {
		var currentBalance float64
		err = tx.QueryRow("SELECT balance_grams FROM team_tea_accounts WHERE team_id = $1 FOR UPDATE", teamId).Scan(&currentBalance)
		if err != nil {
			return TeamTeaOperation{}, fmt.Errorf("查询当前余额失败: %v", err)
		}
		if currentBalance < amount {
			return TeamTeaOperation{}, fmt.Errorf("余额不足")
		}
	}

	// 创建操作记录
	var operationStatus string
	if memberCount == 1 {
		// 如果团队只有1个成员，直接批准操作
		operationStatus = TeamOperationStatus_Approved
	} else {
		// 多人团队，需要审批
		operationStatus = TeamOperationStatus_Pending
	}

	operation := TeamTeaOperation{
		TeamId:         teamId,
		OperationType:  operationType,
		AmountGrams:    amount,
		Status:         operationStatus,
		OperatorUserId: operatorUserId,
		TargetTeamId:   targetTeamId,
		TargetUserId:   targetUserId,
		Notes:          notes,
		ExpiresAt:      time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:      time.Now(),
	}

	err = tx.QueryRow(`INSERT INTO team_tea_operations 
		(team_id, operation_type, amount_grams, status, operator_user_id, target_team_id, target_user_id, notes, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, uuid`,
		teamId, operationType, amount, operationStatus, operatorUserId, targetTeamId, targetUserId, notes, operation.ExpiresAt).
		Scan(&operation.Id, &operation.Uuid)
	if err != nil {
		return TeamTeaOperation{}, fmt.Errorf("创建操作记录失败: %v", err)
	}

	// 如果是单人团队，立即执行操作
	if memberCount == 1 {
		// 设置审批人为操作人自己
		approverUserId := operatorUserId
		now := time.Now()

		_, err = tx.Exec("UPDATE team_tea_operations SET status = $1, approver_user_id = $2, approved_at = $3, updated_at = $4 WHERE id = $5",
			TeamOperationStatus_Approved, approverUserId, now, now, operation.Id)
		if err != nil {
			return TeamTeaOperation{}, fmt.Errorf("更新操作状态失败: %v", err)
		}

		// 执行操作
		err = executeTeamTeaOperationInTx(tx, operation, approverUserId)
		if err != nil {
			return TeamTeaOperation{}, fmt.Errorf("执行操作失败: %v", err)
		}

		// 重新加载操作信息
		err = tx.QueryRow(`SELECT id, uuid, team_id, operation_type, amount_grams, status, 
			operator_user_id, approver_user_id, approved_at FROM team_tea_operations WHERE id = $1`, operation.Id).
			Scan(&operation.Id, &operation.Uuid, &operation.TeamId, &operation.OperationType,
				&operation.AmountGrams, &operation.Status, &operation.OperatorUserId,
				&operation.ApproverUserId, &operation.ApprovedAt)
		if err != nil {
			return TeamTeaOperation{}, fmt.Errorf("重新加载操作信息失败: %v", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeamTeaOperation{}, fmt.Errorf("提交事务失败: %v", err)
	}

	return operation, nil
}

// executeTeamTeaOperationInTx 在事务中执行团队茶叶操作
func executeTeamTeaOperationInTx(tx *sql.Tx, operation TeamTeaOperation, approverUserId int) error {
	// 获取团队账户当前余额
	var currentBalance float64
	err := tx.QueryRow("SELECT balance_grams FROM team_tea_accounts WHERE team_id = $1 FOR UPDATE", operation.TeamId).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("查询当前余额失败: %v", err)
	}

	var newBalance float64
	var transactionType string

	// 根据操作类型处理
	switch operation.OperationType {
	case TeamOperationType_Deposit:
		newBalance = currentBalance + operation.AmountGrams
		transactionType = TeamTransactionType_Deposit
	case TeamOperationType_Withdraw, TeamOperationType_TransferOut:
		if currentBalance < operation.AmountGrams {
			return fmt.Errorf("余额不足")
		}
		newBalance = currentBalance - operation.AmountGrams
		transactionType = TeamTransactionType_Withdraw
		if operation.OperationType == TeamOperationType_TransferOut {
			transactionType = TeamTransactionType_TransferOut
		}
	case TeamOperationType_TransferIn:
		newBalance = currentBalance + operation.AmountGrams
		transactionType = TeamTransactionType_TransferIn
	default:
		return fmt.Errorf("不支持的操作类型")
	}

	// 如果是转出操作，需要处理目标方
	if operation.OperationType == TeamOperationType_TransferOut && operation.TargetTeamId != nil {
		// 转账到其他团队账户
		var targetBalance float64
		err = tx.QueryRow("SELECT balance_grams FROM team_tea_accounts WHERE team_id = $1 FOR UPDATE", *operation.TargetTeamId).Scan(&targetBalance)
		if err != nil {
			return fmt.Errorf("查询目标团队余额失败: %v", err)
		}

		// 更新目标团队余额
		targetNewBalance := targetBalance + operation.AmountGrams
		_, err = tx.Exec("UPDATE team_tea_accounts SET balance_grams = $2, updated_at = $3 WHERE team_id = $1",
			*operation.TargetTeamId, targetNewBalance, time.Now())
		if err != nil {
			return fmt.Errorf("更新目标团队余额失败: %v", err)
		}

		// 记录目标团队交易流水
		_, err = tx.Exec(`INSERT INTO team_tea_transactions 
			(team_id, operation_id, transaction_type, amount_grams, balance_before, balance_after, description, related_team_id, related_user_id) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			*operation.TargetTeamId, &operation.Uuid, TeamTransactionType_TransferIn, operation.AmountGrams,
			targetBalance, targetNewBalance, "团队转账转入", &operation.TeamId, &approverUserId)
		if err != nil {
			return fmt.Errorf("记录目标团队交易流水失败: %v", err)
		}
	}

	// 转账到用户账户
	if operation.OperationType == TeamOperationType_TransferOut && operation.TargetUserId != nil {
		// 确保目标用户有茶叶账户
		err = EnsureTeaAccountExists(*operation.TargetUserId)
		if err != nil {
			return fmt.Errorf("确保用户茶叶账户失败: %v", err)
		}

		var userBalance float64
		err = tx.QueryRow("SELECT balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", *operation.TargetUserId).Scan(&userBalance)
		if err != nil {
			return fmt.Errorf("查询用户余额失败: %v", err)
		}

		// 更新用户余额
		userNewBalance := userBalance + operation.AmountGrams
		_, err = tx.Exec("UPDATE tea_accounts SET balance_grams = $2, updated_at = $3 WHERE user_id = $1",
			*operation.TargetUserId, userNewBalance, time.Now())
		if err != nil {
			return fmt.Errorf("更新用户余额失败: %v", err)
		}

		// 记录用户交易流水
		_, err = tx.Exec(`INSERT INTO tea_transactions 
			(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			*operation.TargetUserId, &operation.Uuid, TransactionType_TransferIn, operation.AmountGrams,
			userBalance, userNewBalance, "团队转账转入", nil)
		if err != nil {
			return fmt.Errorf("记录用户交易流水失败: %v", err)
		}
	}

	// 更新团队余额
	_, err = tx.Exec("UPDATE team_tea_accounts SET balance_grams = $2, updated_at = $3 WHERE team_id = $1",
		operation.TeamId, newBalance, time.Now())
	if err != nil {
		return fmt.Errorf("更新团队余额失败: %v", err)
	}

	// 记录交易流水
	description := "茶叶存入"
	switch operation.OperationType {
	case TeamOperationType_Withdraw:
		description = "茶叶提取"
	case TeamOperationType_TransferOut:
		description = "茶叶转出"
	case TeamOperationType_TransferIn:
		description = "茶叶转入"
	}

	_, err = tx.Exec(`INSERT INTO team_tea_transactions 
		(team_id, operation_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		operation.TeamId, &operation.Uuid, transactionType, operation.AmountGrams,
		currentBalance, newBalance, description, &approverUserId)
	if err != nil {
		return fmt.Errorf("记录交易流水失败: %v", err)
	}

	return nil
}

// ApproveTeamTeaOperation 审批团队茶叶操作
func ApproveTeamTeaOperation(operationUuid string, approverUserId int) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取操作信息
	var operation TeamTeaOperation
	err = tx.QueryRow(`SELECT id, uuid, team_id, operation_type, amount_grams, status, expires_at, 
		target_team_id, target_user_id FROM team_tea_operations WHERE uuid = $1 FOR UPDATE`, operationUuid).
		Scan(&operation.Id, &operation.Uuid, &operation.TeamId, &operation.OperationType,
			&operation.AmountGrams, &operation.Status, &operation.ExpiresAt,
			&operation.TargetTeamId, &operation.TargetUserId)
	if err != nil {
		return fmt.Errorf("操作记录不存在: %v", err)
	}

	// 验证状态
	if operation.Status != TeamOperationStatus_Pending {
		return fmt.Errorf("操作状态异常")
	}
	if time.Now().After(operation.ExpiresAt) {
		// 操作已过期，更新状态
		_, _ = tx.Exec("UPDATE team_tea_operations SET status = $1, updated_at = $2 WHERE id = $3",
			TeamOperationStatus_Expired, time.Now(), operation.Id)
		return fmt.Errorf("操作已过期")
	}

	// 获取团队账户当前余额
	var currentBalance float64
	err = tx.QueryRow("SELECT balance_grams FROM team_tea_accounts WHERE team_id = $1 FOR UPDATE", operation.TeamId).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("查询当前余额失败: %v", err)
	}

	var newBalance float64
	var transactionType string

	// 根据操作类型处理
	switch operation.OperationType {
	case TeamOperationType_Deposit:
		newBalance = currentBalance + operation.AmountGrams
		transactionType = TeamTransactionType_Deposit
	case TeamOperationType_Withdraw, TeamOperationType_TransferOut:
		if currentBalance < operation.AmountGrams {
			return fmt.Errorf("余额不足")
		}
		newBalance = currentBalance - operation.AmountGrams
		transactionType = TeamTransactionType_Withdraw
		if operation.OperationType == TeamOperationType_TransferOut {
			transactionType = TeamTransactionType_TransferOut
		}
	case TeamOperationType_TransferIn:
		newBalance = currentBalance + operation.AmountGrams
		transactionType = TeamTransactionType_TransferIn
	default:
		return fmt.Errorf("不支持的操作类型")
	}

	// 如果是转出操作，需要处理目标方
	if operation.OperationType == TeamOperationType_TransferOut && operation.TargetTeamId != nil {
		// 转账到其他团队账户
		var targetBalance float64
		err = tx.QueryRow("SELECT balance_grams FROM team_tea_accounts WHERE team_id = $1 FOR UPDATE", *operation.TargetTeamId).Scan(&targetBalance)
		if err != nil {
			return fmt.Errorf("查询目标团队余额失败: %v", err)
		}

		// 更新目标团队余额
		targetNewBalance := targetBalance + operation.AmountGrams
		_, err = tx.Exec("UPDATE team_tea_accounts SET balance_grams = $2, updated_at = $3 WHERE team_id = $1",
			*operation.TargetTeamId, targetNewBalance, time.Now())
		if err != nil {
			return fmt.Errorf("更新目标团队余额失败: %v", err)
		}

		// 记录目标团队交易流水
		_, err = tx.Exec(`INSERT INTO team_tea_transactions 
			(team_id, operation_id, transaction_type, amount_grams, balance_before, balance_after, description, related_team_id, related_user_id) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			*operation.TargetTeamId, &operation.Uuid, TeamTransactionType_TransferIn, operation.AmountGrams,
			targetBalance, targetNewBalance, "团队转账转入", &operation.TeamId, &approverUserId)
		if err != nil {
			return fmt.Errorf("记录目标团队交易流水失败: %v", err)
		}
	}

	// 转账到用户账户
	if operation.OperationType == TeamOperationType_TransferOut && operation.TargetUserId != nil {
		// 确保目标用户有茶叶账户
		err = EnsureTeaAccountExists(*operation.TargetUserId)
		if err != nil {
			return fmt.Errorf("确保用户茶叶账户失败: %v", err)
		}

		var userBalance float64
		err = tx.QueryRow("SELECT balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", *operation.TargetUserId).Scan(&userBalance)
		if err != nil {
			return fmt.Errorf("查询用户余额失败: %v", err)
		}

		// 更新用户余额
		userNewBalance := userBalance + operation.AmountGrams
		_, err = tx.Exec("UPDATE tea_accounts SET balance_grams = $2, updated_at = $3 WHERE user_id = $1",
			*operation.TargetUserId, userNewBalance, time.Now())
		if err != nil {
			return fmt.Errorf("更新用户余额失败: %v", err)
		}

		// 记录用户交易流水
		_, err = tx.Exec(`INSERT INTO tea_transactions 
			(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			*operation.TargetUserId, &operation.Uuid, TransactionType_TransferIn, operation.AmountGrams,
			userBalance, userNewBalance, "团队转账转入", nil)
		if err != nil {
			return fmt.Errorf("记录用户交易流水失败: %v", err)
		}
	}

	// 更新团队余额
	_, err = tx.Exec("UPDATE team_tea_accounts SET balance_grams = $2, updated_at = $3 WHERE team_id = $1",
		operation.TeamId, newBalance, time.Now())
	if err != nil {
		return fmt.Errorf("更新团队余额失败: %v", err)
	}

	// 更新操作状态
	approvedAt := time.Now()
	_, err = tx.Exec("UPDATE team_tea_operations SET status = $1, approver_user_id = $2, approved_at = $3, updated_at = $4 WHERE id = $5",
		TeamOperationStatus_Approved, approverUserId, approvedAt, approvedAt, operation.Id)
	if err != nil {
		return fmt.Errorf("更新操作状态失败: %v", err)
	}

	// 记录交易流水
	description := "茶叶存入"
	switch operation.OperationType {
	case TeamOperationType_Withdraw:
		description = "茶叶提取"
	case TeamOperationType_TransferOut:
		description = "茶叶转出"
	case TeamOperationType_TransferIn:
		description = "茶叶转入"
	}

	_, err = tx.Exec(`INSERT INTO team_tea_transactions 
		(team_id, operation_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		operation.TeamId, &operation.Uuid, transactionType, operation.AmountGrams,
		currentBalance, newBalance, description, &approverUserId)
	if err != nil {
		return fmt.Errorf("记录交易流水失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// RejectTeamTeaOperation 拒绝团队茶叶操作
func RejectTeamTeaOperation(operationUuid string, approverUserId int, reason string) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取操作信息
	var operation TeamTeaOperation
	err = tx.QueryRow("SELECT id, status FROM team_tea_operations WHERE uuid = $1 FOR UPDATE", operationUuid).
		Scan(&operation.Id, &operation.Status)
	if err != nil {
		return fmt.Errorf("操作记录不存在: %v", err)
	}

	// 验证状态
	if operation.Status != TeamOperationStatus_Pending {
		return fmt.Errorf("操作状态异常")
	}

	// 更新操作状态
	_, err = tx.Exec("UPDATE team_tea_operations SET status = $1, approver_user_id = $2, rejection_reason = $3, updated_at = $4 WHERE id = $5",
		TeamOperationStatus_Rejected, approverUserId, reason, time.Now(), operation.Id)
	if err != nil {
		return fmt.Errorf("更新操作状态失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetTeamPendingOperations 获取团队待审批操作列表
func GetTeamPendingOperations(teamId int, page, limit int) ([]TeamTeaOperation, error) {
	offset := (page - 1) * limit
	rows, err := db.Query(`SELECT id, uuid, team_id, operation_type, amount_grams, status, 
		operator_user_id, approver_user_id, target_team_id, target_user_id, notes, 
		rejection_reason, expires_at, approved_at, created_at, updated_at 
		FROM team_tea_operations WHERE team_id = $1 AND status = $2 AND expires_at > NOW() 
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		teamId, TeamOperationStatus_Pending, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待审批操作失败: %v", err)
	}
	defer rows.Close()

	var operations []TeamTeaOperation
	for rows.Next() {
		var operation TeamTeaOperation
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.TeamId, &operation.OperationType,
			&operation.AmountGrams, &operation.Status, &operation.OperatorUserId, &operation.ApproverUserId,
			&operation.TargetTeamId, &operation.TargetUserId, &operation.Notes,
			&operation.RejectionReason, &operation.ExpiresAt, &operation.ApprovedAt, &operation.CreatedAt, &operation.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// GetTeamTeaTransactions 获取团队交易流水
func GetTeamTeaTransactions(teamId int, page, limit int, transactionType string) ([]TeamTeaTransaction, error) {
	offset := (page - 1) * limit
	var rows *sql.Rows
	var err error

	if transactionType == "" {
		rows, err = db.Query(`SELECT id, uuid, team_id, operation_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, related_team_id, related_user_id, created_at 
			FROM team_tea_transactions WHERE team_id = $1 
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`, teamId, limit, offset)
	} else {
		rows, err = db.Query(`SELECT id, uuid, team_id, operation_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, related_team_id, related_user_id, created_at 
			FROM team_tea_transactions WHERE team_id = $1 AND transaction_type = $2 
			ORDER BY created_at DESC LIMIT $3 OFFSET $4`, teamId, transactionType, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("查询交易流水失败: %v", err)
	}
	defer rows.Close()

	var transactions []TeamTeaTransaction
	for rows.Next() {
		var transaction TeamTeaTransaction
		err = rows.Scan(&transaction.Id, &transaction.Uuid, &transaction.TeamId, &transaction.OperationId,
			&transaction.TransactionType, &transaction.AmountGrams, &transaction.BalanceBefore,
			&transaction.BalanceAfter, &transaction.Description, &transaction.RelatedTeamId, &transaction.RelatedUserId, &transaction.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描交易流水失败: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetTeamTeaOperations 获取团队操作历史
func GetTeamTeaOperations(teamId int, page, limit int) ([]TeamTeaOperation, error) {
	offset := (page - 1) * limit
	rows, err := db.Query(`SELECT id, uuid, team_id, operation_type, amount_grams, status, 
		operator_user_id, approver_user_id, target_team_id, target_user_id, notes, 
		rejection_reason, expires_at, approved_at, created_at, updated_at 
		FROM team_tea_operations WHERE team_id = $1 
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, teamId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询操作历史失败: %v", err)
	}
	defer rows.Close()

	var operations []TeamTeaOperation
	for rows.Next() {
		var operation TeamTeaOperation
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.TeamId, &operation.OperationType,
			&operation.AmountGrams, &operation.Status, &operation.OperatorUserId, &operation.ApproverUserId,
			&operation.TargetTeamId, &operation.TargetUserId, &operation.Notes,
			&operation.RejectionReason, &operation.ExpiresAt, &operation.ApprovedAt, &operation.CreatedAt, &operation.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// EnsureTeamTeaAccountExists 确保团队有茶叶账户
func EnsureTeamTeaAccountExists(teamId int) error {
	// 自由人团队不应该有茶叶资产
	if teamId == TeamIdFreelancer {
		return nil
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM team_tea_accounts WHERE team_id = $1)", teamId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查团队账户存在性失败: %v", err)
	}

	if !exists {
		account := &TeamTeaAccount{
			TeamId:       teamId,
			BalanceGrams: 0.0,
			Status:       TeamTeaAccountStatus_Normal,
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
	err := db.QueryRow("SELECT status, frozen_reason FROM team_tea_accounts WHERE team_id = $1", teamId).
		Scan(&status, &frozenReason)
	if err != nil {
		return false, "", fmt.Errorf("查询团队账户状态失败: %v", err)
	}

	if status == TeamTeaAccountStatus_Frozen {
		return true, frozenReason.String, nil
	}

	return false, "", nil
}

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

// GetTeamTeaOperationByUuid 根据UUID获取团队茶叶操作
func GetTeamTeaOperationByUuid(uuid string) (TeamTeaOperation, error) {
	operation := TeamTeaOperation{}
	err := db.QueryRow(`SELECT id, uuid, team_id, operation_type, amount_grams, status, 
		operator_user_id, approver_user_id, target_team_id, target_user_id, notes, 
		rejection_reason, expires_at, approved_at, created_at, updated_at 
		FROM team_tea_operations WHERE uuid = $1`, uuid).
		Scan(&operation.Id, &operation.Uuid, &operation.TeamId, &operation.OperationType,
			&operation.AmountGrams, &operation.Status, &operation.OperatorUserId, &operation.ApproverUserId,
			&operation.TargetTeamId, &operation.TargetUserId, &operation.Notes,
			&operation.RejectionReason, &operation.ExpiresAt, &operation.ApprovedAt, &operation.CreatedAt, &operation.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return operation, fmt.Errorf("团队茶叶操作不存在")
		}
		return operation, fmt.Errorf("查询团队茶叶操作失败: %v", err)
	}
	return operation, nil
}

// IsTeamMember 检查用户是否是团队成员
func IsTeamMember(userId, teamId int) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM team_members 
		WHERE team_id = $1 AND user_id = $2 AND status = $3
	`, teamId, userId, TeMemberStatusActive).Scan(&count)

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

	rows, err := db.Query(query, userId, TeMemberStatusActive, TeamIdFreelancer)
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
