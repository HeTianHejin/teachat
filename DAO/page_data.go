package data

// UserBiographyPagedata 个人页面数据
type UserBiographyPageData struct {
	SessUser User
	User     User
	IsAuthor bool
	Message  string
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
type ObjectiveDetailPageData struct {
	SessUser    User // 当前会话用户
	Objective   Objective
	Master      User
	Open        bool
	IsEdited    bool
	ProjectList []Project // objective下所有projects
}

// 茶话会页面（集锦）页面渲染所需数据
type ObjectiveSquarePData struct {
	SessUser      User        // 当前会话用户
	ObjectiveList []Objective // 广场上所有茶话会
}

// 某个茶台详情页面渲染所需的动态数据
type ProjectDetailPageData struct {
	SessUser            User // 当前会话用户
	Project             Project
	Master              User
	Open                bool
	IsEdited            bool
	ThreadAndAuthorList []ThreadAndAuthorBean // project下所有Threads和作者资料夹
	ThreadCount         int                   // project下所有Threads个数
	IsOverTwelve        bool                  //是否超过12个
}

// 茶议对象和作者资料夹（豆荚一样有许多个单元）
type ThreadAndAuthorBean struct {
	Thread        Thread
	Count         int // 附属对象计数
	Status        string
	CreatedAtDate string
	Author        User // 作者
	DefaultTeam   Team // 作者默认团队
}

// 茶议草稿页面渲染数据
type DThreadDetailPageData struct {
	SessUser    User
	DraftThread DraftThread
}

// ThreadDetailPageData struct 用于茶议详情页面渲染
type ThreadDetailPageData struct {
	SessUser User // 当前会话用户
	// 反对百分比整数
	ProgressOppose int
	// 支持百分比整数
	ProgressSupport int
	IsInput         bool // 是否需要显示输入面板
	Thread          Thread
	PostList        []Post // 跟贴队列
}

// 用于跟贴详情页面渲染
type PostDetailPageData struct {
	SessUser          User // 当前会话用户
	Post              Post
	IsAuthor          bool     // 是否为品味作者
	QuoteThread       Thread   // 引用的茶议
	QuoteThreadAuthor User     // 引用茶议的作者
	ThreadList        []Thread // 针对此品味的茶议队列
	IsInput           bool     // 是否需要显示输入面板
	IsOverTwelve      bool     // 是否超过12个
}

// 用于茶团详情页面渲染
type TeamDetailPageData struct {
	SessUser             User
	Team                 Team
	Founder              User // 茶团创建者
	TeamMemberCount      int
	CoreMemberDataList   []TeamMemberBean
	NormalMemberDataList []TeamMemberBean
	IsAuthor             bool
	Open                 bool
}

// 茶团成员资料夹
type TeamMemberBean struct {
	User           User
	DefaultTeam    Team
	TeamMemberRole string
}

// 用于茶团队列页面渲染
type TeamsPageData struct {
	SessUser User
	TeamList []Team
}

// 用于index页面渲染
type IndexPageData struct {
	SessUser            User                  // 当前会话用户
	ThreadAndAuthorList []ThreadAndAuthorBean // Threads和作者资料夹
}

// 用户信箱页面数据
type LetterboxPageData struct {
	SessUser       User
	InvitationList []Invitation
}

type InvitationDetailPageData struct {
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
}

type ConnectionFriendPageData struct {
	SessUser User
}
