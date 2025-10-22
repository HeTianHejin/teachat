-- 团队技能记录表
CREATE TABLE skill_teams (
    id                    SERIAL PRIMARY KEY,
    skill_id              INTEGER REFERENCES skills(id),
    team_id               INTEGER REFERENCES teams(id),
    level                 INTEGER NOT NULL CHECK (level >= 1 AND level <= 9),
    status                INTEGER NOT NULL DEFAULT 2,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_skill_teams_team_id ON skill_teams(team_id);
CREATE INDEX idx_skill_teams_skill_id ON skill_teams(skill_id);
CREATE INDEX idx_skill_teams_status ON skill_teams(status);
CREATE INDEX idx_skill_teams_deleted_at ON skill_teams(deleted_at);

-- 添加注释
COMMENT ON TABLE skill_teams IS '团队技能记录表';
COMMENT ON COLUMN skill_teams.id IS '主键ID';
COMMENT ON COLUMN skill_teams.skill_id IS '技能ID';
COMMENT ON COLUMN skill_teams.team_id IS '团队ID';
COMMENT ON COLUMN skill_teams.level IS '团队掌握该技能的等级(1-9)';
COMMENT ON COLUMN skill_teams.status IS '团队技能状态(0-不可用,1-低效,2-正常,3-高效)';
COMMENT ON COLUMN skill_teams.created_at IS '创建时间';
COMMENT ON COLUMN skill_teams.updated_at IS '更新时间';
COMMENT ON COLUMN skill_teams.deleted_at IS '软删除时间';

-- 团队法力记录表
CREATE TABLE magic_teams (
    id                    SERIAL PRIMARY KEY,
    magic_id              INTEGER REFERENCES magics(id),
    team_id               INTEGER REFERENCES teams(id),
    level                 INTEGER NOT NULL CHECK (level >= 1 AND level <= 9),
    status                INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_magic_teams_team_id ON magic_teams(team_id);
CREATE INDEX idx_magic_teams_magic_id ON magic_teams(magic_id);
CREATE INDEX idx_magic_teams_status ON magic_teams(status);
CREATE INDEX idx_magic_teams_deleted_at ON magic_teams(deleted_at);

-- 添加注释
COMMENT ON TABLE magic_teams IS '团队法力记录表';
COMMENT ON COLUMN magic_teams.id IS '主键ID';
COMMENT ON COLUMN magic_teams.magic_id IS '法力ID';
COMMENT ON COLUMN magic_teams.team_id IS '团队ID';
COMMENT ON COLUMN magic_teams.level IS '团队掌握该法力的段位(1-9)';
COMMENT ON COLUMN magic_teams.status IS '团队法力状态(0-混乱,1-清晰,2-协同,3-创新)';
COMMENT ON COLUMN magic_teams.created_at IS '创建时间';
COMMENT ON COLUMN magic_teams.updated_at IS '更新时间';
COMMENT ON COLUMN magic_teams.deleted_at IS '软删除时间';
