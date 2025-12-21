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
// 注意不能转出0/负数，不能转给自己和自由人团队id=TeamIdFreelancer
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

// GetTeamTeaTransactions 获取团队交易流水（使用统一的 tea_transactions 表）
func GetTeamTeaTransactions(teamId int, page, limit int, transactionType string) ([]TransactionRecord, error) {
	offset := (page - 1) * limit
	var rows *sql.Rows
	var err error

	if transactionType == "" {
		rows, err = DB.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, target_user_id, target_team_id, target_type, created_at 
			FROM tea.transaction_records WHERE (user_id = $1 OR target_team_id = $1) 
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`, teamId, limit, offset)
	} else {
		rows, err = DB.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, target_user_id, target_team_id, target_type, created_at 
			FROM tea.transaction_records WHERE (user_id = $1 OR target_team_id = $1) AND transaction_type = $2 
			ORDER BY created_at DESC LIMIT $3 OFFSET $4`, teamId, transactionType, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("查询团队交易流水失败: %v", err)
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
			return nil, fmt.Errorf("扫描团队交易流水失败: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// CreateTeamTeaTransaction 创建团队交易流水记录（使用统一的 tea_transactions 表）
func CreateTeamTeaTransaction(teamId int, transactionType string, amountGrams, balanceBefore, balanceAfter float64, description string, targetUserId, targetTeamId, operatorUserId, approverUserId *int, targetType string) error {
	// 对于团队交易，user_id 字段存储 team_id
	_, err := DB.Exec(`INSERT INTO tea.transaction_records 
		(user_id, transfer_id, transaction_type, amount_grams, balance_before, balance_after, description, target_user_id, target_team_id, target_type) 
		VALUES ($1, NULL, $2, $3, $4, $5, $6, $7, $8, $9)`,
		teamId, transactionType, amountGrams, balanceBefore, balanceAfter, description, targetUserId, targetTeamId, targetType)
	if err != nil {
		return fmt.Errorf("创建团队交易流水失败: %v", err)
	}
	return nil
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

// IsTeamMember 检查用户是否是团队成员
func IsTeamMember(userId, teamId int) (bool, error) {
	var count int
	err := DB.QueryRow(`
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

	rows, err := DB.Query(query, userId, TeMemberStatusActive, TeamIdFreelancer)
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
