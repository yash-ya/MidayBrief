package api

import (
	"MidayBrief/db"
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
	if strings.Contains(text, "config") || strings.Contains(text, "post time") || strings.Contains(text, "timezone") {
		handleCombinedConfig(event)
	} else {
		log.Printf("New DM from user %s: %s", event.Event.User, event.Event.Text)
		if err := db.SaveUserMessage(event.TeamID, event.Event.User, event.Event.Text); err != nil {
			log.Printf("Failed to save user message: %v", err)
		} else {
			log.Printf("User message saved for team %s, user %s", event.TeamID, event.Event.User)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handleCombinedConfig(event SlackEvent) {
	text := event.Event.Text

	reChan := regexp.MustCompile(`<#(C\w+)\|?[^>]*>`)
	reTime := regexp.MustCompile(`post time (\d{2}:\d{2})`)
	reZone := regexp.MustCompile(`timezone ([A-Za-z]+/[A-Za-z_]+)`)

	channelMatch := reChan.FindStringSubmatch(text)
	timeMatch := reTime.FindStringSubmatch(text)
	zoneMatch := reZone.FindStringSubmatch(text)

	var updates []string

	if len(channelMatch) == 2 {
		if err := db.UpdateChannelID(event.TeamID, channelMatch[1]); err == nil {
			updates = append(updates, "channel")
		} else {
			log.Printf("Failed to update channel for team %s: %v", event.TeamID, err)
		}
	}

	if len(timeMatch) == 2 {
		if _, err := time.Parse("15:04", timeMatch[1]); err == nil {
			if err := db.UpdatePostTime(event.TeamID, timeMatch[1]); err == nil {
				updates = append(updates, "post time")
			} else {
				log.Printf("Failed to update post time for team %s: %v", event.TeamID, err)
			}
		}
	}

	if len(zoneMatch) == 2 {
		if _, err := time.LoadLocation(zoneMatch[1]); err == nil {
			if err := db.UpdateTimezone(event.TeamID, zoneMatch[1]); err == nil {
				updates = append(updates, "timezone")
			} else {
				log.Printf("Failed to update timezone for team %s: %v", event.TeamID, err)
			}
		}
	}

	if len(updates) > 0 {
		msg := fmt.Sprintf("Updated: %s", strings.Join(updates, ", "))
		sendDM(event.TeamID, event.Event.Channel, msg)
	} else {
		sendDM(event.TeamID, event.Event.Channel,
			"No valid configuration found. Try: `config #channel`, `post time 17:00`, or `timezone Asia/Kolkata`.")
	}
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