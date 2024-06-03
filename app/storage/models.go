package storage

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	TgId       int64 `gorm:"primaryKey"`
	ChatId     int64
	TgUsername string
}

type Message struct {
	gorm.Model
	TgMsgId       int64 `gorm:"primaryKey"`
	Text          string
	RepyToTgMsgId int64
	ChatId        int64
	UserTgId      int64
	Role          string
}
