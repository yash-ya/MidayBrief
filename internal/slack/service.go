package slack

import (
	"MidayBrief/db"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func sendDM(teamID, userChannelID, message string) {
	team, err := db.GetTeamConfig(teamID);
	if err != nil {
		fmt.Println("Not able to find the team in db", err)
		return
	}

	var url = "https://slack.com/api/chat.postMessage"
	payload := map[string]string{
		"channel": userChannelID,
		"text":    message,
	}

	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+team.AccessToken)

	client := &http.Client{}
	client.Do(req)
}

func postToStandUpsChannel(teamID, userID, message string) {
	team, err := db.GetTeamConfig(teamID);
	if err != nil {
		fmt.Println("Not able to find the team in db", err)
		return
	}

	saveErrMessage := db.SaveUserMessage(teamID, userID, message)
	if saveErrMessage != nil {
		fmt.Println("Unable to save user message in DB", saveErrMessage)
	}else {
		fmt.Println("User message is saved in DB")
	}

	payload := map[string]string{

		"channel": team.ChannelID,
		"text":    fmt.Sprintf("*Update from <@%s>*:\n%s", userID, message),
	}
	jsonBody, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+team.AccessToken)

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