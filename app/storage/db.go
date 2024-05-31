package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupDb(db *gorm.DB) {
	db.AutoMigrate(&User{}, &Message{})

}

func GetDb() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	SetupDb(db)
	return db
}
