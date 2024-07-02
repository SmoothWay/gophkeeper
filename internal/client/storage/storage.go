package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrInternalError = errors.New("internal error")
	ErrItemNotFound  = errors.New("item not found")
)

func Migrate(storagePath string) error {
	db, err := newSQLDB(storagePath)
	if err != nil {
		return err
	}

	err = migrate(db, 1)
	if err != nil {
		return err
	}

	return nil
}

func newSQLDB(storagePath string) (*sql.DB, error) {
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		file, err := os.Create(storagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", ErrInternalError)
	}

	// Check if the connection is actually established
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
