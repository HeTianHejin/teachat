package dao

import (
	"context"
	"time"
)

// FamilyRelation 家庭关联关系，用于标识不同Family记录之间的关系
// 主要用途：利益回避机制，识别三代以内近亲关系
type FamilyRelation struct {
	Id           int
	Uuid         string
	FamilyId1    int    // 第一个家庭ID
	FamilyId2    int    // 第二个家庭ID
	RelationType int    // 关系类型
	ConfirmedBy  int    // 确认者用户ID
	Status       int    // 状态：0-单方声明，1-双方确认，2-已拒绝
	Note         string // 关系说明
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time
}

// 家庭关系类型常量
const (
	FamilyRelationSamePerspective = 1 // 同一家庭不同视角（如夫妻各自登记）
	FamilyRelationSequential      = 2 // 前后家庭（离婚再婚）
	FamilyRelationParentChild     = 3 // 父母-子女家庭（上下代）
	FamilyRelationAdoption        = 4 // 领养关系
	FamilyRelationSibling         = 5 // 兄弟姐妹家庭（同代）
)

// 关系状态常量
const (
	FamilyRelationStatusUnilateral = 0 // 单方声明
	FamilyRelationStatusConfirmed  = 1 // 双方确认
	FamilyRelationStatusRejected   = 2 // 已拒绝
)

// GetRelationType 获取关系类型文本
func (fr *FamilyRelation) GetRelationType() string {
	switch fr.RelationType {
	case FamilyRelationSamePerspective:
		return "同一家庭"
	case FamilyRelationSequential:
		return "前后家庭"
	case FamilyRelationParentChild:
		return "父母子女"
	case FamilyRelationAdoption:
		return "领养关系"
	case FamilyRelationSibling:
		return "兄弟姐妹"
	default:
		return "未知"
	}
}

// GetStatus 获取状态文本
func (fr *FamilyRelation) GetStatus() string {
	switch fr.Status {
	case FamilyRelationStatusUnilateral:
		return "待确认"
	case FamilyRelationStatusConfirmed:
		return "已确认"
	case FamilyRelationStatusRejected:
		return "已拒绝"
	default:
		return "未知"
	}
}

// Create 创建家庭关联
func (fr *FamilyRelation) Create() error {
	ctx, cancel := getContext()
	defer cancel()

	query := `INSERT INTO family_relations (uuid, family_id_1, family_id_2, relation_type, 
		confirmed_by, status, note, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, uuid`

	err := DB.QueryRowContext(ctx, query, Random_UUID(), fr.FamilyId1, fr.FamilyId2,
		fr.RelationType, fr.ConfirmedBy, fr.Status, fr.Note, time.Now()).
		Scan(&fr.Id, &fr.Uuid)

	return wrapError("FamilyRelation.Create", err)
}

// Get 根据ID获取家庭关联
func (fr *FamilyRelation) Get() error {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT id, uuid, family_id_1, family_id_2, relation_type, confirmed_by, 
		status, note, created_at, updated_at, deleted_at 
		FROM family_relations WHERE id = $1 AND deleted_at IS NULL`

	err := DB.QueryRowContext(ctx, query, fr.Id).Scan(
		&fr.Id, &fr.Uuid, &fr.FamilyId1, &fr.FamilyId2, &fr.RelationType,
		&fr.ConfirmedBy, &fr.Status, &fr.Note, &fr.CreatedAt, &fr.UpdatedAt, &fr.DeletedAt)

	return wrapError("FamilyRelation.Get", err)
}

// Confirm 确认家庭关联（双方确认）
func (fr *FamilyRelation) Confirm(userId int) error {
	ctx, cancel := getContext()
	defer cancel()

	now := time.Now()
	fr.Status = FamilyRelationStatusConfirmed
	fr.UpdatedAt = &now

	query := `UPDATE family_relations SET status = $1, updated_at = $2 
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := DB.ExecContext(ctx, query, fr.Status, now, fr.Id)
	return wrapError("FamilyRelation.Confirm", err)
}

// Reject 拒绝家庭关联
func (fr *FamilyRelation) Reject(userId int) error {
	ctx, cancel := getContext()
	defer cancel()

	now := time.Now()
	fr.Status = FamilyRelationStatusRejected
	fr.UpdatedAt = &now

	query := `UPDATE family_relations SET status = $1, updated_at = $2 
		WHERE id = $3 AND deleted_at IS NULL`

	_, err := DB.ExecContext(ctx, query, fr.Status, now, fr.Id)
	return wrapError("FamilyRelation.Reject", err)
}

// GetRelatedFamilies 获取与指定家庭相关的所有家庭
func GetRelatedFamilies(familyId int) ([]Family, error) {
	ctx, cancel := getContext()
	defer cancel()

	query := `SELECT DISTINCT f.id, f.uuid, f.author_id, f.name, f.introduction, 
		f.is_married, f.has_child, f.husband_from_family_id, f.wife_from_family_id, 
		f.status, f.created_at, f.updated_at, f.logo, f.is_open, f.deleted_at, f.perspective_user_id
		FROM families f
		INNER JOIN family_relations fr ON (f.id = fr.family_id_1 OR f.id = fr.family_id_2)
		WHERE (fr.family_id_1 = $1 OR fr.family_id_2 = $1) 
		AND fr.status = $2 
		AND fr.deleted_at IS NULL 
		AND f.deleted_at IS NULL 
		AND f.id != $1`

	rows, err := DB.QueryContext(ctx, query, familyId, FamilyRelationStatusConfirmed)
	if err != nil {
		return nil, wrapError("GetRelatedFamilies", err)
	}
	return scanFamilies(rows)
}

// GetThreeGenerationUsers 获取三代以内的所有用户（用于利益回避）
// 包括：父母辈、本人辈、子女辈
func GetThreeGenerationUsers(userId int, ctx context.Context) ([]int, error) {
	// 1. 获取用户所在的所有家庭
	userFamilies, err := GetAllFamilies(userId, ctx)
	if err != nil {
		return nil, err
	}

	userIdMap := make(map[int]bool)
	userIdMap[userId] = true // 包含自己

	for _, family := range userFamilies {
		// 2. 获取当前家庭的所有成员（本人辈）
		members, err := family.AllMembers()
		if err != nil {
			continue
		}
		for _, member := range members {
			userIdMap[member.UserId] = true
		}

		// 3. 获取父母辈家庭
		parentFamilyIds := []int{}
		if family.HusbandFromFamilyId > 0 {
			parentFamilyIds = append(parentFamilyIds, family.HusbandFromFamilyId)
		}
		if family.WifeFromFamilyId > 0 {
			parentFamilyIds = append(parentFamilyIds, family.WifeFromFamilyId)
		}

		// 获取父母辈家庭的所有成员
		for _, parentFamilyId := range parentFamilyIds {
			parentFamily := Family{Id: parentFamilyId}
			if err := parentFamily.Get(); err == nil {
				parentMembers, _ := parentFamily.AllMembers()
				for _, pm := range parentMembers {
					userIdMap[pm.UserId] = true
				}
			}
		}

		// 4. 获取子女辈家庭（通过family_relations）
		query := `SELECT DISTINCT fm.user_id 
			FROM family_relations fr
			INNER JOIN families f ON (f.id = fr.family_id_1 OR f.id = fr.family_id_2)
			INNER JOIN family_members fm ON fm.family_id = f.id
			WHERE (fr.family_id_1 = $1 OR fr.family_id_2 = $1)
			AND fr.relation_type = $2
			AND fr.status = $3
			AND fr.deleted_at IS NULL
			AND f.deleted_at IS NULL`

		rows, err := DB.QueryContext(ctx, query, family.Id,
			FamilyRelationParentChild, FamilyRelationStatusConfirmed)
		if err == nil {
			for rows.Next() {
				var childUserId int
				if err := rows.Scan(&childUserId); err == nil {
					userIdMap[childUserId] = true
				}
			}
			rows.Close()
		}
	}

	// 转换为切片
	result := make([]int, 0, len(userIdMap))
	for uid := range userIdMap {
		result = append(result, uid)
	}

	return result, nil
}

// ShouldAvoidConflict 判断两个用户是否需要利益回避
func ShouldAvoidConflict(userId1, userId2 int, ctx context.Context) (bool, error) {
	if userId1 == userId2 {
		return true, nil
	}

	// 获取userId1的三代以内亲属
	relatives, err := GetThreeGenerationUsers(userId1, ctx)
	if err != nil {
		return false, err
	}

	// 检查userId2是否在其中
	for _, relativeId := range relatives {
		if relativeId == userId2 {
			return true, nil
		}
	}

	return false, nil
}
