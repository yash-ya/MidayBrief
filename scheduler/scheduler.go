package scheduler

import (
	"MidayBrief/api"
	"MidayBrief/db"
	"MidayBrief/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	promptMessage         = "Good day! 👋\n\nHope you're doing well. Let's kick off your daily standup.\n\n🕐 First up — *What did you work on yesterday?*\nFeel free to share key highlights or any progress you made."
	schedulerInterval     = 1 * time.Minute
	checkExpiredTimeout   = 30 * time.Second // Increased timeout for expired prompt check
	promptSessionDuration = 30 * time.Minute // Example: prompt session expires after 30 minutes if not completed
)

func StartScheduler() {
	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()

	log.Println("[INFO] Scheduler started...")

	for now := range ticker.C {
		processSchedule(now)
	}
}

func processSchedule(now time.Time) {
	teams, err := db.GetAllTeamConfigs()
	if err != nil {
		log.Printf("[ERROR] Failed to fetch team configs: %v", err)
		return
	}

	for _, team := range teams {
		if team.PostTime == "" || team.PromptTime == "" || team.Timezone == "" || team.ChannelID == "" {
			log.Printf("[WARN] Skipping team %s due to incomplete configuration.", team.TeamID)
			continue
		}

		loc, err := time.LoadLocation(team.Timezone)
		if err != nil {
			log.Printf("[WARN] Invalid timezone for team %s: %s - %v", team.TeamID, team.Timezone, err)
			continue
		}

		localTime := now.In(loc)
		formatted := localTime.Format("15:04")

		if formatted == team.PromptTime {
			log.Printf("[INFO] Triggering prompt for team %s at %s (%s)", team.TeamID, localTime.Format("15:04:05"), team.Timezone)
			go triggerPromptForTeam(team)
		}

		if formatted == team.PostTime {
			log.Printf("[INFO] Triggering post summary for team %s at %s (%s)", team.TeamID, localTime.Format("15:04:05"), team.Timezone)
			go postSummaryForTeam(team, loc)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), checkExpiredTimeout)
	defer cancel()
	checkExpiredPrompts(ctx)
}

func checkExpiredPrompts(ctx context.Context) {
	var cursor uint64
	for {
		keys, newCursor, err := utils.RedisClient.Scan(ctx, cursor, "prompt_expiry:*", 100).Result()
		if err != nil {
			log.Printf("[ERROR] Redis SCAN failed for expired prompts: %v", err)
			break
		}
		cursor = newCursor

		for _, key := range keys {
			ttl, err := utils.RedisClient.TTL(ctx, key).Result()
			if err != nil {
				log.Printf("[ERROR] Failed to get TTL for key %s during expiry check: %v", key, err)
				continue
			}

			if ttl > 0 && ttl != -1 {
				continue
			}

			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				log.Printf("[WARN] Invalid prompt expiry key format: %s", key)
				_ = utils.RedisClient.Del(ctx, key).Err()
				continue
			}
			teamID, userID := parts[1], parts[2]

			state, err := utils.GetPromptState(teamID, userID, ctx)
			if err != nil {
				log.Printf("[INFO] Stale prompt expiry key found for user %s in team %s. Deleting.", userID, teamID)
				_ = utils.RedisClient.Del(ctx, key).Err()
				continue
			}

			if state.Step < 4 {
				team, err := db.GetTeamConfig(teamID)
				if err != nil {
					log.Printf("[ERROR] Team config fetch failed for team %s: %v\n", teamID, err)
				}
				log.Printf("[INFO] Prompt session expired for user %s in team %s (Step: %d). Notifying user and cleaning up.", userID, teamID, state.Step)
				api.SendMessage(team.AccessToken, userID, "⏰ Your prompt session expired. To submit your update, reply with `update` again.")
			} else {
				log.Printf("[INFO] Prompt session for user %s in team %s completed (Step: %d), but expiry key was still present and expired. Cleaning up.", userID, teamID, state.Step)
			}
			if err := utils.RedisClient.Del(ctx, key).Err(); err != nil {
				log.Printf("[ERROR] Failed to delete expired prompt key %s: %v", key, err)
			}
			if err := utils.DeletePromptState(teamID, userID, ctx); err != nil {
				log.Printf("[ERROR] Failed to delete prompt state for user %s in team %s: %v", userID, teamID, err)
			}
		}

		if cursor == 0 {
			break
		}
	}
}

func triggerPromptForTeam(team db.TeamConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := db.GetAllPromptUser(team.TeamID)
	if err != nil {
		log.Printf("[ERROR] Failed to get prompt users for team %s: %v", team.TeamID, err)
		return
	}

	for _, user := range users {
		state := utils.PromptState{
			Step:      1,
			Responses: make(map[string]string),
			StartedAt: time.Now(),
		}

		if err := utils.SetPromptState(team.TeamID, user.UserID, state, ctx); err != nil {
			log.Printf("[ERROR] Failed to set prompt state for user %s in team %s: %v", user.UserID, team.TeamID, err)
			continue
		}

		if err := utils.SetPromptExpiry(team.TeamID, user.UserID, promptSessionDuration, ctx); err != nil {
			log.Printf("[ERROR] Failed to set prompt expiry for user %s in team %s: %v", user.UserID, team.TeamID, err)
			continue
		}

		if err := api.SendMessage(team.AccessToken, user.UserID, promptMessage); err != nil {
			log.Printf("[ERROR] Failed to send prompt to user %s in team %s: %v", user.UserID, team.TeamID, err)
		}
	}
}

func postSummaryForTeam(team db.TeamConfig, location *time.Location) {
	if team.AccessToken == "" || team.ChannelID == "" {
		log.Printf("[ERROR] Missing access token or channel ID for team %s. Cannot post summary.", team.TeamID)
		return
	}

	messages, err := db.GetMessagesForTeamToday(team.TeamID, location)
	if err != nil {
		log.Printf("[ERROR] Error fetching messages for team %s: %v", team.TeamID, err)
		return
	}

	if len(messages) == 0 {
		log.Printf("[INFO] No messages found for team %s to summarize.", team.TeamID)
		return
	}

	summary := formatSummary(messages)
	if err := api.SendMessage(team.AccessToken, team.ChannelID, summary); err != nil {
		log.Printf("[ERROR] Failed to post summary for team %s to channel %s: %v", team.TeamID, team.ChannelID, err)
	}

	if err := db.CleanupMessages(team.TeamID); err != nil {
		log.Printf("[ERROR] Failed to clean messages for team %s after posting summary: %v", team.TeamID, err)
	}
}

func formatSummary(messages []db.UserMessage) string {
	userMap := make(map[string][]string)

	for _, msg := range messages {
		userMap[msg.UserID] = append(userMap[msg.UserID], msg.Message)
	}

	var summary strings.Builder
	summary.WriteString("📝 *Team Daily Standup Summary*\n\n")

	for userID, encryptedMessages := range userMap {
		summary.WriteString(fmt.Sprintf("• <@%s>\n", userID))

		for _, enc := range encryptedMessages {
			decrypted, err := utils.Decrypt(enc)
			log.Printf("[INFO] User Message %s", decrypted)
			if err != nil {
				log.Printf("[ERROR] Error decrypting message for user %s: %v", userID, err)
				continue
			}

			var parsed map[string]string
			if err := json.Unmarshal([]byte(decrypted), &parsed); err != nil {
				log.Printf("[WARN] Failed to unmarshal message for user %s. Falling back to raw message: %s", userID, decrypted)
				summary.WriteString(fmt.Sprintf("   - %s\n", decrypted))
				continue
			}

			if y, ok := parsed["Yesterday"]; ok && y != "" {
				summary.WriteString(fmt.Sprintf("   📌 *Yesterday:*\n      - %s\n", y))
			}
			if t, ok := parsed["Today"]; ok && t != "" {
				summary.WriteString(fmt.Sprintf("   🎯 *Today:*\n      - %s\n", t))
			}
			if b, ok := parsed["Blockers"]; ok && b != "" {
				summary.WriteString(fmt.Sprintf("   🚧 *Blockers:*\n      - %s\n", b))
			}
			summary.WriteString("\n")
		}
	}

	return summary.String()
}
