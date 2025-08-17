package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/risk/new
func HandleNewRisk(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		RiskNewGet(w, r)
	case http.MethodPost:
		RiskNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/risk/new
func RiskNewGet(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	user, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	var riskData struct {
		SessUser  data.User
		ReturnURL string
	}
	riskData.SessUser = user
	riskData.ReturnURL = r.URL.Query().Get("return_url")

	renderHTML(w, &riskData, "layout", "navbar.private", "risk.new")
}

// POST /v1/risk/new
func RiskNewPost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	user, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 验证必填字段
	name := strings.TrimSpace(r.PostFormValue("name"))
	description := strings.TrimSpace(r.PostFormValue("description"))
	
	if name == "" {
		report(w, r, "风险名称不能为空。")
		return
	}
	if description == "" {
		report(w, r, "风险描述不能为空。")
		return
	}

	// 解析表单数据
	severity, _ := strconv.Atoi(r.PostFormValue("severity"))
	if severity < 1 || severity > 5 {
		severity = 3 // 默认中风险
	}

	risk := data.Risk{
		UserId:      user.Id,
		Name:        name,
		Nickname:    strings.TrimSpace(r.PostFormValue("nickname")),
		Keywords:    strings.TrimSpace(r.PostFormValue("keywords")),
		Description: description,
		Source:      strings.TrimSpace(r.PostFormValue("source")),
		Severity:    data.RiskSeverityLevel(severity),
	}

	if err := risk.Create(); err != nil {
		util.Debug("Cannot create risk", err)
		report(w, r, "创建风险记录失败，请重试。")
		return
	}

	// 获取返回URL参数
	returnURL := r.PostFormValue("return_url")
	if returnURL == "" {
		returnURL = r.URL.Query().Get("return_url")
	}
	if returnURL == "" {
		returnURL = "/v1/"
	} else {
		// 如果有返回URL，添加新创建的风险ID参数
		if returnURL != "/v1/" {
			if strings.Contains(returnURL, "?") {
				returnURL += fmt.Sprintf("&new_risk_id=%d", risk.Id)
			} else {
				returnURL += fmt.Sprintf("?new_risk_id=%d", risk.Id)
			}
		}
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}