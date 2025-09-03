package data

// 某个茶话页面渲染所需的动态数据
type PublicPData struct {
	IsAuthor bool // 是否为作者
}

// 某个茶话会详情页面渲染
type ObjectiveDetail struct {
	SessUser User // 当前会话用户
	//IsAuthor   bool // 是否为作者
	IsAdmin    bool // 是否为管理员
	IsMaster   bool // 是否为管理员
	IsVerifier bool // 是否为见证员
	IsGuest    bool // 是否为游客
	IsInvited  bool // 是否受邀请茶友

	SessUserDefaultFamily    Family
	SessUserSurvivalFamilies []Family

	SessUserDefaultTeam   Team
	SessUserSurvivalTeams []Team

	SessUserDefaultPlace Place
	SessUserBindPlaces   []Place

	ObjectiveBean    ObjectiveBean // 该茶话会资料夹
	ProjectBeanSlice []ProjectBean // objective下所有projects
}

// 茶话会页面（集合）页面渲染所需数据
type ObjectiveSquare struct {
	SessUser           User            // 当前会话用户
	ObjectiveBeanSlice []ObjectiveBean // 广场上所有茶话会
	IsOverTwelve       bool
}
type ObjectiveBean struct {
	Objective     Objective
	Open          bool
	CreatedAtDate string
	Status        string
	Author        User   // 作者
	AuthorFamily  Family // 作者发帖时选择的家庭，
	AuthorTeam    Team   // 作者发帖时选择的团队，

	ProjectCount int // 附属对象茶台计数
}
type ProjectBean struct {
	Project       Project
	Open          bool
	CreatedAtDate string
	Status        string

	Author       User   // 作者
	AuthorFamily Family // 作者发帖时选择的家庭，
	AuthorTeam   Team   // 作者发帖时选择的团队，

	Place Place //项目发生的地方,地点

	ThreadCount int  // 附属对象茶议计数
	IsApproved  bool // 是否入围
}

// 某个茶台详情页面渲染所需的动态数据
type ProjectDetail struct {
	SessUser                 User   // 当前会话用户
	IsAdmin                  bool   // 是否为茶围管理员
	IsMaster                 bool   // 是否为茶台管理员
	IsVerifier               bool   // 是否为见证员
	IsInvited                bool   // 是否受邀请茶友
	IsGuest                  bool   // 是否为游客
	IsInput                  bool   // 是否需要显示输入面板
	SessUserDefaultFamily    Family //当前会话用户默认&家庭茶团
	SessUserSurvivalFamilies []Family
	SessUserDefaultTeam      Team
	SessUserSurvivalTeams    []Team
	SessUserDefaultPlace     Place
	SessUserBindPlaces       []Place

	ProjectBean      ProjectBean             //当前浏览茶台资料夹
	IsApproved       bool                    //是否入围
	Approved6Threads ProjectApproved6Threads //入围茶台必备6茶议

	IsEdited bool

	QuoteObjectiveBean ObjectiveBean // 引用的茶围

	ThreadBeanSlice []ThreadBean // project下普通Threads和作者资料荚

	ThreadCount           int // project下所有Threads个数
	ThreadIsApprovedCount int //project（茶台）已采纳茶议数量

	IsAppointmentCompleted bool // 约茶是否完成
	IsSeeSeekCompleted     bool // 看看是否完成
	IsBrainFireCompleted   bool // 脑火是否完成
	IsSuggestionCompleted  bool // 建议是否完成
	IsGoodsReady           bool // 物资是否备齐
	IsHandicraftCompleted  bool // 手工艺是否完成

	IsOverTwelve bool //是否超过12个
}

// 入围茶台必备6茶议
type ProjectApproved6Threads struct {
	ThreadBeanAppointment     ThreadBean
	ThreadBeanSeeSeekSlice    []ThreadBean
	ThreadBeanBrainFireSlice  []ThreadBean
	ThreadBeanSuggestionSlice []ThreadBean
	ThreadBeanGoodsSlice      []ThreadBean
	ThreadBeanHandicraftSlice []ThreadBean
}

// 茶议草稿页面渲染数据
type DThreadDetail struct {
	SessUser              User
	SessUserDefaultTeam   Team
	SessUserSurvivalTeams []Team

	DraftThread DraftThread
}

// 用于茶议详情页面渲染
type ThreadDetail struct {
	SessUser                 User // 当前会话用户
	SessUserDefaultFamily    Family
	SessUserSurvivalFamilies []Family
	SessUserDefaultTeam      Team
	SessUserSurvivalTeams    []Team
	SessUserDefaultPlace     Place
	SessUserBindPlaces       []Place
	IsGuest                  bool // 是否为游客
	IsAdmin                  bool // 是否为茶围管理员
	IsMaster                 bool // 是否为茶台管理员
	IsVerifier               bool // 是否为见证员
	IsInput                  bool // 是否需要显示新茶议输入面板
	IsPostExist              bool // 是否已经回复过了

	NumSupport int // 支持人数
	NumOppose  int // 反对人数

	ProgressSupport int // 支持率（百分比整数）
	ProgressOppose  int // 反对率（百分比整数）

	QuoteObjectiveBean ObjectiveBean // 引用的茶围（愿景）豆荚

	QuoteProjectBean ProjectBean // 引用的茶台豆荚（实现茶围愿景所需的项目or节点之一）
	QuotePostBean    PostBean    // 引用的品味豆荚（议中议）

	ThreadBean ThreadBean // 当前茶议豆荚

	PostBeanSlice []PostBean // 普通跟贴豆荚队列

	PostBeanAdminSlice []PostBean //茶围管理团队回复切片

	Appointment ProjectAppointment // 约茶
	SeeSeek     SeeSeek            // 看看
	BrainFire   BrainFire          //脑火
	Suggestion  Suggestion         //建议
	Goods       Goods              //物资
	Handicraft  Handicraft         //手工艺

	StatsSet StatsSet //涉及人事统计值集合
}

// 茶议对象和作者资料荚（豆荚一样有许多个单元）
type ThreadBean struct {
	Thread    Thread
	PostCount int // 附属品味计数

	Author       User   // 作者
	AuthorFamily Family //作者发帖时选择的&受益家庭
	AuthorTeam   Team   // 作者创建发帖时选择的$责任团队

	IsApproved bool // 主张方案是否被采纳

	StatsSet StatsSet //涉及人事统计数值集合

}
type ThreadSupplement struct {
	SessUser   User // 当前会话用户
	IsAdmin    bool // 是否为茶围管理员
	IsMaster   bool // 是否为茶台管理员
	IsVerifier bool // 是否为见证员

	QuoteObjectiveBean ObjectiveBean // 引用的茶围（目的/愿景）豆荚

	QuoteProjectBean ProjectBean // 引用的茶台豆荚（实现茶围愿景所需的项目or节点之一）
	QuotePostBean    PostBean    // 引用的品味豆荚（议中议）

	ThreadBean         ThreadBean // 当前茶议豆荚
	PostBeanSlice      []PostBean // 普通跟贴豆荚队列
	PostBeanAdminSlice []PostBean //茶围管理团队回复切片

	AppointmentStatus int     // 约茶状态
	SeeSeek           SeeSeek // SeeSeek
	BrainFire         BrainFire
}

// 用于跟贴详情页面渲染
type PostDetail struct {
	SessUser                 User     // 当前会话用户
	IsGuest                  bool     // 是否为游客
	IsAuthor                 bool     // 是否为品味作者
	IsAdmin                  bool     // 是否为茶围管理成员
	IsMaster                 bool     // 是否为茶台管理成员
	IsVerifier               bool     // 是否为见证员
	SessUserDefaultFamily    Family   // 当前会话用户默认&家庭茶团
	SessUserSurvivalFamilies []Family // 当前会话用户全部&家庭茶团
	SessUserDefaultTeam      Team
	SessUserSurvivalTeams    []Team
	SessUserDefaultPlace     Place
	SessUserBindPlaces       []Place
	IsInput                  bool // 是否需要显示输入面板

	PostBean        PostBean     // 跟贴豆荚
	ThreadBeanSlice []ThreadBean // 针对此品味的茶议队列

	QuoteThreadBean ThreadBean // 引用的茶议豆荚

	QuoteProjectBean ProjectBean // 引用的茶台豆荚

	QuoteObjectiveBean ObjectiveBean // 引用的茶围豆荚

	IsOverTwelve bool // 是否超过12个
}
type PostBean struct {
	Post          Post
	ThreadCount   int    // 附属对象茶议计数
	Attitude      string // 表态立场，支持or反对
	CreatedAtDate string
	Author        User   // 作者
	AuthorFamily  Family // 作者发帖时选择的家庭，或者默认&家庭茶团
	AuthorTeam    Team   // 作者创建发帖时选择的团队，或者默认$团队
}

// 用于index页面渲染
type IndexPageData struct {
	SessUser        User         // 当前会话用户
	ThreadBeanSlice []ThreadBean // Threads和作者资料荚
}

// 接纳茶语消息页面数据
type AcceptMessagePageData struct {
	SessUser           User
	AcceptMessageSlice []AcceptMessage
}

// 接纳茶语对象页面数据
type AcceptObjectPageData struct {
	SessUser User
	Title    string //标题
	Body     string //内容
	Id       int    //ao_id
}
