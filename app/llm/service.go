package llm

import (
	"amArbaoui/yaggptbot/app/config"
	"context"
	"fmt"
)

type LlmService struct {
	Providers map[string]ChatProvider
}

func (l LlmService) GetCompletionMessage(ctx context.Context, messages []CompletionRequestMessage, userPromt string, model string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	
	providerName, ok := config.ModelMap[model]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrModelNotFound, model)
	}
	provider, ok := l.Providers[providerName]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrProviderNotFound, model)
	}
	return provider.GetCompletionMessage(ctx, messages, userPromt, model)
}
