package config

import (
	"amArbaoui/yaggptbot/app/llm/openrouter"

	"github.com/sashabaranov/go-openai"
)

var ModelMap = map[string]string{
	ChatGPT4o:         OpenAI,
	ClaudeSonnet3Dot7: OpenRouter,
	ClaudeSonnet4:     OpenRouter,
	Gemini2Dot5Flash:  OpenRouter,
	Gemini2Dot5Pro:    OpenRouter,
	DeepseekV3Dot1:    OpenRouter,
}
var OpenaiModelMapping = map[string]string{
	ChatGPT4o: openai.GPT4oLatest,
}

var OpenRouterModelMapping = map[string]openrouter.Model{
	ClaudeSonnet3Dot7: "anthropic/claude-3.7-sonnet",
	ClaudeSonnet4:     "anthropic/claude-sonnet-4",
	Gemini2Dot5Flash:  "google/gemini-2.5-flash",
	Gemini2Dot5Pro:    "google/gemini-2.5-pro",
	DeepseekV3Dot1:    "deepseek/deepseek-chat-v3.1",
}
