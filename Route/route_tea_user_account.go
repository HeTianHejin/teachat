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

// 星茶用户账户相关响应结构体
type TeaUsrAccountResponse struct {
	Uuid              string `json:"uuid"`
	UserId            int    `json:"user_id"`
	BalanceMilligrams int64  `json:"balance_milligrams"`
	Status            string `json:"status"`
	FrozenReason      string `json:"frozen_reason,omitempty"`
	CreatedAt         string `json:"created_at"`
}

// 用户对用户转账响应结构体
type UserToUserTransferOutResponse struct {
	Uuid             string  `json:"uuid"`
	FromUserId       int     `json:"from_user_id"`
	FromUserName     string  `json:"from_user_name,omitempty"`
	ToUserId         int     `json:"to_user_id"`
	ToUserName       string  `json:"to_user_name,omitempty"`
	AmountMilligrams int64   `json:"amount_milligrams"`
	Status           string  `json:"status"`
	PaymentTime      *string `json:"payment_time,omitempty"`
	Notes            string  `json:"notes"`
	RejectionReason  string  `json:"rejection_reason,omitempty"`
	ExpiresAt        string  `json:"expires_at"`
	CreatedAt        string  `json:"created_at"`
}

// 用户对团队转账响应结构体
type UserToTeamTransferResponse struct {
	Uuid             string  `json:"uuid"`
	FromUserId       int     `json:"from_user_id"`
	FromUserName     string  `json:"from_user_name,omitempty"`
	ToTeamId         int     `json:"to_team_id"`
	ToTeamName       string  `json:"to_team_name,omitempty"`
	AmountMilligrams int64   `json:"amount_milligrams"`
	Status           string  `json:"status"`
	PaymentTime      *string `json:"payment_time,omitempty"`
	Notes            string  `json:"notes"`
	RejectionReason  string  `json:"rejection_reason,omitempty"`
	ExpiresAt        string  `json:"expires_at"`
	CreatedAt        string  `json:"created_at"`
}

// 用户来自用户转账接收响应结构体
type UserFromUserTransferInResponse struct {
	Uuid                    string `json:"uuid"`
	UserToUserTransferOutId int    `json:"user_to_user_transfer_out_id"`
	FromUserId              int    `json:"from_user_id"`
	FromUserName            string `json:"from_user_name"`
	ToUserId                int    `json:"to_user_id"`
	ToUserName              string `json:"to_user_name"`
	AmountMilligrams        int64  `json:"amount_milligrams"`
	BalanceAfterReceipt     int64  `json:"balance_after_receipt"`
	Status                  string `json:"status"`
	Notes                   string `json:"notes"`

	IsConfirmed       bool   `json:"is_confirmed"`
	OperationalUserId int    `json:"operational_user_id"`
	RejectionReason   string `json:"rejection_reason,omitempty"`
	ExpiresAt         string `json:"expires_at"`
	CreatedAt         string `json:"created_at"`
}

// 用户来自团队转账接收响应结构体
type UserFromTeamTransferInResponse struct {
	Uuid                    string `json:"uuid"`
	TeamToUserTransferOutId int    `json:"team_to_user_transfer_out_id"`
	FromTeamId              int    `json:"from_team_id"`
	FromTeamName            string `json:"from_team_name"`
	ToUserId                int    `json:"to_user_id"`
	ToUserName              string `json:"to_user_name"`
	AmountMilligrams        int64  `json:"amount_milligrams"`
	BalanceAfterReceipt     int64  `json:"balance_after_receipt"`
	Status                  string `json:"status"`
	Notes                   string `json:"notes"`

	IsConfirmed     bool   `json:"is_confirmed"`
	RejectionReason string `json:"rejection_reason,omitempty"`
	ExpiresAt       string `json:"expires_at"`
	CreatedAt       string `json:"created_at"`
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

// HandleTeaUserAccount 处理用户星茶账户（星茶罐）页面请求
func HandleTeaUserAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TeaUserAcountGet(w, r)
}

// TeaUserAcountGet 获取星茶罐页面
func TeaUserAcountGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		// 不阻止流程，即使账户创建失败也显示页面
	}

	// 获取用户星茶账户信息
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	var accountInfo *dao.TeaUserAccount
	if err == nil {
		accountInfo = &account
	} else {
		// 如果获取失败，创建一个空的账户信息
		accountInfo = &dao.TeaUserAccount{
			UserId:            s_u.Id,
			BalanceMilligrams: 0,
			Status:            dao.TeaAccountStatus_Normal,
		}
	}

	// 获取用户待确认来自用户转账数量
	pendingFromUserCount, err := dao.GetTeaUserToUserPendingTransferOutsCount(s_u.Id)
	if err != nil {
		util.Debug("cannot get pending transfers count", err)
		report(w, s_u, "你好，茶博士失魂鱼，获取您的待确认星茶转账记录失败。")
		return
	}
	// 获取用户待确认来自团队转账数量
	pendingFromTeamCount, err := dao.TeaUserInFromTeamPendingTransferOutsCount(s_u.Id)
	if err != nil {
		util.Debug("cannot get pending team transfers count", err)
		report(w, s_u, "你好，茶博士失魂鱼，获取您的待确认星茶转账记录失败。")
		return
	}

	// 创建星茶罐数据结构
	var deskData struct {
		SessUser    dao.User
		TeaAccount  *dao.TeaUserAccount
		AccountInfo struct {
			BalanceDisplay          string
			LockedBalanceDisplay    string
			AvailableBalanceDisplay string
			StatusDisplay           string
			IsFrozen                bool
		}
		PendingTransferFromUserCount int
		PendingTransferFromTeamCount int
	}

	deskData.SessUser = s_u
	deskData.TeaAccount = accountInfo
	deskData.PendingTransferFromUserCount = pendingFromUserCount
	deskData.PendingTransferFromTeamCount = pendingFromTeamCount

	// 状态显示
	if accountInfo.Status == dao.TeaAccountStatus_Frozen {
		deskData.AccountInfo.StatusDisplay = "已冻结"
		deskData.AccountInfo.IsFrozen = true
	} else {
		deskData.AccountInfo.StatusDisplay = "正常"
		deskData.AccountInfo.IsFrozen = false
	}

	// 余额显示
	deskData.AccountInfo.BalanceDisplay = fmt.Sprintf("%d 毫克", accountInfo.BalanceMilligrams)
	deskData.AccountInfo.LockedBalanceDisplay = fmt.Sprintf("%d 毫克", accountInfo.LockedBalanceMilligrams)
	availableBalance := accountInfo.BalanceMilligrams - accountInfo.LockedBalanceMilligrams
	deskData.AccountInfo.AvailableBalanceDisplay = fmt.Sprintf("%d 毫克", availableBalance)

	generateHTML(w, &deskData, "layout", "navbar.private", "tea.user.account")
}

// GetTeaUserAccountAPI 获取用户星茶账户信息API
func GetTeaUserAccountAPI(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(user.Id)
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
		Uuid:              account.Uuid,
		UserId:            account.UserId,
		BalanceMilligrams: account.BalanceMilligrams,
		Status:            account.Status,
		FrozenReason:      account.FrozenReason,
		CreatedAt:         account.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	respondWithSuccess(w, "获取账户信息成功", response)
}

// CreateTeaUserToUserTransferAPI 发起用户对用户星茶转账
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
		ToUserId         int    `json:"to_user_id"`
		AmountMilligrams int64  `json:"amount_milligrams"`
		Notes            string `json:"notes"`
		ExpireHours      int    `json:"expire_hours"`
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
	if req.AmountMilligrams <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账金额必须大于0")
		return
	}
	if req.ExpireHours <= 0 || req.ExpireHours > 168 { // 最多7天
		req.ExpireHours = 24 // 默认24小时
	}

	// 检查账户是否被冻结
	frozen, reason, err := dao.CheckTeaUserAccountFrozen(user.Id)
	if err != nil {
		util.Debug("CheckTeaUserAccountFrozen error:", err)
		respondWithError(w, http.StatusInternalServerError, "检查账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("账户已被冻结: %s", reason))
		return
	}
	toUser, err := dao.GetUser(req.ToUserId)
	if err != nil {
		util.Debug("GetUser error:", err)
		respondWithError(w, http.StatusBadRequest, "接收方用户不存在")
		return
	}
	if toUser.Id == user.Id {
		respondWithError(w, http.StatusBadRequest, "不能向自己转账")
		return
	}
	// 检查接收方用户账户是否被冻结
	frozen, reason, err = dao.CheckTeaUserAccountFrozen(toUser.Id)
	if err != nil {
		util.Debug("CheckTeaUserAccountFrozen error:", err)
		respondWithError(w, http.StatusInternalServerError, "检查接收方账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("接收方账户已被冻结: %s", reason))
		return
	}

	// 创建转出方用户对用户转账OUT记录
	transfer, err := dao.CreateTeaUserToUserTransferOut(user.Id, user.Name, toUser.Id, toUser.Name, req.AmountMilligrams, req.Notes, req.ExpireHours)
	if err != nil {
		util.Debug("CreateTeaUserToUserTransferOut error:", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := UserToUserTransferOutResponse{
		Uuid:             transfer.Uuid,
		FromUserId:       transfer.FromUserId,
		FromUserName:     transfer.FromUserName,
		ToUserId:         transfer.ToUserId,
		ToUserName:       transfer.ToUserName,
		AmountMilligrams: transfer.AmountMilligrams,
		Status:           transfer.Status,
		Notes:            transfer.Notes,
		ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
		CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	// 用户对用户转账无需审核，需要等待对方接收后，创建用户对用户转入IN记录
	// _, err = dao.CreateTeaUserFromUserTransferIn(transfer.Id, toUser.Id, toUser.Name, transfer.FromUserId, transfer.FromUserName, transfer.AmountMilligrams, transfer.Notes, transfer.ExpiresAt)
	// if err != nil {
	// 	util.Debug("CreateTeaUserFromUserTransferIn error:", err)
	// 	respondWithError(w, http.StatusInternalServerError, "创建接收方转账记录失败")
	// 	return
	// }

	respondWithSuccess(w, "用户对用户转账发起成功", response)
}

// CreateTeaUserToTeamTransferAPI 发起用户对团队星茶转账
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
		ToTeamId         int    `json:"to_team_id"`
		AmountMilligrams int64  `json:"amount_milligrams"`
		Notes            string `json:"notes"`
		ExpireHours      int    `json:"expire_hours"`
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
	if req.AmountMilligrams <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账金额必须大于0")
		return
	}
	if req.ExpireHours <= 0 || req.ExpireHours > 168 { // 最多7天
		req.ExpireHours = 24 // 默认24小时
	}

	// 检查是否向自由人团队转账（自由人团队ID为2）
	if req.ToTeamId == dao.TeamIdFreelancer {
		respondWithError(w, http.StatusBadRequest, "不能向自由人团队转账，自由人团队不支持星茶资产")
		return
	}

	// 检查账户是否被冻结
	frozen, reason, err := dao.CheckTeaUserAccountFrozen(user.Id)
	if err != nil {
		util.Debug("CheckTeaUserAccountFrozen error:", err)
		respondWithError(w, http.StatusInternalServerError, "检查账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("账户已被冻结: %s", reason))
		return
	}
	toTeam, err := dao.GetTeam(req.ToTeamId)
	if err != nil {
		util.Debug("GetTeam error:", err)
		respondWithError(w, http.StatusBadRequest, "接收方团队不存在")
		return
	}
	// 检查接收方团队账户是否被冻结
	frozen, reason, err = dao.CheckTeaTeamAccountFrozen(toTeam.Id)
	if err != nil {
		util.Debug("CheckTeaTeamAccountFrozen error:", err)
		respondWithError(w, http.StatusInternalServerError, "检查接收方团队账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("接收方团队账户已被冻结: %s", reason))
		return
	}

	// 创建用户对团队转账
	transfer, err := dao.CreateTeaUserToTeamTransferOut(user.Id, user.Name, toTeam.Id, toTeam.Name, req.AmountMilligrams, req.Notes, req.ExpireHours)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := UserToTeamTransferResponse{
		Uuid:             transfer.Uuid,
		FromUserId:       transfer.FromUserId,
		FromUserName:     transfer.FromUserName,
		ToTeamId:         transfer.ToTeamId,
		ToTeamName:       transfer.ToTeamName,
		AmountMilligrams: transfer.AmountMilligrams,
		Status:           transfer.Status,
		Notes:            transfer.Notes,
		ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
		CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	respondWithSuccess(w, "用户对团队转账发起成功", response)
}

// GetTeaUserPendingUserToUserTransfersAPI 获取用户待确认（接收）的"用户对用户"转账列表
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
	transfers, err := dao.GetTeaUserToUserPendingTransferOuts(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待确认转账失败")
		return
	}

	// 转换响应格式
	var responses []UserToUserTransferOutResponse
	for _, transfer := range transfers {

		response := UserToUserTransferOutResponse{
			Uuid:             transfer.Uuid,
			FromUserId:       transfer.FromUserId,
			FromUserName:     transfer.FromUserName,
			ToUserId:         transfer.ToUserId,
			ToUserName:       transfer.ToUserName,
			AmountMilligrams: transfer.AmountMilligrams,
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取待确认用户对用户转账成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeaUserFromUserCompletedTransfersAPI 获取用户来自用户已完成的转入记录（从接收方视角）- 收入记录（仅已完成）
func GetTeaUserFromUserCompletedTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户来自用户已完成的转入记录（仅已完成状态）
	transfers, err := dao.TeaUserFromUserCompletedTransferIns(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户已完成来自用户收入记录失败")
		return
	}

	// 转换响应格式
	var responses []UserFromUserTransferInResponse
	for _, transfer := range transfers {
		response := UserFromUserTransferInResponse{
			Uuid:                    transfer.Uuid,
			UserToUserTransferOutId: transfer.UserToUserTransferOutId,
			ToUserId:                transfer.ToUserId,
			ToUserName:              transfer.ToUserName,
			FromUserId:              transfer.FromUserId,
			FromUserName:            transfer.FromUserName,
			AmountMilligrams:        transfer.AmountMilligrams,
			BalanceAfterReceipt:     transfer.BalanceAfterReceipt,
			Status:                  transfer.Status,
			Notes:                   transfer.Notes,
			IsConfirmed:             transfer.IsConfirmed,
			OperationalUserId:       transfer.OperationalUserId,
			RejectionReason:         transfer.RejectionReason,
			ExpiresAt:               transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:               transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户来自用户收入已完成记录成功", responses, page, limit, 0)
}

// GetTeaUserToUserExpiredTransfersAPI 获取用户对用户超时转出记录
func GetTeaUserToUserExpiredTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}
	// 获取分页参数
	page, limit := getPaginationParams(r)
	// 获取用户对用户转出已经过期记录
	transfers, err := dao.TeaUserToUserExpiredTransferOuts(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对用户转出已过期记录失败")
		return
	}
	// 转换响应格式
	var responses []UserToUserTransferOutResponse
	for _, transfer := range transfers {

		response := UserToUserTransferOutResponse{
			Uuid:             transfer.Uuid,
			FromUserId:       transfer.FromUserId,
			FromUserName:     transfer.FromUserName,
			ToUserId:         transfer.ToUserId,
			ToUserName:       transfer.ToUserName,
			AmountMilligrams: transfer.AmountMilligrams,
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}
	respondWithPagination(w, "获取用户对用户已经过期记录成功", responses, page, limit, 0)
}

// GetTeaUserToTeamExpiredTransfersAPI 获取用户对团队超时转出记录
func GetTeaUserToTeamExpiredTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对团队转出已经过期记录
	transfers, err := dao.TeaUserToTeamExpiredTransferOuts(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对团队转出已过期记录失败")
		return
	}

	// 转换响应格式
	var responses []UserToTeamTransferResponse
	for _, transfer := range transfers {

		response := UserToTeamTransferResponse{
			Uuid:             transfer.Uuid,
			FromUserId:       transfer.FromUserId,
			FromUserName:     transfer.FromUserName,
			ToTeamId:         transfer.ToTeamId,
			ToTeamName:       transfer.ToTeamName,
			AmountMilligrams: transfer.AmountMilligrams,
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对团队已经过期记录成功", responses, page, limit, 0)
}

// GetTeaUserPendingUserToTeamTransfersAPI 获取用户发起的待团队确认的用户对团队转账列表
func GetTeaUserPendingUserToTeamTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对团队待确认转账记录
	transfers, err := dao.TeaUserOutToTeamPendingTransfers(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对团队待确认转账失败")
		return
	}

	// 转换响应格式
	var responses []UserToTeamTransferResponse
	for _, transfer := range transfers {
		response := UserToTeamTransferResponse{
			Uuid:             transfer.Uuid,
			FromUserId:       transfer.FromUserId,
			FromUserName:     transfer.FromUserName,
			ToTeamId:         transfer.ToTeamId,
			ToTeamName:       transfer.ToTeamName,
			AmountMilligrams: transfer.AmountMilligrams,
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对团队待确认转账成功", responses, page, limit, 0)
}

// GetTeaUserPendingTeamToUserTransferOutsAPI 获取用户待确认的团队对用户转账列表
func GetTeaUserPendingTeamToUserTransferOutsAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待用户确认，团队对用户转账记录
	transfers, err := dao.TeaUserInFromTeamPendingTransfers(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待确认转账失败")
		return
	}

	// 转换响应格式
	var responses []UserFromTeamTransferInResponse
	for _, transfer := range transfers {
		// 获取用户名信息
		toUser, _ := dao.GetUser(transfer.ToUserId)

		response := UserFromTeamTransferInResponse{
			Uuid:             transfer.Uuid,
			ToUserId:         transfer.ToUserId,
			FromTeamId:       transfer.FromTeamId,
			AmountMilligrams: transfer.AmountMilligrams,
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if toUser.Name != "" {
			response.ToUserName = toUser.Name
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取待确认用户对团队转账成功", responses, page, limit, 0)
}

// GetTeaUserFromTeamTransferInsAPI 获取用户从团队转入已完成记录（从接收方视角）
func GetTeaUserFromTeamTransferInsAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户从团队转入已完成记录
	transfers, err := dao.TeaUserFromTeamCompletedTransferIns(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户从团队转入记录失败")
		return
	}

	// 转换响应格式
	var responses []UserFromTeamTransferInResponse
	for _, transfer := range transfers {
		response := UserFromTeamTransferInResponse{
			Uuid:                    transfer.Uuid,
			TeamToUserTransferOutId: transfer.TeamToUserTransferOutId,
			ToUserId:                transfer.ToUserId,
			ToUserName:              transfer.ToUserName,
			FromTeamId:              transfer.FromTeamId,
			FromTeamName:            transfer.FromTeamName,
			AmountMilligrams:        transfer.AmountMilligrams,
			BalanceAfterReceipt:     transfer.BalanceAfterReceipt,
			Status:                  transfer.Status,
			IsConfirmed:             transfer.IsConfirmed,
			RejectionReason:         transfer.RejectionReason,
			ExpiresAt:               transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:               transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户从团队已完成转入记录成功", responses, page, limit, 0)
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

// HandleTeaUserFromUserTransferIns 处理用户转入记录页面请求 - 接收历史（所有状态）
func HandleTeaUserFromUserTransferIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserFromUserTransferInsGet(w, r)
}

// HandleTeaUserCompletedTransfers 处理用户已完成转入记录页面请求 - 收入记录（仅已完成）
func HandleTeaUserCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserFromUserCompletedTransferInsGet(w, r)
}

// HandleTeaUserFromTeamCompletedTransfers 处理用户从团队转入记录页面请求
func HandleTeaUserFromTeamCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserFromTeamCompletedTransferInsGet(w, r)
}

// PendingUserToUserTransfersGet 获取待对方用户确认,由当前用户发起转账列表页面
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

	// 确保当前用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取当前用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待对方确认用户对用户转账
	transfers, err := dao.TeaUserOutToUserPendingTransfers(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending transfers", err)
		report(w, s_u, "获取待确认用户对用户转账失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedPendingTransfer struct {
		dao.TeaUserToUserTransferOut
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
		enhanced.AmountDisplay = fmt.Sprintf("%d 毫克", transfer.AmountMilligrams)

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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
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

// FreezeTeaUserAccountAPI 冻结星茶账户（管理员功能）
func FreezeTeaUserAccountAPI(w http.ResponseWriter, r *http.Request) {
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

// UnfreezeTeaUserAccountAPI 解冻星茶账户（管理员功能）
func UnfreezeTeaUserAccountAPI(w http.ResponseWriter, r *http.Request) {
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

// ConfirmTeaUserFromUserTransferInAPI 确认接收用户对用户转账
func ConfirmTeaUserFromUserTransferInAPI(w http.ResponseWriter, r *http.Request) {
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
		FromUserId int `json:"from_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.FromUserId <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账用户ID无效")
		return
	}

	// 确认接收转账
	err = dao.TeaUserConfirmFromUserTransferIn(user.Id, req.FromUserId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对用户转账确认接收成功", nil)
}

// RejectTeaUserFromUserTransferInAPI 拒绝接收用户对用户转账
func RejectTeaUserFromUserTransferInAPI(w http.ResponseWriter, r *http.Request) {
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
		FromUserId int    `json:"from_user_id"`
		Reason     string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.FromUserId <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账用户ID无效")
		return
	}

	// 拒绝接收用户对用户转账
	err = dao.TeaUserRejectFromUserTransferIn(req.FromUserId, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对用户转账拒绝成功", nil)
}

// ConfirmTeaUserFromTeamTransferAInPI 当前用户确认接收,来自团队转账
func ConfirmTeaUserFromTeamTransferAInPI(w http.ResponseWriter, r *http.Request) {
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
		FromTeamId int `json:"from_team_id"`
		TeamId     int `json:"team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.FromTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "转账团队ID无效")
		return
	}

	if req.TeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 确认接收来自团队转账
	err = dao.TeaUserConfirmFromTeamTransferIn(user.Id, req.FromTeamId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户接收来自团队转账确认成功", nil)

}

// TeaUserRejectFromTeamTransferInAPI 当前用户拒绝接收,来自团队转账
func TeaUserRejectFromTeamTransferInAPI(w http.ResponseWriter, r *http.Request) {
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
		// TODO:2026-01-31以下方法需要更新
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
	if err := dao.TeaUserProcessToUserExpiredTransfers(); err != nil {
		return fmt.Errorf("处理用户对用户过期转账失败: %v", err)
	}

	// 处理用户对团队过期转账
	if err := dao.TeaUserProcessToTeamExpiredTransfers(); err != nil {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待确认用户对团队转账
	transfers, err := dao.GetTeaUserFromUserPendingTransferIns(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending transfers", err)
		report(w, s_u, "获取待确认用户对团队转账失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedPendingTransfer struct {
		dao.TeaUserFromUserTransferIn
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
			TeaUserFromUserTransferIn: transfer,
			CanAccept:                 !transfer.ExpiresAt.Before(time.Now()),
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
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

// UserFromUserTransferInsGet 获取用户来自用户转入待确认记录页面
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户待确认状态来自用户转账记录
	transfers, err := dao.GetTeaUserFromUserPendingTransferIns(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get transfer ins", err)
		report(w, s_u, "获取用户对用户待确认状态转账记录失败。")
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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
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

// UserFromUserCompletedTransferInsGet 获取用户来自用户已完成转入记录页面 - 收入记录（仅已完成）
func UserFromUserCompletedTransferInsGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户来自用户已完成的转入记录（仅已完成状态）
	transfers, err := dao.TeaUserFromUserCompletedTransferIns(s_u.Id, page, limit)
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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
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

// UserFromTeamCompletedTransferInsGet 获取用户从团队转入已完成记录页面
func UserFromTeamCompletedTransferInsGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户从团队转入已完成状态记录
	transfers, err := dao.TeaUserFromTeamCompletedTransferIns(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get transfer ins from team", err)
		report(w, s_u, "获取用户从团队转入已完成状态转账记录失败。")
		return
	}

	// 增强转账数据，添加团队信息和状态显示
	type EnhancedTransferIn struct {
		dao.TeaUserFromTeamTransferIn
		FromTeamName  string
		ToUserName    string
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaUserFromTeamTransferIn: transfer,
		}

		// 获取发送方团队信息
		team, _ := dao.GetTeam(transfer.FromTeamId)
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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
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

// 辅助函数：安全获取time.Time值，处理nil和any类型
// func safeTime(val any) time.Time {
// 	if val == nil {
// 		return time.Time{}
// 	}
// 	if t, ok := val.(time.Time); ok {
// 		return t
// 	}
// 	return time.Time{}
// }

// GetTeaUserToUserCompletedTransfersAPI 获取用户对用户转出已完成记录列表API(仅已完成状态)
func GetTeaUserToUserCompletedTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对用户转出已完成记录
	transfers, err := dao.TeaUserToUserCompletedTransferOuts(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对用户转出已完成记录失败")
		return
	}

	// 转换响应格式
	var responses []UserToUserTransferOutResponse
	for _, transfer := range transfers {
		response := UserToUserTransferOutResponse{
			Uuid:             transfer.Uuid,
			FromUserId:       transfer.FromUserId,
			FromUserName:     transfer.FromUserName,
			ToUserId:         transfer.ToUserId,
			ToUserName:       transfer.ToUserName,
			AmountMilligrams: transfer.AmountMilligrams,
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对用户转出已完成记录成功", responses, page, limit, 0)
}

// GetTeaUserToUserCompletedTransfers 获取用户对用户转出已完成记录列表页面(仅已完成状态)
func GetTeaUserToUserCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserToUserCompletedTransfersGet(w, r)
}

// UserToUserCompletedTransfersGet 获取用户对用户转出已完成记录页面
func UserToUserCompletedTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对用户转出已完成记录
	transfers, err := dao.TeaUserToUserCompletedTransferOuts(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get completed transfer outs", err)
		report(w, s_u, "获取用户对用户转出已完成记录失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedTransferOut struct {
		dao.TeaUserToUserTransferOut
		FromUserName  string
		ToUserName    string
		StatusDisplay string
		AmountDisplay string
	}

	var enhancedTransfers []EnhancedTransferOut
	for _, transfer := range transfers {
		enhanced := EnhancedTransferOut{
			TeaUserToUserTransferOut: transfer,
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

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransferOut
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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.completed_transfer_outs")
}

// GetTeaUserToTeamCompletedTransfersAPI 获取用户对团队转出已完成记录列表API(仅已完成状态)
func GetTeaUserToTeamCompletedTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对团队转出已完成记录
	transfers, err := dao.TeaUserToTeamCompletedTransferOuts(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户对团队转出已完成记录失败")
		return
	}

	// 转换响应格式
	var responses []UserToTeamTransferResponse
	for _, transfer := range transfers {
		response := UserToTeamTransferResponse{
			Uuid:             transfer.Uuid,
			FromUserId:       transfer.FromUserId,
			FromUserName:     transfer.FromUserName,
			ToTeamId:         transfer.ToTeamId,
			ToTeamName:       transfer.ToTeamName,
			AmountMilligrams: transfer.AmountMilligrams,
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对团队转出已完成记录成功", responses, page, limit, 0)
}

// GetTeaUserToTeamCompletedTransfers 获取用户对团队转出已完成记录列表页面(仅已完成状态)
func GetTeaUserToTeamCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserToTeamCompletedTransfersGet(w, r)
}

// UserToTeamCompletedTransfersGet 获取用户对团队转出已完成记录页面
func UserToTeamCompletedTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对团队转出已完成记录
	transfers, err := dao.TeaUserToTeamCompletedTransferOuts(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get completed transfer outs to team", err)
		report(w, s_u, "获取用户对团队转出已完成记录失败。")
		return
	}

	// 增强转账数据，添加团队信息和状态显示
	type EnhancedTransferOut struct {
		dao.TeaUserToTeamTransferOut
		FromUserName  string
		ToTeamName    string
		StatusDisplay string
		AmountDisplay string
	}

	var enhancedTransfers []EnhancedTransferOut
	for _, transfer := range transfers {
		enhanced := EnhancedTransferOut{
			TeaUserToTeamTransferOut: transfer,
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方团队信息
		toTeam, _ := dao.GetTeam(transfer.ToTeamId)
		if toTeam.Id > 0 {
			enhanced.ToTeamName = toTeam.Name
		}

		// 添加状态显示（只有已完成状态）
		enhanced.StatusDisplay = "已完成"

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransferOut
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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.completed_transfer_outs_to_team")
}

// GetTeaUserToUserExpiredTransfers 获取用户对用户转出已超时记录列表页面(仅已超时状态)
func GetTeaUserToUserExpiredTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserToUserExpiredTransfersGet(w, r)
}

// UserToUserExpiredTransfersGet 获取用户对用户转出已超时记录页面
func UserToUserExpiredTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对用户转出已超时记录
	transfers, err := dao.TeaUserToUserExpiredTransferOuts(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get expired transfer outs", err)
		report(w, s_u, "获取用户对用户转出已超时记录失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedTransferOut struct {
		dao.TeaUserToUserTransferOut
		FromUserName  string
		ToUserName    string
		StatusDisplay string
		AmountDisplay string
	}

	var enhancedTransfers []EnhancedTransferOut
	for _, transfer := range transfers {
		enhanced := EnhancedTransferOut{
			TeaUserToUserTransferOut: transfer,
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

		// 添加状态显示（只有已超时状态）
		enhanced.StatusDisplay = "已超时"

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransferOut
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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.expired_transfer_outs")
}

// GetTeaUserToTeamExpiredTransfers 获取用户对团队转出已超时记录列表页面(仅已超时状态)
func GetTeaUserToTeamExpiredTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserToTeamExpiredTransfersGet(w, r)
}

// UserToTeamExpiredTransfersGet 获取用户对团队转出已超时记录页面
func UserToTeamExpiredTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户对团队转出已超时记录
	transfers, err := dao.TeaUserToTeamExpiredTransferOuts(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get expired transfer outs to team", err)
		report(w, s_u, "获取用户对团队转出已超时记录失败。")
		return
	}

	// 增强转账数据，添加团队信息和状态显示
	type EnhancedTransferOut struct {
		dao.TeaUserToTeamTransferOut
		FromUserName  string
		ToTeamName    string
		StatusDisplay string
		AmountDisplay string
	}

	var enhancedTransfers []EnhancedTransferOut
	for _, transfer := range transfers {
		enhanced := EnhancedTransferOut{
			TeaUserToTeamTransferOut: transfer,
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方团队信息
		toTeam, _ := dao.GetTeam(transfer.ToTeamId)
		if toTeam.Id > 0 {
			enhanced.ToTeamName = toTeam.Name
		}

		// 添加状态显示（只有已超时状态）
		enhanced.StatusDisplay = "已超时"

		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaUserAccount
		Transfers               []EnhancedTransferOut
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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.expired_transfer_outs_to_team")
}

// GetTeaUserFromUserPendingTransfers 等待用户确认接收用户转入记录列表页面 - 待确认状态
func GetTeaUserFromUserPendingTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserFromUserPendingTransfersGet(w, r)
}

// UserFromUserPendingTransfersGet 获取用户来自用户待确认转入记录页面
func UserFromUserPendingTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户来自用户待确认转入记录
	transfers, err := dao.GetTeaUserToUserPendingTransferOuts(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending transfer outs from user", err)
		report(w, s_u, "获取用户来自用户待确认转入记录失败。")
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
		CanAccept     bool
		TimeRemaining string
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaUserFromUserTransferIn: transfer,
			CanAccept:                 !transfer.ExpiresAt.Before(time.Now()),
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
		enhanced.StatusDisplay = "待确认"

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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.pending_transfer_ins_from_user")
}

// GetTeaUserFromTeamPendingTransfersAPI 等待用户确认接收来自团队转入记录API - 待确认状态
func GetTeaUserFromTeamPendingTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户来自团队待确认转入记录
	transfers, err := dao.TeaUserInFromTeamPendingTransfers(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户来自团队待确认转入记录失败")
		return
	}

	// 转换响应格式
	var responses []UserFromTeamTransferInResponse
	for _, transfer := range transfers {
		response := UserFromTeamTransferInResponse{
			Uuid:                    transfer.Uuid,
			TeamToUserTransferOutId: transfer.TeamToUserTransferOutId,
			ToUserId:                transfer.ToUserId,
			ToUserName:              transfer.ToUserName,
			FromTeamId:              transfer.FromTeamId,
			FromTeamName:            transfer.FromTeamName,
			AmountMilligrams:        transfer.AmountMilligrams,
			BalanceAfterReceipt:     transfer.BalanceAfterReceipt,
			Status:                  transfer.Status,
			Notes:                   transfer.Notes,
			IsConfirmed:             transfer.IsConfirmed,
			RejectionReason:         transfer.RejectionReason,
			ExpiresAt:               transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:               transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户来自团队待确认转入记录成功", responses, page, limit, 0)
}

// ConfirmTeaUserFromTeamTransferInAPI 用户确认接收来自团队转账API
func ConfirmTeaUserFromTeamTransferInAPI(w http.ResponseWriter, r *http.Request) {
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

	// 团队(某个成员)确认接收用户对团队转账
	err = dao.TeaUserConfirmFromTeamTransferIn(req.TransferUuid, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对团队转账确认接收成功", nil)
}

// HandleTeaUserFromTeamCompletedTransfersAPI 用户已经确认接收团队转入记录API - 收入记录（仅已完成）
func HandleTeaUserFromTeamCompletedTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户来自团队已完成转入记录
	transfers, err := dao.TeaUserFromTeamCompletedTransferIns(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户来自团队已完成转入记录失败")
		return
	}

	// 转换响应格式
	var responses []UserFromTeamTransferInResponse
	for _, transfer := range transfers {
		response := UserFromTeamTransferInResponse{
			Uuid:                    transfer.Uuid,
			TeamToUserTransferOutId: transfer.TeamToUserTransferOutId,
			ToUserId:                transfer.ToUserId,
			ToUserName:              transfer.ToUserName,
			FromTeamId:              transfer.FromTeamId,
			FromTeamName:            transfer.FromTeamName,
			AmountMilligrams:        transfer.AmountMilligrams,
			BalanceAfterReceipt:     transfer.BalanceAfterReceipt,
			Status:                  transfer.Status,
			Notes:                   transfer.Notes,
			IsConfirmed:             transfer.IsConfirmed,
			RejectionReason:         transfer.RejectionReason,
			ExpiresAt:               transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:               transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户来自团队已完成转入记录成功", responses, page, limit, 0)
}

// HandleTeaUserFromUserPendingTransfers 等待用户确认接收来自团队转入记录页面 - 待确认状态
func HandleTeaUserFromUserPendingTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserFromTeamPendingTransfersGet(w, r)
}

// UserFromTeamPendingTransfersGet 获取用户来自团队待确认转入记录页面
func UserFromTeamPendingTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 确保用户有星茶账户
	err = dao.TeaUserEnsureAccountExists(s_u.Id)
	if err != nil {
		util.Debug("cannot ensure tea account exists", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取用户星茶账户
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	if err != nil {
		util.Debug("cannot get tea account", err)
		report(w, s_u, "获取星茶账户失败。")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取用户来自团队待确认转入记录
	transfers, err := dao.TeaUserInFromTeamPendingTransfers(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending transfer ins from team", err)
		report(w, s_u, "获取用户来自团队待确认转入记录失败。")
		return
	}

	// 增强转账数据，添加团队信息和状态显示
	type EnhancedTransferIn struct {
		dao.TeaUserFromTeamTransferIn
		FromTeamName  string
		ToUserName    string
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
		CanAccept     bool
		TimeRemaining string
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaUserFromTeamTransferIn: transfer,
			CanAccept:                 !transfer.ExpiresAt.Before(time.Now()),
		}

		// 获取发送方团队信息
		fromTeam, _ := dao.GetTeam(transfer.FromTeamId)
		if fromTeam.Id > 0 {
			enhanced.FromTeamName = fromTeam.Name
		}

		// 获取接收方用户信息
		toUser, _ := dao.GetUser(transfer.ToUserId)
		if toUser.Id > 0 {
			enhanced.ToUserName = toUser.Name
		}

		// 添加状态显示
		enhanced.StatusDisplay = "待确认"

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

	// 状态显示
	if account.Status == dao.TeaAccountStatus_Frozen {
		if account.FrozenReason != "-" {
			pageData.StatusDisplay = "已冻结 (" + account.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.pending_transfer_ins_from_team")
}
