package route

import (
	"encoding/json"
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
	dao.TeamTeaAccount
	AvailableBalanceGrams float64
}

// PageData 用于模板渲染
type PageData struct {
	SessUser         dao.User
	Team             *dao.Team
	TeamAccount      TeamAccountWithAvailable
	Transactions     []dao.TeaTransaction
	UserIsCoreMember bool
}

// GetTeamTeaAccount 获取团队茶叶账户信息
func GetTeamTeaAccount(w http.ResponseWriter, r *http.Request) {
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
	err = dao.EnsureTeamTeaAccountExists(teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队茶叶账户失败")
		return
	}

	// 获取团队茶叶账户
	account, err := dao.GetTeamTeaAccountByTeamId(teamId)
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

// GetTeamTeaTransactions 获取团队交易流水记录
func GetTeamTeaTransactions(w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看交易流水")
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	transactionType := r.URL.Query().Get("transaction_type")

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

	// 获取团队交易流水
	transactions, err := dao.GetTeamTeaTransactions(teamId, page, limit, transactionType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取交易流水失败")
		return
	}

	respondWithSuccess(w, "获取团队交易流水成功", transactions)
}

// FreezeTeamAccount 冻结团队茶叶账户
func FreezeTeamAccount(w http.ResponseWriter, r *http.Request) {
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
	account, err := dao.GetTeamTeaAccountByTeamId(req.TeamId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "团队账户不存在")
		return
	}

	err = account.UpdateStatus(dao.TeamTeaAccountStatus_Frozen, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "冻结账户失败")
		return
	}

	respondWithSuccess(w, "团队账户冻结成功", nil)
}

// UnfreezeTeamAccount 解冻团队茶叶账户
func UnfreezeTeamAccount(w http.ResponseWriter, r *http.Request) {
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
	account, err := dao.GetTeamTeaAccountByTeamId(req.TeamId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "团队账户不存在")
		return
	}

	err = account.UpdateStatus(dao.TeamTeaAccountStatus_Normal, "")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "解冻账户失败")
		return
	}

	respondWithSuccess(w, "团队账户解冻成功", nil)
}

// HandleTeamTeaAccount 处理团队茶叶账户页面请求
func HandleTeamTeaAccount(w http.ResponseWriter, r *http.Request) {
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
	err = dao.EnsureTeamTeaAccountExists(team.Id)
	if err != nil {
		util.Debug("cannot ensure team tea account exists", err)
		report(w, s_u, "获取团队茶叶账户失败。")
		return
	}

	// 获取团队茶叶账户
	teamAccount, err := dao.GetTeamTeaAccountByTeamId(team.Id)
	if err != nil {
		util.Debug("cannot get team tea account", err)
		report(w, s_u, "获取团队茶叶账户失败。")
		return
	}

	// 计算可用余额
	teamAccountWithAvailable := TeamAccountWithAvailable{
		TeamTeaAccount:        teamAccount,
		AvailableBalanceGrams: teamAccount.BalanceGrams - teamAccount.LockedBalanceGrams,
	}

	// 获取团队交易流水
	transactions, err := dao.GetTeamTeaTransactions(team.Id, 1, 20, "")
	if err != nil {
		util.Debug("cannot get team transactions", err)
		transactions = []dao.TeaTransaction{}
	}

	// 创建页面数据结构
	// 判断是否核心成员
	isCoreMember, _ := dao.CanUserManageTeamAccount(s_u.Id, team.Id)
	pageData := PageData{
		SessUser:         s_u,
		Team:             singleTeam,
		TeamAccount:      teamAccountWithAvailable,
		Transactions:     transactions,
		UserIsCoreMember: isCoreMember,
	}
	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.account")
}

// HandleTeamTeaTransactions 处理团队交易流水页面请求
func HandleTeamTeaTransactions(w http.ResponseWriter, r *http.Request) {
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
		report(w, user, "您不是该团队成员，无法查看交易流水。")
		return
	}

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	transactionType := r.URL.Query().Get("type")

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

	// 获取团队交易流水
	transactions, err := dao.GetTeamTeaTransactions(teamId, page, limit, transactionType)
	if err != nil {
		util.Debug("cannot get team transactions", err)
		transactions = []dao.TeaTransaction{}
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser        dao.User
		Team            *dao.Team
		Transactions    []dao.TeaTransaction
		CurrentPage     int
		Limit           int
		TransactionType string
	}

	pageData.SessUser = user
	pageData.Team = &team
	pageData.Transactions = transactions
	pageData.CurrentPage = page
	pageData.Limit = limit
	pageData.TransactionType = transactionType

	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "team.tea.transactions")
}
