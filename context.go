package main

import (
	"fmt"
	"io"
	"net/http"
)

func getContextText(topic, gistID string) string {
	url := fmt.Sprintf("https://gist.github.com/akhandagrwl-29/%s/raw", gistID)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("HTTP error while getting response from gist: %s for topic: %s and error:%v\n", getTrimmedText(gistID), getTrimmedText(topic), err)
		return ""
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)
	fmt.Println("Headers:", resp.Header)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read data error from gist: ", err)
		return ""
	}
	//fmt.Printf("Body for topic: %s is: %s\n", topic, string(body))
	return string(body)
}
