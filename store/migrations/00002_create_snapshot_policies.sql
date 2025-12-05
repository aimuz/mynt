-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS snapshot_policies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    schedule TEXT NOT NULL,
    retention TEXT NOT NULL,
    datasets TEXT, -- JSON array of dataset names
    enabled BOOLEAN DEFAULT 1,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS snapshot_policies;
-- +goose StatementEnd
