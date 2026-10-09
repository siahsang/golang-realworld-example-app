package core

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/siahsang/blog/internal/utils/databaseutils"
)

type Core struct {
	log         *slog.Logger
	session     *databaseutils.SQLSession
	sqlTemplate *databaseutils.SQLTemplate
}

func NewCore(log *slog.Logger, session *databaseutils.SQLSession, sqlTemplate *databaseutils.SQLTemplate) *Core {
	return &Core{
		log:         log,
		session:     session,
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
