# Family结构软删除功能实现

## 概述
为Family结构添加了软删除功能，允许标记家庭为已删除状态而不是物理删除数据。

## 主要更改

### 1. 结构体更改
- **文件**: `DAO/family.go`
- **更改**: 在`Family`结构体中添加了`DeletedAt *time.Time`字段
- **说明**: 软删除时间戳，NULL表示未删除

### 2. 新增方法

#### 软删除相关方法
- `SoftDelete()`: 软删除家庭
- `Restore()`: 恢复软删除的家庭  
- `IsDeleted()`: 检查家庭是否已被软删除

#### 查询方法
- `GetDeletedFamiliesByAuthorId(authorId int)`: 获取用户已删除的家庭列表
- `GetFamilyIncludingDeleted(family_id int)`: 获取家庭（包括已删除的）

### 3. 现有方法更新

#### 查询方法添加软删除过滤
以下方法已更新，添加了`AND deleted_at IS NULL`条件：
- `Family.Get()`
- `Family.GetByUuid()`
- `GetLastDefaultFamily()`
- `ResignMemberFamilies()`
- `GetAllAuthorFamilies()`
- `CountAllAuthorFamilies()`
- `GetAllFamilies()`
- `GetFamiliesByAuthorId()`
- `queryFamiliesByUserRole()`

#### 扫描函数更新
- `scanFamilies()`: 添加了`DeletedAt`字段的扫描

### 4. 数据库架构更改
- **文件**: `sql/schema.sql`
- **更改**: 
  - 在`families`表中添加`deleted_at TIMESTAMP`字段
  - 添加索引：
    - `idx_families_deleted_at`
    - `idx_families_author_id` 
    - `idx_families_author_deleted`

### 5. 测试文件
- **文件**: `DAO/family_soft_delete_test.go`
- **内容**: 包含软删除功能的单元测试

## 使用示例

### 软删除家庭
```go
family := Family{Id: 1}
err := family.Get()
if err != nil {
    // 处理错误
}

err = family.SoftDelete()
if err != nil {
    // 处理错误
}
```

### 恢复家庭
```go
family := Family{Id: 1}
family, err := GetFamilyIncludingDeleted(1)
if err != nil {
    // 处理错误
}

err = family.Restore()
if err != nil {
    // 处理错误
}
```

### 检查删除状态
```go
if family.IsDeleted() {
    // 家庭已被软删除
}
```

### 获取已删除的家庭
```go
deletedFamilies, err := GetDeletedFamiliesByAuthorId(userId)
if err != nil {
    // 处理错误
}
```

## 注意事项

1. **向后兼容性**: 所有现有的查询方法都会自动过滤已删除的记录
2. **性能**: 添加了适当的数据库索引来优化软删除查询
3. **数据完整性**: 软删除不会影响外键关系
4. **恢复功能**: 可以通过`Restore()`方法恢复已删除的家庭

## 数据库迁移

如果在现有数据库上应用这些更改，需要执行以下SQL：

```sql
-- 添加deleted_at字段
ALTER TABLE families ADD COLUMN deleted_at TIMESTAMP;

-- 添加索引
CREATE INDEX idx_families_deleted_at ON families(deleted_at);
CREATE INDEX idx_families_author_id ON families(author_id);
CREATE INDEX idx_families_author_deleted ON families(author_id, deleted_at);
```

## 测试

运行测试以验证功能：
```bash
go test -v ./DAO -run 'TestFamily.*Delete'
```

或者运行所有软删除相关测试：
```bash
go test -v ./DAO -run TestFamilySoftDelete
go test -v ./DAO -run TestFamilyQueryWithSoftDelete
```