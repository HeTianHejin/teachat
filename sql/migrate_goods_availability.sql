-- 迁移脚本：将 goods 表的 availability 字段移动到 goods_families 和 goods_teams 表

-- 为 goods_families 表添加字段
ALTER TABLE goods_families 
ADD COLUMN availability INTEGER DEFAULT 0,
ADD COLUMN updated_at TIMESTAMP;

-- 为 goods_teams 表添加字段  
ALTER TABLE goods_teams 
ADD COLUMN availability INTEGER DEFAULT 0,
ADD COLUMN updated_at TIMESTAMP;

-- 从 goods 表移除 availability 字段
ALTER TABLE goods DROP COLUMN availability;

-- 添加注释
COMMENT ON COLUMN goods_families.availability IS '物资在该家庭中的使用状态：0-可用，1-使用中，2-闲置，3-已报废，4-已遗失，5-已转让';
COMMENT ON COLUMN goods_teams.availability IS '物资在该团队中的使用状态：0-可用，1-使用中，2-闲置，3-已报废，4-已遗失，5-已转让';