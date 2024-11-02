package user

import (
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

func (us *UserServiceImpl) GetUserByTgId(tgId int64) (*User, error) {
	entity, err := us.rep.GetUserByTgId(tgId)
	if err != nil {
		return nil, err
	}
	return NewUserFromDbEntity(entity), nil
}

func (us *UserServiceImpl) GetUsers() ([]User, error) {
	var users []User
	entities, err := us.rep.GetUsers()
	if err != nil {
		return users, err
	}
	for _, elem := range entities {
		users = append(users, *NewUserFromDbEntity(&elem))
	}
	return users, nil
}

func (us *UserServiceImpl) GetUsersDetails() ([]UserDetails, error) {
	var usersDetails []UserDetails
	entities, err := us.rep.GetUsers()
	if err != nil {
		return usersDetails, err
	}
	for _, elem := range entities {
		usersDetails = append(usersDetails, *NewUserDetails(&elem))
	}
	return usersDetails, nil
}

func (us *UserServiceImpl) UpdateUser(user *User) error {
	return us.rep.UpdateUser(user)
}

func (us *UserServiceImpl) SaveUser(user *User) error {
	if _, err := us.GetUserByTgId(user.Id); err == nil {
		return errors.New("user already exists")
	}
	return us.rep.SaveUser(user)
}

func (us *UserServiceImpl) ValidateTgUser(tgUser *tgbotapi.User) error {
	if tgUser.IsBot {
		return ErrBotsNotAllowed
	}
	_, err := us.rep.GetUserByTgId(tgUser.ID)
	return err
}

func (us *UserServiceImpl) GetUserPromptByTgId(tgId int64) (*UserPrompt, error) {
	user, err := us.rep.GetUserByTgId(tgId)
	if err != nil {
		return nil, err
	}
	prompt, err := us.rep.GetUserPrompt(user.ID)
	if err != nil {
		return nil, ErrPromptNotFound
	}
	return &UserPrompt{UserID: prompt.UserID, Prompt: *prompt.Prompt}, nil
}

func (us *UserServiceImpl) SetUserPrompt(prompt *UserPrompt) error {
	return us.rep.SetUserPrompt(prompt)

}

func (us *UserServiceImpl) RemoveUserPromt(tgId int64) error {
	user, err := us.rep.GetUserByTgId(tgId)
	if err != nil {
		return err
	}
	return us.rep.RemoveUserPromt(user.ID)

}

func (us *UserServiceImpl) GetUserState(tgId int64) (State, error) {
	user, err := us.rep.GetUserByTgId(tgId)
	if err != nil {
		return "", err
	}
	return us.rep.GetUserState(user.ID)

}

func (us *UserServiceImpl) SetUserState(tgId int64, state State) error {
	user, err := us.rep.GetUserByTgId(tgId)
	if err != nil {
		return err
	}
	return us.rep.SetUserState(user.ID, state)

}

func (us *UserServiceImpl) ResetUserState(tgId int64) error {
	user, err := us.rep.GetUserByTgId(tgId)
	if err != nil {
		return err
	}
	return us.rep.ResetUserState(user.ID)

}
