package data

import (
	"database/sql"
	"strconv"
	"time"
)

// Family 家庭喝茶团，原始的社会单位，社区生活构成的基础单元
// 除了美猴王孙悟空是从石头里蹦出来的之外，其他人都是上有父母，中间配偶（可能结偶状态为false），下有子女（可能子女数为0）
type Family struct {
	Id                  int
	Uuid                string
	AuthorId            int    // 创建者id
	Name                string // 家庭名称，默认是“丈夫-妻子”联合名字组合，例如：比尔及梅琳达·盖茨（Bill & Melinda Gates)基金会（Foundation)【命名方法】
	Introduction        string // 家庭简介
	IsMarried           bool   // 是否已结婚？（法律上的领取结婚证）
	HasChild            bool   // 这个家庭是否有子女（包括领养的）？
	HusbandFromFamilyId int    // 丈夫来自的家庭id，如果是0表示未登记,parents' home
	WifeFromFamilyId    int    // 妻子来自的家庭id，in-laws,如果是0表示未登记
	Status              int    // 状态指数，0、保密，1、独身，2、已婚，3、同居，4、分居，5、离异，6、未知
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Logo                string // 家庭标志图片名
}

// Family.GetStatus()
func (f *Family) GetStatus() string {
	switch f.Status {
	case 0:
		return "保密"
	case 1:
		return "单身"
	case 2:
		return "已婚"
	case 3:
		return "同居"
	case 4:
		return "分居"
	case 5:
		return "离异"
	default:
		return "未知"
	}
}

// FamilyMember 家庭成员，包括丈夫、妻子、子女
// 某一个家庭的子女，这里明确要求为未成年的，年龄小于18岁；
// 如果子女已成年（age>18），可以承担民事责任，就同时算是另一个家庭的（单身家庭）成员，
type FamilyMember struct {
	Id               int
	Uuid             string
	FamilyId         int    // 家庭id
	UserId           int    // 茶友id
	Role             int    // 家庭角色，0、未知，1、丈夫，2、妻子，3、女儿， 4、儿子，5、宠物？！
	IsAdult          bool   // 是否成年?
	NickName         string // 父母对孩童时期的昵称，例如：狗剩
	IsAdopted        bool   // 是否被领养?例如：木偶人匹诺曹Pinocchio
	Age              int    // 年龄,如果是0表示未知
	OrderOfSeniority int    // 家中排行老几？孩子的年长先后顺序，1、2、3 ...,如果是0表示未知
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// FamilyMember.GetRole()
func (fm *FamilyMember) GetRole() string {
	switch fm.Role {
	case 0:
		return "未知"
	case 1:
		return "丈夫"
	case 2:
		return "妻子"
	case 3:
		return "女儿"
	case 4:
		return "儿子"
	case 5:
		return "宠物"
	default:
		return "秘密"
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
	statement := "INSERT INTO user_default_families (user_id, family_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(udf.UserId, udf.FamilyId, time.Now()).Scan(&udf.Id)
	if err != nil {
		return
	}
	return
}

// (user *User) GetLastDefaultFamily() 根据user.Id从user_default_families和families，获取用户最后一次设定的“默认家庭”，return (family Family, err error)
func (user *User) GetLastDefaultFamily() (family Family, err error) {
	family = Family{}
	statement := "SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, f.status, f.created_at, f.updated_at, f.logo FROM user_default_families udf LEFT JOIN families f ON udf.family_id = f.id WHERE udf.user_id = $1 ORDER BY udf.created_at DESC LIMIT 1"
	err = Db.QueryRow(statement, user.Id).Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo)
	if err != nil {
		if err == sql.ErrNoRows {
			//如果找不到设置记录，则返回id=0，表示“默认家庭=温暖之家”
			return Family{Id: 0, Uuid: "x", Name: "温暖之家"}, nil
		}
		return
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
	statement := "INSERT INTO families (uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), f.AuthorId, f.Name, f.Introduction, f.IsMarried, f.HasChild, f.HusbandFromFamilyId, f.WifeFromFamilyId, f.Status, time.Now(), time.Now(), f.Logo).Scan(&f.Id)
	if err != nil {
		return
	}
	return
}

// Family.Update() 更新家庭
func (f *Family) Update() (err error) {
	statement := "UPDATE families SET name=$1, introduction=$2, is_married=$3, has_child=$4, husband_from_family_id=$5, wife_from_family_id=$6, status=$7, updated_at=$8, logo=$9 WHERE id=$10"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(f.Name, f.Introduction, f.IsMarried, f.HasChild, f.HusbandFromFamilyId, f.WifeFromFamilyId, f.Status, time.Now(), f.Logo, f.Id)
	if err != nil {
		return
	}
	return
}

// Family.Get() 根据id获取家庭
func (f *Family) Get() (err error) {
	statement := "SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo FROM families WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(f.Id).Scan(&f.Id, &f.Uuid, &f.AuthorId, &f.Name, &f.Introduction, &f.IsMarried, &f.HasChild, &f.HusbandFromFamilyId, &f.WifeFromFamilyId, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.Logo)
	if err != nil {
		return
	}
	return
}

// GetFamiliesByAuthorId() 根据author_id获取家庭列表
func GetFamiliesByAuthorId(authorId int) (families []Family, err error) {
	rows, err := Db.Query("SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo FROM families WHERE author_id=$1", authorId)
	if err != nil {
		return
	}
	for rows.Next() {
		family := Family{}
		err = rows.Scan(&family.Id, &family.Uuid, &family.AuthorId, &family.Name, &family.Introduction, &family.IsMarried, &family.HasChild, &family.HusbandFromFamilyId, &family.WifeFromFamilyId, &family.Status, &family.CreatedAt, &family.UpdatedAt, &family.Logo)
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
	statement := "SELECT id, uuid, author_id, name, introduction, is_married, has_child, husband_from_family_id, wife_from_family_id, status, created_at, updated_at, logo FROM families WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(f.Uuid).Scan(&f.Id, &f.Uuid, &f.AuthorId, &f.Name, &f.Introduction, &f.IsMarried, &f.HasChild, &f.HusbandFromFamilyId, &f.WifeFromFamilyId, &f.Status, &f.CreatedAt, &f.UpdatedAt, &f.Logo)
	if err != nil {
		return
	}
	return
}

// Family.Founder() 获取家庭创建者
func (f *Family) Founder() (user User, err error) {
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", f.AuthorId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Family.IsMember() 是否是家庭成员
func (f *Family) IsMember(user_id int) (isMember bool, err error) {
	statement := "SELECT COUNT(*) FROM family_members WHERE family_id=$1 AND user_id=$2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(f.Id, user_id).Scan(&count)
	if err != nil {
		return
	}
	isMember = count > 0
	return
}

// Family.GetLogo()
func (f *Family) GetLogo() string {
	if f.Logo == "" {
		return "familyLogo.jpeg"
	}
	return f.Logo
}

// FamilyMember.Create() 创建家庭成员
func (fm *FamilyMember) Create() (err error) {
	statement := "INSERT INTO family_members (uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), fm.FamilyId, fm.UserId, fm.Role, fm.IsAdult, fm.NickName, fm.IsAdopted, fm.Age, fm.OrderOfSeniority, time.Now(), time.Now()).Scan(&fm.Id)
	if err != nil {
		return
	}
	return
}

// FamilyMember.GetById() 根据id获取家庭成员
func (fm *FamilyMember) GetById() (err error) {
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

// FamilyMember.GetByUserId() 根据user_id获取家庭成员
func (fm *FamilyMember) GetByUserId() (err error) {
	statement := "SELECT id, uuid, family_id, user_id, role, is_adult, nick_name, is_adopted, age, order_of_seniority, created_at, updated_at FROM family_members WHERE user_id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(fm.UserId).Scan(&fm.Id, &fm.Uuid, &fm.FamilyId, &fm.UserId, &fm.Role, &fm.IsAdult, &fm.NickName, &fm.IsAdopted, &fm.Age, &fm.OrderOfSeniority, &fm.CreatedAt, &fm.UpdatedAt)
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
