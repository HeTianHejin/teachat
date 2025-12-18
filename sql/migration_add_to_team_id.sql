-- 为 tea_transfers 表添加 to_team_id 列以支持用户向团队转账
-- Migration: Add to_team_id column to tea_transfers table

-- 添加 to_team_id 列
ALTER TABLE tea_transfers ADD COLUMN to_team_id INTEGER REFERENCES teams(id);

-- 添加注释
COMMENT ON COLUMN tea_transfers.to_team_id IS '接收转账的团队ID，如果是用户间转账则为NULL';

-- 创建索引
CREATE INDEX idx_tea_transfers_to_team ON tea_transfers(to_team_id);

-- 添加约束：to_user_id 和 to_team_id 不能同时为空，也不能同时有值
ALTER TABLE tea_transfers ADD CONSTRAINT check_tea_transfer_target 
    CHECK (
        (to_user_id IS NOT NULL AND to_team_id IS NULL) OR 
        (to_user_id IS NULL AND to_team_id IS NOT NULL)
    );

-- 修改原有的 to_user_id 列允许为 NULL
ALTER TABLE tea_transfers ALTER COLUMN to_user_id DROP NOT NULL;

-- 添加注释说明
COMMENT ON COLUMN tea_transfers.to_user_id IS '接收转账的用户ID，如果是向团队转账则为NULL';