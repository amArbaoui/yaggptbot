package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LlmService interface { // TODO: set promt
	GetCompletionMessage(messages []llm.CompletionRequestMessage) (string, error)
}

type MessageService interface {
	GetMessage(messageId int64) (*models.Message, error)
	SendMessage(botAPI *tgbotapi.BotAPI, SendMsgRequest models.Message) (*tgbotapi.Message, error)
	SaveMessage(message *tgbotapi.Message, role string) error
}

type UserService interface {
	DeleteUser(tgId int64) error
	GetUserByTgId(tgId int64) (*models.User, error)
	GetUsers() ([]models.User, error)
	GetUsersDetails() ([]models.UserDetails, error)
	UpdateUser(*models.User) error
	ValidateTgUser(tgUser *tgbotapi.User) error
}

type MessageRepository interface {
	GetMessageById(messageId int64) (*storage.Message, error)
	SaveMessage(message *tgbotapi.Message, role string) error
}
