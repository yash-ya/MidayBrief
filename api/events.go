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

	text := strings.ToLower(event.Event.Text)
	if strings.HasPrefix(text, "config <#") || strings.HasPrefix(text, "post time") || strings.HasPrefix(text, "timezone") {
		handleCombinedConfig(event, team)
	} else {
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

	w.WriteHeader(http.StatusOK)
}

func handleCombinedConfig(event SlackEvent, team *db.TeamConfig) {
	if event.Event.User != team.AdminUserID {
		sendDM(team.TeamID, event.Event.Channel, "Only the admin can update team settings.")
		return
	}

	reChan := regexp.MustCompile(`config <#(C\w+)\|?[^>]*>`)
	reTime := regexp.MustCompile(`post time (\d{2}:\d{2})`)
	reZone := regexp.MustCompile(`timezone ([A-Za-z]+/[A-Za-z_]+)`)

	channelMatch := reChan.FindStringSubmatch(event.Event.Text)
	timeMatch := reTime.FindStringSubmatch(event.Event.Text)
	zoneMatch := reZone.FindStringSubmatch(event.Event.Text)

	var updates []string
	var errors []string

	// Channel
	if len(channelMatch) == 2 {
		if err := db.UpdateChannelID(event.TeamID, channelMatch[1]); err == nil {
			updates = append(updates, "channel")
		} else {
			errors = append(errors, "Failed to update channel.")
		}
	}

	// Post Time
	if len(timeMatch) == 2 {
		if _, err := time.Parse("15:04", timeMatch[1]); err == nil {
			if err := db.UpdatePostTime(event.TeamID, timeMatch[1]); err == nil {
				updates = append(updates, "post time")
			} else {
				errors = append(errors, "Failed to update post time.")
			}
		} else {
			errors = append(errors, "Invalid time format. Use 24-hr format like: post time 17:00.")
		}
	}

	// Timezone
	if len(zoneMatch) == 2 {
		if _, err := time.LoadLocation(zoneMatch[1]); err == nil {
			if err := db.UpdateTimezone(event.TeamID, zoneMatch[1]); err == nil {
				updates = append(updates, "timezone")
			} else {
				errors = append(errors, "Failed to update timezone.")
			}
		} else {
			errors = append(errors, fmt.Sprintf("Invalid timezone: '%s'. Use format like: timezone Asia/Kolkata.", zoneMatch[1]))
		}
	}

	if len(updates) > 0 {
		msg := fmt.Sprintf("Updated: %s", strings.Join(updates, ", "))
		if len(errors) > 0 {
			msg += "\n\n" + strings.Join(errors, "\n")
		}
		sendDM(event.TeamID, event.Event.Channel, msg)
		return
	}

	if len(errors) > 0 {
		sendDM(event.TeamID, event.Event.Channel, strings.Join(errors, "\n"))
		return
	}

	sendDM(event.TeamID, event.Event.Channel,
		"No valid configuration found. Try: config #channel, post time 17:00, or timezone Asia/Kolkata.")
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
