package models

import (
	"amArbaoui/yaggptbot/app/storage"
)

type Message struct {
	Id       int64
	Text     string
	RepyToId int64
	ChatId   int64
	Role     string
}

func NewMessageFromEntity(entity storage.Message) Message {
	return Message{Id: entity.TgMsgId, Text: entity.Text, RepyToId: entity.RepyToTgMsgId, ChatId: entity.ChatId, Role: entity.Role}
}
