package report

import (
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	sqlUsersWithLastMessages = "select u.tg_username, max(m.created_at) as last_message from message m inner join user u on m.chat_id  = u.tg_chat_id group by u.tg_username order by u.tg_username COLLATE NOCASE"
	sqlUsersMessagesAfter    = "select u.tg_username, count(1) as messages from message m inner join user u on m.chat_id  = u.tg_chat_id where role = 'user' and m.created_at  > :created_after group by u.tg_username  order by count(1);"
)

type UserLastMessage struct {
	TgName      string `db:"tg_username"`
	LastMessage int64  `db:"last_message"`
}

type UserMessageStat struct {
	TgName    string `db:"tg_username"`
	MessageQt int64  `db:"messages"`
}

type DbReportRepository struct {
	db *sqlx.DB
}

func NewDbReportRepository(db *sqlx.DB) DbReportRepository {
	return DbReportRepository{db: db}
}

func (rep *DbReportRepository) GetUserLastMessage() ([]*UserLastMessage, error) {
	var usersLastMessages = []*UserLastMessage{}
	err := rep.db.Select(&usersLastMessages, sqlUsersWithLastMessages)
	if err != nil {
		return nil, err
	}
	return usersLastMessages, nil

}

func (rep *DbReportRepository) GetUserMessagesStat(after time.Time) ([]*UserMessageStat, error) {
	var usersMessagesStats = []*UserMessageStat{}
	err := rep.db.Select(&usersMessagesStats, sqlUsersMessagesAfter, after.Unix())
	if err != nil {
		return nil, err
	}
	return usersMessagesStats, nil

}
