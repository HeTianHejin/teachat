package data

// UserBeanPagedata 个人主页，渲染所需数据
type UserBean struct {
	SessUser User //会话茶友
	IsAuthor bool //是否本人

	User User // 目标茶友

	DefaultFamilyBean           FamilyBean   //目标茶友默认的家庭茶团资料夹
	ParentMemberFamilyBeanSlice []FamilyBean //目标茶友管理（作为家长成员）家庭茶团资料夹
	ChildMemberFamilyBeanSlice  []FamilyBean //作为子女成员家庭茶团资料夹
	OtherMemberFamilyBeanSlice  []FamilyBean //作为其他成员角色的家庭茶团资料夹
	ResignMemberFamilyBeanSlice []FamilyBean //目标茶友声明离开的家庭茶团资料夹

	DefaultTeamBean     TeamBean   //目标茶友默认事业茶团资料夹
	ManageTeamBeanSlice []TeamBean //目标茶友管理的事业茶团资料夹
	JoinTeamBeanSlice   []TeamBean //目标茶友已加入的事业茶团资料夹
	ResignTeamBeanSlice []TeamBean //目标茶友已离开的事业茶团资料夹

	DefaultPlace Place //目标茶友首选品茶地点

	Message string // 给目标茶友的通知消息
}

// 用户数据结构
type UserPageData struct {
	User             User
	DefaultFamily    Family
	SurvivalFamilies []Family
	DefaultTeam      Team
	SurvivalTeams    []Team
	DefaultPlace     Place
	BindPlaces       []Place
}

// 涉及人事统计值集合
type StatsSet struct {
	PersonCount int // 人数
	FamilyCount int // 家庭数
	TeamCount   int // 团队数
}

// 用于家庭茶团资料集合页面渲染
type FamilySquare struct {
	SessUser User

	IsEmpty bool //是否为空茶团资料夹队列

	DefaultFamilyBean    FamilyBean   //茶友的默认家庭茶团资料夹
	OtherFamilyBeanSlice []FamilyBean //其他家庭茶团资料夹队列
}
type FamilyDetail struct {
	SessUser User
	IsParent bool //当前茶友是否为人父母角色？
	IsChild  bool //当前茶友是否为人子女角色？
	IsOther  bool //当前茶友是否为其他类型的家庭成员？

	FamilyBean            FamilyBean
	ParentMemberBeanSlice []FamilyMemberBean //男主人和女主人
	ChildMemberBeanSlice  []FamilyMemberBean //孩子们
	OtherMemberBeanSlice  []FamilyMemberBean //其他类型的家庭成员，例如：猫猫，狗狗……

	IsNewMember        bool               //是否为新成员声明书提及茶友？
	NewMember          User               //新成员声明书提及茶友
	FamilyMemberSignIn FamilyMemberSignIn //提及当前茶友的家庭成员声明书

}
type FamilyBean struct {
	Family      Family
	Founder     User
	FounderTeam Team // 发起人默认所在的团队

	PersonCount int //成员计数
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
type FamilyMemberSignInBeanSlice struct {
	SessUser                    User
	FamilyMemberSignInBeanSlice []FamilyMemberSignInBean //申请书队列
}

//查询某个用户全部加盟茶团申请书状态列表
// type MemberApplicationSliceByUser struct {
// 	SessUser                  User
// 	MemberApplicationBeanSlice []MemberApplicationBean //申请书队列
// }

// 用户信箱页面数据
type LetterboxPageData struct {
	SessUser User

	InvitationBeanSlice []InvitationBean //team邀请函

	GroupInvitationBeanSlice   []GroupInvitationBean //集团邀请函
	GroupInvitationUnreadCount int                   //未读集团邀请函数量
}
