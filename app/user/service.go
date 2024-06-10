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

func (us *UserServiceImpl) DeleteUser(tgId int64) error {
	return us.rep.DeleteUser(tgId)
}

func (us *UserServiceImpl) GetUserByTgId(tgId int64) (*models.User, error) {
	entity, err := us.rep.GetUserByTgId(tgId)
	if err != nil {
		return nil, err
	}
	return models.NewUserFromDbEntity(entity), nil
}

func (us *UserServiceImpl) GetUsers() ([]models.User, error) {
	var users []models.User
	entities, err := us.rep.GetUsers()
	if err != nil {
		return users, err
	}
	for _, elem := range entities {
		users = append(users, *models.NewUserFromDbEntity(&elem))
	}
	return users, nil
}

func (us *UserServiceImpl) GetUsersDetails() ([]models.UserDetails, error) {
	var usersDetails []models.UserDetails
	entities, err := us.rep.GetUsers()
	if err != nil {
		return usersDetails, err
	}
	for _, elem := range entities {
		usersDetails = append(usersDetails, *models.NewUserDetails(&elem))
	}
	return usersDetails, nil
}

func (us *UserServiceImpl) UpdateUser(user *models.User) error {
	return us.rep.UpdateUser(user)
}

func (us *UserServiceImpl) SaveUser(user *models.User) error {
	if _, err := us.GetUserByTgId(user.Id); err == nil {
		return errors.New("user already exists")
	}
	return us.rep.SaveUser(user)
}

func (us *UserServiceImpl) ValidateTgUser(tgUser *tgbotapi.User) error {
	if tgUser.IsBot {
		return errors.New("bots are restricted to use this service")
	}
	_, err := us.rep.GetUserByTgId(tgUser.ID)
	return err
}
