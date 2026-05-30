package test_utils

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/siahsang/blog/internal/utils/databaseutils"
)

var testDB *sql.DB

func CreateTestDB() {
	logger := slog.Default()
	db, err := databaseutils.OpenDBConnection(logger, "postgres://postgres:postgres@localhost/?sslmode=disable")
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer db.Close()
	testDBName := "myblog_test"
	_, err = db.Exec(fmt.Sprintf(`
    			-- Terminate all connections to the test database
    			SELECT pg_terminate_backend(pg_stat_activity.pid)
    			FROM pg_stat_activity
    			WHERE pg_stat_activity.datname = '%s'
    			AND pid <> pg_backend_pid();
   `, testDBName))

	if err != nil {
		logger.Error("Failed to terminate all test database", "error", err)
		os.Exit(1)
	}

	// drop the database if not exist
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		logger.Error("Failed to drop test database", "error", err)
		os.Exit(1)
	}

	// create the test database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		logger.Error("Failed to create test database", "error", err)
		os.Exit(1)
	} else {
		logger.Info("Database %s created successfully", testDBName)
	}

}

func SetupTestDB() {
	logger := slog.Default()
	db, err := databaseutils.OpenDBConnection(logger, "postgres://postgres:postgres@localhost/myblog_test?sslmode=disable")
	if err != nil {
		logger.Error("Failed to connect to test database", "error", err)
		os.Exit(1)
	}

	testDB = db
}

func ResetTestDB() {
	logger := slog.Default()
	logger.Info("Truncating test database...")

	if testDB == nil {
		logger.Error("Test database is not initialized")
		return
	}

	// Get all table names from the public schema
	rows, err := testDB.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`)
	if err != nil {
		logger.Error("Failed to get table names", "error", err)
		os.Exit(1)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			logger.Error("Failed to scan table name", "error", err)
			os.Exit(1)
		}
		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		logger.Info("No tables found to truncate")
		return
	}

	// Disable foreign key checks and truncate all tables
	_, err = testDB.Exec("SET session_replication_role = 'replica';")
	if err != nil {
		logger.Error("Failed to disable foreign key checks", "error", err)
		os.Exit(1)
	}

	for _, table := range tables {
		_, err = testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			logger.Error("Failed to truncate table", "table", table, "error", err)
		} else {
			logger.Info("Truncated table", "table", table)
		}
	}

	// Re-enable foreign key checks
	_, err = testDB.Exec("SET session_replication_role = 'origin';")
	if err != nil {
		logger.Error("Failed to re-enable foreign key checks", "error", err)
		os.Exit(1)
	}

	logger.Info("Test database tables truncated successfully")
}

func TearDownTestDB() {
	logger := slog.Default()
	if testDB == nil {
		logger.Error("Test database is not initialized")
		return
	}

	testDB.Close()
}

// RunMigrations applies all database migrations to the specified database
func RunMigrations(db *sql.DB, migrationsPath string) error {
	logger := slog.Default()
	logger.Info("Running database migrations", "path", migrationsPath)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	logger.Info("Migrations completed successfully")
	return nil
}
