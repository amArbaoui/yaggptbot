package telegram

import (
	"amArbaoui/yaggptbot/app/user"
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BaseHandler func(ctx context.Context, bot *GPTBot, update *tgbotapi.Update)
type StateHandler func(ctx context.Context, bot *GPTBot, update *tgbotapi.Update)
type Command func(ctx context.Context, bot *GPTBot, update *tgbotapi.Update)

type Dispatcher struct {
	baseHandlers map[string]BaseHandler
	stateHandler map[user.State]StateHandler
	commands     map[string]Command
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		baseHandlers: make(map[string]BaseHandler),
		stateHandler: make(map[user.State]StateHandler),
		commands:     make(map[string]Command),
	}
}

func GetUserDispatcher() *Dispatcher {
	userDispatcher := NewDispatcher()
	userDispatcher.RegisterBaseHandler("callback", CallbackHandler)
	userDispatcher.RegisterBaseHandler("document", UserDocumentHandler)
	userDispatcher.RegisterBaseHandler("photo", UserPhotoHandler)
	userDispatcher.RegisterBaseHandler("message", UserMesasgeHandler)
	userDispatcher.RegisterStateHandler(user.SETTING_PROMT, SetPromptHandler)
	userDispatcher.RegisterCommand("start", StartCommand)
	userDispatcher.RegisterCommand("promptset", SetPromptCommand)
	userDispatcher.RegisterCommand("promptreset", ResetPromtCommand)
	userDispatcher.RegisterCommand("setmodel", SetModelCommand)

	return userDispatcher

}

func (d *Dispatcher) RegisterBaseHandler(updateType string, handler BaseHandler) {
	d.baseHandlers[updateType] = handler
}

func (d *Dispatcher) RegisterStateHandler(state user.State, stateHandler StateHandler) {
	d.stateHandler[state] = stateHandler
}

func (d *Dispatcher) RegisterCommand(path string, command Command) {
	d.commands[path] = command
}

func (d *Dispatcher) HandleUpdate(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if update.CallbackQuery != nil {
		d.baseHandlers["callback"](ctx, bot, update)
		return
	}
	if update.Message != nil && update.Message.IsCommand() {
		commandText := update.Message.Command()
		cmd, ok := d.commands[commandText]
		if !ok {
			log.Printf("received unknown command, %s", commandText)
			return
		}
		cmd(ctx, bot, update)
		return
	}
	if update.Message == nil || bot.ValidateUpdate(update) != nil {
		return
	}
	if update.Message.Document != nil {
		d.baseHandlers["document"](ctx, bot, update)
		return
	}
	if update.Message.Photo != nil {
		d.baseHandlers["photo"](ctx, bot, update)
		return
	}
	state, err := bot.userService.GetUserState(ctx, update.SentFrom().ID)
	if err == nil && state != "" {
		d.stateHandler[state](ctx, bot, update)
		return
	}
	d.baseHandlers["message"](ctx, bot, update)
}

func (b *GPTBot) ValidateUpdate(update *tgbotapi.Update) error {
	ctx := context.Background()
	err := b.userService.ValidateTgUser(ctx, update.SentFrom())
	if err != nil {
		log.Printf("got message (%s) from not authenticaded user %s", update.Message.Text, update.Message.From.UserName)
		messageText := "Looks like you are not authenticated yet to use this bot. Please use /start command and wait while admin processes your request"
		_, _ = b.chatService.SendMessage(ctx, MessageOut{ChatId: update.Message.Chat.ID, RepyToId: int64(update.Message.MessageID), Text: messageText})
	}
	return err
}
