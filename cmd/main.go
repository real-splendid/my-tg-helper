package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/real-splendid/my-tg-helper/internal"
)

type Update struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

var config *internal.Config
var sberToken string

func init() {
	config = internal.LoadConfig()

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))

	// Get initial Sber token
	var err error
	sberToken, err = internal.GetSberToken(config.SberAuthKey)
	if err != nil {
		slog.Error("failed to get Sber token", "error", err)
		os.Exit(1)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("received message", "text", update.Message.Text)

	// Convert text to speech
	audioData, err := internal.SynthesizeText(update.Message.Text, sberToken)
	if err != nil {
		slog.Error("failed to synthesize speech", "error", err)
		// Try to refresh token and retry once
		sberToken, err = internal.GetSberToken(config.SberAuthKey)
		if err != nil {
			slog.Error("failed to refresh Sber token", "error", err)
			sendMessage(map[string]interface{}{
				"chat_id": update.Message.Chat.ID,
				"text":    "Sorry, I couldn't convert your text to speech.",
			})
			return
		}
		audioData, err = internal.SynthesizeText(update.Message.Text, sberToken)
		if err != nil {
			slog.Error("failed to synthesize speech after token refresh", "error", err)
			sendMessage(map[string]interface{}{
				"chat_id": update.Message.Chat.ID,
				"text":    "Sorry, I couldn't convert your text to speech.",
			})
			return
		}
	}

	// Send voice message
	sendVoice(update.Message.Chat.ID, audioData)
}

func sendVoice(chatID int64, audioData []byte) {
	url := "https://api.telegram.org/bot" + config.BotToken + "/sendVoice"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add chat_id field
	_ = writer.WriteField("chat_id", fmt.Sprintf("%d", chatID))

	// Add voice file
	part, err := writer.CreateFormFile("voice", "voice.wav")
	if err != nil {
		slog.Error("failed to create form file", "error", err)
		return
	}
	if _, err = io.Copy(part, bytes.NewReader(audioData)); err != nil {
		slog.Error("failed to copy audio data", "error", err)
		return
	}

	if err = writer.Close(); err != nil {
		slog.Error("failed to close writer", "error", err)
		return
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send voice message", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("failed to send voice message", "status", resp.Status)
	}
}

func sendMessage(reply map[string]interface{}) {
	url := "https://api.telegram.org/bot" + config.BotToken + "/sendMessage"

	jsonReply, err := json.Marshal(reply)
	if err != nil {
		slog.Error("error marshaling reply", "error", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonReply))
	if err != nil {
		slog.Error("error sending message", "error", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("failed to send message", "status", resp.Status)
	}
}

func setWebhook() {
	certFile, err := os.Open(config.CertPath)
	if err != nil {
		slog.Error("failed to open certificate file", "error", err)
		os.Exit(1)
	}
	defer certFile.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("certificate", filepath.Base(config.CertPath))
	if err != nil {
		slog.Error("failed to create form file", "error", err)
		os.Exit(1)
	}

	if _, err = io.Copy(part, certFile); err != nil {
		slog.Error("failed to copy certificate", "error", err)
		os.Exit(1)
	}

	if err = writer.WriteField("url", config.WebhookURL); err != nil {
		slog.Error("failed to write webhook URL", "error", err)
		os.Exit(1)
	}

	if err = writer.Close(); err != nil {
		slog.Error("failed to close writer", "error", err)
		os.Exit(1)
	}

	url := "https://api.telegram.org/bot" + config.BotToken + "/setWebhook"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send request", "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("failed to set webhook", "status", resp.Status)
		os.Exit(1)
	}

	slog.Info("webhook set successfully")
}

func main() {
	setWebhook()

	http.HandleFunc("/webhook", handler)
	slog.Info("starting server", "port", config.Port)
	if err := http.ListenAndServeTLS(":"+config.Port, config.CertPath, config.KeyFile, nil); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
