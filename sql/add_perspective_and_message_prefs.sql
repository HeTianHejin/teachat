-- 添加视角字段和消息偏好功能

-- 1. 为families表添加perspective_user_id字段
ALTER TABLE families ADD COLUMN IF NOT EXISTS perspective_user_id INTEGER;

-- 更新现有数据：perspective_user_id默认等于author_id
UPDATE families SET perspective_user_id = author_id WHERE perspective_user_id IS NULL;

-- 添加注释
COMMENT ON COLUMN families.perspective_user_id IS '视角所属用户ID，表示这是谁眼中的家庭，默认等于author_id';

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_families_perspective_user ON families(perspective_user_id);

-- 2. 创建家庭消息偏好表
CREATE TABLE IF NOT EXISTS family_message_preferences (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER REFERENCES users(id),
    family_id             INTEGER REFERENCES families(id),
    receive_messages      BOOLEAN DEFAULT true,
    notification_type     INTEGER DEFAULT 2,
    muted_until           TIMESTAMP,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

-- 添加注释
COMMENT ON TABLE family_message_preferences IS '家庭消息偏好设置，用户可以选择接收哪些家庭成员的消息';
COMMENT ON COLUMN family_message_preferences.notification_type IS '0-关闭, 1-仅重要, 2-全部';

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_family_message_prefs_user ON family_message_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_family_message_prefs_family ON family_message_preferences(family_id);
CREATE INDEX IF NOT EXISTS idx_family_message_prefs_user_family ON family_message_preferences(user_id, family_id);
