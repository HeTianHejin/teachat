package dao

import (
	"database/sql"
	"fmt"
	"time"
)

/*
-- 星茶是一种特殊虚拟现实资源/物品，（毫克）作为重量单位
-- 用户获得星茶来自向茶庄团队购买（1元/克）
-- 如果需要兑现，就向茶庄出售星茶（1克/元）

注意：用户星茶账户相关定义（TeaUserToUserTransferOut、TeaUserFromTeamTransferIn……）已迁移至tea_account.go文件，保留此注释以提示开发者--

团队星茶账户转账流程：
1、发起方法：团队转出额度星茶，无论接收方是团队还是用户（个人），都要求由1成员填写转账表单，第一步创建待审核转账表单，
2、审核方法：由任意1核心成员审批待审核表单，
2.1、如果审核批准，转账表单状态更新为已批准，执行第3步锁定转账额度，
2.2、如果审核否决，转账表单状态更新为已否决，记录审批操作记录，流程结束。
3、锁定方法：已批准的转出团队账户转出额度星茶数量被锁定，防止重复发起转账；
4.1、接收方法：目标接受方，用户或者团队任意1状态正常成员TeamMemberStatusActive(1)，有效期内，操作接收，继续第5步结算双方账户，
4.2、拒收方法，目标接受方，用户或者团队任意1状态正常成员，有效期内，操作拒收，转账表单状态更新为已拒收，解锁被转出方锁定星茶，记录拒收操作用户id及时间原因，流程结束。
5、结算方法：按锁定额度（接收额度）清算双方账户数额，创建出入流水记录；
6、超时处理：自动解锁转出用户账户被锁定额度星茶，不创建交易流水明细记录。
*/

// // 团队星茶账户状态常量
const (
	TeaTeamAccountStatus_Normal  = "normal"
	TeaTeamAccountStatus_Frozen  = "frozen"
	TeaTeamAccountStatus_Deleted = "deleted"
)

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

	// 注意流程：审批通过后，才会创建待接收记录（TeaUserFromTeamTransferIn）
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
	IsConfirmed              bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId        int    // 操作用户id，确认接收或者拒绝接收的用户id（团队成员）
	ReceptionRejectionReason string // 如果拒绝，填写原因,默认值:'-'

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

	// 接收方ToTeam成员操作，Confirmed/Rejected二选一
	IsConfirmed              bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId        int    // 操作用户id，确认接收或者拒绝接收的用户id（团队成员）
	ReceptionRejectionReason string // 如果拒绝，填写原因,默认值:'-'

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

// CheckTeamAccountFrozen 检查团队账户是否被冻结
func CheckTeamAccountFrozen(teamId int) (bool, string, error) {
	// 自由人团队没有星茶资产，视为冻结状态
	if teamId == TeamIdFreelancer {
		return true, "自由人团队不支持星茶资产", nil
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

// ProcessTeamToUserExpiredTransfers 处理过期的团队对用户转账，解锁相应的锁定金额
func ProcessTeamToUserExpiredTransfers() error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 查找所有过期且仍为pending_receipt状态的团队对用户转账
	rows, err := tx.Query(`
		SELECT id, from_team_id, amount_milligrams 
		FROM tea.team_to_user_transfer_out 
		WHERE status = $1 AND expires_at < $2`,
		TeaTransferStatusPendingReceipt, time.Now())
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
		return nil // 没有过期转账需要处理
	}

	// 处理每个过期转账：更新状态并解锁金额
	for _, et := range expiredTransfers {
		// 获取当前锁定余额
		var currentLockedBalance int64
		err = tx.QueryRow("SELECT locked_balance_milligrams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", et.FromTeamId).Scan(&currentLockedBalance)
		if err != nil {
			return fmt.Errorf("查询锁定余额失败: %v", err)
		}

		// 检查锁定余额是否足够
		if currentLockedBalance < et.Amount {
			// 锁定余额不足，记录警告并跳过
			continue
		}

		// 更新转账状态为过期
		_, err = tx.Exec("UPDATE tea.team_to_user_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			TeaTransferStatusExpired, time.Now(), et.Id)
		if err != nil {
			return fmt.Errorf("更新过期转账状态失败: %v", err)
		}

		// 解锁相应的锁定金额
		newLockedBalance := currentLockedBalance - et.Amount
		_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_milligrams = $1, updated_at = $2 WHERE team_id = $3",
			newLockedBalance, time.Now(), et.FromTeamId)
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

// ProcessTeamToTeamExpiredTransfers 处理过期的团队对团队转账，解锁相应的锁定金额
func ProcessTeamToTeamExpiredTransfers() error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 查找所有过期且仍为pending_receipt状态的团队对团队转账
	rows, err := tx.Query(`
		SELECT id, from_team_id, amount_milligrams 
		FROM tea.team_to_team_transfer_out 
		WHERE status = $1 AND expires_at < $2`,
		TeaTransferStatusPendingReceipt, time.Now())
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
		return nil // 没有过期转账需要处理
	}

	// 处理每个过期转账：更新状态并解锁金额
	for _, et := range expiredTransfers {
		// 获取当前锁定余额
		var currentLockedBalance int64
		err = tx.QueryRow("SELECT locked_balance_milligrams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", et.FromTeamId).Scan(&currentLockedBalance)
		if err != nil {
			return fmt.Errorf("查询锁定余额失败: %v", err)
		}

		// 检查锁定余额是否足够
		if currentLockedBalance < et.Amount {
			// 锁定余额不足，记录警告并跳过
			continue
		}

		// 更新转账状态为过期
		_, err = tx.Exec("UPDATE tea.team_to_team_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			TeaTransferStatusExpired, time.Now(), et.Id)
		if err != nil {
			return fmt.Errorf("更新过期转账状态失败: %v", err)
		}

		// 解锁相应的锁定金额
		newLockedBalance := currentLockedBalance - et.Amount
		_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_milligrams = $1, updated_at = $2 WHERE team_id = $3",
			newLockedBalance, time.Now(), et.FromTeamId)
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
