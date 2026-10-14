package route

import (
	"net/http"
	"strconv"
	dao "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/workbook/user?uuid=
func WorkbookUser(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Error("cannot get user from session %d: %v", sess.Id, err)
		report(w, s_u, "你好，世人都晓神仙好，只有金银忘不了！请稍后再试。")
		return
	}
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		// 如果uuid参数为空，则返回错误信息
		http.Error(w, "uuid is required", http.StatusBadRequest)
		return
	}

	var tUser dao.User
	if uuid == s_u.Uuid {
		// 翻阅自己的作业本
		// 标记是本人， 如pageData.IsAuthor = true
		tUser = s_u
	} else {
		//pageData.IsAuthor = false
		// 根据uuid获取目标茶友资料
		tUser, err = dao.GetUserByUuidOrId(uuid)
		if err != nil {
			util.Debug("cannot get user by uuid %s: %v", uuid, err)
			report(w, s_u, "你好，找不到提及的茶友资料，请确认后再试。")
			return
		}
	}

	// 读取目标茶友的历史作业记录
	// 订单案例归个人：订单完成时，以订单成员快照记录实际参与者、承担角色和当时关联的技能。案例保留订单发生时的团队、服务版本等背景，但不以团队当前状态作为展示前提。
	// 团队只作历史背景：团队可以关联或展示成员的技能，但不拥有个人技能档案。团队解散后，历史案例仍可显示“当时以某团队成员身份参与”。
	// 作业本汇总事实，不授予认证：集中展示案例数、同行评价、需求方满意度及争议处理结果；明确它们是参与者提供的记录，不代表平台认证或能力保证。未完成、取消或仍在争议处理中的订单不计入已完成案例。
	page := 1
	if requestedPage, parseErr := strconv.Atoi(r.URL.Query().Get("page")); parseErr == nil && requestedPage > 0 {
		page = requestedPage
	}
	const pageSize = 20
	totalCount, err := dao.GetCompletedTeaOrderSummaryCountByUserId(r.Context(), tUser.Id)
	if err != nil {
		util.Debug("cannot count completed orders for user %d: %v", tUser.Id, err)
		report(w, s_u, "你好，未能读取这位茶友的作业本，请稍后再试。")
		return
	}
	totalPages := (totalCount + pageSize - 1) / pageSize
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}
	orders, err := dao.GetCompletedTeaOrderSummariesByUserId(r.Context(), tUser.Id, page, pageSize)
	if err != nil {
		util.Debug("cannot read completed orders for user %d: %v", tUser.Id, err)
		report(w, s_u, "你好，未能读取这位茶友的作业本，请稍后再试。")
		return
	}

	type WorkbookPageData struct {
		SessUser    dao.User
		TargetUser  dao.User
		Orders      []dao.CompletedTeaOrderSummary
		TotalCount  int
		CurrentPage int
		TotalPages  int
		HasPrev     bool
		HasNext     bool
	}
	pageData := WorkbookPageData{
		SessUser: s_u, TargetUser: tUser, Orders: orders, TotalCount: totalCount,
		CurrentPage: page, TotalPages: totalPages,
		HasPrev: page > 1, HasNext: page < totalPages,
	}
	generateHTML(w, &pageData, "layout", "navbar.private", "workbook.user")

}
