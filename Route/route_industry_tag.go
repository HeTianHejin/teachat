package route

import (
	"encoding/json"
	"net/http"
	"strconv"

	dao "teachat/DAO"
	util "teachat/Util"
)

// industryTagNode 级联下拉用的行业标签节点
type industryTagNode struct {
	Id          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	LevelName   string `json:"level_name"`
	HasChildren bool   `json:"has_children"`
}

// IndustryTagChildren GET /v1/team/industry_children?parent=<id>
// 返回指定行业标签节点下的直接子节点（JSON），供"门类→大类→中类→小类"单页级联下拉使用。
// parent 缺省或 0 时返回全部门类（level=1）。
// has_children=false 表示已到叶子，前端可停止下钻。
func IndustryTagChildren(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}
	if _, err := s.User(); err != nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}

	parentID := 0
	if v := r.URL.Query().Get("parent"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			parentID = p
		}
	}

	children, err := dao.GetIndustryTagChildren(parentID)
	if err != nil {
		util.Debug("Cannot get industry tag children %v", err)
		http.Error(w, "获取行业分类失败", http.StatusInternalServerError)
		return
	}

	nodes := make([]industryTagNode, 0, len(children))
	for _, c := range children {
		count, err := dao.CountIndustryTagChildren(c.Id)
		if err != nil {
			util.Debug("Cannot count industry tag children %v", err)
		}
		nodes = append(nodes, industryTagNode{
			Id:          c.Id,
			Code:        c.Code,
			Name:        c.Name,
			Level:       c.Level,
			LevelName:   dao.IndustryTagLevelName(c.Level),
			HasChildren: count > 0,
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		util.Debug("Cannot encode industry tag children %v", err)
	}
}
