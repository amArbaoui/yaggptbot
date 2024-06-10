package user

import (
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/storage"
)

type UserRepository interface {
	DeleteUser(tgId int64) error
	GetUsers() ([]storage.User, error)
	GetUserByTgId(tgId int64) (*storage.User, error)
	SaveUser(*models.User) error
	UpdateUser(*models.User) error
	GetUserPrompt(userId int64) (*storage.Prompt, error)
	SetUserPrompt(*models.UserPrompt) error
	RemoveUserPromt(userId int64) error
	GetUserState(userId int64) (State, error)
	SetUserState(userId int64, state State) error
	ResetUserState(userId int64) error
}
