package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/magic/new
func HandleNewMagic(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		MagicNewGet(w, r)
	case http.MethodPost:
		MagicNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handler /v1/magic/detail
func HandleMagicDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	MagicDetailGet(w, r)
}

// Handler /v1/magic/list
func HandleMagicList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	MagicListGet(w, r)
}

// GET /v1/magic/new
func MagicNewGet(w http.ResponseWriter, r *http.Request) {
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

	var magicData struct {
		SessUser  data.User
		ReturnURL string
	}
	magicData.SessUser = user
	magicData.ReturnURL = r.URL.Query().Get("return_url")

	generateHTML(w, &magicData, "layout", "navbar.private", "magic.new")
}

// POST /v1/magic/new
func MagicNewPost(w http.ResponseWriter, r *http.Request) {
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
		report(w, r, "法力名称不能为空。")
		return
	}
	if description == "" {
		report(w, r, "法力描述不能为空。")
		return
	}

	// 解析表单数据
	category, _ := strconv.Atoi(r.PostFormValue("category"))
	if category < 1 || category > 2 {
		category = 1 // 默认理性
	}

	intelligenceLevel, _ := strconv.Atoi(r.PostFormValue("intelligence_level"))
	if intelligenceLevel < 1 || intelligenceLevel > 5 {
		intelligenceLevel = 3 // 默认中等
	}

	difficultyLevel, _ := strconv.Atoi(r.PostFormValue("difficulty_level"))
	if difficultyLevel < 1 || difficultyLevel > 5 {
		difficultyLevel = 3 // 默认中等
	}

	level, _ := strconv.Atoi(r.PostFormValue("level"))
	if level < 1 || level > 5 {
		level = 1 // 默认入门
	}

	magic := data.Magic{
		UserId:            user.Id,
		Name:              name,
		Nickname:          strings.TrimSpace(r.PostFormValue("nickname")),
		Description:       description,
		IntelligenceLevel: data.IntelligenceLevel(intelligenceLevel),
		DifficultyLevel:   data.DifficultyLevel(difficultyLevel),
		Category:          data.MagicCategory(category),
		Level:             level,
	}

	if err := magic.Create(r.Context()); err != nil {
		util.Debug("Cannot create magic", err)
		report(w, r, "创建法力记录失败，请重试。")
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
		// 如果有返回URL，添加新创建的法力ID参数
		if returnURL != "/v1/" {
			if strings.Contains(returnURL, "?") {
				returnURL += fmt.Sprintf("&new_magic_id=%d", magic.Id)
			} else {
				returnURL += fmt.Sprintf("?new_magic_id=%d", magic.Id)
			}
		}
	}
	http.Redirect(w, r, returnURL, http.StatusFound)
}

// GET /v1/magic/detail?id=123
func MagicDetailGet(w http.ResponseWriter, r *http.Request) {
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
	uuidStr := r.URL.Query().Get("uuid")
	if idStr == "" && uuidStr == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	var magic data.Magic
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			report(w, r, "你好，假作真时真亦假，无为有处有还无？")
			return
		}
		magic.Id = id
	} else {
		magic.Uuid = uuidStr
	}

	if err := magic.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get magic by id/uuid", magic.Id, magic.Uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 创建MagicBean结构
	magicBean := struct {
		Magic data.Magic
	}{
		Magic: magic,
	}

	var magicData struct {
		SessUser   data.User
		IsVerifier bool
		IsAdmin    bool
		IsMaster   bool
		IsInvited  bool
		MagicBean  interface{}
	}
	magicData.SessUser = user
	magicData.MagicBean = magicBean

	generateHTML(w, &magicData, "layout", "navbar.private", "magic.detail")
}

// GET /v1/magic/list
func MagicListGet(w http.ResponseWriter, r *http.Request) {
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

	// 获取所有法力
	magics, err := data.GetAllMagics(r.Context())
	if err != nil {
		util.Debug("Cannot get magics", err)
		report(w, r, "获取法力列表失败，请重试。")
		return
	}

	// 创建法力Bean列表
	var magicBeans []struct {
		Magic data.Magic
	}
	for _, magic := range magics {
		magicBeans = append(magicBeans, struct {
			Magic data.Magic
		}{
			Magic: magic,
		})
	}

	var magicData struct {
		SessUser data.User
		Magics   []struct {
			Magic data.Magic
		}
	}
	magicData.SessUser = user
	magicData.Magics = magicBeans

	generateHTML(w, &magicData, "layout", "navbar.private", "magic.list")
}