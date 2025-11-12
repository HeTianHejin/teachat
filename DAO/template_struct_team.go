package data

// 集团邀请团队加盟邀请函
type GroupInvitationBean struct {
	Invitation GroupInvitation
	Group      Group //发出邀请函的集团
	AuthorCEO  User  //集团首席执行官

	InviteUser User   //受邀请的团队CEO
	Team       Team   //手邀请团队
	Status     string //邀请函目前状态
}

// 集团队列页面动态渲染
type GroupDetail struct {
	SessUser  User
	CanManage bool //有管理权

	GroupBean     GroupBean
	TeamBeanSlice []TeamBean
	//FirstTeamBean TeamBean // 集团第一/顶级管理团队（董事会？）
	IsOverTwelve bool
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

// InvitationBean 团队邀请茶友加盟邀请函
type InvitationBean struct {
	Invitation Invitation
	Team       Team   //发出邀请函的团队
	AuthorCEO  User   //撰写邀请函的时任团队首席执行官
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

// 某个茶团的全部成员User邀请函页面数据
type InvitationsPageData struct {
	SessUser User

	Team            Team
	InvitationSlice []Invitation
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

	Team                          Team // 声明所属团队
	TeamMemberRoleNoticeBeanSlice []TeamMemberRoleNoticeBean
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

// 查询某个茶团全部加盟申请书状态列表
type MemberApplicationSlice struct {
	SessUser                   User
	Team                       Team                    //当前茶团
	MemberApplicationBeanSlice []MemberApplicationBean //申请书队列
}
type MemberApplicationBean struct {
	MemberApplication MemberApplication //申请书
	Status            string            //申请书状态

	Team Team //欲加盟的茶团资料

	Author        User   //申请人
	AuthorTeam    Team   // 申请人发帖时选择的团队，或者默认$团队
	CreatedAtDate string //申请时间
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

// 用于茶团队列页面渲染
type TeamSquare struct {
	SessUser User

	IsEmpty bool //是否没有加入任何茶团

	TeamBeanSlice []TeamBean
}
type TeamBean struct {
	Team          Team
	CreatedAtDate string
	//Open                 bool
	Founder              User
	FounderDefaultFamily Family //发起人默认&家庭茶团
	FounderTeam          Team   // 发起人默认所在的团队
	CEO                  User   // CEO
	CEOTeam              Team   // CEO所在默认团队
	CEODefaultFamily     Family // CEO默认&家庭茶团

	MembersCount int //成员计数
}

// 用于某个茶团详情页面渲染
type TeamDetail struct {
	SessUser     User //当前访问用户
	IsFounder    bool //是否为创建者
	IsCEO        bool //是否CEO
	IsCoreMember bool //是否核心成员
	IsMember     bool //是否成员

	TeamBean TeamBean //$事业茶团资料夹

	//IsAuthor              bool
	CoreMemberBeanSlice   []TeamMemberBean //核心成员资料夹
	NormalMemberBeanSlice []TeamMemberBean //普通成员资料夹

	HasApplication bool //是否有新的加盟申请书

	GroupBean *GroupBean //所属集团资料夹（如果有）
}
