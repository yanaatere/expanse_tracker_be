ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS receipt_image_url TEXT;

-- Down
-- ALTER TABLE transactions DROP COLUMN IF EXISTS receipt_image_url;
