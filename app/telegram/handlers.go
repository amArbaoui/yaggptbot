package telegram

import (
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/user"
	"errors"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func UserMesasgeHandler(bot *GPTBot, update *tgbotapi.Update) {
	var promptText string
	m := update.Message
	llmCompetionRequest, err := bot.GetConversationChain(m)
	if err != nil {
		if errors.Is(err, ErrMessageNotFound) {
			replyText := "Failed to find reply message(s). Please send your question as a new message"
			bot.TextReply(replyText, m)
			return
		}

	}
	llmCompetionRequest = append(llmCompetionRequest, llm.CompletionRequestMessage{Text: m.Text, Role: "user"})
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

	aiResp := Message{Id: m.Chat.ID, Text: llmResp, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID, Role: "assistant"}
	msg, err := bot.chatService.SendMessage(aiResp)
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
	bot.TextReply(text, m)

}

func UserPhotoHandler(bot *GPTBot, update *tgbotapi.Update) {
	var text string
	var caption string
	m := update.Message
	err := bot.msgService.SaveMessage(m, "user")
	if err != nil {
		log.Println(err)
	}
	imageUrl, err := bot.botAPI.GetFileDirectURL(m.Photo[len(m.Photo)-1].FileID)
	if err != nil {
		text = "Failed to handle photo, please try again"
		bot.TextReply(text, m)
	}
	if m.Caption != "" {
		caption = m.Caption
	} else {
		caption = "Describe this image"
	}
	conversationChain, err := bot.GetConversationChain(m)
	if err != nil {
		log.Println(err)

	}
	conversationChain = append(conversationChain, llm.CompletionRequestMessage{
		Text:     caption,
		Role:     "user",
		ImageUrl: &imageUrl,
	},
	)
	llmResp, err := bot.llmService.GetCompletionMessage(conversationChain, "")
	if err != nil {
		log.Println(err)
	}
	msg, err := bot.TextReply(llmResp, m)
	if err != nil {
		log.Printf("failed to send ai response, %s", err)
		return
	}
	err = bot.msgService.SaveMessage(msg, "assistant")
	if err != nil {
		log.Println(err)
	}

}

func (b *GPTBot) GetConversationChain(m *tgbotapi.Message) ([]llm.CompletionRequestMessage, error) {
	completionRequest := make([]llm.CompletionRequestMessage, 0)
	if m.ReplyToMessage == nil {
		return completionRequest, nil
	}
	messageChain, err := b.msgService.GetMessageChain(int64(m.ReplyToMessage.MessageID), b.botOptions.MaxConversationDepth)
	if err != nil {
		return completionRequest, err
	}
	for _, message := range messageChain {
		completionRequest = append(completionRequest, llm.CompletionRequestMessage{Text: message.Text, Role: message.Role, ImageUrl: nil})

	}
	return completionRequest, nil

}
