package data

import (
	"strings"
)

// UpdateFamilyNameWithSpouse 当配偶确认加入家庭时，自动更新家庭名称
// 将占位符"*"替换为实际配偶姓名
func (f *Family) UpdateFamilyNameWithSpouse(spouseUserId int) error {
	// 检查家庭名称是否包含占位符*
	if !strings.Contains(f.Name, "*") {
		return nil // 已经是完整名称，无需更新
	}

	// 获取家庭的父母成员
	parentMembers, err := f.ParentMembers()
	if err != nil {
		return wrapError("UpdateFamilyNameWithSpouse.GetParentMembers", err)
	}

	// 找出男主人和女主人
	var husband, wife *User
	for _, member := range parentMembers {
		memberUser, err := GetUser(member.UserId)
		if err != nil {
			continue
		}

		switch member.Role {
		case FamilyMemberRoleHusband:
			husband = &memberUser
		case FamilyMemberRoleWife:
			wife = &memberUser
		}
	}

	// 生成新的家庭名称
	var newName string
	if husband != nil && wife != nil {
		// 有男女主人，使用"男主人&女主人"格式
		newName = husband.Name + "&" + wife.Name
	} else if husband != nil {
		// 只有男主人
		newName = husband.Name + "&*"
	} else if wife != nil {
		// 只有女主人
		newName = wife.Name + "&*"
	} else {
		// 没有父母成员，保持原名
		return nil
	}

	// 更新家庭名称
	f.Name = newName
	return f.Update()
}

// GetDisplayName 获取家庭显示名称
// 如果包含占位符*，返回友好的显示文本
func (f *Family) GetDisplayName() string {
	if strings.Contains(f.Name, "*") {
		// 替换*为"待确认"
		return strings.Replace(f.Name, "*", "待确认", -1)
	}
	return f.Name
}

// HasPlaceholder 检查家庭名称是否包含占位符
func (f *Family) HasPlaceholder() bool {
	return strings.Contains(f.Name, "*")
}
