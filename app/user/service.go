package user

import (
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/storage"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByTgId(tgId int64) (models.User, error)
	SaveUser(models.User) error
}

type UserServiceImpl struct {
	rep UserRepository
}

func NewUserService(db *gorm.DB) *UserServiceImpl {
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

type UserDbRepository struct {
	db *gorm.DB
}

func (rep *UserDbRepository) GetUserByTgId(tgId int64) (models.User, error) {
	var user storage.User
	result := rep.db.First(&user, "tg_id = ?", tgId)
	if result.Error != nil {
		return models.User{}, errors.New("user not found")
	}
	return models.User{Id: int64(user.ID), TgName: user.TgUsername, ChatId: user.ChatId}, nil

}

func (rep *UserDbRepository) SaveUser(user models.User) error {
	result := rep.db.Create(&storage.User{TgId: user.Id, ChatId: user.ChatId, TgUsername: user.TgName})
	if result.Error != nil {
		return errors.New("user not created")
	}
	return nil

}
