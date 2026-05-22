CREATE TABLE deposits (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bank_name     TEXT NOT NULL,
    amount        NUMERIC(14,2) NOT NULL CHECK (amount > 0),
    interest_rate NUMERIC(5,2)  NOT NULL CHECK (interest_rate >= 0),
    start_date    DATE NOT NULL,
    end_date      DATE NOT NULL,
    note          TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE credits (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type              TEXT NOT NULL CHECK (type IN ('consumer','card')),
    bank_name         TEXT NOT NULL,
    total_amount      NUMERIC(14,2) NOT NULL CHECK (total_amount > 0),
    remaining_balance NUMERIC(14,2) NOT NULL CHECK (remaining_balance >= 0),
    interest_rate     NUMERIC(5,2)  NOT NULL CHECK (interest_rate >= 0),
    monthly_payment   NUMERIC(12,2) NOT NULL DEFAULT 0,
    note              TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
