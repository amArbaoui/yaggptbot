package llm

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type CompletionRequestMessage struct {
	Text string
	Role string
}

type OpenAiService struct {
	Client    openai.Client
	MaxTokens int
}

func NewOpenAiService(secretToken string, maxTokens int) *OpenAiService {
	client := openai.NewClient(secretToken)
	return &OpenAiService{Client: *client, MaxTokens: maxTokens}
}

func (o *OpenAiService) GetCompletionMessage(messages []CompletionRequestMessage) (string, error) {
	ctx := context.Background()
	completionMessages := make([]openai.ChatCompletionMessage, 0)
	for _, message := range messages {
		completionMessages = append(completionMessages, openai.ChatCompletionMessage{Content: message.Text, Role: message.Role})

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
