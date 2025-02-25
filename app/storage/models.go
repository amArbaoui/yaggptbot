package storage

type User struct {
	ID         int64  `db:"user_id" json:"-"`
	TgId       int64  `db:"tg_user_id" json:"tg_id"`
	ChatId     int64  `db:"tg_chat_id" json:"chat_id"`
	TgUsername string `db:"tg_username" json:"tg_username"`
	CreatedAt  int64  `db:"created_at" json:"created_at"`
	UpdatedAt  *int64 `db:"updated_at" json:"updated_at"`
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

type Prompt struct {
	UserID    int64   `db:"user_id"`
	Prompt    *string `db:"prompt"`
	CreatedAt int64   `db:"created_at"`
	UpdatedAt *int64  `db:"updated_at"`
}

type Model struct {
	UserID    int64   `db:"user_id"`
	Model     *string `db:"model"`
	CreatedAt int64   `db:"created_at"`
	UpdatedAt *int64  `db:"updated_at"`
}
