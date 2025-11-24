package user

import "amArbaoui/yaggptbot/app/storage"

type User struct {
	Id     int64
	TgName string
	ChatId int64
}

func NewUserFromDbEntity(entity *storage.User) *User {
	return &User{Id: int64(entity.ID), TgName: entity.TgUsername, ChatId: entity.ChatId}
}

type UserDetails struct {
	User
	Model     string
	CreatedAt int64
	UpdatedAt *int64
}

func NewUserDetails(entity *storage.User, model string) *UserDetails {
	usr := NewUserFromDbEntity(entity)
	return &UserDetails{User: *usr, Model: model, CreatedAt: entity.CreatedAt, UpdatedAt: entity.UpdatedAt}
}

type UserPrompt struct {
	UserID int64
	Prompt string
}

type UserModel struct {
	UserID int64
	Model  string
}
