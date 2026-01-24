package dao

import "time"

type TeaOrderDeposit struct {
	Id               int
	Uuid             string
	TeaOrderId       int                 // 茶围订单ID
	Type             TeaOrderDepositType // 款项类型：1：约茶，2：探茶，3：看看，4：验茶，5：脑火,6：手工艺，7：其他
	PayerTeamId      int                 // 支付团队ID，款项来源方
	BankTeamId       int                 // 托管团队ID，款项托管方
	PayeeTeamId      int                 // 解题团队ID，款项最终接收方
	AmountMilligrams int64               // 托管星茶数量，以 毫克（0.001克） 为单位

	// 关联的转账记录（用于追踪星茶流向）
	TransferOutId int //  支付方→托管的转出记录,REFERENCES tea.team_to_team_transfer_out(id)
	TransferInId  int //  托管方的接收记录,REFERENCES tea.team_from_team_transfer_in(id)

	Status     TeaOrderDepositStatus
	Notes      string // 备注说明
	HasDispute bool   // 是否存在争议（快速查询标识）

	//时间节点
	ExpiredAt *time.Time // 支付过期时间，（超时自动取消）
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time // 软删除时间
}

// TeaOrderDepositStatus 托管状态枚举
type TeaOrderDepositStatus int

const (
	DepositStatusPendingPayment  TeaOrderDepositStatus = iota // 待支付处理
	DepositStatusPendingDeposit                               // 待托管处理
	DepositStatusPaid                                         // 已支付,托管方已接收成功
	DepositStatusReleasedToPayee                              // 已释放给解题方
	DepositStatusRefundedToPayer                              // 已退款给需求方
	DepositStatusDisputed                                     // 争议中(解题方与需求方争议,需要仲裁)
	DepositStatusCancelled                                    // 已取消(订单审批被撤销)
)

type TeaOrderDepositType int

// TeaOrderDepositType 款项类型枚举
const (
	DepositTypeTeaAppointment = iota + 1 // 约茶
	DepositTypeTeaExplore                // 探茶
	DepositTypeTeaSeeSeek                // 看看
	DepositTypeTeaExamine                // 验茶
	DepositTypeBrainFire                 // 脑火
	DepositTypeHandicraft                // 手工艺
	DepositTypeOther                     // 其他
)

// StatusString 返回托管状态的中文描述
func (tod *TeaOrderDeposit) StatusString() string {
	switch tod.Status {
	case DepositStatusPendingPayment:
		return "待支付处理"
	case DepositStatusPendingDeposit:
		return "待托管处理"
	case DepositStatusPaid:
		return "已支付"
	case DepositStatusReleasedToPayee:
		return "已释放给解题方"
	case DepositStatusRefundedToPayer:
		return "已退款给需求方"
	case DepositStatusDisputed:
		return "争议中"
	case DepositStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}
