-- New improved data model for Sampatti

-- Users who log in via Google OAuth
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    google_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Investment plans (families/groups)
CREATE TABLE IF NOT EXISTS plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    owner_user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users (id)
);

-- Users who can access a plan
CREATE TABLE IF NOT EXISTS plan_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    invited_by_user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES plans (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (invited_by_user_id) REFERENCES users (id),
    UNIQUE(plan_id, user_id)
);

-- External investment platform credentials
CREATE TABLE IF NOT EXISTS investment_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_id INTEGER NOT NULL,
    platform TEXT NOT NULL CHECK(platform IN ('kuvera', 'stockal')),
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES plans (id)
);

-- Kuvera data fetches
CREATE TABLE IF NOT EXISTS kuvera_fetches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_account_id INTEGER NOT NULL,
    portfolio_data TEXT,
    holdings_data TEXT,
    gold_price_data TEXT,
    fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    error_message TEXT,
    FOREIGN KEY (investment_account_id) REFERENCES investment_accounts (id)
);

-- Stockal data fetches
CREATE TABLE IF NOT EXISTS stockal_fetches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_account_id INTEGER NOT NULL,
    account_summary_data TEXT,
    portfolio_data TEXT,
    fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    error_message TEXT,
    FOREIGN KEY (investment_account_id) REFERENCES investment_accounts (id)
);