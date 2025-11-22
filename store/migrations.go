package store

import (
	"embed"
)

// sqlMigrations is used to embed sql migrations
//
//go:embed migrations/*.sql
var sqlMigrations embed.FS
