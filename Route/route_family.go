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

// GET /v1/family/default?id=
// 设置某个茶友的默认家庭茶团
func SetDefaultFamily(w http.ResponseWriter, r *http.Request) {
	// 1. get session
	s, err := session(r)
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
	// 2. get family id
	family_uuid := r.URL.Query().Get("id")
	//check family is valid
	if family_uuid == data.FamilyUuidUnknown {
		report(w, r, "你好，茶博士摸摸头竟然说，陛下这个特殊家庭茶团不允许私用呢。")
		return
	}
	t_family := data.Family{
		Uuid: family_uuid,
	}
	//fetch family
	if err = t_family.GetByUuid(); err != nil {
		util.Debug("Cannot get family by uuid", err)
		report(w, r, "你好，茶博士摸摸头，竟然说这个家庭茶团不存在。")
		return
	}
	//check user is family member
	ok, err := t_family.IsMember(s_u.Id)
	if err != nil {
		util.Debug("Cannot check user is family member", err)
		report(w, r, "你好，茶博士摸摸头，竟然说这个家庭茶团不存在。")
		return
	}
	//if not member
	if !ok {
		report(w, r, "你好，茶博士摸摸头竟然说，陛下真的和这个家庭茶团有关系吗？")
		return
	}

	//检查这个新的默认家庭茶团是否已经设置，避免重复记录
	//fetch user default family
	lastDefaultFamily, err := s_u.GetLastDefaultFamily()
	if err != nil {
		util.Debug("Cannot get user's last default family", err)
		report(w, r, "你好，茶博士摸摸头，竟然说墨水用完了，设置默认家庭茶团失败。")
		return
	}
	//if last default family is not Unknown
	if lastDefaultFamily.Id > data.FamilyIdUnknown {
		//if last default family is  equal to the new default family
		if lastDefaultFamily.Id == t_family.Id {
			report(w, r, "你好，茶博士竟然说,请勿重复设置默认家庭茶团。")
			return
		}

	}

	// set default family
	new_user_default_family := data.UserDefaultFamily{
		UserId:   s_u.Id,
		FamilyId: t_family.Id,
	}
	if err = new_user_default_family.Create(); err != nil {
		util.Debug("Cannot create user default family", err)
		report(w, r, "你好，茶博士摸摸头，竟然说墨水用完了，设置默认家庭茶团失败。")
		return
	}

	// redirect
	http.Redirect(w, r, "/v1/families/home", http.StatusFound)
}

// Get /v1/families/home
// 浏览&家庭茶团队列
func HomeFamilies(w http.ResponseWriter, r *http.Request) {
	// 1. get session
	s, err := session(r)
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
	// 2. get user's family
	family_slice, err := data.GetAllFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's family given id")
		report(w, r, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
		return
	}

	f_b_slice, err := fetchFamilyBeanSlice(family_slice)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's family")
		report(w, r, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
		return
	}

	var fSPD data.FamilySquare

	f_b_l_len := len(f_b_slice)
	if f_b_l_len != 0 {
		//如果len(f_b_slice)!=0,说明用户已经登记有家庭茶团，

		fSPD.IsEmpty = false

		//2.1 get user's default family
		l_default_family, err := s_u.GetLastDefaultFamily()
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				l_default_family = data.FamilyUnknown
			} else {
				util.Debug(s_u.Id, "Cannot get user's default family")
				report(w, r, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有默认家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
				return
			}

		}

		for i, bean := range f_b_slice {
			//截短 family.introduction 内容为66中文字，方便排版浏览
			bean.Family.Introduction = subStr(bean.Family.Introduction, 66)

			//if l_default_family.id == bean.family.id ,fSPD.DefaultFamilyBean = bean
			if bean.Family.Id == l_default_family.Id {

				fSPD.DefaultFamilyBean = bean
				//remove this bean from f_b_slice
				f_b_slice = append(f_b_slice[:i], f_b_slice[i+1:]...)
			}

		}

		fSPD.OtherFamilyBeanSlice = f_b_slice

	} else {
		//如果len(f_b_slice)==0,说明用户还没有登记任何家庭茶团，那么标识为空
		fSPD.IsEmpty = true
	}

	fSPD.SessUser = s_u

	// 3. render
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.home")
}

// GET /v1/family/detail?id=XXX
// 查看&家庭茶团详情
// 需要检查会话用户是否被这个家庭声明为新成员，这影响是否展示新成员声明
// 如果会话用户是家庭成员，可以直接查看详情，
// 如果这个家庭设置isopen==false，检查会话用户不是家庭成员，也不是被声明为新成员，那么不能查看家庭茶团资料
func FamilyDetail(w http.ResponseWriter, r *http.Request) {
	// 1. get session
	s, err := session(r)
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

	var fD data.FamilyDetail

	// 2. get family
	family_uuid := r.URL.Query().Get("id")

	//用户如果没有设置默认家庭，则其uuid为x.
	//报告无信息可供查看。
	if family_uuid == data.FamilyUuidUnknown {
		report(w, r, "盛世无饥馑，四海可为家。")
		return
	}

	family := data.Family{
		Uuid: family_uuid,
	}
	if err = family.GetByUuid(); err != nil {
		util.Debug("Cannot get family by UUID", err)
		report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}

	fD.IsNewMember = false
	isMember := false
	// 3. check user is member of family
	isMember, err = family.IsMember(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot check user is_member of family")
		report(w, r, "你好，茶博士摸摸满头大汗，说因为外星人突然出现导致未能查看&家庭茶团详情。")
		return
	}
	if !isMember {
		// 不是家庭成员，则尝试读取这个家庭的增加新成员声明，看当前用户是否是某个声明对象
		// check user is new member of family
		family_member_sign_in := data.FamilyMemberSignIn{
			FamilyId: family.Id,
			UserId:   s_u.Id,
		}
		if err = family_member_sign_in.GetByFamilyIdMemberUserId(); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				//查询资料出现失误
				util.Debug("Cannot get family member sign in", err)
				report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
				return
			}
			fD.IsNewMember = false
		} else {
			//是新成员声明书中的茶友
			fD.IsNewMember = true
			fD.FamilyMemberSignIn = family_member_sign_in
		}

	}

	//检查当前用户是否可以查看这个家庭茶团资料
	// 如果不是家庭成员或者新成员，
	// 检查家庭是否被设置为公开，否则不能查看
	if !family.IsOpen {
		if !isMember && !fD.IsNewMember {
			report(w, r, "你好，茶博士摸摸头，竟然说你不是这个&家庭茶团的成员，未能查看&家庭茶团详情。")
			return
		}
	}

	//读取目标家庭的资料夹
	family_bean, err := fetchFamilyBean(family)
	if err != nil {
		util.Debug(family.Id, "Cannot fetch bean given family")
		report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	f := data.Family{
		Id: family.Id,
	}
	f_p_members, err := f.ParentMembers()
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's parent members")
		report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	parent_member_bean_slice, err := fetchFamilyMemberBeanSlice(f_p_members)
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's parent members bean")
		report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}

	c_members, err := f.ChildMembers()
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's child members")
		report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	child_member_bean_slice, err := fetchFamilyMemberBeanSlice(c_members)
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's child members")
		report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	other_members, err := f.OtherMembers()
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's other members")
		report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	other_member_bean_slice, err := fetchFamilyMemberBeanSlice(other_members)
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's other members bean")
		report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}

	// 3.1 check user is parent of family
	fD.IsParent = false
	for _, f_p_member := range f_p_members {
		if f_p_member.UserId == s_u.Id {
			fD.IsParent = true
		}
	}

	//3.2 check user is child of family
	fD.IsChild = false
	for _, c_member := range c_members {
		if c_member.UserId == s_u.Id {
			fD.IsChild = true
		}
	}

	// 3.3 check user is other member of family
	fD.IsOther = false
	for _, o_member := range other_members {
		if o_member.UserId == s_u.Id {
			fD.IsOther = true
		}
	}

	fD.SessUser = s_u
	fD.FamilyBean = family_bean
	fD.ParentMemberBeanSlice = parent_member_bean_slice
	fD.ChildMemberBeanSlice = child_member_bean_slice
	fD.OtherMemberBeanSlice = other_member_bean_slice

	// 4. render
	generateHTML(w, &fD, "layout", "navbar.private", "family.detail", "component_avatar_name_gender")
}

// HandleNewFamily() /v1/family/new
func HandleNewFamily(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		NewFamily(w, r)
	case http.MethodPost:
		// return family id
		SaveFamily(w, r)
	default:
		// return error
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

// POST /v1/family/new
// create new family
func SaveFamily(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, r, "你好，茶博士失魂鱼，未能开新茶团，请稍后再试。")
		return
	}
	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	// 读取提交的家庭状态
	status_str := r.PostFormValue("status")
	// change str into int
	status_int, err := strconv.Atoi(status_str)
	if err != nil {
		report(w, r, "你好，茶博士摸摸头，竟然说&家庭茶团状态看不懂，未能创建新茶团。")
		return
	}
	// 0 =< status_int <= 5
	if status_int < 0 || status_int > 5 {
		report(w, r, "你好，茶博士摸摸头，竟然说&家庭茶团状态看不懂，未能创建新茶团。")
		return
	}

	introduction := r.PostFormValue("introduction")
	// 检测introduction是否在min-int(util.Config.ThreadMaxWord)中文字符
	lenI := cnStrLen(introduction)
	if lenI < int(util.Config.ThreadMinWord) || lenI > int(util.Config.ThreadMaxWord) {
		report(w, r, "你好，茶博士摸摸头，竟然说&家庭茶团价绍字数太多或者太少，未能创建新茶团。")
		return
	}

	//声明一个空白&家庭茶团
	var new_family data.Family

	new_family.Status = status_int

	//读取提交的is_open的checkbox值，判断&家庭茶团是否公开
	is_open_str := r.PostFormValue("is_open")
	//fmt.Println(is_open_str)
	if is_open_str == "on" {
		new_family.IsOpen = true
	} else {
		new_family.IsOpen = false
	}

	child := r.PostFormValue("child")
	switch child {
	case "yes":
		new_family.HasChild = true
	case "no":
		new_family.HasChild = false
	default:
		report(w, r, "你好，茶博士摸摸头，&家庭茶团是否有孩子？看不懂提交内容，未能创建新茶团。")
		return
	}

	new_family.Name = s_u.Name + "&"

	new_family.AuthorId = s_u.Id
	new_family.Introduction = introduction
	//初始化家庭茶团默认参数
	new_family.HusbandFromFamilyId = data.FamilyIdUnknown
	new_family.WifeFromFamilyId = data.FamilyIdUnknown
	new_family.Logo = "familyLogo"

	//保存到数据库中,返回新家庭茶团的id
	if err := new_family.Create(); err != nil {
		util.Debug(s_u.Email, "Cannot create new family")
		report(w, r, "你好，茶博士摸摸头，未能创建新茶团，请稍后再试。")
		return
	}

	//把创建者登记为默认&家庭茶团成员之一
	//声明一个家庭成员
	author_member := data.FamilyMember{
		FamilyId: new_family.Id,
		UserId:   s_u.Id,
		IsAdult:  true,
	}
	//根据茶友性别，设置其相应的男主
	if s_u.Gender == data.User_Gender_Male {
		author_member.Role = data.FamilyMemberRoleHusband
	} else {
		//或者女主角色
		author_member.Role = data.FamilyMemberRoleWife
	}
	if err := author_member.Create(); err != nil {
		util.Debug(s_u.Email, "Cannot create author family member")
		report(w, r, "你好，茶博士摸摸头，未能创建新茶团，请稍后再试。")
		return
	}

	//检查会话茶友是否已经设置了默认首选家庭茶团
	//如果已经设置了默认首选家庭茶团，则不再设置
	df, err := s_u.GetLastDefaultFamily()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			df = data.FamilyUnknown
		} else {
			util.Debug(s_u.Id, "Cannot get user's default family")
			report(w, r, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有默认家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
			return
		}
	}

	if df.Id == data.FamilyIdUnknown {
		//还没有设置默认家庭
		udf := data.UserDefaultFamily{
			UserId:   s_u.Id,
			FamilyId: new_family.Id,
		}
		//把这个新家庭茶团设为默认
		if err := udf.Create(); err != nil {
			util.Debug(s_u.Email, "Cannot create user's default family")
			report(w, r, "你好，茶博士摸摸头，未能创建默认家庭茶团，请稍后再试。")
			return
		}
	}

	//报告用户登记家庭茶团成功
	text := ""
	if s_u.Gender == data.User_Gender_Female {
		text = fmt.Sprintf("%s 女士，你好，登记 %s 家庭茶团成功，可以到我的家庭中查看详情，祝愿拥有快乐品茶时光。", s_u.Name, new_family.Name)
	} else {
		text = fmt.Sprintf("%s 先生，你好，登记 %s 家庭茶团成功，可以到我的家庭中查看详情，祝愿拥有美好品茶时光。", s_u.Name, new_family.Name)
	}
	report(w, r, text)
}

// GET /v1/family/new
// 返回一张空白的家庭填写表格（页面）
func NewFamily(w http.ResponseWriter, r *http.Request) {
	s, err := session(r)
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
	var fSPD data.FamilySquare
	fSPD.SessUser = s_u

	generateHTML(w, &fSPD, "layout", "navbar.private", "family.new")
}
