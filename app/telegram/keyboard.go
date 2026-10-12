package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"fmt"
	"maps"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewRegistrationKeyboard(tgUsername string, chatId int64) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅", fmt.Sprintf("user:register:%s:%d", tgUsername, chatId)),
		),
	)
}

func NewModelKeyboard(chatId int64, currentModel string) tgbotapi.InlineKeyboardMarkup {
	buttons := make([][]tgbotapi.InlineKeyboardButton, 0, len(config.ModelMap))
	for _, k := range slices.Sorted(maps.Keys(config.ModelMap)) {
		label := k
		if k == currentModel {
			label = "✅ " + k
		}
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("model:set:%s:%d", k, chatId))})
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
}
