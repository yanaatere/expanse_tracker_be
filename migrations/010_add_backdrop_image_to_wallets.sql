ALTER TABLE wallets ADD COLUMN IF NOT EXISTS backdrop_image VARCHAR(500);

-- Down
-- ALTER TABLE wallets DROP COLUMN IF EXISTS backdrop_image;
