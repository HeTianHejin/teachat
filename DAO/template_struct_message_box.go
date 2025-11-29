package data

// 消息盒子详情页面数据
type MessageBoxDetail struct {
	SessUser     User      // 当前登录用户
	MessageBox   MessageBox // 消息盒子信息
	Team         Team      // 团队信息
	IsMember     bool      // 是否为团队成员
	IsCoreMember bool      // 是否为核心成员
	Messages     []Message // 消息列表
	MessageCount int       // 用户可见的消息总数
}

// 发送纸条页面数据
type MessageSendPageData struct {
	SessUser   User      // 当前登录用户
	Team       Team      // 团队信息
	Receiver   User      // 接收者用户信息
	MessageBox MessageBox // 消息盒子信息
}

// 发送布告页面数据
type MessageAnnouncementSendPageData struct {
	SessUser User // 当前登录用户
	Team     Team // 团队信息
}