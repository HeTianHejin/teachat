package dao

/*
在一个茶围objective目标里，某个项目project被选中“入围”之后，系统将询问围主是否需要启动线下作业服务以解决问题。
-否，没有TeaOrder启动，继续线上讨论；
-是，启动tea-order，这是一个类似”大观园的诗社活动“，有见证方verify team（主持&裁判），需求方payer team（出题），解题方payee team（作答），一个解题的过程是一个handicraft（手工艺），
可能需要多个handicraft才能完成work，为了慎独，另外引入监护方（care team）与解题方共同承担责任风险；
-启动tea-order之后，系统会生成一个tea-order实体，记录该解题服务的相关信息；
-一个tea-order可以包含多个handicraft，每个handicraft对应一个具体的解题任务(见handicraft.go文件)；
(作为存档记录，tea-order涉及的茶围objective与项目project信息及参与方实体将固定（不可删除）以实现历史还原)
。。。
*/
type TeaOrder struct {
	Id           int
	Uuid         string
	ObjectiveId  int    // 茶围目标ID
	ProjectId    int    // 项目ID
	Status       string // tea-order状态：pending/active/completed/cancelled
	VerifyTeamId int    // 见证方团队ID
	PayerTeamId  int    // 需求方团队ID
	PayeeTeamId  int    // 解题方团队ID
	CareTeamId   int    // 监护方团队ID
	Score        int    // 解题评分
	CreatedAt    int64
	UpdatedAt    int64
}

const (
	TeaOrderStatusPending   = "pending"
	TeaOrderStatusActive    = "active"
	TeaOrderStatusCompleted = "completed"
	TeaOrderStatusCancelled = "cancelled"
)
