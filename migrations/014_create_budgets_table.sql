CREATE TABLE budgets (
    id                   SERIAL PRIMARY KEY,
    user_id              INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id          INT,
    category_name        VARCHAR(100) NOT NULL,
    budget_limit         NUMERIC(15, 2) NOT NULL,
    period               VARCHAR(20) NOT NULL DEFAULT 'monthly'
                             CHECK (period IN ('daily', 'weekly', 'monthly')),
    title                VARCHAR(255),
    notification_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_budgets_user_category
    ON budgets(user_id, category_name);
