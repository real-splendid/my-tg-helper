package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Error initializing bot: %v", err)
	}

	debug, _ := strconv.ParseBool(os.Getenv("BOT_DEBUG_MODE"))
	bot.Debug = debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	webhookURL := os.Getenv("BOT_WEBHOOK_URL") + "/" + bot.Token
	certPath := os.Getenv("BOT_CERT_PATH")
	wh, _ := tgbotapi.NewWebhookWithCert(webhookURL, tgbotapi.FilePath(certPath))

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatalf("Error setting webhook: %v", err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatalf("Error getting webhook info: %v", err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	port := os.Getenv("BOT_SERVER_PORT")
	go func() {
		err := http.ListenAndServeTLS("0.0.0.0:"+port, certPath, os.Getenv("BOT_KEY_PATH"), nil)
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	for update := range updates {
		log.Printf("Received update: %+v\n", update)
	}
}
