package data

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// skill of person,performance
// 技能，某种可以量化，培训传递的完成某种作业的能力，相对magic来说，不需要太多的文本资料阅读积累以及复杂的逻辑推理，
// 驾驶车辆，飞行器，安装空调，书法，剪发，和面团，抹墙灰，拧螺丝等，通常是操作机械工具去做功的操作组合？
type Skill struct {
	Id              int
	Uuid            string
	UserId          int // 创建者用户ID
	Name            string
	Nickname        string
	Description     string          //对此技能的描述，例如：驾驶车辆，飞行器，安装空调等
	StrengthLevel   StrengthLevel   // 体力耗费等级(1-5)，肌肉消耗能量数量？
	DifficultyLevel DifficultyLevel // 掌握作业能力的学习课程难度等级(1-5)，例如，学习驾驶汽车需要先识字，再阅读理解交通规则，伴随反复上公共道路积累行驶经验...
	Category        SkillCategory   // 分类：0、未知类型，1、通用软技能：如沟通，健康与情绪管理等，2、通用硬技能：可以设立试卷考试的科目技能，如：驾驶车辆，控制计算机，拧螺丝...
	Level           int             // 等级，
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	DeletedAt       *time.Time //软删除
}
type SkillCategory int

const (
	UnknownSkillCategory SkillCategory = iota //初始化默认值
	GeneralSoftSkill                          // 通用软技能：如沟通，健康与情绪管理等，
	GeneralHardSkill                          // 通用硬技能：可以设立试卷考试的科目技能，如：驾驶车辆，控制计算机，限时拆装轮胎...
)

type StrengthLevel int // 体力耗费等级(1-5)

const (
	UnknownStrength StrengthLevel = iota //初始化默认值
	VeryLowStrength
	LowStrength
	ModerateStrength
	HighStrength
	VeryHighStrength
)

type DifficultyLevel int // 掌握能力的学习课程难度等级(1-5)

const (
	UnknownDifficulty DifficultyLevel = iota //初始化默认值
	VeryLowDifficulty
	LowDifficulty
	ModerateDifficulty
	HighDifficulty
	VeryHighDifficulty
)

// 用户【技能记录】
type SkillUser struct {
	Id      int
	SkillId int // 技能ID
	UserId  int // 用户ID
	Level   int // 用户掌握该技能的等级，1-9

	// 用户技能状态:
	// 0、失能 --嗑药、喝酒、病残、受伤，
	// 1、弱能 --老人婴童、饥饿、发烧生病，
	// 2、中能，--普通人成年技能，
	// 3、强能，--运动员，特种兵，专业技师
	Status    SkillUserStatus
	CreatedAt time.Time // 创建时间
	UpdatedAt *time.Time
	DeletedAt *time.Time //软删除
}
type SkillUserStatus int

const (
	DisabledSkillUserStatus SkillUserStatus = iota // 0、失能
	WeakSkillUserStatus                            // 1、弱能
	NormalSkillUserStatus                          // 2、中能
	StrongSkillUserStatus                          // 3、强能
)

// CRUD 操作方法

// Skill.Create 创建新的技能
func (s *Skill) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO skills 
		(uuid, user_id, name, nickname, description, strength_level, difficulty_level, category, level) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), s.UserId, s.Name, s.Nickname, s.Description,
		s.StrengthLevel, s.DifficultyLevel, s.Category, s.Level).Scan(&s.Id, &s.Uuid)
	return err
}

// Skill.GetByIdOrUUID 根据ID或UUID获取技能
func (s *Skill) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if s.Id <= 0 && s.Uuid == "" {
		return errors.New("invalid Skill ID or UUID")
	}
	statement := `SELECT id, uuid, user_id, name, nickname, description, strength_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM skills WHERE (id=$1 OR uuid=$2) AND deleted_at IS NULL`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, s.Id, s.Uuid).Scan(&s.Id, &s.Uuid, &s.UserId, &s.Name, &s.Nickname, &s.Description,
		&s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	return err
}

// Skill.Update 更新技能
func (s *Skill) Update() error {
	statement := `UPDATE skills SET name = $2, nickname = $3, description = $4, 
		strength_level = $5, difficulty_level = $6, category = $7, level = $8, updated_at = $9  
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(s.Id, s.Name, s.Nickname, s.Description, s.StrengthLevel, s.DifficultyLevel, s.Category, s.Level, time.Now())
	return err
}

// Skill.Delete 软删除技能
func (s *Skill) Delete() error {
	statement := `UPDATE skills SET deleted_at = $2 WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(s.Id, now)
	if err == nil {
		s.DeletedAt = &now
	}
	return err
}

// Skill.CreatedDateTime 格式化技能创建时间
func (s *Skill) CreatedDateTime() string {
	return s.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// Skill.CategoryString 获取分类字符串
func (s *Skill) CategoryString() string {
	switch s.Category {
	case GeneralSoftSkill:
		return "通用软技能"
	case GeneralHardSkill:
		return "通用硬技能"
	default:
		return "未知分类"
	}
}

// Skill.StrengthLevelString 获取体力等级字符串
func (s *Skill) StrengthLevelString() string {
	switch s.StrengthLevel {
	case VeryLowStrength:
		return "极低"
	case LowStrength:
		return "较低"
	case ModerateStrength:
		return "中等"
	case HighStrength:
		return "较高"
	case VeryHighStrength:
		return "极高"
	default:
		return "未知等级"
	}
}

// Skill.DifficultyLevelString 获取难度等级字符串
func (s *Skill) DifficultyLevelString() string {
	switch s.DifficultyLevel {
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

// Skill.IsHighLevel 判断是否为高等级技能
func (s *Skill) IsHighLevel() bool {
	return s.Level >= 4
}

// Skill.IsHighStrength 判断是否为高体力要求
func (s *Skill) IsHighStrength() bool {
	return s.StrengthLevel >= HighStrength
}

// Skill.IsHighDifficulty 判断是否为高难度掌握
func (s *Skill) IsHighDifficulty() bool {
	return s.DifficultyLevel >= HighDifficulty
}

// Skill.GetSkillsByCategory 根据分类获取技能列表
func GetSkillsByCategory(category SkillCategory, ctx context.Context) ([]Skill, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, user_id, name, nickname, description, strength_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM skills WHERE category = $1 AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`
	rows, err := db.QueryContext(ctx, statement, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		err := rows.Scan(&s.Id, &s.Uuid, &s.UserId, &s.Name, &s.Nickname, &s.Description,
			&s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
		if err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// Skill.GetSkillsByStrengthLevel 根据体力等级获取技能列表
func GetSkillsByStrengthLevel(strengthLevel StrengthLevel, ctx context.Context) ([]Skill, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, user_id, name, nickname, description, strength_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM skills WHERE strength_level = $1 AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`
	rows, err := db.QueryContext(ctx, statement, strengthLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		err := rows.Scan(&s.Id, &s.Uuid, &s.UserId, &s.Name, &s.Nickname, &s.Description,
			&s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
		if err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// Skill.GetSkillsByDifficultyLevel 根据难度等级获取技能列表
func GetSkillsByDifficultyLevel(difficultyLevel DifficultyLevel, ctx context.Context) ([]Skill, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, user_id, name, nickname, description, strength_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM skills WHERE difficulty_level = $1 AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`
	rows, err := db.QueryContext(ctx, statement, difficultyLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		err := rows.Scan(&s.Id, &s.Uuid, &s.UserId, &s.Name, &s.Nickname, &s.Description,
			&s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
		if err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// Skill.GetSkillsByRecordUserId() 根据user_id获取其登记的全部技能列表
func GetSkillsByRecordUserId(userId int, ctx context.Context) ([]Skill, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, user_id, name, nickname, description, strength_level, difficulty_level,
		category, level, created_at, updated_at, deleted_at
		FROM skills WHERE user_id = $1 AND deleted_at IS NULL ORDER BY category, level DESC, created_at DESC`
	rows, err := db.QueryContext(ctx, statement, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		err := rows.Scan(&s.Id, &s.Uuid, &s.UserId, &s.Name, &s.Nickname, &s.Description,
			&s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
		if err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// Skill.CountSkillsByRecordUserId() 统计user登记的技能数量
func CountSkillsByRecordUserId(userId int, ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM skills WHERE user_id = $1 AND deleted_at IS NULL`
	var count int
	err := db.QueryRowContext(ctx, statement, userId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// SkillUser CRUD 方法

// SkillUser.Create 创建用户【技能记录】
func (su *SkillUser) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `INSERT INTO skill_users (skill_id, user_id, level, status) VALUES ($1, $2, $3, $4) RETURNING id`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, su.SkillId, su.UserId, su.Level, su.Status).Scan(&su.Id)
}

// SkillUser.GetByUserAndSkill 根据用户ID和技能ID获取用户【技能记录】
func (su *SkillUser) GetByUserAndSkill(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, skill_id, user_id, level, status, created_at, updated_at, deleted_at
		FROM skill_users WHERE user_id = $1 AND skill_id = $2 AND deleted_at IS NULL`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, su.UserId, su.SkillId).Scan(&su.Id, &su.SkillId, &su.UserId, &su.Level, &su.Status, &su.CreatedAt, &su.UpdatedAt, &su.DeletedAt)
}

// SkillUser.Update 更新用户【技能记录】
func (su *SkillUser) Update() error {
	statement := `UPDATE skill_users SET level = $2, status = $3, updated_at = $4 WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(su.Id, su.Level, su.Status, time.Now())
	return err
}

// SkillUser.Delete 软删除用户【技能记录】
func (su *SkillUser) Delete() error {
	statement := `UPDATE skill_users SET deleted_at = $2 WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(su.Id, now)
	if err == nil {
		su.DeletedAt = &now
	}
	return err
}

// SkillUser.StatusString 获取【技能记录】状态字符串
func (su *SkillUser) StatusString() string {
	switch su.Status {
	case DisabledSkillUserStatus:
		return "失能"
	case WeakSkillUserStatus:
		return "弱能"
	case NormalSkillUserStatus:
		return "中能"
	case StrongSkillUserStatus:
		return "强能"
	default:
		return "未知状态"
	}
}

// GetUserSkills 获取用户的个人【技能记录】列表
func GetUserSkills(userId int, ctx context.Context) ([]SkillUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, skill_id, user_id, level, status, created_at, updated_at, deleted_at
		FROM skill_users WHERE user_id = $1 AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`
	rows, err := db.QueryContext(ctx, statement, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userSkills []SkillUser
	for rows.Next() {
		var su SkillUser
		err := rows.Scan(&su.Id, &su.SkillId, &su.UserId, &su.Level, &su.Status, &su.CreatedAt, &su.UpdatedAt, &su.DeletedAt)
		if err != nil {
			return nil, err
		}
		userSkills = append(userSkills, su)
	}
	return userSkills, nil
}

// 统计用户有的【技能记录】数量
func CountUserSkills(userId int, ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM skill_users WHERE user_id = $1 AND deleted_at IS NULL`
	var count int
	err := db.QueryRowContext(ctx, statement, userId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 验证用户【技能记录】level值是否有效
func (s *SkillUser) IsValidSkillLevel() error {
	if s.Level < 1 || s.Level > 9 {
		return fmt.Errorf("level must be between 1-9, got %d", s.Level)
	}
	if s.SkillId <= 0 {
		return errors.New("skill_id is required")
	}
	if s.UserId <= 0 {
		return errors.New("user_id is required")
	}
	return nil
}

// 判断用户技能是否可用
func (s *SkillUser) IsUsableSkillStatus() bool {
	return s.Status >= NormalSkillUserStatus && s.Status <= StrongSkillUserStatus && s.DeletedAt == nil
}

// Skill.GetById 根据ID获取技能记录
func (s *Skill) GetById(id int, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, user_id, name, nickname, description, strength_level, difficulty_level, 
		category, level, created_at, updated_at, deleted_at
		FROM skills WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, id).Scan(&s.Id, &s.Uuid, &s.UserId, &s.Name, &s.Nickname, &s.Description,
		&s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
}

// GetSkillsBySkillUsers 根据【技能记录】获取技能列表
func GetSkillsBySkillUsers(skillUsers []SkillUser, ctx context.Context) ([]Skill, error) {
	if len(skillUsers) == 0 {
		return []Skill{}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	skillIds := make([]int, len(skillUsers))
	for i, su := range skillUsers {
		skillIds[i] = su.SkillId
	}

	placeholders := make([]string, len(skillIds))
	args := make([]interface{}, len(skillIds))
	for i, id := range skillIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	statement := fmt.Sprintf(`SELECT id, uuid, user_id, name, nickname, description, strength_level, difficulty_level,
		category, level, created_at, updated_at, deleted_at
		FROM skills WHERE id IN (%s) AND deleted_at IS NULL ORDER BY level DESC, created_at DESC`,
		strings.Join(placeholders, ","))

	rows, err := db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		err := rows.Scan(&s.Id, &s.Uuid, &s.UserId, &s.Name, &s.Nickname, &s.Description,
			&s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
		if err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// LoadAllSkills 根据UserId从SkillUsers和Skills表获取 []Skill
func (u *User) LoadAllSkills(ctx context.Context) ([]Skill, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT s.id, s.uuid, s.user_id, s.name, s.nickname, s.description, 
		s.strength_level, s.difficulty_level, s.category, s.level, s.created_at, s.updated_at, s.deleted_at
		FROM skills s INNER JOIN skill_users su ON s.id = su.skill_id 
		WHERE su.user_id = $1 AND s.deleted_at IS NULL AND su.deleted_at IS NULL 
		ORDER BY su.level DESC, su.created_at DESC`
	rows, err := db.QueryContext(ctx, statement, u.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		err := rows.Scan(&s.Id, &s.Uuid, &s.UserId, &s.Name, &s.Nickname, &s.Description,
			&s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
		if err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}
// EnsureDefaultSkills 确保用户拥有默认技能
func EnsureDefaultSkills(userId int, ctx context.Context) error {
	// 检查用户是否已有技能记录
	count, err := CountUserSkills(userId, ctx)
	if err != nil {
		return err
	}
	
	// 如果用户已有技能记录，跳过初始化
	if count > 0 {
		return nil
	}
	
	// 默认技能ID列表（对应setup_default_values.sql中的前6个技能）
	defaultSkillIds := []int{1, 2, 3, 4, 5, 6}
	
	for _, skillId := range defaultSkillIds {
		skillUser := SkillUser{
			SkillId: skillId,
			UserId:  userId,
			Level:   1,                          // 默认等级1
			Status:  NormalSkillUserStatus,      // 默认中能状态
		}
		if err := skillUser.Create(ctx); err != nil {
			// 记录错误但继续处理其他技能
			continue
		}
	}
	return nil
}