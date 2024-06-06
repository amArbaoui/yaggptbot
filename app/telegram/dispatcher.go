package telegram

import (
	"amArbaoui/yaggptbot/app/models"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler func(bot *GPTBot, update *tgbotapi.Update)

type Dispatcher struct {
	handlers map[string]BotHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]BotHandler),
	}
}

func GetUserDispatcher() *Dispatcher {
	userDispatcher := NewDispatcher()
	userDispatcher.RegisterHandler("command", UserComamndHandler)
	userDispatcher.RegisterHandler("document", UserDocumentHandler)
	userDispatcher.RegisterHandler("photo", UserPhotoHandler)
	userDispatcher.RegisterHandler("message", UserMesasgeHandler)
	return userDispatcher

}

func (d *Dispatcher) RegisterHandler(updateType string, handler BotHandler) {
	d.handlers[updateType] = handler
}

func (d *Dispatcher) HandleUpdate(bot *GPTBot, update *tgbotapi.Update) {
	if update.Message == nil || bot.ValidateUpdate(update) != nil {
		return
	}

	if update.Message.IsCommand() {
		d.handlers["command"](bot, update)
		return
	}
	if update.Message.Document != nil {
		d.handlers["document"](bot, update)
		return
	}
	if update.Message.Photo != nil {
		d.handlers["photo"](bot, update)
		return
	}
	d.handlers["message"](bot, update)
}

func (b *GPTBot) ValidateUpdate(update *tgbotapi.Update) error {
	err := b.userService.ValidateTgUser(update.SentFrom())
	if err != nil {
		fmt.Printf("Got message (%s) for not authenticaded user %s", update.Message.Text, update.Message.From.UserName)
		messageText := fmt.Sprintf(
			"Looks like you are not authenticated to use this bot. Plesae send this info to administrator:\n"+
				"```javascript\n"+
				`{"tg_id": "%d", "tg_username": "%s", "chat_id": %d}`+
				"```\n",
			update.Message.From.ID, update.Message.From.UserName, update.Message.Chat.ID)

		_, _ = b.msgService.SendMessage(b.botAPI, models.Message{ChatId: update.Message.Chat.ID, RepyToId: int64(update.Message.MessageID), Text: messageText})
	}
	return err
}
