package main

import (
	"amArbaoui/yaggptbot/app/api"
	"amArbaoui/yaggptbot/app/config"
	"amArbaoui/yaggptbot/app/llm"
	"amArbaoui/yaggptbot/app/report"
	"amArbaoui/yaggptbot/app/storage"
	"amArbaoui/yaggptbot/app/telegram"
	"amArbaoui/yaggptbot/app/user"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var wg sync.WaitGroup
	wg.Add(3)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-stop
		log.Printf("Received interrupt signal, shutting down")
		cancel()
	}()

	db := storage.GetDB()
	cnf := config.NewConfig()
	encryptionService := storage.NewEncryptionService(cnf.EncryptionKey)
	botApi, err := tgbotapi.NewBotAPI(cnf.TgToken)
	if err != nil {
		log.Panic(err)
	}
	msgService := telegram.NewMessageDbService(db, encryptionService)
	chatService := telegram.NewChatService(botApi)
	llmService := llm.NewOpenAiService(cnf.OpenAiToken, config.OpenAiMaxTokens, config.DefaultPromt)
	userService := user.NewUserService(db)
	reportService := report.NewReportService(chatService, cnf.NotificationChatId, db)
	reportScheduler, _ := report.NewReportScheduler(&reportService)
	botOptions := telegram.BotOptions{MaxConversationDepth: config.MaxMessageContextDepth, BotDebugEnabled: config.BotDebugEnabled, BotAdminChatId: cnf.AdminChatId, NotificationChatId: cnf.NotificationChatId}
	bot := telegram.NewGPTBot(botApi, chatService, llmService, msgService, userService, botOptions)
	apiServer := api.NewServer(cnf.SrvAddr, cnf.ApiKey, userService, chatService, llmService)
	go bot.StartPolling(ctx, &wg)
	go apiServer.Run(ctx, &wg)
	go reportScheduler.Run(ctx, &wg)
	wg.Wait()

}
