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
		SendMessage(team.AccessToken, event.Event.Channel, "Sorry, only the team admin can change these settings. If you need something updated, please reach out to them!")
		log.Printf("[INFO] Unauthorized config attempt by user %s in team %s\n", event.Event.User, team.TeamID)
		return
	}

	text := event.Event.Text
	var updates, errors []string

	handleChannelConfig(team, text, &updates, &errors)
	handlePostTimeConfig(team, text, &updates, &errors)
	handleTimezoneConfig(team, text, &updates, &errors)
	handlePromptTimeConfig(team, text, &updates, &errors)
	handleAddAllUsersConfig(team, text, &updates, &errors)
	handleAddUserConfig(team, text, &updates, &errors)
	handleRemoveUserConfig(team, text, &updates, &errors)

	sendCombinedConfigResponse(event.Event.Channel, team.AccessToken, updates, errors)
}

func handleChannelConfig(team *db.TeamConfig, text string, updates, errors *[]string) {
	if channelID := extractChannelID(text); channelID != "" {
		if err := db.UpdateChannelID(team.TeamID, channelID); err == nil {
			*updates = append(*updates, fmt.Sprintf("channel set to <#%s>", channelID)) // Use channel mention
			log.Printf("[INFO] Channel ID updated for team %s to %s\n", team.TeamID, channelID)
		} else {
			*errors = append(*errors, "Couldn't update the channel. Please try again.")
			log.Printf("[ERROR] Failed to update channel for team %s: %v\n", team.TeamID, err)
		}
	}
}

func handlePostTimeConfig(team *db.TeamConfig, text string, updates, errors *[]string) {
	if timeStr := extractValue(text, `post time (\d{2}:\d{2})`); timeStr != "" {
		if _, err := time.Parse("15:04", timeStr); err == nil {
			if err := db.UpdatePostTime(team.TeamID, timeStr); err == nil {
				*updates = append(*updates, fmt.Sprintf("daily summary time set to %s", timeStr))
				log.Printf("[INFO] Post time updated for team %s to %s\n", team.TeamID, timeStr)
			} else {
				*errors = append(*errors, "Failed to set the daily summary time.")
				log.Printf("[ERROR] Failed to update post time for team %s: %v\n", team.TeamID, err)
			}
		} else {
			*errors = append(*errors, "That time format looks a bit off. Please use the 24-hour format, like `post time 17:00`.")
			log.Printf("[ERROR] Invalid post time format by team %s: %s\n", team.TeamID, timeStr)
		}
	}
}

func handleTimezoneConfig(team *db.TeamConfig, text string, updates, errors *[]string) {
	if zone := extractValue(text, `timezone ([A-Za-z]+/[A-Za-z_]+)`); zone != "" {
		if _, err := time.LoadLocation(zone); err == nil {
			if err := db.UpdateTimezone(team.TeamID, zone); err == nil {
				*updates = append(*updates, fmt.Sprintf("timezone set to %s", zone))
				log.Printf("[INFO] Timezone updated for team %s to %s\n", team.TeamID, zone)
			} else {
				*errors = append(*errors, "Couldn't update your team's timezone.")
				log.Printf("[ERROR] Failed to update timezone for team %s: %v\n", team.TeamID, err)
			}
		} else {
			*errors = append(*errors, fmt.Sprintf("Hmm, '%s' doesn't seem to be a valid timezone. Try something like `timezone Asia/Kolkata`.", zone))
			log.Printf("[ERROR] Invalid timezone value for team %s: %s\n", team.TeamID, zone)
		}
	}
}

func handlePromptTimeConfig(team *db.TeamConfig, text string, updates, errors *[]string) {
	if promptTime := extractValue(text, `prompt time (\d{2}:\d{2})`); promptTime != "" {
		if _, err := time.Parse("15:04", promptTime); err == nil {
			if err := db.UpdatePromptTime(team.TeamID, promptTime); err == nil {
				*updates = append(*updates, fmt.Sprintf("prompt time set to %s", promptTime))
				log.Printf("[INFO] Prompt time updated for team %s to %s\n", team.TeamID, promptTime)
			} else {
				*errors = append(*errors, "Failed to set the prompt time.")
				log.Printf("[ERROR] Failed to update prompt time for team %s: %v\n", team.TeamID, err)
			}
		} else {
			*errors = append(*errors, "That time format looks a bit off for prompts. Please use the 24-hour format, like `prompt time 10:00`.")
			log.Printf("[ERROR] Invalid prompt time format by team %s: %s\n", team.TeamID, promptTime)
		}
	}
}

func handleAddAllUsersConfig(team *db.TeamConfig, text string, updates, errors *[]string) {
	if strings.Contains(strings.TrimSpace(strings.ToLower(text)), "add all users") {
		users, err := getAllTeamUsers(team.AccessToken)
		if err != nil {
			*errors = append(*errors, "I couldn’t get the list of users from Slack. Please check my permissions.")
			log.Printf("[ERROR] Failed to fetch Slack users for team %s: %v\n", team.TeamID, err)
			return
		}

		count := 0
		for _, userID := range users {
			if !db.IsPromptUserExist(team.TeamID, userID) {
				if err := db.AddPromptUser(team.TeamID, userID); err == nil {
					count++
					if team.AdminUserID != userID {
						go FireAndForgetDM(team.AccessToken, userID, slackUserWelcomeMessage)
					}
				} else {
					log.Printf("[ERROR] Failed to add user %s to team %s: %v\n", userID, team.TeamID, err)
				}
			}
		}

		if count == 0 {
			*errors = append(*errors, "No new users were added because everyone is already on the prompt list!")
			log.Printf("[INFO] No new users added for team %s; all were already configured\n", team.TeamID)
		} else {
			*updates = append(*updates, fmt.Sprintf("added *%d new users* to the prompt list", count))
			log.Printf("[INFO] Successfully added %d new users for team %s\n", count, team.TeamID)
		}
	}
}

func handleAddUserConfig(team *db.TeamConfig, text string, updates, errors *[]string) {
	if strings.HasPrefix(strings.ToLower(text), "add user ") {
		addUsers := extractUserIDs(text)
		if len(addUsers) == 0 {
			*errors = append(*errors, "Please mention at least one user to add, like `add user @alice`.")
			return
		}
		for _, userID := range addUsers {
			if !db.IsPromptUserExist(team.TeamID, userID) {
				if err := db.AddPromptUser(team.TeamID, userID); err == nil {
					*updates = append(*updates, fmt.Sprintf("added <@%s> to the prompt list", userID))
					if team.AdminUserID != userID {
						go FireAndForgetDM(team.AccessToken, userID, slackUserWelcomeMessage)
					}
					log.Printf("[INFO] Added user %s to team %s\n", userID, team.TeamID)
				} else {
					*errors = append(*errors, fmt.Sprintf("Couldn’t add <@%s> right now due to an issue.", userID))
					log.Printf("[ERROR] Failed to add user %s to team %s: %v\n", userID, team.TeamID, err)
				}
			} else {
				*errors = append(*errors, fmt.Sprintf("<@%s> is already scheduled for daily stand-ups.", userID))
				log.Printf("[INFO] Skipped adding existing prompt user %s to team %s\n", userID, team.TeamID)
			}
		}
	}
}

func handleRemoveUserConfig(team *db.TeamConfig, text string, updates, errors *[]string) {
	if strings.HasPrefix(strings.ToLower(text), "remove user ") {
		removeUsers := extractUserIDs(text)
		if len(removeUsers) == 0 {
			*errors = append(*errors, "Please mention at least one user to remove, like `remove user @bob`.")
			return
		}
		for _, userID := range removeUsers {
			if db.IsPromptUserExist(team.TeamID, userID) {
				if err := db.RemovePromptUser(team.TeamID, userID); err == nil {
					*updates = append(*updates, fmt.Sprintf("removed <@%s> from the prompt list", userID))
					log.Printf("[INFO] Removed user %s from team %s\n", userID, team.TeamID)
				} else {
					*errors = append(*errors, fmt.Sprintf("Couldn’t remove <@%s> right now due to an issue.", userID))
					log.Printf("[ERROR] Failed to remove user %s from team %s: %v\n", userID, team.TeamID, err)
				}
			} else {
				*errors = append(*errors, fmt.Sprintf("<@%s> isn't currently receiving daily stand-up reminders.", userID))
				log.Printf("[INFO] Skipped removal; user %s not found in prompt list for team %s\n", userID, team.TeamID)
			}
		}
	}
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

func sendCombinedConfigResponse(channel, token string, updates, errors []string) {
	var response strings.Builder

	if len(updates) > 0 {
		response.WriteString("✅ *Success! Here's what I've updated for you:*\n")
		for _, u := range updates {
			response.WriteString("• " + u + "\n")
		}
	}

	if len(errors) > 0 {
		if len(updates) > 0 {
			response.WriteString("\n")
		}
		response.WriteString("⚠️ *Just a heads-up, I ran into a couple of snags:*\n")
		for _, e := range errors {
			response.WriteString("• " + e + "\n")
		}
	}

	if response.Len() == 0 {
		response.WriteString(unrecognizedCommandMessage)
	}

	SendMessage(token, channel, response.String())
}
