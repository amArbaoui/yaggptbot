package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm/openrouter"
	"context"
	"fmt"
	"log"
	"strings"
)

type OpenRouterProvider struct {
	client        *openrouter.Client
	DefaultPrompt string
}

func NewOpenrouterProvider(client *openrouter.Client, defaultPrompt string) *OpenRouterProvider {
	return &OpenRouterProvider{client: client, DefaultPrompt: defaultPrompt}

}
func (o *OpenRouterProvider) GetCompletionMessage(ctx context.Context, messages []CompletionRequestMessage, userPromt string, model string, opts CompletionOptions) (string, error) {
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
	if opts.WebTools {
		req.Tools = webTools()
	}
	completion, err := o.client.GetChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(completion.Choices) == 0 || completion.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("%w: model %s", ErrEmptyCompletion, model)
	}
	if toolUse := completion.Usage.ServerToolUse; toolUse != nil && toolUse.WebSearchRequests > 0 {
		log.Printf("model %s made %d web searches", model, toolUse.WebSearchRequests)
	}
	message := completion.Choices[0].Message
	return withSources(message.Content, message.Annotations), nil
}

func webTools() []openrouter.Tool {
	return []openrouter.Tool{
		{Type: openrouter.ToolWebSearch, Parameters: openrouter.WebSearchParameters{
			Engine:     config.WebSearchEngine,
			MaxResults: config.WebSearchMaxResults,
			MaxUses:    config.WebSearchMaxUses,
		}},
		{Type: openrouter.ToolWebFetch, Parameters: openrouter.WebFetchParameters{
			Engine:           config.WebFetchEngine,
			MaxUses:          config.WebFetchMaxUses,
			MaxContentTokens: config.WebFetchMaxContentTokens,
		}},
	}
}

// withSources appends cited web pages, which the model didn't link in the text, as a markdown list
func withSources(content string, annotations []openrouter.Annotation) string {
	var sources strings.Builder
	seen := make(map[string]bool)
	count := 0
	for _, a := range annotations {
		if a.URLCitation == nil || a.URLCitation.URL == "" || seen[a.URLCitation.URL] || strings.Contains(content, a.URLCitation.URL) {
			continue
		}
		seen[a.URLCitation.URL] = true
		if count == config.MaxSources {
			break
		}
		count++
		title := a.URLCitation.Title
		if title == "" {
			title = a.URLCitation.URL
		}
		title = strings.NewReplacer("[", "(", "]", ")").Replace(title)
		fmt.Fprintf(&sources, "\n%d. [%s](<%s>)", count, title, a.URLCitation.URL)
	}
	if count == 0 {
		return content
	}
	return content + "\n\n**Sources:**" + sources.String()
}
