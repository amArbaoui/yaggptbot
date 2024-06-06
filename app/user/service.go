package user

import (
	"amArbaoui/yaggptbot/app/models"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

type UserServiceImpl struct {
	rep UserRepository
}

func NewUserService(db *sqlx.DB) *UserServiceImpl {
	repository := UserDbRepository{db: db}
	return &UserServiceImpl{rep: &repository}
}

func (us *UserServiceImpl) ValidateTgUser(tgUser *tgbotapi.User) error {
	if tgUser.IsBot {
		return errors.New("bots are restricted to use this service")
	}
	_, err := us.rep.GetUserByTgId(tgUser.ID)
	return err
}

func (us *UserServiceImpl) GetUserByTgId(tgId int64) (models.User, error) {
	return us.rep.GetUserByTgId(tgId)
}

func (us *UserServiceImpl) SaveUser(user models.User) error {
	if _, err := us.GetUserByTgId(user.Id); err == nil {
		return errors.New("user already exists")
	}
	return us.rep.SaveUser(user)
}
