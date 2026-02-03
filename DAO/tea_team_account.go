package dao

import (
	"database/sql"
	"fmt"
	util "teachat/Util"
	"time"
)

/*
-- 星茶是一种特殊虚拟现实资源/物品，（毫克）作为重量单位
-- 用户获得星茶来自向茶庄团队购买（1元/克）
-- 如果需要兑现，就向茶庄出售星茶（1克/元）

注意：用户星茶账户相关定义（TeaUserToUserTransferOut、TeaUserFromTeamTransferIn……）已迁移至tea_account.go文件，保留此注释以提示开发者--

【重要：团队转账与用户转账的处理方式差异】
- 用户转账：使用 locked_balance_milligrams 锁定金额（等待对方确认后才扣减余额）
- 团队转账：直接扣减 balance_milligrams（进入待审批状态，审批通过后直接完成）
原因：团队转账有额外的审批流程，资金需要立即冻结，所以直接扣减余额更合适。

团队星茶账户转账流程：
1、发起方法：团队转出额度星茶，无论接收方是团队还是用户（个人），都要求由1成员填写转账表单，第一步创建待审核转账表单，
2、审核方法：由任意1核心成员审批待审核表单，
2.1、如果审核批准，转账表单状态更新为已批准，执行第3步锁定转账额度，
2.2、如果审核否决，转账表单状态更新为已否决，记录审批操作记录，流程结束。
3、锁定方法：已批准的转出团队账户转出额度星茶数量被锁定，防止重复发起转账；
4.1、接收方法：目标接受方，用户或者团队任意1状态正常成员TeamMemberStatusActive(1)，有效期内，操作“确认”，继续第5步结算双方账户，
4.2、拒收方法，目标接受方，用户或者团队任意1状态正常成员，有效期内，操作“拒收”。系统创建转入（transferIn）记录状态“已拒收”，同时out转账表单状态更新为“已拒收”，解锁被转出方锁定星茶，记录拒收操作用户id及时间原因，流程结束。
5、结算方法：按锁定额度（接收额度）清算双方账户数额，创建转入（transferIn）记录；
6、超时处理：自动解锁转出团队账户被锁定额度星茶，不创建接收目标方转入（transferIn）记录。
*/

// // 团队星茶账户状态常量
const (
	TeaTeamAccountStatus_Normal  = "normal"
	TeaTeamAccountStatus_Frozen  = "frozen"
	TeaTeamAccountStatus_Deleted = "deleted"
)

// 转账状态常量（统一状态枚举）
// const (
// 	// 团队转出特有状态
// 	TeaTransferStatusPendingApproval  = "pending_approval"  // 待团队审批
// 	TeaTransferStatusApproved         = "approved"          // 团队审批通过
// 	TeaTransferStatusApprovalRejected = "approval_rejected" // 团队审批拒绝

// 	// 用户和团队通用状态
// 	TeaTransferStatusPendingReceipt = "pending_receipt" // 待接收方确认
// 	TeaTransferStatusCompleted      = "completed"       // 转账完成
// 	TeaTransferStatusRejected       = "rejected"        // 接收方拒绝
// 	TeaTransferStatusExpired        = "expired"         // 已超时
// )

// 团队星茶账户结构体
type TeaTeamAccount struct {
	Id                      int
	Uuid                    string
	TeamId                  int
	BalanceMilligrams       int64  // 星茶数量(毫克）
	LockedBalanceMilligrams int64  // 被锁定的星茶数量(毫克）
	Status                  string // normal, frozen
	FrozenReason            string // 冻结原因,默认值:'-'
	CreatedAt               time.Time
	UpdatedAt               *time.Time
}

// 团队对用户转账结构体（完全匹配数据库表结构）
type TeaTeamToUserTransferOut struct {
	Id           int
	Uuid         string
	FromTeamId   int    // 转出团队ID
	FromTeamName string // 转出团队名称
	ToUserId     int    // 接收用户ID
	ToUserName   string // 接收用户名称

	InitiatorUserId  int    // 发起转账的用户id（团队成员）
	AmountMilligrams int64  // 转账星茶数量(毫克），即锁定数量
	Notes            string // 转账备注,默认值:'-'

	// 审批相关（团队转出时使用）
	IsOnlyOneMemberTeam bool // 默认值false:多人团队审批(必填)，true:单人团队自动批准
	// 审批人填写，必填
	IsApproved              bool      // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          int       // 审批人ID，团队核心成员id，多人团队不能是发起人id，（必填）单人团队自动批准时，审批人是发起人自己
	ApprovalRejectionReason string    // 审批意见，如果拒绝，填写原因,默认值:'-'
	ApprovedAt              time.Time // 审批时间

	// 注意流程：审批通过而且对方确认接受之后，才会创建待接收记录（TeaUserFromTeamTransferIn）
	Status               string // 包含审批状态，待审批，已批准，已拒绝，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer int64  // 转账后余额(毫克）
	CreatedAt            time.Time
	ExpiresAt            time.Time  // 转账请求过期时间，同时是解除锁定额度时间
	PaymentTime          *time.Time // 实际支付时间，关联接收时间，（有值条件：审批通过+接收成功才有值）
	UpdatedAt            *time.Time
}

// 团队对团队转账结构体（完全匹配数据库表结构）
type TeaTeamToTeamTransferOut struct {
	Id           int
	Uuid         string
	FromTeamId   int    // 转出团队ID
	FromTeamName string // 转出团队名称
	ToTeamId     int    // 接收团队ID
	ToTeamName   string // 接收团队名称

	InitiatorUserId  int    // 发起转账的用户id（团队成员）
	AmountMilligrams int64  // 转账星茶数量(毫克），即锁定数量
	Notes            string // 转账备注,默认值:'-'

	// 审批相关（团队转出时使用）
	IsOnlyOneMemberTeam bool // 默认值false:多人团队审批(必填)，true:单人团队自动批准
	// 审批人填写，必填
	IsApproved              bool      // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          int       // 审批人ID，团队核心成员id，多人团队不能是发起人id，（必填）单人团队自动批准时，审批人是发起人自己
	ApprovalRejectionReason string    // 审批意见，如果拒绝，填写原因,默认值:'-'
	ApprovedAt              time.Time // 审批时间

	// 注意流程：审批通过，对方确认接收后，才会创建待接收记录（TeaTeamFromTeamTransferIn）
	Status               string // 包含审批状态，待审批，已批准，已拒绝，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer int64  // 转账后余额(毫克）
	CreatedAt            time.Time
	ExpiresAt            time.Time  // 转账请求过期时间，同时是解除锁定额度时间
	PaymentTime          *time.Time // 实际支付时间，关联接收时间，（有值条件：审批通过+接收成功才有值）
	UpdatedAt            *time.Time
}

// 团队接收用户转入记录结构体（完全匹配数据库表结构）
type TeaTeamFromUserTransferIn struct {
	Id                      int
	Uuid                    string
	UserToTeamTransferOutId int    // 引用-用户对团队转出记录id
	ToTeamId                int    // 接收团队ID
	ToTeamName              string // 接收团队名称
	FromUserId              int    // 转出用户ID
	FromUserName            string // 转出用户名称
	AmountMilligrams        int64  // 转账星茶数量(毫克）
	Notes                   string // 转账备注,默认值:'-'
	Status                  string // 包含接收状态，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer    int64  // 转账后余额(毫克）

	// 接收方ToTeam成员操作，Confirmed/Rejected二选一
	IsConfirmed       bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId int    // 操作用户id，确认接收或者拒绝接收的用户id（团队成员）
	RejectionReason   string // 如果拒绝，填写原因,默认值:'-'

	ExpiresAt time.Time // 过期时间，接收截止时间，也是FromUser解锁额度时间
	CreatedAt time.Time // 必填，如果接收，是接收、清算时间；如果拒绝，是拒绝时间
}

// 团队接收团队转入记录结构体（完全匹配数据库表结构）
type TeaTeamFromTeamTransferIn struct {
	Id                      int
	Uuid                    string
	TeamToTeamTransferOutId int    // 引用-团队对团队转出记录id
	ToTeamId                int    // 接收团队ID
	ToTeamName              string // 接收团队名称
	FromTeamId              int    // 转出团队ID
	FromTeamName            string // 转出团队名称
	AmountMilligrams        int64  // 转账星茶数量(毫克）
	Notes                   string // 转账备注,默认值:'-'
	Status                  string // 包含接收状态，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer    int64  // 转账后余额(毫克）

	// 接收方ToTeam成员操作，Confirmed(Completed)/Rejected二选一
	IsConfirmed       bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId int    // 操作用户id，确认接收或者拒绝接收的用户id（团队成员）
	RejectionReason   string // 如果拒绝，填写原因,默认值:'-'

	ExpiresAt time.Time // 过期时间，接收截止时间，也是FromTeam解锁额度时间
	CreatedAt time.Time // 必填，如果接收，是接收、清算时间；如果拒绝，是拒绝时间
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

// GetTeaTeamAccountByTeamId 根据团队ID获取星茶账户
func GetTeaTeamAccountByTeamId(teamId int) (TeaTeamAccount, error) {
	// 自由人团队没有星茶资产，返回特殊的冻结账户
	if teamId == TeamIdFreelancer {
		reason := "自由人团队不支持星茶资产"
		account := TeaTeamAccount{
			TeamId:            TeamIdFreelancer,
			BalanceMilligrams: 0,
			Status:            TeaTeamAccountStatus_Frozen,
			FrozenReason:      reason,
		}
		return account, nil
	}

	account := TeaTeamAccount{}
	err := DB.QueryRow("SELECT id, uuid, team_id, balance_milligrams, locked_balance_milligrams, status, frozen_reason, created_at, updated_at FROM tea.team_accounts WHERE team_id = $1", teamId).
		Scan(&account.Id, &account.Uuid, &account.TeamId, &account.BalanceMilligrams, &account.LockedBalanceMilligrams, &account.Status, &account.FrozenReason, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return account, fmt.Errorf("团队星茶账户不存在")
		}
		return account, fmt.Errorf("查询团队星茶账户失败: %v", err)
	}
	return account, nil
}

// Create 创建团队星茶账户
func (account *TeaTeamAccount) Create() error {
	statement := "INSERT INTO tea.team_accounts (team_id, balance_milligrams, status) VALUES ($1, $2, $3) RETURNING id, uuid"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(account.TeamId, account.BalanceMilligrams, account.Status).Scan(&account.Id, &account.Uuid)
	if err != nil {
		return fmt.Errorf("创建团队星茶账户失败: %v", err)
	}
	return nil
}

// UpdateStatus 更新团队星茶账户状态
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
		account.FrozenReason = reason
	}
	return nil
}

// EnsureTeaTeamAccountExists 确保团队有星茶账户
func EnsureTeaTeamAccountExists(teamId int) error {
	// 自由人团队不应该有星茶资产
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
			TeamId:            teamId,
			BalanceMilligrams: 0,
			Status:            TeaTeamAccountStatus_Normal,
		}
		return account.Create()
	}

	return nil
}

// CheckTeaTeamAccountFrozen 检查团队账户是否被冻结
func CheckTeaTeamAccountFrozen(teamId int) (bool, string, error) {
	// 自由人团队没有星茶资产，视为冻结状态
	if teamId == TeamIdFreelancer {
		return true, "自由人团队不支持星茶资产", nil
	}

	var status string
	frozenReason := "-" // 默认值
	err := DB.QueryRow("SELECT status, frozen_reason FROM tea.team_accounts WHERE team_id = $1", teamId).
		Scan(&status, &frozenReason)
	if err != nil {
		return false, "", fmt.Errorf("查询团队账户状态失败: %v", err)
	}

	if status == TeaTeamAccountStatus_Frozen {
		return true, frozenReason, nil
	}

	return false, frozenReason, nil
}

// processExpiredTransfers 处理过期转账的通用函数
func processExpiredTransfers(tableName string) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`SELECT id, from_team_id, amount_milligrams FROM %s WHERE status = $1 AND expires_at < $2`, tableName)
	rows, err := tx.Query(query, TeaTransferStatusPendingReceipt, time.Now())
	if err != nil {
		return fmt.Errorf("查询过期转账失败: %v", err)
	}
	defer rows.Close()

	var expiredTransfers []struct {
		Id         int
		FromTeamId int
		Amount     int64
	}

	for rows.Next() {
		var et struct {
			Id         int
			FromTeamId int
			Amount     int64
		}
		if err := rows.Scan(&et.Id, &et.FromTeamId, &et.Amount); err != nil {
			return fmt.Errorf("扫描过期转账失败: %v", err)
		}
		expiredTransfers = append(expiredTransfers, et)
	}

	if len(expiredTransfers) == 0 {
		return nil
	}

	for _, et := range expiredTransfers {
		var currentLockedBalance int64
		err = tx.QueryRow("SELECT locked_balance_milligrams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", et.FromTeamId).Scan(&currentLockedBalance)
		if err != nil {
			return fmt.Errorf("查询锁定余额失败: %v", err)
		}

		if currentLockedBalance < et.Amount {
			util.Warning("团队锁定余额不足，无法处理过期转账", et.FromTeamId, et.Id, currentLockedBalance, et.Amount)
			continue
		}

		updateQuery := fmt.Sprintf("UPDATE %s SET status = $1, updated_at = $2 WHERE id = $3", tableName)
		_, err = tx.Exec(updateQuery, TeaTransferStatusExpired, time.Now(), et.Id)
		if err != nil {
			return fmt.Errorf("更新过期转账状态失败: %v", err)
		}

		newLockedBalance := currentLockedBalance - et.Amount
		_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_milligrams = $1, updated_at = $2 WHERE team_id = $3",
			newLockedBalance, time.Now(), et.FromTeamId)
		if err != nil {
			return fmt.Errorf("解锁过期转账金额失败: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	return nil
}

// ProcessTeamToUserExpiredTransfers 处理过期的团队对用户转账，解锁相应的锁定金额
func ProcessTeamToUserExpiredTransfers() error {
	return processExpiredTransfers("tea.team_to_user_transfer_out")
}

// ProcessTeamToTeamExpiredTransfers 处理过期的团队对团队转账，解锁相应的锁定金额
func ProcessTeamToTeamExpiredTransfers() error {
	return processExpiredTransfers("tea.team_to_team_transfer_out")
}

// GetTeaTeamToTeamCompletedTransferOuts 获取团队对团队已经完成状态交易记录
func GetTeaTeamToTeamCompletedTransferOuts(team_id, page, limit int) ([]TeaTeamToTeamTransferOut, error) {
	if team_id == TeamIdNone {
		// 错误参数
		return nil, fmt.Errorf("团队ID不能为0")
	}
	if team_id == TeamIdFreelancer {
		return nil, fmt.Errorf("自由人团队没有星茶帐户")
	}
	rows, err := DB.Query(`
		SELECT id, uuid, from_team_id, from_team_name, to_team_id, to_team_name,
		       initiator_user_id, amount_milligrams, notes, is_only_one_member_team,
		       is_approved, approver_user_id, approval_rejection_reason, approved_at,
		       status, balance_after_transfer, created_at, expires_at, payment_time, updated_at
		FROM tea.team_to_team_transfer_out
		WHERE from_team_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, team_id, TeaTransferStatusCompleted, limit, (page-1)*limit)
	if err != nil {
		// amazonq-ignore-next-line
		return nil, fmt.Errorf("查询团队对团队已完成转出记录失败: %v", err)
	}
	defer rows.Close()

	transfers := []TeaTeamToTeamTransferOut{}
	for rows.Next() {
		var transfer TeaTeamToTeamTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.FromTeamName,
			&transfer.ToTeamId, &transfer.ToTeamName, &transfer.InitiatorUserId,
			&transfer.AmountMilligrams, &transfer.Notes, &transfer.IsOnlyOneMemberTeam,
			&transfer.IsApproved, &transfer.ApproverUserId, &transfer.ApprovalRejectionReason,
			&transfer.ApprovedAt, &transfer.Status, &transfer.BalanceAfterTransfer,
			&transfer.CreatedAt, &transfer.ExpiresAt, &transfer.PaymentTime, &transfer.UpdatedAt); err != nil {
			// amazonq-ignore-next-line
			return nil, fmt.Errorf("扫描团队对团队已完成转出记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	return transfers, nil
}

// CountPendingTeamReceipts 统计团队待接收的来自团队的转账数量
func CountPendingTeamReceipts(teamId int) (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) 
		FROM tea.team_from_team_transfer_in 
		WHERE to_team_id = $1 AND status = $2`, teamId, TeaTransferStatusPendingReceipt).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询团队待接收转账数量失败: %v", err)
	}
	return count, nil
}

// CountPendingUserReceipts 统计团队待接收的来自用户的转账数量
func CountPendingUserReceipts(teamId int) (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) 
		FROM tea.team_from_user_transfer_in 
		WHERE to_team_id = $1 AND status = $2`, teamId, TeaTransferStatusPendingReceipt).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询团队待接收用户转账数量失败: %v", err)
	}
	return count, nil
}

// CountPendingTeamApprovals 统计团队待审批的转出操作数量
func CountPendingTeamApprovals(teamId int) (int, error) {
	var count int
	// 统计团队对用户和团队对团队的待审批转账
	err := DB.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM tea.team_to_user_transfer_out WHERE from_team_id = $1 AND status = $2) +
			(SELECT COUNT(*) FROM tea.team_to_team_transfer_out WHERE from_team_id = $1 AND status = $2)`,
		teamId, TeaTransferStatusPendingApproval).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询团队待审批操作数量失败: %v", err)
	}
	return count, nil
}

// GetTeamTeaTransactions 获取团队星茶交易历史
func GetTeamTeaTransactions(teamId, page, limit int) ([]map[string]any, error) {
	if teamId == TeamIdFreelancer {
		return []map[string]any{}, nil
	}

	// 合并查询团队的所有转账记录
	query := `
		SELECT 'team_to_user' as type, id, amount_milligrams, status, created_at, notes
		FROM tea.team_to_user_transfer_out 
		WHERE from_team_id = $1
		UNION ALL
		SELECT 'team_to_team' as type, id, amount_milligrams, status, created_at, notes
		FROM tea.team_to_team_transfer_out 
		WHERE from_team_id = $1
		UNION ALL
		SELECT 'user_to_team' as type, id, amount_milligrams, status, created_at, notes
		FROM tea.team_from_user_transfer_in 
		WHERE to_team_id = $1
		UNION ALL
		SELECT 'team_from_team' as type, id, amount_milligrams, status, created_at, notes
		FROM tea.team_from_team_transfer_in 
		WHERE to_team_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := DB.Query(query, teamId, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询团队交易历史失败: %v", err)
	}
	defer rows.Close()

	transactions := []map[string]any{}
	for rows.Next() {
		var txType string
		var id int
		var amount int64
		var status string
		var createdAt time.Time
		var notes string

		if err := rows.Scan(&txType, &id, &amount, &status, &createdAt, &notes); err != nil {
			return nil, fmt.Errorf("扫描交易记录失败: %v", err)
		}

		transaction := map[string]any{
			"type":       txType,
			"id":         id,
			"amount":     amount,
			"status":     status,
			"created_at": createdAt,
			"notes":      notes,
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetTeamTransferOutOperations 获取团队转出操作记录
func GetTeamTransferOutOperations(teamId, page, limit int) ([]map[string]any, error) {
	if teamId == TeamIdFreelancer {
		return []map[string]any{}, nil
	}

	query := `
		SELECT 'team_to_user' as type, uuid, to_user_id as target_id, '' as target_name,
		       amount_milligrams, notes, status, created_at, expires_at
		FROM tea.team_to_user_transfer_out 
		WHERE from_team_id = $1
		UNION ALL
		SELECT 'team_to_team' as type, uuid, to_team_id as target_id, to_team_name as target_name,
		       amount_milligrams, notes, status, created_at, expires_at
		FROM tea.team_to_team_transfer_out 
		WHERE from_team_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := DB.Query(query, teamId, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询团队转出操作失败: %v", err)
	}
	defer rows.Close()

	operations := []map[string]any{}
	for rows.Next() {
		var opType, uuid, targetName, notes, status string
		var targetId int
		var amount int64
		var createdAt, expiresAt time.Time

		if err := rows.Scan(&opType, &uuid, &targetId, &targetName, &amount, &notes, &status, &createdAt, &expiresAt); err != nil {
			return nil, fmt.Errorf("扫描转出操作记录失败: %v", err)
		}

		operation := map[string]any{
			"type":        opType,
			"uuid":        uuid,
			"target_id":   targetId,
			"target_name": targetName,
			"amount":      amount,
			"notes":       notes,
			"status":      status,
			"created_at":  createdAt,
			"expires_at":  expiresAt,
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// GetTeamTransferInOperations 获取团队转入操作记录
func GetTeamTransferInOperations(teamId, page, limit int) ([]map[string]any, error) {
	if teamId == TeamIdFreelancer {
		return []map[string]any{}, nil
	}

	query := `
		SELECT 'user_to_team' as type, uuid, from_user_id as source_id, from_user_name as source_name,
		       amount_milligrams, notes, status, created_at, expires_at
		FROM tea.team_from_user_transfer_in 
		WHERE to_team_id = $1
		UNION ALL
		SELECT 'team_to_team' as type, uuid, from_team_id as source_id, from_team_name as source_name,
		       amount_milligrams, notes, status, created_at, expires_at
		FROM tea.team_from_team_transfer_in 
		WHERE to_team_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := DB.Query(query, teamId, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询团队转入操作失败: %v", err)
	}
	defer rows.Close()

	operations := []map[string]any{}
	for rows.Next() {
		var opType, uuid, sourceName, notes, status string
		var sourceId int
		var amount int64
		var createdAt, expiresAt time.Time

		if err := rows.Scan(&opType, &uuid, &sourceId, &sourceName, &amount, &notes, &status, &createdAt, &expiresAt); err != nil {
			return nil, fmt.Errorf("扫描转入操作记录失败: %v", err)
		}

		operation := map[string]any{
			"type":        opType,
			"uuid":        uuid,
			"source_id":   sourceId,
			"source_name": sourceName,
			"amount":      amount,
			"notes":       notes,
			"status":      status,
			"created_at":  createdAt,
			"expires_at":  expiresAt,
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// GetPendingTeamIncomingTransfers 获取团队待确认转入转账
func GetPendingTeamIncomingTransfers(teamId, page, limit int) ([]map[string]any, error) {
	if teamId == TeamIdFreelancer {
		return []map[string]any{}, nil
	}

	query := `
		SELECT 'user_to_team' as transfer_type, uuid, from_user_id as from_id, from_user_name as from_name,
		       amount_milligrams, notes, created_at, expires_at
		FROM tea.team_from_user_transfer_in 
		WHERE to_team_id = $1 AND status = $2
		UNION ALL
		SELECT 'team_to_team' as transfer_type, uuid, from_team_id as from_id, from_team_name as from_name,
		       amount_milligrams, notes, created_at, expires_at
		FROM tea.team_from_team_transfer_in 
		WHERE to_team_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := DB.Query(query, teamId, TeaTransferStatusPendingReceipt, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询团队待确认转入转账失败: %v", err)
	}
	defer rows.Close()

	transfers := []map[string]any{}
	for rows.Next() {
		var transferType, uuid, fromName, notes string
		var fromId int
		var amount int64
		var createdAt, expiresAt time.Time

		if err := rows.Scan(&transferType, &uuid, &fromId, &fromName, &amount, &notes, &createdAt, &expiresAt); err != nil {
			return nil, fmt.Errorf("扫描待确认转入转账失败: %v", err)
		}

		transfer := map[string]any{
			"transfer_type":     transferType,
			"uuid":              uuid,
			"from_id":           fromId,
			"from_name":         fromName,
			"amount_milligrams": amount,
			"notes":             notes,
			"created_at":        createdAt,
			"expires_at":        expiresAt,
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// GetPendingTeamToUserOperations 获取团队对用户待确认操作
func GetPendingTeamToUserOperations(teamId, page, limit int) ([]map[string]any, error) {
	rows, err := DB.Query(`
		SELECT uuid, to_user_id, to_user_name, amount_milligrams, notes, status, created_at, expires_at
		FROM tea.team_to_user_transfer_out 
		WHERE from_team_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, teamId, TeaTransferStatusPendingReceipt, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询团队对用户待确认操作失败: %v", err)
	}
	defer rows.Close()

	operations := []map[string]any{}
	for rows.Next() {
		var uuid, toUserName, notes, status string
		var toUserId int
		var amount int64
		var createdAt, expiresAt time.Time

		if err := rows.Scan(&uuid, &toUserId, &toUserName, &amount, &notes, &status, &createdAt, &expiresAt); err != nil {
			return nil, fmt.Errorf("扫描团队对用户待确认操作失败: %v", err)
		}

		operation := map[string]any{
			"uuid":              uuid,
			"to_user_id":        toUserId,
			"to_user_name":      toUserName,
			"amount_milligrams": amount,
			"notes":             notes,
			"status":            status,
			"created_at":        createdAt,
			"expires_at":        expiresAt,
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// GetPendingTeamToTeamOperations 获取团队对团队待确认操作
func GetPendingTeamToTeamOperations(teamId, page, limit int) ([]map[string]any, error) {
	rows, err := DB.Query(`
		SELECT uuid, to_team_id, to_team_name, amount_milligrams, notes, status, created_at, expires_at
		FROM tea.team_to_team_transfer_out 
		WHERE from_team_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, teamId, TeaTransferStatusPendingReceipt, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询团队对团队待确认操作失败: %v", err)
	}
	defer rows.Close()

	operations := []map[string]any{}
	for rows.Next() {
		var uuid, toTeamName, notes, status string
		var toTeamId int
		var amount int64
		var createdAt, expiresAt time.Time

		if err := rows.Scan(&uuid, &toTeamId, &toTeamName, &amount, &notes, &status, &createdAt, &expiresAt); err != nil {
			return nil, fmt.Errorf("扫描团队对团队待确认操作失败: %v", err)
		}

		operation := map[string]any{
			"uuid":              uuid,
			"to_team_id":        toTeamId,
			"to_team_name":      toTeamName,
			"amount_milligrams": amount,
			"notes":             notes,
			"status":            status,
			"created_at":        createdAt,
			"expires_at":        expiresAt,
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// CreateTeaTeamToUserTransferOut 创建团队对用户转账
func CreateTeaTeamToUserTransferOut(fromTeamId, initiatorUserId, toUserId int, amountMilligrams int64, notes string, expireHours int) (TeaTeamToUserTransferOut, error) {
	var transfer TeaTeamToUserTransferOut

	if fromTeamId == TeamIdFreelancer {
		return transfer, fmt.Errorf("自由人团队不支持星茶转账")
	}

	if amountMilligrams <= 0 {
		return transfer, fmt.Errorf("转账金额必须大于0")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return transfer, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 1. 锁定并检查转出团队账户余额
	var balance int64
	var status string
	err = tx.QueryRow(`
		SELECT balance_milligrams, status 
		FROM tea.team_accounts 
		WHERE team_id = $1 
		FOR UPDATE`,
		fromTeamId).Scan(&balance, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return transfer, fmt.Errorf("转出团队星茶账户不存在")
		}
		return transfer, fmt.Errorf("查询转出团队账户失败: %v", err)
	}

	if status == TeaTeamAccountStatus_Frozen {
		return transfer, fmt.Errorf("团队账户已冻结")
	}

	// 2. 检查余额是否足够（团队转账直接扣减余额，不锁定）
	if balance < amountMilligrams {
		return transfer, fmt.Errorf("团队星茶余额不足，当前余额: %d 毫克，需要: %d 毫克", balance, amountMilligrams)
	}

	// 3. 直接扣减团队账户余额（团队转账进入待审批状态，直接扣减）
	_, err = tx.Exec(`
		UPDATE tea.team_accounts 
		SET balance_milligrams = balance_milligrams - $1,
		    updated_at = $2
		WHERE team_id = $3 
		AND balance_milligrams >= $1`,
		amountMilligrams, time.Now(), fromTeamId)
	if err != nil {
		return transfer, fmt.Errorf("扣减团队账户余额失败: %v", err)
	}

	// 4. 创建转账记录
	expiresAt := time.Now().Add(time.Duration(expireHours) * time.Hour)
	err = tx.QueryRow(`
		INSERT INTO tea.team_to_user_transfer_out 
		(from_team_id, to_user_id, initiator_user_id, amount_milligrams, notes, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, uuid, created_at`,
		fromTeamId, toUserId, initiatorUserId, amountMilligrams, notes,
		TeaTransferStatusPendingApproval, expiresAt).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.CreatedAt)
	if err != nil {
		return transfer, fmt.Errorf("创建团队对用户转账失败: %v", err)
	}

	transfer.FromTeamId = fromTeamId
	transfer.ToUserId = toUserId
	transfer.InitiatorUserId = initiatorUserId
	transfer.AmountMilligrams = amountMilligrams
	transfer.Notes = notes
	transfer.Status = TeaTransferStatusPendingApproval
	transfer.ExpiresAt = expiresAt

	// 5. 提交事务
	if err = tx.Commit(); err != nil {
		return transfer, fmt.Errorf("提交事务失败: %v", err)
	}

	return transfer, nil
}

// CreateTeaTeamToTeamTransferOut 创建团队对团队转账
func CreateTeaTeamToTeamTransferOut(fromTeamId, initiatorUserId, toTeamId int, amountMilligrams int64, notes string, expireHours int) (TeaTeamToTeamTransferOut, error) {
	var transfer TeaTeamToTeamTransferOut

	if fromTeamId == TeamIdFreelancer || toTeamId == TeamIdFreelancer {
		return transfer, fmt.Errorf("自由人团队不支持星茶转账")
	}

	if amountMilligrams <= 0 {
		return transfer, fmt.Errorf("转账金额必须大于0")
	}

	if fromTeamId == toTeamId {
		return transfer, fmt.Errorf("不能向自己的团队转账")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return transfer, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 1. 锁定并检查转出团队账户余额
	var balance int64
	var status string
	err = tx.QueryRow(`
		SELECT balance_milligrams, status 
		FROM tea.team_accounts 
		WHERE team_id = $1 
		FOR UPDATE`,
		fromTeamId).Scan(&balance, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return transfer, fmt.Errorf("转出团队星茶账户不存在")
		}
		return transfer, fmt.Errorf("查询转出团队账户失败: %v", err)
	}

	if status == TeaTeamAccountStatus_Frozen {
		return transfer, fmt.Errorf("团队账户已冻结")
	}

	// 2. 检查余额是否足够（团队转账直接扣减余额，不锁定）
	if balance < amountMilligrams {
		return transfer, fmt.Errorf("团队星茶余额不足，当前余额: %d 毫克，需要: %d 毫克", balance, amountMilligrams)
	}

	// 3. 直接扣减团队账户余额（团队转账进入待审批状态，直接扣减）
	_, err = tx.Exec(`
		UPDATE tea.team_accounts 
		SET balance_milligrams = balance_milligrams - $1,
		    updated_at = $2
		WHERE team_id = $3 
		AND balance_milligrams >= $1`,
		amountMilligrams, time.Now(), fromTeamId)
	if err != nil {
		return transfer, fmt.Errorf("扣减团队账户余额失败: %v", err)
	}

	// 4. 创建转账记录
	expiresAt := time.Now().Add(time.Duration(expireHours) * time.Hour)
	err = tx.QueryRow(`
		INSERT INTO tea.team_to_team_transfer_out 
		(from_team_id, to_team_id, initiator_user_id, amount_milligrams, notes, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, uuid, created_at`,
		fromTeamId, toTeamId, initiatorUserId, amountMilligrams, notes,
		TeaTransferStatusPendingApproval, expiresAt).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.CreatedAt)
	if err != nil {
		return transfer, fmt.Errorf("创建团队对团队转账失败: %v", err)
	}

	transfer.FromTeamId = fromTeamId
	transfer.ToTeamId = toTeamId
	transfer.InitiatorUserId = initiatorUserId
	transfer.AmountMilligrams = amountMilligrams
	transfer.Notes = notes
	transfer.Status = TeaTransferStatusPendingApproval
	transfer.ExpiresAt = expiresAt

	// 5. 提交事务
	if err = tx.Commit(); err != nil {
		return transfer, fmt.Errorf("提交事务失败: %v", err)
	}

	return transfer, nil
}

// TeaConfirmUserToTeamTransferOut 团队(某个成员)确认接收来自用户转账
func TeaConfirmUserToTeamTransferOut(transferUuid string, toTeamId, operationalUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var fromUserId int
	var notes, toTeamName, fromUserName string
	err = tx.QueryRow(`
		SELECT id, from_user_id, amount_milligrams, notes, to_team_name, from_user_name
		FROM tea.user_to_team_transfer_out 
		WHERE uuid = $1 AND to_team_id = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		transferUuid, toTeamId, TeaTransferStatusPendingReceipt,
	).Scan(&transferOutID, &fromUserId, &amountMg, &notes, &toTeamName, &fromUserName)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到待确认的转账记录")
		}
		return fmt.Errorf("查询转账记录失败: %v", err)
	}

	if amountMg <= 0 {
		return fmt.Errorf("转账金额无效: %d", amountMg)
	}

	now := time.Now()

	// 1. 更新转账状态
	_, err = tx.Exec(`
		UPDATE tea.user_to_team_transfer_out
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4`,
		TeaTransferStatusCompleted, now, transferOutID, TeaTransferStatusPendingReceipt)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 2. 更新转出用户账户（减少余额和锁定金额）
	_, err = tx.Exec(`
		UPDATE tea.user_accounts 
		SET balance_milligrams = balance_milligrams - $1,
			locked_balance_milligrams = locked_balance_milligrams - $1,
			updated_at = $2
		WHERE user_id = $3 
		AND locked_balance_milligrams >= $1
		RETURNING id`,
		amountMg, now, fromUserId)
	if err != nil {
		return fmt.Errorf("更新转出用户账户失败: %v", err)
	}

	// 3. 更新接收团队账户（增加余额）
	var receiverBalanceAfter int64
	err = tx.QueryRow(`
		UPDATE tea.team_accounts 
		SET balance_milligrams = balance_milligrams + $1,
			updated_at = $2
		WHERE team_id = $3
		RETURNING balance_milligrams`,
		amountMg, now, toTeamId).Scan(&receiverBalanceAfter)
	if err != nil {
		return fmt.Errorf("更新接收团队账户失败: %v", err)
	}

	// 4. 创建接收记录
	_, err = tx.Exec(`
		INSERT INTO tea.team_from_user_transfer_in (
			user_to_team_transfer_out_id, to_team_id, to_team_name, 
			from_user_id, from_user_name, amount_milligrams, notes, 
			balance_after_transfer, status, is_confirmed, operational_user_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		transferOutID, toTeamId, toTeamName, fromUserId, fromUserName,
		amountMg, notes, receiverBalanceAfter, TeaTransferStatusCompleted,
		true, operationalUserId, now)
	if err != nil {
		return fmt.Errorf("创建接收来自用户转账记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	return nil
}

// TeaConfirmTeamToTeamTransferOut 团队(某个成员)确认接收来自团队转账
func TeaConfirmTeamToTeamTransferOut(transferUuid string, toTeamId, operationalUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var fromTeamId int
	var notes, toTeamName, fromTeamName string
	err = tx.QueryRow(`
		SELECT id, from_team_id, amount_milligrams, notes, to_team_name, from_team_name
		FROM tea.team_to_team_transfer_out 
		WHERE uuid = $1 AND to_team_id = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		transferUuid, toTeamId, TeaTransferStatusPendingReceipt,
	).Scan(&transferOutID, &fromTeamId, &amountMg, &notes, &toTeamName, &fromTeamName)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到待确认的转账记录")
		}
		return fmt.Errorf("查询转账记录失败: %v", err)
	}

	if amountMg <= 0 {
		return fmt.Errorf("转账金额无效: %d", amountMg)
	}

	now := time.Now()

	// 1. 更新转账状态
	_, err = tx.Exec(`
		UPDATE tea.team_to_team_transfer_out
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4`,
		TeaTransferStatusCompleted, now, transferOutID, TeaTransferStatusPendingReceipt)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 2. 更新转出团队账户（减少余额和锁定金额）
	_, err = tx.Exec(`
		UPDATE tea.team_accounts 
		SET balance_milligrams = balance_milligrams - $1,
			locked_balance_milligrams = locked_balance_milligrams - $1,
			updated_at = $2
		WHERE team_id = $3 
		AND locked_balance_milligrams >= $1
		RETURNING id`,
		amountMg, now, fromTeamId)
	if err != nil {
		return fmt.Errorf("更新转出团队账户失败: %v", err)
	}

	// 3. 更新接收团队账户（增加余额）
	var receiverBalanceAfter int64
	err = tx.QueryRow(`
		UPDATE tea.team_accounts 
		SET balance_milligrams = balance_milligrams + $1,
			updated_at = $2
		WHERE team_id = $3
		RETURNING balance_milligrams`,
		amountMg, now, toTeamId).Scan(&receiverBalanceAfter)
	if err != nil {
		return fmt.Errorf("更新接收团队账户失败: %v", err)
	}

	// 4. 创建接收记录
	_, err = tx.Exec(`
		INSERT INTO tea.team_from_team_transfer_in (
			team_to_team_transfer_out_id, to_team_id, to_team_name, 
			from_team_id, from_team_name, amount_milligrams, notes, 
			balance_after_transfer, status, is_confirmed, operational_user_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		transferOutID, toTeamId, toTeamName, fromTeamId, fromTeamName,
		amountMg, notes, receiverBalanceAfter, TeaTransferStatusCompleted,
		true, operationalUserId, now)
	if err != nil {
		return fmt.Errorf("创建接收来自团队转账记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	return nil
}

// TeaTeamRejectFromUserTransferIn 团队(某个成员)拒绝接收来自用户转账
func TeaTeamRejectFromUserTransferIn(transferUuid string, toTeamId, operatorUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var fromUserId int
	var toTeamName, fromUserName, notes string
	err = tx.QueryRow(`
		SELECT id, from_user_id, amount_milligrams, to_team_name, from_user_name, notes
		FROM tea.user_to_team_transfer_out 
		WHERE uuid = $1 AND to_team_id = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		transferUuid, toTeamId, TeaTransferStatusPendingReceipt,
	).Scan(&transferOutID, &fromUserId, &amountMg, &toTeamName, &fromUserName, &notes)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到待拒绝的转账记录")
		}
		return fmt.Errorf("查询转账记录失败: %v", err)
	}

	if amountMg <= 0 {
		return fmt.Errorf("转账金额无效: %d", amountMg)
	}

	now := time.Now()

	// 1. 更新转账状态为已拒绝
	_, err = tx.Exec(`
		UPDATE tea.user_to_team_transfer_out
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4`,
		TeaTransferStatusRejected, now, transferOutID, TeaTransferStatusPendingReceipt)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 2. 解锁转出用户账户的锁定金额
	_, err = tx.Exec(`
		UPDATE tea.user_accounts 
		SET locked_balance_milligrams = locked_balance_milligrams - $1,
			updated_at = $2
		WHERE user_id = $3 
		AND locked_balance_milligrams >= $1`,
		amountMg, now, fromUserId)
	if err != nil {
		return fmt.Errorf("解锁转出用户账户失败: %v", err)
	}

	// 3. 创建拒绝接收记录
	_, err = tx.Exec(`
		INSERT INTO tea.team_from_user_transfer_in (
			user_to_team_transfer_out_id, to_team_id, to_team_name,
			from_user_id, from_user_name, amount_milligrams, notes,
			status, operational_user_id, rejection_reason, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		transferOutID, toTeamId, toTeamName, fromUserId,
		fromUserName, amountMg,
		notes, TeaTransferStatusRejected,
		operatorUserId,
		reason,
		now)
	if err != nil {
		return fmt.Errorf("创建拒绝接收记录失败: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	return nil
}

// TeaTeamRejectFromTeamTransferIn 团队(某个成员),拒绝接收,来自团队转账
func TeaTeamRejectFromTeamTransferIn(transferUuid string, toTeamId, operatorUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var fromTeamId int
	var toTeamName, fromTeamName, notes string
	err = tx.QueryRow(`
		SELECT id, from_team_id, amount_milligrams, to_team_name, from_team_name, notes
		FROM tea.team_to_team_transfer_out 
		WHERE uuid = $1 AND to_team_id = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		transferUuid, toTeamId, TeaTransferStatusPendingReceipt,
	).Scan(&transferOutID, &fromTeamId, &amountMg, &toTeamName, &fromTeamName, &notes)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到待拒绝的转账记录")
		}
		return fmt.Errorf("查询转账记录失败: %v", err)
	}

	if amountMg <= 0 {
		return fmt.Errorf("转账金额无效: %d", amountMg)
	}

	now := time.Now()

	// 1. 更新转账状态为已拒绝
	_, err = tx.Exec(`
		UPDATE tea.team_to_team_transfer_out
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4`,
		TeaTransferStatusRejected, now, transferOutID, TeaTransferStatusPendingReceipt)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 2. 解锁转出团队账户的锁定金额
	_, err = tx.Exec(`
		UPDATE tea.team_accounts 
		SET locked_balance_milligrams = locked_balance_milligrams - $1,
			updated_at = $2
		WHERE team_id = $3 
		AND locked_balance_milligrams >= $1`,
		amountMg, now, fromTeamId)
	if err != nil {
		return fmt.Errorf("解锁转出团队账户失败: %v", err)
	}

	// 3. 创建拒绝接收记录
	_, err = tx.Exec(`
		INSERT INTO tea.team_from_team_transfer_in (
			team_to_team_transfer_out_id, to_team_id, to_team_name,
			from_team_id, from_team_name, amount_milligrams, notes,
			status, operational_user_id, rejection_reason, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		transferOutID, toTeamId, toTeamName, fromTeamId,
		fromTeamName, amountMg,
		notes, TeaTransferStatusRejected,
		operatorUserId,
		reason,
		now)
	if err != nil {
		return fmt.Errorf("创建拒绝接收记录失败: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// TeaTeamApproveToUserTransferOut 某个团队核心成员,审批通过,团队对用户转账
func TeaTeamApproveToUserTransferOut(fromTeamId int, transferUuid string, approverUserId int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var toUserId int
	var toUserName, fromTeamName, notes string
	err = tx.QueryRow(`
		SELECT id, amount_milligrams, to_user_id, to_user_name, from_team_name, notes
		FROM tea.team_to_user_transfer_out 
		WHERE from_team_id = $1 AND uuid = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		fromTeamId, transferUuid, TeaTransferStatusPendingApproval,
	).Scan(&transferOutID, &amountMg, &toUserId, &toUserName, &fromTeamName, &notes)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到待审批的转账记录")
		}
		return fmt.Errorf("查询转账记录失败: %v", err)
	}

	if amountMg <= 0 {
		return fmt.Errorf("转账金额无效: %d", amountMg)
	}

	now := time.Now()

	// 1. 更新转账状态为已批准
	_, err = tx.Exec(`
		UPDATE tea.team_to_user_transfer_out 
		SET status = $1, approver_user_id = $2, approved_at = $3, updated_at = $4
		WHERE id = $5 AND status = $6`,
		TeaTransferStatusApproved, approverUserId, now, now, transferOutID, TeaTransferStatusPendingApproval)
	if err != nil {
		return fmt.Errorf("更新转账审批状态失败: %v", err)
	}

	// 2. 锁定转出团队账户的金额
	_, err = tx.Exec(`
		UPDATE tea.team_accounts 
		SET locked_balance_milligrams = locked_balance_milligrams + $1,
			updated_at = $2
		WHERE team_id = $3`,
		amountMg, now, fromTeamId)
	if err != nil {
		return fmt.Errorf("锁定转出团队账户金额失败: %v", err)
	}

	// 3. 创建接收记录（用户确认 接收/拒绝 操作后创建）
	// _, err = tx.Exec(`
	// 	INSERT INTO tea.user_from_team_transfer_in (
	// 		team_to_user_transfer_out_id, to_user_id, to_user_name,
	// 		from_team_id, from_team_name, amount_milligrams, notes,
	// 		status, is_confirmed, operational_user_id, rejection_reason, expires_at, created_at
	// 	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
	// 	transferOutID, toUserId, toUserName, fromTeamId,
	// 	fromTeamName, amountMg, notes, TeaTransferStatusPendingReceipt,
	// 	false, approverUserId, "-", now.Add(24*time.Hour), now)
	// if err != nil {
	// 	return fmt.Errorf("创建待接收记录失败: %v", err)
	// }

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// TeaTeamApproveToTeamTransferOut 某个团队核心成员,审批通过,团队对团队转账
func TeaTeamApproveToTeamTransferOut(fromTeamId int, transferUuid string, approverUserId int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var toTeamId int
	var toTeamName, fromTeamName, notes string
	err = tx.QueryRow(`
		SELECT id, amount_milligrams, to_team_id, to_team_name, from_team_name, notes
		FROM tea.team_to_team_transfer_out 
		WHERE from_team_id = $1 AND uuid = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		fromTeamId, transferUuid, TeaTransferStatusPendingApproval,
	).Scan(&transferOutID, &amountMg, &toTeamId, &toTeamName, &fromTeamName, &notes)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到待审批的转账记录")
		}
		return fmt.Errorf("查询转账记录失败: %v", err)
	}

	if amountMg <= 0 {
		return fmt.Errorf("转账金额无效: %d", amountMg)
	}

	now := time.Now()

	// 1. 更新转账状态为已批准
	_, err = tx.Exec(`
		UPDATE tea.team_to_team_transfer_out 
		SET status = $1, approver_user_id = $2, approved_at = $3, updated_at = $4
		WHERE id = $5 AND status = $6`,
		TeaTransferStatusApproved, approverUserId, now, now, transferOutID, TeaTransferStatusPendingApproval)
	if err != nil {
		return fmt.Errorf("更新转账审批状态失败: %v", err)
	}

	// 2. 锁定转出团队账户的金额
	_, err = tx.Exec(`
		UPDATE tea.team_accounts 
		SET locked_balance_milligrams = locked_balance_milligrams + $1,
			updated_at = $2
		WHERE team_id = $3`,
		amountMg, now, fromTeamId)
	if err != nil {
		return fmt.Errorf("锁定转出团队账户金额失败: %v", err)
	}

	// 3. 创建待接收记录（接收团队需要确认接收）
	// _, err = tx.Exec(`
	// 	INSERT INTO tea.team_from_team_transfer_in (
	// 		team_to_team_transfer_out_id, to_team_id, to_team_name,
	// 		from_team_id, from_team_name, amount_milligrams, notes,
	// 		status, is_confirmed, operational_user_id, rejection_reason, expires_at, created_at
	// 	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
	// 	transferOutID, toTeamId, toTeamName, fromTeamId,
	// 	fromTeamName, amountMg, notes, TeaTransferStatusPendingReceipt,
	// 	false, approverUserId, "-", now.Add(24*time.Hour), now)
	// if err != nil {
	// 	return fmt.Errorf("创建待接收记录失败: %v", err)
	// }

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// TeaTeamRejectToUserTransferOut 某个团队核心成员,拒绝审批,团队对用户转账
func TeaTeamRejectToUserTransferOut(fromTeamId int, transferUuid string, approverUserId int, reason string) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var toUserId int
	var toUserName, fromTeamName, notes string
	err = tx.QueryRow(`
		SELECT id, amount_milligrams, to_user_id, to_user_name, from_team_name, notes
		FROM tea.team_to_user_transfer_out 
		WHERE from_team_id = $1 AND uuid = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		fromTeamId, transferUuid, TeaTransferStatusPendingApproval,
	).Scan(&transferOutID, &amountMg, &toUserId, &toUserName, &fromTeamName, &notes)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到待拒绝的转账记录")
		}
		return fmt.Errorf("查询转账记录失败: %v", err)
	}

	if amountMg <= 0 {
		return fmt.Errorf("转账金额无效: %d", amountMg)
	}

	now := time.Now()

	// 1. 更新转账状态为已拒绝
	_, err = tx.Exec(`
		UPDATE tea.team_to_user_transfer_out 
		SET status = $1, approver_user_id = $2, approval_rejection_reason = $3, approved_at = $4, updated_at = $5
		WHERE id = $6 AND status = $7`,
		TeaTransferStatusApprovalRejected, approverUserId, reason, now, now, transferOutID, TeaTransferStatusPendingApproval)
	if err != nil {
		return fmt.Errorf("更新转账拒绝状态失败: %v", err)
	}

	// 2. 释放转出团队账户的锁定金额
	_, err = tx.Exec(`
		UPDATE tea.team_accounts 
		SET locked_balance_milligrams = locked_balance_milligrams - $1,
			updated_at = $2
		WHERE team_id = $3 
		AND locked_balance_milligrams >= $1`,
		amountMg, now, fromTeamId)
	if err != nil {
		return fmt.Errorf("释放转出团队账户锁定金额失败: %v", err)
	}

	// 3. 创建拒绝记录（用于历史追溯）
	// _, err = tx.Exec(`
	// 	INSERT INTO tea.user_from_team_transfer_in (
	// 		team_to_user_transfer_out_id, to_user_id, to_user_name,
	// 		from_team_id, from_team_name, amount_milligrams, notes,
	// 		status, is_confirmed, operational_user_id, rejection_reason, expires_at, created_at
	// 	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
	// 	transferOutID, toUserId, toUserName, fromTeamId,
	// 	fromTeamName, amountMg, notes, TeaTransferStatusRejected,
	// 	false, approverUserId, reason, now, now)
	// if err != nil {
	// 	return fmt.Errorf("创建拒绝记录失败: %v", err)
	// }

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// TeaTeamRejectToTeamTransferOut 某个团队核心成员,拒绝审批团队对团队转账
func TeaTeamRejectToTeamTransferOut(fromTeamId int, transferUuid string, approverUserId int, reason string) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var toTeamId int
	var toTeamName, fromTeamName, notes string
	err = tx.QueryRow(`
		SELECT id, amount_milligrams, to_team_id, to_team_name, from_team_name, notes
		FROM tea.team_to_team_transfer_out 
		WHERE from_team_id = $1 AND uuid = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		fromTeamId, transferUuid, TeaTransferStatusPendingApproval,
	).Scan(&transferOutID, &amountMg, &toTeamId, &toTeamName, &fromTeamName, &notes)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到待拒绝的转账记录")
		}
		return fmt.Errorf("查询转账记录失败: %v", err)
	}

	if amountMg <= 0 {
		return fmt.Errorf("转账金额无效: %d", amountMg)
	}

	now := time.Now()

	// 1. 更新转账状态为已拒绝
	_, err = tx.Exec(`
		UPDATE tea.team_to_team_transfer_out 
		SET status = $1, approver_user_id = $2, approval_rejection_reason = $3, approved_at = $4, updated_at = $5
		WHERE id = $6 AND status = $7`,
		TeaTransferStatusApprovalRejected, approverUserId, reason, now, now, transferOutID, TeaTransferStatusPendingApproval)
	if err != nil {
		return fmt.Errorf("更新转账拒绝状态失败: %v", err)
	}

	// 2. 释放转出团队账户的锁定金额
	_, err = tx.Exec(`
		UPDATE tea.team_accounts 
		SET locked_balance_milligrams = locked_balance_milligrams - $1,
			updated_at = $2
		WHERE team_id = $3 
		AND locked_balance_milligrams >= $1`,
		amountMg, now, fromTeamId)
	if err != nil {
		return fmt.Errorf("释放转出团队账户锁定金额失败: %v", err)
	}

	// 3. 创建拒绝记录（用于历史追溯）
	// _, err = tx.Exec(`
	// 	INSERT INTO tea.team_from_team_transfer_in (
	// 		team_to_team_transfer_out_id, to_team_id, to_team_name,
	// 		from_team_id, from_team_name, amount_milligrams, notes,
	// 		status, is_confirmed, operational_user_id, rejection_reason, expires_at, created_at
	// 	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
	// 	transferOutID, toTeamId, toTeamName, fromTeamId,
	// 	fromTeamName, amountMg, notes, TeaTransferStatusRejected,
	// 	false, approverUserId, reason, now, now)
	// if err != nil {
	// 	return fmt.Errorf("创建拒绝记录失败: %v", err)
	// }

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}
