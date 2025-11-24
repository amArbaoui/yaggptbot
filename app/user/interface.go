package user

import (
	"amArbaoui/yaggptbot/app/storage"
	"context"
)

type UserRepository interface {
	DeleteUser(ctx context.Context, tgId int64) error
	GetUsers(ctx context.Context) ([]storage.User, error)
	GetUserByTgId(ctx context.Context, tgId int64) (*storage.User, error)
	SaveUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	GetUserPrompt(ctx context.Context, userId int64) (*storage.Prompt, error)
	SetUserPrompt(ctx context.Context, prompt *UserPrompt) error
	RemoveUserPromt(ctx context.Context, userId int64) error
	GetUserState(ctx context.Context, userId int64) (State, error)
	SetUserState(ctx context.Context, userId int64, state State) error
	ResetUserState(ctx context.Context, userId int64) error
	GetAllUsersModels(ctx context.Context) ([]storage.Model, error)
	GetUserModel(ctx context.Context, userId int64) (*storage.Model, error)
	SetUserModel(ctx context.Context, model *UserModel) error
	SetDefaultModel(ctx context.Context, model string) error
}
