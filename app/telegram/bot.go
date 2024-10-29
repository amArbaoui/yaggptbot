package telegram

import (
	"context"
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotOptions struct {
	MaxConversationDepth int
	BotDebugEnabled      bool
}

type GPTBot struct {
	botAPI         *tgbotapi.BotAPI
	chatService    ChatService
	llmService     LlmService
	msgService     MessageService
	userService    UserService
	botOptions     BotOptions
	userDispatcher *Dispatcher
}

func NewGPTBot(botApi *tgbotapi.BotAPI,
	chatService ChatService,
	llmservice LlmService,
	messageService MessageService,
	userService UserService,
	botOptions BotOptions) GPTBot {
	return GPTBot{botAPI: botApi,
		chatService:    chatService,
		llmService:     llmservice,
		msgService:     messageService,
		userService:    userService,
		botOptions:     botOptions,
		userDispatcher: GetUserDispatcher()}
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
	if update.Message == nil {
		fmt.Printf("ignoring update, not a message")
		return
	}
	if update.Message.Chat.IsPrivate() {
		b.userDispatcher.HandleUpdate(b, update)
	} else {
		fmt.Println("recieved group update, not implemented")

	}
}

func (b *GPTBot) TextReply(replyText string, m *tgbotapi.Message) (*tgbotapi.Message, error) {
	resp := Message{Id: m.Chat.ID,
		Text:     replyText,
		RepyToId: int64(m.MessageID),
		ChatId:   m.Chat.ID,
		Role:     "service"}
	msg, err := b.chatService.SendMessage(resp)
	if err != nil {
		log.Println(err)
	}
	return msg, err
}
