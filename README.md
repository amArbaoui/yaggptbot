# Yaggptbot

Yaggptbot - Yet Another Golang GPT Bot.

## Description

This is just a simple TG bot, allowing users to do only one thing - chat with GPT models, using API.

**Features:**
* Uses openrouter, multiple models supported;
* Holds context using TG reply to message;
* All conversations are encrypted and stored in SQLite DB;
* User control via WebAPI;
* Chat completion WebAPI endpoint;
* Small footprint - image size about 20MB;


## Getting Started

### Dependencies
Docker

### Installing
1. Register bot and get telegram token.
2. Create notification chat and add bot with manage privileges.
3. Get openai & openrouter tokens.
4. Drop env file in app folder.

```bash
SERV_ADDR=:8081
TG_TOKEN=TG_TOKEN
OPENAI_TOKEN=OPENAI_TOKEN
OPENROUTER_TOKEN=OPENROUTER_TOKEN
X_API_KEY=API-KEY
ENCRYPTION_KEY="dA6ED5MJXumah90N1irZ7KUj6LGP0pGAeN9Aj8uj9b8=" # 32 bit key string
ADMADMIN_CHAT_ID=12345
NOTIFICATION_CHAT_ID=12345
DEFAULT_PROMPT="new prompt" # this is optional
```
Run docker compose up  


## Version History
* 0.4.3
    * Use chatgpt-4o-latest
    * Fixed bug with image handling
    * Improved stability
* 0.4.2
    * Added long messages support (gt telegram limit)
    * Added optional DEFAULT_PROMPT env variable
* 0.4.1
    * Fixed bug with model update
* 0.4.0
    * Add Openrouter provider
    * Add abiity to select model (/setmodel)
    * Add claude-3.5/3.7, o3-mini-high, gemini-2.0
    * Set default model to gemini-2.0.
* 0.3.0
    * Use callback for registration
    * Add notification chat
* 0.2.0
    * Add reports
* 0.1.3
    * Add ability to work with images
* 0.1.2
    * Add user prompts
    * Moved migrations from Makefile to app code
* 0.1.1 
    * Add completion WebAPI endpoint
* 0.1.0
    * Initial Release
