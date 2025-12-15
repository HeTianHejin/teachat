package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	dao "teachat/DAO"
)

// 茶叶账户相关响应结构体
type TeaAccountResponse struct {
	Uuid         string  `json:"uuid"`
	UserId       int     `json:"user_id"`
	BalanceGrams float64 `json:"balance_grams"`
	Status       string  `json:"status"`
	FrozenReason string  `json:"frozen_reason,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// 转账请求结构体
type CreateTransferRequest struct {
	ToUserId    int     `json:"to_user_id"`
	AmountGrams float64 `json:"amount_grams"`
	Notes       string  `json:"notes"`
	ExpireHours int     `json:"expire_hours"`
}

// 转账响应结构体
type TransferResponse struct {
	Uuid            string  `json:"uuid"`
	FromUserId      int     `json:"from_user_id"`
	ToUserId        int     `json:"to_user_id"`
	AmountGrams     float64 `json:"amount_grams"`
	Status          string  `json:"status"`
	PaymentTime     *string `json:"payment_time,omitempty"`
	Notes           string  `json:"notes"`
	RejectionReason *string  `json:"rejection_reason,omitempty"`
	ExpiresAt       string  `json:"expires_at"`
	CreatedAt       string  `json:"created_at"`
	FromUserName    string  `json:"from_user_name,omitempty"`
	ToUserName      string  `json:"to_user_name,omitempty"`
}

// 交易流水响应结构体
type TransactionResponse struct {
	Uuid            string  `json:"uuid"`
	UserId          int     `json:"user_id"`
	TransferId      *string `json:"transfer_id,omitempty"`
	TransactionType string  `json:"transaction_type"`
	AmountGrams     float64 `json:"amount_grams"`
	BalanceBefore   float64 `json:"balance_before"`
	BalanceAfter    float64 `json:"balance_after"`
	Description     string  `json:"description"`
	RelatedUserId   *int    `json:"related_user_id,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// 通用API响应结构体
type ApiResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	PageInfo *PageInfo   `json:"page_info,omitempty"`
}

// 分页信息结构体
type PageInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// GetTeaAccount 获取用户茶叶账户信息
func GetTeaAccount(w http.ResponseWriter, r *http.Request) {
	// 检查是否已经登录
	s, err := session(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	user, err := s.User()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	// 确保用户有茶叶账户
	err = dao.EnsureTeaAccountExists(user.Id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取账户信息失败")
		return
	}

	// 获取账户信息
	account, err := dao.GetTeaAccountByUserId(user.Id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取账户信息失败")
		return
	}

	response := TeaAccountResponse{
		Uuid:         account.Uuid,
		UserId:       account.UserId,
		BalanceGrams: account.BalanceGrams,
		Status:       account.Status,
		FrozenReason: account.FrozenReason,
		CreatedAt:    account.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	respondWithSuccess(w, "获取账户信息成功", response)
}

// CreateTeaTransfer 发起茶叶转账
func CreateTeaTransfer(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 解析请求体
	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	// 验证参数
	if req.ToUserId <= 0 {
		respondWithError(w, http.StatusBadRequest, "接收方用户ID无效")
		return
	}
	if req.AmountGrams <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账金额必须大于0")
		return
	}
	if req.ExpireHours <= 0 || req.ExpireHours > 168 { // 最多7天
		req.ExpireHours = 24 // 默认24小时
	}

	// 检查账户是否被冻结
	frozen, reason, err := dao.CheckAccountFrozen(user.Id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("账户已被冻结: %s", reason))
		return
	}

	// 创建转账
	transfer, err := dao.CreateTeaTransfer(user.Id, req.ToUserId, req.AmountGrams, req.Notes, req.ExpireHours)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 获取用户名信息
	fromUser, _ := dao.GetUser(user.Id)
	toUser, _ := dao.GetUser(req.ToUserId)

	response := TransferResponse{
		Uuid:        transfer.Uuid,
		FromUserId:  transfer.FromUserId,
		ToUserId:    transfer.ToUserId,
		AmountGrams: transfer.AmountGrams,
		Status:      transfer.Status,
		Notes:       transfer.Notes,
		ExpiresAt:   transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
		CreatedAt:   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if fromUser.Name != "" {
		response.FromUserName = fromUser.Name
	}
	if toUser.Name != "" {
		response.ToUserName = toUser.Name
	}

	respondWithSuccess(w, "转账发起成功", response)
}

// ConfirmTeaTransfer 确认接收转账
func ConfirmTeaTransfer(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 解析请求体
	var req struct {
		TransferUuid string `json:"transfer_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}

	// 确认转账
	err = dao.ConfirmTeaTransfer(req.TransferUuid, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "转账确认成功", nil)
}

// RejectTeaTransfer 拒绝转账
func RejectTeaTransfer(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "请求方法错误")
		return
	}

	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 解析请求体
	var req struct {
		TransferUuid string `json:"transfer_uuid"`
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

	// 拒绝转账
	err = dao.RejectTeaTransfer(req.TransferUuid, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "转账拒绝成功", nil)
}

// GetPendingTransfers 获取待确认转账列表
func GetPendingTransfers(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待确认转账
	transfers, err := dao.GetPendingTransfers(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待确认转账失败")
		return
	}

	// 转换响应格式
	var responses []TransferResponse
	for _, transfer := range transfers {
		// 获取用户名
		fromUser, _ := dao.GetUser(transfer.FromUserId)

		response := TransferResponse{
			Uuid:        transfer.Uuid,
			FromUserId:  transfer.FromUserId,
			ToUserId:    transfer.ToUserId,
			AmountGrams: transfer.AmountGrams,
			Status:      transfer.Status,
			Notes:       transfer.Notes,
			ExpiresAt:   transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if fromUser.Name != "" {
			response.FromUserName = fromUser.Name
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取待确认转账成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTransferHistory 获取转账历史
func GetTransferHistory(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取转账历史
	transfers, err := dao.GetTransferHistory(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取转账历史失败")
		return
	}

	// 转换响应格式
	var responses []TransferResponse
	for _, transfer := range transfers {
		// 获取用户名
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		toUser, _ := dao.GetUser(transfer.ToUserId)

		response := TransferResponse{
			Uuid:            transfer.Uuid,
			FromUserId:      transfer.FromUserId,
			ToUserId:        transfer.ToUserId,
			AmountGrams:     transfer.AmountGrams,
			Status:          transfer.Status,
			Notes:           transfer.Notes,
			RejectionReason: transfer.RejectionReason,
			ExpiresAt:       transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:       transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if fromUser.Name != "" {
			response.FromUserName = fromUser.Name
		}
		if toUser.Name != "" {
			response.ToUserName = toUser.Name
		}
		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取转账历史成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetUserTransactions 获取用户交易流水
func GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取交易类型过滤参数
	transactionType := r.URL.Query().Get("type")

	// 获取交易流水
	transactions, err := dao.GetUserTransactions(user.Id, page, limit, transactionType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取交易流水失败")
		return
	}

	// 转换响应格式
	var responses []TransactionResponse
	for _, tx := range transactions {
		response := TransactionResponse{
			Uuid:            tx.Uuid,
			UserId:          tx.UserId,
			TransferId:      tx.TransferId,
			TransactionType: tx.TransactionType,
			AmountGrams:     tx.AmountGrams,
			BalanceBefore:   tx.BalanceBefore,
			BalanceAfter:    tx.BalanceAfter,
			Description:     tx.Description,
			RelatedUserId:   tx.RelatedUserId,
			CreatedAt:       tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取交易流水成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// FreezeTeaAccount 冻结茶叶账户（管理员功能）
func FreezeTeaAccount(w http.ResponseWriter, r *http.Request) {
	// 验证管理员权限
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	if user.Role != dao.User_Role_Captain && user.Role != dao.User_Role_TeaOffice {
		respondWithError(w, http.StatusForbidden, "无权限执行此操作")
		return
	}

	// 解析请求体
	var req struct {
		UserId int    `json:"user_id"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.UserId <= 0 {
		respondWithError(w, http.StatusBadRequest, "用户ID无效")
		return
	}
	if req.Reason == "" {
		respondWithError(w, http.StatusBadRequest, "冻结原因不能为空")
		return
	}

	// 获取账户并冻结
	account, err := dao.GetTeaAccountByUserId(req.UserId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "账户不存在")
		return
	}

	err = account.UpdateStatus(dao.TeaAccountStatus_Frozen, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "冻结账户失败")
		return
	}

	respondWithSuccess(w, "账户冻结成功", nil)
}

// UnfreezeTeaAccount 解冻茶叶账户（管理员功能）
func UnfreezeTeaAccount(w http.ResponseWriter, r *http.Request) {
	// 验证管理员权限
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	if user.Role != dao.User_Role_Captain && user.Role != dao.User_Role_TeaOffice {
		respondWithError(w, http.StatusForbidden, "无权限执行此操作")
		return
	}

	// 解析请求体
	var req struct {
		UserId int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.UserId <= 0 {
		respondWithError(w, http.StatusBadRequest, "用户ID无效")
		return
	}

	// 获取账户并解冻
	account, err := dao.GetTeaAccountByUserId(req.UserId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "账户不存在")
		return
	}

	err = account.UpdateStatus(dao.TeaAccountStatus_Normal, "")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "解冻账户失败")
		return
	}

	respondWithSuccess(w, "账户解冻成功", nil)
}

// 辅助函数：获取分页参数
func getPaginationParams(r *http.Request) (page, limit int) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page = 1
	limit = 20

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

	return page, limit
}

// 辅助函数：返回成功响应
func respondWithSuccess(w http.ResponseWriter, message string, data interface{}) {
	response := ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// 辅助函数：返回错误响应
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	response := ApiResponse{
		Success: false,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// 辅助函数：获取当前用户
func getCurrentUserFromSession(r *http.Request) (dao.User, error) {
	s, err := session(r)
	if err != nil {
		return dao.User{}, err
	}
	return s.User()
}

// 辅助函数：返回分页响应
func respondWithPagination(w http.ResponseWriter, message string, data interface{}, page, limit, total int) {
	totalPages := (total + limit - 1) / limit
	pageInfo := PageInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	response := ApiResponse{
		Success:  true,
		Message:  message,
		Data:     data,
		PageInfo: &pageInfo,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
