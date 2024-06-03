package telegram

import (
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/storage"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type MessageDbService struct {
	rep MessageRepository
}

func NewMessageDbService(db *gorm.DB) *MessageDbService {
	repository := MessageDbRepository{db: db}
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
	parseMode := "MarkdownV2"
	escapedText := tgbotapi.EscapeText(parseMode, SendMsgRequest.Text)
	msgConfig := tgbotapi.NewMessage(SendMsgRequest.ChatId, escapedText) // TODO: escape markdown v2
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

type MessageDbRepository struct {
	db *gorm.DB
}

func (rep *MessageDbRepository) GetMessageById(messageId int64) (storage.Message, error) {
	var msg storage.Message
	result := rep.db.Find(&msg, "tg_msg_id = ?", messageId)
	if result.Error != nil {
		return storage.Message{}, errors.New("message not found")
	}
	return msg, nil
}

func (rep *MessageDbRepository) SaveMessage(message *tgbotapi.Message, role string) error {
	var replyTo int64
	if reply := message.ReplyToMessage; reply != nil {
		replyTo = int64(reply.MessageID)
	}

	result := rep.db.Create(&storage.Message{
		TgMsgId:       int64(message.MessageID),
		Text:          message.Text,
		RepyToTgMsgId: replyTo,
		ChatId:        message.Chat.ID,
		UserTgId:      message.From.ID,
		Role:          role,
	})
	return result.Error

}
