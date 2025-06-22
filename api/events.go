package api

import (
	"MidayBrief/db"
	"MidayBrief/utils"
	"context"
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

	// Check if user is in the middle of a prompt flow
	ctx := context.Background()
	if state, err := utils.GetPromptState(team.TeamID, event.Event.User, ctx); err == nil && state != nil {
		handlePromptStep(event, team, *state, ctx)
		w.WriteHeader(http.StatusOK)
		return
	}

	if isConfig(event.Event.Text) {
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
		SendMessage(team.AccessToken, event.Event.Channel, "Looks like you've already sent this update today.")
		return
	}

	encryptedMessage, _ := utils.Encrypt(event.Event.Text)
	if err := db.SaveUserMessage(event.TeamID, event.Event.User, encryptedMessage); err != nil {
		log.Printf("Failed to save user message: %v", err)
	} else {
		log.Printf("User message saved for team %s, user %s", event.TeamID, event.Event.User)
		SendMessage(team.AccessToken, event.Event.Channel, "Got your update for today!")
	}
}

func handleCombinedConfig(event SlackEvent, team *db.TeamConfig) {
	if event.Event.User != team.AdminUserID {
		SendMessage(team.AccessToken, event.Event.Channel, "Only the admin can update team settings.")
		return
	}

	text := event.Event.Text
	var updates, errors []string

	if channelID := extractChannelID(text); channelID != "" {
		if err := db.UpdateChannelID(team.TeamID, channelID); err == nil {
			updates = append(updates, fmt.Sprintf("channel updated to %s#", channelID))
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
			errors = append(errors, fmt.Sprintf("Failed to fetch user list for adding. Error - %s", err))
		} else {
			count := 0
			for _, userID := range users {
				if err := db.AddPromptUser(team.TeamID, userID); err == nil {
					count++
					FireAndForgetDM(team.AccessToken, userID, slackUserWelcomeMessage)
				}
			}
			updates = append(updates, fmt.Sprintf("added %d users for prompts", count))
		}
	}

	if strings.HasPrefix(strings.ToLower(text), "add user ") {
		addUsers := extractUserIDs(text)
		for _, userID := range addUsers {

			if err := db.AddPromptUser(team.TeamID, userID); err == nil {
				updates = append(updates, fmt.Sprintf("added <@%s>", userID))
				FireAndForgetDM(team.AccessToken, userID, slackUserWelcomeMessage)
			} else {
				errors = append(errors, fmt.Sprintf("Failed to add <@%s>", userID))
			}
		}
	}

	if strings.HasPrefix(strings.ToLower(text), "remove user ") {
		removeUsers := extractUserIDs(text)
		for _, userID := range removeUsers {
			if err := db.RemovePromptUser(team.TeamID, userID); err == nil {
				updates = append(updates, fmt.Sprintf("removed <@%s>", userID))
			} else {
				errors = append(errors, fmt.Sprintf("Failed to remove <@%s>", userID))
			}
		}
	}

	var response strings.Builder
	if len(updates) > 0 {
		response.WriteString("✅ Updates:\n")
		for _, u := range updates {
			response.WriteString("\t• " + u + "\n")
		}
	}
	if len(errors) > 0 {
		response.WriteString("\n⚠️ Issues:\n")
		for _, e := range errors {
			response.WriteString("\t• " + e + "\n")
		}
	}
	if response.Len() == 0 {
		response.WriteString("No valid configuration found.\nTry: `config #channel`, `post time 17:00`, `timezone Asia/Kolkata`, `add all`, `add/remove @user`.")
	}

	SendMessage(team.AccessToken, event.Event.Channel, response.String())
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

func handlePromptStep(event SlackEvent, team *db.TeamConfig, state utils.PromptState, ctx context.Context) {
	userID := event.Event.User
	teamID := team.TeamID
	accessToken := team.AccessToken
	text := strings.TrimSpace(event.Event.Text)

	switch state.Step {
	case 1:
		state.Responses["yesterday"] = text
		state.Step = 2
		utils.SetPromptState(teamID, userID, state, ctx)
		SendMessage(accessToken, userID, "Got it! What are your plans for today?")
	case 2:
		state.Responses["today"] = text
		state.Step = 3
		utils.SetPromptState(teamID, userID, state, ctx)
		SendMessage(accessToken, userID, "Thanks! Do you have any blockers?")
	case 3:
		state.Responses["blockers"] = text
		saveFinalPrompt(teamID, userID, state)
		utils.DeletePromptState(teamID, userID, ctx)
		SendMessage(accessToken, userID, "All set! Your standup update has been recorded.")
	default:
		utils.DeletePromptState(teamID, userID, ctx)
		SendMessage(accessToken, userID, "Unexpected error. Prompt session cleared. Please try again.")
	}
}

func saveFinalPrompt(teamID, userID string, state utils.PromptState) {
	jsonResponse, err := json.Marshal(state.Responses)
	if err != nil {
		log.Printf("Failed to marshal prompt responses: %v", err)
		return
	}

	encrypted, err := utils.Encrypt(string(jsonResponse))
	if err != nil {
		log.Printf("Failed to encrypt prompt response: %v", err)
		return
	}

	if err := db.SaveUserMessage(teamID, userID, encrypted); err != nil {
		log.Printf("Failed to save final prompt message: %v", err)
	}
}
