package data

import "time"

// 一个地点可以有多个地址
type PlaceAddress struct {
	PlaceId   int
	AddressId int
}

// 动态经纬度定位数据
type Location struct {
	Time      time.Time
	Longitude float64 // 经度
	Latitude  float64 // 纬度
	Altitude  float64 // 高度
	Direction float64 // 方向
	Speed     float64 // 速度
	Accuracy  float64 // 精度
	Adcode    int     //行政区划代码：由9位阿拉伯数字组成，前两位数字表示省，第三、四位表示市，第五、六位表示县，第七至九位表示乡、镇；六位数则表示明确到区（县）
	Provider  string  // 供应商
	Addr      string  // 定位服务供应商提供地址
}
type LocationHistory struct {
	Id      int
	Uuid    string
	UserId  int
	PlaceId int
	Location
}

// LocationHistory.create() 创建1定位历史记录
func (lh *LocationHistory) Create() (err error) {
	statement := "INSERT INTO location_history (uuid, user_id, place_id, time, longitude, latitude, altitude, direction, speed, accuracy, adcode, provider, addr) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), lh.UserId, lh.PlaceId, lh.Time, lh.Longitude, lh.Latitude, lh.Altitude, lh.Direction, lh.Speed, lh.Accuracy, lh.Adcode, lh.Provider, lh.Addr).Scan(&lh.PlaceId)
	return
}

// 地点，洞天福地，茶话会举办的空间。
// 可能是一个建筑物大楼、大厅，房间 ...也可能是行驶中的机舱，船舱,还可能是野外的一块草地，一个帐篷。 maybe a cave like Water Curtain Cave
type Place struct {
	Id             int
	Uuid           string
	Name           string // 名称 ，如：《红楼梦》里的大观园
	Nickname       string // 别名，如：绛红轩
	Description    string // 地方描述
	Icon           string // 图标
	OccupantUserId int    // 洞主，物业使用者（负责人）ID，如：贾宝玉
	OwnerUserId    int    // 物业产权登记所有者 如：贾政
	Level          int    // 等级： 1：特级（普京的城堡），2：一级（别墅）飞机，3:独栋，4联排，5公寓楼，6保用十年以上亭棚，7保用十年以下棚，8帐篷等临时遮蔽物业
	Category       int    // 类型 ：0:虚拟空间， 1:私人住宅，2:公共建筑空间，3:户外，4:机舱，5:酒店或商业租赁场所，6:野外
	IsPublic       bool   // 是否公开
	IsGovernment   bool   // 是否政府单位
	UserId         int    // 登记者id
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// place.Create() 保存1地点记录，用queryRow方法插入places表，返回id,uuid
func (p *Place) Create() (err error) {
	statement := "INSERT INTO places (uuid, name, nickname, description, icon, occupant_user_id, owner_user_id, level, category, is_public, is_government, user_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(p.Uuid, p.Name, p.Nickname, p.Description, p.Icon, p.OccupantUserId, p.OwnerUserId, p.Level, p.Category, p.IsPublic, p.IsGovernment, p.UserId, time.Now(), time.Now()).Scan(&p.Id)
	return
}

// place.GetById() ��据id获取1地点记录
func (p *Place) GetById() (err error) {
	err = Db.QueryRow("SELECT id, uuid, name, nickname, description, icon, occupant_user_id, owner_user_id, level, category, is_public, is_government, user_id, created_at, updated_at FROM places WHERE id = $1", p.Id).
		Scan(&p.Id, &p.Uuid, &p.Name, &p.Nickname, &p.Description, &p.Icon, &p.OccupantUserId, &p.OwnerUserId, &p.Level, &p.Category, &p.IsPublic, &p.IsGovernment, &p.UserId, &p.CreatedAt, &p.UpdatedAt)
	return
}

// place.Update() 更新1地点记录
func (p *Place) Update() (err error) {
	statement := "UPDATE places SET name = $2, nickname = $3, description = $4, icon = $5, occupant_user_id = $6, owner_user_id = $7, level = $8, category = $9, is_public = $10, is_government = $11, user_id = $12, updated_at = $13 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(p.Id, p.Name, p.Nickname, p.Description, p.Icon, p.OccupantUserId, p.OwnerUserId, p.Level, p.Category, p.IsPublic, p.IsGovernment, p.UserId, time.Now())
	return
}

// place.Delete() 删除1地点记录
func (p *Place) Delete() (err error) {
	statement := "DELETE FROM places WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(p.Id)
	return
}

// 物流快递地址，如大清帝国京都金陵市世袭园区贾府大街1号大观园
type Address struct {
	Id           int
	Uuid         string
	Nation       string // 国家
	Province     string // 省
	City         string // 市
	District     string // 区
	Town         string // 镇/街道
	Village      string // 村/楼盘小区
	Street       string // 道路
	Building     string // 楼栋
	Unit         string // 单元
	PortalNumber string // 门牌号
	PostalCode   string // 邮政编码（邮政部门发布，末端是基层邮局）
	Category     int    // 类别
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// address.Create() 保存1地址记录，用queryRow方法��入addresses表，返回id,
func (a *Address) Create() (err error) {
	statement := "INSERT INTO addresses (uuid, nation, province, city, district, town, village, street, building, unit, portal_number, postal_code, category, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(a.Uuid, a.Nation, a.Province, a.City, a.District, a.Town, a.Village, a.Street, a.Building, a.Unit, a.PortalNumber, a.PostalCode, a.Category, time.Now(), time.Now()).Scan(&a.Id)
	return
}

// address.GetById() ��据id获取1地址记录
func (a *Address) GetById() (err error) {
	err = Db.QueryRow("SELECT id, uuid, nation, province, city, district, town, village, street, building, unit, portal_number, postal_code, category, created_at, updated_at FROM addresses WHERE id = $1", a.Id).
		Scan(&a.Id, &a.Uuid, &a.Nation, &a.Province, &a.City, &a.District, &a.Town, &a.Village, &a.Street, &a.Building, &a.Unit, &a.PortalNumber, &a.PostalCode, &a.Category, &a.CreatedAt, &a.UpdatedAt)
	return
}

// address.Update() 更新1地址记录
func (a *Address) Update() (err error) {
	statement := "UPDATE addresses SET uuid = $2, nation = $3, province = $4, city = $5, district = $6, town = $7, village = $8, street = $9, building = $10, unit = $11, portal_number = $12,  postal_code = $13, category = $14, updated_at = $15 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(a.Id, a.Uuid, a.Nation, a.Province, a.City, a.District, a.Town, a.Village, a.Street, a.Building, a.Unit, a.PortalNumber, a.PostalCode, a.Category, time.Now())
	return
}

// address.Delete() 删除1地址记录
func (a *Address) Delete() (err error) {
	statement := "DELETE FROM addresses WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(a.Id)
	return
}
