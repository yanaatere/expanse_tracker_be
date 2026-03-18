ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS wallet_id INT REFERENCES wallets(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS sub_category_id INT REFERENCES categories(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);

-- Down
-- ALTER TABLE transactions DROP COLUMN IF EXISTS wallet_id;
-- ALTER TABLE transactions DROP COLUMN IF EXISTS sub_category_id;
