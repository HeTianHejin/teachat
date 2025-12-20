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

// 团队交易类型常量（已统一使用 tea_transactions 表）
// 为了向后兼容保留这些常量，但实际使用 tea_account.go 中的 TransactionType_*
const (
	TransactionType_Deposit  = "deposit"
	TransactionType_Withdraw = "withdraw"
)

// 流通对象类型常量
const (
	TransactionTargetType_User = "u" // 个人
	TransactionTargetType_Team = "t" // 团队
)

// 茶叶账户流转规则：
// 个人对个人或者团队转账茶叶，无需审批，操作转出方的额定茶叶数量被锁定，接收方需要在有效期内确认接收，
// 接收方个人如果确认接受，按锁定额度清算双方账户数额并记录流水明细。
// 团队转出茶叶，无论对团队还是个人，都要求1成员发起转账操作，1核心成员审批，转出操作才生效；
// 团队审批转出的茶叶额定数量同个人账户一样会被锁定，等待接收方有效期内接收/拒绝，
// 团队接收茶叶转入，有效期内仅需要任意1成员确认接收即可结算双方账户，记录出入流水记录；
// 如果对方接收，才真正清算双方账户数额，创建实际流通交易流水记录，如果被接收方拒绝或者超时，解锁被转出方锁定茶叶，不创建交易流水记录。
// 超时处理，解锁转出方被锁定茶叶，双方无交易流水，有操作记录。

// 团队茶叶账户结构体
type TeamTeaAccount struct {
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

// GetTeamTeaAccountByTeamId 根据团队ID获取茶叶账户
func GetTeamTeaAccountByTeamId(teamId int) (TeamTeaAccount, error) {
	// 自由人团队没有茶叶资产，返回特殊的冻结账户
	if teamId == TeamIdFreelancer {
		reason := "自由人团队不支持茶叶资产"
		account := TeamTeaAccount{
			TeamId:       TeamIdFreelancer,
			BalanceGrams: 0.0,
			Status:       TeamTeaAccountStatus_Frozen,
			FrozenReason: &reason,
		}
		return account, nil
	}

	account := TeamTeaAccount{}
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
func (account *TeamTeaAccount) Create() error {
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
func (account *TeamTeaAccount) UpdateStatus(status, reason string) error {
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

// EnsureTeamTeaAccountExists 确保团队有茶叶账户
func EnsureTeamTeaAccountExists(teamId int) error {
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
	err := DB.QueryRow("SELECT status, frozen_reason FROM tea.team_accounts WHERE team_id = $1", teamId).
		Scan(&status, &frozenReason)
	if err != nil {
		return false, "", fmt.Errorf("查询团队账户状态失败: %v", err)
	}

	if status == TeamTeaAccountStatus_Frozen {
		return true, frozenReason.String, nil
	}

	return false, "", nil
}

// GetTeamTeaTransactions 获取团队交易流水（使用统一的 tea_transactions 表）
func GetTeamTeaTransactions(teamId int, page, limit int, transactionType string) ([]TeaTransaction, error) {
	offset := (page - 1) * limit
	var rows *sql.Rows
	var err error

	if transactionType == "" {
		rows, err = DB.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, target_user_id, target_team_id, target_type, created_at 
			FROM tea_transactions WHERE (user_id = $1 OR target_team_id = $1) 
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`, teamId, limit, offset)
	} else {
		rows, err = DB.Query(`SELECT id, uuid, user_id, transfer_id, transaction_type, 
			amount_grams, balance_before, balance_after, description, target_user_id, target_team_id, target_type, created_at 
			FROM tea_transactions WHERE (user_id = $1 OR target_team_id = $1) AND transaction_type = $2 
			ORDER BY created_at DESC LIMIT $3 OFFSET $4`, teamId, transactionType, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("查询团队交易流水失败: %v", err)
	}
	defer rows.Close()

	var transactions []TeaTransaction
	for rows.Next() {
		var transaction TeaTransaction
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
	_, err := DB.Exec(`INSERT INTO tea_transactions 
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
