package route

import (
	"fmt"
	"net/http"
	"strconv"
	data "teachat/DAO"
	util "teachat/Util"
)

// Get /v1/families/home
// 浏览&家庭茶团队列
func HomeFamilies(w http.ResponseWriter, r *http.Request) {
	// 1. get session
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Warning(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	// 2. get user's family
	family_list, err := data.GetFamiliesByAuthorId(s_u.Id)
	if err != nil {
		util.Warning(err, s_u.Id, "Cannot get user's family given id")
		Report(w, r, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
		return
	}
	f_b_list, err := FetchFamilyBeanList(family_list)
	if err != nil {
		util.Warning(err, s_u.Id, "Cannot get user's family")
		Report(w, r, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
		return
	}
	//截短 family.introduction 内容为66中文字，方便排版浏览
	for _, bean := range f_b_list {
		bean.Family.Introduction = Substr(bean.Family.Introduction, 66)
	}

	// 3. render
	var fSPD data.FamilySquare
	fSPD.SessUser = s_u
	fSPD.FamilyBeanList = f_b_list

	RenderHTML(w, &fSPD, "layout", "navbar.private", "families.home")
}

// GET /v1/family/detail?id=XXX
// 查看&家庭茶团详情
func FamilyDetail(w http.ResponseWriter, r *http.Request) {
	// 1. get session
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		util.Warning(err, "Cannot get user from session")
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}

	var fD data.FamilyDetail

	// 2. get family
	family_uuid := r.URL.Query().Get("id")

	//用户如果没有设置默认家庭，则其uuid为x(数据库查找方法返回特殊值)，则报告无信息可供查看。
	if family_uuid == "x" {
		Report(w, r, "家是世界上最温暖的地方，没有之一。")
		return
	}

	family := data.Family{
		Uuid: family_uuid,
	}
	if err = family.GetByUuid(); err != nil {
		Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}

	//检查当前用户是否可以查看这个家庭茶团资料
	isMember := false
	fD.IsNewMember = false
	// 3. check user is member of family
	isMember, err = family.IsMember(s_u.Id)
	if err != nil {
		util.Warning(err, s_u.Id, "Cannot check user is_member of family")
		Report(w, r, "你好，茶博士摸摸满头大汗，说因为外星人突然出现导致未能查看&家庭茶团详情。")
		return
	}
	if !isMember {
		// 如果不是家庭成员，则尝试读取这个家庭的增加新成员声明，看当前用户是否是某个声明对象
		// check user is new member of family
		family_member_sign_in := data.FamilyMemberSignIn{
			FamilyId: family.Id,
			UserId:   s_u.Id,
		}
		if err = family_member_sign_in.GetByFamilyIdMemberUserId(); err != nil {
			//查询资料出现失误
			util.Warning(err, "Cannot get family member sign in")
			Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
			return
		}
		//是新成员声明书中的茶友
		fD.IsNewMember = true
		fD.FamilyMemberSignIn = family_member_sign_in
		isMember = true
	}

	// 如果不是家庭成员或者新成员，
	// 检查家庭是否被设置为公开，否则不能查看
	if !isMember && !family.IsOpen {
		Report(w, r, "你好，茶博士摸摸头，竟然说你不是这个&家庭茶团的成员，未能查看&家庭茶团详情。")
		return

	}

	//读取目标家庭的资料夹
	family_bean, err := FetchFamilyBean(family)
	if err != nil {
		util.Warning(err, family.Id, "Cannot fetch bean given family")
		Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	f := data.Family{
		Id: family.Id,
	}
	f_p_members, err := f.ParentMembers()
	if err != nil {
		util.Warning(err, family.Id, "Cannot fetch family's parent members")
		Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	parent_member_bean_list, err := FetchFamilyMemberBeanList(f_p_members)
	if err != nil {
		util.Warning(err, family.Id, "Cannot fetch family's parent members bean")
		Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}

	c_members, err := f.ChildMembers()
	if err != nil {
		util.Warning(err, family.Id, "Cannot fetch family's child members")
		Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	child_member_bean_list, err := FetchFamilyMemberBeanList(c_members)
	if err != nil {
		util.Warning(err, family.Id, "Cannot fetch family's child members")
		Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	other_members, err := f.OtherMembers()
	if err != nil {
		util.Warning(err, family.Id, "Cannot fetch family's other members")
		Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	other_member_bean_list, err := FetchFamilyMemberBeanList(other_members)
	if err != nil {
		util.Warning(err, family.Id, "Cannot fetch family's other members bean")
		Report(w, r, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
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
	fD.ParentMemberBeanList = parent_member_bean_list
	fD.ChildMemberBeanList = child_member_bean_list
	fD.OtherMemberBeanList = other_member_bean_list

	// 4. render
	RenderHTML(w, &fD, "layout", "navbar.private", "family.detail")
}

// HandleNewFamily() /v1/family/new
func HandleNewFamily(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		NewFamily(w, r)
	case http.MethodPost:
		// return family id
		CreateFamily(w, r)
	default:
		// return error
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

// POST /v1/family/new
// create new family
func CreateFamily(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Warning(err, " Cannot parse form")
		Report(w, r, "你好，茶博士失魂鱼，未能开新茶团，请稍后再试。")
		return
	}
	s, err := Session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		Report(w, r, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}

	// n := r.PostFormValue("name")
	// // &家庭茶团名称是否在4-24中文字符
	// l := CnStrLen(n)
	// if l < 4 || l > 24 {
	// 	Report(w, r, "你好，茶博士摸摸头，竟然说&家庭茶团名字字数太多或者太少，未能创建新茶团。")
	// 	return
	// }
	introduction := r.PostFormValue("introduction")
	// 检测introduction是否在17-456中文字符
	lenI := CnStrLen(introduction)
	if lenI < 3 || lenI > 456 {
		Report(w, r, "你好，茶博士摸摸头，竟然说&家庭茶团价绍字数太多或者太少，未能创建新茶团。")
		return
	}

	//声明一个空白&家庭茶团
	var new_family data.Family
	//读取提交的is_open的checkbox值，判断&家庭茶团是否公开
	is_open_str := r.PostFormValue("is_open")
	//fmt.Println(is_open_str)
	if is_open_str == "on" {
		new_family.IsOpen = true
	} else {
		new_family.IsOpen = false
	}
	//读取提交的家庭状态
	status_str := r.PostFormValue("status")
	// change str into int
	status_int, err := strconv.Atoi(status_str)
	if err != nil {
		Report(w, r, "你好，茶博士摸摸头，竟然说&家庭茶团状态看不懂，未能创建新茶团。")
		return
	}
	// 0 =< status_int <= 5
	if status_int < 0 || status_int > 5 {
		Report(w, r, "你好，茶博士摸摸头，竟然说&家庭茶团状态看不懂，未能创建新茶团。")
		return
	}
	//fmt.Println(status_int)
	new_family.Status = status_int

	child := r.PostFormValue("child")
	if child == "yes" {
		new_family.HasChild = true
	} else if child == "no" {
		new_family.HasChild = false
	} else {
		Report(w, r, "你好，茶博士摸摸头，&家庭茶团是否有孩子？看不懂提交内容，未能创建新茶团。")
		return
	}
	new_family.AuthorId = s_u.Id
	new_family.Name = s_u.Name + "&"
	new_family.Introduction = introduction
	//初始化家庭茶团默认参数
	new_family.HusbandFromFamilyId = 0
	new_family.WifeFromFamilyId = 0
	new_family.Logo = "familyLogo"

	//保存到数据库中,返回新家庭茶团的id
	if err := new_family.Create(); err != nil {
		util.Warning(err, s_u.Email, "Cannot create new family")
		Report(w, r, "你好，茶博士摸摸头，未能创建新茶团，请稍后再试。")
		return
	}
	//把创建者登记为默认&家庭茶团成员
	//声明一个空白家庭成员
	new_member := data.FamilyMember{
		FamilyId: new_family.Id,
		UserId:   s_u.Id,
		Role:     0,
	}
	if s_u.Gender == 1 {
		new_member.Role = 1
	} else {
		new_member.Role = 2
	}
	if err := new_member.Create(); err != nil {
		util.Warning(err, s_u.Email, "Cannot create new family member")
		Report(w, r, "你好，茶博士摸摸头，未能创建新茶团，请稍后再试。")
		return
	}

	//报告用户登记家庭茶团成功
	text := ""
	if s_u.Gender == 0 {
		text = fmt.Sprintf("%s 女士，你好，登记 %s 成功，祝愿拥有快乐品茶时光。", s_u.Name, new_family.Name)
	} else {
		text = fmt.Sprintf("%s 先生，你好，登记 %s 成功，祝愿拥有美好品茶时光。", s_u.Name, new_family.Name)
	}
	Report(w, r, text)
}

// GET /v1/family/new
// 返回一张空白的家庭填写表格（页面）
func NewFamily(w http.ResponseWriter, r *http.Request) {
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
	var fSPD data.FamilySquare
	fSPD.SessUser = s_u

	RenderHTML(w, &fSPD, "layout", "navbar.private", "family.new")
}
