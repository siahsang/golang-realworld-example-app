package databaseutils

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

func OpenDBConnection(logger *slog.Logger, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(10)

	duration, err := time.ParseDuration("10s")
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctxWithTimeOut, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err = db.PingContext(ctxWithTimeOut)
	if err != nil {
		err := db.Close()
		return nil, err
	}
	logger.Info("Database connection established successfully")

	return db, nil
}
