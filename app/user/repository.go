package user

import (
	"amArbaoui/yaggptbot/app/storage"
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

func (rep *UserDbRepository) GetUsers() ([]storage.User, error) {
	var users = []storage.User{}
	err := rep.db.Select(&users, sqlGetUsers)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrUserNotFound, err)
	}
	return users, nil

}

func (rep *UserDbRepository) GetUserByTgId(tgId int64) (*storage.User, error) {
	var user storage.User
	err := rep.db.Get(&user, sqlGetUserByTgID, tgId)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil

}

func (rep *UserDbRepository) SaveUser(user *User) error {
	var maxUserId int64
	var newUser storage.User
	err := rep.db.Get(&maxUserId, sqlMaxUserID)
	if err != nil {
		return fmt.Errorf("%w:%v", ErrUserNotCreated, err)
	}
	newUser = storage.User{ID: maxUserId + 1, TgId: user.Id, ChatId: user.ChatId, TgUsername: user.TgName, CreatedAt: time.Now().Unix(), UpdatedAt: nil}
	_, err = rep.db.NamedExec(sqlSaveUser, &newUser)
	if err != nil {
		return fmt.Errorf("%w:%v", ErrUserNotCreated, err)
	}
	return nil

}

func (rep *UserDbRepository) UpdateUser(user *User) error {
	_, err := rep.db.Exec(sqlUpdateUser, user.TgName, user.ChatId, user.Id)
	return err

}

func (rep *UserDbRepository) DeleteUser(tgId int64) error {
	_, err := rep.db.Exec(sqlDeleteUser, tgId)
	return err
}

func (rep *UserDbRepository) GetUserPrompt(userId int64) (*storage.Prompt, error) {
	var userPrompt storage.Prompt
	err := rep.db.Get(&userPrompt, sqlGetUserPrompt, userId)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrPromptNotFound, err)
	}
	return &userPrompt, nil
}

func (rep *UserDbRepository) SetUserPrompt(prompt *UserPrompt) error {
	userPrompt := storage.Prompt{UserID: prompt.UserID, Prompt: &prompt.Prompt, CreatedAt: time.Now().Unix(), UpdatedAt: nil}

	_, err := rep.db.NamedExec(sqlSetUserPrompt, &userPrompt)
	if err != nil {
		return fmt.Errorf("%w:%v", ErrPromptNotCreated, err)
	}
	return nil
}

func (rep *UserDbRepository) RemoveUserPromt(userId int64) error {
	_, err := rep.db.Exec(sqlRemoveUserPrompt, userId)
	return err
}

func (rep *UserDbRepository) GetUserState(userId int64) (State, error) {
	var state State
	err := rep.db.Get(&state, sqlGetUserState, userId, state)
	return state, err

}

func (rep *UserDbRepository) SetUserState(userId int64, state State) error {
	err := rep.ResetUserState(userId)
	if err != nil {
		return err
	}
	_, err = rep.db.Exec(sqlSetUserState, userId, state)
	return err

}

func (rep *UserDbRepository) ResetUserState(userId int64) error {
	_, err := rep.db.Exec(sqlRemoveUserState, userId)
	return err

}

func (rep *UserDbRepository) GetUserModel(userId int64) (*storage.Model, error) {
	var userModel storage.Model
	err := rep.db.Get(&userModel, sqlGetUserModel, userId)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrModelNotSet, err)
	}
	return &userModel, nil
}

func (rep *UserDbRepository) ResetUserModel(userId int64) error {
	_, err := rep.db.Exec(sqlRemoveUserModel, userId)
	return err

}

func (rep *UserDbRepository) SetUserModel(model *UserModel) error {
	userModel := storage.Model{UserID: model.UserID, Model: &model.Model, CreatedAt: time.Now().Unix(), UpdatedAt: nil}
	tx := rep.db.MustBegin()
	tx.MustExec(sqlRemoveUserModel, model.UserID)
	tx.NamedExec(sqlSetUserModel, &userModel)
	err := tx.Commit()
	if err != nil {
		return fmt.Errorf("%w:%v", ErrModelNotSet, err)
	}
	return nil
}
