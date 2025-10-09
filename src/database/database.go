package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "embed"

	_ "modernc.org/sqlite"

	"github.com/cnaize/meds/src/core/logger"
)

//go:embed migrations/*
var migrations string

type Database struct {
	Q  *Queries
	DB *sql.DB

	path   string
	logger *logger.Logger
}

func NewDatabase(path string, logger *logger.Logger) *Database {
	return &Database{
		Q:      New(),
		path:   path,
		logger: logger,
	}
}

func (d *Database) Init(ctx context.Context) error {
	db, err := sql.Open("sqlite", d.path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	// TODO: implement migrations
	if _, err := db.ExecContext(ctx, migrations); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	d.DB = db

	return nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
