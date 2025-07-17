package data

import (
	"time"
)

// Group 集团，n个茶团集合，构成一个大组织。 group = team set
type Group struct {
	Id           int
	Uuid         string
	Name         string
	Abbreviation string // 集团简称
	Mission      string
	FounderId    int
	FirstTeamId  int    // 初创团队，
	Class        int    // 1: "开放式集团",2: "封闭式集团",10: "开放式草集团",20: "封闭式草集团"
	Logo         string // 集团标志
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// GroupMember 集团成员，1 team = 1 member
type GroupMember struct {
	Id      int
	GroupId int
	TeamId  int
	Level   int // 等级 1最高级，2次级，3 次次级，...
	Role    int

	//1 正常（活跃成员）   | Active (normal member)
	//2 暂停（临时限制）   | Suspended (temporary)
	Status    int
	UserId    int //登记用户id
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// Group.CreatedAtDate()
func (group *Group) CreatedAtDate() string {
	return group.CreatedAt.Format(FMT_DATE_CN)
}

// Group.Property()
func (group *Group) Property() string {
	switch group.Class {
	case 1:
		return "开放式集团"
	case 2:
		return "封闭式集团"
	case 10:
		return "开放式草集团"
	case 20:
		return "封闭式草集团"
	default:
		return "未知"
	}
}
