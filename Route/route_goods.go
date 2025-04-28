package route

import (
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// HandleGoodsTeamUpdate() /v1/goods/team_update
func HandleGoodsTeamUpdate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GoodsTeamUpdate(w, r)
	case http.MethodPost:
		GoodsTeamUpdatePost(w, r)
	default:
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}

}

// POST /v1/goods/team_update
func GoodsTeamUpdatePost(w http.ResponseWriter, r *http.Request) {
	// Check session
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	g_id_str := r.PostFormValue("id")
	if g_id_str == "" {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	g_id, err := strconv.Atoi(g_id_str)
	if err != nil {
		Report(w, r, "一脸蒙的茶博士，看不懂你的物资资料，请确认后再试一次。")
		return
	}

	team_id_str := r.PostFormValue("team_id")
	if team_id_str == "" {
		Report(w, r, "茶博士耸耸肩说，你无法查看不存在的物资，请确认后再试一次。")
		return
	}
	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	team, err := data.GetTeam(team_id)
	if err != nil {
		util.Error(s.Email, "Cannot get team from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	is_member, err := team.IsMember(s_u.Id)
	if err != nil {
		util.Error(s.Email, "Cannot get team member from database")
		Report(w, r, "茶博士耸耸肩说，今天不可以查看物资的资料，请确认后再试一次。")
		return
	}
	if !is_member {
		Report(w, r, "茶博士耸耸肩说，非成员无权查看物资的资料，请确认后再试一次。")
		return
	}
	tg := data.GoodsTeam{GoodsId: g_id, TeamId: team_id}
	if err = tg.GetByTeamIdAndGoodsId(); err != nil {
		util.Error(s.Email, "Cannot get team goods from database")
		Report(w, r, "一脸蒙的茶博士，表示根据提供的参数无法查到物资资料，请确认后再试一次。")
		return
	}
	g := data.Goods{Id: g_id}
	if err = g.Get(); err != nil {
		util.Error(s.Email, "Cannot get goods from database")
		Report(w, r, "满头大汗的茶博士，表示找不到茶团物资，请稍后再试一次。")
		return
	}

	category_int := 0
	if cat := r.PostFormValue("category"); cat == "0" || cat == "1" {
		category_int, _ = strconv.Atoi(cat) // Safe to ignore error as we've already validated
	} else {
		Report(w, r, "你好，茶博士表示无法理解物资的类型，请确认后再试。")
		return
	}
	// Get goods name
	le := len(r.PostFormValue("goods_name"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的名称太长或者太短，请确认后再试。")
		return
	}
	goods_name := r.PostFormValue("goods_name")
	//nickname
	le = len(r.PostFormValue("nickname"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的昵称太长或者太短，请确认后再试。")
		return
	}
	goods_nickname := r.PostFormValue("nickname")

	//features
	goods_features := 0
	if fea := r.PostFormValue("features"); fea == "0" || fea == "1" {
		goods_features, _ = strconv.Atoi(fea) // Safe to ignore error as we've already validated
	} else {
		Report(w, r, "你好，茶博士表示无法理解物资的特性，请确认后再试。")
		return
	}

	goods_designer := r.PostFormValue("designer")
	le = len(goods_designer)
	if le > 45 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的设计者太长或者太短，请确认后再试。")
		return
	}

	// Get goods description
	describe := r.PostFormValue("describe")
	le = len(describe)
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的描述太长或者太短，请确认后再试。")
		return
	}

	//applicability
	le = len(r.PostFormValue("applicability"))
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的适用范围太长或者太短，请确认后再试。")
		return
	}
	goods_applicability := r.PostFormValue("applicability")

	//brandname
	le = len(r.PostFormValue("brandname"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的品牌名称太长或者太短，请确认后再试。")
		return
	}
	brand_name := r.PostFormValue("brandname")

	//model
	le = len(r.PostFormValue("model"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的型号太长或者太短，请确认后再试。")
		return
	}
	goods_model := r.PostFormValue("model")

	//color
	le = len(r.PostFormValue("color"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的款色描述太长或者太短，请确认后再试。")
		return
	}
	goods_color := r.PostFormValue("color")

	//specification
	le = len(r.PostFormValue("specification"))
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的规格描述太长或者太短，请确认后再试。")
		return
	}
	goods_specification := r.PostFormValue("specification")

	//manufacturer
	le = len(r.PostFormValue("manufacturer"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的制造商太长或者太短，请确认后再试。")
		return
	}
	goods_manufacturer := r.PostFormValue("manufacturer")

	//origin
	le = len(r.PostFormValue("origin"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的产地太长或者太短，请确认后再试。")
		return
	}
	goods_origin := r.PostFormValue("origin")

	goods_price_str := r.PostFormValue("price")
	// Check goods price
	if goods_price_str == "" {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}
	// Convert goods price to float64
	goods_price, err := strconv.ParseFloat(goods_price_str, 64)
	// Check goods price
	if err != nil {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}
	//0<goods_price<100,000,000
	if goods_price < 0 || goods_price > 100000000 {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}

	goods_weight_str := r.PostFormValue("weight")
	// Check goods weight
	if goods_weight_str == "" {
		Report(w, r, "你好，茶博士表示物资的重量太长或者太短，请确认后再试。")
		return
	}
	// Convert goods weight to float64
	goods_weight, err := strconv.ParseFloat(goods_weight_str, 64)
	// Check goods weight
	if err != nil {
		Report(w, r, "你好，茶博士表示物资的重量太长或者太短，请确认后再试。")
		return
	}

	goods_dimensions_str := r.PostFormValue("dimensions")
	if goods_dimensions_str == "" {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}
	le = len(goods_dimensions_str)
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}

	goods_material := r.PostFormValue("material")
	le = len(goods_material)
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的材质太长或者太短，请确认后再试。")
		return
	}

	goods_size := r.PostFormValue("size")
	le = len(goods_size)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}

	goods_net_con_typ := r.PostFormValue("connection_type")
	le = len(goods_net_con_typ)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的网络连接类型太长或者太短，请确认后再试。")
		return
	}

	goods_sn := r.PostFormValue("serial_number")
	le = len(goods_sn)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的序列号太长或者太短，请确认后再试。")
		return
	}

	goods_expi_date := r.PostFormValue("expiration_date")
	le = len(goods_expi_date)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的过期日期太长或者太短，请确认后再试。")
		return
	}

	goods_state := r.PostFormValue("state")
	le = len(goods_state)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的状态太长或者太短，请确认后再试。")
		return
	}

	goods_manu_url_str := r.PostFormValue("official_website")
	le = len(goods_manu_url_str)
	if le > 256 {
		Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
		return
	}
	//check url
	// if goods_manu_url_str == "" {
	// 	Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
	// 	return
	// }
	// _, err = url.Parse(goods_manu_url_str)
	// if err != nil {
	// 	Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
	// 	return
	// }

	goods_engine_typ := r.PostFormValue("engine_type")
	le = len(goods_engine_typ)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的引擎类型太长或者太短，请确认后再试。")
		return
	}

	goods_purchase_url_str := r.PostFormValue("purchase_url")
	le = len(goods_purchase_url_str)
	if le > 256 {
		Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
		return
	}
	//check url
	// if goods_purchase_url_str == "" {
	// 	Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
	// 	return
	// }
	// _, err = url.Parse(goods_purchase_url_str)
	// if err != nil {
	// 	Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
	// 	return
	// }

	o_goods := data.Goods{
		Id:                    g.Id,
		RecorderUserId:        s_u.Id,
		Name:                  goods_name,
		Nickname:              goods_nickname,
		Designer:              goods_designer,
		Describe:              describe,
		Price:                 goods_price,
		Applicability:         goods_applicability,
		Category:              category_int,
		Specification:         goods_specification,
		BrandName:             brand_name,
		Model:                 goods_model,
		Weight:                goods_weight,
		Dimensions:            goods_dimensions_str,
		Material:              goods_material,
		Size:                  goods_size,
		Color:                 goods_color,
		NetworkConnectionType: goods_net_con_typ,
		Features:              goods_features,
		SerialNumber:          goods_sn,
		State:                 goods_state,
		Origin:                goods_origin,
		Manufacturer:          goods_manufacturer,
		ManufacturerURL:       goods_manu_url_str,
		EngineType:            goods_engine_typ,
		PurchaseURL:           goods_purchase_url_str,
	}
	if err := o_goods.Update(); err != nil {
		util.Error(s.Email, "Cannot update goods from database")
		Report(w, r, "一脸蒙的茶博士，表示无法更新物资，请确认后再试一次。")
		return
	}
	http.Redirect(w, r, "/v1/goods/team_detail?id="+g_id_str+"&team_id="+team_id_str, http.StatusFound)
}

// GET /v1/goods/team_update?id=xxx&team_id=xxx
func GoodsTeamUpdate(w http.ResponseWriter, r *http.Request) {

	// Check session
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	g_id_str := r.URL.Query().Get("id")
	if g_id_str == "" {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	g_id, err := strconv.Atoi(g_id_str)
	if err != nil {
		Report(w, r, "一脸蒙的茶博士，看不懂你的物资资料，请确认后再试一次。")
		return
	}

	team_id_str := r.URL.Query().Get("team_id")
	if team_id_str == "" {
		Report(w, r, "茶博士耸耸肩说，你无法查看不存在的物资，请确认后再试一次。")
		return
	}
	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	team, err := data.GetTeam(team_id)
	if err != nil {
		util.Error(s.Email, "Cannot get team from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	is_member, err := team.IsMember(s_u.Id)
	if err != nil {
		util.Error(s.Email, "Cannot get team member from database")
		Report(w, r, "茶博士耸耸肩说，今天不可以查看物资的资料，请确认后再试一次。")
		return
	}
	if !is_member {
		Report(w, r, "茶博士耸耸肩说，非成员无权查看物资的资料，请确认后再试一次。")
		return
	}
	tg := data.GoodsTeam{GoodsId: g_id, TeamId: team_id}
	if err = tg.GetByTeamIdAndGoodsId(); err != nil {
		util.Error(s.Email, "Cannot get team goods from database")
		Report(w, r, "一脸蒙的茶博士，表示根据提供的参数无法查到物资资料，请确认后再试一次。")
		return
	}
	g := data.Goods{Id: g_id}
	if err = g.Get(); err != nil {
		util.Error(s.Email, "Cannot get goods from database")
		Report(w, r, "满头大汗的茶博士，表示找不到茶团物资，请稍后再试一次。")
		return
	}

	var gTD data.GoodsTeamDetail

	gTD.SessUser = s_u
	gTD.IsAdmin = true
	gTD.Team = team
	gTD.Goods = g

	RenderHTML(w, &gTD, "layout", "navbar.private", "goods.team_update")

}

// GET /v1/goods/team_detail?id=xxx&team_id=xxx
func GoodsTeamDetail(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	g_id_str := r.URL.Query().Get("id")
	if g_id_str == "" {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	g_id, err := strconv.Atoi(g_id_str)
	if err != nil {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}

	team_id_str := r.URL.Query().Get("team_id")
	if team_id_str == "" {
		Report(w, r, "茶博士耸耸肩说，你无权查看物资的资料，请确认后再试一次。")
		return
	}
	if team_id_str == "0" || team_id_str == "2" {
		Report(w, r, "茶博士耸耸肩说，你无权查看物资的资料，请确认后再试一次。")
		return
	}

	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	team, err := data.GetTeam(team_id)
	if err != nil {
		util.Error(s.Email, "Cannot get team from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	is_member, err := team.IsMember(s_u.Id)
	if err != nil {
		Report(w, r, "茶博士耸耸肩说，你无权查看物资的资料，请确认后再试一次。")
		return
	}
	if !is_member {
		Report(w, r, "茶博士耸耸肩说，你无权查看物资的资料，请确认后再试一次。")
		return
	}
	tg := data.GoodsTeam{GoodsId: g_id, TeamId: team_id}
	if err = tg.GetByTeamIdAndGoodsId(); err != nil {
		util.Error(s.Email, "Cannot get team goods from database")
		Report(w, r, "一脸蒙的茶博士，表示根据提供的参数无法查到物资资料，请确认后再试一次。")
		return
	}
	g := data.Goods{Id: g_id}
	if err = g.Get(); err != nil {
		util.Error(s.Email, "Cannot get goods from database")
		Report(w, r, "满头大汗的茶博士，表示找不到茶团物资，请稍后再试一次。")
		return
	}

	var tGD data.GoodsTeamDetail

	tGD.SessUser = s_u
	tGD.IsAdmin = true
	tGD.Team = team
	tGD.Goods = g

	RenderHTML(w, &tGD, "layout", "navbar.private", "goods.team_detail")

}

// GET /v1/goods/team?id=
func GoodsTeam(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	team_uuid := r.URL.Query().Get("id")
	if team_uuid == "" {
		Report(w, r, "茶博士耸耸肩说，你无权查看物资的资料，请确认后再试一次。")
		return
	}

	team, err := data.GetTeamByUUID(team_uuid)
	if err != nil {
		util.Error(s.Email, "Cannot get team from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	if team.Id == 0 || team.Id == 2 {
		Report(w, r, "茶博士耸耸肩说，你无权查看物资的资料，请确认后再试一次。")
		return
	}

	is_member, err := team.IsMember(s_u.Id)
	if err != nil {
		util.Error(s.Email, "Cannot get team member from database")
		Report(w, r, "茶博士耸耸肩说，你无权查看物资的资料，请确认后再试一次。")
		return
	}
	if !is_member {
		Report(w, r, "茶博士耸耸肩说，你无权查看物资的资料，请确认后再试一次。")
		return
	}

	// Get []goods from database
	t_g := data.GoodsTeam{TeamId: team.Id}
	t_goods_slice, err := t_g.GetAllGoodsByTeamId()
	if err != nil {
		util.Error(s.Email, "Cannot get goods from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}

	var gTS data.GoodsTeamSlice

	gTS.SessUser = s_u
	gTS.IsAdmin = true
	gTS.Team = team
	gTS.GoodsSlice = t_goods_slice

	RenderHTML(w, &gTS, "layout", "navbar.private", "goods.team")

}

// HandleGoodsFamilyNew() /v1/goods/family_new
func HandleGoodsFamilyNew(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GoodsFamilyNewGet(w, r)
	case http.MethodPost:
		GoodsFamilyNewPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// // 处理茶团新物资登记的办理窗口
// // HandleGoodsTeamNew() /v1/goods/team_new
func HandleGoodsTeamNew(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GoodsTeamNewGet(w, r)
	case http.MethodPost:
		GoodsTeamNewPost(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// POST /v1/goods/team_new
func GoodsTeamNewPost(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// Get form data
	err = r.ParseForm()
	// Check form data
	if err != nil {
		util.Error(s.Email, "Cannot parse form data")
		//http.Redirect(w, r, "/v1/goods/new", http.StatusFound)
		Report(w, r, "一脸蒙的茶博士，表示看不懂你提交的物资资料，请确认后再试一次。")
		return
	}

	team_id_str := r.PostFormValue("team_id")
	// Check team_id
	if team_id_str == "" {
		Report(w, r, "你好，茶博士表示无法理解物资的团队，请确认后再试。")
		return
	}
	if team_id_str == "0" || team_id_str == "2" {
		Report(w, r, "你好，茶博士表示无法理解物资的团队，请确认后再试。")
		return
	}
	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		Report(w, r, "你好，茶博士表示无法理解物资的团队，请确认后再试。")
		return
	}

	team, err := data.GetTeam(team_id)
	if err != nil {
		util.Error(s.Email, "Cannot get team from database")
		Report(w, r, "你好，茶博士表示无法理解物资的团队，请确认后再试。")
		return
	}

	//check team member
	is_member, err := team.IsMember(s_u.Id)
	if err != nil {
		util.Error(s.Email, "Cannot get team member from database")
		Report(w, r, "你好，茶博士表示无法理解你的团队，请确认后再试。")
		return
	}

	if !is_member {
		Report(w, r, "你好，茶博士表示无法理解你的团队，请确认后再试。")
		return
	}
	// Max goods count check
	goods_teams := data.GoodsTeam{TeamId: team.Id}
	count_teams_goods, err := goods_teams.CountByTeamId()
	if err != nil {
		util.Error(s.Email, "Cannot get team goods from database")
		Report(w, r, "你好，茶博士表示无法理解你的团队，请确认后再试。")
		return
	}
	if count_teams_goods >= 9999 {
		Report(w, r, "你好，茶博士表示你的团队已经达到最大物资数量，请确认后再试。")
		return
	}

	category_int := 0
	if cat := r.PostFormValue("category"); cat == "0" || cat == "1" {
		category_int, _ = strconv.Atoi(cat) // Safe to ignore error as we've already validated
	} else {
		Report(w, r, "你好，茶博士表示无法理解物资的类型，请确认后再试。")
		return
	}
	// Get goods name
	le := len(r.PostFormValue("goods_name"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的名称太长或者太短，请确认后再试。")
		return
	}
	goods_name := r.PostFormValue("goods_name")
	//nickname
	le = len(r.PostFormValue("nickname"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的昵称太长或者太短，请确认后再试。")
		return
	}
	goods_nickname := r.PostFormValue("nickname")

	//features
	goods_features := 0
	if fea := r.PostFormValue("features"); fea == "0" || fea == "1" {
		goods_features, _ = strconv.Atoi(fea) // Safe to ignore error as we've already validated
	} else {
		Report(w, r, "你好，茶博士表示无法理解物资的特性，请确认后再试。")
		return
	}

	goods_designer := r.PostFormValue("designer")
	le = len(goods_designer)
	if le > 45 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的设计者太长或者太短，请确认后再试。")
		return
	}

	// Get goods description
	le = len(r.PostFormValue("describe"))
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的描述太长或者太短，请确认后再试。")
		return
	}
	describe := r.PostFormValue("describe")

	//applicability
	le = len(r.PostFormValue("applicability"))
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的适用范围太长或者太短，请确认后再试。")
		return
	}
	goods_applicability := r.PostFormValue("applicability")

	//brandname
	le = len(r.PostFormValue("brandname"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的品牌名称太长或者太短，请确认后再试。")
		return
	}
	brand_name := r.PostFormValue("brandname")

	//model
	le = len(r.PostFormValue("model"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的型号太长或者太短，请确认后再试。")
		return
	}
	goods_model := r.PostFormValue("model")

	//color
	le = len(r.PostFormValue("color"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的款色描述太长或者太短，请确认后再试。")
		return
	}
	goods_color := r.PostFormValue("color")

	//specification
	le = len(r.PostFormValue("specification"))
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的规格描述太长或者太短，请确认后再试。")
		return
	}
	goods_specification := r.PostFormValue("specification")

	//manufacturer
	le = len(r.PostFormValue("manufacturer"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的制造商太长或者太短，请确认后再试。")
		return
	}
	goods_manufacturer := r.PostFormValue("manufacturer")

	//origin
	le = len(r.PostFormValue("origin"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的产地太长或者太短，请确认后再试。")
		return
	}
	goods_origin := r.PostFormValue("origin")

	goods_price_str := r.PostFormValue("price")
	// Check goods price
	if goods_price_str == "" {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}
	// Convert goods price to float64
	goods_price, err := strconv.ParseFloat(goods_price_str, 64)
	// Check goods price
	if err != nil {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}
	//0<goods_price<100,000,000
	if goods_price < 0 || goods_price > 100000000 {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}

	goods_weight_str := r.PostFormValue("weight")
	// Check goods weight
	if goods_weight_str == "" {
		Report(w, r, "你好，茶博士表示物资的重量太长或者太短，请确认后再试。")
		return
	}
	// Convert goods weight to float64
	goods_weight, err := strconv.ParseFloat(goods_weight_str, 64)
	// Check goods weight
	if err != nil {
		Report(w, r, "你好，茶博士表示物资的重量太长或者太短，请确认后再试。")
		return
	}

	goods_dimensions_str := r.PostFormValue("dimensions")
	if goods_dimensions_str == "" {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}
	le = len(goods_dimensions_str)
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}

	goods_material := r.PostFormValue("material")
	le = len(goods_material)
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的材质太长或者太短，请确认后再试。")
		return
	}

	goods_size := r.PostFormValue("size")
	le = len(goods_size)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}

	goods_net_con_typ := r.PostFormValue("connection_type")
	le = len(goods_net_con_typ)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的网络连接类型太长或者太短，请确认后再试。")
		return
	}

	goods_sn := r.PostFormValue("serial_number")
	le = len(goods_sn)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的序列号太长或者太短，请确认后再试。")
		return
	}

	goods_expi_date := r.PostFormValue("expiration_date")
	le = len(goods_expi_date)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的过期日期太长或者太短，请确认后再试。")
		return
	}

	goods_state := r.PostFormValue("state")
	le = len(goods_state)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的状态太长或者太短，请确认后再试。")
		return
	}

	goods_manu_url_str := r.PostFormValue("official_website")
	le = len(goods_manu_url_str)
	if le > 256 {
		Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
		return
	}
	//check url
	// if goods_manu_url_str == "" {
	// 	Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
	// 	return
	// }
	// _, err = url.Parse(goods_manu_url_str)
	// if err != nil {
	// 	Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
	// 	return
	// }

	goods_engine_typ := r.PostFormValue("engine_type")
	le = len(goods_engine_typ)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的引擎类型太长或者太短，请确认后再试。")
		return
	}

	goods_purchase_url_str := r.PostFormValue("purchase_url")
	le = len(goods_purchase_url_str)
	if le > 256 {
		Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
		return
	}
	//check url
	// if goods_purchase_url_str == "" {
	// 	Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
	// 	return
	// }
	// _, err = url.Parse(goods_purchase_url_str)
	// if err != nil {
	// 	Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
	// 	return
	// }

	new_goods := data.Goods{
		RecorderUserId:        s_u.Id,
		Name:                  goods_name,
		Nickname:              goods_nickname,
		Designer:              goods_designer,
		Describe:              describe,
		Price:                 goods_price,
		Applicability:         goods_applicability,
		Category:              category_int,
		Specification:         goods_specification,
		BrandName:             brand_name,
		Model:                 goods_model,
		Weight:                goods_weight,
		Dimensions:            goods_dimensions_str,
		Material:              goods_material,
		Size:                  goods_size,
		Color:                 goods_color,
		NetworkConnectionType: goods_net_con_typ,
		Features:              goods_features,
		SerialNumber:          goods_sn,
		State:                 goods_state,
		Origin:                goods_origin,
		Manufacturer:          goods_manufacturer,
		ManufacturerURL:       goods_manu_url_str,
		EngineType:            goods_engine_typ,
		PurchaseURL:           goods_purchase_url_str,
	}
	if err := new_goods.Create(); err != nil {
		util.Error(s.Email, "Cannot create new goods")
		Report(w, r, "一脸蒙的茶博士，表示无法创建物资，请确认后再试一次。")
		return
	}

	// Create team goods
	tg := data.GoodsTeam{
		TeamId:  team_id,
		GoodsId: new_goods.Id,
	}
	if err := tg.Create(); err != nil {
		util.Error(s.Email, "Cannot create team goods")
		Report(w, r, "一脸蒙的茶博士，表示无法绑定团队物资，请确认后再试一次。")
		return
	}
	http.Redirect(w, r, "/v1/goods/team?id="+team.Uuid, http.StatusFound)

}

// GET /v1/goods/team_new?id=xxx
func GoodsTeamNewGet(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	s_u, err := s.User()
	if err != nil {
		util.Error(s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	team_id_str := r.URL.Query().Get("id")
	team_id, err := strconv.Atoi(team_id_str)
	if err != nil {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}

	team, err := data.GetTeam(team_id)
	if err != nil {
		util.Error(s.Email, "Cannot get team from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资资料，请确认后再试一次。")
		return
	}
	//check if user is member of the team
	is_member, err := team.IsMember(s_u.Id)
	if err != nil {
		util.Error(s.Email, "Cannot get team member from database")
		Report(w, r, "茶博士耸耸肩说，你无权处理茶团物资的资料，请确认后再试一次。")
		return
	}

	if !is_member {
		Report(w, r, "茶博士耸耸肩说，你无权处理茶团物资的资料，请确认后再试一次。")
		return
	}

	var gL data.GoodsTeamSlice

	gL.IsAdmin = true
	gL.Team = team
	gL.SessUser = s_u

	RenderHTML(w, &gL, "layout", "navbar.private", "goods.team_new")
}

// GET /goods/family_new?id=xxx
func GoodsFamilyNewGet(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	s_u, err := s.User()
	if err != nil {
		util.Error(s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	family_id_str := r.URL.Query().Get("id")
	family_id, err := strconv.Atoi(family_id_str)
	if err != nil {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的家庭资料，请确认后再试。")
		return
	}

	if family_id == 0 {
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的家庭资料，请确认后再试一次。")
		return
	}

	family, err := data.GetFamily(family_id)
	if err != nil {
		util.Error(s.Email, "Cannot get family from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的家庭资料，请确认。")
		return
	}
	//check if user is member of the family
	is_member, err := family.IsMember(s_u.Id)
	if err != nil {
		util.Error(s.Email, "Cannot get family member from database")
		Report(w, r, "茶博士耸耸肩说，你无权处理家庭物资的资料，请确认后再试一次。")
		return
	}

	if !is_member {
		Report(w, r, "茶博士耸耸肩说，你无权处理家庭物资的资料，请确认后再试一次。")
		return
	}

	var gL data.GoodsFamilySlice

	gL.IsAdmin = true
	gL.Family = family
	gL.SessUser = s_u

	RenderHTML(w, &gL, "layout", "navbar.private", "goods.new")
}
func GoodsFamilyNewPost(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	// Max goods count check

	// Get form data
	err = r.ParseForm()
	// Check form data
	if err != nil {
		util.Error(s.Email, "Cannot parse form data")
		//http.Redirect(w, r, "/v1/goods/new", http.StatusFound)
		Report(w, r, "一脸蒙的茶博士，表示看不懂你提交的物资资料，请确认后再试一次。")
		return
	}
	// Get family_id
	family_id_str := r.PostFormValue("family_id")
	// Check family_id
	if family_id_str == "" {
		Report(w, r, "你好，茶博士表示无法理解物资的家庭，请确认后再试一次。")
		return
	}

	if family_id_str == "0" {
		Report(w, r, "你好，茶博士表示无法理解物资的家庭，请确认后再试。")
		return
	}
	// Convert family_id to int
	family_id, err := strconv.Atoi(family_id_str)
	// Check family_id
	if err != nil {
		Report(w, r, "你好，茶博士表示无法理解物资的家庭，请确认后再试。")
		return
	}
	family, err := data.GetFamily(family_id)
	// Check family
	if err != nil {
		Report(w, r, "你好，茶博士表示无法理解物资的家庭，请确认后再试。")
		return
	}
	//check if user is member of the family
	is_member, err := family.IsMember(s_u.Id)
	// Check family member
	if err != nil {
		Report(w, r, "你好，茶博士表示无法理解物资的家庭，请确认后再试。")
		return
	}
	// Check family member
	if !is_member {
		Report(w, r, "你好，茶博士表示无法理解物资的家庭，请确认后再试。")
		return
	}

	category_int := 0
	if cat := r.PostFormValue("category"); cat == "0" || cat == "1" {
		category_int, _ = strconv.Atoi(cat) // Safe to ignore error as we've already validated
	} else {
		Report(w, r, "你好，茶博士表示无法理解物资的类型，请确认后再试。")
		return
	}
	// Get goods name
	le := len(r.PostFormValue("goods_name"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的名称太长或者太短，请确认后再试。")
		return
	}
	goods_name := r.PostFormValue("goods_name")
	//nickname
	le = len(r.PostFormValue("nickname"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的昵称太长或者太短，请确认后再试。")
		return
	}
	goods_nickname := r.PostFormValue("nickname")

	//features
	goods_features := 0
	if fea := r.PostFormValue("features"); fea == "0" || fea == "1" {
		goods_features, _ = strconv.Atoi(fea) // Safe to ignore error as we've already validated
	} else {
		Report(w, r, "你好，茶博士表示无法理解物资的特性，请确认后再试。")
		return
	}

	goods_designer := r.PostFormValue("designer")
	le = len(goods_designer)
	if le > 45 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的设计者太长或者太短，请确认后再试。")
		return
	}

	// Get goods description
	describe := r.PostFormValue("describe")
	le = len(describe)
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的描述太长或者太短，请确认后再试。")
		return
	}

	//applicability
	le = len(r.PostFormValue("applicability"))
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的适用范围太长或者太短，请确认后再试。")
		return
	}
	goods_applicability := r.PostFormValue("applicability")

	//brandname
	le = len(r.PostFormValue("brandname"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的品牌名称太长或者太短，请确认后再试。")
		return
	}
	brand_name := r.PostFormValue("brandname")

	//model
	le = len(r.PostFormValue("model"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的型号太长或者太短，请确认后再试。")
		return
	}
	goods_model := r.PostFormValue("model")

	//color
	le = len(r.PostFormValue("color"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的款色描述太长或者太短，请确认后再试。")
		return
	}
	goods_color := r.PostFormValue("color")

	//specification
	le = len(r.PostFormValue("specification"))
	if le > 456 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的规格描述太长或者太短，请确认后再试。")
		return
	}
	goods_specification := r.PostFormValue("specification")

	//manufacturer
	le = len(r.PostFormValue("manufacturer"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的制造商太长或者太短，请确认后再试。")
		return
	}
	goods_manufacturer := r.PostFormValue("manufacturer")

	//origin
	le = len(r.PostFormValue("origin"))
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的产地太长或者太短，请确认后再试。")
		return
	}
	goods_origin := r.PostFormValue("origin")

	goods_price_str := r.PostFormValue("price")
	// Check goods price
	if goods_price_str == "" {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}
	// Convert goods price to float64
	goods_price, err := strconv.ParseFloat(goods_price_str, 64)
	// Check goods price
	if err != nil {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}
	//0<goods_price<100,000,000
	if goods_price < 0 || goods_price > 100000000 {
		Report(w, r, "你好，茶博士表示物资的价格太长或者太短，请确认后再试。")
		return
	}

	goods_weight_str := r.PostFormValue("weight")
	// Check goods weight
	if goods_weight_str == "" {
		Report(w, r, "你好，茶博士表示物资的重量太长或者太短，请确认后再试。")
		return
	}
	// Convert goods weight to float64
	goods_weight, err := strconv.ParseFloat(goods_weight_str, 64)
	// Check goods weight
	if err != nil {
		Report(w, r, "你好，茶博士表示物资的重量太长或者太短，请确认后再试。")
		return
	}

	goods_dimensions_str := r.PostFormValue("dimensions")
	if goods_dimensions_str == "" {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}
	le = len(goods_dimensions_str)
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}

	goods_material := r.PostFormValue("material")
	le = len(goods_material)
	if le > 50 || le < 1 {
		Report(w, r, "你好，茶博士表示物资的材质太长或者太短，请确认后再试。")
		return
	}

	goods_size := r.PostFormValue("size")
	le = len(goods_size)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的尺寸太长或者太短，请确认后再试。")
		return
	}

	goods_net_con_typ := r.PostFormValue("connection_type")
	le = len(goods_net_con_typ)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的网络连接类型太长或者太短，请确认后再试。")
		return
	}

	goods_sn := r.PostFormValue("serial_number")
	le = len(goods_sn)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的序列号太长或者太短，请确认后再试。")
		return
	}

	goods_expi_date := r.PostFormValue("expiration_date")
	le = len(goods_expi_date)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的过期日期太长或者太短，请确认后再试。")
		return
	}

	goods_state := r.PostFormValue("state")
	le = len(goods_state)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的状态太长或者太短，请确认后再试。")
		return
	}

	goods_manu_url_str := r.PostFormValue("official_website")
	le = len(goods_manu_url_str)
	if le > 256 {
		Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
		return
	}
	//check url
	// if goods_manu_url_str == "" {
	// 	Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
	// 	return
	// }
	// _, err = url.Parse(goods_manu_url_str)
	// if err != nil {
	// 	Report(w, r, "你好，茶博士表示物资的官方网站太长或者太短，请确认后再试。")
	// 	return
	// }

	goods_engine_typ := r.PostFormValue("engine_type")
	le = len(goods_engine_typ)
	if le > 50 {
		Report(w, r, "你好，茶博士表示物资的引擎类型太长或者太短，请确认后再试。")
		return
	}

	goods_purchase_url_str := r.PostFormValue("purchase_url")
	le = len(goods_purchase_url_str)
	if le > 256 {
		Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
		return
	}
	//check url
	// if goods_purchase_url_str == "" {
	// 	Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
	// 	return
	// }
	// _, err = url.Parse(goods_purchase_url_str)
	// if err != nil {
	// 	Report(w, r, "你好，茶博士表示物资的购买链接太长或者太短，请确认后再试。")
	// 	return
	// }

	new_goods := data.Goods{
		RecorderUserId:        s_u.Id,
		Name:                  goods_name,
		Nickname:              goods_nickname,
		Designer:              goods_designer,
		Describe:              describe,
		Price:                 goods_price,
		Applicability:         goods_applicability,
		Category:              category_int,
		Specification:         goods_specification,
		BrandName:             brand_name,
		Model:                 goods_model,
		Weight:                goods_weight,
		Dimensions:            goods_dimensions_str,
		Material:              goods_material,
		Size:                  goods_size,
		Color:                 goods_color,
		NetworkConnectionType: goods_net_con_typ,
		Features:              goods_features,
		SerialNumber:          goods_sn,
		State:                 goods_state,
		Origin:                goods_origin,
		Manufacturer:          goods_manufacturer,
		ManufacturerURL:       goods_manu_url_str,
		EngineType:            goods_engine_typ,
		PurchaseURL:           goods_purchase_url_str,
	}
	if err := new_goods.Create(); err != nil {
		util.Error(s.Email, "Cannot create new goods")
		Report(w, r, "一脸蒙的茶博士，表示无法创建物资，请确认后再试一次。")
		return
	}

	fg := data.GoodsFamily{
		FamilyId: family.Id,
		GoodsId:  new_goods.Id,
	}
	if err := fg.Create(); err != nil {
		util.Error(s.Email, "Cannot create new goods family")
		Report(w, r, "一脸蒙的茶博士，表示无法创建物资，请确认后再试一次。")
		return
	}

	http.Redirect(w, r, "/v1/goods/family?id="+family_id_str, http.StatusFound)

}

// GET /v1/goods/collect?id=xxx
func GoodsCollect(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	goods_uuid := r.URL.Query().Get("id")

	t_goods := data.Goods{Uuid: goods_uuid}

	if err = t_goods.GetByUuid(); err != nil {
		util.Error(s.Email, "Cannot get goods from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资，请确认。")
		return
	}

	//检查用户是否已经收藏过

	t_goods_user := data.GoodsUser{
		UserId:  s_u.Id,
		GoodsId: t_goods.Id,
	}
	exist, err := t_goods_user.CheckUserGoodsExist()
	if err != nil {
		util.Error(s.Email, "Cannot check goods user from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资，请确认。")
		return
	}
	//如果已经收藏过，就不再收藏
	if exist {
		Report(w, r, "你已经收藏过了，不用再收藏了。")
		return
	}
	//max < 99
	//count
	count, err := t_goods_user.CountByUserId()
	if err != nil {
		util.Error(s.Email, "Cannot count goods user from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资，请确认。")
		return
	}
	if count > 99 {
		Report(w, r, "你已经收藏太多了，地库都藏不下了。")
		return
	}
	//insert
	if err = t_goods_user.Create(); err != nil {
		util.Error(s.Email, "Cannot create goods user from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资，请确认。")
		return
	}

	http.Redirect(w, r, "/v1/goods/eye_on", http.StatusFound)
}

// GoodsEyeOn() /v1/goods/eye_on
func GoodsEyeOn(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Error("Cannot get user from session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	gu := data.GoodsUser{UserId: s_u.Id}
	goods_slice, err := gu.GetGoodsByUserId()
	if err != nil {
		util.Error(s.Email, "Cannot get goods from database")
		Report(w, r, "一脸蒙的茶博士，表示看不懂你的物资，请确认。")
		return
	}

	gL := data.GoodsUserSlice{
		SessUser:   s_u,
		GoodsSlice: goods_slice,
	}

	RenderHTML(w, &gL, "layout", "navbar.private", "goods.eye_on")
}
