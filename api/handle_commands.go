package api

import (
	"MidayBrief/db"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

func handleCombinedConfig(event SlackEvent, team *db.TeamConfig) {
	if event.Event.User != team.AdminUserID {
		SendMessage(team.AccessToken, event.Event.Channel, "Only the admin can update team settings.")
		log.Printf("[INFO] Unauthorized config attempt by user %s in team %s\n", event.Event.User, team.TeamID)
		return
	}

	text := event.Event.Text
	var updates, errors []string

	if channelID := extractChannelID(text); channelID != "" {
		if err := db.UpdateChannelID(team.TeamID, channelID); err == nil {
			updates = append(updates, fmt.Sprintf("channel updated to %s#", channelID))
			log.Printf("[INFO] Channel ID updated for team %s\n", team.TeamID)
		} else {
			errors = append(errors, "Failed to update channel.")
			log.Printf("[ERROR] Failed to update channel for team %s: %v\n", team.TeamID, err)
		}
	}

	if timeStr := extractValue(text, `post time (\d{2}:\d{2})`); timeStr != "" {
		if _, err := time.Parse("15:04", timeStr); err == nil {
			if err := db.UpdatePostTime(team.TeamID, timeStr); err == nil {
				updates = append(updates, fmt.Sprintf("post time updated to %s", timeStr))
				log.Printf("[INFO] Post time updated for team %s\n", team.TeamID)
			} else {
				errors = append(errors, "Failed to update post time.")
				log.Printf("[ERROR] Failed to update post time for team %s: %v\n", team.TeamID, err)
			}
		} else {
			errors = append(errors, "Invalid time format. Use 24-hr format like: post time 17:00.")
			log.Printf("[ERROR] Invalid post time format by team %s: %s\n", team.TeamID, timeStr)
		}
	}

	if zone := extractValue(text, `timezone ([A-Za-z]+/[A-Za-z_]+)`); zone != "" {
		if _, err := time.LoadLocation(zone); err == nil {
			if err := db.UpdateTimezone(team.TeamID, zone); err == nil {
				updates = append(updates, fmt.Sprintf("timezone updated to %s", zone))
				log.Printf("[INFO] Timezone updated for team %s\n", team.TeamID)
			} else {
				errors = append(errors, "Failed to update timezone.")
				log.Printf("[ERROR] Failed to update timezone for team %s: %v\n", team.TeamID, err)
			}
		} else {
			errors = append(errors, fmt.Sprintf("Invalid timezone: '%s'. Use format like: timezone Asia/Kolkata.", zone))
			log.Printf("[ERROR] Invalid timezone value for team %s: %s\n", team.TeamID, zone)
		}
	}

	if promptTime := extractValue(text, `prompt time (\d{2}:\d{2})`); promptTime != "" {
		if _, err := time.Parse("15:04", promptTime); err == nil {
			if err := db.UpdatePromptTime(team.TeamID, promptTime); err == nil {
				updates = append(updates, fmt.Sprintf("prompt time updated to %s", promptTime))
				log.Printf("[INFO] Prompt time updated for team %s\n", team.TeamID)
			} else {
				errors = append(errors, "Failed to update prompt time.")
				log.Printf("[ERROR] Failed to update prompt time for team %s: %v\n", team.TeamID, err)
			}
		} else {
			errors = append(errors, "Invalid time format. Use 24-hr format like: prompt time 10:00.")
			log.Printf("[ERROR] Invalid prompt time format by team %s: %s\n", team.TeamID, promptTime)
		}
	}

	if strings.Contains(strings.TrimSpace(strings.ToLower(text)), "add all users") {
		users, err := getAllTeamUsers(team.AccessToken)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to fetch user list for adding. Error - %s", err))
			log.Printf("[ERROR] Failed to fetch Slack users for team %s: %v\n", team.TeamID, err)
		} else {
			count := 0
			for _, userID := range users {
				if err := db.AddPromptUser(team.TeamID, userID); err == nil {
					count++
					go FireAndForgetDM(team.AccessToken, userID, slackUserWelcomeMessage)
				}
			}
			updates = append(updates, fmt.Sprintf("added %d users for prompts", count))
			log.Printf("[INFO] Added %d users for team %s\n", count, team.TeamID)
		}
	}

	if strings.HasPrefix(strings.ToLower(text), "add user ") {
		addUsers := extractUserIDs(text)
		for _, userID := range addUsers {
			if err := db.AddPromptUser(team.TeamID, userID); err == nil {
				updates = append(updates, fmt.Sprintf("added <@%s>", userID))
				go FireAndForgetDM(team.AccessToken, userID, slackUserWelcomeMessage)
				log.Printf("[INFO] Added user %s to team %s\n", userID, team.TeamID)
			} else {
				errors = append(errors, fmt.Sprintf("Failed to add <@%s>", userID))
				log.Printf("[ERROR] Failed to add user %s to team %s: %v\n", userID, team.TeamID, err)
			}
		}
	}

	if strings.HasPrefix(strings.ToLower(text), "remove user ") {
		removeUsers := extractUserIDs(text)
		for _, userID := range removeUsers {
			if err := db.RemovePromptUser(team.TeamID, userID); err == nil {
				updates = append(updates, fmt.Sprintf("removed <@%s>", userID))
				log.Printf("[INFO] Removed user %s from team %s\n", userID, team.TeamID)
			} else {
				errors = append(errors, fmt.Sprintf("Failed to remove <@%s>", userID))
				log.Printf("[ERROR] Failed to remove user %s from team %s: %v\n", userID, team.TeamID, err)
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
		log.Printf("[INFO] No valid configuration commands found for team %s\n", team.TeamID)
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
