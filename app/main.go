package main

import (
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/telegram"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadenv() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	loadenv()
	db := storage.GetDb()
	tgToken := os.Getenv("TG_TOKEN")
	openAiToken := os.Getenv("OPENAI_TOKEN")
	bot := telegram.NewGPTBot(tgToken, openAiToken, db)
	bot.ListenAndServe()
}
