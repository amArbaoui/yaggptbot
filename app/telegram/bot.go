package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/models"
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotOptions struct {
	MaxConversationDepth int
	BotDebugEnabled      bool
}

type GPTBot struct {
	botAPI      *tgbotapi.BotAPI
	llmService  LlmService
	msgService  MessageService
	userService UserService
	botOptions  BotOptions
}

func NewGPTBot(tgToken string,
	llmservice LlmService,
	messageService MessageService,
	userService UserService,
	botOptions BotOptions) GPTBot {
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = botOptions.BotDebugEnabled
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return GPTBot{botAPI: bot, llmService: llmservice, msgService: messageService, userService: userService, botOptions: botOptions}
}

func (b *GPTBot) ListenAndServe(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.botAPI.GetUpdatesChan(u)
	for {
		select {
		case update := <-updates:
			message := update.Message
			if message == nil {
				continue
			}
			if err := b.ValidateUpdate(update); err != nil {
				continue
			}
			b.RespondToTextMessage(message)
		case <-ctx.Done():
			log.Println("shutting down bot")
			return
		}
	}
}

func (b *GPTBot) ValidateUpdate(update tgbotapi.Update) error {
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

func (b *GPTBot) RespondToTextMessage(m *tgbotapi.Message) {
	llmCompetionRequest, err := b.getConversationChain(m)
	if err != nil {
		log.Println(err)

	}
	err = b.msgService.SaveMessage(m, "user")
	if err != nil {
		log.Println(err)
	}

	llmResp, err := b.llmService.GetCompletionMessage(llmCompetionRequest)
	if err != nil {
		log.Println(err)

	}

	aiResp := models.Message{Id: m.Chat.ID, Text: llmResp, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID, Role: "assistant"}
	msg, err := b.msgService.SendMessage(b.botAPI, aiResp)
	if err != nil {
		log.Printf("failed to send ai response, %s", err)
		return
	}
	err = b.msgService.SaveMessage(&msg, "assistant")
	if err != nil {
		log.Println(err)
	}

}

func (b *GPTBot) getConversationChain(m *tgbotapi.Message) ([]llm.CompletionRequestMessage, error) {
	var messageChain []llm.CompletionRequestMessage
	var replyMessageId int64
	depth := 0
	if m.Text == "" {
		return messageChain, errors.New("recieved empty message")
	}
	messageChain = append(messageChain, llm.CompletionRequestMessage{Text: m.Text, Role: "user"})
	if replyMessage := m.ReplyToMessage; replyMessage != nil {
		replyMessageId = int64(replyMessage.MessageID)
	}
	for replyMessageId > 0 && depth < b.botOptions.MaxConversationDepth {
		reply, err := b.msgService.GetMessage(replyMessageId)
		if err != nil {
			log.Println(err)
			break
		}
		messageChain = append(messageChain, llm.CompletionRequestMessage{Text: reply.Text, Role: reply.Role})
		replyMessageId = reply.RepyToId
		depth++
	}
	slices.Reverse(messageChain)
	return messageChain, nil

}
