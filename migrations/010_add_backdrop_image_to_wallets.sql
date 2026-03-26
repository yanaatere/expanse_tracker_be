DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'wallets' AND column_name = 'backdrop_image') THEN
        ALTER TABLE wallets ADD COLUMN backdrop_image VARCHAR(500);
    END IF;
END $$;

-- Down
-- ALTER TABLE wallets DROP COLUMN IF EXISTS backdrop_image;
