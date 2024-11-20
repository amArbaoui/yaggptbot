package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	SrvAddr            string
	TgToken            string
	OpenAiToken        string
	ApiKey             string
	EncryptionKey      string
	NotificationChatId int64
	AdminChatId        int64
}

func NewConfig() *Config {
	configMap := ReadConfigMap()
	if !ValidConfigMap(configMap) {
		log.Println("Trying to get env variable from .env file")
		loadEnv()
		configMap = ReadConfigMap()
		if !ValidConfigMap(configMap) {
			log.Fatal("failed to set up environment")
		}

	}
	notificationChat, err := strconv.Atoi(configMap["NOTIFICATION_CHAT_ID"])
	if err != nil {
		log.Fatal("Failed to parse NOTIFICATION_CHAT_ID")
	}
	adminChatId, err := strconv.Atoi(configMap["ADMIN_CHAT_ID"])
	if err != nil {
		log.Fatal("Failed to parse ADMIN_CHAT_ID")
	}

	return &Config{SrvAddr: configMap["SERV_ADDR"],
		TgToken:            configMap["TG_TOKEN"],
		OpenAiToken:        configMap["OPENAI_TOKEN"],
		ApiKey:             configMap["X_API_KEY"],
		EncryptionKey:      configMap["ENCRYPTION_KEY"],
		NotificationChatId: int64(notificationChat),
		AdminChatId:        int64(adminChatId),
	}

}

func ReadConfigMap() map[string]string {
	m := map[string]string{
		"SERV_ADDR":            os.Getenv("SERV_ADDR"),
		"TG_TOKEN":             os.Getenv("TG_TOKEN"),
		"OPENAI_TOKEN":         os.Getenv("OPENAI_TOKEN"),
		"X_API_KEY":            os.Getenv("X_API_KEY"),
		"ENCRYPTION_KEY":       os.Getenv("ENCRYPTION_KEY"),
		"NOTIFICATION_CHAT_ID": os.Getenv("NOTIFICATION_CHAT_ID"),
		"ADMIN_CHAT_ID":        os.Getenv("ADMIN_CHAT_ID"),
	}
	return m
}

func ValidConfigMap(configMap map[string]string) bool {
	isValid := true
	for k, v := range configMap {
		if v == "" && v != k {
			isValid = false
			log.Printf("%s environment variable has empty/default value\n", k)
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
