package config

import (
	"amArbaoui/yaggptbot/app/llm/openrouter"

	"github.com/sashabaranov/go-openai"
)

var ModelMap = map[string]string{
	ClaudeSonnet5:    OpenRouter,
	Gemini3Dot7Flash: OpenRouter,
	DeepseekV4Flash:  OpenRouter,
	Gpt5Dot6Luna:     OpenRouter,
}

var ModelReasoningEffort = map[string]string{
	ClaudeSonnet5:    "none",
	Gemini3Dot7Flash: "minimal",
	DeepseekV4Flash:  "none",
	Gpt5Dot6Luna:     "none",
}
var OpenaiModelMapping = map[string]string{
	Gpt5Dot6Luna: openai.GPT5ChatLatest,
}

var OpenRouterModelMapping = map[string]openrouter.Model{
	ClaudeSonnet5:    "anthropic/claude-sonnet-5",
	Gemini3Dot7Flash: "google/gemini-3.7-flash",
	DeepseekV4Flash:  "deepseek/deepseek-v4-flash-0731",
	Gpt5Dot6Luna:     "openai/gpt-5.6-luna",
}
