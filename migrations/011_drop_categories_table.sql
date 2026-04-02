-- Remove FK constraints referencing categories from transactions
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_category_id_fkey;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_sub_category_id_fkey;

-- Drop the categories table (also drops its self-referential FK on parent_id)
DROP TABLE IF EXISTS categories;
