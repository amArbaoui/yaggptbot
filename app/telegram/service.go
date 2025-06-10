package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/util"
	"log"

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

func (ms *MessageDbService) SaveMessage(message *tgbotapi.Message, role string) error {
	return ms.rep.SaveMessage(message, role)

}
func (ms *MessageDbService) SaveMessages(messages []*tgbotapi.Message, role string) error {
	return ms.rep.SaveMessages(messages, role)

}

type ChatServiceImpl struct {
	botApi *tgbotapi.BotAPI
}

func NewChatService(botApi *tgbotapi.BotAPI) *ChatServiceImpl {
	return &ChatServiceImpl{botApi: botApi}
}

func (ch *ChatServiceImpl) SendMessage(SendMsgRequest MessageOut) ([]*tgbotapi.Message, error) {
	splitedText := util.SliceString(SendMsgRequest.Text, config.TelegramMessageLimit)
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
	parseMode := tgbotapi.ModeMarkdown
	msgConfig := tgbotapi.NewMessage(chatId, text)
	if replyToId > 0 {
		msgConfig.ReplyToMessageID = int(replyToId)
	}
	msgConfig.ParseMode = parseMode
	msg, err := ch.botApi.Send(msgConfig)
	if err != nil {
		msgConfig.Text = tgbotapi.EscapeText(parseMode, text)
		log.Println("failed to send original message, sending escaped version")
		msg, err = ch.botApi.Send(msgConfig)
		if err != nil {
			log.Println("failed to send message, error:", err)
			ch.botApi.Send(tgbotapi.NewMessage(chatId, "Error, please try again"))
		}
	}
	return &msg, err
}
