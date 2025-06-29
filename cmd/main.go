package main

import (
	"context"
	"database/sql"
	"github.com/siahsang/blog/internal/data"
	"log/slog"
	"os"
	"time"
)

type application struct {
	models data.Models
	logger *slog.Logger
}

func main() {
	logger := configLogger()
	logger.Info("Starting application...")
	db, err := openDBConnection()

	if err != nil {
		logger.Error("Error opening database connection: %v", err)
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Error closing database connection: %v", err)
			os.Exit(1)
		}
	}()

	logger.Info("Database connection established successfully")

	app := application{
		models: data.NewModels(db, logger),
	}

}

func configLogger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(handler)
	return logger
}

func openDBConnection() (*sql.DB, error) {
	db, err := sql.Open("postgres", "")
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(10)

	duration, err := time.ParseDuration("10s")
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(context.Background())
	if err != nil {
		return nil, err
	}

	return db, nil

}
