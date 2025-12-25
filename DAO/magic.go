package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// 法力，构思，把一个主意（有意义的想法）转化为一个实例（应用成品）的思考;
// 了不起的个人技能,例如原创一首优美的格律诗，并且用工具记录下来（草稿）;
// 观察，收集，整理，推理、判断、解决问题能力
// DS：“匠心”
type Magic struct {
	Id                int
	Uuid              string
	UserId            int    // 创建者用户ID
	Name              string //例如，红楼梦的“会作诗”就是一种法力，“填词”是另一种类似法力，解数学方程，找到设备故障能力，debug也是一种法力？？
	Nickname          string
	Description       string            //对此法力的描述，例如：七步成诗，能构建出解决某种疑难问题思路方法
	IntelligenceLevel IntelligenceLevel // 智力耗费等级(1-5)Mental effort level required.大脑神经元消耗葡萄糖数量？解方程比吟诗更耗费脑力？
	DifficultyLevel   DifficultyLevel   // 掌握构思能力的学习课程难度等级(1-5)，
	Category          MagicCategory     // 类型：0、未知，1、理性， 2、感性
	Level             int               // 段位，
	CreatedAt         time.Time
	UpdatedAt         *time.Time
	DeletedAt         *time.Time //软删除
}
type MagicCategory int

const (
	UnknownMagicCategory MagicCategory = iota
	Rational                           // 理性
	Sensual                            // 感性
)

type MagicUser struct {
	Id      int
	UserId  int // 用户ID
	MagicId int // 法力ID
	Level   int // 掌握法力的段位(1-9)，

	//个人状态：
	//0、迷糊、醉酒、昏迷
	//1、清醒，思路清晰
	//2、专注，心无旁骛
	//3、灵感迸发，妙笔生花
	Status    MagicUserStatus
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time //软删除
}
type MagicUserStatus int

const (
	Confused MagicUserStatus = iota // 迷糊、醉酒、昏迷
	Clear                           // 清醒，思路清晰
	Focused                         // 专注，心无旁骛
	Inspired                        // 灵感迸发，妙笔生花
)

// 团队【法力记录】
type MagicTeam struct {
	Id      int
	MagicId int // 法力ID
	TeamId  int // 团队ID
	Level   int // 团队掌握该法力的段位(1-9)

	// 团队法力状态:
	// 0、混乱 --团队内部冲突、沟通不畅、目标不明
	// 1、清晰 --团队思路清晰、目标明确
	// 2、协同 --团队配合默契、高效协作
	// 3、创新 --团队创新力强、突破性思维
	Status    MagicTeamStatus
	CreatedAt time.Time // 创建时间
	UpdatedAt *time.Time
	DeletedAt *time.Time //软删除
}

type MagicTeamStatus int

const (
	ChaoticMagicTeamStatus       MagicTeamStatus = iota // 0、混乱
	ClearMagicTeamStatus                                // 1、清晰
	CollaborativeMagicTeamStatus                        // 2、协同
	InnovativeMagicTeamStatus                           // 3、创新
)

// IntelligenceLevel 智力耗费等级
// 用于估量完成任务所需的学习成本、推理深度和创造性思维需求
// 等级范围从1到5，数值越高表示智力需求越复杂
// 估量学习成本大小，推理深度层级，灵光一闪（创新）的概率
type IntelligenceLevel int //智力耗费等级(1-5)

const (
	// UnknownIntelligence 未知智力等级
	// 默认初始值，表示尚未评估或无法分类的智力需求
	UnknownIntelligence IntelligenceLevel = iota

	// VeryLowIntelligence 极低智力需求 (等级1)
	// 特征：几乎无需学习，机械性重复操作
	// - 学习成本：接近零，可立即上手
	// - 推理深度：表面层次，无需复杂思考
	// - 灵光一闪概率：极低，主要依赖肌肉记忆
	// 示例：简单数据录入、基础组装工作、按固定流程操作
	VeryLowIntelligence

	// LowIntelligence 低智力需求 (等级2)
	// 特征：需要基础理解，但复杂度有限
	// - 学习成本：较低，短期培训即可掌握
	// - 推理深度：浅层推理，涉及简单问题解决
	// - 灵光一闪概率：较低，主要依赖既定程序
	// 示例：常规客服应答、标准流程执行、基础设备操作
	LowIntelligence

	// ModerateIntelligence 中等智力需求 (等级3)
	// 特征：需要系统学习和实践应用
	// - 学习成本：中等，需要专门培训和实践
	// - 推理深度：中等深度，涉及多因素分析
	// - 灵光一闪概率：中等，可能出现创新性解决方案
	// 示例：技术故障排查、项目计划制定、中等复杂度分析
	ModerateIntelligence

	// HighIntelligence 高智力需求 (等级4)
	// 特征：需要深度专业知识和创造性思维
	// - 学习成本：较高，需要长期专业积累
	// - 推理深度：深层推理，涉及复杂系统分析
	// - 灵光一闪概率：较高，经常需要创新突破
	// 示例：科研问题解决、复杂系统设计、战略规划制定
	HighIntelligence

	// VeryHighIntelligence 极高智力需求 (等级5)
	// 特征：需要顶尖专业水平和突破性思维
	// - 学习成本：极高，需要领域内顶尖专业知识
	// - 推理深度：极其复杂，涉及跨学科综合推理
	// - 灵光一闪概率：很高，依赖突破性创新思维
	// 示例：前沿科学研究、重大技术突破、复杂危机处理
	VeryHighIntelligence
)

// CRUD 操作方法

// Create 创建新的法力记录
func (m *Magic) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO magics 
		(uuid, user_id, name, nickname, description, intelligence_level, difficulty_level, category, level) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), m.UserId, m.Name, m.Nickname, m.Description,
		m.IntelligenceLevel, m.DifficultyLevel, m.Category, m.Level).Scan(&m.Id, &m.Uuid)
	return err
}

// GetByIdOrUUID 根据ID或UUID获取法力记录
func (m *Magic) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if m.Id <= 0 && m.Uuid == "" {
		return errors.New("invalid Magic ID or UUID")
	}
	statement := `SELECT id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM magics WHERE (id=$1 OR uuid=$2) AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, m.Id, m.Uuid).Scan(&m.Id, &m.Uuid, &m.UserId, &m.Name, &m.Nickname, &m.Description,
		&m.IntelligenceLevel, &m.DifficultyLevel, &m.Category, &m.Level, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
	return err
}

// Update 更新法力记录
func (m *Magic) Update() error {
	statement := `UPDATE magics SET name = $2, nickname = $3, description = $4, 
		intelligence_level = $5, difficulty_level = $6, category = $7, level = $8, updated_at = $9  
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.Id, m.Name, m.Nickname, m.Description, m.IntelligenceLevel, m.DifficultyLevel, m.Category, m.Level, time.Now())
	return err
}

// Delete 软删除法力记录
func (m *Magic) Delete() error {
	statement := `UPDATE magics SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(m.Id, now)
	if err == nil {
		m.DeletedAt = &now
	}
	return err
}

// CreatedDateTime 格式化创建时间
func (m *Magic) CreatedDateTime() string {
	return m.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// CategoryString 获取分类字符串
func (m *Magic) CategoryString() string {
	switch m.Category {
	case Rational:
		return "理性"
	case Sensual:
		return "感性"
	default:
		return "未知分类"
	}
}

// LevelString 返回智力等级的字符串表示
func (i IntelligenceLevel) LevelString() string {
	switch i {
	case UnknownIntelligence:
		return "未知智力等级"
	case VeryLowIntelligence:
		return "极低智力需求"
	case LowIntelligence:
		return "低智力需求"
	case ModerateIntelligence:
		return "中等智力需求"
	case HighIntelligence:
		return "高智力需求"
	case VeryHighIntelligence:
		return "极高智力需求"
	default:
		return "未定义等级"
	}
}

// Description 返回智力等级的详细描述
func (i IntelligenceLevel) Description() string {
	switch i {
	case UnknownIntelligence:
		return "尚未评估或无法分类的智力需求等级"
	case VeryLowIntelligence:
		return "几乎无需学习，机械性重复操作，推理深度浅，创新需求极低"
	case LowIntelligence:
		return "需要基础理解但复杂度有限，涉及简单问题解决，创新需求较低"
	case ModerateIntelligence:
		return "需要系统学习和实践应用，涉及多因素分析，有中等创新需求"
	case HighIntelligence:
		return "需要深度专业知识和创造性思维，涉及复杂系统分析，创新需求较高"
	case VeryHighIntelligence:
		return "需要顶尖专业水平和突破性思维，涉及跨学科推理，创新需求极高"
	default:
		return "未定义的智力需求等级"
	}
}

// IntelligenceLevelString 获取智力等级字符串
func (m *Magic) IntelligenceLevelString() string {
	return m.IntelligenceLevel.LevelString()
}

// DifficultyLevelString 获取难度等级字符串
func (m *Magic) DifficultyLevelString() string {
	switch m.DifficultyLevel {
	case VeryLowDifficulty:
		return "极易"
	case LowDifficulty:
		return "较易"
	case ModerateDifficulty:
		return "中等"
	case HighDifficulty:
		return "较难"
	case VeryHighDifficulty:
		return "极难"
	default:
		return "未知难度"
	}
}

// IsHighLevel 判断是否为高等级法力
func (m *Magic) IsHighLevel() bool {
	return m.Level >= 4
}

// IsHighIntelligence 判断是否为高智力要求
func (m *Magic) IsHighIntelligence() bool {
	return m.IntelligenceLevel >= HighIntelligence
}

// IsHighDifficulty 判断是否为高难度掌握
func (m *Magic) IsHighDifficulty() bool {
	return m.DifficultyLevel >= HighDifficulty
}

// 根据分类获取法力列表
func GetMagicsByCategory(category MagicCategory, ctx context.Context) ([]Magic, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM magics WHERE category = $1 AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`
	rows, err := DB.QueryContext(ctx, statement, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var magics []Magic
	for rows.Next() {
		var m Magic
		err := rows.Scan(&m.Id, &m.Uuid, &m.UserId, &m.Name, &m.Nickname, &m.Description,
			&m.IntelligenceLevel, &m.DifficultyLevel, &m.Category, &m.Level, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
		if err != nil {
			return nil, err
		}
		magics = append(magics, m)
	}
	return magics, nil
}

// 根据难度等级获取法力列表
func GetMagicsByDifficultyLevel(difficultyLevel DifficultyLevel, ctx context.Context) ([]Magic, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM magics WHERE difficulty_level = $1 AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`
	rows, err := DB.QueryContext(ctx, statement, difficultyLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var magics []Magic
	for rows.Next() {
		var m Magic
		err := rows.Scan(&m.Id, &m.Uuid, &m.UserId, &m.Name, &m.Nickname, &m.Description,
			&m.IntelligenceLevel, &m.DifficultyLevel, &m.Category, &m.Level, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
		if err != nil {
			return nil, err
		}
		magics = append(magics, m)
	}
	return magics, nil
}

// 获取所有法力列表
func GetAllMagics(ctx context.Context) ([]Magic, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM magics WHERE deleted_at IS NULL ORDER BY category, level DESC, created_at DESC`
	rows, err := DB.QueryContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var magics []Magic
	for rows.Next() {
		var m Magic
		err := rows.Scan(&m.Id, &m.Uuid, &m.UserId, &m.Name, &m.Nickname, &m.Description,
			&m.IntelligenceLevel, &m.DifficultyLevel, &m.Category, &m.Level, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
		if err != nil {
			return nil, err
		}
		magics = append(magics, m)
	}
	return magics, nil
}

// MagicUser CRUD 方法

// MagicUser.Create 创建用户法力记录
func (mu *MagicUser) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `INSERT INTO magic_users (magic_id, user_id, level, status) VALUES ($1, $2, $3, $4) RETURNING id`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, mu.MagicId, mu.UserId, mu.Level, mu.Status).Scan(&mu.Id)
}

// MagicUser.GetByUserAndMagic 根据用户ID和法力ID获取用户法力记录
func (mu *MagicUser) GetByUserAndMagic(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, magic_id, user_id, level, status, created_at, updated_at, deleted_at
		FROM magic_users WHERE user_id = $1 AND magic_id = $2 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, mu.UserId, mu.MagicId).Scan(&mu.Id, &mu.MagicId, &mu.UserId, &mu.Level, &mu.Status, &mu.CreatedAt, &mu.UpdatedAt, &mu.DeletedAt)
}

// MagicUser.Update 更新用户法力记录
func (mu *MagicUser) Update() error {
	statement := `UPDATE magic_users SET level = $2, status = $3, updated_at = $4 WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(mu.Id, mu.Level, mu.Status, time.Now())
	return err
}

// MagicUser.Delete 软删除用户法力记录
func (mu *MagicUser) Delete() error {
	statement := `UPDATE magic_users SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(mu.Id, now)
	if err == nil {
		mu.DeletedAt = &now
	}
	return err
}

// MagicUser.StatusString 获取状态字符串
func (mu *MagicUser) StatusString() string {
	switch mu.Status {
	case Confused:
		return "迷糊"
	case Clear:
		return "清醒"
	case Focused:
		return "专注"
	case Inspired:
		return "灵感迸发"
	default:
		return "未知状态"
	}
}

// GetUserMagics 获取用户的所有法力
func GetUserMagics(userId int, ctx context.Context) ([]MagicUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, magic_id, user_id, level, status, created_at, updated_at, deleted_at
		FROM magic_users WHERE user_id = $1 AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`
	rows, err := DB.QueryContext(ctx, statement, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userMagics []MagicUser
	for rows.Next() {
		var mu MagicUser
		err := rows.Scan(&mu.Id, &mu.MagicId, &mu.UserId, &mu.Level, &mu.Status, &mu.CreatedAt, &mu.UpdatedAt, &mu.DeletedAt)
		if err != nil {
			return nil, err
		}
		userMagics = append(userMagics, mu)
	}
	return userMagics, nil
}

// MagicUser.GetById 根据ID获取用户法力记录
func (mu *MagicUser) GetById(id int, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, magic_id, user_id, level, status, created_at, updated_at, deleted_at
		FROM magic_users WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, id).Scan(&mu.Id, &mu.MagicId, &mu.UserId, &mu.Level, &mu.Status, &mu.CreatedAt, &mu.UpdatedAt, &mu.DeletedAt)
}

// MagicTeam CRUD 方法

// MagicTeam.Create 创建团队【法力记录】
func (mt *MagicTeam) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `INSERT INTO magic_teams (magic_id, team_id, level, status) VALUES ($1, $2, $3, $4) RETURNING id`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, mt.MagicId, mt.TeamId, mt.Level, mt.Status).Scan(&mt.Id)
}

// MagicTeam.GetByTeamAndMagic 根据团队ID和法力ID获取团队【法力记录】
func (mt *MagicTeam) GetByTeamAndMagic(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, magic_id, team_id, level, status, created_at, updated_at, deleted_at
		FROM magic_teams WHERE team_id = $1 AND magic_id = $2 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, mt.TeamId, mt.MagicId).Scan(&mt.Id, &mt.MagicId, &mt.TeamId, &mt.Level, &mt.Status, &mt.CreatedAt, &mt.UpdatedAt, &mt.DeletedAt)
}

// MagicTeam.Update 更新团队【法力记录】
func (mt *MagicTeam) Update() error {
	statement := `UPDATE magic_teams SET level = $2, status = $3, updated_at = $4 WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(mt.Id, mt.Level, mt.Status, time.Now())
	return err
}

// MagicTeam.Delete 软删除团队【法力记录】
func (mt *MagicTeam) Delete() error {
	statement := `UPDATE magic_teams SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(mt.Id, now)
	if err == nil {
		mt.DeletedAt = &now
	}
	return err
}

// MagicTeam.StatusString 获取【法力记录】状态字符串
func (mt *MagicTeam) StatusString() string {
	switch mt.Status {
	case ChaoticMagicTeamStatus:
		return "混乱"
	case ClearMagicTeamStatus:
		return "清晰"
	case CollaborativeMagicTeamStatus:
		return "协同"
	case InnovativeMagicTeamStatus:
		return "创新"
	default:
		return "未知状态"
	}
}

// GetTeamMagics 获取团队的所有法力记录
func GetTeamMagics(teamId int, ctx context.Context) ([]MagicTeam, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, magic_id, team_id, level, status, created_at, updated_at, deleted_at
		FROM magic_teams WHERE team_id = $1 AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`
	rows, err := DB.QueryContext(ctx, statement, teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teamMagics []MagicTeam
	for rows.Next() {
		var mt MagicTeam
		err := rows.Scan(&mt.Id, &mt.MagicId, &mt.TeamId, &mt.Level, &mt.Status, &mt.CreatedAt, &mt.UpdatedAt, &mt.DeletedAt)
		if err != nil {
			return nil, err
		}
		teamMagics = append(teamMagics, mt)
	}
	return teamMagics, nil
}

// CountTeamMagics 统计团队有的法力记录数量
func CountTeamMagics(teamId int, ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM magic_teams WHERE team_id = $1 AND deleted_at IS NULL`
	var count int
	err := DB.QueryRowContext(ctx, statement, teamId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// MagicTeam.GetById 根据ID获取团队法力记录
func (mt *MagicTeam) GetById(id int, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, magic_id, team_id, level, status, created_at, updated_at, deleted_at
		FROM magic_teams WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, id).Scan(&mt.Id, &mt.MagicId, &mt.TeamId, &mt.Level, &mt.Status, &mt.CreatedAt, &mt.UpdatedAt, &mt.DeletedAt)
}

// SearchMagicByName 按名称搜索法力
func SearchMagicByName(keyword string, limit int, ctx context.Context) ([]Magic, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `SELECT id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level, 
              category, level, created_at, updated_at, deleted_at FROM magics 
              WHERE (name ILIKE $1 OR nickname ILIKE $1 OR description ILIKE $1) AND deleted_at IS NULL
              ORDER BY name 
              LIMIT $2`

	rows, err := DB.QueryContext(ctx, query, "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var magics []Magic
	for rows.Next() {
		var magic Magic
		err := rows.Scan(&magic.Id, &magic.Uuid, &magic.UserId, &magic.Name, &magic.Nickname, &magic.Description,
			&magic.IntelligenceLevel, &magic.DifficultyLevel, &magic.Category, &magic.Level, &magic.CreatedAt, &magic.UpdatedAt, &magic.DeletedAt)
		if err != nil {
			return nil, err
		}
		magics = append(magics, magic)
	}

	return magics, nil
}

// EnsureDefaultMagics 确保用户拥有默认法力
func EnsureDefaultMagics(userId int, ctx context.Context) error {
	// 检查用户是否已有法力记录
	count, err := CountUserMagics(userId, ctx)
	if err != nil {
		return err
	}

	// 如果用户已有法力记录，跳过初始化
	if count > 0 {
		return nil
	}

	// 默认法力ID列表（对应setup_default_values.sql中的法力）
	defaultMagicIds := []int{1, 2, 3, 4, 5, 6}

	for _, magicId := range defaultMagicIds {
		magicUser := MagicUser{
			MagicId: magicId,
			UserId:  userId,
			Level:   1,     // 默认等级1
			Status:  Clear, // 默认清醒状态
		}
		if err := magicUser.Create(ctx); err != nil {
			// 记录错误但继续处理其他法力
			continue
		}
	}
	return nil
}

// CountUserMagics 统计用户有的法力记录数量
func CountUserMagics(userId int, ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM magic_users WHERE user_id = $1 AND deleted_at IS NULL`
	var count int
	err := DB.QueryRowContext(ctx, statement, userId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetMagicsByMagicUsers 根据用户法力记录获取法力列表
func GetMagicsByMagicUsers(magicUsers []MagicUser, ctx context.Context) ([]Magic, error) {
	if len(magicUsers) == 0 {
		return []Magic{}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	magicIds := make([]int, len(magicUsers))
	for i, mu := range magicUsers {
		magicIds[i] = mu.MagicId
	}

	placeholders := make([]string, len(magicIds))
	args := make([]any, len(magicIds))
	for i, id := range magicIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	statement := fmt.Sprintf(`SELECT id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level,
		category, level, created_at, updated_at, deleted_at
		FROM magics WHERE id IN (%s) AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`,
		strings.Join(placeholders, ","))

	rows, err := DB.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var magics []Magic
	for rows.Next() {
		var m Magic
		err := rows.Scan(&m.Id, &m.Uuid, &m.UserId, &m.Name, &m.Nickname, &m.Description,
			&m.IntelligenceLevel, &m.DifficultyLevel, &m.Category, &m.Level, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
		if err != nil {
			return nil, err
		}
		magics = append(magics, m)
	}
	return magics, nil
}

// GetMagicsByMagicTeams 根据团队法力记录获取法力列表
func GetMagicsByMagicTeams(magicTeams []MagicTeam, ctx context.Context) ([]Magic, error) {
	if len(magicTeams) == 0 {
		return []Magic{}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	magicIds := make([]int, len(magicTeams))
	for i, mt := range magicTeams {
		magicIds[i] = mt.MagicId
	}

	placeholders := make([]string, len(magicIds))
	args := make([]any, len(magicIds))
	for i, id := range magicIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	statement := fmt.Sprintf(`SELECT id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level,
		category, level, created_at, updated_at, deleted_at
		FROM magics WHERE id IN (%s) AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`,
		strings.Join(placeholders, ","))

	rows, err := DB.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var magics []Magic
	for rows.Next() {
		var m Magic
		err := rows.Scan(&m.Id, &m.Uuid, &m.UserId, &m.Name, &m.Nickname, &m.Description,
			&m.IntelligenceLevel, &m.DifficultyLevel, &m.Category, &m.Level, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
		if err != nil {
			return nil, err
		}
		magics = append(magics, m)
	}
	return magics, nil
}
