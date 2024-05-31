package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/user"
	"errors"
	"log"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type GPTBot struct {
	botAPI      *tgbotapi.BotAPI
	llmService  LlmService
	msgService  MessageService
	userService UserService
}

func NewGPTBot(tgToken string, openAiToken string, db *gorm.DB) GPTBot {
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	llmService := llm.NewOpenAiService(openAiToken)
	msgService := NewMessageDbService(db)
	userService := user.NewUserService(db)

	return GPTBot{botAPI: bot, llmService: llmService, msgService: msgService, userService: userService}
}

func (b *GPTBot) ListenAndServe() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.botAPI.GetUpdatesChan(u)

	for update := range updates {
		if message := update.Message; update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			err := b.userService.ValidateTgUser(update.SentFrom())
			if err != nil {
				b.msgService.SendMessage(b.botAPI, models.Message{ChatId: update.Message.Chat.ID, RepyToId: update.Message.From.ID, Text: "You are not authenticated"})
			} else {
				b.handleMessage(message)
			}

		}
	}
}

func (b *GPTBot) handleMessage(m *tgbotapi.Message) {
	llmCompetionRequest, err := b.getConversationChain(m)
	if err != nil {
		log.Println(err)

	}
	b.msgService.SaveMessage(m, "user")

	llmResp, err := b.llmService.GetCompletionMessage(llmCompetionRequest)
	if err != nil {
		log.Println(err)

	}

	aiResp := models.Message{Id: m.Chat.ID, Text: llmResp, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID, Role: "assistant"}
	msg, err := b.msgService.SendMessage(b.botAPI, aiResp)
	if err != nil {
		log.Printf("failed to send ai respinse")
	}
	b.msgService.SaveMessage(&msg, "assistant")

}

func (b *GPTBot) getConversationChain(m *tgbotapi.Message) ([]llm.CompletionRequestMessage, error) {
	var messageChain []llm.CompletionRequestMessage
	var replyMessageId int64
	if m.Text == "" {
		return messageChain, errors.New("recieved empty message")
	}
	messageChain = append(messageChain, llm.CompletionRequestMessage{Text: m.Text, Role: "user"})
	if replyMessage := m.ReplyToMessage; replyMessage != nil {
		replyMessageId = int64(replyMessage.MessageID)
	}
	for replyMessageId > 0 {
		reply, err := b.msgService.GetMessage(replyMessageId)
		if err != nil {
			log.Println(err)
			break
		}
		messageChain = append(messageChain, llm.CompletionRequestMessage{Text: reply.Text, Role: reply.Role})
		replyMessageId = reply.RepyToId
	}
	slices.Reverse(messageChain)
	return messageChain, nil

}
