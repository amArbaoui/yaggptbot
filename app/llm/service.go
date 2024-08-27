package llm

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type CompletionRequestMessage struct {
	Text     string
	Role     string
	ImageUrl *string
}

type OpenAiService struct {
	Client        openai.Client
	MaxTokens     int
	DefaultPrompt string
}

func NewOpenAiService(secretToken string, maxTokens int, defaultPromt string) *OpenAiService {
	client := openai.NewClient(secretToken)
	return &OpenAiService{Client: *client, MaxTokens: maxTokens, DefaultPrompt: defaultPromt}
}

func (o *OpenAiService) GetCompletionMessage(messages []CompletionRequestMessage, userPromt string) (string, error) {
	ctx := context.Background()
	prompt := o.DefaultPrompt
	if userPromt != "" {
		prompt = userPromt
	}
	systemPromt := openai.ChatCompletionMessage{
		Role:    "system",
		Content: prompt,
	}

	completionMessages := make([]openai.ChatCompletionMessage, 0)
	completionMessages = append(completionMessages, systemPromt)
	for _, message := range messages {
		content := make([]openai.ChatMessagePart, 0)
		content = append(content, openai.ChatMessagePart{Type: openai.ChatMessagePartTypeText, Text: message.Text})
		if message.ImageUrl != nil {
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
		Model:     openai.GPT4o,
		MaxTokens: o.MaxTokens,
		Messages:  completionMessages,
	}
	resp, err := o.Client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil

}
