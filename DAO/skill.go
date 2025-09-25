package data

import "time"

// skill of person,performance
// 技能，某种可以量化，培训传递的完成某种作业的能力，相对magic来说，不需要太多的文本资料阅读积累以及复杂的逻辑推理，
// 驾驶车辆，飞行器，安装空调，书法，剪发，和面团，抹墙灰，拧螺丝等，通常是操作机械工具去做功的操作组合？
type Skill struct {
	Id              int
	Uuid            string
	Name            string
	Nickname        string
	Description     string          //对此技能的描述，例如：驾驶车辆，飞行器，安装空调等
	StrengthLevel   StrengthLevel   // 体力耗费等级(0-5)，肌肉消耗能量数量？
	DifficultyLevel DifficultyLevel // 掌握作业能力的学习课程难度等级(0-5)，例如，学习驾驶汽车需要先识字，再阅读理解交通规则，伴随反复上公共道路积累行驶经验...
	Category        SkillCategory   // 分类：0、未知类型，1、通用软技能：如沟通，健康与情绪管理等，2、通用硬技能：可以设立试卷考试的科目技能，如：驾驶车辆，控制计算机，拧螺丝...
	Level           int             // 等级，
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}
type SkillCategory int

const (
	UnknownSkillCategory SkillCategory = iota //初始化默认值
	GeneralSoftSkill                          // 通用软技能：如沟通，健康与情绪管理等，
	GeneralHardSkill                          // 通用硬技能：可以设立试卷考试的科目技能，如：驾驶车辆，控制计算机，限时拆装手机...
)

type StrengthLevel int // 体力耗费等级(0-5)

const (
	UnknownStrength StrengthLevel = iota //初始化默认值
	VeryLowStrength
	LowStrength
	ModerateStrength
	HighStrength
	VeryHighStrength
)

type DifficultyLevel int // 掌握能力的学习课程难度等级(0-5)

const (
	UnknownDifficulty DifficultyLevel = iota //初始化默认值
	VeryLowDifficulty
	LowDifficulty
	ModerateDifficulty
	HighDifficulty
	VeryHighDifficulty
)
