package api

import (
	"MidayBrief/db"
	"MidayBrief/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func sendDM(teamID, userChannelID, message string) {
	team, err := db.GetTeamConfig(teamID)
	if err != nil {
		log.Printf("sendDM: failed to get team config for %s: %v", teamID, err)
		return
	}

	payload := map[string]string{
		"channel": userChannelID,
		"text":    message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("sendDM: failed to marshal payload: %v", err)
		return
	}

	accessToken, _ := utils.Decrypt(team.AccessToken)
	req, err := http.NewRequest("POST", slackPostMessagesURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("sendDM: failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("sendDM: failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("sendDM: Slack API responded with status %s", resp.Status)
	}
}

func getUserTimeZone(accessToken, userID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?user=%s", slackUserInfoURL, userID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		OK   bool `json:"ok"`
		User struct {
			TZ string `json:"tz"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if !result.OK {
		return "", fmt.Errorf("slack api error: could not get timezone")
	}

	return result.User.TZ, nil
}

// func postToStandUpsChannel(teamID, userID, message string) {
// 	team, err := db.GetTeamConfig(teamID)
// 	if err != nil {
// 		log.Printf("postToStandUpsChannel: failed to get team config for %s: %v", teamID, err)
// 		return
// 	}

// 	if err := db.SaveUserMessage(teamID, userID, message); err != nil {
// 		log.Printf("postToStandUpsChannel: failed to save user message: %v", err)
// 	} else {
// 		log.Printf("User message saved for team %s, user %s", teamID, userID)
// 	}

// 	payload := map[string]string{
// 		"channel": team.ChannelID,
// 		"text":    fmt.Sprintf("Update from <@%s>:\n%s", userID, message),
// 	}

// 	body, err := json.Marshal(payload)
// 	if err != nil {
// 		log.Printf("postToStandUpsChannel: failed to marshal payload: %v", err)
// 		return
// 	}

// 	req, err := http.NewRequest("POST", slackPostMessagesURL, bytes.NewBuffer(body))
// 	if err != nil {
// 		log.Printf("postToStandUpsChannel: failed to create request: %v", err)
// 		return
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+team.AccessToken)

// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Printf("postToStandUpsChannel: failed to send request: %v", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		log.Printf("postToStandUpsChannel: Slack API responded with status %s", resp.Status)
// 	}
// }
