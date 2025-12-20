-- 添加流通对象类型字段到团队茶叶交易流水表
ALTER TABLE tea.team_transactions 
ADD COLUMN target_type VARCHAR(1) DEFAULT 'u';

-- 添加字段注释
COMMENT ON COLUMN tea.team_transactions.target_type IS '流通对象类型: u-个人, t-团队';

-- 添加约束检查
ALTER TABLE tea.team_transactions 
ADD CONSTRAINT check_team_tea_transaction_target_type 
CHECK (target_type IN ('u', 't'));

-- 创建索引
CREATE INDEX idx_tea_team_transactions_target_type ON tea.team_transactions(target_type);

-- 更新现有记录的target_type
-- 根据related_team_id和related_user_id来判断
UPDATE tea.team_transactions 
SET target_type = 't' 
WHERE transaction_type IN ('transfer_out', 'transfer_in') 
AND related_team_id IS NOT NULL;

UPDATE tea.team_transactions 
SET target_type = 'u' 
WHERE transaction_type IN ('transfer_out') 
AND related_user_id IS NOT NULL 
AND related_team_id IS NULL;

-- 对于其他交易类型，默认设为'u'（操作人）
UPDATE tea.team_transactions 
SET target_type = 'u' 
WHERE target_type IS NULL;