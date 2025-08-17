package data

import "time"

// “看看”（查勘）及手工（施工）作业潜在风险，（ --DeeSeek && ClaudeSonnet补充完善）
// 默认直接责任方是归属作业执行茶团
type Risk struct {
	Id          int
	Uuid        string
	UserId      int    // 记录人ID
	Name        string // 风险名称（如"高空坠落风险"）
	Nickname    string // 风险别名（行业术语，如"坠高险"）
	Keywords    string // 风险关键词
	Description string // 需包含风险触发机制（如"在≥2米无护栏平台作业时触发"）
	Source      string // 细分来源类型（设备/环境/操作）

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

// 操作安全防护措施（针对作业风险）
type SafetyProtection struct {
	Id     int
	Uuid   string
	RiskId int // 关联的风险ID
	UserId int // 负责人ID

	Title       string // 防护措施标题
	Description string // 防护措施描述
	Type        int    // 防护类型（个人防护/集体防护/技术防护）
	Priority    int    // 优先级（1-5）
	Status      int    // 状态（计划中/执行中/已完成）

	Equipment     string     // 所需防护设备
	PlannedDate   *time.Time // 计划执行时间
	CompletedDate *time.Time // 完成时间
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

const (
	ProtectionTypePersonal   = 1 // 个人防护（安全带、防毒面具等）
	ProtectionTypeCollective = 2 // 集体防护（安全网、警戒线等）
	ProtectionTypeTechnical  = 3 // 技术防护（监测设备、自动报警等）
)

const (
	ProtectionStatusPlanned    = 1 // 计划中
	ProtectionStatusInProgress = 2 // 执行中
	ProtectionStatusCompleted  = 3 // 已完成
	ProtectionStatusCancelled  = 4 // 已取消
)

func (p *SafetyProtection) TypeName() string {
	switch p.Type {
	case ProtectionTypePersonal:
		return "个人防护"
	case ProtectionTypeCollective:
		return "集体防护"
	case ProtectionTypeTechnical:
		return "技术防护"
	default:
		return "未知"
	}
}

func (p *SafetyProtection) StatusName() string {
	switch p.Status {
	case ProtectionStatusPlanned:
		return "计划中"
	case ProtectionStatusInProgress:
		return "执行中"
	case ProtectionStatusCompleted:
		return "已完成"
	case ProtectionStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// 获取风险的所有防护措施
func (r *Risk) GetSafetyProtections() ([]SafetyProtection, error) {
	rows, err := Db.Query("SELECT id, uuid, risk_id, user_id, title, description, type, priority, status, equipment, planned_date, completed_date, created_at, updated_at FROM safety_protections WHERE risk_id = $1 ORDER BY priority DESC, created_at DESC", r.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var protections []SafetyProtection
	for rows.Next() {
		var p SafetyProtection
		err := rows.Scan(&p.Id, &p.Uuid, &p.RiskId, &p.UserId, &p.Title, &p.Description, &p.Type, &p.Priority, &p.Status, &p.Equipment, &p.PlannedDate, &p.CompletedDate, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		protections = append(protections, p)
	}
	return protections, nil
}

// 创建安全防护措施
func (p *SafetyProtection) Create() error {
	statement := "INSERT INTO safety_protections (uuid, risk_id, user_id, title, description, type, priority, status, equipment, planned_date, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, created_at"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	p.Uuid = Random_UUID()
	err = stmt.QueryRow(p.Uuid, p.RiskId, p.UserId, p.Title, p.Description, p.Type, p.Priority, p.Status, p.Equipment, p.PlannedDate, time.Now()).Scan(&p.Id, &p.CreatedAt)
	return err
}

// 创建风险
func (r *Risk) Create() error {
	statement := "INSERT INTO risks (uuid, user_id, name, nickname, keywords, description, source, severity, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	r.Uuid = Random_UUID()
	err = stmt.QueryRow(r.Uuid, r.UserId, r.Name, r.Nickname, r.Keywords, r.Description, r.Source, r.Severity, time.Now()).Scan(&r.Id, &r.CreatedAt)
	return err
}

// 更新风险
func (r *Risk) Update() error {
	statement := "UPDATE risks SET name = $2, nickname = $3, keywords = $4, description = $5, source = $6, severity = $7, updated_at = $8 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(r.Id, r.Name, r.Nickname, r.Keywords, r.Description, r.Source, r.Severity, time.Now())
	return err
}

// 根据ID获取风险
func GetRiskById(id int) (Risk, error) {
	var r Risk
	err := Db.QueryRow("SELECT id, uuid, user_id, name, nickname, keywords, description, source, severity, created_at, updated_at FROM risks WHERE id = $1", id).Scan(&r.Id, &r.Uuid, &r.UserId, &r.Name, &r.Nickname, &r.Keywords, &r.Description, &r.Source, &r.Severity, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}
