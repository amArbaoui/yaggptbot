package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"fmt"
)

type LlmService struct {
	Providers map[string]ChatProvider
}

func (l LlmService) GetCompletionMessage(messages []CompletionRequestMessage, userPromt string, model string) (string, error) {
	providerName, ok := config.ModelMap[model]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrModelNotFound, model)
	}
	provider, ok := l.Providers[providerName]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrProviderNotFound, model)
	}
	return provider.GetCompletionMessage(messages, userPromt, model)
}
