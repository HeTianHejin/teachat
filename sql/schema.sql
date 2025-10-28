-- TeaChat 数据库架构定义
-- 创建数据库
DROP DATABASE IF EXISTS teachat;
CREATE DATABASE teachat;

-- ============================================
-- 核心用户与组织表
-- ============================================

-- 用户表
CREATE TABLE users (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name                  VARCHAR(255),
    email                 VARCHAR(255) NOT NULL UNIQUE,
    password              VARCHAR(255) NOT NULL,
    biography             TEXT,
    role                  VARCHAR(64),
    gender                INTEGER,
    avatar                VARCHAR(255),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 家庭表
CREATE TABLE families (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(255) DEFAULT gen_random_uuid(),
    author_id             INTEGER,
    name                  VARCHAR(255),
    introduction          TEXT,
    is_married            BOOLEAN DEFAULT true,
    has_child             BOOLEAN,
    husband_from_family_id INTEGER DEFAULT 0,
    wife_from_family_id   INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 1,
    logo                  VARCHAR(255),
    is_open               BOOLEAN DEFAULT true,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 团队表
CREATE TABLE teams (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name                  VARCHAR(255),
    mission               TEXT,
    founder_id            INTEGER REFERENCES users(id),
    class                 INTEGER,
    abbreviation          INTEGER,
    logo                  VARCHAR(255),
    superior_team_id      INTEGER DEFAULT 0,
    subordinate_team_id   INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- ============================================
-- 地址与场所表
-- ============================================

-- 地址表
CREATE TABLE addresses (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    nation                VARCHAR(255),
    province              VARCHAR(255),
    city                  VARCHAR(255),
    district              VARCHAR(255),
    town                  VARCHAR(255),
    village               VARCHAR(255),
    street                VARCHAR(255),
    building              VARCHAR(255),
    unit                  VARCHAR(255),
    portal_number         VARCHAR(255),
    postal_code           VARCHAR(20) DEFAULT '0',
    category              INTEGER DEFAULT 0,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 场所表
CREATE TABLE places (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    description           TEXT,
    icon                  VARCHAR(255) DEFAULT 'bootstrap-icons/bank.svg',
    occupant_user_id      INTEGER,
    owner_user_id         INTEGER,
    level                 INTEGER DEFAULT 0,
    category              INTEGER DEFAULT 0,
    is_public             BOOLEAN DEFAULT true,
    is_government         BOOLEAN DEFAULT false,
    user_id               INTEGER,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- ============================================
-- 关系映射表
-- ============================================

-- 家庭成员表
CREATE TABLE family_members (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(255) DEFAULT gen_random_uuid(),
    family_id             INTEGER,
    user_id               INTEGER,
    role                  INTEGER DEFAULT 0,
    is_adult              BOOLEAN DEFAULT true,
    nick_name             VARCHAR(255) DEFAULT ':P',
    is_adopted            BOOLEAN DEFAULT false,
    age                   INTEGER DEFAULT 0,
    order_of_seniority    INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 团队成员表
CREATE TABLE team_members (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    team_id               INTEGER REFERENCES teams(id),
    user_id               INTEGER REFERENCES users(id),
    role                  VARCHAR(255),
    class                 INTEGER DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 用户默认家庭表
CREATE TABLE user_default_families (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER,
    family_id             INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 用户默认团队表
CREATE TABLE user_default_teams (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER REFERENCES users(id),
    team_id               INTEGER REFERENCES teams(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 用户场所关联表
CREATE TABLE user_place (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER,
    place_id              INTEGER,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户默认场所表
CREATE TABLE user_default_place (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER,
    place_id              INTEGER,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户地址关联表
CREATE TABLE user_address (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER REFERENCES users(id),
    address_id            INTEGER REFERENCES addresses(id),
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户默认地址表
CREATE TABLE user_default_address (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER REFERENCES users(id),
    address_id            INTEGER REFERENCES addresses(id),
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 场所地址关联表
CREATE TABLE place_addresses (
    place_id              INTEGER REFERENCES places(id),
    address_id            INTEGER REFERENCES addresses(id),
    PRIMARY KEY (place_id, address_id)
);

-- ============================================
-- 会话与认证表
-- ============================================

-- 会话表
CREATE TABLE sessions (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    email                 VARCHAR(255),
    user_id               INTEGER REFERENCES users(id),
    gender                INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 最近查询表
CREATE TABLE last_queries (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER REFERENCES users(id),
    path                  VARCHAR(255),
    query                 VARCHAR(255),
    query_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 目标与项目表
-- ============================================

-- 目标表
CREATE TABLE objectives (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    title                 VARCHAR(64) NOT NULL,
    body                  TEXT,
    user_id               INTEGER NOT NULL,
    class                 INTEGER NOT NULL,
    family_id             INTEGER DEFAULT 0,
    team_id               INTEGER NOT NULL DEFAULT 2,
    cover                 VARCHAR(64),
    is_private            BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edit_at               TIMESTAMP
);

-- 项目表
CREATE TABLE projects (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    title                 VARCHAR(64) NOT NULL,
    body                  TEXT,
    objective_id          INTEGER,
    user_id               INTEGER,
    class                 INTEGER,
    family_id             INTEGER NOT NULL DEFAULT 0,
    team_id               INTEGER NOT NULL DEFAULT 2,
    cover                 VARCHAR(64) DEFAULT 'default-pr-cover',
    is_private            BOOLEAN DEFAULT false,
    status                INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edit_at               TIMESTAMP
);

-- 项目批准表
CREATE TABLE project_approved (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER NOT NULL,
    project_id            INTEGER NOT NULL,
    objective_id          INTEGER NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 项目邀请团队表
CREATE TABLE project_invited_teams (
    id                    SERIAL PRIMARY KEY,
    project_id            INTEGER REFERENCES projects(id),
    team_id               INTEGER REFERENCES teams(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 目标邀请团队表
CREATE TABLE objective_invited_teams (
    id                    SERIAL PRIMARY KEY,
    objective_id          INTEGER REFERENCES objectives(id),
    team_id               INTEGER REFERENCES teams(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 项目场所关联表
CREATE TABLE project_place (
    id                    SERIAL PRIMARY KEY,
    project_id            INTEGER,
    place_id              INTEGER,
    user_id               INTEGER,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 项目预约表
CREATE TABLE project_appointments (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    project_id            INTEGER NOT NULL,
    note                  VARCHAR(255),
    start_time            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    place_id              INTEGER NOT NULL DEFAULT 0,
    payer_user_id         INTEGER,
    payer_team_id         INTEGER,
    payer_family_id       INTEGER,
    payee_user_id         INTEGER,
    payee_team_id         INTEGER,
    payee_family_id       INTEGER,
    verifier_user_id      INTEGER,
    verifier_family_id    INTEGER,
    verifier_team_id      INTEGER,
    status                SMALLINT NOT NULL DEFAULT 0,
    confirmed_at          TIMESTAMP,
    rejected_at           TIMESTAMP,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 帖子与讨论表
-- ============================================

-- 帖子表
CREATE TABLE posts (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    body                  TEXT,
    user_id               INTEGER REFERENCES users(id),
    thread_id             INTEGER,
    attitude              BOOLEAN,
    family_id             INTEGER DEFAULT 0,
    team_id               INTEGER NOT NULL DEFAULT 2,
    is_private            BOOLEAN DEFAULT false,
    class                 INTEGER DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edit_at               TIMESTAMP
);

-- 主题表
CREATE TABLE threads (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    title                 VARCHAR(64),
    body                  TEXT,
    user_id               INTEGER REFERENCES users(id),
    project_id            INTEGER REFERENCES projects(id),
    post_id               INTEGER,
    class                 INTEGER DEFAULT 10,
    family_id             INTEGER DEFAULT 0,
    team_id               INTEGER NOT NULL DEFAULT 2,
    type                  INTEGER DEFAULT 0,
    category              INTEGER DEFAULT 0,
    is_private            BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edit_at               TIMESTAMP
);

-- 草稿帖子表
CREATE TABLE draft_posts (
    id                    SERIAL PRIMARY KEY,
    body                  TEXT,
    user_id               INTEGER,
    thread_id             INTEGER,
    attitude              BOOLEAN,
    class                 INTEGER DEFAULT 0,
    team_id               INTEGER NOT NULL DEFAULT 2,
    is_private            BOOLEAN DEFAULT false,
    family_id             INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 草稿主题表
CREATE TABLE draft_threads (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER NOT NULL,
    project_id            INTEGER NOT NULL,
    title                 VARCHAR(64),
    body                  TEXT,
    class                 INTEGER DEFAULT 0,
    type                  INTEGER DEFAULT 0,
    post_id               INTEGER DEFAULT 0,
    team_id               INTEGER NOT NULL DEFAULT 2,
    is_private            BOOLEAN DEFAULT false,
    family_id             INTEGER DEFAULT 0,
    category              INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 阅读记录表
CREATE TABLE reads (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER,
    thread_id             INTEGER REFERENCES threads(id),
    read_at               TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 主题批准表
CREATE TABLE thread_approved (
    id                    SERIAL PRIMARY KEY,
    project_id            INTEGER NOT NULL REFERENCES projects(id),
    thread_id             INTEGER NOT NULL REFERENCES threads(id),
    user_id               INTEGER NOT NULL REFERENCES users(id),
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 邀请与消息表
-- ============================================

-- 邀请表
CREATE TABLE invitations (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    team_id               INTEGER REFERENCES teams(id),
    invite_email          VARCHAR(255),
    role                  VARCHAR(50),
    invite_word           TEXT,
    status                INTEGER,
    author_user_id        INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 邀请回复表
CREATE TABLE invitation_replies (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(50) NOT NULL DEFAULT gen_random_uuid(),
    invitation_id         INTEGER,
    user_id               INTEGER,
    reply_word            TEXT NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 接受对象表
CREATE TABLE accept_objects (
    id                    SERIAL PRIMARY KEY,
    object_type           INTEGER DEFAULT 0,
    object_id             INTEGER
);

-- 接受消息表
CREATE TABLE accept_messages (
    id                    SERIAL PRIMARY KEY,
    from_user_id          INTEGER REFERENCES users(id),
    to_user_id            INTEGER REFERENCES users(id),
    title                 VARCHAR(64),
    content               TEXT,
    accept_object_id      INTEGER REFERENCES accept_objects(id),
    class                 INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 接受记录表
CREATE TABLE acceptances (
    id                    SERIAL PRIMARY KEY,
    accept_object_id      INTEGER,
    x_accept              BOOLEAN DEFAULT false,
    x_user_id             INTEGER,
    x_accepted_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    y_accept              BOOLEAN DEFAULT false,
    y_user_id             INTEGER DEFAULT 0,
    y_accepted_at         TIMESTAMP
);

-- 新消息计数表
CREATE TABLE new_message_counts (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER,
    count                 INTEGER DEFAULT 0
);

-- ============================================
-- 成员管理表
-- ============================================

-- 家庭成员签到表
CREATE TABLE family_member_sign_ins (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(255) DEFAULT gen_random_uuid(),
    family_id             INTEGER,
    user_id               INTEGER,
    role                  INTEGER DEFAULT 0,
    is_adult              BOOLEAN DEFAULT true,
    title                 VARCHAR(255),
    content               TEXT,
    place_id              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    is_adopted            BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 家庭成员签到回复表
CREATE TABLE family_member_sign_in_replies (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(255) DEFAULT gen_random_uuid(),
    sign_in_id            INTEGER,
    user_id               INTEGER,
    is_confirm            BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 家庭成员签出表
CREATE TABLE family_member_sign_outs (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(255) DEFAULT gen_random_uuid(),
    family_id             INTEGER,
    user_id               INTEGER,
    role                  INTEGER DEFAULT 0,
    is_adult              BOOLEAN DEFAULT true,
    title                 VARCHAR(255),
    content               TEXT,
    place_id              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    is_adopted            BOOLEAN DEFAULT false,
    author_user_id        INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 团队成员角色通知表
CREATE TABLE team_member_role_notices (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    team_id               INTEGER REFERENCES teams(id),
    ceo_id                INTEGER REFERENCES users(id),
    member_id             INTEGER,
    member_current_role   VARCHAR(64),
    new_role              VARCHAR(64),
    title                 VARCHAR(64),
    content               TEXT,
    status                INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 团队成员辞职表
CREATE TABLE team_member_resignations (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) DEFAULT gen_random_uuid(),
    team_id               INTEGER,
    ceo_user_id           INTEGER,
    core_member_user_id   INTEGER,
    member_id             INTEGER,
    member_user_id        INTEGER,
    member_current_role   VARCHAR(36),
    title                 VARCHAR(255),
    content               TEXT,
    status                SMALLINT,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 成员申请表
CREATE TABLE member_applications (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) DEFAULT gen_random_uuid(),
    team_id               INTEGER,
    user_id               INTEGER,
    content               TEXT,
    status                SMALLINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 成员申请回复表
CREATE TABLE member_application_replies (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) DEFAULT gen_random_uuid(),
    member_application_id INTEGER,
    team_id               INTEGER,
    user_id               INTEGER,
    reply_content         VARCHAR(255),
    status                SMALLINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- ============================================
-- 物资管理表
-- ============================================

-- 物资表
CREATE TABLE goods (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    recorder_user_id      INTEGER,
    name                  VARCHAR(255),
    nickname              VARCHAR(255),
    designer              VARCHAR(255),
    describe              TEXT,
    price                 FLOAT,
    applicability         VARCHAR(255),
    category              INTEGER,
    specification         VARCHAR(255),
    brand_name            VARCHAR(255),
    model                 VARCHAR(255),
    weight                FLOAT,
    dimensions            VARCHAR(255),
    material              VARCHAR(255),
    size                  VARCHAR(255),
    color                 VARCHAR(255),
    network_connection_type VARCHAR(255),
    features              INTEGER,
    serial_number         VARCHAR(255),
    physical_state        INTEGER DEFAULT 0,
    operational_state     INTEGER DEFAULT 0,
    origin                VARCHAR(255),
    manufacturer          VARCHAR(255),
    manufacturer_url      VARCHAR(255),
    engine_type           VARCHAR(255),
    purchase_url          VARCHAR(255),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 团队物资表
CREATE TABLE goods_teams (
    id                    SERIAL PRIMARY KEY,
    team_id               INTEGER,
    goods_id              INTEGER,
    availability          INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 用户物资表
CREATE TABLE goods_users (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER,
    goods_id              INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 家庭物资表
CREATE TABLE goods_families (
    id                    SERIAL PRIMARY KEY,
    family_id             INTEGER,
    goods_id              INTEGER,
    availability          INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 项目物资表
CREATE TABLE goods_projects (
    id                    SERIAL PRIMARY KEY,
    project_id            INTEGER NOT NULL REFERENCES projects(id),
    responsible_user_id   INTEGER NOT NULL REFERENCES users(id),
    goods_id              INTEGER NOT NULL REFERENCES goods(id),
    provider_type         INTEGER NOT NULL DEFAULT 1,
    expected_usage        TEXT,
    quantity              INTEGER NOT NULL DEFAULT 1,
    category              INTEGER NOT NULL DEFAULT 1,
    status                INTEGER NOT NULL DEFAULT 0,
    notes                 TEXT,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 项目物资准备状态表
CREATE TABLE goods_project_readiness (
    id                    SERIAL PRIMARY KEY,
    project_id            INTEGER NOT NULL,
    is_ready              BOOLEAN NOT NULL DEFAULT FALSE,
    user_id               INTEGER NOT NULL,
    notes                 TEXT,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE
);

-- ============================================
-- 技能与法力表
-- ============================================

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

-- ============================================
-- 手工艺表
-- ============================================

-- 手工艺表
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
    type                  INTEGER NOT NULL DEFAULT 1,
    category              INTEGER NOT NULL DEFAULT 0,
    status                INTEGER NOT NULL DEFAULT 0,
    skill_difficulty      INTEGER NOT NULL DEFAULT 3,
    magic_difficulty      INTEGER NOT NULL DEFAULT 3,
    contributor_count     INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 手工艺协助者表
CREATE TABLE handicraft_contributors (
    id                    SERIAL PRIMARY KEY,
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    user_id               INTEGER REFERENCES users(id),
    contribution_rate     INTEGER DEFAULT 50,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 手工艺技能关联表
CREATE TABLE handicraft_skills (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    skill_id              INTEGER REFERENCES skills(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 手工艺法力关联表
CREATE TABLE handicraft_magics (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    handicraft_id         INTEGER REFERENCES handicrafts(id),
    magic_id              INTEGER REFERENCES magics(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 开工仪式表
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

-- 过程记录表
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

-- 结束仪式表
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

-- ============================================
-- 凭据表
-- ============================================

-- 凭据主表
CREATE TABLE evidences (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) NOT NULL UNIQUE,
    description           TEXT,
    recorder_user_id      INTEGER NOT NULL REFERENCES users(id),
    note                  TEXT,
    category              INTEGER NOT NULL DEFAULT 0,
    path                  VARCHAR(500),
    original_url          VARCHAR(500),
    filename              VARCHAR(255),
    mime_type             VARCHAR(100),
    file_size             BIGINT NOT NULL DEFAULT 0,
    file_hash             VARCHAR(64),
    width                 INTEGER DEFAULT 0,
    height                INTEGER DEFAULT 0,
    duration              INTEGER DEFAULT 0,
    visibility            INTEGER NOT NULL DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 手工艺凭据关联表
CREATE TABLE handicraft_evidences (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) NOT NULL UNIQUE,
    handicraft_id         INTEGER NOT NULL REFERENCES handicrafts(id),
    evidence_id           INTEGER NOT NULL REFERENCES evidences(id),
    note                  TEXT,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- ============================================
-- 环境与安全表
-- ============================================

-- 环境表
CREATE TABLE environments (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name                  VARCHAR(255),
    summary               TEXT,
    temperature           INTEGER,
    humidity              INTEGER,
    pm25                  INTEGER,
    noise                 INTEGER,
    light                 INTEGER,
    wind                  INTEGER,
    flow                  INTEGER,
    rain                  INTEGER,
    pressure              INTEGER,
    smoke                 INTEGER,
    dust                  INTEGER,
    odor                  INTEGER,
    visibility            INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 隐患表
CREATE TABLE hazards (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER REFERENCES users(id),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    keywords              VARCHAR(255),
    description           TEXT,
    source                VARCHAR(255),
    severity              INTEGER NOT NULL DEFAULT 1,
    category              INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 风险表
CREATE TABLE risks (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER REFERENCES users(id),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    keywords              VARCHAR(255),
    description           TEXT,
    source                VARCHAR(255),
    severity              INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 安全措施表
CREATE TABLE safety_measures (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    hazard_id             INTEGER REFERENCES hazards(id),
    user_id               INTEGER REFERENCES users(id),
    title                 VARCHAR(255) NOT NULL,
    description           TEXT,
    priority              INTEGER NOT NULL DEFAULT 3,
    status                INTEGER NOT NULL DEFAULT 1,
    planned_date          TIMESTAMP,
    completed_date        TIMESTAMP,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 安全防护表
CREATE TABLE safety_protections (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    risk_id               INTEGER REFERENCES risks(id),
    user_id               INTEGER REFERENCES users(id),
    title                 VARCHAR(255) NOT NULL,
    description           TEXT,
    type                  INTEGER NOT NULL DEFAULT 1,
    priority              INTEGER NOT NULL DEFAULT 3,
    status                INTEGER NOT NULL DEFAULT 1,
    equipment             VARCHAR(255),
    planned_date          TIMESTAMP,
    completed_date        TIMESTAMP,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- ============================================
-- 看看(检查)相关表
-- ============================================

-- 看看主表
CREATE TABLE see_seeks (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    description           TEXT,
    place_id              INTEGER,
    project_id            INTEGER,
    start_time            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time              TIMESTAMP DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    payer_user_id         INTEGER,
    payer_team_id         INTEGER,
    payer_family_id       INTEGER,
    payee_user_id         INTEGER,
    payee_team_id         INTEGER,
    payee_family_id       INTEGER,
    verifier_user_id      INTEGER,
    verifier_team_id      INTEGER,
    verifier_family_id    INTEGER,
    category              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    step                  INTEGER DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看环境关联表
CREATE TABLE see_seek_environments (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    environment_id        INTEGER REFERENCES environments(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看隐患关联表
CREATE TABLE see_seek_hazards (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    hazard_id             INTEGER REFERENCES hazards(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看风险关联表
CREATE TABLE see_seek_risks (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    risk_id               INTEGER REFERENCES risks(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看-看表
CREATE TABLE see_seek_looks (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    outline               TEXT,
    is_deform             BOOLEAN DEFAULT false,
    skin                  TEXT,
    is_graze              BOOLEAN DEFAULT false,
    color                 VARCHAR(255),
    is_change             BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看-听表
CREATE TABLE see_seek_listens (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    sound                 TEXT,
    is_abnormal           BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看-闻表
CREATE TABLE see_seek_smells (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    odour                 TEXT,
    is_foul_odour         BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看-摸表
CREATE TABLE see_seek_touches (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    temperature           VARCHAR(255),
    is_fever              BOOLEAN DEFAULT false,
    stretch               VARCHAR(255),
    is_stiff              BOOLEAN DEFAULT false,
    shake                 VARCHAR(255),
    is_shake              BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看检查报告表
CREATE TABLE see_seek_examination_reports (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 1,
    status                INTEGER DEFAULT 0,
    name                  VARCHAR(255),
    nickname              VARCHAR(255),
    description           TEXT,
    sample_type           VARCHAR(255),
    sample_order          VARCHAR(255),
    instrument_goods_id   INTEGER,
    report_title          VARCHAR(255),
    report_content        TEXT,
    master_user_id        INTEGER,
    reviewer_user_id      INTEGER,
    report_date           TIMESTAMP,
    attachment            TEXT,
    tags                  VARCHAR(255),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看检查项目表
CREATE TABLE see_seek_examination_items (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    classify              INTEGER DEFAULT 1,
    see_seek_examination_report_id INTEGER REFERENCES see_seek_examination_reports(id),
    item_code             VARCHAR(50),
    item_name             VARCHAR(255) NOT NULL,
    result                TEXT,
    result_unit           VARCHAR(50),
    reference_min         DECIMAL(10,4),
    reference_max         DECIMAL(10,4),
    remark                TEXT,
    abnormal_flag         BOOLEAN DEFAULT false,
    method                VARCHAR(255),
    operator              VARCHAR(255),
    status                INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 看看凭据关联表
CREATE TABLE see_seek_look_evidences (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) NOT NULL UNIQUE,
    see_seek_id           INTEGER NOT NULL REFERENCES see_seeks(id),
    evidence_id           INTEGER NOT NULL REFERENCES evidences(id),
    note                  TEXT,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- ============================================
-- 其他功能表
-- ============================================

-- 脑火(头脑风暴)表
CREATE TABLE brain_fires (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    project_id            INTEGER NOT NULL,
    start_time            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    environment_id        INTEGER,
    title                 VARCHAR(255) NOT NULL,
    inference             TEXT,
    diagnose              TEXT,
    judgement             TEXT,
    payer_user_id         INTEGER,
    payer_team_id         INTEGER,
    payer_family_id       INTEGER,
    payee_user_id         INTEGER,
    payee_team_id         INTEGER,
    payee_family_id       INTEGER,
    verifier_user_id      INTEGER,
    verifier_family_id    INTEGER,
    verifier_team_id      INTEGER,
    status                INTEGER DEFAULT 1,
    brain_fire_class      INTEGER DEFAULT 1,
    brain_fire_type       INTEGER DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 建议表
CREATE TABLE suggestions (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER NOT NULL,
    project_id            INTEGER NOT NULL,
    resolution            BOOLEAN NOT NULL DEFAULT false,
    body                  TEXT NOT NULL,
    category              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 足迹表
CREATE TABLE footprints (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER,
    team_id               INTEGER,
    team_name             VARCHAR(255),
    team_type             SMALLINT,
    content               TEXT,
    content_id            INTEGER,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 警示词表
CREATE TABLE watchwords (
    id                    SERIAL PRIMARY KEY,
    word                  VARCHAR(255) NOT NULL,
    administrator_id      INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 索引创建
-- ============================================

-- 技能相关索引
CREATE INDEX idx_skills_category ON skills(category);
CREATE INDEX idx_skills_strength_level ON skills(strength_level);
CREATE INDEX idx_skills_difficulty_level ON skills(difficulty_level);
CREATE INDEX idx_skills_level ON skills(level);
CREATE INDEX idx_skills_deleted_at ON skills(deleted_at);

CREATE INDEX idx_skill_users_user_id ON skill_users(user_id);
CREATE INDEX idx_skill_users_skill_id ON skill_users(skill_id);
CREATE INDEX idx_skill_users_status ON skill_users(status);
CREATE INDEX idx_skill_users_deleted_at ON skill_users(deleted_at);

CREATE INDEX idx_skill_teams_team_id ON skill_teams(team_id);
CREATE INDEX idx_skill_teams_skill_id ON skill_teams(skill_id);
CREATE INDEX idx_skill_teams_status ON skill_teams(status);
CREATE INDEX idx_skill_teams_deleted_at ON skill_teams(deleted_at);

-- 法力相关索引
CREATE INDEX idx_magics_category ON magics(category);
CREATE INDEX idx_magics_intelligence_level ON magics(intelligence_level);
CREATE INDEX idx_magics_difficulty_level ON magics(difficulty_level);
CREATE INDEX idx_magics_level ON magics(level);
CREATE INDEX idx_magics_deleted_at ON magics(deleted_at);

CREATE INDEX idx_magic_users_user_id ON magic_users(user_id);
CREATE INDEX idx_magic_users_magic_id ON magic_users(magic_id);
CREATE INDEX idx_magic_users_status ON magic_users(status);
CREATE INDEX idx_magic_users_deleted_at ON magic_users(deleted_at);

CREATE INDEX idx_magic_teams_team_id ON magic_teams(team_id);
CREATE INDEX idx_magic_teams_magic_id ON magic_teams(magic_id);
CREATE INDEX idx_magic_teams_status ON magic_teams(status);
CREATE INDEX idx_magic_teams_deleted_at ON magic_teams(deleted_at);

-- 手工艺相关索引
CREATE INDEX idx_handicrafts_project_id ON handicrafts(project_id);
CREATE INDEX idx_handicrafts_recorder_user_id ON handicrafts(recorder_user_id);
CREATE INDEX idx_handicrafts_type ON handicrafts(type);
CREATE INDEX idx_handicrafts_category ON handicrafts(category);
CREATE INDEX idx_handicrafts_status ON handicrafts(status);
CREATE INDEX idx_handicrafts_deleted_at ON handicrafts(deleted_at);

CREATE INDEX idx_handicraft_contributors_handicraft_id ON handicraft_contributors(handicraft_id);
CREATE INDEX idx_handicraft_contributors_user_id ON handicraft_contributors(user_id);

CREATE INDEX idx_handicraft_skills_handicraft_id ON handicraft_skills(handicraft_id);
CREATE INDEX idx_handicraft_skills_skill_id ON handicraft_skills(skill_id);

CREATE INDEX idx_handicraft_magics_handicraft_id ON handicraft_magics(handicraft_id);
CREATE INDEX idx_handicraft_magics_magic_id ON handicraft_magics(magic_id);

CREATE INDEX idx_inaugurations_handicraft_id ON inaugurations(handicraft_id);
CREATE INDEX idx_inaugurations_recorder_user_id ON inaugurations(recorder_user_id);

CREATE INDEX idx_process_records_handicraft_id ON process_records(handicraft_id);
CREATE INDEX idx_process_records_recorder_user_id ON process_records(recorder_user_id);
CREATE INDEX idx_process_records_deleted_at ON process_records(deleted_at);

CREATE INDEX idx_endings_handicraft_id ON endings(handicraft_id);
CREATE INDEX idx_endings_recorder_user_id ON endings(recorder_user_id);

-- 凭据相关索引
CREATE INDEX idx_evidences_uuid ON evidences(uuid);
CREATE INDEX idx_evidences_recorder_user_id ON evidences(recorder_user_id);
CREATE INDEX idx_evidences_category ON evidences(category);
CREATE INDEX idx_evidences_visibility ON evidences(visibility);
CREATE INDEX idx_evidences_file_hash ON evidences(file_hash);
CREATE INDEX idx_evidences_deleted_at ON evidences(deleted_at);
CREATE INDEX idx_evidences_created_at ON evidences(created_at);

CREATE INDEX idx_handicraft_evidences_handicraft_id ON handicraft_evidences(handicraft_id);
CREATE INDEX idx_handicraft_evidences_evidence_id ON handicraft_evidences(evidence_id);
CREATE INDEX idx_handicraft_evidences_deleted_at ON handicraft_evidences(deleted_at);

CREATE INDEX idx_see_seek_look_evidences_see_seek_id ON see_seek_look_evidences(see_seek_id);
CREATE INDEX idx_see_seek_look_evidences_evidence_id ON see_seek_look_evidences(evidence_id);
CREATE INDEX idx_see_seek_look_evidences_deleted_at ON see_seek_look_evidences(deleted_at);

-- 物资相关索引
CREATE INDEX idx_goods_projects_project_id ON goods_projects(project_id);
CREATE INDEX idx_goods_projects_goods_id ON goods_projects(goods_id);
CREATE INDEX idx_goods_projects_provider_type ON goods_projects(provider_type);
CREATE INDEX idx_goods_projects_deleted_at ON goods_projects(deleted_at);

CREATE INDEX idx_goods_project_readiness_project_id ON goods_project_readiness(project_id);
CREATE INDEX idx_goods_project_readiness_user_id ON goods_project_readiness(user_id);

-- 安全相关索引
CREATE INDEX idx_safety_measures_hazard_id ON safety_measures(hazard_id);
CREATE INDEX idx_safety_measures_status ON safety_measures(status);

CREATE INDEX idx_safety_protections_risk_id ON safety_protections(risk_id);
CREATE INDEX idx_safety_protections_status ON safety_protections(status);

CREATE INDEX idx_hazards_severity ON hazards(severity);
CREATE INDEX idx_hazards_category ON hazards(category);

-- 看看相关索引
CREATE INDEX idx_see_seeks_project_id ON see_seeks(project_id);
CREATE INDEX idx_see_seeks_status ON see_seeks(status);

CREATE INDEX idx_see_seek_environments_see_seek_id ON see_seek_environments(see_seek_id);
CREATE INDEX idx_see_seek_hazards_see_seek_id ON see_seek_hazards(see_seek_id);
CREATE INDEX idx_see_seek_risks_see_seek_id ON see_seek_risks(see_seek_id);

CREATE INDEX idx_see_seek_looks_see_seek_id ON see_seek_looks(see_seek_id);
CREATE INDEX idx_see_seek_listens_see_seek_id ON see_seek_listens(see_seek_id);
CREATE INDEX idx_see_seek_smells_see_seek_id ON see_seek_smells(see_seek_id);
CREATE INDEX idx_see_seek_touches_see_seek_id ON see_seek_touches(see_seek_id);

CREATE INDEX idx_see_seek_examination_reports_see_seek_id ON see_seek_examination_reports(see_seek_id);
CREATE INDEX idx_see_seek_examination_items_report_id ON see_seek_examination_items(see_seek_examination_report_id);

-- 其他索引
CREATE INDEX idx_brain_fires_project_id ON brain_fires(project_id);
CREATE INDEX idx_brain_fires_status ON brain_fires(status);
CREATE INDEX idx_brain_fires_start_time ON brain_fires(start_time);

CREATE INDEX idx_suggestions_project_id ON suggestions(project_id);
CREATE INDEX idx_suggestions_user_id ON suggestions(user_id);
CREATE INDEX idx_suggestions_status ON suggestions(status);

-- ============================================
-- 表注释
-- ============================================

COMMENT ON TABLE skills IS '技能表';
COMMENT ON TABLE skill_users IS '用户技能记录表';
COMMENT ON TABLE skill_teams IS '团队技能记录表';
COMMENT ON TABLE magics IS '法力表';
COMMENT ON TABLE magic_users IS '用户法力记录表';
COMMENT ON TABLE magic_teams IS '团队法力记录表';
COMMENT ON TABLE handicrafts IS '手艺表';
COMMENT ON TABLE handicraft_contributors IS '手艺协助者表';
COMMENT ON TABLE inaugurations IS '开工仪式表';
COMMENT ON TABLE process_records IS '过程记录表';
COMMENT ON TABLE endings IS '结束仪式表';
COMMENT ON TABLE evidences IS '凭据表';
COMMENT ON TABLE handicraft_evidences IS '手工艺凭据关联表';
COMMENT ON TABLE see_seek_look_evidences IS '"看看"凭据关联表';
COMMENT ON TABLE goods_projects IS '项目物资表';
COMMENT ON TABLE goods_project_readiness IS '项目物资准备状态表';
COMMENT ON TABLE hazards IS '隐患表';
COMMENT ON TABLE risks IS '风险表';
COMMENT ON TABLE safety_measures IS '安全措施表';
COMMENT ON TABLE safety_protections IS '安全防护表';
COMMENT ON TABLE see_seeks IS '看看检查表';
COMMENT ON TABLE brain_fires IS '脑火(头脑风暴)表';
COMMENT ON TABLE suggestions IS '建议表';
