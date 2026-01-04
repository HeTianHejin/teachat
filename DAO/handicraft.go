package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// 手工艺（作业），技能操作，需要集中注意力身体手眼协调平衡配合完成的动作。
// 例如：制作特色食品，补牙洞，高空攀爬作业...
type Handicraft struct {
	Id             int
	Uuid           string
	TeaOrderId     int // 所属tea-order ID
	RecorderUserId int //记录人id

	Name        string
	Nickname    string
	Description string // 手工艺总览，任务综合描述

	ProjectId int // 发生的茶台ID，项目

	InitiatorId int // 策动人ID
	OwnerId     int // 主理/执行人ID

	Type            HandicraftType   // 类型
	Category        int              // 分类：0、公开，1、私密，仅当事家庭/团队可见内容
	Status          HandicraftStatus // 状态
	SkillDifficulty int              // 技能操作难度(1-5)，引用 skill.DifficultyLevel
	MagicDifficulty int              // 创意思维难度(1-5)，引用 magic.DifficultyLevel

	// 统计字段
	ContributorCount int           // 协助者/助攻人计数
	FinalScore       sql.NullInt64 // 最终得分。NULL代表还没有评分；0代表已评分但是得了0分

	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time //软删除

}

// 评分记录结构体
type HandicraftRating struct {
	Id           int
	HandicraftId int
	RaterUserId  int
	RawScore     int
	Comment      string
	CreatedAt    time.Time
}

// 协助者/助攻人ID列表
type HandicraftContributor struct {
	Id               int
	HandicraftId     int
	UserId           int // 助攻人ID
	ContributionRate int // 贡献值(1-100)
	CreatedAt        time.Time
	DeletedAt        *time.Time //软删除
}

// MagicSlice 手工艺作业的法力集合id
type HandicraftMagic struct {
	Id           int
	Uuid         string
	HandicraftId int
	MagicId      int // magic.go -> Magic{}
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time //软删除
}

// SkillSlice 手工艺作业的技能集合id
type HandicraftSkill struct {
	Id           int
	Uuid         string
	HandicraftId int
	SkillId      int // skill.go -> Skill{}
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time //软删除
}

const (
	HandicraftCategoryPublic = iota // 公开
	HandicraftCategorySecret        // 私密
)

// 手工艺类型 - 根据体力需求和技能复杂度划分
type HandicraftType int

const (
	UnknownWork     HandicraftType = iota // 未知类型/初始化默认值
	LightWork                             // 轻体力（普通人都可以完成，如：喂水、简单清洁）
	MediumWork                            // 中等体力（介于轻重之间，如：给饮水机更换水桶、搬家具）
	HeavyWork                             // 重体力（需要较高强度体能才能完成，如：搬运洗衣机、重物搬运）
	SkillfulWork                          // 轻巧力（需要特定体能加上精细手艺，如：刺绣、精细木工）
	MediumSkillWork                       // 中巧力（中等体力+中等技能，如：家电维修、普通木工）
	HeavySkillWork                        // 重巧力（需要特定体能加上载重力，如：铁艺制作、大型雕塑）
)

// 手工艺状态
type HandicraftStatus int

const (
	NotStarted HandicraftStatus = iota // 未开始，初始化默认值
	InProgress                         // 已开始（进行中）
	Paused                             // 中途暂停
	Completed                          // 已完成（顺利结束）
	Abandoned                          // 已放弃（因故未完成）
)

// 事前，开场状态记录
// 手工艺作业开工仪式，到岗准备开工。例如，书法的起手式，准备动手前一刻的快照
type Inauguration struct {
	Id             int
	Uuid           string
	HandicraftId   int    // 手工艺Id
	Name           string // 某某作业（活动）开始（启动）仪式
	Description    string // 备注描述
	RecorderUserId int    //记录人id
	EvidenceId     int    // 音视频等视觉证据，默认值为 0，表示没有值.指向：evidence.go -> Evidence{}
	Status         int    // 状态： 0、未记录，1、已记录（提交）
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

// Create 创建开工仪式记录
func (i *Inauguration) Create() (err error) {
	statement := `INSERT INTO inaugurations 
		(uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), i.HandicraftId, i.Name, i.Description, i.RecorderUserId, i.EvidenceId, i.Status, time.Now()).Scan(&i.Id, &i.Uuid)
	return err
}

// 事中，过程
// 作业记录仪记录
type ProcessRecord struct {
	Id             int
	Uuid           string
	HandicraftId   int    // 手工艺Id
	Name           string // 某某作业（活动）过程记录
	Description    string // 备注描述
	RecorderUserId int    // 记录人id
	EvidenceId     int    // 音视频等视觉证据，默认值为 0，表示没有值.指向：evidence.go -> Evidence{}
	Status         int    // 状态：0、未记录，1、已记录（提交）
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time //软删除
}

// Create 创建过程记录
func (p *ProcessRecord) Create() (err error) {
	statement := `INSERT INTO process_records 
		(uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), p.HandicraftId, p.Name, p.Description, p.RecorderUserId, p.EvidenceId, p.Status, time.Now()).Scan(&p.Id, &p.Uuid)
	return err
}

// 事终，收尾，
// 手工艺作业结束仪式，离手（场）快照。
type Ending struct {
	Id             int
	Uuid           string
	HandicraftId   int    // 手工艺Id
	Name           string // 某某作业（活动）结束（闭幕）仪式
	Description    string // 备注描述
	RecorderUserId int    // 记录人id
	EvidenceId     int    // 音视频等视觉证据，默认值为 0，表示没有值.指向：evidence.go -> Evidence{}
	Status         int    // 状态：0、未记录，1、已记录（提交）
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

// Create 创建结束仪式记录
func (e *Ending) Create() (err error) {
	statement := `INSERT INTO endings 
		(uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), e.HandicraftId, e.Name, e.Description, e.RecorderUserId, e.EvidenceId, e.Status, time.Now()).Scan(&e.Id, &e.Uuid)
	return err
}

// HandicraftDifficulty 手工艺二维难度结构
type HandicraftDifficulty struct {
	SkillLevel int // 技能操作难度(1-5)
	MagicLevel int // 创意思维难度(1-5)
}

// GetDifficulty 获取手工艺的二维难度
func (h *Handicraft) GetDifficulty() HandicraftDifficulty {
	return HandicraftDifficulty{
		SkillLevel: h.SkillDifficulty,
		MagicLevel: h.MagicDifficulty,
	}
}

// GetOverallDifficulty 计算综合难度等级(1-5)
func (h *Handicraft) GetOverallDifficulty() int {
	// 使用加权平均或最大值等策略
	// 这里使用简单的平均值向上取整
	total := h.SkillDifficulty + h.MagicDifficulty
	return (total + 1) / 2
}

// GetDifficultyType 获取难度类型特征
func (h *Handicraft) GetDifficultyType() string {
	skill, magic := h.SkillDifficulty, h.MagicDifficulty

	if skill >= 4 && magic >= 4 {
		return "高技能高创意" // 大师级作业
	} else if skill >= 4 && magic <= 2 {
		return "高技能低创意" // 熟练工作业
	} else if skill <= 2 && magic >= 4 {
		return "低技能高创意" // 创意设计作业
	} else if skill <= 2 && magic <= 2 {
		return "低技能低创意" // 简单作业
	}
	return "中等难度" // 其他情况
}

// IsHighDifficulty 判断是否为高难度作业
func (h *Handicraft) IsHighDifficulty() bool {
	return h.SkillDifficulty >= 4 || h.MagicDifficulty >= 4
}

// CRUD 操作方法

// Create() 创建新的手工艺记录
func (h *Handicraft) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO handicrafts 
		(uuid, tea_order_id, recorder_user_id, name, nickname, description, project_id, initiator_id, owner_id, 
		 type, category, status, skill_difficulty, magic_difficulty, contributor_count, final_score) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), h.TeaOrderId, h.RecorderUserId, h.Name, h.Nickname, h.Description,
		h.ProjectId, h.InitiatorId, h.OwnerId, h.Type, h.Category, h.Status, h.SkillDifficulty, h.MagicDifficulty, h.ContributorCount, h.FinalScore).Scan(&h.Id, &h.Uuid)
	return err
}

// GetByIdOrUUID() 根据ID或UUID获取手工艺记录
func (h *Handicraft) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if h.Id <= 0 && h.Uuid == "" {
		return errors.New("invalid Handicraft ID or UUID")
	}
	statement := `SELECT id, uuid, tea_order_id, recorder_user_id, name, nickname, description, project_id, 
		initiator_id, owner_id, type, category, status, skill_difficulty, magic_difficulty, 
		contributor_count, final_score, created_at, updated_at, deleted_at
		FROM handicrafts WHERE (id=$1 OR uuid=$2) AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, h.Id, h.Uuid).Scan(&h.Id, &h.Uuid, &h.TeaOrderId, &h.RecorderUserId, &h.Name, &h.Nickname, &h.Description,
		&h.ProjectId, &h.InitiatorId, &h.OwnerId, &h.Type, &h.Category, &h.Status, &h.SkillDifficulty, &h.MagicDifficulty,
		&h.ContributorCount, &h.FinalScore, &h.CreatedAt, &h.UpdatedAt, &h.DeletedAt)
	return err
}

// Update() 更新手工艺记录
func (h *Handicraft) Update() error {
	statement := `UPDATE handicrafts SET name = $2, nickname = $3, description = $4, 
		status = $5, skill_difficulty = $6, magic_difficulty = $7, contributor_count = $8, final_score = $9, updated_at = $10  
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(h.Id, h.Name, h.Nickname, h.Description, h.Status, h.SkillDifficulty, h.MagicDifficulty, h.ContributorCount, h.FinalScore, time.Now())
	return err
}

// Delete() 软删除手工艺记录
func (h *Handicraft) Delete() error {
	statement := `UPDATE handicrafts SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(h.Id, now)
	if err == nil {
		h.DeletedAt = &now
	}
	return err
}

// CreatedDateTime 格式化创建时间
func (h *Handicraft) CreatedDateTime() string {
	return h.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// StatusString 获取状态字符串
func (h *Handicraft) StatusString() string {
	switch h.Status {
	case NotStarted:
		return "未开始"
	case InProgress:
		return "进行中"
	case Paused:
		return "已暂停"
	case Completed:
		return "已完成"
	case Abandoned:
		return "已放弃"
	default:
		return "未知状态"
	}
}

// TypeString 返回类型的名称字符串
func (h HandicraftType) TypeString() string {
	switch h {
	case UnknownWork:
		return "未知类型"
	case LightWork:
		return "轻体力"
	case MediumWork:
		return "中等体力"
	case HeavyWork:
		return "重体力"
	case SkillfulWork:
		return "轻巧力"
	case MediumSkillWork:
		return "中巧力"
	case HeavySkillWork:
		return "重巧力"
	default:
		return "未定义"
	}
}

// Description 返回类型的详细描述
func (h HandicraftType) Description() string {
	switch h {
	case UnknownWork:
		return "未知工作类型"
	case LightWork:
		return "普通人都可以完成的轻体力工作，如：喂水、简单清洁"
	case MediumWork:
		return "介于轻重之间的中等体力工作，如：更换水桶、搬家具"
	case HeavyWork:
		return "需要较高强度体能才能完成的重体力工作，如：搬运洗衣机"
	case SkillfulWork:
		return "需要精细手艺的轻巧工作，如：刺绣、精细木工"
	case MediumSkillWork:
		return "中等体力加中等技能的工作，如：家电维修、普通木工"
	case HeavySkillWork:
		return "需要特定体能加上载重力的工作，如：铁艺制作、大型雕塑"
	default:
		return "未定义的工作类型"
	}
}

// GetHandicraftsByProjectId() 根据项目ID获取手工艺记录
func GetHandicraftsByProjectId(projectId int, ctx context.Context) ([]Handicraft, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, tea_order_id, recorder_user_id, name, nickname, description, project_id, 
		initiator_id, owner_id, type, category, status, skill_difficulty, magic_difficulty, 
		contributor_count, final_score, created_at, updated_at, deleted_at
		FROM handicrafts WHERE project_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := DB.QueryContext(ctx, statement, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var handicrafts []Handicraft
	for rows.Next() {
		var h Handicraft
		err := rows.Scan(&h.Id, &h.Uuid, &h.TeaOrderId, &h.RecorderUserId, &h.Name, &h.Nickname, &h.Description,
			&h.ProjectId, &h.InitiatorId, &h.OwnerId, &h.Type, &h.Category, &h.Status, &h.SkillDifficulty, &h.MagicDifficulty,
			&h.ContributorCount, &h.FinalScore, &h.CreatedAt, &h.UpdatedAt, &h.DeletedAt)
		if err != nil {
			return nil, err
		}
		handicrafts = append(handicrafts, h)
	}
	return handicrafts, nil
}

// IsAllHandicraftsCompleted() 判断项目的全部手工艺作业是否已完成
func IsAllHandicraftsCompleted(projectId int, ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) as total, 
		COUNT(CASE WHEN status = $2 THEN 1 END) as completed
		FROM handicrafts WHERE project_id = $1 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var total, completed int
	err = stmt.QueryRowContext(ctx, projectId, Completed).Scan(&total, &completed)
	if err != nil {
		return false, err
	}

	// 如果没有手工艺作业，返回false,ErrNoRows
	if total == 0 {
		return false, sql.ErrNoRows
	}

	// 全部完成返回true
	return total == completed, nil
}

// HandicraftContributor CRUD 操作

// Create 创建协助者记录
func (hc *HandicraftContributor) Create() (err error) {
	statement := `INSERT INTO handicraft_contributors 
		(handicraft_id, user_id, contribution_rate, created_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(hc.HandicraftId, hc.UserId, hc.ContributionRate, time.Now()).Scan(&hc.Id)
	if err != nil {
		return
	}

	// 更新 handicrafts 表的 contributor_count
	updateStmt := `UPDATE handicrafts SET contributor_count = (
		SELECT COUNT(*) FROM handicraft_contributors 
		WHERE handicraft_id = $1 AND deleted_at IS NULL
	) WHERE id = $1`
	_, err = DB.Exec(updateStmt, hc.HandicraftId)
	return err
}

// Delete 删除协助者记录
func (hc *HandicraftContributor) Delete() error {
	statement := `UPDATE handicraft_contributors SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(hc.Id, now)
	if err == nil {
		hc.DeletedAt = &now
	}
	return err
}

// 获取手工艺的协助者列表，按贡献值降序排列
func (h *Handicraft) GetContributors() ([]HandicraftContributor, error) {
	rows, err := DB.Query("SELECT id, handicraft_id, user_id, contribution_rate, created_at, deleted_at FROM handicraft_contributors WHERE handicraft_id = $1 AND deleted_at IS NULL ORDER BY contribution_rate DESC", h.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributors []HandicraftContributor
	for rows.Next() {
		var hc HandicraftContributor
		err := rows.Scan(&hc.Id, &hc.HandicraftId, &hc.UserId, &hc.ContributionRate, &hc.CreatedAt, &hc.DeletedAt)
		if err != nil {
			return nil, err
		}
		contributors = append(contributors, hc)
	}
	return contributors, nil
}

// HandicraftMagic CRUD 操作

// Create 创建手工艺法力关联
func (hm *HandicraftMagic) Create() (err error) {
	statement := `INSERT INTO handicraft_magics 
		(uuid, handicraft_id, magic_id, created_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), hm.HandicraftId, hm.MagicId, time.Now()).Scan(&hm.Id, &hm.Uuid)
	return err
}

// Delete 删除手工艺法力关联
func (hm *HandicraftMagic) Delete() error {
	statement := `UPDATE handicraft_magics SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(hm.Id, now)
	if err == nil {
		hm.DeletedAt = &now
	}
	return err
}

// HandicraftSkill CRUD 操作

// Create 创建手工艺技能关联
func (hs *HandicraftSkill) Create() (err error) {
	statement := `INSERT INTO handicraft_skills 
		(uuid, handicraft_id, skill_id, created_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), hs.HandicraftId, hs.SkillId, time.Now()).Scan(&hs.Id, &hs.Uuid)
	return err
}

// Delete 删除手工艺技能关联
func (hs *HandicraftSkill) Delete() error {
	statement := `UPDATE handicraft_skills SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(hs.Id, now)
	if err == nil {
		hs.DeletedAt = &now
	}
	return err
}

// GetHandicraftCompletionStatus 获取项目手工艺完成状态统计
func GetHandicraftCompletionStatus(projectId int, ctx context.Context) (total, completed, inProgress, notStarted int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT 
		COUNT(*) as total,
		COUNT(CASE WHEN status = $2 THEN 1 END) as completed,
		COUNT(CASE WHEN status = $3 THEN 1 END) as in_progress,
		COUNT(CASE WHEN status = $4 THEN 1 END) as not_started
		FROM handicrafts WHERE project_id = $1 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, projectId, Completed, InProgress, NotStarted).Scan(&total, &completed, &inProgress, &notStarted)
	return
}

// GetInaugurationsByHandicraftId 根据手工艺ID获取开工仪式记录
func GetInaugurationsByHandicraftId(handicraftId int) ([]Inauguration, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at, updated_at FROM inaugurations WHERE handicraft_id = $1 ORDER BY created_at DESC", handicraftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inaugurations []Inauguration
	for rows.Next() {
		var i Inauguration
		err := rows.Scan(&i.Id, &i.Uuid, &i.HandicraftId, &i.Name, &i.Description, &i.RecorderUserId, &i.EvidenceId, &i.Status, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		inaugurations = append(inaugurations, i)
	}
	return inaugurations, nil
}

// GetProcessRecordsByHandicraftId 根据手工艺ID获取过程记录
func GetProcessRecordsByHandicraftId(handicraftId int) ([]ProcessRecord, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at, updated_at, deleted_at FROM process_records WHERE handicraft_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC", handicraftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ProcessRecord
	for rows.Next() {
		var p ProcessRecord
		err := rows.Scan(&p.Id, &p.Uuid, &p.HandicraftId, &p.Name, &p.Description, &p.RecorderUserId, &p.EvidenceId, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, p)
	}
	return records, nil
}

// GetEndingsByHandicraftId 根据手工艺ID获取结束仪式记录
func GetEndingsByHandicraftId(handicraftId int) ([]Ending, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at, updated_at FROM endings WHERE handicraft_id = $1 ORDER BY created_at DESC", handicraftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endings []Ending
	for rows.Next() {
		var e Ending
		err := rows.Scan(&e.Id, &e.Uuid, &e.HandicraftId, &e.Name, &e.Description, &e.RecorderUserId, &e.EvidenceId, &e.Status, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		endings = append(endings, e)
	}
	return endings, nil
}

// GetHandicraftSkills 根据手工艺ID获取技能关联
func GetHandicraftSkills(handicraftId int) ([]HandicraftSkill, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, skill_id, created_at, updated_at, deleted_at FROM handicraft_skills WHERE handicraft_id = $1 AND deleted_at IS NULL", handicraftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []HandicraftSkill
	for rows.Next() {
		var hs HandicraftSkill
		err := rows.Scan(&hs.Id, &hs.Uuid, &hs.HandicraftId, &hs.SkillId, &hs.CreatedAt, &hs.UpdatedAt, &hs.DeletedAt)
		if err != nil {
			return nil, err
		}
		skills = append(skills, hs)
	}
	return skills, nil
}

// GetHandicraftMagics 根据手工艺ID获取法力关联
func GetHandicraftMagics(handicraftId int) ([]HandicraftMagic, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, magic_id, created_at, updated_at, deleted_at FROM handicraft_magics WHERE handicraft_id = $1 AND deleted_at IS NULL", handicraftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var magics []HandicraftMagic
	for rows.Next() {
		var hm HandicraftMagic
		err := rows.Scan(&hm.Id, &hm.Uuid, &hm.HandicraftId, &hm.MagicId, &hm.CreatedAt, &hm.UpdatedAt, &hm.DeletedAt)
		if err != nil {
			return nil, err
		}
		magics = append(magics, hm)
	}
	return magics, nil
}

// GetInaugurationsByEvidenceId 根据凭证ID获取开工仪式
func GetInaugurationsByEvidenceId(evidenceId int) ([]Inauguration, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at, updated_at FROM inaugurations WHERE evidence_id = $1", evidenceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inaugurations []Inauguration
	for rows.Next() {
		var i Inauguration
		err := rows.Scan(&i.Id, &i.Uuid, &i.HandicraftId, &i.Name, &i.Description, &i.RecorderUserId, &i.EvidenceId, &i.Status, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, err
		}
		inaugurations = append(inaugurations, i)
	}
	return inaugurations, nil
}

// GetProcessRecordsByEvidenceId 根据凭证ID获取过程记录
func GetProcessRecordsByEvidenceId(evidenceId int) ([]ProcessRecord, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at, updated_at, deleted_at FROM process_records WHERE evidence_id = $1 AND deleted_at IS NULL", evidenceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ProcessRecord
	for rows.Next() {
		var p ProcessRecord
		err := rows.Scan(&p.Id, &p.Uuid, &p.HandicraftId, &p.Name, &p.Description, &p.RecorderUserId, &p.EvidenceId, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, p)
	}
	return records, nil
}

// GetEndingsByEvidenceId 根据凭证ID获取结束仪式
func GetEndingsByEvidenceId(evidenceId int) ([]Ending, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, name, description, recorder_user_id, evidence_id, status, created_at, updated_at FROM endings WHERE evidence_id = $1", evidenceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endings []Ending
	for rows.Next() {
		var e Ending
		err := rows.Scan(&e.Id, &e.Uuid, &e.HandicraftId, &e.Name, &e.Description, &e.RecorderUserId, &e.EvidenceId, &e.Status, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		endings = append(endings, e)
	}
	return endings, nil
}

// GetEvidencesByHandicraftId 根据手工艺ID获取凭证列表
func GetEvidencesByHandicraftId(handicraftId int) ([]Evidence, error) {
	rows, err := DB.Query(`SELECT DISTINCT e.id, e.uuid, e.file_name, e.file_path, e.file_size, e.mime_type, 
		e.description, e.category, e.uploader_user_id, e.created_at, e.updated_at, e.deleted_at 
		FROM evidences e 
		LEFT JOIN inaugurations i ON e.id = i.evidence_id 
		LEFT JOIN process_records p ON e.id = p.evidence_id 
		LEFT JOIN endings en ON e.id = en.evidence_id 
		WHERE (i.handicraft_id = $1 OR p.handicraft_id = $1 OR en.handicraft_id = $1) 
		AND e.deleted_at IS NULL`, handicraftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidences []Evidence
	for rows.Next() {
		var e Evidence
		err := rows.Scan(&e.Id, &e.Uuid, &e.FileName, &e.Path, &e.FileSize, &e.MimeType,
			&e.Description, &e.Category, &e.RecorderUserId, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt)
		if err != nil {
			return nil, err
		}
		evidences = append(evidences, e)
	}
	return evidences, nil
}
