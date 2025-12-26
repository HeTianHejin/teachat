# 消息盒子重复问题修复指南

## 问题描述
数据库中存在重复的消息盒子记录，导致同一个团队有多个消息盒子：
```sql
teachat=# select id,object_id,type,is_empty from message_boxes;
 id | object_id | type | is_empty 
----+-----------+------+----------
  1 |         4 |    2 | t
  2 |         4 |    2 | t
  3 |         4 |    2 | t
  4 |         4 |    2 | t
  5 |         4 |    2 | f
  6 |         4 |    2 | t
  7 |         4 |    2 | t
  8 |         4 |    2 | t
  9 |         4 |    2 | t
(9 rows)
```

## 根本原因
1. **并发问题**：多个用户同时访问同一团队时，可能同时创建消息盒子
2. **缺少唯一约束**：数据库表缺少防止重复的约束
3. **错误处理不当**：代码无法区分"记录不存在"和"数据库错误"

## 修复步骤

### 第1步：代码修复（已完成）
已修改以下文件：
- `DAO/message.go`：添加了 `GetOrCreateMessageBox` 方法
- `Route/route_message_box.go`：使用新的安全方法

### 第2步：清理现有重复数据
```bash
psql -d teachat -f sql/fix_duplicate_message_boxes.sql
```

### 第3步：添加数据库约束
```bash
psql -d teachat -f sql/add_unique_constraint_message_boxes.sql
```

## 修复后的行为
1. **原子操作**：`GetOrCreateMessageBox` 方法确保获取或创建操作的原子性
2. **并发安全**：防止多个请求同时创建重复记录
3. **数据库保护**：唯一约束从根本上防止重复
4. **错误处理**：正确区分不同类型的错误

## 验证方法
```sql
-- 检查是否还有重复记录
SELECT object_id, type, COUNT(*) as count
FROM message_boxes 
WHERE deleted_at IS NULL
GROUP BY object_id, type
HAVING COUNT(*) > 1;

-- 检查唯一约束
SELECT conname, contype 
FROM pg_constraint 
WHERE conrelid = 'message_boxes'::regclass;
```

## 注意事项
1. 在执行数据库脚本前，建议先备份数据库
2. 清理重复数据时会保留有消息的记录，优先保留最早创建的
3. 添加约束后，任何尝试创建重复记录的操作都会失败