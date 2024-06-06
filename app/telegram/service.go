package telegram

import (
	"amArbaoui/yaggptbot/app/models"
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

func (ms *MessageDbService) GetMessage(messageId int64) (models.Message, error) {
	msg, err := ms.rep.GetMessageById(messageId)
	if err != nil {
		return models.Message{}, err
	}

	return models.Message{Id: msg.TgMsgId, Text: msg.Text, RepyToId: msg.RepyToTgMsgId, ChatId: msg.ChatId, Role: msg.Role}, nil

}

func (ms *MessageDbService) SendMessage(botAPI *tgbotapi.BotAPI, SendMsgRequest models.Message) (tgbotapi.Message, error) {
	parseMode := tgbotapi.ModeMarkdown
	msgConfig := tgbotapi.NewMessage(SendMsgRequest.ChatId, SendMsgRequest.Text)
	if SendMsgRequest.RepyToId > 0 {
		msgConfig.ReplyToMessageID = int(SendMsgRequest.RepyToId)
	}
	msgConfig.ParseMode = parseMode
	msg, err := botAPI.Send(msgConfig)
	return msg, err
}

func (ms *MessageDbService) SaveMessage(message *tgbotapi.Message, role string) error {
	return ms.rep.SaveMessage(message, role)

}
