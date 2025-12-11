# 利益回避机制使用示例

## 核心API

### 1. 检查两人是否需要回避
```go
shouldAvoid, err := ShouldAvoidConflict(userId1, userId2)
if shouldAvoid {
    // 需要回避
}
```

### 2. 获取某人的三代以内亲属
```go
relatives, err := GetThreeGenerationUsers(userId)
// 返回所有三代以内亲属的user_id列表
```

### 3. 建立家庭关联
```go
relation := FamilyRelation{
    FamilyId1:    parentFamilyId,
    FamilyId2:    childFamilyId,
    RelationType: FamilyRelationParentChild,
    ConfirmedBy:  userId,
    Status:       FamilyRelationStatusConfirmed,
}
relation.Create()
```

## 实际应用场景

### 场景1：项目投票时自动排除亲属

```go
// 获取有资格投票的成员
func GetEligibleVoters(projectId int, proposerUserId int) ([]User, error) {
    // 1. 获取项目所有成员
    allMembers, err := GetProjectMembers(projectId)
    if err != nil {
        return nil, err
    }
    
    // 2. 过滤掉提案人的三代以内亲属
    eligibleVoters := []User{}
    for _, member := range allMembers {
        shouldAvoid, err := ShouldAvoidConflict(proposerUserId, member.Id)
        if err != nil {
            continue
        }
        
        if !shouldAvoid {
            eligibleVoters = append(eligibleVoters, member)
        }
    }
    
    return eligibleVoters, nil
}
```

### 场景2：审批流程自动跳过亲属

```go
// 获取下一个审批人（跳过有亲属关系的）
func GetNextApprover(applicantUserId int, approverList []int) (int, error) {
    for _, approverId := range approverList {
        shouldAvoid, err := ShouldAvoidConflict(applicantUserId, approverId)
        if err != nil {
            continue
        }
        
        if !shouldAvoid {
            return approverId, nil // 返回第一个无亲属关系的审批人
        }
    }
    
    return 0, errors.New("所有审批人都需要回避")
}
```

### 场景3：团队成员招募时提示亲属关系

```go
// 检查新成员是否与现有成员有亲属关系
func CheckTeamMemberConflict(teamId int, newUserId int) ([]string, error) {
    // 获取团队现有成员
    existingMembers, err := GetTeamMembers(teamId)
    if err != nil {
        return nil, err
    }
    
    conflicts := []string{}
    for _, member := range existingMembers {
        shouldAvoid, err := ShouldAvoidConflict(newUserId, member.UserId)
        if err != nil {
            continue
        }
        
        if shouldAvoid {
            conflicts = append(conflicts, 
                fmt.Sprintf("%s与%s存在亲属关系", 
                    GetUserName(newUserId), 
                    GetUserName(member.UserId)))
        }
    }
    
    return conflicts, nil
}
```

### 场景4：评审委员会组建时自动回避

```go
// 组建评审委员会，自动排除与申请人有亲属关系的评委
func FormReviewCommittee(applicantUserId int, candidateReviewers []int, requiredCount int) ([]int, error) {
    committee := []int{}
    
    for _, reviewerId := range candidateReviewers {
        if len(committee) >= requiredCount {
            break
        }
        
        shouldAvoid, err := ShouldAvoidConflict(applicantUserId, reviewerId)
        if err != nil {
            continue
        }
        
        if !shouldAvoid {
            committee = append(committee, reviewerId)
        }
    }
    
    if len(committee) < requiredCount {
        return nil, fmt.Errorf("无法组建足够的评审委员会，需要%d人，只找到%d人", 
            requiredCount, len(committee))
    }
    
    return committee, nil
}
```

## 不同家庭形式的支持

### 传统家庭
```
祖父母 → 父母 → 子女
系统自动识别三代关系
```

### 同性家庭
```
Alice & Bob (同性伴侣)
    ↓
Charlie (通过精子库)
系统同样识别为父母-子女关系
```

### 单亲家庭
```
单亲妈妈
    ↓
孩子
系统识别为母子关系
```

### 领养家庭
```
生父母家庭 ←领养关系→ 养父母家庭
                        ↓
                    养子女家庭
可以同时识别生父母和养父母关系
```

### 离婚再婚
```
前配偶家庭(已删除) → 现配偶家庭
系统只识别现有关系，除非用户主动声明前家庭关系
```

## 隐私保护机制

### 1. 不公开家庭
```go
family := Family{
    IsOpen: false, // 不公开
}
// 其他人无法搜索到这个家庭
// 但利益回避机制仍然有效
```

### 2. 单方声明
```go
// 乔布斯可以不承认生父母关系
relation := FamilyRelation{
    Status: FamilyRelationStatusUnilateral, // 单方声明
}
// 生父母声明了关系，但乔布斯不确认
// 系统不会将其纳入乔布斯的回避范围
```

### 3. 拒绝关系
```go
relation.Reject(userId)
// 明确拒绝某个家庭关系声明
```

## 测试验证

```bash
# 测试传统家庭回避
go test -v ./DAO -run TestFamilyRelationAvoidance

# 测试同性家庭回避
go test -v ./DAO -run TestSameGenderFamily
```

## 性能优化建议

### 1. 缓存三代亲属列表
```go
// 对于频繁查询的用户，可以缓存其三代亲属列表
cache.Set(fmt.Sprintf("relatives:%d", userId), relatives, 1*time.Hour)
```

### 2. 批量检查回避关系
```go
// 一次性检查多个用户对
func BatchCheckAvoidance(userId int, targetUserIds []int) (map[int]bool, error) {
    relatives, err := GetThreeGenerationUsers(userId)
    if err != nil {
        return nil, err
    }
    
    relativeMap := make(map[int]bool)
    for _, rid := range relatives {
        relativeMap[rid] = true
    }
    
    result := make(map[int]bool)
    for _, targetId := range targetUserIds {
        result[targetId] = relativeMap[targetId]
    }
    
    return result, nil
}
```

## 关键优势

✅ **通用性** - 支持所有家庭形式  
✅ **自动化** - 无需手动维护回避名单  
✅ **准确性** - 基于用户主动登记的关系  
✅ **隐私性** - 尊重用户的隐私选择  
✅ **灵活性** - 支持单方声明和双方确认  
✅ **可追溯** - 保留历史关系记录
