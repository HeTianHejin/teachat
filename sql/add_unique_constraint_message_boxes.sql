-- 添加唯一约束防止重复消息盒子
-- 在修复现有重复数据后执行此脚本

-- 1. 首先清理重复数据（如果还没有执行的话）
-- 运行 fix_duplicate_message_boxes.sql

-- 2. 创建部分唯一索引，防止同一类型和对象ID出现重复记录（仅对未删除的记录）
CREATE UNIQUE INDEX unique_message_box_type_object 
ON message_boxes (type, object_id) 
WHERE deleted_at IS NULL;

-- 3. 验证索引创建成功
SELECT indexname, tablename 
FROM pg_indexes 
WHERE tablename = 'message_boxes' 
AND indexname = 'unique_message_box_type_object';

-- 4. 测试约束：尝试插入重复记录应该失败
-- BEGIN;
-- INSERT INTO message_boxes (uuid, type, object_id, is_empty, max_count, created_at) 
-- VALUES ('test-uuid', 2, 4, true, 199, CURRENT_TIMESTAMP);
-- INSERT INTO message_boxes (uuid, type, object_id, is_empty, max_count, created_at) 
-- VALUES ('test-uuid-2', 2, 4, true, 199, CURRENT_TIMESTAMP);
-- ROLLBACK;