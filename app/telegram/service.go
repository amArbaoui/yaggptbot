package telegram

import (
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/storage"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type MessageDbService struct {
	rep MessageRepository
}

func NewMessageDbService(db *gorm.DB, encryptService *storage.EncryptionService) *MessageDbService {
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

type MessageDbRepository struct {
	db             *gorm.DB
	encryptService *storage.EncryptionService
}

func (rep *MessageDbRepository) GetMessageById(messageId int64) (storage.Message, error) {
	var msg storage.Message
	decodeError := fmt.Errorf("failed to decode message text")
	result := rep.db.Find(&msg, "tg_msg_id = ?", messageId)
	if result.Error != nil {
		return storage.Message{}, errors.New("message not found")
	}
	decodedB64MessageText, err := storage.DecodeKeyFromString(msg.Text)
	if err != nil {
		return storage.Message{}, decodeError
	}
	decodedBytes, err := rep.encryptService.Decrypt(decodedB64MessageText)
	if err != nil {
		return storage.Message{}, decodeError
	}
	msg.Text = string(decodedBytes)
	return msg, nil
}

func (rep *MessageDbRepository) SaveMessage(message *tgbotapi.Message, role string) error {
	var replyTo int64
	encodeError := fmt.Errorf("failed to decode message text")
	if reply := message.ReplyToMessage; reply != nil {
		replyTo = int64(reply.MessageID)
	}
	encodedMessageText, err := rep.encryptService.Encrypt([]byte(message.Text))
	if err != nil {
		return encodeError
	}
	b64DecodedMessageText := storage.EncodeKeyToString(encodedMessageText)
	result := rep.db.Create(&storage.Message{
		TgMsgId:       int64(message.MessageID),
		Text:          b64DecodedMessageText,
		RepyToTgMsgId: replyTo,
		ChatId:        message.Chat.ID,
		UserTgId:      message.From.ID,
		Role:          role,
	})
	return result.Error

}
