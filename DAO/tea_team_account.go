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
4.1、接收方法：目标接受方，用户或者团队任意1状态正常成员TeamMemberStatusActive(1)，有效期内，操作接收，继续第5步结算双方账户，
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

// 团队对用户转账结构体（完全匹配数据库表结构）
type TeaTeamToUserTransferOut struct {
	Id           int
	Uuid         string
	FromTeamId   int    // 转出团队ID
	FromTeamName string // 转出团队名称
	ToUserId     int    // 接收用户ID
	ToUserName   string // 接收用户名称

	InitiatorUserId int     // 发起转账的用户id（团队成员）
	AmountGrams     float64 // 转账茶叶数量(克)，即锁定数量
	Notes           string  // 转账备注,默认值:'-'

	// 审批相关（团队转出时使用）
	IsOnlyOneMemberTeam bool // 默认值false:多人团队审批(必填)，true:单人团队自动批准
	// 审批人填写，必填
	IsApproved              bool      // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          int       // 审批人ID，团队核心成员id，多人团队不能是发起人id，（必填）单人团队自动批准时，审批人是发起人自己
	ApprovalRejectionReason string    // 审批意见，如果拒绝，填写原因,默认值:'-'
	ApprovedAt              time.Time // 审批时间

	// 注意流程：审批通过后，才会创建待接收记录（TeaUserFromTeamTransferIn）
	Status               string  // 包含审批状态，待审批，已批准，已拒绝，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer float64 // 转账后余额(克)
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

	InitiatorUserId int     // 发起转账的用户id（团队成员）
	AmountGrams     float64 // 转账茶叶数量(克)，即锁定数量
	Notes           string  // 转账备注,默认值:'-'

	// 审批相关（团队转出时使用）
	IsOnlyOneMemberTeam bool // 默认值false:多人团队审批(必填)，true:单人团队自动批准
	// 审批人填写，必填
	IsApproved              bool      // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          int       // 审批人ID，团队核心成员id，多人团队不能是发起人id，（必填）单人团队自动批准时，审批人是发起人自己
	ApprovalRejectionReason string    // 审批意见，如果拒绝，填写原因,默认值:'-'
	ApprovedAt              time.Time // 审批时间

	// 注意流程：审批通过后，才会创建待接收记录（TeaTeamFromTeamTransferIn）
	Status               string  // 包含审批状态，待审批，已批准，已拒绝，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer float64 // 转账后余额(克)
	CreatedAt            time.Time
	ExpiresAt            time.Time  // 转账请求过期时间，同时是解除锁定额度时间
	PaymentTime          *time.Time // 实际支付时间，关联接收时间，（有值条件：审批通过+接收成功才有值）
	UpdatedAt            *time.Time
}

// 团队接收用户转入记录结构体（完全匹配数据库表结构）
type TeaTeamFromUserTransferIn struct {
	Id                      int
	Uuid                    string
	UserToTeamTransferOutId int     // 引用-用户对团队转出记录id
	ToTeamId                int     // 接收团队ID
	ToTeamName              string  // 接收团队名称
	FromUserId              int     // 转出用户ID
	FromUserName            string  // 转出用户名称
	AmountGrams             float64 // 转账茶叶数量(克)
	Notes                   string  // 转账备注,默认值:'-'
	Status                  string  // 包含接收状态，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer    float64 // 转账后余额(克)

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
	TeamToTeamTransferOutId int     // 引用-团队对团队转出记录id
	ToTeamId                int     // 接收团队ID
	ToTeamName              string  // 接收团队名称
	FromTeamId              int     // 转出团队ID
	FromTeamName            string  // 转出团队名称
	AmountGrams             float64 // 转账茶叶数量(克)
	Notes                   string  // 转账备注,默认值:'-'
	Status                  string  // 包含接收状态，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer    float64 // 转账后余额(克)

	// 接收方ToTeam成员操作，Confirmed/Rejected二选一
	IsConfirmed              bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId        int    // 操作用户id，确认接收或者拒绝接收的用户id（团队成员）
	ReceptionRejectionReason string // 如果拒绝，填写原因,默认值:'-'

	ExpiresAt time.Time // 过期时间，接收截止时间，也是FromTeam解锁额度时间
	CreatedAt time.Time // 必填，如果接收，是接收、清算时间；如果拒绝，是拒绝时间
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

	// 查询团队作为转出方向用户的交易历史
	rows, err := DB.Query(`
		SELECT 'outgoing' as transaction_type,
			   'user' as target_type,
			   uuid, from_team_id, initiator_user_id, to_user_id, to_team_id,
			   amount_grams, status, notes, payment_time, created_at
		FROM tea.team_to_user_transfer_out
		WHERE from_team_id = $1 AND status = 'completed'

		UNION ALL

		-- 查询团队作为转出方向其他团队的交易历史
		SELECT 'outgoing' as transaction_type,
			   'team' as target_type,
			   uuid, from_team_id, initiator_user_id, to_user_id, to_team_id,
			   amount_grams, status, notes, payment_time, created_at
		FROM tea.team_to_team_transfer_out
		WHERE from_team_id = $1 AND status = 'completed'

		UNION ALL

		-- 查询团队作为接收方的交易历史（从用户转入）
		SELECT 'incoming' as transaction_type,
			   'user' as target_type,
			   uto.uuid, NULL as from_team_id, uto.from_user_id as initiator_user_id,
			   uto.to_user_id, uto.to_team_id,
			   uto.amount_grams, uto.status, uto.notes, uto.payment_time, uto.created_at
		FROM tea.user_to_team_transfer_out uto
		WHERE uto.to_team_id = $1 AND uto.status = 'completed'

		UNION ALL

		-- 查询团队作为接收方的交易历史（从其他团队转入）
		SELECT 'incoming' as transaction_type,
			   'team' as target_type,
			   tto.uuid, tto.from_team_id, tto.initiator_user_id,
			   NULL as to_user_id, tto.to_team_id,
			   tto.amount_grams, tto.status, tto.notes, tto.payment_time, tto.created_at
		FROM tea.team_to_team_transfer_out tto
		WHERE tto.to_team_id = $1 AND tto.status = 'completed'

		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		teamId, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("查询团队交易历史失败: %v", err)
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var transactionType, targetType, uuid, status, notes string
		var fromTeamId, initiatorUserId sql.NullInt64
		var toUserId, toTeamId sql.NullInt64
		var amountGrams float64
		var paymentTime, createdAt sql.NullTime

		err = rows.Scan(&transactionType, &targetType, &uuid, &fromTeamId, &initiatorUserId, &toUserId, &toTeamId,
			&amountGrams, &status, &notes, &paymentTime, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("扫描团队交易记录失败: %v", err)
		}

		transaction := map[string]interface{}{
			"transaction_type":  transactionType,
			"target_type":       targetType,
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

// IsTeamActiveMember 检查用户是否是团队成员正常状态TeamMemberStatusActive(1)
func IsTeamActiveMember(userId, teamId int) (bool, error) {
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

// CreateTeaTransferTeamToUser 创建团队向用户转账记录
func CreateTeaTransferTeamToUser(fromTeamId, initiatorUserId, toUserId int, amount float64, notes string, expireHours int) (TeaTeamToUserTransferOut, error) {
	// 验证参数
	if amount <= 0 {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：转账金额必须大于0")
	}
	if fromTeamId == TeamIdFreelancer {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：自由人团队不能发起转账")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出团队账户余额和锁定金额
	var teamBalance, teamLockedBalance float64
	var teamStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", fromTeamId).
		Scan(&teamBalance, &teamLockedBalance, &teamStatus)
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：查询转出团队账户失败 - %v", err)
	}

	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := teamBalance - teamLockedBalance
	if availableBalance < amount {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：团队可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", teamBalance, teamLockedBalance, availableBalance)
	}

	// 检查发起人是否是团队成员
	isMember, err := IsTeamActiveMember(initiatorUserId, fromTeamId)
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：只有团队成员才能发起团队转账")
	}

	// 确保接收方用户账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea.user_accounts WHERE user_id = $1", toUserId).Scan(&toAccountId)
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：接收方用户账户不存在 - %v", err)
	}

	// 获取团队名称和用户名称
	var fromTeamName, toUserName string
	err = tx.QueryRow("SELECT name FROM teams WHERE id = $1", fromTeamId).Scan(&fromTeamName)
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：获取转出团队名称失败 - %v", err)
	}
	err = tx.QueryRow("SELECT name FROM users WHERE id = $1", toUserId).Scan(&toUserName)
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：获取接收用户名称失败 - %v", err)
	}

	// 确定转账类型（单人团队自动批准，多人团队需要审批）
	var approverUserId *int
	var approvedAt *time.Time
	var status string
	teamMemberCount, err := getTeamMemberCount(tx, fromTeamId)
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("查询团队成员数量失败: %v", err)
	}
	if teamMemberCount == 1 {
		// 单人团队自动批准
		status = TeaTransferStatusPendingReceipt
		approverUserId = &initiatorUserId
		now := time.Now()
		approvedAt = &now
	} else {
		// 多人团队需要审批
		status = TeaTransferStatusPendingApproval
	}

	// 创建转账记录
	transfer := TeaTeamToUserTransferOut{
		FromTeamId:      fromTeamId,
		FromTeamName:    fromTeamName,
		ToUserId:        toUserId,
		ToUserName:      toUserName,
		InitiatorUserId: initiatorUserId,
		AmountGrams:     amount,
		Notes:           notes,
		Status:          status,
		ExpiresAt:       time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:       time.Now(),
	}

	// 如果是单人团队，设置审批信息
	if teamMemberCount == 1 {
		transfer.IsApproved = true
		transfer.ApproverUserId = initiatorUserId
		transfer.ApprovedAt = *approvedAt
		transfer.IsOnlyOneMemberTeam = true
	}

	err = tx.QueryRow(`INSERT INTO tea.team_to_user_transfer_out
		(from_team_id, from_team_name, to_user_id, to_user_name, initiator_user_id, amount_grams, notes, status, approver_user_id, approved_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid`,
		fromTeamId, fromTeamName, toUserId, toUserName, initiatorUserId, amount, notes, status, approverUserId, approvedAt, transfer.ExpiresAt).
		Scan(&transfer.Id, &transfer.Uuid)
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：创建团队转出记录失败 - %v", err)
	}

	// 锁定团队账户的相应金额
	_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE team_id = $3",
		amount, time.Now(), fromTeamId)
	if err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：锁定团队账户金额失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaTeamToUserTransferOut{}, fmt.Errorf("转账失败：提交事务失败 - %v", err)
	}

	return transfer, nil
}

// CreateTeaTransferTeamToTeam 创建团队向团队转账记录
func CreateTeaTransferTeamToTeam(fromTeamId, initiatorUserId, toTeamId int, amount float64, notes string, expireHours int) (TeaTeamToTeamTransferOut, error) {
	// 验证参数
	if amount <= 0 {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：转账金额必须大于0")
	}
	if fromTeamId == TeamIdFreelancer || toTeamId == TeamIdFreelancer {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：自由人团队不能参与转账")
	}
	if fromTeamId == toTeamId {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：不能向自己团队转账")
	}

	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 检查转出团队账户余额和锁定金额
	var teamBalance, teamLockedBalance float64
	var teamStatus string
	err = tx.QueryRow("SELECT balance_grams, locked_balance_grams, status FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", fromTeamId).
		Scan(&teamBalance, &teamLockedBalance, &teamStatus)
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：查询转出团队账户失败 - %v", err)
	}

	// 计算可用余额 = 总余额 - 锁定金额
	availableBalance := teamBalance - teamLockedBalance
	if availableBalance < amount {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：团队可用余额不足。总余额: %.3f克, 锁定余额: %.3f克, 可用余额: %.3f克", teamBalance, teamLockedBalance, availableBalance)
	}

	// 检查发起人是否是团队成员
	isMember, err := IsTeamActiveMember(initiatorUserId, fromTeamId)
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：只有团队成员才能发起团队转账")
	}

	// 确保接收方团队账户存在
	var toAccountId int
	err = tx.QueryRow("SELECT id FROM tea.team_accounts WHERE team_id = $1", toTeamId).Scan(&toAccountId)
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：接收方团队账户不存在 - %v", err)
	}

	// 获取团队名称
	var fromTeamName, toTeamName string
	err = tx.QueryRow("SELECT name FROM teams WHERE id = $1", fromTeamId).Scan(&fromTeamName)
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：获取转出团队名称失败 - %v", err)
	}
	err = tx.QueryRow("SELECT name FROM teams WHERE id = $1", toTeamId).Scan(&toTeamName)
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：获取接收团队名称失败 - %v", err)
	}

	// 确定转账类型（单人团队自动批准，多人团队需要审批）
	var approverUserId *int
	var approvedAt *time.Time
	var status string
	teamMemberCount, err := getTeamMemberCount(tx, fromTeamId)
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("查询团队成员数量失败: %v", err)
	}
	if teamMemberCount == 1 {
		// 单人团队自动批准
		status = TeaTransferStatusPendingReceipt
		approverUserId = &initiatorUserId
		now := time.Now()
		approvedAt = &now
	} else {
		// 多人团队需要审批
		status = TeaTransferStatusPendingApproval
	}

	// 创建转账记录
	transfer := TeaTeamToTeamTransferOut{
		FromTeamId:      fromTeamId,
		FromTeamName:    fromTeamName,
		ToTeamId:        toTeamId,
		ToTeamName:      toTeamName,
		InitiatorUserId: initiatorUserId,
		AmountGrams:     amount,
		Notes:           notes,
		Status:          status,
		ExpiresAt:       time.Now().Add(time.Duration(expireHours) * time.Hour),
		CreatedAt:       time.Now(),
	}

	// 如果是单人团队，设置审批信息
	if teamMemberCount == 1 {
		transfer.IsApproved = true
		transfer.ApproverUserId = initiatorUserId
		transfer.ApprovedAt = *approvedAt
		transfer.IsOnlyOneMemberTeam = true
	}

	err = tx.QueryRow(`INSERT INTO tea.team_to_team_transfer_out
		(from_team_id, from_team_name, to_team_id, to_team_name, initiator_user_id, amount_grams, notes, status, approver_user_id, approved_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid`,
		fromTeamId, fromTeamName, toTeamId, toTeamName, initiatorUserId, amount, notes, status, approverUserId, approvedAt, transfer.ExpiresAt).
		Scan(&transfer.Id, &transfer.Uuid)
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：创建团队转出记录失败 - %v", err)
	}

	// 锁定团队账户的相应金额
	_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_grams = locked_balance_grams + $1, updated_at = $2 WHERE team_id = $3",
		amount, time.Now(), fromTeamId)
	if err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：锁定团队账户金额失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return TeaTeamToTeamTransferOut{}, fmt.Errorf("转账失败：提交事务失败 - %v", err)
	}

	return transfer, nil
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

// ============================================
// 团队对用户转账审批和确认相关函数
// ============================================

// ApproveTeamToUserTransfer 审批团队向用户转账
func ApproveTeamToUserTransfer(transferUuid string, approverUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("审批失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账记录
	var transfer TeaTeamToUserTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, initiator_user_id, amount_grams, status 
		FROM tea.team_to_user_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.InitiatorUserId,
			&transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("审批失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingApproval {
		return fmt.Errorf("审批失败：转账状态异常")
	}

	// 检查审批人是否是团队成员（不能自己审批自己）
	isMember, err := IsTeamActiveMember(approverUserId, transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return fmt.Errorf("审批失败：只有团队成员才能审批")
	}
	if approverUserId == transfer.InitiatorUserId {
		return fmt.Errorf("审批失败：不能自己审批自己发起的操作")
	}

	// 更新转账状态为已批准，并更新为待接收
	approvedAt := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_to_user_transfer_out SET 
		status = $1, approver_user_id = $2, approved_at = $3, updated_at = $4 
		WHERE id = $5`,
		TeaTransferStatusPendingReceipt, approverUserId, approvedAt, approvedAt, transfer.Id)
	if err != nil {
		return fmt.Errorf("审批失败：更新转账状态失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// RejectTeamToUserTransfer 拒绝团队向用户转账
func RejectTeamToUserTransfer(transferUuid string, approverUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("拒绝失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账记录
	var transfer TeaTeamToUserTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, initiator_user_id, amount_grams, status 
		FROM tea.team_to_user_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.InitiatorUserId,
			&transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("拒绝失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingApproval {
		return fmt.Errorf("拒绝失败：转账状态异常")
	}

	// 检查审批人是否是团队成员（不能自己审批自己）
	isMember, err := IsTeamActiveMember(approverUserId, transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return fmt.Errorf("拒绝失败：只有团队成员才能审批")
	}
	if approverUserId == transfer.InitiatorUserId {
		return fmt.Errorf("拒绝失败：不能自己审批自己发起的操作")
	}

	// 获取团队账户锁定金额
	var teamLockedBalance float64
	err = tx.QueryRow("SELECT locked_balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", transfer.FromTeamId).Scan(&teamLockedBalance)
	if err != nil {
		return fmt.Errorf("查询团队账户锁定金额失败: %v", err)
	}

	// 解锁团队账户的锁定金额
	newLockedBalance := teamLockedBalance - transfer.AmountGrams
	if newLockedBalance < 0 {
		newLockedBalance = 0
	}

	_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_grams = $1, updated_at = $2 WHERE team_id = $3",
		newLockedBalance, time.Now(), transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("解锁团队账户金额失败: %v", err)
	}

	// 更新转账状态为已拒绝
	rejectedAt := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_to_user_transfer_out SET 
		status = $1, approver_user_id = $2, approval_rejection_reason = $3, rejected_by = $4, rejected_at = $5, updated_at = $6 
		WHERE id = $7`,
		TeaTransferStatusApprovalRejected, approverUserId, reason, approverUserId, rejectedAt, rejectedAt, transfer.Id)
	if err != nil {
		return fmt.Errorf("拒绝失败：更新转账状态失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// ConfirmTeamToUserTransfer 确认接收团队向用户转账
func ConfirmTeamToUserTransfer(transferUuid string, confirmUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("确认转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaTeamToUserTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, to_user_id, amount_grams, status, expires_at 
		FROM tea.team_to_user_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.ToUserId,
			&transfer.AmountGrams, &transfer.Status, &transfer.ExpiresAt)
	if err != nil {
		return fmt.Errorf("确认转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingReceipt {
		return fmt.Errorf("确认转账失败：转账状态异常")
	}
	if time.Now().After(transfer.ExpiresAt) {
		// 转账已过期，更新状态
		_, _ = tx.Exec("UPDATE tea.team_to_user_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			TeaTransferStatusExpired, time.Now(), transfer.Id)
		return fmt.Errorf("确认转账失败：转账已过期")
	}

	// 检查确认权限（只有接收方用户才能确认）
	if transfer.ToUserId != confirmUserId {
		return fmt.Errorf("无权确认此转账")
	}

	// 获取接收用户账户余额
	var toUserBalance float64
	err = tx.QueryRow("SELECT balance_grams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", transfer.ToUserId).Scan(&toUserBalance)
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
		toUserBalance+transfer.AmountGrams, time.Now(), transfer.ToUserId)
	if err != nil {
		return fmt.Errorf("更新接收用户账户余额失败: %v", err)
	}

	// 更新转账状态，设置实际支付时间，记录转出后余额
	paymentTime := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_to_user_transfer_out SET 
		status = $1, 
		payment_time = $2, 
		balance_after_transfer = $3,
		updated_at = $4 
		WHERE id = $5`,
		TeaTransferStatusCompleted, paymentTime, newFromTeamBalance, paymentTime, transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建转入记录（团队向用户转账确认时创建转入记录，关联到确认用户，记录接收后余额）
	_, err = tx.Exec(`INSERT INTO tea.user_from_team_transfer_in
		(team_to_user_transfer_out_id, to_user_id, to_user_name, from_team_id, from_team_name,
		amount_grams, notes, balance_after_receipt, status, is_confirmed, operational_user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		transfer.Id, transfer.ToUserId, transfer.ToUserName, transfer.FromTeamId, transfer.FromTeamName,
		transfer.AmountGrams, transfer.Notes, toUserBalance+transfer.AmountGrams, TeaTransferStatusCompleted, true, confirmUserId, transfer.ExpiresAt, paymentTime)
	if err != nil {
		return fmt.Errorf("创建转入记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// RejectTeamToUserTransferReceipt 拒绝接收团队向用户转账
func RejectTeamToUserTransferReceipt(transferUuid string, rejectUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("拒绝转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaTeamToUserTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, to_user_id, amount_grams, status 
		FROM tea.team_to_user_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.ToUserId,
			&transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("拒绝转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingReceipt {
		return fmt.Errorf("拒绝转账失败：转账状态异常")
	}

	// 检查拒绝权限（只有接收方用户才能拒绝）
	if transfer.ToUserId != rejectUserId {
		return fmt.Errorf("无权拒绝此转账")
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

	// 更新转账状态为已拒收
	_, err = tx.Exec(`UPDATE tea.team_to_user_transfer_out SET 
		status = $1, 
		updated_at = $2 
		WHERE id = $3`,
		TeaTransferStatusRejected, time.Now(), transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建拒收记录
	_, err = tx.Exec(`INSERT INTO tea.user_from_team_transfer_in
		(team_to_user_transfer_out_id, to_user_id, to_user_name, from_team_id, from_team_name,
		amount_grams, notes, status, is_confirmed, operational_user_id, reception_rejection_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		transfer.Id, transfer.ToUserId, transfer.ToUserName, transfer.FromTeamId, transfer.FromTeamName,
		transfer.AmountGrams, transfer.Notes, TeaTransferStatusRejected, false, rejectUserId, reason, time.Now())
	if err != nil {
		return fmt.Errorf("创建拒收记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// ============================================
// 团队对团队转账审批和确认相关函数
// ============================================

// ApproveTeamToTeamTransfer 审批团队向团队转账
func ApproveTeamToTeamTransfer(transferUuid string, approverUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("审批失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账记录
	var transfer TeaTeamToTeamTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, initiator_user_id, amount_grams, status 
		FROM tea.team_to_team_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.InitiatorUserId,
			&transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("审批失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingApproval {
		return fmt.Errorf("审批失败：转账状态异常")
	}

	// 检查审批人是否是团队成员（不能自己审批自己）
	isMember, err := IsTeamActiveMember(approverUserId, transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return fmt.Errorf("审批失败：只有团队成员才能审批")
	}
	if approverUserId == transfer.InitiatorUserId {
		return fmt.Errorf("审批失败：不能自己审批自己发起的操作")
	}

	// 更新转账状态为已批准，并更新为待接收
	approvedAt := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_to_team_transfer_out SET 
		status = $1, approver_user_id = $2, approved_at = $3, updated_at = $4 
		WHERE id = $5`,
		TeaTransferStatusPendingReceipt, approverUserId, approvedAt, approvedAt, transfer.Id)
	if err != nil {
		return fmt.Errorf("审批失败：更新转账状态失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// RejectTeamToTeamTransfer 拒绝团队向团队转账
func RejectTeamToTeamTransfer(transferUuid string, approverUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("拒绝失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账记录
	var transfer TeaTeamToTeamTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, initiator_user_id, amount_grams, status 
		FROM tea.team_to_team_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.InitiatorUserId,
			&transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("拒绝失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingApproval {
		return fmt.Errorf("拒绝失败：转账状态异常")
	}

	// 检查审批人是否是团队成员（不能自己审批自己）
	isMember, err := IsTeamActiveMember(approverUserId, transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return fmt.Errorf("拒绝失败：只有团队成员才能审批")
	}
	if approverUserId == transfer.InitiatorUserId {
		return fmt.Errorf("拒绝失败：不能自己审批自己发起的操作")
	}

	// 获取团队账户锁定金额
	var teamLockedBalance float64
	err = tx.QueryRow("SELECT locked_balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", transfer.FromTeamId).Scan(&teamLockedBalance)
	if err != nil {
		return fmt.Errorf("查询团队账户锁定金额失败: %v", err)
	}

	// 解锁团队账户的锁定金额
	newLockedBalance := teamLockedBalance - transfer.AmountGrams
	if newLockedBalance < 0 {
		newLockedBalance = 0
	}

	_, err = tx.Exec("UPDATE tea.team_accounts SET locked_balance_grams = $1, updated_at = $2 WHERE team_id = $3",
		newLockedBalance, time.Now(), transfer.FromTeamId)
	if err != nil {
		return fmt.Errorf("解锁团队账户金额失败: %v", err)
	}

	// 更新转账状态为已拒绝
	rejectedAt := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_to_team_transfer_out SET 
		status = $1, approver_user_id = $2, approval_rejection_reason = $3, rejected_by = $4, rejected_at = $5, updated_at = $6 
		WHERE id = $7`,
		TeaTransferStatusApprovalRejected, approverUserId, reason, approverUserId, rejectedAt, rejectedAt, transfer.Id)
	if err != nil {
		return fmt.Errorf("拒绝失败：更新转账状态失败 - %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// ConfirmTeamToTeamTransfer 确认接收团队向团队转账
func ConfirmTeamToTeamTransfer(transferUuid string, confirmUserId int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("确认转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaTeamToTeamTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, to_team_id, amount_grams, status, expires_at 
		FROM tea.team_to_team_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.ToTeamId,
			&transfer.AmountGrams, &transfer.Status, &transfer.ExpiresAt)
	if err != nil {
		return fmt.Errorf("确认转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingReceipt {
		return fmt.Errorf("确认转账失败：转账状态异常")
	}
	if time.Now().After(transfer.ExpiresAt) {
		// 转账已过期，更新状态
		_, _ = tx.Exec("UPDATE tea.team_to_team_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			TeaTransferStatusExpired, time.Now(), transfer.Id)
		return fmt.Errorf("确认转账失败：转账已过期")
	}

	// 检查确认权限（只有接收方团队成员才能确认）
	isMember, err := IsTeamActiveMember(confirmUserId, transfer.ToTeamId)
	if err != nil {
		return fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return fmt.Errorf("只有团队成员才能确认团队转账")
	}

	// 获取接收团队账户余额
	var toTeamBalance float64
	err = tx.QueryRow("SELECT balance_grams FROM tea.team_accounts WHERE team_id = $1 FOR UPDATE", transfer.ToTeamId).Scan(&toTeamBalance)
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
		toTeamBalance+transfer.AmountGrams, time.Now(), transfer.ToTeamId)
	if err != nil {
		return fmt.Errorf("更新接收团队账户余额失败: %v", err)
	}

	// 更新转账状态，设置实际支付时间，记录转出后余额
	paymentTime := time.Now()
	_, err = tx.Exec(`UPDATE tea.team_to_team_transfer_out SET 
		status = $1, 
		payment_time = $2, 
		balance_after_transfer = $3,
		updated_at = $4 
		WHERE id = $5`,
		TeaTransferStatusCompleted, paymentTime, newFromTeamBalance, paymentTime, transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建转入记录（团队转账确认时创建转入记录，关联到确认用户，记录接收后余额）
	_, err = tx.Exec(`INSERT INTO tea.team_from_team_transfer_in
		(team_to_team_transfer_out_id, to_team_id, to_team_name, from_team_id, from_team_name,
		amount_grams, notes, balance_after_transfer, status, is_confirmed, operational_user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		transfer.Id, transfer.ToTeamId, transfer.ToTeamName, transfer.FromTeamId, transfer.FromTeamName,
		transfer.AmountGrams, transfer.Notes, toTeamBalance+transfer.AmountGrams, TeaTransferStatusCompleted, true, confirmUserId, transfer.ExpiresAt, paymentTime)
	if err != nil {
		return fmt.Errorf("创建转入记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// RejectTeamToTeamTransferReceipt 拒绝接收团队向团队转账
func RejectTeamToTeamTransferReceipt(transferUuid string, rejectUserId int, reason string) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("拒绝转账失败：开始事务失败 - %v", err)
	}
	defer tx.Rollback()

	// 获取转账信息
	var transfer TeaTeamToTeamTransferOut
	err = tx.QueryRow(`SELECT id, uuid, from_team_id, to_team_id, amount_grams, status 
		FROM tea.team_to_team_transfer_out WHERE uuid = $1 FOR UPDATE`, transferUuid).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromTeamId, &transfer.ToTeamId,
			&transfer.AmountGrams, &transfer.Status)
	if err != nil {
		return fmt.Errorf("拒绝转账失败：转账记录不存在 - %v", err)
	}

	// 验证状态
	if transfer.Status != TeaTransferStatusPendingReceipt {
		return fmt.Errorf("拒绝转账失败：转账状态异常")
	}

	// 检查拒绝权限（只有接收方团队成员才能拒绝）
	isMember, err := IsTeamActiveMember(rejectUserId, transfer.ToTeamId)
	if err != nil {
		return fmt.Errorf("检查团队成员身份失败: %v", err)
	}
	if !isMember {
		return fmt.Errorf("只有团队成员才能拒绝团队转账")
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

	// 更新转账状态为已拒收
	_, err = tx.Exec(`UPDATE tea.team_to_team_transfer_out SET 
		status = $1, 
		updated_at = $2 
		WHERE id = $3`,
		TeaTransferStatusRejected, time.Now(), transfer.Id)
	if err != nil {
		return fmt.Errorf("更新转账状态失败: %v", err)
	}

	// 创建拒收记录
	_, err = tx.Exec(`INSERT INTO tea.team_from_team_transfer_in
		(team_to_team_transfer_out_id, to_team_id, to_team_name, from_team_id, from_team_name,
		amount_grams, notes, status, is_confirmed, operational_user_id, reception_rejection_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		transfer.Id, transfer.ToTeamId, transfer.ToTeamName, transfer.FromTeamId, transfer.FromTeamName,
		transfer.AmountGrams, transfer.Notes, TeaTransferStatusRejected, false, rejectUserId, reason, time.Now())
	if err != nil {
		return fmt.Errorf("创建拒收记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// ============================================
// 团队转账查询相关函数
// ============================================

// GetPendingTeamToUserOperations 获取团队待审批团队向用户转账列表
func GetPendingTeamToUserOperations(teamId int, page, limit int) ([]TeaTeamToUserTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`SELECT id, uuid, from_team_id, from_team_name, to_user_id, to_user_name,
		initiator_user_id, amount_grams, status, notes, balance_after_transfer, expires_at, created_at
		FROM tea.team_to_user_transfer_out
		WHERE from_team_id = $1 AND status = $2 AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		teamId, TeaTransferStatusPendingApproval, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待审批操作失败: %v", err)
	}
	defer rows.Close()

	var operations []TeaTeamToUserTransferOut
	for rows.Next() {
		var operation TeaTeamToUserTransferOut
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.FromTeamId, &operation.FromTeamName,
			&operation.ToUserId, &operation.ToUserName, &operation.InitiatorUserId,
			&operation.AmountGrams, &operation.Status, &operation.Notes,
			&operation.BalanceAfterTransfer, &operation.ExpiresAt, &operation.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// GetPendingTeamToTeamOperations 获取团队待审批团队向团队转账列表
func GetPendingTeamToTeamOperations(teamId int, page, limit int) ([]TeaTeamToTeamTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`SELECT id, uuid, from_team_id, from_team_name, to_team_id, to_team_name,
		initiator_user_id, amount_grams, status, notes, balance_after_transfer, expires_at, created_at
		FROM tea.team_to_team_transfer_out
		WHERE from_team_id = $1 AND status = $2 AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		teamId, TeaTransferStatusPendingApproval, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待审批操作失败: %v", err)
	}
	defer rows.Close()

	var operations []TeaTeamToTeamTransferOut
	for rows.Next() {
		var operation TeaTeamToTeamTransferOut
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.FromTeamId, &operation.FromTeamName,
			&operation.ToTeamId, &operation.ToTeamName, &operation.InitiatorUserId,
			&operation.AmountGrams, &operation.Status, &operation.Notes,
			&operation.BalanceAfterTransfer, &operation.ExpiresAt, &operation.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// GetPendingTeamToUserReceipts 获取团队待接收用户向团队转账列表
func GetPendingTeamToUserReceipts(teamId int, page, limit int) ([]TeaUserToTeamTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`SELECT id, uuid, from_user_id, from_user_name, to_team_id, to_team_name,
		amount_grams, status, notes, balance_after_transfer, expires_at, created_at
		FROM tea.user_to_team_transfer_out
		WHERE to_team_id = $1 AND status = $2 AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		teamId, TeaTransferStatusPendingReceipt, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待接收转账失败: %v", err)
	}
	defer rows.Close()

	var operations []TeaUserToTeamTransferOut
	for rows.Next() {
		var operation TeaUserToTeamTransferOut
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.FromUserId, &operation.FromUserName,
			&operation.ToTeamId, &operation.ToTeamName, &operation.AmountGrams, &operation.Status,
			&operation.Notes, &operation.BalanceAfterTransfer, &operation.ExpiresAt, &operation.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// GetPendingTeamToTeamReceipts 获取团队待接收团队向团队转账列表
func GetPendingTeamToTeamReceipts(teamId int, page, limit int) ([]TeaTeamToTeamTransferOut, error) {
	offset := (page - 1) * limit
	rows, err := DB.Query(`SELECT id, uuid, from_team_id, from_team_name, to_team_id, to_team_name,
		initiator_user_id, amount_grams, status, notes, balance_after_transfer, expires_at, created_at
		FROM tea.team_to_team_transfer_out
		WHERE to_team_id = $1 AND status = $2 AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		teamId, TeaTransferStatusPendingReceipt, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询待接收转账失败: %v", err)
	}
	defer rows.Close()

	var operations []TeaTeamToTeamTransferOut
	for rows.Next() {
		var operation TeaTeamToTeamTransferOut
		err = rows.Scan(&operation.Id, &operation.Uuid, &operation.FromTeamId, &operation.FromTeamName,
			&operation.ToTeamId, &operation.ToTeamName, &operation.InitiatorUserId,
			&operation.AmountGrams, &operation.Status, &operation.Notes,
			&operation.BalanceAfterTransfer, &operation.ExpiresAt, &operation.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描操作记录失败: %v", err)
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

// CountPendingTeamReceipts 获取团队待确认接收操作数量
func CountPendingTeamReceipts(teamId int) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM (
		SELECT id FROM tea.user_to_team_transfer_out
		WHERE to_team_id = $1 AND status = $2 AND expires_at > NOW()
		UNION ALL
		SELECT id FROM tea.team_to_team_transfer_out
		WHERE to_team_id = $1 AND status = $2 AND expires_at > NOW()
	) AS combined`,
		teamId, TeaTransferStatusPendingReceipt).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询待确认接收操作数量失败: %v", err)
	}
	return count, nil
}

// ============================================
// 兼容性函数（为向后兼容保留的统一接口）
// ============================================

// ApproveTeamTransfer 审批团队转账（兼容函数，自动判断转账类型）
// 注意：这是兼容函数，建议使用具体的 ApproveTeamToUserTransfer 或 ApproveTeamToTeamTransfer
func ApproveTeamTransfer(transferUuid string, approverUserId int) error {
	// 先尝试团队对用户转账
	err := ApproveTeamToUserTransfer(transferUuid, approverUserId)
	if err == nil {
		return nil
	}

	// 如果不是团队对用户转账，尝试团队对团队转账
	err = ApproveTeamToTeamTransfer(transferUuid, approverUserId)
	if err != nil {
		return fmt.Errorf("审批失败：找不到对应的转账记录")
	}

	return nil
}

// RejectTeamTransfer 拒绝团队转账（兼容函数，自动判断转账类型）
// 注意：这是兼容函数，建议使用具体的 RejectTeamToUserTransfer 或 RejectTeamToTeamTransfer
func RejectTeamTransfer(transferUuid string, approverUserId int, reason string) error {
	// 先尝试团队对用户转账
	err := RejectTeamToUserTransfer(transferUuid, approverUserId, reason)
	if err == nil {
		return nil
	}

	// 如果不是团队对用户转账，尝试团队对团队转账
	err = RejectTeamToTeamTransfer(transferUuid, approverUserId, reason)
	if err != nil {
		return fmt.Errorf("拒绝失败：找不到对应的转账记录")
	}

	return nil
}

// ConfirmTeamTransfer 确认接收团队转账（兼容函数，自动判断转账类型）
// 注意：这是兼容函数，建议使用具体的 ConfirmTeamToUserTransfer 或 ConfirmTeamToTeamTransfer
func ConfirmTeamTransfer(transferUuid string, confirmUserId int) error {
	// 先尝试团队向用户转账
	err := ConfirmTeamToUserTransfer(transferUuid, confirmUserId)
	if err == nil {
		return nil
	}

	// 如果不是团队向用户转账，尝试团队向团队转账
	err = ConfirmTeamToTeamTransfer(transferUuid, confirmUserId)
	if err != nil {
		return fmt.Errorf("确认失败：找不到对应的转账记录")
	}

	return nil
}

// RejectTeamTransferReceipt 拒绝接收团队转账（兼容函数，自动判断转账类型）
// 注意：这是兼容函数，建议使用具体的 RejectTeamToUserTransferReceipt 或 RejectTeamToTeamTransferReceipt
func RejectTeamTransferReceipt(transferUuid string, rejectUserId int, reason string) error {
	// 先尝试团队向用户转账
	err := RejectTeamToUserTransferReceipt(transferUuid, rejectUserId, reason)
	if err == nil {
		return nil
	}

	// 如果不是团队向用户转账，尝试团队向团队转账
	err = RejectTeamToTeamTransferReceipt(transferUuid, rejectUserId, reason)
	if err != nil {
		return fmt.Errorf("拒绝失败：找不到对应的转账记录")
	}

	return nil
}

// GetPendingTeamOperations 获取团队待审批操作列表（兼容函数）
// 注意：这是兼容函数，建议使用具体的 GetPendingTeamToUserOperations 或 GetPendingTeamToTeamOperations
func GetPendingTeamOperations(teamId int, page, limit int) (interface{}, error) {
	// 获取团队对用户待审批操作
	userTransfers, err := GetPendingTeamToUserOperations(teamId, page, limit)
	if err != nil {
		return nil, err
	}

	// 如果有结果，直接返回
	if len(userTransfers) > 0 {
		return userTransfers, nil
	}

	// 如果没有结果，尝试获取团队对团队待审批操作
	teamTransfers, err := GetPendingTeamToTeamOperations(teamId, page, limit)
	if err != nil {
		return nil, err
	}

	return teamTransfers, nil
}

// GetPendingTeamTransfers 获取团队待接收转账列表（兼容函数）
// 注意：这是兼容函数，建议使用具体的 GetPendingTeamToUserReceipts 或 GetPendingTeamToTeamReceipts
func GetPendingTeamTransfers(teamId int, page, limit int) (interface{}, error) {
	// 获取用户向团队待接收操作
	userReceipts, err := GetPendingTeamToUserReceipts(teamId, page, limit)
	if err != nil {
		return nil, err
	}

	// 如果有结果，直接返回
	if len(userReceipts) > 0 {
		return userReceipts, nil
	}

	// 如果没有结果，尝试获取团队向团队待接收操作
	teamReceipts, err := GetPendingTeamToTeamReceipts(teamId, page, limit)
	if err != nil {
		return nil, err
	}

	return teamReceipts, nil
}

// ============================================
// 团队转账历史查询相关函数
// ============================================

// GetTeamTransferInOperations 获取团队所有转入记录（接收）
func GetTeamTransferInOperations(teamId int, page, limit int) ([]map[string]interface{}, error) {
	offset := (page - 1) * limit
	transfers := []map[string]interface{}{}

	// 查询团队接收用户转入记录
	rows, err := DB.Query(`
		SELECT 'user' as transfer_type,
			   ti.id, ti.uuid, ti.user_to_team_transfer_out_id,
			   ti.to_team_id, ti.to_team_name,
			   ti.from_user_id, ti.from_user_name,
			   ti.amount_grams, ti.notes,
			   ti.status, ti.balance_after_transfer,
			   ti.is_confirmed, ti.operational_user_id, ti.reception_rejection_reason,
			   ti.expires_at, ti.created_at
		FROM tea.team_from_user_transfer_in ti
		WHERE ti.to_team_id = $1
		ORDER BY ti.created_at DESC LIMIT $2 OFFSET $3`,
		teamId, limit, offset)

	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var transferType string
			var receptionRejectionReason sql.NullString
			var id, userToTeamTransferOutId, toTeamId, fromUserId int
			var uuid, toTeamName, fromUserName, notes, status string
			var amountGrams, balanceAfterTransfer float64
			var isConfirmed bool
			var operationalUserId int
			var expiresAt, createdAt time.Time

			err = rows.Scan(
				&transferType,
				&id, &uuid, &userToTeamTransferOutId,
				&toTeamId, &toTeamName,
				&fromUserId, &fromUserName,
				&amountGrams, &notes,
				&status, &balanceAfterTransfer,
				&isConfirmed, &operationalUserId, &receptionRejectionReason,
				&expiresAt, &createdAt)
			if err == nil {
				transfer := map[string]interface{}{
					"transfer_type":                transferType,
					"id":                           id,
					"uuid":                         uuid,
					"user_to_team_transfer_out_id": userToTeamTransferOutId,
					"to_team_id":                   toTeamId,
					"to_team_name":                 toTeamName,
					"from_user_id":                 fromUserId,
					"from_user_name":               fromUserName,
					"amount_grams":                 amountGrams,
					"notes":                        notes,
					"status":                       status,
					"balance_after_transfer":       balanceAfterTransfer,
					"is_confirmed":                 isConfirmed,
					"operational_user_id":          operationalUserId,
					"reception_rejection_reason":   getNullableString(receptionRejectionReason),
					"expires_at":                   expiresAt,
					"created_at":                   createdAt,
				}
				transfers = append(transfers, transfer)
			}
		}
	}

	// 查询团队接收团队转入记录
	rows2, err := DB.Query(`
		SELECT 'team' as transfer_type,
			   ti.id, ti.uuid, ti.team_to_team_transfer_out_id,
			   ti.to_team_id, ti.to_team_name,
			   ti.from_team_id, ti.from_team_name,
			   ti.amount_grams, ti.notes,
			   ti.status, ti.balance_after_transfer,
			   ti.is_confirmed, ti.operational_user_id, ti.reception_rejection_reason,
			   ti.expires_at, ti.created_at
		FROM tea.team_from_team_transfer_in ti
		WHERE ti.to_team_id = $1
		ORDER BY ti.created_at DESC LIMIT $2 OFFSET $3`,
		teamId, limit, offset)

	if err == nil {
		defer rows2.Close()

		for rows2.Next() {
			var transferType string
			var receptionRejectionReason sql.NullString
			var id, teamToTeamTransferOutId, toTeamId, fromTeamId int
			var uuid, toTeamName, fromTeamName, notes, status string
			var amountGrams, balanceAfterTransfer float64
			var isConfirmed bool
			var operationalUserId int
			var expiresAt, createdAt time.Time

			err = rows2.Scan(
				&transferType,
				&id, &uuid, &teamToTeamTransferOutId,
				&toTeamId, &toTeamName,
				&fromTeamId, &fromTeamName,
				&amountGrams, &notes,
				&status, &balanceAfterTransfer,
				&isConfirmed, &operationalUserId, &receptionRejectionReason,
				&expiresAt, &createdAt)
			if err == nil {
				transfer := map[string]interface{}{
					"transfer_type":                transferType,
					"id":                           id,
					"uuid":                         uuid,
					"team_to_team_transfer_out_id": teamToTeamTransferOutId,
					"to_team_id":                   toTeamId,
					"to_team_name":                 toTeamName,
					"from_team_id":                 fromTeamId,
					"from_team_name":               fromTeamName,
					"amount_grams":                 amountGrams,
					"notes":                        notes,
					"status":                       status,
					"balance_after_transfer":       balanceAfterTransfer,
					"is_confirmed":                 isConfirmed,
					"operational_user_id":          operationalUserId,
					"reception_rejection_reason":   getNullableString(receptionRejectionReason),
					"expires_at":                   expiresAt,
					"created_at":                   createdAt,
				}
				transfers = append(transfers, transfer)
			}
		}
	}

	return transfers, nil
}

// GetTeamTransferOutOperations 获取团队所有转出操作历史（包括各种状态）
func GetTeamTransferOutOperations(teamId int, page, limit int) ([]map[string]interface{}, error) {
	offset := (page - 1) * limit
	operations := []map[string]interface{}{}

	// 查询团队向用户转出操作历史
	rows, err := DB.Query(`
		SELECT 'user' as transfer_type,
			   id, uuid, from_team_id, from_team_name,
			   to_user_id, to_user_name,
			   initiator_user_id, amount_grams, notes,
			   status, approver_user_id, approved_at,
			   approval_rejection_reason, rejected_by, rejected_at,
			   balance_after_transfer, expires_at, payment_time, created_at
		FROM tea.team_to_user_transfer_out
		WHERE from_team_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		teamId, limit, offset)

	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var transferType string
			var approverUserId, rejectedBy sql.NullInt64
			var approvedAt, rejectedAt, paymentTime sql.NullTime
			var approvalRejectionReason sql.NullString
			var id, fromTeamId, toUserId, initiatorUserId int
			var uuid, fromTeamName, toUserName, notes, status string
			var amountGrams, balanceAfterTransfer float64
			var expiresAt, createdAt time.Time

			err = rows.Scan(
				&transferType,
				&id, &uuid, &fromTeamId, &fromTeamName,
				&toUserId, &toUserName,
				&initiatorUserId, &amountGrams, &notes,
				&status, &approverUserId, &approvedAt,
				&approvalRejectionReason, &rejectedBy, &rejectedAt,
				&balanceAfterTransfer, &expiresAt, &paymentTime, &createdAt)
			if err == nil {
				operation := map[string]interface{}{
					"transfer_type":             transferType,
					"id":                        id,
					"uuid":                      uuid,
					"from_team_id":              fromTeamId,
					"from_team_name":            fromTeamName,
					"to_user_id":                toUserId,
					"to_user_name":              toUserName,
					"initiator_user_id":         initiatorUserId,
					"amount_grams":              amountGrams,
					"notes":                     notes,
					"status":                    status,
					"approver_user_id":          getNullableInt64(approverUserId),
					"approved_at":               getNullableTime(approvedAt),
					"approval_rejection_reason": getNullableString(approvalRejectionReason),
					"rejected_by":               getNullableInt64(rejectedBy),
					"rejected_at":               getNullableTime(rejectedAt),
					"balance_after_transfer":    balanceAfterTransfer,
					"expires_at":                expiresAt,
					"payment_time":              getNullableTime(paymentTime),
					"created_at":                createdAt,
				}
				operations = append(operations, operation)
			}
		}
	}

	// 查询团队向团队转出操作历史
	rows2, err := DB.Query(`
		SELECT 'team' as transfer_type,
			   id, uuid, from_team_id, from_team_name,
			   to_team_id, to_team_name,
			   initiator_user_id, amount_grams, notes,
			   status, approver_user_id, approved_at,
			   approval_rejection_reason, rejected_by, rejected_at,
			   balance_after_transfer, expires_at, payment_time, created_at
		FROM tea.team_to_team_transfer_out
		WHERE from_team_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		teamId, limit, offset)

	if err == nil {
		defer rows2.Close()

		for rows2.Next() {
			var transferType string
			var approverUserId, rejectedBy sql.NullInt64
			var approvedAt, rejectedAt, paymentTime sql.NullTime
			var approvalRejectionReason sql.NullString
			var id, fromTeamId, toTeamId, initiatorUserId int
			var uuid, fromTeamName, toTeamName, notes, status string
			var amountGrams, balanceAfterTransfer float64
			var expiresAt, createdAt time.Time

			err = rows2.Scan(
				&transferType,
				&id, &uuid, &fromTeamId, &fromTeamName,
				&toTeamId, &toTeamName,
				&initiatorUserId, &amountGrams, &notes,
				&status, &approverUserId, &approvedAt,
				&approvalRejectionReason, &rejectedBy, &rejectedAt,
				&balanceAfterTransfer, &expiresAt, &paymentTime, &createdAt)
			if err == nil {
				operation := map[string]interface{}{
					"transfer_type":             transferType,
					"id":                        id,
					"uuid":                      uuid,
					"from_team_id":              fromTeamId,
					"from_team_name":            fromTeamName,
					"to_team_id":                toTeamId,
					"to_team_name":              toTeamName,
					"initiator_user_id":         initiatorUserId,
					"amount_grams":              amountGrams,
					"notes":                     notes,
					"status":                    status,
					"approver_user_id":          getNullableInt64(approverUserId),
					"approved_at":               getNullableTime(approvedAt),
					"approval_rejection_reason": getNullableString(approvalRejectionReason),
					"rejected_by":               getNullableInt64(rejectedBy),
					"rejected_at":               getNullableTime(rejectedAt),
					"balance_after_transfer":    balanceAfterTransfer,
					"expires_at":                expiresAt,
					"payment_time":              getNullableTime(paymentTime),
					"created_at":                createdAt,
				}
				operations = append(operations, operation)
			}
		}
	}

	return operations, nil
}

// GetPendingTeamIncomingTransfers 获取团队所有待确认转入转账（包括用户转入和团队转入）
func GetPendingTeamIncomingTransfers(teamId int, page, limit int) ([]map[string]interface{}, error) {
	offset := (page - 1) * limit
	transfers := []map[string]interface{}{}

	// 查询用户向团队的待接收转账
	rows, err := DB.Query(`
		SELECT 'user_to_team' as transfer_type,
			   id, uuid, from_user_id, from_user_name, to_team_id, to_team_name,
			   amount_grams, status, notes, balance_after_transfer, expires_at, created_at
		FROM tea.user_to_team_transfer_out
		WHERE to_team_id = $1 AND status = $2 AND expires_at > NOW()
		UNION ALL
		SELECT 'team_to_team' as transfer_type,
			   id, uuid, from_team_id, from_team_name, to_team_id, to_team_name,
			   amount_grams, status, notes, balance_after_transfer, expires_at, created_at
		FROM tea.team_to_team_transfer_out
		WHERE to_team_id = $1 AND status = $2 AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		teamId, TeaTransferStatusPendingReceipt, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("查询团队待确认转入转账失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var transferType string
		var id, fromId, toTeamId int
		var uuid, fromName, toTeamName, notes, status string
		var amountGrams float64
		var balanceAfterTransfer sql.NullFloat64
		var expiresAt, createdAt time.Time

		err = rows.Scan(&transferType, &id, &uuid, &fromId, &fromName, &toTeamId, &toTeamName,
			&amountGrams, &status, &notes, &balanceAfterTransfer, &expiresAt, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("扫描转账记录失败: %v", err)
		}

		transfer := map[string]interface{}{
			"transfer_type":          transferType,
			"id":                     id,
			"uuid":                   uuid,
			"from_id":                fromId,
			"from_name":              fromName,
			"to_team_id":             toTeamId,
			"to_team_name":           toTeamName,
			"amount_grams":           amountGrams,
			"status":                 status,
			"notes":                  notes,
			"balance_after_transfer": getNullableFloat64(balanceAfterTransfer),
			"expires_at":             expiresAt,
			"created_at":             createdAt,
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// 辅助函数：处理sql.NullString
func getNullableString(nullString sql.NullString) interface{} {
	if nullString.Valid {
		return nullString.String
	}
	return nil
}

// 辅助函数：处理sql.NullFloat64
func getNullableFloat64(nullFloat sql.NullFloat64) interface{} {
	if nullFloat.Valid {
		return nullFloat.Float64
	}
	return nil
}
