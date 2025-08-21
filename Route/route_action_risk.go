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

// Handler /v1/risk/detail
func HandleRiskDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	RiskDetailGet(w, r)
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
		Severity:    severity,
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

// GET /v1/risk/detail?id=123
func RiskDetailGet(w http.ResponseWriter, r *http.Request) {
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

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	risk := data.Risk{Id: id}
	if err := risk.GetByIdOrUUID(); err != nil {
		util.Debug("Cannot get risk by id", id, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 获取记录者信息
	recorder, err := data.GetUser(risk.UserId)
	if err != nil {
		util.Debug("Cannot get recorder user", risk.UserId, err)
		// 如果获取记录者失败，使用默认值
		recorder = data.User{Id: 0, Name: "未知用户"}
	}

	var riskData struct {
		SessUser data.User
		Risk     data.Risk
		Recorder data.User
	}
	riskData.SessUser = user
	riskData.Risk = risk
	riskData.Recorder = recorder

	renderHTML(w, &riskData, "layout", "navbar.private", "risk.detail")
}
