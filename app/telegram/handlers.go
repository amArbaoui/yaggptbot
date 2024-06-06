package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/models"
	"errors"
	"log"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func UserMesasgeHandler(bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	llmCompetionRequest, err := bot.GetConversationChain(update.Message)
	if err != nil {
		log.Println(err)

	}
	err = bot.msgService.SaveMessage(m, "user")
	if err != nil {
		log.Println(err)
	}

	llmResp, err := bot.llmService.GetCompletionMessage(llmCompetionRequest)
	if err != nil {
		log.Println(err)

	}

	aiResp := models.Message{Id: m.Chat.ID, Text: llmResp, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID, Role: "assistant"}
	msg, err := bot.msgService.SendMessage(bot.botAPI, aiResp)
	if err != nil {
		log.Printf("failed to send ai response, %s", err)
		return
	}
	err = bot.msgService.SaveMessage(msg, "assistant")
	if err != nil {
		log.Println(err)
	}

}

func UserComamndHandler(bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	resp := models.Message{Id: m.Chat.ID, Text: "I don't support any commands by now. Plese send text message to get answer from AI", RepyToId: int64(m.MessageID), ChatId: m.Chat.ID, Role: "assistant"}
	_, err := bot.msgService.SendMessage(bot.botAPI, resp)
	if err != nil {
		log.Println(err)
	}

}

func UserDocumentHandler(bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	resp := models.Message{Id: m.Chat.ID, Text: "Documents not supported", RepyToId: int64(m.MessageID), ChatId: m.Chat.ID, Role: "assistant"}
	_, err := bot.msgService.SendMessage(bot.botAPI, resp)
	if err != nil {
		log.Println(err)
	}

}

func UserPhotoHandler(bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	resp := models.Message{Id: m.Chat.ID, Text: "Photo not supported", RepyToId: int64(m.MessageID), ChatId: m.Chat.ID, Role: "assistant"}
	_, err := bot.msgService.SendMessage(bot.botAPI, resp)
	if err != nil {
		log.Println(err)
	}

}

func (b *GPTBot) GetConversationChain(m *tgbotapi.Message) ([]llm.CompletionRequestMessage, error) {
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
