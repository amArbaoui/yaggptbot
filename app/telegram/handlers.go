package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm"
	"fmt"

	"amArbaoui/yaggptbot/app/user"
	"errors"
	"log"
	"strconv"
	"strings"

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
	model, err := bot.userService.GetUserModelByTgId(update.SentFrom().ID)
	if err != nil {
		log.Println(err)
	}
	llmResp, err := bot.llmService.GetCompletionMessage(llmCompetionRequest, promptText, model.Model)
	if err != nil {
		log.Println(err)

	}

	aiResp := MessageOut{Text: llmResp, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID}
	messages, err := bot.chatService.SendMessage(aiResp)
	if err != nil {
		log.Printf("failed to send ai response, %s", err)
		return
	}

	err = bot.msgService.SaveMessages(messages, "assistant")
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
	llmResp, err := bot.llmService.GetCompletionMessage(conversationChain, "", config.ChatGPT4o)
	if err != nil {
		log.Println(err)
	}
	msg, err := bot.TextReply(llmResp, m)
	if err != nil {
		log.Printf("failed to send ai response, %s", err)
		return
	}
	for _, m := range msg {
		err = bot.msgService.SaveMessage(m, "assistant")
		if err != nil {
			log.Println(err)
		}
	}

}

func CallbackHandler(bot *GPTBot, update *tgbotapi.Update) {
	data := update.CallbackData()
	if strings.HasPrefix(data, "user:") {
		handleUserCallback(update, bot)
	}
	if strings.HasPrefix(data, "model:") {
		handleModelSelectionCallback(update, bot)
	}

}

func handleUserCallback(update *tgbotapi.Update, bot *GPTBot) {
	data := update.CallbackData()
	split := strings.Split(data, ":")
	operation := split[1]
	userName := split[2]
	userId, _ := strconv.Atoi(split[3])
	if operation == "register" {
		err := bot.userService.SaveUser(
			&user.User{
				Id:     int64(userId),
				TgName: userName,
				ChatId: int64(userId),
			},
		)
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
		if _, err := bot.botAPI.Request(callback); err != nil {
			fmt.Println(err)
		}
		if err == nil {
			_, err := bot.chatService.SendMessage(
				MessageOut{
					Text:     tgbotapi.EscapeText("Markdown", config.GreetUserMessage),
					ChatId:   int64(userId),
					RepyToId: 0,
				},
			)
			if err != nil {
				log.Println("failed to greet user")
			}
			_, err = bot.chatService.SendMessage(
				MessageOut{
					Text:     tgbotapi.EscapeText("Markdown", config.HowToUseItMessage),
					ChatId:   int64(userId),
					RepyToId: 0,
				},
			)
			if err != nil {
				log.Println("failed to send instruction")
			}

		}

	}
}

func handleModelSelectionCallback(update *tgbotapi.Update, bot *GPTBot) {
	var reply string
	data := update.CallbackData()
	split := strings.Split(data, ":")
	operation := split[1]
	modelName := split[2]
	userId, _ := strconv.Atoi(split[3])
	userEntity, err := bot.userService.GetUserByTgId(int64(userId))
	if err != nil {
		reply = "error, failed to set model"
		bot.botAPI.Send(tgbotapi.NewMessage(int64(userId), reply))
		return
	}
	if operation == "set" {
		err := bot.userService.SetUserModel(&user.UserModel{UserID: userEntity.Id, Model: modelName})
		if err != nil {
			reply = "error, failed to set model"
		}
		reply = fmt.Sprintf("you're now using %s", modelName)
		bot.botAPI.Send(tgbotapi.NewMessage(int64(userId), reply))

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
