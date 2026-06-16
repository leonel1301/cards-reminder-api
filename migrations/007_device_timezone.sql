ALTER TABLE device_tokens
    ADD COLUMN IF NOT EXISTS timezone TEXT NOT NULL DEFAULT 'America/Lima';
