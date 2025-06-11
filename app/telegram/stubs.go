package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChatServiceStub struct {
}

func (ch *ChatServiceStub) SendMessage(ctx context.Context, SendMsgRequest MessageOut) ([]*tgbotapi.Message, error) {
	return nil, nil
}
