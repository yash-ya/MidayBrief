package slack

import (
	"MidayBrief/db"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
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

	// Handle event callback
	var event SlackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Bad event format", http.StatusBadRequest)
		return
	}
	// Check for direct messages to the bot
	if event.Event.Type == "message" && event.Event.ChannelType == "im" {
		if strings.HasPrefix(event.Event.Text, "config <#") {
			handleChannelConfig(event);
		} else {
			fmt.Printf("New DM from user %s: %s\n", event.Event.User, event.Event.Text)
			postToStandUpsChannel(event.TeamID, event.Event.User, event.Event.Text)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handleChannelConfig(event SlackEvent) {
	re := regexp.MustCompile(`<#(C\w+)\|?[^>]*>`)
	matches := re.FindStringSubmatch(event.Event.Text)
	if len(matches) < 2 {
		sendDM(event.TeamID, event.Event.Channel, "Couldn't find a valid channel reference. Try: `config #standups` (use autocomplete).")
		return
	}
	channelID := matches[1]

	if err := db.UpdateChannelID(event.TeamID, channelID); err != nil {
		log.Println("Failed to update channelID in team config:", err)
	} else {
		log.Println("ChannelID is updated in team config for:", event.TeamID)
	}
	sendDM(event.TeamID, event.Event.Channel, fmt.Sprintf("Got it! I'll post your updates to <#%s>", channelID))
}