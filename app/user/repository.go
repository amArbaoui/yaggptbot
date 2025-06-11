package user

import (
	"amArbaoui/yaggptbot/app/storage"
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	sqlMaxUserID        = "select coalesce(max(user_id), 0) from user"
	sqlGetUsers         = "select user_id, tg_user_id, tg_chat_id, tg_username, created_at, updated_at from user"
	sqlGetUserByTgID    = "select user_id, tg_user_id, tg_chat_id, tg_username, created_at, updated_at from user where tg_user_id = $1"
	sqlSaveUser         = "insert into user (user_id, tg_user_id, tg_chat_id, tg_username, created_at, updated_at) values (:user_id, :tg_user_id, :tg_chat_id, :tg_username, :created_at, :updated_at)"
	sqlUpdateUser       = "update user set tg_username = $1, tg_chat_id = $2, updated_at = unixepoch() where tg_user_id = $3"
	sqlDeleteUser       = "delete from user where tg_user_id = $1"
	sqlGetUserPrompt    = "select  user_id, prompt, created_at, updated_at from user_prompt where user_id = $1"
	sqlSetUserPrompt    = "insert or replace into user_prompt (user_id, prompt, created_at, updated_at) values (:user_id, :prompt, :created_at, :updated_at)"
	sqlRemoveUserPrompt = "delete from user_prompt where user_id = $1"
	sqlGetUserState     = "select state from user_state where user_id = $1 "
	sqlSetUserState     = "insert into user_state (user_id, state) values ($1, $2)"
	sqlRemoveUserState  = "delete from user_state where user_id = $1"
	sqlRemoveUserModel  = "delete from user_model where user_id = $1"
	sqlSetUserModel     = "insert or replace into user_model (user_id, model, created_at, updated_at) values (:user_id, :model, :created_at, :updated_at)"
	sqlGetUserModel     = "select  user_id, model, created_at, updated_at from user_model where user_id = $1"
)

type UserDbRepository struct {
	db *sqlx.DB
}

func (rep *UserDbRepository) GetUsers(ctx context.Context) ([]storage.User, error) {
	var users = []storage.User{}
	err := rep.db.SelectContext(ctx, &users, sqlGetUsers)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrUserNotFound, err)
	}
	return users, nil
}

func (rep *UserDbRepository) GetUserByTgId(ctx context.Context, tgId int64) (*storage.User, error) {
	var user storage.User
	err := rep.db.GetContext(ctx, &user, sqlGetUserByTgID, tgId)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (rep *UserDbRepository) SaveUser(ctx context.Context, user *User) error {
	var maxUserId int64
	var newUser storage.User
	err := rep.db.GetContext(ctx, &maxUserId, sqlMaxUserID)
	if err != nil {
		return fmt.Errorf("%w:%v", ErrUserNotCreated, err)
	}
	newUser = storage.User{ID: maxUserId + 1, TgId: user.Id, ChatId: user.ChatId, TgUsername: user.TgName, CreatedAt: time.Now().Unix(), UpdatedAt: nil}
	_, err = rep.db.NamedExecContext(ctx, sqlSaveUser, &newUser)
	if err != nil {
		return fmt.Errorf("%w:%v", ErrUserNotCreated, err)
	}
	return nil
}

func (rep *UserDbRepository) UpdateUser(ctx context.Context, user *User) error {
	_, err := rep.db.ExecContext(ctx, sqlUpdateUser, user.TgName, user.ChatId, user.Id)
	return err
}

func (rep *UserDbRepository) DeleteUser(ctx context.Context, tgId int64) error {
	_, err := rep.db.ExecContext(ctx, sqlDeleteUser, tgId)
	return err
}

func (rep *UserDbRepository) GetUserPrompt(ctx context.Context, userId int64) (*storage.Prompt, error) {
	var userPrompt storage.Prompt
	err := rep.db.GetContext(ctx, &userPrompt, sqlGetUserPrompt, userId)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrPromptNotFound, err)
	}
	return &userPrompt, nil
}

func (rep *UserDbRepository) SetUserPrompt(ctx context.Context, prompt *UserPrompt) error {
	userPrompt := storage.Prompt{UserID: prompt.UserID, Prompt: &prompt.Prompt, CreatedAt: time.Now().Unix(), UpdatedAt: nil}
	tx := rep.db.MustBegin()
	tx.MustExecContext(ctx, sqlRemoveUserPrompt, prompt.UserID)
	tx.NamedExecContext(ctx, sqlSetUserPrompt, &userPrompt)
	err := tx.Commit()
	if err != nil {
		return fmt.Errorf("%w:%v", ErrPromptNotCreated, err)
	}
	return nil
}

func (rep *UserDbRepository) RemoveUserPromt(ctx context.Context, userId int64) error {
	_, err := rep.db.ExecContext(ctx, sqlRemoveUserPrompt, userId)
	return err
}

func (rep *UserDbRepository) GetUserState(ctx context.Context, userId int64) (State, error) {
	var state State
	err := rep.db.GetContext(ctx, &state, sqlGetUserState, userId, state)
	return state, err
}

func (rep *UserDbRepository) SetUserState(ctx context.Context, userId int64, state State) error {
	err := rep.ResetUserState(ctx, userId)
	if err != nil {
		return err
	}
	_, err = rep.db.ExecContext(ctx, sqlSetUserState, userId, state)
	return err
}

func (rep *UserDbRepository) ResetUserState(ctx context.Context, userId int64) error {
	_, err := rep.db.ExecContext(ctx, sqlRemoveUserState, userId)
	return err
}

func (rep *UserDbRepository) GetUserModel(ctx context.Context, userId int64) (*storage.Model, error) {
	var userModel storage.Model
	err := rep.db.GetContext(ctx, &userModel, sqlGetUserModel, userId)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrModelNotSet, err)
	}
	return &userModel, nil
}

func (rep *UserDbRepository) ResetUserModel(ctx context.Context, userId int64) error {
	_, err := rep.db.ExecContext(ctx, sqlRemoveUserModel, userId)
	return err
}

func (rep *UserDbRepository) SetUserModel(ctx context.Context, model *UserModel) error {
	userModel := storage.Model{UserID: model.UserID, Model: &model.Model, CreatedAt: time.Now().Unix(), UpdatedAt: nil}
	tx := rep.db.MustBegin()
	tx.MustExecContext(ctx, sqlRemoveUserModel, model.UserID)
	tx.NamedExecContext(ctx, sqlSetUserModel, &userModel)
	err := tx.Commit()
	if err != nil {
		return fmt.Errorf("%w:%v", ErrModelNotSet, err)
	}
	return nil
}
