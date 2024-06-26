package storage

import (
	"database/sql"
	"errors"

	migrate4 "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func migrate(db *sql.DB) error {

	instance, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}
	m, err := migrate4.NewWithDatabaseInstance("file://migrations", "sqlite", instance)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate4.ErrNoChange) {
		return err
	}

	return nil
}
