-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS disk_smart (
    disk_name TEXT PRIMARY KEY,
    passed BOOLEAN NOT NULL DEFAULT 1,
    temperature INTEGER DEFAULT 0,
    power_on_hours INTEGER DEFAULT 0,
    power_cycle_count INTEGER DEFAULT 0,
    reallocated_sectors INTEGER DEFAULT 0,
    pending_sectors INTEGER DEFAULT 0,
    uncorrectable_errors INTEGER DEFAULT 0,
    attributes TEXT, -- JSON array of SMART attributes
    updated_at DATETIME NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS disk_smart;
-- +goose StatementEnd
