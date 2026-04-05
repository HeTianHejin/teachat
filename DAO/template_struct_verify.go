package dao

type VerifierWorkspagePageData struct {
	SessUser            User
	PendingOrders       []*TeaOrderBean
	ActiveOrders        []*TeaOrderBean
	CancelledOrders     []*TeaOrderBean
	CompletedOrders     []*TeaOrderBean
	PendingOrderCount   int
	ActiveOrderCount    int
	PauseOrderCount     int
	CancelledOrderCount int
	CompletedOrderCount int
}
type TeaOrderBean struct {
	TeaOrder         *TeaOrder
	ObjectiveBean    *ObjectiveBean
	ProjectBean      *ProjectBean
	PayerTeam        *Team  // 需求方团队
	PayeeTeam        *Team  // 解题方团队
	CareTeam         *Team  // 监护方团队
	VerifyTeam       *Team  // 见证方团队
	OperatorUser     *User  // 入围操作人
	ApproverUser     *User  // 见证批准人
	StatusLabelClass string // 根据状态返回Bootstrap标签类，如"warning"、"success"等
	CreatedDateTime  string // 格式化后的时间
}
