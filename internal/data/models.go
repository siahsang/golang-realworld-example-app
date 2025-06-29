package data

import (
	"database/sql"
	"log/slog"
)

type Models struct {
	Users UserModel
}

func NewModels(DB *sql.DB, log *slog.Logger) Models {
	return Models{
		Users: UserModel{
			DB:  DB,
			log: log,
		},
	}
}
