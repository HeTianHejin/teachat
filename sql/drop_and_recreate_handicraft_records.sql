-- 删除旧表
DROP TABLE IF EXISTS endings CASCADE;
DROP TABLE IF EXISTS process_records CASCADE;
DROP TABLE IF EXISTS inaugurations CASCADE;

-- 创建开工仪式表
CREATE TABLE inaugurations (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    recorder_user_id      INTEGER REFERENCES users(id),
    evidence_id           INTEGER DEFAULT 0,
    status                INTEGER NOT NULL DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 创建过程记录表
CREATE TABLE process_records (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    recorder_user_id      INTEGER REFERENCES users(id),
    evidence_id           INTEGER DEFAULT 0,
    status                INTEGER NOT NULL DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 创建结束仪式表
CREATE TABLE endings (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    recorder_user_id      INTEGER REFERENCES users(id),
    evidence_id           INTEGER DEFAULT 0,
    status                INTEGER NOT NULL DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_inaugurations_handicraft_id ON inaugurations(handicraft_id);
CREATE INDEX idx_inaugurations_recorder_user_id ON inaugurations(recorder_user_id);

CREATE INDEX idx_process_records_handicraft_id ON process_records(handicraft_id);
CREATE INDEX idx_process_records_recorder_user_id ON process_records(recorder_user_id);
CREATE INDEX idx_process_records_deleted_at ON process_records(deleted_at);

CREATE INDEX idx_endings_handicraft_id ON endings(handicraft_id);
CREATE INDEX idx_endings_recorder_user_id ON endings(recorder_user_id);

-- 添加注释
COMMENT ON TABLE inaugurations IS '开工仪式表';
COMMENT ON COLUMN inaugurations.handicraft_id IS '手工艺ID';
COMMENT ON COLUMN inaugurations.recorder_user_id IS '记录人用户ID';
COMMENT ON COLUMN inaugurations.evidence_id IS '证据ID，0表示无';

COMMENT ON TABLE process_records IS '过程记录表';
COMMENT ON COLUMN process_records.handicraft_id IS '手工艺ID';
COMMENT ON COLUMN process_records.recorder_user_id IS '记录人用户ID';
COMMENT ON COLUMN process_records.evidence_id IS '证据ID，0表示无';

COMMENT ON TABLE endings IS '结束仪式表';
COMMENT ON COLUMN endings.handicraft_id IS '手工艺ID';
COMMENT ON COLUMN endings.recorder_user_id IS '记录人用户ID';
COMMENT ON COLUMN endings.evidence_id IS '证据ID，0表示无';
