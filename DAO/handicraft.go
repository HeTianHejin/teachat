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
	Uuid         string
	HandicraftId int
	MagicId      int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// SkillSlice 手工艺作业的技能集合id
type HandicraftSkill struct {
	Id           int
	Uuid         string
	HandicraftId int
	SkillId      int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// ToolSlice  手工艺作业的装备或工具（短期租赁的商品）清单单号，完成这个部分作业，可能需要多个工具（装备），例如写一首古诗，需要毛笔、墨、纸、砚台和水，书桌等
type HandicraftTool struct {
	Id           int
	Uuid         string
	HandicraftId int
	ToolId       int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// GoodsSlice 手工艺作业的消耗品，材料，（货物）商品清单单号
type HandicraftGoods struct {
	Id           int
	Uuid         string
	HandicraftId int
	GoodsId      int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// 手工艺作业场所安全隐患，风险因素，例如高空堕落、机械挤压、电击、化学腐蚀……
type HandicraftHazard struct {
	Id           int
	Uuid         string
	HandicraftId int
	HazardId     int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type StrengthLevel int // 体力耗费等级(1-5)

const (
	VeryLowStrength  StrengthLevel = 1
	LowStrength      StrengthLevel = 2
	ModerateStrength StrengthLevel = 3
	HighStrength     StrengthLevel = 4
	VeryHighStrength StrengthLevel = 5
)

type IntelligenceLevel int //智力耗费等级(1-5)

const (
	VeryLowIntelligence  IntelligenceLevel = 1
	LowIntelligence      IntelligenceLevel = 2
	ModerateIntelligence IntelligenceLevel = 3
	HighIntelligence     IntelligenceLevel = 4
	VeryHighIntelligence IntelligenceLevel = 5
)

type DifficultyLevel int //掌握能力的学习课程难度等级(1-5)

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

// 手工艺作业场所自然（物理）环境条件，如：温度，湿度，扬尘，光照，风力，流速……
type HandicraftEnvironment struct {
	Id            int
	Uuid          string
	HandicraftId  int
	EnvironmentId int
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// 手工艺作业场所安全措施，例如：断电，隔离，上锁……
type HandicraftSafetyMeasure struct {
	Id              int
	Uuid            string
	HandicraftId    int
	SafetyMeasureId int
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

// EvidenceSlice  手工艺作业的音视频等视觉证据，默认值为 0，表示没有
type HandicraftEvidence struct {
	Id           int
	Uuid         string
	HandicraftId int
	EvidenceId   int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
