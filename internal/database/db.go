package database

import (
	"database/sql"
	"log/slog"
)

type DB struct {
	log *slog.Logger
	db  *sql.DB
}

func NewDB(dbConn *sql.DB, log *slog.Logger) *DB {
	return &DB{
		log: log,
		db:  dbConn,
	}
}
