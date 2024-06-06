package storage

type User struct {
	ID         int64  `db:"user_id"`
	TgId       int64  `db:"tg_user_id"`
	ChatId     int64  `db:"tg_chat_id"`
	TgUsername string `db:"tg_username"`
	CreatedAt  int64  `db:"created_at"`
	UpdatedAt  *int64 `db:"updated_at"`
}

type Message struct {
	ChatId        int64  `db:"chat_id"`
	TgMsgId       int64  `db:"message_id"`
	RepyToTgMsgId int64  `db:"reply_id"`
	Text          string `db:"message_text"`
	Role          string `db:"role"`
	CreatedAt     int64  `db:"created_at"`
	UpdatedAt     *int64 `db:"updated_at"`
}

type Promt struct {
	UserID    int64   `db:"user_id"`
	Promt     *string `db:"promt"`
	CreatedAt int64   `db:"created_at"`
	UpdatedAt *int64  `db:"updated_at"`
}
