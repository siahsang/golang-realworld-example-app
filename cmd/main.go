package main

import (
	"context"
	"database/sql"
	"github.com/golang-cz/devslog"
	_ "github.com/lib/pq"
	"github.com/siahsang/blog/internal/data"
	"log/slog"
	"os"
	"sync"
	"time"
)

type application struct {
	models data.Models
	logger *slog.Logger
	wg     sync.WaitGroup
}

func main() {
	logger := configLogger()
	logger.Info("Starting application...")
	db, err := openDBConnection()

	if err != nil {
		logger.Error("Errors opening database connection: %v", err)
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Errors closing database connection: %v", err)
			os.Exit(1)
		}
	}()

	logger.Info("Database connection established successfully")

	app := application{
		models: data.NewModels(db, logger),
		logger: logger,
		wg:     sync.WaitGroup{},
	}

	if err := app.serve(); err != nil {
		logger.Error("ErrorStack starting server: %v", err)
		os.Exit(1)
	}
}

func configLogger() *slog.Logger {
	handler := devslog.NewHandler(
		os.Stdout, &devslog.Options{
			HandlerOptions: &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
			NewLineAfterLog: false,
		})

	logger := slog.New(handler)
	return logger
}

func openDBConnection() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost/myblog?sslmode=disable")
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
