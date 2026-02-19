package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func sendTelegramMessage(botToken string, chatID int64, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	body := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal sendMessage body: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("sendMessage request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sendMessage status %d", resp.StatusCode)
	}
	return nil
}
