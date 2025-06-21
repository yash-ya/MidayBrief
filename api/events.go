package api

import (
	"MidayBrief/db"
	"MidayBrief/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func HandleSlackEvents(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	var verification urlVerification
	if err := json.Unmarshal(body, &verification); err == nil && verification.Type == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(verification.Challenge))
		return
	}

	var event SlackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid Slack event format", http.StatusBadRequest)
		return
	}

	if event.Event.Text == "" {
		http.Error(w, "Empty text message", http.StatusBadRequest)
		return
	}

	team, err := db.GetTeamConfig(event.TeamID)
	if err != nil {
		http.Error(w, "Team not configured", http.StatusBadRequest)
		return
	}

	if event.Event.Type != "message" || event.Event.ChannelType != "im" || event.Event.User == team.BotUserID {
		w.WriteHeader(http.StatusOK)
		return
	}

	text := event.Event.Text
	log.Printf("User message %s", text)
	if isConfig(text) {
		handleCombinedConfig(event, team)
	} else {
		handleUserMessage(event, team)
	}

	w.WriteHeader(http.StatusOK)
}

func isConfig(text string) bool {
	lowered := strings.ToLower(text)
	isConfig := regexp.MustCompile(`config\s+<#(C\w+)\|?[^>]*>`).MatchString(text)
	isPostTime := regexp.MustCompile(`post time\s+\d{2}:\d{2}`).MatchString(lowered)
	isTimezone := regexp.MustCompile(`timezone\s+[A-Za-z]+/[A-Za-z_]+`).MatchString(text)
	isPromptTime := regexp.MustCompile(`prompt time\s+\d{2}:\d{2}`).MatchString(lowered)
	isAddAll := strings.Contains(lowered, "add all users")
	isAddUser := regexp.MustCompile(`add user\s+(<@U[0-9A-Z]+>\s*)+`).MatchString(text)
	isRemoveUser := regexp.MustCompile(`remove user\s+(<@U[0-9A-Z]+>\s*)+`).MatchString(text)

	return isConfig || isPostTime || isTimezone || isPromptTime || isAddAll || isAddUser || isRemoveUser
}

func handleUserMessage(event SlackEvent, team *db.TeamConfig) {
	hash := utils.Hash(event.Event.Text)
	if db.IsDuplicateMessage(event.TeamID, event.Event.User, hash, team.Timezone) {
		sendDM(event.TeamID, event.Event.Channel, "Looks like you've already sent this update today.")
		return
	}

	encryptedMessage, _ := utils.Encrypt(event.Event.Text)
	if err := db.SaveUserMessage(event.TeamID, event.Event.User, encryptedMessage); err != nil {
		log.Printf("Failed to save user message: %v", err)
	} else {
		log.Printf("User message saved for team %s, user %s", event.TeamID, event.Event.User)
		sendDM(event.TeamID, event.Event.Channel, "Got your update for today!")
	}
}

func handleCombinedConfig(event SlackEvent, team *db.TeamConfig) {
	if event.Event.User != team.AdminUserID {
		sendDM(team.TeamID, event.Event.Channel, "Only the admin can update team settings.")
		return
	}

	text := event.Event.Text
	var updates, errors []string

	if channelID := extractChannelID(text); channelID != "" {
		if err := db.UpdateChannelID(team.TeamID, channelID); err == nil {
			updates = append(updates, fmt.Sprintf("channel updated to %s", channelID))
		} else {
			errors = append(errors, "Failed to update channel.")
		}
	}

	if timeStr := extractValue(text, `post time (\d{2}:\d{2})`); timeStr != "" {
		if _, err := time.Parse("15:04", timeStr); err == nil {
			if err := db.UpdatePostTime(team.TeamID, timeStr); err == nil {
				updates = append(updates, fmt.Sprintf("post time updated to %s", timeStr))
			} else {
				errors = append(errors, "Failed to update post time.")
			}
		} else {
			errors = append(errors, "Invalid time format. Use 24-hr format like: post time 17:00.")
		}
	}

	if zone := extractValue(text, `timezone ([A-Za-z]+/[A-Za-z_]+)`); zone != "" {
		if _, err := time.LoadLocation(zone); err == nil {
			if err := db.UpdateTimezone(team.TeamID, zone); err == nil {
				updates = append(updates, fmt.Sprintf("timezone updated to %s", zone))
			} else {
				errors = append(errors, "Failed to update timezone.")
			}
		} else {
			errors = append(errors, fmt.Sprintf("Invalid timezone: '%s'. Use format like: timezone Asia/Kolkata.", zone))
		}
	}

	if promptTime := extractValue(text, `prompt time (\d{2}:\d{2})`); promptTime != "" {
		if _, err := time.Parse("15:04", promptTime); err == nil {
			if err := db.UpdatePromptTime(team.TeamID, promptTime); err == nil {
				updates = append(updates, fmt.Sprintf("prompt time updated to %s", promptTime))
			} else {
				errors = append(errors, "Failed to update prompt time.")
			}
		} else {
			errors = append(errors, "Invalid time format. Use 24-hr format like: prompt time 10:00.")
		}
	}

	if strings.Contains(strings.TrimSpace(strings.ToLower(text)), "add all users") {
		users, err := getAllTeamUsers(team.AccessToken)
		if err != nil {
			errors = append(errors, "Failed to fetch user list for adding.")
		} else {
			count := 0
			for _, userID := range users {
				if err := db.AddPromptUser(team.TeamID, userID); err == nil {
					count++
				}
			}
			updates = append(updates, fmt.Sprintf("added %d users for prompts", count))
		}
	}

	if strings.HasPrefix(strings.ToLower(text), "add users ") {
		addUsers := extractUserIDs(text)
		for _, userID := range addUsers {

			if err := db.AddPromptUser(team.TeamID, userID); err == nil {
				updates = append(updates, fmt.Sprintf("added @%s", userID))
			} else {
				errors = append(errors, fmt.Sprintf("Failed to add @%s", userID))
			}
		}
	}

	if strings.HasPrefix(strings.ToLower(text), "remove users ") {
		removeUsers := extractUserIDs(text)
		for _, userID := range removeUsers {
			if err := db.RemovePromptUser(team.TeamID, userID); err == nil {
				updates = append(updates, fmt.Sprintf("removed @%s", userID))
			} else {
				errors = append(errors, fmt.Sprintf("Failed to remove @%s", userID))
			}
		}
	}

	var response strings.Builder
	if len(updates) > 0 {
		response.WriteString("✅ Updates:\n")
		for _, u := range updates {
			response.WriteString("- " + u + "\n")
		}
	}
	if len(errors) > 0 {
		response.WriteString("\n⚠️ Issues:\n")
		for _, e := range errors {
			response.WriteString("- " + e + "\n")
		}
	}
	if response.Len() == 0 {
		response.WriteString("No valid configuration found.\nTry: `config #channel`, `post time 17:00`, `timezone Asia/Kolkata`, `add all`, `add/remove @user`.")
	}

	sendDM(team.TeamID, event.Event.Channel, response.String())
}

func extractChannelID(text string) string {
	re := regexp.MustCompile(`config\s+<#(C\w+)\|?[^>]*>`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func extractValue(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func extractUserIDs(text string) []string {
	re := regexp.MustCompile(`<@U[0-9A-Z]+>`)
	matches := re.FindAllString(text, -1)
	var users []string
	for _, match := range matches {
		userID := strings.Trim(match, "<@>")
		users = append(users, userID)
	}
	return users
}

type SlackMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func SendMessage(token, channel, text string) error {
	msg := SlackMessage{
		Channel: channel,
		Text:    text,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("SendMessage: failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", slackPostMessagesURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("SendMessage: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
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
