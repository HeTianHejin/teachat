package route

import (
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

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
		util.Info(err, s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var pL data.PlaceList
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
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//限制用户登记的地方最大数量为99,防止暴表
	if count_place, err := data.CountPlaceByUserId(s_u.Id); err != nil || count_place >= 99 {
		util.Warning(err, "Cannot get user place count")
		Report(w, r, "你好，闪电考拉表示您已经提交了多得数不过来，就要爆表的地方，请确定后再试。")
		return
	}
	r.ParseForm()
	category_str := r.PostFormValue("category")
	category_int, err := strconv.Atoi(category_str)
	if err != nil {
		util.Warning(err, " Cannot convert class to int")
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
	name := r.PostFormValue("name")
	nickname := r.PostFormValue("nickname")
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
		util.Danger(err, "Cannot create place")
		Report(w, r, "你好，茶博士居然说墨水用完了无法记录新地方，请确认后再试。")
		return
	}
	//把这个地方绑定到当前用户名下
	up := data.UserPlace{
		UserId:  s_u.Id,
		PlaceId: place.Id,
	}
	if err = up.Create(); err != nil {
		util.Danger(err, "cannot create user-place")
		Report(w, r, "你好，闪电考拉正在飞速为你写字服务中，请确认后再试。")
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
			util.Danger(err, "cannot create user default place")
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
		util.Info(err, s.Email, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var pL data.PlaceList
	places, err := s_u.GetAllBindPlaces()
	if err != nil {
		util.Warning(err, "Cannot get places from user")
		Report(w, r, "你好，茶博士表示无法获取您收集的地方，请稍后再试。")
		return
	}
	pL.PlaceList = places
	pL.SessUser = s_u
	RenderHTML(w, &pL, "layout", "navbar.private", "places.my")
}

// GET  /v1/place/detail?id=
func PlaceDetail(w http.ResponseWriter, r *http.Request) {
	//获取地方的uuid
	r.ParseForm()
	place_uuid := r.FormValue("id")
	t_place := data.Place{
		Uuid: place_uuid,
	}

	if err := t_place.GetByUuid(); err != nil {
		util.Warning(err, "Cannot get place by uuid")
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
		util.Info(err, "Cannot get user from session")
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
	RenderHTML(w, &pD, "layout", "navbar.private", "place.detail")
}
