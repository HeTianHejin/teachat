package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// GET /v1/search/by_tag?tag=xxx&type=team|group
// 根据标签搜索团队或集团
func SearchByTag(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	searchType := r.URL.Query().Get("type")

	if tag == "" {
		report(w, data.UserUnknown, "你好，请输入搜索标签。")
		return
	}

	// 获取会话用户
	var sessUser data.User
	s, err := session(r)
	if err != nil {
		sessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
		}
	} else {
		sessUser, _ = s.User()
	}

	var pageData struct {
		SessUser       data.User
		Tag            string
		SearchType     string
		TeamBeanSlice  []data.TeamBean
		GroupBeanSlice []data.GroupBean
	}
	pageData.SessUser = sessUser
	pageData.Tag = tag
	pageData.SearchType = searchType

	// 根据类型搜索
	if searchType == "group" {
		// 搜索集团
		groups, err := data.SearchGroupsByTag(tag)
		if err != nil {
			util.Debug("Cannot search groups by tag", err)
		}
		// 转换为GroupBean
		groupBeans := make([]data.GroupBean, 0, len(groups))
		for _, g := range groups {
			founder, _ := data.GetUser(g.FounderId)
			groupBeans = append(groupBeans, data.GroupBean{
				Group:         g,
				CreatedAtDate: g.CreatedAtDate(),
				Open:          g.Class == data.GroupClassOpen,
				Founder:       founder,
			})
		}
		pageData.GroupBeanSlice = groupBeans
	} else {
		// 默认搜索团队
		teams, err := data.SearchTeamsByTag(tag)
		if err != nil {
			util.Debug("Cannot search teams by tag", err)
		}
		// 转换为TeamBean
		teamBeans, err := fetchTeamBeanSlice(teams)
		if err != nil {
			util.Debug("Cannot fetch team bean slice", err)
		}
		pageData.TeamBeanSlice = teamBeans
	}

	// 渲染页面
	if sessUser.Id == data.UserId_None {
		generateHTML(w, &pageData, "layout", "navbar.public", "search.tag_result")
	} else {
		generateHTML(w, &pageData, "layout", "navbar.private", "search.tag_result")
	}
}

// GET /v1/tags/hot
// 显示热门标签页面
func HotTags(w http.ResponseWriter, r *http.Request) {
	// 获取会话用户
	var sessUser data.User
	s, err := session(r)
	if err != nil {
		sessUser = data.User{
			Id:   data.UserId_None,
			Name: "游客",
		}
	} else {
		sessUser, _ = s.User()
	}

	// 预定义热门标签
	hotTags := []string{
		"诗词书法", "家电维修", "软件开发", "绘画摄影",
		"音乐舞蹈", "电脑维护", "语言培训", "法律咨询",
		"中医养生", "电子商务", "装修设计", "财务会计",
	}

	var pageData struct {
		SessUser data.User
		HotTags  []string
	}
	pageData.SessUser = sessUser
	pageData.HotTags = hotTags

	// 渲染页面
	if sessUser.Id == data.UserId_None {
		generateHTML(w, &pageData, "layout", "navbar.public", "tags.hot")
	} else {
		generateHTML(w, &pageData, "layout", "navbar.private", "tags.hot")
	}
}
