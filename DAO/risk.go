package data

import "time"

// “看看”作业潜在风险，（ --DeeSeek补充）
// 默认直接责任方是归属作业执行茶团
type Risk struct {
	Id          int
	Uuid        string
	Name        string   // 风险名称（如"高空坠落风险"）
	Nickname    string   // 风险别名（行业术语，如"坠高险"）
	Keywords    []string // 建议改为切片，便于检索（如["高空","安全带","坠落"]）
	Description string   // 需包含风险触发机制（如"在≥2米无护栏平台作业时触发"）
	Source      string   // 细分来源类型（设备/环境/操作）

	// 分级维度
	Severity RiskSeverityLevel // 严重程度（1-5） 等级	Level值	判定标准	示例场景
	// Probability     string // 发生概率（A-D）等级	代码标识	发生频率定义（每年）	典型场景
	// Controllability int    // 可控性（Ⅰ-Ⅲ） 等级	Class值	管控措施有效性	对应措施举例
	// CompositeLevel  string // 综合等级（如"4BⅡ"）

	// 扩展字段
	// Threshold     float64 // 风险阈值（如气体浓度≥100ppm）
	// EmergencyPlan string  // 应急预案ID

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 作业风险分级维度
type RiskSeverityLevel int

const (
	RiskSeverityLevel1 = iota + 1 // 1	可忽略	无实际伤害风险	设备表面灰尘接触
	RiskSeverityLevel2            // 2	轻度	轻微伤害可现场处理	皮肤轻微割伤
	RiskSeverityLevel3            // 3	中度	可逆性伤害需专业医疗干预	化学灼伤（二级）
	RiskSeverityLevel4            // 4	严重	可能造成重大伤害（如骨折、中毒住院）	有毒气体急性中毒
	RiskSeverityLevel5            // 5	致命	可能导致死亡/永久性伤残	高空坠落（无防护）、高压电击
)

func RiskSeverityLevelString(level int) string {
	switch level {
	case RiskSeverityLevel1:
		return "可忽略"
	case RiskSeverityLevel2:
		return "轻度"
	case RiskSeverityLevel3:
		return "中度"
	case RiskSeverityLevel4:
		return "严重"
	case RiskSeverityLevel5:
		return "致命"
	default:
		return "未知"
	}
}

// const (
// 	RiskProbabilityLevelA = "极高" // A	极高	>10次	老旧设备电路漏电
// 	RiskProbabilityLevelB = "高"  // B	高	1~10次	高空作业安全带失效
// 	RiskProbabilityLevelC = "中"  // C	中	0.1~1次	有毒气体微量泄漏
// 	RiskProbabilityLevelD = "低"  // D	低	<0.1次	设备罕见材料疲劳断裂
// )

// const (
// 	RiskControllabilityLevel1 = iota + 1 // Ⅰ	难控	现有技术无法完全消除风险	密闭空间突发爆炸
// 	RiskControllabilityLevel2            // Ⅱ	部分可控	需专业设备和严格规程	高压设备带电检修
// 	RiskControllabilityLevel3            // Ⅲ	易控	通过基础防护可规避	佩戴防毒面具处理低毒气体
// )
