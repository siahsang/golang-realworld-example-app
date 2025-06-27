package data

import (
	"database/sql"
	"log"
)

type Models struct {
	Users UserModel
}

func NewModels(DB *sql.DB, log *log.Logger) Models {
	return Models{
		Users: UserModel{
			DB:  DB,
			log: log,
		},
	}
}
