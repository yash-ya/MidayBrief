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

	if event.Event.Type != "message" || event.Event.ChannelType != "im" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch {
	case strings.HasPrefix(event.Event.Text, "config <#"):
		handleChannelConfig(event)
	case strings.HasPrefix(event.Event.Text, "post time"):
		handlePostTime(event)
	case team.BotUserID != event.Event.User:
		fmt.Printf("New DM from user %s: %s\n", event.Event.User, event.Event.Text)
		postToStandUpsChannel(event.TeamID, event.Event.User, event.Event.Text)
	}

	w.WriteHeader(http.StatusOK)
}

func handlePostTime(event SlackEvent) {
	postTime := extractValueAfterCommand(event.Event.Text, "post time")
	if postTime == "" {
		sendDM(event.TeamID, event.Event.Channel, "Please provide time like: post time 17:00")
		return
	}

	if _, err := time.Parse("15:04", postTime); err != nil {
		sendDM(event.TeamID, event.Event.Channel, "Invalid time format. Use 24-hr format like 17:00 (UTC).")
		return
	}

	if err := db.UpdatePostTime(event.TeamID, postTime); err != nil {
		handleConfigError("post time", event.TeamID, err, event)
		return
	}

	sendDM(event.TeamID, event.Event.Channel, fmt.Sprintf("Got it! I'll post your team's updates daily at %s.", postTime))
}

func handleChannelConfig(event SlackEvent) {
	channelID := extractChannelID(event.Event.Text)
	if channelID == "" {
		sendDM(event.TeamID, event.Event.Channel, "Couldn't find a valid channel reference. Try: config #standups (use autocomplete).")
		return
	}

	if err := db.UpdateChannelID(event.TeamID, channelID); err != nil {
		handleConfigError("channel ID", event.TeamID, err, event)
		return
	}

	sendDM(event.TeamID, event.Event.Channel, fmt.Sprintf("Got it! I'll post your updates to <#%s>", channelID))
}

func extractValueAfterCommand(text, cmd string) string {
	text = strings.TrimSpace(text)
	parts := strings.SplitN(text, cmd, 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func extractChannelID(text string) string {
	re := regexp.MustCompile(`<#(C\w+)\|?[^>]*>`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func handleConfigError(field, teamID string, err error, event SlackEvent) {
	log.Printf("Failed to update %s for team %s: %v", field, teamID, err)
	sendDM(event.TeamID, event.Event.Channel, fmt.Sprintf("An error occurred while updating your %s.", field))
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
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", slackPostMessages, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Slack API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack api returned non-200: %s", resp.Status)
	}

	return nil
}
