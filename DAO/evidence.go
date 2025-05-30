package data

import "time"

// 凭据，依据，指音视频等视觉证据，证明手工艺作业符合描述的资料,
// 最好能反映作业劳动成就。或者人力消耗、工具的折旧情况。
type Evidence struct {
	Id             int
	Uuid           string
	HandicraftId   int    // 标记属于那一个手工艺，
	Description    string // 描述记录
	RecorderUserId int    // 记录人id
	Note           string //备注,特别说明
	Category       int    //分类：1、图片，2、视频，3、音频，4、其他
	Link           string // 储存链接（地址）
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}
