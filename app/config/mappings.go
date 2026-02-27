package config

import (
	"amArbaoui/yaggptbot/app/llm/openrouter"

	"github.com/sashabaranov/go-openai"
)

var ModelMap = map[string]string{
	ClaudeSonnet3Dot7: OpenRouter,
	ClaudeSonnet4:     OpenRouter,
	Gemini2Dot5Flash:  OpenRouter,
	Gemini2Dot5Pro:    OpenRouter,
	DeepseekV3Dot1:    OpenRouter,
	Gpt5Dot1Chat:      OpenRouter,
}

var ModelReasoningEffort = map[string]string{
	ClaudeSonnet3Dot7: "none",
	ClaudeSonnet4:     "none",
	Gemini2Dot5Flash:  "none",
	Gemini2Dot5Pro:    "minimal",
	DeepseekV3Dot1:    "none",
	Gpt5Dot1Chat:      "none",
}
var OpenaiModelMapping = map[string]string{
	Gpt5Dot1Chat: openai.GPT5ChatLatest,
}

var OpenRouterModelMapping = map[string]openrouter.Model{
	ClaudeSonnet3Dot7: "anthropic/claude-3.7-sonnet",
	ClaudeSonnet4:     "anthropic/claude-sonnet-4",
	Gemini2Dot5Flash:  "google/gemini-2.5-flash",
	Gemini2Dot5Pro:    "google/gemini-2.5-pro",
	DeepseekV3Dot1:    "deepseek/deepseek-chat-v3.1",
	Gpt5Dot1Chat:      "openai/gpt-5.1-chat",
}
