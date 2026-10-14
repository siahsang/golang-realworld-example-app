package core

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/siahsang/blog/internal/utils/databaseutils"
)

type Core struct {
	log     *slog.Logger
	session *databaseutils.SQLSession
}

func NewCore(log *slog.Logger, session *databaseutils.SQLSession) *Core {
	return &Core{
		log:     log,
		session: session,
	}
}

func IsUserNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) ||
		errors.Is(err, databaseutils.ErrNoRowsFound) ||
		errors.Is(err, NoRecordFound) {
		return true
	}
	return false
}
