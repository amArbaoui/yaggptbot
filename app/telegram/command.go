package telegram

import (
	"amArbaoui/yaggptbot/app/user"
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	fromId := update.Message.From.ID
	tgName := update.Message.From.UserName
	err := bot.userService.ValidateTgUser(ctx, update.SentFrom())
	if err == nil {
		return
	}
	if update.Message.From.ID == bot.botOptions.BotAdminChatId {
		err = bot.userService.SaveUser(ctx, &user.User{
			Id:     fromId,
			TgName: tgName,
			ChatId: fromId,
		})
		if err != nil {
			return
		}
		log.Printf("registered admin user %s", tgName)
	}
	msg := tgbotapi.NewMessage(bot.botOptions.NotificationChatId, fmt.Sprintf("🧐 Register %s?", tgName))
	msg.ReplyMarkup = NewRegistrationKeyboard(
		tgName,
		fromId,
	)
	_, err = bot.botAPI.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func SetPromptCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	currentPrompt, err := bot.userService.GetUserPromptByTgId(ctx, m.From.ID)
	var respText string
	if err != nil {
		respText = "Please enter new prompt"
	} else {
		respText = fmt.Sprintf("Current prompt is: %s\nPlease set new prompt", currentPrompt.Prompt)
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

func SetModelCommand(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	model, err := bot.userService.GetUserModelByTgId(ctx, m.From.ID)
	if err != nil {
		log.Printf("failed to get modle for user %d, %v", m.From.ID, err)
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("You are using %s. Please select new model", model.Model))
	msg.ReplyMarkup = NewModelKeyboard(update.Message.Chat.ID)
	bot.botAPI.Send(msg)
}
