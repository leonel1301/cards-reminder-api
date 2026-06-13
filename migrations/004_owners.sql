CREATE TABLE IF NOT EXISTS owners (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    salary_day  SMALLINT CHECK (salary_day IS NULL OR salary_day BETWEEN 1 AND 31),
    is_self     BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_owners_user_id ON owners (user_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_owners_one_self_per_user
    ON owners (user_id)
    WHERE is_self = true;

ALTER TABLE cards
    ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES owners(id);

-- Backfill: create default (self) owner for each existing user
INSERT INTO owners (user_id, name, is_self)
SELECT u.id, COALESCE(u.display_name, 'Yo'), true
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM owners o WHERE o.user_id = u.id AND o.is_self = true
);

-- Assign existing cards to the user's self owner
UPDATE cards c
SET owner_id = o.id
FROM owners o
WHERE o.user_id = c.user_id
  AND o.is_self = true
  AND c.owner_id IS NULL;

ALTER TABLE cards
    ALTER COLUMN owner_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_cards_owner_id ON cards (owner_id);
