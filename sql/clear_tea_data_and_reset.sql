-- 清空茶叶交易相关测试数据并重置新表结构
-- 在执行前请备份重要数据！

-- 1. 清空旧表数据
TRUNCATE TABLE team_tea_operations CASCADE;
TRUNCATE TABLE tea.team_transactions CASCADE;
TRUNCATE TABLE tea_transfers CASCADE;
TRUNCATE TABLE tea_transactions CASCADE;

-- 2. 重置茶叶账户余额（可选，如果需要从零开始测试）
UPDATE tea_accounts SET balance_grams = 0, locked_balance_grams = 0, updated_at = NOW();
UPDATE tea.team_accounts SET balance_grams = 0, locked_balance_grams = 0, updated_at = NOW() WHERE team_id != 1; -- 不包括自由人团队

-- 3. 确保所有必要的茶叶账户存在
INSERT INTO tea_accounts (user_id, balance_grams, status, created_at, updated_at)
SELECT DISTINCT tm.user_id, 0, 'normal', NOW(), NOW()
FROM team_members tm
WHERE tm.user_id NOT IN (SELECT user_id FROM tea_accounts)
  AND tm.status = 'active';

-- 4. 确保所有团队茶叶账户存在（除自由人团队外）
INSERT INTO tea.team_accounts (team_id, balance_grams, status, created_at, updated_at)
SELECT DISTINCT t.id, 0, 'normal', NOW(), NOW()
FROM teams t
WHERE t.id NOT IN (SELECT team_id FROM tea.team_accounts)
  AND t.id != 2  -- 排除自由人团队
  AND t.deleted_at IS NULL;

-- 5. 给测试用户添加一些初始茶叶（可选）
-- 给用户ID为1的用户添加100克茶叶作为测试
UPDATE tea_accounts SET balance_grams = 100 WHERE user_id = 1;

-- 6. 记录系统发放流水（可选）
INSERT INTO tea_transactions (user_id, transaction_type, amount_grams, balance_before, balance_after, description, target_user_id, target_type, created_at)
SELECT user_id, 'system_grant', 100, 0, 100, '系统初始化茶叶', user_id, 'u', NOW()
FROM tea_accounts WHERE user_id = 1 AND balance_grams = 100;

-- 7. 清理孤立的数据
DELETE FROM team_tea_operations WHERE team_id NOT IN (SELECT id FROM teams);
DELETE FROM tea.team_transactions WHERE team_id NOT IN (SELECT id FROM teams);

-- 8. 验证数据一致性
-- 检查账户余额总和
SELECT 
    'tea_accounts' as table_name,
    COUNT(*) as account_count,
    COALESCE(SUM(balance_grams), 0) as total_balance,
    COALESCE(SUM(locked_balance_grams), 0) as total_locked
FROM tea_accounts
UNION ALL
SELECT 
    'tea.team_accounts' as table_name,
    COUNT(*) as account_count,
    COALESCE(SUM(balance_grams), 0) as total_balance,
    COALESCE(SUM(locked_balance_grams), 0) as total_locked
FROM tea.team_accounts
WHERE team_id != 2; -- 排除自由人团队

-- 9. 显示清理完成信息
SELECT 
    '数据清理完成' as status,
    NOW() as completed_at,
    '所有茶叶交易相关表已清空，账户余额已重置，新表结构已准备就绪' as message;