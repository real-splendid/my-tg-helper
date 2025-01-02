package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Update struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

var (
	botToken   string
	webhookURL string
	port       string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken = os.Getenv("BOT_TOKEN")
	webhookURL = os.Getenv("BOT_WEBHOOK_URL")
	port = os.Getenv("PORT")

	if botToken == "" || webhookURL == "" {
		log.Fatal("BOT_TOKEN and BOT_WEBHOOK_URL must be set in the env")
	}

	if port == "" {
		port = "8443" // default port if not specified
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received message: %s", update.Message.Text)

	reply := map[string]interface{}{
		"chat_id": update.Message.Chat.ID,
		"text":    "Echo: " + update.Message.Text,
	}

	sendMessage(reply)
}

func sendMessage(reply map[string]interface{}) {
	url := "https://api.telegram.org/bot" + botToken + "/sendMessage"

	jsonReply, err := json.Marshal(reply)
	if err != nil {
		log.Println("Error marshaling reply:", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonReply))
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to send message: %s", resp.Status)
	}
}

func setWebhook() {
	url := "https://api.telegram.org/bot" + botToken + "/setWebhook?url=" + webhookURL
	resp, err := http.Get(url)

	if err != nil {
		log.Fatalf("Failed to set webhook: %v", err)
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["ok"].(bool) {
		log.Println("Webhook set successfully", result)
	} else {
		log.Printf("Failed to set webhook: %v", result["description"])
	}
}

func main() {
	setWebhook()
	http.HandleFunc("/webhook", handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
