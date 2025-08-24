package core

import (
	"database/sql"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"log/slog"
)

type Core struct {
	log         *slog.Logger
	db          *sql.DB
	sqlTemplate *databaseutils.SQLTemplate
}

func NewCore(dbConn *sql.DB, log *slog.Logger, sqlTemplate *databaseutils.SQLTemplate) *Core {
	return &Core{
		log:         log,
		db:          dbConn,
		sqlTemplate: sqlTemplate,
	}
}
