package route

import (
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/group/new
func NewGroup(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	u, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var tpd data.TeamSquare
	tpd.SessUser = u
	GenerateHTML(w, &tpd, "layout", "navbar.private", "group.new")
}

// POST /v1/group/create
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	su, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	name := r.PostFormValue("name")
	le := CnStrLen(name)
	if le < 4 || le > 24 {
		Report(w, r, "你好，茶博士摸摸头，竟然说集团名字字数太多或者太少，未能创建新集团。")
		return
	}
	_, err = data.GetGroupByName(name)
	if err == nil {
		//重名
		Report(w, r, "你好，茶博士摸摸头，竟然说集团名称重复，未能创建新集团。")
		return
	}
	abbr := r.PostFormValue("abbreviation")
	// 简称是否在4-6中文字符
	lenA := CnStrLen(abbr)
	if lenA < 4 || lenA > 6 {
		Report(w, r, "你好，茶博士摸摸头，竟然说队名简称字数太多或者太少，未能创建新集团。")
		return
	}
	_, err = data.GetGroupByAbbreviation(abbr)
	if err == nil {
		//重名
		Report(w, r, "你好，茶博士摸摸头，竟然说集团简称重复，未能创建新集团。")
		return
	}

	mission := r.PostFormValue("mission")
	// 检测mission是否在17-456中文字符
	lenM := CnStrLen(mission)
	if lenM < 17 || lenM > 456 {
		Report(w, r, "你好，茶博士摸摸头，竟然说愿景字数太多或者太少，未能创建新集团。")
		return
	}
	class, err := strconv.Atoi(r.PostFormValue("class"))
	if err != nil {
		util.Info(err, " Cannot convert class to int")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
		return
	}
	//检测class是否合规
	switch class {
	case 10, 20:
		break
	default:
		Report(w, r, "你好，茶博士摸摸头，竟然说集团类别太多或者太少，未能创建新集团。")
		return
	}
	team, err := su.GetLastDefaultTeam()
	if err != nil {
		util.Info(err, "Cannot get last default team")
		Report(w, r, "你好，茶博士彬彬有礼的说，找不到您的默认团队信息，请稍后再试。")
		return
	}

	g := data.Group{
		Name:         name,
		Abbreviation: abbr,
		Mission:      mission,
		Class:        class,
		Logo:         "groupLogo",
		FounderId:    su.Id,
		FirstTeamId:  team.Id,
	}
	err = g.Create()
	if err != nil {
		util.Warning(err, "Cannot create group")
		Report(w, r, "你好，茶博士因为的笔没有墨水了，未能创建新集团，请稍后再试。")
		return
	}
	// 创建一条友邻盲评,是否接纳 新集团的记录
	aO := data.AcceptObject{
		ObjectId:   g.Id,
		ObjectType: 6,
	}
	if err = aO.Create(); err != nil {
		util.Warning(err, "Cannot create accept_object")
		Report(w, r, "你好，茶博士失魂鱼，未能创建新集团，请稍后再试。")
	}

	// 发送盲评请求消息给两个在线用户
	//构造消息
	mess := data.AcceptMessage{
		FromUserId:     1,
		Title:          "新茶语邻座评审邀请",
		Content:        "您被茶棚选中为新茶语评审官啦，请及时审理新茶。",
		AcceptObjectId: aO.Id,
	}
	//发送消息
	if err = AcceptMessageSendExceptUserId(su.Id, mess); err != nil {
		Report(w, r, "你好，茶博士迷路了，未能发送盲评请求消息。")
		return
	}
	t := fmt.Sprintf("你好，新集团 %s 已准备妥当，稍等有缘茶友评审通过之后，即行昭告天下。", g.Abbreviation)
	// 提示用户草稿保存成功
	Report(w, r, t)
}

// Get /v1/group/detail
func GroupDetail(w http.ResponseWriter, r *http.Request) {

	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	su, err := s.User()
	if err != nil {
		util.Info(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	var gd data.GroupDetail

	vals := r.URL.Query()
	uuid := vals.Get("id")
	//获取group
	g, err := data.GetGroupByUuid(uuid)
	if err != nil {
		Report(w, r, "你好，茶博士摸摸头，居然说集团可能不存在，未能查看集团详情。")
		return
	}
	// 读取集团资料荚
	gb, err := GetGroupBean(g)
	if err != nil {
		Report(w, r, "你好，茶博士摸摸头，居然说集团可能不存在，未能查看集团详情。")
		return
	}
	// 读取全部下属茶团资料
	team_list, err := data.GetTeamsByGroupId(g.Id)
	if err != nil {
		Report(w, r, "你好，茶博士摸摸头，居然说集团可能存在黑洞，未能查看集团详情。")
		return
	}
	// 读取第一管理团队
	f_team, _ := data.GetTeamById(g.FirstTeamId)
	gd.FirstTeamBean, err = GetTeamBean(f_team)
	if err != nil {
		Report(w, r, "你好，����������头，��然说集��可能存在黑��，未能查看集��详情。")
		return
	}

	//remove f_team from team_list，余下的是普通团队,移除，删除 重复的team
	for i, team := range team_list {
		if team.Id == f_team.Id {
			team_list = append(team_list[:i], team_list[i+1:]...)
			break
		}
	}

	gd.TeamBeanList, err = GetTeamBeanList(team_list)
	if err != nil {
		Report(w, r, "你好，茶博士摸摸头，居然说集团可能不存在，未能查看集团详情。")
		return
	}

	gd.SessUser = su
	gd.GroupBean = gb
	// 是否超过12个
	len := len(team_list)
	if len > 12 {
		gd.IsOverTwelve = true
	} else {
		gd.IsOverTwelve = false
	}

	GenerateHTML(w, &gd, "layout", "navbar.private", "group.detail", "teams.public")
}
