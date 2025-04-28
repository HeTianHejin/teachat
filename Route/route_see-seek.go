package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handler /v1/see-seek/new
func HandleNewSeeSeek(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		SeeSeekNewGet(w, r)
	case http.MethodPost:
		SeeSeekNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func SeeSeekNewPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// GET /v1/see-seek/new?id=xxx
func SeeSeekNewGet(w http.ResponseWriter, r *http.Request) {
	var err error
	vals := r.URL.Query()
	uuid := vals.Get("id")
	t_post := data.Post{Uuid: uuid}
	if err = t_post.GetByUuid(); err != nil {
		util.Error(" Cannot get post detail", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	sess, err := Session(r)
	if err != nil {
		util.Error(" Cannot get session", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	s_u, err := sess.User()
	if err != nil {
		util.Error(" Cannot get user from session", err)
		Report(w, r, "你好，茶博士失魂鱼，有眼不识泰山。")
		return
	}
	t_thread, err := t_post.Thread()
	if err != nil {
		util.Error(" Cannot get thread from post", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	master_team, err := data.GetTeam(t_thread.TeamId)
	if err != nil {
		util.Error(" Cannot get master team", err)
		Report(w, r, "你好，茶博士失魂鱼，未能读取专属资料。")
		return
	}
	is_member, err := master_team.IsMember(s_u.Id)
	if err != nil {
		util.Error(" Cannot check master-team-member given team_id,s_u.Email", master_team.Id, s_u.Email)
		Report(w, r, "你好，茶博士失魂鱼，未能读取茶议团队会员资格资料。")
		return
	}
	// if not member, return
	if !is_member {
		Report(w, r, "你好，茶博士表示惊讶，成员资格检查未通过。")
		return
	}

}
