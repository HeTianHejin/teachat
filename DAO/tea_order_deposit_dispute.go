package dao

import "time"

// TeaOrderDepositDispute 托管争议表
type TeaOrderDepositDispute struct {
	Id               int
	Uuid             string
	DepositId        int           // 关联的托管记录ID, REFERENCES tea_order_deposit(id)
	InitiatorTeamId  int           // 发起争议方团队ID
	RespondentTeamId int           // 被诉方团队ID
	Reason           string        // 争议原因
	Status           DisputeStatus // 争议状态

	// 仲裁信息
	ArbitratorTeamId *int          // 仲裁方团队ID,未仲裁时为 nil
	ArbitrationNotes string        // 仲裁说明
	Result           DisputeResult // 仲裁结果

	// 时间节点
	CreatedAt    time.Time  // 争议发起时间
	ArbitratedAt *time.Time // 仲裁完成时间
	UpdatedAt    *time.Time
	DeletedAt    *time.Time // 软删除
}

// DisputeStatus 争议状态枚举
type DisputeStatus int

const (
	DisputeStatusPending    DisputeStatus = iota + 1 // 争议中
	DisputeStatusArbitrated                          // 已仲裁
	DisputeStatusWithdrawn                           // 已撤销
)

// DisputeResult 仲裁结果枚举
type DisputeResult int

const (
	DisputeResultNone     DisputeResult = iota // 未仲裁
	DisputeResultPayeeWin                      // 解题方胜出
	DisputeResultPayerWin                      // 需求方胜出
)

// StatusString 返回争议状态的中文描述
func (d *TeaOrderDepositDispute) StatusString() string {
	switch d.Status {
	case DisputeStatusPending:
		return "争议中"
	case DisputeStatusArbitrated:
		return "已仲裁"
	case DisputeStatusWithdrawn:
		return "已撤销"
	default:
		return "未知状态"
	}
}

// ResultString 返回仲裁结果的中文描述
func (d *TeaOrderDepositDispute) ResultString() string {
	switch d.Result {
	case DisputeResultNone:
		return "未仲裁"
	case DisputeResultPayeeWin:
		return "解题方胜出"
	case DisputeResultPayerWin:
		return "需求方胜出"
	default:
		return "未知结果"
	}
}
