package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm"
	"context"
	"fmt"

	"amArbaoui/yaggptbot/app/user"
	"errors"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func UserMesasgeHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	var promptText string
	m := update.Message
	llmCompetionRequest, err := bot.GetConversationChain(ctx, m)
	if err != nil {
		if errors.Is(err, ErrMessageNotFound) {
			replyText := "Failed to find reply message(s). Please send your question as a new message or if my previous message were splitted, reply to the first chunk."
			bot.TextReplyWithContext(ctx, replyText, m)
			return
		}

	}
	llmCompetionRequest = append(llmCompetionRequest, llm.CompletionRequestMessage{Text: m.Text, Role: "user"})
	err = bot.msgService.SaveMessage(ctx, m, "user")
	if err != nil {
		log.Println(err)
	}

	prompt, err := bot.userService.GetUserPromptByTgId(ctx, update.SentFrom().ID)
	switch err {
	case nil:
		promptText = prompt.Prompt
	case user.ErrPromptNotFound:
		promptText = bot.config.DefaultPrompt
	default:
		promptText = bot.config.DefaultPrompt
		log.Println(err)
	}

	model, err := bot.userService.GetUserModelByTgId(ctx, update.SentFrom().ID)
	if err != nil {
		log.Println(err)
	}
	llmResp, err := bot.llmService.GetCompletionMessage(ctx, llmCompetionRequest, promptText, model.Model)
	if err != nil {
		log.Println(err)

	}

	aiResp := MessageOut{Text: llmResp, RepyToId: int64(m.MessageID), ChatId: m.Chat.ID}
	messages, err := bot.chatService.SendMessage(ctx, aiResp)
	if err != nil {
		log.Printf("failed to send ai response, %s", err)
		return
	}

	err = bot.msgService.SaveMessages(ctx, messages, "assistant")
	if err != nil {
		log.Println(err)
	}

}

func UserDocumentHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	m := update.Message
	text := "Documents not supported"
	bot.TextReplyWithContext(ctx, text, m)

}

func UserPhotoHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	var text string
	var caption string
	m := update.Message
	err := bot.msgService.SaveMessage(ctx, m, "user")
	if err != nil {
		log.Println(err)
	}
	imageUrl, err := bot.botAPI.GetFileDirectURL(m.Photo[len(m.Photo)-1].FileID)
	if err != nil {
		text = "Failed to handle photo, please try again"
		bot.TextReplyWithContext(ctx, text, m)
	}
	if m.Caption != "" {
		caption = m.Caption
	} else {
		caption = "Describe this image"
	}
	conversationChain, err := bot.GetConversationChain(ctx, m)
	if err != nil {
		log.Println(err)

	}
	conversationChain = append(conversationChain, llm.CompletionRequestMessage{
		Text:     caption,
		Role:     "user",
		ImageUrl: &imageUrl,
	},
	)
	llmResp, err := bot.llmService.GetCompletionMessage(ctx, conversationChain, bot.config.DefaultPrompt, config.Gpt5Dot1Chat)
	if err != nil {
		log.Println(err)
	}
	msg, err := bot.TextReplyWithContext(ctx, llmResp, m)
	if err != nil {
		log.Printf("failed to send ai response, %s", err)
		return
	}
	for _, m := range msg {
		err = bot.msgService.SaveMessage(ctx, m, "assistant")
		if err != nil {
			log.Println(err)
		}
	}

}

func CallbackHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	data := update.CallbackData()
	if strings.HasPrefix(data, "user:") {
		handleUserCallback(ctx, update, bot)
	}
	if strings.HasPrefix(data, "model:") {
		handleModelSelectionCallback(ctx, update, bot)
	}

}

func handleUserCallback(ctx context.Context, update *tgbotapi.Update, bot *GPTBot) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	data := update.CallbackData()
	split := strings.Split(data, ":")
	operation := split[1]
	userName := split[2]
	userId, _ := strconv.Atoi(split[3])
	if operation == "register" {
		err := bot.userService.SaveUser(
			ctx,
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
				ctx,
				MessageOut{
					Text:     tgbotapi.EscapeText(tgbotapi.ModeMarkdown, config.GreetUserMessage),
					ChatId:   int64(userId),
					RepyToId: 0,
				},
			)
			if err != nil {
				log.Println("failed to greet user")
			}
			_, err = bot.chatService.SendMessage(
				ctx,
				MessageOut{
					Text:     tgbotapi.EscapeText(tgbotapi.ModeMarkdown, config.HowToUseItMessage),
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

func handleModelSelectionCallback(ctx context.Context, update *tgbotapi.Update, bot *GPTBot) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	var reply string
	data := update.CallbackData()
	split := strings.Split(data, ":")
	operation := split[1]
	modelName := split[2]
	userId, _ := strconv.Atoi(split[3])
	userEntity, err := bot.userService.GetUserByTgId(ctx, int64(userId))
	if err != nil {
		reply = "error, failed to set model"
		bot.botAPI.Send(tgbotapi.NewMessage(int64(userId), reply))
		return
	}
	if operation == "set" {
		err := bot.userService.SetUserModel(ctx, &user.UserModel{UserID: userEntity.Id, Model: modelName})
		if err != nil {
			reply = "error, failed to set model"
		}
		reply = fmt.Sprintf("you're now using %s", modelName)
		bot.botAPI.Send(tgbotapi.NewMessage(int64(userId), reply))

	}
}

func (b *GPTBot) GetConversationChain(ctx context.Context, m *tgbotapi.Message) ([]llm.CompletionRequestMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	completionRequest := make([]llm.CompletionRequestMessage, 0)
	if m.ReplyToMessage == nil {
		return completionRequest, nil
	}
	messageChain, err := b.msgService.GetMessageChain(ctx, int64(m.ReplyToMessage.MessageID), b.botOptions.MaxConversationDepth)
	if err != nil {
		return completionRequest, err
	}
	for _, message := range messageChain {
		completionRequest = append(completionRequest, llm.CompletionRequestMessage{Text: message.Text, Role: message.Role, ImageUrl: nil})
	}
	return completionRequest, nil
}
