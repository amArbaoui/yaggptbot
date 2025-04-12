package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm/openrouter"
)

type OpenRouterProvider struct {
	client        *openrouter.Client
	DefaultPrompt string
}

func NewOpenrouterProvider(client openrouter.Client, defaultPrompt string) *OpenRouterProvider {
	return &OpenRouterProvider{client: &client, DefaultPrompt: defaultPrompt}

}
func (o *OpenRouterProvider) GetCompletionMessage(messages []CompletionRequestMessage, userPromt string, model string) (string, error) {
	providerModel, ok := config.OpenRouterModelMapping[model]
	if !ok {
		return "", ErrProviderNotFound
	}
	completionMessages := make([]openrouter.CompletionMessage, 0, len(messages))
	completionMessages = append(completionMessages, openrouter.CompletionMessage{
		Role: "user", Content: []openrouter.CompletionContent{{
			Type: "text",
			Text: &userPromt,
		},
		},
	},
	)
	for _, message := range messages {
		var completionContent interface{}
		switch {
		case message.Role == "assistant":
			completionContent = message.Text
		case message.ImageUrl != nil:
			completionContent = []openrouter.CompletionContent{{
				Type: "image_url",
				ImageUrl: &openrouter.ImageURL{
					Url: *message.ImageUrl,
				},
			},
			}
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

	req := openrouter.ChatCompletionRequest{
		Model:    openrouter.Model(providerModel),
		Messages: completionMessages,
	}
	completion, err := o.client.GetChatCompletion(req)
	if err != nil {
		return "", err
	}
	return completion.Choices[0].Message.Content, nil
}
