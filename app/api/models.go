package api

import "amArbaoui/yaggptbot/app/models"

type NewUserRequest struct {
	TgId       float64 `json:"tg_id"`
	ChatId     float64 `json:"chat_id"`
	TgUsername string  `json:"tg_username"`
}

func NewUserFromAddUserRequest(newUserRequest NewUserRequest) models.User {
	return models.User{
		Id:     int64(newUserRequest.TgId),
		ChatId: int64(newUserRequest.ChatId),
		TgName: newUserRequest.TgUsername,
	}
}
