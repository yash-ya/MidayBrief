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
		http.Error(w, "Unable to read request", http.StatusBadRequest)
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
		http.Error(w, "Bad event format", http.StatusBadRequest)
		return
	}

	team, err := db.GetTeamConfig(event.TeamID)

	if err != nil {
		http.Error(w, "Team not configured", http.StatusBadRequest)
		return
	}

	if event.Event.Type == "message" && event.Event.ChannelType == "im" {
		if strings.HasPrefix(event.Event.Text, "config <#") {
			handleChannelConfig(event)
		} else if strings.HasPrefix(event.Event.Text, "post time"){
			handlePostTime(event)
		} else if team.BotUserID != event.Event.User {
			fmt.Printf("New DM from user %s: %s\n", event.Event.User, event.Event.Text)
			postToStandUpsChannel(event.TeamID, event.Event.User, event.Event.Text)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handlePostTime(event SlackEvent) {
	postTime := extractValueAfterCommand(event.Event.Text, "post time")
	if postTime == "" {
		sendDM(event.TeamID, event.Event.Channel, "Please provide time like: `post time 17:00`")
		return
	}

	if _, err := time.Parse("15:04", postTime); err != nil {
		sendDM(event.TeamID, event.Event.Channel, "Invalid time format. Use 24-hr format like `17:00` (UTC).")
		return
	}

	if err := db.UpdatePostTime(event.TeamID, postTime); err != nil {
		handleConfigError("post time", event.TeamID, err, event)
		return
	}

	sendDM(event.TeamID, event.Event.Channel,
		fmt.Sprintf("Got it! I'll post your team's updates daily at `%s`.", postTime))
}

func handleChannelConfig(event SlackEvent) {
	channelID := extractChannelID(event.Event.Text)
	if channelID == "" {
		sendDM(event.TeamID, event.Event.Channel,
			"Couldn't find a valid channel reference. Try: `config #standups` (use autocomplete).")
		return
	}

	if err := db.UpdateChannelID(event.TeamID, channelID); err != nil {
		handleConfigError("channel ID", event.TeamID, err, event)
		return
	}

	sendDM(event.TeamID, event.Event.Channel,
		fmt.Sprintf("Got it! I'll post your updates to <#%s>", channelID))
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
	log.Printf("Failed to update %s in team config for %s: %v", field, teamID, err)
	sendDM(event.TeamID, event.Event.Channel, fmt.Sprintf("Something went wrong while updating your %s.", field))
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
	body, _ := json.Marshal(msg)

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned non-200: %s", resp.Status)
	}

	return nil
}