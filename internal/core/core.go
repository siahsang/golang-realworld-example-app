package core

import (
	"database/sql"
	"github.com/siahsang/blog/internal/utils/database"
	"log/slog"
	"time"
)

type Core struct {
	log         *slog.Logger
	db          *sql.DB
	sqlTemplate *database.SQLTemplate
}

func NewDB(dbConn *sql.DB, log *slog.Logger, timeout time.Duration) *Core {
	return &Core{
		log: log,
		db:  dbConn,
		sqlTemplate: &database.SQLTemplate{
			DB:      dbConn,
			Timeout: timeout,
		},
	}
}
