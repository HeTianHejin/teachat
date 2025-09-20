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

-- 添加索引以提高查询性能
CREATE INDEX idx_goods_projects_project_id ON goods_projects(project_id);
CREATE INDEX idx_goods_projects_goods_id ON goods_projects(goods_id);
CREATE INDEX idx_goods_projects_provider_type ON goods_projects(provider_type);
CREATE INDEX idx_goods_projects_deleted_at ON goods_projects(deleted_at);

-- 添加注释
COMMENT ON TABLE goods_projects IS '项目物资表';
COMMENT ON COLUMN goods_projects.id IS '主键ID';
COMMENT ON COLUMN goods_projects.project_id IS '项目ID';
COMMENT ON COLUMN goods_projects.responsible_user_id IS '物资责任人用户ID';
COMMENT ON COLUMN goods_projects.goods_id IS '物资ID';
COMMENT ON COLUMN goods_projects.provider_type IS '物资提供方：1-出茶方，2-收茶方';
COMMENT ON COLUMN goods_projects.expected_usage IS '预期用途说明';
COMMENT ON COLUMN goods_projects.quantity IS '数量';
COMMENT ON COLUMN goods_projects.category IS '物资类别：1-工具装备，2-消耗品';
COMMENT ON COLUMN goods_projects.status IS '物资状态：0-可用，1-使用中，2-闲置，3-已报废，4-已遗失，5-已转让';
COMMENT ON COLUMN goods_projects.notes IS '备注';
COMMENT ON COLUMN goods_projects.created_at IS '创建时间';
COMMENT ON COLUMN goods_projects.updated_at IS '更新时间';
COMMENT ON COLUMN goods_projects.deleted_at IS '软删除时间';