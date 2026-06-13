CREATE TABLE IF NOT EXISTS card_payments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id    UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    cycle_end  DATE NOT NULL,
    paid_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    notes      TEXT,
    UNIQUE (card_id, cycle_end)
);

CREATE INDEX IF NOT EXISTS idx_card_payments_card_id ON card_payments (card_id);
