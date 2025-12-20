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

// 茶叶账户相关响应结构体
type TeaAccountResponse struct {
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

// 转账响应结构体
type TransferResponse struct {
	Uuid            string  `json:"uuid"`
	FromUserId      int     `json:"from_user_id"`
	ToUserId        int     `json:"to_user_id"`
	ToTeamId        int     `json:"to_team_id"`
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
	TargetUserId    *int    `json:"target_user_id,omitempty"`
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
	if req.ToUserId <= 0 && req.ToTeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "必须指定接收方用户ID或团队ID")
		return
	}
	if req.ToUserId > 0 && req.ToTeamId > 0 {
		respondWithError(w, http.StatusBadRequest, "只能指定一种接收方类型（用户或团队）")
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
	if req.ToTeamId == 2 {
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

	// 创建转账
	var transfer dao.TeaTransfer

	if req.ToUserId > 0 {
		// 用户间转账
		transfer, err = dao.CreateTeaTransfer(user.Id, req.ToUserId, req.AmountGrams, req.Notes, req.ExpireHours)
	} else {
		// 用户向团队转账
		transfer, err = dao.CreateTeaTransferToTeam(user.Id, req.ToTeamId, req.AmountGrams, req.Notes, req.ExpireHours)
	}

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 获取用户名信息
	fromUser, _ := dao.GetUser(user.Id)
	var toUser dao.User
	if req.ToUserId > 0 {
		toUser, _ = dao.GetUser(req.ToUserId)
	}

	response := TransferResponse{
		Uuid:        transfer.Uuid,
		FromUserId:  *transfer.FromUserId,
		ToUserId:    getToUserId(transfer.ToUserId),
		ToTeamId:    getToTeamId(transfer.ToTeamId),
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
		fromUser, _ := dao.GetUser(*transfer.FromUserId)

		response := TransferResponse{
			Uuid:        transfer.Uuid,
			FromUserId:  *transfer.FromUserId,
			ToUserId:    getToUserId(transfer.ToUserId),
			ToTeamId:    getToTeamId(transfer.ToTeamId),
			AmountGrams: transfer.AmountGrams,
			Status:      transfer.Status,
			Notes:       transfer.Notes,
			ExpiresAt:   transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if fromUser.Name != "" {
			response.FromUserName = fromUser.Name
		}

		// 如果是用户间转账，获取接收用户名
		if transfer.ToUserId != nil && *transfer.ToUserId > 0 {
			toUser, _ := dao.GetUser(*transfer.ToUserId)
			if toUser.Id > 0 {
				response.ToUserName = toUser.Name
			}
		} else if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
			// 如果是团队转账，获取团队名称
			team, _ := dao.GetTeam(*transfer.ToTeamId)
			if team.Id > 0 {
				response.ToUserName = team.Name + " (团队)"
			}
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取待确认转账成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// HandlePendingTransfers 处理待确认转账页面请求
func HandlePendingTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	PendingTransfersGet(w, r)
}

// PendingTransfersGet 获取待确认转账页面
func PendingTransfersGet(w http.ResponseWriter, r *http.Request) {
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

	// 获取待确认转账
	transfers, err := dao.GetPendingTransfers(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending transfers", err)
		report(w, s_u, "获取待确认转账失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedPendingTransfer struct {
		dao.TeaTransfer
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
			TeaTransfer: transfer,
			CanAccept:   !transfer.ExpiresAt.Before(time.Now()),
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(*transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方信息（用户或团队）
		if transfer.ToUserId != nil && *transfer.ToUserId > 0 {
			toUser, _ := dao.GetUser(*transfer.ToUserId)
			if toUser.Id > 0 {
				enhanced.ToUserName = toUser.Name
			}
		} else if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
			team, _ := dao.GetTeam(*transfer.ToTeamId)
			if team.Id > 0 {
				enhanced.ToUserName = team.Name + " (团队)"
			}
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
		TeaAccount              dao.TeaAccount
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

	generateHTML(w, &pageData, "layout", "navbar.private", "pending_transfers")
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
		fromUser, _ := dao.GetUser(*transfer.FromUserId)
		var toUser dao.User
		if transfer.ToUserId != nil && *transfer.ToUserId > 0 {
			toUser, _ = dao.GetUser(*transfer.ToUserId)
		}

		response := TransferResponse{
			Uuid:            transfer.Uuid,
			FromUserId:      *transfer.FromUserId,
			ToUserId:        getToUserId(transfer.ToUserId),
			ToTeamId:        getToTeamId(transfer.ToTeamId),
			AmountGrams:     transfer.AmountGrams,
			Status:          transfer.Status,
			Notes:           transfer.Notes,
			RejectionReason: transfer.ReceptionRejectionReason,
			ExpiresAt:       transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:       transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if fromUser.Name != "" {
			response.FromUserName = fromUser.Name
		}
		if toUser.Name != "" {
			response.ToUserName = toUser.Name
		}

		// 如果是团队转账，显示团队名称
		if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
			team, _ := dao.GetTeam(*transfer.ToTeamId)
			if team.Id > 0 {
				response.ToUserName = team.Name + " (团队)"
			}
		}
		if transfer.PaymentTime != nil {
			paymentTime := transfer.PaymentTime.Format("2006-01-02 15:04:05")
			response.PaymentTime = &paymentTime
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取转账历史成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// HandleTransferHistory 处理转账历史页面请求
func HandleTransferHistory(w http.ResponseWriter, r *http.Request) {
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
		dao.TeaTransfer
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
			TeaTransfer: transfer,
		}

		// 判断是否为收入（转给当前用户或转给当前用户所在的团队）
		if transfer.ToUserId != nil && *transfer.ToUserId == s_u.Id {
			enhanced.IsIncoming = true
		} else if transfer.ToTeamId != nil {
			// 检查用户是否是该团队成员
			isMember, _ := dao.IsTeamMember(s_u.Id, *transfer.ToTeamId)
			enhanced.IsIncoming = isMember
		} else {
			enhanced.IsIncoming = false
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(*transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
		}

		// 获取接收方信息（用户或团队）
		if transfer.ToUserId != nil && *transfer.ToUserId > 0 {
			// 用户间转账
			toUser, _ := dao.GetUser(*transfer.ToUserId)
			if toUser.Id > 0 {
				enhanced.ToUserName = toUser.Name
				enhanced.ToUserType = "user"
				enhanced.ToUserId = *transfer.ToUserId
			}
		} else if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
			// 团队转账
			team, _ := dao.GetTeam(*transfer.ToTeamId)
			if team.Id > 0 {
				enhanced.ToUserName = team.Name
				enhanced.ToUserType = "team"
				enhanced.ToTeamId = *transfer.ToTeamId
				enhanced.ToTeamUuid = team.Uuid
			}
		}

		// 添加状态显示
		switch transfer.Status {
		case dao.StatusPendingApproval:
			enhanced.StatusDisplay = "待确认"
		case dao.StatusApproved, dao.StatusCompleted:
			enhanced.StatusDisplay = "已完成"
		case dao.StatusRejected:
			enhanced.StatusDisplay = "已拒绝"
		case dao.StatusExpired:
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
		TeaAccount              dao.TeaAccount
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

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.transacter_history")
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
			TargetUserId:    tx.TargetUserId,
			CreatedAt:       tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取交易流水成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// HandleUserTransactions 处理个人交易流水页面请求
func HandleUserTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserTransactionsGet(w, r)
}

// UserTransactionsGet 获取个人交易流水页面
func UserTransactionsGet(w http.ResponseWriter, r *http.Request) {
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

	// 获取交易类型过滤参数
	transactionType := r.URL.Query().Get("type")

	// 获取交易流水
	transactions, err := dao.GetUserTransactions(s_u.Id, page, limit, transactionType)
	if err != nil {
		util.Debug("cannot get user transactions", err)
		report(w, s_u, "获取交易流水失败。")
		return
	}

	// 增强交易数据，添加目标用户信息和目标团队信息
	type EnhancedTransaction struct {
		dao.TeaTransaction
		TargetUserName string
		TargetTeamName string
		TargetTeamId   int
		TypeDisplay    string
		IsIncome       bool
	}

	var enhancedTransactions []EnhancedTransaction
	for _, tx := range transactions {
		enhanced := EnhancedTransaction{
			TeaTransaction: tx,
		}

		// 获取目标用户信息或目标团队信息
		if tx.TargetUserId != nil {
			// 目标用户相关
			targetUser, _ := dao.GetUser(*tx.TargetUserId)
			if targetUser.Id > 0 {
				enhanced.TargetUserName = targetUser.Name
			}
		} else if tx.TargetTeamId != nil {
			// 目标团队相关
			targetTeam, _ := dao.GetTeam(*tx.TargetTeamId)
			if targetTeam.Id > 0 {
				enhanced.TargetTeamName = targetTeam.Name
				enhanced.TargetTeamId = *tx.TargetTeamId
			}
		}

		// 添加交易类型显示和收入/支出判断
		switch tx.TransactionType {
		case dao.TransactionType_TransferIn:
			enhanced.TypeDisplay = "转入"
			enhanced.IsIncome = true
		case dao.TransactionType_TransferOut:
			enhanced.TypeDisplay = "转出"
			enhanced.IsIncome = false
		case dao.TransactionType_SystemGrant:
			enhanced.TypeDisplay = "系统发放"
			enhanced.IsIncome = true
		case dao.TransactionType_SystemDeduct:
			enhanced.TypeDisplay = "系统扣除"
			enhanced.IsIncome = false
		case dao.TransactionType_Refund:
			enhanced.TypeDisplay = "退款"
			enhanced.IsIncome = true
		default:
			enhanced.TypeDisplay = "未知"
			enhanced.IsIncome = false
		}

		enhancedTransactions = append(enhancedTransactions, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		TeaAccount              dao.TeaAccount
		Transactions            []EnhancedTransaction
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
		FilterType              string
	}

	pageData.SessUser = s_u
	pageData.TeaAccount = account
	pageData.Transactions = enhancedTransactions

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
	pageData.FilterType = transactionType

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.transactions")
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

// 辅助函数：获取ToUserId，处理nil指针
func getToUserId(toUserId *int) int {
	if toUserId == nil {
		return 0
	}
	return *toUserId
}

// 辅助函数：获取ToTeamId，处理nil指针
func getToTeamId(toTeamId *int) int {
	if toTeamId == nil {
		return 0
	}
	return *toTeamId
}

// GetTeamPendingIncomingTransfers 获取团队待确认转入转账列表
func GetTeamPendingIncomingTransfers(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取团队ID
	teamIdStr := r.URL.Query().Get("team_id")
	if teamIdStr == "" {
		respondWithError(w, http.StatusBadRequest, "团队ID不能为空")
		return
	}
	teamId, err := strconv.Atoi(teamIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 检查用户是否是团队成员
	isMember, err := dao.IsTeamMember(user.Id, teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查团队成员身份失败")
		return
	}
	if !isMember {
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看团队待确认转账")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取团队待确认转入转账
	transfers, err := dao.GetPendingTransfersForTeam(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队待确认转账失败")
		return
	}

	// 转换响应格式
	var responses []TransferResponse
	for _, transfer := range transfers {
		// 获取用户名
		fromUser, _ := dao.GetUser(*transfer.FromUserId)

		response := TransferResponse{
			Uuid:        transfer.Uuid,
			FromUserId:  *transfer.FromUserId,
			ToUserId:    getToUserId(transfer.ToUserId),
			ToTeamId:    getToTeamId(transfer.ToTeamId),
			AmountGrams: transfer.AmountGrams,
			Status:      transfer.Status,
			Notes:       transfer.Notes,
			ExpiresAt:   transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if fromUser.Name != "" {
			response.FromUserName = fromUser.Name
		}

		// 如果是团队转账，获取团队名称
		if transfer.ToTeamId != nil && *transfer.ToTeamId > 0 {
			team, _ := dao.GetTeam(*transfer.ToTeamId)
			if team.Id > 0 {
				response.ToUserName = team.Name + " (团队)"
			}
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取团队待确认转账成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// HandleTeamPendingIncomingTransfers 处理团队待确认转入转账页面请求
func HandleTeamPendingIncomingTransfers(w http.ResponseWriter, r *http.Request) {
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
		report(w, s_u, "您不是该团队成员，无法查看待确认转入。")
		return
	}

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

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取团队待确认转入转账
	transfers, err := dao.GetPendingTransfersForTeam(team.Id, page, limit)
	if err != nil {
		util.Debug("cannot get team pending incoming transfers", err)
		report(w, s_u, "获取团队待确认转入失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedPendingIncomingTransfer struct {
		dao.TeaTransfer
		FromUserName  string
		AmountDisplay string
		IsExpired     bool
		CanAccept     bool
		TimeRemaining string
	}

	var enhancedTransfers []EnhancedPendingIncomingTransfer
	for _, transfer := range transfers {
		enhanced := EnhancedPendingIncomingTransfer{
			TeaTransfer: transfer,
			CanAccept:   !transfer.ExpiresAt.Before(time.Now()),
		}

		// 获取发送方用户信息
		fromUser, _ := dao.GetUser(*transfer.FromUserId)
		if fromUser.Id > 0 {
			enhanced.FromUserName = fromUser.Name
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
		Team                    *dao.Team
		TeamAccount             dao.TeamTeaAccount
		Transfers               []EnhancedPendingIncomingTransfer
		BalanceDisplay          string
		LockedBalanceDisplay    string
		AvailableBalanceDisplay string
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Transfers = enhancedTransfers

	// 格式化余额显示
	if teamAccount.BalanceGrams >= 1 {
		pageData.BalanceDisplay = util.FormatFloat(teamAccount.BalanceGrams, 2) + " 克"
	} else {
		pageData.BalanceDisplay = util.FormatFloat(teamAccount.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if teamAccount.LockedBalanceGrams >= 1 {
		pageData.LockedBalanceDisplay = util.FormatFloat(teamAccount.LockedBalanceGrams, 2) + " 克"
	} else {
		pageData.LockedBalanceDisplay = util.FormatFloat(teamAccount.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := teamAccount.BalanceGrams - teamAccount.LockedBalanceGrams
	if availableBalance >= 1 {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		pageData.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if teamAccount.Status == dao.TeamTeaAccountStatus_Frozen {
		if teamAccount.FrozenReason != nil {
			pageData.StatusDisplay = "已冻结 (" + *teamAccount.FrozenReason + ")"
		} else {
			pageData.StatusDisplay = "已冻结"
		}
	} else {
		pageData.StatusDisplay = "正常"
	}

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "team_pending_incoming_transfers")
}

// ProcessExpiredTransfersJob 定时任务：处理过期转账
func ProcessExpiredTransfersJob() error {
	return dao.ProcessExpiredTransfers()
}
