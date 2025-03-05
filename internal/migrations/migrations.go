package migrations

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Up(db *sql.DB, dir ...string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	path, err := parseArgs(dir...)
	if err != nil {
		return fmt.Errorf("up migrations: %w", err)
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

	path, err := parseArgs(dir...)
	if err != nil {
		return fmt.Errorf("down migrations: %w", err)
	}

	err = goose.Down(db, path)
	if err != nil {
		return err
	}

	return nil
}

func parseArgs(args ...string) (string, error) {
	switch len(args) {
	case 0:
		return "migrations", nil
	case 1:
		return args[0], nil
	default:
		return "", fmt.Errorf("invalid dir args: %d items", len(args))
	}
}
