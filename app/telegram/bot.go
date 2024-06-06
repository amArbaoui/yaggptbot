package telegram

import (
	"context"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotOptions struct {
	MaxConversationDepth int
	BotDebugEnabled      bool
}

type GPTBot struct {
	botAPI      *tgbotapi.BotAPI
	llmService  LlmService
	msgService  MessageService
	userService UserService
	botOptions  BotOptions
}

func NewGPTBot(tgToken string,
	llmservice LlmService,
	messageService MessageService,
	userService UserService,
	botOptions BotOptions) GPTBot {
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = botOptions.BotDebugEnabled
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return GPTBot{botAPI: bot, llmService: llmservice, msgService: messageService, userService: userService, botOptions: botOptions}
}

func (b *GPTBot) StartPolling(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	userDispatcher := NewDispatcher()
	userDispatcher.RegisterHandler("command", UserComamndHandler)
	userDispatcher.RegisterHandler("document", UserDocumentHandler)
	userDispatcher.RegisterHandler("photo", UserPhotoHandler)
	userDispatcher.RegisterHandler("message", UserMesasgeHandler)

	updates := b.botAPI.GetUpdatesChan(u)
	for {
		select {
		case update := <-updates:
			userDispatcher.HandleUpdate(b, &update)
		case <-ctx.Done():
			log.Println("shutting down bot")
			return
		}
	}
}
