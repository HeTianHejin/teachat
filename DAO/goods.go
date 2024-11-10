package data

import "time"

// 好东西，举办茶话会活动需要的物资。包括装备、物品和材料等.通常是可以交易的商品，例如车辆、工具、食物、耗材...
type Goods struct {
	Id                    int
	Uuid                  string
	UserId                int       // 物资登记人ID
	Name                  string    // 名称
	Nickname              string    // 昵称
	Designer              string    // 设计者
	Describe              string    // 描述
	Price                 float64   // 价格
	Applicability         string    // 用途
	Category              int       // 类别：0-虚拟（不受重力影响）；1-物体（受重力影响）
	Specification         string    // 标准规格
	Brandname             string    // 品牌名称
	Model                 string    // 型号
	Weight                string    // 重量
	Dimensions            string    // 尺寸
	Material              string    // 材质
	Size                  string    // 大小
	Color                 string    // 款色
	NetworkConnectionType string    // 网络连接类型
	Features              int       // 特点: 0-可以买卖；1-不可交易（例如象牙，人体器官？）
	SerialNumber          string    // 序列号
	ProductionDate        time.Time // 生产日期
	ExpirationDate        time.Time // 到期日期
	State                 string    // 新旧程度
	Origin                string    // 原产地
	Manufacturer          string    // 生产商
	ManufacturerLink      string    // 制造商链接
	EngineType            string    // 动力类型？锂离子电池驱动？燃油内燃机？市电？
	PurchaseLink          string    // 网购链接
	CreatedTime           time.Time // 创建时间
	UpdatedTime           time.Time // 更新时间
}

// 用户收集的物资清单
type UserGoods struct {
	Id        int
	UserId    int
	GoodsId   int
	CreatedAt time.Time
}

// userGoods.Create()
func (userGoods *UserGoods) Create() (err error) {
	if err = Db.QueryRow("INSERT INTO user_goods(user_id, goods_id) VALUES($1, $2) RETURNING id",
		userGoods.UserId, userGoods.GoodsId).Scan(&userGoods.Id); err != nil {
		return err
	}
	return err
}

// userGoods.Delete()
func (userGoods *UserGoods) Delete() (err error) {
	_, err = Db.Exec("DELETE FROM user_goods WHERE user_id = $1 AND goods_id = $2", userGoods.UserId, userGoods.GoodsId)
	return err
}

// GetGoodsListByUserId from user_goods 根据用户id，读取其收集的全部物资信息清单
func GetGoodsListByUserId(userId int) (goods_list []Goods, err error) {
	rows, err := Db.Query("SELECT goods_id FROM user_goods WHERE user_id = $1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var goods Goods
		//逐一登记物资id
		if err = rows.Scan(&goods.Id); err != nil {
			return nil, err
		}
		//根据ID获取物资信息
		if err = goods.Get(); err != nil {
			return nil, err
		}
		goods_list = append(goods_list, goods)
	}
	return goods_list, err
}

// (thread *Thread) GetAllGoods() return (goods Goods,err error)
func (thread *Thread) GetAllGoods() (goods_list []Goods, err error) {
	thread_goods_list, err := thread.GetThreadGoodsByThreadId()
	if err != nil {
		return nil, err
	}
	for _, v := range thread_goods_list {
		g := Goods{Id: v.GoodsId}
		if err = g.Get(); err != nil {
			return nil, err
		}
		goods_list = append(goods_list, g)
	}
	return goods_list, err
}

// goods.Create()
func (goods *Goods) Create() (err error) {
	if err = Db.QueryRow("INSERT INTO goods(uuid, user_id, name, nickname, designer, describe, price, applicability, category, specification, brand_name, model, weight, dimensions, material, size, color, network_connection_type, features, serial_number, state, origin, manufacturer, manufacturer_link, engine_type, purchase_link) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26) RETURNING id",
		Random_UUID(), goods.UserId, goods.Name, goods.Nickname, goods.Designer, goods.Describe, goods.Price, goods.Applicability, goods.Category, goods.Specification, goods.Brandname, goods.Model, goods.Weight, goods.Dimensions, goods.Material, goods.Size, goods.Color, goods.NetworkConnectionType, goods.Features, goods.SerialNumber, goods.State, goods.Origin, goods.Manufacturer, goods.ManufacturerLink, goods.EngineType, goods.PurchaseLink).Scan(&goods.Id); err != nil {
		return err
	}
	return err
}

// goods.Get()
func (goods *Goods) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, user_id, name, nickname, designer, describe, price, applicability, category, specification, brand_name, model, weight, dimensions, material, size, color, network_connection_type, features, serial_number, production_date, expiration_date, state, origin, manufacturer, manufacturer_link, engine_type, purchase_link, created_time, updated_time FROM goods WHERE id = $1", goods.Id).Scan(&goods.Id, &goods.Uuid, &goods.UserId, &goods.Name, &goods.Nickname, &goods.Designer, &goods.Describe, &goods.Price, &goods.Applicability, &goods.Category, &goods.Specification, &goods.Brandname, &goods.Model, &goods.Weight, &goods.Dimensions, &goods.Material, &goods.Size, &goods.Color, &goods.NetworkConnectionType, &goods.Features, &goods.SerialNumber, &goods.ProductionDate, &goods.ExpirationDate, &goods.State, &goods.Origin, &goods.Manufacturer, &goods.ManufacturerLink, &goods.EngineType, &goods.PurchaseLink, &goods.CreatedTime, &goods.UpdatedTime)
	return err
}

// goods.Update()
func (goods *Goods) Update() (err error) {
	_, err = Db.Exec("UPDATE goods SET uuid = $1, user_id = $2, name = $3, nickname = $4, designer = $5, describe = $6, price = $7, applicability = $8, category = $9, specification = $10, brand_name = $11, model = $12, weight = $13, dimensions = $14, material = $15, size = $16, color = $17, network_connection_type = $18, features = $19, serial_number = $20, production_date = $21, expiration_date = $22, state = $23, origin = $24, manufacturer = $25, manufacturer_link = $26, engine_type = $27, purchase_link = $28, updated_time = $29 WHERE id = $30",
		goods.Uuid, goods.UserId, goods.Name, goods.Nickname, goods.Designer, goods.Describe, goods.Price, goods.Applicability, goods.Category, goods.Specification, goods.Brandname, goods.Model, goods.Weight, goods.Dimensions, goods.Material, goods.Size, goods.Color, goods.NetworkConnectionType, goods.Features, goods.SerialNumber, goods.ProductionDate, goods.ExpirationDate, goods.State, goods.Origin, goods.Manufacturer, goods.ManufacturerLink, goods.EngineType, goods.PurchaseLink, time.Now(), goods.Id)
	return err
}

// goods.Delete()
func (goods *Goods) Delete() (err error) {
	_, err = Db.Exec("DELETE FROM goods WHERE id = $1", goods.Id)
	return err
}

// GetGoodsByUserId() 根据goods.user_id获取goods
func GetGoodsByUserId(userId int) (goodsList []Goods, err error) {
	rows, err := Db.Query("SELECT id, uuid, user_id, name, nickname, designer, describe, price, applicability, category, specification, brand_name, model, weight, dimensions, material, size, color, network_connection_type, features, serial_number, production_date, expiration_date, state, origin, manufacturer, manufacturer_link, engine_type, purchase_link, created_time, updated_time FROM goods WHERE user_id = $1", userId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var goods Goods
		if err = rows.Scan(&goods.Id, &goods.Uuid, &goods.UserId, &goods.Name, &goods.Nickname, &goods.Designer, &goods.Describe, &goods.Price, &goods.Applicability, &goods.Category, &goods.Specification, &goods.Brandname, &goods.Model, &goods.Weight, &goods.Dimensions, &goods.Material, &goods.Size, &goods.Color, &goods.NetworkConnectionType, &goods.Features, &goods.SerialNumber, &goods.ProductionDate, &goods.ExpirationDate, &goods.State, &goods.Origin, &goods.Manufacturer, &goods.ManufacturerLink, &goods.EngineType, &goods.PurchaseLink, &goods.CreatedTime, &goods.UpdatedTime); err != nil {
			return nil, err
		}
		goodsList = append(goodsList, goods)
	}
	return goodsList, nil
}

// CountGoodsByUserId() ��据goods.user_id获取goods数量
func CountGoodsByUserId(userId int) (count int, err error) {
	err = Db.QueryRow("SELECT COUNT(*) FROM goods WHERE user_id = $1", userId).Scan(&count)
	return count, err
}
