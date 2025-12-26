# 家庭关联与利益回避机制

## 核心功能

实现三代以内近亲关系识别，用于投票、审批等场景的自动利益回避。

## 设计理念

1. **主观登记** - 每个人按自己的认知登记家庭
2. **关系确认** - 通过family_relations建立客观联系
3. **自动回避** - 系统自动识别三代以内亲属关系

## 数据结构

### FamilyRelation 家庭关联表

```go
type FamilyRelation struct {
    FamilyId1    int    // 第一个家庭ID
    FamilyId2    int    // 第二个家庭ID  
    RelationType int    // 关系类型
    ConfirmedBy  int    // 确认者
    Status       int    // 0-单方声明，1-双方确认，2-已拒绝
}
```

### 关系类型

| 类型 | 说明 | 示例 |
|------|------|------|
| 1 | 同一家庭不同视角 | 夫妻各自登记同一个家 |
| 2 | 前后家庭 | 离婚再婚 |
| 3 | 父母-子女家庭 | 上下代关系 |
| 4 | 领养关系 | 养父母-养子女 |
| 5 | 兄弟姐妹家庭 | 同代关系 |

## 使用示例

### 1. 建立父母-子女关系

```go
// 张三登记：他来自父母家庭#50，现在建立了自己的家庭#101
relation := FamilyRelation{
    FamilyId1:    50,  // 父母家庭
    FamilyId2:    101, // 张三的家庭
    RelationType: FamilyRelationParentChild,
    ConfirmedBy:  张三UserId,
    Status:       FamilyRelationStatusUnilateral, // 单方声明
}
relation.Create()

// 父母确认
relation.Confirm(父亲UserId)
```

### 2. 同一家庭不同视角

```go
// 张三登记的家庭#101，李四登记的家庭#102，实际是同一个家
relation := FamilyRelation{
    FamilyId1:    101, // 张三视角
    FamilyId2:    102, // 李四视角
    RelationType: FamilyRelationSamePerspective,
    ConfirmedBy:  李四UserId,
    Status:       FamilyRelationStatusConfirmed,
}
relation.Create()
```

### 3. 检查利益回避

```go
// 检查张三和李四是否需要回避
shouldAvoid, err := ShouldAvoidConflict(张三UserId, 李四UserId)
if shouldAvoid {
    // 需要回避，不能参与同一投票/审批
}
```

### 4. 获取三代以内亲属

```go
// 获取张三的所有三代以内亲属
relatives, err := GetThreeGenerationUsers(张三UserId)
// 返回：父母、配偶、子女、兄弟姐妹、祖父母、孙子女等
```

## 适用场景

### ✅ 传统家庭
```
祖父母家庭(#10)
     |
  父母家庭(#50)
     |
  张三家庭(#101)
```

### ✅ 同性家庭
```
Alice & Bob家庭(#201)
- Alice (女主人)
- Bob (女主人)  
- Charlie (儿子，通过精子库)
```

### ✅ 单亲家庭
```
单亲妈妈家庭(#301)
- 妈妈 (女主人)
- 孩子 (通过精子库)
```

### ✅ 领养家庭
```
生父母家庭(#50) ----领养关系----> 养父母家庭(#100)
                                      |
                                  乔布斯家庭(#201)
```

### ✅ 离婚再婚
```
前家庭(#103, 已删除) --前后关系--> 现家庭(#101)
```

## 回避规则

系统自动识别以下关系需要回避：

1. **配偶** - 同一家庭的男女主人
2. **父母** - 通过HusbandFromFamilyId/WifeFromFamilyId
3. **子女** - 通过family_relations的父母-子女关系
4. **兄弟姐妹** - 同一父母家庭的子女
5. **祖父母/孙子女** - 三代关系

## 实际应用

### 投票场景

```go
// 项目投票时，自动排除三代以内亲属
func GetEligibleVoters(projectId int, initiatorUserId int) ([]int, error) {
    // 获取所有项目成员
    allMembers := GetProjectMembers(projectId)
    
    // 获取发起人的三代以内亲属
    relatives, _ := GetThreeGenerationUsers(initiatorUserId)
    
    // 排除亲属
    eligibleVoters := []int{}
    for _, member := range allMembers {
        shouldAvoid, _ := ShouldAvoidConflict(initiatorUserId, member)
        if !shouldAvoid {
            eligibleVoters = append(eligibleVoters, member)
        }
    }
    
    return eligibleVoters, nil
}
```

### 审批场景

```go
// 申请审批时，自动跳过有亲属关系的审批人
func GetNextApprover(applicantUserId int, approvers []int) (int, error) {
    for _, approver := range approvers {
        shouldAvoid, _ := ShouldAvoidConflict(applicantUserId, approver)
        if !shouldAvoid {
            return approver, nil // 返回第一个无亲属关系的审批人
        }
    }
    return 0, errors.New("所有审批人都需要回避")
}
```

## 隐私保护

1. **IsOpen字段** - 用户可以选择不公开家庭信息
2. **单方声明** - 可以声明关系但对方不确认（如乔布斯不承认生父母）
3. **软删除** - 历史家庭关系可以软删除但保留记录

## 测试

```bash
# 运行利益回避测试
go test -v ./DAO -run TestFamilyRelationAvoidance

# 运行同性家庭测试
go test -v ./DAO -run TestSameGenderFamily
```

## 数据库迁移

```sql
-- 添加family_relations表
CREATE TABLE family_relations (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    family_id_1           INTEGER REFERENCES families(id),
    family_id_2           INTEGER REFERENCES families(id),
    relation_type         INTEGER NOT NULL,
    confirmed_by          INTEGER REFERENCES users(id),
    status                INTEGER DEFAULT 0,
    note                  TEXT,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 添加索引
CREATE INDEX idx_family_relations_family1 ON family_relations(family_id_1);
CREATE INDEX idx_family_relations_family2 ON family_relations(family_id_2);
CREATE INDEX idx_family_relations_status ON family_relations(status);
CREATE INDEX idx_family_relations_families ON family_relations(family_id_1, family_id_2, status);
```

## 关键优势

✅ 支持所有家庭形式（传统、同性、单亲、领养等）  
✅ 尊重个人主观认知  
✅ 自动识别三代以内关系  
✅ 灵活的确认机制  
✅ 完善的隐私保护  
✅ 适用于投票、审批等多种场景
