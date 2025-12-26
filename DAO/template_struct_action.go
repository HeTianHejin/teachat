package dao

// Appointment
type AppointmentTemplateData struct {
	SessUser   User
	IsVerifier bool
	IsAdmin    bool // 是否为茶围目标管理员
	IsMaster   bool
	IsInvited  bool

	QuoteObjectiveBean ObjectiveBean
	ProjectBean        ProjectBean
	AppointmentBean    ProjectAppointmentBean
}
type ProjectAppointmentBean struct {
	Appointment ProjectAppointment

	Project Project

	Payer       User
	PayerFamily Family
	PayerTeam   Team

	Payee       User
	PayeeFamily Family
	PayeeTeam   Team

	Verifier       User
	VerifierFamily Family
	VerifierTeam   Team
}

// 新建“看看”，
// “看看”详情页面数据
type SeeSeekDetailTemplateData struct {
	SessUser       User
	IsVerifier     bool
	IsAdmin        bool
	IsMaster       bool
	IsInvited      bool
	SessUserFamily Family
	SessUserTeam   Team

	Payer       User
	PayerFamily Family
	PayerTeam   Team

	Payee       User
	PayeeFamily Family
	PayeeTeam   Team

	Verifier       User
	VerifierFamily Family
	VerifierTeam   Team

	ProjectAppointment ProjectAppointmentBean //约茶预约资料夹
	SeeSeekBean        SeeSeekBean

	ProjectBean        ProjectBean
	QuoteObjectiveBean ObjectiveBean

	Environments []Environment //场所环境

	// 状态管理相关字段
	CurrentSeeSeek        *SeeSeek //当前进行中的SeeSeek记录
	ExistingEnvironmentId int      //已选择的环境ID
	ExistingHazardIds     []int    //已选择的隐患ID列表
	ExistingRiskIds       []int    //已选择的风险ID列表
	ExistingHazardIdsStr  string   //隐患ID字符串（逗号分隔）
	ExistingRiskIdsStr    string   //风险ID字符串（逗号分隔）
}
type SeeSeekBean struct {
	SeeSeek SeeSeek
	IsOpen  bool

	SeeSeekLook   SeeSeekLook   //观察
	SeeSeekListen SeeSeekListen //聆听
	SeeSeekSmell  SeeSeekSmell  //闻闻（嗅）
	SeeSeekTouch  SeeSeekTouch  //触摸

	SeeSeekExaminationReport []SeeSeekExaminationReport //附加专项检测报告，通常是使用专业工具客观验证

	Environment Environment //场所环境
	Hazard      []Hazard    //场所隐患
	Risk        []Risk      //作业风险
	Goods       []Goods     //物资

	Project Project
	Place   Place
}

// 用于继续“看看”步骤的模版数据
type SeeSeekStepTemplateData struct {
	SessUser   User
	IsVerifier bool
	IsAdmin    bool
	IsMaster   bool

	Verifier       User
	VerifierFamily Family
	VerifierTeam   Team

	SeeSeekBean SeeSeekBean

	ProjectBean        ProjectBean
	QuoteObjectiveBean ObjectiveBean

	// 状态管理相关字段
	CurrentSeeSeek   *SeeSeek //当前进行中的SeeSeek记录
	CompletedSteps   int
	CurrentStep      int
	SeeSeekStepTitle string

	ExistingEnvironmentId int    //已选择的环境ID
	ExistingHazardIds     []int  //已选择的隐患ID列表
	ExistingRiskIds       []int  //已选择的风险ID列表
	ExistingHazardIdsStr  string //隐患ID字符串（逗号分隔）
	ExistingRiskIdsStr    string //风险ID字符串（逗号分隔）

	DefaultHazards []Hazard //默认隐患列表
	DefaultRisks   []Risk   //默认风险列表
}

// 新建“脑火”页面
// 展示“脑火”详情页面数据
type BrainFireDetailTemplateData struct {
	SessUser       User
	IsVerifier     bool
	IsAdmin        bool
	IsMaster       bool
	IsInvited      bool
	SessUserFamily Family
	SessUserTeam   Team

	Payer       User
	PayerFamily Family
	PayerTeam   Team

	Payee       User
	PayeeFamily Family
	PayeeTeam   Team

	Verifier       User
	VerifierFamily Family
	VerifierTeam   Team

	ProjectAppointment ProjectAppointmentBean //约茶预约资料夹
	BrainFireBean      BrainFireBean

	ProjectBean        ProjectBean
	QuoteObjectiveBean ObjectiveBean

	Environments []Environment //场所环境
}
type BrainFireBean struct {
	BrainFire   BrainFire
	IsOpen      bool
	Environment Environment //场所环境
	Project     Project
}

// 新建"建议"页面
// 展示"建议"详情页面数据
type SuggestionDetailTemplateData struct {
	SessUser   User
	IsVerifier bool
	IsAdmin    bool
	IsMaster   bool
	IsInvited  bool

	SuggestionBean     SuggestionBean
	ProjectBean        ProjectBean
	QuoteObjectiveBean ObjectiveBean
}

type SuggestionBean struct {
	Suggestion Suggestion
	IsOpen     bool
	Project    Project
}

type SkillDetailTemplateData struct {
	SessUser User

	Skill Skill
}
type MagicDetailTemplateData struct {
	SessUser User
	// IsVerifier bool
	// IsAdmin    bool
	// IsMaster   bool
	// IsInvited  bool

	Magics Magic
}

type HandicraftDetailTemplateData struct {
	SessUser   User
	IsVerifier bool
	IsAdmin    bool
	IsMaster   bool
	IsInvited  bool

	HandicraftBean     HandicraftBean
	ProjectBean        ProjectBean
	QuoteObjectiveBean ObjectiveBean

	Skills []Skill
	Magics []Magic

	EvidenceHandicraftBean []EvidenceHandicraftBean
}
type HandicraftBean struct {
	Handicraft Handicraft
	IsOpen     bool
	Project    Project

	Contributors   []HandicraftContributor
	Inauguration   *Inauguration
	ProcessRecords []ProcessRecord
	Ending         *Ending
}

type SkillUserBean struct {
	User       User
	Skills     []Skill
	SkillUsers []SkillUser
}

// 法力用户Bean
type MagicUserBean struct {
	User       User
	MagicUsers []MagicUser
	Magics     []Magic
}
type SkillTeamBean struct {
	Team       Team
	Skills     []Skill
	SkillTeams []SkillTeam
}

type MagicTeamBean struct {
	Team       Team
	Magics     []Magic
	MagicTeams []MagicTeam
}
type EvidenceSeeSeekBean struct {
	Evidences []Evidence
	SeeSeek   SeeSeek
}
type EvidenceHandicraftBean struct {
	Evidences  []Evidence
	Handicraft Handicraft
}
