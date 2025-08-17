package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/hazard/new
func HandleNewHazard(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HazardNewGet(w, r)
	case http.MethodPost:
		HazardNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/hazard/new
func HazardNewGet(w http.ResponseWriter, r *http.Request) {
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

	var hazardData struct {
		SessUser  data.User
		ReturnURL string
	}
	hazardData.SessUser = user
	hazardData.ReturnURL = r.URL.Query().Get("return_url")

	renderHTML(w, &hazardData, "layout", "navbar.private", "hazard.new")
}

// POST /v1/hazard/new
func HazardNewPost(w http.ResponseWriter, r *http.Request) {
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
		report(w, r, "隐患名称不能为空。")
		return
	}
	if description == "" {
		report(w, r, "隐患描述不能为空。")
		return
	}

	// 解析表单数据
	severity, _ := strconv.Atoi(r.PostFormValue("severity"))
	if severity < 1 || severity > 5 {
		severity = 3 // 默认中风险
	}

	category, _ := strconv.Atoi(r.PostFormValue("category"))
	if category < 1 || category > 6 {
		category = 6 // 默认其他
	}

	hazard := data.Hazard{
		UserId:      user.Id,
		Name:        name,
		Nickname:    strings.TrimSpace(r.PostFormValue("nickname")),
		Keywords:    strings.TrimSpace(r.PostFormValue("keywords")),
		Description: description,
		Source:      strings.TrimSpace(r.PostFormValue("source")),
		Severity:    severity,
		Category:    data.HazardCategory(category),
	}

	if err := hazard.Create(); err != nil {
		util.Debug("Cannot create hazard", err)
		report(w, r, "创建隐患记录失败，请重试。")
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
		// 如果有返回URL，添加新创建的隐患ID参数
		if returnURL != "/v1/" {
			if strings.Contains(returnURL, "?") {
				returnURL += fmt.Sprintf("&new_hazard_id=%d", hazard.Id)
			} else {
				returnURL += fmt.Sprintf("?new_hazard_id=%d", hazard.Id)
			}
		}
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}