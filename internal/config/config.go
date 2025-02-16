package config

import (
	"cmp"
	"fmt"
	"os"
)

type Config struct {
	User       string
	Password   string
	Host       string
	Port       string
	Database   string
	ServerPort string
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Database,
	)
}

func Load() *Config {
	return &Config{
		User:       cmp.Or(os.Getenv("DATABASE_USER"), "postgres"),
		Password:   cmp.Or(os.Getenv("DATABASE_PASSWORD"), "postgres"),
		Host:       cmp.Or(os.Getenv("DATABASE_HOST"), "localhost"),
		Port:       cmp.Or(os.Getenv("DATABASE_PORT"), "5432"),
		Database:   cmp.Or(os.Getenv("DATABASE_NAME"), "shop"),
		ServerPort: cmp.Or(os.Getenv("SERVER_PORT"), "8080"),
	}
}
