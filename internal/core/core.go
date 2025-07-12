package core

import (
	"database/sql"
	"log/slog"
)

type Core struct {
	log *slog.Logger
	db  *sql.DB
}

func NewDB(dbConn *sql.DB, log *slog.Logger) *Core {
	return &Core{
		log: log,
		db:  dbConn,
	}
}
