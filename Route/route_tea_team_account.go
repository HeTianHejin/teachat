package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	dao "teachat/DAO"
	util "teachat/Util"
	"time"
)

// 团队星茶账户相关响应结构体
type TeamTeaAccountResponse struct {
	Uuid              string `json:"uuid"`
	TeamId            int    `json:"team_id"`
	TeamName          string `json:"team_name,omitempty"`
	BalanceMilligrams int64  `json:"balance_milligrams"`
	Status            string `json:"status"`
	FrozenReason      string `json:"frozen_reason,omitempty"`
	CreatedAt         string `json:"created_at"`
}

// 团队星茶账户解冻请求结构体
type UnfreezeTeamAccountRequest struct {
	TeamId int `json:"team_id"`
}

// TeamAccountWithAvailable 用于模板渲染，包含可用余额
type TeamAccountWithAvailable struct {
	dao.TeaTeamAccount
	AvailableBalanceMilligrams int64
}

// PageData 用于模板渲染
type PageData struct {
	SessUser                 dao.User
	Team                     *dao.Team
	TeamAccount              TeamAccountWithAvailable
	TransactionHistory       []map[string]any
	UserIsCoreMember         bool
	PendingIncomingTeamCount int
	PendingIncomingUserCount int
	PendingApprovalCount     int
}

// GetTeaTeamAccountAPI 获取团队星茶账户信息
func GetTeaTeamAccountAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := session(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "未登录")
		return
	}

	user, err := sess.User()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 必须指定团队ID
	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看星茶账户")
		return
	}

	// 确保团队有星茶账户
	err = dao.EnsureTeaTeamAccountExists(teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队星茶账户失败")
		return
	}

	// 获取团队星茶账户
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "团队星茶账户不存在")
		return
	}

	// 获取团队名称
	var teamName string
	if team, err := dao.GetTeam(teamId); err == nil {
		teamName = team.Name
	}

	response := TeamTeaAccountResponse{
		Uuid:              account.Uuid,
		TeamId:            account.TeamId,
		TeamName:          teamName,
		BalanceMilligrams: account.BalanceMilligrams,
		Status:            account.Status,
		FrozenReason:      account.FrozenReason,
		CreatedAt:         account.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	respondWithSuccess(w, "获取团队星茶账户成功", response)
}

// GetTeaTeamToTeamCompletedTransferOutsAPI 获取团队对团队转账已完成状态记录
func GetTeaTeamToTeamCompletedTransferOutsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := session(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "未登录")
		return
	}

	user, err := sess.User()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看团队对团队已完成状态转账纪录")
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 获取团队对团队已经完成状态交易记录
	transactions, err := dao.GetTeaTeamToTeamCompletedTransferOuts(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队对团队已完成状态转账纪录失败")
		return
	}

	respondWithSuccess(w, "获取团队对团队已完成状态转账纪录成功", transactions)
}

// FreezeTeaTeamAccountAPI 冻结团队星茶账户
func FreezeTeaTeamAccountAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := session(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "未登录")
		return
	}

	user, err := sess.User()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	type Request struct {
		TeamId int    `json:"team_id"`
		Reason string `json:"reason"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	if req.Reason == "" {
		respondWithError(w, http.StatusBadRequest, "冻结原因不能为空")
		return
	}

	// 检查用户权限
	canFreeze, err := dao.IsTeamActiveMember(user.Id, dao.TeamIdSpaceshipCrew)
	if err != nil || !canFreeze {
		respondWithError(w, http.StatusForbidden, "您没有权限冻结团队账户")
		return
	}

	// 获取账户并冻结
	account, err := dao.GetTeaTeamAccountByTeamId(req.TeamId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "团队账户不存在")
		return
	}

	err = account.UpdateStatus(dao.TeaTeamAccountStatus_Frozen, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "冻结账户失败")
		return
	}

	respondWithSuccess(w, "团队账户冻结成功", nil)
}

// UnfreezeTeaTeamAccountAPI 解冻团队星茶账户
func UnfreezeTeaTeamAccountAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := session(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "未登录")
		return
	}

	user, err := sess.User()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	var req UnfreezeTeamAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户权限
	canFreeze, err := dao.IsTeamActiveMember(user.Id, dao.TeamIdSpaceshipCrew)
	if err != nil || !canFreeze {
		respondWithError(w, http.StatusForbidden, "您没有权限冻结团队账户")
		return
	}

	// 获取账户并解冻
	account, err := dao.GetTeaTeamAccountByTeamId(req.TeamId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "团队账户不存在")
		return
	}

	err = account.UpdateStatus(dao.TeaTeamAccountStatus_Normal, "")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "解冻账户失败")
		return
	}

	respondWithSuccess(w, "团队账户解冻成功", nil)
}

// HandleTeaTeamAccount 处理团队星茶账户页面请求
func HandleTeaTeamAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TeaTeamAccountGet(w, r)
}

// TeaTeamAccountGet 获取团队星茶账户页面
func TeaTeamAccountGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 必须指定团队ID
	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		report(w, s_u, "必须指定团队ID。")
		return
	}

	// 显示指定团队的星茶账户
	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, s_u, "团队ID无效。")
		return
	}

	// 获取指定团队信息
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, s_u, "团队不存在。")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		report(w, s_u, "您不是该团队成员，无法查看星茶账户。")
		return
	}

	singleTeam := &team

	// 确保团队有星茶账户
	err = dao.EnsureTeaTeamAccountExists(team.Id)
	if err != nil {
		util.Debug("cannot ensure team tea account exists", err)
		report(w, s_u, "获取团队星茶账户失败。")
		return
	}

	// 获取团队星茶账户
	teamAccount, err := dao.GetTeaTeamAccountByTeamId(team.Id)
	if err != nil {
		util.Debug("cannot get team tea account", err)
		report(w, s_u, "获取团队星茶账户失败。")
		return
	}

	// 计算可用余额
	//TODO： 团队转账直接扣减余额，无需计算锁定余额
	teamAccountWithAvailable := TeamAccountWithAvailable{
		TeaTeamAccount:             teamAccount,
		AvailableBalanceMilligrams: teamAccount.BalanceMilligrams - teamAccount.LockedBalanceMilligrams,
	}

	// 获取待确认接收来自团队转账操作数量
	pendingIncomingCount, err := dao.CountPendingTeamReceipts(team.Id)
	if err != nil {
		util.Debug("cannot get pending incoming transfers count", err)
		pendingIncomingCount = 0
	}
	//TODO： 获取待确认接收来自用户转账操作数量
	pendingIncomingUserCount, err := dao.CountPendingUserReceipts(team.Id)
	if err != nil {
		util.Debug("cannot get pending incoming user transfers count", err)
		pendingIncomingUserCount = 0
	}

	// 获取帐户转出，待审批操作数量
	pendingApprovalCount, err := dao.CountPendingTeamApprovals(team.Id)
	if err != nil {
		util.Debug("cannot get pending approval operations count", err)
		pendingApprovalCount = 0
	}

	// 创建页面数据结构
	// 判断是否核心成员
	isCoreMember, _ := dao.CanUserManageTeamAccount(s_u.Id, team.Id)
	pageData := PageData{
		SessUser:                 s_u,
		Team:                     singleTeam,
		TeamAccount:              teamAccountWithAvailable,
		UserIsCoreMember:         isCoreMember,
		PendingIncomingTeamCount: pendingIncomingCount,
		PendingIncomingUserCount: pendingIncomingUserCount,
		PendingApprovalCount:     pendingApprovalCount,
	}
	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.account")
}

// HandleTeaTeamTransactionHistory 处理团队交易历史页面请求
func HandleTeaTeamTransactionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	user, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, user, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 必须指定团队ID
	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		report(w, user, "必须指定团队ID。")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, user, "团队ID无效。")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, user, "团队不存在。")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		report(w, user, "您不是该团队成员，无法查看交易历史。")
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 获取团队星茶账户信息
	teamAccount, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get team tea account", err)
		teamAccount = dao.TeaTeamAccount{}
	}

	// 获取过滤器类型
	filterType := r.URL.Query().Get("type")

	// 获取团队交易历史
	transactions, err := dao.GetTeamTeaTransactions(teamId, page, limit)
	if err != nil {
		util.Debug("cannot get team transaction history", err)
		transactions = []map[string]any{}
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser       dao.User
		Team           *dao.Team
		TeamAccount    dao.TeaTeamAccount
		Transactions   []map[string]any
		CurrentPage    int
		Limit          int
		FilterType     string
		BalanceDisplay string
		StatusDisplay  string
	}

	pageData.SessUser = user
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Transactions = transactions
	pageData.CurrentPage = page
	pageData.Limit = limit
	pageData.FilterType = filterType

	// 格式化余额显示
	pageData.BalanceDisplay = fmt.Sprintf("%d 毫克", teamAccount.BalanceMilligrams)

	// 状态显示
	if teamAccount.Status == "frozen" {
		if teamAccount.FrozenReason != "" {
			pageData.StatusDisplay = "已冻结 (" + teamAccount.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.transactions")
}

// HandleTeaTeamOperationsHistory 处理团队操作历史页面请求
func HandleTeaTeamOperationsHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	user, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, user, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 必须指定团队ID
	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		report(w, user, "必须指定团队ID。")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, user, "团队ID无效。")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, user, "团队不存在。")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		report(w, user, "您不是该团队成员，无法查看操作历史。")
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 获取团队所有操作历史（包括待审批、已审批、已完成、已拒绝等所有状态）
	operations, err := dao.GetTeamTransferOutOperations(teamId, page, limit)
	if err != nil {
		util.Debug("cannot get team operations history", err)
		operations = []map[string]any{}
	}

	// 获取团队星茶账户信息
	teamAccount, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get team tea account", err)
		teamAccount = dao.TeaTeamAccount{}
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser    dao.User
		Team        *dao.Team
		TeamAccount struct {
			dao.TeaTeamAccount
			BalanceDisplay string
			StatusDisplay  string
		}
		Operations  []map[string]any
		CurrentPage int
		Limit       int
	}

	pageData.SessUser = user
	pageData.Team = &team
	pageData.TeamAccount.TeaTeamAccount = teamAccount
	pageData.Operations = operations
	pageData.CurrentPage = page
	pageData.Limit = limit

	// 格式化余额显示
	pageData.TeamAccount.BalanceDisplay = fmt.Sprintf("%d 毫克", teamAccount.BalanceMilligrams)

	// 状态显示
	if teamAccount.Status == "frozen" {
		if teamAccount.FrozenReason != "" {
			pageData.TeamAccount.StatusDisplay = "已冻结 (" + teamAccount.FrozenReason + ")"
		} else {
			pageData.TeamAccount.StatusDisplay = "已冻结"
		}
	} else {
		pageData.TeamAccount.StatusDisplay = "正常"
	}

	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.operations.history")
}

// ============================================
// 团队转账API路由处理函数（对应用户版功能）
// ============================================

// CreateTeaTeamToUserTransferAPI 发起团队对用户星茶转账
func CreateTeaTeamToUserTransferAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		FromTeamId       int    `json:"from_team_id"`
		ToUserId         int    `json:"to_user_id"`
		AmountMilligrams int64  `json:"amount_milligrams"`
		Notes            string `json:"notes"`
		ExpireHours      int    `json:"expire_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.FromTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "必须指定转出团队ID")
		return
	}
	if req.ToUserId <= 0 {
		respondWithError(w, http.StatusBadRequest, "必须指定接收用户ID")
		return
	}
	if req.AmountMilligrams <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账数额必须大于0")
		return
	}
	if req.ExpireHours <= 0 || req.ExpireHours > 168 {
		req.ExpireHours = 24
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, req.FromTeamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能发起团队转账")
		return
	}

	// 创建团队对用户转账
	transfer, err := dao.CreateTeaTeamToUserTransferOut(req.FromTeamId, user.Id, req.ToUserId, req.AmountMilligrams, req.Notes, req.ExpireHours)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队对用户转账发起成功", transfer)
}

// CreateTeaTeamToTeamTransferAPI 发起团队对团队星茶转账
func CreateTeaTeamToTeamTransferAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		FromTeamId       int    `json:"from_team_id"`
		ToTeamId         int    `json:"to_team_id"`
		AmountMilligrams int64  `json:"amount_milligrams"`
		Notes            string `json:"notes"`
		ExpireHours      int    `json:"expire_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.FromTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "必须指定转出团队ID")
		return
	}
	if req.ToTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "必须指定接收团队ID")
		return
	}
	if req.AmountMilligrams <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账数额必须大于0")
		return
	}
	if req.ExpireHours <= 0 || req.ExpireHours > 168 {
		req.ExpireHours = 24
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, req.FromTeamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能发起团队转账")
		return
	}

	// 创建团队对团队转账
	transfer, err := dao.CreateTeaTeamToTeamTransferOut(req.FromTeamId, user.Id, req.ToTeamId, req.AmountMilligrams, req.Notes, req.ExpireHours)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队对团队转账发起成功", transfer)
}

// GetTeaTeamToUserTransferOutsAPI 获取团队对用户转出记录API
func GetTeaTeamToUserTransferOutsAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看转出记录")
		return
	}

	page, limit := getPaginationParams(r)
	operations, err := dao.GetTeamTransferOutOperations(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队转出记录失败")
		return
	}

	respondWithPagination(w, "获取团队对用户转出记录成功", operations, page, limit, 0)
}

// GetTeaTeamToTeamTransferOutsAPI 获取团队对团队转出记录API
func GetTeaTeamToTeamTransferOutsAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看转出记录")
		return
	}

	page, limit := getPaginationParams(r)
	operations, err := dao.GetTeamTransferOutOperations(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队转出记录失败")
		return
	}

	respondWithPagination(w, "获取团队对团队转出记录成功", operations, page, limit, 0)
}

// GetTeaTeamPendingTeamToUserTransfersAPI 获取团队待确认团队对用户转账API
func GetTeaTeamPendingTeamToUserTransfersAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看待确认转账")
		return
	}

	page, limit := getPaginationParams(r)
	operations, err := dao.GetPendingTeamToUserOperations(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待确认转账失败")
		return
	}

	respondWithPagination(w, "获取团队待确认团队对用户转账成功", operations, page, limit, 0)
}

// GetTeaTeamPendingTeamToTeamTransfersAPI 获取团队待确认团队对团队转账API
func GetTeaTeamPendingTeamToTeamTransfersAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看待确认转账")
		return
	}

	page, limit := getPaginationParams(r)
	operations, err := dao.GetPendingTeamToTeamOperations(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待确认转账失败")
		return
	}

	respondWithPagination(w, "获取团队待确认团队对团队转账成功", operations, page, limit, 0)
}

// GetTeaTeamToUserTransferHistoryAPI 获取团队对用户转账历史API
func GetTeaTeamToUserTransferHistoryAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看转账历史")
		return
	}

	page, limit := getPaginationParams(r)
	transactions, err := dao.GetTeamTeaTransactions(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取转账历史失败")
		return
	}

	respondWithPagination(w, "获取团队对用户转账历史成功", transactions, page, limit, 0)
}

// GetTeaTeamToTeamTransferHistoryAPI 获取团队对团队转账已完成状态记录API
func GetTeaTeamToTeamTransferHistoryAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看转账历史")
		return
	}

	page, limit := getPaginationParams(r)
	transactions, err := dao.GetTeamTeaTransactions(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取转账历史失败")
		return
	}

	respondWithPagination(w, "获取团队对团队转账历史成功", transactions, page, limit, 0)
}

// GetTeaTeamFromUserTransferInsAPI 获取团队接收用户转入记录API
func GetTeaTeamFromUserTransferInsAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看转入记录")
		return
	}

	page, limit := getPaginationParams(r)
	transfers, err := dao.GetTeamTransferInOperations(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取转入记录失败")
		return
	}

	respondWithPagination(w, "获取团队接收用户转入记录成功", transfers, page, limit, 0)
}

// GetTeaTeamFromTeamTransferInsAPI 获取团队接收团队转入记录API
func GetTeaTeamFromTeamTransferInsAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看转入记录")
		return
	}

	page, limit := getPaginationParams(r)
	transfers, err := dao.GetTeamTransferInOperations(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取转入记录失败")
		return
	}

	respondWithPagination(w, "获取团队接收团队转入记录成功", transfers, page, limit, 0)
}

// ============================================
// 团队转账确认/拒绝API路由处理函数
// ============================================

// ConfirmTeaTeamFromUserTransferAPI 团队确认接收来自用户转账API
func ConfirmTeaTeamFromUserTransferAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		TransferUuid string `json:"transfer_uuid"`
		ToTeamId     int    `json:"to_team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}
	if req.ToTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, req.ToTeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队成员身份失败")
		return
	}
	if !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能确认接收用户转账")
		return
	}

	err = dao.TeaConfirmUserToTeamTransferOut(req.TransferUuid, req.ToTeamId, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队确认接收来自用户转账成功", nil)
}

// RejectTeaTeamFromUserTransferAPI 团队拒绝接收来自用户转账API
func RejectTeaTeamFromUserTransferAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		TransferUuid string `json:"transfer_uuid"`
		ToTeamId     int    `json:"to_team_id"`
		Reason       string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}
	if req.ToTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, req.ToTeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队成员身份失败")
		return
	}
	if !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能拒绝接收用户转账")
		return
	}

	err = dao.TeaTeamRejectFromUserTransferIn(req.TransferUuid, req.ToTeamId, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队拒绝接收来自用户转账成功", nil)
}

// ConfirmTeaTeamFromTeamTransferAPI 团队确认接收来自团队转账API
func ConfirmTeaTeamFromTeamTransferAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		TransferUuid string `json:"transfer_uuid"`
		ToTeamId     int    `json:"to_team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}

	if req.ToTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, req.ToTeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队成员身份失败")
		return
	}
	if !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能确认接收团队转账")
		return
	}

	err = dao.TeaConfirmTeamToTeamTransferOut(req.TransferUuid, req.ToTeamId, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队确认接收来自团队转账成功", nil)
}

// RejectTeaTeamFromTeamTransferAPI 团队拒绝接收来自团队转账API
func RejectTeaTeamFromTeamTransferAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		TransferUuid string `json:"transfer_uuid"`
		ToTeamId     int    `json:"to_team_id"`
		Reason       string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}

	if req.ToTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, req.ToTeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队成员身份失败")
		return
	}
	if !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能拒绝接收团队转账")
		return
	}

	err = dao.TeaTeamRejectFromTeamTransferIn(req.TransferUuid, req.ToTeamId, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队拒绝接收来自团队转账成功", nil)
}

// ============================================
// 团队转账审批API路由处理函数
// ============================================

// ApproveTeaTeamToUserTransferAPI 审批团队对用户转账API
func ApproveTeaTeamToUserTransferAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		TransferUuid string `json:"transfer_uuid"`
		FromTeamId   int    `json:"from_team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}
	if req.FromTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队核心成员
	CoreManage, err := dao.CanUserManageTeamAccount(user.Id, req.FromTeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队核心成员身份失败")
		return
	}
	if !CoreManage {
		respondWithError(w, http.StatusForbidden, "只有团队核心成员才能审批团队对用户转账")
		return
	}

	err = dao.TeaTeamApproveToUserTransferOut(req.FromTeamId, req.TransferUuid, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队对用户转账审批通过", nil)
}

// ApproveTeaTeamToTeamTransferAPI 审批团队对团队转账API
func ApproveTeaTeamToTeamTransferAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		TransferUuid string `json:"transfer_uuid"`
		FromTeamId   int    `json:"from_team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}
	if req.FromTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队核心成员
	CoreManage, err := dao.CanUserManageTeamAccount(user.Id, req.FromTeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队核心成员身份失败")
		return
	}
	if !CoreManage {
		respondWithError(w, http.StatusForbidden, "只有团队核心成员才能审批团队对团队转账")
		return
	}

	err = dao.TeaTeamApproveToTeamTransferOut(req.FromTeamId, req.TransferUuid, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队对团队转账审批通过", nil)
}

// RejectTeaTeamToUserTransferApprovalAPI 拒绝审批团队对用户转账API
func RejectTeaTeamToUserTransferApprovalAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		TransferUuid string `json:"transfer_uuid"`
		FromTeamId   int    `json:"from_team_id"`
		Reason       string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}
	if req.FromTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队核心成员
	CoreManage, err := dao.CanUserManageTeamAccount(user.Id, req.FromTeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队核心成员身份失败")
		return
	}
	if !CoreManage {
		respondWithError(w, http.StatusForbidden, "只有团队核心成员才能审批团队对用户转账")
		return
	}

	err = dao.TeaTeamRejectToUserTransferOut(req.FromTeamId, req.TransferUuid, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队对用户转账审批拒绝", nil)
}

// RejectTeaTeamToTeamTransferApprovalAPI 拒绝审批团队对团队转账API
func RejectTeaTeamToTeamTransferApprovalAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	var req struct {
		TransferUuid string `json:"transfer_uuid"`
		FromTeamId   int    `json:"from_team_id"`
		Reason       string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}
	if req.FromTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队核心成员
	CoreManage, err := dao.CanUserManageTeamAccount(user.Id, req.FromTeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队核心成员身份失败")
		return
	}
	if !CoreManage {
		respondWithError(w, http.StatusForbidden, "只有团队核心成员才能审批团队对团队转账")
		return
	}

	err = dao.TeaTeamRejectToTeamTransferOut(req.FromTeamId, req.TransferUuid, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队对团队转账审批拒绝", nil)
}

// ============================================
// 团队转账页面路由处理函数
// ============================================

// HandleTeaTeamPendingTeamToUserTransfers 处理团队待确认团队对用户转账页面请求
func HandleTeaTeamPendingTeamToUserTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 实现页面渲染逻辑
	respondWithError(w, http.StatusNotImplemented, "页面功能待实现")
}

// HandleTeaTeamPendingTeamToTeamTransfers 处理团队待确认团队对团队转账页面请求
func HandleTeaTeamPendingTeamToTeamTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 实现页面渲染逻辑
	respondWithError(w, http.StatusNotImplemented, "页面功能待实现")
}

// HandleTeaTeamToUserTransferHistory 处理团队对用户转账历史页面请求
func HandleTeaTeamToUserTransferHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 实现页面渲染逻辑
	respondWithError(w, http.StatusNotImplemented, "页面功能待实现")
}

// HandleTeaTeamToTeamTransferHistory 处理团队对团队转账历史页面请求
func HandleTeaTeamToTeamTransferHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 实现页面渲染逻辑
	respondWithError(w, http.StatusNotImplemented, "页面功能待实现")
}

// HandleTeaTeamFromUserTransferIns 处理团队接收用户转入记录页面请求
func HandleTeaTeamFromUserTransferIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 实现页面渲染逻辑
	respondWithError(w, http.StatusNotImplemented, "页面功能待实现")
}

// HandleTeaTeamFromTeamTransferIns 处理团队接收团队转入记录页面请求
func HandleTeaTeamFromTeamTransferIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 实现页面渲染逻辑
	respondWithError(w, http.StatusNotImplemented, "页面功能待实现")
}

// HandleTeaTeamPendingIncomingTransfers 处理团队待确认转入转账页面请求
func HandleTeaTeamPendingIncomingTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TeamPendingIncomingTransfersGet(w, r)
}

// TeamPendingIncomingTransfersGet 获取团队待确认转入转账页面
func TeamPendingIncomingTransfersGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 必须指定团队ID
	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		report(w, s_u, "必须指定团队ID。")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, s_u, "团队ID无效。")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, s_u, "团队不存在。")
		return
	}

	// 检查用户是否是团队正常成员
	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		report(w, s_u, "只有团队成员才能查看待确认转账。")
		return
	}

	// 获取团队星茶账户
	teamAccount, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get team tea account", err)
		report(w, s_u, "获取团队星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待确认转入转账
	transfers, err := dao.GetPendingTeamIncomingTransfers(teamId, page, limit)
	if err != nil {
		util.Debug("cannot get pending incoming transfers", err)
		report(w, s_u, "获取待确认转账失败。")
		return
	}

	// 增强转账数据，添加预处理字段
	type EnhancedTransfer struct {
		Uuid          string
		TransferType  string
		FromId        int
		FromName      string
		TeamId        int
		ToTeamName    string
		AmountDisplay string
		Notes         string
		CreatedAt     string
		ExpiresAt     string
		StatusDisplay string
		IsExpired     bool
		IsNearExpiry  bool
		CanConfirm    bool
	}

	var enhancedTransfers []EnhancedTransfer
	for _, transfer := range transfers {
		enhanced := EnhancedTransfer{
			Uuid:         safeString(transfer["uuid"]),
			TransferType: safeString(transfer["transfer_type"]),
			FromId:       int(safeInt64(transfer["from_id"])),
			FromName:     safeString(transfer["from_name"]),
			Notes:        safeString(transfer["notes"]),
			CreatedAt:    safeString(transfer["created_at"]),
			ExpiresAt:    safeString(transfer["expires_at"]),
		}

		// 获取发送方信息
		if enhanced.TransferType == "user_to_team" {
			enhanced.FromName = safeString(transfer["from_user_name"])
		} else if enhanced.TransferType == "team_to_team" {
			enhanced.ToTeamName = safeString(transfer["from_team_name"])
		}

		// 格式化金额显示
		amountMilligrams := safeInt64(transfer["amount_milligrams"])
		enhanced.AmountDisplay = fmt.Sprintf("%d 毫克", amountMilligrams)

		// 格式化时间显示和检查过期状态
		if createdAt, ok := transfer["created_at"].(time.Time); ok {
			enhanced.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		} else {
			enhanced.CreatedAt = safeString(transfer["created_at"])
		}

		if expiresAt, ok := transfer["expires_at"].(time.Time); ok {
			enhanced.ExpiresAt = expiresAt.Format("2006-01-02 15:04:05")
			// 检查是否过期和即将过期
			nowTime := time.Now()
			enhanced.IsExpired = expiresAt.Before(nowTime)
			enhanced.IsNearExpiry = !enhanced.IsExpired && expiresAt.Sub(nowTime) < time.Hour
		} else {
			enhanced.ExpiresAt = safeString(transfer["expires_at"])
			// 如果无法解析时间，默认为未过期
			enhanced.IsExpired = false
			enhanced.IsNearExpiry = false
		}
		enhanced.CanConfirm = !enhanced.IsExpired

		// 状态显示
		if enhanced.IsExpired {
			enhanced.StatusDisplay = "已过期"
		} else if enhanced.IsNearExpiry {
			enhanced.StatusDisplay = "即将过期"
		} else {
			enhanced.StatusDisplay = "可确认"
		}

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser    dao.User
		Team        *dao.Team
		TeamAccount dao.TeaTeamAccount
		Transfers   []EnhancedTransfer
		CurrentPage int
		Limit       int
	}

	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Transfers = enhancedTransfers
	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.pending_incoming_transfers")
}

// GetTeaTeamPendingIncomingTransfersAPI 获取团队待确认转入转账API
func GetTeaTeamPendingIncomingTransfersAPI(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "必须指定团队ID")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看待确认转账")
		return
	}

	page, limit := getPaginationParams(r)
	transfers, err := dao.GetPendingTeamIncomingTransfers(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待确认转账失败")
		return
	}

	respondWithPagination(w, "获取团队待确认转入转账成功", transfers, page, limit, 0)
}

// 辅助函数：安全获取字符串值
func safeString(val any) string {
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", val)
}

// HandleTeaTeamToTeamCompletedTransferOuts 处理团队对团队已完成状态转账记录页面请求
func HandleTeaTeamToTeamCompletedTransferOuts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	user, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, user, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 必须指定团队ID
	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		report(w, user, "必须指定团队ID。")
		return
	}

	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		report(w, user, "团队ID无效。")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, user, "团队不存在。")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		report(w, user, "您不是该团队成员，无法查看已完成转账记录。")
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 获取团队对团队已完成状态转账记录
	transferOuts, err := dao.GetTeaTeamToTeamCompletedTransferOuts(teamId, page, limit)
	if err != nil {
		util.Debug("cannot get team-to-team completed transfers", err)
		transferOuts = []dao.TeaTeamToTeamTransferOut{}
	}

	// 转换为模板需要的map格式
	transactionMaps := make([]map[string]any, len(transferOuts))
	for i, t := range transferOuts {
		transactionMaps[i] = map[string]any{
			"created_at":        t.CreatedAt,
			"amount_milligrams": t.AmountMilligrams,
			"notes":             t.Notes,
			"to_team_name":      t.ToTeamName,
			"to_team_id":        t.ToTeamId,
			"status":            t.Status,
		}
	}

	// 获取团队星茶账户信息
	teamAccount, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get team tea account", err)
		teamAccount = dao.TeaTeamAccount{}
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser       dao.User
		Team           *dao.Team
		TeamAccount    dao.TeaTeamAccount
		Transactions   []map[string]any
		CurrentPage    int
		Limit          int
		BalanceDisplay string
		StatusDisplay  string
	}

	pageData.SessUser = user
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Transactions = transactionMaps
	pageData.CurrentPage = page
	pageData.Limit = limit

	// 格式化余额显示
	pageData.BalanceDisplay = fmt.Sprintf("%d 毫克", teamAccount.BalanceMilligrams)

	// 状态显示
	if teamAccount.Status == "frozen" {
		if teamAccount.FrozenReason != "" {
			pageData.StatusDisplay = "已冻结 (" + teamAccount.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.completed_team_to_team_transfers")
}

// 辅助函数：安全获取int64值
func safeInt64(val any) int64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}
