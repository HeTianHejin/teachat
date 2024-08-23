package data

import "time"

// 洞天福地，茶话会举办的空间地点。
// 可能是一个建筑物大楼、大厅，房间 ...也可能是行驶中的机舱，船舱,还可能是野外的一块草地，一个帐篷。 maybe a cave like Water Curtain Cave
type Place struct {
	Id                    int
	Uuid                  string
	Name                  string // 名称 ，如贾宝玉的“怡红快绿”院
	Nickname              string // 别名： 绛红轩
	Description           string // 附加说明
	Icon                  string // 图标
	PropertyManagerUserId int    //物业管理人ID 贾宝玉
	MasterUserId          int    // 洞主？物业所有者 如贾府的贾政
	Level                 int    // 等级： 1：特级（城堡），2：一级（别墅），3:独栋，4联排，5公寓楼
	Category              int    // 类型 ：1:私人住宅，2:公共建筑空间，3:户外，4:机舱，5:酒店或商业租赁场所，6:野外
	IsPublic              bool   // 是否公开
	AddressId             int    // （可选）物流快递地址Id
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// 物流快递地址，如大清国金陵市世袭区贾府大街1号大观园
type Address struct {
	Id           int
	Uuid         string
	Name         string
	Nickname     string // 别名
	Description  string // 说明
	Icon         string // 图标
	Country      string // 国家
	Province     string // 省
	City         string // 市
	District     string // 区
	Town         string // 镇/街道
	Village      string // 村/楼盘小区
	Street       string // 街道路
	Building     string // 楼栋
	Unit         string // 单元
	PortalNumber string // 门牌号
	PostalCode   string // 邮政编码
	Category     int    // 类别
	IsPublic     bool   // 是否公开
	IsGovernment bool   // 是否政府单位
	UserId       int    // 登记者id
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
