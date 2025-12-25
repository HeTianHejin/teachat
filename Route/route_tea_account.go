package route

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	dao "teachat/DAO"
	util "teachat/Util"
	"time"
)

// 茶叶用户账户相关响应结构体
type TeaUsrAccountResponse struct {
	Uuid         string  `json:"uuid"`
	UserId       int     `json:"user_id"`
	BalanceGrams float64 `json:"balance_grams"`
	Status       string  `json:"status"`
	FrozenReason *string `json:"frozen_reason,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// 转账请求结构体
type CreateTransferRequest struct {
	ToUserId    int     `json:"to_user_id,omitempty"`
	ToTeamId    int     `json:"to_team_id,omitempty"`
	AmountGrams float64 `json:"amount_grams"`
	Notes       string  `json:"notes"`
	ExpireHours int     `json:"expire_hours"`
}

// 用户对用户转账响应结构体
type UserToUserTransferResponse struct {
	Uuid            string  `json:"uuid"`
	FromUserId      int     `json:"from_user_id"`
	ToUserId        int     `json:"to_user_id"`
	AmountGrams     float64 `json:"amount_grams"`
	Status          string  `json:"status"`
	PaymentTime     *string `json:"payment_time,omitempty"`
	Notes           string  `json:"notes"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
	ExpiresAt       string  `json:"expires_at"`
	CreatedAt       string  `json:"created_at"`
	FromUserName    string  `json:"from_user_name,omitempty"`
	ToUserName      string  `json:"to_user_name,omitempty"`
}

// 用户对团队转账响应结构体
type UserToTeamTransferResponse struct {
	Uuid            string  `json:"uuid"`
	FromUserId      int     `json:"from_user_id"`
	ToTeamId        int     `json:"to_team_id"`
	AmountGrams     float64 `json:"amount_grams"`
	Status          string  `json:"status"`
	PaymentTime     *string `json:"payment_time,omitempty"`
	Notes           string  `json:"notes"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
	ExpiresAt       string  `json:"expires_at"`
	CreatedAt       string  `json:"created_at"`
	FromUserName    string  `json:"from_user_name,omitempty"`
	ToTeamName      string  `json:"to_team_name,omitempty"`
}

// 用户对用户转账接收响应结构体
type UserFromUserTransferResponse struct {
	Uuid                    string  `json:"uuid"`
	UserToUserTransferOutId int     `json:"user_to_user_transfer_out_id"`
	ToUserId                int     `json:"to_user_id"`
	ToUserName              string  `json:"to_user_name"`
	FromUserId              int     `json:"from_user_id"`
	FromUserName            string  `json:"from_user_name"`
	AmountGrams             float64 `json:"amount_grams"`
	BalanceAfterReceipt     float64 `json:"balance_after_receipt"`
	Status                  string  `json:"status"`
	IsConfirmed             bool    `json:"is_confirmed"`
	OperationalUserId       int     `json:"operational_user_id"`
	RejectionReason         string  `json:"rejection_reason,omitempty"`
	ExpiresAt               string  `json:"expires_at"`
	CreatedAt               string  `json:"created_at"`
}

// 用户对团队转账接收响应结构体
type UserFromTeamTransferResponse struct {
	Uuid                    string  `json:"uuid"`
	TeamToUserTransferOutId int     `json:"team_to_user_transfer_out_id"`
	ToUserId                int     `json:"to_user_id"`
	ToUserName              string  `json:"to_user_name"`
	FromTeamId              int     `json:"from_team_id"`
	FromTeamName            string  `json:"from_team_name"`
	AmountGrams             float64 `json:"amount_grams"`
	BalanceAfterReceipt     float64 `json:"balance_after_receipt"`
	Status                  string  `json:"status"`
	IsConfirmed             bool    `json:"is_confirmed"`
	RejectionReason         *string `json:"rejection_reason,omitempty"`
	ExpiresAt               string  `json:"expires_at"`
	CreatedAt               string  `json:"created_at"`
}

// 交易历史响应结构体
type TransactionHistoryResponse struct {
	TransactionType string  `json:"transaction_type"` // "incoming" 或 "outgoing"
	Uuid            string  `json:"uuid"`
	FromUserId      int     `json:"from_user_id"`
	ToUserId        int     `json:"to_user_id"`
	ToTeamId        int     `json:"to_team_id"`
	AmountGrams     float64 `json:"amount_grams"`
	Status          string  `json:"status"`
	Notes           string  `json:"notes"`
	PaymentTime     string  `json:"payment_time"`
	CreatedAt       string  `json:"created_at"`
	FromUserName    string  `json:"from_user_name,omitempty"`
	ToUserName      string  `json:"to_user_name,omitempty"`
}

// 通用API响应结构体
type ApiResponse struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Data     any       `json:"data,omitempty"`
	PageInfo *PageInfo `json:"page_info,omitempty"`
}

// 分页信息结构体
type PageInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// GetTeaUserAccount 获取用户茶叶账户信息
func GetTeaUserAccount(w http.ResponseWriter, r *http.Request) {
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

	response := TeaUsrAccountResponse{
		Uuid:         account.Uuid,
		UserId:       account.UserId,
		BalanceGrams: account.BalanceGrams,
		Status:       account.Status,
		FrozenReason: account.FrozenReason,
		CreatedAt:    account.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	respondWithSuccess(w, "获取账户信息成功", response)
}

// CreateTeaUserToUserTransferAPI 发起用户对用户茶叶转账
func CreateTeaUserToUserTransferAPI(w http.ResponseWriter, r *http.Request) {
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
		ToUserId    int     `json:"to_user_id"`
		AmountGrams float64 `json:"amount_grams"`
		Notes       string  `json:"notes"`
		ExpireHours int     `json:"expire_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	// 验证参数
	if req.ToUserId <= 0 {
		respondWithError(w, http.StatusBadRequest, "必须指定接收方用户ID")
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

	// 创建用户对用户转账
	transfer, err := dao.CreateTeaTransferUserToUser(user.Id, req.ToUserId, req.AmountGrams, req.Notes, req.ExpireHours)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 获取用户名信息
	fromUser, _ := dao.GetUser(user.Id)
	toUser, _ := dao.GetUser(req.ToUserId)

	response := UserToUserTransferResponse{
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

	respondWithSuccess(w, "用户对用户转账发起成功", response)
}

// CreateTeaUserToTeamTransferAPI 发起用户对团队茶叶转账
func CreateTeaUserToTeamTransferAPI(w http.ResponseWriter, r *http.Request) {
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
		ToTeamId    int     `json:"to_team_id"`
		AmountGrams float64 `json:"amount_grams"`
		Notes       string  `json:"notes"`
		ExpireHours int     `json:"expire_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	// 验证参数
	if req.ToTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "必须指定接收方团队ID")
		return
	}
	if req.AmountGrams <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账金额必须大于0")
		return
	}
	if req.ExpireHours <= 0 || req.ExpireHours > 168 { // 最多7天
		req.ExpireHours = 24 // 默认24小时
	}

	// 检查是否向自由人团队转账（自由人团队ID为2）
	if req.ToTeamId == dao.TeamIdFreelancer {
		respondWithError(w, http.StatusBadRequest, "不能向自由人团队转账，自由人团队不支持茶叶资产")
		return
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

	// 创建用户对团队转账
	transfer, err := dao.CreateTeaTransferUserToTeam(user.Id, req.ToTeamId, req.AmountGrams, req.Notes, req.ExpireHours)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 获取用户名和团队名信息
	fromUser, _ := dao.GetUser(user.Id)
	team, _ := dao.GetTeam(req.ToTeamId)

	response := UserToTeamTransferResponse{
		Uuid:        transfer.Uuid,
		FromUserId:  transfer.FromUserId,
		ToTeamId:    transfer.ToTeamId,
		AmountGrams: transfer.AmountGrams,
		Status:      transfer.Status,
		Notes:       transfer.Notes,
		ExpiresAt:   transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
		CreatedAt:   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if fromUser.Name != "" {
		response.FromUserName = fromUser.Name
	}
	if team.Name != "" {
		response.ToTeamName = team.Name
	}

	respondWithSuccess(w, "用户对团队转账发起成功", response)
}

// GetTeaUserPendingUserToUserTransfersAPI 获取用户待确认的用户对用户转账列表
func GetTeaUserPendingUserToUserTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待确认用户对用户转账
	transfers, err := dao.GetPendingTransfers(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待确认转账失败")
		return
	}

	// 转换响应格式
	var responses []UserToUserTransferResponse
	for _, transfer := range transfers {
		// 获取用户名
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		toUser, _ := dao.GetUser(transfer.ToUserId)

		response := UserToUserTransferResponse{
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

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取待确认用户对用户转账成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeaUserFromUserTransferInsAPI 获取用户转入记录（从接收方视角）- 接收历史（所有状态）
func GetTeaUserFromUserTransferInsAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户转入记录（所有状态）
	transfers, err := dao.GetTransferIns(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取转入记录失败")
		return
	}

	// 转换响应格式
	var responses []UserFromUserTransferResponse
	for _, transfer := range transfers {
		response := UserFromUserTransferResponse{
			Uuid:                    transfer.Uuid,
			UserToUserTransferOutId: transfer.UserToUserTransferOutId,
			ToUserId:                transfer.ToUserId,
			ToUserName:              transfer.ToUserName,
			FromUserId:              transfer.FromUserId,
			FromUserName:            transfer.FromUserName,
			AmountGrams:             transfer.AmountGrams,
			BalanceAfterReceipt:     transfer.BalanceAfterReceipt,
			Status:                  transfer.Status,
			IsConfirmed:             transfer.IsConfirmed,
			OperationalUserId:       transfer.OperationalUserId,
			RejectionReason:         transfer.ReceptionRejectionReason.String,
			ExpiresAt:               transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:               transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户转入记录成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeaUserCompletedTransferInsAPI 获取用户已完成的转入记录（从接收方视角）- 收入记录（仅已完成）
func GetTeaUserCompletedTransferInsAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户已完成的转入记录（仅已完成状态）
	transfers, err := dao.GetCompletedTransferIns(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取收入记录失败")
		return
	}

	// 转换响应格式
	var responses []UserFromUserTransferResponse
	for _, transfer := range transfers {
		response := UserFromUserTransferResponse{
			Uuid:                    transfer.Uuid,
			UserToUserTransferOutId: transfer.UserToUserTransferOutId,
			ToUserId:                transfer.ToUserId,
			ToUserName:              transfer.ToUserName,
			FromUserId:              transfer.FromUserId,
			FromUserName:            transfer.FromUserName,
			AmountGrams:             transfer.AmountGrams,
			BalanceAfterReceipt:     transfer.BalanceAfterReceipt,
			Status:                  transfer.Status,
			IsConfirmed:             transfer.IsConfirmed,
			OperationalUserId:       transfer.OperationalUserId,
			RejectionReason:         transfer.ReceptionRejectionReason.String,
			ExpiresAt:               transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:               transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取收入记录成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeaUserToTeamTransferOutsAPI 获取用户对团队转出记录（从转出方视角）
func GetTeaUserToTeamTransferOutsAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对团队转出记录（暂时使用用户对用户转出记录，因为团队功能暂时禁用）
	transfers, err := dao.GetTransferOuts(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对团队转出记录失败")
		return
	}

	// 转换响应格式
	var responses []UserToTeamTransferResponse
	for _, transfer := range transfers {
		// 获取用户名信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)

		response := UserToTeamTransferResponse{
			Uuid:        transfer.Uuid,
			FromUserId:  transfer.FromUserId,
			ToTeamId:    0, // 团队功能暂时禁用，设置为0
			AmountGrams: transfer.AmountGrams,
			Status:      transfer.Status,
			Notes:       transfer.Notes,
			ExpiresAt:   transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if fromUser.Name != "" {
			response.FromUserName = fromUser.Name
		}
		// 团队功能暂时禁用，设置为空
		response.ToTeamName = "团队功能暂不可用"
		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对团队转出记录成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeaUserPendingUserToTeamTransfersAPI 获取用户待确认的用户对团队转账列表
func GetTeaUserPendingUserToTeamTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待确认用户对团队转账（暂时使用用户对用户转账，因为团队功能暂时禁用）
	transfers, err := dao.GetPendingTransfers(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待确认转账失败")
		return
	}

	// 转换响应格式
	var responses []UserToTeamTransferResponse
	for _, transfer := range transfers {
		// 获取用户名信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)

		response := UserToTeamTransferResponse{
			Uuid:        transfer.Uuid,
			FromUserId:  transfer.FromUserId,
			ToTeamId:    0, // 团队功能暂时禁用，设置为0
			AmountGrams: transfer.AmountGrams,
			Status:      transfer.Status,
			Notes:       transfer.Notes,
			ExpiresAt:   transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if fromUser.Name != "" {
			response.FromUserName = fromUser.Name
		}
		// 团队功能暂时禁用，设置为空
		response.ToTeamName = "团队功能暂不可用"

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取待确认用户对团队转账成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeaUserToTeamTransferHistoryAPI 获取用户对团队转账历史（包含转出和转入）
func GetTeaUserToTeamTransferHistoryAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对团队转账历史（暂时使用用户对用户转账历史，因为团队功能暂时禁用）
	transfers, err := dao.GetTransferHistory(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对团队转账历史失败")
		return
	}

	// 转换响应格式
	var responses []UserToTeamTransferResponse
	for _, transfer := range transfers {
		// 获取用户名信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)

		response := UserToTeamTransferResponse{
			Uuid:        transfer.Uuid,
			FromUserId:  transfer.FromUserId,
			ToTeamId:    0, // 团队功能暂时禁用，设置为0
			AmountGrams: transfer.AmountGrams,
			Status:      transfer.Status,
			Notes:       transfer.Notes,
			ExpiresAt:   transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if fromUser.Name != "" {
			response.FromUserName = fromUser.Name
		}
		// 团队功能暂时禁用，设置为空
		response.ToTeamName = "团队功能暂不可用"
		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对团队转账历史成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeaUserFromTeamTransferInsAPI 获取用户从团队转入记录（从接收方视角）
func GetTeaUserFromTeamTransferInsAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户从团队转入记录
	transfers, err := dao.GetTransferIns(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取从团队转入记录失败")
		return
	}

	// 转换响应格式
	var responses []UserFromTeamTransferResponse
	for _, transfer := range transfers {
		response := UserFromTeamTransferResponse{
			Uuid:                    transfer.Uuid,
			TeamToUserTransferOutId: transfer.UserToUserTransferOutId, // 暂时使用相同字段
			ToUserId:                transfer.ToUserId,
			ToUserName:              transfer.ToUserName,
			FromTeamId:              transfer.FromUserId,   // 暂时使用FromUserId作为FromTeamId
			FromTeamName:            transfer.FromUserName, // 暂时使用FromUserName作为FromTeamName
			AmountGrams:             transfer.AmountGrams,
			BalanceAfterReceipt:     transfer.BalanceAfterReceipt,
			Status:                  transfer.Status,
			IsConfirmed:             transfer.IsConfirmed,
			RejectionReason:         getNullableString(transfer.ReceptionRejectionReason),
			ExpiresAt:               transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:               transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户从团队转入记录成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// HandleTeaUserPendingUserToUserTransfers 处理待确认用户对用户转账页面请求
func HandleTeaUserPendingUserToUserTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	PendingUserToUserTransfersGet(w, r)
}

// HandleTeaUserPendingUserToTeamTransfers 处理待确认用户对团队转账页面请求
func HandleTeaUserPendingUserToTeamTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	PendingUserToTeamTransfersGet(w, r)
}

// HandleTeaTeamTransferHistory 处理团队转账历史页面请求
func HandleTeaTeamTransferHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TeamTransferHistoryGet(w, r)
}

// HandleTeaUserFromUserTransferIns 处理用户转入记录页面请求 - 接收历史（所有状态）
func HandleTeaUserFromUserTransferIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserFromUserTransferInsGet(w, r)
}

// HandleTeaUserCompletedTransferIns 处理用户已完成转入记录页面请求 - 收入记录（仅已完成）
func HandleTeaUserCompletedTransferIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserCompletedTransferInsGet(w, r)
}

// HandleTeaUserFromTeamTransferIns 处理用户从团队转入记录页面请求
func HandleTeaUserFromTeamTransferIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserFromTeamTransferInsGet(w, r)
}

// PendingUserToUserTransfersGet 获取待确认用户对用户转账页面
func PendingUserToUserTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有茶叶账户
	err = dao.EnsureTeaAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取用户茶叶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待确认用户对用户转账
	transfers, err := dao.GetPendingTransfers(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending transfers", err)
		report(w, s_u, "获取待确认用户对用户转账失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedPendingTransfer struct {
		dao.TeaUserToUserTransferOut
		FromUserName  string
		ToUserName    string
		AmountDisplay string
		IsExpired     bool
		CanAccept     bool
		TimeRemaining string
	}

	var enhancedTransfers []EnhancedPendingTransfer
	for _, transfer := range transfers {
		enhanced := EnhancedPendingTransfer{
			TeaUserToUserTransferOut: transfer,
			CanAccept:                !transfer.ExpiresAt.Before(time.Now()),
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方用户信息
		toUser, _ := dao.GetUser(transfer.ToUserId)
		if toUser.Id > 0 {
			enhanced.ToUserName = toUser.Name
		}

		// 格式化金额显示
		if transfer.AmountGrams >= 1 {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams, 3) + " 克"
		} else {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams*1000, 0) + " 毫克"
		}

		// 检查是否过期
		enhanced.IsExpired = transfer.ExpiresAt.Before(time.Now())

		// 计算剩余时间
		if enhanced.CanAccept {
			timeRemaining := time.Until(transfer.ExpiresAt)
			if timeRemaining > time.Hour {
				enhanced.TimeRemaining = fmt.Sprintf("%.0f小时", timeRemaining.Hours())
			} else if timeRemaining > time.Minute {
				enhanced.TimeRemaining = fmt.Sprintf("%.0f分钟", timeRemaining.Minutes())
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
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedPendingTransfer
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.TeaAccount = account
	pageData.Transfers = enhancedTransfers

	// 格式化余额显示
	if account.BalanceGrams >= 1 {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams, 2) + " 克"
	} else {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if account.LockedBalanceGrams >= 1 {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams, 2) + " 克"
	} else {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := account.BalanceGrams - account.LockedBalanceGrams
	if availableBalance >= 1 {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.pending_user_to_user_transfers")
}

// GetTeaUserToUserTransferOutsAPI 获取用户对用户转出记录（从转出方视角）
func GetTeaUserToUserTransferOutsAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对用户转出记录
	transfers, err := dao.GetTransferOuts(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对用户转出记录失败")
		return
	}

	// 转换响应格式
	var responses []UserToUserTransferResponse
	for _, transfer := range transfers {
		// 获取用户名
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		toUser, _ := dao.GetUser(transfer.ToUserId)

		response := UserToUserTransferResponse{
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
		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对用户转出记录成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeaUserToUserTransferHistoryAPI 获取用户对用户转账历史（包含转出和转入）
func GetTeaUserToUserTransferHistoryAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对用户转账历史
	transfers, err := dao.GetTransferHistory(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对用户转账历史失败")
		return
	}

	// 转换响应格式
	var responses []UserToUserTransferResponse
	for _, transfer := range transfers {
		// 获取用户名
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		toUser, _ := dao.GetUser(transfer.ToUserId)

		response := UserToUserTransferResponse{
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
		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对用户转账历史成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// HandleTeaUserTransferHistory 处理用户转账历史页面请求
func HandleTeaUserTransferHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TransferHistoryGet(w, r)
}

// TransferHistoryGet 获取转账历史页面
func TransferHistoryGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有茶叶账户
	err = dao.EnsureTeaAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取用户茶叶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取转账历史
	transfers, err := dao.GetTransferHistory(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get transfer history", err)
		report(w, s_u, "获取转账历史失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedTransfer struct {
		dao.TeaUserToUserTransferOut
		FromUserName  string
		ToUserName    string
		ToUserType    string // "user" 或 "team"
		ToUserId      int    // 接收方用户ID（如果是用户转账）
		ToTeamId      int    // 接收方团队ID（如果是团队转账）
		ToTeamUuid    string // 接收方团队UUID（用于链接）
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
		IsIncoming    bool
	}

	var enhancedTransfers []EnhancedTransfer
	for _, transfer := range transfers {
		enhanced := EnhancedTransfer{
			TeaUserToUserTransferOut: transfer,
		}

		// 判断是否为收入（转给当前用户）
		if transfer.ToUserId == s_u.Id {
			enhanced.IsIncoming = true
		} else {
			enhanced.IsIncoming = false
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方用户信息
		toUser, _ := dao.GetUser(transfer.ToUserId)
		if toUser.Id > 0 {
			enhanced.ToUserName = toUser.Name
			enhanced.ToUserType = "user"
			enhanced.ToUserId = transfer.ToUserId
		}

		// 添加状态显示
		switch transfer.Status {
		case dao.TeaTransferStatusPendingApproval:
			enhanced.StatusDisplay = "待确认"
		case dao.TeaTransferStatusApproved, dao.TeaTransferStatusCompleted:
			enhanced.StatusDisplay = "已完成"
		case dao.TeaTransferStatusRejected:
			enhanced.StatusDisplay = "已拒绝"
		case dao.TeaTransferStatusExpired:
			enhanced.StatusDisplay = "已过期"
		default:
			enhanced.StatusDisplay = "未知"
		}

		// 格式化金额显示
		if transfer.AmountGrams >= 1 {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams, 3) + " 克"
		} else {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams*1000, 0) + " 毫克"
		}

		// 检查是否过期
		enhanced.IsExpired = transfer.ExpiresAt.Before(time.Now())

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransfer
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.TeaAccount = account
	pageData.Transfers = enhancedTransfers

	// 格式化余额显示
	if account.BalanceGrams >= 1 {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams, 2) + " 克"
	} else {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if account.LockedBalanceGrams >= 1 {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams, 2) + " 克"
	} else {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := account.BalanceGrams - account.LockedBalanceGrams
	if availableBalance >= 1 {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.transfer_history")
}

// FreezeTeaUserAccount 冻结茶叶账户（管理员功能）
func FreezeTeaUserAccount(w http.ResponseWriter, r *http.Request) {
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

// UnfreezeTeaUserAccount 解冻茶叶账户（管理员功能）
func UnfreezeTeaUserAccount(w http.ResponseWriter, r *http.Request) {
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
func respondWithSuccess(w http.ResponseWriter, message string, data any) {
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
func respondWithPagination(w http.ResponseWriter, message string, data any, page, limit, total int) {
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

// ConfirmTeaUserToUserTransferAPI 确认接收用户对用户转账
func ConfirmTeaUserToUserTransferAPI(w http.ResponseWriter, r *http.Request) {
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

	// 确认接收转账
	err = dao.ConfirmTeaTransfer(req.TransferUuid, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对用户转账确认接收成功", nil)
}

// RejectTeaUserToUserTransferAPI 拒绝接收用户对用户转账
func RejectTeaUserToUserTransferAPI(w http.ResponseWriter, r *http.Request) {
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

	// 拒绝接收转账
	err = dao.RejectTeaTransfer(req.TransferUuid, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对用户转账拒绝成功", nil)
}

// ConfirmTeaUserToTeamTransferAPI 确认接收用户对团队转账
func ConfirmTeaUserToTeamTransferAPI(w http.ResponseWriter, r *http.Request) {
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
		TeamId       int    `json:"team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}

	if req.TeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, req.TeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队成员身份失败")
		return
	}
	if !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能确认接收团队转账")
		return
	}

	// 确认接收转账
	err = dao.ConfirmTeaTransfer(req.TransferUuid, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对团队转账确认接收成功", nil)

}

// RejectTeaUserToTeamTransferAPI 拒绝接收用户对团队转账
func RejectTeaUserToTeamTransferAPI(w http.ResponseWriter, r *http.Request) {
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
		TeamId       int    `json:"team_id"`
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

	if req.TeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamActiveMember(user.Id, req.TeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队成员身份失败")
		return
	}
	if !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能拒绝接收团队转账")
		return
	}

	// 拒绝接收转账
	// 暂时注释掉团队相关功能，待用户对用户功能稳定后再恢复
	respondWithError(w, http.StatusServiceUnavailable, "团队功能暂不可用，待用户对用户功能稳定后恢复")
	/*
		err = dao.RejectTeaTransfer(req.TransferUuid, user.Id, req.Reason)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithSuccess(w, "用户对团队转账拒绝成功", nil)
	*/
}

// ProcessExpiredTransfersJob 定时任务：处理过期转账
func ProcessExpiredTransfersJob() error {
	// 处理用户对用户过期转账
	if err := dao.ProcessUserToUserExpiredTransfers(); err != nil {
		return fmt.Errorf("处理用户对用户过期转账失败: %v", err)
	}

	// 处理用户对团队过期转账
	if err := dao.ProcessUserToTeamExpiredTransfers(); err != nil {
		return fmt.Errorf("处理用户对团队过期转账失败: %v", err)
	}

	// 处理团队对用户过期转账
	if err := dao.ProcessTeamToUserExpiredTransfers(); err != nil {
		return fmt.Errorf("处理团队对用户过期转账失败: %v", err)
	}

	// 处理团队对团队过期转账
	if err := dao.ProcessTeamToTeamExpiredTransfers(); err != nil {
		return fmt.Errorf("处理团队对团队过期转账失败: %v", err)
	}

	return nil
}

// PendingUserToTeamTransfersGet 获取待确认用户对团队转账页面
func PendingUserToTeamTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有茶叶账户
	err = dao.EnsureTeaAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取用户茶叶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待确认用户对团队转账
	transfers, err := dao.GetPendingTransfers(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending transfers", err)
		report(w, s_u, "获取待确认用户对团队转账失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedPendingTransfer struct {
		dao.TeaUserToUserTransferOut
		FromUserName  string
		ToTeamName    string
		AmountDisplay string
		IsExpired     bool
		CanAccept     bool
		TimeRemaining string
	}

	var enhancedTransfers []EnhancedPendingTransfer
	for _, transfer := range transfers {
		enhanced := EnhancedPendingTransfer{
			TeaUserToUserTransferOut: transfer,
			CanAccept:                !transfer.ExpiresAt.Before(time.Now()),
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方团队信息（团队功能暂时禁用）
		// team, _ := dao.GetTeam(transfer.ToTeamId)
		// if team.Id > 0 {
		// 	enhanced.ToTeamName = team.Name
		// }
		enhanced.ToTeamName = "团队功能暂不可用"

		// 格式化金额显示
		if transfer.AmountGrams >= 1 {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams, 3) + " 克"
		} else {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams*1000, 0) + " 毫克"
		}

		// 检查是否过期
		enhanced.IsExpired = transfer.ExpiresAt.Before(time.Now())

		// 计算剩余时间
		if enhanced.CanAccept {
			timeRemaining := time.Until(transfer.ExpiresAt)
			if timeRemaining > time.Hour {
				enhanced.TimeRemaining = fmt.Sprintf("%.0f小时", timeRemaining.Hours())
			} else if timeRemaining > time.Minute {
				enhanced.TimeRemaining = fmt.Sprintf("%.0f分钟", timeRemaining.Minutes())
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
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedPendingTransfer
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.TeaAccount = account
	pageData.Transfers = enhancedTransfers

	// 格式化余额显示
	if account.BalanceGrams >= 1 {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams, 2) + " 克"
	} else {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if account.LockedBalanceGrams >= 1 {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams, 2) + " 克"
	} else {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := account.BalanceGrams - account.LockedBalanceGrams
	if availableBalance >= 1 {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.pending_user_to_team_transfers")
}

// TeamTransferHistoryGet 获取团队转账历史页面
func TeamTransferHistoryGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有茶叶账户
	err = dao.EnsureTeaAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取用户茶叶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取团队转账历史
	transfers, err := dao.GetTransferHistory(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get transfer history", err)
		report(w, s_u, "获取团队转账历史失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedTransfer struct {
		dao.TeaUserToUserTransferOut
		FromUserName  string
		ToTeamName    string
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
		IsIncoming    bool
	}

	var enhancedTransfers []EnhancedTransfer
	for _, transfer := range transfers {
		enhanced := EnhancedTransfer{
			TeaUserToUserTransferOut: transfer,
		}

		// 判断是否为收入（转给当前用户）
		if transfer.ToUserId == s_u.Id {
			enhanced.IsIncoming = true
		} else {
			enhanced.IsIncoming = false
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方团队信息（团队功能暂时禁用）
		// team, _ := dao.GetTeam(transfer.ToTeamId)
		// if team.Id > 0 {
		// 	enhanced.ToTeamName = team.Name
		// }
		enhanced.ToTeamName = "团队功能暂不可用"

		// 添加状态显示
		switch transfer.Status {
		case dao.TeaTransferStatusPendingApproval:
			enhanced.StatusDisplay = "待确认"
		case dao.TeaTransferStatusApproved, dao.TeaTransferStatusCompleted:
			enhanced.StatusDisplay = "已完成"
		case dao.TeaTransferStatusRejected:
			enhanced.StatusDisplay = "已拒绝"
		case dao.TeaTransferStatusExpired:
			enhanced.StatusDisplay = "已过期"
		default:
			enhanced.StatusDisplay = "未知"
		}

		// 格式化金额显示
		if transfer.AmountGrams >= 1 {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams, 3) + " 克"
		} else {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams*1000, 0) + " 毫克"
		}

		// 检查是否过期
		enhanced.IsExpired = transfer.ExpiresAt.Before(time.Now())

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransfer
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.TeaAccount = account
	pageData.Transfers = enhancedTransfers

	// 格式化余额显示
	if account.BalanceGrams >= 1 {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams, 2) + " 克"
	} else {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if account.LockedBalanceGrams >= 1 {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams, 2) + " 克"
	} else {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := account.BalanceGrams - account.LockedBalanceGrams
	if availableBalance >= 1 {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.team.transfer_history")
}

// UserFromUserTransferInsGet 获取用户转入记录页面 - 接收历史（所有状态）
func UserFromUserTransferInsGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有茶叶账户
	err = dao.EnsureTeaAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取用户茶叶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户转入记录（所有状态）
	transfers, err := dao.GetTransferIns(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get transfer ins", err)
		report(w, s_u, "获取用户转入记录失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedTransferIn struct {
		dao.TeaUserFromUserTransferIn
		FromUserName  string
		ToUserName    string
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaUserFromUserTransferIn: transfer,
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方用户信息
		toUser, _ := dao.GetUser(transfer.ToUserId)
		if toUser.Id > 0 {
			enhanced.ToUserName = toUser.Name
		}

		// 添加状态显示
		switch transfer.Status {
		case dao.TeaTransferStatusPendingApproval:
			enhanced.StatusDisplay = "待确认"
		case dao.TeaTransferStatusApproved, dao.TeaTransferStatusCompleted:
			enhanced.StatusDisplay = "已完成"
		case dao.TeaTransferStatusRejected:
			enhanced.StatusDisplay = "已拒绝"
		case dao.TeaTransferStatusExpired:
			enhanced.StatusDisplay = "已过期"
		default:
			enhanced.StatusDisplay = "未知"
		}

		// 格式化金额显示
		if transfer.AmountGrams >= 1 {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams, 3) + " 克"
		} else {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams*1000, 0) + " 毫克"
		}

		// 检查是否过期
		enhanced.IsExpired = transfer.ExpiresAt.Before(time.Now())

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransferIn
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.TeaAccount = account
	pageData.Transfers = enhancedTransfers

	// 格式化余额显示
	if account.BalanceGrams >= 1 {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams, 2) + " 克"
	} else {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if account.LockedBalanceGrams >= 1 {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams, 2) + " 克"
	} else {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := account.BalanceGrams - account.LockedBalanceGrams
	if availableBalance >= 1 {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.from_user_transfer_ins")
}

// UserCompletedTransferInsGet 获取用户已完成转入记录页面 - 收入记录（仅已完成）
func UserCompletedTransferInsGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有茶叶账户
	err = dao.EnsureTeaAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取用户茶叶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户已完成的转入记录（仅已完成状态）
	transfers, err := dao.GetCompletedTransferIns(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get completed transfer ins", err)
		report(w, s_u, "获取收入记录失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedTransferIn struct {
		dao.TeaUserFromUserTransferIn
		FromUserName  string
		ToUserName    string
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaUserFromUserTransferIn: transfer,
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方用户信息
		toUser, _ := dao.GetUser(transfer.ToUserId)
		if toUser.Id > 0 {
			enhanced.ToUserName = toUser.Name
		}

		// 添加状态显示（只有已完成状态）
		enhanced.StatusDisplay = "已完成"

		// 格式化金额显示
		if transfer.AmountGrams >= 1 {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams, 3) + " 克"
		} else {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams*1000, 0) + " 毫克"
		}

		// 检查是否过期（已完成的不会过期）
		enhanced.IsExpired = false

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransferIn
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.TeaAccount = account
	pageData.Transfers = enhancedTransfers

	// 格式化余额显示
	if account.BalanceGrams >= 1 {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams, 2) + " 克"
	} else {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if account.LockedBalanceGrams >= 1 {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams, 2) + " 克"
	} else {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := account.BalanceGrams - account.LockedBalanceGrams
	if availableBalance >= 1 {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.completed_transfer_ins")
}

// UserFromTeamTransferInsGet 获取用户从团队转入记录页面
func UserFromTeamTransferInsGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有茶叶账户
	err = dao.EnsureTeaAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取用户茶叶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取茶叶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户从团队转入记录
	transfers, err := dao.GetTransferIns(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get transfer ins from team", err)
		report(w, s_u, "获取用户从团队转入记录失败。")
		return
	}

	// 增强转账数据，添加团队信息和状态显示
	type EnhancedTransferIn struct {
		dao.TeaUserFromUserTransferIn
		FromTeamName  string
		ToUserName    string
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaUserFromUserTransferIn: transfer,
		}

		// 获取发送方团队信息
		team, _ := dao.GetTeam(transfer.FromUserId) // 暂时使用FromUserId作为团队ID
		if team.Id > 0 {
			enhanced.FromTeamName = team.Name
		}

		// 获取接收方用户信息
		toUser, _ := dao.GetUser(transfer.ToUserId)
		if toUser.Id > 0 {
			enhanced.ToUserName = toUser.Name
		}

		// 添加状态显示
		switch transfer.Status {
		case dao.TeaTransferStatusPendingApproval:
			enhanced.StatusDisplay = "待确认"
		case dao.TeaTransferStatusApproved, dao.TeaTransferStatusCompleted:
			enhanced.StatusDisplay = "已完成"
		case dao.TeaTransferStatusRejected:
			enhanced.StatusDisplay = "已拒绝"
		case dao.TeaTransferStatusExpired:
			enhanced.StatusDisplay = "已过期"
		default:
			enhanced.StatusDisplay = "未知"
		}

		// 格式化金额显示
		if transfer.AmountGrams >= 1 {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams, 3) + " 克"
		} else {
			enhanced.AmountDisplay = util.FormatFloat(transfer.AmountGrams*1000, 0) + " 毫克"
		}

		// 检查是否过期
		enhanced.IsExpired = transfer.ExpiresAt.Before(time.Now())

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransferIn
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.TeaAccount = account
	pageData.Transfers = enhancedTransfers

	// 格式化余额显示
	if account.BalanceGrams >= 1 {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams, 2) + " 克"
	} else {
		pageData.BalanceDisplay = util.FormatFloat(account.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if account.LockedBalanceGrams >= 1 {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams, 2) + " 克"
	} else {
		pageData.LockedBalanceDisplay = util.FormatFloat(account.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := account.BalanceGrams - account.LockedBalanceGrams
	if availableBalance >= 1 {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.from_team_transfer_ins")
}

// 辅助函数：安全获取int值，处理nil和any类型
func safeInt(val any) int {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// 辅助函数：安全获取time.Time值，处理nil和any类型
func safeTime(val any) time.Time {
	if val == nil {
		return time.Time{}
	}
	if t, ok := val.(time.Time); ok {
		return t
	}
	return time.Time{}
}

// 辅助函数：处理sql.NullString，返回*string（有效时返回指针，无效时返回nil）
func getNullableString(nullString sql.NullString) *string {
	if nullString.Valid {
		return &nullString.String
	}
	return nil
}
