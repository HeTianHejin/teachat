# Family软删除功能快速参考

## 核心API

### 软删除操作
```go
// 软删除家庭
family := Family{Id: familyId}
err := family.SoftDelete()

// 恢复家庭
err := family.Restore()

// 检查是否已删除
isDeleted := family.IsDeleted()
```

### 查询操作
```go
// 查询未删除的家庭（自动过滤已删除）
family, err := GetFamily(familyId)

// 查询包括已删除的家庭
family, err := GetFamilyIncludingDeleted(familyId)

// 获取用户已删除的家庭列表
deletedFamilies, err := GetDeletedFamiliesByAuthorId(userId)

// 获取用户的所有家庭（自动过滤已删除）
families, err := GetAllAuthorFamilies(userId)
```

## 数据库字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `deleted_at` | TIMESTAMP | 软删除时间戳，NULL表示未删除 |

## 数据库索引

- `idx_families_deleted_at`: 优化软删除查询
- `idx_families_author_id`: 优化按作者查询
- `idx_families_author_deleted`: 复合索引，优化按作者和删除状态查询

## 测试验证

```bash
# 运行所有软删除相关测试
go test -v ./DAO -run TestFamilySoftDelete
go test -v ./DAO -run TestFamilyQueryWithSoftDelete
```

## 关键特性

✅ 向后兼容 - 现有代码无需修改  
✅ 自动过滤 - 所有查询自动排除已删除记录  
✅ 可恢复 - 支持恢复已删除的家庭  
✅ 性能优化 - 添加了适当的数据库索引  
✅ 完整测试 - 包含单元测试验证