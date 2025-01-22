package route

import (
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handle() /v1/family_member/sign_in
// 处理&家庭茶团的登记新成员窗口
// 根据提交的某个茶友邮箱地址，将其申报为&家庭茶团成员
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

// GET /v1/family_member/sign_in?id=xxx
// 给用户返回一张空白的&家庭茶团新成员登记表格（页面）
func FamilyMemberSignInGet(w http.ResponseWriter, r *http.Request) {
	//读取会话资料
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	family_member_user_uuid := r.URL.Query().Get("id")
	if family_member_user_uuid == "" {
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 欲声明为家庭成员的茶友资料
	family_member_user, err := data.GetUserByUUID(family_member_user_uuid)
	if err != nil {
		util.Info(err, "cannot get family by uuid")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}

	//读取当前用户的相关资料
	s_u, s_d_family, s_all_families, s_d_team, s_survival_teams, s_d_place, s_places, err := FetchUserRelatedData(s)
	if err != nil {
		util.Danger(err, "cannot fetch s_u s_teams given session")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	var fms data.FamilyMemberSignInPageData
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

	fms.FamilyMemberUser = family_member_user

	//渲染页面
	RenderHTML(w, &fms, "layout", "navbar.private", "family_member.sign_in")

}

// POST /v1/family_member/sign_in
// 处理增加&家庭茶团成员声明的提交事务
func FamilyMemberSignInPost(w http.ResponseWriter, r *http.Request) {
	// 获取session
	s, err := Session(r)
	if err != nil {
		util.Danger(err, " Cannot get session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Warning(err, "Cannot get user from session")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 解析表单内容，获取当前用户提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Danger(err, " Cannot parse form")
		Report(w, r, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}

	m_email := r.PostFormValue("m_email")
	// 检查提交的成员邮箱
	if ok := IsEmail(m_email); !ok {
		Report(w, r, "你好，涨红了脸的茶博士，竟然强词夺理说，电子邮箱格式太复杂看不懂，请确认后再试一次。")
		return
	}
	//读取声明退出的成员资料
	t_user, err := data.GetUserByEmail(m_email)
	if err != nil {
		util.Warning(err, m_email, "Cannot get user by email")
		Report(w, r, "你好，茶博士正在无事忙之中，稍后再试。")
		return
	}
	// 读取提及的家庭资料
	// 提及的家庭
	family_uuid := r.PostFormValue("family_uuid")
	t_family := data.Family{
		Uuid: family_uuid,
	}
	// 检查提及的家庭是否存在
	if err = t_family.GetByUuid(); err != nil {
		util.Warning(err, t_family.Uuid, "Cannot get family by uuid")
		Report(w, r, "你好，茶博士找不到提及的家庭资料，请确认后再试。")
		return
	}

	// 声明标题
	title := "关于" + t_family.Name + "家庭茶团增加新成员的声明"

	// 提交的声明内容
	cont := r.PostFormValue("content")
	// 检查提交的声明内容字数是否>3 and <456
	lenCont := CnStrLen(cont)
	if lenCont < 3 || lenCont > 456 {
		Report(w, r, "你好，茶博士认为内容字数太长或者太短，请确认后再试。")
		return
	}

	ok := false
	// 检查提及的茶友是否已经是提及的家庭的成员
	if ok, err = t_family.IsMember(t_user.Id); ok || err != nil {
		Report(w, r, "你好，茶博士认为提及的茶友已经是家庭的成员，请确认后再试。")
		util.Warning(err, t_user.Id, "Cannot check if user is member of family")
		return
	}

	// check if user is member of family
	if ok, err = t_family.IsMember(s_u.Id); err != nil || !ok {
		Report(w, r, "你好，茶博士认为你不是这个家庭的成员，请确认后再试。")
		util.Warning(err, s_u.Id, "Cannot check if user is member of family")
		return
	}

	// 检查当前用户是否这个家庭的父母角色
	parent_members, err := t_family.ParentMembers()
	if err != nil {
		util.Warning(err, t_family.Id, "Cannot get parent members of family")
		Report(w, r, "你好，茶博士认为你不是这个家庭的成员，请确认后再试。")
		return
	}
	for _, p := range parent_members {
		if p.UserId == s_u.Id {
			ok = true
			break
		}
	}
	if !ok {
		Report(w, r, "你好，茶博士认为你无权声明这个家庭增加新成员，请确认后再试。")
		return
	}

	//读取提及的place资料
	place_uuid := r.PostFormValue("place_uuid")
	t_place := data.Place{
		Uuid: place_uuid,
	}
	// 检查提及的品茶地点是否存在
	if err = t_place.GetByUuid(); err != nil {
		util.Warning(err, t_place.Uuid, "Cannot get place by uuid")
		Report(w, r, "你好，茶博士找不到提及的地点，请确认后再试。")
		return
	}

	//读取提交的角色
	role_str := r.PostFormValue("role")
	// 检查提交的角色是否存在
	if role_str == "" {
		Report(w, r, "你好，茶博士认为你没有选择角色，请确认后再试。")
		return
	}
	role_int, err := strconv.Atoi(role_str)
	if err != nil {
		Report(w, r, "你好，茶博士处理选择的角色出现了问题，请稍后再试。")
		return
	}
	if role_int < 0 || role_int > 5 {
		Report(w, r, "你好，茶博士认为你选择的角色不存在，请确认后再试。")
		return
	}

	//检查这个角色是否被占用
	t_family_member := data.FamilyMember{
		Role:     role_int,
		FamilyId: t_family.Id,
	}
	//查看成员角色，分类处理：0、秘密，1、男主人，2、女主人，3、女儿， 4、儿子，5、宠物,
	switch role_int {
	case 0, 3, 4, 5:
		// ok，角色可以共用
		break
	case 1, 2:
		//角色是唯一的的，检查是否被占用
		if err = t_family_member.GetByRoleFamilyId(); err == nil {
			Report(w, r, "你好，茶博士认为你选择的角色已经被占用，请确认后再试。")
			return
		}
	default:
		Report(w, r, "你好，茶博士认为你选择的角色不存在，请确认后再试。")
	}

	// 提交的是否为成年人参数
	is_adult_str := r.PostFormValue("is_adult")
	if is_adult_str == "" {
		Report(w, r, "你好，茶博士认为你没有选择是否为成年人，请确认后再试。")
		return
	}
	// 检查提交的是否为成年人参数是否合法
	is_adult, err := strconv.ParseBool(is_adult_str)
	if err != nil {
		Report(w, r, "你好，茶博士认为你选择的是否为成年人不合法，请确认后再试。")
		return
	}
	// 检查是否为成年人，如果不是成年人，检查是否已经有成年人

	// 读取提交的是否领养参数
	is_adopted_str := r.PostFormValue("is_adopted")
	// 检查提交的是否领养参数是否合法
	is_adopted, err := strconv.ParseBool(is_adopted_str)
	if err != nil {
		Report(w, r, "你好，茶博士看不懂你声明的成员是否领养情况，请确认后再试。")
		return
	}
	// 新声明
	t_family_member_sign_in := data.FamilyMemberSignIn{
		FamilyId:  t_family.Id,
		UserId:    t_user.Id,
		Role:      role_int,
		IsAdult:   is_adult,
		Title:     title,
		Content:   cont,
		PlaceId:   t_place.Id,
		IsAdopted: is_adopted,
	}
	//检查是否已经存在重复的声明
	if err = t_family_member_sign_in.GetByFamilyIdMemberUserId(); err == nil {
		Report(w, r, "你好，茶博士认为你已经提交过这个声明，请确认后再试。")
		return
	}

	// 保存新声明
	if err = t_family_member_sign_in.Create(); err != nil {
		util.Warning(err, "Cannot create family member sign in")
		Report(w, r, "你好，满头大汗的茶博士说，因为眼镜太模糊导致增加成员的声明保存失败，请确认后再试。")
		return
	}

	//报告声明保存成功
	report := fmt.Sprintf("你好，%s 已经保存成功。请自行联系你的家人，查找访问你的家庭详情，阅读声明并确认后生效。", title)
	Report(w, r, report)

}
