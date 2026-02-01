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

用户星茶账户转账流程：
1、发起方法：用户填写对用户/团队、id、转出额度，无需审批，直接创建转帐表单；
2、锁定方法：转出的用户账户转出额度星茶数量被锁定，防止重复发起转账；
3.1、接收方法：在有效期内，接收方用户或者团队任意1正常状态成员TeamMemberStatusActive(1)，操作接收，创建接收记录，继续第4步；
3.2、拒绝方法：在有效期内，接收方用户或者团队任意1状态正常成员，操作拒收，创建拒收原因及拒收用户id、时间记录，流程结束；
4、清算方法：接收方确认接收后，按锁定额度（接收额度）清算双方账户数额，创建实际流通流水明细记录。
5、超时处理：自动解锁转出用户账户被锁定额度星茶，不创建交易流水明细记录。
*/

/* --团队星茶账户相关定义已迁移至tea_team_account.go文件，保留此注释以提示开发者--
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
/*
// 团队星茶账户状态常量
const (
	TeaTeamAccountStatus_Normal  = "normal"
	TeaTeamAccountStatus_Frozen  = "frozen"
	TeaTeamAccountStatus_Deleted = "deleted"
)

// 团队星茶账户结构体
type TeaTeamAccount struct {
	Id                 int
	Uuid               string
	TeamId             int
	BalanceMilligrams       int64 // 星茶数量(毫克)
	LockedBalanceMilligrams int64 // 被锁定的星茶数量(毫克)
	Status             string  // normal, frozen
	FrozenReason       string // 冻结原因，默认值:'-'
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
	AmountMilligrams     int64 // 转账星茶数量(毫克)，即锁定数量
	Notes           string  // 转账备注，默认值:'-'

	// 审批相关（团队转出时使用）
	IsOnlyOneMemberTeam bool // 默认值false:多人团队审批(必填)，true:单人团队自动批准
	// 审批人填写，必填
	IsApproved              bool      // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          int       // 审批人ID，团队核心成员id，多人团队不能是发起人id，（必填）单人团队自动批准时，审批人是发起人自己
	ApprovalRejectionReason string // 审批意见，如果拒绝，填写原因,默认值:'-'
	ApprovedAt              time.Time // 审批时间

	// 注意流程：审批通过而且对方确认接受之后，才会创建待接收记录（TeaUserFromTeamTransferIn）
	Status               string  // 包含审批状态，待审批，已批准，已拒绝，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer int64 // 转账后余额(毫克)
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
	AmountMilligrams     int64 // 转账星茶数量(毫克)，即锁定数量
	Notes           string  // 转账备注，默认值:'-'

	// 审批相关（团队转出时使用）
	IsOnlyOneMemberTeam bool // 默认值false:多人团队审批(必填)，true:单人团队自动批准
	// 审批人填写，必填
	IsApproved              bool      // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          int       // 审批人ID，团队核心成员id，多人团队不能是发起人id，（必填）单人团队自动批准时，审批人是发起人自己
	ApprovalRejectionReason string // 审批意见，如果拒绝，填写原因,默认值:'-'
	ApprovedAt              time.Time // 审批时间

	// 注意流程：审批通过，对方确认接收后，才会创建待接收记录（TeaTeamFromTeamTransferIn）
	Status               string  // 包含审批状态，待审批，已批准，已拒绝，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer int64 // 转账后余额(毫克)
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
	AmountMilligrams             int64 // 转账星茶数量(毫克)
	Notes                   string  // 转账备注，默认值:'-'
	Status                  string  // 包含接收状态，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer    int64 // 转账后余额(毫克)

	// 接收方ToTeam成员操作，Confirmed/Rejected二选一
	IsConfirmed              bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId        int    // 操作用户id，确认接收或者拒绝接收的用户id（团队成员）
	RejectionReason string // 如果拒绝，填写原因，默认值:'-'

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
	AmountMilligrams             int64 // 转账星茶数量(毫克)
	Notes                   string  // 转账备注，默认值:'-'
	Status                  string  // 包含接收状态，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer    int64 // 转账后余额(毫克)

	// 接收方ToTeam成员操作，Confirmed/Rejected二选一
	IsConfirmed              bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId        int    // 操作用户id，确认接收或者拒绝接收的用户id（团队成员）
	RejectionReason string // 如果拒绝，填写原因，默认值:'-'

	ExpiresAt time.Time // 过期时间，接收截止时间，也是FromTeam解锁额度时间
	CreatedAt time.Time // 必填，如果接收，是接收、清算时间；如果拒绝，是拒绝时间
}
	--团队星茶账户相关定义已迁移至tea_team_account.go文件，保留此注释以提示开发者--
*/

// 星茶账户状态常量
const (
	TeaAccountStatus_Normal = "normal"
	TeaAccountStatus_Frozen = "frozen"
)

// 星茶账户持有人类型常量
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

	// 用户和团队通用状态
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

// 用户星茶账户结构体
type TeaUserAccount struct {
	Id                      int
	Uuid                    string
	UserId                  int
	BalanceMilligrams       int64  // 星茶数量(毫克，1克=1000毫克)
	LockedBalanceMilligrams int64  // 交易有效期被锁定的星茶数量(毫克)
	Status                  string // normal, frozen
	FrozenReason            string // 冻结原因，默认值:'-'
	CreatedAt               time.Time
	UpdatedAt               *time.Time
}

// 用户对用户，专用星茶转账结构体，
// 注意：不能转出0或者负数，不能转给自己，不能转给被冻结User星茶账户
type TeaUserToUserTransferOut struct {
	Id           int
	Uuid         string
	FromUserId   int    // 转出方用户ID，对账单审计用
	FromUserName string // 转出方用户名称，对账单审计用
	ToUserId     int    // 接收方用户ID，对账单审计用
	ToUserName   string // 接收方用户名称，对账单审计用

	//转出用户填写
	AmountMilligrams int64  // 转账额度（毫克）
	Notes            string // 转账备注

	// 待接收	StatusPendingReceipt   = "pending_receipt"
	// 已完成	StatusCompleted        = "completed"
	// 以拒收	StatusRejected         = "rejected"
	// 已超时	StatusExpired          = "expired"
	Status               string     // 系统填写
	BalanceAfterTransfer int64      // 转出后账户余额（毫克），对账单审计用
	CreatedAt            time.Time  // 创建，流程开始时间，也是锁定额度起始时间
	ExpiresAt            time.Time  // 过期时间，也是锁定额度截止时间
	PaymentTime          *time.Time // （清算成功才有值）实际支付时间，关联对方确认接收时间
	UpdatedAt            *time.Time
}

// 用户对团队，专用星茶转账结构体
// 注意不能转出0/负数，不能转给自己、自由人团队id=TeamIdFreelancer(2)，不能转给被冻结Team星茶账户
type TeaUserToTeamTransferOut struct {
	Id           int
	Uuid         string
	FromUserId   int    // 转出方用户ID，对账单审计用
	FromUserName string // 转出方用户名称，对账单审计用
	ToTeamId     int    // 接收方团队ID，对账单审计用
	ToTeamName   string // 接收方团队名称，对账单审计用

	//转出用户填写
	AmountMilligrams int64
	Notes            string // 转账备注

	// 待接收	StatusPendingReceipt   = "pending_receipt"
	// 已完成	StatusCompleted        = "completed"
	// 以拒收	StatusRejected         = "rejected"
	// 已超时	StatusExpired          = "expired"
	Status               string // 系统填写
	BalanceAfterTransfer int64  // 转出后，FromUser账户余额（毫克），对账单审计用

	CreatedAt   time.Time  // 创建，流程开始时间，也是锁定额度起始时间
	ExpiresAt   time.Time  // 过期时间，也是锁定额度截止时间
	PaymentTime *time.Time // （清算成功才有值）实际支付时间，关联对方确认接收时间
	UpdatedAt   *time.Time
}

// 用户对用户，专用星茶转账接收记录结构体
type TeaUserFromUserTransferIn struct {
	Id   int
	Uuid string
	// 系统填写，对接转出表单，必填
	UserToUserTransferOutId int    // 用户对用户转出记录id
	ToUserId                int    // 接收用户id，账户持有人ID
	ToUserName              string // 接收用户名称，对账单审计用
	FromUserId              int    // 转出用户id
	FromUserName            string // 转出用户名称，对账单审计用

	AmountMilligrams    int64  // 接收转账额度（毫克），对账单审计用
	Notes               string // 转出方备注（从转出表复制过来）
	BalanceAfterReceipt int64  // 接收后账户余额（毫克），对账单审计用

	// 已完成	StatusCompleted        = "completed"
	// 已拒收	StatusRejected         = "rejected"
	Status string //方便阅读，对账单审计用

	// 接收方ToUser操作，completed/Rejected二选一
	IsConfirmed       bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId int    // 操作用户id，确认接收或者拒绝接收的用户id
	RejectionReason   string // 如果拒绝，填写原因，默认值:'-'

	ExpiresAt time.Time // 过期时间，接收截止时间，也是FromUser解锁额度时间
	CreatedAt time.Time // 必填，如果接收，是接收、清算时间；如果拒绝，是拒绝时间
}

// 用户对团队，专用星茶转账接收记录结构体
type TeaUserFromTeamTransferIn struct {
	Id   int
	Uuid string
	// 系统填写，对接转出表单，必填
	TeamToUserTransferOutId int    // 团队对用户转出记录id
	ToUserId                int    // 接收用户id，账户持有人ID
	ToUserName              string // 接收用户名称，对账单审计用
	FromTeamId              int    // 转出团队id
	FromTeamName            string // 转出团队名称，对账单审计用
	AmountMilligrams        int64  // 接收转账额度（毫克），对账单审计用
	Notes                   string // 转出方备注（从转出表复制过来），默认值:'-'
	BalanceAfterReceipt     int64  // 接收后账户余额（毫克），对账单审计用

	// 已完成	StatusCompleted        = "completed"
	// 已拒收	StatusRejected         = "rejected"
	Status string //方便阅读，对账单审计用
	// 接收方ToUser操作，Confirmed/Rejected二选一
	IsConfirmed     bool      // 默认false，默认不接收，避免转账错误被误接收
	RejectionReason string    // 如果拒绝，填写原因，默认值:'-'
	ExpiresAt       time.Time // 过期时间，接收截止时间，也是FromTeam解锁额度时间
	CreatedAt       time.Time // 必填，如果接收，是清算时间；如果拒绝，是拒绝时间
}

// TeaUserEnsureAccountExists 确保用户有星茶账户
func TeaUserEnsureAccountExists(userId int) error {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM tea.user_accounts WHERE user_id = $1)", userId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查账户存在性失败: %v", err)
	}

	if !exists {
		account := &TeaUserAccount{
			UserId:            userId,
			BalanceMilligrams: 0,
			Status:            TeaAccountStatus_Normal,
		}
		return account.Create()
	}

	return nil
}

// TeaUserAccount.Create 创建用户星茶账户
func (account *TeaUserAccount) Create() error {
	account.CreatedAt = time.Now()
	err := DB.QueryRow(`
		INSERT INTO tea.user_accounts (user_id, balance_milligrams, status, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, account.UserId, account.BalanceMilligrams, account.Status, account.CreatedAt).Scan(&account.Id)
	if err != nil {
		return fmt.Errorf("创建用户星茶账户失败: %v", err)
	}
	return nil
}

// updateStatus()
func (account *TeaUserAccount) UpdateStatus(status, reason string) error {
	statement := "UPDATE tea.user_accounts SET status = $2, frozen_reason = $3, updated_at = $4 WHERE id = $1"
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

// TeaUserProcessToUserExpiredTransfers 处理过期的用户对用户转账，解锁相应的锁定金额
func TeaUserProcessToUserExpiredTransfers() error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 查找所有过期且仍为pending_receipt状态的用户对用户转账
	rows, err := tx.Query(`
		SELECT id, from_user_id, amount_milligrams 
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
		Amount     int64
	}

	for rows.Next() {
		var et struct {
			Id         int
			FromUserId int
			Amount     int64
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
		var currentLockedBalance int64
		err = tx.QueryRow("SELECT locked_balance_milligrams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", et.FromUserId).Scan(&currentLockedBalance)
		if err != nil {
			return fmt.Errorf("查询锁定余额失败: %v", err)
		}

		// 检查锁定余额是否足够
		if currentLockedBalance < et.Amount {
			// 锁定余额不足，记录警告并跳过
			util.Error("用户对用户转账过期处理时，锁定余额不足，无法解锁，转账ID：%d，用户ID：%d，锁定余额：%d，转账金额：%d", et.Id, et.FromUserId, currentLockedBalance, et.Amount)
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
		_, err = tx.Exec("UPDATE tea.user_accounts SET locked_balance_milligrams = $1, updated_at = $2 WHERE user_id = $3",
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

// TeaUserProcessToTeamExpiredTransfers 处理过期的用户对团队转账，解锁相应的锁定金额
func TeaUserProcessToTeamExpiredTransfers() error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 查找所有过期且仍为pending_receipt状态的用户对团队转账
	rows, err := tx.Query(`
		SELECT id, from_user_id, amount_milligrams 
		FROM tea.user_to_team_transfer_out 
		WHERE status = $1 AND expires_at < $2`,
		TeaTransferStatusPendingReceipt, time.Now())
	if err != nil {
		return fmt.Errorf("查询过期转账失败: %v", err)
	}
	defer rows.Close()

	var expiredTransfers []struct {
		Id         int
		FromUserId int
		Amount     int64
	}

	for rows.Next() {
		var et struct {
			Id         int
			FromUserId int
			Amount     int64
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
		var currentLockedBalance int64
		err = tx.QueryRow("SELECT locked_balance_milligrams FROM tea.user_accounts WHERE user_id = $1 FOR UPDATE", et.FromUserId).Scan(&currentLockedBalance)
		if err != nil {
			return fmt.Errorf("查询锁定余额失败: %v", err)
		}

		// 检查锁定余额是否足够
		if currentLockedBalance < et.Amount {
			// 锁定余额不足，记录警告并跳过
			util.Error("用户对团队转账过期处理时，锁定余额不足，无法解锁，转账ID：%d，用户ID：%d，锁定余额：%d，转账金额：%d", et.Id, et.FromUserId, currentLockedBalance, et.Amount)
			continue
		}

		// 更新转账状态为过期
		_, err = tx.Exec("UPDATE tea.user_to_team_transfer_out SET status = $1, updated_at = $2 WHERE id = $3",
			TeaTransferStatusExpired, time.Now(), et.Id)
		if err != nil {
			return fmt.Errorf("更新过期转账状态失败: %v", err)
		}

		// 解锁相应的锁定金额
		newLockedBalance := currentLockedBalance - et.Amount
		_, err = tx.Exec("UPDATE tea.user_accounts SET locked_balance_milligrams = $1, updated_at = $2 WHERE user_id = $3",
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

// GetTeaAccountByUserId 根据用户ID获取星茶账户
func GetTeaAccountByUserId(userId int) (TeaUserAccount, error) {
	account := TeaUserAccount{}
	err := DB.QueryRow(`
		SELECT id, uuid, user_id, balance_milligrams, locked_balance_milligrams, 
		       status, frozen_reason, created_at, updated_at
		FROM tea.user_accounts 
		WHERE user_id = $1`, userId).
		Scan(&account.Id, &account.Uuid, &account.UserId, &account.BalanceMilligrams,
			&account.LockedBalanceMilligrams, &account.Status, &account.FrozenReason,
			&account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return account, fmt.Errorf("用户星茶账户不存在")
		}
		return account, fmt.Errorf("查询用户星茶账户失败: %v", err)
	}
	return account, nil
}

// CheckTeaUserAccountFrozen 检查用户星茶账户是否被冻结
func CheckTeaUserAccountFrozen(userId int) (bool, string, error) {
	var frozen bool
	var reason string
	err := DB.QueryRow(`
		SELECT status = 'frozen', frozen_reason
		FROM tea.user_accounts 
		WHERE user_id = $1`, userId).Scan(&frozen, &reason)
	if err != nil {
		return false, "", fmt.Errorf("查询用户星茶账户冻结状态失败: %v", err)
	}
	return frozen, reason, nil
}

func CreateTeaUserToUserTransferOut(fromUserId int, from_user_name string, toUserId int, to_user_name string, amount_milligrams int64, notes string, expireHours int) (TeaUserToUserTransferOut, error) {
	transfer := TeaUserToUserTransferOut{}
	err := DB.QueryRow(`
		INSERT INTO tea.user_to_user_transfer_out (from_user_id, to_user_id, from_user_name, to_user_name, amount_milligrams, notes, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, uuid, from_user_id,from_user_name, to_user_id, to_user_name , amount_milligrams, notes, status, expires_at, created_at, updated_at`,
		fromUserId, toUserId, from_user_name, to_user_name, amount_milligrams, notes, TeaTransferStatusPendingReceipt, time.Now().Add(time.Duration(expireHours)*time.Hour)).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName, &transfer.ToUserId, &transfer.ToUserName, &transfer.AmountMilligrams, &transfer.Notes, &transfer.Status, &transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
	if err != nil {
		return transfer, fmt.Errorf("创建用户间转账失败: %v", err)
	}
	return transfer, nil
}
func CreateTeaUserFromUserTransferIn(userToUserTransferOutId int, toUserId int, to_user_name string, fromUserId int, from_user_name string, amount_milligrams int64, notes string, balanceAfterReceipt int64, expiresAt time.Time) (TeaUserFromUserTransferIn, error) {
	transfer := TeaUserFromUserTransferIn{}
	err := DB.QueryRow(`
		INSERT INTO tea.user_from_user_transfer_in (user_to_user_transfer_out_id, to_user_id, to_user_name, from_user_id, from_user_name, amount_milligrams, notes, balance_after_receipt, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, uuid, user_to_user_transfer_out_id, to_user_id, to_user_name, from_user_id, from_user_name,
		          amount_milligrams, notes, balance_after_receipt, status, expires_at, created_at`,
		userToUserTransferOutId, toUserId, to_user_name, fromUserId, from_user_name,
		amount_milligrams, notes, balanceAfterReceipt,
		TeaTransferStatusPendingReceipt, expiresAt, time.Now()).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.UserToUserTransferOutId, &transfer.ToUserId, &transfer.ToUserName,
			&transfer.FromUserId, &transfer.FromUserName, &transfer.AmountMilligrams, &transfer.Notes,
			&transfer.BalanceAfterReceipt, &transfer.Status, &transfer.ExpiresAt, &transfer.CreatedAt)
	if err != nil {
		return transfer, fmt.Errorf("创建用户间转账接收记录失败: %v", err)
	}
	return transfer, nil
}

func CreateTeaUserToTeamTransferOut(from_user_id int, from_user_name string, to_team_id int, to_team_name string, amount_milligrams int64, notes string, expire_hours int) (TeaUserToTeamTransferOut, error) {
	transfer := TeaUserToTeamTransferOut{}
	err := DB.QueryRow(`
		INSERT INTO tea.user_to_team_transfer_out (from_user_id, from_user_name, to_team_id, to_team_name, amount_milligrams, notes, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, uuid, from_user_id, from_user_name,to_team_id,to_team_name , amount_milligrams, notes, status, expires_at, created_at, updated_at`,
		from_user_id, from_user_name, to_team_id, to_team_name, amount_milligrams, notes, TeaTransferStatusPendingReceipt, time.Now().Add(time.Duration(expire_hours)*time.Hour)).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName, &transfer.ToTeamId, &transfer.ToTeamName, &transfer.AmountMilligrams, &transfer.Notes, &transfer.Status, &transfer.ExpiresAt, &transfer.CreatedAt, &transfer.UpdatedAt)
	if err != nil {
		return transfer, fmt.Errorf("创建用户组间转账失败: %v", err)
	}
	return transfer, nil
}
func CreateTeaUserFromTeamTransferIn(teamToUserTransferOutId int, to_user_id int, to_user_name string, from_team_id int, from_team_name string, amount_milligrams int64, notes string, balance_after_receipt int64, expires_at time.Time) (TeaUserFromTeamTransferIn, error) {
	transfer := TeaUserFromTeamTransferIn{}
	err := DB.QueryRow(`
		INSERT INTO tea.user_from_team_transfer_in (team_to_user_transfer_out_id, to_user_id, to_user_name, from_team_id, from_team_name, amount_milligrams, notes, balance_after_receipt, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, uuid, team_to_user_transfer_out_id, to_user_id, to_user_name, from_team_id, from_team_name,
		          amount_milligrams, notes, balance_after_receipt, status, expires_at, created_at`,
		teamToUserTransferOutId, to_user_id, to_user_name, from_team_id, from_team_name,
		amount_milligrams, notes, balance_after_receipt, TeaTransferStatusPendingReceipt, expires_at, time.Now()).
		Scan(&transfer.Id, &transfer.Uuid, &transfer.TeamToUserTransferOutId, &transfer.ToUserId, &transfer.ToUserName,
			&transfer.FromTeamId, &transfer.FromTeamName, &transfer.AmountMilligrams, &transfer.Notes,
			&transfer.BalanceAfterReceipt, &transfer.Status, &transfer.ExpiresAt, &transfer.CreatedAt)
	if err != nil {
		return transfer, fmt.Errorf("创建用户对团队转账接收记录失败: %v", err)
	}
	return transfer, nil
}

// TeaUserOutToUserPendingTransfersCount 获取用户星茶账户发起的，待对方用户确认接收状态的转账数量
func TeaUserOutToUserPendingTransfersCount(fromUserId int) (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) 
		FROM tea.user_to_user_transfer_out 
		WHERE from_user_id = $1 AND status = $2`, fromUserId, TeaTransferStatusPendingReceipt).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询用户发起的，用户对用户,待处理状态转账数量失败: %v", err)
	}
	return count, nil
}

// TeaUserOutToTeamPendingTransfersCount 获取用户星茶账户发起的，待对方团队确认接收状态的转账数量
func TeaUserOutToTeamPendingTransfersCount(from_user_id int) (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) 
		FROM tea.user_to_team_transfer_out 
		WHERE from_user_id = $1 AND status = $2`, from_user_id, TeaTransferStatusPendingReceipt).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询用户发起的，用户对团队,待处理状态转账数量失败: %v", err)
	}
	return count, nil
}

// TeaUserOutToUserPendingTransfers 获取用户星茶账户发起的，待对方用户确认接收状态的转账记录
func TeaUserOutToUserPendingTransfers(from_user_id int, page, limit int) ([]TeaUserToUserTransferOut, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name,
		amount_milligrams, notes, status, expires_at, created_at, updated_at
		FROM tea.user_to_user_transfer_out
		WHERE from_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, from_user_id, TeaTransferStatusPendingReceipt, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户发起的，用户对用户,待处理状态转账失败: %v", err)
	}

	defer rows.Close()

	transfers := []TeaUserToUserTransferOut{}
	for rows.Next() {
		var transfer TeaUserToUserTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.AmountMilligrams,
			&transfer.Notes, &transfer.Status, &transfer.ExpiresAt,
			&transfer.CreatedAt, &transfer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户发起的，用户对用户,待处理状态转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户发起的，用户对用户,待处理状态转账记录失败: %v", err)
	}
	return transfers, nil
}

// TeaUserOutToTeamPendingTransfers 获取用户星茶账户发起的，待对方团队确认接收状态的转账记录
func TeaUserOutToTeamPendingTransfers(from_user_id int, page, limit int) ([]TeaUserToTeamTransferOut, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, from_user_name, to_team_id, to_team_name,
		amount_milligrams, notes, status, expires_at, created_at, updated_at
		FROM tea.user_to_team_transfer_out
		WHERE from_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, from_user_id, TeaTransferStatusPendingReceipt, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户发起的，用户对团队,待处理状态转账失败: %v", err)
	}

	defer rows.Close()

	transfers := []TeaUserToTeamTransferOut{}
	for rows.Next() {
		var transfer TeaUserToTeamTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToTeamId, &transfer.ToTeamName, &transfer.AmountMilligrams,
			&transfer.Notes, &transfer.Status, &transfer.ExpiresAt,
			&transfer.CreatedAt, &transfer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户发起的，用户对团队,待处理状态转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户发起的，用户对团队,待处理状态转账记录失败: %v", err)
	}
	return transfers, nil
}

// TeaUserInFromTeamPendingTransfers 获取用户星茶账户，待接收状态，来自团队转账记录
func TeaUserInFromTeamPendingTransfers(to_user_id int, page, limit int) ([]TeaTeamToUserTransferOut, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, to_user_id, to_user_name, from_team_id, from_team_name,
		amount_milligrams, notes, status, expires_at, created_at, updated_at
		FROM tea.team_to_user_transfer_out
		WHERE to_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, to_user_id, TeaTransferStatusPendingReceipt, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户待确认状态，来自团队转账失败: %v", err)
	}

	defer rows.Close()

	transfers := []TeaTeamToUserTransferOut{}
	for rows.Next() {
		var transfer TeaTeamToUserTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.ToUserId, &transfer.ToUserName,
			&transfer.FromTeamId, &transfer.FromTeamName, &transfer.AmountMilligrams,
			&transfer.Notes, &transfer.Status, &transfer.ExpiresAt,
			&transfer.CreatedAt, &transfer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户待确认状态，来自团队转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	return transfers, nil
}

// TeaUserInFromTeamPendingTransferOutsCount 获取用户星茶账户待处理状态,来自团队对用户转账数量
func TeaUserInFromTeamPendingTransferOutsCount(to_user_id int) (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) 
		FROM tea.team_to_user_transfer_out 
		WHERE to_user_id = $1 AND status = $2`, to_user_id, TeaTransferStatusPendingReceipt).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询用户待确认状态团队对用户转账数量失败: %v", err)
	}
	return count, nil
}

// TeaUserInFromUserPendingTransfers 获取用户星茶账户，待接收状态，来自其他用户转账记录
func TeaUserInFromUserPendingTransfers(to_user_id int, page, limit int) ([]TeaUserToUserTransferOut, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name,
		amount_milligrams, notes, status, expires_at, created_at, updated_at
		FROM tea.user_to_user_transfer_out
		WHERE to_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, to_user_id, TeaTransferStatusPendingReceipt, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户星茶账户待接收状态，来自其他用户转账失败: %v", err)
	}

	defer rows.Close()

	transfers := []TeaUserToUserTransferOut{}
	for rows.Next() {
		var transfer TeaUserToUserTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.AmountMilligrams,
			&transfer.Notes, &transfer.Status, &transfer.ExpiresAt,
			&transfer.CreatedAt, &transfer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户星茶账户待接收状态，来自其他用户转账记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户星茶账户待接收状态，来自其他用户转账记录失败: %v", err)
	}
	return transfers, nil
}

// TeaUserInFromUserPendingTransferOutsCount 获取用户星茶账户待处理状态,来自其他用户对用户转账数量
func TeaUserInFromUserPendingTransferOutsCount(to_user_id int) (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) 
		FROM tea.user_to_user_transfer_out 
		WHERE to_user_id = $1 AND status = $2`, to_user_id, TeaTransferStatusPendingReceipt).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询用户星茶账户待处理用户对用户转账数量失败: %v", err)
	}
	return count, nil
}

// 用户确认接收来自某个用户转账
func TeaUserConfirmFromUserTransferIn(transferUuid string, to_user_id int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID, amountMg int
	var fromUserId int
	var notes, toUserName, fromUserName string
	err = tx.QueryRow(`
		SELECT id, from_user_id, amount_milligrams, notes, to_user_name, from_user_name
		FROM tea.user_to_user_transfer_out 
		WHERE uuid = $1 AND to_user_id = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		transferUuid, to_user_id, TeaTransferStatusPendingReceipt,
	).Scan(&transferOutID, &fromUserId, &amountMg, &notes, &toUserName, &fromUserName)

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
		UPDATE tea.user_to_user_transfer_out
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
		AND locked_balance_milligrams >= $1`,
		amountMg, now, fromUserId)
	if err != nil {
		return fmt.Errorf("更新转出用户账户失败: %v", err)
	}

	// 3. 更新接收用户账户（增加余额）
	var receiverBalanceAfter int
	err = tx.QueryRow(`
		UPDATE tea.user_accounts 
		SET balance_milligrams = balance_milligrams + $1,
			updated_at = $2
		WHERE user_id = $3
		RETURNING balance_milligrams`,
		amountMg, now, to_user_id).Scan(&receiverBalanceAfter)
	if err != nil {
		return fmt.Errorf("更新接收用户账户失败: %v", err)
	}

	// 4. 创建接收记录
	_, err = tx.Exec(`
		INSERT INTO tea.user_from_user_transfer_in (
			user_to_user_transfer_out_id, to_user_id, to_user_name, 
			from_user_id, from_user_name, amount_milligrams, notes, 
			balance_after_receipt, status, is_confirmed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		transferOutID, to_user_id, toUserName, fromUserId, fromUserName,
		amountMg, notes, receiverBalanceAfter, TeaTransferStatusCompleted,
		true, now)
	if err != nil {
		return fmt.Errorf("创建接收来自用户转账记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	return nil
}

// 用户确认接收来自某个团队转账
func TeaUserConfirmFromTeamTransferIn(transferUuid string, to_user_id int) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID, amountMg, fromUserID int
	var fromTeamId int
	var notes, toUserName, fromTeamName string
	err = tx.QueryRow(`
		SELECT id, from_team_id, amount_milligrams, from_user_id, notes, to_user_name, from_team_name
		FROM tea.team_to_user_transfer_out 
		WHERE uuid = $1 AND to_user_id = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		transferUuid, to_user_id, TeaTransferStatusPendingReceipt,
	).Scan(&transferOutID, &fromTeamId, &amountMg, &fromUserID, &notes, &toUserName, &fromTeamName)

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
		UPDATE tea.team_to_user_transfer_out
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
		AND locked_balance_milligrams >= $1`,
		amountMg, now, fromUserID)
	if err != nil {
		return fmt.Errorf("更新转出用户账户失败: %v", err)
	}

	// 3. 更新接收用户账户（增加余额）
	var receiverBalanceAfter int
	err = tx.QueryRow(`
		UPDATE tea.user_accounts 
		SET balance_milligrams = balance_milligrams + $1,
			updated_at = $2
		WHERE user_id = $3
		RETURNING balance_milligrams`,
		amountMg, now, to_user_id).Scan(&receiverBalanceAfter)
	if err != nil {
		return fmt.Errorf("更新接收用户账户失败: %v", err)
	}

	// 4. 创建接收记录
	_, err = tx.Exec(`
		INSERT INTO tea.user_from_team_transfer_in (
			team_to_user_transfer_out_id, to_user_id, to_user_name, 
			from_team_id, from_team_name, amount_milligrams, notes, 
			balance_after_receipt, status, is_confirmed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		transferOutID, to_user_id, toUserName, fromTeamId, fromTeamName,
		amountMg, notes, receiverBalanceAfter, TeaTransferStatusCompleted,
		true, now)
	if err != nil {
		return fmt.Errorf("创建接收来自团队星茶转账记录失败: %v", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	return nil
}

// 获取用户来自用户已完成的转入记录（仅已完成状态）
func TeaUserFromUserCompletedTransferIns(user_id int, page, limit int) ([]TeaUserFromUserTransferIn, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, user_to_user_transfer_out_id, to_user_id, to_user_name, from_user_id, from_user_name,
		       amount_milligrams, notes, balance_after_receipt, status, is_confirmed, 
		       operational_user_id, rejection_reason, expires_at, created_at
		FROM tea.user_from_user_transfer_in
		WHERE to_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, user_id, TeaTransferStatusCompleted, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户来自用户已完成转入记录失败: %v", err)
	}
	defer rows.Close()

	transfers := []TeaUserFromUserTransferIn{}
	for rows.Next() {
		var transfer TeaUserFromUserTransferIn
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.UserToUserTransferOutId,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.AmountMilligrams, &transfer.Notes,
			&transfer.BalanceAfterReceipt,
			&transfer.Status,
			&transfer.IsConfirmed,
			&transfer.OperationalUserId,
			&transfer.RejectionReason,
			&transfer.ExpiresAt,
			&transfer.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户来自用户已完成转入记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户来自用户已完成转入记录失败: %v", err)
	}
	return transfers, nil
}

// 获取用户来自团队已完成的转入记录（仅已完成状态）
func TeaUserFromTeamCompletedTransferIns(user_id int, page, limit int) ([]TeaUserFromTeamTransferIn, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, team_to_user_transfer_out_id, to_user_id, to_user_name, from_team_id, from_team_name,
		       amount_milligrams, notes, balance_after_receipt, status, is_confirmed, 
		       rejection_reason, expires_at, created_at
		FROM tea.user_from_team_transfer_in
		WHERE to_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, user_id, TeaTransferStatusCompleted, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户来自团队已完成转入记录失败: %v", err)
	}
	defer rows.Close()

	transfers := []TeaUserFromTeamTransferIn{}
	for rows.Next() {
		var transfer TeaUserFromTeamTransferIn
		if err := rows.Scan(&transfer.Id, &transfer.Uuid,
			&transfer.TeamToUserTransferOutId,
			&transfer.ToUserId, &transfer.ToUserName,
			&transfer.FromTeamId, &transfer.FromTeamName,
			&transfer.AmountMilligrams, &transfer.Notes,
			&transfer.BalanceAfterReceipt,
			&transfer.Status,
			&transfer.IsConfirmed,
			&transfer.RejectionReason,
			&transfer.ExpiresAt,
			&transfer.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户来自团队已完成转入记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户来自团队已完成转入记录失败: %v", err)
	}
	return transfers, nil
}

// TeaUserRejectFromUserTransferIn 某个用户,拒绝接收,用户对用户转账
func TeaUserRejectFromUserTransferIn(transferUuid string, toUserId int, reason string) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var fromUserId int
	var toUserName, fromUserName, notes string
	var expiresAt time.Time
	err = tx.QueryRow(`
		SELECT id, from_user_id, amount_milligrams, to_user_name, from_user_name, notes, expires_at
		FROM tea.user_to_user_transfer_out 
		WHERE uuid = $1 AND to_user_id = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		transferUuid, toUserId, TeaTransferStatusPendingReceipt,
	).Scan(&transferOutID, &fromUserId, &amountMg, &toUserName, &fromUserName, &notes, &expiresAt)

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

	// 1. 更新转出记录状态为拒绝
	_, err = tx.Exec(`
		UPDATE tea.user_to_user_transfer_out
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4`,
		TeaTransferStatusRejected, now, transferOutID, TeaTransferStatusPendingReceipt)
	if err != nil {
		return fmt.Errorf("更新用户对用户转账状态为拒绝失败: %v", err)
	}

	// 2. 释放转出方账户锁定金额
	_, err = tx.Exec(`
		UPDATE tea.user_accounts
		SET locked_balance_milligrams = locked_balance_milligrams - $1,
			updated_at = $2
		WHERE user_id = $3
		AND locked_balance_milligrams >= $1`,
		amountMg, now, fromUserId)
	if err != nil {
		return fmt.Errorf("释放用户对用户转账锁定金额失败: %v", err)
	}

	// 3. 创建拒绝接收记录（用于历史审计）
	_, err = tx.Exec(`
		INSERT INTO tea.user_from_user_transfer_in (
			user_to_user_transfer_out_id, to_user_id, to_user_name, from_user_id, from_user_name,
			amount_milligrams, notes, status, is_confirmed, operational_user_id, rejection_reason, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		transferOutID, toUserId, toUserName, fromUserId, fromUserName,
		amountMg, notes, TeaTransferStatusRejected, false, toUserId, reason, expiresAt, now)
	if err != nil {
		return fmt.Errorf("创建用户对用户转账拒绝接收记录失败: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// TeaUserRejectFromTeamTransferIn 某个用户,拒绝接收,来自团队转账
func TeaUserRejectFromTeamTransferIn(transferUuid string, toUserId int, reason string) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 首先锁定并获取转账记录详情
	var transferOutID int
	var amountMg int64
	var fromTeamId int
	var toUserName, fromTeamName, notes string
	var expiresAt time.Time
	err = tx.QueryRow(`
		SELECT id, from_team_id, amount_milligrams, to_user_name, from_team_name, notes, expires_at
		FROM tea.team_to_user_transfer_out 
		WHERE uuid = $1 AND to_user_id = $2 AND status = $3
		FOR UPDATE SKIP LOCKED`,
		transferUuid, toUserId, TeaTransferStatusPendingReceipt,
	).Scan(&transferOutID, &fromTeamId, &amountMg, &toUserName, &fromTeamName, &notes, &expiresAt)

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

	// 1. 更新转出记录状态为拒绝
	_, err = tx.Exec(`
		UPDATE tea.team_to_user_transfer_out
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4`,
		TeaTransferStatusRejected, now, transferOutID, TeaTransferStatusPendingReceipt)
	if err != nil {
		return fmt.Errorf("更新团队对用户转账状态为拒绝失败: %v", err)
	}

	// 2. 释放转出团队账户锁定金额
	_, err = tx.Exec(`
		UPDATE tea.team_accounts
		SET locked_balance_milligrams = locked_balance_milligrams - $1,
			updated_at = $2
		WHERE team_id = $3
		AND locked_balance_milligrams >= $1`,
		amountMg, now, fromTeamId)
	if err != nil {
		return fmt.Errorf("释放团队对用户转账锁定金额失败: %v", err)
	}

	// 3. 创建拒绝接收记录（用于历史审计）
	_, err = tx.Exec(`
		INSERT INTO tea.user_from_team_transfer_in (
			team_to_user_transfer_out_id, to_user_id, to_user_name, from_team_id, from_team_name,
			amount_milligrams, notes, status, is_confirmed, operational_user_id, rejection_reason, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		transferOutID, toUserId, toUserName, fromTeamId, fromTeamName,
		amountMg, notes, TeaTransferStatusRejected, false, toUserId, reason, expiresAt, now)
	if err != nil {
		return fmt.Errorf("创建用户拒绝接收团队转账记录失败: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// 获取某个用户,已经拒绝,来自用户转入记录（已拒绝状态）
func TeaUserFromUserRejectedTransferIns(toUserId int, page, limit int) ([]TeaUserFromUserTransferIn, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, user_to_user_transfer_out_id, to_user_id, to_user_name, from_user_id, from_user_name,
		       amount_milligrams, notes, balance_after_receipt, status, is_confirmed,
		       operational_user_id, rejection_reason, expires_at, created_at
		FROM tea.user_from_user_transfer_in
		WHERE to_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, toUserId, TeaTransferStatusRejected, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户对用户已拒绝转入记录失败: %v", err)
	}
	defer rows.Close()

	transfers := []TeaUserFromUserTransferIn{}
	for rows.Next() {
		var transfer TeaUserFromUserTransferIn
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.UserToUserTransferOutId,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.AmountMilligrams, &transfer.Notes,
			&transfer.BalanceAfterReceipt,
			&transfer.Status,
			&transfer.IsConfirmed,
			&transfer.OperationalUserId,
			&transfer.RejectionReason,
			&transfer.ExpiresAt,
			&transfer.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户对用户已拒绝转入记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户对用户已拒绝转入记录失败: %v", err)
	}
	return transfers, nil
}

// 获取某个用户,已经拒绝,来自团队转入记录（已拒绝状态）
func TeaUserRejectedTeamToUserTransferIns(to_user_id int, page, limit int) ([]TeaUserFromTeamTransferIn, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, team_to_user_transfer_out_id, to_user_id, to_user_name, from_team_id, from_team_name,
		       amount_milligrams, notes, balance_after_receipt, status, is_confirmed,
		       rejection_reason, expires_at, created_at
		FROM tea.user_from_team_transfer_in
		WHERE to_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, to_user_id, TeaTransferStatusRejected, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户来自团队已拒绝转入记录失败: %v", err)
	}
	defer rows.Close()

	transfers := []TeaUserFromTeamTransferIn{}
	for rows.Next() {
		var transfer TeaUserFromTeamTransferIn
		if err := rows.Scan(&transfer.Id, &transfer.Uuid,
			&transfer.TeamToUserTransferOutId,
			&transfer.ToUserId, &transfer.ToUserName,
			&transfer.FromTeamId, &transfer.FromTeamName,
			&transfer.AmountMilligrams, &transfer.Notes,
			&transfer.BalanceAfterReceipt,
			&transfer.Status,
			&transfer.IsConfirmed,
			&transfer.RejectionReason,
			&transfer.ExpiresAt,
			&transfer.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户来自团队已拒绝转入记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户来自团队已拒绝转入记录失败: %v", err)
	}
	return transfers, nil
}

// 获取用户对用户转出已经过期记录
func TeaUserToUserExpiredTransferOuts(user_id, page, limit int) ([]TeaUserToUserTransferOut, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name,
		       amount_milligrams, notes, status, balance_after_transfer,
		       expires_at, created_at, payment_time, updated_at
		FROM tea.user_to_user_transfer_out
		WHERE from_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, user_id, TeaTransferStatusExpired, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户对用户已过期转出记录失败: %v", err)
	}
	defer rows.Close()
	transfers := []TeaUserToUserTransferOut{}
	for rows.Next() {
		var transfer TeaUserToUserTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.AmountMilligrams, &transfer.Notes,
			&transfer.Status, &transfer.BalanceAfterTransfer, &transfer.ExpiresAt, &transfer.CreatedAt,
			&transfer.PaymentTime, &transfer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户对用户已过期转出记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户对用户已过期转出记录失败: %v", err)
	}
	return transfers, nil
}

// 获取用户对团队转出已经过期记录
func TeaUserToTeamExpiredTransferOuts(user_id, page, limit int) ([]TeaUserToTeamTransferOut, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, from_user_name, to_team_id, to_team_name,
		       amount_milligrams, notes, status, balance_after_transfer,
		       expires_at, created_at, payment_time, updated_at
		FROM tea.user_to_team_transfer_out
		WHERE from_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, user_id, TeaTransferStatusExpired, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户对团队已过期转出记录失败: %v", err)
	}
	defer rows.Close()
	transfers := []TeaUserToTeamTransferOut{}
	for rows.Next() {
		var transfer TeaUserToTeamTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToTeamId, &transfer.ToTeamName, &transfer.AmountMilligrams, &transfer.Notes,
			&transfer.Status, &transfer.BalanceAfterTransfer, &transfer.ExpiresAt, &transfer.CreatedAt,
			&transfer.PaymentTime, &transfer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户对团队已过期转出记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户对团队已过期转出记录失败: %v", err)
	}
	return transfers, nil
}

// TeaUserToUserCompletedTransferOuts 获取用户对用户转出已完成记录（仅已完成状态）
func TeaUserToUserCompletedTransferOuts(from_user_id int, page, limit int) ([]TeaUserToUserTransferOut, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, from_user_name, to_user_id, to_user_name,
		       amount_milligrams, notes, status, balance_after_transfer,
		       expires_at, created_at, payment_time, updated_at
		FROM tea.user_to_user_transfer_out
		WHERE from_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, from_user_id, TeaTransferStatusCompleted, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户对用户转出已完成记录失败: %v", err)
	}
	defer rows.Close()

	transfers := []TeaUserToUserTransferOut{}
	for rows.Next() {
		var transfer TeaUserToUserTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToUserId, &transfer.ToUserName, &transfer.AmountMilligrams, &transfer.Notes,
			&transfer.Status, &transfer.BalanceAfterTransfer, &transfer.ExpiresAt, &transfer.CreatedAt,
			&transfer.PaymentTime, &transfer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户对用户转出已完成记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户对用户转出已完成记录失败: %v", err)
	}
	return transfers, nil
}

// TeaUserToTeamCompletedTransferOuts 获取用户对团队转出已完成记录（仅已完成状态）
func TeaUserToTeamCompletedTransferOuts(from_user_id int, page, limit int) ([]TeaUserToTeamTransferOut, error) {
	rows, err := DB.Query(`
		SELECT id, uuid, from_user_id, from_user_name, to_team_id, to_team_name,
		       amount_milligrams, notes, status, balance_after_transfer,
		       expires_at, created_at, payment_time, updated_at
		FROM tea.user_to_team_transfer_out
		WHERE from_user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, from_user_id, TeaTransferStatusCompleted, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("查询用户对团队转出已完成记录失败: %v", err)
	}
	defer rows.Close()

	transfers := []TeaUserToTeamTransferOut{}
	for rows.Next() {
		var transfer TeaUserToTeamTransferOut
		if err := rows.Scan(&transfer.Id, &transfer.Uuid, &transfer.FromUserId, &transfer.FromUserName,
			&transfer.ToTeamId, &transfer.ToTeamName, &transfer.AmountMilligrams, &transfer.Notes,
			&transfer.Status, &transfer.BalanceAfterTransfer, &transfer.ExpiresAt, &transfer.CreatedAt,
			&transfer.PaymentTime, &transfer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户对团队转出已完成记录失败: %v", err)
		}
		transfers = append(transfers, transfer)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历用户对团队转出已完成记录失败: %v", err)
	}
	return transfers, nil
}
