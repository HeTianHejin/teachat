package route

import (
	"net/http"
	"strconv"

	data "teachat/DAO"
	util "teachat/Util"
)

// HandleGoodsProjectNew 处理项目物资新增页面和创建
func HandleGoodsProjectNew(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	if !isVerifier(s_u.Id) {
		report(w, r, "只有见证员才可以添加项目物资，请联系管理员。")
		return
	}

	switch r.Method {
	case http.MethodGet:
		GoodsProjectNewGet(w, r, s_u)
	case http.MethodPost:
		GoodsProjectNewPost(w, r, s_u)
	default:
		report(w, r, "一脸蒙的茶博士，表示看不懂你的项目物资资料，请确认后再试一次。")
		return
	}
}

// GET /v1/goods/project_new?project_id=xxx
func GoodsProjectNewGet(w http.ResponseWriter, r *http.Request, s_u data.User) {

	project_id_str := r.URL.Query().Get("project_id")
	if project_id_str == "" {
		report(w, r, "一脸蒙的茶博士，表示看不懂你的项目资料，请确认后再试一次。")
		return
	}
	project_id, err := strconv.Atoi(project_id_str)
	if err != nil {
		report(w, r, "一脸蒙的茶博士，表示看不懂你的项目资料，请确认后再试一次。")
		return
	}

	// 验证项目是否存在
	project := data.Project{Id: project_id}
	if err := project.Get(); err != nil {
		util.Debug("cannot get project from database", err)
		report(w, r, "一脸蒙的茶博士，表示看不懂你的项目资料，请确认后再试一次。")
		return
	}
	// 读取“约茶”资料以查询出茶叶方和收茶叶方
	p_a, err := data.GetAppointmentByProjectId(project_id, r.Context())
	if err != nil {
		util.Debug("cannot get project appointment from database given pr_id", project_id, err)
		report(w, r, "一脸蒙的茶博士，表示看不懂你的项目资料，请确认后再试一次。")
		return
	}
	f_id := 0
	t_id := 0
	t_s_id := 0
	var goods_slice_payer []data.Goods
	var goods_slice_payee []data.Goods
	if project.IsPrivate {
		t_id = p_a.PayeeTeamId
		f_id = p_a.PayerFamilyId
		goods_slice_t, err := data.GetAvailabilityGoodsByTeamId(t_id, r.Context())
		if err != nil {
			util.Debug("cannot get goods from database given team_id", t_id, err)
			report(w, r, "一脸蒙的茶博士，表示看不懂你的项目物资资料，请确认后再试一次。")
			return
		}
		goods_slice_f, err := data.GetAvailabilityGoodsByFamilyId(f_id, r.Context())
		if err != nil {
			util.Debug("cannot get goods from database given family_id", f_id, err)
			report(w, r, "一脸蒙的茶博士，表示看不懂你的项目物资资料，请确认后再试一次。")
			return
		}
		goods_slice_payee = goods_slice_t
		goods_slice_payer = goods_slice_f

	} else {
		t_id = p_a.PayerTeamId
		t_s_id = p_a.PayeeTeamId
		goods_slice_t, err := data.GetAvailabilityGoodsByTeamId(t_id, r.Context())
		if err != nil {
			util.Debug("cannot get goods from database given team_id", project.TeamId, err)
			report(w, r, "一脸蒙的茶博士，表示看不懂你的项目物资资料，请确认后再试一次。")
			return
		}
		goods_slice_t_s, err := data.GetAvailabilityGoodsByTeamId(t_s_id, r.Context())
		if err != nil {
			util.Debug("cannot get goods from database given team_id", project.TeamId, err)
			report(w, r, "一脸蒙的茶博士，表示看不懂你的项目物资资料，请确认后再试一次。")
			return
		}
		goods_slice_payer = goods_slice_t
		goods_slice_payee = goods_slice_t_s
	}

	var gPD data.GoodsProjectSlice
	gPD.SessUser = s_u
	gPD.IsAdmin = true
	gPD.Project = project
	gPD.GoodsSlicePayee = goods_slice_payee
	gPD.GoodsSlicePayer = goods_slice_payer

	generateHTML(w, &gPD, "layout", "navbar.private", "goods.project_new")
}

// POST /v1/goods/project_new
func GoodsProjectNewPost(w http.ResponseWriter, r *http.Request, s_u data.User) {

	// 解析表单
	err := r.ParseForm()
	if err != nil {
		util.Debug("cannot parse form data", err)
		report(w, r, "一脸蒙的茶博士，表示看不懂你提交的项目物资资料，请确认后再试一次。")
		return
	}

	// 获取项目ID
	project_id_str := r.PostFormValue("project_id")
	if project_id_str == "" {
		report(w, r, "你好，茶博士表示无法理解项目，请确认后再试。")
		return
	}
	project_id, err := strconv.Atoi(project_id_str)
	if err != nil {
		report(w, r, "你好，茶博士表示无法理解项目，请确认后再试。")
		return
	}

	// 验证项目存在
	project := data.Project{Id: project_id}
	if err := project.Get(); err != nil {
		util.Debug("cannot get project from database", err)
		report(w, r, "你好，茶博士表示无法理解项目，请确认后再试。")
		return
	}

	// 获取物资ID
	goods_id_str := r.PostFormValue("goods_id")
	if goods_id_str == "" {
		report(w, r, "你好，茶博士表示无法理解物资，请确认后再试。")
		return
	}
	goods_id, err := strconv.Atoi(goods_id_str)
	if err != nil {
		report(w, r, "你好，茶博士表示无法理解物资，请确认后再试。")
		return
	}

	// 获取数量
	quantity_str := r.PostFormValue("quantity")
	if quantity_str == "" {
		report(w, r, "你好，茶博士表示物资数量不能为空，请确认后再试。")
		return
	}
	quantity, err := strconv.Atoi(quantity_str)
	if err != nil || quantity <= 0 {
		report(w, r, "你好，茶博士表示物资数量格式错误，请确认后再试。")
		return
	}

	// 获取类别
	category_int := 0
	if cat := r.PostFormValue("category"); cat == "1" || cat == "2" {
		category_int, _ = strconv.Atoi(cat)
	} else {
		report(w, r, "你好，茶博士表示无法理解物资的类别，请确认后再试。")
		return
	}

	// 获取责任人ID（必填）
	responsible_str := r.PostFormValue("responsible_user_id")
	if responsible_str == "" {
		report(w, r, "你好，茶博士表示责任人ID不能为空，请确认后再试。")
		return
	}
	responsible_user_id, err := strconv.Atoi(responsible_str)
	if err != nil || responsible_user_id <= 0 {
		report(w, r, "你好，茶博士表示责任人ID格式错误，请确认后再试。")
		return
	}

	// 获取预期用途
	expected_usage := r.PostFormValue("expected_usage")
	le := len(expected_usage)
	if le > int(util.Config.ThreadMaxWord) {
		report(w, r, "你好，茶博士表示预期用途描述太长，请确认后再试。")
		return
	}

	// 获取提供方类型
	provider_type_str := r.PostFormValue("provider_type")
	if provider_type_str == "" {
		report(w, r, "你好，茶博士表示提供方类型不能为空，请确认后再试。")
		return
	}
	provider_type, err := strconv.Atoi(provider_type_str)
	if err != nil || (provider_type != data.ProviderTypePayee && provider_type != data.ProviderTypePayer) {
		report(w, r, "你好，茶博士表示提供方类型格式错误，请确认后再试。")
		return
	}

	// 获取备注
	notes := r.PostFormValue("notes")
	le = len(notes)
	if le > int(util.Config.ThreadMaxWord) {
		report(w, r, "你好，茶博士表示备注太长，请确认后再试。")
		return
	}

	// 创建项目物资记录
	goods_project := data.GoodsProject{
		ProjectId:         project_id,
		ResponsibleUserId: responsible_user_id,
		GoodsId:           goods_id,
		ProviderType:      provider_type,
		ExpectedUsage:     expected_usage,
		Quantity:          quantity,
		Category:          category_int,
		Status:            data.Available, // 默认状态为可用
		Notes:             notes,
	}

	if err := goods_project.Create(r.Context()); err != nil {
		util.Debug("cannot create project goods", err)
		report(w, r, "一脸蒙的茶博士，表示无法创建项目物资，请确认后再试一次。")
		return
	}

	http.Redirect(w, r, "/v1/goods/project_detail?uuid="+project.Uuid, http.StatusFound)
}

// HandleGoodsProjectDetail 处理项目物资详情页面
func HandleGoodsProjectDetail(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GoodsProjectDetail(w, r)
	default:
		report(w, r, "一脸蒙的茶博士，表示看不懂你的项目物资资料，请确认后再试一次。")
		return
	}
}

// GET /v1/goods/project_detail?uuid=xxx
func GoodsProjectDetail(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 获取项目信息
	project := data.Project{Uuid: uuid}
	if err := project.GetByUuid(); err != nil {
		util.Debug("Cannot get project by uuid", uuid, err)
		report(w, r, "你好，假作真时真亦假，无为有处有还无？")
		return
	}

	// 获取目标信息
	ob, err := project.Objective()
	if err != nil {
		util.Debug("Cannot get objective", err)
		report(w, r, "获取目标信息失败")
		return
	}

	// 获取项目物资列表
	goodsProjectList, err := data.GetGoodsByProjectId(project.Id, r.Context())
	if err != nil {
		util.Debug("Cannot get goods by project id", project.Id, err)
		report(w, r, "获取项目物资列表失败")
		return
	}

	// 获取物资详细信息
	var goodsList []data.Goods
	for _, gp := range goodsProjectList {
		goods := data.Goods{Id: gp.GoodsId}
		if err := goods.GetByIdOrUUID(r.Context()); err != nil {
			util.Debug("Cannot get goods by id", gp.GoodsId, err)
			continue
		}
		goodsList = append(goodsList, goods)
	}

	// 获取ProjectBean和ObjectiveBean
	projectBean, err := fetchProjectBean(project)
	if err != nil {
		util.Debug("Cannot fetch project bean", err)
		report(w, r, "获取项目详情失败")
		return
	}

	objectiveBean, err := fetchObjectiveBean(ob)
	if err != nil {
		util.Debug("Cannot fetch objective bean", err)
		report(w, r, "获取目标详情失败")
		return
	}

	// 准备页面数据
	templateData := data.GoodsProjectList{
		SessUser:           s_u,
		ProjectBean:        projectBean,
		QuoteObjectiveBean: objectiveBean,
		GoodsProjectList:   goodsProjectList,
		GoodsList:          goodsList,
	}

	// 权限检查
	is_admin, err := checkObjectiveAdminPermission(&ob, s_u.Id)
	if err != nil {
		util.Debug("Admin permission check failed", "userId", s_u.Id, "objectiveId", ob.Id, "error", err)
		report(w, r, "你好，玉烛滴干风里泪，晶帘隔破月中痕。")
		return
	}
	templateData.IsAdmin = is_admin

	if !is_admin {
		is_master, err := checkProjectMasterPermission(&project, s_u.Id)
		if err != nil {
			util.Debug("Permission check failed", "user_id:", s_u.Id, "error:", err)
			report(w, r, "你好，疏是枝条艳是花，春妆儿女竞奢华。")
			return
		}
		templateData.IsMaster = is_master
	}

	if !is_admin && !templateData.IsMaster {
		is_verifier := isVerifier(s_u.Id)
		templateData.IsVerifier = is_verifier
	}

	generateHTML(w, &templateData, "layout", "navbar.private", "goods.project_detail", "component_project_simple_detail", "component_sess_capacity", "component_avatar_name_gender")
}
