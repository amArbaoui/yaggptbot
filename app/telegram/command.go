package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/user"
	"context"
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	fromId := m.From.ID
	tgName := m.From.UserName
	err := bot.userService.ValidateTgUser(ctx, update.SentFrom())
	switch {
	case err == nil:
		HelpCommand(ctx, bot, update)
		return
	case errors.Is(err, user.ErrBotsNotAllowed):
		return
	}
	if fromId == bot.botOptions.BotAdminChatId {
		err = bot.userService.SaveUser(ctx, &user.User{
			Id:     fromId,
			TgName: tgName,
			ChatId: fromId,
		})
		if err != nil {
			log.Printf("failed to register admin user %s, %v", tgName, err)
			return
		}
		log.Printf("registered admin user %s", tgName)
		HelpCommand(ctx, bot, update)
		return
	}
	if _, pending := bot.pendingRegistrations.LoadOrStore(fromId, struct{}{}); pending {
		bot.TextReplyWithContext(ctx, config.RegistrationPendingMessage, m)
		return
	}
	msg := tgbotapi.NewMessage(bot.botOptions.NotificationChatId, fmt.Sprintf("🧐 Register %s?", tgName))
	msg.ReplyMarkup = NewRegistrationKeyboard(
		tgName,
		fromId,
	)
	_, err = bot.botAPI.Send(msg)
	if err != nil {
		log.Println(err)
		bot.pendingRegistrations.Delete(fromId)
		bot.TextReplyWithContext(ctx, config.RegistrationFailedMessage, m)
		return
	}
	bot.TextReplyWithContext(ctx, config.RegistrationRequestedMessage, m)
}

func HelpCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	bot.TextReplyWithContext(ctx, config.HowToUseItMessage, update.Message)
}

func SetPromptCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	currentPrompt, err := bot.userService.GetUserPromptByTgId(ctx, m.From.ID)
	var respText string
	if err != nil {
		respText = config.EnterPromptMessage
	} else {
		respText = fmt.Sprintf(config.CurrentPromptMessage, currentPrompt.Prompt)
	}
	resp := MessageOut{Text: respText, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID}
	err = bot.userService.SetUserState(ctx, m.From.ID, user.SETTING_PROMT)
	if err != nil {
		log.Printf("failed to set state %v", err)
		return
	}
	bot.chatService.SendMessage(ctx, resp)
}

func ResetPromtCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	respText := "Prompt removed"
	resp := MessageOut{Text: respText, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID}
	err := bot.userService.RemoveUserPromt(ctx, m.From.ID)
	if err != nil {
		log.Printf("failed to remove prompt %v", err)
		return
	}
	bot.userService.ResetUserState(ctx, m.From.ID)
	bot.chatService.SendMessage(ctx, resp)
}

func CancelCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	respText := config.CancelledMessage
	state, err := bot.userService.GetUserState(ctx, m.From.ID)
	if err != nil || state == "" {
		respText = config.NothingToCancelMessage
	}
	if err := bot.userService.ResetUserState(ctx, m.From.ID); err != nil {
		log.Printf("failed to reset state %v", err)
	}
	bot.TextReplyWithContext(ctx, respText, m)
}

func SetModelCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	model, err := bot.userService.GetUserModelByTgId(ctx, m.From.ID)
	if err != nil {
		log.Printf("failed to get modle for user %d, %v", m.From.ID, err)
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(config.SelectModelMessage, model.Model))
	msg.ReplyMarkup = NewModelKeyboard(update.Message.Chat.ID, model.Model)
	bot.botAPI.Send(msg)
}
