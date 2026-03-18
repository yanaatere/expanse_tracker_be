-- Add parent_id to categories for sub-category support (self-referential)
ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS parent_id INT REFERENCES categories(id) ON DELETE CASCADE;

-- Index for fast sub-category lookups
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);

-- Down
-- ALTER TABLE categories DROP COLUMN IF EXISTS parent_id;
