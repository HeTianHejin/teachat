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

-- "看看"凭据关联表
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

-- 添加索引以提高查询性能
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

-- 添加注释
COMMENT ON TABLE evidences IS '凭据表';
COMMENT ON COLUMN evidences.id IS '主键ID';
COMMENT ON COLUMN evidences.uuid IS '唯一标识符';
COMMENT ON COLUMN evidences.description IS '描述记录';
COMMENT ON COLUMN evidences.recorder_user_id IS '记录人ID';
COMMENT ON COLUMN evidences.note IS '备注说明';
COMMENT ON COLUMN evidences.category IS '分类：0-未知，1-图片，2-视频，3-音频，4-其他';
COMMENT ON COLUMN evidences.path IS '储存路径';
COMMENT ON COLUMN evidences.original_url IS '原始URL';
COMMENT ON COLUMN evidences.filename IS '文件名';
COMMENT ON COLUMN evidences.mime_type IS 'MIME类型';
COMMENT ON COLUMN evidences.file_size IS '文件大小(字节)';
COMMENT ON COLUMN evidences.file_hash IS '文件哈希值';
COMMENT ON COLUMN evidences.width IS '图片/视频宽度(像素)';
COMMENT ON COLUMN evidences.height IS '图片/视频高度(像素)';
COMMENT ON COLUMN evidences.duration IS '视频/音频时长(秒)';
COMMENT ON COLUMN evidences.visibility IS '可见性：0-公开，1-私有';
COMMENT ON COLUMN evidences.created_at IS '创建时间';
COMMENT ON COLUMN evidences.updated_at IS '更新时间';
COMMENT ON COLUMN evidences.deleted_at IS '软删除时间';

COMMENT ON TABLE handicraft_evidences IS '手工艺凭据关联表';
COMMENT ON COLUMN handicraft_evidences.id IS '主键ID';
COMMENT ON COLUMN handicraft_evidences.uuid IS '唯一标识符';
COMMENT ON COLUMN handicraft_evidences.handicraft_id IS '手工艺ID';
COMMENT ON COLUMN handicraft_evidences.evidence_id IS '凭据ID';
COMMENT ON COLUMN handicraft_evidences.note IS '备注说明';
COMMENT ON COLUMN handicraft_evidences.created_at IS '创建时间';
COMMENT ON COLUMN handicraft_evidences.updated_at IS '更新时间';
COMMENT ON COLUMN handicraft_evidences.deleted_at IS '软删除时间';

COMMENT ON TABLE see_seek_look_evidences IS '"看看"凭据关联表';
COMMENT ON COLUMN see_seek_look_evidences.id IS '主键ID';
COMMENT ON COLUMN see_seek_look_evidences.uuid IS '唯一标识符';
COMMENT ON COLUMN see_seek_look_evidences.see_seek_id IS '"看看"ID';
COMMENT ON COLUMN see_seek_look_evidences.evidence_id IS '凭据ID';
COMMENT ON COLUMN see_seek_look_evidences.note IS '备注说明';
COMMENT ON COLUMN see_seek_look_evidences.created_at IS '创建时间';
COMMENT ON COLUMN see_seek_look_evidences.updated_at IS '更新时间';
COMMENT ON COLUMN see_seek_look_evidences.deleted_at IS '软删除时间';
