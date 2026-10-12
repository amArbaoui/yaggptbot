package llm

import (
	"context"
	"errors"
)

var ErrProviderNotFound = errors.New("unknown provider")
var ErrModelNotFound = errors.New("unknown model")
var ErrEmptyCompletion = errors.New("empty completion")

type ChatProvider interface {
	GetCompletionMessage(ctx context.Context, messages []CompletionRequestMessage, userPromt string, model string, opts CompletionOptions) (string, error)
}

type CompletionOptions struct {
	// WebTools lets the model search the web and fetch pages when it decides it needs to
	WebTools bool
}

type CompletionRequestMessage struct {
	Text     string
	Role     string
	ImageUrl *string
}
