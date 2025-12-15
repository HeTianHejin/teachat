-- 茶叶支付系统数据库架构定义
-- 为TeaChat添加茶叶账户和转账功能

-- ============================================
-- 茶叶支付系统表
-- ============================================

-- 茶叶账户表
CREATE TABLE tea_accounts (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    user_id               INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance_grams         DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 茶叶数量，精确到毫克
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
COMMENT ON COLUMN tea_accounts.status IS '账户状态: normal-正常, frozen-冻结';
COMMENT ON COLUMN tea_accounts.frozen_reason IS '账户冻结原因说明';

-- 茶叶转账记录表
CREATE TABLE tea_transfers (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    from_user_id          INTEGER NOT NULL REFERENCES users(id),
    to_user_id            INTEGER NOT NULL REFERENCES users(id),
    amount_grams          DECIMAL(15,3) NOT NULL,
    status                VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, confirmed, rejected, expired
    payment_time          TIMESTAMP, -- 实际支付时间
    notes                 TEXT, -- 转账备注
    rejection_reason      TEXT, -- 拒绝原因
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_transfers_from_user ON tea_transfers(from_user_id);
CREATE INDEX idx_tea_transfers_to_user ON tea_transfers(to_user_id);
CREATE INDEX idx_tea_transfers_status ON tea_transfers(status);
CREATE INDEX idx_tea_transfers_created_at ON tea_transfers(created_at);
CREATE INDEX idx_tea_transfers_expires_at ON tea_transfers(expires_at);

-- 添加表注释
COMMENT ON TABLE tea_transfers IS '茶叶转账记录表';
COMMENT ON COLUMN tea_transfers.amount_grams IS '转账茶叶数量，单位为克';
COMMENT ON COLUMN tea_transfers.status IS '转账状态: pending-待确认, confirmed-已确认, rejected-已拒绝, expired-已过期';
COMMENT ON COLUMN tea_transfers.payment_time IS '实际转账完成时间';
COMMENT ON COLUMN tea_transfers.notes IS '转账备注信息';
COMMENT ON COLUMN tea_transfers.rejection_reason IS '拒绝转账的原因';
COMMENT ON COLUMN tea_transfers.expires_at IS '转账过期时间，超过此时间自动失效';

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
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_transactions_user_id ON tea_transactions(user_id);
CREATE INDEX idx_tea_transactions_type ON tea_transactions(transaction_type);
CREATE INDEX idx_tea_transactions_created_at ON tea_transactions(created_at);
CREATE INDEX idx_tea_transactions_transfer_id ON tea_transactions(transfer_id);

-- 添加表注释
COMMENT ON TABLE tea_transactions IS '茶叶交易流水记录表';
COMMENT ON COLUMN tea_transactions.transaction_type IS '交易类型: transfer_out-转出, transfer_in-转入, system_grant-系统发放, system_deduct-系统扣除, refund-退款';
COMMENT ON COLUMN tea_transactions.balance_before IS '交易前余额';
COMMENT ON COLUMN tea_transactions.balance_after IS '交易后余额';
COMMENT ON COLUMN tea_transactions.description IS '交易描述';
COMMENT ON COLUMN tea_transactions.related_user_id IS '交易相关用户ID（如转账的对方用户）';

-- ============================================
-- 团队茶叶账户表（基于用户账户系统扩展）
-- ============================================

-- 团队茶叶账户表
CREATE TABLE team_tea_accounts (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    team_id               INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    balance_grams         DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 茶叶数量，精确到毫克
    status                VARCHAR(20) NOT NULL DEFAULT 'normal', -- normal, frozen
    frozen_reason         TEXT, -- 冻结原因
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX idx_team_tea_accounts_team_id ON team_tea_accounts(team_id);
CREATE INDEX idx_team_tea_accounts_status ON team_tea_accounts(status);

-- 添加表注释
COMMENT ON TABLE team_tea_accounts IS '团队茶叶账户表';
COMMENT ON COLUMN team_tea_accounts.balance_grams IS '团队茶叶余额，单位为克，精确到3位小数(毫克)';
COMMENT ON COLUMN team_tea_accounts.status IS '团队账户状态: normal-正常, frozen-冻结';
COMMENT ON COLUMN team_tea_accounts.frozen_reason IS '团队账户冻结原因说明';

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

-- 创建索引
CREATE INDEX idx_team_tea_operations_team_id ON team_tea_operations(team_id);
CREATE INDEX idx_team_tea_operations_status ON team_tea_operations(status);
CREATE INDEX idx_team_tea_operations_operator ON team_tea_operations(operator_user_id);
CREATE INDEX idx_team_tea_operations_approver ON team_tea_operations(approver_user_id);
CREATE INDEX idx_team_tea_operations_created_at ON team_tea_operations(created_at);
CREATE INDEX idx_team_tea_operations_expires_at ON team_tea_operations(expires_at);

-- 添加表注释
COMMENT ON TABLE team_tea_operations IS '团队茶叶操作记录表';
COMMENT ON COLUMN team_tea_operations.operation_type IS '操作类型: deposit-存入, withdraw-提取, transfer_out-转出, transfer_in-转入';
COMMENT ON COLUMN team_tea_operations.status IS '操作状态: pending-待审批, approved-已审批, rejected-已拒绝, expired-已过期';
COMMENT ON COLUMN team_tea_operations.operator_user_id IS '操作发起人ID';
COMMENT ON COLUMN team_tea_operations.approver_user_id IS '审批人ID';
COMMENT ON COLUMN team_tea_operations.target_team_id IS '转账目标团队ID';
COMMENT ON COLUMN team_tea_operations.target_user_id IS '转账目标用户ID';
COMMENT ON COLUMN team_tea_operations.notes IS '操作备注信息';
COMMENT ON COLUMN team_tea_operations.rejection_reason IS '拒绝操作的原因';
COMMENT ON COLUMN team_tea_operations.expires_at IS '审批过期时间，超过此时间自动失效';
COMMENT ON COLUMN team_tea_operations.approved_at IS '实际审批完成时间';

-- 团队茶叶交易流水表
CREATE TABLE team_tea_transactions (
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

-- 创建索引
CREATE INDEX idx_team_tea_transactions_team_id ON team_tea_transactions(team_id);
CREATE INDEX idx_team_tea_transactions_type ON team_tea_transactions(transaction_type);
CREATE INDEX idx_team_tea_transactions_created_at ON team_tea_transactions(created_at);
CREATE INDEX idx_team_tea_transactions_operation_id ON team_tea_transactions(operation_id);

-- 添加表注释
COMMENT ON TABLE team_tea_transactions IS '团队茶叶交易流水记录表';
COMMENT ON COLUMN team_tea_transactions.transaction_type IS '交易类型: deposit-存入, withdraw-提取, transfer_out-转出, transfer_in-转入, system_grant-系统发放, system_deduct-系统扣除';
COMMENT ON COLUMN team_tea_transactions.balance_before IS '交易前余额';
COMMENT ON COLUMN team_tea_transactions.balance_after IS '交易后余额';
COMMENT ON COLUMN team_tea_transactions.description IS '交易描述';
COMMENT ON COLUMN team_tea_transactions.related_team_id IS '交易相关团队ID（如转账的对方团队）';
COMMENT ON COLUMN team_tea_transactions.related_user_id IS '交易相关用户ID（如操作人、审批人）';

-- ============================================
-- 系统配置和约束
-- ============================================

-- 账户状态枚举约束
ALTER TABLE tea_accounts ADD CONSTRAINT check_tea_account_status 
    CHECK (status IN ('normal', 'frozen'));

-- 转账状态枚举约束
ALTER TABLE tea_transfers ADD CONSTRAINT check_tea_transfer_status 
    CHECK (status IN ('pending', 'confirmed', 'rejected', 'expired'));

-- 交易类型枚举约束
ALTER TABLE tea_transactions ADD CONSTRAINT check_tea_transaction_type 
    CHECK (transaction_type IN ('transfer_out', 'transfer_in', 'system_grant', 'system_deduct', 'refund'));

-- 团队账户状态枚举约束
ALTER TABLE team_tea_accounts ADD CONSTRAINT check_team_tea_account_status 
    CHECK (status IN ('normal', 'frozen'));

-- 团队操作状态枚举约束
ALTER TABLE team_tea_operations ADD CONSTRAINT check_team_tea_operation_status 
    CHECK (status IN ('pending', 'approved', 'rejected', 'expired'));

-- 团队操作类型枚举约束
ALTER TABLE team_tea_operations ADD CONSTRAINT check_team_tea_operation_type 
    CHECK (operation_type IN ('deposit', 'withdraw', 'transfer_out', 'transfer_in'));

-- 团队交易类型枚举约束
ALTER TABLE team_tea_transactions ADD CONSTRAINT check_team_tea_transaction_type 
    CHECK (transaction_type IN ('deposit', 'withdraw', 'transfer_out', 'transfer_in', 'system_grant', 'system_deduct'));

-- 金额不能为负数约束
ALTER TABLE tea_accounts ADD CONSTRAINT check_tea_account_balance_positive 
    CHECK (balance_grams >= 0);

ALTER TABLE tea_transfers ADD CONSTRAINT check_tea_transfer_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE tea_transactions ADD CONSTRAINT check_tea_transaction_amount_positive 
    CHECK (amount_grams > 0);

-- 团队账户金额约束
ALTER TABLE team_tea_accounts ADD CONSTRAINT check_team_tea_account_balance_positive 
    CHECK (balance_grams >= 0);

ALTER TABLE team_tea_operations ADD CONSTRAINT check_team_tea_operation_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE team_tea_transactions ADD CONSTRAINT check_team_tea_transaction_amount_positive 
    CHECK (amount_grams > 0);

-- ============================================
-- 触发器：自动更新updated_at字段
-- ============================================

-- 茶叶账户表更新时间触发器
CREATE OR REPLACE FUNCTION update_tea_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER tea_accounts_updated_at_trigger
    BEFORE UPDATE ON tea_accounts
    FOR EACH ROW EXECUTE FUNCTION update_tea_accounts_updated_at();

-- 茶叶转账表更新时间触发器
CREATE OR REPLACE FUNCTION update_tea_transfers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER tea_transfers_updated_at_trigger
    BEFORE UPDATE ON tea_transfers
    FOR EACH ROW EXECUTE FUNCTION update_tea_transfers_updated_at();

-- 团队茶叶账户表更新时间触发器
CREATE TRIGGER team_tea_accounts_updated_at_trigger
    BEFORE UPDATE ON team_tea_accounts
    FOR EACH ROW EXECUTE FUNCTION update_tea_transfers_updated_at();

-- 团队茶叶操作表更新时间触发器
CREATE TRIGGER team_tea_operations_updated_at_trigger
    BEFORE UPDATE ON team_tea_operations
    FOR EACH ROW EXECUTE FUNCTION update_tea_transfers_updated_at();

-- ============================================
-- 视图：账户汇总信息
-- ============================================

-- 用户账户汇总视图
CREATE VIEW user_tea_account_summary AS
SELECT 
    u.id as user_id,
    u.uuid as user_uuid,
    u.name as user_name,
    u.email,
    COALESCE(ta.balance_grams, 0) as tea_balance,
    COALESCE(ta.status, 'no_account') as account_status,
    COALESCE(ta.frozen_reason, '') as frozen_reason,
    -- 收到转账数量
    (SELECT COUNT(*) FROM tea_transfers WHERE to_user_id = u.id AND status = 'pending') as pending_received_count,
    -- 发出转账数量
    (SELECT COUNT(*) FROM tea_transfers WHERE from_user_id = u.id AND status = 'pending') as pending_sent_count,
    -- 总交易次数
    (SELECT COUNT(*) FROM tea_transactions WHERE user_id = u.id) as total_transactions,
    -- 账户创建时间
    ta.created_at as account_created_at
FROM users u
LEFT JOIN tea_accounts ta ON u.id = ta.user_id;

COMMENT ON VIEW user_tea_account_summary IS '用户茶叶账户汇总信息视图';

-- 团队账户汇总视图
CREATE VIEW team_tea_account_summary AS
SELECT 
    t.id as team_id,
    t.uuid as team_uuid,
    t.name as team_name,
    t.abbreviation,
    COALESCE(tta.balance_grams, 0) as tea_balance,
    COALESCE(tta.status, 'no_account') as account_status,
    COALESCE(tta.frozen_reason, '') as frozen_reason,
    -- 待审批操作数量
    (SELECT COUNT(*) FROM team_tea_operations WHERE team_id = t.id AND status = 'pending') as pending_operations_count,
    -- 总操作次数
    (SELECT COUNT(*) FROM team_tea_operations WHERE team_id = t.id) as total_operations,
    -- 总交易次数
    (SELECT COUNT(*) FROM team_tea_transactions WHERE team_id = t.id) as total_transactions,
    -- 账户创建时间
    tta.created_at as account_created_at
FROM teams t
LEFT JOIN team_tea_accounts tta ON t.id = tta.team_id;

COMMENT ON VIEW team_tea_account_summary IS '团队茶叶账户汇总信息视图';

-- ============================================
-- 初始化数据：为现有用户创建茶叶账户
-- ============================================

-- 为现有用户创建茶叶账户（如果还没有的话）
INSERT INTO tea_accounts (user_id, balance_grams, status)
SELECT id, 0.000, 'normal'
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM tea_accounts ta 
    WHERE ta.user_id = u.id
);

-- 为现有团队创建茶叶账户（如果还没有的话）
INSERT INTO team_tea_accounts (team_id, balance_grams, status)
SELECT id, 0.000, 'normal'
FROM teams t
WHERE NOT EXISTS (
    SELECT 1 FROM team_tea_accounts tta 
    WHERE tta.team_id = t.id
);

-- ============================================
-- 示例查询
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
WHERE status = 'pending' AND expires_at > CURRENT_TIMESTAMP;

-- 查看账户汇总信息
SELECT * FROM user_tea_account_summary WHERE user_id = ?;
*/