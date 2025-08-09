package data

import "time"

// 限时思考？脑火，指在限时场景下，智力资源运作消耗糖的表观记录
// 烧脑活动，根据“看看”所得线索作出快刀斩乱麻的判断
type BrainFire struct {
	Id        int
	Uuid      string
	ProjectId int
	// 日期时间
	StartTime     time.Time // 开始时间，默认为当前时间
	EndTime       time.Time // 结束时间，默认为开始时间+1小时
	EnvironmentId int       // 环境id

	Title     string
	Inference string //快速推理
	Diagnose  string //诊断、验证
	Judgement string //断言、最终解决方案

	PayerUserId   int //付茶叶代表人Id
	PayerTeamId   int //付茶叶团队Id
	PayerFamilyId int //付茶叶家庭Id

	PayeeUserId   int //收茶叶代表人Id
	PayeeTeamId   int //收茶叶团队Id
	PayeeFamilyId int //收茶叶家庭Id

	VerifierUserId   int
	VerifierFamilyId int
	VerifierTeamId   int
	CreatedAt        time.Time
	UpdatedAt        *time.Time

	//1、未点火
	//2、已点火
	//3、燃烧中
	//4、已熄灭
	Status BrainFireStatus

	//1、文艺类
	//2、理工类
	BrainFireClass int

	//1、公开
	//2、私密（专利？）
	BrainFireType int
}
type BrainFireStatus int

const (
	BrainFireStatusUnlit        BrainFireStatus = iota + 1 // 未点火
	BrainFireStatusLit                                     // 已点火
	BrainFireStatusBurning                                 // 燃烧中
	BrainFireStatusExtinguished                            // 已熄灭
)
