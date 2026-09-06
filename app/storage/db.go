package storage

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"

	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const (
	DbPath        = "db/yaggptbot.db"
	MigrationPath = "storage/migrations"
)

func GetDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite", DbPath)
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
	driver, err := sqlite.WithInstance(db.DB, &sqlite.Config{})
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
