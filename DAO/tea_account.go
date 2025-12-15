package dao

import (
	"database/sql"
	"fmt"
	"time"
)

// 茶叶账户状态常量
const (
	TeaAccountStatus_Normal = "normal"
	TeaAccountStatus_Frozen = "frozen"
)

// 转账状态常量
const (
	TransferStatus_Pending   = "pending"
	TransferStatus_Confirmed = "confirmed"
	TransferStatus_Rejected  = "rejected"
	TransferStatus_Expired   = "expired"
)

// 交易类型常量
// TransactionType_TransferOut 表示转账支出交易类型
// TransactionType_TransferIn 表示转账收入交易类型
// TransactionType_SystemGrant 表示系统茶庄发放交易类型
// TransactionType_SystemDeduct 表示系统扣除交易类型
// TransactionType_Refund 表示退款交易类型
const (
	TransactionType_TransferOut  = "transfer_out"
	TransactionType_TransferIn   = "transfer_in"
	TransactionType_SystemGrant  = "system_grant"
	TransactionType_SystemDeduct = "system_deduct"
	TransactionType_Refund       = "refund"
)

// 茶叶账户结构体
type TeaAccount struct {
	Id           int
	Uuid         string
	UserId       int
	BalanceGrams float64 // 茶叶数量(克)
	Status       string  // normal, frozen
	FrozenReason string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// 茶叶转账结构体
type TeaTransfer struct {
	Id              int
	Uuid            string
	FromUserId      int
	ToUserId        int
	AmountGrams     float64
	Status          string // pending, confirmed, rejected, expired
	PaymentTime     *time.Time
	Notes           string
	RejectionReason *string
	ExpiresAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

// 茶叶交易流水结构体
type TeaTransaction struct {
	Id              int
	Uuid            string
	UserId          int
	TransferId      *string
	TransactionType string // transfer_out, transfer_in, system_grant, system_deduct, refund
	AmountGrams     float64
	BalanceBefore   float64
	BalanceAfter    float64
	Description     string
	RelatedUserId   *int
	CreatedAt       time.Time
}

// GetTeaAccountByUserId 根据用户ID获取茶叶账户
func GetTeaAccountByUserId(userId int) (TeaAccount, error) {
	account := TeaAccount{}
	err := db.QueryRow("SELECT id, uuid, user_id, balance_grams, status, frozen_reason, created_at, updated_at FROM tea_accounts WHERE user_id = $1", userId).
		Scan(&account.Id, &account.Uuid, &account.UserId, &account.BalanceGrams, &account.Status, &account.FrozenReason, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return account, fmt.Errorf("用户茶叶账户不存在")
		}
		return account, fmt.Errorf("查询茶叶账户失败: %v", err)
	}
	return account, nil
}

// Create 创建茶叶账户
func (account *TeaAccount) Create() error {
	statement := "INSERT INTO tea_accounts (user_id, balance_grams, status) VALUES ($1, $2, $3) RETURNING id, uuid"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(account.UserId, account.BalanceGrams, account.Status).Scan(&account.Id, &account.Uuid)
	if err != nil {
		return fmt.Errorf("创建茶叶账户失败: %v", err)
	}
	return nil
}

// UpdateStatus 更新账户状态
func (account *TeaAccount) UpdateStatus(status, reason string) error {
	statement := "UPDATE tea_accounts SET status = $2, frozen_reason = $3, updated_at = $4 WHERE id = $1"
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

// SystemAdjustBalance 系统调整余额
func (account *TeaAccount) SystemAdjustBalance(amount float64, description string, adminUserId int) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取当前余额
	var currentBalance float64
	err = tx.QueryRow("SELECT balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", account.UserId).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("获取当前余额失败: %v", err)
	}

	// 计算新余额
	newBalance := currentBalance + amount
	if newBalance < 0 {
		return fmt.Errorf("余额不能为负数")
	}

	// 更新账户余额
	_, err = tx.Exec("UPDATE tea_accounts SET balance_grams = $2, updated_at = $3 WHERE user_id = $1",
		account.UserId, newBalance, time.Now())
	if err != nil {
		return fmt.Errorf("更新账户余额失败: %v", err)
	}

	// 记录交易流水
	transactionType := TransactionType_SystemGrant
	if amount < 0 {
		transactionType = TransactionType_SystemDeduct
		amount = -amount // 取绝对值
	}

	_, err = tx.Exec(`INSERT INTO tea_transactions 
		(user_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		account.UserId, transactionType, amount, currentBalance, newBalance, description, &adminUserId)
	if err != nil {
		return fmt.Errorf("记录交易流水失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	account.BalanceGrams = newBalance
	return nil
}

// CreateTeaTransfer 创建转账记录
func CreateTeaTransfer(fromUserId, toUserId int, amount float64, notes string, expireHours int) (TeaTransfer, error) {
	// 验证参数
	if amount <= 0 {
		return TeaTransfer{}, fmt.Errorf("转账金额必须大于0")
	}
	if fromUserId == toUserId {
		return TeaTransfer{}, fmt.Errorf("不能给自己转账")
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出账户余额
	var fromBalance float64
	err = tx.QueryRow("SELECT balance_grams, status FROM tea_accounts WHERE user_id = $1 FOR UPDATE", fromUserId).
		Scan(&fromBalance, new(string))
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("查询转出账户失败: %v", err)
	}
	if fromBalance < amount {
		return TeaTransfer{}, fmt.Errorf("余额不足")
	}

	// 确保接收方账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea_accounts WHERE user_id = $1", toUserId).Scan(&toAccountId)
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("接收方账户不存在: %v", err)
	}

	// 创建转账记录
	transfer := TeaTransfer{
		FromUserId:  fromUserId,
		ToUserId:    toUserId,
		AmountGrams: amount,
		Status:      TransferStatus_Pending,
		Notes:       notes,
		ExpiresAt:   time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:   time.Now(),
	}

	err = tx.QueryRow(`INSERT INTO tea_transfers 
		(from_user_id, to_user_id, amount_grams, status, notes, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, uuid`,
		fromUserId, toUserId, amount, TransferStatus_Pending, notes, transfer.ExpiresAt).
		Scan(&transfer.Id, &transfer.Uuid)
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("创建转账记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaTransfer{}, fmt.Errorf("提交事务失败: %v", err)
	}

	return transfer, nil
}

// ConfirmTeaTransfer 确认接收转账
func ConfirmTeaTransfer(transferUuid string, toUserId int) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaTransfer
	var fromUserId int
	err = tx.QueryRow(`SELECT id, uuid, from_user_id, to_user_id, amount_grams, status, expires_at 
		FROM tea_transfers WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &fromUserId, &transfer.ToUserId,
			&transfer.AmountGrams, &transfer.Status, &transfer.ExpiresAt)
	if err != nil {
		return fmt.Errorf("转账记录不存在: %v", err)
	}

	// 验证状态
	if transfer.Status != TransferStatus_Pending {
		return fmt.Errorf("转账状态异常")
	}
	if transfer.ToUserId != toUserId {
		return fmt.Errorf("无权确认此转账")
	}
	if time.Now().After(transfer.ExpiresAt) {
		// 转账已过期，更新状态
		_, _ = tx.Exec("UPDATE tea_transfers SET status = $1, updated_at = $2 WHERE id = $3",
			TransferStatus_Expired, time.Now(), transfer.Id)
		return fmt.Errorf("转账已过期")
	}

	// 获取转出账户当前余额并冻结资金
	var fromBalance float64
	err = tx.QueryRow("SELECT balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", fromUserId).Scan(&fromBalance)
	if err != nil {
		return fmt.Errorf("查询转出账户余额失败: %v", err)
	}
	if fromBalance < transfer.AmountGrams {
		return fmt.Errorf("转出账户余额不足")
	}

	// 获取接收账户当前余额
	var toBalance float64
	err = tx.QueryRow("SELECT balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", toUserId).Scan(&toBalance)
	if err != nil {
		return fmt.Errorf("查询接收账户余额失败: %v", err)
	}

	// 执行转账：扣除转出账户余额，增加接收账户余额
	_, err = tx.Exec("UPDATE tea_accounts SET balance_grams = $2, updated_at = $3 WHERE user_id = $1",
		fromUserId, fromBalance-transfer.AmountGrams, time.Now())
	if err != nil {
		return fmt.Errorf("更新转出账户余额失败: %v", err)
	}

	_, err = tx.Exec("UPDATE tea_accounts SET balance_grams = $2, updated_at = $3 WHERE user_id = $1",
		toUserId, toBalance+transfer.AmountGrams, time.Now())
	if err != nil {
		return fmt.Errorf("更新接收账户余额失败: %v", err)
	}

	// 更新转账状态
	paymentTime := time.Now()
	_, err = tx.Exec("UPDATE tea_transfers SET status = $1, payment_time = $2, updated_at = $3 WHERE id = $4",
		TransferStatus_Confirmed, paymentTime, paymentTime, transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 记录交易流水：转出方
	_, err = tx.Exec(`INSERT INTO tea_transactions 
		(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		fromUserId, &transfer.Uuid, TransactionType_TransferOut, transfer.AmountGrams,
		fromBalance, fromBalance-transfer.AmountGrams, "转账转出", &toUserId)
	if err != nil {
		return fmt.Errorf("记录转出交易流水失败: %v", err)
	}

	// 记录交易流水：接收方
	_, err = tx.Exec(`INSERT INTO tea_transactions 
		(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		toUserId, &transfer.Uuid, TransactionType_TransferIn, transfer.AmountGrams,
		toBalance, toBalance+transfer.AmountGrams, "转账转入", &fromUserId)
	if err != nil {
		return fmt.Errorf("记录接收交易流水失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// RejectTeaTransfer 拒绝转账
func RejectTeaTransfer(transferUuid string, toUserId int, reason string) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaTransfer
	err = tx.QueryRow("SELECT id, to_user_id, status FROM tea_transfers WHERE uuid = $1 FOR UPDATE", transferUuid).
		Scan(&transfer.Id, &transfer.ToUserId, &transfer.Status)
	if err != nil {
		return fmt.Errorf("转账记录不存在: %v", err)
	}

	// 验证状态
	if transfer.Status != TransferStatus_Pending {
		return fmt.Errorf("转账状态异常")
	}
	if transfer.ToUserId != toUserId {
		return fmt.Errorf("无权拒绝此转账")
	}

	// 更新转账状态
	_, err = tx.Exec("UPDATE tea_transfers SET status = $1, rejection_reason = $2, updated_at = $3 WHERE id = $4",
		TransferStatus_Rejected, reason, time.Now(), transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetPendingTransfers 获取用户待确认转账列表
func GetPendingTransfers(userId int, page, limit int) ([]TeaTransfer, error) {
	offset := (page - 1) * limit
	rows, err := db.Query(`SELECT id, uuid, from_user_id, to_user_id, amount_grams, status, 
		payment_time, notes, rejection_reason, expires_at, created_at, updated_at 
		FROM tea_transfers WHERE to_user_id = $1 AND status = $2 AND expires_at > NOW() 
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		userId, TransferStatus_Pending, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待确认转账失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaTransfer
	for rows.Next() {
		var transfer TeaTransfer
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId,
			&transfer.AmountGrams, &transfer.Status, &transfer.PaymentTime, &transfer.Notes,
			&transfer.RejectionReason, &transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetUserTransactions 获取用户交易流水
func GetUserTransactions(userId int, page, limit int, transactionType string) ([]TeaTransaction, error) {
	offset := (page - 1) * limit
	var rows *sql.Rows
	var err error

	if transactionType == "" {
		rows, err = db.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, related_user_id, created_at 
			FROM tea_transactions WHERE user_id = $1 
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userId, limit, offset)
	} else {
		rows, err = db.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, related_user_id, created_at 
			FROM tea_transactions WHERE user_id = $1 AND transaction_type = $2 
			ORDER BY created_at DESC LIMIT $3 OFFSET $4`, userId, transactionType, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("查询交易流水失败: %v", err)
	}
	defer rows.Close()

	var transactions []TeaTransaction
	for rows.Next() {
		var transaction TeaTransaction
		err = rows.Scan(&transaction.Id, &transaction.Uuid, &transaction.UserId, &transaction.TransferId,
			&transaction.TransactionType, &transaction.AmountGrams, &transaction.BalanceBefore,
			&transaction.BalanceAfter, &transaction.Description, &transaction.RelatedUserId, &transaction.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描交易流水失败: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetTransferHistory 获取用户转账历史
func GetTransferHistory(userId int, page, limit int) ([]TeaTransfer, error) {
	offset := (page - 1) * limit
	rows, err := db.Query(`SELECT id, uuid, from_user_id, to_user_id, amount_grams, status, 
		payment_time, notes, rejection_reason, expires_at, created_at, updated_at 
		FROM tea_transfers WHERE from_user_id = $1 OR to_user_id = $1 
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询转账历史失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaTransfer
	for rows.Next() {
		var transfer TeaTransfer
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId,
			&transfer.AmountGrams, &transfer.Status, &transfer.PaymentTime, &transfer.Notes,
			&transfer.RejectionReason, &transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// EnsureTeaAccountExists 确保用户有茶叶账户
func EnsureTeaAccountExists(userId int) error {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM tea_accounts WHERE user_id = $1)", userId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查账户存在性失败: %v", err)
	}

	if !exists {
		account := &TeaAccount{
			UserId:       userId,
			BalanceGrams: 0.0,
			Status:       TeaAccountStatus_Normal,
		}
		return account.Create()
	}

	return nil
}

// CheckAccountFrozen 检查账户是否被冻结
func CheckAccountFrozen(userId int) (bool, string, error) {
	var status string
	var frozenReason sql.NullString
	err := db.QueryRow("SELECT status, frozen_reason FROM tea_accounts WHERE user_id = $1", userId).
		Scan(&status, &frozenReason)
	if err != nil {
		return false, "", fmt.Errorf("查询账户状态失败: %v", err)
	}

	if status == TeaAccountStatus_Frozen {
		return true, frozenReason.String, nil
	}

	return false, "", nil
}
