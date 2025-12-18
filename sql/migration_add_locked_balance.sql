-- 添加锁定金额字段到用户茶叶账户表
ALTER TABLE tea_accounts ADD COLUMN IF NOT EXISTS locked_balance_grams DECIMAL(15,3) NOT NULL DEFAULT 0.000;

-- 添加锁定金额字段到团队茶叶账户表  
ALTER TABLE team_tea_accounts ADD COLUMN IF NOT EXISTS locked_balance_grams DECIMAL(15,3) NOT NULL DEFAULT 0.000;

-- 添加注释
COMMENT ON COLUMN tea_accounts.locked_balance_grams IS '被锁定的茶叶余额，单位为克，精确到3位小数(毫克)。用于待确认转账期间的资金锁定';
COMMENT ON COLUMN team_tea_accounts.locked_balance_grams IS '团队被锁定的茶叶余额，单位为克，精确到3位小数(毫克)。用于待确认转账期间的资金锁定';

-- 添加约束：锁定金额不能为负数
ALTER TABLE tea_accounts ADD CONSTRAINT check_tea_account_locked_balance_non_negative 
    CHECK (locked_balance_grams >= 0);

ALTER TABLE team_tea_accounts ADD CONSTRAINT check_team_tea_account_locked_balance_non_negative 
    CHECK (locked_balance_grams >= 0);

-- 更新可用余额检查约束：实际可用余额 = balance_grams - locked_balance_grams
-- 删除旧约束
ALTER TABLE tea_accounts DROP CONSTRAINT IF EXISTS check_tea_account_balance_positive;
ALTER TABLE team_tea_accounts DROP CONSTRAINT IF EXISTS check_team_tea_account_balance_positive;

-- 添加新约束：总余额不能为负数
ALTER TABLE tea_accounts ADD CONSTRAINT check_tea_account_balance_positive 
    CHECK (balance_grams >= 0);

ALTER TABLE team_tea_accounts ADD CONSTRAINT check_team_tea_account_balance_positive 
    CHECK (balance_grams >= 0);