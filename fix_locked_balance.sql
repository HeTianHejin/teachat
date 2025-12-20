-- 修复锁定余额数据不一致的脚本

-- 1. 首先备份相关表
CREATE TABLE tea_accounts_backup AS SELECT * FROM tea_accounts;
CREATE TABLE tea.team.accounts_backup AS SELECT * FROM tea.team.accounts;
CREATE TABLE tea_transfers_backup AS SELECT * FROM tea_transfers;

-- 2. 计算每个用户的实际待确认转账金额
WITH user_pending_transfers AS (
    SELECT 
        from_user_id,
        COALESCE(SUM(amount_grams), 0) as total_pending_amount
    FROM tea_transfers 
    WHERE status = 'pending' AND expires_at > NOW()
    GROUP BY from_user_id
),
-- 3. 更新用户账户的锁定余额为实际的待确认金额
updated_user_accounts AS (
    UPDATE tea_accounts 
    SET locked_balance_grams = COALESCE(upt.total_pending_amount, 0)
    FROM user_pending_transfers upt
    WHERE tea_accounts.user_id = upt.from_user_id
    RETURNING tea_accounts.*
)
-- 4. 处理没有待确认转账的用户，将其锁定余额设为0
UPDATE tea_accounts 
SET locked_balance_grams = 0
WHERE user_id NOT IN (SELECT DISTINCT from_user_id FROM tea_transfers WHERE status = 'pending' AND expires_at > NOW())
  AND locked_balance_grams != 0;

-- 5. 同样的逻辑处理团队账户（如果团队有待确认的转出操作）
-- 注意：根据当前的实现，团队可能不会有转出操作，但为了完整性还是包含

-- 检查修复结果
SELECT 
    'user_accounts_after_fix' as check_type,
    COUNT(*) as total_accounts,
    COUNT(CASE WHEN locked_balance_grams < 0 THEN 1 END) as negative_locked_balance,
    COUNT(CASE WHEN (balance_grams - locked_balance_grams) < 0 THEN 1 END) as negative_available_balance
FROM tea_accounts
UNION ALL
SELECT 
    'team_accounts_after_fix' as check_type,
    COUNT(*) as total_accounts,
    COUNT(CASE WHEN locked_balance_grams < 0 THEN 1 END) as negative_locked_balance,
    COUNT(CASE WHEN (balance_grams - locked_balance_grams) < 0 THEN 1 END) as negative_available_balance
FROM tea.team.accounts;