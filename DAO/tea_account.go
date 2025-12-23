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

// 茶叶账户持有人类型常量
const (
	TeaAccountHolderType_User = "u" // 用户
	TeaAccountHolderType_Team = "t" // 团队
)

// 转账状态常量（统一状态枚举）
const (
	// 团队转出特有状态
	TeaTransferStatusPendingApproval  = "pending_approval"  // 待团队审批
	TeaTransferStatusApproved         = "approved"          // 团队审批通过
	TeaTransferStatusApprovalRejected = "approval_rejected" // 团队审批拒绝

	// 通用状态
	TeaTransferStatusPendingReceipt = "pending_receipt" // 待接收方确认
	TeaTransferStatusCompleted      = "completed"       // 转账完成
	TeaTransferStatusRejected       = "rejected"        // 接收方拒绝
	TeaTransferStatusExpired        = "expired"         // 已超时
)

// 转账类型常量
const (
	TeaTransferType_UserToUser = "user_to_user"
	TeaTransferType_UserToTeam = "user_to_team"
	TeaTransferType_TeamToUser = "team_to_user"
	TeaTransferType_TeamToTeam = "team_to_team"
)

// 交易类型常量
// TeaTransactionType_TransferOut 表示转账支出交易类型
// TeaTransactionType_TransferIn 表示转账收入交易类型
// TeaTransactionType_SystemGrant 表示系统发放交易类型
// TeaTransactionType_SystemDeduct 表示系统扣除交易类型
const (
	TeaTransactionType_TransferOut  = "transfer_out"
	TeaTransactionType_TransferIn   = "transfer_in"
	TeaTransactionType_Withdraw     = "withdraw"
	TeaTransactionType_SystemGrant  = "system_grant"
	TeaTransactionType_SystemDeduct = "system_deduct"
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

// 用户对用户，专用茶叶转账结构体，
// 注意：不能转出0或者负数，不能转给自己，不能转给被冻结User茶叶账户
type TeaUserToUserTransferOut struct {
	Id           int
	Uuid         string
	FromUserId   int    // 转出方用户ID，对账单审计用
	FromUserName string // 转出方用户名称，对账单审计用
	ToUserId     int    // 接收方用户ID，对账单审计用
	ToUserName   string // 接收方用户名称，对账单审计用

	//转出用户填写
	AmountGrams float64 // 转账额度（克）
	Notes       string  // 转账备注

	// 待接收	StatusPendingReceipt   = "pending_receipt"
	// 已完成	StatusCompleted        = "completed"
	// 以拒收	StatusRejected         = "rejected"
	// 已超时	StatusExpired          = "expired"
	Status string // 系统填写
	//TransferType string 从表名获取

	//系统填写
	BalanceAfterTransfer float64    // 转出后账户余额（克），对账单审计用
	CreatedAt            time.Time  // 创建，流程开始时间，也是锁定额度起始时间
	ExpiresAt            time.Time  // 过期时间，也是锁定额度截止时间
	PaymentTime          *time.Time // （清算成功才有值）实际支付时间，关联对方确认接收时间
	UpdatedAt            *time.Time
}

// 用户对团队，专用茶叶转账结构体
// 注意不能转出0/负数，不能转给自己、自由人团队id=TeamIdFreelancer(2)，不能转给被冻结Team茶叶账户
type TeaUserToTeamTransferOut struct {
	Id           int
	Uuid         string
	FromUserId   int    // 转出方用户ID，对账单审计用
	FromUserName string // 转出方用户名称，对账单审计用
	ToTeamId     int    // 接收方团队ID，对账单审计用
	ToTeamName   string // 接收方团队名称，对账单审计用

	//转出用户填写
	AmountGrams float64
	Notes       string // 转账备注

	// 待接收	StatusPendingReceipt   = "pending_receipt"
	// 已完成	StatusCompleted        = "completed"
	// 以拒收	StatusRejected         = "rejected"
	// 已超时	StatusExpired          = "expired"
	Status string // 系统填写
	//TransferType string 从表名获取

	//系统填写
	BalanceAfterTransfer float64 // 转出后，FromUser账户余额，对账单审计用

	CreatedAt   time.Time  // 创建，流程开始时间，也是锁定额度起始时间
	ExpiresAt   time.Time  // 过期时间，也是锁定额度截止时间
	PaymentTime *time.Time // （清算成功才有值）实际支付时间，关联对方确认接收时间
	UpdatedAt   *time.Time
}

// 用户对用户，专用茶叶转账接收记录结构体
type TeaUserFromUserTransferIn struct {
	Id   int
	Uuid string
	// 系统填写，对接转出表单，必填
	UserToUserTransferOutId int    // 用户对用户转出记录id
	ToUserId                int    // 接收用户id，账户持有人ID
	ToUserName              string // 接收用户名称，对账单审计用
	FromUserId              int    // 转出用户id
	FromUserName            string // 转出用户名称，对账单审计用

	AmountGrams         float64 // 接收转账额度（克），对账单审计用
	Notes               string  // 转出方备注（从转出表复制过来）
	BalanceAfterReceipt float64 // 接收后账户余额，对账单审计用

	// 已完成	StatusCompleted        = "completed"
	// 已拒收	StatusRejected         = "rejected"
	Status string //方便阅读，对账单审计用

	// 接收方ToUser操作，Confirmed/Rejected二选一
	IsConfirmed              bool           // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId        int            // 操作用户id，确认接收或者拒绝接收的用户id
	ReceptionRejectionReason sql.NullString // 如果拒绝，填写原因

	ExpiresAt time.Time // 过期时间，接收截止时间，也是FromUser解锁额度时间
	CreatedAt time.Time // 必填，如果接收，是接收、清算时间；如果拒绝，是拒绝时间
}

// 用户对团队，专用茶叶转账接收记录结构体
type TeaUserFromTeamTransferIn struct {
	Id   int
	Uuid string
	// 系统填写，对接转出表单，必填
	TeamToUserTransferOutId int     // 团队对用户转出记录id
	ToUserId                int     // 接收用户id，账户持有人ID
	ToUserName              string  // 接收用户名称，对账单审计用
	FromTeamId              int     // 转出团队id
	FromTeamName            string  // 转出团队名称，对账单审计用
	AmountGrams             float64 // 接收转账额度（克），对账单审计用
	Notes                   string  // 转出方备注（从转出表复制过来）
	BalanceAfterReceipt     float64 // 接收后账户余额，对账单审计用

	// 已完成	StatusCompleted        = "completed"
	// 已拒收	StatusRejected         = "rejected"
	Status string //方便阅读，对账单审计用
	// 接收方ToUser操作，Confirmed/Rejected二选一
	IsConfirmed              bool      // 默认false，默认不接收，避免转账错误被误接收
	ReceptionRejectionReason *string   // 如果拒绝，填写原因
	ExpiresAt                time.Time // 过期时间，接收截止时间，也是FromTeam解锁额度时间
	CreatedAt                time.Time // 必填，如果接收，是清算时间；如果拒绝，是拒绝时间
}

// 注：原TransactionRecord结构体已废弃，交易流水数据可以从转出表和转入表中推导出来

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
		return fmt.Errorf("开始事务失败: %v", err)
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

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	account.BalanceGrams = newBalance
	return nil
}

// CreateTeaTransferUserToUser 创建用户对用户类型转账记录
func CreateTeaTransferUserToUser(fromUserId, toUserId int, amount float64, notes string, expireHours int) (TeaUserToUserTransferOut, error) {
	// 验证参数
	if amount <= 0 {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：转账金额必须大于0")
	}
	if fromUserId == toUserId {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：不能给自己转账")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaUserToUserTransferOut{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出账户余额和锁定金额
	var fromBalance, fromLockedBalance float64
	var fromStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", fromUserId).
		Scan(&fromBalance, &fromLockedBalance, &fromStatus)
	if err != nil {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：查询转出账户失败 - %v", err)
	}
	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := fromBalance - fromLockedBalance
	if availableBalance < amount {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", fromBalance, fromLockedBalance, availableBalance)
	}

	// 确保接收方账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea.user_accounts WHERE user_id = $1", toUserId).Scan(&toAccountId)
	if err != nil {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：接收方账户不存在 - %v", err)
	}

	// 获取用户名称用于审计
	var fromUserName, toUserName string
	err = tx.QueryRow("SELECT name FROM users WHERE id = $1", fromUserId).Scan(&fromUserName)
	if err != nil {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：获取转出用户名称失败 - %v", err)
	}
	err = tx.QueryRow("SELECT name FROM users WHERE id = $1", toUserId).Scan(&toUserName)
	if err != nil {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：获取接收用户名称失败 - %v", err)
	}

	// 创建转账记录
	transfer := TeaUserToUserTransferOut{
		FromUserId:   fromUserId,
		FromUserName: fromUserName,
		ToUserId:     toUserId,
		ToUserName:   toUserName,
		AmountGrams:  amount,
		Status:       TeaTransferStatusPendingReceipt,
		Notes:        notes,
		ExpiresAt:    time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:    time.Now(),
	}

	err = tx.QueryRow(`INSERT INTO tea.user_to_user_transfer_out 
		(from_user_id, from_user_name, to_user_id, to_user_name, amount_grams, status, notes, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, uuid`,
		fromUserId, fromUserName, toUserId, toUserName, amount, TeaTransferStatusPendingReceipt, notes, transfer.ExpiresAt).
		Scan(&transfer.Id, &transfer.Uuid)
	if err != nil {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：创建转账记录失败 - %v", err)
	}

	// 锁定转出账户的相应金额
	_, err = tx.Exec("UPDATE tea.user_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), fromUserId)
	if err != nil {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：锁定转出账户金额失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaUserToUserTransferOut{}, fmt.Errorf("转账失败：提交事务失败 - %v", err)
	}

	return transfer, nil
}

// CreateTeaTransferUserToTeam 发起用户向团队的茶叶转账
func CreateTeaTransferUserToTeam(fromUserId, toTeamId int, amount float64, notes string, expireHours int) (TeaUserToTeamTransferOut, error) {
	// 验证参数
	if amount <= 0 {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：转账金额必须大于0")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出账户余额和锁定金额
	var fromBalance, fromLockedBalance float64
	var fromStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", fromUserId).
		Scan(&fromBalance, &fromLockedBalance, &fromStatus)
	if err != nil {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：查询转出账户失败 - %v", err)
	}
	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := fromBalance - fromLockedBalance
	if availableBalance < amount {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", fromBalance, fromLockedBalance, availableBalance)
	}

	// 确保接收方团队账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea.team_accounts WHERE team_id = $1", toTeamId).Scan(&toAccountId)
	if err != nil {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：接收方团队账户不存在 - %v", err)
	}

	// 获取用户名称和团队名称用于审计
	var fromUserName, toTeamName string
	err = tx.QueryRow("SELECT name FROM users WHERE id = $1", fromUserId).Scan(&fromUserName)
	if err != nil {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：获取转出用户名称失败 - %v", err)
	}
	err = tx.QueryRow("SELECT name FROM teams WHERE id = $1", toTeamId).Scan(&toTeamName)
	if err != nil {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：获取接收团队名称失败 - %v", err)
	}

	// 创建转账记录
	transfer := TeaUserToTeamTransferOut{
		FromUserId:   fromUserId,
		FromUserName: fromUserName,
		ToTeamId:     toTeamId,
		ToTeamName:   toTeamName,
		AmountGrams:  amount,
		Status:       TeaTransferStatusPendingReceipt,
		Notes:        notes,
		ExpiresAt:    time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:    time.Now(),
	}

	err = tx.QueryRow(`INSERT INTO tea.user_to_team_transfer_out 
		(from_user_id, from_user_name, to_team_id, to_team_name, amount_grams, status, notes, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, uuid`,
		fromUserId, fromUserName, toTeamId, toTeamName, amount, TeaTransferStatusPendingReceipt, notes, transfer.ExpiresAt).
		Scan(&transfer.Id, &transfer.Uuid)
	if err != nil {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：创建转账记录失败 - %v", err)
	}

	// 锁定转出账户的相应金额
	_, err = tx.Exec("UPDATE tea.user_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), fromUserId)
	if err != nil {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：锁定转出账户金额失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaUserToTeamTransferOut{}, fmt.Errorf("转账失败：提交事务失败 - %v", err)
	}

	return transfer, nil
}

// getAndValidateUserToUserTransfer 获取并验证用户对用户转账信息
func getAndValidateUserToUserTransfer(tx *sql.Tx, transferUuid string, toUserId int) (TeaUserToUserTransferOut, error) {
	var transfer TeaUserToUserTransferOut
	err := tx.QueryRow(`SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name, 
		amount_grams, status, expires_at, notes
		FROM tea.user_to_user_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.AmountGrams,
			&transfer.Status, &transfer.ExpiresAt, &transfer.Notes)
	if err != nil {
		return transfer, fmt.Errorf("确认转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingReceipt {
		return transfer, fmt.Errorf("确认转账失败：转账状态异常")
	}
	if time.Now().After(transfer.ExpiresAt) {
		// 转账已过期，更新状态
		_, _ = tx.Exec("UPDATE tea.user_to_user_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			TeaTransferStatusExpired, time.Now(), transfer.Id)
		return transfer, fmt.Errorf("确认转账失败：转账已过期")
	}

	// 检查确认权限
	if transfer.ToUserId != toUserId {
		return transfer, fmt.Errorf("无权确认此转账")
	}

	return transfer, nil
}

// getAndValidateUserToTeamTransfer 获取并验证用户对团队转账信息
// func getAndValidateUserToTeamTransfer(tx *sql.Tx, transferUuid string) (TeaUserToTeamTransferOut, error) {
// 	var transfer TeaUserToTeamTransferOut
// 	err := tx.QueryRow(`SELECT id, uuid, from_user_id, from_user_name, to_team_id, to_team_name,
// 		amount_grams, status, expires_at, notes
// 		FROM tea.user_to_team_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
// 		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
// 			&transfer.ToTeamId, &transfer.ToTeamName, &transfer.AmountGrams,
// 			&transfer.Status, &transfer.ExpiresAt, &transfer.Notes)
// 	if err != nil {
// 		return transfer, fmt.Errorf("确认转账失败：转账记录不存在 - %v", err)
// 	}

// 	// 验证状态
// 	if transfer.Status != TeaTransferStatusPendingReceipt {
// 		return transfer, fmt.Errorf("确认转账失败：转账状态异常")
// 	}
// 	if time.Now().After(transfer.ExpiresAt) {
// 		// 转账已过期，更新状态
// 		_, _ = tx.Exec("UPDATE tea.user_to_team_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
// 			TeaTransferStatusExpired, time.Now(), transfer.Id)
// 		return transfer, fmt.Errorf("确认转账失败：转账已过期")
// 	}

// 	// 检查确认权限（团队成员才能确认）
// 	// 暂时注释掉团队相关功能
// 	// isMember, err := IsTeamMember(toUserId, transfer.ToTeamId)
// 	// if err != nil {
// 	// 	return transfer, fmt.Errorf("检查团队成员身份失败: %v", err)
// 	// }
// 	// if !isMember {
// 	// 	return transfer, fmt.Errorf("只有团队成员才能确认团队转账")
// 	// }

// 	return transfer, nil
// }

// confirmUserToUserTransfer 确认用户对用户转账
func confirmUserToUserTransfer(tx *sql.Tx, transfer TeaUserToUserTransferOut, toUserId int) error {
	// 获取接收方账户余额
	var toUserBalance float64
	err := tx.QueryRow("SELECT balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", toUserId).Scan(&toUserBalance)
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
		toUserBalance+transfer.AmountGrams, time.Now(), toUserId)
	if err != nil {
		return fmt.Errorf("更新接收账户余额失败: %v", err)
	}

	// 更新转账状态，记录转出后余额
	paymentTime := time.Now()
	_, err = tx.Exec(`UPDATE tea.user_to_user_transfer_out SET 
		status = $1, 
		balance_after_transfer = $2,
		payment_time = $3, 
		updated_at = $4 
		WHERE id = $5`,
		TeaTransferStatusCompleted, newFromBalance, paymentTime, paymentTime, transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建转入记录，记录接收后余额
	_, err = tx.Exec(`INSERT INTO tea.user_from_user_transfer_in
		(user_to_user_transfer_out_id, to_user_id, to_user_name, from_user_id, from_user_name,
		amount_grams, notes, balance_after_receipt, status, is_confirmed, operational_user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		transfer.Id, toUserId, transfer.ToUserName, transfer.FromUserId, transfer.FromUserName,
		transfer.AmountGrams, transfer.Notes, toUserBalance+transfer.AmountGrams, TeaTransferStatusCompleted, true, toUserId, transfer.ExpiresAt, paymentTime)
	if err != nil {
		return fmt.Errorf("创建转入记录失败: %v", err)
	}

	return nil
}

// recordTeamTransferTransactions 记录团队转账的交易流水（已废弃，不再使用交易流水表）
// func recordTeamTransferTransactions(transfer TeaUserTransferOut, fromBalance, newFromBalance, teamBalance float64) error {
// 	// 注：交易流水表已废弃，不再记录交易流水
// 	return nil
// }

// recordPersonalTransferTransactions 记录用户转账的交易流水（已废弃，不再使用交易流水表）
// func recordPersonalTransferTransactions(tx *sql.Tx, transfer TeaUserTransferOut, fromBalance, newFromBalance, toBalance float64, toUserId int) error {
// 	// 注：交易流水表已废弃，不再记录交易流水
// 	return nil
// }

// ConfirmTeaTransfer 确认接收转账（支持用户间转账和用户向团队转账）
func ConfirmTeaTransfer(transferUuid string, toUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 根据转账类型获取并验证转账信息
	// 先尝试用户对用户转账
	userToUserTransfer, err := getAndValidateUserToUserTransfer(tx, transferUuid, toUserId)
	if err == nil {
		err = confirmUserToUserTransfer(tx, userToUserTransfer, toUserId)
		if err != nil {
			return err
		}
		// 提交事务
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %v", err)
		}
		return nil
	}

	// 如果不是用户对用户转账，尝试用户对团队转账
	// 暂时注释掉团队相关功能
	// userToTeamTransfer, err := getAndValidateUserToTeamTransfer(tx, transferUuid, toUserId)
	// if err == nil {
	// 	err = confirmUserToTeamTransfer(tx, userToTeamTransfer, toUserId)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// 提交事务
	// 	if err = tx.Commit(); err != nil {
	// 		return fmt.Errorf("提交事务失败: %v", err)
	// 	}
	// 	return nil
	// }

	return fmt.Errorf("确认转账失败：找不到对应的转账记录")
}

// RejectTeaTransfer 拒绝用户对用户转账
func RejectTeaTransfer(transferUuid string, toUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaUserToUserTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name, 
		amount_grams, status, expires_at, notes
		FROM tea.user_to_user_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.AmountGrams,
			&transfer.Status, &transfer.ExpiresAt, &transfer.Notes)
	if err != nil {
		return fmt.Errorf("拒绝转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingReceipt {
		return fmt.Errorf("拒绝转账失败：转账状态异常")
	}

	// 检查拒绝权限
	if transfer.ToUserId != toUserId {
		return fmt.Errorf("无权拒绝此转账")
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
	_, err = tx.Exec(`UPDATE tea.user_to_user_transfer_out SET 
		status = $1, 
		updated_at = $2 
		WHERE id = $3`,
		TeaTransferStatusRejected, time.Now(), transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建拒收记录
	_, err = tx.Exec(`INSERT INTO tea.user_from_user_transfer_in
		(user_to_user_transfer_out_id, to_user_id, to_user_name, from_user_id, from_user_name,
		amount_grams, notes, status, is_confirmed, operational_user_id, reception_rejection_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		transfer.Id, toUserId, transfer.ToUserName, transfer.FromUserId, transfer.FromUserName,
		transfer.AmountGrams, transfer.Notes, TeaTransferStatusRejected, false, toUserId, reason, time.Now())
	if err != nil {
		return fmt.Errorf("创建拒收记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// ProcessExpiredTransfers 处理过期的用户对用户转账，解锁相应的锁定金额
func ProcessExpiredTransfers() error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 查找所有过期且仍为pending状态的用户对用户转账
	rows, err := tx.Query(`
		SELECT id, from_user_id, amount_grams 
		FROM tea.user_to_user_transfer_out 
		WHERE status = $1 AND expires_at < $2`,
		TeaTransferStatusPendingReceipt, time.Now())
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
		_, err = tx.Exec("UPDATE tea.user_to_user_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			TeaTransferStatusExpired, time.Now(), et.Id)
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

// GetPendingTransfers 获取用户待确认的用户对用户转账列表
func GetPendingTransfers(userId int, page, limit int) ([]TeaUserToUserTransferOut, error) {
	offset := (page - 1) * limit
	// 查询用户待确认的用户对用户转账
	rows, err := DB.Query(`SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name, 
		amount_grams, status, notes, expires_at, created_at, updated_at 
		FROM tea.user_to_user_transfer_out 
		WHERE to_user_id = $1 
		AND status = $2 AND expires_at > NOW() 
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		userId, TeaTransferStatusPendingReceipt, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待确认转账失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaUserToUserTransferOut
	for rows.Next() {
		var transfer TeaUserToUserTransferOut
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.AmountGrams, &transfer.Status,
			&transfer.Notes, &transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetUserTransactions 获取用户交易历史（仅限用户对用户转账）
func GetUserTransactions(userId int, page, limit int) ([]map[string]interface{}, error) {
	offset := (page - 1) * limit

	// 查询用户作为转出方的交易历史
	rows, err := DB.Query(`
		SELECT 
			'outgoing' as transaction_type,
			uuid,
			from_user_id,
			from_user_name,
			to_user_id,
			to_user_name,
			amount_grams,
			status,
			notes,
			payment_time,
			created_at
		FROM tea.user_to_user_transfer_out 
		WHERE from_user_id = $1 AND status = 'completed'
		
		UNION ALL
		
		-- 查询用户作为接收方的交易历史
		SELECT 
			'incoming' as transaction_type,
			uto.uuid,
			uto.from_user_id,
			uto.from_user_name,
			uto.to_user_id,
			uto.to_user_name,
			uto.amount_grams,
			uto.status,
			uto.notes,
			uto.payment_time,
			uto.created_at
		FROM tea.user_to_user_transfer_out uto
		INNER JOIN tea.user_from_user_transfer_in uin ON uto.id = uin.user_to_user_transfer_out_id
		WHERE uto.to_user_id = $1 AND uto.status = 'completed' AND uin.status = 'completed'
		
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userId, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("查询交易历史失败: %v", err)
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var transactionType, uuid, status, notes, fromUserName, toUserName string
		var fromUserId, toUserId int
		var amountGrams float64
		var paymentTime, createdAt sql.NullTime

		err = rows.Scan(&transactionType, &uuid, &fromUserId, &fromUserName, &toUserId, &toUserName,
			&amountGrams, &status, &notes, &paymentTime, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("扫描交易记录失败: %v", err)
		}

		transaction := map[string]interface{}{
			"transaction_type": transactionType,
			"uuid":             uuid,
			"from_user_id":     fromUserId,
			"from_user_name":   fromUserName,
			"to_user_id":       toUserId,
			"to_user_name":     toUserName,
			"amount_grams":     amountGrams,
			"status":           status,
			"notes":            notes,
			"payment_time":     getNullableTime(paymentTime),
			"created_at":       getNullableTime(createdAt),
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetTransferHistory 获取用户转账历史（仅限用户对用户转账）
func GetTransferHistory(userId int, page, limit int) ([]TeaUserToUserTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name, 
		amount_grams, status, notes, balance_after_transfer, expires_at, payment_time, created_at, updated_at 
		FROM tea.user_to_user_transfer_out 
		WHERE from_user_id = $1 OR to_user_id = $1 
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询转账历史失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaUserToUserTransferOut
	for rows.Next() {
		var transfer TeaUserToUserTransferOut
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.AmountGrams, &transfer.Status,
			&transfer.Notes, &transfer.BalanceAfterTransfer, &transfer.ExpiresAt,
			&transfer.PaymentTime, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetTransferOuts 获取用户转出记录（从转出方视角，仅限用户对用户转账）
func GetTransferOuts(userId int, page, limit int) ([]TeaUserToUserTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name, 
			   amount_grams, status, notes, balance_after_transfer, expires_at, payment_time, created_at, updated_at 
		FROM tea.user_to_user_transfer_out 
		WHERE from_user_id = $1 
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询转出记录失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaUserToUserTransferOut
	for rows.Next() {
		var transfer TeaUserToUserTransferOut
		err = rows.Scan(
			&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName, &transfer.ToUserId, &transfer.ToUserName,
			&transfer.AmountGrams, &transfer.Status, &transfer.Notes, &transfer.BalanceAfterTransfer,
			&transfer.ExpiresAt, &transfer.PaymentTime, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转出记录失败: %v", err)
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

// GetTransferIns 获取用户转入记录（从接收方视角，仅限用户对用户转账）
func GetTransferIns(userId int, page, limit int) ([]TeaUserFromUserTransferIn, error) {
	offset := (page - 1) * limit
	// 查询用户转入记录（从user_from_user_transfer_in表）
	rows, err := DB.Query(`
		SELECT id, uuid, user_to_user_transfer_out_id, to_user_id, to_user_name, from_user_id, from_user_name,
			   amount_grams, notes, balance_after_receipt, status, is_confirmed, operational_user_id, reception_rejection_reason, expires_at, created_at
		FROM tea.user_from_user_transfer_in
		WHERE to_user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询转入记录失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaUserFromUserTransferIn
	for rows.Next() {
		var transfer TeaUserFromUserTransferIn
		err = rows.Scan(
			&transfer.Id, &transfer.Uuid, &transfer.UserToUserTransferOutId, &transfer.ToUserId, &transfer.ToUserName,
			&transfer.FromUserId, &transfer.FromUserName, &transfer.AmountGrams, &transfer.Notes, &transfer.BalanceAfterReceipt,
			&transfer.Status, &transfer.IsConfirmed, &transfer.OperationalUserId, &transfer.ReceptionRejectionReason,
			&transfer.ExpiresAt, &transfer.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转入记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// 辅助函数
// getNullableInt64 处理sql.NullInt64，返回正确的值（有效时返回int64，无效时返回nil）
func getNullableInt64(nullInt sql.NullInt64) interface{} {
	if nullInt.Valid {
		return nullInt.Int64
	}
	return nil
}

// 辅助函数
// getNullableTime 处理sql.NullTime，返回正确的值（有效时返回time.Time，无效时返回nil）
func getNullableTime(nullTime sql.NullTime) interface{} {
	if nullTime.Valid {
		return nullTime.Time
	}
	return nil
}
