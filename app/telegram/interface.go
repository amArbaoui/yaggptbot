package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/user"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LlmService interface {
	GetCompletionMessage(messages []llm.CompletionRequestMessage, prompt string) (string, error)
}

type MessageService interface {
	GetMessage(messageId int64) (*Message, error)
	GetMessageChain(topMessageId int64, maxConversationDepth int) ([]*storage.Message, error)
	SaveMessage(message *tgbotapi.Message, role string) error
}

type ChatService interface {
	SendMessage(SendMsgRequest MessageOut) (*tgbotapi.Message, error)
}

type UserService interface {
	DeleteUser(tgId int64) error
	GetUserByTgId(tgId int64) (*user.User, error)
	GetUsers() ([]user.User, error)
	GetUsersDetails() ([]user.UserDetails, error)
	GetUserPromptByTgId(tgId int64) (*user.UserPrompt, error)
	RemoveUserPromt(tg_id int64) error
	SetUserPrompt(prompt *user.UserPrompt) error
	UpdateUser(*user.User) error
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
