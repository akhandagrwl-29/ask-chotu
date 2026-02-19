package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Telegram API types for getUpdates
type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

type TelegramMessage struct {
	MessageID int           `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      TelegramChat  `json:"chat"`
	Date      int64         `json:"date"`
	Text      string        `json:"text,omitempty"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type getUpdatesResponse struct {
	OK     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

var (
	contextMap           = make(map[string]string)
	historyMap           = make(map[int64]*ChatHistory)
	messageCountMap      = make(map[int64]int)
	updateIDProcessedMap = make(map[int]bool)
)

func listenTelegramUpdates(botToken string, context string) {
	baseURL := fmt.Sprintf("https://api.telegram.org/bot%s", botToken)
	offset := 0

	for {
		url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=60", baseURL, offset)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Telegram getUpdates connection error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var updates getUpdatesResponse
		if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
			log.Printf("Telegram getUpdates decode error: %v", err)
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()

		if !updates.OK {
			log.Printf("Telegram getUpdates returned ok=false")
			time.Sleep(2 * time.Second)
			continue
		}

		if n := len(updates.Result); n > 0 {
			u := updates.Result[n-1]
			updateProcessed := updateIDProcessedMap[u.UpdateID]
			if updateProcessed {
				log.Printf("Telegram update %d already processed, skipping", u.UpdateID)
				continue
			}
			offset = u.UpdateID + 1
			if u.Message != nil && u.Message.Text != "" {
				chatID := u.Message.Chat.ID
				text := strings.TrimSpace(u.Message.Text)
				log.Printf("[chat %d] New message: %s", chatID, text)

				history := historyMap[chatID]
				if history == nil {
					history = &ChatHistory{}
					historyMap[chatID] = history
				}

				count := messageCountMap[chatID]
				count++
				messageCountMap[chatID] = count

				if count >= 140 {
					msg := l2
					if count == 140 {
						msg = l1
					}
					time.Sleep(1 * time.Second)
					if err := sendTelegramMessage(botToken, chatID, msg); err != nil {
						log.Printf("[chat %d] error sending response: %v", chatID, err)
					}
				} else {
					chatbotResponse, err := getChatbotResponse(text, context, history)
					if err != nil {
						log.Printf("[chat %d] error getting chatbot response: %v", chatID, err)
					} else {
						log.Printf("[chat %d] Chatbot response: %s", chatID, chatbotResponse)
						if err := sendTelegramMessage(botToken, chatID, chatbotResponse); err != nil {
							log.Printf("[chat %d] error sending response: %v", chatID, err)
						}
					}
				}

			}
			updateIDProcessedMap[u.UpdateID] = true
		}

		// brief pause when no updates to avoid hammering API
		if len(updates.Result) == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		botToken = os.Getenv("BOT_TOKEN")
	}
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN or BOT_TOKEN env variable not set")
	}

	gistsEnv := os.Getenv("GISTS")
	if gistsEnv == "" {
		log.Fatal("GISTS env variable not set")
	}
	gists := strings.Split(gistsEnv, ",")
	// Use first gist for context (single context for the bot)
	gistID := strings.TrimSpace(gists[0])
	context := getContextText("telegram", gistID)
	log.Printf("Loaded context from gist %s", getTrimmedText(gistID))

	go listenTelegramUpdates(botToken, context)

	log.Println("Telegram bot started. Listening for messages...")
	select {}
}
