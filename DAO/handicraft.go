package data

import "time"

// 法力，构思，把一个主意（有意义的想法）转化为一个实例（应用成品）的思考
// 了不起的个人技能,例如原创一首优美的格律诗，并且用工具记录下来（草稿）
type Magic struct {
	Id                int
	Uuid              string
	Name              string //例如，红楼梦的“会作诗”就是一种法力，“填词”是另一种类似法力，解数学方程，找到设备故障能力，debug也是一种法力？？
	Nickname          string
	Description       string //对此法力的描述，例如：七步成诗，能构建出解决某种疑难问题思路方法
	IntelligenceLevel int    // 智力耗费等级(1-5)Mental effort level required.大脑神经元消耗葡萄糖数量？解方程比吟诗更耗费脑力？
	DifficultyLevel   int    // 掌握构思能力的学习课程难度等级(1-5)，例如，学习作诗需要先识字（读和写），再大量阅读前人示范优秀诗词，理解，伴随反复思考实验...
	Category          int    // 类型，
	Level             int    // 等级，
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}

// skill of person,performance
// 技能，某种可以量化，培训传递的完成某种作业的能力，相对magic来说，不需要太多的文本资料阅读积累以及复杂的逻辑推理，
// 驾驶车辆，飞行器，安装空调，书法，剪发，和面团，抹墙灰，拧螺丝等，通常是操作机械工具去做功的操作组合？
type Skill struct {
	Id              int
	Uuid            string
	Name            string
	Nickname        string
	Description     string //对此技能的描述，例如：驾驶车辆，飞行器，安装空调等
	StrengthLevel   int    // 体力耗费等级(1-5)，肌肉消耗能量数量？
	DifficultyLevel int    // 掌握作业能力的学习课程难度等级(1-5)，例如，学习驾驶汽车需要先识字，再阅读理解交通规则，伴随反复上公共道路积累行驶经验...
	Category        int    // 分类，
	Level           int    // 等级，
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

// Magic.Create() 创建法力
func (m *Magic) Create() (err error) {
	statement := "INSERT INTO magics (uuid, name, nickname, description, intelligence_level, difficulty_level, category, level, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), m.Name, m.Nickname, m.Description, m.IntelligenceLevel, m.DifficultyLevel, m.Category, m.Level, time.Now()).Scan(&m.Id, &m.Uuid)
	if err != nil {
		return
	}
	return
}

// Magic.Get() 根据id获取法力
func (m *Magic) Get() (err error) {
	statement := "SELECT id, uuid, name, nickname, description, intelligence_level, difficulty_level, category, level, created_at, updated_at FROM magics WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(m.Id).Scan(&m.Id, &m.Uuid, &m.Name, &m.Nickname, &m.Description, &m.IntelligenceLevel, &m.DifficultyLevel, &m.Category, &m.Level, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// Magic.GetByUuid() 根据uuid获取法力
func (m *Magic) GetByUuid() (err error) {
	statement := "SELECT id, uuid, name, nickname, description, intelligence_level, difficulty_level, category, level, created_at, updated_at FROM magics WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(m.Uuid).Scan(&m.Id, &m.Uuid, &m.Name, &m.Nickname, &m.Description, &m.IntelligenceLevel, &m.DifficultyLevel, &m.Category, &m.Level, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// Skill.Create() 创建技能
func (s *Skill) Create() (err error) {
	statement := "INSERT INTO skills (uuid, name, nickname, description, strength_level, difficulty_level, category, level, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), s.Name, s.Nickname, s.Description, s.StrengthLevel, s.DifficultyLevel, s.Category, s.Level, time.Now()).Scan(&s.Id)
	if err != nil {
		return
	}
	return
}

// Skill.Get() 根据id获取技能记录
func (s *Skill) Get() (err error) {
	statement := "SELECT id, uuid, name, nickname, description, strength_level, difficulty_level, category, level, created_at, updated_at FROM skills WHERE id=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.Id).Scan(&s.Id, &s.Uuid, &s.Name, &s.Nickname, &s.Description, &s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return
	}
	return
}
func (s *Skill) GetByUuid() (err error) {
	statement := "SELECT id, uuid, name, nickname, description, strength_level, difficulty_level, category, level, created_at, updated_at FROM skills WHERE uuid=$1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s.Uuid).Scan(&s.Id, &s.Uuid, &s.Name, &s.Nickname, &s.Description, &s.StrengthLevel, &s.DifficultyLevel, &s.Category, &s.Level, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return
	}
	return
}

// 手工艺（作业），技能操作，需要集中注意力身体手眼协调平衡配合完成的动作。
// 例如：制作特色食品，补牙洞，高空攀爬作业...
type Handicraft struct {
	Id             int
	Uuid           string
	RecorderUserId int // 记录人id，ID of the person recording the handicraft details

	Name        string
	Nickname    string
	Description string // 手工艺总览，任务综合描述。例如，在绢纸上用毛笔（沾墨）抄写一首格律诗词（从草稿眷录为成品）。

	ProjectId int  // 发生的茶台ID，项目，
	IsPrivate bool // 服务对象类型，对&家庭（family）=true，对$团队（team）=false.默认是false

	TeamId_Requester   int // 需求（甲方）客户,需求团队ID，客户必须是一个茶团（$team），非团队则“team_id=0”
	FamilyId_Requester int // 需求（甲方）客户,需求家庭ID，客户必须是一个家庭（family），非家庭则“family_id=0”
	TeamId_Provider    int // 报价（乙方）服务提供者团队id,必须是一个茶团（$team），非团队则“team_id=0”
	FamilyId_Provider  int // 报价（乙方）服务提供者家庭ID，必须是一个家庭（family），非家庭则“family_id=0”

	Category        int // 类型，0:日常普通作业，1:非物质文化遗产手艺？
	Status          int // 状态
	DifficultyLevel int // 新增：难度分级

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 手工艺作业成品，
// 可能是半成品，局部成品
// 作业产品商品Id,例子，1或n首写在白纸上的命题古诗，白纸是作业目标，手工艺内容是在纸上留下美丽的墨迹。如果写了多首，每一份诗（可交易标的物）都可以是一个手艺成品。
type HandicraftProduct struct {
	Id           int
	HandicraftId int
	GoodsId      int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// MagicSlice 手工艺作业的法力集合id
type HandicraftMagic struct {
	Id           int
	HandicraftId int
	MagicId      int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// SkillSlice 手工艺作业的技能集合id
type HandicraftSkill struct {
	Id           int
	HandicraftId int
	SkillId      int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// ToolSlice  手工艺作业的装备或工具（短期租赁的商品）清单单号，完成这个部分作业，可能需要多个工具（装备），例如写一首古诗，需要毛笔、墨、纸、砚台和水，书桌等
type HandicraftTool struct {
	Id           int
	HandicraftId int
	ToolId       int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// GoodsSlice 手工艺作业的消耗品，材料，（货物）商品清单单号
type HandicraftGoods struct {
	Id           int
	HandicraftId int
	GoodsId      int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// 手工艺作业场所环境条件，如：温度，湿度，扬尘，光照，风力，流速……
type HandicraftEnvironment struct {
	Id            int
	HandicraftId  int
	EnvironmentId int
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// 对作业环境模糊（口头）记录，用于茶话会交流
// 手工艺作业环境属性
type Environment struct {
	Id      int
	Uuid    string
	Summary string //概述

	Temperature int //温度
	Humidity    int //湿度
	PM25        int //粉尘
	Noise       int //噪声
	Light       int //光照
	Wind        int //风力
	Flow        int //流速
	Rain        int //雨量
	Pressure    int //气压
	Smoke       int //烟雾
	Dust        int //扬尘
	Odor        int //异味:
	Visibility  int //能见度

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 手工艺作业环境属性分级映射
var LevelMaps = map[string]map[int]string{
	// 异味
	"Odor": {
		1: "极臭（Extreme Stench）",
		2: "浓烈臭味（Strong Odor）",
		3: "明显异味（Noticeable Smell）",
		4: "轻微气味（Faint Odor）",
		5: "无异味（Odorless）",
	},
	// 噪声（分贝逻辑：数字越小越安静）
	"Noise": {
		1: "震耳欲聋（Deafening）", // >90dB
		2: "嘈杂（Noisy）",       // 70-90dB
		3: "一般（Moderate）",    // 50-70dB
		4: "安静（Quiet）",       // 30-50dB
		5: "极静（Silent）",      // <30dB
	},
	// 温度（℃）
	"Temperature": {
		1: "极热（Scorching）",   // >40℃
		2: "炎热（Hot）",         // 30-40℃
		3: "舒适（Comfortable）", // 18-30℃
		4: "微凉（Cool）",        // 5-18℃
		5: "寒冷（Freezing）",    // <5℃
	},
	// 湿度（%RH）
	"Humidity": {
		1: "极湿（Suffocating）", // >90%
		2: "潮湿（Humid）",       // 70-90%
		3: "适宜（Balanced）",    // 40-70%
		4: "干燥（Dry）",         // 20-40%
		5: "极干（Arid）",        // <20%
	},
	// PM2.5（μg/m³）
	"PM25": {
		1: "毒雾（Hazardous）",       // >250
		2: "重污染（Very Unhealthy）", // 150-250
		3: "中度污染（Unhealthy）",     // 55-150
		4: "轻度污染（Moderate）",      // 35-55
		5: "优良（Good）",            // <35
	},
	// 光照（Lux）
	"Light": {
		1: "刺眼（Blinding）", // >10,000
		2: "明亮（Bright）",   // 1,000-10,000
		3: "适中（Normal）",   // 300-1,000
		4: "昏暗（Dim）",      // 50-300
		5: "黑暗（Dark）",     // <50
	},
	// 风力（m/s）
	"Wind": {
		1: "飓风（Hurricane）", // >32.7
		2: "强风（Gale）",      // 10.8-32.7
		3: "微风（Breeze）",    // 3.3-10.8
		4: "轻风（Light）",     // 1.5-3.3
		5: "无风（Calm）",      // <1.5
	},
	// 流速（m/s，通用流体）
	"Flow": {
		1: "湍急（Rapid）",    // >3
		2: "较快（Swift）",    // 1-3
		3: "平稳（Steady）",   // 0.3-1
		4: "缓慢（Slow）",     // 0.1-0.3
		5: "静止（Stagnant）", // <0.1
	},
	// 雨量（mm/h）
	"Rain": {
		1: "暴雨（Torrential）", // >50
		2: "大雨（Heavy）",      // 15-50
		3: "中雨（Moderate）",   // 5-15
		4: "小雨（Light）",      // 1-5
		5: "无雨（None）",       // <1
	},
	// 气压（hPa）
	"Pressure": {
		1: "极高（Very High）", // >1030
		2: "偏高（High）",      // 1015-1030
		3: "正常（Normal）",    // 990-1015
		4: "偏低（Low）",       // 970-990
		5: "极低（Very Low）",  // <970
	},
	// 烟雾（浓度指数）
	"Smoke": {
		1: "严重烟雾（Dense）", // 高浓度
		2: "明显烟雾（Thick）", // 中高浓度
		3: "轻度烟雾（Hazy）",  // 可察觉
		4: "微量烟雾（Trace）", // 轻微
		5: "无烟雾（Clear）",  // 无
	},
	// 扬尘（μg/m³）
	"Dust": {
		1: "沙尘暴（Dust Storm）", // >500
		2: "重度扬尘（Heavy）",     // 200-500
		3: "中度扬尘（Moderate）",  // 100-200
		4: "轻度扬尘（Light）",     // 50-100
		5: "无尘（Clean）",       // <50
	},
	// 能见度（km）
	"Visibility": {
		1: "极差（Zero）",      // <0.1
		2: "很差（Poor）",      // 0.1-1
		3: "一般（Fair）",      // 1-5
		4: "良好（Good）",      // 5-10
		5: "极佳（Excellent）", // >10
	},
}

// 根据字段名和等级返回作业环境描述
// 安全获取分级描述（处理无效字段或等级）
func GetLevelDescription(field string, level int) string {
	if level < 1 || level > 5 {
		return "无效等级"
	}
	if m, ok := LevelMaps[field]; ok {
		return m[level]
	}
	return "未知字段：" + field
}

// 手工艺作业场所安全隐患，风险因素，例如高空堕落、机械挤压、电击、化学腐蚀……
type HandicraftSafetyHazard struct {
	Id             int
	HandicraftId   int
	SafetyHazardId int
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

// 作业场所安全隐患
type SafetyHazard struct {
	Id   int
	Uuid string

	Name        string //隐患名称
	Nickname    string //隐患别名
	Keywords    string //隐患关键词
	Description string //隐患描述
	Source      string //隐患来源
	Level       int    //隐患等级
	Class       int    //隐患类型

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 手工艺作业场所安全措施，例如：断电，隔离，上锁……
type HandicraftSafetyMeasure struct {
	Id              int
	HandicraftId    int
	SafetyMeasureId int
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

// EvidenceSlice  手工艺作业的音视频等视觉证据，默认值为 0，表示没有
type HandicraftEvidence struct {
	Id           int
	HandicraftId int
	EvidenceId   int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type StrengthLevel int

const (
	VeryLowStrength  StrengthLevel = 1
	LowStrength      StrengthLevel = 2
	ModerateStrength StrengthLevel = 3
	HighStrength     StrengthLevel = 4
	VeryHighStrength StrengthLevel = 5
)

type IntelligenceLevel int

const (
	VeryLowIntelligence  IntelligenceLevel = 1
	LowIntelligence      IntelligenceLevel = 2
	ModerateIntelligence IntelligenceLevel = 3
	HighIntelligence     IntelligenceLevel = 4
	VeryHighIntelligence IntelligenceLevel = 5
)

type DifficultyLevel int

const (
	VeryLowDifficulty  DifficultyLevel = 1
	LowDifficulty      DifficultyLevel = 2
	ModerateDifficulty DifficultyLevel = 3
	HighDifficulty     DifficultyLevel = 4
	VeryHighDifficulty DifficultyLevel = 5
)

// 手工艺作业开工仪式，到岗准备开工。例如，书法的起手式，准备动手前一刻的快照
type Inauguration struct {
	Id             int
	Uuid           string
	HandicraftId   int // 手工艺Id
	Name           string
	Nickname       string
	ArtistUserId   int    // 手艺人Id。如果是集团，则填first_team_id作为代表。例如，贾宝玉和他的女仆组成一个作古诗小组，如果一个人自己完成，则为单人成员组。
	RecorderUserId int    // 记录人id
	Description    string // 作业内容描述
	EvidenceId     int    // 音视频等视觉证据，默认值为 0，表示没有
	Status         int    // 状态
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

type Tool struct {
	Id             int
	Uuid           string
	HandicraftId   int
	GoodsId        int    //好东西（商品）id，例如，砚台，砂石，水，墨，毛笔，纸，书桌等
	RecorderUserId int    // 记录人id
	Note           string //备注,特别说明
	Category       int    //类型
	Class          int    //分类
	Level          int    //等级
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

// 收尾，手工艺作业结束仪式，离手（场）快照。
type Ending struct {
	Id             int
	Uuid           string
	HandicraftId   int // 手工艺Id
	Name           string
	Nickname       string
	ArtistUserId   int    // 完成这部分作业的手艺人id。每一个人的操作就是一环节，1 part。例如，贾宝玉写下了一首或者几首诗都可以视作一个环节，
	RecorderUserId int    // 记录人id
	Description    string // 作业内容，成就快照描述
	EvidenceId     int    // 默认值为 0，表示没有
	Status         int    // 状态。0:失败作业，1:已完成作业，2:需要延期作业
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

// 凭据，依据，指音视频等视觉证据，证明手工艺作业符合描述的资料,
// 最好能反映作业劳动成就。或者人力消耗、工具的折旧情况。
type Evidence struct {
	Id             int
	Uuid           string
	HandicraftId   int    // 标记属于那一个手工艺，
	Description    string // 描述记录
	RecorderUserId int    // 记录人id
	Note           string //备注,特别说明
	Category       int    //分类：1、图片，2、视频，3、音频，4、其他
	Link           string // 储存链接（地址）
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}
