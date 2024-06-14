# Yaggptbot

Yaggptbot - Yet Another Golang GPT Bot.

## Description

This is just a simple TG bot, allowing users to do only one thing - chat with GPT model, using OpenAI API.

**Features:**
* Works on GPT-4o;
* Holds context using TG reply to message;
* All conversations are encrypted and stored in SQLite DB;
* User control via WebAPI;
* Small footprint - image size about 20MB;

**TBD:**
* User prompt selection;
* Model selection;
* Image support;
* Chat completion WebAPI endpoint;



## Getting Started

### Dependencies
Docker

### Installing
Drop env file in app folder.
```bash
SERV_ADDR=:8081
TG_TOKEN=TG_TOKEN
OPENAI_TOKEN=OPENAI_TOKEN
X_API_KEY=API-KEY
ENCRYPTION_KEY="dA6ED5MJXumah90N1irZ7KUj6LGP0pGAeN9Aj8uj9b8=" # 32 bit key string
```
Run docker compose up  

When you will first write to bot,  it will send you a payload for a post request to webapi, so run
```bash
curl -X POST --header "X-API-KEY:API-KEY" host:8081/user -d '{"tg_id": 123456, "tg_username": "myuser", "chat_id": 12345}'  
```


## Version History

* 0.1.0
    * Initial Release
