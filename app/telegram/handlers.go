package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/models"
	"amArbaoui/yaggptbot/app/user"
	"errors"
	"fmt"
	"log"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func UserMesasgeHandler(bot *GPTBot, update *tgbotapi.Update) {
	var promptText string
	m := update.Message
	llmCompetionRequest, err := bot.GetConversationChain(m)
	if err != nil {
		if errors.Is(err, ErrMessageNotFound) {
			replyText := "Failed to find reply message(s). Please send your question as a new message"
			_ = bot.TextReply(replyText, m)
			return
		}

	}
	err = bot.msgService.SaveMessage(m, "user")
	if err != nil {
		log.Println(err)
	}

	prompt, err := bot.userService.GetUserPromptByTgId(update.SentFrom().ID)
	switch err {
	case nil:
		promptText = prompt.Prompt
	case user.ErrPromptNotFound:
	default:
		log.Println(err)
	}

	if err == nil {
		promptText = prompt.Prompt
	}
	llmResp, err := bot.llmService.GetCompletionMessage(llmCompetionRequest, promptText)
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

func UserDocumentHandler(bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	text := "Documents not supported"
	_ = bot.TextReply(text, m)

}

func UserPhotoHandler(bot *GPTBot, update *tgbotapi.Update) {
	m := update.Message
	text := "Photo not supported"
	bot.TextReply(text, m)

}

func (b *GPTBot) GetConversationChain(m *tgbotapi.Message) ([]llm.CompletionRequestMessage, error) {
	var messageChain []llm.CompletionRequestMessage
	var replyMessageId int64
	depth := 0
	if m.Text == "" {
		return messageChain, fmt.Errorf("recieved empty message")
	}
	messageChain = append(messageChain, llm.CompletionRequestMessage{Text: m.Text, Role: "user"})
	if replyMessage := m.ReplyToMessage; replyMessage != nil {
		replyMessageId = int64(replyMessage.MessageID)
	}
	for replyMessageId > 0 && depth < b.botOptions.MaxConversationDepth {
		reply, err := b.msgService.GetMessage(replyMessageId)
		if err != nil {
			return nil, err
		}
		messageChain = append(messageChain, llm.CompletionRequestMessage{Text: reply.Text, Role: reply.Role})
		replyMessageId = reply.RepyToId
		depth++
	}
	slices.Reverse(messageChain)
	return messageChain, nil

}
