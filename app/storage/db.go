package storage

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DbPath        = "db/yaggptbot.db"
	MigrationPath = "storage/migrations"
)

func GetDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", DbPath)
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	err = runMigrations(db)
	if err != nil {
		log.Fatalf("failed to apply database migrations %v", err)
	}
	return db
}

func runMigrations(db *sqlx.DB) error {
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", MigrationPath),
		"ql", driver,
	)
	if err != nil {
		return err
	}
	m.Up()
	return nil
}
