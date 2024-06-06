package user

import "amArbaoui/yaggptbot/app/models"

type UserRepository interface {
	GetUserByTgId(tgId int64) (models.User, error)
	SaveUser(models.User) error
}
