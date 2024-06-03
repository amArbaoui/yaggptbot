package main

import (
	"amArbaoui/yaggptbot/app/api"
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/telegram"
	"amArbaoui/yaggptbot/app/user"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	var wg sync.WaitGroup
	wg.Add(2)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("Received interrupt signal, shutting down")
		cancel()
	}()

	db := storage.GetDb()
	srvAddr := os.Getenv("SERV_ADDR")
	tgToken := os.Getenv("TG_TOKEN")
	openAiToken := os.Getenv("OPENAI_TOKEN")
	apiKey := os.Getenv("X_API_KEY")
	fmt.Println(srvAddr)
	llmService := llm.NewOpenAiService(openAiToken, OPENAI_MAX_TOKENS)
	msgService := telegram.NewMessageDbService(db)
	userService := user.NewUserService(db)
	botOptions := telegram.BotOptions{MaxConversationDepth: TG_MAX_CONVERSATION_DEPTH}
	bot := telegram.NewGPTBot(tgToken, openAiToken, llmService, msgService, userService, botOptions)
	apiServer := api.NewServer(srvAddr, apiKey, userService)
	go bot.ListenAndServe(ctx, &wg)
	go apiServer.Run(ctx, &wg)
	wg.Wait()

}
