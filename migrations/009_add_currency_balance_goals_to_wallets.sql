DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'wallets' AND column_name = 'currency') THEN
        ALTER TABLE wallets ADD COLUMN currency VARCHAR(10) NOT NULL DEFAULT 'IDR';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'wallets' AND column_name = 'balance') THEN
        ALTER TABLE wallets ADD COLUMN balance NUMERIC(15,2) NOT NULL DEFAULT 0;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'wallets' AND column_name = 'goals') THEN
        ALTER TABLE wallets ADD COLUMN goals VARCHAR(255);
    END IF;
END $$;

-- Down
-- ALTER TABLE wallets DROP COLUMN IF EXISTS goals, DROP COLUMN IF EXISTS balance, DROP COLUMN IF EXISTS currency;
