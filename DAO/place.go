package data

import "time"

// 地方，洞天福地，茶话会举办的空间。
// 可能是一个建筑物大楼、大厅，房间 ...也可能是行驶中的机舱，船舱,还可能是野外的一块草地，一个帐篷。 maybe a cave like Water Curtain Cave
type Place struct {
	Id             int
	Uuid           string
	Name           string // 名称 ，如：《红楼梦》大观园里的怡红院
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

// place.CountByUser() 统计某个用户登记的地方总数量
func CountPlaceByUserId(user_id int) (count int, err error) {
	err = Db.QueryRow("SELECT COUNT(*) FROM places WHERE user_id = $1", user_id).Scan(&count)
	return
}

// 一个地方可以有多个地址
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

// 某个用户设置的“默认地方”
type UserDefaultPlace struct {
	Id        int
	UserId    int
	PlaceId   int
	CreatedAt time.Time
}

// userPlace 用户绑定的地方
type UserPlace struct {
	Id        int
	UserId    int
	PlaceId   int
	CreatedAt time.Time
}

// UserPlace.Create() 创建用户绑定的地方,返回id
func (up *UserPlace) Create() (err error) {
	statement := "INSERT INTO user_place (user_id, place_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(up.UserId, up.PlaceId, time.Now()).Scan(&up.Id)
	return
}

// UserPlace.GetByUserId() 根据用户ID获取某个用户绑定的全部地方
func (up *UserPlace) GetByUserId() (userPlaces []UserPlace, err error) {
	rows, err := Db.Query("SELECT id, user_id, place_id, created_at FROM user_place WHERE user_id = $1", up.UserId)
	if err != nil {
		return
	}
	for rows.Next() {
		var userPlace UserPlace
		err = rows.Scan(&userPlace.Id, &userPlace.UserId, &userPlace.PlaceId, &userPlace.CreatedAt)
		if err != nil {
			return
		}
		userPlaces = append(userPlaces, userPlace)
	}
	rows.Close()
	return
}

// user_place.count() 统计某个用户绑定的地方数量
func (up *UserPlace) Count() (count int) {
	rows, err := Db.Query("SELECT COUNT(*) FROM user_place WHERE user_id = $1", up.UserId)
	if err != nil {
		return
	}
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return
		}
	}
	rows.Close()
	return
}

// UserDefaultPlace.Create() 创建1默认地方
func (udp *UserDefaultPlace) Create() (err error) {
	statement := "INSERT INTO user_default_place (user_id, place_id) VALUES ($1, $2) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(udp.UserId, udp.PlaceId).Scan(&udp.Id)
	return
}

// User.GetLastDefaultPlace() 根据user_id，从user_default_place表中获取最后一条记录，再根据记录中place_id，从places表获取1个place
func (u *User) GetLastDefaultPlace() (place Place, err error) {
	statement := "SELECT id, user_id, place_id, created_at FROM user_default_place WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	var udp UserDefaultPlace
	err = stmt.QueryRow(u.Id).Scan(&udp.Id, &udp.UserId, &udp.PlaceId, &udp.CreatedAt)
	if err != nil {
		return
	}
	place = Place{Id: udp.PlaceId}
	err = place.Get()
	return
}

// user.GetAllBindPlaces() 获取某个用户绑定的全部地方
func (u *User) GetAllBindPlaces() (Places []Place, err error) {
	up := UserPlace{UserId: u.Id}
	userPlaces, err := up.GetByUserId()
	if err != nil {
		return
	}
	for _, userPlace := range userPlaces {
		place := Place{Id: userPlace.PlaceId}
		err = place.Get()
		if err != nil {
			return
		}
		Places = append(Places, place)
	}
	return
}

// 用户关联地址
type UserAddress struct {
	Id        int
	UserId    int
	AddressId int
	CreatedAt time.Time
}

// UserAddress.GetByUserId() 根据用户ID获取全部用户地址
func (ua *UserAddress) GetByUserId() (userAddresses []UserAddress, err error) {
	rows, err := Db.Query("SELECT id, user_id, address_id, created_at FROM user_address WHERE user_id = $1", ua.UserId)
	if err != nil {
		return
	}
	for rows.Next() {
		var userAddress UserAddress
		err = rows.Scan(&userAddress.Id, &userAddress.UserId, &userAddress.AddressId, &userAddress.CreatedAt)
		if err != nil {
			return
		}
		userAddresses = append(userAddresses, userAddress)
	}
	rows.Close()
	return
}

// user.GetDefaultAddress() 获取用户默认地址,return (address Address,err error)
func (u *User) GetDefaultAddress() (address Address, err error) {
	udp := UserDefaultAddress{UserId: u.Id}
	err = udp.Get()
	if err != nil {
		return
	}
	address = Address{Id: udp.AddressId}
	err = address.Get()
	return
}

// user.GetAllAddress() 获取全部用户地址
func (u *User) GetAllAddress() (Addresses []Address, err error) {
	ua := UserAddress{UserId: u.Id}
	userAddresses, err := ua.GetByUserId()
	if err != nil {
		return
	}
	for _, userAddress := range userAddresses {
		address := Address{Id: userAddress.AddressId}
		err = address.Get()
		if err != nil {
			return
		}
		Addresses = append(Addresses, address)
	}
	return
}

// UserAddress.Create() 创建1用户地址
func (ua *UserAddress) Create() (err error) {
	statement := "INSERT INTO user_address (user_id, address_id) VALUES ($1, $2) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ua.UserId, ua.AddressId).Scan(&ua.Id)
	return
}

// UserAddress.Get() 获取1用户地址
func (ua *UserAddress) Get() (err error) {
	err = Db.QueryRow("SELECT id, user_id, address_id, created_at FROM user_address WHERE id = $1", ua.Id).Scan(&ua.Id, &ua.UserId, &ua.AddressId, &ua.CreatedAt)
	return
}

// UserAddress.Delete() 删除1用户地址
func (ua *UserAddress) Delete() (err error) {
	statement := "DELETE FROM user_address WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(ua.Id)
	return
}

// 用户默认地址
type UserDefaultAddress struct {
	Id        int
	UserId    int
	AddressId int
	CreatedAt time.Time
}

// UserDefaultAddress.Create() 创建1用户默认地址
func (uda *UserDefaultAddress) Create() (err error) {
	statement := "INSERT INTO user_default_address (user_id, address_id) VALUES ($1, $2) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(uda.UserId, uda.AddressId).Scan(&uda.Id)
	return
}

// UserDefaultAddress.Get() 获取1用户默认地址
func (uda *UserDefaultAddress) Get() (err error) {
	err = Db.QueryRow("SELECT id, user_id, address_id, created_at FROM user_default_address WHERE id = $1", uda.Id).Scan(&uda.Id, &uda.UserId, &uda.AddressId, &uda.CreatedAt)
	return
}

// LocationHistory.get() 获取1定位历��记录
func (lh *LocationHistory) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, user_id, place_id, time, longitude, latitude, altitude, direction, speed, accuracy, adcode, provider, addr FROM location_history WHERE id = $1", lh.Id).Scan(&lh.Id, &lh.Uuid, &lh.UserId, &lh.PlaceId, &lh.Time, &lh.Longitude, &lh.Latitude, &lh.Altitude, &lh.Direction, &lh.Speed, &lh.Accuracy, &lh.Adcode, &lh.Provider, &lh.Addr)
	return
}

// LocationHistory.create() 创建1定位历史记录
func (lh *LocationHistory) Create() (err error) {
	statement := "INSERT INTO location_history (uuid, user_id, place_id, time, longitude, latitude, altitude, direction, speed, accuracy, adcode, provider, addr) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), lh.UserId, lh.PlaceId, lh.Time, lh.Longitude, lh.Latitude, lh.Altitude, lh.Direction, lh.Speed, lh.Accuracy, lh.Adcode, lh.Provider, lh.Addr).Scan(&lh.Id)
	return
}
func (lh *LocationHistory) GetLocationHistoryByPlaceId() (locationHistory []LocationHistory, err error) {
	rows, err := Db.Query("SELECT id, uuid, user_id, place_id, time, longitude, latitude, altitude, direction, speed, accuracy, adcode, provider, addr FROM location_history WHERE place_id = $1", lh.PlaceId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		lh := LocationHistory{}
		err = rows.Scan(&lh.Id, &lh.Uuid, &lh.UserId, &lh.PlaceId, &lh.Time, &lh.Longitude, &lh.Latitude, &lh.Altitude, &lh.Direction, &lh.Speed, &lh.Accuracy, &lh.Adcode, &lh.Provider, &lh.Addr)
		if err != nil {
			return
		}
		locationHistory = append(locationHistory, lh)
	}
	err = rows.Err()
	return
}

// place.Create() 保存1地方记录，用queryRow方法插入places表，返回id
func (p *Place) Create() (err error) {
	statement := "INSERT INTO places (uuid, name, nickname, description, icon, occupant_user_id, owner_user_id, level, category, is_public, is_government, user_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), p.Name, p.Nickname, p.Description, p.Icon, p.OccupantUserId, p.OwnerUserId, p.Level, p.Category, p.IsPublic, p.IsGovernment, p.UserId, time.Now(), time.Now()).Scan(&p.Id)
	return
}

// place.GetById() 根据id获取1地方记录
func (p *Place) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, name, nickname, description, icon, occupant_user_id, owner_user_id, level, category, is_public, is_government, user_id, created_at, updated_at FROM places WHERE id = $1", p.Id).
		Scan(&p.Id, &p.Uuid, &p.Name, &p.Nickname, &p.Description, &p.Icon, &p.OccupantUserId, &p.OwnerUserId, &p.Level, &p.Category, &p.IsPublic, &p.IsGovernment, &p.UserId, &p.CreatedAt, &p.UpdatedAt)
	return
}

// user.GetAllRecordPlaces 根据登记者user_id获取全部登记地方
func (u *User) GetAllRecordPlaces() (places []Place, err error) {
	rows, err := Db.Query("SELECT id, uuid, name, nickname, description, icon, occupant_user_id, owner_user_id, level, category, is_public, is_government, user_id, created_at, updated_at FROM places WHERE user_id = $1", u.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		var place Place
		err = rows.Scan(&place.Id, &place.Uuid, &place.Name, &place.Nickname, &place.Description, &place.Icon, &place.OccupantUserId, &place.OwnerUserId, &place.Level, &place.Category, &place.IsPublic, &place.IsGovernment, &place.UserId, &place.CreatedAt, &place.UpdatedAt)
		if err != nil {
			return
		}
		places = append(places, place)
	}
	rows.Close()
	return
}

// place.GetByUuid() 根据uuid获取1地方记录
func (p *Place) GetByUuid() (err error) {
	err = Db.QueryRow("SELECT id, uuid, name, nickname, description, icon, occupant_user_id, owner_user_id, level, category, is_public, is_government, user_id, created_at, updated_at FROM places WHERE uuid = $1", p.Uuid).
		Scan(&p.Id, &p.Uuid, &p.Name, &p.Nickname, &p.Description, &p.Icon, &p.OccupantUserId, &p.OwnerUserId, &p.Level, &p.Category, &p.IsPublic, &p.IsGovernment, &p.UserId, &p.CreatedAt, &p.UpdatedAt)
	return
}

// place.Update() 更新1地方记录
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

// place.Delete() 删除1地方记录
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

// address.GetById() 根据id获取1地址记录
func (a *Address) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, nation, province, city, district, town, village, street, building, unit, portal_number, postal_code, category, created_at, updated_at FROM addresses WHERE id = $1", a.Id).
		Scan(&a.Id, &a.Uuid, &a.Nation, &a.Province, &a.City, &a.District, &a.Town, &a.Village, &a.Street, &a.Building, &a.Unit, &a.PortalNumber, &a.PostalCode, &a.Category, &a.CreatedAt, &a.UpdatedAt)
	return
}

// address.GetByUuid() 根据uuid获取1地址记录
func (a *Address) GetByUuid() (err error) {
	err = Db.QueryRow("SELECT id, uuid, nation, province, city, district, town, village, street, building, unit, portal_number, postal_code, category, created_at, updated_at FROM addresses WHERE uuid = $1", a.Uuid).
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
