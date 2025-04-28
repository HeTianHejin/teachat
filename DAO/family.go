package data

import (
	"strconv"
	"time"
)

// Family 家庭，原始的社会单位，社区生活构成的基础单元，
// 一个狭义的家庭是由男主人（hero）、女主人公（heroine）、「或有」未成年儿子、女儿和宠物组成，
// 已成年的儿子和女儿，成年后算为自动分家离开，计算为另一个（未来）家庭的成员，
// 除了美猴王孙悟空是从石头里蹦出来的之外，其他人都是上有父母（上一代家庭），中间兄弟姐妹/配偶（可能结偶状态为false），下有子女（可能子女数为0）（下一代家庭），
// 上中下亲缘家庭集合为大家族。
type Family struct {
	Id                  int
	Uuid                string
	AuthorId            int    // 创建者id
	Name                string // 家庭名称，默认是“丈夫&妻子”联合名字组合，例如：比尔及梅琳达·盖茨（Bill & Melinda Gates)基金会（Foundation)【命名方法】
	Introduction        string // 家庭简介
	IsMarried           bool   // 是否已结婚？（法律上的领取结婚证）
	HasChild            bool   // 这个家庭是否有子女（包括领养的）？
	HusbandFromFamilyId int    // 丈夫来自的家庭id，如果是0表示未登记,parents' home
	WifeFromFamilyId    int    // 妻子来自的家庭id，in-laws,如果是0表示未登记
	Status              int    // 状态指数，0、保密，1、单身，2、同居，3、已婚，4、分居，5、离婚，其他、未知
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	Logo                string // 家庭标志图片名
	IsOpen              bool   // 是否公开，公开的家庭可以被搜索到，不公开的家庭不可以被搜索到
}

// Family.GetStatus()
func (f *Family) GetStatus() string {
	switch f.Status {
	case 0:
		return "保密"
	case 1:
		return "单身"
	case 2:
		return "同居"
	case 3:
		return "已婚"
	case 4:
		return "分居"
	case 5:
		return "离婚"
	default:
		return "未知"
	}
}

// FamilyMember 家庭成员，包括男主人公（hero）、女主人公（heroine）、儿子、女儿
// 某一个家庭的子女，这里明确要求为未成年的，年龄小于18岁；
// 如果子女已成年（age>18），可以承担民事责任，就同时算是另一个家庭的（单身家庭）成员，
type FamilyMember struct {
	Id               int
	Uuid             string
	FamilyId         int    // 家庭id
	UserId           int    // 茶友id
	Role             int    // 家庭角色，0、秘密，1、男主人公，2、女主人公，3、女儿， 4、儿子，5、宠物,
	IsAdult          bool   // 是否成年?
	NickName         string // 父母对孩童时期的昵称，例如：狗剩
	IsAdopted        bool   // 是否被领养?例如：木偶人匹诺曹Pinocchio
	Age              int    // 年龄,如果是0表示未知
	OrderOfSeniority int    // 家中排行老几？孩子的年长先后顺序，1、2、3 ...,如果是0表示未知
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

// FamilyMember.GetRole()
func (fm *FamilyMember) GetRole() string {
	switch fm.Role {
	case 0:
		return "秘密"
	case 1:
		return "男主人"
	case 2:
		return "女主人"
	case 3:
		return "女儿"
	case 4:
		return "儿子"
	case 5:
		return "宠物"

	default:
		return "未知"
	}
}

// 增加&家庭茶团成员声明
type FamilyMemberSignIn struct {
	Id           int
	Uuid         string
	FamilyId     int    //“家庭茶团成员声明”所指向的&家庭茶团id
	UserId       int    //被声明为新成员的茶友id
	Role         int    // 家庭成员角色：0、秘密，1、男主人公，2、女主人公，3、女儿， 4、儿子，5、宠物,
	IsAdult      bool   //是否成年
	Title        string //标题
	Content      string //声明内容
	PlaceId      int    //“家庭茶团成员声明”所指向的地点id
	Status       int    //状态：0、未读，1、已读， 2、已确认， 3、已否认，
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	IsAdopted    bool //是否领养
	AuthorUserId int  //声明书作者茶友id
}

// 离开&家庭茶团成员声明
type FamilyMemberSignOut struct {
	Id           int
	Uuid         string
	FamilyId     int    //“家庭茶团成员声明”所指向的&家庭茶团id
	UserId       int    //被声明为离开成员的茶友id
	Role         int    //家庭成员角色：0、秘密，1、男主人公，2、女主人公，3、女儿， 4、儿子，5、宠物,
	IsAdult      bool   //是否成年
	Title        string //标题
	Content      string //声明内容
	PlaceId      int    //“家庭茶团成员声明”所指向的地点id
	Status       int    //状态：0、未读，1、已读， 2、已确认， 3、已否认，
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	IsAdopted    bool //是否领养
	AuthorUserId int  //声明书作者茶友id
}

// FamilyMemberSignOut.Create() 创建“离开&家庭茶团成员声明”
func (fms *FamilyMemberSignOut) Create() (err error) {
	statement := "INSERT INTO family_member_sign_outs (uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, is_adopted, author_user_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), fms.FamilyId, fms.UserId, fms.Role, fms.IsAdult, fms.Title, fms.Content, fms.PlaceId, fms.Status, time.Now(), fms.IsAdopted, fms.AuthorUserId).Scan(&fms.Id, &fms.Uuid)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignOut.Get()根据“id”获取“离开&家庭茶团成员声明”
func (fms *FamilyMemberSignOut) Get() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_outs WHERE id = $1"
	err = Db.QueryRow(statement, fms.Id).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignOut.GetByFamilyIdUserId() 根据“家庭茶团id”和“茶友id”获取“离开&家庭茶团成员声明”
func (fms *FamilyMemberSignOut) GetByFamilyIdUserId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_outs WHERE family_id = $1 AND user_id = $2"
	err = Db.QueryRow(statement, fms.FamilyId, fms.UserId).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignOut.GetByUuid() 根据“uuid”获取“离开&家庭茶团成员声明”
func (fms *FamilyMemberSignOut) GetByUuid() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_outs WHERE uuid = $1"
	err = Db.QueryRow(statement, fms.Uuid).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignOut.Update() 更新“离开&家庭茶团成员声明”
func (fms *FamilyMemberSignOut) Update() (err error) {
	statement := "UPDATE family_member_sign_outs SET family_id = $2, user_id = $3, role = $4, is_adult = $5, title = $6, content = $7, place_id = $8, status = $9, updated_at = $10, is_adopted = $11, author_user_id = $12 WHERE id = $1"
	_, err = Db.Exec(statement, fms.Id, fms.FamilyId, fms.UserId, fms.Role, fms.IsAdult, fms.Title, fms.Content, fms.PlaceId, fms.Status, time.Now(), fms.IsAdopted, fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignInReply struct 针对“增加&家庭茶团成员声明”的答复
type FamilyMemberSignInReply struct {
	Id        int
	Uuid      string
	SignInId  int  //“家庭茶团成员声明”所指向的&家庭茶团id
	UserId    int  //被声明为新成员的茶友id
	IsConfirm bool //答复结果: true: 已确认，false: 已否认
	CreatedAt time.Time
}

// FamilyMemberSignInReply.Create() 创建“家庭茶团成员声明回复”
func (fmsr *FamilyMemberSignInReply) Create() (err error) {
	statement := "INSERT INTO family_member_sign_in_replies (uuid, sign_in_id, user_id, is_confirm, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), fmsr.SignInId, fmsr.UserId, fmsr.IsConfirm, time.Now()).Scan(&fmsr.Id, &fmsr.Uuid)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignInReply.GetByUuid() 根据uuid获取“家庭茶团成员声明回复”
func (fmsr *FamilyMemberSignInReply) GetByUuid() (err error) {
	statement := "SELECT id, uuid, sign_in_id, user_id, is_confirm, created_at FROM family_member_sign_in_replies WHERE uuid = $1"
	err = Db.QueryRow(statement, fmsr.Uuid).Scan(&fmsr.Id, &fmsr.Uuid, &fmsr.SignInId, &fmsr.UserId, &fmsr.IsConfirm, &fmsr.CreatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignInReply.Get() 根据id获取“家庭茶团成员声明回复”
func (fmsr *FamilyMemberSignInReply) Get() (err error) {
	statement := "SELECT id, uuid, sign_in_id, user_id, is_confirm, created_at FROM family_member_sign_in_replies WHERE id = $1"
	err = Db.QueryRow(statement, fmsr.Id).Scan(&fmsr.Id, &fmsr.Uuid, &fmsr.SignInId, &fmsr.UserId, &fmsr.IsConfirm, &fmsr.CreatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignInReply.GetResult() string{} 如果isconfirn == true,return "已确认" else return “已否认”
func (fmsr *FamilyMemberSignInReply) GetResult() string {
	if fmsr.IsConfirm {
		return "已确认"
	} else {
		return "已否认"
	}
}

// FamilyMemberSignIn.Check() 如果status > 1 ,表示已经处理完毕，返回true，表示无需再处理
func (fms *FamilyMemberSignIn) Check() bool {
	return fms.Status > 1
}

// FamilyMemberSignIn.GetRole()
func (fms *FamilyMemberSignIn) GetRole() string {
	switch fms.Role {
	case 0:
		return "秘密"
	case 1:
		return "男主人"
	case 2:
		return "女主人"
	case 3:
		return "女儿"
	case 4:
		return "儿子"
	case 5:
		return "宠物"

	default:
		return "未知"
	}
}

// FamilyMemberSignIn.Create() 创建“家庭茶团成员声明”
func (fms *FamilyMemberSignIn) Create() (err error) {
	statement := "INSERT INTO family_member_sign_ins (uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, is_adopted, author_user_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), fms.FamilyId, fms.UserId, fms.Role, fms.IsAdult, fms.Title, fms.Content, fms.PlaceId, fms.Status, time.Now(), fms.IsAdopted, fms.AuthorUserId).Scan(&fms.Id, &fms.Uuid)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignIn.Get() 根据id获取“家庭茶团成员声明”
func (fms *FamilyMemberSignIn) Get() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_ins WHERE id = $1"
	err = Db.QueryRow(statement, fms.Id).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}
func (fms *FamilyMemberSignIn) Update() (err error) {
	statement := "UPDATE family_member_sign_ins SET role = $2, is_adult = $3, title = $4, content = $5, place_id=$6, status = $7, updated_at = $8 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(fms.Id, fms.Role, fms.IsAdult, fms.Title, fms.Content, fms.PlaceId, fms.Status, time.Now())
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignIn.GetByUUID() 根据uuid获取“家庭茶团成员声明”
func (fms *FamilyMemberSignIn) GetByUuid() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_ins WHERE uuid = $1"
	err = Db.QueryRow(statement, fms.Uuid).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignIn.GetByFamilyIdMemberUserId() 根据family_id和user_id获取“家庭茶团成员声明”
func (fms *FamilyMemberSignIn) GetByFamilyIdMemberUserId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_ins WHERE family_id = $1 AND user_id = $2"
	err = Db.QueryRow(statement, fms.FamilyId, fms.UserId).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignIn.CreatedAtDate() 根据创建时间获取日期
func (fms *FamilyMemberSignIn) CreatedAtDate() string {
	return fms.CreatedAt.Format("2006-01-02")
}

// FamilyMemberSignIn.GetStatus() 获取状态
func (fms *FamilyMemberSignIn) GetStatus() string {
	switch fms.Status {
	case 0:
		return "未读"
	case 1:
		return "已读"
	case 2:
		return "已确认"
	case 3:
		return "已否认"
	default:
		return "未知"
	}
}

// UserDefaultFamily 用户的“默认家庭”设置记录
type UserDefaultFamily struct {
	Id        int
	UserId    int
	FamilyId  int
	CreatedAt time.Time
}

// UserDefaultFamily.Create() 创建用户的“默认家庭”设置记录
func (udf *UserDefaultFamily) Create() (err error) {
	statement := "INSERT INTO user_default_families (user_id, family_id) VALUES ($1, $2) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(udf.UserId, udf.FamilyId).Scan(&udf.Id)
	if err != nil {
		return
	}
	return
}

// (user *User) GetLastDefaultFamily() 根据user.Id从user_default_families和families，获取用户最后一次设定的“默认家庭”，return (family Family, err error)
func (user *User) GetLastDefaultFamily() (family Family, err error) {
	family = Family{}
	statement := "SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo, f.is_open FROM user_default_families udf LEFT JOIN families f ON udf.family_id = f.id WHERE udf.user_id = $1 ORDER BY udf.created_at DESC"
	err = Db.QueryRow(statement, user.Id).Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
	if err != nil {

		return family, err
	}
	return
}

// ParentMemberFamilies() 用户user担任核心（男、女主人）父母角色的全部&家庭茶团，
// user.id = family_member.user_id,
// family.id = family_member.family_id,
// family_member.role = 1 or 2,
// return (Families []Family, err error)
func ParentMemberFamilies(user_id int) (families []Family, err error) {
	statement := "SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo, f.is_open FROM family_members fm LEFT JOIN families f ON fm.family_id = f.id WHERE fm.user_id = $1 AND (fm.role = 1 OR fm.role = 2) ORDER BY fm.created_at DESC"
	rows, err := Db.Query(statement, user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		family := Family{}
		err = rows.Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
		if err != nil {
			return
		}
		families = append(families, family)
	}
	return
}

// ChildMemberFamilies() 用户担任子女角色的全部&家庭茶团，family_member.role = 3 or 4,return (Families []Family, err error)
func ChildMemberFamilies(user_id int) (families []Family, err error) {
	statement := "SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo, f.is_open FROM family_members fm LEFT JOIN families f ON fm.family_id = f.id WHERE fm.user_id = $1 AND fm.role IN (3,4) ORDER BY fm.created_at DESC"
	rows, err := Db.Query(statement, user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		family := Family{}
		err = rows.Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
		if err != nil {
			return
		}
		families = append(families, family)
	}
	return
}

// OtherMemberFamilies() 用户担任其他角色的全部&家庭茶团，family_member.role = 5,return (Families []Family, err error)
func OtherMemberFamilies(user_id int) (families []Family, err error) {
	statement := "SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo, f.is_open FROM family_members fm LEFT JOIN families f ON fm.family_id = f.id WHERE fm.user_id = $1 AND fm.role = 5 ORDER BY fm.created_at DESC"
	rows, err := Db.Query(statement, user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		family := Family{}
		err = rows.Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
		if err != nil {
			return
		}
		families = append(families, family)
	}
	return
}

// ResignMemberFamilies() 用户声明离开的&家庭茶团，family_member_sign_out.user_id == family_member.user_id ,return (Families []Family, err error)
func ResignMemberFamilies(user_id int) (families []Family, err error) {
	statement := "SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo, f.is_open FROM family_member_sign_outs fmso LEFT JOIN families f ON fmso.family_id = f.id WHERE fmso.user_id = $1 ORDER BY fmso.created_at DESC"
	rows, err := Db.Query(statement, user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		family := Family{}
		err = rows.Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
		if err != nil {
			return
		}
		families = append(families, family)
	}
	return
}

// GetAllAuthorFamilies() 根据user.Id从families，获取用户登记的全部家庭资料，返回 (Families []Family, err error)
func GetAllAuthorFamilies(user_id int) (families []Family, err error) {
	//families = []Family{}
	statement := "SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open FROM families WHERE author_id = $1 ORDER BY created_at DESC"
	rows, err := Db.Query(statement, user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		family := Family{}
		err = rows.Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
		if err != nil {
			return
		}
		families = append(families, family)
	}
	return
}

// CountAllAuthorFamilies() 统计用户登记的全部家庭数量值
func CountAllAuthorFamilies(user_id int) (count int, err error) {
	statement := "SELECT COUNT(*) FROM families WHERE author_id = $1"
	err = Db.QueryRow(statement, user_id).Scan(&count)
	if err != nil {
		return
	}
	return
}

// CountAllfamilies() 根据family_member.user_id,统计某个茶友是多少个家庭茶团的成员，return count int, err error
func CountAllfamilies(user_id int) (count int, err error) {
	statement := "SELECT COUNT(*) FROM family_members WHERE user_id = $1"
	err = Db.QueryRow(statement, user_id).Scan(&count)
	if err != nil {
		return
	}
	return
}

// GetAllFamilies() 根据family_member.user_id,获取某个茶友是多少个家庭茶团的成员，return (families []Family, err error)
func GetAllFamilies(user_id int) (families []Family, err error) {
	statement := "SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo, f.is_open FROM family_members fm LEFT JOIN families f ON fm.family_id = f.id WHERE fm.user_id = $1 ORDER BY f.created_at DESC"
	rows, err := Db.Query(statement, user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		family := Family{}
		err = rows.Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
		if err != nil {
			return
		}
		families = append(families, family)
	}
	return
}

// FamilyMember.GetIsAdult() 获取是否成年
func (fm *FamilyMember) GetIsAdult() string {
	if fm.IsAdult {
		return "成年"
	}
	return "未成年"
}

// FamilyMember.GetAge() 获取年龄
func (fm *FamilyMember) GetAge() string {
	if fm.Age == 0 {
		return "未知"
	}
	return strconv.Itoa(fm.Age)
}

// FamilyMember.GetSeniority() 获取出生顺序号，排行老几
func (fm *FamilyMember) GetSeniority() string {
	if fm.OrderOfSeniority == 0 {
		return "未知"
	}
	return strconv.Itoa(fm.OrderOfSeniority)
}

// Family.CreatedAtDate() 创建日期
func (f *Family) CreatedAtDate() string {
	return f.CreatedAt.Format("2006-01-02")
}

// Family.Create() 创建家庭
func (f *Family) Create() (err error) {
	statement := "INSERT INTO families (uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, logo, is_open) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), f.AuthorId, f.Name, f.Introduction, f.IsMarried, f.HasChild, f.HusbandFromFamilyId, f.WifeFromFamilyId, f.Status, time.Now(), f.Logo, f.IsOpen).Scan(&f.Id, &f.Uuid)
	if err != nil {
		return
	}
	return
}

// Family.Update() 更新家庭
func (f *Family) Update() (err error) {
	statement := "UPDATE families SET name=$1, introduction=$2, is_married=$3, has_child=$4, husband_from_family_id=$5, wife_from_family_id=$6, status=$7, updated_at=$8, logo=$9, is_open=$10 WHERE id=$11"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(f.Name, f.Introduction, f.IsMarried, f.HasChild, f.HusbandFromFamilyId, f.WifeFromFamilyId, f.Status, time.Now(), f.Logo, f.IsOpen, f.Id)
	if err != nil {
		return
	}
	return
}

// Family.Get() 根据id获取家庭
func (f *Family) Get() (err error) {
	statement := "SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open FROM families WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(f.Id).Scan(&f.Id, &f.Uuid, &f.AuthorId, &f.Name, &f.Introduction, &f.IsMarried, &f.HasChild, &f.HusbandFromFamilyId, &f.WifeFromFamilyId, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.Logo, &f.IsOpen)
	if err != nil {
		return
	}
	return
}

// GetFamily(id int)
func GetFamily(id int) (family Family, err error) {
	err = Db.QueryRow("SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open FROM families WHERE id = $1", id).
		Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
	return
}

// GetFamiliesByAuthorId() 根据author_id获取家庭列表
func GetFamiliesByAuthorId(authorId int) (families []Family, err error) {
	rows, err := Db.Query("SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open FROM families WHERE author_id=$1", authorId)
	if err != nil {
		return
	}
	for rows.Next() {
		family := Family{}
		err = rows.Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen)
		if err != nil {
			return
		}
		families = append(families, family)
	}
	rows.Close()
	return
}

// Family.GetByUuid() 根据uuid获取家庭
func (f *Family) GetByUuid() (err error) {
	statement := "SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open FROM families WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(f.Uuid).Scan(&f.Id, &f.Uuid, &f.AuthorId, &f.Name, &f.Introduction, &f.IsMarried, &f.HasChild, &f.HusbandFromFamilyId, &f.WifeFromFamilyId, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.Logo, &f.IsOpen)
	if err != nil {
		return
	}
	return
}

// Family.Founder() 获取家庭登记者
func (f *Family) Founder() (user User, err error) {
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", f.AuthorId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Family.IsMember() 根据id，检查用户是否是家庭成员
func (f *Family) IsMember(user_id int) (isMember bool, err error) {

	statement := "SELECT COUNT(*) FROM family_members WHERE family_id=$1 AND user_id=$2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(f.Id, user_id).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Family.AllMembers() 获取家庭所有成员
func (f *Family) AllMembers() (members []FamilyMember, err error) {
	rows, err := Db.Query("SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at FROM family_members WHERE family_id=$1", f.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		member := FamilyMember{}
		err = rows.Scan(&member.Id, &member.Uuid, &member.FamilyId, &member.UserId, &member.Role, &member.IsAdult, &member.NickName, &member.IsAdopted, &member.Age, &member.OrderOfSeniority, &member.CreatedAt, &member.UpdatedAt)
		if err != nil {
			return
		}
		members = append(members, member)
	}
	rows.Close()
	return
}

// Family.GetAllMembersUserIdsByFamilyId() 获取家庭所有成员的UserId切片
func GetAllMembersUserIdsByFamilyId(family_id int) (userIds []int, err error) {
	rows, err := Db.Query("SELECT user_id FROM family_members WHERE family_id=$1", family_id)
	if err != nil {
		return
	}
	for rows.Next() {
		var userId int
		err = rows.Scan(&userId)
		if err != nil {
			return
		}
		userIds = append(userIds, userId)
	}
	rows.Close()
	return
}

// IsFamilyExist(user_id, partner_user_id int)  在family_members表里，是否存在同一个family_id，family_member.user_id和partner_user_id是同一家庭成员，而且role=1 or 2，
// 返回 exist bool, err error
func IsFamilyExist(user_id, partner_user_id int) (exist bool, err error) {
	statement := "SELECT COUNT(*) FROM family_members WHERE (user_id=$1 AND role IN (1, 2)) AND family_id IN (SELECT family_id FROM family_members WHERE user_id=$2 AND role IN (1, 2))"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(user_id, partner_user_id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FamilyMember.Create() 创建家庭成员
func (fm *FamilyMember) Create() (err error) {
	statement := "INSERT INTO family_members (uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), fm.FamilyId, fm.UserId, fm.Role, fm.IsAdult, fm.NickName, fm.IsAdopted, fm.Age, fm.OrderOfSeniority, time.Now()).Scan(&fm.Id, &fm.Uuid)
	if err != nil {
		return
	}
	return
}

// FamilyMember.Get() 根据id获取家庭成员
func (fm *FamilyMember) Get() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at FROM family_members WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(fm.Id).Scan(&fm.Id, &fm.Uuid, &fm.FamilyId, &fm.UserId, &fm.Role, &fm.IsAdult, &fm.NickName, &fm.IsAdopted, &fm.Age, &fm.OrderOfSeniority, &fm.CreatedAt, &fm.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMember.GetByUserIdFamilyId() 根据user_id获取指定的家庭成员
func (fm *FamilyMember) GetByUserIdFamilyId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at FROM family_members WHERE user_id=$1 and family_id = $2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(fm.UserId, fm.FamilyId).Scan(&fm.Id, &fm.Uuid, &fm.FamilyId, &fm.UserId, &fm.Role, &fm.IsAdult, &fm.NickName, &fm.IsAdopted, &fm.Age, &fm.OrderOfSeniority, &fm.CreatedAt, &fm.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMember.GetByRoleFamilyId() 根据role获取指定的家庭成员
func (fm *FamilyMember) GetByRoleFamilyId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at FROM family_members WHERE role=$1 and family_id = $2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(fm.Role, fm.FamilyId).Scan(&fm.Id, &fm.Uuid, &fm.FamilyId, &fm.UserId, &fm.Role, &fm.IsAdult, &fm.NickName, &fm.IsAdopted, &fm.Age, &fm.OrderOfSeniority, &fm.CreatedAt, &fm.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMember.ParentMember() 获取家庭成员的父母成员,return parentMembers []FamilyMember,err error
func (f *Family) ParentMembers() (parent_members []FamilyMember, err error) {
	rows, err := Db.Query("SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at FROM family_members WHERE family_id=$1 AND role IN (1, 2)", f.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		parentMember := FamilyMember{}
		err = rows.Scan(&parentMember.Id, &parentMember.Uuid, &parentMember.FamilyId, &parentMember.UserId, &parentMember.Role, &parentMember.IsAdult, &parentMember.NickName, &parentMember.IsAdopted, &parentMember.Age, &parentMember.OrderOfSeniority, &parentMember.CreatedAt, &parentMember.UpdatedAt)
		if err != nil {
			return
		}
		parent_members = append(parent_members, parentMember)
	}
	rows.Close()
	return
}

// FamilyMember.ChildMembers() 获取家庭成员的子女成员列表
func (f *Family) ChildMembers() (child_members []FamilyMember, err error) {
	rows, err := Db.Query("SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at FROM family_members WHERE family_id=$1 AND role IN (3, 4)", f.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		childMember := FamilyMember{}
		err = rows.Scan(&childMember.Id, &childMember.Uuid, &childMember.FamilyId, &childMember.UserId, &childMember.Role, &childMember.IsAdult, &childMember.NickName, &childMember.IsAdopted, &childMember.Age, &childMember.OrderOfSeniority, &childMember.CreatedAt, &childMember.UpdatedAt)
		if err != nil {
			return
		}
		child_members = append(child_members, childMember)
	}
	rows.Close()
	return
}

// FamilyMember.OtherMembers() 获取家庭成员的其他成员列表
func (f *Family) OtherMembers() (other_members []FamilyMember, err error) {
	rows, err := Db.Query("SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at FROM family_members WHERE family_id=$1 AND role IN (0, 5)", f.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		otherMember := FamilyMember{}
		err = rows.Scan(&otherMember.Id, &otherMember.Uuid, &otherMember.FamilyId, &otherMember.UserId, &otherMember.Role, &otherMember.IsAdult, &otherMember.NickName, &otherMember.IsAdopted, &otherMember.Age, &otherMember.OrderOfSeniority, &otherMember.CreatedAt, &otherMember.UpdatedAt)
		if err != nil {
			return
		}
		other_members = append(other_members, otherMember)
	}
	rows.Close()
	return
}

// FamilyMember.CreatedAtDate() 登记日期
func (fm *FamilyMember) CreatedAtDate() string {
	return fm.CreatedAt.Format("2006-01-02")
}

// 统计某个家庭的成员数量
func CountFamilyMembers(familyId int) (count int, err error) {
	err = Db.QueryRow("SELECT COUNT(*) FROM family_members WHERE family_id=$1", familyId).Scan(&count)
	if err != nil {
		return
	}
	return
}
