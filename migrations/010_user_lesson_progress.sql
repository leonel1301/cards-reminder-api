-- Lifetime Learn progress: which catalog lessons a user has marked as completed.
-- lesson_id matches the stable string IDs in the iOS LearnCatalog (e.g. basic_what_is_credit).

CREATE TABLE IF NOT EXISTS user_lesson_progress (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id   TEXT NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, lesson_id),
    CONSTRAINT user_lesson_progress_lesson_id_format
        CHECK (lesson_id ~ '^[a-z0-9_]{1,64}$')
);

CREATE INDEX IF NOT EXISTS idx_user_lesson_progress_user_id
    ON user_lesson_progress (user_id);
