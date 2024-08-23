package data

import "time"

// UserBiographyPagedata 个人页面数据
type UserBiography struct {
	SessUser           User
	User               User // 作者资料
	DefaultTeam        TeamBean
	ManageTeamBeanList []TeamBean
	JoinTeamBeanList   []TeamBean
	QuitTeamBeanList   []TeamBean
	IsAuthor           bool
	Message            string // 通知消息
}
type MonologueBean struct {
	Monologue Monologue
	Author    User
	Team      Team
}

// 某个茶话页面渲染所需的动态数据
type PublicPData struct {
	IsAuthor bool // 是否为作者
}

// type ProjectPData struct {
// 	IsAuthor bool // 是否为作者
// }
// type ThreadPData struct {
// 	IsAuthor bool
// }
// type PostPData struct {
// 	IsAuthor bool
// }

// 某个茶话会详情页面渲染
type ObjectiveDetail struct {
	SessUser        User // 当前会话用户
	ObjectiveBean   ObjectiveBean
	ProjectBeanList []ProjectBean // objective下所有projects
}

// 茶话会页面（集锦）页面渲染所需数据
type ObjectiveSquare struct {
	SessUser          User            // 当前会话用户
	ObjectiveBeanList []ObjectiveBean // 广场上所有茶话会
	IsOverTwelve      bool
}
type ObjectiveBean struct {
	Objective     Objective
	Open          bool
	CreatedAtDate string
	Status        string
	Count         int // 附属对象计数
	Author        User
	AuthorTeam    Team // 作者默认团队
}
type ProjectBean struct {
	Project       Project
	Open          bool
	CreatedAtDate string
	Status        string
	Count         int // 附属对象计数
	Author        User
	AuthorTeam    Team // 作者默认团队
}

// 某个茶台详情页面渲染所需的动态数据
type ProjectDetail struct {
	SessUser   User // 当前会话用户
	Project    Project
	Master     User
	MasterTeam Team
	Open       bool
	IsEdited   bool

	QuoteObjective           Objective // 引用的茶围
	QuoteObjectiveAuthor     User      // 引用的茶围作者
	QuoteObjectiveAuthorTeam Team      // 引用的茶围作者所在的默认茶团

	ThreadBeanList []ThreadBean // project下所有Threads和作者资料荚
	ThreadCount    int          // project下所有Threads个数
	IsOverTwelve   bool         //是否超过12个
}

// 茶议草稿页面渲染数据
type DThreadDetail struct {
	SessUser    User
	DraftThread DraftThread
}

// 用于茶议详情页面渲染
type ThreadDetail struct {
	SessUser User // 当前会话用户

	NumSupport int // 支持人数
	NumOppose  int // 反对人数

	ProgressOppose  int // 反对百分比整数
	ProgressSupport int // 支持百分比整数

	QuoteProject           Project
	QuoteProjectAuthor     User
	QuoteProjectAuthorTeam Team

	QuotePost           Post
	QuotePostAuthor     User
	QuotePostAuthorTeam Team

	IsInput      bool // 是否需要显示输入面板
	ThreadBean   ThreadBean
	PostBeanList []PostBean // 跟贴豆荚队列
}

// 茶议对象和作者资料荚（豆荚一样有许多个单元）
type ThreadBean struct {
	Thread        Thread
	Count         int // 附属对象计数
	Status        string
	CreatedAtDate string
	Author        User // 作者
	AuthorTeam    Team // 作者默认团队
}

// 用于跟贴详情页面渲染
type PostDetail struct {
	SessUser              User // 当前会话用户
	PostBean              PostBean
	IsAuthor              bool         // 是否为品味作者
	QuoteThread           Thread       // 引用的茶议
	QuoteThreadAuthor     User         // 引用茶议的作者
	QuoteThreadAuthorTeam Team         // 引用茶议的作者所在的默认茶团
	ThreadBeanList        []ThreadBean // 针对此品味的茶议队列
	IsInput               bool         // 是否需要显示输入面板
	IsOverTwelve          bool         // 是否超过12个
}
type PostBean struct {
	Post          Post
	Count         int    // 附属对象计数
	Attitude      string // 表态立场，支持or反对
	CreatedAtDate string
	Author        User // 作者
	AuthorTeam    Team // 作者默认团队
}

// 用于茶团详情页面渲染
type TeamDetail struct {
	SessUser             User
	Team                 Team
	Founder              User // 茶团创建者
	FounderTeam          Team // 发起人默认所在的团队
	CreatedAtDate        string
	TeamMemberCount      int
	CoreMemberDataList   []TeamMemberBean
	NormalMemberDataList []TeamMemberBean
	IsAuthor             bool
	Open                 bool
}

// 茶团成员资料荚
type TeamMemberBean struct {
	User           User
	AuthorTeam     Team
	CreatedAtDate  string
	TeamMemberRole string
}

// 集团队列页面动态渲染
type GroupDetail struct {
	SessUser      User
	GroupBean     GroupBean
	TeamBeanList  []TeamBean
	FirstTeamBean TeamBean // 集团第一/顶级管理团队（董事会？）
	IsOverTwelve  bool
}

// 集团详情资料荚
type GroupBean struct {
	Group         Group
	CreatedAtDate string
	Open          bool
	Founder       User
	FounderTeam   Team // 发起人默认所在的团队
	TeamsCount    int  // 下属团队计数
	Count         int  // 集团总成员计数，包括全部附属团队的人员数
}

// 用于茶团队列页面渲染
type TeamSquare struct {
	SessUser     User
	TeamBeanList []TeamBean
}
type TeamBean struct {
	Team          Team
	CreatedAtDate string
	Open          bool
	Founder       User
	FounderTeam   Team // 发起人默认所在的团队
	Count         int  //成员计数
}

// 用于index页面渲染
type IndexPageData struct {
	SessUser       User         // 当前会话用户
	ThreadBeanList []ThreadBean // Threads和作者资料荚
}

// 用户信箱页面数据
type LetterboxPageData struct {
	SessUser       User
	InvitationList []Invitation
}

type InvitationDetail struct {
	SessUser User
}

// 某个茶团的全部邀请函页面数据
type InvitationsPageData struct {
	SessUser       User
	Team           Team
	InvitationList []Invitation
}

type AcceptMessagePageData struct {
	SessUser          User
	AcceptMessageList []AcceptMessage
}

type AcceptObjectPageData struct {
	SessUser User
	Title    string //标题
	Body     string //内容
	Id       int    //ao_id
}

type ConnectionFriendPageData struct {
	SessUser User
}

// 动态定位数据
type Location struct {
	Time      time.Time
	Longitude float64 // 经度
	Latitude  float64 // 纬度
	Altitude  float64 // 高度
	Direction float64 // 方向
	Speed     float64 // 速度
	Accuracy  float64 // 精度
	Provider  string  // 供应商
	Address   string  // 邮政地址？航班航线名称？
}
