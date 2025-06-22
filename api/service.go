package api

import (
	"MidayBrief/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func FireAndForgetDM(accessToken, userChannel, message string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic in DM goroutine:", r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := SendMessageWithContext(ctx, accessToken, userChannel, message); err != nil {
			log.Printf("Failed to send DM to %s: %v", userChannel, err)
		}
	}()
}

func SendMessageWithContext(ctx context.Context, accessToken, userChannel, message string) error {
	decryptedAccessToken, err := utils.Decrypt(accessToken)
	if err != nil {
		log.Printf("SendMessage: Error in decrypt access token %s", err)
		decryptedAccessToken = accessToken
	}

	payload := map[string]string{
		"channel": userChannel,
		"text":    message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", slackPostMessagesURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", decryptedAccessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status: %s", resp.Status)
	}

	return nil
}

func SendMessage(accessToken, userChannel, message string) error {

	decryptedAccessToken, err := utils.Decrypt(accessToken)
	if err != nil {
		log.Printf("SendMessage: Error in decrypt access token %s", err)
		decryptedAccessToken = accessToken
	}

	payload := map[string]string{
		"channel": userChannel,
		"text":    message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("SendMessage: failed to marshal message: %w", err)
	}
	req, err := http.NewRequest("POST", slackPostMessagesURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("SendMessage: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", decryptedAccessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("SendMessage: failed to send request to Slack API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SendMessage: Slack API responded with status %s", resp.Status)
	}

	return nil
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

func getAllTeamUsers(token string) ([]string, error) {
	req, err := http.NewRequest("GET", slackUsersListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to Slack failed: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		OK      bool `json:"ok"`
		Members []struct {
			ID      string `json:"id"`
			IsBot   bool   `json:"is_bot"`
			Name    string `json:"name"`
			Deleted bool   `json:"deleted"`
		} `json:"members"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	if !result.OK {
		return nil, fmt.Errorf("slack api returned not OK")
	}

	var userIDs []string
	for _, member := range result.Members {
		if !member.IsBot && !member.Deleted && member.Name != "slackbot" {
			userIDs = append(userIDs, member.ID)
		}
	}
	return userIDs, nil
}
