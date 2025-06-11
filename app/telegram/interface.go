package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/user"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LlmService interface {
	GetCompletionMessage(ctx context.Context, messages []llm.CompletionRequestMessage, prompt string, model string) (string, error)
}

type MessageService interface {
	GetMessage(ctx context.Context, messageId int64) (*Message, error)
	GetMessageChain(ctx context.Context, topMessageId int64, maxConversationDepth int) ([]*storage.Message, error)
	SaveMessage(ctx context.Context, message *tgbotapi.Message, role string) error
	SaveMessages(ctx context.Context, messages []*tgbotapi.Message, role string) error
}

type ChatService interface {
	SendMessage(ctx context.Context, SendMsgRequest MessageOut) ([]*tgbotapi.Message, error)
}

type UserService interface {
	DeleteUser(ctx context.Context, tgId int64) error
	GetUserByTgId(ctx context.Context, tgId int64) (*user.User, error)
	GetUsers(ctx context.Context) ([]user.User, error)
	GetUsersDetails(ctx context.Context) ([]user.UserDetails, error)
	GetUserPromptByTgId(ctx context.Context, tgId int64) (*user.UserPrompt, error)
	RemoveUserPromt(ctx context.Context, tg_id int64) error
	SetUserPrompt(ctx context.Context, prompt *user.UserPrompt) error
	UpdateUser(ctx context.Context, user *user.User) error
	ValidateTgUser(ctx context.Context, tgUser *tgbotapi.User) error
	SaveUser(ctx context.Context, user *user.User) error
	GetUserState(ctx context.Context, tgId int64) (user.State, error)
	SetUserState(ctx context.Context, tgId int64, state user.State) error
	ResetUserState(ctx context.Context, tgId int64) error
	GetUserModel(ctx context.Context, userId int64) (*user.UserModel, error)
	GetUserModelByTgId(ctx context.Context, tgId int64) (*user.UserModel, error)
	SetUserModel(ctx context.Context, model *user.UserModel) error
}

type MessageRepository interface {
	GetMessageById(ctx context.Context, messageId int64) (*storage.Message, error)
	GetMessageChain(ctx context.Context, topMessageId int64, maxConversationDepth int) ([]*storage.Message, error)
	SaveMessage(ctx context.Context, message *tgbotapi.Message, role string) error
	SaveMessages(ctx context.Context, messages []*tgbotapi.Message, role string) error
}
