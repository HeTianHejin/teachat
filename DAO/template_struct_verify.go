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
	PayerTeam        *Team
	PayeeTeam        *Team
	CareTeam         *Team
	VerifyTeam       *Team  // 可选
	StatusLabelClass string // 根据状态返回Bootstrap标签类，如"warning"、"success"等
	CreatedDateTime  string // 格式化后的时间
}
