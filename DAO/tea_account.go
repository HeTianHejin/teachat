package dao

import (
	"database/sql"
	"fmt"
	"time"
)

/*
用户茶叶账户转账流程：
1、发起方法：用户填写对用户/团队、id、转出额度，无需审批，直接创建转帐表单；
2、锁定方法：转出的用户账户转出额度茶叶数量被锁定，防止重复发起转账；
3.1、接收方法：在有效期内，接收方用户或者团队任意1正常状态成员，操作接收，创建接收记录，继续第4步；
3.2、拒绝方法：在有效期内，接收方用户或者团队任意1状态正常成员，操作拒收，创建拒收原因及拒收用户id、时间记录，流程结束；
4、清算方法：接收方确认接收后，按锁定额度（接收额度）清算双方账户数额，创建实际流通流水明细记录。
5、超时处理：自动解锁转出用户账户被锁定额度茶叶，不创建交易流水明细记录。
*/

// 茶叶账户状态常量
const (
	TeaAccountStatus_Normal = "normal"
	TeaAccountStatus_Frozen = "frozen"
)

// 流通对象类型常量
const (
	TransactionTargetType_User = "u" // 用户
	TransactionTargetType_Team = "t" // 团队
)

// 转账状态常量（统一状态枚举）
const (
	StatusPendingApproval  = "pending_approval"  // 待审批（团队转出）
	StatusApproved         = "approved"          // 审批通过
	StatusApprovalRejected = "approval_rejected" // 审批拒绝
	StatusPendingReceipt   = "pending_receipt"   // 待接收
	StatusCompleted        = "completed"         // 已完成
	StatusRejected         = "rejected"          // 接收拒绝
	StatusExpired          = "expired"           // 已超时
)

// 交易类型常量
// TransactionType_TransferOut 表示转账支出交易类型
// TransactionType_TransferIn 表示转账收入交易类型
// TransactionType_SystemGrant 表示系统发放交易类型
// TransactionType_SystemDeduct 表示系统扣除交易类型
const (
	TransactionType_TransferOut  = "transfer_out"
	TransactionType_TransferIn   = "transfer_in"
	TransactionType_Withdraw     = "withdraw"
	TransactionType_SystemGrant  = "system_grant"
	TransactionType_SystemDeduct = "system_deduct"
)

// 用户茶叶账户结构体
type TeaUserAccount struct {
	Id                 int
	Uuid               string
	UserId             int
	BalanceGrams       float64 // 茶叶数量(克)
	LockedBalanceGrams float64 // 交易有效期被锁定的茶叶数量(克)
	Status             string  // normal, frozen
	FrozenReason       *string
	CreatedAt          time.Time
	UpdatedAt          *time.Time
}

// 用户专用茶叶转账结构体，
// payer转出，发起流程
// 转帐发起操作记录（从转出方视角）
// 注意不能转出0/负数，不能转给自己自由人团队id=TeamIdFreelancer
type TeaUserTransferOut struct {
	Id         int
	Uuid       string
	FromUserId int // 户主，转出方用户ID

	//用户填写，同时也是转出方操作记录内容
	// u/t必须二选一
	ToUserId    *int
	ToTeamId    *int
	AmountGrams float64 // 转账额度（克）
	Notes       string  // 转账备注

	// 待接收	StatusPendingReceipt   = "pending_receipt"
	// 已完成	StatusCompleted        = "completed"
	// 以拒收	StatusRejected         = "rejected"
	// 已超时	StatusExpired          = "expired"
	Status string // 系统填写
	//TransferType string 从表名获取

	//系统填写
	CreatedAt   time.Time  // 创建，流程开始时间，也是锁定额度起始时间
	ExpiresAt   time.Time  // 过期时间，也是锁定额度截止时间
	PaymentTime *time.Time // 实际支付时间（已批准+已接收）
	UpdatedAt   *time.Time
}

// payee转入，如果payer是团队发起，需内部批准后才有这个处理步骤
// 本人或者成员点击接收通知上“处理”才打开表单，提交参数后创建记录，同时处理对方转出记录
// 接收操作历史记录（从接收方视角），仅内部可见，用于查询、对账
// 接收流程相同，用户和团队共用
type TeaTransferIn struct {
	Id     int
	Uuid   string
	UserId int // 户主id，接收者用户ID

	// 系统填写，对接转出表单，必填二选一
	UserTransferOutId *int // 用户转出记录id
	TeamTransferOutId *int // 团队转出记录id

	// 已完成	StatusCompleted        = "completed"
	// 接收拒绝	StatusRejected         = "rejected"
	Status string
	//TransferType string 从表名获取

	// 接收方操作（接收方填写），C/R二选一
	// 如果TeamTransferOut.ToTeamId不为空，必须检查确认人是否为Team团队成员
	ConfirmedBy              *int
	ReceptionRejectionReason *string
	RejectedBy               *int

	CreatedAt time.Time
}

// 交易记录结构体
// 用户茶叶交易流水，实时查询出入表数据生成
type TransactionRecord struct {
	Id              int
	Uuid            string
	UserId          int
	TransferId      *string
	TransactionType string
	AmountGrams     float64
	BalanceBefore   float64
	BalanceAfter    float64
	Description     string
	TargetUserId    *int
	TargetTeamId    *int
	TargetType      string
	CreatedAt       time.Time
}

// GetTeaAccountByUserId 根据用户ID获取茶叶账户
func GetTeaAccountByUserId(userId int) (TeaUserAccount, error) {
	account := TeaUserAccount{}
	err := DB.QueryRow("SELECT id, uuid, user_id, balance_grams, locked_balance_grams, status, frozen_reason, created_at, updated_at FROM tea.user_accounts WHERE user_id = $1", userId).
		Scan(&account.Id, &account.Uuid, &account.UserId, &account.BalanceGrams, &account.LockedBalanceGrams, &account.Status, &account.FrozenReason, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return account, fmt.Errorf("用户茶叶账户不存在")
		}
		return account, fmt.Errorf("查询茶叶账户失败 - %v", err)
	}
	return account, nil
}

// Create 创建用户茶叶账户
func (account *TeaUserAccount) Create() error {
	statement := "INSERT INTO tea.user_accounts (user_id, balance_grams, status) VALUES ($1, $2, $3) RETURNING id, uuid"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(account.UserId, account.BalanceGrams, account.Status).Scan(&account.Id, &account.Uuid)
	if err != nil {
		return fmt.Errorf("创建茶叶账户失败 - %v", err)
	}
	return nil
}

// UpdateStatus 更新账户状态
func (account *TeaUserAccount) UpdateStatus(status, reason string) error {
	statement := "UPDATE tea.user_accounts SET status = $2, frozen_reason = $3, updated_at = $4 WHERE id = $1"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(account.Id, status, reason, time.Now())
	if err != nil {
		return fmt.Errorf("更新账户状态失败 - %v", err)
	}

	account.Status = status
	if reason != "" {
		account.FrozenReason = &reason
	} else {
		account.FrozenReason = nil
	}
	return nil
}

// SystemAdjustBalance 系统调整余额
func (account *TeaUserAccount) SystemAdjustBalance(amount float64, description string, adminUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("确认转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取当前余额
	var currentBalance float64
	err = tx.QueryRow("SELECT balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", account.UserId).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("获取当前余额失败: %v", err)
	}

	// 计算新余额
	newBalance := currentBalance + amount
	if newBalance < 0 {
		return fmt.Errorf("余额不能为负数")
	}

	// 更新账户余额
	_, err = tx.Exec("UPDATE tea.user_accounts SET balance_grams = $2, updated_at = $3 WHERE user_id = $1",
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

	_, err = tx.Exec(`INSERT INTO tea.transaction_records 
		(user_id, transaction_type, amount_grams, balance_before, balance_after, description, target_user_id, target_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		account.UserId, transactionType, amount, currentBalance, newBalance, description, &adminUserId, TransactionTargetType_User)
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

// CreateUtoUTeaTransfer 创建用户对用户类型转账记录
func CreateUtoUTeaTransfer(fromUserId, toUserId int, amount float64, notes string, expireHours int) (TeaUserTransferOut, error) {
	// 验证参数
	if amount <= 0 {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：转账金额必须大于0")
	}
	if fromUserId == toUserId {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：不能给自己转账")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出账户余额和锁定金额
	var fromBalance, fromLockedBalance float64
	var fromStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", fromUserId).
		Scan(&fromBalance, &fromLockedBalance, &fromStatus)
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：查询转出账户失败 - %v", err)
	}
	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := fromBalance - fromLockedBalance
	if availableBalance < amount {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", fromBalance, fromLockedBalance, availableBalance)
	}

	// 确保接收方账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea.user_accounts WHERE user_id = $1", toUserId).Scan(&toAccountId)
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：接收方账户不存在 - %v", err)
	}

	// 创建转账记录
	transfer := TeaUserTransferOut{
		FromUserId:  fromUserId,
		ToUserId:    &toUserId,
		ToTeamId:    nil,
		AmountGrams: amount,
		Status:      StatusPendingReceipt,
		Notes:       notes,
		ExpiresAt:   time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:   time.Now(),
	}

	err = tx.QueryRow(`INSERT INTO tea.user_transfer_out 
		(from_user_id, to_user_id, amount_grams, status, notes, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, uuid`,
		fromUserId, toUserId, amount, StatusPendingReceipt, notes, transfer.ExpiresAt).
		Scan(&transfer.Id, &transfer.Uuid)
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：创建转账记录失败 - %v", err)
	}

	// 锁定转出账户的相应金额
	_, err = tx.Exec("UPDATE tea.user_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), fromUserId)
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：锁定转出账户金额失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：提交事务失败 - %v", err)
	}

	return transfer, nil
}

// CreatePtoTTeaTransferToTeam 发起用户向团队的茶叶转账
func CreatePtoTTeaTransferToTeam(fromUserId, toTeamId int, amount float64, notes string, expireHours int) (TeaUserTransferOut, error) {
	// 验证参数
	if amount <= 0 {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：转账金额必须大于0")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出账户余额和锁定金额
	var fromBalance, fromLockedBalance float64
	var fromStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", fromUserId).
		Scan(&fromBalance, &fromLockedBalance, &fromStatus)
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：查询转出账户失败 - %v", err)
	}
	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := fromBalance - fromLockedBalance
	if availableBalance < amount {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", fromBalance, fromLockedBalance, availableBalance)
	}

	// 确保接收方团队账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea.team_accounts WHERE team_id = $1", toTeamId).Scan(&toAccountId)
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：接收方团队账户不存在 - %v", err)
	}

	// 创建转账记录
	transfer := TeaUserTransferOut{
		FromUserId:  fromUserId,
		ToUserId:    nil,
		ToTeamId:    &toTeamId,
		AmountGrams: amount,
		Status:      StatusPendingReceipt,
		Notes:       notes,
		ExpiresAt:   time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:   time.Now(),
	}

	err = tx.QueryRow(`INSERT INTO tea.user_transfer_out 
		(from_user_id, to_team_id, amount_grams, status, notes, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, uuid`,
		fromUserId, toTeamId, amount, StatusPendingReceipt, notes, transfer.ExpiresAt).
		Scan(&transfer.Id, &transfer.Uuid)
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：创建转账记录失败 - %v", err)
	}

	// 锁定转出账户的相应金额
	_, err = tx.Exec("UPDATE tea.user_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), fromUserId)
	if err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：锁定转出账户金额失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaUserTransferOut{}, fmt.Errorf("转账失败：提交事务失败 - %v", err)
	}

	return transfer, nil
}

// getAndValidateTransfer 获取并验证转账信息
func getAndValidateTransfer(tx *sql.Tx, transferUuid string, toUserId int) (TeaUserTransferOut, error) {
	var transfer TeaUserTransferOut
	err := tx.QueryRow(`SELECT id, uuid, from_user_id, to_user_id, to_team_id, amount_grams, status, expires_at
		FROM tea.user_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId, &transfer.ToTeamId,
			&transfer.AmountGrams, &transfer.Status, &transfer.ExpiresAt)
	if err != nil {
		return transfer, fmt.Errorf("确认转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != StatusPendingReceipt {
		return transfer, fmt.Errorf("确认转账失败：转账状态异常")
	}
	if time.Now().After(transfer.ExpiresAt) {
		// 转账已过期，更新状态
		_, _ = tx.Exec("UPDATE tea.user_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			StatusExpired, time.Now(), transfer.Id)
		return transfer, fmt.Errorf("确认转账失败：转账已过期")
	}

	// 检查确认权限
	if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
		// 团队转账：检查用户是否是团队成员
		isMember, err := IsTeamMember(toUserId, *transfer.ToTeamId)
		if err != nil {
			return transfer, fmt.Errorf("检查团队成员身份失败: %v", err)
		}
		if !isMember {
			return transfer, fmt.Errorf("只有团队成员才能确认团队转账")
		}
	} else {
		// 用户间转账：检查接收用户ID
		if transfer.ToUserId == nil || *transfer.ToUserId != toUserId {
			return transfer, fmt.Errorf("无权确认此转账")
		}
	}

	return transfer, nil
}

// confirmTeamTransfer 确认团队转账
func confirmTeamTransfer(tx *sql.Tx, transfer TeaUserTransferOut, toUserId int) error {
	// 确保团队账户存在
	var teamBalance float64
	err := tx.QueryRow("SELECT balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", *transfer.ToTeamId).Scan(&teamBalance)
	if err != nil {
		return fmt.Errorf("查询团队账户余额失败: %v", err)
	}

	// 获取转出账户信息
	var fromBalance, fromLockedBalance float64
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", transfer.FromUserId).
		Scan(&fromBalance, &fromLockedBalance)
	if err != nil {
		return fmt.Errorf("查询转出账户信息失败: %v", err)
	}

	// 检查余额是否足够
	if fromLockedBalance < transfer.AmountGrams {
		return fmt.Errorf("锁定余额不足，无法完成转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromLockedBalance, transfer.AmountGrams)
	}
	if fromBalance < transfer.AmountGrams {
		return fmt.Errorf("账户余额不足，无法完成转账。当前余额: %.3f克, 转账金额: %.3f克", fromBalance, transfer.AmountGrams)
	}

	// 更新账户余额
	newFromBalance := fromBalance - transfer.AmountGrams
	newFromLockedBalance := fromLockedBalance - transfer.AmountGrams

	_, err = tx.Exec("UPDATE tea.user_accounts SET balance_grams = $1, locked_balance_grams = $2, updated_at = $3 WHERE user_id = $4",
		newFromBalance, newFromLockedBalance, time.Now(), transfer.FromUserId)
	if err != nil {
		return fmt.Errorf("更新转出账户余额失败: %v", err)
	}

	_, err = tx.Exec("UPDATE tea.team_accounts SET balance_grams = $1, updated_at = $2 WHERE team_id = $3",
		teamBalance+transfer.AmountGrams, time.Now(), *transfer.ToTeamId)
	if err != nil {
		return fmt.Errorf("更新团队账户余额失败: %v", err)
	}

	// 更新转账状态
	paymentTime := time.Now()
	_, err = tx.Exec(`UPDATE tea.user_transfer_out SET 
		status = $1, 
		payment_time = $2, 
		updated_at = $3 
		WHERE id = $4`,
		StatusCompleted, paymentTime, paymentTime, transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建转入记录
	_, err = tx.Exec(`INSERT INTO tea.transfer_in 
		(user_id, user_transfer_out_id, status, confirmed_by, created_at) 
		VALUES ($1, $2, $3, $4, $5)`,
		toUserId, transfer.Id, StatusCompleted, toUserId, paymentTime)
	if err != nil {
		return fmt.Errorf("创建转入记录失败: %v", err)
	}

	// 记录交易流水
	err = recordTeamTransferTransactions(tx, transfer, fromBalance, newFromBalance, teamBalance)
	if err != nil {
		return err
	}

	return nil
}

// confirmPersonalTransfer 确认用户间转账
func confirmPersonalTransfer(tx *sql.Tx, transfer TeaUserTransferOut, toUserId int) error {
	// 获取接收方账户余额
	var toBalance float64
	err := tx.QueryRow("SELECT balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", toUserId).Scan(&toBalance)
	if err != nil {
		return fmt.Errorf("查询接收账户余额失败: %v", err)
	}

	// 获取转出账户信息
	var fromBalance, fromLockedBalance float64
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", transfer.FromUserId).
		Scan(&fromBalance, &fromLockedBalance)
	if err != nil {
		return fmt.Errorf("查询转出账户信息失败: %v", err)
	}

	// 检查余额是否足够
	if fromLockedBalance < transfer.AmountGrams {
		return fmt.Errorf("锁定余额不足，无法完成转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromLockedBalance, transfer.AmountGrams)
	}
	if fromBalance < transfer.AmountGrams {
		return fmt.Errorf("账户余额不足，无法完成转账。当前余额: %.3f克, 转账金额: %.3f克", fromBalance, transfer.AmountGrams)
	}

	// 更新账户余额
	newFromBalance := fromBalance - transfer.AmountGrams
	newFromLockedBalance := fromLockedBalance - transfer.AmountGrams

	_, err = tx.Exec("UPDATE tea.user_accounts SET balance_grams = $1, locked_balance_grams = $2, updated_at = $3 WHERE user_id = $4",
		newFromBalance, newFromLockedBalance, time.Now(), transfer.FromUserId)
	if err != nil {
		return fmt.Errorf("更新转出账户余额失败: %v", err)
	}

	_, err = tx.Exec("UPDATE tea.user_accounts SET balance_grams = $1, updated_at = $2 WHERE user_id = $3",
		toBalance+transfer.AmountGrams, time.Now(), toUserId)
	if err != nil {
		return fmt.Errorf("更新接收账户余额失败: %v", err)
	}

	// 更新转账状态
	paymentTime := time.Now()
	_, err = tx.Exec(`UPDATE tea.user_transfer_out SET 
		status = $1, 
		payment_time = $2, 
		updated_at = $3 
		WHERE id = $4`,
		StatusCompleted, paymentTime, paymentTime, transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建转入记录
	_, err = tx.Exec(`INSERT INTO tea.transfer_in 
		(user_id, user_transfer_out_id, status, confirmed_by, created_at) 
		VALUES ($1, $2, $3, $4, $5)`,
		toUserId, transfer.Id, StatusCompleted, toUserId, paymentTime)
	if err != nil {
		return fmt.Errorf("创建转入记录失败: %v", err)
	}

	// 记录交易流水
	err = recordPersonalTransferTransactions(tx, transfer, fromBalance, newFromBalance, toBalance, toUserId)
	if err != nil {
		return err
	}

	return nil
}

// recordTeamTransferTransactions 记录团队转账的交易流水
func recordTeamTransferTransactions(tx *sql.Tx, transfer TeaUserTransferOut, fromBalance, newFromBalance, teamBalance float64) error {
	// 记录转出方交易流水
	_, err := tx.Exec(`INSERT INTO tea.transaction_records 
		(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, target_team_id, target_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		transfer.FromUserId, &transfer.Uuid, TransactionType_TransferOut, transfer.AmountGrams,
		fromBalance, newFromBalance, fmt.Sprintf("向团队转账: %s", transfer.Notes), transfer.ToTeamId, TransactionTargetType_Team)
	if err != nil {
		return fmt.Errorf("记录转出交易流水失败: %v", err)
	}

	// 记录团队转入交易流水
	_, err = tx.Exec(`INSERT INTO tea.transaction_records 
		(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, target_user_id, target_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		*transfer.ToTeamId, &transfer.Uuid, TransactionType_TransferIn, transfer.AmountGrams,
		teamBalance, teamBalance+transfer.AmountGrams, "用户转账转入", &transfer.FromUserId, TransactionTargetType_User)
	if err != nil {
		return fmt.Errorf("记录团队转入交易流水失败: %v", err)
	}

	return nil
}

// recordPersonalTransferTransactions 记录用户转账的交易流水
func recordPersonalTransferTransactions(tx *sql.Tx, transfer TeaUserTransferOut, fromBalance, newFromBalance, toBalance float64, toUserId int) error {
	// 记录转出方交易流水
	_, err := tx.Exec(`INSERT INTO tea.transaction_records 
		(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, target_user_id, target_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		transfer.FromUserId, &transfer.Uuid, TransactionType_TransferOut, transfer.AmountGrams,
		fromBalance, newFromBalance, fmt.Sprintf("转账给用户: %s", transfer.Notes), toUserId, TransactionTargetType_User)
	if err != nil {
		return fmt.Errorf("记录转出交易流水失败: %v", err)
	}

	// 记录接收方交易流水
	_, err = tx.Exec(`INSERT INTO tea.transaction_records 
		(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, target_user_id, target_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		toUserId, &transfer.Uuid, TransactionType_TransferIn, transfer.AmountGrams,
		toBalance, toBalance+transfer.AmountGrams, "转账转入", &transfer.FromUserId, TransactionTargetType_User)
	if err != nil {
		return fmt.Errorf("记录转入交易流水失败: %v", err)
	}

	return nil
}

// ConfirmTeaTransfer 确认接收转账（支持用户间转账和用户向团队转账）
func ConfirmTeaTransfer(transferUuid string, toUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("确认转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取并验证转账信息
	transfer, err := getAndValidateTransfer(tx, transferUuid, toUserId)
	if err != nil {
		return err
	}

	// 根据转账类型执行不同的确认逻辑
	if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
		err = confirmTeamTransfer(tx, transfer, toUserId)
	} else {
		err = confirmPersonalTransfer(tx, transfer, toUserId)
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

// RejectTeaTransfer 拒绝转账（支持用户间转账和用户向团队转账）
func RejectTeaTransfer(transferUuid string, toUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("确认转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息（包含团队信息和金额）
	var transfer TeaUserTransferOut
	err = tx.QueryRow("SELECT id, from_user_id, to_user_id, to_team_id, amount_grams, status FROM tea.user_transfer_out WHERE uuid = $1 FOR UPDATE", transferUuid).
		Scan(&transfer.Id, &transfer.FromUserId, &transfer.ToUserId, &transfer.ToTeamId, &transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("确认转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != StatusPendingReceipt {
		return fmt.Errorf("确认转账失败：转账状态异常")
	}

	// 检查拒绝权限
	if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
		// 团队转账：检查用户是否是团队成员
		isMember, err := IsTeamMember(toUserId, *transfer.ToTeamId)
		if err != nil {
			return fmt.Errorf("检查团队成员身份失败: %v", err)
		}
		if !isMember {
			return fmt.Errorf("只有团队成员才能拒绝团队转账")
		}
	} else {
		// 用户间转账：检查接收用户ID
		if transfer.ToUserId == nil || *transfer.ToUserId != toUserId {
			return fmt.Errorf("无权拒绝此转账")
		}
	}

	// 获取转出账户的锁定金额信息
	var fromLockedBalance float64
	err = tx.QueryRow("SELECT locked_balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", transfer.FromUserId).Scan(&fromLockedBalance)
	if err != nil {
		return fmt.Errorf("查询转出账户锁定金额失败: %v", err)
	}

	// 解锁转出账户的锁定金额
	newFromLockedBalance := fromLockedBalance - transfer.AmountGrams

	// 检查锁定余额是否足够解锁
	if newFromLockedBalance < 0 {
		return fmt.Errorf("锁定余额不足，无法拒绝转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromLockedBalance, transfer.AmountGrams)
	}

	_, err = tx.Exec("UPDATE tea.user_accounts SET locked_balance_grams = $1, updated_at = $2 WHERE user_id = $3",
		newFromLockedBalance, time.Now(), transfer.FromUserId)
	if err != nil {
		return fmt.Errorf("解锁转出账户金额失败: %v", err)
	}

	// 更新转账状态
	_, err = tx.Exec(`UPDATE tea.user_transfer_out SET 
		status = $1, 
		updated_at = $2 
		WHERE id = $3`,
		StatusRejected, time.Now(), transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建拒收记录
	_, err = tx.Exec(`INSERT INTO tea.transfer_in 
		(user_id, user_transfer_out_id, status, rejected_by, reception_rejection_reason, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		toUserId, transfer.Id, StatusRejected, toUserId, reason, time.Now())
	if err != nil {
		return fmt.Errorf("创建拒收记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// ProcessExpiredTransfers 处理过期的转账，解锁相应的锁定金额
func ProcessExpiredTransfers() error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("确认转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 查找所有过期且仍为pending状态的转账
	rows, err := tx.Query(`
		SELECT id, from_user_id, amount_grams 
		FROM tea.user_transfer_out 
		WHERE status = $1 AND expires_at < $2`,
		StatusPendingReceipt, time.Now())
	if err != nil {
		return fmt.Errorf("查询过期转账失败: %v", err)
	}
	defer rows.Close()

	var expiredTransfers []struct {
		Id         int
		FromUserId int
		Amount     float64
	}

	for rows.Next() {
		var et struct {
			Id         int
			FromUserId int
			Amount     float64
		}
		if err := rows.Scan(&et.Id, &et.FromUserId, &et.Amount); err != nil {
			return fmt.Errorf("扫描过期转账失败: %v", err)
		}
		expiredTransfers = append(expiredTransfers, et)
	}

	if len(expiredTransfers) == 0 {
		return nil // 没有过期转账需要处理
	}

	// 处理每个过期转账：更新状态并解锁金额
	for _, et := range expiredTransfers {
		// 获取当前锁定余额
		var currentLockedBalance float64
		err = tx.QueryRow("SELECT locked_balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", et.FromUserId).Scan(&currentLockedBalance)
		if err != nil {
			return fmt.Errorf("查询锁定余额失败: %v", err)
		}

		// 检查锁定余额是否足够
		if currentLockedBalance < et.Amount {
			// 锁定余额不足，记录警告并跳过
			continue
		}

		// 更新转账状态为过期
		_, err = tx.Exec("UPDATE tea.user_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			StatusExpired, time.Now(), et.Id)
		if err != nil {
			return fmt.Errorf("更新过期转账状态失败: %v", err)
		}

		// 解锁相应的锁定金额
		newLockedBalance := currentLockedBalance - et.Amount
		_, err = tx.Exec("UPDATE tea.user_accounts SET locked_balance_grams = $1, updated_at = $2 WHERE user_id = $3",
			newLockedBalance, time.Now(), et.FromUserId)
		if err != nil {
			return fmt.Errorf("解锁过期转账金额失败: %v", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetPendingTransfers 获取用户待确认转账列表（支持用户间转账和团队转账）
func GetPendingTransfers(userId int, page, limit int) ([]TeaUserTransferOut, error) {
	offset := (page - 1) * limit
	// 查询用户用户待确认转账 + 用户所属团队的待确认转账
	rows, err := DB.Query(`SELECT DISTINCT t.id, t.uuid, t.from_user_id, t.to_user_id, t.to_team_id, 
		t.amount_grams, t.status, t.notes, t.expires_at, t.created_at, t.updated_at 
		FROM tea.user_transfer_out t 
		LEFT JOIN team_members tm ON t.to_team_id = tm.team_id AND tm.user_id = $1
		WHERE (t.to_user_id = $1 OR (tm.user_id = $1 AND t.to_team_id IS NOT NULL)) 
		AND t.status = $2 AND t.expires_at > NOW() 
		ORDER BY t.created_at DESC LIMIT $3 OFFSET $4`,
		userId, StatusPendingReceipt, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待确认转账失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaUserTransferOut
	for rows.Next() {
		var transfer TeaUserTransferOut
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId,
			&transfer.ToTeamId, &transfer.AmountGrams, &transfer.Status, &transfer.Notes,
			&transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetUserTransactions 获取用户交易流水
func GetUserTransactions(userId int, page, limit int, transactionType string) ([]TransactionRecord, error) {
	offset := (page - 1) * limit
	var rows *sql.Rows
	var err error

	if transactionType == "" {
		rows, err = DB.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, target_user_id, target_team_id, target_type, created_at 
			FROM tea.transaction_records WHERE user_id = $1 
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userId, limit, offset)
	} else {
		rows, err = DB.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, target_user_id, target_team_id, target_type, created_at 
			FROM tea.transaction_records WHERE user_id = $1 AND transaction_type = $2 
			ORDER BY created_at DESC LIMIT $3 OFFSET $4`, userId, transactionType, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("查询交易流水失败: %v", err)
	}
	defer rows.Close()

	var transactions []TransactionRecord
	for rows.Next() {
		var transaction TransactionRecord
		err = rows.Scan(&transaction.Id, &transaction.Uuid, &transaction.UserId, &transaction.TransferId,
			&transaction.TransactionType, &transaction.AmountGrams, &transaction.BalanceBefore,
			&transaction.BalanceAfter, &transaction.Description, &transaction.TargetUserId,
			&transaction.TargetTeamId, &transaction.TargetType, &transaction.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描交易流水失败: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetTransferHistory 获取用户转账历史
func GetTransferHistory(userId int, page, limit int) ([]TeaUserTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`SELECT id, uuid, from_user_id, to_user_id, to_team_id, amount_grams, status, 
		payment_time, notes, expires_at, created_at, updated_at 
		FROM tea.user_transfer_out WHERE from_user_id = $1 OR to_user_id = $1 
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询转账历史失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaUserTransferOut
	for rows.Next() {
		var transfer TeaUserTransferOut
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId,
			&transfer.ToTeamId, &transfer.AmountGrams, &transfer.Status, &transfer.PaymentTime, &transfer.Notes,
			&transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetPendingTransfersForTeam 获取转入指定团队的待确认转账列表
func GetPendingTransfersForTeam(teamId int, page, limit int) ([]TeaUserTransferOut, error) {
	offset := (page - 1) * limit
	// 查询转入团队的待确认转账
	rows, err := DB.Query(`SELECT t.id, t.uuid, t.from_user_id, t.to_user_id, t.to_team_id, 
		t.amount_grams, t.status, t.payment_time, t.notes, t.expires_at, t.created_at, t.updated_at 
		FROM tea.user_transfer_out t 
		WHERE t.to_team_id = $1 
		AND t.status = $2 AND t.expires_at > NOW() 
		ORDER BY t.created_at DESC LIMIT $3 OFFSET $4`,
		teamId, StatusPendingReceipt, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询团队待确认转账失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaUserTransferOut
	for rows.Next() {
		var transfer TeaUserTransferOut
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId,
			&transfer.ToTeamId, &transfer.AmountGrams, &transfer.Status, &transfer.PaymentTime, &transfer.Notes,
			&transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
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
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM tea.user_accounts WHERE user_id = $1)", userId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查账户存在性失败: %v", err)
	}

	if !exists {
		account := &TeaUserAccount{
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
	err := DB.QueryRow("SELECT status, frozen_reason FROM tea.user_accounts WHERE user_id = $1", userId).
		Scan(&status, &frozenReason)
	if err != nil {
		return false, "", fmt.Errorf("查询账户状态失败: %v", err)
	}

	if status == TeaAccountStatus_Frozen {
		return true, frozenReason.String, nil
	}

	return false, "", nil
}

// GetTransferOuts 获取用户转出记录（从转出方视角）
func GetTransferOuts(userId int, page, limit int) ([]TeaUserTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, to_user_id, to_team_id, 
			   amount_grams, status, payment_time, notes,
			   expires_at, created_at, updated_at 
		FROM tea.user_transfer_out 
		WHERE from_user_id = $1 
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询转出记录失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaUserTransferOut
	for rows.Next() {
		var transfer TeaUserTransferOut
		err = rows.Scan(
			&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId, &transfer.ToTeamId,
			&transfer.AmountGrams, &transfer.Status, &transfer.PaymentTime, &transfer.Notes,
			&transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转出记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetTransferIns 获取用户转入记录（从接收方视角）
func GetTransferIns(userId int, page, limit int) ([]TeaTransferIn, error) {
	offset := (page - 1) * limit
	// 查询用户转入记录（从transfer_in表）
	rows, err := DB.Query(`
		SELECT ti.id, ti.uuid, ti.user_id, ti.user_transfer_out_id, ti.team_transfer_out_id,
			   ti.status, ti.confirmed_by, ti.rejected_by, ti.reception_rejection_reason, ti.created_at
		FROM tea.transfer_in ti
		WHERE ti.user_id = $1
		ORDER BY ti.created_at DESC LIMIT $2 OFFSET $3`,
		userId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询转入记录失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaTransferIn
	for rows.Next() {
		var transfer TeaTransferIn
		err = rows.Scan(
			&transfer.Id, &transfer.Uuid, &transfer.UserId, &transfer.UserTransferOutId, &transfer.TeamTransferOutId,
			&transfer.Status, &transfer.ConfirmedBy, &transfer.RejectedBy, &transfer.ReceptionRejectionReason, &transfer.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转入记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}


