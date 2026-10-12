package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/user"
	"context"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetPromptHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	var respText string
	m := update.Message
	messageText := m.Text
	tgId := update.SentFrom().ID
	u, err := bot.userService.GetUserByTgId(ctx, tgId)
	if err != nil {
		log.Printf("failed to get user %v", err)
		return
	}
	if strings.TrimSpace(messageText) == "" {
		bot.TextReplyWithContext(ctx, config.InvalidPromptMessage, m)
		return
	}
	prompt := user.UserPrompt{UserID: u.Id, Prompt: messageText}
	err = bot.userService.SetUserPrompt(ctx, &prompt)
	if err != nil {
		respText = "Error, failed to set new prompt. Please try again"
		log.Print(respText)
	}
	if respText == "" {
		respText = "Prompt updated"
		bot.userService.ResetUserState(ctx, tgId)
	}
	reply := MessageOut{Text: respText, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID}
	bot.chatService.SendMessage(ctx, reply)
}
