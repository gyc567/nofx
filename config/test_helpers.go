package config

import "database/sql"

// NewTestDatabase creates a Database instance with a provided sql.DB connection.
// This is only available in tests (or if this file is included in the build).
// Since this file is in package config, it can access private fields.
func NewTestDatabase(db *sql.DB) *Database {
	return &Database{db: db}
}
