package config

import (
	"amArbaoui/yaggptbot/app/llm/openrouter"

	"github.com/sashabaranov/go-openai"
)

var ModelMap = map[string]string{
	ChatGPT4o:         OpenAI,
	ChatGPTO3MiniHigh: OpenRouter,
	ClaudeSonnet3Dot5: OpenRouter,
	ClaudeSonnet3Dot7: OpenRouter,
	Gemini2Dot0:       OpenRouter,
}
var OpenaiModelMapping = map[string]string{
	ChatGPT4o: openai.GPT4oLatest,
}

var OpenRouterModelMapping = map[string]openrouter.Model{
	Gemini2Dot0:       "google/gemini-2.0-flash-001",
	ChatGPTO3MiniHigh: "openai/o3-mini-high",
	ClaudeSonnet3Dot5: "anthropic/claude-3.5-sonnet",
	ClaudeSonnet3Dot7: "anthropic/claude-3.7-sonnet",
}
