package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"context"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotOptions struct {
	MaxConversationDepth int
	BotDebugEnabled      bool
	NotificationChatId   int64
	BotAdminChatId       int64
}

type GPTBot struct {
	botAPI         *tgbotapi.BotAPI
	chatService    ChatService
	llmService     LlmService
	msgService     MessageService
	userService    UserService
	botOptions     BotOptions
	userDispatcher *Dispatcher
	config         config.Config
}

func NewGPTBot(botApi *tgbotapi.BotAPI,
	chatService ChatService,
	llmservice LlmService,
	messageService MessageService,
	userService UserService,
	botOptions BotOptions,
	config *config.Config) GPTBot {
	return GPTBot{botAPI: botApi,
		chatService:    chatService,
		llmService:     llmservice,
		msgService:     messageService,
		userService:    userService,
		botOptions:     botOptions,
		userDispatcher: GetUserDispatcher(),
		config:         *config}
}

func (b *GPTBot) StartPolling(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.botAPI.GetUpdatesChan(u)
	for {
		select {
		case update := <-updates:
			b.Handle(&update)
		case <-ctx.Done():
			log.Println("shutting down bot")
			return
		}
	}
}

func (b *GPTBot) Handle(update *tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		log.Printf("ignoring update, not a message, nor callback")
		return
	}
	b.userDispatcher.HandleUpdate(b, update)

}

func (b *GPTBot) TextReply(replyText string, m *tgbotapi.Message) ([]*tgbotapi.Message, error) {
	resp := MessageOut{
		Text:     replyText,
		RepyToId: int64(m.MessageID),
		ChatId:   m.Chat.ID}
	msg, err := b.chatService.SendMessage(resp)
	if err != nil {
		log.Println(err)
	}
	return msg, err
}
