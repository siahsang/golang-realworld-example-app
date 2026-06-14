package core

import (
	"database/sql"
	"errors"
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

func IsUserNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, NoRecordFound) {
		return true
	}
	errStr := err.Error()
	return errStr == "no Rows Found" || errStr == "No record found"
}
