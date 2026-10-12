package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm/openrouter"
	"context"
	"fmt"
)

type OpenRouterProvider struct {
	client        *openrouter.Client
	DefaultPrompt string
}

func NewOpenrouterProvider(client *openrouter.Client, defaultPrompt string) *OpenRouterProvider {
	return &OpenRouterProvider{client: client, DefaultPrompt: defaultPrompt}

}
func (o *OpenRouterProvider) GetCompletionMessage(ctx context.Context, messages []CompletionRequestMessage, userPromt string, model string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	providerModel, ok := config.OpenRouterModelMapping[model]
	if !ok {
		return "", ErrProviderNotFound
	}
	completionMessages := make([]openrouter.CompletionMessage, 0, len(messages)+1)
	if userPromt != "" {
		completionMessages = append(completionMessages, openrouter.CompletionMessage{
			Role:    "system",
			Content: userPromt,
		})
	}
	for _, message := range messages {
		var completionContent interface{}
		switch {
		case message.Role == "assistant":
			completionContent = message.Text
		case message.ImageUrl != nil:
			content := make([]openrouter.CompletionContent, 0, 2)
			if message.Text != "" {
				content = append(content, openrouter.CompletionContent{Type: "text", Text: &message.Text})
			}
			completionContent = append(content, openrouter.CompletionContent{
				Type: "image_url",
				ImageUrl: &openrouter.ImageURL{
					Url: *message.ImageUrl,
				},
			})
		default:
			completionContent = []openrouter.CompletionContent{{
				Type: "text",
				Text: &message.Text,
			},
			}
		}
		completionMessages = append(completionMessages, openrouter.CompletionMessage{
			Role:    message.Role,
			Content: completionContent})
	}
	reasoningEffort, ok := config.ModelReasoningEffort[model]
	if !ok {
		reasoningEffort = "minimal"
	}
	req := openrouter.ChatCompletionRequest{
		Model:     openrouter.Model(providerModel),
		Messages:  completionMessages,
		Reasoning: openrouter.Reasoning{Effort: reasoningEffort},
	}
	completion, err := o.client.GetChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(completion.Choices) == 0 || completion.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("%w: model %s", ErrEmptyCompletion, model)
	}
	return completion.Choices[0].Message.Content, nil
}
