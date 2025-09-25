package data

import "time"

// 手工艺（作业），技能操作，需要集中注意力身体手眼协调平衡配合完成的动作。
// 例如：制作特色食品，补牙洞，高空攀爬作业...
type Handicraft struct {
	Id             int
	Uuid           string
	RecorderUserId int //记录人id

	Name        string
	Nickname    string
	Description string // 手工艺总览，任务综合描述

	ProjectId int // 发生的茶台ID，项目

	InitiatorId int // 策动人ID
	OwnerId     int // 主理/执行人ID

	Category         HandicraftCategory // 分类
	Status           HandicraftStatus   // 状态
	SkillDifficulty  int                // 技能操作难度(1-5)，引用 skill.DifficultyLevel
	MagicDifficulty  int                // 创意思维难度(1-5)，引用 magic.DifficultyLevel

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 协助者/助攻人ID列表
type HandicraftContributor struct {
	Id           int
	HandicraftId int
	UserId       int // 助攻人ID
	CreatedAt    time.Time
	DeletedAt    *time.Time //软删除
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

// 手工艺分类
type HandicraftCategory int

const (
	UnknownWork    HandicraftCategory = iota //初始化默认值
	LightWork                                // 轻体力（普通人都可以完成）
	HeavyWork                                // 重体力（需要较高强度体能才能完成）
	SkillfulWork                             // 轻巧力（需要特定体能加上精细手艺）
	HeavySkillWork                           // 重巧力（需要特定体能加上载重力）
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