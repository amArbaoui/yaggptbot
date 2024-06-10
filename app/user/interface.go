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
}
