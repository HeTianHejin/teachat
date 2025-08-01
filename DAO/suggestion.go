package data

import "time"

// 建议,
// 根据检查或者踏勘结果，拟采取处置方案
type Suggestion struct {
	Id        int
	Uuid      string
	Title     string
	Body      string
	UserId    int
	ThreadId  int
	Attitude  bool //表态：肯定（颔首）或者否定（摇头）
	FamilyId  int  //作者发帖时选择的家庭id
	TeamId    int  //作者发帖时选择的成员所属茶团id（team/family）
	IsPrivate bool //指定责任（受益）权属类型，代表&家庭（family）=true，代表$团队（team）=false。默认是false
	CreatedAt time.Time
	UpdatedAt *time.Time

	Class  int
	Status int
}
