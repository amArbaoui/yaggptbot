package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/user"
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BaseHandler func(ctx context.Context, bot *GPTBot, update *tgbotapi.Update)
type StateHandler func(ctx context.Context, bot *GPTBot, update *tgbotapi.Update)
type Command func(ctx context.Context, bot *GPTBot, update *tgbotapi.Update)

type Dispatcher struct {
	baseHandlers   map[string]BaseHandler
	stateHandler   map[user.State]StateHandler
	commands       map[string]Command
	publicCommands map[string]bool
}

// botCommands are shown in the telegram commands menu
var botCommands = []tgbotapi.BotCommand{
	{Command: "setmodel", Description: "Choose the model"},
	{Command: "promptset", Description: "Set a custom prompt"},
	{Command: "promptreset", Description: "Remove the custom prompt"},
	{Command: "cancel", Description: "Cancel the current operation"},
	{Command: "help", Description: "How to use the bot"},
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		baseHandlers:   make(map[string]BaseHandler),
		stateHandler:   make(map[user.State]StateHandler),
		commands:       make(map[string]Command),
		publicCommands: make(map[string]bool),
	}
}

func GetUserDispatcher() *Dispatcher {
	userDispatcher := NewDispatcher()
	userDispatcher.RegisterBaseHandler("callback", CallbackHandler)
	userDispatcher.RegisterBaseHandler("document", UserDocumentHandler)
	userDispatcher.RegisterBaseHandler("photo", UserPhotoHandler)
	userDispatcher.RegisterBaseHandler("message", UserMesasgeHandler)
	userDispatcher.RegisterStateHandler(user.SETTING_PROMT, SetPromptHandler)
	userDispatcher.RegisterPublicCommand("start", StartCommand)
	userDispatcher.RegisterPublicCommand("help", HelpCommand)
	userDispatcher.RegisterCommand("promptset", SetPromptCommand)
	userDispatcher.RegisterCommand("promptreset", ResetPromtCommand)
	userDispatcher.RegisterCommand("setmodel", SetModelCommand)
	userDispatcher.RegisterCommand("cancel", CancelCommand)

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

// RegisterPublicCommand registers a command available to not registered users
func (d *Dispatcher) RegisterPublicCommand(path string, command Command) {
	d.RegisterCommand(path, command)
	d.publicCommands[path] = true
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
			bot.TextReplyWithContext(ctx, config.UnknownCommandMessage, update.Message)
			return
		}
		if !d.publicCommands[commandText] && bot.ValidateUpdate(update) != nil {
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
