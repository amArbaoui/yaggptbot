package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupDB(db *gorm.DB) {
	err := db.AutoMigrate(&User{}, &Message{})
	if err != nil {
		panic("failed to migrate db")
	}
}

func GetDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	SetupDB(db)
	return db
}
