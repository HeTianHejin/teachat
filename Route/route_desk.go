package route

import (
	"net/http"
	dao "teachat/DAO"
	util "teachat/Util"
)

// HandleDesk 处理写字台页面请求
func HandleDesk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	DeskGet(w, r)
}

// DeskGet 获取写字台页面
func DeskGet(w http.ResponseWriter, r *http.Request) {
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
		// 不阻止流程，即使账户创建失败也显示页面
	}

	// 获取用户茶叶账户信息
	account, err := dao.GetTeaAccountByUserId(s_u.Id)
	var accountInfo *dao.TeaUserAccount
	if err == nil {
		accountInfo = &account
	} else {
		// 如果获取失败，创建一个空的账户信息
		accountInfo = &dao.TeaUserAccount{
			UserId:       s_u.Id,
			BalanceGrams: 0.0,
			Status:       dao.TeaAccountStatus_Normal,
		}
	}

	// 获取待确认转账数量
	pendingTransfers, err := dao.GetPendingTransfers(s_u.Id, 1, 1) // 只需要获取数量
	var pendingCount int
	if err == nil {
		pendingCount = len(pendingTransfers)
	}

	// 创建写字台数据结构
	var deskData struct {
		SessUser     dao.User
		TeaAccount   *dao.TeaUserAccount
		AccountInfo  struct {
			BalanceDisplay         string
			LockedBalanceDisplay   string
			AvailableBalanceDisplay string
			StatusDisplay          string
			IsFrozen               bool
		}
		PendingTransferCount int
	}

	deskData.SessUser = s_u
	deskData.TeaAccount = accountInfo
	deskData.PendingTransferCount = pendingCount

	// 格式化总余额显示
	if accountInfo.BalanceGrams >= 1 {
		deskData.AccountInfo.BalanceDisplay = util.FormatFloat(accountInfo.BalanceGrams, 2) + " 克"
	} else {
		deskData.AccountInfo.BalanceDisplay = util.FormatFloat(accountInfo.BalanceGrams*1000, 0) + " 毫克"
	}

	// 格式化锁定余额显示
	if accountInfo.LockedBalanceGrams >= 1 {
		deskData.AccountInfo.LockedBalanceDisplay = util.FormatFloat(accountInfo.LockedBalanceGrams, 2) + " 克"
	} else {
		deskData.AccountInfo.LockedBalanceDisplay = util.FormatFloat(accountInfo.LockedBalanceGrams*1000, 0) + " 毫克"
	}

	// 计算和格式化可用余额显示
	availableBalance := accountInfo.BalanceGrams - accountInfo.LockedBalanceGrams
	if availableBalance >= 1 {
		deskData.AccountInfo.AvailableBalanceDisplay = util.FormatFloat(availableBalance, 2) + " 克"
	} else {
		deskData.AccountInfo.AvailableBalanceDisplay = util.FormatFloat(availableBalance*1000, 0) + " 毫克"
	}

	// 状态显示
	if accountInfo.Status == dao.TeaAccountStatus_Frozen {
		deskData.AccountInfo.StatusDisplay = "已冻结"
		deskData.AccountInfo.IsFrozen = true
	} else {
		deskData.AccountInfo.StatusDisplay = "正常"
		deskData.AccountInfo.IsFrozen = false
	}

	generateHTML(w, &deskData, "layout", "navbar.private", "desk")
}
