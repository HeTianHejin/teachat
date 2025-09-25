package data

import "time"

// 法力，构思，把一个主意（有意义的想法）转化为一个实例（应用成品）的思考
// 了不起的个人技能,例如原创一首优美的格律诗，并且用工具记录下来（草稿）
// DS：“匠心”
type Magic struct {
	Id                int
	Uuid              string
	Name              string //例如，红楼梦的“会作诗”就是一种法力，“填词”是另一种类似法力，解数学方程，找到设备故障能力，debug也是一种法力？？
	Nickname          string
	Description       string            //对此法力的描述，例如：七步成诗，能构建出解决某种疑难问题思路方法
	IntelligenceLevel IntelligenceLevel // 智力耗费等级(0-5)Mental effort level required.大脑神经元消耗葡萄糖数量？解方程比吟诗更耗费脑力？
	DifficultyLevel   DifficultyLevel   // 掌握构思能力的学习课程难度等级(0-5)，例如，学习作诗需要先识字（读和写），再大量阅读前人示范优秀诗词，理解，伴随反复思考实验...
	Category          MagicCategory     // 类型：0、未知，1、理性， 2、感性
	Level             int               // 段位，
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}
type MagicCategory int

const (
	UnknownMagicCategory MagicCategory = iota
	Rational                           // 理性
	Sensual                            // 感性
)

type IntelligenceLevel int //智力耗费等级(0-5)

const (
	UnknownIntelligence IntelligenceLevel = iota
	VeryLowIntelligence
	LowIntelligence
	ModerateIntelligence
	HighIntelligence
	VeryHighIntelligence
)
