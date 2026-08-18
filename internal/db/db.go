// Package db opens the MySQL pool and applies schema migrations on boot.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/towgenik/sharing-vision-backend/internal/config"
)

// Open connects to MySQL, runs the Golang migration that ensures the
// `article` database exists, then opens the connection pool.
//
// Spec (Backend_Test-Sharing_Vision.md §2-2): "Create the `article` database
// using migrations in either Golang or Python (20 points)". The posts TABLE
// itself is created manually — see sql/manual_create.sql (the 80-point task).
func Open(cfg *config.Config) (*sql.DB, error) {
	// DATABASE_URL (e.g. a managed/remote DSN) already names a database and
	// bypasses the bootstrap migration.
	if cfg.DatabaseURL != "" {
		return openPool(cfg.DatabaseURL)
	}

	serverDSN := mysqlDSN(cfg, "")
	if err := pingRetry(serverDSN); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := ensureDatabase(cfg); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	pool, err := openPool(mysqlDSN(cfg, cfg.MySQLDatabase))
	if err != nil {
		return nil, err
	}
	// The posts table is a MANUAL creation step (80 pts); fail fast with a
	// clear message if it is missing so the manual task is unmissable.
	if err := requireTable(pool); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// openPool opens, pings, and tunes the *sql.DB pool.
func openPool(dsn string) (*sql.DB, error) {
	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(20)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(5 * time.Minute)
	if err := pool.Ping(); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// mysqlDSN builds a go-sql-driver DSN, optionally binding a database name
// (empty name connects at server level, required for migration).
func mysqlDSN(cfg *config.Config, dbName string) string {
	mc := mysql.NewConfig()
	mc.User = cfg.MySQLUser
	mc.Passwd = cfg.MySQLPassword
	mc.Net = "tcp"
	mc.Addr = fmt.Sprintf("%s:%s", cfg.MySQLHost, cfg.MySQLPort)
	mc.DBName = dbName
	mc.ParseTime = true
	mc.MultiStatements = true
	mc.Collation = "utf8mb4_unicode_ci"
	return mc.FormatDSN()
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

// ensureDatabase implements the Golang migration: it creates the `article`
// database if it does not already exist (idempotent).
func ensureDatabase(cfg *config.Config) error {
	server, err := sql.Open("mysql", mysqlDSN(cfg, ""))
	if err != nil {
		return err
	}
	defer server.Close()

	var n int
	if err := server.QueryRow(
		"SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?",
		cfg.MySQLDatabase).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		log.Printf("migrate: database %q already exists (no-op)", cfg.MySQLDatabase)
		return nil
	}
	log.Printf("migrate: creating database %q", cfg.MySQLDatabase)
	_, err = server.Exec(fmt.Sprintf(
		"CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.MySQLDatabase))
	return err
}

// requireTable fails fast with a pointer to the manual SQL task when the
// posts table has not been created yet.
func requireTable(pool *sql.DB) error {
	if _, err := pool.Exec("SELECT 1 FROM posts LIMIT 1"); err == nil {
		return nil
	}
	return fmt.Errorf(
		"posts table not found in database — run the manual SQL first: " +
			"mysql article < sql/manual_create.sql (Database manual task, 80 pts)")
}
