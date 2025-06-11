package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"context"
	"log"
	"runtime/debug"
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
			handleCtx, cancel := context.WithCancel(ctx)
			wg.Add(1)
			go func(ctx context.Context, upd *tgbotapi.Update) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("CRITICAL: Recovered from panic in handler. Panic: %v\n", r)
						log.Printf("Trace:\n%s", debug.Stack())
					}
				}()
				defer wg.Done()
				defer cancel()
				b.Handle(ctx, upd)
			}(handleCtx, &update)
		case <-ctx.Done():
			log.Println("shutting down bot")
			return
		}
	}
}

func (b *GPTBot) Handle(ctx context.Context, update *tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		log.Printf("ignoring update, not a message, nor callback")
		return
	}
	select {
	case <-ctx.Done():
		return
	default:
	}
	b.userDispatcher.HandleUpdate(ctx, b, update)
}

func (b *GPTBot) TextReply(replyText string, m *tgbotapi.Message) ([]*tgbotapi.Message, error) {
	return b.TextReplyWithContext(context.Background(), replyText, m)
}

func (b *GPTBot) TextReplyWithContext(ctx context.Context, replyText string, m *tgbotapi.Message) ([]*tgbotapi.Message, error) {
	resp := MessageOut{
		Text:     replyText,
		RepyToId: int64(m.MessageID),
		ChatId:   m.Chat.ID}
	msg, err := b.chatService.SendMessage(ctx, resp)
	if err != nil {
		log.Println(err)
	}
	return msg, err
}
