package dao

import (
	"context"
	"testing"
)

// TestFamilyRelationAvoidance 测试利益回避机制
func TestFamilyRelationAvoidance(t *testing.T) {
	// 场景：张三和李四是夫妻，王五是张三的父亲
	// 在投票时，张三、李四、王五应该互相回避

	// 创建张三的父母家庭
	parentFamily := Family{
		AuthorId:     10, // 王五
		Name:         "王家",
		Introduction: "张三的父母家庭",
		IsMarried:    true,
		Status:       FamilyStatusMarried,
		IsOpen:       true,
	}
	err := parentFamily.Create()
	if err != nil {
		t.Fatalf("创建父母家庭失败: %v", err)
	}

	// 创建张三和李四的家庭（张三视角）
	zhangFamily := Family{
		AuthorId:            1, // 张三
		Name:                "张三&李四的家",
		Introduction:        "张三和李四的家庭",
		IsMarried:           true,
		HusbandFromFamilyId: parentFamily.Id, // 张三来自父母家庭
		Status:              FamilyStatusMarried,
		IsOpen:              true,
	}
	err = zhangFamily.Create()
	if err != nil {
		t.Fatalf("创建张三家庭失败: %v", err)
	}

	// 创建家庭关联：父母-子女关系
	relation := FamilyRelation{
		FamilyId1:    parentFamily.Id,
		FamilyId2:    zhangFamily.Id,
		RelationType: FamilyRelationParentChild,
		ConfirmedBy:  1, // 张三确认
		Status:       FamilyRelationStatusConfirmed,
		Note:         "张三来自王家",
	}
	err = relation.Create()
	if err != nil {
		t.Fatalf("创建家庭关联失败: %v", err)
	}

	// 添加家庭成员
	// 父母家庭：王五（父亲）
	parentMember := FamilyMember{
		FamilyId: parentFamily.Id,
		UserId:   10, // 王五
		Role:     FamilyMemberRoleHusband,
		IsAdult:  true,
	}
	err = parentMember.Create()
	if err != nil {
		t.Fatalf("添加父亲成员失败: %v", err)
	}

	// 张三家庭：张三（丈夫）
	zhangMember := FamilyMember{
		FamilyId: zhangFamily.Id,
		UserId:   1, // 张三
		Role:     FamilyMemberRoleHusband,
		IsAdult:  true,
	}
	err = zhangMember.Create()
	if err != nil {
		t.Fatalf("添加张三成员失败: %v", err)
	}

	// 张三家庭：李四（妻子）
	liMember := FamilyMember{
		FamilyId: zhangFamily.Id,
		UserId:   2, // 李四
		Role:     FamilyMemberRoleWife,
		IsAdult:  true,
	}
	err = liMember.Create()
	if err != nil {
		t.Fatalf("添加李四成员失败: %v", err)
	}

	// 测试利益回避
	shouldAvoid, err := ShouldAvoidConflict(1, 2, context.Background())
	if err != nil {
		t.Fatalf("检查回避关系失败: %v", err)
	}
	if !shouldAvoid {
		t.Error("张三和李四应该回避，但系统判断不需要回避")
	}

	// 2. 张三和王五应该回避（父子）
	shouldAvoid, err = ShouldAvoidConflict(1, 10, context.Background())
	if err != nil {
		t.Fatalf("检查回避关系失败: %v", err)
	}
	if !shouldAvoid {
		t.Error("张三和王五应该回避，但系统判断不需要回避")
	}

	// 3. 李四和王五应该回避（婆媳）
	shouldAvoid, err = ShouldAvoidConflict(2, 10, context.Background())
	if err != nil {
		t.Fatalf("检查回避关系失败: %v", err)
	}
	if !shouldAvoid {
		t.Error("李四和王五应该回避，但系统判断不需要回避")
	}

	// 4. 张三和陌生人不应该回避
	shouldAvoid, err = ShouldAvoidConflict(1, 999, context.Background())
	if err != nil {
		t.Fatalf("检查回避关系失败: %v", err)
	}
	if shouldAvoid {
		t.Error("张三和陌生人不应该回避，但系统判断需要回避")
	}

	t.Log("利益回避机制测试通过")
}

// TestSameGenderFamily 测试同性家庭的回避机制
func TestSameGenderFamily(t *testing.T) {
	// 场景：Alice和Bob是同性伴侣，通过精子库有一个孩子Charlie
	// 在投票时，Alice、Bob、Charlie应该互相回避

	aliceBobFamily := Family{
		AuthorId:     20, // Alice
		Name:         "Alice & Bob的家",
		Introduction: "同性家庭",
		IsMarried:    true,
		Status:       FamilyStatusMarried,
		IsOpen:       true,
	}
	err := aliceBobFamily.Create()
	if err != nil {
		t.Fatalf("创建同性家庭失败: %v", err)
	}

	// 添加成员
	members := []FamilyMember{
		{FamilyId: aliceBobFamily.Id, UserId: 20, Role: FamilyMemberRoleWife, IsAdult: true}, // Alice
		{FamilyId: aliceBobFamily.Id, UserId: 21, Role: FamilyMemberRoleWife, IsAdult: true}, // Bob
		{FamilyId: aliceBobFamily.Id, UserId: 22, Role: FamilyMemberRoleSon, IsAdult: false}, // Charlie
	}

	for _, member := range members {
		if err := member.Create(); err != nil {
			t.Fatalf("添加家庭成员失败: %v", err)
		}
	}

	// 测试回避
	shouldAvoid, err := ShouldAvoidConflict(20, 21, context.Background()) // Alice和Bob
	if err != nil {
		t.Fatalf("检查回避关系失败: %v", err)
	}
	if !shouldAvoid {
		t.Error("Alice和Bob应该回避")
	}

	shouldAvoid, err = ShouldAvoidConflict(20, 22, context.Background()) // Alice和Charlie
	if err != nil {
		t.Fatalf("检查回避关系失败: %v", err)
	}
	if !shouldAvoid {
		t.Error("Alice和Charlie应该回避")
	}

	t.Log("同性家庭回避机制测试通过")
}
