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

// 用户收集的物资清单
type UserGoods struct {
	Id        int
	UserId    int
	GoodsId   int
	CreatedAt time.Time
}
