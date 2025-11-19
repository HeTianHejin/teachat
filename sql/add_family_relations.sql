-- 添加家庭关联表（用于利益回避机制）

-- 创建family_relations表
CREATE TABLE IF NOT EXISTS family_relations (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    family_id_1           INTEGER REFERENCES families(id),
    family_id_2           INTEGER REFERENCES families(id),
    relation_type         INTEGER NOT NULL,
    confirmed_by          INTEGER REFERENCES users(id),
    status                INTEGER DEFAULT 0,
    note                  TEXT,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- 添加注释
COMMENT ON TABLE family_relations IS '家庭关联表，用于识别三代以内近亲关系实现利益回避';
COMMENT ON COLUMN family_relations.relation_type IS '1-同一家庭不同视角, 2-前后家庭, 3-父母子女, 4-领养关系, 5-兄弟姐妹';
COMMENT ON COLUMN family_relations.status IS '0-单方声明, 1-双方确认, 2-已拒绝';

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_family_relations_family1 ON family_relations(family_id_1);
CREATE INDEX IF NOT EXISTS idx_family_relations_family2 ON family_relations(family_id_2);
CREATE INDEX IF NOT EXISTS idx_family_relations_status ON family_relations(status);
CREATE INDEX IF NOT EXISTS idx_family_relations_type ON family_relations(relation_type);
CREATE INDEX IF NOT EXISTS idx_family_relations_deleted ON family_relations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_family_relations_families ON family_relations(family_id_1, family_id_2, status);
