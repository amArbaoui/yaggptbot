package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/user"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LlmService interface {
	GetCompletionMessage(messages []llm.CompletionRequestMessage, prompt string) (string, error)
}

type MessageService interface {
	GetMessage(messageId int64) (*models.Message, error)
	GetMessageChain(topMessageId int64, maxConversationDepth int) ([]*storage.Message, error)
	SendMessage(botAPI *tgbotapi.BotAPI, SendMsgRequest models.Message) (*tgbotapi.Message, error)
	SaveMessage(message *tgbotapi.Message, role string) error
}

type UserService interface {
	DeleteUser(tgId int64) error
	GetUserByTgId(tgId int64) (*models.User, error)
	GetUsers() ([]models.User, error)
	GetUsersDetails() ([]models.UserDetails, error)
	GetUserPromptByTgId(tgId int64) (*models.UserPrompt, error)
	RemoveUserPromt(tg_id int64) error
	SetUserPrompt(prompt *models.UserPrompt) error
	UpdateUser(*models.User) error
	ValidateTgUser(tgUser *tgbotapi.User) error
	GetUserState(tgId int64) (user.State, error)
	SetUserState(tgId int64, state user.State) error
	ResetUserState(tgId int64) error
}

type MessageRepository interface {
	GetMessageById(messageId int64) (*storage.Message, error)
	GetMessageChain(topMessageId int64, maxConversationDepth int) ([]*storage.Message, error)
	SaveMessage(message *tgbotapi.Message, role string) error
}
