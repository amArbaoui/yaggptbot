package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAiProvider struct {
	Client        openai.Client
	DefaultPrompt string
	MaxTokens     int
}

func NewOpenAiProvider(apiKey string, maxTokens int, defaultPrompt string) *OpenAiProvider {
	client := openai.NewClient(apiKey)
	return &OpenAiProvider{Client: *client, MaxTokens: maxTokens, DefaultPrompt: defaultPrompt}
}

func (o *OpenAiProvider) GetCompletionMessage(messages []CompletionRequestMessage, userPromt string, model string) (string, error) {
	ctx := context.Background()
	openaiModel, ok := config.OpenaiModelMapping[model]
	if !ok {
		return "", ErrModelNotFound
	}
	systemPromt := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userPromt,
	}

	completionMessages := make([]openai.ChatCompletionMessage, 0)
	completionMessages = append(completionMessages, systemPromt)
	for _, message := range messages {
		content := make([]openai.ChatMessagePart, 0)
		if message.ImageUrl == nil {
			content = append(content, openai.ChatMessagePart{Type: openai.ChatMessagePartTypeText, Text: message.Text})

		} else {
			content = append(content,
				openai.ChatMessagePart{Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    *message.ImageUrl,
						Detail: openai.ImageURLDetailAuto,
					}})
		}
		completionMessages = append(completionMessages, openai.ChatCompletionMessage{MultiContent: content, Role: message.Role})

	}
	req := openai.ChatCompletionRequest{
		Model:    openaiModel,
		Messages: completionMessages,
	}
	resp, err := o.Client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil

}
