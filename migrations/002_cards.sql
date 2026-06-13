CREATE TABLE IF NOT EXISTS cards (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    last_four_digits  CHAR(4) NOT NULL,
    issuer            TEXT,
    billing_cycle_day SMALLINT NOT NULL CHECK (billing_cycle_day BETWEEN 1 AND 31),
    payment_due_day   SMALLINT NOT NULL CHECK (payment_due_day BETWEEN 1 AND 31),
    color_hex         TEXT,
    notes             TEXT,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cards_user_id ON cards (user_id);
