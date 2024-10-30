package telegram

import (
	"amArbaoui/yaggptbot/app/user"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetPromptCommand(bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	currentPrompt, err := bot.userService.GetUserPromptByTgId(m.From.ID)
	var respText string
	if err != nil {
		respText = "Please enter new prompt"
	} else {
		respText = fmt.Sprintf("Current prompt is: %s\nPlease set new prompt", currentPrompt.Prompt)
	}
	resp := MessageOut{Text: respText, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID}
	err = bot.userService.SetUserState(m.From.ID, user.SETTING_PROMT)
	if err != nil {
		log.Printf("failed to set state %v", err)
		return
	}
	bot.chatService.SendMessage(resp)
}

func ResetPromtCommand(bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	respText := "Prompt removed"
	resp := MessageOut{Text: respText, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID}
	err := bot.userService.RemoveUserPromt(m.From.ID)
	if err != nil {
		log.Printf("failed to remove prompt %v", err)
		return
	}
	bot.userService.ResetUserState(m.From.ID)
	bot.chatService.SendMessage(resp)
}
