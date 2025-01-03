package internal

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken    string
	WebhookURL  string
	Port        string
	CertPath    string
	KeyFile     string
	SberAuthKey string
	RqUID       string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	config := &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		WebhookURL:  os.Getenv("BOT_WEBHOOK_URL"),
		Port:        os.Getenv("BOT_SERVER_PORT"),
		CertPath:    os.Getenv("BOT_CERT_PATH"),
		KeyFile:     os.Getenv("BOT_KEY_PATH"),
		SberAuthKey: os.Getenv("SBER_AUTH_KEY"),
		RqUID:       os.Getenv("SBER_RQUID"),
	}

	if config.BotToken == "" || config.WebhookURL == "" || config.Port == "" || config.CertPath == "" || config.SberAuthKey == "" || config.RqUID == "" {
		log.Fatal("Missing required environment variables")
	}

	return config
}
