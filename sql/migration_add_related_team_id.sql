-- 添加related_team_id列到tea_transactions表
-- 用于支持团队转账功能的个人交易流水记录

ALTER TABLE tea_transactions 
ADD COLUMN related_team_id INTEGER REFERENCES teams(id);

-- 添加字段注释
COMMENT ON COLUMN tea_transactions.related_team_id IS '交易相关团队ID（如转账的对方团队）';

-- 创建索引
CREATE INDEX idx_tea_transactions_related_team_id ON tea_transactions(related_team_id);

-- 添加约束检查：related_user_id和related_team_id只能有一个有值
ALTER TABLE tea_transactions 
ADD CONSTRAINT check_tea_transaction_target 
CHECK (
    (related_user_id IS NOT NULL AND related_team_id IS NULL) OR
    (related_user_id IS NULL AND related_team_id IS NOT NULL) OR
    (related_user_id IS NULL AND related_team_id IS NULL) -- 允许两者都为NULL（如系统操作）
);