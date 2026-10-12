package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/tgformat"
	"context"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

type MessageDbService struct {
	rep MessageRepository
}

func NewMessageDbService(db *sqlx.DB, encryptService *storage.EncryptionService) *MessageDbService {
	repository := MessageDbRepository{db: db, encryptService: encryptService}
	return &MessageDbService{rep: &repository}
}

func (ms *MessageDbService) GetMessage(ctx context.Context, chatId int64, messageId int64) (*Message, error) {
	msg, err := ms.rep.GetMessageById(ctx, chatId, messageId)
	if err != nil {
		return nil, err
	}

	return &Message{Id: msg.TgMsgId, Text: msg.Text, RepyToId: msg.RepyToTgMsgId, ChatId: msg.ChatId, Role: msg.Role}, nil
}

func (ms *MessageDbService) GetMessageChain(ctx context.Context, chatId int64, topMessageId int64, maxConversationDepth int) ([]*storage.Message, error) {
	msgChain, err := ms.rep.GetMessageChain(ctx, chatId, topMessageId, maxConversationDepth)
	if err != nil {
		return nil, err
	}

	return msgChain, nil
}

func (ms *MessageDbService) GetLastAssistantMessageId(ctx context.Context, chatId int64, after time.Time) (int64, error) {
	return ms.rep.GetLastAssistantMessageId(ctx, chatId, after)
}

func (ms *MessageDbService) ResetConversation(ctx context.Context, chatId int64, messageId int64) error {
	return ms.rep.ResetConversation(ctx, chatId, messageId)
}

func (ms *MessageDbService) SaveMessage(ctx context.Context, message *tgbotapi.Message, role string) error {
	return ms.rep.SaveMessage(ctx, message, role)
}

func (ms *MessageDbService) SaveMessages(ctx context.Context, messages []*tgbotapi.Message, text string, role string) error {
	return ms.rep.SaveMessages(ctx, messages, text, role)
}

type ChatServiceImpl struct {
	botApi *tgbotapi.BotAPI
}

func NewChatService(botApi *tgbotapi.BotAPI) *ChatServiceImpl {
	return &ChatServiceImpl{botApi: botApi}
}

func (ch *ChatServiceImpl) SendMessage(ctx context.Context, SendMsgRequest MessageOut) ([]*tgbotapi.Message, error) {
	splitedText := tgformat.Render(SendMsgRequest.Text, config.TelegramMessageLimit)
	sentMessages := make([]*tgbotapi.Message, 0, len(splitedText))
	for _, m := range splitedText {
		msg, err := ch.send(SendMsgRequest.ChatId, SendMsgRequest.RepyToId, m)
		if err != nil {
			return sentMessages, err
		}
		sentMessages = append(sentMessages, msg)
	}
	return sentMessages, nil
}

func (ch *ChatServiceImpl) send(chatId int64, replyToId int64, text string) (*tgbotapi.Message, error) {
	msgConfig := tgbotapi.NewMessage(chatId, text)
	if replyToId > 0 {
		msgConfig.ReplyToMessageID = int(replyToId)
	}
	msgConfig.ParseMode = tgbotapi.ModeHTML
	msg, err := ch.botApi.Send(msgConfig)
	if err != nil {
		log.Printf("failed to send formatted message, sending plain text: %v", err)
		msgConfig.Text = tgformat.Plain(text)
		msgConfig.ParseMode = ""
		msg, err = ch.botApi.Send(msgConfig)
		if err != nil {
			log.Println("failed to send message, error:", err)
			ch.botApi.Send(tgbotapi.NewMessage(chatId, "Error, please try again"))
		}
	}
	return &msg, err
}
