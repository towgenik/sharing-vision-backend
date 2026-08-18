// Package db opens the MySQL pool and applies schema migrations on boot.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/towgenik/sharing-vision-backend/internal/config"
	"github.com/towgenik/sharing-vision-backend/migrations"
)

// Open connects to MySQL, verifies connectivity, and runs pending migrations.
func Open(cfg *config.Config) (*sql.DB, error) {
	var dsn string
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		mc := mysql.NewConfig()
		mc.User = cfg.MySQLUser
		mc.Passwd = cfg.MySQLPassword
		mc.Net = "tcp"
		mc.Addr = fmt.Sprintf("%s:%s", cfg.MySQLHost, cfg.MySQLPort)
		mc.DBName = cfg.MySQLDatabase
		mc.ParseTime = true
		mc.MultiStatements = true
		mc.Collation = "utf8mb4_unicode_ci"
		dsn = mc.FormatDSN()
	}

	if err := pingRetry(dsn); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(20)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(5 * time.Minute)

	if err := runMigrations(dsn); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return pool, nil
}

// pingRetry waits for MySQL to be ready (it can be slow to start in a container).
func pingRetry(dsn string) error {
	var err error
	for i := 0; i < 30; i++ {
		var probe *sql.DB
		probe, err = sql.Open("mysql", dsn)
		if err == nil {
			err = probe.Ping()
			probe.Close()
			if err == nil {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return err
}

// runMigrations applies embedded migrations/ via golang-migrate.
func runMigrations(dsn string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	mysqlURL := fmt.Sprintf("mysql://%s?multiStatements=true", dsn)
	m, err := migrate.NewWithSourceInstance("iofs", src, mysqlURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
