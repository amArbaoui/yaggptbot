package telegram

import (
	"amArbaoui/yaggptbot/app/storage"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

const (
	sqlGetMessge   = "select chat_id, message_id, reply_id, message_text, role, created_at, updated_at from message where message_id = $1"
	sqlSaveMessage = "insert into message (chat_id, message_id, reply_id, message_text, role, created_at, updated_at) values (:chat_id, :message_id, :reply_id, :message_text, :role, :created_at, :updated_at)"
)

type MessageDbRepository struct {
	db             *sqlx.DB
	encryptService *storage.EncryptionService
}

func (rep *MessageDbRepository) GetMessageById(messageId int64) (*storage.Message, error) {
	var msg storage.Message
	err := rep.db.Get(&msg, sqlGetMessge, messageId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMessageNotFound, err)
	}
	decodedB64MessageText, err := storage.DecodeKeyFromString(msg.Text)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrMessageNotDecoded, err)
	}
	decodedBytes, err := rep.encryptService.Decrypt(decodedB64MessageText)
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrMessageNotDecoded, err)
	}
	msg.Text = string(decodedBytes)
	return &msg, nil
}

func (rep *MessageDbRepository) SaveMessage(message *tgbotapi.Message, role string) error {
	var replyTo int64
	var newMessage storage.Message
	if reply := message.ReplyToMessage; reply != nil {
		replyTo = int64(reply.MessageID)
	}
	encodedMessageText, err := rep.encryptService.Encrypt([]byte(message.Text))
	if err != nil {
		return fmt.Errorf("%w:%v", ErrMessageNotEncoded, err)
	}
	b64DecodedMessageText := storage.EncodeKeyToString(encodedMessageText)
	newMessage = storage.Message{
		ChatId:        message.Chat.ID,
		TgMsgId:       int64(message.MessageID),
		RepyToTgMsgId: replyTo,
		Text:          b64DecodedMessageText,
		Role:          role,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     nil,
	}
	_, err = rep.db.NamedExec(sqlSaveMessage, &newMessage)
	if err != nil {
		return fmt.Errorf("%w:%v", ErrMessageNotCreated, err)
	}
	return nil

}
