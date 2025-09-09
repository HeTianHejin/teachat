package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// 好东西，举办茶话会活动需要的物资。包括装备、物品和材料等.
// 通常是可以交易的商品，例如车辆、工具、食物、耗材...
type Goods struct {
	Id                    int              //postgreSQL serial主键
	Uuid                  string           // 全局唯一标识符，用于在分布式系统中唯一标识物资
	RecorderUserId        int              // 物资登记人ID
	Name                  string           // 名称,通常是官方正式名称。
	Nickname              string           // 昵称或别名，便于用户记忆和使用。
	Designer              string           // 设计者，适用于有设计元素的物资。
	Describe              string           // 详细描述，帮助用户了解物资的具体信息。
	Price                 float64          // 价格，通常以货币单位表示。
	Applicability         string           // 用途，描述物资的使用场景。
	Category              int              // 类别：0-虚拟（不受重力影响）；1-物体（受重力影响）
	Specification         string           // 标准规格,如技术参数等。
	BrandName             string           // 品牌名称
	Model                 string           // 型号
	Weight                float64          // 重量
	Dimensions            string           // 尺寸
	Material              string           // 材质，适用于物理物品。
	Size                  string           // 物资的大小，适用于物理物品。
	Color                 string           // 物资的颜色或款式。
	NetworkConnectionType string           // 网络连接类型，适用于需要联网的设备。可以考虑使用枚举类型，如 WiFi, Bluetooth, Ethernet 等。
	Features              int              // 物资的特点，0表示可以买卖，1表示不可交易（例如象牙，人体器官？）。
	SerialNumber          string           // 物资的序列号，用于唯一标识每个物资。
	PhysicalState         PhysicalState    // 物理状态
	OperationalState      OperationalState // 运行状态
	Availability          Availability     // 可用性状态
	Origin                string           // 原产地
	Manufacturer          string           // 生产商名称。
	ManufacturerURL       string           // 制造商的官方网站或链接。
	EngineType            string           // 动力类型，适用于有动力系统的物资。锂离子电池驱动？燃油内燃机？市电？
	PurchaseURL           string           // 网购链接，方便用户直接购买。
	CreatedAt             time.Time        // 物资记录的创建时间。
	UpdatedAt             *time.Time       // 物资记录的最后更新时间。
}

// 物理状态（物资的新旧程度）
type PhysicalState int

const (
	PhysicalNew     PhysicalState = iota // 全新
	PhysicalUsed                         // 已使用
	PhysicalWorn                         // 磨损
	PhysicalDamaged                      // 损坏
)

// 功能状态
type OperationalState int

const (
	OperationalNormal      OperationalState = iota // 正常
	OperationalFaulty                              // 故障
	OperationalMaintenance                         // 维修中
	OperationalExpired                             // 已过期
)

// 可用性状态（管理状态）
type Availability int

const (
	Available   Availability = iota // 可用
	InUse                           // 使用中
	Idle                            // 闲置
	Discarded                       // 已报废
	Lost                            // 已遗失
	Transferred                     // 已转让
)

const (
	GoodsCategoryPhysical = iota // 物理物资,受重力影响
	GoodsCategoryVirtual         // 虚拟物资,不受重力影响
)
const (
	GoodsFeatureTradable    = iota // 可买卖
	GoodsFeatureNonTradable        // 不可交易（例如象牙，人体器官？）
)

// Goods.Create() 创建一个Goods记录
func (g *Goods) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO goods 
		(uuid, recorder_user_id, name, nickname, designer, describe, price, applicability, category, 
		 specification, brand_name, model, weight, dimensions, material, size, color, 
		 network_connection_type, features, serial_number, physical_state, operational_state, 
		 availability, origin, manufacturer, manufacturer_url, engine_type, purchase_url) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), g.RecorderUserId, g.Name, g.Nickname, g.Designer,
		g.Describe, g.Price, g.Applicability, g.Category, g.Specification, g.BrandName, g.Model,
		g.Weight, g.Dimensions, g.Material, g.Size, g.Color, g.NetworkConnectionType, g.Features,
		g.SerialNumber, g.PhysicalState, g.OperationalState, g.Availability, g.Origin,
		g.Manufacturer, g.ManufacturerURL, g.EngineType, g.PurchaseURL).Scan(&g.Id, &g.Uuid)
	return err
}

// Goods.Update() 更新Goods记录
func (g *Goods) Update(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE goods SET recorder_user_id = $2, name = $3, nickname = $4, designer = $5, 
		describe = $6, price = $7, applicability = $8, category = $9, specification = $10, 
		brand_name = $11, model = $12, weight = $13, dimensions = $14, material = $15, size = $16, 
		color = $17, network_connection_type = $18, features = $19, serial_number = $20, 
		physical_state = $21, operational_state = $22, availability = $23, origin = $24, 
		manufacturer = $25, manufacturer_url = $26, engine_type = $27, purchase_url = $28, 
		updated_at = $29 WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, g.Id, g.RecorderUserId, g.Name, g.Nickname, g.Designer,
		g.Describe, g.Price, g.Applicability, g.Category, g.Specification, g.BrandName, g.Model,
		g.Weight, g.Dimensions, g.Material, g.Size, g.Color, g.NetworkConnectionType, g.Features,
		g.SerialNumber, g.PhysicalState, g.OperationalState, g.Availability, g.Origin,
		g.Manufacturer, g.ManufacturerURL, g.EngineType, g.PurchaseURL, time.Now())
	return err
}

// Goods.GetByIdOrUUID() 根据ID或UUID获取Goods记录
func (g *Goods) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, uuid, recorder_user_id, name, nickname, designer, describe, price, 
		applicability, category, specification, brand_name, model, weight, dimensions, material, 
		size, color, network_connection_type, features, serial_number, physical_state, 
		operational_state, availability, origin, manufacturer, manufacturer_url, engine_type, 
		purchase_url, created_at, updated_at
		FROM goods WHERE id=$1 OR uuid=$2`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, g.Id, g.Uuid).Scan(&g.Id, &g.Uuid, &g.RecorderUserId,
		&g.Name, &g.Nickname, &g.Designer, &g.Describe, &g.Price, &g.Applicability, &g.Category,
		&g.Specification, &g.BrandName, &g.Model, &g.Weight, &g.Dimensions, &g.Material, &g.Size,
		&g.Color, &g.NetworkConnectionType, &g.Features, &g.SerialNumber, &g.PhysicalState,
		&g.OperationalState, &g.Availability, &g.Origin, &g.Manufacturer, &g.ManufacturerURL,
		&g.EngineType, &g.PurchaseURL, &g.CreatedAt, &g.UpdatedAt)
	return err
}

// GetGoodsByRecorderUserId 根据recorder_user_id查找Goods记录
func GetGoodsByRecorderUserId(recorderUserId int, ctx context.Context) ([]Goods, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, uuid, recorder_user_id, name, nickname, designer, describe, price, 
		applicability, category, specification, brand_name, model, weight, dimensions, material, 
		size, color, network_connection_type, features, serial_number, physical_state, 
		operational_state, availability, origin, manufacturer, manufacturer_url, engine_type, 
		purchase_url, created_at, updated_at
		FROM goods WHERE recorder_user_id = $1 ORDER BY created_at DESC`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, recorderUserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goodsSlice []Goods
	for rows.Next() {
		var g Goods
		err = rows.Scan(&g.Id, &g.Uuid, &g.RecorderUserId, &g.Name, &g.Nickname, &g.Designer,
			&g.Describe, &g.Price, &g.Applicability, &g.Category, &g.Specification, &g.BrandName,
			&g.Model, &g.Weight, &g.Dimensions, &g.Material, &g.Size, &g.Color,
			&g.NetworkConnectionType, &g.Features, &g.SerialNumber, &g.PhysicalState,
			&g.OperationalState, &g.Availability, &g.Origin, &g.Manufacturer, &g.ManufacturerURL,
			&g.EngineType, &g.PurchaseURL, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		goodsSlice = append(goodsSlice, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return goodsSlice, nil
}

// Goods.PhysicalStateString() 返回物理状态的中文描述
func (g *Goods) PhysicalStateString() string {
	switch g.PhysicalState {
	case PhysicalNew:
		return "全新"
	case PhysicalUsed:
		return "已使用"
	case PhysicalWorn:
		return "磨损"
	case PhysicalDamaged:
		return "损坏"
	default:
		return "未知状态"
	}
}

// Goods.OperationalStateString() 返回功能状态的中文描述
func (g *Goods) OperationalStateString() string {
	switch g.OperationalState {
	case OperationalNormal:
		return "正常"
	case OperationalFaulty:
		return "故障"
	case OperationalMaintenance:
		return "维修中"
	case OperationalExpired:
		return "已过期"
	default:
		return "未知状态"
	}
}

// Goods.AvailabilityString() 返回可用性状态的中文描述
func (g *Goods) AvailabilityString() string {
	switch g.Availability {
	case Available:
		return "可用"
	case InUse:
		return "使用中"
	case Idle:
		return "闲置"
	case Discarded:
		return "已报废"
	case Lost:
		return "已遗失"
	case Transferred:
		return "已转让"
	default:
		return "未知状态"
	}
}

// Goods.CreatedDateTime() 返回格式化的创建时间
func (g *Goods) CreatedDateTime() string {
	return g.CreatedAt.Format(FMT_DATE_TIME_CN)
}

type GoodsFamily struct {
	Id        int
	FamilyId  int
	GoodsId   int
	CreatedAt time.Time
}

// GoodsFamily.Create() 保存1家庭收集的物资记录
func (fg *GoodsFamily) Create() (err error) {
	statement := "INSERT INTO goods_families (family_id, goods_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(fg.FamilyId, fg.GoodsId, time.Now()).Scan(&fg.Id)
	return
}

// GoodsFamily.Get() 获取1家庭收集的物资记录
func (fg *GoodsFamily) Get() (err error) {
	err = db.QueryRow("SELECT id, family_id, goods_id, created_at FROM goods_families WHERE id = $1", fg.Id).
		Scan(&fg.Id, &fg.FamilyId, &fg.GoodsId, &fg.CreatedAt)
	return
}

// GoodsFamily.Update() 更新1家庭收集的物资记录
func (fg *GoodsFamily) Update() (err error) {
	statement := "UPDATE goods_families SET family_id = $2, goods_id = $3, created_at = $4 WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(fg.Id, fg.FamilyId, fg.GoodsId, fg.CreatedAt)
	return
}

// GoodsFamily.Delete() 删除1家庭收集的物资记录
func (fg *GoodsFamily) Delete() (err error) {
	statement := "DELETE FROM goods_families WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(fg.Id)
	return
}

// GoodsFamily.GetByFamilyIdAndGoodsId() 获取1家庭收集的物资记录
func (fg *GoodsFamily) GetByFamilyIdAndGoodsId() (err error) {
	err = db.QueryRow("SELECT id, family_id, goods_id, created_at FROM goods_families WHERE family_id = $1 AND goods_id = $2", fg.FamilyId, fg.GoodsId).
		Scan(&fg.Id, &fg.FamilyId, &fg.GoodsId, &fg.CreatedAt)
	return
}

// GoodsFamily.GetAllByFamilyId()  获取家庭收集的所有物资记录
func (fg *GoodsFamily) GetAllByFamilyId() (goodsFamilySlice []GoodsFamily, err error) {
	rows, err := db.Query("SELECT id, family_id, goods_id, created_at FROM goods_families WHERE family_id = $1", fg.FamilyId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var goodsFamily GoodsFamily
		err = rows.Scan(&goodsFamily.Id, &goodsFamily.FamilyId, &goodsFamily.GoodsId, &goodsFamily.CreatedAt)
		if err != nil {
			return
		}
		goodsFamilySlice = append(goodsFamilySlice, goodsFamily)
	}
	return
}

// GetGoodsByFamilyId()  获取家庭收集的所有物资记录
func (fg *GoodsFamily) GetGoodsByFamilyId() (goodsSlice []Goods, err error) {
	rows, err := db.Query("SELECT id, uuid, recorder_user_id, name, nickname, designer, describe, price, applicability, category, specification, brand_name, model, weight, dimensions, material, size, color, network_connection_type, features, serial_number, physical_state, operational_state, availability, origin, manufacturer, manufacturer_url, engine_type, purchase_url, created_at, updated_at FROM goods WHERE id IN (SELECT goods_id FROM goods_families WHERE family_id = $1)", fg.FamilyId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var goods Goods
		err = rows.Scan(&goods.Id, &goods.Uuid, &goods.RecorderUserId, &goods.Name, &goods.Nickname, &goods.Designer, &goods.Describe, &goods.Price, &goods.Applicability, &goods.Category, &goods.Specification, &goods.BrandName, &goods.Model, &goods.Weight, &goods.Dimensions, &goods.Material, &goods.Size, &goods.Color, &goods.NetworkConnectionType, &goods.Features, &goods.SerialNumber, &goods.PhysicalState, &goods.OperationalState, &goods.Availability, &goods.Origin, &goods.Manufacturer, &goods.ManufacturerURL, &goods.EngineType, &goods.PurchaseURL, &goods.CreatedAt, &goods.UpdatedAt)
		if err != nil {
			return
		}
		goodsSlice = append(goodsSlice, goods)
	}
	return
}

// CheckGoodsByFamilyId()  检查家庭收集的物资记录是否存在
func (fg *GoodsFamily) CheckGoodsByFamilyId() (exists bool, err error) {
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM goods_families WHERE family_id = $1 AND goods_id = $2", fg.FamilyId, fg.GoodsId).Scan(&count)
	if err != nil {
		return
	}
	exists = count > 0
	return
}

// GoodsFamily.CountByFamilyId()
func (fg *GoodsFamily) CountByFamilyId() (count int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM goods_families WHERE family_id = $1", fg.FamilyId).Scan(&count)
	return
}

type GoodsTeam struct {
	Id        int
	TeamId    int
	GoodsId   int
	CreatedAt time.Time
}

// GoodsTeam.Create() 保存1团队收集的物资记录
func (tg *GoodsTeam) Create() (err error) {
	statement := "INSERT INTO goods_teams (team_id, goods_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(tg.TeamId, tg.GoodsId, time.Now()).Scan(&tg.Id)
	return
}

// GoodsTeam.Get() 获取1团队收集的物资记录
func (tg *GoodsTeam) Get() (err error) {
	err = db.QueryRow("SELECT id, team_id, goods_id, created_at FROM goods_teams WHERE id = $1", tg.Id).
		Scan(&tg.Id, &tg.TeamId, &tg.GoodsId, &tg.CreatedAt)
	return
}

// GoodsTeam.GetByTeamIdAndGoodsId() 获取1团队收集的物资记录
func (tg *GoodsTeam) GetByTeamIdAndGoodsId() (err error) {
	err = db.QueryRow("SELECT id, team_id, goods_id, created_at FROM goods_teams WHERE team_id = $1 AND goods_id = $2", tg.TeamId, tg.GoodsId).
		Scan(&tg.Id, &tg.TeamId, &tg.GoodsId, &tg.CreatedAt)
	return
}

// GoodsTeam.Update() 更新1团队收集的物资记录
func (tg *GoodsTeam) Update() (err error) {
	statement := "UPDATE goods_teams SET team_id = $2, goods_id = $3, created_at = $4 WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(tg.Id, tg.TeamId, tg.GoodsId, tg.CreatedAt)
	return
}

// GoodsTeam.Delete() 删除1团队收集的物资记录
func (tg *GoodsTeam) Delete() (err error) {
	statement := "DELETE FROM goods_teams WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(tg.Id)
	return
}

// GoodsTeam.GetAllByTeamId()  获取团队收集的所有物资记录
func (tg *GoodsTeam) GetAllByTeamId() (goodsTeamSlice []GoodsTeam, err error) {
	rows, err := db.Query("SELECT id, team_id, goods_id, created_at FROM goods_teams WHERE team_id = $1", tg.TeamId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var goodsTeam GoodsTeam
		err = rows.Scan(&goodsTeam.Id, &goodsTeam.TeamId, &goodsTeam.GoodsId, &goodsTeam.CreatedAt)
		if err != nil {
			return
		}
		goodsTeamSlice = append(goodsTeamSlice, goodsTeam)
	}
	return
}

// 根据团队收集的所有物资记录，获取全部团队物资，return []Goods
func (tg *GoodsTeam) GetAllGoodsByTeamId() (goodsSlice []Goods, err error) {
	rows, err := db.Query("SELECT id, team_id, goods_id, created_at FROM goods_teams WHERE team_id = $1", tg.TeamId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var goodsTeam GoodsTeam
		err = rows.Scan(&goodsTeam.Id, &goodsTeam.TeamId, &goodsTeam.GoodsId, &goodsTeam.CreatedAt)
		if err != nil {
			return
		}
		var goods Goods
		goods.Id = goodsTeam.GoodsId
		ctx := context.Background()
		err = goods.GetByIdOrUUID(ctx)
		if err != nil {
			return
		}
		goodsSlice = append(goodsSlice, goods)
	}
	return
}

// GoodsTeam.CountByTeamId()  获取团队收集的物资记录数量
func (tg *GoodsTeam) CountByTeamId() (count int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM goods_teams WHERE team_id = $1", tg.TeamId).Scan(&count)
	return
}

// CheckTeamGoodsExist() 检查团队收集的物资记录是否存在
func (tg *GoodsTeam) CheckTeamGoodsExist() (exist bool, err error) {
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM goods_teams WHERE team_id = $1 AND goods_id = $2)", tg.TeamId, tg.GoodsId).Scan(&exist)
	return
}

// 用户看上（收藏/标记）的物质
type GoodsUser struct {
	Id        int
	UserId    int
	GoodsId   int
	CreatedAt time.Time
}

// GoodsUser.Create() 保存1用户收集的物资记录
func (ug *GoodsUser) Create() (err error) {
	statement := "INSERT INTO goods_users (user_id, goods_id, created_at) VALUES ($1, $2, $3) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(ug.UserId, ug.GoodsId, time.Now()).Scan(&ug.Id)
	return
}

// GoodsUser.Get() 获取1用户收集的物资记录
func (ug *GoodsUser) Get() (err error) {
	err = db.QueryRow("SELECT id, user_id, goods_id, created_at FROM goods_users WHERE id = $1", ug.Id).
		Scan(&ug.Id, &ug.UserId, &ug.GoodsId, &ug.CreatedAt)
	return
}

// GoodsUser.GetByUserIdAndGoodsId() 获取1用户收集的物资记录
func (ug *GoodsUser) GetByUserIdAndGoodsId() (err error) {
	err = db.QueryRow("SELECT id, user_id, goods_id, created_at FROM goods_users WHERE user_id = $1 AND goods_id = $2", ug.UserId, ug.GoodsId).
		Scan(&ug.Id, &ug.UserId, &ug.GoodsId, &ug.CreatedAt)
	return
}

// GoodsUser.Delete() 删除1用户收集的物资记录
func (ug *GoodsUser) Delete() (err error) {
	statement := "DELETE FROM goods_users WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(ug.Id)
	return
}

// GoodsUser.GetAllByUserId()  获取用户收集的所有物资记录
func (ug *GoodsUser) GetAllByUserId() (goodsUserSlice []GoodsUser, err error) {
	rows, err := db.Query("SELECT id, user_id, goods_id, created_at FROM goods_users WHERE user_id = $1", ug.UserId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var goodsUser GoodsUser
		err = rows.Scan(&goodsUser.Id, &goodsUser.UserId, &goodsUser.GoodsId, &goodsUser.CreatedAt)
		if err != nil {
			return
		}
		goodsUserSlice = append(goodsUserSlice, goodsUser)
	}
	return
}

// GoodsUser.CountByUserId()  获取用户收集的物资记录数量
func (gu *GoodsUser) CountByUserId() (count int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM goods_users WHERE user_id = $1", gu.UserId).Scan(&count)
	return
}

// GetGoodsByUserId 获取用户收集的物资记录 [经DeepSeek优化]
// 返回:
//   - 物资切片(可能为空)
//   - 错误(包括sql.ErrNoRows当用户无物资时)
func (gu *GoodsUser) GetGoodsByUserId() ([]Goods, error) {
	// 使用JOIN一次性获取所有数据，避免N+1查询
	query := `
        SELECT g.id, g.uuid, g.recorder_user_id, g.name, g.nickname, g.designer, g.describe, g.price, g.applicability, g.category, g.specification, g.brand_name, g.model, g.weight, g.dimensions, g.material, g.size, g.color, g.network_connection_type, g.features, g.serial_number, g.physical_state, g.operational_state, g.availability, g.origin, g.manufacturer, g.manufacturer_url, g.engine_type, g.purchase_url, g.created_at, g.updated_at
        FROM goods g
        JOIN goods_users gu ON g.id = gu.goods_id
        WHERE gu.user_id = $1
    `

	rows, err := db.Query(query, gu.UserId)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var goodsSlice []Goods
	for rows.Next() {
		var goods Goods
		if err := rows.Scan(&goods.Id, &goods.Uuid, &goods.RecorderUserId, &goods.Name, &goods.Nickname, &goods.Designer, &goods.Describe, &goods.Price, &goods.Applicability, &goods.Category, &goods.Specification, &goods.BrandName, &goods.Model, &goods.Weight, &goods.Dimensions, &goods.Material, &goods.Size, &goods.Color, &goods.NetworkConnectionType, &goods.Features, &goods.SerialNumber, &goods.PhysicalState, &goods.OperationalState, &goods.Availability, &goods.Origin, &goods.Manufacturer, &goods.ManufacturerURL, &goods.EngineType, &goods.PurchaseURL, &goods.CreatedAt, &goods.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan row failed: %w", err)
		}
		goodsSlice = append(goodsSlice, goods)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	// 如果没有任何记录，返回ErrNoRows
	if len(goodsSlice) == 0 {
		return nil, sql.ErrNoRows
	}

	return goodsSlice, nil
}

// CheckUserGoodsExist() 检查用户是否收藏了该物资
func (ug *GoodsUser) CheckUserGoodsExist() (exist bool, err error) {
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM goods_users WHERE user_id = $1 AND goods_id = $2)", ug.UserId, ug.GoodsId).Scan(&exist)
	return
}
