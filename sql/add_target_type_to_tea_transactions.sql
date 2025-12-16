-- 添加流通对象类型字段到个人茶叶交易流水表
ALTER TABLE tea_transactions 
ADD COLUMN target_type VARCHAR(1) DEFAULT 'u';

-- 添加字段注释
COMMENT ON COLUMN tea_transactions.target_type IS '流通对象类型: u-个人, t-团队';

-- 添加约束检查
ALTER TABLE tea_transactions 
ADD CONSTRAINT check_tea_transaction_target_type 
CHECK (target_type IN ('u', 't'));

-- 创建索引
CREATE INDEX idx_tea_transactions_target_type ON tea_transactions(target_type);

-- 更新现有记录的target_type
-- 根据描述文字和交易类型来判断
UPDATE tea_transactions 
SET target_type = 't' 
WHERE transaction_type = 'transfer_out' 
AND description LIKE '向团队转账:%';

UPDATE tea_transactions 
SET target_type = 'u' 
WHERE transaction_type IN ('transfer_out', 'transfer_in', 'system_grant', 'system_deduct', 'refund') 
AND description NOT LIKE '向团队转账:%';

-- 对于其他交易类型，默认设为'u'（个人账户操作）
UPDATE tea_transactions 
SET target_type = 'u' 
WHERE target_type IS NULL;