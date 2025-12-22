-- ============================================
-- 茶叶账户系统重建脚本
-- 仅重建tea schema，保留public schema中的其他表数据
-- ============================================

-- 第一步：删除现有的tea schema及其所有对象
DROP SCHEMA IF EXISTS tea CASCADE;

-- 第二步：重新创建tea schema和所有表结构
-- 创建 tea schema
CREATE SCHEMA IF NOT EXISTS tea;

-- 用户茶叶账户表
CREATE TABLE tea.user_accounts (
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
CREATE UNIQUE INDEX idx_tea_user_accounts_user_id ON tea.user_accounts(user_id);
CREATE INDEX idx_tea_user_accounts_status ON tea.user_accounts(status);

-- 用户转出记录表
CREATE TABLE tea.user_transfer_out (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    from_user_id          INTEGER NOT NULL REFERENCES users(id),
    to_user_id            INTEGER REFERENCES users(id), -- 用户接收（与to_team_id二选一）
    to_team_id            INTEGER REFERENCES teams(id), -- 团队接收（与to_user_id二选一）
    amount_grams          DECIMAL(15,3) NOT NULL,
    status                VARCHAR(20) NOT NULL DEFAULT 'pending_receipt', -- 转账状态
    notes                 TEXT, -- 转账备注
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    payment_time          TIMESTAMP, -- 实际支付时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_user_transfer_out_from_user ON tea.user_transfer_out(from_user_id);
CREATE INDEX idx_tea_user_transfer_out_to_user ON tea.user_transfer_out(to_user_id);
CREATE INDEX idx_tea_user_transfer_out_to_team ON tea.user_transfer_out(to_team_id);
CREATE INDEX idx_tea_user_transfer_out_status ON tea.user_transfer_out(status);
CREATE INDEX idx_tea_user_transfer_out_expires_at ON tea.user_transfer_out(expires_at);

-- 转账转入记录表（接收方视角）
CREATE TABLE tea.transfer_in (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    user_id               INTEGER NOT NULL REFERENCES users(id),
    user_transfer_out_id  INTEGER REFERENCES tea.user_transfer_out(id), -- 用户转出记录ID
    team_transfer_out_id  INTEGER, -- 团队转出记录ID（预留）
    status                VARCHAR(20) NOT NULL, -- 转入状态
    confirmed_by          INTEGER REFERENCES users(id), -- 确认人
    rejected_by           INTEGER REFERENCES users(id), -- 拒绝人
    reception_rejection_reason TEXT, -- 拒收原因
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_transfer_in_user_id ON tea.transfer_in(user_id);
CREATE INDEX idx_tea_transfer_in_user_transfer_out ON tea.transfer_in(user_transfer_out_id);
CREATE INDEX idx_tea_transfer_in_status ON tea.transfer_in(status);

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

-- 团队茶叶转出记录表
CREATE TABLE tea.team_transfer_out (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    from_team_id          INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE, -- 转出方团队
    initiator_user_id     INTEGER NOT NULL REFERENCES users(id), -- 发起人（必须是团队成员）
    to_user_id            INTEGER REFERENCES users(id), -- 用户接收（与to_team_id二选一）
    to_team_id            INTEGER REFERENCES teams(id), -- 团队接收（与to_user_id二选一）
    amount_grams          DECIMAL(15,3) NOT NULL,
    notes                 TEXT, -- 转账备注
    status                VARCHAR(20) NOT NULL DEFAULT 'pending_approval', -- 转账状态
    transfer_type         VARCHAR(30) NOT NULL, -- team_initiated, team_approval_required
    approver_user_id      INTEGER REFERENCES users(id), -- 审批人ID
    approved_at           TIMESTAMP, -- 审批时间
    approval_rejection_reason TEXT, -- 审批拒绝原因
    rejected_by           INTEGER REFERENCES users(id), -- 拒绝人ID
    rejected_at           TIMESTAMP, -- 拒绝时间
    payment_time          TIMESTAMP, -- 实际支付时间（接收方确认后）
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_team_transfer_out_from_team ON tea.team_transfer_out(from_team_id);
CREATE INDEX idx_tea_team_transfer_out_to_user ON tea.team_transfer_out(to_user_id);
CREATE INDEX idx_tea_team_transfer_out_to_team ON tea.team_transfer_out(to_team_id);
CREATE INDEX idx_tea_team_transfer_out_status ON tea.team_transfer_out(status);
CREATE INDEX idx_tea_team_transfer_out_expires_at ON tea.team_transfer_out(expires_at);
CREATE INDEX idx_tea_team_transfer_out_transfer_type ON tea.team_transfer_out(transfer_type);

-- 团队茶叶操作记录表
CREATE TABLE tea.team_operations (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    team_id               INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    operation_type        VARCHAR(30) NOT NULL, -- deposit, withdraw, transfer_in
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
    operation_id          UUID REFERENCES tea.team_operations(uuid), -- 关联的操作ID
    transaction_type      VARCHAR(30) NOT NULL, -- deposit, withdraw, transfer_out, transfer_in, system_grant, system_deduct
    amount_grams          DECIMAL(15,3) NOT NULL,
    balance_before        DECIMAL(15,3) NOT NULL,
    balance_after         DECIMAL(15,3) NOT NULL,
    description           TEXT,
    target_team_id        INTEGER REFERENCES teams(id), -- 交易目标团队（如转账对方团队）
    target_user_id        INTEGER REFERENCES users(id), -- 交易目标用户（如转账对方用户）
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 第三步：添加约束
-- 账户状态枚举约束
ALTER TABLE tea.user_accounts ADD CONSTRAINT check_tea_user_account_status 
    CHECK (status IN ('normal', 'frozen'));

ALTER TABLE tea.user_transfer_out ADD CONSTRAINT check_tea_user_transfer_out_status 
    CHECK (status IN ('pending_receipt', 'completed', 'rejected', 'expired'));

ALTER TABLE tea.transfer_in ADD CONSTRAINT check_tea_transfer_in_status 
    CHECK (status IN ('completed', 'rejected'));

ALTER TABLE tea.user_transfer_out ADD CONSTRAINT check_tea_user_transfer_out_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE tea.team_accounts ADD CONSTRAINT check_tea_team_account_status 
    CHECK (status IN ('normal', 'frozen'));

ALTER TABLE tea.team_transfer_out ADD CONSTRAINT check_tea_team_transfer_out_status 
    CHECK (status IN ('pending_approval', 'approved', 'approval_rejected', 'pending_receipt', 'completed', 'rejected', 'expired'));

ALTER TABLE tea.team_transfer_out ADD CONSTRAINT check_tea_team_transfer_out_type 
    CHECK (transfer_type IN ('team_initiated', 'team_approval_required'));

ALTER TABLE tea.team_operations ADD CONSTRAINT check_tea_team_operation_status 
    CHECK (status IN ('pending', 'approved', 'rejected', 'expired'));

ALTER TABLE tea.team_operations ADD CONSTRAINT check_tea_team_operation_type 
    CHECK (operation_type IN ('deposit', 'withdraw', 'transfer_out', 'transfer_in'));

ALTER TABLE tea.team_transactions ADD CONSTRAINT check_tea_team_transaction_type 
    CHECK (transaction_type IN ('deposit', 'withdraw', 'transfer_out', 'transfer_in', 'system_grant', 'system_deduct'));

ALTER TABLE tea.user_accounts ADD CONSTRAINT check_tea_user_account_balance_positive 
    CHECK (balance_grams >= 0);

ALTER TABLE tea.team_accounts ADD CONSTRAINT check_tea_team_account_balance_positive 
    CHECK (balance_grams >= 0);

ALTER TABLE tea.team_operations ADD CONSTRAINT check_tea_team_operation_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE tea.team_transactions ADD CONSTRAINT check_tea_team_transaction_amount_positive 
    CHECK (amount_grams > 0);

-- 第四步：创建触发器和视图
-- 更新时间触发器函数
CREATE OR REPLACE FUNCTION update_tea_user_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION update_tea_user_transfer_out_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 茶叶账户表更新时间触发器
CREATE TRIGGER tea_user_accounts_updated_at_trigger
    BEFORE UPDATE ON tea.user_accounts
    FOR EACH ROW EXECUTE FUNCTION update_tea_user_accounts_updated_at();

-- 团队茶叶账户表更新时间触发器
CREATE TRIGGER tea_team_accounts_updated_at_trigger
    BEFORE UPDATE ON tea.team_accounts
    FOR EACH ROW EXECUTE FUNCTION update_tea_user_accounts_updated_at();

-- 团队茶叶操作表更新时间触发器
CREATE TRIGGER tea_team_operations_updated_at_trigger
    BEFORE UPDATE ON tea.team_operations
    FOR EACH ROW EXECUTE FUNCTION update_tea_user_accounts_updated_at();

-- 用户转出记录表更新时间触发器
CREATE TRIGGER tea_user_transfer_out_updated_at_trigger
    BEFORE UPDATE ON tea.user_transfer_out
    FOR EACH ROW EXECUTE FUNCTION update_tea_user_transfer_out_updated_at();

-- 用户账户汇总视图
CREATE VIEW tea.user_account_summary AS
SELECT 
    u.id as user_id,
    u.uuid as user_uuid,
    u.name as user_name,
    u.email,
    COALESCE(tua.balance_grams, 0) as tea_balance,
    COALESCE(tua.locked_balance_grams, 0) as locked_balance,
    (COALESCE(tua.balance_grams, 0) - COALESCE(tua.locked_balance_grams, 0)) as available_balance,
    COALESCE(tua.status, 'no_account') as account_status,
    COALESCE(tua.frozen_reason, '') as frozen_reason,
    -- 收到转账数量（用户间转账 + 团队转账）
    (SELECT COUNT(*) FROM tea.user_transfer_out 
     WHERE (to_user_id = u.id OR (to_team_id IS NOT NULL AND EXISTS (
         SELECT 1 FROM team_members WHERE team_id = to_team_id AND user_id = u.id AND status = 1
     ))) 
     AND status = 'pending_receipt' AND expires_at > NOW()) as pending_received_count,
    -- 发出转账数量
    (SELECT COUNT(*) FROM tea.user_transfer_out 
     WHERE from_user_id = u.id AND status = 'pending_receipt' AND expires_at > NOW()) as pending_sent_count,
    -- 总交易次数（从转出表和转入表中计算）
    (SELECT COUNT(*) FROM tea.user_transfer_out 
     WHERE (from_user_id = u.id OR to_user_id = u.id) AND status = 'completed') as total_transactions,
    -- 账户创建时间
    tua.created_at as account_created_at
FROM users u
LEFT JOIN tea.user_accounts tua ON u.id = tua.user_id;

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
    (SELECT COUNT(*) FROM tea.team_operations WHERE team_id = t.id AND status = 'pending') as pending_operations_count,
    -- 总操作次数
    (SELECT COUNT(*) FROM tea.team_operations WHERE team_id = t.id) as total_operations,
    -- 总交易次数
    (SELECT COUNT(*) FROM tea.team_transactions WHERE team_id = t.id) as total_transactions,
    -- 账户创建时间
    tta.created_at as account_created_at
FROM teams t
LEFT JOIN tea.team_accounts tta ON t.id = tta.team_id;

-- 第五步：初始化数据
-- 为现有用户创建茶叶账户（如果还没有的话）
INSERT INTO tea.user_accounts (user_id, balance_grams, locked_balance_grams, status)
SELECT id, 0.000, 0.000, 'normal'
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM tea.user_accounts tua 
    WHERE tua.user_id = u.id
);

-- 为现有团队创建茶叶账户（如果还没有的话）
INSERT INTO tea.team_accounts (team_id, balance_grams, locked_balance_grams, status)
SELECT id, 0.000, 0.000, 'normal'
FROM teams t
WHERE NOT EXISTS (
    SELECT 1 FROM tea.team_accounts tta 
    WHERE tta.team_id = t.id
);

-- 第六步：验证重建结果
SELECT '重建完成' as status;
SELECT COUNT(*) as user_accounts_count FROM tea.user_accounts;
SELECT COUNT(*) as team_accounts_count FROM tea.team_accounts;