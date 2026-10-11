ALTER TABLE skills
    ADD COLUMN IF NOT EXISTS review_status INTEGER NOT NULL DEFAULT 3
        CHECK (review_status BETWEEN 1 AND 4),
    ADD COLUMN IF NOT EXISTS review_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS reviewed_by INTEGER REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_skills_review_status
    ON skills(review_status) WHERE deleted_at IS NULL;