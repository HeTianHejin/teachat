package dao

import (
	"context"
	"testing"
)

// TestFamilySoftDelete 测试家庭软删除功能
func TestFamilySoftDelete(t *testing.T) {
	// 注意：这个测试需要数据库连接，实际运行时需要确保数据库配置正确

	// 创建测试家庭
	family := Family{
		AuthorId:     1,
		Name:         "测试家庭",
		Introduction: "这是一个测试家庭",
		IsMarried:    true,
		HasChild:     false,
		Status:       FamilyStatusMarried,
		Logo:         "test-logo.png",
		IsOpen:       true,
	}

	// 测试创建
	err := family.Create()
	if err != nil {
		t.Fatalf("创建家庭失败: %v", err)
	}

	// 验证家庭未被删除
	if family.IsDeleted() {
		t.Error("新创建的家庭不应该被标记为已删除")
	}

	// 测试软删除
	err = family.SoftDelete(context.Background())
	if err != nil {
		t.Fatalf("软删除家庭失败: %v", err)
	}

	// 验证家庭已被软删除
	if !family.IsDeleted() {
		t.Error("家庭应该被标记为已删除")
	}

	// 验证DeletedAt字段不为空
	if family.DeletedAt == nil {
		t.Error("DeletedAt字段应该不为空")
	}

	// 测试恢复
	err = family.Restore(context.Background())
	if err != nil {
		t.Fatalf("恢复家庭失败: %v", err)
	}

	// 验证家庭已被恢复
	if family.IsDeleted() {
		t.Error("家庭应该被恢复，不再标记为已删除")
	}

	// 验证DeletedAt字段为空
	if family.DeletedAt != nil {
		t.Error("DeletedAt字段应该为空")
	}
}

// TestFamilyQueryWithSoftDelete 测试带软删除的查询功能
func TestFamilyQueryWithSoftDelete(t *testing.T) {
	// 创建测试家庭
	family := Family{
		AuthorId:     1,
		Name:         "查询测试家庭",
		Introduction: "用于测试查询的家庭",
		IsMarried:    true,
		HasChild:     false,
		Status:       FamilyStatusMarried,
		Logo:         "query-test-logo.png",
		IsOpen:       true,
	}

	// 创建家庭
	err := family.Create()
	if err != nil {
		t.Fatalf("创建家庭失败: %v", err)
	}

	familyId := family.Id

	// 测试正常查询
	retrievedFamily, err := GetFamily(familyId)
	if err != nil {
		t.Fatalf("查询家庭失败: %v", err)
	}
	if retrievedFamily.Id != familyId {
		t.Error("查询到的家庭ID不匹配")
	}

	// 软删除家庭
	err = family.SoftDelete(context.Background())
	if err != nil {
		t.Fatalf("软删除家庭失败: %v", err)
	}

	// 测试查询已删除的家庭（应该返回错误）
	_, err = GetFamily(familyId)
	if err == nil {
		t.Error("查询已删除的家庭应该返回错误")
	}

	// 测试包含已删除家庭的查询
	retrievedFamily, err = GetFamilyIncludingDeleted(familyId)
	if err != nil {
		t.Fatalf("查询包含已删除家庭失败: %v", err)
	}
	if retrievedFamily.Id != familyId {
		t.Error("查询到的家庭ID不匹配")
	}
	if !retrievedFamily.IsDeleted() {
		t.Error("查询到的家庭应该被标记为已删除")
	}
}
