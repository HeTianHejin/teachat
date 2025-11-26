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

// GET /v1/family/default?uuid=
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
	if family_uuid == data.FamilyUuidUnknown || family_uuid == "" {
		report(w, s_u, "你好，茶博士摸摸头竟然说，陛下这个特殊家庭茶团不允许私用呢。")
		return
	}
	t_family := data.Family{
		Uuid: family_uuid,
	}
	//fetch family
	if err = t_family.GetByUuid(); err != nil {
		util.Debug("Cannot get family by uuid", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个家庭茶团不存在。")
		return
	}
	//check family is open
	if !t_family.IsOpen {
		report(w, s_u, "你好，茶博士摸摸头竟然说，这个家庭茶团未公开，不能设为默认家庭。")
		return
	}
	//check user is family member
	ok, err := t_family.IsMember(s_u.Id)
	if err != nil {
		util.Debug("Cannot check user is family member", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个家庭茶团不存在。")
		return
	}
	//if not member
	if !ok {
		report(w, s_u, "你好，茶博士摸摸头竟然说，陛下真的和这个家庭茶团有关系吗？")
		return
	}

	//检查这个新的默认家庭茶团是否已经设置，避免重复记录
	//fetch user default family
	lastDefaultFamily, err := s_u.GetLastDefaultFamily()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		util.Debug("Cannot get user's last default family", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说墨水用完了，设置默认家庭茶团失败。")
		return
	} else if errors.Is(err, sql.ErrNoRows) || lastDefaultFamily.Id > data.FamilyIdUnknown {
		//if last default family is not Unknown or NoRows
		if lastDefaultFamily.Id == t_family.Id {
			//if last default family is  equal to the new default family
			report(w, s_u, "你好，茶博士竟然说,请勿重复设置默认家庭茶团。")
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
		report(w, s_u, "你好，茶博士摸摸头，竟然说墨水用完了，设置默认家庭茶团失败。")
		return
	}

	// redirect
	http.Redirect(w, r, "/v1/family/home", http.StatusFound)
}

// Get /v1/family/home
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
	// 2. get user's family - 默认显示公开家庭
	family_slice, err := data.ParentMemberOpenFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's family given id", err)
		report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
		return
	}

	f_b_slice, err := fetchFamilyBeanSlice(family_slice)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's family", err)
		report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
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
				util.Debug(s_u.Id, "Cannot get user's default family", err)
				report(w, s_u, "你好，乱花渐欲迷人眼，未能查看家庭茶团列表。")
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

// GET /v1/family/tree?id=
// 查看家族树
func FamilyTree(w http.ResponseWriter, r *http.Request) {
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

	family_uuid := r.URL.Query().Get("id")
	if family_uuid == data.FamilyUuidUnknown {
		report(w, s_u, "盛世无饥馑，四海可为家。")
		return
	}

	family := data.Family{Uuid: family_uuid}
	if err = family.GetByUuid(); err != nil {
		util.Debug("Cannot get family by uuid", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记。")
		return
	}

	isMember, err := family.IsMember(s_u.Id)
	if err != nil {
		util.Debug("Cannot check user is family member", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说查询出错了。")
		return
	}
	if !family.IsOpen && !isMember {
		report(w, s_u, "你好，茶博士摸摸头，竟然说你不是这个&家庭茶团的成员，未能查看家族树。")
		return
	}

	type FamilyTreeData struct {
		SessUser data.User
		Family   data.Family
	}

	var ftd FamilyTreeData
	ftd.SessUser = s_u
	ftd.Family = family

	generateHTML(w, &ftd, "layout", "navbar.private", "family.tree")
}

// GET /v1/families/parent
// 查看父代家庭（用户作为子女的家庭）
func ParentFamilies(w http.ResponseWriter, r *http.Request) {
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

	family_slice, err := data.ChildMemberFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's parent families", err)
		report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有父代家庭茶团。", s_u.Email))
		return
	}

	f_b_slice, err := fetchFamilyBeanSlice(family_slice)
	if err != nil {
		util.Debug(s_u.Id, "Cannot fetch family beans", err)
		report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，未能查看这个用户%s父代家庭茶团列表。", s_u.Email))
		return
	}

	var fSPD data.FamilySquare
	fSPD.IsEmpty = len(f_b_slice) == 0
	if !fSPD.IsEmpty {
		for i := range f_b_slice {
			f_b_slice[i].Family.Introduction = subStr(f_b_slice[i].Family.Introduction, 66)
		}
		fSPD.OtherFamilyBeanSlice = f_b_slice
	}
	fSPD.SessUser = s_u
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.parent")
}

// GET /v1/families/child
// 查看子代家庭（用户子女的家庭）
func ChildFamilies(w http.ResponseWriter, r *http.Request) {
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

	// 查找用户作为父母的家庭，然后查找这些家庭的子女成员，再查找子女的家庭
	parent_families, err := data.ParentMemberFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's families", err)
		report(w, s_u, s_u, "你好，茶博士摸摸头，未能查看子代家庭茶团列表。")
		return
	}

	var child_families []data.Family
	for _, pf := range parent_families {
		children, _ := pf.ChildMembers()
		for _, child := range children {
			if child.IsAdult {
				child_fams, _ := data.ParentMemberFamilies(child.UserId)
				child_families = append(child_families, child_fams...)
			}
		}
	}

	f_b_slice, err := fetchFamilyBeanSlice(child_families)
	if err != nil {
		util.Debug(s_u.Id, "Cannot fetch family beans", err)
		report(w, s_u, s_u, "你好，茶博士摸摸头，未能查看子代家庭茶团列表。")
		return
	}

	var fSPD data.FamilySquare
	fSPD.IsEmpty = len(f_b_slice) == 0
	if !fSPD.IsEmpty {
		for i := range f_b_slice {
			f_b_slice[i].Family.Introduction = subStr(f_b_slice[i].Family.Introduction, 66)
		}
		fSPD.OtherFamilyBeanSlice = f_b_slice
	}
	fSPD.SessUser = s_u
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.child")
}

// GET /v1/families/in-laws
// 查看外家姻亲（配偶的父代和子代家庭）
func InLawsFamilies(w http.ResponseWriter, r *http.Request) {
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

	// 查找用户作为父母的家庭
	my_families, err := data.ParentMemberFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's families", err)
		report(w, s_u, s_u, "你好，茶博士摸摸头，未能查看外家姻亲茶团列表。")
		return
	}

	var inlaw_families []data.Family
	// 遍历用户的家庭，找到配偶
	for _, my_fam := range my_families {
		parents, _ := my_fam.ParentMembers()
		for _, parent := range parents {
			// 找到配偶（不是自己的父母成员）
			if parent.UserId != s_u.Id {
				// 获取配偶作为子女的家庭（配偶父代）
				spouse_parent_fams, _ := data.ChildMemberFamilies(parent.UserId)
				inlaw_families = append(inlaw_families, spouse_parent_fams...)

				// 获取配偶子女的家庭（配偶子代）
				spouse_child_fams, _ := data.ParentMemberFamilies(parent.UserId)
				for _, scf := range spouse_child_fams {
					children, _ := scf.ChildMembers()
					for _, child := range children {
						if child.IsAdult {
							child_fams, _ := data.ParentMemberFamilies(child.UserId)
							inlaw_families = append(inlaw_families, child_fams...)
						}
					}
				}
			}
		}
	}

	f_b_slice, err := fetchFamilyBeanSlice(inlaw_families)
	if err != nil {
		util.Debug(s_u.Id, "Cannot fetch family beans", err)
		report(w, s_u, s_u, "你好，茶博士摸摸头，未能查看外家姻亲茶团列表。")
		return
	}

	var fSPD data.FamilySquare
	fSPD.IsEmpty = len(f_b_slice) == 0
	if !fSPD.IsEmpty {
		for i := range f_b_slice {
			f_b_slice[i].Family.Introduction = subStr(f_b_slice[i].Family.Introduction, 66)
		}
		fSPD.OtherFamilyBeanSlice = f_b_slice
	}
	fSPD.SessUser = s_u
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.in-laws")
}

// GET /v1/families/gone
// 查看随风飘逝的家庭（已退出或已删除）
func GoneFamilies(w http.ResponseWriter, r *http.Request) {
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

	// 获取用户声明离开的家庭
	resign_families, err := data.ResignMemberFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's resign families", err)
		report(w, s_u, "你好，茶博士摸摸头，未能查看随风飘逝茶团列表。")
		return
	}

	// 获取用户已删除的家庭
	deleted_families, err := data.GetDeletedFamiliesByAuthorId(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's deleted families", err)
		report(w, s_u, "你好，茶博士摸摸头，未能查看随风飘逝茶团列表。")
		return
	}

	// 合并两个列表
	all_gone_families := append(resign_families, deleted_families...)

	f_b_slice, err := fetchFamilyBeanSlice(all_gone_families)
	if err != nil {
		util.Debug(s_u.Id, "Cannot fetch family beans", err)
		report(w, s_u, "你好，茶博士摸摸头，未能查看随风飘逝茶团列表。")
		return
	}

	var fSPD data.FamilySquare
	fSPD.IsEmpty = len(f_b_slice) == 0
	if !fSPD.IsEmpty {
		for i := range f_b_slice {
			f_b_slice[i].Family.Introduction = subStr(f_b_slice[i].Family.Introduction, 66)
		}
		fSPD.OtherFamilyBeanSlice = f_b_slice
	}
	fSPD.SessUser = s_u
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.gone")
}

// Get /v1/family/home/private
// 浏览私密&家庭茶团队列
func HomePrivateFamilies(w http.ResponseWriter, r *http.Request) {
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
	family_slice, err := data.ParentMemberPrivateFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's private family", err)
		report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
		return
	}

	f_b_slice, err := fetchFamilyBeanSlice(family_slice)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's family", err)
		report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
		return
	}

	var fSPD data.FamilySquare
	fSPD.IsEmpty = len(f_b_slice) == 0
	if !fSPD.IsEmpty {
		for i := range f_b_slice {
			f_b_slice[i].Family.Introduction = subStr(f_b_slice[i].Family.Introduction, 66)
		}
		fSPD.OtherFamilyBeanSlice = f_b_slice
	}
	fSPD.SessUser = s_u
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.home.private")
}

// GET /v1/families/parent/private
// 查看私密父代家庭（用户作为子女的家庭）
func ParentPrivateFamilies(w http.ResponseWriter, r *http.Request) {
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

	family_slice, err := data.ChildMemberFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's parent families", err)
		report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有父代家庭茶团。", s_u.Email))
		return
	}

	var private_families []data.Family
	for _, fam := range family_slice {
		if !fam.IsOpen {
			private_families = append(private_families, fam)
		}
	}

	f_b_slice, err := fetchFamilyBeanSlice(private_families)
	if err != nil {
		util.Debug(s_u.Id, "Cannot fetch family beans", err)
		report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，未能查看这个用户%s父代家庭茶团列表。", s_u.Email))
		return
	}

	var fSPD data.FamilySquare
	fSPD.IsEmpty = len(f_b_slice) == 0
	if !fSPD.IsEmpty {
		for i := range f_b_slice {
			f_b_slice[i].Family.Introduction = subStr(f_b_slice[i].Family.Introduction, 66)
		}
		fSPD.OtherFamilyBeanSlice = f_b_slice
	}
	fSPD.SessUser = s_u
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.parent.private")
}

// GET /v1/families/in-laws/private
// 查看私密外家姻亲（配偶的父代和子代家庭）
func InLawsPrivateFamilies(w http.ResponseWriter, r *http.Request) {
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

	my_families, err := data.ParentMemberFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's families", err)
		report(w, s_u, "你好，茶博士摸摸头，未能查看外家姻亲茶团列表。")
		return
	}

	var inlaw_families []data.Family
	for _, my_fam := range my_families {
		parents, _ := my_fam.ParentMembers()
		for _, parent := range parents {
			if parent.UserId != s_u.Id {
				spouse_parent_fams, _ := data.ChildMemberFamilies(parent.UserId)
				for _, spf := range spouse_parent_fams {
					if !spf.IsOpen {
						inlaw_families = append(inlaw_families, spf)
					}
				}

				spouse_child_fams, _ := data.ParentMemberFamilies(parent.UserId)
				for _, scf := range spouse_child_fams {
					if !scf.IsOpen {
						children, _ := scf.ChildMembers()
						for _, child := range children {
							if child.IsAdult {
								child_fams, _ := data.ParentMemberFamilies(child.UserId)
								for _, cf := range child_fams {
									if !cf.IsOpen {
										inlaw_families = append(inlaw_families, cf)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	f_b_slice, err := fetchFamilyBeanSlice(inlaw_families)
	if err != nil {
		util.Debug(s_u.Id, "Cannot fetch family beans", err)
		report(w, s_u, "你好，茶博士摸摸头，未能查看外家姻亲茶团列表。")
		return
	}

	var fSPD data.FamilySquare
	fSPD.IsEmpty = len(f_b_slice) == 0
	if !fSPD.IsEmpty {
		for i := range f_b_slice {
			f_b_slice[i].Family.Introduction = subStr(f_b_slice[i].Family.Introduction, 66)
		}
		fSPD.OtherFamilyBeanSlice = f_b_slice
	}
	fSPD.SessUser = s_u
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.in-laws.private")
}

// GET /v1/families/gone/private
// 查看私密随风飘逝的家庭（已退出或已删除）
func GonePrivateFamilies(w http.ResponseWriter, r *http.Request) {
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

	resign_families, err := data.ResignMemberFamilies(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's resign families", err)
		report(w, s_u, "你好，茶博士摸摸头，未能查看随风飘逝茶团列表。")
		return
	}

	deleted_families, err := data.GetDeletedFamiliesByAuthorId(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot get user's deleted families", err)
		report(w, s_u, "你好，茶博士摸摸头，未能查看随风飘逝茶团列表。")
		return
	}

	var private_gone_families []data.Family
	for _, fam := range append(resign_families, deleted_families...) {
		if !fam.IsOpen {
			private_gone_families = append(private_gone_families, fam)
		}
	}

	f_b_slice, err := fetchFamilyBeanSlice(private_gone_families)
	if err != nil {
		util.Debug(s_u.Id, "Cannot fetch family beans", err)
		report(w, s_u, "你好，茶博士摸摸头，未能查看随风飘逝茶团列表。")
		return
	}

	var fSPD data.FamilySquare
	fSPD.IsEmpty = len(f_b_slice) == 0
	if !fSPD.IsEmpty {
		for i := range f_b_slice {
			f_b_slice[i].Family.Introduction = subStr(f_b_slice[i].Family.Introduction, 66)
		}
		fSPD.OtherFamilyBeanSlice = f_b_slice
	}
	fSPD.SessUser = s_u
	generateHTML(w, &fSPD, "layout", "navbar.private", "families.gone.private")
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
		report(w, s_u, "盛世无饥馑，四海可为家。")
		return
	}

	family := data.Family{
		Uuid: family_uuid,
	}
	if err = family.GetByUuid(); err != nil {
		util.Debug("Cannot get family by UUID", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}

	fD.IsNewMember = false
	isMember := false
	// 3. check user is member of family
	isMember, err = family.IsMember(s_u.Id)
	if err != nil {
		util.Debug(s_u.Id, "Cannot check user is_member of family", err)
		report(w, s_u, "你好，茶博士摸摸满头大汗，说因为外星人突然出现导致未能查看&家庭茶团详情。")
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
				report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
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
			report(w, s_u, "你好，茶博士摸摸头，竟然说你不是这个&家庭茶团的成员，未能查看&家庭茶团详情。")
			return
		}
	}

	//读取目标家庭的资料夹
	family_bean, err := fetchFamilyBean(family)
	if err != nil {
		util.Debug(family.Id, "Cannot fetch bean given family", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	f := data.Family{
		Id: family.Id,
	}
	f_p_members, err := f.ParentMembers()
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's parent members", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	parent_member_bean_slice, err := fetchFamilyMemberBeanSlice(f_p_members)
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's parent members bean", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}

	c_members, err := f.ChildMembers()
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's child members", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	child_member_bean_slice, err := fetchFamilyMemberBeanSlice(c_members)
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's child members", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	other_members, err := f.OtherMembers()
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's other members", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
		return
	}
	other_member_bean_slice, err := fetchFamilyMemberBeanSlice(other_members)
	if err != nil {
		util.Debug(family.Id, "Cannot fetch family's other members bean", err)
		report(w, s_u, "你好，茶博士摸摸头，竟然说这个&家庭茶团没有登记，未能查看&家庭茶团详情。")
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
		NewFamilyGet(w, r)
	case http.MethodPost:
		// return family id
		NewFamilyPost(w, r)
	default:
		// return error
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

// POST /v1/family/new
// create a new family
func NewFamilyPost(w http.ResponseWriter, r *http.Request) {

	s, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/v1/login", http.StatusFound)
		return
	}
	s_u, err := s.User()
	if err != nil {
		report(w, s_u, "你好，茶博士失魂鱼，未能创建新茶团，请稍后再试。")
		return
	}
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能开新茶团，请稍后再试。")
		return
	}
	// 读取提交的家庭状态
	status_str := r.PostFormValue("status")
	// change str into int
	status_int, err := strconv.Atoi(status_str)
	if err != nil {
		report(w, s_u, "你好，茶博士摸摸头，竟然说&家庭茶团状态看不懂，未能创建新茶团。")
		return
	}
	// 0 =< status_int <= 5
	if status_int < 0 || status_int > 5 {
		report(w, s_u, "你好，茶博士摸摸头，竟然说&家庭茶团状态看不懂，未能创建新茶团。")
		return
	}

	introduction := r.PostFormValue("introduction")
	// 检测introduction是否在min-int(util.Config.ThreadMaxWord)中文字符
	lenI := cnStrLen(introduction)
	if lenI < int(util.Config.ThreadMinWord) || lenI > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "你好，茶博士摸摸头，竟然说&家庭茶团价绍字数太多或者太少，未能创建新茶团。")
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
		report(w, s_u, "你好，茶博士摸摸头，&家庭茶团是否有孩子？看不懂提交内容，未能创建新茶团。")
		return
	}

	// 使用占位符*生成家庭名称，防止冒用他人姓名
	new_family.Name = s_u.Name + "&*"
	new_family.AuthorId = s_u.Id
	new_family.PerspectiveUserId = s_u.Id // 视角所属用户，默认等于AuthorId
	new_family.Introduction = introduction
	//初始化家庭茶团默认参数
	new_family.HusbandFromFamilyId = data.FamilyIdUnknown
	new_family.WifeFromFamilyId = data.FamilyIdUnknown
	new_family.Logo = "familyLogo"

	//保存到数据库中,返回新家庭茶团的id
	if err := new_family.Create(); err != nil {
		util.Debug(s_u.Email, "Cannot create new family", err)
		report(w, s_u, "你好，茶博士摸摸头，未能创建新茶团，请稍后再试。")
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
		util.Debug(s_u.Email, "Cannot create author family member", err)
		report(w, s_u, "你好，茶博士摸摸头，未能创建新茶团，请稍后再试。")
		return
	}

	//检查会话茶友是否已经设置了默认首选家庭茶团
	//如果已经设置了默认首选家庭茶团，则不再设置
	df, err := s_u.GetLastDefaultFamily()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			df = data.FamilyUnknown
		} else {
			util.Debug(s_u.Id, "Cannot get user's default family", err)
			report(w, s_u, fmt.Sprintf("你好，茶博士摸摸头，竟然说这个用户%s没有默认家庭茶团，未能查看&家庭茶团列表。", s_u.Email))
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
			util.Debug(s_u.Email, "Cannot create user's default family", err)
			report(w, s_u, "你好，茶博士摸摸头，未能创建默认家庭茶团，请稍后再试。")
			return
		}
	}

	//报告用户登记家庭茶团成功
	// text := ""
	// if s_u.Gender == data.User_Gender_Female {
	// 	text = fmt.Sprintf("%s 女士，你好，登记 %s 家庭茶团成功，可以到我的家庭中查看详情，祝愿拥有快乐品茶时光。", s_u.Name, new_family.Name)
	// } else {
	// 	text = fmt.Sprintf("%s 先生，你好，登记 %s 家庭茶团成功，可以到我的家庭中查看详情，祝愿拥有美好品茶时光。", s_u.Name, new_family.Name)
	// }
	// report(w, text)

	//跳转新建的家庭详情页面
	http.Redirect(w, r, "/v1/family/detail?id="+new_family.Uuid, http.StatusFound)
}

// GET /v1/family/new
// 返回一张空白的家庭填写表格（页面）
func NewFamilyGet(w http.ResponseWriter, r *http.Request) {
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

// HandleEditFamily 处理编辑家庭
func HandleEditFamily(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		EditFamilyGet(w, r)
	case http.MethodPost:
		EditFamilyPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /v1/family/edit?id=xxx
func EditFamilyGet(w http.ResponseWriter, r *http.Request) {
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

	family_uuid := r.URL.Query().Get("id")
	family := data.Family{Uuid: family_uuid}
	if err = family.GetByUuid(); err != nil {
		report(w, s_u, "未找到家庭资料")
		return
	}

	isParent, _ := family.IsParentMember(s_u.Id)
	if !isParent {
		report(w, s_u, "只有父母角色可以编辑家庭资料")
		return
	}

	familyBean, err := fetchFamilyBean(family)
	if err != nil {
		report(w, s_u, "获取家庭资料失败")
		return
	}

	parentMembers, _ := family.ParentMembers()
	var parentFamilies []data.Family
	for _, pm := range parentMembers {
		childFamilies, _ := data.ChildMemberFamilies(pm.UserId)
		parentFamilies = append(parentFamilies, childFamilies...)
	}

	type EditData struct {
		SessUser       data.User
		FamilyBean     data.FamilyBean
		ParentFamilies []data.Family
	}

	generateHTML(w, &EditData{s_u, familyBean, parentFamilies}, "layout", "navbar.private", "family.edit")
}

// POST /v1/family/edit
func EditFamilyPost(w http.ResponseWriter, r *http.Request) {
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

	family_uuid := r.PostFormValue("family_id")
	family := data.Family{Uuid: family_uuid}
	if err = family.GetByUuid(); err != nil {
		report(w, s_u, "未找到家庭资料")
		return
	}

	isParent, _ := family.IsParentMember(s_u.Id)
	if !isParent {
		report(w, s_u, "只有父母角色可以编辑家庭资料")
		return
	}

	introduction := r.PostFormValue("introduction")
	lenI := cnStrLen(introduction)
	if lenI < int(util.Config.ThreadMinWord) || lenI > int(util.Config.ThreadMaxWord) {
		report(w, s_u, "家庭简介字数不符合要求")
		return
	}

	status_str := r.PostFormValue("status")
	status, err := strconv.Atoi(status_str)
	if err != nil || status < 0 || status > 5 {
		report(w, s_u, "家庭状态无效")
		return
	}

	has_child := r.PostFormValue("has_child") == "true"
	is_married := r.PostFormValue("is_married") == "true"

	husband_family_id_str := r.PostFormValue("husband_family_id")
	if husband_family_id_str != "" {
		if hfid, err := strconv.Atoi(husband_family_id_str); err == nil {
			family.HusbandFromFamilyId = hfid
		}
	}

	wife_family_id_str := r.PostFormValue("wife_family_id")
	if wife_family_id_str != "" {
		if wfid, err := strconv.Atoi(wife_family_id_str); err == nil {
			family.WifeFromFamilyId = wfid
		}
	}

	family.Introduction = introduction
	family.Status = status
	family.HasChild = has_child
	family.IsMarried = is_married
	family.IsOpen = r.PostFormValue("is_open") == "on"

	if err = family.Update(); err != nil {
		util.Debug("更新家庭资料失败", err)
		report(w, s_u, "保存失败，请稍后再试")
		return
	}

	http.Redirect(w, r, "/v1/family/detail?id="+family.Uuid, http.StatusFound)
}

// GET /v1/family/member_add?uuid=
// 显示家庭搜索用户页面，--Claude sonnet4.5按要求协助创建
func FamilyMemberAddGet(w http.ResponseWriter, r *http.Request) {
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

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		report(w, s_u, "你好，缺少家庭标识。")
		return
	}

	family := data.Family{Uuid: uuid}
	if err = family.GetByUuid(); err != nil {
		util.Debug("Cannot get family by uuid", err)
		report(w, s_u, "你好，未能找到该家庭。")
		return
	}

	// 检查权限：必须是父母角色
	isParent, _ := family.IsParentMember(s_u.Id)
	if !isParent {
		report(w, s_u, "你好，只有父母角色才能添加家庭成员。")
		return
	}

	var pageData struct {
		SessUser data.User
		Family   data.Family
	}
	pageData.SessUser = s_u
	pageData.Family = family

	generateHTML(w, &pageData, "layout", "navbar.private", "family.member_add")
}

// HandleFamilySearchUser POST /v1/family/search_user
// 处理家庭搜索用户请求 -> 返回搜索结果 --Claude sonnet4.5按要求协助创建
func HandleFamilySearchUser(w http.ResponseWriter, r *http.Request) {

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
	err = r.ParseForm()
	if err != nil {
		util.Debug(" Cannot parse form", err)
		report(w, s_u, "你好，茶博士失魂鱼，未能开新茶团，请稍后再试。")
		return
	}
	familyUuid := r.PostFormValue("family_uuid")
	searchType := r.PostFormValue("search_type")
	keyword := r.PostFormValue("keyword")

	// 验证关键词长度
	if len(keyword) < 1 || len(keyword) > 32 {
		report(w, s_u, "你好，茶博士摸摸头，说关键词太长了记不住呢，请确认后再试。")
		return
	}

	// 获取家庭信息
	family := data.Family{Uuid: familyUuid}
	if err = family.GetByUuid(); err != nil {
		util.Debug("Cannot get family by uuid", err)
		report(w, s_u, "你好，未能找到该家庭。")
		return
	}

	// 检查权限
	isParent, _ := family.IsParentMember(s_u.Id)
	if !isParent {
		report(w, s_u, "你好，只有父母角色才能添加家庭成员。")
		return
	}

	var pageData struct {
		SessUser                 data.User
		Family                   data.Family
		UserDefaultDataBeanSlice []data.UserDefaultDataBean
		IsEmpty                  bool
	}
	pageData.SessUser = s_u
	pageData.Family = family
	pageData.IsEmpty = true

	switch searchType {
	case "user_id":
		// 按茶友号查询
		userId, err := strconv.Atoi(keyword)
		if err != nil || userId <= 0 {
			report(w, s_u, "茶友号必须是正整数")
			return
		}

		user, err := data.GetUser(userId)
		if err == nil && user.Id > 0 {
			userBean, err := fetchUserDefaultBean(user)
			if err == nil {
				pageData.UserDefaultDataBeanSlice = append(pageData.UserDefaultDataBeanSlice, userBean)
				pageData.IsEmpty = false
			}
		}

	case "user_email":
		// 按邮箱查询
		if !isEmail(keyword) {
			report(w, s_u, "你好，请输入有效的电子邮箱地址。")
			return
		}

		user, err := data.GetUserByEmail(keyword, r.Context())
		if err == nil && user.Id > 0 {
			userBean, err := fetchUserDefaultBean(user)
			if err == nil {
				pageData.UserDefaultDataBeanSlice = append(pageData.UserDefaultDataBeanSlice, userBean)
				pageData.IsEmpty = false
			}
		}

	case "user_name":
		// 按花名查询
		userSlice, err := data.SearchUserByNameKeyword(keyword, int(util.Config.DefaultSearchResultNum), r.Context())
		if err == nil && len(userSlice) >= 1 {
			userBeanSlice, err := fetchUserDefaultDataBeanSlice(userSlice)
			if err == nil && len(userBeanSlice) >= 1 {
				pageData.UserDefaultDataBeanSlice = userBeanSlice
				pageData.IsEmpty = false
			}
		}

	default:
		report(w, s_u, "你好，请选择正确的查询方式。")
		return
	}

	generateHTML(w, &pageData, "layout", "navbar.private", "family.search_user_result", "component_avatar_name_gender")
}
