package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/user"
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"

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
	// tg ids of users waiting for the admin approval, to avoid duplicate requests
	pendingRegistrations *sync.Map
}

func NewGPTBot(botApi *tgbotapi.BotAPI,
	chatService ChatService,
	llmservice LlmService,
	messageService MessageService,
	userService UserService,
	botOptions BotOptions,
	config *config.Config) GPTBot {
	return GPTBot{botAPI: botApi,
		chatService:          chatService,
		llmService:           llmservice,
		msgService:           messageService,
		userService:          userService,
		botOptions:           botOptions,
		userDispatcher:       GetUserDispatcher(),
		config:               *config,
		pendingRegistrations: &sync.Map{}}
}

func (b *GPTBot) StartPolling(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	if _, err := b.botAPI.Request(tgbotapi.NewSetMyCommands(botCommands...)); err != nil {
		log.Printf("failed to set bot commands, %v", err)
	}
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

func (b *GPTBot) replyLlmError(ctx context.Context, m *tgbotapi.Message, model string, err error) {
	log.Printf("failed to get completion from %s for user %d: %v", model, m.From.ID, err)
	if ctx.Err() != nil {
		return
	}
	b.TextReplyWithContext(ctx, fmt.Sprintf(config.LlmErrorMessage, model), m)
}

// keepTyping shows the "typing…" chat action until the returned stop func is called.
// Telegram clears the action after ~5 seconds, so it is resent periodically.
func (b *GPTBot) keepTyping(ctx context.Context, chatId int64) (stop func()) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(config.TypingActionInterval)
		defer ticker.Stop()
		for {
			if _, err := b.botAPI.Request(tgbotapi.NewChatAction(chatId, tgbotapi.ChatTyping)); err != nil {
				log.Printf("failed to send typing action, %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return cancel
}

// systemPrompt returns the default prompt with the current date, extended with the user's custom prompt, if any
func (b *GPTBot) systemPrompt(ctx context.Context, tgId int64) string {
	// models don't know the current date and treat their training cutoff as "now"
	defaultPrompt := fmt.Sprintf("%s\nCurrent date: %s", b.config.DefaultPrompt, time.Now().UTC().Format("2006-01-02"))
	prompt, err := b.userService.GetUserPromptByTgId(ctx, tgId)
	if err != nil {
		if !errors.Is(err, user.ErrPromptNotFound) {
			log.Println(err)
		}
		return defaultPrompt
	}
	return fmt.Sprintf("%s\nUSER_PROMPT: %s", defaultPrompt, prompt.Prompt)
}

// answer sends the conversation to the user's model and replies with the completion
func (b *GPTBot) answer(ctx context.Context, m *tgbotapi.Message, conversation []llm.CompletionRequestMessage) {
	model := b.userModel(ctx, m.From.ID)
	if hasImages(conversation) && !config.ModelSupportsImages[model] {
		log.Printf("model %s doesn't support images, using %s", model, config.DefaultVisionModel)
		model = config.DefaultVisionModel
	}
	stopTyping := b.keepTyping(ctx, m.Chat.ID)
	opts := llm.CompletionOptions{WebTools: config.WebToolsEnabled}
	llmResp, err := b.llmService.GetCompletionMessage(ctx, conversation, b.systemPrompt(ctx, m.From.ID), model, opts)
	stopTyping()
	if err != nil {
		b.replyLlmError(ctx, m, model, err)
		return
	}
	messages, err := b.TextReplyWithContext(ctx, llmResp, m)
	if err != nil {
		log.Printf("failed to send ai response, %s", err)
		return
	}
	if err := b.msgService.SaveMessages(ctx, messages, llmResp, "assistant"); err != nil {
		log.Println(err)
	}
}

func (b *GPTBot) userModel(ctx context.Context, tgId int64) string {
	model, err := b.userService.GetUserModelByTgId(ctx, tgId)
	if err != nil {
		log.Println(err)
		return config.DefaultModel
	}
	return model.Model
}

// withImplicitReply makes a message without a reply continue the latest conversation in the chat:
// the one the bot answered within config.ConversationTimeout, unless it was reset with /new.
func (b *GPTBot) withImplicitReply(ctx context.Context, m *tgbotapi.Message) *tgbotapi.Message {
	if m.ReplyToMessage != nil {
		return m
	}
	lastMessageId, err := b.msgService.GetLastAssistantMessageId(ctx, m.Chat.ID, time.Now().Add(-config.ConversationTimeout))
	if err != nil {
		log.Printf("failed to get last message in chat %d, %v", m.Chat.ID, err)
		return m
	}
	if lastMessageId == 0 {
		return m
	}
	msg := *m
	msg.ReplyToMessage = &tgbotapi.Message{MessageID: int(lastMessageId), Chat: m.Chat}
	return &msg
}
