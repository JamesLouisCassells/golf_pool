CREATE TABLE tournament_config (
    year INT PRIMARY KEY,
    entry_deadline TIMESTAMPTZ,
    start_date DATE,
    end_date DATE,
    groups JSONB,
    mutt_multiplier NUMERIC NOT NULL DEFAULT 2,
    old_mutt_multiplier NUMERIC NOT NULL DEFAULT 3,
    pool_payouts JSONB,
    frl_winner TEXT,
    frl_payout INT NOT NULL DEFAULT 500000,
    active BOOLEAN NOT NULL DEFAULT false
);
