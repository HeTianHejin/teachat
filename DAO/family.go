package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Family 家庭，生命传承功能完整的单位，社区构成的基础单元，
// 这里一个家庭是指由1男主人（hero）、1女主人公（heroine）、「或有」未成年儿子、女儿和宠物等成员组成茶团，
// 从构成数值来看，1女1男主和1女儿1男孩结构家庭，是一个标准家庭/理想中位数？
// 已成年的儿子和女儿，成年后算为自动分家离开，计算为另一个（新）家庭的成员，
// 除了美猴王孙悟空是从石头里蹦出来的之外，其他人都是上有父母（上一代家庭），中间兄弟姐妹，如果有子女（下一代家庭），必有配偶（可能结偶状态为false），
// 上中下亲缘家庭集合为大家族，近亲是指连续性三代。
type Family struct {
	Id                  int
	Uuid                string
	AuthorId            int    // 创建者茶友id，权限必须是男主人或者女主人角色
	Name                string // 家庭名称，默认是“丈夫&妻子”联合名字组合，例如：比尔及梅琳达·盖茨（Bill & Melinda Gates)基金会（Foundation)【命名方法】
	Introduction        string // 家庭简介
	IsMarried           bool   // 是否已结婚？（法律上的领取结婚证）
	HasChild            bool   // 这个家庭是否有子女（包括领养的）？
	HusbandFromFamilyId int    // 丈夫来自的家庭id，如果是0表示未登记,parents' home
	WifeFromFamilyId    int    // 妻子来自的家庭id，in-laws,如果是0表示未登记
	Status              int    // 状态指数，0、保密，1、单身，2、同居，3、已婚，4、分居，5、离婚，其他、未知
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	Logo                string     // 家庭标志图片名
	IsOpen              bool       // 是否公开，公开的家庭可以被搜索到，不公开的家庭不可以被搜索到
	DeletedAt           *time.Time // 软删除时间戳，NULL表示未删除
	PerspectiveUserId   int        // 视角所属用户ID，表示这是谁眼中的家庭，等于AuthorId
}

// 未明确家庭资料的茶友，其家庭资料统一虚拟为"四海为家",id=0
// 任何生物人均是来自某个家庭，但是单独的个体，即使成年，属于一个未来家庭的成员之一，不能视为一个家庭。
var FamilyUnknown = Family{
	Id:           FamilyIdUnknown,
	Uuid:         FamilyUuidUnknown,
	Name:         "四海为家",
	AuthorId:     UserId_Captain_Spaceship, //表示系统预设的值
	Introduction: "存在但未明确资料的家庭",
}

// 未知的家庭ID常量，=="四海为家"，家庭ID为0
const FamilyIdUnknown = 0
const FamilyUuidUnknown = "x" //代表未知数

// 家庭状态常量
const (
	FamilyStatusSecret   = 0 // 保密
	FamilyStatusSingle   = 1 // 单身
	FamilyStatusLivingIn = 2 // 同居
	FamilyStatusMarried  = 3 // 已婚
	FamilyStatusSeparate = 4 // 分居
	FamilyStatusDivorced = 5 // 离婚
)

// GetStatus 获取家庭状态文本
func (f *Family) GetStatus() string {
	switch f.Status {
	case FamilyStatusSecret:
		return "保密"
	case FamilyStatusSingle:
		return "单身"
	case FamilyStatusLivingIn:
		return "同居"
	case FamilyStatusMarried:
		return "已婚"
	case FamilyStatusSeparate:
		return "分居"
	case FamilyStatusDivorced:
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
	FamilyId         int        // 家庭id
	UserId           int        // 茶友id
	Role             int        // 家庭角色，0、秘密，1、男主人公，2、女主人公，3、女儿， 4、儿子，5、宠物,
	IsAdult          bool       // 是否成年?
	NickName         string     // 父母对孩童时期的昵称，例如：狗剩
	IsAdopted        bool       // 是否被领养?例如：木偶人匹诺曹Pinocchio
	Birthday         *time.Time // 生日，NULL表示未知
	DeathDate        *time.Time // 忌日，NULL表示在世
	OrderOfSeniority int        // 家中排行老几？孩子的年长先后顺序，1、2、3 ...,如果是0表示未知
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

// 家庭成员角色常量
const (
	FamilyMemberRoleUnknown  = iota // 0、秘密
	FamilyMemberRoleHusband         // 1、男主人公
	FamilyMemberRoleWife            // 2、女主人公
	FamilyMemberRoleDaughter        // 3、女儿
	FamilyMemberRoleSon             // 4、儿子
	FamilyMemberRolePet             // 5、宠物
)

// 成员声明状态常量
const (
	SignInStatusUnread    = 0 // 未读
	SignInStatusRead      = 1 // 已读
	SignInStatusConfirmed = 2 // 已确认
	SignInStatusDenied    = 3 // 已否认
)

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
	AuthorUserId int  //声明书作者茶友id，权限必须是男主人或者女主人角色
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
	stmt, err := DB.Prepare(statement)
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
	err = DB.QueryRow(statement, fms.Id).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignOut.GetByFamilyIdUserId() 根据“家庭茶团id”和“茶友id”获取“离开&家庭茶团成员声明”
func (fms *FamilyMemberSignOut) GetByFamilyIdUserId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_outs WHERE family_id = $1 AND user_id = $2"
	err = DB.QueryRow(statement, fms.FamilyId, fms.UserId).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignOut.GetByUuid() 根据“uuid”获取“离开&家庭茶团成员声明”
func (fms *FamilyMemberSignOut) GetByUuid() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_outs WHERE uuid = $1"
	err = DB.QueryRow(statement, fms.Uuid).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignOut.Update() 更新“离开&家庭茶团成员声明”
func (fms *FamilyMemberSignOut) Update() (err error) {
	statement := "UPDATE family_member_sign_outs SET family_id = $2, user_id = $3, role = $4, is_adult = $5, title = $6, content = $7, place_id = $8, status = $9, updated_at = $10, is_adopted = $11, author_user_id = $12 WHERE id = $1"
	_, err = DB.Exec(statement, fms.Id, fms.FamilyId, fms.UserId, fms.Role, fms.IsAdult, fms.Title, fms.Content, fms.PlaceId, fms.Status, time.Now(), fms.IsAdopted, fms.AuthorUserId)
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
	stmt, err := DB.Prepare(statement)
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
	err = DB.QueryRow(statement, fmsr.Uuid).Scan(&fmsr.Id, &fmsr.Uuid, &fmsr.SignInId, &fmsr.UserId, &fmsr.IsConfirm, &fmsr.CreatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignInReply.Get() 根据id获取“家庭茶团成员声明回复”
func (fmsr *FamilyMemberSignInReply) Get() (err error) {
	statement := "SELECT id, uuid, sign_in_id, user_id, is_confirm, created_at FROM family_member_sign_in_replies WHERE id = $1"
	err = DB.QueryRow(statement, fmsr.Id).Scan(&fmsr.Id, &fmsr.Uuid, &fmsr.SignInId, &fmsr.UserId, &fmsr.IsConfirm, &fmsr.CreatedAt)
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

// Check 检查声明是否已处理
func (fms *FamilyMemberSignIn) Check() bool {
	return fms.Status > SignInStatusRead
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
	stmt, err := DB.Prepare(statement)
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
	err = DB.QueryRow(statement, fms.Id).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}
func (fms *FamilyMemberSignIn) Update() (err error) {
	statement := "UPDATE family_member_sign_ins SET role = $2, is_adult = $3, title = $4, content = $5, place_id=$6, status = $7, updated_at = $8 WHERE id = $1"
	stmt, err := DB.Prepare(statement)
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
	err = DB.QueryRow(statement, fms.Uuid).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignIn.GetByFamilyIdMemberUserId() 根据family_id和user_id获取“家庭茶团成员声明”
func (fms *FamilyMemberSignIn) GetByFamilyIdMemberUserId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, title, content, place_id, status, created_at, updated_at, is_adopted, author_user_id FROM family_member_sign_ins WHERE family_id = $1 AND user_id = $2"
	err = DB.QueryRow(statement, fms.FamilyId, fms.UserId).Scan(&fms.Id, &fms.Uuid, &fms.FamilyId, &fms.UserId, &fms.Role, &fms.IsAdult, &fms.Title, &fms.Content, &fms.PlaceId, &fms.Status, &fms.CreatedAt, &fms.UpdatedAt, &fms.IsAdopted, &fms.AuthorUserId)
	if err != nil {
		return
	}
	return
}

// FamilyMemberSignIn.CreatedAtDate() 根据创建时间获取日期
func (fms *FamilyMemberSignIn) CreatedAtDate() string {
	return fms.CreatedAt.Format("2006-01-02")
}

// GetStatus 获取声明状态文本
func (fms *FamilyMemberSignIn) GetStatus() string {
	switch fms.Status {
	case SignInStatusUnread:
		return "未读"
	case SignInStatusRead:
		return "已读"
	case SignInStatusConfirmed:
		return "已确认"
	case SignInStatusDenied:
		return "已否认"
	default:
		return "未知"
	}
}

// UserDefaultFamily 用户的“默认家庭”设置记录
// 将自动对外展示，所以选中的家庭必须是开放属性（family.IsOpen==true）
type UserDefaultFamily struct {
	Id        int
	UserId    int
	FamilyId  int
	CreatedAt time.Time
}

// UserDefaultFamily.Create() 创建用户的“默认家庭”设置记录
func (udf *UserDefaultFamily) Create() (err error) {
	statement := "INSERT INTO user_default_families (user_id, family_id) VALUES ($1, $2) RETURNING id"
	stmt, err := DB.Prepare(statement)
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
// --DeepSeek优化
func (user *User) GetLastDefaultFamily() (Family, error) {
	const query = `
        SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, 
               f.is_married, f.has_child, f.husband_from_family_id, 
               f.wife_from_family_id, f.status, f.created_at, 
               f.updated_at, f.logo, f.is_open, f.deleted_at, f.perspective_user_id 
        FROM user_default_families udf 
        JOIN families f ON udf.family_id = f.id 
        WHERE udf.user_id = $1 AND f.deleted_at IS NULL 
        ORDER BY udf.created_at DESC 
        LIMIT 1`

	var family Family
	err := DB.QueryRow(query, user.Id).Scan(
		&family.Id, &family.Uuid, &family.AuthorId, &family.Name,
		&family.Introduction, &family.IsMarried, &family.HasChild,
		&family.HusbandFromFamilyId, &family.WifeFromFamilyId,
		&family.Status, &family.CreatedAt, &family.UpdatedAt,
		&family.Logo, &family.IsOpen, &family.DeletedAt, &family.PerspectiveUserId,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Family{}, err
		}
		return Family{}, fmt.Errorf("failed to query default family: %w", err)
	}

	return family, nil
}

// ParentMemberFamilies 用户担任父母角色的全部家庭
func ParentMemberFamilies(user_id int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()
	return queryFamiliesByUserRole(ctx, user_id, []int{FamilyMemberRoleHusband, FamilyMemberRoleWife})
}

// ParentMemberOpenFamilies 用户担任父母角色的公开家庭
func ParentMemberOpenFamilies(user_id int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()
	return queryFamiliesByUserRoleAndOpen(ctx, user_id, []int{FamilyMemberRoleHusband, FamilyMemberRoleWife}, true)
}

// ParentMemberPrivateFamilies 用户担任父母角色的私密家庭
func ParentMemberPrivateFamilies(user_id int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()
	return queryFamiliesByUserRoleAndOpen(ctx, user_id, []int{FamilyMemberRoleHusband, FamilyMemberRoleWife}, false)
}

// ChildMemberFamilies 用户担任子女角色的全部家庭
func ChildMemberFamilies(user_id int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()
	return queryFamiliesByUserRole(ctx, user_id, []int{FamilyMemberRoleDaughter, FamilyMemberRoleSon})
}

// OtherMemberFamilies 用户担任其他角色的全部家庭
func OtherMemberFamilies(user_id int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()
	return queryFamiliesByUserRole(ctx, user_id, []int{FamilyMemberRolePet})
}

// ResignMemberFamilies 用户声明离开的家庭
func ResignMemberFamilies(user_id int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, 
		f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo, f.is_open, f.deleted_at, f.perspective_user_id 
		FROM family_member_sign_outs fmso LEFT JOIN families f ON fmso.family_id = f.id 
		WHERE fmso.user_id = $1 AND f.deleted_at IS NULL ORDER BY fmso.created_at DESC`

	rows, err := DB.QueryContext(ctx, query, user_id)
	if err != nil {
		return nil, wrapError("ResignMemberFamilies", err)
	}
	return scanFamilies(rows)
}

// GetAllAuthorFamilies 获取用户登记的全部家庭
func GetAllAuthorFamilies(user_id int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT id, uuid, author_id, name, introduction, is_married, has_child, 
		husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open, deleted_at, perspective_user_id 
		FROM families WHERE author_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	rows, err := DB.QueryContext(ctx, query, user_id)
	if err != nil {
		return nil, wrapError("GetAllAuthorFamilies", err)
	}
	return scanFamilies(rows)
}

// CountAllAuthorFamilies 统计用户登记的全部家庭数量
func CountAllAuthorFamilies(user_id int) (int, error) {
	ctx, cancel := getContext()
	defer cancel()

	var count int
	err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM families WHERE author_id = $1 AND deleted_at IS NULL", user_id).Scan(&count)
	return count, wrapError("CountAllAuthorFamilies", err)
}

// CountAllfamilies 统计用户所属家庭数量
func CountAllfamilies(user_id int) (int, error) {
	ctx, cancel := getContext()
	defer cancel()

	var count int
	err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM family_members WHERE user_id = $1", user_id).Scan(&count)
	return count, wrapError("CountAllfamilies", err)
}

// 获取用户视角

// GetAllFamilies 获取用户所属的全部家庭
func GetAllFamilies(user_id int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, 
		f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo, f.is_open, f.deleted_at, f.perspective_user_id 
		FROM family_members fm LEFT JOIN families f ON fm.family_id = f.id 
		WHERE fm.user_id = $1 AND f.deleted_at IS NULL ORDER BY f.created_at DESC`

	rows, err := DB.QueryContext(ctx, query, user_id)
	if err != nil {
		return nil, wrapError("GetAllFamilies", err)
	}
	return scanFamilies(rows)
}

// FamilyMember.GetIsAdult() 获取是否成年
func (fm *FamilyMember) GetIsAdult() string {
	if fm.IsAdult {
		return "成年"
	}
	return "未成年"
}

// GetAge 计算年龄
func (fm *FamilyMember) GetAge() int {
	if fm.Birthday == nil {
		return 0
	}
	endDate := time.Now()
	if fm.DeathDate != nil {
		endDate = *fm.DeathDate
	}
	age := endDate.Year() - fm.Birthday.Year()
	if endDate.Month() < fm.Birthday.Month() || (endDate.Month() == fm.Birthday.Month() && endDate.Day() < fm.Birthday.Day()) {
		age--
	}
	return age
}

// GetAgeString 获取年龄字符串
func (fm *FamilyMember) GetAgeString() string {
	age := fm.GetAge()
	if age == 0 {
		return "未知"
	}
	return strconv.Itoa(age)
}

// IsAlive 是否在世
func (fm *FamilyMember) IsAlive() bool {
	return fm.DeathDate == nil
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

// Create 创建单身1人家庭
func (f *Family) Create() error {
	if err := f.Validate(); err != nil {
		return err
	}
	ctx, cancel := getContext()
	defer cancel()

	// PerspectiveUserId默认等于AuthorId
	if f.PerspectiveUserId == 0 {
		f.PerspectiveUserId = f.AuthorId
	}

	query := `INSERT INTO families (uuid, author_id, name, introduction, is_married, has_child, 
		husband_from_family_id, wife_from_family_id, status, created_at, logo, is_open, perspective_user_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id, uuid`

	err := DB.QueryRowContext(ctx, query, Random_UUID(), f.AuthorId, f.Name, f.Introduction,
		f.IsMarried, f.HasChild, f.HusbandFromFamilyId, f.WifeFromFamilyId, f.Status,
		time.Now(), f.Logo, f.IsOpen, f.PerspectiveUserId).Scan(&f.Id, &f.Uuid)

	return wrapError("Family.Create", err)
}

// Family.Update() 更新家庭
func (f *Family) Update() (err error) {
	statement := "UPDATE families SET name=$1, introduction=$2, is_married=$3, has_child=$4, husband_from_family_id=$5, wife_from_family_id=$6, status=$7, updated_at=$8, logo=$9, is_open=$10 WHERE id=$11"
	stmt, err := DB.Prepare(statement)
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
	if f.Id == 0 {
		return fmt.Errorf("family not found with id: %d", f.Id)
	}
	statement := "SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open, deleted_at, perspective_user_id FROM families WHERE id=$1 AND deleted_at IS NULL"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	if err = stmt.QueryRow(f.Id).Scan(&f.Id, &f.Uuid, &f.AuthorId, &f.Name, &f.Introduction, &f.IsMarried, &f.HasChild, &f.HusbandFromFamilyId, &f.WifeFromFamilyId, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.Logo, &f.IsOpen, &f.DeletedAt, &f.PerspectiveUserId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("family not found with id: %d", f.Id)
		}
		return fmt.Errorf("failed to query family: %w", err)
	}
	return
}

// GetFamily retrieves family information by family ID.
// Returns UnknownFamily (with nil error) when family_id is 0.
// Returns error when family not found or database operation fails.
func GetFamily(family_id int) (family Family, err error) {
	if family_id == 0 {
		return FamilyUnknown, nil
	}

	family = Family{Id: family_id}
	if err = family.Get(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Family{}, fmt.Errorf("family not found with id: %d", family_id)
		}
		return Family{}, fmt.Errorf("failed to get family: %w", err)
	}

	return family, nil
}

// GetFamiliesByAuthorId 根据作者ID获取家庭列表
func GetFamiliesByAuthorId(authorId int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT id, uuid, author_id, name, introduction, is_married, has_child, 
		husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open, deleted_at, perspective_user_id 
		FROM families WHERE author_id=$1 AND deleted_at IS NULL`

	rows, err := DB.QueryContext(ctx, query, authorId)
	if err != nil {
		return nil, wrapError("GetFamiliesByAuthorId", err)
	}
	return scanFamilies(rows)
}

// Family.GetByUuid() 根据uuid获取家庭
func (f *Family) GetByUuid() (err error) {
	if f.Uuid == FamilyUuidUnknown {
		*f = FamilyUnknown
		return nil
	}
	statement := "SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open, deleted_at, perspective_user_id FROM families WHERE uuid=$1 AND deleted_at IS NULL"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(f.Uuid).Scan(&f.Id, &f.Uuid, &f.AuthorId, &f.Name, &f.Introduction, &f.IsMarried, &f.HasChild, &f.HusbandFromFamilyId, &f.WifeFromFamilyId, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.Logo, &f.IsOpen, &f.DeletedAt, &f.PerspectiveUserId)
	if err != nil {
		return
	}
	return
}

// Family.Founder() 获取家庭登记者
func (f *Family) Founder() (user User, err error) {
	err = DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", f.AuthorId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Family.IsMember() 根据user_id，检查用户是否是family_id家庭成员
func (f *Family) IsMember(user_id int) (isMember bool, err error) {
	if f.Id == FamilyIdUnknown {
		return false, fmt.Errorf("family not found with id: %d", f.Id)
	}
	statement := "SELECT COUNT(*) FROM family_members WHERE family_id=$1 AND user_id=$2"
	stmt, err := DB.Prepare(statement)
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

// Family.IsParentMember() 根据user_id，检查用户是否是家庭男女主人(父母)成员
func (f *Family) IsParentMember(user_id int) (isMember bool, err error) {
	statement := "SELECT COUNT(*) FROM family_members WHERE family_id=$1 AND user_id=$2 AND role IN ($3, $4)"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(f.Id, user_id, FamilyMemberRoleHusband, FamilyMemberRoleWife).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Family.AllMembers() 获取家庭所有成员
func (f *Family) AllMembers() (members []FamilyMember, err error) {
	rows, err := DB.Query("SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, birthday, death_date, order_of_seniority, created_at, updated_at FROM family_members WHERE family_id=$1", f.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		member := FamilyMember{}
		err = rows.Scan(&member.Id, &member.Uuid, &member.FamilyId, &member.UserId, &member.Role, &member.IsAdult, &member.NickName, &member.IsAdopted, &member.Birthday, &member.DeathDate, &member.OrderOfSeniority, &member.CreatedAt, &member.UpdatedAt)
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
	if family_id == FamilyIdUnknown {
		return nil, fmt.Errorf("family not found with id: %d", family_id)
	}
	rows, err := DB.Query("SELECT user_id FROM family_members WHERE family_id=$1", family_id)
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

// Create 创建家庭成员
func (fm *FamilyMember) Create() error {
	if err := fm.Validate(); err != nil {
		return err
	}
	ctx, cancel := getContext()
	defer cancel()

	query := `INSERT INTO family_members (uuid, family_id, user_id, role, is_adult, nick_name, 
		is_adopted, birthday, death_date, order_of_seniority, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, uuid`

	err := DB.QueryRowContext(ctx, query, Random_UUID(), fm.FamilyId, fm.UserId, fm.Role,
		fm.IsAdult, fm.NickName, fm.IsAdopted, fm.Birthday, fm.DeathDate, fm.OrderOfSeniority, time.Now()).Scan(&fm.Id, &fm.Uuid)

	return wrapError("FamilyMember.Create", err)
}

// FamilyMember.Get() 根据id获取家庭成员
func (fm *FamilyMember) Get() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, birthday, death_date, order_of_seniority, created_at, updated_at FROM family_members WHERE id=$1"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(fm.Id).Scan(&fm.Id, &fm.Uuid, &fm.FamilyId, &fm.UserId, &fm.Role, &fm.IsAdult, &fm.NickName, &fm.IsAdopted, &fm.Birthday, &fm.DeathDate, &fm.OrderOfSeniority, &fm.CreatedAt, &fm.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMember.GetByUserIdFamilyId() 根据user_id获取指定的家庭成员
func (fm *FamilyMember) GetByUserIdFamilyId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, birthday, death_date, order_of_seniority, created_at, updated_at FROM family_members WHERE user_id=$1 and family_id = $2"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(fm.UserId, fm.FamilyId).Scan(&fm.Id, &fm.Uuid, &fm.FamilyId, &fm.UserId, &fm.Role, &fm.IsAdult, &fm.NickName, &fm.IsAdopted, &fm.Birthday, &fm.DeathDate, &fm.OrderOfSeniority, &fm.CreatedAt, &fm.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMember.GetByRoleFamilyId() 根据role获取指定的家庭成员
func (fm *FamilyMember) GetByRoleFamilyId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, birthday, death_date, order_of_seniority, created_at, updated_at FROM family_members WHERE role=$1 and family_id = $2"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(fm.Role, fm.FamilyId).Scan(&fm.Id, &fm.Uuid, &fm.FamilyId, &fm.UserId, &fm.Role, &fm.IsAdult, &fm.NickName, &fm.IsAdopted, &fm.Birthday, &fm.DeathDate, &fm.OrderOfSeniority, &fm.CreatedAt, &fm.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// FamilyMember.ParentMember() 获取家庭成员的父母成员,return parentMembers []FamilyMember,err error
func (f *Family) ParentMembers() (parent_members []FamilyMember, err error) {
	rows, err := DB.Query("SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, birthday, death_date, order_of_seniority, created_at, updated_at FROM family_members WHERE family_id=$1 AND role IN ($2, $3)", f.Id, FamilyMemberRoleHusband, FamilyMemberRoleWife)
	if err != nil {
		return
	}
	for rows.Next() {
		parentMember := FamilyMember{}
		err = rows.Scan(&parentMember.Id, &parentMember.Uuid, &parentMember.FamilyId, &parentMember.UserId, &parentMember.Role, &parentMember.IsAdult, &parentMember.NickName, &parentMember.IsAdopted, &parentMember.Birthday, &parentMember.DeathDate, &parentMember.OrderOfSeniority, &parentMember.CreatedAt, &parentMember.UpdatedAt)
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
	rows, err := DB.Query("SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, birthday, death_date, order_of_seniority, created_at, updated_at FROM family_members WHERE family_id=$1 AND role IN ($2, $3)", f.Id, FamilyMemberRoleDaughter, FamilyMemberRoleSon)
	if err != nil {
		return
	}
	for rows.Next() {
		childMember := FamilyMember{}
		err = rows.Scan(&childMember.Id, &childMember.Uuid, &childMember.FamilyId, &childMember.UserId, &childMember.Role, &childMember.IsAdult, &childMember.NickName, &childMember.IsAdopted, &childMember.Birthday, &childMember.DeathDate, &childMember.OrderOfSeniority, &childMember.CreatedAt, &childMember.UpdatedAt)
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
	rows, err := DB.Query("SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, birthday, death_date, order_of_seniority, created_at, updated_at FROM family_members WHERE family_id=$1 AND role IN ($2, $3)", f.Id, FamilyMemberRoleUnknown, FamilyMemberRolePet)
	if err != nil {
		return
	}
	for rows.Next() {
		otherMember := FamilyMember{}
		err = rows.Scan(&otherMember.Id, &otherMember.Uuid, &otherMember.FamilyId, &otherMember.UserId, &otherMember.Role, &otherMember.IsAdult, &otherMember.NickName, &otherMember.IsAdopted, &otherMember.Birthday, &otherMember.DeathDate, &otherMember.OrderOfSeniority, &otherMember.CreatedAt, &otherMember.UpdatedAt)
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

// FamilyMember.GetByUuid() 根据uuid获取家庭成员
func (fm *FamilyMember) GetByUuid() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, birthday, death_date, order_of_seniority, created_at, updated_at FROM family_members WHERE uuid=$1"
	err = DB.QueryRow(statement, fm.Uuid).Scan(&fm.Id, &fm.Uuid, &fm.FamilyId, &fm.UserId, &fm.Role, &fm.IsAdult, &fm.NickName, &fm.IsAdopted, &fm.Birthday, &fm.DeathDate, &fm.OrderOfSeniority, &fm.CreatedAt, &fm.UpdatedAt)
	return
}

// FamilyMember.UpdateMemberInfo() 更新家庭成员信息
func (fm *FamilyMember) UpdateMemberInfo() (err error) {
	statement := "UPDATE family_members SET nick_name=$1, birthday=$2, death_date=$3, order_of_seniority=$4, updated_at=$5 WHERE id=$6"
	_, err = DB.Exec(statement, fm.NickName, fm.Birthday, fm.DeathDate, fm.OrderOfSeniority, time.Now(), fm.Id)
	return
}

// ParseDate 解析日期字符串
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// CountFamilyMembers 统计家庭总成员数
func CountFamilyMembers(familyId int) (int, error) {
	ctx, cancel := getContext()
	defer cancel()

	var count int
	err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM family_members WHERE family_id=$1", familyId).Scan(&count)
	return count, wrapError("CountFamilyMembers", err)
}

// 统计家庭父母加子女角色成员的数量
func CountFamilyParentAndChildMembers(familyId int, ctx context.Context) (count int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM family_members WHERE family_id=$1 AND role IN ($2, $3, $4, $5)", familyId, FamilyMemberRoleHusband, FamilyMemberRoleWife, FamilyMemberRoleDaughter, FamilyMemberRoleSon).Scan(&count)
	if err != nil {
		return
	}
	return
}

// Family.IsOnlyOneMember() 检查家庭是否只有一个成员
func (f *Family) IsOnlyOneMember() (isOnlyOne bool, err error) {
	count, err := CountFamilyMembers(f.Id)
	if err != nil {
		return
	}
	return count == 1, nil
}

// SoftDelete 软删除家庭
func (f *Family) SoftDelete() error {
	ctx, cancel := getContext()
	defer cancel()

	now := time.Now()
	f.DeletedAt = &now

	query := "UPDATE families SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL"
	result, err := DB.ExecContext(ctx, query, now, now, f.Id)
	if err != nil {
		return wrapError("Family.SoftDelete", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return wrapError("Family.SoftDelete", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("family not found or already deleted: %d", f.Id)
	}

	return nil
}

// Restore 恢复软删除的家庭
func (f *Family) Restore() error {
	ctx, cancel := getContext()
	defer cancel()

	f.DeletedAt = nil
	now := time.Now()
	f.UpdatedAt = &now

	query := "UPDATE families SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL"
	result, err := DB.ExecContext(ctx, query, now, f.Id)
	if err != nil {
		return wrapError("Family.Restore", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return wrapError("Family.Restore", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("family not found or not deleted: %d", f.Id)
	}

	return nil
}

// IsDeleted 检查家庭是否已被软删除
func (f *Family) IsDeleted() bool {
	return f.DeletedAt != nil
}

// GetDeletedFamiliesByAuthorId 获取用户已删除的家庭列表
func GetDeletedFamiliesByAuthorId(authorId int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT id, uuid, author_id, name, introduction, is_married, has_child, 
		husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open, deleted_at, perspective_user_id 
		FROM families WHERE author_id=$1 AND deleted_at IS NOT NULL ORDER BY deleted_at DESC`

	rows, err := DB.QueryContext(ctx, query, authorId)
	if err != nil {
		return nil, wrapError("GetDeletedFamiliesByAuthorId", err)
	}
	return scanFamilies(rows)
}

// GetFamilyIncludingDeleted 获取家庭（包括已删除的）
func GetFamilyIncludingDeleted(family_id int) (family Family, err error) {
	if family_id == 0 {
		return FamilyUnknown, nil
	}

	ctx, cancel := getContext()
	defer cancel()

	query := "SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open, deleted_at, perspective_user_id FROM families WHERE id=$1"
	err = DB.QueryRowContext(ctx, query, family_id).Scan(
		&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction,
		&family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId,
		&family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo, &family.IsOpen, &family.DeletedAt, &family.PerspectiveUserId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Family{}, fmt.Errorf("family not found with id: %d", family_id)
		}
		return Family{}, wrapError("GetFamilyIncludingDeleted", err)
	}

	return family, nil
}

// SearchFamilyByName 根据家庭名称关键词搜索公开的家庭
func SearchFamilyByName(keyword string, limit int, ctx context.Context) ([]Family, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, uuid, author_id, name, introduction, is_married, has_child, 
		husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo, is_open, deleted_at, perspective_user_id 
		FROM families WHERE name LIKE $1 AND deleted_at IS NULL AND is_open = true LIMIT $2`

	rows, err := DB.QueryContext(ctx, query, "%"+keyword+"%", limit)
	if err != nil {
		return nil, wrapError("SearchFamilyByName", err)
	}
	return scanFamilies(rows)
}
