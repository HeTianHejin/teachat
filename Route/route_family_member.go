package route

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// Handle() /v1/family_member/sign_in_new
// 处理&家庭茶团的登记新成员窗口
// 根据提交的某个茶友邮箱地址，将其申报为&家庭茶团成员
func HandleFamilyMemberSignInNew(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		FamilyMemberSignInNewGet(w, r)
	case http.MethodPost:
		FamilyMemberSignInNewPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/family_member/sign_in_new?id=xxx
// 给用户返回一张空白的&家庭茶团新成员登记表格（页面）
func FamilyMemberSignInNewGet(w http.ResponseWriter, r *http.Request) {
	//读取会话资料
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", s.Email, err)
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	family_member_user_uuid := r.URL.Query().Get("id")
	if family_member_user_uuid == "" {
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 欲声明为家庭成员的茶友资料
	family_member_user, err := data.GetUserByUUID(family_member_user_uuid)
	if err != nil {
		util.Debug("cannot get family by uuid", err)
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	//发声明家庭
	family_uuid := r.URL.Query().Get("family_uuid")
	if family_uuid == "" {
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	family := data.Family{Uuid: family_uuid}
	if err = family.GetByUuid(); err != nil {
		util.Debug("cannot get family by uuid:", family_uuid, err)
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}

	var fms data.FamilyMemberSignInNew
	//将当前用户的资料填入表格
	fms.SessUser = s_u
	//将当前用户的默认茶团资料填入表格
	fms.Family = family
	fms.NewMemberUser = family_member_user

	//渲染页面
	generateHTML(w, &fms, "layout", "navbar.private", "family_member.sign_in")

}

// POST /v1/family_member/sign_in_new
// 处理增加&家庭茶团成员声明的提交事务
func FamilyMemberSignInNewPost(w http.ResponseWriter, r *http.Request) {
	// 获取session
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug("Cannot get user from session", err)
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}
	// 解析表单内容，获取当前用户提交的内容
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，柳丝榆荚自芳菲，不管桃飘与李飞。请稍后再试。")
		return
	}

	m_email := r.PostFormValue("m_email")
	if m_email == "" {
		report(w, s_u, "你好，茶博士认为你没有填写茶友的电子邮箱，请确认后再试。")
		return
	}
	// 检查提交的成员邮箱
	if ok := isEmail(m_email); !ok {
		report(w, s_u, "你好，涨红了脸的茶博士，竟然强词夺理说，电子邮箱格式太复杂看不懂，请确认后再试一次。")
		return
	}
	//读取声明增加的成员资料
	t_user, err := data.GetUserByEmail(m_email, r.Context())
	if err != nil {
		util.Debug(m_email, "Cannot get user by email", err)
		report(w, s_u, "你好，茶博士正在无事忙之中，稍后再试。")
		return
	}
	// 读取提及的家庭资料
	family_uuid := r.PostFormValue("family_uuid")

	// 如果family_uuid=“x“特殊值，这是虚值，报告错误
	if family_uuid == data.FamilyUuidUnknown || family_uuid == "" {
		report(w, s_u, "你好，茶博士认为你没有提及具体的家庭，或者提及的&家庭茶团还没有登记，请确认后再试。")
		return
	}

	t_family := data.Family{
		Uuid: family_uuid,
	}
	// 检查提及的家庭是否存在
	if err = t_family.GetByUuid(); err != nil {
		//util.PanicTea(util.LogError(err), t_family.Uuid, "Cannot get family by uuid")
		report(w, s_u, "你好，茶博士找不到提及的家庭资料，请确认后再试。")
		return
	}

	// 声明标题
	title := "关于" + t_family.Name + "家庭茶团增加新成员的声明"

	// 提交的声明内容
	cont := r.PostFormValue("content")
	// 检查提交的声明内容字数是否>threadMinWord and <int(util.Config.ThreadMaxWord)
	lenCont := cnStrLen(cont)
	if lenCont < int(util.Config.ThreadMinWord) || lenCont > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "你好，茶博士认为内容字数太长或者太短，请确认后再试。")
		return
	}

	// check if session user is parent member of family
	if isPMember, err := t_family.IsParentMember(s_u.Id); err != nil || !isPMember {
		util.Debug(s_u.Id, "Cannot check if user is member of family", err)
		report(w, s_u, "你好，茶博士认为你无权声明这个家庭增加新成员，请确认后再试。")
		return
	}
	isMember := false
	// 检查提及的茶友是否已经是提及的家庭的成员
	if isMember, err = t_family.IsMember(t_user.Id); isMember || err != nil {
		util.Debug(t_user.Id, "Cannot check if user is member of family", err)
		report(w, s_u, "你好，茶博士认为提及的茶友已经是家庭的成员，请勿重复添加。")
		return
	}

	//读取提交的角色
	role_str := r.PostFormValue("role")
	// 检查提交的角色是否合法
	if role_str == "" {
		report(w, s_u, "你好，茶博士认为你没有选择角色，请确认后再试。")
		return
	}
	role_int, err := strconv.Atoi(role_str)
	if err != nil {
		report(w, s_u, "你好，茶博士处理选择的角色出现了问题，请稍后再试。")
		return
	}
	if role_int < data.FamilyMemberRoleUnknown || role_int > data.FamilyMemberRolePet {
		report(w, s_u, "你好，茶博士认为你选择的角色不存在，请确认后再试。")
		return
	}

	//检查这个角色是否被占用
	t_family_member := data.FamilyMember{
		Role:     role_int,
		FamilyId: t_family.Id,
	}
	//查看成员角色，分类处理：0、秘密，1、男主人，2、女主人，3、女儿， 4、儿子，5、宠物,
	switch role_int {
	case data.FamilyMemberRoleUnknown, data.FamilyMemberRoleDaughter, data.FamilyMemberRoleSon, data.FamilyMemberRolePet:
		// ok，角色可以共用
		break
	case data.FamilyMemberRoleHusband, data.FamilyMemberRoleWife:
		//角色是唯一的，检查是否被占用
		if err = t_family_member.GetByRoleFamilyId(); err == nil {
			report(w, s_u, "你好，茶博士认为你选择的角色已经被占用，请确认后再试。")
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			break
		} else {
			util.Debug(t_family_member.Id, "Cannot get family member by role and family id", err)
			report(w, s_u, "你好，茶博士处理选择的角色出现了问题，请稍后再试。")
			return
		}
	default:
		report(w, s_u, "你好，茶博士认为你选择的角色不存在，请确认后再试。")
		return
	}

	// 提交的是否为成年人参数
	is_adult_str := r.PostFormValue("is_adult")
	if is_adult_str == "" {
		report(w, s_u, "你好，茶博士认为你没有选择是否为成年人，请确认后再试。")
		return
	}
	// 检查提交的是否为成年人参数是否合法
	is_adult, err := strconv.ParseBool(is_adult_str)
	if err != nil {
		report(w, s_u, "你好，茶博士认为你选择的是否为成年人不合法，请确认后再试。")
		return
	}
	// 检查是否为成年人，如果不是成年人，检查是否已经有成年人

	// 读取提交的是否领养参数
	is_adopted_str := r.PostFormValue("is_adopted")
	// 检查提交的是否领养参数是否合法
	is_adopted, err := strconv.ParseBool(is_adopted_str)
	if err != nil {
		report(w, s_u, "你好，茶博士看不懂你声明的成员是否领养情况，请确认后再试。")
		return
	}
	// 新声明
	new_family_member_sign_in := data.FamilyMemberSignIn{
		FamilyId:     t_family.Id,
		UserId:       t_user.Id,
		Role:         role_int,
		IsAdult:      is_adult,
		Title:        title,
		Content:      cont,
		PlaceId:      data.PlaceIdSpaceshipTeabar,
		IsAdopted:    is_adopted,
		AuthorUserId: s_u.Id,
	}
	//检查是否已经存在重复的声明
	if err = new_family_member_sign_in.GetByFamilyIdMemberUserId(); err == nil {
		report(w, s_u, "你好，茶博士认为你已经提交过这个声明，请确认后再试。")
		return
	}

	// 保存新声明
	if err = new_family_member_sign_in.Create(); err != nil {
		util.Debug("Cannot create family member sign in", err)
		report(w, s_u, "你好，满头大汗的茶博士说，因为眼镜太模糊导致增加成员的声明保存失败，请确认后再试。")
		return
	}

	//报告声明保存成功
	rt := fmt.Sprintf("你好，%s 已经保存成功。请自行联系你的家人，查找访问你的家庭详情，阅读声明并确认后生效。", title)
	report(w, s_u, rt)

}

// Handle() /v1/family_member/sign_in
// 处理&家庭茶团的声明增加成员窗口
// 根据答复结果，来决定是否将其添加为&家庭茶团成员
func HandleFamilyMemberSignIn(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		FamilyMemberSignInRead(w, r)
	case http.MethodPost:
		FamilyMemberSignInReply(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 为声明提及新成员办理取阅声明书，
// GET /v1/family_member/sign_in?id=
func FamilyMemberSignInRead(w http.ResponseWriter, r *http.Request) {
	// 获取session
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Debug(" Cannot get user", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 获取请求参数
	family_member_sign_in_uuid := r.URL.Query().Get("id")
	// 读取增加家庭成员声明资料
	family_member_sign_in := data.FamilyMemberSignIn{
		Uuid: family_member_sign_in_uuid,
	}
	if err := family_member_sign_in.GetByUuid(); err != nil {
		util.Debug(" Cannot get family_member_sign_in given uuid", err)
		report(w, s_u, "读取声明书失误，请稍后再试一次。")
		return
	}

	// 检查声明是否属于会话用户
	if family_member_sign_in.UserId != s_u.Id {
		report(w, s_u, "你好，柳丝榆荚自芳菲，声明资料满天飞。请稍后再试。")
		return
	}

	var fMSID data.FamilyMemberSignInDetail
	// 读取声明书详细资料
	family_member_sign_in_bean, err := fetchFamilyMemberSignInBean(family_member_sign_in)
	if err != nil {
		util.Debug(family_member_sign_in.Id, " Cannot get family_member_sign_in_bean")
		report(w, s_u, "读取声明书失误，请稍后再试一次。")
		return
	}
	//更新声明书状态为已读
	family_member_sign_in.Status = data.SignInStatusRead
	if err := family_member_sign_in.Update(); err != nil {
		util.Debug(" Cannot update family_member_sign_in", err)
		report(w, s_u, "更新声明书失误，请稍后再试一次。")
		return
	}

	//填写页面数据
	fMSID.SessUser = s_u
	fMSID.FamilyMemberSignInBean = family_member_sign_in_bean

	//渲染页面给用户
	generateHTML(w, &fMSID, "layout", "navbar.private", "family_member.sign_in_read")

}

// POST /v1/family_member/sign_in
// 答复家庭茶团成员声明
func FamilyMemberSignInReply(w http.ResponseWriter, r *http.Request) {
	// 获取session
	s, err := session(r)
	if err != nil {
		util.Debug(" Cannot get session", err)
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 根据会话信息读取茶友资料
	s_u, err := s.User()
	if err != nil {
		util.Debug(s.Email, "Cannot get user from session")
		report(w, s_u, "你好，满地梨花一片天，请稍后再试一次")
		return
	}

	//解析表单内容，获取茶友提交的参数
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
		return
	}
	// 检查提交的是否为成年人参数是否合法
	reply_str := r.PostFormValue("reply")
	reply_bool, err := strconv.ParseBool(reply_str)
	if err != nil {
		report(w, s_u, "你好，茶博士看不懂你选择的是否为家庭成员结果，请确认后再试。")
		return
	}
	//获取声明书id
	family_member_sign_in_uuid := r.PostFormValue("id")
	// 读取声明书资料
	family_member_sign_in := data.FamilyMemberSignIn{
		Uuid: family_member_sign_in_uuid,
	}
	if err = family_member_sign_in.GetByUuid(); err != nil {
		util.Debug(" Cannot get family_member_sign_in given uuid", err)
		report(w, s_u, "你好，茶博士正在忙碌中，厚厚的眼镜不见了，稍后再试。")
		return
	}
	// 检查声明是否属于会话用户
	if family_member_sign_in.UserId != s_u.Id {
		report(w, s_u, "你好，声明资料满天飞。各人自有各人家，请勿乱入别人家。")
		return
	}
	// 检查声明书状态是否已读但未处理，status==1是已读未处理，其它值都是非法的值
	if family_member_sign_in.Status != data.SignInStatusRead {
		report(w, s_u, "你好，柳丝榆荚自芳菲，声明资料满天飞。请稍后再试。")
		return
	}

	family_member_sign_in_reply := data.FamilyMemberSignInReply{
		SignInId: family_member_sign_in.Id,
		UserId:   s_u.Id,
	}

	//根据reply_bool值，true表示同意加入家庭，false表示拒绝加入家庭
	if reply_bool {
		//同意加入家庭
		//读取声明书资料
		family_member := data.FamilyMember{
			UserId:    family_member_sign_in.UserId,
			FamilyId:  family_member_sign_in.FamilyId,
			Role:      family_member_sign_in.Role,
			IsAdult:   family_member_sign_in.IsAdult,
			IsAdopted: family_member_sign_in.IsAdopted,
		}
		//保存家庭成员
		if err = family_member.Create(); err != nil {
			util.Debug(" Cannot create family_member", err)
			report(w, s_u, "你好，茶博士正在忙碌中，厚厚的眼镜失踪了，稍后再试。")
			return
		}
		//如果role==1，2，表示家庭成员是家庭的父母角色，那么需要更新家庭的名称
		if family_member.Role == 1 || family_member.Role == 2 {
			family := data.Family{
				Id: family_member.FamilyId,
			}
			if err = family.Get(); err != nil {
				util.Debug(family.Id, " Cannot get family given id")
				report(w, s_u, "你好，茶博士正在忙碌中，厚厚的眼镜不见了，稍后再试。")
				return
			}
			//使用新方法自动更新家庭名称，将占位符*替换为实际配偶姓名
			if err = family.UpdateFamilyNameWithSpouse(family_member.UserId); err != nil {
				util.Debug(" Cannot update family name", err)
				// 不阻断流程，只记录错误
			}
		}

		//更新声明书状态为"已确认“ 2
		family_member_sign_in.Status = data.SignInStatusConfirmed
		if err = family_member_sign_in.Update(); err != nil {
			util.Debug(" Cannot update family_member_sign_in", err)
			report(w, s_u, "你好，茶博士正在忙碌中，厚厚的眼镜不见了，稍后再试。")
			return
		}
		family_member_sign_in_reply.IsConfirm = true

	} else {
		//拒绝加入家庭
		//在声明书状态中更新为“已否认”
		family_member_sign_in.Status = data.SignInStatusDenied
		if err = family_member_sign_in.Update(); err != nil {
			util.Debug(" Cannot update family_member_sign_in", err)
			report(w, s_u, "你好，茶博士正在忙碌中，厚厚的眼镜不见了，稍后再试。")
			return
		}
		family_member_sign_in_reply.IsConfirm = false

	}
	//保存家庭成员声明书答复
	if err = family_member_sign_in_reply.Create(); err != nil {
		util.Debug(" Cannot create family_member_sign_in_reply", err)
		report(w, s_u, "你好，茶博士正在忙碌中，乱花渐欲迷人眼，请稍后再试。")
		return
	}

	if reply_bool {
		//跳转到家庭茶团页面,成员列表上有该茶友，表示已经加入成功
		family := data.Family{
			Id: family_member_sign_in.FamilyId,
		}
		if err = family.Get(); err != nil {
			util.Debug(family.Id, " Cannot get family given id")
			report(w, s_u, "你好，茶博士正在忙碌中，乱花渐欲迷人眼，请稍后再试。")
			return
		}
		http.Redirect(w, r, "/v1/family/detail?id="+(family.Uuid), http.StatusFound)
		return
	}

	//报告保存(否认是成员)成功
	t := fmt.Sprintf("你好，茶博士已经保存关于 %s 否认是成员答复。", family_member_sign_in.Title)
	report(w, s_u, t)

}

// HandleFamilyMemberEdit 处理编辑家庭成员资料
func HandleFamilyMemberEdit(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		FamilyMemberEditGet(w, r)
	case http.MethodPost:
		FamilyMemberEditPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/family_member/edit?id=xxx
func FamilyMemberEditGet(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	member_uuid := r.URL.Query().Get("id")
	fm := data.FamilyMember{Uuid: member_uuid}
	if err = fm.GetByUuid(); err != nil {
		report(w, s_u, "未找到成员资料")
		return
	}

	family := data.Family{Id: fm.FamilyId}
	if err = family.Get(); err != nil {
		report(w, s_u, "未找到家庭资料")
		return
	}

	isParent, _ := family.IsParentMember(s_u.Id)
	if !isParent {
		report(w, s_u, "只有父母角色可以编辑成员资料")
		return
	}

	fmBean, err := fetchFamilyMemberBean(fm)
	if err != nil {
		report(w, s_u, "获取成员资料失败")
		return
	}

	familyBean, err := fetchFamilyBean(family)
	if err != nil {
		report(w, s_u, "获取家庭资料失败")
		return
	}

	type EditData struct {
		SessUser         data.User
		FamilyBean       data.FamilyBean
		FamilyMemberBean data.FamilyMemberBean
	}

	generateHTML(w, &EditData{s_u, familyBean, fmBean}, "layout", "navbar.private", "family_member.edit")
}

// POST /v1/family_member/edit
func FamilyMemberEditPost(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	if err = r.ParseForm(); err != nil {
		report(w, s_u, "表单解析失败")
		return
	}

	member_uuid := r.PostFormValue("member_id")
	fm := data.FamilyMember{Uuid: member_uuid}
	if err = fm.GetByUuid(); err != nil {
		report(w, s_u, "未找到成员资料")
		return
	}

	family := data.Family{Id: fm.FamilyId}
	if err = family.Get(); err != nil {
		report(w, s_u, "未找到家庭资料")
		return
	}

	isParent, _ := family.IsParentMember(s_u.Id)
	if !isParent {
		report(w, s_u, "只有父母角色可以编辑成员资料")
		return
	}

	fm.NickName = r.PostFormValue("nickname")

	if birthday := r.PostFormValue("birthday"); birthday != "" {
		if t, err := data.ParseDate(birthday); err == nil {
			fm.Birthday = &t
		}
	}

	if deathDate := r.PostFormValue("death_date"); deathDate != "" {
		if t, err := data.ParseDate(deathDate); err == nil {
			fm.DeathDate = &t
		}
	} else {
		fm.DeathDate = nil
	}

	if order := r.PostFormValue("order"); order != "" {
		if o, err := strconv.Atoi(order); err == nil {
			fm.OrderOfSeniority = o
		}
	}

	if err = fm.UpdateMemberInfo(); err != nil {
		util.Debug("更新成员资料失败", err)
		report(w, s_u, "保存失败，请稍后再试")
		return
	}

	http.Redirect(w, r, "/v1/family/detail?id="+family.Uuid, http.StatusFound)
}

// GET /v1/family_member/detail?id=xxx
func FamilyMemberDetail(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	member_uuid := r.URL.Query().Get("id")
	fm := data.FamilyMember{Uuid: member_uuid}
	if err = fm.GetByUuid(); err != nil {
		report(w, s_u, "未找到成员资料")
		return
	}

	family := data.Family{Id: fm.FamilyId}
	if err = family.Get(); err != nil {
		report(w, s_u, "未找到家庭资料")
		return
	}

	isMember, _ := family.IsMember(s_u.Id)
	if !family.IsOpen && !isMember {
		report(w, s_u, "无权查看此成员资料")
		return
	}

	fmBean, err := fetchFamilyMemberBean(fm)
	if err != nil {
		report(w, s_u, "获取成员资料失败")
		return
	}

	familyBean, err := fetchFamilyBean(family)
	if err != nil {
		report(w, s_u, "获取家庭资料失败")
		return
	}

	isParent, _ := family.IsParentMember(s_u.Id)

	type DetailData struct {
		SessUser         data.User
		FamilyBean       data.FamilyBean
		FamilyMemberBean data.FamilyMemberBean
		IsParent         bool
	}

	generateHTML(w, &DetailData{s_u, familyBean, fmBean, isParent}, "layout", "navbar.private", "family_member.detail", "component_avatar_name_gender")
}
