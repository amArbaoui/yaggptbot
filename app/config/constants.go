package config

const GreetUserMessage = `👋 Hello and welcome! The adminstrator has approved your request to use this bot`
const HowToUseItMessage = `🚀 Let's get started!
1. Start a Conversation
   - Simply send a message to the bot to begin interaction.

2. Responding to Messages:
   - The bot will use the message's context to provide a relevant response, so use reply function.

3. Setting a Custom Prompt:
   - Use the command /promtset <your_prompt> to set a specific prompt that the bot should consider in its replies.

4. Resetting the Prompt:
   - Use the command /promtreset to remove any previously set prompt and return to default settings.

Happy chatting!
`

const (
	OpenAI                 = "openai"
	OpenRouter             = "openrouter"
	OpenRouterApiUrl       = "https://openrouter.ai/api/v1"
	MaxMessageContextDepth = 30
	OpenAiMaxTokens        = 1000
	TelegramMessageLimit   = 4096
	BotDebugEnabled        = false
	DefaultPrompt          = `SYSTEM PROMPT:
		You should reply using STRICTLY valid telegram markdown v1 markup.
		Don't tell user about it, it's internal detail\n`

	ChatGPT4o         = "chatgpt-4o"
	ClaudeSonnet3Dot7 = "claude-3.7-sonnet"
	ClaudeSonnet4     = "claude-4-sonnet"
	Gemini2Dot5Flash  = "gemini-2.5-flash"
	Gemini2Dot5Pro    = "gemini-2.5-pro"
	DeepseekV3Dot1    = "deepseek-chat-v3.1"
)

const DefaultModel = DeepseekV3Dot1
