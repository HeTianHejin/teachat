-- 旧数据迁移语句：将 category 字段改为 type，并添加新的 category 字段

-- 1. 添加新的 type 字段
ALTER TABLE handicrafts ADD COLUMN IF NOT EXISTS type INTEGER NOT NULL DEFAULT 1;

-- 2. 将旧的 category 数据迁移到 type 字段
UPDATE handicrafts SET type = category WHERE type = 1;

-- 3. 修改 category 字段为新的含义（公开/私密），默认设为公开
ALTER TABLE handicrafts ALTER COLUMN category SET DEFAULT 0;
UPDATE handicrafts SET category = 0;

-- 4. 删除旧的 category 索引，创建新索引
DROP INDEX IF EXISTS idx_handicrafts_category;
CREATE INDEX idx_handicrafts_type ON handicrafts(type);
CREATE INDEX idx_handicrafts_category ON handicrafts(category);

-- 5. 添加注释
COMMENT ON COLUMN handicrafts.type IS '类型：1轻体力,2中等体力,3重体力,4轻巧力,5中巧力,6重巧力';
COMMENT ON COLUMN handicrafts.category IS '分类：0公开,1私密';
