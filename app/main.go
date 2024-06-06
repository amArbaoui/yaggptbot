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

const (
	MaxMessageContextDepth = 30
	OpenAiMaxTokens        = 1000
	BotDebugEnabled        = false
)

type Config struct {
	srvAddr       string
	tgToken       string
	openAiToken   string
	apiKey        string
	encryptionKey string
}

func NewConfig() *Config {
	configMap := ReadConfigMap()
	if !ValidConfigMap(configMap) {
		fmt.Println("Trying to get env variable from .env file")
		loadEnv()
		configMap = ReadConfigMap()
		if !ValidConfigMap(configMap) {
			log.Fatal("failed to set up environment")
		}

	}

	return &Config{srvAddr: configMap["SERV_ADDR"],
		tgToken:       configMap["TG_TOKEN"],
		openAiToken:   configMap["OPENAI_TOKEN"],
		apiKey:        configMap["X_API_KEY"],
		encryptionKey: configMap["ENCRYPTION_KEY"]}

}

func ReadConfigMap() map[string]string {
	m := map[string]string{
		"SERV_ADDR":      os.Getenv("SERV_ADDR"),
		"TG_TOKEN":       os.Getenv("TG_TOKEN"),
		"OPENAI_TOKEN":   os.Getenv("OPENAI_TOKEN"),
		"X_API_KEY":      os.Getenv("X_API_KEY"),
		"ENCRYPTION_KEY": os.Getenv("ENCRYPTION_KEY"),
	}
	return m
}

func ValidConfigMap(configMap map[string]string) bool {
	isValid := true
	for k, v := range configMap {
		if v == "" && v != k {
			isValid = false
			fmt.Printf("%s environment variable has empty/default value\n", k)
		}
	}
	return isValid

}

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
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

	db := storage.GetDB()
	cnf := NewConfig()
	llmService := llm.NewOpenAiService(cnf.openAiToken, OpenAiMaxTokens)
	encryptionService := storage.NewEncryptionService(cnf.encryptionKey)
	msgService := telegram.NewMessageDbService(db, encryptionService)
	userService := user.NewUserService(db)
	botOptions := telegram.BotOptions{MaxConversationDepth: MaxMessageContextDepth, BotDebugEnabled: BotDebugEnabled}
	bot := telegram.NewGPTBot(cnf.tgToken, llmService, msgService, userService, botOptions)
	apiServer := api.NewServer(cnf.srvAddr, cnf.apiKey, userService)
	go bot.StartPolling(ctx, &wg)
	go apiServer.Run(ctx, &wg)
	wg.Wait()

}
