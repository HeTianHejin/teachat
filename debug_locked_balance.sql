-- 检查是否有锁定余额为负数的记录
SELECT 
    'user_accounts' as account_type,
    id,
    user_id,
    balance_grams,
    locked_balance_grams,
    (balance_grams - locked_balance_grams) as available_balance,
    status
FROM tea_accounts 
WHERE locked_balance_grams < 0
UNION ALL
SELECT 
    'team_accounts' as account_type,
    id,
    team_id as user_id,
    balance_grams,
    locked_balance_grams,
    (balance_grams - locked_balance_grams) as available_balance,
    status
FROM team_tea_accounts 
WHERE locked_balance_grams < 0;

-- 检查是否有可用余额为负数的记录
SELECT 
    'user_accounts' as account_type,
    id,
    user_id,
    balance_grams,
    locked_balance_grams,
    (balance_grams - locked_balance_grams) as available_balance,
    status
FROM tea_accounts 
WHERE (balance_grams - locked_balance_grams) < 0
UNION ALL
SELECT 
    'team_accounts' as account_type,
    id,
    team_id as user_id,
    balance_grams,
    locked_balance_grams,
    (balance_grams - locked_balance_grams) as available_balance,
    status
FROM team_tea_accounts 
WHERE (balance_grams - locked_balance_grams) < 0;

-- 检查待确认转账的总金额与锁定余额的关系
SELECT 
    u.id as user_id,
    u.balance_grams,
    u.locked_balance_grams,
    COALESCE(SUM(t.amount_grams), 0) as pending_transfer_amount,
    (u.locked_balance_grams - COALESCE(SUM(t.amount_grams), 0)) as locked_balance_diff
FROM tea_accounts u
LEFT JOIN tea_transfers t ON u.user_id = t.from_user_id 
    AND t.status = 'pending' 
    AND t.expires_at > NOW()
GROUP BY u.id, u.balance_grams, u.locked_balance_grams
HAVING ABS(u.locked_balance_grams - COALESCE(SUM(t.amount_grams), 0)) > 0.001
ORDER BY locked_balance_diff DESC;