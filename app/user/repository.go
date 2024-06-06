package user

import (
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/storage"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	sqlGetUser  = "select user_id, tg_user_id, tg_chat_id, tg_username, created_at, updated_at from user where tg_user_id = $1"
	sqlSaveUSer = "insert into user (user_id, tg_user_id, tg_chat_id, tg_username, created_at, updated_at) values (:user_id, :tg_user_id, :tg_chat_id, :tg_username, :created_at, :updated_at)"
)

type UserDbRepository struct {
	db *sqlx.DB
}

func (rep *UserDbRepository) GetUserByTgId(tgId int64) (models.User, error) {
	var user storage.User
	err := rep.db.Get(&user, sqlGetUser, tgId)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}
	return models.User{Id: int64(user.ID), TgName: user.TgUsername, ChatId: user.ChatId}, nil

}

func (rep *UserDbRepository) SaveUser(user models.User) error {
	var maxUserId int64
	var newUser storage.User

	failedToCreateErr := errors.New("failed to create user")
	err := rep.db.Get(&maxUserId, "select coalesce(max(user_id), 0) from user")
	if err != nil {
		fmt.Println(err)
		return failedToCreateErr
	}
	newUser = storage.User{ID: maxUserId + 1, TgId: user.Id, ChatId: user.ChatId, TgUsername: user.TgName, CreatedAt: time.Now().Unix(), UpdatedAt: nil}
	_, err = rep.db.NamedExec(sqlSaveUSer, &newUser)
	if err != nil {
		fmt.Println(err)
		return failedToCreateErr
	}
	return nil

}
