package data

import "time"

// 作业场所安全隐患，
// 直接责任方默认是归属场所管理茶团（相对的“risk风险”默认责任是作业执行团队方）
// 识别安全隐患的能力也是一种magic
type Hazard struct {
	Id   int
	Uuid string

	Name        string //隐患名称
	Nickname    string //隐患别名
	Keywords    string //隐患关键词
	Description string //隐患描述
	Source      string //隐患来源

	// 分级管理
	Severity int // 隐患严重度（1-5级）
	Category int // 隐患类型枚举（电气/机械/化学等）

	CreatedAt time.Time
	UpdatedAt *time.Time
}

const (
	HazardCategoryElectrical = iota + 1 //电气
	HazardCategoryMechanical            //机械
	HazardCategoryChemical              //化学
	HazardCategoryBiological            //生物
	HazardCategoryErgonomic             //工效学，人机工程学
	HazardCategoryOther                 //其他
)

func (h *Hazard) CategoryName() string {
	switch h.Category {
	case HazardCategoryElectrical:
		return "电气"
	case HazardCategoryMechanical:
		return "机械"
	case HazardCategoryChemical:
		return "化学"
	case HazardCategoryBiological:
		return "生物"
	case HazardCategoryErgonomic:
		return "工效学"
	case HazardCategoryOther:
		return "其他"
	default:
		return "未知"
	}
}

const (
	HazardSeverityNegligible = 1 // 可忽略
	HazardSeverityLow        = 2 // 低风险
	HazardSeverityMedium     = 3 // 中风险
	HazardSeverityHigh       = 4 // 高风险
	HazardSeverityCritical   = 5 // 危急
)

func (h *Hazard) SeverityName() string {
	switch h.Severity {
	case HazardSeverityNegligible:
		return "可忽略"
	case HazardSeverityLow:
		return "低风险"
	case HazardSeverityMedium:
		return "中风险"
	case HazardSeverityHigh:
		return "高风险"
	case HazardSeverityCritical:
		return "危急"
	default:
		return "未知"
	}
}
