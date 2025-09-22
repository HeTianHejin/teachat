-- 项目物资准备状态表
CREATE TABLE goods_project_readiness (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    is_ready BOOLEAN NOT NULL DEFAULT FALSE,
    user_id INTEGER NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);

-- 添加索引
CREATE INDEX idx_goods_project_readiness_project_id ON goods_project_readiness(project_id);
CREATE INDEX idx_goods_project_readiness_user_id ON goods_project_readiness(user_id);

-- 添加注释
COMMENT ON TABLE goods_project_readiness IS '项目物资准备状态表';
COMMENT ON COLUMN goods_project_readiness.id IS '主键ID';
COMMENT ON COLUMN goods_project_readiness.project_id IS '项目ID';
COMMENT ON COLUMN goods_project_readiness.is_ready IS '是否全部物资已备齐';
COMMENT ON COLUMN goods_project_readiness.user_id IS '确认人ID';
COMMENT ON COLUMN goods_project_readiness.notes IS '备注说明';
COMMENT ON COLUMN goods_project_readiness.created_at IS '创建时间';
COMMENT ON COLUMN goods_project_readiness.updated_at IS '更新时间';