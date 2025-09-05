package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/environment/new
func HandleNewEnvironment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		EnvironmentNewGet(w, r)
	case http.MethodPost:
		EnvironmentNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handler /v1/environment/detail
func HandleEnvironmentDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	EnvironmentDetailGet(w, r)
}

// GET /v1/environment/new
func EnvironmentNewGet(w http.ResponseWriter, r *http.Request) {
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

	var envData struct {
		SessUser  data.User
		ReturnURL string
	}
	envData.SessUser = user
	envData.ReturnURL = r.URL.Query().Get("return_url")

	renderHTML(w, &envData, "layout", "navbar.private", "project.environment.new")
}

// POST /v1/environment/new
func EnvironmentNewPost(w http.ResponseWriter, r *http.Request) {
	sess, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	// 验证当前用户身份是否见证者茶团成员
	if !(isVerifier(s_u.Id)) {
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	env := data.Environment{
		Name:        r.PostFormValue("name"),
		Summary:     r.PostFormValue("summary"),
		UserId:      s_u.Id, // 记录见证者ID
		Temperature: parseInt(r.PostFormValue("temperature")),
		Humidity:    parseInt(r.PostFormValue("humidity")),
		PM25:        parseInt(r.PostFormValue("pm25")),
		Noise:       parseInt(r.PostFormValue("noise")),
		Light:       parseInt(r.PostFormValue("light")),
		Wind:        parseInt(r.PostFormValue("wind")),
		Flow:        parseInt(r.PostFormValue("flow")),
		Rain:        parseInt(r.PostFormValue("rain")),
		Pressure:    parseInt(r.PostFormValue("pressure")),
		Smoke:       parseInt(r.PostFormValue("smoke")),
		Dust:        parseInt(r.PostFormValue("dust")),
		Odor:        parseInt(r.PostFormValue("odor")),
		Visibility:  parseInt(r.PostFormValue("visibility")),
	}

	if err := env.Create(); err != nil {
		util.Debug("Cannot create environment", err)
		report(w, r, "创建环境条件失败，请重试。")
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
		// 如果有返回URL，添加新创建的环境ID参数
		if returnURL != "/v1/" {
			if strings.Contains(returnURL, "?") {
				returnURL += fmt.Sprintf("&new_env_id=%d", env.Id)
			} else {
				returnURL += fmt.Sprintf("?new_env_id=%d", env.Id)
			}
		}
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}

// GET /v1/environment/detail?id=123
func EnvironmentDetailGet(w http.ResponseWriter, r *http.Request) {
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

	env := data.Environment{Id: id}
	if err := env.GetByIdOrUUID(); err != nil {
		util.Debug("Cannot get environment by id", id, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 获取记录者信息
	recorder, err := data.GetUser(env.UserId)
	if err != nil {
		util.Debug("Cannot get recorder user", env.UserId, err)
		// 如果获取记录者失败，使用默认值
		recorder = data.User{Id: 0, Name: "未知用户"}
	}

	var envData struct {
		SessUser    data.User
		Environment data.Environment
		Recorder    data.User
	}
	envData.SessUser = user
	envData.Environment = env
	envData.Recorder = recorder

	renderHTML(w, &envData, "layout", "navbar.private", "project.environment.detail")
}

func parseInt(s string) int {
	if s == "" {
		return 3 // 默认值
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 3 // 默认值
	}
	if i < 1 || i > 5 {
		return 3 // 默认值
	}
	return i
}
