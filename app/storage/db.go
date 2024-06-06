package storage

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func GetDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "yaggptbot.db")
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
