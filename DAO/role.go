package dao

import "strings"

// NormalizeUserRole 把表单/外部输入的申报身份归一化为数据库中的规范值。
// 表单历史上提交过 "agi"/"ufo"，而常量是 "AGI"/"UFO"，此处统一做大小写与空白归一。
func NormalizeUserRole(raw string) (role string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "human":
		return User_Role_Human, true
	case "agi":
		return User_Role_AGI, true
	case "ufo":
		return User_Role_UFO, true
	default:
		return "", false
	}
}

// IsHuman 是否为申报的人类用户。
// 历史数据中存在 captain/traveller 等旧角色值；除明确申报为 AGI/UFO 外，
// 一律视为人类，避免旧账号因角色字符串不同而被当成代理。
func (u User) IsHuman() bool { return !u.IsAGI() && !u.IsUFO() }

// IsAGI 是否为申报的 AGI 代理用户。
func (u User) IsAGI() bool { return u.Role == User_Role_AGI }

// IsUFO 是否为申报的 UFO 用户。
func (u User) IsUFO() bool { return u.Role == User_Role_UFO }

// RoleLabel 返回页面上展示的身份标签。人类用户不显示徽标。
func (u User) RoleLabel() string {
	switch u.Role {
	case User_Role_AGI:
		return "AGI"
	case User_Role_UFO:
		return "UFO"
	default:
		return ""
	}
}
