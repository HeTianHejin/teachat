package route

import (
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// 处理新物资的办理窗口
func HandleNewGoods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		NewGoods(w, r)
	case "POST":
		CreateGoods(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// POST /v1/goods/create
// 创建新的货物
// 从表单中读取数据，创建新的货物，存入数据库
func CreateGoods(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//限制用户登记的物资最大数量为999,防止机器人暴力填表
	if count_goods, err := data.CountGoodsByUserId(s_u.Id); err != nil || count_goods >= 999 {
		util.Warning(err, "Cannot get user place count")
		Report(w, r, "你好，闪电考拉表示您已经提交了多得数不过来，就要爆表的物资，请确定后再试。")
		return
	}
	// 从表单中读取数据
	r.ParseForm()
	category_str := r.PostFormValue("category")
	category_int, err := strconv.Atoi(category_str)
	if err != nil {
		util.Warning(err, " Cannot convert class to int")
		Report(w, r, "你好，茶博士表示无法理解物资的类型，请稍后再试。")
		return
	}
	// check category 参数是否合法
	switch category_int {
	case 1, 0:
		break
	default:
		Report(w, r, "你好，茶博士表示无法理解物资的类型，请确认后再试。")
		return
	}
	features_str := r.PostFormValue("features")
	features_int, _ := strconv.Atoi(features_str)
	// check features 参数是否合法
	switch features_int {
	case 1, 0:
		break
	default:
		Report(w, r, "你好，������表示无法理解物资的特��，请确认后再试。")
		return
	}
	name := r.PostFormValue("name")
	nickname := r.PostFormValue("nickname")
	describe := r.PostFormValue("describe")
	applicability := r.PostFormValue("applicability")
	brandname := r.PostFormValue("brandname")
	model := r.PostFormValue("model")
	color := r.PostFormValue("color")
	specification := r.PostFormValue("specification")
	manufacturer := r.PostFormValue("manufacturer")
	origin := r.PostFormValue("origin")

	//填表
	goods := &data.Goods{
		UserId:                s_u.Id,
		Name:                  name,
		Nickname:              nickname,
		Designer:              "-",
		Describe:              describe,
		Price:                 0,
		Applicability:         applicability,
		Category:              category_int,
		Specification:         specification,
		Brandname:             brandname,
		Model:                 model,
		Weight:                "-",
		Dimensions:            "-",
		Material:              "-",
		Size:                  "-",
		Color:                 color,
		NetworkConnectionType: "-",
		Features:              features_int,
		SerialNumber:          "-",
		State:                 "-",
		Origin:                origin,
		Manufacturer:          manufacturer,
		ManufacturerLink:      "-",
		EngineType:            "-",
		PurchaseLink:          "-",
	}
	//入库存档
	if err = goods.Create(); err != nil {
		util.Warning(err, s_u.Id, "Cannot create goods")
		Report(w, r, "你好，闪电考拉表示无法理解暗黑物资，请确认后再试。")
		return
	}
	//绑定这个物质信息到当前用户名下
	ug := data.UserGoods{
		UserId:  s_u.Id,
		GoodsId: goods.Id,
	}
	if err = ug.Create(); err != nil {
		util.Warning(err, s_u.Id, "Cannot create user goods")
		Report(w, r, "你好，闪电考拉表示无法理解暗黑物资，请确认后再试。")
		return
	}
	// 重定向到我的物资页面
	http.Redirect(w, r, "/v1/goods/mine", http.StatusFound)
}

// GET /v1/goods/new
// 响应需求，向用户返回登记新货物的表单页面
func NewGoods(w http.ResponseWriter, r *http.Request) {
	// ���查用户是否已经登录
	// 如果用户未登录，重定向到登录页面
	// 如果用户已登录，显示登记新货物的表单页面
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var gL data.GoodsList
	gL.SessUser = s_u
	//生成html，返回给用户
	RenderHTML(w, &gL, "layout", "navbar.private", "goods.new")
}

// GET /v1/goods/mine
// 读取登记在当前用户的全部物资（好东西）
func MyGoods(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Info(err, s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var gL data.GoodsList
	goods, err := data.GetGoodsByUserId(s_u.Id)
	if err != nil {
		util.Warning(err, "cannot get goods list given user id")
		Report(w, r, "你好，闪电考拉摸摸头，表示需要先找到眼镜，才能帮你找到私己宝贝资料。")
		return
	}
	gL.GoodsList = goods
	gL.SessUser = s_u
	// print html and send to user
	RenderHTML(w, &gL, "layout", "navbar.private", "goods.mine")
}
