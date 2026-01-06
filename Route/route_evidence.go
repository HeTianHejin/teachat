package route

import (
	"net/http"
	"strconv"
	dao "teachat/DAO"
	util "teachat/Util"
)

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
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	templateData := struct {
		SessUser dao.User
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
		report(w, s_u, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	if !dao.IsVerifier(s_u.Id) {
		report(w, s_u, "你没有权限创建证据记录")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		util.Debug("Cannot parse multipart form", err)
		report(w, s_u, "表单数据解析失败")
		return
	}

	description := r.FormValue("description")
	note := r.FormValue("note")
	categoryStr := r.FormValue("category")
	originalURL := r.FormValue("original_url")
	visibilityStr := r.FormValue("visibility")

	if description == "" {
		report(w, s_u, "请填写可视凭据描述")
		return
	}

	category, _ := strconv.Atoi(categoryStr)
	visibility, _ := strconv.Atoi(visibilityStr)

	evidence := dao.Evidence{
		Description:    description,
		RecorderUserId: s_u.Id,
		Note:           note,
		Category:       dao.EvidenceCategory(category),
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
			report(w, s_u, "文件上传失败")
			return
		}

		evidence.Path = filePath
		evidence.FileName = header.Filename
		evidence.MimeType = header.Header.Get("Content-Type")
		evidence.FileSize = fileSize
	} else if originalURL == "" {
		report(w, s_u, "请上传文件或填写原始URL")
		return
	}

	if err := evidence.Create(r.Context()); err != nil {
		util.Debug("Cannot create evidence", err)
		report(w, s_u, "创建证据记录失败")
		return
	}

	// 返回 JSON 响应，方便前端获取新创建的证据ID
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true, "id": ` + strconv.Itoa(evidence.Id) + `, "message": "证据创建成功"}`))
}

// Handler /v1/evidence/detail
func HandleEvidenceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	EvidenceDetailGet(w, r)
}

// GET /v1/evidence/detail?id=xxx
func EvidenceDetailGet(w http.ResponseWriter, r *http.Request) {
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

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		report(w, s_u, "凭证ID缺失")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		report(w, s_u, "无效的凭证ID")
		return
	}

	evidence := dao.Evidence{Id: id}
	if err := evidence.GetByIdOrUUID(r.Context()); err != nil {
		util.Debug("Cannot get evidence by id", id, err)
		report(w, s_u, "凭证不存在")
		return
	}

	// 检查权限：如果是私密凭证，需要验证权限
	if evidence.Visibility == dao.VisibilityPrivate {
		// 获取凭证关联的手工艺
		var handicraft dao.Handicraft
		if inaugurations, err := dao.GetInaugurationsByEvidenceId(evidence.Id); err == nil && len(inaugurations) > 0 {
			handicraft.Id = inaugurations[0].HandicraftId
		} else if processRecords, err := dao.GetProcessRecordsByEvidenceId(evidence.Id); err == nil && len(processRecords) > 0 {
			handicraft.Id = processRecords[0].HandicraftId
		} else if endings, err := dao.GetEndingsByEvidenceId(evidence.Id); err == nil && len(endings) > 0 {
			handicraft.Id = endings[0].HandicraftId
		}

		if handicraft.Id > 0 {
			if err := handicraft.GetByIdOrUUID(r.Context()); err == nil {
				project := dao.Project{Id: handicraft.ProjectId}
				if err := project.Get(); err == nil {
					objective, err := project.Objective()
					if err == nil {
						is_master, _ := checkProjectMasterPermission(&project, s_u.Id)
						is_admin, _ := checkObjectiveAdminPermission(&objective, s_u.Id)
						is_invited, _ := objective.IsInvitedMember(s_u.Id)
						if !is_master && !is_admin && !is_invited {
							report(w, s_u, "你好，身后有余忘缩手，眼前无路想回头。")
							return
						}
					}
				}
			}
		}
	}

	templateData := struct {
		SessUser   dao.User
		Evidence   dao.Evidence
		IsVerifier bool
	}{
		SessUser:   s_u,
		Evidence:   evidence,
		IsVerifier: dao.IsVerifier(s_u.Id),
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "evidence.detail")
}
