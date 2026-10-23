-- 增量迁移：为已有数据库补充 AGI 代理账号申报表。
-- 用法：psql -d teachat -f sql/migration_agent_profiles.sql
-- 可重复执行。

SET search_path TO chat, public;

-- 1) 归一化历史身份值（旧表单曾提交小写 agi/ufo，与常量 "AGI"/"UFO" 不一致）
UPDATE users SET role = 'AGI' WHERE lower(role) = 'agi';
UPDATE users SET role = 'UFO' WHERE lower(role) = 'ufo';

-- 2) AGI 不申报性别：清除历史误填的性别值，用 NULL 表示“未申报/不适用”
UPDATE users SET gender = NULL WHERE role = 'AGI' AND gender IS NOT NULL;

-- 3) AGI 代理账号申报表
CREATE TABLE IF NOT EXISTS agent_profiles (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    sponsor_team_id       INTEGER NOT NULL REFERENCES teams(id),
    sponsor_user_id       INTEGER NOT NULL REFERENCES users(id),
    model_family          VARCHAR(128) NOT NULL,
    model_version         VARCHAR(128) NOT NULL,
    provider              VARCHAR(128) NOT NULL DEFAULT '',
    agency_scope          VARCHAR(500) NOT NULL DEFAULT '',
    autonomy_level        INTEGER NOT NULL DEFAULT 1,
    declared_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMPTZ,
    revoked_at            TIMESTAMPTZ,
    revoked_reason        TEXT
);

CREATE INDEX IF NOT EXISTS idx_agent_profiles_sponsor_team ON agent_profiles(sponsor_team_id);
CREATE INDEX IF NOT EXISTS idx_agent_profiles_sponsor_user ON agent_profiles(sponsor_user_id);

COMMENT ON TABLE agent_profiles IS 'AGI代理账号的公开申报资料，赞助团队与赞助人承担背书责任';
COMMENT ON COLUMN agent_profiles.agency_scope IS '代理范围白名单，逗号分隔：read,post,thread,trade,sign';
COMMENT ON COLUMN agent_profiles.autonomy_level IS '自主程度：1-辅助，2-受监督，3-自主';
COMMENT ON COLUMN agent_profiles.revoked_at IS '赞助关系撤销时间，NULL表示有效';
