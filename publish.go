package main

import (
	"fmt"
	"net/http"
	"strings"
)

func publishResponse(topic, response string) error {
	_, err := http.Post(
		fmt.Sprintf("https://ntfy.sh/%s", topic),
		"text/plain",
		strings.NewReader(
			fmt.Sprintf("Chotu: %s", response),
		),
	)
	if err != nil {
		fmt.Printf("Error publishing response to topic %s: %v\n", topic, err)
		return err
	}
	return nil
}
