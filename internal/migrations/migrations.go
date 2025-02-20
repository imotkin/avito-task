package migrations

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Up(db *sql.DB, dir ...string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	var path string

	switch len(dir) {
	case 0:
		path = "migrations"
	case 1:
		path = dir[0]
	default:
		return errors.New("too many args")
	}

	err = goose.Up(db, path)
	if err != nil {
		return err
	}

	return nil
}

func Down(db *sql.DB, dir ...string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	var path string

	switch len(dir) {
	case 0:
		path = "migrations"
	case 1:
		path = dir[0]
	default:
		return errors.New("too many args")
	}

	err = goose.Down(db, path)
	if err != nil {
		return err
	}

	return nil
}
