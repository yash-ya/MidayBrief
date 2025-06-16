package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func sendDM(userChannelID, message string) {
	var url = "https://slack.com/api/chat.postMessage"
	botToken := os.Getenv("SLACK_BOT_TOKEN")

	payload := map[string]string{
		"channel": userChannelID,
		"text":    message,
	}

	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+botToken)

	client := &http.Client{}
	client.Do(req)
}

func postToStandUpsChannel(userID, message string) {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	channelID := userChannelMap[userID]
	if channelID == "" {
		channelID = "C091DGFPE6S" // fallback if not set
	}

	payload := map[string]string{

		"channel": channelID,
		"text":    fmt.Sprintf("*Update from <@%s>*:\n%s", userID, message),
	}
	jsonBody, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+botToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failed to post message:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Slack API responded with:", resp.Status)
	}
}