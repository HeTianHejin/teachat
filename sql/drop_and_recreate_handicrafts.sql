-- 删除所有手艺相关表（按依赖顺序）
DROP TABLE IF EXISTS handicraft_evidences CASCADE;
DROP TABLE IF EXISTS handicraft_magics CASCADE;
DROP TABLE IF EXISTS handicraft_skills CASCADE;
DROP TABLE IF EXISTS handicraft_contributors CASCADE;
DROP TABLE IF EXISTS handicrafts CASCADE;

-- 重新创建手艺表
CREATE TABLE handicrafts (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    recorder_user_id      INTEGER REFERENCES users(id),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    description           TEXT NOT NULL,
    project_id            INTEGER REFERENCES projects(id),
    initiator_id          INTEGER REFERENCES users(id),
    owner_id              INTEGER REFERENCES users(id),
    category              INTEGER NOT NULL DEFAULT 1,
    status                INTEGER NOT NULL DEFAULT 0,
    skill_difficulty      INTEGER NOT NULL DEFAULT 3,
    magic_difficulty      INTEGER NOT NULL DEFAULT 3,
    contributor_count     INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 手艺协助者表
CREATE TABLE handicraft_contributors (
    id                    SERIAL PRIMARY KEY,
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    user_id               INTEGER REFERENCES users(id),
    contribution_rate     INTEGER DEFAULT 50,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 手艺技能关联表
CREATE TABLE handicraft_skills (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    skill_id              INTEGER REFERENCES skills(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 手艺法力关联表
CREATE TABLE handicraft_magics (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    magic_id              INTEGER REFERENCES magics(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 手艺凭据关联表
CREATE TABLE handicraft_evidences (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    evidence_id           INTEGER REFERENCES evidences(id),
    note                  TEXT,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_handicrafts_project_id ON handicrafts(project_id);
CREATE INDEX idx_handicrafts_recorder_user_id ON handicrafts(recorder_user_id);
CREATE INDEX idx_handicrafts_category ON handicrafts(category);
CREATE INDEX idx_handicrafts_status ON handicrafts(status);
CREATE INDEX idx_handicrafts_deleted_at ON handicrafts(deleted_at);

CREATE INDEX idx_handicraft_contributors_handicraft_id ON handicraft_contributors(handicraft_id);
CREATE INDEX idx_handicraft_contributors_user_id ON handicraft_contributors(user_id);

CREATE INDEX idx_handicraft_skills_handicraft_id ON handicraft_skills(handicraft_id);
CREATE INDEX idx_handicraft_skills_skill_id ON handicraft_skills(skill_id);

CREATE INDEX idx_handicraft_magics_handicraft_id ON handicraft_magics(handicraft_id);
CREATE INDEX idx_handicraft_magics_magic_id ON handicraft_magics(magic_id);

CREATE INDEX idx_handicraft_evidences_handicraft_id ON handicraft_evidences(handicraft_id);
CREATE INDEX idx_handicraft_evidences_evidence_id ON handicraft_evidences(evidence_id);

-- 添加注释
COMMENT ON TABLE handicrafts IS '手艺表';
COMMENT ON COLUMN handicrafts.recorder_user_id IS '记录人用户ID';
COMMENT ON COLUMN handicraft_contributors.contribution_rate IS '贡献值(1-100)';
