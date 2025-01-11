package route

import (
	"net/http"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handle() /v1/family_member/sign_in
// 处理家庭&茶团的登记新成员窗口
// 根据提交的某个茶友邮箱地址，将其申报为家庭&茶团成员
func HandleFamilyMemberSignIn(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		FamilyMemberSignInGet(w, r)
	case http.MethodPost:
		FamilyMemberSignInPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 给用户返回一张空白的家庭&茶团新成员登记表格（页面）
func FamilyMemberSignInGet(w http.ResponseWriter, r *http.Request) {
	//读取会话资料
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	//读取当前用户的相关资料
	s_u, s_d_family, s_all_families, s_d_team, s_survival_teams, s_d_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		util.Danger(err, "cannot fetch s_u s_teams given session")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	var fms data.FamilyMemberSignIn
	//将当前用户的资料填入表格
	fms.SessUser = s_u
	//将当前用户的默认茶团资料填入表格
	fms.SessUserDefaultFamily = s_d_family
	fms.SessUserAllFamilies = s_all_families
	//将当前用户的默认茶团资料填入表格
	fms.SessUserDefaultTeam = s_d_team
	//将当前用户的所有茶团资料填入表格
	fms.SessUserSurvivalTeams = s_survival_teams
	fms.SessUserDefaultPlace = s_d_place
	//将当前用户的所有地点资料填入表格
	fms.SessUserBindPlaces = s_places

	//渲染页面
	RenderHTML(w, &fms, "layout", "navbar.public", "family_member.sign_in")

}

func FamilyMemberSignInPost(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
