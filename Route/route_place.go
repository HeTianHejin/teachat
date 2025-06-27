package route

import (
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/place/collect?id=xxx
// func PlaceCollect() 把指定id的place添加到茶友地点收藏本里
func PlaceCollect(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
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
	err = r.ParseForm()
	if err != nil {
		util.Debug("Cannot parse form", err)
		Report(w, r, "你好，茶博士表示无法理解地方的类型，请稍后再试。")
		return
	}
	id_str := r.FormValue("id")
	t_place_id, err := strconv.Atoi(id_str)
	if err != nil {
		util.Debug("Cannot convert id to int", err)
		Report(w, r, "你好，茶博士表示无法理解地方的类型，请稍后再试。")
		return
	}
	//检查地方是否存在
	t_place := data.Place{Id: t_place_id}
	if err := t_place.Get(); err != nil {
		util.Debug("Cannot get place by id", err)
		Report(w, r, "你好，茶博士表示无法收藏地方，请稍后再试。")
		return
	}
	//检查用户是否已经收藏过该地方
	exist, err := data.CheckUserPlace(s_u.Id, t_place_id)
	if err != nil {
		util.Debug(s_u.Id, t_place_id, "Cannot check user place")
		Report(w, r, "你好，茶博士表示该地方有外星人出没，请稍后再试。")
		return
	}
	if exist {
		Report(w, r, "你好，茶博士表示您已经收藏过该地方，请不要重复收藏。")
		return
	}
	//检查用户收藏的地方数量是否超过99
	count, err := data.CountUserPlace(s_u.Id)
	if err != nil {
		util.Debug("Cannot get user place count", err)
		Report(w, r, "你好，满头大汗的茶博士居然找不到提及的地方，请确定后再试。")
		return
	}
	if count >= 99 {
		Report(w, r, "你好，茶博士表示您已经提交了多得数不过来，就要爆表的地方，请确定后再试。")
		return
	}

	//收藏该地方
	user_place := data.UserPlace{UserId: s_u.Id, PlaceId: t_place_id}
	if err := user_place.Create(); err != nil {
		util.Debug("Cannot collect place", err)
		Report(w, r, "你好，茶博士表示无法收藏地方，请稍后再试。")
		return
	}

	//检查用户是否已经设置默认品茶地点，
	//如果没有设置，把这个地点设为默认地点
	//如果设置了，不做任何操作
	old_default_place, err := s_u.GetLastDefaultPlace()
	if err != nil {
		util.Debug("Cannot get last default place", err)
		Report(w, r, "你好，茶博士表示无法收藏地方，请稍后再试。")
		return
	}
	if old_default_place.Id == 0 {
		//这是茶棚占位地点，还没有设置用户默认地点
		udp := data.UserDefaultPlace{
			UserId:  s_u.Id,
			PlaceId: user_place.Id,
		}
		if err = udp.Create(); err != nil {
			util.Debug("Cannot create user default place", err)
			Report(w, r, "你好，茶博士表示收藏地方失误，请稍后再试。")
			return
		}

	}

	//重定向到用户地点本
	http.Redirect(w, r, "/v1/place/my", http.StatusFound)

}

// GET /v1/place/new
// 向用户返回创建新地方表单页面
func NewPlace(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var pL data.PlaceSlice
	pL.SessUser = s_u
	RenderHTML(w, &pL, "layout", "navbar.private", "place.new")

}

// POST /v1/place/new
// 处理用户创建新地方的请求
func CreatePlace(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
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
	//限制用户登记的地方最大数量为99,防止暴表
	if count_place, err := data.CountPlaceByUserId(s_u.Id); err != nil || count_place >= 99 {
		util.Debug("Cannot get user place count", err)
		Report(w, r, "你好，茶博士表示您已经提交了多得数不过来，就要爆表的地方，请确定后再试。")
		return
	}
	err = r.ParseForm()
	// Check form data
	if err != nil {
		util.Debug(s.Email, "Cannot parse form data")
		//http.Redirect(w, r, "/v1/goods/new", http.StatusFound)
		Report(w, r, "一脸蒙的茶博士，表示看不懂你提交的物资资料，请确认后再试一次。")
		return
	}
	category_str := r.PostFormValue("category")
	category_int, err := strconv.Atoi(category_str)
	if err != nil {
		util.Debug(" Cannot convert class to int", err)
		Report(w, r, "你好，茶博士表示无法理解地方的类型，请稍后再试。")
		return
	}
	// check category 参数是否合法
	switch category_int {
	case 1, 0:
		break
	default:
		Report(w, r, "你好，茶博士表示无法理解地方的类型，请确认后再试。")
		return
	}

	le := len(r.PostFormValue("name"))
	// Check name length
	if le < 2 || le > 50 {
		Report(w, r, "你好，茶博士表示地方名称长度不合法，请确认后再试。")
		return
	}
	name := r.PostFormValue("name")
	le = CnStrLen(r.PostFormValue("nickname"))
	// Check nickname length
	if le < 2 || le > 50 {
		Report(w, r, "你好，茶博士表示地方昵称长度不合法，请确认后再试。")
		return
	}
	nickname := r.PostFormValue("nickname")
	le = CnStrLen(r.PostFormValue("description"))
	// Check description length
	if le < 2 || le > 500 {
		Report(w, r, "你好，茶博士表示地方描述长度不合法，请确认后再试。")
		return
	}
	description := r.PostFormValue("description")
	is_public := r.PostFormValue("is_public") == "0"
	place := data.Place{
		Name:        name,
		Nickname:    nickname,
		UserId:      s_u.Id,
		Description: description,
		Category:    category_int,
		IsPublic:    is_public,
		Icon:        "bootstrap-icons/bank.svg",
	}
	if err = place.Create(); err != nil {
		util.Debug("Cannot create place", err)
		Report(w, r, "你好，茶博士居然说墨水用完了无法记录新地方，请确认后再试。")
		return
	}
	//把这个地方绑定到当前用户名下
	up := data.UserPlace{
		UserId:  s_u.Id,
		PlaceId: place.Id,
	}
	if err = up.Create(); err != nil {
		util.Debug("cannot create user-place", err)
		Report(w, r, "你好，茶博士正在飞速为你写字服务中，请确认后再试。")
		return
	}
	// 统计用户绑定的地方数量，如果是 == 1，那么就把这个地方设置为该用户的默认地方
	place_count := up.Count()
	if place_count == 1 {
		udp := data.UserDefaultPlace{
			UserId:  s_u.Id,
			PlaceId: place.Id,
		}
		if err = udp.Create(); err != nil {
			util.Debug("cannot create user default place", err)
			Report(w, r, "你好，娇羞默默同谁诉，倦倚西风夜已昏。稍后再试。")
			return
		}
	}

	http.Redirect(w, r, "/v1/place/my", http.StatusFound)
}

// GET /v1/place/my
// 读取当前用户的全部绑定（收集）地方
func MyPlace(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var pL data.PlaceSlice
	places, err := s_u.GetAllBindPlaces()
	if err != nil {
		util.Debug("Cannot get places from user", err)
		Report(w, r, "你好，茶博士表示无法获取您收集的地方，请稍后再试。")
		return
	}
	pL.PlaceSlice = places
	pL.SessUser = s_u
	RenderHTML(w, &pL, "layout", "navbar.private", "places.my", "place.media-object")
}

// GET  /v1/place/detail?id=
func PlaceDetail(w http.ResponseWriter, r *http.Request) {
	//获取地方的uuid
	r.ParseForm()
	place_uuid := r.FormValue("id")

	//如果uuid为茶棚系统值"x",这是一个占位值，跳转首页
	if place_uuid == data.PlaceUuidSpaceshipTeabar {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	t_place := data.Place{
		Uuid: place_uuid,
	}

	if err := t_place.GetByUuid(); err != nil {
		util.Debug("Cannot get place by uuid", err)
		Report(w, r, "你好，������表示无法获取您要查看的地方，请稍后再试。")
		return
	}

	s, err := Session(r)
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

	var pD data.PlaceDetail
	pD.Place = t_place
	pD.SessUser = s_u
	//是否地方登记者
	if t_place.UserId == s_u.Id {
		pD.IsAuthor = true
	} else {
		pD.IsAuthor = false
	}
	RenderHTML(w, &pD, "layout", "navbar.private", "place.detail", "place.media-object")
}
