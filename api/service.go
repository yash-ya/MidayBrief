package api

import (
	"MidayBrief/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func FireAndForgetDM(accessToken, userChannel, message string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[WARN] Recovered from panic in DM goroutine: %v\n", r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := SendMessageWithContext(ctx, accessToken, userChannel, message); err != nil {
			log.Printf("[ERROR] Failed to send DM to %s: %v\n", userChannel, err)
		}
	}()
}

func SendMessageWithContext(ctx context.Context, accessToken, userChannel, message string) error {
	decryptedAccessToken, err := utils.Decrypt(accessToken)
	if err != nil {
		log.Printf("[WARN] Failed to decrypt access token, using fallback: %v\n", err)
		decryptedAccessToken = accessToken
	}

	payload := map[string]string{
		"channel": userChannel,
		"text":    message,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", slackPostMessagesURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", decryptedAccessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read SlackMessage response body: %v\n", err)
		return fmt.Errorf("failed to read SlackMessage response body: %v", err)
	}
	defer resp.Body.Close()

	var messageResponse MessageResponse
	if err := json.Unmarshal(responseBody, &messageResponse); err != nil {
		log.Printf("[ERROR] Failed to unmarshal OAuth response: %v\n", err)
	}

	if !messageResponse.Ok {
		log.Printf("[ERROR] Message response returned error: %s\n", messageResponse.Error)
		return fmt.Errorf("message response returned error: %s", messageResponse.Error)
	}

	return nil
}

func SendMessage(accessToken, userChannel, message string) error {
	decryptedAccessToken, err := utils.Decrypt(accessToken)
	if err != nil {
		log.Printf("[WARN] Failed to decrypt access token, using fallback: %v\n", err)
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

	req, err := http.NewRequest("POST", slackPostMessagesURL, bytes.NewBuffer(body))
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

func getUserTimeZone(accessToken, userID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?user=%s", slackUserInfoURL, userID), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("slack request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK   bool `json:"ok"`
		User struct {
			TZ string `json:"tz"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response failed: %w", err)
	}

	if !result.OK {
		return "", fmt.Errorf("slack API error: could not get timezone")
	}

	return result.User.TZ, nil
}

func getAllTeamUsers(accessToken string) ([]string, error) {
	decryptedAccessToken, err := utils.Decrypt(accessToken)
	if err != nil {
		log.Printf("[WARN] Failed to decrypt access token, using fallback: %v\n", err)
		decryptedAccessToken = accessToken
	}
	req, err := http.NewRequest("GET", slackUsersListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+decryptedAccessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("slack request failed: %w", err)
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
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !result.OK {
		return nil, fmt.Errorf("slack API returned not OK")
	}

	var userIDs []string
	for _, member := range result.Members {
		if !member.IsBot && !member.Deleted && member.Name != "slackbot" {
			userIDs = append(userIDs, member.ID)
		}
	}

	log.Printf("[INFO] Retrieved %d valid team users from Slack\n", len(userIDs))
	return userIDs, nil
}
