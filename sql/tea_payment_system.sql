-- ============================================
-- 茶叶支付系统数据库架构定义
-- 为TeaChat添加茶叶账户和转账功能
-- ============================================

-- 创建 tea schema
CREATE SCHEMA IF NOT EXISTS tea;

-- ============================================
-- 第一部分：基础架构定义
-- ============================================

-- 茶叶账户表
CREATE TABLE tea_accounts (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    user_id               INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance_grams         DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 茶叶数量，精确到毫克
    locked_balance_grams  DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 锁定余额
    status                VARCHAR(20) NOT NULL DEFAULT 'normal', -- normal, frozen
    frozen_reason         TEXT, -- 冻结原因
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX idx_tea_accounts_user_id ON tea_accounts(user_id);
CREATE INDEX idx_tea_accounts_status ON tea_accounts(status);

-- 添加表注释
COMMENT ON TABLE tea_accounts IS '用户茶叶账户表';
COMMENT ON COLUMN tea_accounts.balance_grams IS '茶叶余额，单位为克，精确到3位小数(毫克)';
COMMENT ON COLUMN tea_accounts.locked_balance_grams IS '被锁定的茶叶数量，单位为克';
COMMENT ON COLUMN tea_accounts.status IS '账户状态: normal-正常, frozen-冻结';
COMMENT ON COLUMN tea_accounts.frozen_reason IS '账户冻结原因说明';

-- 茶叶转账记录表（扩展版本）
CREATE TABLE tea_transfers (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    from_user_id          INTEGER NOT NULL REFERENCES users(id),
    to_user_id            INTEGER NOT NULL REFERENCES users(id),
    from_team_id          INTEGER REFERENCES teams(id), -- 转出方团队ID
    to_team_id            INTEGER REFERENCES teams(id), -- 接收方团队ID
    amount_grams          DECIMAL(15,3) NOT NULL,
    transfer_type         VARCHAR(30) NOT NULL DEFAULT 'personal', -- 转账类型
    status                VARCHAR(20) NOT NULL DEFAULT 'pending_approval', -- 扩展状态
    payment_time          TIMESTAMP, -- 实际支付时间
    notes                 TEXT, -- 转账备注
    rejection_reason      TEXT, -- 拒绝原因
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    -- 审批相关字段
    initiator_user_id     INTEGER REFERENCES users(id),
    approver_user_id      INTEGER REFERENCES users(id),
    approved_at           TIMESTAMP,
    approval_rejection_reason TEXT,
    -- 接收确认相关字段
    confirmed_by          INTEGER REFERENCES users(id),
    confirmed_at          TIMESTAMP,
    reception_rejection_reason TEXT,
    rejected_by           INTEGER REFERENCES users(id),
    rejected_at           TIMESTAMP,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_transfers_from_user ON tea_transfers(from_user_id);
CREATE INDEX idx_tea_transfers_to_user ON tea_transfers(to_user_id);
CREATE INDEX idx_tea_transfers_to_team ON tea_transfers(to_team_id);
CREATE INDEX idx_tea_transfers_from_team ON tea_transfers(from_team_id);
CREATE INDEX idx_tea_transfers_status ON tea_transfers(status);
CREATE INDEX idx_tea_transfers_transfer_type ON tea_transfers(transfer_type);
CREATE INDEX idx_tea_transfers_created_at ON tea_transfers(created_at);
CREATE INDEX idx_tea_transfers_expires_at ON tea_transfers(expires_at);

-- 添加表注释
COMMENT ON TABLE tea_transfers IS '茶叶转账记录表';
COMMENT ON COLUMN tea_transfers.amount_grams IS '转账茶叶数量，单位为克';
COMMENT ON COLUMN tea_transfers.transfer_type IS '转账类型: personal-个人转账, team_initiated-团队发起转账, team_approval_required-团队转账需审批';
COMMENT ON COLUMN tea_transfers.status IS '转账状态: pending_approval-待审批, pending_receipt-待接收, approved-已审批, approval_rejected-审批拒绝, completed-已完成, rejected-接收拒绝, expired-已过期';
COMMENT ON COLUMN tea_transfers.to_team_id IS '接收方团队ID（团队转账时使用）';
COMMENT ON COLUMN tea_transfers.from_team_id IS '转出方团队ID（团队转出时使用）';
COMMENT ON COLUMN tea_transfers.initiator_user_id IS '发起人ID（团队转账时使用）';
COMMENT ON COLUMN tea_transfers.approver_user_id IS '审批人ID（团队转账时使用）';

-- 茶叶交易流水表
CREATE TABLE tea_transactions (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    user_id               INTEGER NOT NULL REFERENCES users(id),
    transfer_id           UUID REFERENCES tea_transfers(uuid), -- 关联的转账ID
    transaction_type      VARCHAR(30) NOT NULL, -- transfer_out, transfer_in, system_grant, system_deduct, refund
    amount_grams          DECIMAL(15,3) NOT NULL,
    balance_before        DECIMAL(15,3) NOT NULL,
    balance_after         DECIMAL(15,3) NOT NULL,
    description           TEXT,
    related_user_id       INTEGER REFERENCES users(id), -- 交易相关用户（如转账对方）
    target_team_id        INTEGER REFERENCES teams(id), -- 交易相关团队
    target_type           VARCHAR(10) NOT NULL DEFAULT 'u', -- 目标类型: u-用户, t-团队
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_transactions_user_id ON tea_transactions(user_id);
CREATE INDEX idx_tea_transactions_type ON tea_transactions(transaction_type);
CREATE INDEX idx_tea_transactions_created_at ON tea_transactions(created_at);
CREATE INDEX idx_tea_transactions_transfer_id ON tea_transactions(transfer_id);
CREATE INDEX idx_tea_transactions_target_team ON tea_transactions(target_team_id);
CREATE INDEX idx_tea_transactions_target_type ON tea_transactions(target_type);

-- 添加表注释
COMMENT ON TABLE tea_transactions IS '茶叶交易流水记录表';
COMMENT ON COLUMN tea_transactions.target_team_id IS '交易相关团队ID';
COMMENT ON COLUMN tea_transactions.target_type IS '目标类型: u-用户, t-团队';

-- ============================================
-- 团队茶叶账户表（基于用户账户系统扩展）
-- ============================================

-- 团队茶叶账户表
CREATE TABLE tea.team_accounts (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    team_id               INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    balance_grams         DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 茶叶数量，精确到毫克
    locked_balance_grams  DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 锁定余额
    status                VARCHAR(20) NOT NULL DEFAULT 'normal', -- normal, frozen
    frozen_reason         TEXT, -- 冻结原因
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX idx_tea_team_accounts_team_id ON tea.team_accounts(team_id);
CREATE INDEX idx_tea_team_accounts_status ON tea.team_accounts(status);

-- 添加表注释
COMMENT ON TABLE tea.team_accounts IS '团队茶叶账户表';
COMMENT ON COLUMN tea.team_accounts.locked_balance_grams IS '团队被锁定的茶叶数量，单位为克';

-- 团队茶叶操作记录表（需要双重审批）
CREATE TABLE team_tea_operations (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    team_id               INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    operation_type        VARCHAR(30) NOT NULL, -- deposit, withdraw, transfer_out, transfer_in
    amount_grams          DECIMAL(15,3) NOT NULL,
    status                VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, expired
    operator_user_id      INTEGER NOT NULL REFERENCES users(id), -- 操作人
    approver_user_id      INTEGER REFERENCES users(id), -- 审批人
    target_team_id        INTEGER REFERENCES teams(id), -- 转账目标团队（transfer操作时使用）
    target_user_id        INTEGER REFERENCES users(id), -- 转账目标用户（转入用户账户时使用）
    notes                 TEXT, -- 操作备注
    rejection_reason      TEXT, -- 拒绝原因
    expires_at            TIMESTAMP NOT NULL, -- 审批过期时间
    approved_at           TIMESTAMP, -- 审批时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 团队茶叶交易流水表
CREATE TABLE tea.team_transactions (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    team_id               INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    operation_id          UUID REFERENCES team_tea_operations(uuid), -- 关联的操作ID
    transaction_type      VARCHAR(30) NOT NULL, -- deposit, withdraw, transfer_out, transfer_in, system_grant, system_deduct
    amount_grams          DECIMAL(15,3) NOT NULL,
    balance_before        DECIMAL(15,3) NOT NULL,
    balance_after         DECIMAL(15,3) NOT NULL,
    description           TEXT,
    related_team_id       INTEGER REFERENCES teams(id), -- 交易相关团队（如转账对方团队）
    related_user_id       INTEGER REFERENCES users(id), -- 交易相关用户（如操作人、审批人）
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 第二部分：系统配置和约束
-- ============================================

-- 账户状态枚举约束
ALTER TABLE tea_accounts ADD CONSTRAINT check_tea_account_status 
    CHECK (status IN ('normal', 'frozen'));

-- 转账状态枚举约束（扩展版本）
ALTER TABLE tea_transfers ADD CONSTRAINT check_tea_transfer_status 
    CHECK (status IN ('pending_approval', 'pending_receipt', 'approved', 'approval_rejected', 'completed', 'rejected', 'expired'));

-- 转账类型枚举约束
ALTER TABLE tea_transfers ADD CONSTRAINT check_tea_transfer_type 
    CHECK (transfer_type IN ('personal', 'team_initiated', 'team_approval_required'));

-- 交易类型枚举约束
ALTER TABLE tea_transactions ADD CONSTRAINT check_tea_transaction_type 
    CHECK (transaction_type IN ('transfer_out', 'transfer_in', 'system_grant', 'system_deduct', 'refund'));

-- 目标类型枚举约束
ALTER TABLE tea_transactions ADD CONSTRAINT check_tea_transaction_target_type 
    CHECK (target_type IN ('u', 't'));

-- 团队账户状态枚举约束
ALTER TABLE tea.team_accounts ADD CONSTRAINT check_tea_team_account_status 
    CHECK (status IN ('normal', 'frozen'));

-- 团队操作状态枚举约束
ALTER TABLE team_tea_operations ADD CONSTRAINT check_team_tea_operation_status 
    CHECK (status IN ('pending', 'approved', 'rejected', 'expired'));

-- 团队操作类型枚举约束
ALTER TABLE team_tea_operations ADD CONSTRAINT check_team_tea_operation_type 
    CHECK (operation_type IN ('deposit', 'withdraw', 'transfer_out', 'transfer_in'));

-- 团队交易类型枚举约束
ALTER TABLE tea.team_transactions ADD CONSTRAINT check_team_tea_transaction_type 
    CHECK (transaction_type IN ('deposit', 'withdraw', 'transfer_out', 'transfer_in', 'system_grant', 'system_deduct'));

-- 金额不能为负数约束
ALTER TABLE tea_accounts ADD CONSTRAINT check_tea_account_balance_positive 
    CHECK (balance_grams >= 0);

ALTER TABLE tea_transfers ADD CONSTRAINT check_tea_transfer_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE tea_transactions ADD CONSTRAINT check_tea_transaction_amount_positive 
    CHECK (amount_grams > 0);

-- 团队账户金额约束
ALTER TABLE tea.team_accounts ADD CONSTRAINT check_tea_team_account_balance_positive 
    CHECK (balance_grams >= 0);

ALTER TABLE team_tea_operations ADD CONSTRAINT check_team_tea_operation_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE tea.team_transactions ADD CONSTRAINT check_team_tea_transaction_amount_positive 
    CHECK (amount_grams > 0);

-- ============================================
-- 第三部分：触发器和视图
-- ============================================

-- 更新时间触发器函数
CREATE OR REPLACE FUNCTION update_tea_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION update_tea_transfers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 茶叶账户表更新时间触发器
CREATE TRIGGER tea_accounts_updated_at_trigger
    BEFORE UPDATE ON tea_accounts
    FOR EACH ROW EXECUTE FUNCTION update_tea_accounts_updated_at();

-- 茶叶转账表更新时间触发器
CREATE TRIGGER tea_transfers_updated_at_trigger
    BEFORE UPDATE ON tea_transfers
    FOR EACH ROW EXECUTE FUNCTION update_tea_transfers_updated_at();

-- 团队茶叶账户表更新时间触发器
CREATE TRIGGER tea_team_accounts_updated_at_trigger
    BEFORE UPDATE ON tea.team_accounts
    FOR EACH ROW EXECUTE FUNCTION update_tea_transfers_updated_at();

-- 团队茶叶操作表更新时间触发器
CREATE TRIGGER team_tea_operations_updated_at_trigger
    BEFORE UPDATE ON team_tea_operations
    FOR EACH ROW EXECUTE FUNCTION update_tea_transfers_updated_at();

-- 用户账户汇总视图
CREATE VIEW user_tea_account_summary AS
SELECT 
    u.id as user_id,
    u.uuid as user_uuid,
    u.name as user_name,
    u.email,
    COALESCE(ta.balance_grams, 0) as tea_balance,
    COALESCE(ta.locked_balance_grams, 0) as locked_balance,
    (COALESCE(ta.balance_grams, 0) - COALESCE(ta.locked_balance_grams, 0)) as available_balance,
    COALESCE(ta.status, 'no_account') as account_status,
    COALESCE(ta.frozen_reason, '') as frozen_reason,
    -- 收到转账数量
    (SELECT COUNT(*) FROM tea_transfers WHERE to_user_id = u.id AND status = 'pending_receipt') as pending_received_count,
    -- 发出转账数量
    (SELECT COUNT(*) FROM tea_transfers WHERE from_user_id = u.id AND status IN ('pending_approval', 'pending_receipt')) as pending_sent_count,
    -- 总交易次数
    (SELECT COUNT(*) FROM tea_transactions WHERE user_id = u.id) as total_transactions,
    -- 账户创建时间
    ta.created_at as account_created_at
FROM users u
LEFT JOIN tea_accounts ta ON u.id = ta.user_id;

COMMENT ON VIEW user_tea_account_summary IS '用户茶叶账户汇总信息视图';

-- 团队账户汇总视图
CREATE VIEW tea.team_account_summary AS
SELECT 
    t.id as team_id,
    t.uuid as team_uuid,
    t.name as team_name,
    t.abbreviation,
    COALESCE(tta.balance_grams, 0) as tea_balance,
    COALESCE(tta.locked_balance_grams, 0) as locked_balance,
    (COALESCE(tta.balance_grams, 0) - COALESCE(tta.locked_balance_grams, 0)) as available_balance,
    COALESCE(tta.status, 'no_account') as account_status,
    COALESCE(tta.frozen_reason, '') as frozen_reason,
    -- 待审批操作数量
    (SELECT COUNT(*) FROM team_tea_operations WHERE team_id = t.id AND status = 'pending') as pending_operations_count,
    -- 总操作次数
    (SELECT COUNT(*) FROM team_tea_operations WHERE team_id = t.id) as total_operations,
    -- 总交易次数
    (SELECT COUNT(*) FROM tea.team_transactions WHERE team_id = t.id) as total_transactions,
    -- 账户创建时间
    tta.created_at as account_created_at
FROM teams t
LEFT JOIN tea.team_accounts tta ON t.id = tta.team_id;

COMMENT ON VIEW tea.team_account_summary IS '团队茶叶账户汇总信息视图';

-- ============================================
-- 第四部分：数据修复和一致性检查
-- ============================================

-- 修复锁定余额数据不一致的脚本
CREATE OR REPLACE FUNCTION fix_locked_balance_data()
RETURNS TABLE(
    check_type TEXT,
    total_accounts BIGINT,
    negative_locked_balance BIGINT,
    negative_available_balance BIGINT
) AS $$
BEGIN
    -- 备份相关表（可选，在生产环境中谨慎使用）
    -- CREATE TABLE IF NOT EXISTS tea_accounts_backup AS SELECT * FROM tea_accounts;
    -- CREATE TABLE IF NOT EXISTS tea_team_accounts_backup AS SELECT * FROM tea.team_accounts;
    -- CREATE TABLE IF NOT EXISTS tea_transfers_backup AS SELECT * FROM tea_transfers;

    -- 计算每个用户的实际待确认转账金额
    WITH user_pending_transfers AS (
        SELECT 
            from_user_id,
            COALESCE(SUM(amount_grams), 0) as total_pending_amount
        FROM tea_transfers 
        WHERE status IN ('pending_approval', 'pending_receipt') AND expires_at > NOW()
        GROUP BY from_user_id
    ),
    -- 更新用户账户的锁定余额为实际的待确认金额
    updated_user_accounts AS (
        UPDATE tea_accounts 
        SET locked_balance_grams = COALESCE(upt.total_pending_amount, 0)
        FROM user_pending_transfers upt
        WHERE tea_accounts.user_id = upt.from_user_id
        RETURNING tea_accounts.*
    )
    -- 处理没有待确认转账的用户，将其锁定余额设为0
    UPDATE tea_accounts 
    SET locked_balance_grams = 0
    WHERE user_id NOT IN (
        SELECT DISTINCT from_user_id 
        FROM tea_transfers 
        WHERE status IN ('pending_approval', 'pending_receipt') AND expires_at > NOW()
    )
    AND locked_balance_grams != 0;

    -- 同样的逻辑处理团队账户
    WITH team_pending_operations AS (
        SELECT 
            team_id,
            COALESCE(SUM(amount_grams), 0) as total_pending_amount
        FROM team_tea_operations 
        WHERE status = 'pending' AND expires_at > NOW()
        GROUP BY team_id
    )
    UPDATE tea.team_accounts 
    SET locked_balance_grams = COALESCE(tpo.total_pending_amount, 0)
    FROM team_pending_operations tpo
    WHERE tea.team_accounts.team_id = tpo.team_id;

    -- 处理没有待确认操作的团队，将其锁定余额设为0
    UPDATE tea.team_accounts 
    SET locked_balance_grams = 0
    WHERE team_id NOT IN (
        SELECT DISTINCT team_id 
        FROM team_tea_operations 
        WHERE status = 'pending' AND expires_at > NOW()
    )
    AND locked_balance_grams != 0;

    -- 返回修复结果
    RETURN QUERY
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
    FROM tea.team_accounts;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 第五部分：调试工具和状态检查
-- ============================================

-- 检查锁定余额数据问题的函数
CREATE OR REPLACE FUNCTION debug_locked_balance()
RETURNS TABLE(
    debug_type TEXT,
    account_type TEXT,
    id INTEGER,
    user_id INTEGER,
    balance_grams DECIMAL(15,3),
    locked_balance_grams DECIMAL(15,3),
    available_balance DECIMAL(15,3),
    status TEXT
) AS $$
BEGIN
    -- 检查锁定余额为负数的记录
    RETURN QUERY
    SELECT 
        'negative_locked_balance' as debug_type,
        'user_accounts' as account_type,
        ta.id,
        ta.user_id,
        ta.balance_grams,
        ta.locked_balance_grams,
        (ta.balance_grams - ta.locked_balance_grams) as available_balance,
        ta.status
    FROM tea_accounts ta
    WHERE ta.locked_balance_grams < 0
    UNION ALL
    SELECT 
        'negative_locked_balance' as debug_type,
        'team_accounts' as account_type,
        tta.id,
        tta.team_id as user_id,
        tta.balance_grams,
        tta.locked_balance_grams,
        (tta.balance_grams - tta.locked_balance_grams) as available_balance,
        tta.status
    FROM tea.team_accounts tta
    WHERE tta.locked_balance_grams < 0;

    -- 检查可用余额为负数的记录
    RETURN QUERY
    SELECT 
        'negative_available_balance' as debug_type,
        'user_accounts' as account_type,
        ta.id,
        ta.user_id,
        ta.balance_grams,
        ta.locked_balance_grams,
        (ta.balance_grams - ta.locked_balance_grams) as available_balance,
        ta.status
    FROM tea_accounts ta
    WHERE (ta.balance_grams - ta.locked_balance_grams) < 0
    UNION ALL
    SELECT 
        'negative_available_balance' as debug_type,
        'team_accounts' as account_type,
        tta.id,
        tta.team_id as user_id,
        tta.balance_grams,
        tta.locked_balance_grams,
        (tta.balance_grams - tta.locked_balance_grams) as available_balance,
        tta.status
    FROM tea.team_accounts tta
    WHERE (tta.balance_grams - tta.locked_balance_grams) < 0;

    -- 检查待确认转账的总金额与锁定余额的关系
    RETURN QUERY
    SELECT 
        'locked_balance_mismatch' as debug_type,
        'user_accounts' as account_type,
        u.id,
        u.user_id,
        u.balance_grams,
        u.locked_balance_grams,
        (u.balance_grams - u.locked_balance_grams) as available_balance,
        u.status
    FROM tea_accounts u
    LEFT JOIN tea_transfers t ON u.user_id = t.from_user_id 
        AND t.status IN ('pending_approval', 'pending_receipt')
        AND t.expires_at > NOW()
    GROUP BY u.id, u.user_id, u.balance_grams, u.locked_balance_grams, u.status
    HAVING ABS(u.locked_balance_grams - COALESCE(SUM(t.amount_grams), 0)) > 0.001;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 第六部分：初始化数据和示例查询
-- ============================================

-- 为现有用户创建茶叶账户（如果还没有的话）
INSERT INTO tea_accounts (user_id, balance_grams, locked_balance_grams, status)
SELECT id, 0.000, 0.000, 'normal'
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM tea_accounts ta 
    WHERE ta.user_id = u.id
);

-- 为现有团队创建茶叶账户（如果还没有的话）
INSERT INTO tea.team_accounts (team_id, balance_grams, locked_balance_grams, status)
SELECT id, 0.000, 0.000, 'normal'
FROM teams t
WHERE NOT EXISTS (
    SELECT 1 FROM tea.team_accounts tta 
    WHERE tta.team_id = t.id
);

-- ============================================
-- 使用说明和示例查询
-- ============================================

/*
-- 查看用户茶叶账户信息
SELECT * FROM tea_accounts WHERE user_id = ?;

-- 查看用户转账历史
SELECT * FROM tea_transfers 
WHERE from_user_id = ? OR to_user_id = ?
ORDER BY created_at DESC;

-- 查看用户交易流水
SELECT * FROM tea_transactions 
WHERE user_id = ?
ORDER BY created_at DESC;

-- 查看待确认转账
SELECT * FROM tea_transfers 
WHERE status IN ('pending_approval', 'pending_receipt') AND expires_at > CURRENT_TIMESTAMP;

-- 查看账户汇总信息
SELECT * FROM user_tea_account_summary WHERE user_id = ?;

-- 运行数据修复
SELECT * FROM fix_locked_balance_data();

-- 调试锁定余额问题
SELECT * FROM debug_locked_balance();

-- 检查系统状态
SELECT 
    (SELECT COUNT(*) FROM tea_accounts) as total_user_accounts,
    (SELECT COUNT(*) FROM tea.team_accounts) as total_team_accounts,
    (SELECT COUNT(*) FROM tea_transfers WHERE status IN ('pending_approval', 'pending_receipt')) as pending_transfers,
    (SELECT COUNT(*) FROM team_tea_operations WHERE status = 'pending') as pending_team_operations;
*/
