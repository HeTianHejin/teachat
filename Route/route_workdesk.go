package route

import (
	"database/sql"
	"net/http"
	"strconv"
	dao "teachat/DAO"
	util "teachat/Util"
)

// HandleWorkDesk 处理团队工作台路由 —— 茶订单跟踪入口
func HandleWorkDesk(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		WorkDeskGet(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/work-desk?team_id=xxx
// 团队工作台，展示本团参与的茶订单简要列表，按"需求方/解题方"Tab区分
func WorkDeskGet(w http.ResponseWriter, r *http.Request) {
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

	teamUuid := r.URL.Query().Get("team_id")
	if teamUuid == "" {
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的团队。")
		return
	}

	// 查询团队
	team, err := dao.GetTeamByUUID(teamUuid)
	if err != nil {
		util.Debug("Cannot get team by uuid:", teamUuid, err)
		report(w, s_u, "你好，茶博士失魂鱼，未能找到指定的团队。请确认后再试。")
		return
	}

	// 检查当前用户是否为团队正常状态成员（限制非成员、非正常状态成员访问）
	isMember, err := team.IsActiveMember(s_u.Id)
	if err != nil {
		util.Debug("Cannot check team membership", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能验证团队身份。请稍后再试。")
		return
	}
	if !isMember {
		report(w, s_u, "你好，您不是该团队成员，无法查看工作台。")
		return
	}

	ctx := r.Context()
	pageSize := 20

	// 获取当前页
	page := 0
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}

	// 确定当前激活的Tab（默认"payer"需求方）
	activeTab := r.URL.Query().Get("tab")
	if activeTab != "payer" && activeTab != "payee" {
		activeTab = "payer"
	}

	// 获取团队作为需求方（出题方）的茶订单
	payerOrders, err := dao.GetTeaOrdersByPayerTeamId(ctx, team.Id, page, pageSize)
	if err != nil && err != sql.ErrNoRows {
		util.Debug("Cannot get payer tea orders", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取需求方订单。请稍后再试。")
		return
	}

	// 获取团队作为解题方的茶订单
	payeeOrders, err := dao.GetTeaOrdersByPayeeTeamId(ctx, team.Id, page, pageSize)
	if err != nil && err != sql.ErrNoRows {
		util.Debug("Cannot get payee tea orders", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取解题方订单。请稍后再试。")
		return
	}

	// 获取各角色订单数量
	payerCount, err := dao.GetTeaOrderCountByPayerTeamId(ctx, team.Id)
	if err != nil {
		util.Debug("Cannot get payer tea order count", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取需求方订单数量。请稍后再试。")
		return
	}

	payeeCount, err := dao.GetTeaOrderCountByPayeeTeamId(ctx, team.Id)
	if err != nil {
		util.Debug("Cannot get payee tea order count", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能获取解题方订单数量。请稍后再试。")
		return
	}

	// 转换为TeaOrderBean
	payerOrderBeans, err := fetchTeaOrderBeanSlice(payerOrders)
	if err != nil {
		util.Debug("Cannot convert payer orders to beans", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备需求方订单数据。请稍后再试。")
		return
	}

	payeeOrderBeans, err := fetchTeaOrderBeanSlice(payeeOrders)
	if err != nil {
		util.Debug("Cannot convert payee orders to beans", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能准备解题方订单数据。请稍后再试。")
		return
	}

	// 检查是否为见证者
	isVerifier := dao.IsVerifier(s_u.Id)

	// 准备页面数据
	type WorkDeskPageData struct {
		Team            dao.Team
		SessUser        dao.User
		PayerOrders     []*dao.TeaOrderBean
		PayeeOrders     []*dao.TeaOrderBean
		PayerOrderCount int
		PayeeOrderCount int
		ActiveTab       string
		IsVerifier      bool
		Page            int
	}

	pageData := WorkDeskPageData{
		Team:            team,
		SessUser:        s_u,
		PayerOrders:     payerOrderBeans,
		PayeeOrders:     payeeOrderBeans,
		PayerOrderCount: payerCount,
		PayeeOrderCount: payeeCount,
		ActiveTab:       activeTab,
		IsVerifier:      isVerifier,
		Page:            page,
	}

	// 渲染工作台页面
	generateHTML(w, &pageData, "layout", "navbar.private", "workdesk", "component_sess_capacity")
}

// HandleTeaOrderDetail 处理茶订单详情路由（团队工作台入口，见证者/参与团队均可查看）
func HandleTeaOrderDetail(w http.ResponseWriter, r *http.Request) {
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

	// 权限检查：见证者 OR 参与团队（需求方/解题方/监护方）成员均可查看
	isVerifier := dao.IsVerifier(s_u.Id)
	hasAccess := isVerifier

	if !hasAccess {
		// 检查是否为需求方团队成员
		if teaOrder.PayerTeamId > 0 {
			payerTeam := &dao.Team{Id: teaOrder.PayerTeamId}
			if err := payerTeam.Get(); err == nil {
				if isMember, _ := payerTeam.IsActiveMember(s_u.Id); isMember {
					hasAccess = true
				}
			}
		}
	}
	if !hasAccess {
		// 检查是否为解题方团队成员
		if teaOrder.PayeeTeamId > 0 {
			payeeTeam := &dao.Team{Id: teaOrder.PayeeTeamId}
			if err := payeeTeam.Get(); err == nil {
				if isMember, _ := payeeTeam.IsActiveMember(s_u.Id); isMember {
					hasAccess = true
				}
			}
		}
	}
	if !hasAccess {
		// 检查是否为监护方团队成员
		if teaOrder.CareTeamId > 0 {
			careTeam := &dao.Team{Id: teaOrder.CareTeamId}
			if err := careTeam.Get(); err == nil {
				if isMember, _ := careTeam.IsActiveMember(s_u.Id); isMember {
					hasAccess = true
				}
			}
		}
	}

	if !hasAccess {
		report(w, s_u, "你好，您没有权限查看该茶订单详情。")
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

	// 获取来源标识（从哪个页面跳转来的）
	referer := r.URL.Query().Get("from")
	teamUuid := r.URL.Query().Get("team_id")

	// 准备页面数据
	type DetailPageData struct {
		SessUser     dao.User
		TeaOrderBean *dao.TeaOrderBean
		WitnessLogs  []*dao.WitnessLog
		IsVerifier   bool
		Referer      string // "workdesk" 或 "verifier" 或空
		TeamUuid     string // 若从工作台跳转，带回团队UUID用于返回链接
	}
	pageData := DetailPageData{
		SessUser:     s_u,
		TeaOrderBean: teaOrderBean,
		WitnessLogs:  witnessLogs,
		IsVerifier:   isVerifier,
		Referer:      referer,
		TeamUuid:     teamUuid,
	}

	// 渲染详情页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea-order.detail", "component_tea_order_bean", "component_sess_capacity")
}
