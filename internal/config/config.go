// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
)

// Config holds all runtime knobs for the API.
type Config struct {
	// MySQL connection
	MySQLHost     string
	MySQLPort     string
	MySQLDatabase string
	MySQLUser     string
	MySQLPassword string
	// DATABASE_URL, when set, overrides the MYSQL_* fields entirely.
	DatabaseURL string
	// HTTP listen port
	Port string
}

// Load reads configuration from the environment, applying sane defaults.
func Load() (*Config, error) {
	c := &Config{
		MySQLHost:     envOr("MYSQL_HOST", "db"),
		MySQLPort:     envOr("MYSQL_PORT", "3306"),
		MySQLDatabase: envOr("MYSQL_DATABASE", "article"),
		MySQLUser:     envOr("MYSQL_USER", "root"),
		MySQLPassword: envOr("MYSQL_PASSWORD", ""),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          envOr("PORT", "8080"),
	}
	if c.Port == "" {
		return nil, fmt.Errorf("PORT must not be empty")
	}
	return c, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
