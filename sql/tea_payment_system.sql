-- ============================================
-- 全新茶叶支付系统数据库架构定义（基于重构的结构体）
-- 为TeaChat添加茶叶账户和转账功能
-- 所有表统一在tea schema中，与Go结构体完全匹配
-- 
-- 删除脚本（重建前执行）：
-- DROP SCHEMA IF EXISTS tea CASCADE;
-- ============================================

-- 删除旧的tea schema及其所有对象
DROP SCHEMA IF EXISTS tea CASCADE;

-- 创建全新的tea schema
CREATE SCHEMA tea;

-- ============================================
-- 第一部分：用户相关表结构
-- ============================================

-- 用户茶叶账户表（完全匹配TeaUserAccount结构体）
CREATE TABLE tea.user_accounts (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    user_id               INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance_grams         DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 茶叶数量(克)
    locked_balance_grams  DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 交易有效期被锁定的茶叶数量(克)
    status                VARCHAR(20) NOT NULL DEFAULT 'normal', -- normal, frozen
    frozen_reason         TEXT, -- 冻结原因
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX idx_tea_user_accounts_user_id ON tea.user_accounts(user_id);
CREATE INDEX idx_tea_user_accounts_status ON tea.user_accounts(status);

-- 添加表注释
COMMENT ON TABLE tea.user_accounts IS '用户茶叶账户表（匹配TeaUserAccount结构体）';
COMMENT ON COLUMN tea.user_accounts.balance_grams IS '茶叶数量(克)，精确到3位小数(毫克)';
COMMENT ON COLUMN tea.user_accounts.locked_balance_grams IS '交易有效期被锁定的茶叶数量(克)';
COMMENT ON COLUMN tea.user_accounts.status IS '账户状态: normal-正常, frozen-冻结';

-- 用户对用户转账记录表（完全匹配TeaUserToUserTransferOut结构体）
CREATE TABLE tea.user_to_user_transfer_out (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    from_user_id          INTEGER NOT NULL REFERENCES users(id),
    from_user_name        VARCHAR(255) NOT NULL, -- 转出方用户名称
    to_user_id            INTEGER NOT NULL REFERENCES users(id), -- 接收方用户ID
    to_user_name          VARCHAR(255) NOT NULL, -- 接收方用户名称
    amount_grams          DECIMAL(15,3) NOT NULL, -- 转账额度（克）
    notes                 TEXT NOT NULL DEFAULT '-', -- 转账备注
    status                VARCHAR(20) NOT NULL DEFAULT 'pending_receipt', -- 转账状态
    balance_after_transfer DECIMAL(15,3), -- 转出后账户余额（克）
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    payment_time          TIMESTAMP, -- 实际支付时间
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_user_to_user_from_user ON tea.user_to_user_transfer_out(from_user_id);
CREATE INDEX idx_tea_user_to_user_to_user ON tea.user_to_user_transfer_out(to_user_id);
CREATE INDEX idx_tea_user_to_user_status ON tea.user_to_user_transfer_out(status);
CREATE INDEX idx_tea_user_to_user_expires_at ON tea.user_to_user_transfer_out(expires_at);

-- 添加表注释
COMMENT ON TABLE tea.user_to_user_transfer_out IS '用户对用户转账记录表（匹配TeaUserToUserTransferOut结构体）';
COMMENT ON COLUMN tea.user_to_user_transfer_out.status IS '转账状态: pending_receipt-待接收, completed-已完成, rejected-接收拒绝, expired-已过期';

-- 用户对团队转账记录表（完全匹配TeaUserToTeamTransferOut结构体）
CREATE TABLE tea.user_to_team_transfer_out (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    from_user_id          INTEGER NOT NULL REFERENCES users(id),
    from_user_name        VARCHAR(255) NOT NULL, -- 转出方用户名称
    to_team_id            INTEGER NOT NULL REFERENCES teams(id), -- 接收方团队ID
    to_team_name          VARCHAR(255) NOT NULL, -- 接收方团队名称
    amount_grams          DECIMAL(15,3) NOT NULL, -- 转账额度（克）
    notes                 TEXT NOT NULL DEFAULT '-', -- 转账备注
    status                VARCHAR(20) NOT NULL DEFAULT 'pending_receipt', -- 转账状态
    balance_after_transfer DECIMAL(15,3), -- 转出后账户余额（克）
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    payment_time          TIMESTAMP, -- 实际支付时间
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_user_to_team_from_user ON tea.user_to_team_transfer_out(from_user_id);
CREATE INDEX idx_tea_user_to_team_to_team ON tea.user_to_team_transfer_out(to_team_id);
CREATE INDEX idx_tea_user_to_team_status ON tea.user_to_team_transfer_out(status);
CREATE INDEX idx_tea_user_to_team_expires_at ON tea.user_to_team_transfer_out(expires_at);

-- 添加表注释
COMMENT ON TABLE tea.user_to_team_transfer_out IS '用户对团队转账记录表（匹配TeaUserToTeamTransferOut结构体）';
COMMENT ON COLUMN tea.user_to_team_transfer_out.status IS '转账状态: pending_receipt-待接收, completed-已完成, rejected-接收拒绝, expired-已过期';

-- 用户对用户转账接收记录表（完全匹配TeaUserFromUserTransferIn结构体）
CREATE TABLE tea.user_from_user_transfer_in (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    user_to_user_transfer_out_id INTEGER NOT NULL REFERENCES tea.user_to_user_transfer_out(id),
    to_user_id            INTEGER NOT NULL REFERENCES users(id), -- 接收用户id
    to_user_name          VARCHAR(255) NOT NULL, -- 接收用户名称
    from_user_id          INTEGER NOT NULL REFERENCES users(id), -- 转出用户id
    from_user_name        VARCHAR(255) NOT NULL, -- 转出用户名称
    amount_grams          DECIMAL(15,3) NOT NULL, -- 接收转账额度（克）
    notes                 TEXT NOT NULL DEFAULT '-', -- 转出方备注（从转出表复制过来）
    balance_after_receipt DECIMAL(15,3), -- 接收后账户余额
    status                VARCHAR(20) NOT NULL, -- 转入状态
    is_confirmed          BOOLEAN NOT NULL DEFAULT FALSE, -- 是否确认接收
    operational_user_id   INTEGER NOT NULL REFERENCES users(id), -- 操作用户id
    reception_rejection_reason TEXT NOT NULL DEFAULT '-', -- 拒收原因
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_user_from_user_to_user ON tea.user_from_user_transfer_in(to_user_id);
CREATE INDEX idx_tea_user_from_user_status ON tea.user_from_user_transfer_in(status);
CREATE INDEX idx_tea_user_from_user_transfer_out ON tea.user_from_user_transfer_in(user_to_user_transfer_out_id);

-- 添加表注释
COMMENT ON TABLE tea.user_from_user_transfer_in IS '用户对用户转账接收记录表（匹配TeaUserFromUserTransferIn结构体）';
COMMENT ON COLUMN tea.user_from_user_transfer_in.status IS '转入状态: completed-已完成, rejected-接收拒绝';

-- 用户对团队转账接收记录表（完全匹配TeaUserFromTeamTransferIn结构体）
CREATE TABLE tea.user_from_team_transfer_in (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    team_to_user_transfer_out_id INTEGER NOT NULL, -- 引用-团队对用户转出记录id
    to_user_id            INTEGER NOT NULL REFERENCES users(id), -- 接收用户id
    to_user_name          VARCHAR(255) NOT NULL, -- 接收用户名称
    from_team_id          INTEGER NOT NULL REFERENCES teams(id), -- 转出团队id
    from_team_name        VARCHAR(255) NOT NULL, -- 转出团队名称
    amount_grams          DECIMAL(15,3) NOT NULL, -- 接收转账额度（克）
    notes                 TEXT NOT NULL DEFAULT '-', -- 转出方备注（从转出表复制过来）
    balance_after_receipt DECIMAL(15,3), -- 接收后账户余额
    status                VARCHAR(20) NOT NULL, -- 转入状态
    is_confirmed          BOOLEAN NOT NULL DEFAULT FALSE, -- 是否确认接收
    operational_user_id   INTEGER NOT NULL REFERENCES users(id), -- 操作用户id
    reception_rejection_reason TEXT NOT NULL DEFAULT '-', -- 拒收原因
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_user_from_team_to_user ON tea.user_from_team_transfer_in(to_user_id);
CREATE INDEX idx_tea_user_from_team_status ON tea.user_from_team_transfer_in(status);

-- 添加表注释
COMMENT ON TABLE tea.user_from_team_transfer_in IS '用户对团队转账接收记录表（匹配TeaUserFromTeamTransferIn结构体）';
COMMENT ON COLUMN tea.user_from_team_transfer_in.status IS '转入状态: completed-已完成, rejected-接收拒绝';

-- ============================================
-- 第二部分：团队相关表结构
-- ============================================

-- 团队茶叶账户表（完全匹配TeaTeamAccount结构体）
CREATE TABLE tea.team_accounts (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    team_id               INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    balance_grams         DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 茶叶数量(克)
    locked_balance_grams  DECIMAL(15,3) NOT NULL DEFAULT 0.000, -- 被锁定的茶叶数量(克)
    status                VARCHAR(20) NOT NULL DEFAULT 'normal', -- normal, frozen
    frozen_reason         TEXT, -- 冻结原因
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX idx_tea_team_accounts_team_id ON tea.team_accounts(team_id);
CREATE INDEX idx_tea_team_accounts_status ON tea.team_accounts(status);

-- 添加表注释
COMMENT ON TABLE tea.team_accounts IS '团队茶叶账户表（匹配TeaTeamAccount结构体）';

-- 团队对用户转账记录表（完全匹配TeaTeamToUserTransferOut结构体）
CREATE TABLE tea.team_to_user_transfer_out (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    from_team_id          INTEGER NOT NULL REFERENCES teams(id),
    from_team_name        VARCHAR(255) NOT NULL, -- 转出团队名称
    to_user_id            INTEGER NOT NULL REFERENCES users(id), -- 接收用户ID
    to_user_name          VARCHAR(255) NOT NULL, -- 接收用户名称
    initiator_user_id     INTEGER NOT NULL REFERENCES users(id), -- 发起转账的用户id
    amount_grams          DECIMAL(15,3) NOT NULL, -- 转账茶叶数量(克)
    notes                 TEXT NOT NULL DEFAULT '-', -- 转账备注
    status                VARCHAR(20) NOT NULL DEFAULT 'pending_approval', -- 转账状态
    approver_user_id      INTEGER REFERENCES users(id), -- 审批人ID
    approved_at           TIMESTAMP, -- 审批时间
    approval_rejection_reason TEXT  NOT NULL DEFAULT '-', -- 审批拒绝原因
    rejected_by           INTEGER REFERENCES users(id), -- 拒绝人ID
    rejected_at           TIMESTAMP, -- 拒绝时间
    balance_after_transfer DECIMAL(15,3), -- 转账后余额(克)
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at            TIMESTAMP NOT NULL, -- 转账请求过期时间
    payment_time          TIMESTAMP, -- 实际支付时间
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_team_to_user_from_team ON tea.team_to_user_transfer_out(from_team_id);
CREATE INDEX idx_tea_team_to_user_to_user ON tea.team_to_user_transfer_out(to_user_id);
CREATE INDEX idx_tea_team_to_user_status ON tea.team_to_user_transfer_out(status);
CREATE INDEX idx_tea_team_to_user_expires_at ON tea.team_to_user_transfer_out(expires_at);

-- 添加表注释
COMMENT ON TABLE tea.team_to_user_transfer_out IS '团队对用户转账记录表（匹配TeaTeamToUserTransferOut结构体）';
COMMENT ON COLUMN tea.team_to_user_transfer_out.status IS '转账状态: pending_approval-待审批, approved-审批通过, approval_rejected-审批拒绝, pending_receipt-待接收, completed-已完成, rejected-接收拒绝, expired-已超时';

-- 团队对团队转账记录表（完全匹配TeaTeamToTeamTransferOut结构体）
CREATE TABLE tea.team_to_team_transfer_out (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    from_team_id          INTEGER NOT NULL REFERENCES teams(id),
    from_team_name        VARCHAR(255) NOT NULL, -- 转出团队名称
    to_team_id            INTEGER NOT NULL REFERENCES teams(id), -- 接收团队ID
    to_team_name          VARCHAR(255) NOT NULL, -- 接收团队名称
    initiator_user_id     INTEGER NOT NULL REFERENCES users(id), -- 发起转账的用户id
    amount_grams          DECIMAL(15,3) NOT NULL, -- 转账茶叶数量(克)
    notes                 TEXT NOT NULL DEFAULT '-', -- 转账备注
    status                VARCHAR(20) NOT NULL DEFAULT 'pending_approval', -- 转账状态
    approver_user_id      INTEGER REFERENCES users(id), -- 审批人ID
    approved_at           TIMESTAMP, -- 审批时间
    approval_rejection_reason TEXT  NOT NULL DEFAULT '-', -- 审批拒绝原因
    rejected_by           INTEGER REFERENCES users(id), -- 拒绝人ID
    rejected_at           TIMESTAMP, -- 拒绝时间
    balance_after_transfer DECIMAL(15,3), -- 转账后余额(克)
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at            TIMESTAMP NOT NULL, -- 转账请求过期时间
    payment_time          TIMESTAMP, -- 实际支付时间
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_team_to_team_from_team ON tea.team_to_team_transfer_out(from_team_id);
CREATE INDEX idx_tea_team_to_team_to_team ON tea.team_to_team_transfer_out(to_team_id);
CREATE INDEX idx_tea_team_to_team_status ON tea.team_to_team_transfer_out(status);
CREATE INDEX idx_tea_team_to_team_expires_at ON tea.team_to_team_transfer_out(expires_at);

-- 添加表注释
COMMENT ON TABLE tea.team_to_team_transfer_out IS '团队对团队转账记录表（匹配TeaTeamToTeamTransferOut结构体）';
COMMENT ON COLUMN tea.team_to_team_transfer_out.status IS '转账状态: pending_approval-待审批, approved-审批通过, approval_rejected-审批拒绝, pending_receipt-待接收, completed-已完成, rejected-接收拒绝, expired-已超时';

-- 团队接收用户转入记录表（完全匹配TeaTeamFromUserTransferIn结构体）
CREATE TABLE tea.team_from_user_transfer_in (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    user_to_team_transfer_out_id INTEGER NOT NULL REFERENCES tea.user_to_team_transfer_out(id),
    to_team_id            INTEGER NOT NULL REFERENCES teams(id), -- 接收团队ID
    to_team_name          VARCHAR(255) NOT NULL, -- 接收团队名称
    from_user_id          INTEGER NOT NULL REFERENCES users(id), -- 转出用户ID
    from_user_name        VARCHAR(255) NOT NULL, -- 转出用户名称
    amount_grams          DECIMAL(15,3) NOT NULL, -- 转账茶叶数量(克)
    notes                 TEXT NOT NULL DEFAULT '-', -- 转账备注
    status                VARCHAR(20) NOT NULL, -- 转入状态
    is_confirmed          BOOLEAN NOT NULL DEFAULT FALSE, -- 是否确认接收
    operational_user_id   INTEGER NOT NULL REFERENCES users(id), -- 操作用户id
    reception_rejection_reason TEXT NOT NULL DEFAULT '-', -- 拒收原因
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_team_from_user_to_team ON tea.team_from_user_transfer_in(to_team_id);
CREATE INDEX idx_tea_team_from_user_status ON tea.team_from_user_transfer_in(status);

-- 添加表注释
COMMENT ON TABLE tea.team_from_user_transfer_in IS '团队接收用户转入记录表（匹配TeaTeamFromUserTransferIn结构体）';
COMMENT ON COLUMN tea.team_from_user_transfer_in.status IS '转入状态: completed-已完成, rejected-接收拒绝';

-- 团队接收团队转入记录表（完全匹配TeaTeamFromTeamTransferIn结构体）
CREATE TABLE tea.team_from_team_transfer_in (
    id                    SERIAL PRIMARY KEY,
    uuid                  UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    team_to_team_transfer_out_id INTEGER NOT NULL, -- 引用-团队对团队转出记录id
    to_team_id            INTEGER NOT NULL REFERENCES teams(id), -- 接收团队ID
    to_team_name          VARCHAR(255) NOT NULL, -- 接收团队名称
    from_team_id          INTEGER NOT NULL REFERENCES teams(id), -- 转出团队ID
    from_team_name        VARCHAR(255) NOT NULL, -- 转出团队名称
    amount_grams          DECIMAL(15,3) NOT NULL, -- 转账茶叶数量(克)
    notes                 TEXT NOT NULL DEFAULT '-', -- 转账备注
    status                VARCHAR(20) NOT NULL, -- 转入状态
    is_confirmed          BOOLEAN NOT NULL DEFAULT FALSE, -- 是否确认接收
    operational_user_id   INTEGER NOT NULL REFERENCES users(id), -- 操作用户id
    reception_rejection_reason TEXT NOT NULL DEFAULT '-', -- 拒收原因
    expires_at            TIMESTAMP NOT NULL, -- 过期时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tea_team_from_team_to_team ON tea.team_from_team_transfer_in(to_team_id);
CREATE INDEX idx_tea_team_from_team_status ON tea.team_from_team_transfer_in(status);

-- 添加表注释
COMMENT ON TABLE tea.team_from_team_transfer_in IS '团队接收团队转入记录表（匹配TeaTeamFromTeamTransferIn结构体）';
COMMENT ON COLUMN tea.team_from_team_transfer_in.status IS '转入状态: completed-已完成, rejected-接收拒绝';

-- ============================================
-- 第三部分：约束和触发器
-- ============================================

-- 用户账户状态枚举约束
ALTER TABLE tea.user_accounts ADD CONSTRAINT check_tea_user_account_status 
    CHECK (status IN ('normal', 'frozen'));

-- 用户对用户转账状态枚举约束
ALTER TABLE tea.user_to_user_transfer_out ADD CONSTRAINT check_tea_user_to_user_status 
    CHECK (status IN ('pending_receipt', 'completed', 'rejected', 'expired'));

-- 用户对团队转账状态枚举约束
ALTER TABLE tea.user_to_team_transfer_out ADD CONSTRAINT check_tea_user_to_team_status 
    CHECK (status IN ('pending_receipt', 'completed', 'rejected', 'expired'));

-- 用户对用户转账接收状态枚举约束
ALTER TABLE tea.user_from_user_transfer_in ADD CONSTRAINT check_tea_user_from_user_status 
    CHECK (status IN ('completed', 'rejected'));

-- 用户对团队转账接收状态枚举约束
ALTER TABLE tea.user_from_team_transfer_in ADD CONSTRAINT check_tea_user_from_team_status 
    CHECK (status IN ('completed', 'rejected'));

-- 团队账户状态枚举约束
ALTER TABLE tea.team_accounts ADD CONSTRAINT check_tea_team_account_status 
    CHECK (status IN ('normal', 'frozen'));

-- 团队对用户转账状态枚举约束
ALTER TABLE tea.team_to_user_transfer_out ADD CONSTRAINT check_tea_team_to_user_status 
    CHECK (status IN ('pending_approval', 'approved', 'approval_rejected', 'pending_receipt', 'completed', 'rejected', 'expired'));

-- 团队对团队转账状态枚举约束
ALTER TABLE tea.team_to_team_transfer_out ADD CONSTRAINT check_tea_team_to_team_status 
    CHECK (status IN ('pending_approval', 'approved', 'approval_rejected', 'pending_receipt', 'completed', 'rejected', 'expired'));

-- 团队接收用户转入状态枚举约束
ALTER TABLE tea.team_from_user_transfer_in ADD CONSTRAINT check_tea_team_from_user_status 
    CHECK (status IN ('completed', 'rejected'));

-- 团队接收团队转入状态枚举约束
ALTER TABLE tea.team_from_team_transfer_in ADD CONSTRAINT check_tea_team_from_team_status 
    CHECK (status IN ('completed', 'rejected'));

-- 金额不能为负数约束
ALTER TABLE tea.user_accounts ADD CONSTRAINT check_tea_user_account_balance_positive 
    CHECK (balance_grams >= 0);

ALTER TABLE tea.team_accounts ADD CONSTRAINT check_tea_team_account_balance_positive 
    CHECK (balance_grams >= 0);

-- 转账金额必须大于0
ALTER TABLE tea.user_to_user_transfer_out ADD CONSTRAINT check_tea_user_to_user_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE tea.user_to_team_transfer_out ADD CONSTRAINT check_tea_user_to_team_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE tea.team_to_user_transfer_out ADD CONSTRAINT check_tea_team_to_user_amount_positive 
    CHECK (amount_grams > 0);

ALTER TABLE tea.team_to_team_transfer_out ADD CONSTRAINT check_tea_team_to_team_amount_positive 
    CHECK (amount_grams > 0);

-- 不能给自己转账（用户对用户）
ALTER TABLE tea.user_to_user_transfer_out ADD CONSTRAINT check_tea_user_to_user_not_self 
    CHECK (from_user_id != to_user_id);

-- 不能给自己转账（团队对团队）
ALTER TABLE tea.team_to_team_transfer_out ADD CONSTRAINT check_tea_team_to_team_not_self 
    CHECK (from_team_id != to_team_id);

-- ============================================
-- 第四部分：触发器和视图
-- ============================================

-- 更新时间触发器函数
CREATE OR REPLACE FUNCTION update_tea_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有需要更新时间的表创建触发器
CREATE TRIGGER tea_user_accounts_updated_at_trigger
    BEFORE UPDATE ON tea.user_accounts
    FOR EACH ROW EXECUTE FUNCTION update_tea_updated_at();

CREATE TRIGGER tea_user_to_user_transfer_out_updated_at_trigger
    BEFORE UPDATE ON tea.user_to_user_transfer_out
    FOR EACH ROW EXECUTE FUNCTION update_tea_updated_at();

CREATE TRIGGER tea_user_to_team_transfer_out_updated_at_trigger
    BEFORE UPDATE ON tea.user_to_team_transfer_out
    FOR EACH ROW EXECUTE FUNCTION update_tea_updated_at();

CREATE TRIGGER tea_team_accounts_updated_at_trigger
    BEFORE UPDATE ON tea.team_accounts
    FOR EACH ROW EXECUTE FUNCTION update_tea_updated_at();

CREATE TRIGGER tea_team_to_user_transfer_out_updated_at_trigger
    BEFORE UPDATE ON tea.team_to_user_transfer_out
    FOR EACH ROW EXECUTE FUNCTION update_tea_updated_at();

CREATE TRIGGER tea_team_to_team_transfer_out_updated_at_trigger
    BEFORE UPDATE ON tea.team_to_team_transfer_out
    FOR EACH ROW EXECUTE FUNCTION update_tea_updated_at();

-- 用户账户汇总视图（基于新表结构）
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
    -- 待接收转账数量（用户对用户 + 用户对团队）
    (SELECT COUNT(*) FROM tea.user_to_user_transfer_out 
     WHERE to_user_id = u.id AND status = 'pending_receipt' AND expires_at > NOW()) as pending_user_received_count,
    -- 待接收团队转账数量（用户所属团队的待接收转账）
    (SELECT COUNT(*) FROM tea.user_to_team_transfer_out 
     WHERE to_team_id IN (SELECT team_id FROM team_members WHERE user_id = u.id AND status = '1') 
     AND status = 'pending_receipt' AND expires_at > NOW()) as pending_team_received_count,
    -- 发出转账数量（用户对用户 + 用户对团队）
    (SELECT COUNT(*) FROM tea.user_to_user_transfer_out 
     WHERE from_user_id = u.id AND status = 'pending_receipt' AND expires_at > NOW()) as pending_sent_count,
    -- 账户创建时间
    tua.created_at as account_created_at
FROM users u
LEFT JOIN tea.user_accounts tua ON u.id = tua.user_id;

COMMENT ON VIEW tea.user_account_summary IS '用户茶叶账户汇总信息视图（基于新表结构）';

-- 团队账户汇总视图（基于新表结构）
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
    -- 待接收转账数量（用户对团队 + 团队对团队）
    (SELECT COUNT(*) FROM tea.user_to_team_transfer_out 
     WHERE to_team_id = t.id AND status = 'pending_receipt' AND expires_at > NOW()) as pending_user_transfer_count,
    -- 待审批转账数量（团队对用户 + 团队对团队）
    (SELECT COUNT(*) FROM tea.team_to_user_transfer_out 
     WHERE from_team_id = t.id AND status = 'pending_approval' AND expires_at > NOW()) as pending_approval_count,
    -- 账户创建时间
    tta.created_at as account_created_at
FROM teams t
LEFT JOIN tea.team_accounts tta ON t.id = tta.team_id;

COMMENT ON VIEW tea.team_account_summary IS '团队茶叶账户汇总信息视图（基于新表结构）';

-- ============================================
-- 第五部分：一致性检查和维护函数
-- ============================================

-- 检查锁定余额一致性的函数
CREATE OR REPLACE FUNCTION check_tea_balance_consistency()
RETURNS TABLE(
    account_type TEXT,
    total_accounts BIGINT,
    negative_locked_balance BIGINT,
    negative_available_balance BIGINT,
    balance_mismatch BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM (
        -- 用户账户统计
        SELECT 
            'user_accounts' as account_type,
            COUNT(*) as total_accounts,
            COUNT(CASE WHEN tua.locked_balance_grams < 0 THEN 1 END) as negative_locked_balance,
            COUNT(CASE WHEN (tua.balance_grams - tua.locked_balance_grams) < 0 THEN 1 END) as negative_available_balance,
            COUNT(CASE WHEN ABS(tua.locked_balance_grams - COALESCE(upa.total_pending_amount, 0)) > 0.001 THEN 1 END) as balance_mismatch
        FROM tea.user_accounts tua
        LEFT JOIN (
            SELECT 
                from_user_id,
                COALESCE(SUM(amount_grams), 0) as total_pending_amount
            FROM (
                SELECT from_user_id, amount_grams FROM tea.user_to_user_transfer_out 
                WHERE status = 'pending_receipt' AND expires_at > NOW()
                UNION ALL
                SELECT from_user_id, amount_grams FROM tea.user_to_team_transfer_out 
                WHERE status = 'pending_receipt' AND expires_at > NOW()
            ) all_pending
            GROUP BY from_user_id
        ) upa ON tua.user_id = upa.from_user_id

        UNION ALL

        -- 团队账户统计
        SELECT 
            'team_accounts' as account_type,
            COUNT(*) as total_accounts,
            COUNT(CASE WHEN tta.locked_balance_grams < 0 THEN 1 END) as negative_locked_balance,
            COUNT(CASE WHEN (tta.balance_grams - tta.locked_balance_grams) < 0 THEN 1 END) as negative_available_balance,
            COUNT(CASE WHEN tta.locked_balance_grams > 0 THEN 1 END) as balance_mismatch
        FROM tea.team_accounts tta
    ) t;
END;
$$ LANGUAGE plpgsql;

-- 修复锁定余额不一致的函数
CREATE OR REPLACE FUNCTION fix_tea_balance_consistency()
RETURNS TABLE(
    action_type TEXT,
    affected_rows BIGINT
) AS $$
BEGIN
    -- 修复用户账户锁定余额
    WITH user_pending_amounts AS (
        SELECT 
            from_user_id,
            COALESCE(SUM(amount_grams), 0) as total_pending_amount
        FROM (
            SELECT from_user_id, amount_grams FROM tea.user_to_user_transfer_out 
            WHERE status = 'pending_receipt' AND expires_at > NOW()
            UNION ALL
            SELECT from_user_id, amount_grams FROM tea.user_to_team_transfer_out 
            WHERE status = 'pending_receipt' AND expires_at > NOW()
        ) all_pending
        GROUP BY from_user_id
    )
    UPDATE tea.user_accounts 
    SET locked_balance_grams = COALESCE(upa.total_pending_amount, 0)
    FROM user_pending_amounts upa
    WHERE tea.user_accounts.user_id = upa.from_user_id
    AND ABS(tea.user_accounts.locked_balance_grams - COALESCE(upa.total_pending_amount, 0)) > 0.001;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    RETURN QUERY SELECT 'user_accounts_updated' as action_type, affected_rows as affected_rows;
    
    -- 修复没有待确认转账的用户账户锁定余额
    UPDATE tea.user_accounts 
    SET locked_balance_grams = 0
    WHERE user_id NOT IN (
        SELECT DISTINCT from_user_id 
        FROM (
            SELECT from_user_id FROM tea.user_to_user_transfer_out 
            WHERE status = 'pending_receipt' AND expires_at > NOW()
            UNION ALL
            SELECT from_user_id FROM tea.user_to_team_transfer_out 
            WHERE status = 'pending_receipt' AND expires_at > NOW()
        ) all_pending
    ) AND locked_balance_grams != 0;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    RETURN QUERY SELECT 'user_accounts_reset' as action_type, affected_rows as affected_rows;
    
    -- 团队账户目前没有锁定余额机制，将其重置为0
    UPDATE tea.team_accounts 
    SET locked_balance_grams = 0
    WHERE locked_balance_grams != 0;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    RETURN QUERY SELECT 'team_accounts_reset' as action_type, affected_rows as affected_rows;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 第六部分：初始化数据
-- ============================================

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

-- ============================================
-- 第七部分：使用说明和示例查询
-- ============================================

/*
-- 查看用户茶叶账户信息
SELECT * FROM tea.user_accounts WHERE user_id = ?;

-- 查看用户对用户转账记录
SELECT * FROM tea.user_to_user_transfer_out WHERE from_user_id = ? OR to_user_id = ?;

-- 查看用户对团队转账记录
SELECT * FROM tea.user_to_team_transfer_out WHERE from_user_id = ? OR to_team_id = ?;

-- 查看团队对用户转账记录
SELECT * FROM tea.team_to_user_transfer_out WHERE from_team_id = ? OR to_user_id = ?;

-- 查看团队对团队转账记录
SELECT * FROM tea.team_to_team_transfer_out WHERE from_team_id = ? OR to_team_id = ?;

-- 查看待确认转账
SELECT * FROM tea.user_to_user_transfer_out 
WHERE status = 'pending_receipt' AND expires_at > CURRENT_TIMESTAMP;

-- 查看账户汇总信息
SELECT * FROM tea.user_account_summary WHERE user_id = ?;

-- 查看团队账户汇总信息
SELECT * FROM tea.team_account_summary WHERE team_id = ?;

-- 检查数据一致性
SELECT * FROM check_tea_balance_consistency();

-- 修复数据一致性
SELECT * FROM fix_tea_balance_consistency();

-- 检查系统状态
SELECT 
    (SELECT COUNT(*) FROM tea.user_accounts) as total_user_accounts,
    (SELECT COUNT(*) FROM tea.team_accounts) as total_team_accounts,
    (SELECT COUNT(*) FROM tea.user_to_user_transfer_out WHERE status = 'pending_receipt') as pending_user_to_user_transfers,
    (SELECT COUNT(*) FROM tea.user_to_team_transfer_out WHERE status = 'pending_receipt') as pending_user_to_team_transfers,
    (SELECT COUNT(*) FROM tea.team_to_user_transfer_out WHERE status = 'pending_approval') as pending_team_to_user_approvals,
    (SELECT COUNT(*) FROM tea.team_to_team_transfer_out WHERE status = 'pending_approval') as pending_team_to_team_approvals;
*/
