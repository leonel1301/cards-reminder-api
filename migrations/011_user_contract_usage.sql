-- Lifetime contract AI usage during beta (hard cap enforced in the API).

CREATE TABLE IF NOT EXISTS user_contract_usage (
    user_id       UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    analyze_count INT NOT NULL DEFAULT 0 CHECK (analyze_count >= 0),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
