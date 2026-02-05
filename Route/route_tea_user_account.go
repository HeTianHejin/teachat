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
	AmountMilligrams int     `json:"amount_milligrams"`
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
	AmountMilligrams int     `json:"amount_milligrams"`
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
	AmountMilligrams        int    `json:"amount_milligrams"`
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
	pendingFromUserCount, err := dao.TeaUserInFromUserPendingTransferOutsCount(s_u.Id)
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

	// 余额显示（仅数值，单位在模板标题栏显示）
	deskData.AccountInfo.BalanceDisplay = fmt.Sprintf("%d", accountInfo.BalanceMilligrams)
	deskData.AccountInfo.LockedBalanceDisplay = fmt.Sprintf("%d", accountInfo.LockedBalanceMilligrams)
	availableBalance := accountInfo.BalanceMilligrams - accountInfo.LockedBalanceMilligrams
	deskData.AccountInfo.AvailableBalanceDisplay = fmt.Sprintf("%d", availableBalance)

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

// FreezeTeaUserAccountAPI 冻结用户星茶账户（管理员功能）
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

// UnfreezeTeaUserAccountAPI 解冻用户星茶账户（管理员功能）
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
		AmountMilligrams: int(transfer.AmountMilligrams),
		Status:           transfer.Status,
		Notes:            transfer.Notes,
		ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
		CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	// 用户对用户转账无需审核，需要等待对方接收后，创建用户对用户转入IN记录

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
		AmountMilligrams: int(transfer.AmountMilligrams),
		Status:           transfer.Status,
		Notes:            transfer.Notes,
		ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
		CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	respondWithSuccess(w, "用户对团队转账发起成功", response)
}

// GetTeaUserInPendingFromUserTransfers 获取用户待确认,来自用户转入记录页面
func GetTeaUserInPendingFromUserTransfers(w http.ResponseWriter, r *http.Request) {
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
	transfers, err := dao.TeaUserInFromUserPendingTransfers(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get transfer ins", err)
		report(w, s_u, "获取用户对用户待确认状态转账记录失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedTransferIn struct {
		dao.TeaUserToUserTransferOut
		FromUserName  string
		ToUserName    string
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
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

		// 金额显示（仅数值，单位在模板标题栏显示）
		enhanced.AmountDisplay = fmt.Sprintf("%d", int(transfer.AmountMilligrams))

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

	// 余额显示（仅数值，单位在模板标题栏显示）
	pageData.BalanceDisplay = fmt.Sprintf("%d", account.BalanceMilligrams)
	pageData.LockedBalanceDisplay = fmt.Sprintf("%d", account.LockedBalanceMilligrams)
	availableBalance := account.BalanceMilligrams - account.LockedBalanceMilligrams
	pageData.AvailableBalanceDisplay = fmt.Sprintf("%d", availableBalance)

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.from_user_pending_transfer_ins")
}

// GetTeaUserInPendingFromUserTransfersAPI 获取用户待确认的,来自用户转账列表API
func GetTeaUserInPendingFromUserTransfersAPI(w http.ResponseWriter, r *http.Request) {
	// 验证用户登录
	user, err := getCurrentUserFromSession(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待确认用户对用户转账
	transfers, err := dao.TeaUserInFromUserPendingTransfers(user.Id, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取用户待确认转账失败")
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
			AmountMilligrams: int(transfer.AmountMilligrams),
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取待确认用户对用户转账成功", responses, page, limit, 0)
}

// GetTeaUserInPendingFromTeamTransfersAPI 获取用户待确认的,来自团队转账列表
func GetTeaUserInPendingFromTeamTransfersAPI(w http.ResponseWriter, r *http.Request) {
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
			AmountMilligrams: int(transfer.AmountMilligrams),
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

// GetTeaUserInPendingFromTeamTransfers 获取用户待确认的,来自团队转账列表页面
func GetTeaUserInPendingFromTeamTransfers(w http.ResponseWriter, r *http.Request) {
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

	// 获取用户待确认状态来自团队转账记录
	transfers, err := dao.TeaUserInFromTeamPendingTransfers(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get transfer ins", err)
		report(w, s_u, "获取团队对用户待确认状态转账记录失败。")
		return
	}

	// 增强转账数据，添加团队信息和状态显示
	type EnhancedTransferIn struct {
		dao.TeaTeamToUserTransferOut
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaTeamToUserTransferOut: transfer,
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

		// 金额显示
		enhanced.AmountDisplay = fmt.Sprintf("%d", int(transfer.AmountMilligrams))

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
			AmountMilligrams: int(transfer.AmountMilligrams),
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
			AmountMilligrams: int(transfer.AmountMilligrams),
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

// GetTeaUserOutPendingToTeamTransfersAPI 获取用户发起的,待团队确认的用户对团队转账列表API
func GetTeaUserOutPendingToTeamTransfersAPI(w http.ResponseWriter, r *http.Request) {
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
			AmountMilligrams: int(transfer.AmountMilligrams),
			Status:           transfer.Status,
			Notes:            transfer.Notes,
			ExpiresAt:        transfer.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:        transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	respondWithPagination(w, "获取用户对团队待确认转账成功", responses, page, limit, 0)
}

// HandleTeaUserToUserPendingTransfers 由当前用户发起,待对方用户确认,转账列表页面请求
func HandleTeaUserToUserPendingTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetTeaUserToUserPendingTransferOuts(w, r)
}

// HandleTeaUserOutToTeamPendingTransfers 用户发起，待对方团队确认接受转账页面请求
func HandleTeaUserOutToTeamPendingTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetTeaUserToTeamPendingTransferOuts(w, r)
}

// HandleTeaUserFromUserCompletedTransferIns 用户已经确认,来自用户转入记录页面请求 - 收入记录（仅已完成）
func HandleTeaUserFromUserCompletedTransferIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetTeaUserFromUserCompletedTransferIns(w, r)
}

// GetTeaUserFromUserCompletedTransferIns 获取用户来自用户已完成转入记录页面
func GetTeaUserFromUserCompletedTransferIns(w http.ResponseWriter, r *http.Request) {
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
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaUserFromUserTransferIn: transfer,
		}

		// 添加状态显示（只有已完成状态）
		enhanced.StatusDisplay = "已完成"
		// 金额显示
		enhanced.AmountDisplay = fmt.Sprintf("%d", int(transfer.AmountMilligrams))

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

	// 余额显示（仅数值，单位在模板标题栏显示）
	pageData.BalanceDisplay = fmt.Sprintf("%d", account.BalanceMilligrams)
	pageData.LockedBalanceDisplay = fmt.Sprintf("%d", account.LockedBalanceMilligrams)
	availableBalance := account.BalanceMilligrams - account.LockedBalanceMilligrams
	pageData.AvailableBalanceDisplay = fmt.Sprintf("%d", availableBalance)

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.from_user_completed_transfer_ins")
}

// HandleTeaUserFromTeamCompletedTransferIns 获取用户已经确认,来自团队转入记录页面请求 - 收入记录（仅已完成）
func HandleTeaUserFromTeamCompletedTransferIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetTeaUserFromTeamCompletedTransferIns(w, r)
}

// GetTeaUserFromTeamCompletedTransferIns 获取用户从团队转入已完成记录页面
func GetTeaUserFromTeamCompletedTransferIns(w http.ResponseWriter, r *http.Request) {
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
		StatusDisplay string
		AmountDisplay string
		IsExpired     bool
	}

	var enhancedTransfers []EnhancedTransferIn
	for _, transfer := range transfers {
		enhanced := EnhancedTransferIn{
			TeaUserFromTeamTransferIn: transfer,
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

		// 金额显示
		enhanced.AmountDisplay = fmt.Sprintf("%d", int(transfer.AmountMilligrams))
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

	// 余额显示（仅数值，单位在模板标题栏显示）
	pageData.BalanceDisplay = fmt.Sprintf("%d", account.BalanceMilligrams)
	pageData.LockedBalanceDisplay = fmt.Sprintf("%d", account.LockedBalanceMilligrams)
	availableBalance := account.BalanceMilligrams - account.LockedBalanceMilligrams
	pageData.AvailableBalanceDisplay = fmt.Sprintf("%d", availableBalance)

	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.from_team_completed_transfer_ins")
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

// ConfirmTeaUserFromUserTransferInAPI 当前用户确认接收,来自用户转账
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
	err = dao.TeaUserConfirmFromUserTransferIn(req.TransferUuid, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对用户转账确认接收成功", nil)
}

// RejectTeaUserFromUserTransferInAPI 当前用户拒绝接收, 来自用户转账
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

	// 拒绝接收用户对用户转账
	err = dao.TeaUserRejectFromUserTransferIn(req.TransferUuid, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对用户转账拒绝成功", nil)
}

// ConfirmTeaUserFromTeamTransferInAPI 当前用户确认接收,来自团队转账
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
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TransferUuid == "" {
		respondWithError(w, http.StatusBadRequest, "转账UUID不能为空")
		return
	}

	// 确认接收来自团队转账
	err = dao.TeaUserConfirmFromTeamTransferIn(req.TransferUuid, user.Id)
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

	// 拒绝接收来自团体转账

	err = dao.TeaUserRejectFromTeamTransferIn(req.TransferUuid, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "用户对团队转账拒绝成功", nil)

}

// GetTeaUserToUserPendingTransferOuts 获取由当前用户发起,待对方用户确认,转账列表页面
func GetTeaUserToUserPendingTransferOuts(w http.ResponseWriter, r *http.Request) {
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

	// 增强转账数据，添加用户信息和状态显示（转出方视角）
	type EnhancedPendingTransfer struct {
		dao.TeaUserToUserTransferOut
		AmountDisplay string
		IsExpired     bool
		TimeRemaining string
	}

	var enhancedTransfers []EnhancedPendingTransfer
	for _, transfer := range transfers {
		enhanced := EnhancedPendingTransfer{
			TeaUserToUserTransferOut: transfer,
		}
		enhanced.AmountDisplay = fmt.Sprintf("%d", int(transfer.AmountMilligrams))

		// 检查是否过期
		enhanced.IsExpired = transfer.ExpiresAt.Before(time.Now())

		// 计算剩余时间
		if !enhanced.IsExpired {
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

	// 账户余额显示（仅数值，单位在模板标题栏显示）
	pageData.BalanceDisplay = fmt.Sprintf("%d", account.BalanceMilligrams)
	pageData.LockedBalanceDisplay = fmt.Sprintf("%d", account.LockedBalanceMilligrams)
	availableBalance := account.BalanceMilligrams - account.LockedBalanceMilligrams
	pageData.AvailableBalanceDisplay = fmt.Sprintf("%d", availableBalance)

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

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.to_user_pending_transfers")
}

// GetTeaUserToTeamPendingTransferOuts 获取当前用户发起,待对方团队确认,转账页面
func GetTeaUserToTeamPendingTransferOuts(w http.ResponseWriter, r *http.Request) {
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

	// 获取待确认,用户对团队转账
	transfers, err := dao.TeaUserOutToTeamPendingTransfers(s_u.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending transfers", err)
		report(w, s_u, "获取用户发起,待对方团体确认转账失败。")
		return
	}

	// 增强转账数据，添加用户信息和状态显示
	type EnhancedPendingTransfer struct {
		dao.TeaUserToTeamTransferOut
		AmountDisplay string
		IsExpired     bool
		CanAccept     bool
		TimeRemaining string
	}

	var enhancedTransfers []EnhancedPendingTransfer
	for _, transfer := range transfers {
		enhanced := EnhancedPendingTransfer{
			TeaUserToTeamTransferOut: transfer,
			CanAccept:                !transfer.ExpiresAt.Before(time.Now()),
		}

		enhanced.AmountDisplay = fmt.Sprintf("%d", int(transfer.AmountMilligrams))

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

// HandleTeaUserToUserCompletedTransfers 获取用户对用户转出已完成记录列表页面(仅已完成状态)
func HandleTeaUserToUserCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetUserToUserCompletedTransfers(w, r)
}

// GetUserToUserCompletedTransfers 获取用户对用户转出已完成记录页面
func GetUserToUserCompletedTransfers(w http.ResponseWriter, r *http.Request) {
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

	// 增强转账数据，添加状态显示和类型转换
	type EnhancedTransferOut struct {
		Id                   int
		Uuid                 string
		FromUserId           int
		FromUserName         string
		ToUserId             int
		ToUserName           string
		AmountMilligrams     int
		Notes                string
		Status               string
		BalanceAfterTransfer int
		ExpiresAt            time.Time
		CreatedAt            time.Time
		UpdatedAt            time.Time
		StatusDisplay        string
	}

	var enhancedTransfers []EnhancedTransferOut
	for _, transfer := range transfers {
		enhanced := EnhancedTransferOut{
			Id:                   transfer.Id,
			Uuid:                 transfer.Uuid,
			FromUserId:           transfer.FromUserId,
			FromUserName:         transfer.FromUserName,
			ToUserId:             transfer.ToUserId,
			ToUserName:           transfer.ToUserName,
			AmountMilligrams:     int(transfer.AmountMilligrams),
			Notes:                transfer.Notes,
			Status:               transfer.Status,
			BalanceAfterTransfer: int(transfer.BalanceAfterTransfer),
			ExpiresAt:            transfer.ExpiresAt,
			CreatedAt:            transfer.CreatedAt,
			StatusDisplay:        "已完成",
		}
		// UpdatedAt 是指针类型，需要判断是否为 nil
		if transfer.UpdatedAt != nil {
			enhanced.UpdatedAt = *transfer.UpdatedAt
		}
		enhancedTransfers = append(enhancedTransfers, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser                dao.User
		BalanceMilligrams       int
		LockedBalanceMilligrams int
		AvailableBalance        int
		Transfers               []EnhancedTransferOut
		StatusDisplay           string
		CurrentPage             int
		Limit                   int
	}

	pageData.SessUser = s_u
	pageData.BalanceMilligrams = int(account.BalanceMilligrams)
	pageData.LockedBalanceMilligrams = int(account.LockedBalanceMilligrams)
	pageData.AvailableBalance = int(account.BalanceMilligrams - account.LockedBalanceMilligrams)
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

	generateHTML(w, &pageData, "layout", "navbar.private", "tea.user.to_user_completed_transfer_outs")
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
			AmountMilligrams: int(transfer.AmountMilligrams),
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

// HandleTeaUserToTeamCompletedTransfers 获取用户对团队转出已完成记录列表页面(仅已完成状态)
func HandleTeaUserToTeamCompletedTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetUserToTeamCompletedTransfers(w, r)
}

// GetUserToTeamCompletedTransfers 获取用户对团队转出已完成记录页面
func GetUserToTeamCompletedTransfers(w http.ResponseWriter, r *http.Request) {
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

// HandleTeaUserToUserExpiredTransfers 获取用户对用户转出已超时记录列表页面(仅已超时状态)
func HandleTeaUserToUserExpiredTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetUserToUserExpiredTransfers(w, r)
}

// GetUserToUserExpiredTransfers 获取用户对用户转出已超时记录页面
func GetUserToUserExpiredTransfers(w http.ResponseWriter, r *http.Request) {
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

// HandleTeaUserToTeamExpiredTransfers 获取用户对团队转出已超时记录列表页面(仅已超时状态)
func HandleTeaUserToTeamExpiredTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetUserToTeamExpiredTransfers(w, r)
}

// GetUserToTeamExpiredTransfers 获取用户对团队转出已超时记录页面
func GetUserToTeamExpiredTransfers(w http.ResponseWriter, r *http.Request) {
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

// TeaUserFromTeamCompletedTransfersAPI 用户已经确认接收团队转入记录API - 收入记录（仅已完成）
func TeaUserFromTeamCompletedTransfersAPI(w http.ResponseWriter, r *http.Request) {
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
			AmountMilligrams:        int(transfer.AmountMilligrams),
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
