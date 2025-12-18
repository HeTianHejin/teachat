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
	Id               int
	Uuid             string
	UserId           int
	BalanceGrams     float64 // 茶叶数量(克)
	LockedBalanceGrams float64 // 被锁定的茶叶数量(克)
	Status           string  // normal, frozen
	FrozenReason     *string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

// 茶叶转账结构体
type TeaTransfer struct {
	Id              int
	Uuid            string
	FromUserId      int
	ToUserId        *int // 使用指针支持NULL值，团队转账时为nil
	ToTeamId        *int // 使用指针支持NULL值，用户间转账时为nil
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
	err := DB.QueryRow("SELECT id, uuid, user_id, balance_grams, locked_balance_grams, status, frozen_reason, created_at, updated_at FROM tea_accounts WHERE user_id = $1", userId).
		Scan(&account.Id, &account.Uuid, &account.UserId, &account.BalanceGrams, &account.LockedBalanceGrams, &account.Status, &account.FrozenReason, &account.CreatedAt, &account.UpdatedAt)
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
	stmt, err := DB.Prepare(statement)
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

// SystemAdjustBalance 系统调整余额
func (account *TeaAccount) SystemAdjustBalance(amount float64, description string, adminUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
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
		(user_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id, target_type) 
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
	tx, err := DB.Begin()
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出账户余额和锁定金额
	var fromBalance, fromLockedBalance float64
	var fromStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea_accounts WHERE user_id = $1 FOR UPDATE", fromUserId).
		Scan(&fromBalance, &fromLockedBalance, &fromStatus)
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("查询转出账户失败: %v", err)
	}
	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := fromBalance - fromLockedBalance
	if availableBalance < amount {
		return TeaTransfer{}, fmt.Errorf("可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", fromBalance, fromLockedBalance, availableBalance)
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
		ToUserId:    &toUserId,
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

	// 锁定转出账户的相应金额
	_, err = tx.Exec("UPDATE tea_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), fromUserId)
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("锁定转出账户金额失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaTransfer{}, fmt.Errorf("提交事务失败: %v", err)
	}

	return transfer, nil
}

// CreateTeaTransferToTeam 发起用户向团队的茶叶转账
func CreateTeaTransferToTeam(fromUserId, toTeamId int, amount float64, notes string, expireHours int) (TeaTransfer, error) {
	// 验证参数
	if amount <= 0 {
		return TeaTransfer{}, fmt.Errorf("转账金额必须大于0")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出账户余额和锁定金额
	var fromBalance, fromLockedBalance float64
	var fromStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea_accounts WHERE user_id = $1 FOR UPDATE", fromUserId).
		Scan(&fromBalance, &fromLockedBalance, &fromStatus)
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("查询转出账户失败: %v", err)
	}
	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := fromBalance - fromLockedBalance
	if availableBalance < amount {
		return TeaTransfer{}, fmt.Errorf("可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", fromBalance, fromLockedBalance, availableBalance)
	}

	// 确保接收方团队账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM team_tea_accounts WHERE team_id = $1", toTeamId).Scan(&toAccountId)
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("接收方团队账户不存在: %v", err)
	}

	// 创建转账记录
	transfer := TeaTransfer{
		FromUserId:  fromUserId,
		ToUserId:    nil, // 团队转账时to_user_id在数据库中设为NULL，使用to_team_id
		ToTeamId:    &toTeamId,
		AmountGrams: amount,
		Status:      TransferStatus_Pending,
		Notes:       notes,
		ExpiresAt:   time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:   time.Now(),
	}

	err = tx.QueryRow(`INSERT INTO tea_transfers 
		(from_user_id, to_user_id, amount_grams, status, notes, expires_at, to_team_id) 
		VALUES ($1, NULL, $2, $3, $4, $5, $6) RETURNING id, uuid`,
		fromUserId, amount, TransferStatus_Pending, notes, transfer.ExpiresAt, toTeamId).
		Scan(&transfer.Id, &transfer.Uuid)
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("创建转账记录失败: %v", err)
	}

	// 锁定转出账户的相应金额
	_, err = tx.Exec("UPDATE tea_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), fromUserId)
	if err != nil {
		return TeaTransfer{}, fmt.Errorf("锁定转出账户金额失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaTransfer{}, fmt.Errorf("提交事务失败: %v", err)
	}

	return transfer, nil
}

// ConfirmTeaTransfer 确认接收转账（支持用户间转账和用户向团队转账）
func ConfirmTeaTransfer(transferUuid string, toUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息（包含团队信息）
	var transfer TeaTransfer
	err = tx.QueryRow(`SELECT id, uuid, from_user_id, to_user_id, to_team_id, amount_grams, status, expires_at 
		FROM tea_transfers WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId, &transfer.ToTeamId,
			&transfer.AmountGrams, &transfer.Status, &transfer.ExpiresAt)
	if err != nil {
		return fmt.Errorf("转账记录不存在: %v", err)
	}

	// 验证状态
	if transfer.Status != TransferStatus_Pending {
		return fmt.Errorf("转账状态异常")
	}
	if time.Now().After(transfer.ExpiresAt) {
		// 转账已过期，更新状态
		_, _ = tx.Exec("UPDATE tea_transfers SET status = $1, updated_at = $2 WHERE id = $3",
			TransferStatus_Expired, time.Now(), transfer.Id)
		return fmt.Errorf("转账已过期")
	}

	// 检查确认权限
	if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
		// 团队转账：检查用户是否是团队成员
		isMember, err := IsTeamMember(toUserId, *transfer.ToTeamId)
		if err != nil {
			return fmt.Errorf("检查团队成员身份失败: %v", err)
		}
		if !isMember {
			return fmt.Errorf("只有团队成员才能确认团队转账")
		}
	} else {
		// 用户间转账：检查接收用户ID
		if transfer.ToUserId == nil || *transfer.ToUserId != toUserId {
			return fmt.Errorf("无权确认此转账")
		}
	}

	// 获取转出账户信息（包含锁定金额）
	var fromBalance, fromLockedBalance float64
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", transfer.FromUserId).
		Scan(&fromBalance, &fromLockedBalance)
	if err != nil {
		return fmt.Errorf("查询转出账户信息失败: %v", err)
	}

	// 根据转账类型执行不同的确认逻辑
	if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
		// 团队转账确认
		// 确保团队账户存在
		var teamBalance float64
		err = tx.QueryRow("SELECT balance_grams FROM team_tea_accounts WHERE team_id = $1 FOR UPDATE", *transfer.ToTeamId).Scan(&teamBalance)
		if err != nil {
			return fmt.Errorf("查询团队账户余额失败: %v", err)
		}

		// 实际扣除转出账户余额并解锁锁定金额
		newFromBalance := fromBalance - transfer.AmountGrams
		newFromLockedBalance := fromLockedBalance - transfer.AmountGrams
		
		// 检查锁定余额是否足够解锁
		if newFromLockedBalance < 0 {
			return fmt.Errorf("锁定余额不足，无法完成转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromLockedBalance, transfer.AmountGrams)
		}
		
		// 检查账户余额是否足够
		if newFromBalance < 0 {
			return fmt.Errorf("账户余额不足，无法完成转账。当前余额: %.3f克, 转账金额: %.3f克", fromBalance, transfer.AmountGrams)
		}
		
		_, err = tx.Exec("UPDATE tea_accounts SET balance_grams = $1, locked_balance_grams = $2, updated_at = $3 WHERE user_id = $4",
			newFromBalance, newFromLockedBalance, time.Now(), transfer.FromUserId)
		if err != nil {
			return fmt.Errorf("更新转出账户余额失败: %v", err)
		}

		// 增加团队账户余额
		_, err = tx.Exec("UPDATE team_tea_accounts SET balance_grams = $1, updated_at = $2 WHERE team_id = $3",
			teamBalance+transfer.AmountGrams, time.Now(), *transfer.ToTeamId)
		if err != nil {
			return fmt.Errorf("更新团队账户余额失败: %v", err)
		}

		// 更新转账状态
		paymentTime := time.Now()
		_, err = tx.Exec("UPDATE tea_transfers SET status = $1, payment_time = $2, updated_at = $3 WHERE id = $4",
			TransferStatus_Confirmed, paymentTime, paymentTime, transfer.Id)
		if err != nil {
			return fmt.Errorf("更新转账状态失败: %v", err)
		}

		// 记录转出方交易流水
		_, err = tx.Exec(`INSERT INTO tea_transactions 
			(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, related_team_id, target_type) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			transfer.FromUserId, &transfer.Uuid, TransactionType_TransferOut, transfer.AmountGrams,
			fromBalance, newFromBalance, fmt.Sprintf("向团队转账: %s", transfer.Notes), transfer.ToTeamId, TransactionTargetType_Team)
		if err != nil {
			return fmt.Errorf("记录转出交易流水失败: %v", err)
		}

		// 记录团队转入交易流水
		_, err = tx.Exec(`INSERT INTO team_tea_transactions 
			(team_id, operation_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id, target_type) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			*transfer.ToTeamId, nil, TransactionType_TransferIn, transfer.AmountGrams,
			teamBalance, teamBalance+transfer.AmountGrams, "用户转账转入", &transfer.FromUserId, TransactionTargetType_User)
		if err != nil {
			return fmt.Errorf("记录团队转入交易流水失败: %v", err)
		}

	} else {
		// 用户间转账确认
		// 获取接收方账户余额
		var toBalance float64
		err = tx.QueryRow("SELECT balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", toUserId).Scan(&toBalance)
		if err != nil {
			return fmt.Errorf("查询接收账户余额失败: %v", err)
		}

		// 实际扣除转出账户余额并解锁锁定金额
		newFromBalance := fromBalance - transfer.AmountGrams
		newFromLockedBalance := fromLockedBalance - transfer.AmountGrams
		
		// 检查锁定余额是否足够解锁
		if newFromLockedBalance < 0 {
			return fmt.Errorf("锁定余额不足，无法完成转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromLockedBalance, transfer.AmountGrams)
		}
		
		// 检查账户余额是否足够
		if newFromBalance < 0 {
			return fmt.Errorf("账户余额不足，无法完成转账。当前余额: %.3f克, 转账金额: %.3f克", fromBalance, transfer.AmountGrams)
		}
		
		_, err = tx.Exec("UPDATE tea_accounts SET balance_grams = $1, locked_balance_grams = $2, updated_at = $3 WHERE user_id = $4",
			newFromBalance, newFromLockedBalance, time.Now(), transfer.FromUserId)
		if err != nil {
			return fmt.Errorf("更新转出账户余额失败: %v", err)
		}

		// 增加接收账户余额
		_, err = tx.Exec("UPDATE tea_accounts SET balance_grams = $1, updated_at = $2 WHERE user_id = $3",
			toBalance+transfer.AmountGrams, time.Now(), toUserId)
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

		// 记录转出方交易流水
		_, err = tx.Exec(`INSERT INTO tea_transactions 
			(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id, target_type) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			transfer.FromUserId, &transfer.Uuid, TransactionType_TransferOut, transfer.AmountGrams,
			fromBalance, newFromBalance, fmt.Sprintf("转账给用户: %s", transfer.Notes), toUserId, TransactionTargetType_User)
		if err != nil {
			return fmt.Errorf("记录转出交易流水失败: %v", err)
		}

		// 记录接收方交易流水
		_, err = tx.Exec(`INSERT INTO tea_transactions 
			(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, related_user_id, target_type) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			toUserId, &transfer.Uuid, TransactionType_TransferIn, transfer.AmountGrams,
			toBalance, toBalance+transfer.AmountGrams, "转账转入", &transfer.FromUserId, TransactionTargetType_User)
		if err != nil {
			return fmt.Errorf("记录转入交易流水失败: %v", err)
		}
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
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息（包含团队信息和金额）
	var transfer TeaTransfer
	err = tx.QueryRow("SELECT id, from_user_id, to_user_id, to_team_id, amount_grams, status FROM tea_transfers WHERE uuid = $1 FOR UPDATE", transferUuid).
		Scan(&transfer.Id, &transfer.FromUserId, &transfer.ToUserId, &transfer.ToTeamId, &transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("转账记录不存在: %v", err)
	}

	// 验证状态
	if transfer.Status != TransferStatus_Pending {
		return fmt.Errorf("转账状态异常")
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
	err = tx.QueryRow("SELECT locked_balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", transfer.FromUserId).Scan(&fromLockedBalance)
	if err != nil {
		return fmt.Errorf("查询转出账户锁定金额失败: %v", err)
	}

	// 解锁转出账户的锁定金额
	newFromLockedBalance := fromLockedBalance - transfer.AmountGrams
	
	// 检查锁定余额是否足够解锁
	if newFromLockedBalance < 0 {
		return fmt.Errorf("锁定余额不足，无法拒绝转账。当前锁定余额: %.3f克, 转账金额: %.3f克", fromLockedBalance, transfer.AmountGrams)
	}
	
	_, err = tx.Exec("UPDATE tea_accounts SET locked_balance_grams = $1, updated_at = $2 WHERE user_id = $3",
		newFromLockedBalance, time.Now(), transfer.FromUserId)
	if err != nil {
		return fmt.Errorf("解锁转出账户金额失败: %v", err)
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

// ProcessExpiredTransfers 处理过期的转账，解锁相应的锁定金额
func ProcessExpiredTransfers() error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 查找所有过期且仍为pending状态的转账
	rows, err := tx.Query(`
		SELECT id, from_user_id, amount_grams 
		FROM tea_transfers 
		WHERE status = $1 AND expires_at < $2`,
		TransferStatus_Pending, time.Now())
	if err != nil {
		return fmt.Errorf("查询过期转账失败: %v", err)
	}
	defer rows.Close()

	var expiredTransfers []struct {
		Id        int
		FromUserId int
		Amount    float64
	}

	for rows.Next() {
		var et struct {
			Id        int
			FromUserId int
			Amount    float64
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
		err = tx.QueryRow("SELECT locked_balance_grams FROM tea_accounts WHERE user_id = $1 FOR UPDATE", et.FromUserId).Scan(&currentLockedBalance)
		if err != nil {
			return fmt.Errorf("查询锁定余额失败: %v", err)
		}
		
		// 检查锁定余额是否足够
		if currentLockedBalance < et.Amount {
			// 锁定余额不足，记录警告并跳过
			continue
		}
		
		// 更新转账状态为过期
		_, err = tx.Exec("UPDATE tea_transfers SET status = $1, updated_at = $2 WHERE id = $3",
			TransferStatus_Expired, time.Now(), et.Id)
		if err != nil {
			return fmt.Errorf("更新过期转账状态失败: %v", err)
		}

		// 解锁相应的锁定金额
		newLockedBalance := currentLockedBalance - et.Amount
		_, err = tx.Exec("UPDATE tea_accounts SET locked_balance_grams = $1, updated_at = $2 WHERE user_id = $3",
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
func GetPendingTransfers(userId int, page, limit int) ([]TeaTransfer, error) {
	offset := (page - 1) * limit
	// 查询用户个人待确认转账 + 用户所属团队的待确认转账
	rows, err := DB.Query(`SELECT DISTINCT t.id, t.uuid, t.from_user_id, t.to_user_id, t.to_team_id, 
		t.amount_grams, t.status, t.payment_time, t.notes, t.rejection_reason, t.expires_at, t.created_at, t.updated_at 
		FROM tea_transfers t 
		LEFT JOIN team_members tm ON t.to_team_id = tm.team_id AND tm.user_id = $1
		WHERE (t.to_user_id = $1 OR (tm.user_id = $1 AND t.to_team_id IS NOT NULL)) 
		AND t.status = $2 AND t.expires_at > NOW() 
		ORDER BY t.created_at DESC LIMIT $3 OFFSET $4`,
		userId, TransferStatus_Pending, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待确认转账失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaTransfer
	for rows.Next() {
		var transfer TeaTransfer
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId,
			&transfer.ToTeamId, &transfer.AmountGrams, &transfer.Status, &transfer.PaymentTime, &transfer.Notes,
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
		rows, err = DB.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, related_user_id, created_at 
			FROM tea_transactions WHERE user_id = $1 
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userId, limit, offset)
	} else {
		rows, err = DB.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
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
	rows, err := DB.Query(`SELECT id, uuid, from_user_id, to_user_id, to_team_id, amount_grams, status, 
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
			&transfer.ToTeamId, &transfer.AmountGrams, &transfer.Status, &transfer.PaymentTime, &transfer.Notes,
			&transfer.RejectionReason, &transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetPendingTransfersForTeam 获取转入指定团队的待确认转账列表
func GetPendingTransfersForTeam(teamId int, page, limit int) ([]TeaTransfer, error) {
	offset := (page - 1) * limit
	// 查询转入团队的待确认转账
	rows, err := DB.Query(`SELECT t.id, t.uuid, t.from_user_id, t.to_user_id, t.to_team_id, 
		t.amount_grams, t.status, t.payment_time, t.notes, t.rejection_reason, t.expires_at, t.created_at, t.updated_at 
		FROM tea_transfers t 
		WHERE t.to_team_id = $1 
		AND t.status = $2 AND t.expires_at > NOW() 
		ORDER BY t.created_at DESC LIMIT $3 OFFSET $4`,
		teamId, TransferStatus_Pending, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询团队待确认转账失败: %v", err)
	}
	defer rows.Close()

	var transfers []TeaTransfer
	for rows.Next() {
		var transfer TeaTransfer
		err = rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.ToUserId,
			&transfer.ToTeamId, &transfer.AmountGrams, &transfer.Status, &transfer.PaymentTime, &transfer.Notes,
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
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM tea_accounts WHERE user_id = $1)", userId).Scan(&exists)
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
	err := DB.QueryRow("SELECT status, frozen_reason FROM tea_accounts WHERE user_id = $1", userId).
		Scan(&status, &frozenReason)
	if err != nil {
		return false, "", fmt.Errorf("查询账户状态失败: %v", err)
	}

	if status == TeaAccountStatus_Frozen {
		return true, frozenReason.String, nil
	}

	return false, "", nil
}
