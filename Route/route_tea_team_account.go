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

// TeaTeamAccountWithAvailable 用于模板渲染，包含可用余额
type TeaTeamAccountWithAvailable struct {
	dao.TeaTeamAccount
	AvailableBalanceMilligrams int
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

// TeaTeamAccountGet 获取团队星茶账户页面
func TeaTeamAccountGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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
	teaTeamAccount, err := dao.GetTeaTeamAccountByTeamId(team.Id)
	if err != nil {
		util.Debug("cannot get team tea account", err)
		report(w, s_u, "获取团队星茶账户失败。")
		return
	}

	// 计算可用余额
	teaTeamAccountWithAvailable := TeaTeamAccountWithAvailable{
		TeaTeamAccount:             teaTeamAccount,
		AvailableBalanceMilligrams: int(teaTeamAccount.BalanceMilligrams) - int(teaTeamAccount.LockedBalanceMilligrams),
	}

	// 获取待确认接收来自团队转账操作数量
	pendingFromTeamCount, err := dao.TeaTeamCountPendingFromTeamReceipts(team.Id)
	if err != nil {
		util.Debug("cannot get pending incoming transfers count", err)
	}
	// 获取待确认接收来自用户转账操作数量
	pendingFromUserCount, err := dao.TeaTeamCountPendingFromUserReceipts(team.Id)
	if err != nil {
		util.Debug("cannot get pending incoming user transfers count", err)
	}

	// 获取帐户转出，待审批操作数量
	pendingApprovalCount, err := dao.CountTeaTeamPendingApprovals(team.Id)
	if err != nil {
		util.Debug("cannot get pending approval operations count", err)
	}

	// 创建页面数据结构
	// 判断是否核心成员
	isCoreMember, err := dao.CanUserManageTeamAccount(s_u.Id, team.Id)
	if err != nil {
		util.Debug("cannot check if user is core member", err)
	}
	var pageData struct {
		SessUser             dao.User
		Team                 *dao.Team
		TeamAccount          TeaTeamAccountWithAvailable
		UserIsCoreMember     bool
		PendingFromTeamCount int
		PendingFromUserCount int
		PendingApprovalCount int
	}
	pageData.SessUser = s_u
	pageData.Team = singleTeam
	pageData.TeamAccount = teaTeamAccountWithAvailable
	pageData.UserIsCoreMember = isCoreMember
	pageData.PendingFromTeamCount = pendingFromTeamCount
	pageData.PendingFromUserCount = pendingFromUserCount
	pageData.PendingApprovalCount = pendingApprovalCount

	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.account")
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

	err = dao.TeaTeamConfirmFromUserTransfer(req.TransferUuid, req.ToTeamId, user.Id)
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

	err = dao.TeaTeamRejectFromUserTransfer(req.TransferUuid, req.ToTeamId, user.Id, req.Reason)
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

	err = dao.TeaTeamConfirmFromTeamTransfer(req.TransferUuid, req.ToTeamId, user.Id)
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

	err = dao.TeaTeamRejectFromTeamTransfer(req.TransferUuid, req.ToTeamId, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队拒绝接收来自团队转账成功", nil)
}

// ============================================
// 团队转账审批API路由处理函数
// ============================================

// ApproveTeaTeamToUserTransferAPI 审批通过团队对用户转账API
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

// ApproveTeaTeamToTeamTransferAPI 审批通过团队对团队转账API
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

// RejectTeaTeamToUserTransferApprovalAPI 审批拒绝团队对用户转账API
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

// RejectTeaTeamToTeamTransferApprovalAPI 审批拒绝团队对团队转账API 0211
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
// 团队转出页面路由处理函数 (outgoing transfer list)
// ============================================
// CreateTeaTeamToTeamTransferAPI 发起团队对团队星茶转账 0211
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

	if req.ToTeamId == dao.TeamIdFreelancer || req.FromTeamId == dao.TeamIdFreelancer {
		respondWithError(w, http.StatusBadRequest, "无法转账,自由者团队")
		return
	}
	// 检查帐户是否被冻结？
	frozen, reason, err := dao.CheckTeaTeamAccountFrozen(req.FromTeamId)
	if err != nil {
		util.Debug("check tea team account frozen error:", err)
		respondWithError(w, http.StatusInternalServerError, "检查团队账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("账户已被冻结：%s", reason))
		return
	}
	// 检查接收方团队星茶帐户是否被冻结？
	frozen, reason, err = dao.CheckTeaTeamAccountFrozen(req.ToTeamId)
	if err != nil {
		util.Debug("check tea team account frozen error:", err)
		respondWithError(w, http.StatusInternalServerError, "检查接收团队账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("接收方账户已被冻结：%s", reason))
		return
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

// CreateTeaTeamToUserTransferAPI 发起团队对用户星茶转账 0211
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
	if req.FromTeamId == dao.TeamIdFreelancer {
		respondWithError(w, http.StatusBadRequest, "无法转账, 自由者团队")
		return
	}
	// 检查帐户是否被冻结？
	frozen, reason, err := dao.CheckTeaTeamAccountFrozen(req.FromTeamId)
	if err != nil {
		util.Debug("check tea team account frozen error:", err)
		respondWithError(w, http.StatusInternalServerError, "检查团队账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("账户已被冻结：%s", reason))
		return
	}
	// 检查接收方用户星茶帐户是否被冻结？
	frozen, reason, err = dao.CheckTeaUserAccountFrozen(req.ToUserId)
	if err != nil {
		util.Debug("check tea user account frozen error:", err)
		respondWithError(w, http.StatusInternalServerError, "检查接收方账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("接收方账户已被冻结：%s", reason))
		return
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

// GetTeaTeamPendingApproveToTeamTransfers 获取团队转出待审批对团队转账页面 0211
func GetTeaTeamPendingApproveToTeamTransfers(w http.ResponseWriter, r *http.Request) {
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
	team, err := dao.GetTeamByID(teamIdStr)
	if err != nil {
		util.Debug("cannot get team by id:", teamIdStr, err)
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员(查看)
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看待审批转账")
		return
	}
	//检查用户是否核心成员（审批）
	isCoreMember, err := team.IsCoreMember(user.Id)
	if err != nil {
		util.Debug("cannot check if user is core member", err)
	}

	// 确保团队有星茶账户
	err = dao.EnsureTeaTeamAccountExists(teamId)
	if err != nil {
		util.Debug("cannot ensure tea team account exists", err)
		respondWithError(w, http.StatusInternalServerError, "获取团队星茶账户失败")
		return
	}
	// 获取团队星茶账户
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get tea team account", err)
		respondWithError(w, http.StatusInternalServerError, "获取团队星茶账户失败")
		return
	}
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("账户已被冻结：%s", account.FrozenReason))
		return
	}

	page, limit := getPaginationParams(r)
	transfers, err := dao.TeaTeamPendingApprovalToTeamTransferOuts(teamId, page, limit, r.Context())
	if err != nil {
		util.Debug("cannot get pending approve team to team transfer outs", err)
		respondWithError(w, http.StatusInternalServerError, "获取待审批转账记录失败")
		return
	}

	// 增强转账数据，添加团队信息和状态显示
	type EnhancedPendingApproveTransfer struct {
		dao.TeaTeamToTeamTransferOut
		IsExpired     bool
		CanApprove    bool
		TimeRemaining string
	}
	var enhancedTransfers []EnhancedPendingApproveTransfer
	for _, t := range transfers {
		// 计算过期时间,使用UTC时间比较
		expiresAtUTC := t.ExpiresAt.UTC()
		nowUTC := time.Now().UTC()
		enhanced := EnhancedPendingApproveTransfer{
			TeaTeamToTeamTransferOut: t,
			IsExpired:                nowUTC.After(expiresAtUTC),
			CanApprove:               !expiresAtUTC.Before(nowUTC),
		}

		if enhanced.CanApprove {
			timeRemaining := expiresAtUTC.Sub(nowUTC)
			if timeRemaining > time.Hour {
				hours := int(timeRemaining.Hours())
				minutes := int(timeRemaining.Minutes()) % 60
				if minutes > 0 {
					enhanced.TimeRemaining = fmt.Sprintf("%d小时%d分钟", hours, minutes)
				} else {
					enhanced.TimeRemaining = fmt.Sprintf("%d小时", hours)
				}
			} else if timeRemaining > time.Minute {
				minutes := int(timeRemaining.Minutes())
				enhanced.TimeRemaining = fmt.Sprintf("%d分钟", minutes)
			} else {
				enhanced.TimeRemaining = "即将过期"
			}
		} else {
			enhanced.TimeRemaining = "已过期"
		}

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser       dao.User
		Team           *dao.Team
		IsTeamMember   bool
		IsCoreMember   bool
		TeaTeamAccount dao.TeaTeamAccount
		Transfers      []EnhancedPendingApproveTransfer
		StatusDisplay  string
		CurrentPage    int
		Limit          int
	}

	pageData.SessUser = user
	pageData.Team = &team
	pageData.IsTeamMember = isMember
	pageData.IsCoreMember = isCoreMember
	pageData.TeaTeamAccount = account
	pageData.Transfers = enhancedTransfers
	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.pending_approve_to_team_transfers")

}

// GetTeaTeamPendingApproveToUserTransfers 获取团队待审批团队对用户转账页面 0211
func GetTeaTeamPendingApproveToUserTransfers(w http.ResponseWriter, r *http.Request) {
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
	team, err := dao.GetTeamByID(teamIdStr)
	if err != nil {
		util.Debug("cannot get team by id:", teamIdStr, err)
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员(查看)
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看待审批转账")
		return
	}
	//检查用户是否核心成员（审批）
	isCoreMember, err := team.IsCoreMember(user.Id)
	if err != nil {
		util.Debug("cannot check if user is core member", err)
	}

	// 确保团队有星茶账户
	err = dao.EnsureTeaTeamAccountExists(teamId)
	if err != nil {
		util.Debug("cannot ensure tea team account exists", err)
		respondWithError(w, http.StatusInternalServerError, "获取团队星茶账户失败")
		return
	}
	// 获取团队星茶账户
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get tea team account", err)
		respondWithError(w, http.StatusInternalServerError, "获取团队星茶账户失败")
		return
	}
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("账户已被冻结：%s", account.FrozenReason))
		return
	}

	page, limit := getPaginationParams(r)
	transfers, err := dao.TeaTeamPendingApprovalToUserTransferOuts(teamId, page, limit, r.Context())
	if err != nil {
		util.Debug("cannot get pending approve team to user transfer outs", err)
		respondWithError(w, http.StatusInternalServerError, "获取待审批转账记录失败")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedPendingApproveTransfer struct {
		dao.TeaTeamToUserTransferOut
		IsExpired     bool
		CanApprove    bool
		TimeRemaining string
	}
	var enhancedTransfers []EnhancedPendingApproveTransfer
	for _, t := range transfers {
		// 计算过期时间,使用UTC时间比较
		expiresAtUTC := t.ExpiresAt.UTC()
		nowUTC := time.Now().UTC()
		enhanced := EnhancedPendingApproveTransfer{
			TeaTeamToUserTransferOut: t,
			IsExpired:                nowUTC.After(expiresAtUTC),
			CanApprove:               !expiresAtUTC.Before(nowUTC),
		}

		if enhanced.CanApprove {
			timeRemaining := expiresAtUTC.Sub(nowUTC)
			if timeRemaining > time.Hour {
				hours := int(timeRemaining.Hours())
				minutes := int(timeRemaining.Minutes()) % 60
				if minutes > 0 {
					enhanced.TimeRemaining = fmt.Sprintf("%d小时%d分钟", hours, minutes)
				} else {
					enhanced.TimeRemaining = fmt.Sprintf("%d小时", hours)
				}
			} else if timeRemaining > time.Minute {
				minutes := int(timeRemaining.Minutes())
				enhanced.TimeRemaining = fmt.Sprintf("%d分钟", minutes)
			} else {
				enhanced.TimeRemaining = "即将过期"
			}
		} else {
			enhanced.TimeRemaining = "已过期"
		}

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser       dao.User
		Team           *dao.Team
		IsTeamMember   bool
		IsCoreMember   bool
		TeaTeamAccount dao.TeaTeamAccount
		Transfers      []EnhancedPendingApproveTransfer
		StatusDisplay  string
		CurrentPage    int
		Limit          int
	}

	pageData.SessUser = user
	pageData.Team = &team
	pageData.IsTeamMember = isMember
	pageData.IsCoreMember = isCoreMember
	pageData.TeaTeamAccount = account
	pageData.Transfers = enhancedTransfers
	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.pending_approve_to_user_transfers")
}

// GetTeaTeamToTeamCompletedTransfers 获取团队对团队转账已完成状态记录页面 0211
func GetTeaTeamToTeamCompletedTransfers(w http.ResponseWriter, r *http.Request) {
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
	if teamId == dao.TeamIdFreelancer {
		respondWithError(w, http.StatusForbidden, "自由人团队不能查看团队转账纪录")
		return
	}
	team, err := dao.GetTeamByID(teamIdStr)
	if err != nil {
		util.Debug("cannot get team by id:", teamIdStr, err)
		report(w, user, "团队资料缺失")
		return
	}
	// 获取团队帐户
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get tea team account by id", err)
		report(w, user, "团队帐户资料失踪")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看团队转账纪录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取团队对团队已经完成状态交易记录
	transfers, err := dao.TeaTeamToTeamCompletedTransferOuts(teamId, page, limit, r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队对团队已完成状态转账纪录失败")
		return
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser      dao.User
		TeamAccount   dao.TeaTeamAccount
		Team          *dao.Team
		Transfers     []dao.TeaTeamToTeamTransferOut
		CurrentPage   int
		Limit         int
		StatusDisplay string
	}
	pageData.SessUser = user
	pageData.Team = &team
	pageData.TeamAccount = account
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		pageData.StatusDisplay = "已冻结"
	} else {
		pageData.StatusDisplay = "正常"
	}

	// 生成页面
	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.to_team_completed_transfers")

}

// GetTeaTeamToUserCompletedTransfers 获取团队对用户转账已完成状态记录页面 0213
func GetTeaTeamToUserCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}
	// 获取团队ID
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
	if teamId == dao.TeamIdFreelancer {
		respondWithError(w, http.StatusForbidden, "自由人团队不能查看团队转账纪录")
		return
	}
	team, err := dao.GetTeamByID(teamIdStr)
	if err != nil {
		util.Debug("cannot get team by id:", teamIdStr, err)
		report(w, s_u, "团队资料缺失")
		return
	}
	// 获取团队帐户
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get tea team account by id", err)
		report(w, s_u, "团队帐户资料失踪")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看团队转账纪录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取团队对用户已经完成状态交易记录
	transfers, err := dao.TeaTeamToUserCompletedTransferOuts(teamId, page, limit, r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队对用户已完成状态转账纪录失败")
		return
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser      dao.User
		TeamAccount   dao.TeaTeamAccount
		Team          *dao.Team
		Transfers     []dao.TeaTeamToUserTransferOut
		CurrentPage   int
		Limit         int
		StatusDisplay string
	}
	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = account
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		pageData.StatusDisplay = "已冻结"
	} else {
		pageData.StatusDisplay = "正常"
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.to_user_completed_transfers")
}

// GetTeaTeamToTeamOutstandingTransfers 获取团队对团队转账未达成状态记录页面 0222
func GetTeaTeamToTeamOutstandingTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
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
	if teamId == dao.TeamIdFreelancer {
		respondWithError(w, http.StatusForbidden, "自由人团队不能查看团队转账纪录")
		return
	}
	team, err := dao.GetTeamByID(teamIdStr)
	if err != nil {
		util.Debug("cannot get team by id:", teamIdStr, err)
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get tea team account by id", err)
		respondWithError(w, http.StatusInternalServerError, "获取团队帐户失败")
		return
	}

	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看团队转账纪录")
		return
	}

	page, limit := getPaginationParams(r)

	transfers, err := dao.TeaTeamToTeamOutstandingTransferOuts(teamId, page, limit, r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队对团队未达成状态转账纪录失败")
		return
	}

	var pageData struct {
		SessUser      dao.User
		TeamAccount   dao.TeaTeamAccount
		Team          *dao.Team
		Transfers     []dao.TeaTeamToTeamTransferOut
		CurrentPage   int
		Limit         int
		StatusDisplay string
	}
	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = account
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		pageData.StatusDisplay = "已冻结"
	} else {
		pageData.StatusDisplay = "正常"
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.to_team_outstanding_transfers")
}

// GetTeaTeamToUserOutstandingTransfers 获取团队对用户转账未达成状态记录页面 0222
func GetTeaTeamToUserOutstandingTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
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
	if teamId == dao.TeamIdFreelancer {
		respondWithError(w, http.StatusForbidden, "自由人团队不能查看团队转账纪录")
		return
	}
	team, err := dao.GetTeamByID(teamIdStr)
	if err != nil {
		util.Debug("cannot get team by id:", teamIdStr, err)
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get tea team account by id", err)
		respondWithError(w, http.StatusInternalServerError, "获取团队帐户失败")
		return
	}

	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		respondWithError(w, http.StatusForbidden, "您不是该团队成员，无法查看团队转账纪录")
		return
	}

	page, limit := getPaginationParams(r)

	transfers, err := dao.TeaTeamToUserOutstandingTransferOuts(teamId, page, limit, r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队对用户未达成状态转账纪录失败")
		return
	}

	var pageData struct {
		SessUser      dao.User
		TeamAccount   dao.TeaTeamAccount
		Team          *dao.Team
		Transfers     []dao.TeaTeamToUserTransferOut
		CurrentPage   int
		Limit         int
		StatusDisplay string
	}
	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = account
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		pageData.StatusDisplay = "已冻结"
	} else {
		pageData.StatusDisplay = "正常"
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.to_user_outstanding_transfers")
}

// ============================================
// 团队转入页面路由处理函数(incoming transfer)
// ============================================
// GetTeaTeamPendingFromTeamTransfers 获取团队待确认（包含已经超时）转入转账页面 0223
func GetTeaTeamPendingFromTeamTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
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
	transfers, err := dao.TeaTeamPendingFromTeamTransfers(teamId, page, limit, r.Context())
	if err != nil {
		util.Debug("cannot get pending incoming transfers", err)
		report(w, s_u, "获取待确认转账失败。")
		return
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser    dao.User
		Team        *dao.Team
		TeamAccount dao.TeaTeamAccount
		Transfers   []dao.TeaTeamToTeamTransferOut
		CurrentPage int
		Limit       int
	}

	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.pending_from_team_transfers")
}

// GetTeaTeamPendingFromUserTransfers 获取团队待确认（包含已经超时）转入用户转账页面 0224
func GetTeaTeamPendingFromUserTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
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
	transfers, err := dao.TeaTeamPendingFromUserTransfers(teamId, page, limit, r.Context())
	if err != nil {
		util.Debug("cannot get pending from user transfers", err)
		report(w, s_u, "获取待确认转账失败。")
		return
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser    dao.User
		Team        *dao.Team
		TeamAccount dao.TeaTeamAccount
		Transfers   []dao.TeaUserToTeamTransferOut
		CurrentPage int
		Limit       int
	}

	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.pending_from_user_transfers")
}

// GetTeaTeamFromTeamCompletedTransfers 获取团队接收团队转账已完成状态记录页面 0225
func GetTeaTeamFromTeamCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取团队ID
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
	if teamId == dao.TeamIdFreelancer {
		report(w, s_u, "自由人团队不能查看团队转账纪录。")
		return
	}
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, s_u, "团队不存在。")
		return
	}
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get team tea account by id", err)
		report(w, s_u, "获取团队帐户失败。")
		return
	}

	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		report(w, s_u, "您不是该团队成员，无法查看团队转账纪录。")
		return
	}

	page, limit := getPaginationParams(r)

	transfers, err := dao.TeaTeamFromTeamCompletedTransfers(teamId, page, limit, r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队接收团队已完成状态转账纪录失败。")
		return
	}

	var pageData struct {
		SessUser      dao.User
		TeamAccount   dao.TeaTeamAccount
		Team          *dao.Team
		Transfers     []dao.TeaTeamFromTeamTransferIn
		CurrentPage   int
		Limit         int
		StatusDisplay string
	}
	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = account
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		pageData.StatusDisplay = "已冻结"
	} else {
		pageData.StatusDisplay = "正常"
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.from_team_completed_transfers")
}

// GetTeaTeamFromUserCompletedTransfers 获取团队接收用户转账已完成状态记录页面 0225
func GetTeaTeamFromUserCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取团队ID
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
	if teamId == dao.TeamIdFreelancer {
		report(w, s_u, "自由人团队不能查看团队转账纪录。")
		return
	}
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, s_u, "团队不存在。")
		return
	}
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get team tea account by id", err)
		report(w, s_u, "获取团队帐户失败。")
		return
	}

	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		report(w, s_u, "您不是该团队成员，无法查看团队转账纪录。")
		return
	}

	page, limit := getPaginationParams(r)

	transfers, err := dao.TeaTeamFromUserCompletedTransfers(teamId, page, limit, r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队接收用户已完成状态转账纪录失败。")
		return
	}

	var pageData struct {
		SessUser      dao.User
		TeamAccount   dao.TeaTeamAccount
		Team          *dao.Team
		Transfers     []dao.TeaTeamFromUserTransferIn
		CurrentPage   int
		Limit         int
		StatusDisplay string
	}
	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = account
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		pageData.StatusDisplay = "已冻结"
	} else {
		pageData.StatusDisplay = "正常"
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.from_user_completed_transfers")
}

// GetTeaTeamFromTeamRejectedTransfers 获取团队接收团队转账未完成状态记录页面 0226
func GetTeaTeamFromTeamRejectedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取团队ID
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
	if teamId == dao.TeamIdFreelancer {
		report(w, s_u, "自由人团队不能查看团队转账纪录。")
		return
	}
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, s_u, "团队不存在。")
		return
	}
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get team tea account by id", err)
		report(w, s_u, "获取团队帐户失败。")
		return
	}

	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		util.Debug("check user is team member error:", err)
		report(w, s_u, "您不是该团队成员，无法查看团队转账纪录。")
		return
	}

	page, limit := getPaginationParams(r)

	transfers, err := dao.TeaTeamFromTeamRejectedTransfers(teamId, page, limit, r.Context())
	if err != nil {
		util.Debug("cannot get team from team rejected transfers", err)
		report(w, s_u, "获取团队接收团队已拒绝状态转账纪录失败。")
		return
	}
	var pageData struct {
		SessUser      dao.User
		TeamAccount   dao.TeaTeamAccount
		Team          *dao.Team
		Transfers     []dao.TeaTeamFromTeamTransferIn
		CurrentPage   int
		Limit         int
		StatusDisplay string
	}
	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = account
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		pageData.StatusDisplay = "已冻结"
	} else {
		pageData.StatusDisplay = "正常"
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.from_team_rejected_transfers")

}

// GetTeaTeamFromUserRejectedTransfers()  团队接收用户已拒绝记录页面
func GetTeaTeamFromUserRejectedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 验证用户登录
	s_u, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取团队ID
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
	if teamId == dao.TeamIdFreelancer {
		report(w, s_u, "自由人团队不能查看团队转账纪录。")
		return
	}
	team, err := dao.GetTeam(teamId)
	if err != nil {
		util.Debug("cannot get team by id", teamId, err)
		report(w, s_u, "团队不存在。")
		return
	}
	account, err := dao.GetTeaTeamAccountByTeamId(teamId)
	if err != nil {
		util.Debug("cannot get team tea account by id", err)
		report(w, s_u, "获取团队帐户失败。")
		return
	}

	isMember, err := dao.IsTeamActiveMember(s_u.Id, teamId)
	if err != nil || !isMember {
		report(w, s_u, "您不是该团队成员，无法查看团队转账纪录。")
		return
	}

	page, limit := getPaginationParams(r)

	transfers, err := dao.TeaTeamFromUserRejectedTransfers(teamId, page, limit, r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队接收用户已拒绝状态转账纪录失败。")
		return
	}

	var pageData struct {
		SessUser      dao.User
		TeamAccount   dao.TeaTeamAccount
		Team          *dao.Team
		Transfers     []dao.TeaTeamFromUserTransferIn
		CurrentPage   int
		Limit         int
		StatusDisplay string
	}
	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = account
	pageData.Transfers = transfers
	pageData.CurrentPage = page
	pageData.Limit = limit
	if account.Status == dao.TeaTeamAccountStatus_Frozen {
		pageData.StatusDisplay = "已冻结"
	} else {
		pageData.StatusDisplay = "正常"
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.from_user_rejected_transfers")
}
