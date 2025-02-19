package data

import "time"

// 好东西，举办茶话会活动需要的物资。包括装备、物品和材料等.
// 通常是可以交易的商品，例如车辆、工具、食物、耗材...
type Goods struct {
	Id                    int        //postgreSQL serial主键
	Uuid                  string     // 全局唯一标识符，用于在分布式系统中唯一标识物资
	RecorderUserId        int        // 物资登记人ID
	Name                  string     // 名称,通常是官方正式名称。
	Nickname              string     // 昵称或别名，便于用户记忆和使用。
	Designer              string     // 设计者，适用于有设计元素的物资。
	Describe              string     // 详细描述，帮助用户了解物资的具体信息。
	Price                 float64    // 价格，通常以货币单位表示。
	Applicability         string     // 用途，描述物资的使用场景。
	Category              int        // 类别：0-虚拟（不受重力影响）；1-物体（受重力影响）
	Specification         string     // 标准规格,如技术参数等。
	BrandName             string     // 品牌名称
	Model                 string     // 型号
	Weight                float64    // 重量
	Dimensions            string     // 尺寸
	Material              string     // 材质，适用于物理物品。
	Size                  string     // 物资的大小，适用于物理物品。
	Color                 string     // 物资的颜色或款式。
	NetworkConnectionType string     // 网络连接类型，适用于需要联网的设备。可以考虑使用枚举类型，如 WiFi, Bluetooth, Ethernet 等。
	Features              int        // 物资的特点，0表示可以买卖，1表示不可交易（例如象牙，人体器官？）。
	SerialNumber          string     // 物资的序列号，用于唯一标识每个物资。
	ProductionDate        *time.Time // 生产日期，适用于有生产日期的物资。
	ExpirationDate        *time.Time // 到期日期，适用于有有效期的物资。
	State                 string     // 物资的新旧程度，如全新、二手等。
	Origin                string     // 原产地
	Manufacturer          string     // 生产商名称。
	ManufacturerURL       string     // 制造商的官方网站或链接。
	EngineType            string     // 动力类型，适用于有动力系统的物资。锂离子电池驱动？燃油内燃机？市电？
	PurchaseURL           string     // 网购链接，方便用户直接购买。
	CreatedTime           time.Time  // 物资记录的创建时间。
	UpdatedTime           time.Time  // 物资记录的最后更新时间。
}

// Goods.Create() 保存1物资记录，postgreSQL,用queryRow方法存入goods表，返回id,uuid,
func (g *Goods) Create() (err error) {
	statement := "INSERT INTO goods (uuid, recorder_user_id, name, nickname, designer, describe, price, applicability, category, specification, brand_name, model, weight, dimensions, material, size, color, network_connection_type, features, serial_number, production_date, expiration_date, state, origin, manufacturer, manufacturer_url, engine_type, purchase_url, created_time, updated_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), g.RecorderUserId, g.Name, g.Nickname, g.Designer, g.Describe, g.Price, g.Applicability, g.Category, g.Specification, g.BrandName, g.Model, g.Weight, g.Dimensions, g.Material, g.Size, g.Color, g.NetworkConnectionType, g.Features, g.SerialNumber, g.ProductionDate, g.ExpirationDate, g.State, g.Origin, g.Manufacturer, g.ManufacturerURL, g.EngineType, g.PurchaseURL, time.Now(), time.Now()).Scan(&g.Id, &g.Uuid)
	return
}

// Goods.GetById() 根据id获取1物资记录
func (g *Goods) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, recorder_user_id, name, nickname, designer, describe, price, applicability, category, specification, brand_name, model, weight, dimensions, material, size, color, network_connection_type, features, serial_number, production_date, expiration_date, state, origin, manufacturer, manufacturer_url, engine_type, purchase_url, created_time, updated_time FROM goods WHERE id = $1", g.Id).
		Scan(&g.Id, &g.Uuid, &g.RecorderUserId, &g.Name, &g.Nickname, &g.Designer, &g.Describe, &g.Price, &g.Applicability, &g.Category, &g.Specification, &g.BrandName, &g.Model, &g.Weight, &g.Dimensions, &g.Material, &g.Size, &g.Color, &g.NetworkConnectionType, &g.Features, &g.SerialNumber, &g.ProductionDate, &g.ExpirationDate, &g.State, &g.Origin, &g.Manufacturer, &g.ManufacturerURL, &g.EngineType, &g.PurchaseURL, &g.CreatedTime, &g.UpdatedTime)
	return
}

// Goods.GetByUuid() 根据uuid获取1物资记录
func (g *Goods) GetByUuid() (err error) {
	err = Db.QueryRow("SELECT id, uuid, recorder_user_id, name, nickname, designer, describe, price, applicability, category, specification, brand_name, model, weight, dimensions, material, size, color, network_connection_type, features, serial_number, production_date, expiration_date, state, origin, manufacturer, manufacturer_url, engine_type, purchase_url, created_time, updated_time FROM goods WHERE uuid = $1", g.Uuid).
		Scan(&g.Id, &g.Uuid, &g.RecorderUserId, &g.Name, &g.Nickname, &g.Designer, &g.Describe, &g.Price, &g.Applicability, &g.Category, &g.Specification, &g.BrandName, &g.Model, &g.Weight, &g.Dimensions, &g.Material, &g.Size, &g.Color, &g.NetworkConnectionType, &g.Features, &g.SerialNumber, &g.ProductionDate, &g.ExpirationDate, &g.State, &g.Origin, &g.Manufacturer, &g.ManufacturerURL, &g.EngineType, &g.PurchaseURL, &g.CreatedTime, &g.UpdatedTime)
	return
}

// Goods.Update() 更新1物资记录
func (g *Goods) Update() (err error) {
	statement := "UPDATE goods SET recorder_user_id = $2, name = $3, nickname = $4, designer = $5, describe = $6, price = $7, applicability = $8, category = $9, specification = $10, brand_name = $11, model = $12, weight = $13, dimensions = $14, material = $15, size = $16, color = $17, network_connection_type = $18, features = $19, serial_number = $20, production_date = $21, expiration_date = $22, state = $23, origin = $24, manufacturer = $25, manufacturer_url = $26, engine_type = $27, purchase_url = $28, updated_time = $29 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(g.Id, g.RecorderUserId, g.Name, g.Nickname, g.Designer, g.Describe, g.Price, g.Applicability, g.Category, g.Specification, g.BrandName, g.Model, g.Weight, g.Dimensions, g.Material, g.Size, g.Color, g.NetworkConnectionType, g.Features, g.SerialNumber, g.ProductionDate, g.ExpirationDate, g.State, g.Origin, g.Manufacturer, g.ManufacturerURL, g.EngineType, g.PurchaseURL, time.Now())
	return
}

// 用户收集的物资清单
type UserGoods struct {
	Id        int
	UserId    int
	GoodsId   int
	CreatedAt time.Time
}

// UserGoods.Create() 保存1用户收集的物资记录
func (ug *UserGoods) Create() (err error) {
	statement := "INSERT INTO user_goods (user_id, goods_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ug.UserId, ug.GoodsId, time.Now()).Scan(&ug.Id)
	return
}

// UserGoods.Get() 获取1用户收集的物资记录
func (ug *UserGoods) Get() (err error) {
	err = Db.QueryRow("SELECT id, user_id, goods_id, created_at FROM user_goods WHERE id = $1", ug.Id).
		Scan(&ug.Id, &ug.UserId, &ug.GoodsId, &ug.CreatedAt)
	return
}

// UserGoods.GetAllByUserId()  获取用户收集的所有物资记录
func (ug *UserGoods) GetAllByUserId() (userGoodsSlice []UserGoods, err error) {
	rows, err := Db.Query("SELECT id, user_id, goods_id, created_at FROM user_goods WHERE user_id = $1", ug.UserId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var userGoods UserGoods
		err = rows.Scan(&userGoods.Id, &userGoods.UserId, &userGoods.GoodsId, &userGoods.CreatedAt)
		if err != nil {
			return
		}
		userGoodsSlice = append(userGoodsSlice, userGoods)
	}
	return
}

// UserGoods.CountByUserId()  获取用户收集的物资记录数量
func (ug *UserGoods) CountByUserId() (count int, err error) {
	err = Db.QueryRow("SELECT COUNT(*) FROM user_goods WHERE user_id = $1", ug.UserId).Scan(&count)
	return
}
