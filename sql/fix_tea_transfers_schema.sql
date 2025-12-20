-- 修复 tea_transfers 表结构，添加缺失字段
-- 这个脚本修复个人茶叶转账功能中遇到的字段缺失问题

-- 1. 添加 transfer_type 字段
ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS transfer_type VARCHAR(30) NOT NULL DEFAULT 'personal';

-- 2. 添加 to_team_id 字段以支持团队转账
ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS to_team_id INTEGER REFERENCES teams(id);

-- 3. 添加 from_team_id 字段以支持团队转出
ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS from_team_id INTEGER REFERENCES teams(id);

-- 4. 添加 locked_balance_grams 字段到 tea_accounts（如果不存在）
ALTER TABLE tea_accounts 
ADD COLUMN IF NOT EXISTS locked_balance_grams DECIMAL(15,3) NOT NULL DEFAULT 0.000;

-- 5. 添加 locked_balance_grams 字段到 tea.team.accounts（如果不存在）
ALTER TABLE tea.team.accounts 
ADD COLUMN IF NOT EXISTS locked_balance_grams DECIMAL(15,3) NOT NULL DEFAULT 0.000;

-- 6. 添加审批相关字段
ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS initiator_user_id INTEGER REFERENCES users(id);

ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS approver_user_id INTEGER REFERENCES users(id);

ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP;

ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS approval_rejection_reason TEXT;

-- 7. 添加接收确认相关字段
ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS confirmed_by INTEGER REFERENCES users(id);

ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS confirmed_at TIMESTAMP;

ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS reception_rejection_reason TEXT;

ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS rejected_by INTEGER REFERENCES users(id);

ALTER TABLE tea_transfers 
ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMP;

-- 8. 添加 target_type 字段到 tea_transactions（如果不存在）
ALTER TABLE tea_transactions 
ADD COLUMN IF NOT EXISTS target_team_id INTEGER REFERENCES teams(id);

ALTER TABLE tea_transactions 
ADD COLUMN IF NOT EXISTS target_type VARCHAR(10) NOT NULL DEFAULT 'u';

-- 9. 更新约束条件
-- 删除旧的约束（如果存在）
ALTER TABLE tea_transfers DROP CONSTRAINT IF EXISTS check_tea_transfer_status;

-- 添加新的状态约束（支持更多状态）
ALTER TABLE tea_transfers ADD CONSTRAINT check_tea_transfer_status 
    CHECK (status IN ('pending_approval', 'pending_receipt', 'approved', 'approval_rejected', 'completed', 'rejected', 'expired'));

-- 添加转账类型约束
ALTER TABLE tea_transfers ADD CONSTRAINT check_tea_transfer_type 
    CHECK (transfer_type IN ('personal', 'team_initiated', 'team_approval_required'));

-- 添加目标类型约束
ALTER TABLE tea_transactions ADD CONSTRAINT check_tea_transaction_target_type 
    CHECK (target_type IN ('u', 't'));

-- 10. 更新字段注释
COMMENT ON COLUMN tea_transfers.transfer_type IS '转账类型: personal-个人转账, team_initiated-团队发起转账, team_approval_required-团队转账需审批';
COMMENT ON COLUMN tea_transfers.to_team_id IS '接收方团队ID（团队转账时使用）';
COMMENT ON COLUMN tea_transfers.from_team_id IS '转出方团队ID（团队转出时使用）';
COMMENT ON COLUMN tea_transfers.initiator_user_id IS '发起人ID（团队转账时使用）';
COMMENT ON COLUMN tea_transfers.approver_user_id IS '审批人ID（团队转账时使用）';
COMMENT ON COLUMN tea_transfers.approved_at IS '审批时间';
COMMENT ON COLUMN tea_transfers.approval_rejection_reason IS '审批拒绝原因';
COMMENT ON COLUMN tea_transfers.confirmed_by IS '确认接收人ID';
COMMENT ON COLUMN tea_transfers.confirmed_at IS '确认接收时间';
COMMENT ON COLUMN tea_transfers.reception_rejection_reason IS '接收拒绝原因';
COMMENT ON COLUMN tea_transfers.rejected_by IS '拒绝人ID';
COMMENT ON COLUMN tea_transfers.rejected_at IS '拒绝时间';
COMMENT ON COLUMN tea_accounts.locked_balance_grams IS '被锁定的茶叶数量，单位为克';
COMMENT ON COLUMN tea.team.accounts.locked_balance_grams IS '团队被锁定的茶叶数量，单位为克';
COMMENT ON COLUMN tea_transactions.target_team_id IS '交易相关团队ID';
COMMENT ON COLUMN tea_transactions.target_type IS '目标类型: u-用户, t-团队';

-- 11. 创建必要的索引
CREATE INDEX IF NOT EXISTS idx_tea_transfers_to_team ON tea_transfers(to_team_id);
CREATE INDEX IF NOT EXISTS idx_tea_transfers_from_team ON tea_transfers(from_team_id);
CREATE INDEX IF NOT EXISTS idx_tea_transfers_transfer_type ON tea_transfers(transfer_type);
CREATE INDEX IF NOT EXISTS idx_tea_transactions_target_team ON tea_transactions(target_team_id);
CREATE INDEX IF NOT EXISTS idx_tea_transactions_target_type ON tea_transactions(target_type);

-- 12. 修复现有数据
-- 将所有现有记录的 transfer_type 设置为 'personal'
UPDATE tea_transfers SET transfer_type = 'personal' WHERE transfer_type IS NULL OR transfer_type = '';

-- 将所有现有交易记录的 target_type 设置为 'u'
UPDATE tea_transactions SET target_type = 'u' WHERE target_type IS NULL OR target_type = '';

-- 13. 验证修复结果
SELECT 
    'tea_transfers' as table_name,
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'tea_transfers' AND table_schema = 'public'
    AND column_name IN ('transfer_type', 'to_team_id', 'from_team_id', 'locked_balance_grams', 
                        'initiator_user_id', 'approver_user_id', 'approved_at', 'approval_rejection_reason',
                        'confirmed_by', 'confirmed_at', 'reception_rejection_reason', 'rejected_by', 'rejected_at')
ORDER BY column_name;