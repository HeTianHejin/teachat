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
	AmountMilligrams     int64 // 转账星茶数量(毫克)，即锁定数量
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
	Notes           string  // 转账备注,默认值:'-'

	// 审批相关（团队转出时使用）
	IsOnlyOneMemberTeam bool // 默认值false:多人团队审批(必填)，true:单人团队自动批准
	// 审批人填写，必填
	IsApproved              bool      // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          int       // 审批人ID，团队核心成员id，多人团队不能是发起人id，（必填）单人团队自动批准时，审批人是发起人自己
	ApprovalRejectionReason string    // 审批意见，如果拒绝，填写原因,默认值:'-'
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
	Notes                   string  // 转账备注,默认值:'-'
	Status                  string  // 包含接收状态，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer    int64 // 转账后余额(毫克)

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
	AmountMilligrams             int64 // 转账星茶数量(毫克)
	Notes                   string  // 转账备注,默认值:'-'
	Status                  string  // 包含接收状态，待接收，已完成，已拒收，已过期等状态
	BalanceAfterTransfer    int64 // 转账后余额(毫克)

	// 接收方ToTeam成员操作，Confirmed/Rejected二选一
	IsConfirmed              bool   // 默认false，默认不接收，避免转账错误被误接收
	OperationalUserId        int    // 操作用户id，确认接收或者拒绝接收的用户id（团队成员）
	ReceptionRejectionReason string // 如果拒绝，填写原因,默认值:'-'

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
	FrozenReason            *string
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
	Notes            string // 转账备注，默认值:'-'

	// 待接收	StatusPendingReceipt   = "pending_receipt"
	// 已完成	StatusCompleted        = "completed"
	// 以拒收	StatusRejected         = "rejected"
	// 已超时	StatusExpired          = "expired"
	Status string // 系统填写
	//TransferType string 从表名获取

	//系统填写
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
	Notes            string // 转账备注,默认值:'-'

	// 待接收	StatusPendingReceipt   = "pending_receipt"
	// 已完成	StatusCompleted        = "completed"
	// 以拒收	StatusRejected         = "rejected"
	// 已超时	StatusExpired          = "expired"
	Status string // 系统填写
	//TransferType string 从表名获取

	//系统填写
	BalanceAfterTransfer int64 // 转出后，FromUser账户余额（毫克），对账单审计用

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
	Notes               string // 转出方备注（从转出表复制过来）,默认值:'-'
	BalanceAfterReceipt int64  // 接收后账户余额（毫克），对账单审计用

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
	Notes                   string // 转出方备注（从转出表复制过来）,默认值:'-'
	BalanceAfterReceipt     int64  // 接收后账户余额（毫克），对账单审计用

	// 已完成	StatusCompleted        = "completed"
	// 已拒收	StatusRejected         = "rejected"
	Status string //方便阅读，对账单审计用
	// 接收方ToUser操作，Confirmed/Rejected二选一
	IsConfirmed              bool      // 默认false，默认不接收，避免转账错误被误接收
	ReceptionRejectionReason *string   // 如果拒绝，填写原因
	ExpiresAt                time.Time // 过期时间，接收截止时间，也是FromTeam解锁额度时间
	CreatedAt                time.Time // 必填，如果接收，是清算时间；如果拒绝，是拒绝时间
}

// EnsureTeaUserAccountExists 确保用户有星茶账户
func EnsureTeaUserAccountExists(userId int) error {
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
