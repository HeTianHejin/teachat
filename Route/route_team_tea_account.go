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

// 团队茶叶操作请求结构体
type CreateTeamTeaOperationRequest struct {
	OperationType string  `json:"operation_type"`
	AmountGrams   float64 `json:"amount_grams"`
	Notes         string  `json:"notes"`
	ExpireHours   int     `json:"expire_hours"`
	TargetTeamId  *int    `json:"target_team_id,omitempty"`
	TargetUserId  *int    `json:"target_user_id,omitempty"`
}

// 团队茶叶操作响应结构体
type TeamTeaOperationResponse struct {
	Uuid            string  `json:"uuid"`
	TeamId          int     `json:"team_id"`
	OperationType   string  `json:"operation_type"`
	AmountGrams     float64 `json:"amount_grams"`
	Status          string  `json:"status"`
	OperatorUserId  int     `json:"operator_user_id"`
	OperatorName    string  `json:"operator_name,omitempty"`
	ApproverUserId  *int    `json:"approver_user_id,omitempty"`
	ApproverName    string  `json:"approver_name,omitempty"`
	TargetTeamId    *int    `json:"target_team_id,omitempty"`
	TargetTeamName  string  `json:"target_team_name,omitempty"`
	TargetUserId    *int    `json:"target_user_id,omitempty"`
	TargetUserName  string  `json:"target_user_name,omitempty"`
	Notes           string  `json:"notes"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
	ExpiresAt       string  `json:"expires_at"`
	ApprovedAt      *string `json:"approved_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// 团队茶叶交易流水响应结构体
type TeamTeaTransactionResponse struct {
	Uuid            string  `json:"uuid"`
	TeamId          int     `json:"team_id"`
	OperationId     *string `json:"operation_id,omitempty"`
	TransactionType string  `json:"transaction_type"`
	AmountGrams     float64 `json:"amount_grams"`
	BalanceBefore   float64 `json:"balance_before"`
	BalanceAfter    float64 `json:"balance_after"`
	Description     string  `json:"description"`
	RelatedTeamId   *int    `json:"related_team_id,omitempty"`
	RelatedTeamName string  `json:"related_team_name,omitempty"`
	RelatedUserId   *int    `json:"related_user_id,omitempty"`
	RelatedUserName string  `json:"related_user_name,omitempty"`
	TargetType      string  `json:"target_type"`      // 流通对象类型: u-个人, t-团队
	TargetTypeText  string  `json:"target_type_text"` // 流通对象类型显示文本
	CreatedAt       string  `json:"created_at"`
}

// GetTeamTeaAccount 获取团队茶叶账户信息
func GetTeamTeaAccount(w http.ResponseWriter, r *http.Request) {
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

	// 获取团队ID参数
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
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看团队账户")
		return
	}

	// 确保团队有茶叶账户
	err = dao.EnsureTeamTeaAccountExists(teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取账户信息失败")
		return
	}

	// 获取账户信息
	account, err := dao.GetTeamTeaAccountByTeamId(teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取账户信息失败")
		return
	}

	// 获取团队信息
	team, err := dao.GetTeam(teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取团队信息失败")
		return
	}

	response := TeamTeaAccountResponse{
		Uuid:         account.Uuid,
		TeamId:       account.TeamId,
		TeamName:     team.Name,
		BalanceGrams: account.BalanceGrams,
		Status:       account.Status,
		FrozenReason: account.FrozenReason,
		CreatedAt:    account.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	respondWithSuccess(w, "获取团队账户信息成功", response)
}

// CreateTeamTeaOperation 创建团队茶叶操作
func CreateTeamTeaOperation(w http.ResponseWriter, r *http.Request) {
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
	var req CreateTeamTeaOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	// 验证参数
	if req.AmountGrams <= 0 {
		respondWithError(w, http.StatusBadRequest, "操作金额必须大于0")
		return
	}
	if req.ExpireHours <= 0 || req.ExpireHours > 168 { // 最多7天
		req.ExpireHours = 24 // 默认24小时
	}
	if req.OperationType != dao.TeamOperationType_Deposit &&
		req.OperationType != dao.TeamOperationType_Withdraw &&
		req.OperationType != dao.TeamOperationType_TransferOut &&
		req.OperationType != dao.TeamOperationType_TransferIn {
		respondWithError(w, http.StatusBadRequest, "操作类型无效")
		return
	}

	// 如果是转账操作，需要目标信息
	if req.OperationType == dao.TeamOperationType_TransferOut {
		if req.TargetTeamId == nil && req.TargetUserId == nil {
			respondWithError(w, http.StatusBadRequest, "转出操作需要指定目标团队或用户")
			return
		}

		// 检查是否向自由人团队转账
		if req.TargetTeamId != nil && *req.TargetTeamId == dao.TeamIdFreelancer {
			respondWithError(w, http.StatusBadRequest, "不能向自由人团队转账，自由人团队不支持茶叶资产")
			return
		}
	}

	// 获取团队ID（可以从请求参数或session中获取）
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

	// 检查用户是否可以管理团队账户（核心成员）
	canManage, err := dao.CanUserManageTeamAccount(user.Id, teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查管理权限失败")
		return
	}
	if !canManage {
		respondWithError(w, http.StatusForbidden, "只有核心成员才能创建团队茶叶操作")
		return
	}

	// 检查账户是否被冻结
	frozen, reason, err := dao.CheckTeamAccountFrozen(teamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查账户状态失败")
		return
	}
	if frozen {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("团队账户已被冻结: %s", reason))
		return
	}

	// 创建操作
	operation, err := dao.CreateTeamTeaOperation(teamId, user.Id, req.OperationType, req.AmountGrams, req.Notes, req.ExpireHours, req.TargetTeamId, req.TargetUserId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 获取操作人信息
	operatorUser, _ := dao.GetUser(user.Id)

	response := TeamTeaOperationResponse{
		Uuid:           operation.Uuid,
		TeamId:         operation.TeamId,
		OperationType:  operation.OperationType,
		AmountGrams:    operation.AmountGrams,
		Status:         operation.Status,
		OperatorUserId: operation.OperatorUserId,
		OperatorName:   operatorUser.Name,
		Notes:          operation.Notes,
		ExpiresAt:      operation.ExpiresAt.Format("2006-01-02 15:04:05"),
		CreatedAt:      operation.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 如果操作已被批准，添加审批人信息
	if operation.ApproverUserId != nil {
		approverUser, _ := dao.GetUser(*operation.ApproverUserId)
		if approverUser.Id > 0 {
			response.ApproverUserId = operation.ApproverUserId
			response.ApproverName = approverUser.Name
		}
		if operation.ApprovedAt != nil {
			approvedAt := operation.ApprovedAt.Format("2006-01-02 15:04:05")
			response.ApprovedAt = &approvedAt
		}
	}

	// 获取目标信息
	if operation.TargetTeamId != nil {
		targetTeam, _ := dao.GetTeam(*operation.TargetTeamId)
		if targetTeam.Id > 0 {
			response.TargetTeamId = operation.TargetTeamId
			response.TargetTeamName = targetTeam.Name
		}
	}
	if operation.TargetUserId != nil {
		targetUser, _ := dao.GetUser(*operation.TargetUserId)
		if targetUser.Id > 0 {
			response.TargetUserId = operation.TargetUserId
			response.TargetUserName = targetUser.Name
		}
	}

	respondWithSuccess(w, "团队茶叶操作创建成功", response)
}

// ApproveTeamTeaOperation 审批团队茶叶操作
func ApproveTeamTeaOperation(w http.ResponseWriter, r *http.Request) {
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
		OperationUuid string `json:"operation_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.OperationUuid == "" {
		respondWithError(w, http.StatusBadRequest, "操作UUID不能为空")
		return
	}

	// 获取操作信息以检查权限
	operation, err := dao.GetTeamTeaOperationByUuid(req.OperationUuid)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "操作不存在")
		return
	}

	// 检查用户是否可以管理团队账户（核心成员）
	canManage, err := dao.CanUserManageTeamAccount(user.Id, operation.TeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查管理权限失败")
		return
	}
	if !canManage {
		respondWithError(w, http.StatusForbidden, "只有核心成员才能审批团队茶叶操作")
		return
	}

	// 审批操作
	err = dao.ApproveTeamTeaOperation(req.OperationUuid, user.Id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队茶叶操作审批成功", nil)
}

// RejectTeamTeaOperation 拒绝团队茶叶操作
func RejectTeamTeaOperation(w http.ResponseWriter, r *http.Request) {
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
		OperationUuid string `json:"operation_uuid"`
		Reason        string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.OperationUuid == "" {
		respondWithError(w, http.StatusBadRequest, "操作UUID不能为空")
		return
	}

	// 获取操作信息以检查权限
	operation, err := dao.GetTeamTeaOperationByUuid(req.OperationUuid)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "操作不存在")
		return
	}

	// 检查用户是否可以管理团队账户（核心成员）
	canManage, err := dao.CanUserManageTeamAccount(user.Id, operation.TeamId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "检查管理权限失败")
		return
	}
	if !canManage {
		respondWithError(w, http.StatusForbidden, "只有核心成员才能审批团队茶叶操作")
		return
	}

	// 拒绝操作
	err = dao.RejectTeamTeaOperation(req.OperationUuid, user.Id, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithSuccess(w, "团队茶叶操作拒绝成功", nil)
}

// GetTeamPendingOperations 获取团队待审批操作列表
func GetTeamPendingOperations(w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看团队操作")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取待审批操作
	operations, err := dao.GetTeamPendingOperations(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取待审批操作失败")
		return
	}

	// 转换响应格式
	var responses []TeamTeaOperationResponse
	for _, operation := range operations {
		// 获取用户信息
		operatorUser, _ := dao.GetUser(operation.OperatorUserId)

		response := TeamTeaOperationResponse{
			Uuid:           operation.Uuid,
			TeamId:         operation.TeamId,
			OperationType:  operation.OperationType,
			AmountGrams:    operation.AmountGrams,
			Status:         operation.Status,
			OperatorUserId: operation.OperatorUserId,
			OperatorName:   operatorUser.Name,
			Notes:          operation.Notes,
			ExpiresAt:      operation.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:      operation.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 获取审批人信息
		if operation.ApproverUserId != nil {
			approverUser, _ := dao.GetUser(*operation.ApproverUserId)
			if approverUser.Id > 0 {
				response.ApproverUserId = operation.ApproverUserId
				response.ApproverName = approverUser.Name
			}
		}

		// 获取目标信息
		if operation.TargetTeamId != nil {
			targetTeam, _ := dao.GetTeam(*operation.TargetTeamId)
			if targetTeam.Id > 0 {
				response.TargetTeamId = operation.TargetTeamId
				response.TargetTeamName = targetTeam.Name
			}
		}
		if operation.TargetUserId != nil {
			targetUser, _ := dao.GetUser(*operation.TargetUserId)
			if targetUser.Id > 0 {
				response.TargetUserId = operation.TargetUserId
				response.TargetUserName = targetUser.Name
			}
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取待审批操作成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeamTeaOperations 获取团队操作历史
func GetTeamTeaOperations(w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看团队操作历史")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取操作历史
	operations, err := dao.GetTeamTeaOperations(teamId, page, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取操作历史失败")
		return
	}

	// 转换响应格式
	var responses []TeamTeaOperationResponse
	for _, operation := range operations {
		// 获取用户信息
		operatorUser, _ := dao.GetUser(operation.OperatorUserId)

		response := TeamTeaOperationResponse{
			Uuid:            operation.Uuid,
			TeamId:          operation.TeamId,
			OperationType:   operation.OperationType,
			AmountGrams:     operation.AmountGrams,
			Status:          operation.Status,
			OperatorUserId:  operation.OperatorUserId,
			OperatorName:    operatorUser.Name,
			Notes:           operation.Notes,
			RejectionReason: operation.RejectionReason,
			ExpiresAt:       operation.ExpiresAt.Format("2006-01-02 15:04:05"),
			CreatedAt:       operation.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 获取审批人信息
		if operation.ApproverUserId != nil {
			approverUser, _ := dao.GetUser(*operation.ApproverUserId)
			if approverUser.Id > 0 {
				response.ApproverUserId = operation.ApproverUserId
				response.ApproverName = approverUser.Name
			}
		}

		// 获取审批时间
		if operation.ApprovedAt != nil {
			approvedAt := operation.ApprovedAt.Format("2006-01-02 15:04:05")
			response.ApprovedAt = &approvedAt
		}

		// 获取目标信息
		if operation.TargetTeamId != nil {
			targetTeam, _ := dao.GetTeam(*operation.TargetTeamId)
			if targetTeam.Id > 0 {
				response.TargetTeamId = operation.TargetTeamId
				response.TargetTeamName = targetTeam.Name
			}
		}
		if operation.TargetUserId != nil {
			targetUser, _ := dao.GetUser(*operation.TargetUserId)
			if targetUser.Id > 0 {
				response.TargetUserId = operation.TargetUserId
				response.TargetUserName = targetUser.Name
			}
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取操作历史成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// GetTeamTeaTransactions 获取团队交易流水
func GetTeamTeaTransactions(w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusForbidden, "只有团队成员才能查看团队交易流水")
		return
	}

	// 获取分页参数
	page, limit := getPaginationParams(r)

	// 获取交易类型过滤参数
	transactionType := r.URL.Query().Get("type")

	// 获取交易流水
	transactions, err := dao.GetTeamTeaTransactions(teamId, page, limit, transactionType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "获取交易流水失败")
		return
	}

	// 转换响应格式
	var responses []TeamTeaTransactionResponse
	for _, tx := range transactions {
		response := TeamTeaTransactionResponse{
			Uuid:            tx.Uuid,
			TeamId:          tx.TeamId,
			OperationId:     tx.OperationId,
			TransactionType: tx.TransactionType,
			AmountGrams:     tx.AmountGrams,
			BalanceBefore:   tx.BalanceBefore,
			BalanceAfter:    tx.BalanceAfter,
			Description:     tx.Description,
			RelatedTeamId:   tx.RelatedTeamId,
			RelatedUserId:   tx.RelatedUserId,
			TargetType:      tx.TargetType,
			CreatedAt:       tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 设置流通对象类型显示文本
		switch tx.TargetType {
		case dao.TransactionTargetType_User:
			response.TargetTypeText = "个人"
		case dao.TransactionTargetType_Team:
			response.TargetTypeText = "团队"
		default:
			response.TargetTypeText = "未知"
		}

		// 获取相关团队信息
		if tx.RelatedTeamId != nil {
			relatedTeam, _ := dao.GetTeam(*tx.RelatedTeamId)
			if relatedTeam.Id > 0 {
				response.RelatedTeamId = tx.RelatedTeamId
				response.RelatedTeamName = relatedTeam.Name
			}
		}

		// 获取相关用户信息
		if tx.RelatedUserId != nil {
			relatedUser, _ := dao.GetUser(*tx.RelatedUserId)
			if relatedUser.Id > 0 {
				response.RelatedUserId = tx.RelatedUserId
				response.RelatedUserName = relatedUser.Name
			}
		}

		responses = append(responses, response)
	}

	respondWithPagination(w, "获取交易流水成功", responses, page, limit, 0) // TODO: 实现总数统计
}

// FreezeTeamTeaAccount 冻结团队茶叶账户（管理员功能）
func FreezeTeamTeaAccount(w http.ResponseWriter, r *http.Request) {
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
		TeamId int    `json:"team_id"`
		Reason string `json:"reason"`
	}
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

	// 获取账户并冻结
	account, err := dao.GetTeamTeaAccountByTeamId(req.TeamId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队账户不存在")
		return
	}

	err = account.UpdateStatus(dao.TeamTeaAccountStatus_Frozen, req.Reason)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "冻结账户失败")
		return
	}

	respondWithSuccess(w, "团队账户冻结成功", nil)
}

// UnfreezeTeamTeaAccount 解冻团队茶叶账户（管理员功能）
func UnfreezeTeamTeaAccount(w http.ResponseWriter, r *http.Request) {
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
		TeamId int `json:"team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.TeamId <= 0 {
		respondWithError(w, http.StatusBadRequest, "团队ID无效")
		return
	}

	// 获取账户并解冻
	account, err := dao.GetTeamTeaAccountByTeamId(req.TeamId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "团队账户不存在")
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

	// 获取待审批操作
	pendingOperationsData, _ := dao.GetTeamPendingOperations(team.Id, 1, 10)

	// 增强待审批操作数据，添加相关团队和用户信息
	type EnhancedPendingOperation struct {
		dao.TeamTeaOperation
		OperatorName   string
		TargetTeamName string
		TargetUserName string
		TypeDisplay    string
	}

	var pendingOperations []EnhancedPendingOperation
	for _, op := range pendingOperationsData {
		enhanced := EnhancedPendingOperation{
			TeamTeaOperation: op,
		}

		// 获取操作人信息
		operatorUser, _ := dao.GetUser(op.OperatorUserId)
		if operatorUser.Id > 0 {
			enhanced.OperatorName = operatorUser.Name
		}

		// 获取目标团队信息
		if op.TargetTeamId != nil {
			targetTeam, _ := dao.GetTeam(*op.TargetTeamId)
			if targetTeam.Id > 0 {
				enhanced.TargetTeamName = targetTeam.Name
			}
		}

		// 获取目标用户信息
		if op.TargetUserId != nil {
			targetUser, _ := dao.GetUser(*op.TargetUserId)
			if targetUser.Id > 0 {
				enhanced.TargetUserName = targetUser.Name
			}
		}

		// 添加操作类型显示
		switch op.OperationType {
		case dao.TeamOperationType_Deposit:
			enhanced.TypeDisplay = "存入"
		case dao.TeamOperationType_Withdraw:
			enhanced.TypeDisplay = "提取"
		case dao.TeamOperationType_TransferOut:
			enhanced.TypeDisplay = "转出"
		case dao.TeamOperationType_TransferIn:
			enhanced.TypeDisplay = "转入"
		default:
			enhanced.TypeDisplay = "未知"
		}

		pendingOperations = append(pendingOperations, enhanced)
	}

	// 获取最近交易
	recentTransactionsData, _ := dao.GetTeamTeaTransactions(team.Id, 0, 10, "")

	// 增强交易数据，添加相关团队和用户信息
	type EnhancedRecentTransaction struct {
		dao.TeamTeaTransaction
		RelatedTeamName string
		RelatedUserName string
		TargetTypeText  string
	}

	var recentTransactions []EnhancedRecentTransaction
	for _, tx := range recentTransactionsData {
		enhanced := EnhancedRecentTransaction{
			TeamTeaTransaction: tx,
		}

		// 设置流通对象类型显示文本
		switch tx.TargetType {
		case dao.TransactionTargetType_User:
			enhanced.TargetTypeText = "个人"
		case dao.TransactionTargetType_Team:
			enhanced.TargetTypeText = "团队"
		default:
			enhanced.TargetTypeText = "未知"
		}

		// 获取相关团队信息
		if tx.RelatedTeamId != nil {
			relatedTeam, _ := dao.GetTeam(*tx.RelatedTeamId)
			if relatedTeam.Id > 0 {
				enhanced.RelatedTeamName = relatedTeam.Name
			}
		}

		// 获取相关用户信息
		if tx.RelatedUserId != nil {
			relatedUser, _ := dao.GetUser(*tx.RelatedUserId)
			if relatedUser.Id > 0 {
				enhanced.RelatedUserName = relatedUser.Name
			}
		}

		recentTransactions = append(recentTransactions, enhanced)
	}

	// 检查用户权限
	isCoreMember, err := dao.CanUserManageTeamAccount(s_u.Id, team.Id)
	if err != nil {
		util.Debug("cannot check user permission", err)
		report(w, s_u, "获取团队管理信息失败。")
		return
	}

	// 创建团队账户数据
	teamAccountData := []map[string]interface{}{
		{
			"TeamId":             team.Id,
			"TeamUuid":           team.Uuid,
			"TeamName":           team.Name,
			"TeamAbbrev":         team.Abbreviation,
			"TeamAccount":        teamAccount,
			"UserIsCoreMember":   isCoreMember,
			"PendingOperations":  pendingOperations,
			"RecentTransactions": recentTransactions,
			"BalanceDisplay":     formatTeaBalance(teamAccount.BalanceGrams),
			"StatusDisplay":      getTeamAccountStatusDisplay(teamAccount.Status, teamAccount.FrozenReason),
		},
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser        dao.User
		Team            *dao.Team
		TeamAccountData []map[string]interface{}
	}

	pageData.SessUser = s_u
	pageData.Team = singleTeam
	pageData.TeamAccountData = teamAccountData

	generateHTML(w, &pageData, "layout", "navbar.private", "team_tea_account")
}

// formatTeaBalance 格式化茶叶余额显示
func formatTeaBalance(balance float64) string {
	if balance >= 1 {
		return util.FormatFloat(balance, 2) + " 克"
	} else {
		return util.FormatFloat(balance*1000, 0) + " 毫克"
	}
}

// HandleTeamTeaTransactions 处理团队交易流水页面请求
func HandleTeamTeaTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TeamTeaTransactionsGet(w, r)
}

// HandleTeamPendingOperations 处理团队待审批操作页面请求
func HandleTeamPendingOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TeamPendingOperationsGet(w, r)
}

// HandleTeamTeaOperations 处理团队操作记录页面请求
func HandleTeamTeaOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TeamOperationsHistoryGet(w, r)
}

// TeamTeaTransactionsGet 获取团队交易流水页面
func TeamTeaTransactionsGet(w http.ResponseWriter, r *http.Request) {
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

	// 显示指定团队的交易流水
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
		report(w, s_u, "您不是该团队成员，无法查看交易流水。")
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

	// 获取交易类型过滤参数
	transactionType := r.URL.Query().Get("type")

	// 获取交易流水
	transactions, err := dao.GetTeamTeaTransactions(team.Id, page, limit, transactionType)
	if err != nil {
		util.Debug("cannot get team transactions", err)
		report(w, s_u, "获取交易流水失败。")
		return
	}

	// 检查用户权限
	isCoreMember, err := dao.CanUserManageTeamAccount(s_u.Id, team.Id)
	if err != nil {
		util.Debug("cannot check user permission", err)
		report(w, s_u, "获取团队管理信息失败。")
		return
	}

	// 增强交易数据，添加相关团队和用户信息
	type EnhancedTransaction struct {
		dao.TeamTeaTransaction
		RelatedTeamName string
		RelatedUserName string
		TargetTypeText  string
	}

	var enhancedTransactions []EnhancedTransaction
	for _, tx := range transactions {
		enhanced := EnhancedTransaction{
			TeamTeaTransaction: tx,
		}

		// 设置流通对象类型显示文本
		switch tx.TargetType {
		case dao.TransactionTargetType_User:
			enhanced.TargetTypeText = "个人"
		case dao.TransactionTargetType_Team:
			enhanced.TargetTypeText = "团队"
		default:
			enhanced.TargetTypeText = "未知"
		}

		// 获取相关团队信息
		if tx.RelatedTeamId != nil {
			relatedTeam, _ := dao.GetTeam(*tx.RelatedTeamId)
			if relatedTeam.Id > 0 {
				enhanced.RelatedTeamName = relatedTeam.Name
			}
		}

		// 获取相关用户信息
		if tx.RelatedUserId != nil {
			relatedUser, _ := dao.GetUser(*tx.RelatedUserId)
			if relatedUser.Id > 0 {
				enhanced.RelatedUserName = relatedUser.Name
			}
		}

		enhancedTransactions = append(enhancedTransactions, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser         dao.User
		Team             *dao.Team
		TeamAccount      dao.TeamTeaAccount
		Transactions     []EnhancedTransaction
		UserIsCoreMember bool
		BalanceDisplay   string
		StatusDisplay    string
		CurrentPage      int
		Limit            int
		FilterType       string
	}

	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Transactions = enhancedTransactions
	pageData.UserIsCoreMember = isCoreMember
	pageData.BalanceDisplay = formatTeaBalance(teamAccount.BalanceGrams)
	pageData.StatusDisplay = getTeamAccountStatusDisplay(teamAccount.Status, teamAccount.FrozenReason)
	pageData.CurrentPage = page
	pageData.Limit = limit
	pageData.FilterType = transactionType

	generateHTML(w, &pageData, "layout", "navbar.private", "team_tea_transactions")
}

// TeamPendingOperationsGet 获取团队待审批操作页面
func TeamPendingOperationsGet(w http.ResponseWriter, r *http.Request) {
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

	// 显示指定团队的待审批操作
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
		report(w, s_u, "您不是该团队成员，无法查看待审批操作。")
		return
	}

	// 检查用户权限（只有核心成员才能查看待审批操作）
	isCoreMember, err := dao.CanUserManageTeamAccount(s_u.Id, team.Id)
	if err != nil {
		util.Debug("cannot check user permission", err)
		report(w, s_u, "获取团队管理信息失败。")
		return
	}
	if !isCoreMember {
		report(w, s_u, "只有核心成员才能查看待审批操作。")
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

	// 获取待审批操作
	operations, err := dao.GetTeamPendingOperations(team.Id, page, limit)
	if err != nil {
		util.Debug("cannot get pending operations", err)
		report(w, s_u, "获取待审批操作失败。")
		return
	}

	// 增强操作数据，添加相关团队和用户信息
	type EnhancedOperation struct {
		dao.TeamTeaOperation
		OperatorName   string
		ApproverName   string
		TargetTeamName string
		TargetUserName string
	}

	var enhancedOperations []EnhancedOperation
	for _, op := range operations {
		enhanced := EnhancedOperation{
			TeamTeaOperation: op,
		}

		// 获取操作人信息
		operatorUser, _ := dao.GetUser(op.OperatorUserId)
		if operatorUser.Id > 0 {
			enhanced.OperatorName = operatorUser.Name
		}

		// 获取审批人信息
		if op.ApproverUserId != nil {
			approverUser, _ := dao.GetUser(*op.ApproverUserId)
			if approverUser.Id > 0 {
				enhanced.ApproverName = approverUser.Name
			}
		}

		// 获取目标团队信息
		if op.TargetTeamId != nil {
			targetTeam, _ := dao.GetTeam(*op.TargetTeamId)
			if targetTeam.Id > 0 {
				enhanced.TargetTeamName = targetTeam.Name
			}
		}

		// 获取目标用户信息
		if op.TargetUserId != nil {
			targetUser, _ := dao.GetUser(*op.TargetUserId)
			if targetUser.Id > 0 {
				enhanced.TargetUserName = targetUser.Name
			}
		}

		enhancedOperations = append(enhancedOperations, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser         dao.User
		Team             *dao.Team
		TeamAccount      dao.TeamTeaAccount
		Operations       []EnhancedOperation
		UserIsCoreMember bool
		BalanceDisplay   string
		StatusDisplay    string
		CurrentPage      int
		Limit            int
	}

	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Operations = enhancedOperations
	pageData.UserIsCoreMember = isCoreMember
	pageData.BalanceDisplay = formatTeaBalance(teamAccount.BalanceGrams)
	pageData.StatusDisplay = getTeamAccountStatusDisplay(teamAccount.Status, teamAccount.FrozenReason)
	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "team_tea_pending_operations")
}

// TeamOperationsHistoryGet 获取团队操作记录页面
func TeamOperationsHistoryGet(w http.ResponseWriter, r *http.Request) {
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

	// 显示指定团队的操作记录
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
		report(w, s_u, "您不是该团队成员，无法查看操作记录。")
		return
	}

	// 检查用户权限（只有核心成员才能查看操作记录）
	isCoreMember, err := dao.CanUserManageTeamAccount(s_u.Id, team.Id)
	if err != nil {
		util.Debug("cannot check user permission", err)
		report(w, s_u, "获取团队管理信息失败。")
		return
	}
	if !isCoreMember {
		report(w, s_u, "只有核心成员才能查看操作记录。")
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

	// 获取操作记录（包括所有状态）
	operations, err := dao.GetTeamTeaOperations(team.Id, page, limit)
	if err != nil {
		util.Debug("cannot get team operations", err)
		report(w, s_u, "获取操作记录失败。")
		return
	}

	// 增强操作数据，添加相关团队和用户信息
	type EnhancedOperation struct {
		dao.TeamTeaOperation
		OperatorName   string
		ApproverName   string
		TargetTeamName string
		TargetUserName string
		StatusDisplay  string
		TypeDisplay    string
		IsExpired      bool
	}

	var enhancedOperations []EnhancedOperation
	for _, op := range operations {
		enhanced := EnhancedOperation{
			TeamTeaOperation: op,
		}

		// 获取操作人信息
		operatorUser, _ := dao.GetUser(op.OperatorUserId)
		if operatorUser.Id > 0 {
			enhanced.OperatorName = operatorUser.Name
		}

		// 获取审批人信息
		if op.ApproverUserId != nil {
			approverUser, _ := dao.GetUser(*op.ApproverUserId)
			if approverUser.Id > 0 {
				enhanced.ApproverName = approverUser.Name
			}
		}

		// 获取目标团队信息
		if op.TargetTeamId != nil {
			targetTeam, _ := dao.GetTeam(*op.TargetTeamId)
			if targetTeam.Id > 0 {
				enhanced.TargetTeamName = targetTeam.Name
			}
		}

		// 获取目标用户信息
		if op.TargetUserId != nil {
			targetUser, _ := dao.GetUser(*op.TargetUserId)
			if targetUser.Id > 0 {
				enhanced.TargetUserName = targetUser.Name
			}
		}

		// 添加状态显示
		switch op.Status {
		case dao.TeamOperationStatus_Pending:
			enhanced.StatusDisplay = "待审批"
		case dao.TeamOperationStatus_Approved:
			enhanced.StatusDisplay = "已批准"
		case dao.TeamOperationStatus_Rejected:
			enhanced.StatusDisplay = "已拒绝"
		case dao.TeamOperationStatus_Expired:
			enhanced.StatusDisplay = "已过期"
		default:
			enhanced.StatusDisplay = "未知"
		}

		// 添加操作类型显示
		switch op.OperationType {
		case dao.TeamOperationType_Deposit:
			enhanced.TypeDisplay = "存入"
		case dao.TeamOperationType_Withdraw:
			enhanced.TypeDisplay = "提取"
		case dao.TeamOperationType_TransferOut:
			enhanced.TypeDisplay = "转出"
		case dao.TeamOperationType_TransferIn:
			enhanced.TypeDisplay = "转入"
		default:
			enhanced.TypeDisplay = "未知"
		}

		// 检查是否过期
		enhanced.IsExpired = op.ExpiresAt.Before(time.Now())

		enhancedOperations = append(enhancedOperations, enhanced)
	}

	// 创建页面数据结构
	var pageData struct {
		SessUser         dao.User
		Team             *dao.Team
		TeamAccount      dao.TeamTeaAccount
		Operations       []EnhancedOperation
		UserIsCoreMember bool
		BalanceDisplay   string
		StatusDisplay    string
		CurrentPage      int
		Limit            int
	}

	pageData.SessUser = s_u
	pageData.Team = &team
	pageData.TeamAccount = teamAccount
	pageData.Operations = enhancedOperations
	pageData.UserIsCoreMember = isCoreMember
	pageData.BalanceDisplay = formatTeaBalance(teamAccount.BalanceGrams)
	pageData.StatusDisplay = getTeamAccountStatusDisplay(teamAccount.Status, teamAccount.FrozenReason)
	pageData.CurrentPage = page
	pageData.Limit = limit

	generateHTML(w, &pageData, "layout", "navbar.private", "team_tea_operations_history")
}

// getTeamAccountStatusDisplay 获取团队账户状态显示
func getTeamAccountStatusDisplay(status string, frozenReason *string) string {
	switch status {
	case dao.TeamTeaAccountStatus_Frozen:
		if frozenReason != nil {
			return "已冻结 (" + *frozenReason + ")"
		}
		return "已冻结"
	default:
		return "正常"
	}
}
