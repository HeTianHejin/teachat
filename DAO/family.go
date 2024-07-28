package data

import "time"

// Family 家庭，原始的社会单位，社区生活构成的基础单元
// 除了美猴王孙悟空是从石头里蹦出来的之外，其他人都是上有父母，中间配偶（可能结偶状态为false），下有子女（可能子女数为0）
type Family struct {
	Id                     int
	Uuid                   string
	Name                   string // 家庭名称，默认是“丈夫-妻子”联合名字组合，例如：比尔及梅琳达·盖茨（Bill & Melinda Gates)基金会（Foundation)【命名方法】
	Introduction           string
	HusbandUserId          int   // 丈夫角色id
	WifeUserId             int   // 妻子角色id
	ChildUserIdSet         []int // 未成年孩子id集合
	HusbandFromFamilyIdSet []int // 丈夫来自的家庭id集合，默认第一个是血缘父母，其次是继父母...
	WifeFromFamilyIdSet    []int // 妻子来自的家庭id集合，默认第一个是血缘父母，其次是继父母...
	Married                bool  // 是否结婚？
	AdoptedChildUserIdSet  []int // 领养的未成年子女id集合
	StateIndex             int   // 状态指数，活跃，失联...
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Logo                   string // 家庭标志图片名
}

// Save() 记录保存一个家庭资料
func (f *Family) Save() error {
	statement := "INSERT INTO families (uuid, name, introduction, husband_user_id, wife_user_id, child_user_id_set, husband_from_family_id_set, wife_from_family_id_set, married, adopted_child_user_id_set, state_index, created_at, updated_at, logo) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(CreateUuid(), f.Name, f.Introduction, f.HusbandUserId, f.WifeUserId, f.ChildUserIdSet, f.HusbandFromFamilyIdSet, f.WifeFromFamilyIdSet, f.Married, f.AdoptedChildUserIdSet, f.StateIndex, f.CreatedAt, f.UpdatedAt, f.Logo)
	if err != nil {
		return err
	}
	return nil
}

// Update() 更新一个家庭资料根据其id
func (f *Family) Update() error {
	statement := "UPDATE families SET name=$2, introduction=$3, husband_user_id=$4, wife_user_id=$5, child_user_id_set=$6, husband_from_family_id_set=$7, wife_from_family_id_set=$8, married=$9, adopted_child_user_id_set=$10, state_index=$11, updated_at=$12, logo=$13 WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(f.Id, f.Name, f.Introduction, f.HusbandUserId, f.WifeUserId, f.ChildUserIdSet, f.HusbandFromFamilyIdSet, f.WifeFromFamilyIdSet, f.Married, f.AdoptedChildUserIdSet, f.StateIndex, f.UpdatedAt, f.Logo)
	if err != nil {
		return err
	}
	return nil
}

// Get() 获取一个家庭资料根据其Id
func (f *Family) Get() error {
	statement := "SELECT id, uuid, name, introduction, husband_user_id, wife_user_id, child_user_id_set, husband_from_family_id_set, wife_from_family_id_set, married, adopted_child_user_id_set, state_index, created_at, updated_at, logo FROM families WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	row := stmt.QueryRow(f.Id)
	err = row.Scan(&f.Id, &f.Uuid, &f.Name, &f.Introduction, &f.HusbandUserId, &f.WifeUserId, &f.ChildUserIdSet, &f.HusbandFromFamilyIdSet, &f.WifeFromFamilyIdSet, &f.Married, &f.AdoptedChildUserIdSet, &f.StateIndex, &f.CreatedAt, &f.UpdatedAt, &f.Logo)
	if err != nil {
		return err
	}
	return nil
}
