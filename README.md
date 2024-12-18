# Yaggptbot

Yaggptbot - Yet Another Golang GPT Bot.

## Description

This is just a simple TG bot, allowing users to do only one thing - chat with GPT model, using OpenAI API.

**Features:**
* Works on GPT-4o;
* Holds context using TG reply to message;
* All conversations are encrypted and stored in SQLite DB;
* User control via WebAPI;
* Chat completion WebAPI endpoint;
* Small footprint - image size about 20MB;


## Getting Started

### Dependencies
Docker

### Installing
1. Register bot and get tekegran token.
2. Create notification chat and add bot with manage privileges.
3. Get openai token.
4. Drop env file in app folder.

```bash
SERV_ADDR=:8081
TG_TOKEN=TG_TOKEN
OPENAI_TOKEN=OPENAI_TOKEN
X_API_KEY=API-KEY
ENCRYPTION_KEY="dA6ED5MJXumah90N1irZ7KUj6LGP0pGAeN9Aj8uj9b8=" # 32 bit key string
ADMADMIN_CHAT_ID=12345
NOTIFICATION_CHAT_ID=12345
```
Run docker compose up  

When you will first write to bot,  it will send you a payload for a post request to webapi, so run
```bash
curl -X POST --header "X-API-KEY:API-KEY" host:8081/user -d '{"tg_id": 123456, "tg_username": "myuser", "chat_id": 12345}'  
```


## Version History
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
