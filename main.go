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

var (
	contextMap = make(map[string]string)
	historyMap = make(map[string]*ChatHistory)
)

func listenTopic(topic string, context string, history *ChatHistory) {
	url := fmt.Sprintf("https://ntfy.sh/%s/json", topic)
	count := 0

	for {
		log.Printf("Listening on topic: %s\n", getTrimmedText(topic))

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[%s] connection error: %v", getTrimmedText(topic), err)
			time.Sleep(2 * time.Second)
			continue
		}

		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			line := scanner.Bytes()

			var event NtfyEvent
			if err := json.Unmarshal(line, &event); err != nil {
				log.Printf("[%s] invalid json: %v", getTrimmedText(topic), err)
				continue
			}

			log.Printf("[%s] event=%+v\n", getTrimmedText(topic), event)

			if event.Event == "message" && !strings.Contains(event.Message, "Chotu:") {
				count++
				log.Printf("[%s] New message: %s\n", getTrimmedText(topic), event.Message)

				if count >= 20 {
					msg := l2
					if count == 20 {
						msg = l1
					}
					time.Sleep(1 * time.Second)
					_ = publishResponse(topic, msg)
					if err != nil {
						log.Printf("[%s] error publishing response: %v", getTrimmedText(topic), err)
						continue
					}
					continue
				}

				chatbotResponse, err := getChatbotResponse(event.Message, context, history)
				if err != nil {
					log.Printf("[%s] error getting chatbotResponse response: %v", getTrimmedText(topic), err)
					continue
				}

				log.Printf("[%s] Chatbot response: %s\n", getTrimmedText(topic), chatbotResponse)

				err = publishResponse(topic, chatbotResponse)
				if err != nil {
					log.Printf("[%s] error publishing response: %v", getTrimmedText(topic), err)
					continue
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("[%s] scanner error: %v", getTrimmedText(topic), err)
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

	gistsEnv := os.Getenv("GISTS")
	if gistsEnv == "" {
		log.Fatal("GISTS env variable not set")
	}
	gists := strings.Split(gistsEnv, ",")

	for i, gist := range gists {
		topic := topics[i]
		context := getContextText(topic, gist)
		contextMap[topic] = context
		log.Printf("Loaded context for topic %s from gist %s\n", getTrimmedText(topic), getTrimmedText(gist))
	}

	for _, x := range topics {
		topic := x
		context := contextMap[topic]
		history := historyMap[topic]
		if history == nil {
			history = &ChatHistory{}
			historyMap[topic] = history
		}
		go listenTopic(strings.TrimSpace(topic), context, history)
	}

	select {}
}
