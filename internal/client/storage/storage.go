package storage

import (
	"database/sql"
	"errors"

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

	err = migrate(db)
	if err != nil {
		return err
	}

	return nil
}

func newSQLDB(storagePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, err
	}

	return db, nil
}
