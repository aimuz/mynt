-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    state TEXT NOT NULL,
    progress INTEGER DEFAULT 0,
    metadata TEXT,
    result TEXT,
    error TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS disks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    model TEXT,
    serial TEXT NOT NULL,
    size INTEGER,
    type TEXT,
    first_seen DATETIME NOT NULL,
    last_seen DATETIME NOT NULL,
    is_attached BOOLEAN DEFAULT 1,
    UNIQUE(name, serial)
);
CREATE INDEX IF NOT EXISTS idx_disks_serial ON disks(serial);
CREATE INDEX IF NOT EXISTS idx_disks_attached ON disks(is_attached);

CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    data TEXT,
    status TEXT NOT NULL DEFAULT 'unread',
    created_at DATETIME NOT NULL,
    read_at DATETIME,
    acked_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

CREATE TABLE IF NOT EXISTS shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL,
    protocol TEXT NOT NULL,
    read_only BOOLEAN DEFAULT 0,
    browseable BOOLEAN DEFAULT 1,
    guest_ok BOOLEAN DEFAULT 0,
    valid_users TEXT,
    comment TEXT,
    share_type TEXT DEFAULT 'normal',
    created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_shares_protocol ON shares(protocol);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT,
    email TEXT,
    account_type TEXT NOT NULL DEFAULT 'virtual',
    is_admin BOOLEAN DEFAULT 0,
    is_active BOOLEAN DEFAULT 1,
    home_dir TEXT,
    shell TEXT,
    uid INTEGER,
    gid INTEGER,
    created_at DATETIME NOT NULL,
    last_login DATETIME
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_type ON users(account_type);

CREATE TABLE IF NOT EXISTS system_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS system_config;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS shares;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS disks;
DROP TABLE IF EXISTS tasks;
-- +goose StatementEnd
