package telegram

import (
	"amArbaoui/yaggptbot/app/storage"
	"bytes"
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

const (
	sqlGetMessge       = "select chat_id, message_id, reply_id, message_text, role, created_at, updated_at from message where message_id = $1 and role != 'system'"
	sqlSaveMessage     = "insert into message (chat_id, message_id, reply_id, message_text, role, created_at, updated_at) values (:chat_id, :message_id, :reply_id, :message_text, :role, :created_at, :updated_at)"
	sqlGetMessageChain = `
	with recursive x as (
		select
			chat_id,
			message_id,
			reply_id,
			message_text,
			"role",
			created_at,
			updated_at,
			0 as depth
		from
			message
		where
			message_id = $1
			and role != 'system'
		union all
		select
			m.chat_id,
			m.message_id,
			m.reply_id,
			m.message_text,
			m."role",
			m.created_at,
			m.updated_at,
			x.depth + 1
		from
			message m
		join x on
			m.message_id = x.reply_id
		where
			x.reply_id is not null
			and x.reply_id > 0
			and x.depth + 1 < $2
			and m.role != 'system'
	  )
		select
			chat_id,
			message_id,
			reply_id,
			message_text,
			"role",
			created_at,
			updated_at
		from x
		order by depth desc;
	`
)

type MessageDbRepository struct {
	db             *sqlx.DB
	encryptService *storage.EncryptionService
}

func (rep *MessageDbRepository) GetMessageById(ctx context.Context, messageId int64) (*storage.Message, error) {
	var msg storage.Message
	err := rep.db.Get(&msg, sqlGetMessge, messageId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", storage.ErrMessageNotFound, err)
	}
	decoded, err := storage.DecodeStoredText(msg.Text, *rep.encryptService)
	if err != nil {
		return nil, err
	}
	msg.Text = decoded
	return &msg, nil
}

func (rep *MessageDbRepository) GetMessageChain(ctx context.Context, topMessageId int64, maxConversationDepth int) ([]*storage.Message, error) {
	messageChain := make([]*storage.Message, 0)
	err := rep.db.SelectContext(ctx, &messageChain, sqlGetMessageChain, topMessageId, maxConversationDepth)
	if err != nil {
		return messageChain, err
	}
	for _, m := range messageChain {
		decoded, err := storage.DecodeStoredText(m.Text, *rep.encryptService)
		if err != nil {
			return messageChain, err
		}
		m.Text = decoded
	}
	return messageChain, nil
}

func (rep *MessageDbRepository) SaveMessage(ctx context.Context, message *tgbotapi.Message, role string) error {
	var replyTo int64
	var newMessage storage.Message
	if reply := message.ReplyToMessage; reply != nil {
		replyTo = int64(reply.MessageID)
	}
	encodedMessageText, err := rep.encryptService.Encrypt([]byte(message.Text))
	if err != nil {
		return fmt.Errorf("%w:%v", storage.ErrMessageNotEncoded, err)
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
		return fmt.Errorf("%w:%v", storage.ErrMessageNotCreated, err)
	}
	return nil
}

func (rep *MessageDbRepository) SaveMessages(ctx context.Context, messages []*tgbotapi.Message, role string) error {
	var buffer bytes.Buffer

	for _, m := range messages {
		buffer.WriteString(m.Text)
	}
	messageToSave := messages[0]
	messageToSave.Text = buffer.String()
	return rep.SaveMessage(ctx, messageToSave, role)
}
