package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type NtfyEvent struct {
	ID      string `json:"id"`
	Time    int64  `json:"time"`
	Event   string `json:"event"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

func listenTopic(topic string) {
	url := fmt.Sprintf("https://ntfy.sh/%s/json", topic)

	for {
		log.Printf("Listening on topic: %s\n", topic)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[%s] connection error: %v", topic, err)
			time.Sleep(5 * time.Second)
			continue
		}

		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			line := scanner.Bytes()

			var event NtfyEvent
			if err := json.Unmarshal(line, &event); err != nil {
				log.Printf("[%s] invalid json: %v", topic, err)
				continue
			}

			log.Printf("[%s] event=%+v\n", topic, event)

			if event.Event == "message" && !strings.Contains(event.Message, "Chotu") {
				log.Printf("[%s] New message: %s\n", topic, event.Message)

				http.Post(
					fmt.Sprintf("https://ntfy.sh/%s", topic),
					"text/plain",
					strings.NewReader(
						fmt.Sprintf("Ask Chotu: %s", event.Message),
					),
				)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("[%s] scanner error: %v", topic, err)
		}

		resp.Body.Close()

		// reconnect if stream closes
		time.Sleep(2 * time.Second)
	}
}

func main() {
	topicsEnv := os.Getenv("NTFY_TOPICS")
	if topicsEnv == "" {
		log.Fatal("NTFY_TOPICS env variable not set")
	}

	topics := strings.Split(topicsEnv, ",")

	for _, topic := range topics {
		go listenTopic(strings.TrimSpace(topic))
	}

	select {}
}

