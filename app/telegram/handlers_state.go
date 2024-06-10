package telegram

import (
	"amArbaoui/yaggptbot/app/models"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SettingPromptHandler(bot *GPTBot, update *tgbotapi.Update) {
	var respText string
	m := update.Message
	messageText := m.Text
	tgId := update.SentFrom().ID
	user, err := bot.userService.GetUserByTgId(tgId)
	if err != nil {
		log.Printf("failed to get user %v", err)
		return
	}
	if messageText == "" {
		respText = "Please send valid prompt text"

	}
	newPrompt := fmt.Sprintf("SYSTEM_PROMPT: You should reply only in valid Telegram MarkDown V1 markup.  USER_PROMPT: %s", messageText)
	prompt := models.UserPrompt{UserID: user.Id, Prompt: newPrompt}
	err = bot.userService.SetUserPrompt(&prompt)
	if err != nil {
		respText = "Error, failed to set new prompt. Please try again"
		log.Print(respText)
	}
	if respText == "" {
		respText = "Prompt updated"
		bot.userService.ResetUserState(tgId)
	}
	reply := models.Message{Id: m.Chat.ID, Text: respText, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID, Role: "system"}
	bot.msgService.SendMessage(bot.botAPI, reply)
}
