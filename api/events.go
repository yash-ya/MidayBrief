package api

import (
	"MidayBrief/db"
	"MidayBrief/utils"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func HandleSlackEvents(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		log.Printf("[ERROR] Failed to read Slack request body: %v\n", err)
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
		log.Printf("[ERROR] Failed to parse Slack event: %v\n", err)
		return
	}

	if event.Event.Text == "" {
		http.Error(w, "Empty text message", http.StatusBadRequest)
		log.Printf("[WARN] Ignored empty text message\n")
		return
	}

	team, err := db.GetTeamConfig(event.TeamID)
	if err != nil {
		http.Error(w, "Team not configured", http.StatusBadRequest)
		log.Printf("[ERROR] Team config fetch failed for team %s: %v\n", event.TeamID, err)
		return
	}

	if event.Event.Type != "message" || event.Event.ChannelType != "im" || event.Event.User == team.BotUserID {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := context.Background()
	if state, err := utils.GetPromptState(team.TeamID, event.Event.User, ctx); err == nil && state != nil {
		handlePromptStep(event, team, *state, ctx)
		w.WriteHeader(http.StatusOK)
		return
	}

	if isConfig(event.Event.Text) {
		handleCombinedConfig(event, team)
	} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(event.Event.Text)), "update") && !utils.CanUpdateNow(team.PostTime, team.Timezone) {
		SendMessage(team.AccessToken, event.Event.Channel, "You're too close to the posting time. Updates are only allowed until 30 minutes before the summary is posted.")
		return
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
		log.Printf("[ERROR] Failed to save user message: %v\n", err)
	} else {
		log.Printf("[INFO] Saved user message for team %s, user %s\n", event.TeamID, event.Event.User)
		SendMessage(team.AccessToken, event.Event.Channel, "Got your update for today!")
	}
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
		log.Printf("[WARN] Unexpected state step for user %s on team %s. Session cleared.\n", userID, teamID)
	}
}

func saveFinalPrompt(teamID, userID string, state utils.PromptState) {
	jsonResponse, err := json.Marshal(state.Responses)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal prompt responses: %v\n", err)
		return
	}

	encrypted, err := utils.Encrypt(string(jsonResponse))
	if err != nil {
		log.Printf("[ERROR] Failed to encrypt prompt responses: %v\n", err)
		return
	}

	if err := db.SaveUserMessage(teamID, userID, encrypted); err != nil {
		log.Printf("[ERROR] Failed to save encrypted prompt response: %v\n", err)
	}
}
