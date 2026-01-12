package route

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	dao "teachat/DAO"
	util "teachat/Util"
)

// HandleVerifierWorkspace 处理见证者工作间路由
func HandleVerifierWorkspace(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		VerifierWorkspaceGet(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/verifier/workspace
// 见证者工作间页面，展示所有茶订单，按状态分类
func VerifierWorkspaceGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限访问见证者工作间。")
		return
	}

	ctx := r.Context()

	// 获取各状态的茶订单数量
	pendingCount, err := dao.GetTeaOrderCountByStatus(ctx, dao.TeaOrderStatusPending)
	if err != nil {
		util.Debug("Cannot get pending tea order count", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取待审批订单数量。请稍后再试。")
		return
	}

	activeCount, err := dao.GetTeaOrderCountByStatus(ctx, dao.TeaOrderStatusActive)
	if err != nil {
		util.Debug("Cannot get active tea order count", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取进行中订单数量。请稍后再试。")
		return
	}

	pauseCount, err := dao.GetTeaOrderCountByStatus(ctx, dao.TeaOrderStatusPause)
	if err != nil {
		util.Debug("Cannot get pause tea order count", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取暂停订单数量。请稍后再试。")
		return
	}

	cancelledCount, err := dao.GetTeaOrderCountByStatus(ctx, dao.TeaOrderStatusCancelled)
	if err != nil {
		util.Debug("Cannot get cancelled tea order count", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取已取消订单数量。请稍后再试。")
		return
	}

	completedCount, err := dao.GetTeaOrderCountByStatus(ctx, dao.TeaOrderStatusCompleted)
	if err != nil {
		util.Debug("Cannot get completed tea order count", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取已完成订单数量。请稍后再试。")
		return
	}

	// 获取各状态的茶订单列表（每页20条）
	pendingOrders, err := dao.GetTeaOrdersByStatus(ctx, dao.TeaOrderStatusPending, 0, 20)
	if err != nil && err != sql.ErrNoRows {
		util.Debug("Cannot get pending tea orders", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取待审批订单。请稍后再试。")
		return
	}

	activeOrders, err := dao.GetTeaOrdersByStatus(ctx, dao.TeaOrderStatusActive, 0, 20)
	if err != nil && err != sql.ErrNoRows {
		util.Debug("Cannot get active tea orders", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取进行中订单。请稍后再试。")
		return
	}

	cancelledOrders, err := dao.GetTeaOrdersByStatus(ctx, dao.TeaOrderStatusCancelled, 0, 20)
	if err != nil && err != sql.ErrNoRows {
		util.Debug("Cannot get cancelled tea orders", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取异常订单。请稍后再试。")
		return
	}

	completedOrders, err := dao.GetTeaOrdersByStatus(ctx, dao.TeaOrderStatusCompleted, 0, 20)
	if err != nil && err != sql.ErrNoRows {
		util.Debug("Cannot get completed tea orders", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取已完成订单。请稍后再试。")
		return
	}

	// 转换为TeaOrderBean
	pendingOrderBeans, err := fetchTeaOrderBeanSlice(pendingOrders)
	if err != nil {
		util.Debug("Cannot convert pending orders to beans", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备待审批订单数据。请稍后再试。")
		return
	}

	activeOrderBeans, err := fetchTeaOrderBeanSlice(activeOrders)
	if err != nil {
		util.Debug("Cannot convert active orders to beans", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备进行中订单数据。请稍后再试。")
		return
	}

	cancelledOrderBeans, err := fetchTeaOrderBeanSlice(cancelledOrders)
	if err != nil {
		util.Debug("Cannot convert cancelled orders to beans", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备异常订单数据。请稍后再试。")
		return
	}

	completedOrderBeans, err := fetchTeaOrderBeanSlice(completedOrders)
	if err != nil {
		util.Debug("Cannot convert completed orders to beans", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备已完成订单数据。请稍后再试。")
		return
	}

	// 准备页面数据
	pageData := dao.VerifierWorkspagePageData{
		SessUser:            s_u,
		PendingOrders:       pendingOrderBeans,
		ActiveOrders:        activeOrderBeans,
		CancelledOrders:     cancelledOrderBeans,
		CompletedOrders:     completedOrderBeans,
		PendingOrderCount:   pendingCount,
		ActiveOrderCount:    activeCount,
		PauseOrderCount:     pauseCount,
		CancelledOrderCount: cancelledCount,
		CompletedOrderCount: completedCount,
	}

	// 渲染页面
	generateHTML(w, &pageData, "layout", "navbar.private", "verifier.workspace", "component_tea_order_bean", "component_sess_capacity")
}

// HandleVerifierOrderApprove 处理审批茶订单路由
func HandleVerifierOrderApprove(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		VerifierOrderApproveGet(w, r)
	case http.MethodPost:
		VerifierOrderApprovePost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/verifier/order/approve?uuid=xxx
// 审批茶订单表单页面
func VerifierOrderApproveGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限执行审批操作。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 检查订单状态
	if teaOrder.Status != dao.TeaOrderStatusPending {
		report(w, s_u, "你好，该订单状态不允许审批操作。")
		return
	}

	// 获取茶订单Bean
	teaOrderBean, err := fetchTeaOrderBean(*teaOrder)
	if err != nil {
		util.Debug("Cannot convert tea order to bean", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备茶订单数据。请稍后再试。")
		return
	}

	// 准备页面数据
	type ApprovePageData struct {
		SessUser        dao.User
		TeaOrderBean    *dao.TeaOrderBean
	}
	pageData := ApprovePageData{
		SessUser:     s_u,
		TeaOrderBean: teaOrderBean,
	}

	// 渲染审批表单页面
	generateHTML(w, &pageData, "layout", "navbar.private", "verifier.order.approve", "component_tea_order_bean", "component_sess_capacity")
}

// POST /v1/verifier/order/approve
// 处理审批茶订单
func VerifierOrderApprovePost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限执行审批操作。")
		return
	}

	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	teaTopic := r.PostFormValue("tea_topic")
	if teaTopic == "" {
		report(w, s_u, "你好，茶博士失魂鱼，请填写茶会主题。")
		return
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 检查订单状态
	if teaOrder.Status != dao.TeaOrderStatusPending {
		report(w, s_u, "你好，该订单状态不允许审批操作。")
		return
	}

	// 更新订单状态
	teaOrder.TeaTopic = teaTopic
	teaOrder.IsApproved = true
	teaOrder.ApproverUserId = s_u.Id
	teaOrder.ApprovalRejectionReason = "-"
	approvalTime := time.Now()
	teaOrder.ApprovedAt = &approvalTime
	teaOrder.Status = dao.TeaOrderStatusActive

	if err = teaOrder.Update(); err != nil {
		util.Debug("Cannot update tea order", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能更新茶订单。请稍后再试。")
		return
	}

	// 创建见证日志
	witnessLog := &dao.WitnessLog{
		Uuid:       dao.Random_UUID(),
		TeaOrderId: teaOrder.Id,
		Action:     dao.WitnessActionApprove,
		Reason:     "茶订单已审批通过",
		EvidenceId: 0,
		WitnessAt:  approvalTime,
	}
	if err = witnessLog.Create(r.Context()); err != nil {
		util.Debug("Cannot create witness log", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建见证日志。请稍后再试。")
		return
	}

	// 重定向到工作间
	http.Redirect(w, r, "/v1/verifier/workspace", http.StatusFound)
}

// HandleVerifierOrderReject 处理拒绝茶订单路由
func HandleVerifierOrderReject(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		VerifierOrderRejectGet(w, r)
	case http.MethodPost:
		VerifierOrderRejectPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/verifier/order/reject?uuid=xxx
// 拒绝茶订单表单页面
func VerifierOrderRejectGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限执行拒绝操作。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 检查订单状态
	if teaOrder.Status != dao.TeaOrderStatusPending {
		report(w, s_u, "你好，该订单状态不允许拒绝操作。")
		return
	}

	// 获取茶订单Bean
	teaOrderBean, err := fetchTeaOrderBean(*teaOrder)
	if err != nil {
		util.Debug("Cannot convert tea order to bean", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备茶订单数据。请稍后再试。")
		return
	}

	// 准备页面数据
	type RejectPageData struct {
		SessUser        dao.User
		TeaOrderBean    *dao.TeaOrderBean
	}
	pageData := RejectPageData{
		SessUser:     s_u,
		TeaOrderBean: teaOrderBean,
	}

	// 渲染拒绝表单页面
	generateHTML(w, &pageData, "layout", "navbar.private", "verifier.order.reject", "component_tea_order_bean", "component_sess_capacity")
}

// POST /v1/verifier/order/reject
// 处理拒绝茶订单
func VerifierOrderRejectPost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限执行拒绝操作。")
		return
	}

	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	reason := r.PostFormValue("reason")
	if reason == "" {
		report(w, s_u, "你好，茶博士失魂鱼，请填写拒绝原因。")
		return
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 检查订单状态
	if teaOrder.Status != dao.TeaOrderStatusPending {
		report(w, s_u, "你好，该订单状态不允许拒绝操作。")
		return
	}

	// 更新订单状态
	teaOrder.IsApproved = false
	teaOrder.ApproverUserId = s_u.Id
	teaOrder.ApprovalRejectionReason = reason
	teaOrder.Status = dao.TeaOrderStatusCancelled

	if err = teaOrder.Update(); err != nil {
		util.Debug("Cannot update tea order", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能更新茶订单。请稍后再试。")
		return
	}

	// 创建见证日志
	witnessLog := &dao.WitnessLog{
		Uuid:       dao.Random_UUID(),
		TeaOrderId: teaOrder.Id,
		Action:     dao.WitnessActionCancel,
		Reason:     fmt.Sprintf("拒绝原因：%s", reason),
		EvidenceId: 0,
		WitnessAt:  time.Now(),
	}
	if err = witnessLog.Create(r.Context()); err != nil {
		util.Debug("Cannot create witness log", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建见证日志。请稍后再试。")
		return
	}

	// 重定向到工作间
	http.Redirect(w, r, "/v1/verifier/workspace", http.StatusFound)
}

// HandleVerifierOrderPause 处理暂停茶订单路由
func HandleVerifierOrderPause(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		VerifierOrderPauseGet(w, r)
	case http.MethodPost:
		VerifierOrderPausePost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/verifier/order/pause?uuid=xxx
// 暂停茶订单表单页面
func VerifierOrderPauseGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限执行暂停操作。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 检查订单状态
	if teaOrder.Status != dao.TeaOrderStatusActive {
		report(w, s_u, "你好，该订单状态不允许暂停操作。")
		return
	}

	// 获取茶订单Bean
	teaOrderBean, err := fetchTeaOrderBean(*teaOrder)
	if err != nil {
		util.Debug("Cannot convert tea order to bean", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备茶订单数据。请稍后再试。")
		return
	}

	// 准备页面数据
	type PausePageData struct {
		SessUser        dao.User
		TeaOrderBean    *dao.TeaOrderBean
	}
	pageData := PausePageData{
		SessUser:     s_u,
		TeaOrderBean: teaOrderBean,
	}

	// 渲染暂停表单页面
	generateHTML(w, &pageData, "layout", "navbar.private", "verifier.order.pause", "component_tea_order_bean", "component_sess_capacity")
}

// POST /v1/verifier/order/pause
// 处理暂停茶订单
func VerifierOrderPausePost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限执行暂停操作。")
		return
	}

	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	reason := r.PostFormValue("reason")
	if reason == "" {
		reason = "-"
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 检查订单状态
	if teaOrder.Status != dao.TeaOrderStatusActive {
		report(w, s_u, "你好，该订单状态不允许暂停操作。")
		return
	}

	// 更新订单状态
	teaOrder.Status = dao.TeaOrderStatusPause

	if err = teaOrder.Update(); err != nil {
		util.Debug("Cannot update tea order", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能更新茶订单。请稍后再试。")
		return
	}

	// 创建见证日志
	witnessLog := &dao.WitnessLog{
		Uuid:       dao.Random_UUID(),
		TeaOrderId: teaOrder.Id,
		Action:     dao.WitnessActionPause,
		Reason:     reason,
		EvidenceId: 0,
		WitnessAt:  time.Now(),
	}
	if err = witnessLog.Create(r.Context()); err != nil {
		util.Debug("Cannot create witness log", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建见证日志。请稍后再试。")
		return
	}

	// 重定向到工作间
	http.Redirect(w, r, "/v1/verifier/workspace", http.StatusFound)
}

// HandleVerifierOrderCancel 处理终止茶订单路由
func HandleVerifierOrderCancel(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		VerifierOrderCancelGet(w, r)
	case http.MethodPost:
		VerifierOrderCancelPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/verifier/order/cancel?uuid=xxx
// 终止茶订单表单页面
func VerifierOrderCancelGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限执行终止操作。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 检查订单状态
	if teaOrder.Status != dao.TeaOrderStatusActive && teaOrder.Status != dao.TeaOrderStatusPause {
		report(w, s_u, "你好，该订单状态不允许终止操作。")
		return
	}

	// 获取茶订单Bean
	teaOrderBean, err := fetchTeaOrderBean(*teaOrder)
	if err != nil {
		util.Debug("Cannot convert tea order to bean", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备茶订单数据。请稍后再试。")
		return
	}

	// 准备页面数据
	type CancelPageData struct {
		SessUser        dao.User
		TeaOrderBean    *dao.TeaOrderBean
	}
	pageData := CancelPageData{
		SessUser:     s_u,
		TeaOrderBean: teaOrderBean,
	}

	// 渲染终止表单页面
	generateHTML(w, &pageData, "layout", "navbar.private", "verifier.order.cancel", "component_tea_order_bean", "component_sess_capacity")
}

// POST /v1/verifier/order/cancel
// 处理终止茶订单
func VerifierOrderCancelPost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限执行终止操作。")
		return
	}

	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	uuid := r.PostFormValue("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	reason := r.PostFormValue("reason")
	if reason == "" {
		reason = "-"
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 检查订单状态
	if teaOrder.Status != dao.TeaOrderStatusActive && teaOrder.Status != dao.TeaOrderStatusPause {
		report(w, s_u, "你好，该订单状态不允许终止操作。")
		return
	}

	// 更新订单状态
	teaOrder.Status = dao.TeaOrderStatusCancelled

	if err = teaOrder.Update(); err != nil {
		util.Debug("Cannot update tea order", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能更新茶订单。请稍后再试。")
		return
	}

	// 创建见证日志
	witnessLog := &dao.WitnessLog{
		Uuid:       dao.Random_UUID(),
		TeaOrderId: teaOrder.Id,
		Action:     dao.WitnessActionCancel,
		Reason:     reason,
		EvidenceId: 0,
		WitnessAt:  time.Now(),
	}
	if err = witnessLog.Create(r.Context()); err != nil {
		util.Debug("Cannot create witness log", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能创建见证日志。请稍后再试。")
		return
	}

	// 重定向到工作间
	http.Redirect(w, r, "/v1/verifier/workspace", http.StatusFound)
}

// HandleVerifierOrderDetail 处理查看茶订单详情路由
func HandleVerifierOrderDetail(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}

	// 检查用户是否为见证者
	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你好，您没有权限查看茶订单详情。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。")
		return
	}

	// 获取茶订单
	teaOrder := &dao.TeaOrder{Uuid: uuid}
	if err = teaOrder.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get tea order", uuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的茶订单。请确认后再试。")
		return
	}

	// 获取茶订单Bean
	teaOrderBean, err := fetchTeaOrderBean(*teaOrder)
	if err != nil {
		util.Debug("Cannot convert tea order to bean", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备茶订单数据。请稍后再试。")
		return
	}

	// 获取见证日志
	witnessLog := &dao.WitnessLog{TeaOrderId: teaOrder.Id}
	witnessLogs, err := witnessLog.GetByTeaOrderId(r.Context())
	if err != nil && err != sql.ErrNoRows {
		util.Debug("Cannot get witness logs", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取见证日志。请稍后再试。")
		return
	}

	// 准备页面数据
	type DetailPageData struct {
		SessUser        dao.User
		TeaOrderBean    *dao.TeaOrderBean
		WitnessLogs     []*dao.WitnessLog
	}
	pageData := DetailPageData{
		SessUser:     s_u,
		TeaOrderBean: teaOrderBean,
		WitnessLogs:  witnessLogs,
	}

	// 渲染详情页面
	generateHTML(w, &pageData, "layout", "navbar.private", "verifier.order.detail", "component_tea_order_bean", "component_sess_capacity")
}

// fetchTeaOrderBean 根据茶订单获取对应的Bean数据
func fetchTeaOrderBean(teaOrder dao.TeaOrder) (*dao.TeaOrderBean, error) {
	bean := &dao.TeaOrderBean{
		TeaOrder:        &teaOrder,
		CreatedDateTime: teaOrder.CreatedDateTime(),
		StatusLabelClass: getStatusLabelClass(teaOrder.Status),
	}

	// 获取茶围
	objective := dao.Objective{Id: teaOrder.ObjectiveId}
	if err := objective.Get(); err != nil {
		return nil, err
	}
	bean.ObjectiveBean = &dao.ObjectiveBean{
		Objective: objective,
	}

	// 获取茶台
	project := dao.Project{Id: teaOrder.ProjectId}
	if err := project.Get(); err != nil {
		return nil, err
	}
	bean.ProjectBean = &dao.ProjectBean{}

	// 获取需求方团队
	if teaOrder.PayerTeamId > 0 {
		payerTeam, err := dao.GetTeam(teaOrder.PayerTeamId)
		if err != nil {
			return nil, err
		}
		bean.PayerTeam = &payerTeam
	} else {
		bean.PayerTeam = &dao.Team{Name: "未指定"}
	}

	// 获取解题方团队
	if teaOrder.PayeeTeamId > 0 {
		payeeTeam, err := dao.GetTeam(teaOrder.PayeeTeamId)
		if err != nil {
			return nil, err
		}
		bean.PayeeTeam = &payeeTeam
	} else {
		bean.PayeeTeam = &dao.Team{Name: "未指定"}
	}

	// 获取监护方团队
	if teaOrder.CareTeamId > 0 {
		careTeam, err := dao.GetTeam(teaOrder.CareTeamId)
		if err != nil {
			return nil, err
		}
		bean.CareTeam = &careTeam
	} else {
		bean.CareTeam = &dao.Team{Name: "未指定"}
	}

	// 获取见证方团队（可选）
	if teaOrder.VerifyTeamId > 0 {
		verifyTeam, err := dao.GetTeam(teaOrder.VerifyTeamId)
		if err != nil {
			return nil, err
		}
		bean.VerifyTeam = &verifyTeam
	}

	return bean, nil
}

// fetchTeaOrderBeanSlice 批量获取茶订单Bean数据
func fetchTeaOrderBeanSlice(teaOrders []*dao.TeaOrder) ([]*dao.TeaOrderBean, error) {
	beans := make([]*dao.TeaOrderBean, 0, len(teaOrders))
	for _, order := range teaOrders {
		bean, err := fetchTeaOrderBean(*order)
		if err != nil {
			return nil, err
		}
		beans = append(beans, bean)
	}
	return beans, nil
}

// getStatusLabelClass 根据状态返回Bootstrap标签类
func getStatusLabelClass(status string) string {
	switch status {
	case dao.TeaOrderStatusPending:
		return "warning"
	case dao.TeaOrderStatusActive:
		return "success"
	case dao.TeaOrderStatusPause:
		return "info"
	case dao.TeaOrderStatusCompleted:
		return "primary"
	case dao.TeaOrderStatusCancelled:
		return "danger"
	default:
		return "default"
	}
}
