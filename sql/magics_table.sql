-- 法力表
CREATE TABLE magics (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER REFERENCES users(id),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    description           TEXT NOT NULL,
    intelligence_level    INTEGER NOT NULL DEFAULT 3,
    difficulty_level      INTEGER NOT NULL DEFAULT 3,
    category              INTEGER NOT NULL DEFAULT 1,
    level                 INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 用户法力记录表
CREATE TABLE magic_users (
    id                    SERIAL PRIMARY KEY,
    magic_id              INTEGER REFERENCES magics(id),
    user_id               INTEGER REFERENCES users(id),
    level                 INTEGER NOT NULL CHECK (level >= 1 AND level <= 9),
    status                INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_magics_category ON magics(category);
CREATE INDEX idx_magics_intelligence_level ON magics(intelligence_level);
CREATE INDEX idx_magics_difficulty_level ON magics(difficulty_level);
CREATE INDEX idx_magics_level ON magics(level);
CREATE INDEX idx_magics_deleted_at ON magics(deleted_at);

CREATE INDEX idx_magic_users_user_id ON magic_users(user_id);
CREATE INDEX idx_magic_users_magic_id ON magic_users(magic_id);
CREATE INDEX idx_magic_users_status ON magic_users(status);
CREATE INDEX idx_magic_users_deleted_at ON magic_users(deleted_at);

-- 添加注释
COMMENT ON TABLE magics IS '法力表';
COMMENT ON COLUMN magics.id IS '主键ID';
COMMENT ON COLUMN magics.uuid IS '唯一标识符';
COMMENT ON COLUMN magics.user_id IS '创建者用户ID';
COMMENT ON COLUMN magics.name IS '法力名称';
COMMENT ON COLUMN magics.nickname IS '法力别名';
COMMENT ON COLUMN magics.description IS '法力描述';
COMMENT ON COLUMN magics.intelligence_level IS '智力要求等级(1-5)';
COMMENT ON COLUMN magics.difficulty_level IS '掌握难度等级(1-5)';
COMMENT ON COLUMN magics.category IS '法力分类(1-理性,2-感性)';
COMMENT ON COLUMN magics.level IS '法力等级(1-5)';
COMMENT ON COLUMN magics.created_at IS '创建时间';
COMMENT ON COLUMN magics.updated_at IS '更新时间';
COMMENT ON COLUMN magics.deleted_at IS '软删除时间';

COMMENT ON TABLE magic_users IS '用户法力记录表';
COMMENT ON COLUMN magic_users.id IS '主键ID';
COMMENT ON COLUMN magic_users.magic_id IS '法力ID';
COMMENT ON COLUMN magic_users.user_id IS '用户ID';
COMMENT ON COLUMN magic_users.level IS '用户掌握该法力的段位(1-9)';
COMMENT ON COLUMN magic_users.status IS '用户法力状态(0-迷糊,1-清醒,2-专注,3-灵感迸发)';
COMMENT ON COLUMN magic_users.created_at IS '创建时间';
COMMENT ON COLUMN magic_users.updated_at IS '更新时间';
COMMENT ON COLUMN magic_users.deleted_at IS '软删除时间';