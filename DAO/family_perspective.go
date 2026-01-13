package dao

import (
	"context"
	"net/http"
	"time"
)

// FamilyTreeNode 家族树节点
type FamilyTreeNode struct {
	Family       Family
	Generation   int              // 代数：-2祖父母, -1父母, 0本人, 1子女, 2孙子女
	Relationship string           // 关系描述
	Children     []FamilyTreeNode // 子节点
}

// GetUserPerspectiveFamilyTree 获取用户视角的家族树
// 返回以用户为中心的三代家族树
func GetUserPerspectiveFamilyTree(userId int, ctx context.Context) (*FamilyTreeNode, error) {
	// 获取用户所在的所有家庭
	userFamilies, err := GetAllFamilies(userId, ctx)
	if err != nil {
		return nil, err
	}

	if len(userFamilies) == 0 {
		return nil, nil
	}

	// 使用用户的主要家庭（最新的）作为根节点
	rootFamily := userFamilies[0]

	root := &FamilyTreeNode{
		Family:       rootFamily,
		Generation:   0,
		Relationship: "本人家庭",
		Children:     []FamilyTreeNode{},
	}

	// 构建父母辈（-1代）
	if rootFamily.HusbandFromFamilyId > 0 {
		parentFamily := Family{Id: rootFamily.HusbandFromFamilyId}
		if err := parentFamily.Get(); err == nil {
			// 递归获取祖父母辈（-2代）
			grandparents := buildParentGeneration(&parentFamily, -2)

			parentNode := FamilyTreeNode{
				Family:       parentFamily,
				Generation:   -1,
				Relationship: "父母家庭",
				Children:     grandparents,
			}
			root.Children = append(root.Children, parentNode)
		}
	}

	if rootFamily.WifeFromFamilyId > 0 && rootFamily.WifeFromFamilyId != rootFamily.HusbandFromFamilyId {
		parentFamily := Family{Id: rootFamily.WifeFromFamilyId}
		if err := parentFamily.Get(); err == nil {
			grandparents := buildParentGeneration(&parentFamily, -2)

			parentNode := FamilyTreeNode{
				Family:       parentFamily,
				Generation:   -1,
				Relationship: "岳父母家庭",
				Children:     grandparents,
			}
			root.Children = append(root.Children, parentNode)
		}
	}

	// 构建子女辈（+1代）通过family_relations
	childFamilies, err := getChildFamilies(rootFamily.Id)
	if err == nil {
		for _, childFamily := range childFamilies {
			childNode := FamilyTreeNode{
				Family:       childFamily,
				Generation:   1,
				Relationship: "子女家庭",
				Children:     []FamilyTreeNode{},
			}
			root.Children = append(root.Children, childNode)
		}
	}

	return root, nil
}

// buildParentGeneration 构建父母辈节点
func buildParentGeneration(family *Family, generation int) []FamilyTreeNode {
	nodes := []FamilyTreeNode{}

	if family.HusbandFromFamilyId > 0 {
		grandparentFamily := Family{Id: family.HusbandFromFamilyId}
		if err := grandparentFamily.Get(); err == nil {
			node := FamilyTreeNode{
				Family:       grandparentFamily,
				Generation:   generation,
				Relationship: "祖父母家庭",
				Children:     []FamilyTreeNode{},
			}
			nodes = append(nodes, node)
		}
	}

	if family.WifeFromFamilyId > 0 && family.WifeFromFamilyId != family.HusbandFromFamilyId {
		grandparentFamily := Family{Id: family.WifeFromFamilyId}
		if err := grandparentFamily.Get(); err == nil {
			node := FamilyTreeNode{
				Family:       grandparentFamily,
				Generation:   generation,
				Relationship: "外祖父母家庭",
				Children:     []FamilyTreeNode{},
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// getChildFamilies 获取子女家庭
func getChildFamilies(parentFamilyId int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT DISTINCT f.id, f.uuid, f.author_id, f.name, f.introduction, 
		f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, 
		f.status, f.created_at, f.updated_at, f.logo, f.is_open, f.deleted_at, f.perspective_user_id
		FROM families f
		WHERE (f.husband_from_family_id = $1 OR f.wife_from_family_id = $1)
		AND f.deleted_at IS NULL`

	rows, err := DB.QueryContext(ctx, query, parentFamilyId)
	if err != nil {
		return nil, wrapError("getChildFamilies", err)
	}
	return scanFamilies(rows)
}

// FamilyMessagePreference 家庭消息偏好设置
type FamilyMessagePreference struct {
	Id               int
	Uuid             string
	UserId           int        // 用户ID
	FamilyId         int        // 家庭ID
	ReceiveMessages  bool       // 是否接收该家庭成员的消息
	NotificationType int        // 通知类型：0-关闭，1-仅重要，2-全部
	MutedUntil       *time.Time // 静音到期时间
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

// 通知类型常量
const (
	NotificationTypeOff       = 0 // 关闭
	NotificationTypeImportant = 1 // 仅重要
	NotificationTypeAll       = 2 // 全部
)

// Create 创建消息偏好设置
func (fmp *FamilyMessagePreference) Create() error {
	ctx, cancel := getContext()
	defer cancel()

	query := `INSERT INTO family_message_preferences 
		(uuid, user_id, family_id, receive_messages, notification_type, muted_until, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, uuid`

	err := DB.QueryRowContext(ctx, query, Random_UUID(), fmp.UserId, fmp.FamilyId,
		fmp.ReceiveMessages, fmp.NotificationType, fmp.MutedUntil, time.Now()).
		Scan(&fmp.Id, &fmp.Uuid)

	return wrapError("FamilyMessagePreference.Create", err)
}

// Update 更新消息偏好设置
func (fmp *FamilyMessagePreference) Update() error {
	ctx, cancel := getContext()
	defer cancel()

	now := time.Now()
	fmp.UpdatedAt = &now

	query := `UPDATE family_message_preferences 
		SET receive_messages = $1, notification_type = $2, muted_until = $3, updated_at = $4 
		WHERE id = $5`

	_, err := DB.ExecContext(ctx, query, fmp.ReceiveMessages, fmp.NotificationType,
		fmp.MutedUntil, now, fmp.Id)

	return wrapError("FamilyMessagePreference.Update", err)
}

// GetUserFamilyMessagePreferences 获取用户的所有家庭消息偏好
func GetUserFamilyMessagePreferences(userId int) ([]FamilyMessagePreference, error) {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT id, uuid, user_id, family_id, receive_messages, notification_type, 
		muted_until, created_at, updated_at 
		FROM family_message_preferences WHERE user_id = $1`

	rows, err := DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, wrapError("GetUserFamilyMessagePreferences", err)
	}
	defer rows.Close()

	preferences := []FamilyMessagePreference{}
	for rows.Next() {
		var pref FamilyMessagePreference
		err := rows.Scan(&pref.Id, &pref.Uuid, &pref.UserId, &pref.FamilyId,
			&pref.ReceiveMessages, &pref.NotificationType, &pref.MutedUntil,
			&pref.CreatedAt, &pref.UpdatedAt)
		if err != nil {
			continue
		}
		preferences = append(preferences, pref)
	}

	return preferences, rows.Err()
}

// ShouldReceiveMessageFrom 判断用户是否应该接收来自某家庭成员的消息
func ShouldReceiveMessageFrom(userId int, senderUserId int, r *http.Request) (bool, error) {
	// 1. 获取发送者所在的家庭
	senderFamilies, err := GetAllFamilies(senderUserId, r.Context())
	if err != nil {
		return false, err
	}

	// 2. 获取用户的消息偏好设置
	preferences, err := GetUserFamilyMessagePreferences(userId)
	if err != nil {
		return true, nil // 默认接收
	}

	prefMap := make(map[int]*FamilyMessagePreference)
	for i := range preferences {
		prefMap[preferences[i].FamilyId] = &preferences[i]
	}

	// 3. 检查发送者的任一家庭是否在用户的接收列表中
	for _, family := range senderFamilies {
		if pref, exists := prefMap[family.Id]; exists {
			// 检查是否静音
			if pref.MutedUntil != nil && time.Now().Before(*pref.MutedUntil) {
				return false, nil
			}
			return pref.ReceiveMessages, nil
		}
	}

	// 默认接收
	return true, nil
}
