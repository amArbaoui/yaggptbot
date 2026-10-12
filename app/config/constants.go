package config

import "time"

const LlmErrorMessage = `⚠️ Failed to get a response from %s. Please try again or switch model with /setmodel`

const GreetUserMessage = `👋 Hello and welcome! The adminstrator has approved your request to use this bot`
const HowToUseItMessage = `🚀 Let's get started!
1. Start a conversation
   - Just send a message and the bot will answer.

2. Keep the context
   - Reply to the bot's message to continue the conversation. A message without a reply starts a new one.

3. Images
   - Send a photo, optionally with a caption, to ask about it.

4. Commands
   - /setmodel - choose the model
   - /promptset - set a custom prompt, the bot will ask for its text
   - /promptreset - remove your custom prompt
   - /cancel - cancel the current operation
   - /help - show this message

Happy chatting!
`

const (
	UnknownCommandMessage        = "Unknown command. Use /help to see available commands"
	RegistrationRequestedMessage = "Your request has been sent to the admin. You will get a message once it is approved"
	RegistrationPendingMessage   = "Your request is already waiting for the admin's approval"
	RegistrationFailedMessage    = "Failed to send your request to the admin, please try again later"
	SelectModelMessage           = "You are using %s. Please select new model"
	SetModelFailedMessage        = "Failed to set model, please try again"
	EnterPromptMessage           = "Please send the new prompt text, or /cancel"
	CurrentPromptMessage         = "Current prompt is: %s\n\nPlease send the new prompt text, or /cancel"
	InvalidPromptMessage         = "Please send the prompt as a text message, or /cancel"
	CancelledMessage             = "Cancelled"
	NothingToCancelMessage       = "Nothing to cancel"
)

const (
	OpenAI                 = "openai"
	OpenRouter             = "openrouter"
	OpenRouterApiUrl       = "https://openrouter.ai/api/v1"
	MaxMessageContextDepth = 30
	OpenAiMaxTokens        = 1000
	TelegramMessageLimit   = 4096
	BotDebugEnabled        = false
	TypingActionInterval   = 4 * time.Second
	DefaultPrompt          = `SYSTEM PROMPT:
		You are chatting with the user in Telegram. Format replies with standard Markdown.
		LaTeX is not rendered, write formulas in plain text or Unicode.
		Don't tell user about it, it's internal detail`

	Gpt5Dot6Luna     = "chatgpt-5.6-luna"
	ClaudeSonnet5    = "claude-sonnet-5"
	Gemini3Dot7Flash = "gemini-3.7-flash"
	DeepseekV4Flash  = "deepseek-chat-v3.1"
)

const DefaultModel = Gpt5Dot6Luna
