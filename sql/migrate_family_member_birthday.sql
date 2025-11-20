-- 迁移脚本：将 family_members 表的 age 字段改为 birthday 和 death_date
-- 执行日期：2024

-- 1. 添加新字段
ALTER TABLE family_members 
ADD COLUMN birthday TIMESTAMP,
ADD COLUMN death_date TIMESTAMP;

-- 2. 可选：迁移现有数据（将age转换为大致的birthday）
-- 注意：这只是估算，实际生日需要手动更新
UPDATE family_members 
SET birthday = created_at - (age * INTERVAL '1 year')
WHERE age > 0;

-- 3. 删除旧的 age 字段
ALTER TABLE family_members DROP COLUMN age;

-- 4. 添加注释
COMMENT ON COLUMN family_members.birthday IS '生日，NULL表示未知';
COMMENT ON COLUMN family_members.death_date IS '忌日，NULL表示在世';
