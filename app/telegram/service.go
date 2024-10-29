package telegram

import (
	"amArbaoui/yaggptbot/app/storage"

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

func (ms *MessageDbService) GetMessage(messageId int64) (*Message, error) {
	msg, err := ms.rep.GetMessageById(messageId)
	if err != nil {
		return nil, err
	}

	return &Message{Id: msg.TgMsgId, Text: msg.Text, RepyToId: msg.RepyToTgMsgId, ChatId: msg.ChatId, Role: msg.Role}, nil

}

func (ms *MessageDbService) GetMessageChain(topMessageId int64, maxConversationDepth int) ([]*storage.Message, error) {
	msgChain, err := ms.rep.GetMessageChain(topMessageId, maxConversationDepth)
	if err != nil {
		return nil, err
	}

	return msgChain, nil
}

func (ms *MessageDbService) SendMessage(botAPI *tgbotapi.BotAPI, SendMsgRequest Message) (*tgbotapi.Message, error) {
	parseMode := tgbotapi.ModeMarkdown
	msgConfig := tgbotapi.NewMessage(SendMsgRequest.ChatId, SendMsgRequest.Text)
	if SendMsgRequest.RepyToId > 0 {
		msgConfig.ReplyToMessageID = int(SendMsgRequest.RepyToId)
	}
	msgConfig.ParseMode = parseMode
	msg, err := botAPI.Send(msgConfig)
	if err != nil {
		botAPI.Send(tgbotapi.NewMessage(SendMsgRequest.ChatId, "Error, please try again"))
	}
	return &msg, err
}

func (ms *MessageDbService) SaveMessage(message *tgbotapi.Message, role string) error {
	return ms.rep.SaveMessage(message, role)

}

type ChatServiceImpl struct {
	botApi *tgbotapi.BotAPI
}

func NewChatService(botApi *tgbotapi.BotAPI) *ChatServiceImpl {
	return &ChatServiceImpl{botApi: botApi}
}

func (ch *ChatServiceImpl) SendMessage(SendMsgRequest Message) (*tgbotapi.Message, error) {
	parseMode := tgbotapi.ModeMarkdown
	msgConfig := tgbotapi.NewMessage(SendMsgRequest.ChatId, SendMsgRequest.Text)
	if SendMsgRequest.RepyToId > 0 {
		msgConfig.ReplyToMessageID = int(SendMsgRequest.RepyToId)
	}
	msgConfig.ParseMode = parseMode
	msg, err := ch.botApi.Send(msgConfig)
	if err != nil {
		ch.botApi.Send(tgbotapi.NewMessage(SendMsgRequest.ChatId, "Error, please try again"))
	}
	return &msg, err
}
