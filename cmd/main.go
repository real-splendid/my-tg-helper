package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

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
	certPath   string
	keyFile    string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken = os.Getenv("BOT_TOKEN")
	webhookURL = os.Getenv("BOT_WEBHOOK_URL")
	port = os.Getenv("BOT_SERVER_PORT")
	certPath = os.Getenv("BOT_CERT_PATH")
	keyFile = os.Getenv("KEY_FILE")

	if botToken == "" || webhookURL == "" || port == "" || certPath == "" {
		log.Fatal("Required environment variables are not set")
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
	certFile, err := os.Open(certPath)
	if err != nil {
		log.Fatalf("Failed to open certificate file: %v", err)
	}
	defer certFile.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("certificate", filepath.Base(certPath))
	if err != nil {
		log.Fatalf("Failed to create form file: %v", err)
	}
	_, err = io.Copy(part, certFile)
	if err != nil {
		log.Fatalf("Failed to copy certificate to form: %v", err)
	}

	err = writer.WriteField("url", webhookURL)
	if err != nil {
		log.Fatalf("Failed to write webhook URL to form: %v", err)
	}

	writer.Close()

	url := "https://api.telegram.org/bot" + botToken + "/setWebhook"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
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

	err := http.ListenAndServeTLS(":"+port, certPath, keyFile, nil)
	if err != nil {
		log.Fatalf("ListenAndServeTLS failed: %v", err)
	}
}
