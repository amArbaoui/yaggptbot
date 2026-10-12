package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/storage"
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

	m := update.Message
	conversation, err := bot.GetConversationChain(ctx, m)
	if err != nil {
		if errors.Is(err, storage.ErrMessageNotFound) {
			replyText := "Failed to find reply message(s). Please send your question as a new message."
			bot.TextReplyWithContext(ctx, replyText, m)
			return
		}
		log.Println(err)
	}
	conversation = append(conversation, llm.CompletionRequestMessage{Text: m.Text, Role: "user"})
	err = bot.msgService.SaveMessage(ctx, m, "user")
	if err != nil {
		log.Println(err)
	}
	bot.answer(ctx, m, conversation)
}

func UserDocumentHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	m := update.Message
	if imageFileID(m) != "" {
		UserPhotoHandler(ctx, bot, update)
		return
	}
	bot.TextReplyWithContext(ctx, config.DocumentsNotSupportedMessage, m)
}

// UserPhotoHandler handles photos and images sent as files
func UserPhotoHandler(ctx context.Context, bot *GPTBot, update *tgbotapi.Update) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	m := update.Message
	if imageFileSize(m) > config.MaxImageSize {
		bot.TextReplyWithContext(ctx, config.ImageTooLargeMessage, m)
		return
	}
	imageURL, err := bot.imageDataURL(ctx, imageFileID(m))
	if err != nil {
		log.Printf("failed to get image for message %d: %v", m.MessageID, err)
		replyText := config.ImageFailedMessage
		if errors.Is(err, ErrImageTooLarge) {
			replyText = config.ImageTooLargeMessage
		}
		bot.TextReplyWithContext(ctx, replyText, m)
		return
	}
	conversation, err := bot.GetConversationChain(ctx, m)
	if err != nil {
		log.Println(err)
	}
	caption := m.Caption
	if caption == "" {
		caption = config.DefaultImageCaption
	}
	conversation = append(conversation, llm.CompletionRequestMessage{
		Text:     caption,
		Role:     "user",
		ImageUrl: &imageURL,
	})
	err = bot.msgService.SaveMessage(ctx, m, "user")
	if err != nil {
		log.Println(err)
	}
	bot.answer(ctx, m, conversation)
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
		bot.pendingRegistrations.Delete(int64(userId))
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
					Text:     config.GreetUserMessage,
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
					Text:     config.HowToUseItMessage,
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

	query := update.CallbackQuery
	split := strings.Split(query.Data, ":")
	if len(split) < 3 || split[1] != "set" {
		return
	}
	modelName := split[2]
	answer := tgbotapi.NewCallback(query.ID, "")
	defer func() {
		if _, err := bot.botAPI.Request(answer); err != nil {
			log.Printf("failed to answer callback, %v", err)
		}
	}()

	userEntity, err := bot.userService.GetUserByTgId(ctx, query.From.ID)
	if err != nil {
		log.Printf("failed to get user %d, %v", query.From.ID, err)
		answer.Text = config.SetModelFailedMessage
		return
	}
	if _, ok := config.ModelMap[modelName]; !ok {
		log.Printf("unknown model %s selected by user %d", modelName, query.From.ID)
		answer.Text = config.SetModelFailedMessage
		return
	}
	err = bot.userService.SetUserModel(ctx, &user.UserModel{UserID: userEntity.Id, Model: modelName})
	if err != nil {
		log.Printf("failed to set model for user %d, %v", query.From.ID, err)
		answer.Text = config.SetModelFailedMessage
		return
	}
	answer.Text = fmt.Sprintf("You're now using %s", modelName)
	if query.Message == nil {
		return
	}
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		query.Message.Chat.ID,
		query.Message.MessageID,
		fmt.Sprintf(config.SelectModelMessage, modelName),
		NewModelKeyboard(query.Message.Chat.ID, modelName),
	)
	if _, err := bot.botAPI.Request(edit); err != nil {
		log.Printf("failed to update model keyboard, %v", err)
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
	messageChain, err := b.msgService.GetMessageChain(ctx, m.Chat.ID, int64(m.ReplyToMessage.MessageID), b.botOptions.MaxConversationDepth)
	if err != nil {
		return completionRequest, err
	}
	for _, message := range messageChain {
		request := llm.CompletionRequestMessage{Text: message.Text, Role: message.Role}
		if message.ImageFileId != nil {
			imageURL, err := b.imageDataURL(ctx, *message.ImageFileId)
			if err != nil {
				log.Printf("failed to get image of message %d: %v", message.TgMsgId, err)
			} else {
				request.ImageUrl = &imageURL
			}
		}
		if request.Text == "" && request.ImageUrl == nil {
			continue
		}
		completionRequest = append(completionRequest, request)
	}
	return completionRequest, nil
}
