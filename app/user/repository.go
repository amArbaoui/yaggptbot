package user

import (
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/storage"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	sqlMaxUserId     = "select coalesce(max(user_id), 0) from user"
	sqlGetUsers      = "select user_id, tg_user_id, tg_chat_id, tg_username, created_at, updated_at from user"
	sqlGetUserByTgId = "select user_id, tg_user_id, tg_chat_id, tg_username, created_at, updated_at from user where tg_user_id = $1"
	sqlSaveUser      = "insert into user (user_id, tg_user_id, tg_chat_id, tg_username, created_at, updated_at) values (:user_id, :tg_user_id, :tg_chat_id, :tg_username, :created_at, :updated_at)"
	sqlUpdateUser    = "update user set tg_username = $1, tg_chat_id = $2, updated_at = unixepoch() where tg_user_id = $3"
	sqlDeleteUser    = "delete from user where tg_user_id = $1"
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
	err := rep.db.Get(&user, sqlGetUserByTgId, tgId)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil

}

func (rep *UserDbRepository) SaveUser(user *models.User) error {
	var maxUserId int64
	var newUser storage.User
	err := rep.db.Get(&maxUserId, sqlMaxUserId)
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

func (rep *UserDbRepository) UpdateUser(user *models.User) error {
	_, err := rep.db.Exec(sqlUpdateUser, user.TgName, user.ChatId, user.Id)
	return err

}

func (rep *UserDbRepository) DeleteUser(tgId int64) error {
	_, err := rep.db.Exec(sqlDeleteUser, tgId)
	return err
}
