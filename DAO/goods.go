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

	Origin          string     // 原产地
	Manufacturer    string     // 生产商名称。
	ManufacturerURL string     // 制造商的官方网站或链接。
	EngineType      string     // 动力类型，适用于有动力系统的物资。锂离子电池驱动？燃油内燃机？市电？
	PurchaseURL     string     // 网购链接，方便用户直接购买。
	CreatedAt       time.Time  // 物资记录的创建时间。
	UpdatedAt       *time.Time // 物资记录的最后更新时间。
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

// 使用状态（管理状态）
type Availability int

const (
	Available   Availability = iota // 可用
	InUse                           // 使用中
	Idle                            // 闲置
	Discarded                       // 已报废
	Lost                            // 已遗失
	Transferred                     // 已转让
)

// GoodsFamily 家庭物资关系
type GoodsFamily struct {
	Id           int
	FamilyId     int
	GoodsId      int
	Availability Availability // 在该家庭中的使用状态
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// GoodsTeam 团队物资关系
type GoodsTeam struct {
	Id           int
	TeamId       int
	GoodsId      int
	Availability Availability // 在该团队中的使用状态
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

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
		 origin, manufacturer, manufacturer_url, engine_type, purchase_url) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), g.RecorderUserId, g.Name, g.Nickname, g.Designer,
		g.Describe, g.Price, g.Applicability, g.Category, g.Specification, g.BrandName, g.Model,
		g.Weight, g.Dimensions, g.Material, g.Size, g.Color, g.NetworkConnectionType, g.Features,
		g.SerialNumber, g.PhysicalState, g.OperationalState, g.Origin,
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
		physical_state = $21, operational_state = $22, origin = $23, 
		manufacturer = $24, manufacturer_url = $25, engine_type = $26, purchase_url = $27, 
		updated_at = $28 WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, g.Id, g.RecorderUserId, g.Name, g.Nickname, g.Designer,
		g.Describe, g.Price, g.Applicability, g.Category, g.Specification, g.BrandName, g.Model,
		g.Weight, g.Dimensions, g.Material, g.Size, g.Color, g.NetworkConnectionType, g.Features,
		g.SerialNumber, g.PhysicalState, g.OperationalState, g.Origin,
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
		operational_state, origin, manufacturer, manufacturer_url, engine_type, 
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
		&g.OperationalState, &g.Origin, &g.Manufacturer, &g.ManufacturerURL,
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
		operational_state, origin, manufacturer, manufacturer_url, engine_type, 
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
			&g.OperationalState, &g.Origin, &g.Manufacturer, &g.ManufacturerURL,
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

// AvailabilityString 返回可用性状态的中文描述
func AvailabilityString(availability Availability) string {
	switch availability {
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

// GoodsFamily 方法

// Create 创建家庭物资关系
func (gf *GoodsFamily) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO goods_families (family_id, goods_id, availability, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING id`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, gf.FamilyId, gf.GoodsId, gf.Availability, time.Now()).Scan(&gf.Id)
	return err
}

// UpdateAvailability 更新家庭物资的使用状态
func (gf *GoodsFamily) UpdateAvailability(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE goods_families SET availability = $1, updated_at = $2 WHERE id = $3`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, gf.Availability, time.Now(), gf.Id)
	return err
}

// GoodsFamily.CountByFamilyId() 根据family_id统计家庭物资数量
func (gf *GoodsFamily) CountByFamilyId(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT COUNT(*) FROM goods_families WHERE family_id = $1`
	var count int
	err := db.QueryRowContext(ctx, statement, gf.FamilyId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GoodsTeam 方法

// Create 创建团队物资关系
func (gt *GoodsTeam) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO goods_teams (team_id, goods_id, availability, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING id`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, gt.TeamId, gt.GoodsId, gt.Availability, time.Now()).Scan(&gt.Id)
	return err
}

// UpdateAvailability 更新团队物资的使用状态
func (gt *GoodsTeam) UpdateAvailability(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE goods_teams SET availability = $1, updated_at = $2 WHERE id = $3`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, gt.Availability, time.Now(), gt.Id)
	return err
}

// 查询方法

// GetGoodsFamilyByIds 根据家庭ID和物资ID获取家庭物资关系
func GetGoodsFamilyByIds(familyId, goodsId int, ctx context.Context) (*GoodsFamily, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	gf := &GoodsFamily{}
	statement := `SELECT id, family_id, goods_id, availability, created_at, updated_at 
		FROM goods_families WHERE family_id = $1 AND goods_id = $2`
	err := db.QueryRowContext(ctx, statement, familyId, goodsId).Scan(
		&gf.Id, &gf.FamilyId, &gf.GoodsId, &gf.Availability, &gf.CreatedAt, &gf.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return gf, nil
}

// GetGoodsTeamByIds 根据团队ID和物资ID获取团队物资关系
func GetGoodsTeamByIds(teamId, goodsId int, ctx context.Context) (*GoodsTeam, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	gt := &GoodsTeam{}
	statement := `SELECT id, team_id, goods_id, availability, created_at, updated_at 
		FROM goods_teams WHERE team_id = $1 AND goods_id = $2`
	err := db.QueryRowContext(ctx, statement, teamId, goodsId).Scan(
		&gt.Id, &gt.TeamId, &gt.GoodsId, &gt.Availability, &gt.CreatedAt, &gt.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return gt, nil
}

// GetGoodsByFamilyId 获取家庭的所有物资
func GetGoodsByFamilyId(familyId int, ctx context.Context) ([]Goods, []Availability, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT g.id, g.uuid, g.recorder_user_id, g.name, g.nickname, g.designer, 
		g.describe, g.price, g.applicability, g.category, g.specification, g.brand_name, 
		g.model, g.weight, g.dimensions, g.material, g.size, g.color, 
		g.network_connection_type, g.features, g.serial_number, g.physical_state, 
		g.operational_state, g.origin, g.manufacturer, g.manufacturer_url, g.engine_type, 
		g.purchase_url, g.created_at, g.updated_at, gf.availability
		FROM goods g JOIN goods_families gf ON g.id = gf.goods_id 
		WHERE gf.family_id = $1 ORDER BY g.created_at DESC`

	rows, err := db.QueryContext(ctx, statement, familyId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var goodsList []Goods
	var availabilities []Availability
	for rows.Next() {
		var g Goods
		var availability Availability
		err = rows.Scan(&g.Id, &g.Uuid, &g.RecorderUserId, &g.Name, &g.Nickname, &g.Designer,
			&g.Describe, &g.Price, &g.Applicability, &g.Category, &g.Specification, &g.BrandName,
			&g.Model, &g.Weight, &g.Dimensions, &g.Material, &g.Size, &g.Color,
			&g.NetworkConnectionType, &g.Features, &g.SerialNumber, &g.PhysicalState,
			&g.OperationalState, &g.Origin, &g.Manufacturer, &g.ManufacturerURL,
			&g.EngineType, &g.PurchaseURL, &g.CreatedAt, &g.UpdatedAt, &availability)
		if err != nil {
			return nil, nil, err
		}
		goodsList = append(goodsList, g)
		availabilities = append(availabilities, availability)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	return goodsList, availabilities, nil
}

// GoodsTeam.CountByTeamId()
func (gt *GoodsTeam) CountByTeamId(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT COUNT(*) FROM goods_teams WHERE team_id = $1`
	var count int
	err := db.QueryRowContext(ctx, statement, gt.TeamId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetGoodsByTeamId 获取团队的所有物资
func GetGoodsByTeamId(teamId int, ctx context.Context) ([]Goods, []Availability, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT g.id, g.uuid, g.recorder_user_id, g.name, g.nickname, g.designer, 
		g.describe, g.price, g.applicability, g.category, g.specification, g.brand_name, 
		g.model, g.weight, g.dimensions, g.material, g.size, g.color, 
		g.network_connection_type, g.features, g.serial_number, g.physical_state, 
		g.operational_state, g.origin, g.manufacturer, g.manufacturer_url, g.engine_type, 
		g.purchase_url, g.created_at, g.updated_at, gt.availability
		FROM goods g JOIN goods_teams gt ON g.id = gt.goods_id 
		WHERE gt.team_id = $1 ORDER BY g.created_at DESC`

	rows, err := db.QueryContext(ctx, statement, teamId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var goodsList []Goods
	var availabilities []Availability
	for rows.Next() {
		var g Goods
		var availability Availability
		err = rows.Scan(&g.Id, &g.Uuid, &g.RecorderUserId, &g.Name, &g.Nickname, &g.Designer,
			&g.Describe, &g.Price, &g.Applicability, &g.Category, &g.Specification, &g.BrandName,
			&g.Model, &g.Weight, &g.Dimensions, &g.Material, &g.Size, &g.Color,
			&g.NetworkConnectionType, &g.Features, &g.SerialNumber, &g.PhysicalState,
			&g.OperationalState, &g.Origin, &g.Manufacturer, &g.ManufacturerURL,
			&g.EngineType, &g.PurchaseURL, &g.CreatedAt, &g.UpdatedAt, &availability)
		if err != nil {
			return nil, nil, err
		}
		goodsList = append(goodsList, g)
		availabilities = append(availabilities, availability)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	return goodsList, availabilities, nil
}

// Goods.CreatedDateTime() 返回格式化的创建时间
func (g *Goods) CreatedDateTime() string {
	return g.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// Get 获取家庭物资关系
func (gf *GoodsFamily) Get(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, family_id, goods_id, availability, created_at, updated_at 
		FROM goods_families WHERE id = $1`
	err := db.QueryRowContext(ctx, statement, gf.Id).Scan(
		&gf.Id, &gf.FamilyId, &gf.GoodsId, &gf.Availability, &gf.CreatedAt, &gf.UpdatedAt)
	return err
}

// GetByFamilyIdAndGoodsId 根据家庭ID和物资ID获取关系
func (gf *GoodsFamily) GetByFamilyIdAndGoodsId(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, family_id, goods_id, availability, created_at, updated_at 
		FROM goods_families WHERE family_id = $1 AND goods_id = $2`
	err := db.QueryRowContext(ctx, statement, gf.FamilyId, gf.GoodsId).Scan(
		&gf.Id, &gf.FamilyId, &gf.GoodsId, &gf.Availability, &gf.CreatedAt, &gf.UpdatedAt)
	return err
}

// Update 更新家庭物资关系
func (gf *GoodsFamily) Update(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE goods_families SET family_id = $2, goods_id = $3, availability = $4, updated_at = $5 WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, gf.Id, gf.FamilyId, gf.GoodsId, gf.Availability, time.Now())
	return err
}

// Delete 删除家庭物资关系
func (gf *GoodsFamily) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `DELETE FROM goods_families WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, gf.Id)
	return err
}

// Get 获取团队物资关系
func (gt *GoodsTeam) Get(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, team_id, goods_id, availability, created_at, updated_at 
		FROM goods_teams WHERE id = $1`
	err := db.QueryRowContext(ctx, statement, gt.Id).Scan(
		&gt.Id, &gt.TeamId, &gt.GoodsId, &gt.Availability, &gt.CreatedAt, &gt.UpdatedAt)
	return err
}

// GetByTeamIdAndGoodsId 根据团队ID和物资ID获取关系
func (gt *GoodsTeam) GetByTeamIdAndGoodsId(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, team_id, goods_id, availability, created_at, updated_at 
		FROM goods_teams WHERE team_id = $1 AND goods_id = $2`
	err := db.QueryRowContext(ctx, statement, gt.TeamId, gt.GoodsId).Scan(
		&gt.Id, &gt.TeamId, &gt.GoodsId, &gt.Availability, &gt.CreatedAt, &gt.UpdatedAt)
	return err
}

// Update 更新团队物资关系
func (gt *GoodsTeam) Update(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE goods_teams SET team_id = $2, goods_id = $3, availability = $4, updated_at = $5 WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, gt.Id, gt.TeamId, gt.GoodsId, gt.Availability, time.Now())
	return err
}

// Delete 删除团队物资关系
func (gt *GoodsTeam) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `DELETE FROM goods_teams WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, gt.Id)
	return err
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
        SELECT g.id, g.uuid, g.recorder_user_id, g.name, g.nickname, g.designer, g.describe, g.price, g.applicability, g.category, g.specification, g.brand_name, g.model, g.weight, g.dimensions, g.material, g.size, g.color, g.network_connection_type, g.features, g.serial_number, g.physical_state, g.operational_state, g.origin, g.manufacturer, g.manufacturer_url, g.engine_type, g.purchase_url, g.created_at, g.updated_at
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
		if err := rows.Scan(&goods.Id, &goods.Uuid, &goods.RecorderUserId, &goods.Name, &goods.Nickname, &goods.Designer, &goods.Describe, &goods.Price, &goods.Applicability, &goods.Category, &goods.Specification, &goods.BrandName, &goods.Model, &goods.Weight, &goods.Dimensions, &goods.Material, &goods.Size, &goods.Color, &goods.NetworkConnectionType, &goods.Features, &goods.SerialNumber, &goods.PhysicalState, &goods.OperationalState, &goods.Origin, &goods.Manufacturer, &goods.ManufacturerURL, &goods.EngineType, &goods.PurchaseURL, &goods.CreatedAt, &goods.UpdatedAt); err != nil {
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

// SearchGoodsByName 按名称/别名/品牌/型号/规格搜索物资
func SearchGoodsByName(keyword string, limit int, ctx context.Context) ([]Goods, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const query = `
		SELECT id, uuid, recorder_user_id, name, nickname, designer, describe, price, applicability, category, specification, brand_name, model, weight, dimensions, material, size, color, network_connection_type, features, serial_number, physical_state, operational_state, origin, manufacturer, manufacturer_url, engine_type, purchase_url, created_at, updated_at
		FROM goods
		WHERE name ILIKE $1 OR nickname ILIKE $1 OR brand_name ILIKE $1 OR model ILIKE $1 OR specification ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := db.QueryContext(ctx, query, "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goodsSlice []Goods
	for rows.Next() {
		var g Goods
		if err := rows.Scan(&g.Id, &g.Uuid, &g.RecorderUserId, &g.Name, &g.Nickname, &g.Designer, &g.Describe, &g.Price, &g.Applicability, &g.Category, &g.Specification, &g.BrandName, &g.Model, &g.Weight, &g.Dimensions, &g.Material, &g.Size, &g.Color, &g.NetworkConnectionType, &g.Features, &g.SerialNumber, &g.PhysicalState, &g.OperationalState, &g.Origin, &g.Manufacturer, &g.ManufacturerURL, &g.EngineType, &g.PurchaseURL, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		goodsSlice = append(goodsSlice, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return goodsSlice, nil
}

func (gt *GoodsTeam) GetAllGoodsByTeamId(ctx context.Context) ([]Goods, []Availability, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `
		SELECT g.id, g.uuid, g.recorder_user_id, g.name, g.nickname, g.designer, g.describe, g.price, g.applicability, g.category, g.specification, g.brand_name, g.model, g.weight, g.dimensions, g.material, g.size, g.color, g.network_connection_type, g.features, g.serial_number, g.physical_state, g.operational_state, g.origin, g.manufacturer, g.manufacturer_url, g.engine_type, g.purchase_url, g.created_at, g.updated_at, gf.availability
		FROM goods g
		LEFT JOIN goods_families gf ON g.id = gf.goods_id
		WHERE gf.team_id = $1
		ORDER BY g.created_at DESC`

	rows, err := db.QueryContext(ctx, statement, gt.TeamId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var goodsList []Goods
	var availabilities []Availability
	for rows.Next() {
		var g Goods
		var availability Availability
		err = rows.Scan(&g.Id, &g.Uuid, &g.RecorderUserId, &g.Name, &g.Nickname, &g.Designer,
			&g.Describe, &g.Price, &g.Applicability, &g.Category, &g.Specification, &g.BrandName,
			&g.Model, &g.Weight, &g.Dimensions, &g.Material, &g.Size, &g.Color,
			&g.NetworkConnectionType, &g.Features, &g.SerialNumber, &g.PhysicalState,
			&g.OperationalState, &g.Origin, &g.Manufacturer, &g.ManufacturerURL,
			&g.EngineType, &g.PurchaseURL, &g.CreatedAt, &g.UpdatedAt, &availability)
		if err != nil {
			return nil, nil, err
		}
		goodsList = append(goodsList, g)
		availabilities = append(availabilities, availability)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	return goodsList, availabilities, nil
}
