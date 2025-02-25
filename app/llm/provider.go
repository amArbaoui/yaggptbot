package llm

import "errors"

var ErrProviderNotFound = errors.New("unknown provider")
var ErrModelNotFound = errors.New("unknown model")

type ChatProvider interface {
	GetCompletionMessage(messages []CompletionRequestMessage, userPromt string, model string) (string, error)
}

type CompletionRequestMessage struct {
	Text     string
	Role     string
	ImageUrl *string
}
