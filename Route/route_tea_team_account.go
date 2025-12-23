package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	dao "teachat/DAO"
	util "teachat/Util"
)

// 团队茶叶账户相关响应结构体
type TeamTeaAccountResponse struct {
	Uuid         string  `json:"uuid"`
	TeamId       int     `json:"team_id"`
	TeamName     string  `json:"team_name,omitempty"`
	BalanceGrams float64 `json:"balance_grams"`
	Status       string  `json:"status"`
	FrozenReason *string `json:"frozen_reason,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// 团队茶叶账户解冻请求结构体
type UnfreezeTeamAccountRequest struct {
	TeamId int `json:"team_id"`
}

// TeamAccountWithAvailable 用于模板渲染，包含可用余额
type TeamAccountWithAvailable struct {
	dao.TeaTeamAccount
	AvailableBalanceGrams float64
}

// PageData 用于模板渲染
type PageData struct {
	SessUser             dao.User
	Team                 *dao.Team
	TeamAccount          TeamAccountWithAvailable
	TransactionHistory   []map[string]interface{}
	UserIsCoreMember     bool
	PendingIncomingCount int
}

// GetTeaTeamAccount 获取团队茶叶账户信息
func GetTeaTeamAccount(w http.ResponseWriter, r *http.Request) {
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
	isMember, err := dao.IsTeamMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看茶叶账户")
		return
	}

	// 确保团队有茶叶账户
	err = dao.EnsureTeaTeamAccountExists(teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队茶叶账户失败")
		return
	}

	// 获取团队茶叶账户
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "团队茶叶账户不存在")
		return
	}

	// 获取团队名称
	var teamName string
	if team, err := dao.GetTeam(teamId); err == nil {
		teamName = team.Name
	}

	response := TeamTeaAccountResponse{
		Uuid:         account.Uuid,
		TeamId:       account.TeamId,
		TeamName:     teamName,
		BalanceGrams: account.BalanceGrams,
		Status:       account.Status,
		FrozenReason: account.FrozenReason,
		CreatedAt:    account.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	respondWithSuccess(w, "获取团队茶叶账户成功", response)
}

// GetTeaTeamTransactionHistory 获取团队交易历史（从转出表中查询）
func GetTeaTeamTransactionHistory(w http.ResponseWriter, r *http.Request) {
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
	isMember, err := dao.IsTeamMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看交易历史")
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

	// 获取团队交易历史
	transactions, err := dao.GetTeamTeaTransactions(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取交易历史失败")
		return
	}

	respondWithSuccess(w, "获取团队交易历史成功", transactions)
}

// FreezeTeaTeamAccount 冻结团队茶叶账户
func FreezeTeaTeamAccount(w http.ResponseWriter, r *http.Request) {
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
	canManage, err := dao.CanUserManageTeamAccount(user.Id, req.TeamId)
	if err != nil || !canManage {
		respondWithError(w, http.StatusForbidden, "您没有权限管理该团队账户")
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

// UnfreezeTeaTeamAccount 解冻团队茶叶账户
func UnfreezeTeaTeamAccount(w http.ResponseWriter, r *http.Request) {
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
	canManage, err := dao.CanUserManageTeamAccount(user.Id, req.TeamId)
	if err != nil || !canManage {
		respondWithError(w, http.StatusForbidden, "您没有权限管理该团队账户")
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

// HandleTeaTeamTeaAccount 处理团队茶叶账户页面请求
func HandleTeaTeamTeaAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TeamTeaAccountGet(w, r)
}

// TeamTeaAccountGet 获取团队茶叶账户页面
func TeamTeaAccountGet(w http.ResponseWriter, r *http.Request) {
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

	// 显示指定团队的茶叶账户
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
	isMember, err := dao.IsTeamMember(s_u.Id, teamId)
	if err != nil || !isMember {
		report(w, s_u, "您不是该团队成员，无法查看茶叶账户。")
		return
	}

	singleTeam := &team

	// 确保团队有茶叶账户
	err = dao.EnsureTeaTeamAccountExists(team.Id)
	if err != nil {
		util.Debug("cannot ensure team tea account exists", err)
		report(w, s_u, "获取团队茶叶账户失败。")
		return
	}

	// 获取团队茶叶账户
	teamAccount, err := dao.GetTeaTeamAccountByTeamId(team.Id)
	if err != nil {
		util.Debug("cannot get team tea account", err)
		report(w, s_u, "获取团队茶叶账户失败。")
		return
	}

	// 计算可用余额
	teamAccountWithAvailable := TeamAccountWithAvailable{
		TeaTeamAccount:        teamAccount,
		AvailableBalanceGrams: teamAccount.BalanceGrams - teamAccount.LockedBalanceGrams,
	}

	// 获取待确认接收操作数量
	pendingIncomingCount, err := dao.CountPendingTeamReceipts(team.Id)
	if err != nil {
		util.Debug("cannot get pending incoming transfers count", err)
		pendingIncomingCount = 0
	}

	// 创建页面数据结构
	// 判断是否核心成员
	isCoreMember, _ := dao.CanUserManageTeamAccount(s_u.Id, team.Id)
	pageData := PageData{
		SessUser:             s_u,
		Team:                 singleTeam,
		TeamAccount:          teamAccountWithAvailable,
		TransactionHistory:   []map[string]interface{}{}, // 空数组，不再显示最近交易记录
		UserIsCoreMember:     isCoreMember,
		PendingIncomingCount: pendingIncomingCount,
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
	isMember, err := dao.IsTeamMember(user.Id, teamId)
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

	// 获取团队交易历史
	// transactions, err := dao.GetTeamTeaTransactions(teamId, page, limit)
	// if err != nil {
	// 	util.Debug("cannot get team transaction history", err)
	// 	transactions = []map[string]interface{}{}
	// }

	// 获取团队茶叶账户信息
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
		transactions = []map[string]interface{}{}
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser       dao.User
		Team           *dao.Team
		TeamAccount    dao.TeaTeamAccount
		Transactions   []map[string]interface{}
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
	if teamAccount.BalanceGrams >= 1 {
		pageData.BalanceDisplay = fmt.Sprintf("%.2f 克", teamAccount.BalanceGrams)
	} else {
		pageData.BalanceDisplay = fmt.Sprintf("%.0f 毫克", teamAccount.BalanceGrams*1000)
	}

	// 状态显示
	if teamAccount.Status == "frozen" {
		if teamAccount.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *teamAccount.FrozenReason + ")"
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
	isMember, err := dao.IsTeamMember(user.Id, teamId)
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
		operations = []map[string]interface{}{}
	}

	// 获取团队茶叶账户信息
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
		Operations  []map[string]interface{}
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
	if teamAccount.BalanceGrams >= 1 {
		pageData.TeamAccount.BalanceDisplay = fmt.Sprintf("%.2f 克", teamAccount.BalanceGrams)
	} else {
		pageData.TeamAccount.BalanceDisplay = fmt.Sprintf("%.0f 毫克", teamAccount.BalanceGrams*1000)
	}

	// 状态显示
	if teamAccount.Status == "frozen" {
		if teamAccount.FrozenReason != nil {
			pageData.TeamAccount.StatusDisplay = "已冻结 (" + *teamAccount.FrozenReason + ")"
		} else {
			pageData.TeamAccount.StatusDisplay = "已冻结"
		}
	} else {
		pageData.TeamAccount.StatusDisplay = "正常"
	}

	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.operations.history")
}
