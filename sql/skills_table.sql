-- 技能表
CREATE TABLE skills (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER REFERENCES users(id),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    description           TEXT NOT NULL,
    strength_level        INTEGER NOT NULL DEFAULT 3,
    difficulty_level      INTEGER NOT NULL DEFAULT 3,
    category              INTEGER NOT NULL DEFAULT 2,
    level                 INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 用户技能记录表
CREATE TABLE skill_users (
    id                    SERIAL PRIMARY KEY,
    skill_id              INTEGER REFERENCES skills(id),
    user_id               INTEGER REFERENCES users(id),
    level                 INTEGER NOT NULL CHECK (level >= 1 AND level <= 9),
    status                INTEGER NOT NULL DEFAULT 2,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_skills_category ON skills(category);
CREATE INDEX idx_skills_strength_level ON skills(strength_level);
CREATE INDEX idx_skills_difficulty_level ON skills(difficulty_level);
CREATE INDEX idx_skills_level ON skills(level);
CREATE INDEX idx_skills_deleted_at ON skills(deleted_at);

CREATE INDEX idx_skill_users_user_id ON skill_users(user_id);
CREATE INDEX idx_skill_users_skill_id ON skill_users(skill_id);
CREATE INDEX idx_skill_users_status ON skill_users(status);
CREATE INDEX idx_skill_users_deleted_at ON skill_users(deleted_at);

-- 添加注释
COMMENT ON TABLE skills IS '技能表';
COMMENT ON COLUMN skills.id IS '主键ID';
COMMENT ON COLUMN skills.uuid IS '唯一标识符';
COMMENT ON COLUMN skills.name IS '技能名称';
COMMENT ON COLUMN skills.nickname IS '技能别名';
COMMENT ON COLUMN skills.description IS '技能描述';
COMMENT ON COLUMN skills.strength_level IS '体力要求等级(1-5)';
COMMENT ON COLUMN skills.difficulty_level IS '掌握难度等级(1-5)';
COMMENT ON COLUMN skills.category IS '技能分类(1-通用软技能,2-通用硬技能)';
COMMENT ON COLUMN skills.level IS '技能等级(1-5)';
COMMENT ON COLUMN skills.created_at IS '创建时间';
COMMENT ON COLUMN skills.updated_at IS '更新时间';
COMMENT ON COLUMN skills.deleted_at IS '软删除时间';

COMMENT ON TABLE skill_users IS '用户技能记录表';
COMMENT ON COLUMN skill_users.id IS '主键ID';
COMMENT ON COLUMN skill_users.skill_id IS '技能ID';
COMMENT ON COLUMN skill_users.user_id IS '用户ID';
COMMENT ON COLUMN skill_users.level IS '用户掌握该技能的等级(1-9)';
COMMENT ON COLUMN skill_users.status IS '用户技能状态(0-失能,1-弱能,2-中能,3-强能)';
COMMENT ON COLUMN skill_users.created_at IS '创建时间';
COMMENT ON COLUMN skill_users.updated_at IS '更新时间';
COMMENT ON COLUMN skill_users.deleted_at IS '软删除时间';