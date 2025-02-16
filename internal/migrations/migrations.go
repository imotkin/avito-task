package migrations

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Up(db *sql.DB) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	err = goose.Up(db, "migrations")
	if err != nil {
		return err
	}

	return nil
}

func Down(db *sql.DB) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	err = goose.Down(db, "migrations")
	if err != nil {
		return err
	}

	return nil
}
