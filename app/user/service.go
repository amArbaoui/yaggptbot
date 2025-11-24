package user

import (
	"amArbaoui/yaggptbot/app/config"
	"context"
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

func (us *UserServiceImpl) DeleteUser(ctx context.Context, tgId int64) error {
	return us.rep.DeleteUser(ctx, tgId)
}

func (us *UserServiceImpl) GetUserByTgId(ctx context.Context, tgId int64) (*User, error) {
	entity, err := us.rep.GetUserByTgId(ctx, tgId)
	if err != nil {
		return nil, err
	}
	return NewUserFromDbEntity(entity), nil
}

func (us *UserServiceImpl) GetUsers(ctx context.Context) ([]User, error) {
	var users []User
	entities, err := us.rep.GetUsers(ctx)
	if err != nil {
		return users, err
	}
	for _, elem := range entities {
		users = append(users, *NewUserFromDbEntity(&elem))
	}
	return users, nil
}

func (us *UserServiceImpl) GetUsersDetails(ctx context.Context) ([]UserDetails, error) {
	var usersDetails []UserDetails
	entities, err := us.rep.GetUsers(ctx)
	if err != nil {
		return usersDetails, err
	}
	models, err := us.rep.GetAllUsersModels(ctx)
	if err != nil {
		return usersDetails, err
	}
	userModelMap := make(map[int64]string)
	for _, m := range models {
		userModelMap[m.UserID] = *m.Model
	}
	for _, elem := range entities {
		model, foundModel := userModelMap[elem.ID]
		if !foundModel {
			model = config.DefaultModel
		}
		usersDetails = append(usersDetails, *NewUserDetails(&elem, model))
	}
	return usersDetails, nil
}

func (us *UserServiceImpl) UpdateUser(ctx context.Context, user *User) error {
	return us.rep.UpdateUser(ctx, user)
}

func (us *UserServiceImpl) SaveUser(ctx context.Context, user *User) error {
	if _, err := us.GetUserByTgId(ctx, user.Id); err == nil {
		return errors.New("user already exists")
	}
	return us.rep.SaveUser(ctx, user)
}

func (us *UserServiceImpl) ValidateTgUser(ctx context.Context, tgUser *tgbotapi.User) error {
	if tgUser.IsBot {
		return ErrBotsNotAllowed
	}
	_, err := us.rep.GetUserByTgId(ctx, tgUser.ID)
	return err
}

func (us *UserServiceImpl) GetUserPromptByTgId(ctx context.Context, tgId int64) (*UserPrompt, error) {
	user, err := us.rep.GetUserByTgId(ctx, tgId)
	if err != nil {
		return nil, err
	}
	prompt, err := us.rep.GetUserPrompt(ctx, user.ID)
	if err != nil {
		return nil, ErrPromptNotFound
	}
	return &UserPrompt{UserID: prompt.UserID, Prompt: *prompt.Prompt}, nil
}

func (us *UserServiceImpl) SetUserPrompt(ctx context.Context, prompt *UserPrompt) error {
	return us.rep.SetUserPrompt(ctx, prompt)
}
func (us *UserServiceImpl) GetUserModel(ctx context.Context, userId int64) (*UserModel, error) {
	model, err := us.rep.GetUserModel(ctx, userId)
	if err != nil {
		return &UserModel{UserID: userId, Model: config.DefaultModel}, nil
	}
	return &UserModel{UserID: model.UserID, Model: *model.Model}, nil
}
func (us *UserServiceImpl) GetUserModelByTgId(ctx context.Context, tgId int64) (*UserModel, error) {
	user, err := us.rep.GetUserByTgId(ctx, tgId)
	if err != nil {
		return nil, err
	}

	return us.GetUserModel(ctx, user.ID)
}

func (us *UserServiceImpl) SetUserModel(ctx context.Context, model *UserModel) error {
	return us.rep.SetUserModel(ctx, model)
}

func (us *UserServiceImpl) SetDefaultModel(ctx context.Context, model string) error {
	if _, ok := config.ModelMap[model]; !ok {
		return ErrModelNotFound
	}
	return us.rep.SetDefaultModel(ctx, model)
}

func (us *UserServiceImpl) RemoveUserPromt(ctx context.Context, tgId int64) error {
	user, err := us.rep.GetUserByTgId(ctx, tgId)
	if err != nil {
		return err
	}
	return us.rep.RemoveUserPromt(ctx, user.ID)
}

func (us *UserServiceImpl) GetUserState(ctx context.Context, tgId int64) (State, error) {
	user, err := us.rep.GetUserByTgId(ctx, tgId)
	if err != nil {
		return "", err
	}
	return us.rep.GetUserState(ctx, user.ID)
}

func (us *UserServiceImpl) SetUserState(ctx context.Context, tgId int64, state State) error {
	user, err := us.rep.GetUserByTgId(ctx, tgId)
	if err != nil {
		return err
	}
	return us.rep.SetUserState(ctx, user.ID, state)
}

func (us *UserServiceImpl) ResetUserState(ctx context.Context, tgId int64) error {
	user, err := us.rep.GetUserByTgId(ctx, tgId)
	if err != nil {
		return err
	}
	return us.rep.ResetUserState(ctx, user.ID)
}
