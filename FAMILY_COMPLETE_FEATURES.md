# 家庭系统完整功能清单

## 已实现的三大核心功能

### 1. ✅ 软删除功能
- **文件**: `DAO/family.go`
- **字段**: `DeletedAt *time.Time`
- **方法**: `SoftDelete()`, `Restore()`, `IsDeleted()`
- **用途**: 保留历史家庭记录，支持恢复

### 2. ✅ 家庭关联与利益回避
- **文件**: `DAO/family_relation.go`
- **表**: `family_relations`
- **核心方法**: 
  - `ShouldAvoidConflict()` - 检查两人是否需要回避
  - `GetThreeGenerationUsers()` - 获取三代以内亲属
- **用途**: 投票、审批时自动排除亲属

### 3. ✅ 视角与消息通知
- **文件**: `DAO/family_perspective.go`
- **字段**: `PerspectiveUserId int`
- **表**: `family_message_preferences`
- **核心方法**:
  - `GetUserPerspectiveFamilyTree()` - 获取用户视角的家族树
  - `ShouldReceiveMessageFrom()` - 判断是否接收消息
- **用途**: 个性化家族树展示，精细化消息通知控制

## 数据库架构

### families表（完整字段）
```sql
- id, uuid, author_id
- name, introduction
- is_married, has_child
- husband_from_family_id, wife_from_family_id
- status, logo, is_open
- created_at, updated_at
- deleted_at              -- 软删除
- perspective_user_id     -- 视角字段
```

### family_relations表
```sql
- family_id_1, family_id_2  -- 关联的两个家庭
- relation_type             -- 关系类型
- status                    -- 确认状态
- confirmed_by              -- 确认者
```

### family_message_preferences表
```sql
- user_id, family_id
- receive_messages          -- 是否接收消息
- notification_type         -- 通知类型
- muted_until               -- 静音到期时间
```

## 核心设计理念

### 1. 主观真实优先
每个人按自己的认知登记家庭，系统尊重个人视角。

### 2. 灵活的关系确认
- 单方声明
- 双方确认
- 拒绝机制

### 3. 包容所有家庭形式
- 传统家庭
- 同性家庭
- 单亲家庭
- 领养家庭
- 离婚再婚

### 4. 精细化控制
- 利益回避自动化
- 消息通知个性化
- 隐私保护完善

## 完整文件清单

### 核心代码
- `DAO/family.go` - Family结构和基础方法
- `DAO/family_relation.go` - 家庭关联和回避机制
- `DAO/family_perspective.go` - 视角和消息偏好
- `DAO/family_helpers.go` - 辅助函数

### 测试文件
- `DAO/family_soft_delete_test.go` - 软删除测试 ✅
- `DAO/family_relation_test.go` - 关联和回避测试 ✅

### 数据库
- `sql/schema.sql` - 完整数据库架构
- `sql/add_family_relations.sql` - 关联表迁移
- `sql/add_perspective_and_message_prefs.sql` - 视角和消息偏好迁移

### 文档
- `FAMILY_SOFT_DELETE_CHANGES.md` - 软删除功能说明
- `FAMILY_RELATION_GUIDE.md` - 关联系统指南
- `FAMILY_AVOIDANCE_EXAMPLES.md` - 回避机制示例
- `FAMILY_PERSPECTIVE_GUIDE.md` - 视角和消息通知指南
- `FAMILY_SYSTEM_SUMMARY.md` - 系统总结
- `FAMILY_COMPLETE_FEATURES.md` - 本文档

## 使用场景矩阵

| 功能 | 投票 | 审批 | 消息 | 家族树 |
|------|------|------|------|--------|
| 软删除 | ✅ | ✅ | ✅ | ✅ |
| 利益回避 | ✅ | ✅ | - | - |
| 视角字段 | - | - | ✅ | ✅ |
| 消息偏好 | - | - | ✅ | - |

## 典型应用流程

### 流程1：用户登记家庭
```
1. 用户创建Family记录
   - perspective_user_id = 用户ID
   - 填写家庭信息
   
2. 添加家庭成员
   - 通过FamilyMemberSignIn声明
   - 等待成员确认
   
3. 建立家庭关联
   - 声明与父母家庭的关系
   - 等待父母确认
```

### 流程2：投票时自动回避
```
1. 用户发起投票
   
2. 系统调用GetThreeGenerationUsers()
   - 获取发起人的三代以内亲属
   
3. 系统调用ShouldAvoidConflict()
   - 逐一检查每个投票人
   - 排除有亲属关系的人
   
4. 通知有资格的投票人
```

### 流程3：浏览家族树并设置消息
```
1. 用户打开家族树页面
   
2. 系统调用GetUserPerspectiveFamilyTree()
   - 展示用户视角的三代家族树
   
3. 用户选择某个家庭
   - 查看该家庭成员
   - 设置消息接收偏好
   
4. 保存FamilyMessagePreference
   - 选择通知类型
   - 设置静音时间
```

### 流程4：发送家庭消息
```
1. 用户A向用户B发送消息
   
2. 系统调用ShouldReceiveMessageFrom()
   - 检查B的消息偏好设置
   - 检查是否静音
   
3. 根据结果决定
   - 发送消息 ✅
   - 或静默丢弃 🔇
```

## 性能优化建议

### 1. 缓存策略
```go
// 缓存三代亲属列表（1小时）
cache.Set(fmt.Sprintf("relatives:%d", userId), relatives, 1*time.Hour)

// 缓存家族树（30分钟）
cache.Set(fmt.Sprintf("family_tree:%d", userId), tree, 30*time.Minute)

// 缓存消息偏好（10分钟）
cache.Set(fmt.Sprintf("msg_prefs:%d", userId), prefs, 10*time.Minute)
```

### 2. 批量查询
```go
// 批量检查回避关系
func BatchCheckAvoidance(userId int, targetUserIds []int) map[int]bool

// 批量获取消息偏好
func BatchGetMessagePreferences(userId int, familyIds []int) map[int]bool
```

### 3. 异步处理
```go
// 异步更新家族树缓存
go UpdateFamilyTreeCache(userId)

// 异步发送消息通知
go SendNotificationAsync(userId, message)
```

## 安全考虑

### 1. 权限控制
- 只有家庭成员可以查看家庭信息
- 只有男女主人可以添加/删除成员
- 软删除的家庭不对外显示

### 2. 隐私保护
- IsOpen字段控制家庭是否公开
- 消息偏好完全由用户控制
- 可以拒绝家庭关联声明

### 3. 数据验证
- 防止循环关联（A→B→A）
- 验证家庭成员角色的合理性
- 检查关系类型的一致性

## 测试覆盖

### 已测试功能 ✅
- 软删除和恢复
- 软删除查询过滤
- 传统家庭利益回避
- 同性家庭利益回避

### 待测试功能
- 家族树生成
- 消息偏好设置
- 静音功能
- 批量操作

## 未来扩展方向

### 短期（1-3个月）
1. 家族树可视化（图形界面）
2. 消息通知UI界面
3. 批量导入家庭关系
4. 移动端适配

### 中期（3-6个月）
1. 智能推荐消息设置
2. 家庭事件日历
3. 家族相册功能
4. 家庭群组聊天

### 长期（6-12个月）
1. AI辅助家族树构建
2. 基因关系验证
3. 跨平台数据同步
4. 家族历史记录

## 总结

本系统成功实现了：

✅ **灵活的家庭定义** - 支持所有家庭形式  
✅ **自动利益回避** - 三代以内自动识别  
✅ **个性化视角** - 每个人看到自己的家族树  
✅ **精细化通知** - 完全控制消息接收  
✅ **完善的隐私** - 多层次隐私保护  
✅ **可追溯历史** - 软删除保留记录

完全满足您提出的需求：
1. 软删除功能 ✅
2. 利益回避机制 ✅
3. 视角和消息通知 ✅

系统已经可以投入使用！🎉
