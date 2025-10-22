package route

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
	"time"
)

// saveUploadedFile 保存上传的可视凭证文件
func saveUploadedFile(file multipart.File, header *multipart.FileHeader, userId int) (string, int64, error) {
	// 创建上传目录
	uploadDir := filepath.Join("public", "uploads", "evidence", time.Now().Format("2006-01"))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", 0, err
	}

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	hash := md5.New()
	io.WriteString(hash, fmt.Sprintf("%d_%d_%s", userId, time.Now().UnixNano(), header.Filename))
	fileName := fmt.Sprintf("%x%s", hash.Sum(nil), ext)
	filePath := filepath.Join(uploadDir, fileName)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", 0, err
	}
	defer dst.Close()

	fileSize, err := io.Copy(dst, file)
	if err != nil {
		return "", 0, err
	}

	return filePath, fileSize, nil
}

// Handler /v1/evidence/new
func HandleNewEvidence(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		EvidenceNewGet(w, r)
	case http.MethodPost:
		EvidenceNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/evidence/new
func EvidenceNewGet(w http.ResponseWriter, r *http.Request) {
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

	templateData := struct {
		SessUser data.User
	}{
		SessUser: s_u,
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "evidence.new")
}

// POST /v1/evidence/new
func EvidenceNewPost(w http.ResponseWriter, r *http.Request) {
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

	if !isVerifier(s_u.Id) {
		report(w, r, "你没有权限创建证据记录")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		util.Debug("Cannot parse multipart form", err)
		report(w, r, "表单数据解析失败")
		return
	}

	description := r.FormValue("description")
	note := r.FormValue("note")
	categoryStr := r.FormValue("category")
	originalURL := r.FormValue("original_url")
	visibilityStr := r.FormValue("visibility")

	if description == "" {
		report(w, r, "请填写可视凭据描述")
		return
	}

	category, _ := strconv.Atoi(categoryStr)
	visibility, _ := strconv.Atoi(visibilityStr)

	evidence := data.Evidence{
		Description:    description,
		RecorderUserId: s_u.Id,
		Note:           note,
		Category:       data.EvidenceCategory(category),
		OriginalURL:    originalURL,
		Visibility:     visibility,
	}

	// 处理文件上传
	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()

		// 保存文件
		filePath, fileSize, err := saveUploadedFile(file, header, s_u.Id)
		if err != nil {
			util.Debug("Cannot save uploaded file", err)
			report(w, r, "文件上传失败")
			return
		}

		evidence.Path = filePath
		evidence.FileName = header.Filename
		evidence.MimeType = header.Header.Get("Content-Type")
		evidence.FileSize = fileSize
	} else if originalURL == "" {
		report(w, r, "请上传文件或填写原始URL")
		return
	}

	if err := evidence.Create(r.Context()); err != nil {
		util.Debug("Cannot create evidence", err)
		report(w, r, "创建证据记录失败")
		return
	}

	// 返回 JSON 响应，方便前端获取新创建的证据ID
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true, "id": ` + strconv.Itoa(evidence.Id) + `, "message": "证据创建成功"}`))
}
