package config

import (
	"amArbaoui/yaggptbot/app/llm/openrouter"

	"github.com/sashabaranov/go-openai"
)

var ModelMap = map[string]string{
	ChatGPT4o:         OpenAI,
	ClaudeSonnet3Dot5: OpenRouter,
	ClaudeSonnet3Dot7: OpenRouter,
	Gemini2Dot0:       OpenRouter,
	ClaudeSonnet4:     OpenRouter,
	Gemini2Dot5Flash:  OpenRouter,
	Gemini2Dot5Pro:    OpenRouter,
}
var OpenaiModelMapping = map[string]string{
	ChatGPT4o: openai.GPT4oLatest,
}

var OpenRouterModelMapping = map[string]openrouter.Model{
	Gemini2Dot0:       "google/gemini-2.0-flash-001",
	ClaudeSonnet3Dot5: "anthropic/claude-3.5-sonnet",
	ClaudeSonnet3Dot7: "anthropic/claude-3.7-sonnet",
	ClaudeSonnet4:     "anthropic/claude-sonnet-4",
	Gemini2Dot5Flash:  "google/gemini-2.5-flash",
	Gemini2Dot5Pro:    "google/gemini-2.5-pro",
}
