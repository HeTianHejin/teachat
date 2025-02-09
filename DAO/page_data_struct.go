package data

// UserBeanPagedata 个人主页，渲染所需数据
type UserBean struct {
	SessUser User //会话茶友
	IsAuthor bool //是否作者

	User User // 目标茶友

	DefaultFamilyBean          FamilyBean   //目标茶友默认的家庭茶团资料夹
	ParentMemberFamilyBeanList []FamilyBean //目标茶友管理（作为家长成员）家庭茶团资料夹
	ChildMemberFamilyBeanList  []FamilyBean //作为子女成员家庭茶团资料夹
	OtherMemberFamilyBeanList  []FamilyBean //作为其他成员角色的家庭茶团资料夹
	ResignMemberFamilyBeanList []FamilyBean //目标茶友声明离开的家庭茶团资料夹

	DefaultTeamBean    TeamBean   //目标茶友默认事业茶团资料夹
	ManageTeamBeanList []TeamBean //目标茶友管理的事业茶团资料夹
	JoinTeamBeanList   []TeamBean //目标茶友已加入的事业茶团资料夹
	ResignTeamBeanList []TeamBean //目标茶友已离开的事业茶团资料夹

	DefaultPlace Place //目标茶友首选品茶地点

	Message string // 给目标茶友的通知消息
}

// 个人独白，独角戏资料
type MonologueBean struct {
	Monologue Monologue
	Author    User
	Team      Team
}

// 某个茶话页面渲染所需的动态数据
type PublicPData struct {
	IsAuthor bool // 是否为作者
}

// 某个茶话会详情页面渲染
type ObjectiveDetail struct {
	SessUser              User // 当前会话用户
	SessUserDefaultTeam   Team
	SessUserSurvivalTeams []Team
	SessUserDefaultPlace  Place
	SessUserBindPlaces    []Place
	IsGuest               bool // 是否为游客
	IsInvited             bool // 是否受邀请

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
	AuthorFamily  Family // 作者默认&家庭茶团
	AuthorTeam    Team   // 作者默认$团队
}
type ProjectBean struct {
	Project       Project
	Open          bool
	CreatedAtDate string
	Status        string
	Count         int // 附属对象计数
	Author        User
	AuthorFamily  Family // 作者默认&家庭茶团
	AuthorTeam    Team   // 作者默认团队
	Place         Place  //项目地方
}

// 某个茶台详情页面渲染所需的动态数据
type ProjectDetail struct {
	SessUser              User // 当前会话用户
	SessUserDefaultTeam   Team
	SessUserSurvivalTeams []Team
	SessUserDefaultPlace  Place
	SessUserBindPlaces    []Place
	IsInput               bool // 是否需要显示输入面板
	IsGuest               bool // 是否为游客

	Project      Project //当前浏览茶台
	Master       User    //台主
	MasterFamily Family  // 台主默认的&家庭茶团
	MasterTeam   Team    //台主开台时选择的$事业团队成员身份，所属团队

	Place    Place //茶台(项目)活动地方
	Open     bool
	IsEdited bool

	QuoteObjective             Objective // 引用的茶围
	QuoteObjectiveAuthor       User      // 引用的茶围作者
	QuoteObjectiveAuthorFamily Family    // 引用的茶围作者默认&家庭茶团
	QuoteObjectiveAuthorTeam   Team      // 引用的茶围作者创建时所选择的成员身份，所属团队

	ThreadBeanList        []ThreadBean // project下所有Threads和作者资料荚
	ThreadCount           int          // project下所有Threads个数
	ThreadIsApprovedCount int          //project（茶台）已采纳茶议数量
	IsOverTwelve          bool         //是否超过12个

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
	SessUser              User // 当前会话用户
	IsGuest               bool // 是否为游客
	SessUserDefaultFamily Family
	SessUserDefaultTeam   Team
	SessUserSurvivalTeams []Team
	SessUserDefaultPlace  Place
	SessUserBindPlaces    []Place
	IsInput               bool // 是否需要显示输入面板
	IsPostExist           bool // 是否已经回复过了

	NumSupport int // 支持人数
	NumOppose  int // 反对人数

	ProgressOppose  int // 反对百分比整数
	ProgressSupport int // 支持百分比整数

	QuoteObjective Objective // 引用的茶围

	QuoteProject             Project // 引用的茶台
	QuoteProjectAuthor       User
	QuoteProjectAuthorFamily Family
	QuoteProjectAuthorTeam   Team

	ThreadBean   ThreadBean
	PostBeanList []PostBean // 跟贴豆荚队列

	QuotePost             Post
	QuotePostAuthor       User
	QuotePostAuthorFamily Family
	QuotePostAuthorTeam   Team
}

// 茶议对象和作者资料荚（豆荚一样有许多个单元）
type ThreadBean struct {
	Thread        Thread
	Count         int // 附属对象计数
	Status        string
	CreatedAtDate string

	Author       User   // 作者
	AuthorFamily Family //作者默认&家庭茶团
	AuthorTeam   Team   // 作者默认$团队

	IsMaster bool // 是否为台主
	IsAdmin  bool // 是否为管理员

	Cost       int  // 花费
	TimeSlot   int  // 耗费时间段
	IsApproved bool // 主张方案是否被采纳
}

// 用于跟贴详情页面渲染
type PostDetail struct {
	SessUser              User // 当前会话用户
	IsGuest               bool // 是否为游客
	SessUserDefaultTeam   Team
	SessUserSurvivalTeams []Team
	SessUserDefaultPlace  Place
	SessUserBindPlaces    []Place
	IsAuthor              bool // 是否为品味作者

	PostBean       PostBean
	ThreadBeanList []ThreadBean // 针对此品味的茶议队列

	QuoteThread           Thread // 引用的茶议
	QuoteThreadAuthor     User   // 引用茶议的作者
	QuoteThreadAuthorTeam Team   // 引用茶议的作者所在的默认茶团

	QuoteProject Project // 引用的茶台

	QuoteObjective Objective // 引用的茶围

	IsInput      bool // 是否需要显示输入面板
	IsOverTwelve bool // 是否超过12个
}
type PostBean struct {
	Post          Post
	Count         int    // 附属对象计数
	Attitude      string // 表态立场，支持or反对
	CreatedAtDate string
	Author        User   // 作者
	AuthorFamily  Family // 作者默认&家庭茶团
	AuthorTeam    Team   // 作者默认团队
}

// 用于某个茶团详情页面渲染
type TeamDetail struct {
	SessUser     User //当前访问用户
	IsFounder    bool //是否为创建者
	IsCEO        bool //是否CEO
	IsCoreMember bool //是否核心成员（管理员）
	IsMember     bool //是否成员

	Team                 Team   //茶团
	Founder              User   // 茶团发起人（创建者）
	FounderDefaultFamily Family //发起人默认&家庭茶团
	FounderTeam          Team   // 发起人默认所在的团队
	CEO                  User   // CEO
	CEOTeam              Team   // CEO所在默认团队
	CreatedAtDate        string
	TeamMemberCount      int              //成员数量统计
	CoreMemberBeanList   []TeamMemberBean //核心成员资料夹
	NormalMemberBeanList []TeamMemberBean //普通成员资料夹
	IsAuthor             bool
	Open                 bool //是否开放式茶团

	HasApplication bool //是否有新的加盟申请书
}

// 茶团成员资料荚
type TeamMemberBean struct {
	TeamMember TeamMember

	Member       User
	IsFounder    bool //是否为创建者
	IsCEO        bool //是否CEO
	IsCoreMember bool //是否核心成员（管理员）

	MemberDefaultFamily Family //Member默认&家庭茶团
	MemberDefaultTeam   Team   //First优先茶团
	CreatedAtDate       string
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
	Team                 Team
	CreatedAtDate        string
	Open                 bool
	Founder              User
	FounderDefaultFamily Family //发起人默认&家庭茶团
	FounderTeam          Team   // 发起人默认所在的团队
	Count                int    //成员计数
}

// 用于家庭茶团资料集合页面渲染
type FamilySquare struct {
	SessUser User

	IsEmpty bool //是否为空茶团资料夹队列

	DefaultFamilyBean   FamilyBean   //茶友的默认家庭茶团资料夹
	OtherFamilyBeanList []FamilyBean //其他家庭茶团资料夹队列
}
type FamilyDetail struct {
	SessUser User
	IsParent bool //当前茶友是否为人父母角色？
	IsChild  bool //当前茶友是否为人子女角色？
	IsOther  bool //当前茶友是否为其他类型的家庭成员？

	FamilyBean           FamilyBean
	ParentMemberBeanList []FamilyMemberBean //男主人和女主人
	ChildMemberBeanList  []FamilyMemberBean //孩子们
	OtherMemberBeanList  []FamilyMemberBean //其他类型的家庭成员，例如：猫猫，狗狗……

	IsNewMember        bool               //是否为新成员声明书提及茶友？
	NewMember          User               //新成员声明书提及茶友
	FamilyMemberSignIn FamilyMemberSignIn //提及当前茶友的家庭成员声明书

}
type FamilyBean struct {
	Family      Family
	Founder     User
	FounderTeam Team // 发起人默认所在的团队

	Count int //成员计数
}
type FamilyMemberBean struct {
	FamilyMember FamilyMember
	Member       User
	IsHusband    bool //是否为丈夫
	IsWife       bool //是否为妻子
	IsChild      bool //是否为子女
	IsParent     bool //是否为父母
	IsFounder    bool //是否为创建者

	MemberDefaultFamily Family //当前家庭
	MemberDefaultTeam   Team   //First优先茶团

}

// 申报&家庭茶团新成员页面数据
type FamilyMemberSignInNew struct {
	SessUser              User
	SessUserDefaultFamily Family
	SessUserAllFamilies   []Family
	SessUserDefaultTeam   Team
	SessUserSurvivalTeams []Team
	SessUserDefaultPlace  Place
	SessUserBindPlaces    []Place

	FamilyMemberUser User // 声明为家庭新成员目标茶友
}

// FamilyMemberSignInBeanDetail &家庭茶团新成员声明书详情页面数据
type FamilyMemberSignInDetail struct {
	SessUser User

	FamilyMemberSignInBean FamilyMemberSignInBean //&家庭茶团新成员声明书资料夹
}

// FamilyMemberSignInBean &家庭茶团增加成员声明书资料夹
type FamilyMemberSignInBean struct {
	FamilyMemberSignIn FamilyMemberSignIn //声明书
	Family             Family             //发出声明的家庭
	NewMember          User               //新成员茶友id
	Author             User               //声明书作者
	Place              Place              //声明地点
}
type FamilyMemberSignInBeanList struct {
	SessUser                   User
	FamilyMemberSignInBeanList []FamilyMemberSignInBean //申请书队列
}

// 查询某个茶团全部加盟申请书状态列表
type MemberApplicationList struct {
	SessUser                  User
	Team                      Team                    //当前茶团
	MemberApplicationBeanList []MemberApplicationBean //申请书队列
}
type MemberApplicationBean struct {
	MemberApplication MemberApplication //申请书
	Status            string            //申请书状态

	Team Team //欲加盟的茶团资料

	Author        User   //申请人
	AuthorTeam    Team   // 申请人默认所在的团队
	CreatedAtDate string //申请时间
}

//查询某个用户全部加盟茶团申请书状态列表
// type MemberApplicationListByUser struct {
// 	SessUser                  User
// 	MemberApplicationBeanList []MemberApplicationBean //申请书队列
// }

// 茶团成员退出声明撰写页面数据
type TeamMemberResign struct {
	SessUser User

	Team     Team     //声明人所指的茶团
	TeamBean TeamBean //声明人所指的茶团资料荚

	ResignMember            TeamMember //退出声明茶团成员
	ResignUser              User       //退出声明人
	ResignMemberDefaultTeam Team       //退出声明人默认所属团队

	TeamMemberResignation TeamMemberResignation //退出团队声明书
}

// 好东西，物资清单
type GoodsList struct {
	SessUser  User
	GoodsList []Goods
}

// 物资详情
type GoodsDetail struct {
	SessUser User
	Goods    Goods
	IsAuthor bool
}

// 我的地盘我做主
type PlaceList struct {
	SessUser  User
	PlaceList []Place
}
type PlaceDetail struct {
	SessUser User
	Place    Place
	IsAuthor bool
}

// 用于index页面渲染
type IndexPageData struct {
	SessUser       User         // 当前会话用户
	ThreadBeanList []ThreadBean // Threads和作者资料荚
}

// 用户信箱页面数据
type LetterboxPageData struct {
	SessUser User

	InvitationBeanList []InvitationBean
}

// InvitationBean
type InvitationBean struct {
	Invitation Invitation
	Team       Team   //发出邀请函的团队
	AuthorCEO  User   //团队首席执行官
	InviteUser User   //邀请对象
	Status     string //邀请函目前状态
}

// 茶团加盟邀请函详情页面数据
type InvitationDetail struct {
	SessUser              User
	SessUserDefaultFamily Family
	SessUserAllFamilies   []Family
	SessUserDefaultTeam   Team
	SessUserSurvivalTeams []Team
	SessUserDefaultPlace  Place
	SessUserBindPlaces    []Place

	InvitationBean InvitationBean //邀请函资料夹

}

// 茶团加盟申请书审查页面数据
type ApplicationReview struct {
	SessUser            User
	SessUserDefaultTeam Team //默认所属茶团

	Application MemberApplication //申请书
	Team        Team              // 拟加盟的茶团队

	Applicant            User // 申请人
	ApplicantDefaultTeam Team // 申请人默认所属茶团
}

// 某个茶团的全部邀请函页面数据
type InvitationsPageData struct {
	SessUser User

	Team           Team
	InvitationList []Invitation
}

// 团队成员角色变动声明撰写（new）页面
type TeamMemberRoleChangeNoticePage struct {
	SessUser User
	IsCEO    bool // 会话茶友是否为CEO？

	Team                     Team                     // 声明所属团队
	TeamMemberRoleNoticeBean TeamMemberRoleNoticeBean //团队成员角色变动声明资料夹串
}

// 团队成员角色变动声明集合，浏览（查阅）页面
type TeamMemberRoleChangedNoticesPageData struct {
	SessUser User

	Team                         Team // 声明所属团队
	TeamMemberRoleNoticeBeanList []TeamMemberRoleNoticeBean
}

// 团队成员角色变动声明资料夹
type TeamMemberRoleNoticeBean struct {
	TeamMemberRoleNotice TeamMemberRoleNotice
	Team                 Team //需要调整角色的当前茶团
	Founder              User //团队创建人
	CEO                  User //时任茶团CEO茶友
	Member               User //被调整角色茶友
	MemberDefaultTeam    Team //被调整角色茶友默认所属团队
}

// 接纳茶语消息页面数据
type AcceptMessagePageData struct {
	SessUser          User
	AcceptMessageList []AcceptMessage
}

// 接纳茶语对象页面数据
type AcceptObjectPageData struct {
	SessUser User
	Title    string //标题
	Body     string //内容
	Id       int    //ao_id
}

type ConnectionFriendPageData struct {
	SessUser User
}

// 查询功能，根据关键词查找数据库记录，所得到的数据集合，页面数据
type SearchPageData struct {
	SessUser User

	IsEmpty bool //查询结果为空

	Count int //查询结果个数

	UserBeanList []UserBean //茶友（用户）资料夹队列

	TeamBeanList []TeamBean //茶团资料夹队列
	//ThreadBeanList   []ThreadBean
	PlaceList []Place //品茶地点集合
}
