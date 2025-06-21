package scheduler

import (
	"MidayBrief/api"
	"MidayBrief/db"
	"MidayBrief/utils"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

const promptMessage = "Good day! 👋\n\nHope you're doing well. Let's kick off your daily standup.\n\n🕐 First up — *What did you work on yesterday?*\nFeel free to share key highlights or any progress you made."

func StartScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	log.Println("Scheduler started...")

	for t := range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		processSchedule(ctx, t)
		cancel()
	}
}

func processSchedule(ctx context.Context, now time.Time) {
	teams, err := db.GetAllTeamConfigs()
	if err != nil {
		log.Println("Failed to fetch team configs:", err)
		return
	}

	for _, team := range teams {
		if team.PostTime == "" || team.PromptTime == "" || team.Timezone == "" || team.ChannelID == "" {
			continue
		}

		loc, err := time.LoadLocation(team.Timezone)
		if err != nil {
			log.Printf("Invalid timezone for team %s: %s\n", team.TeamID, team.Timezone)
			continue
		}

		localTime := now.In(loc).Format("15:04")

		if localTime == team.PromptTime {
			log.Printf("Triggering prompt for team %s at %s (%s)", team.TeamID, localTime, team.Timezone)
			go triggerPromptForTeam(team, ctx)
		}

		if localTime == team.PostTime {
			log.Printf("Triggering post summary for team %s at %s (%s)", team.TeamID, localTime, team.Timezone)
			go postSummaryForTeam(team, loc)
			if err := db.CleanupMessages(team.TeamID); err != nil {
				log.Printf("Failed to clean messages for team %s: %v", team.TeamID, err)
			}
		}
	}
}

func triggerPromptForTeam(team db.TeamConfig, ctx context.Context) {
	users, err := db.GetAllPromptUser(team.TeamID)
	if err != nil {
		log.Printf("Failed to get prompt users for %s: %v", team.TeamID, err)
		return
	}

	for _, user := range users {
		state := utils.PromptState{
			Step:      1,
			Responses: make(map[string]string),
		}

		if err := utils.SetPromptState(team.TeamID, user.UserID, state, ctx); err != nil {
			log.Printf("Failed to set prompt state for user %s: %v", user.UserID, err)
			continue
		}

		err := api.SendMessage(team.TeamID, user.UserID, promptMessage)
		if err != nil {
			log.Printf("Failed to send first prompt to user %s: %v", user.UserID, err)
		}
	}
}

func postSummaryForTeam(team db.TeamConfig, location *time.Location) {
	if team.AccessToken == "" || team.ChannelID == "" {
		log.Printf("PostSummaryForTeam: missing credentials for team %s", team.TeamID)
		return
	}

	messages, err := db.GetMessagesForTeamToday(team.TeamID, location)
	if err != nil {
		log.Printf("PostSummaryForTeam: error fetching messages for team %s: %v", team.TeamID, err)
		return
	}

	if len(messages) == 0 {
		log.Printf("PostSummaryForTeam: no messages found for team %s", team.TeamID)
		return
	}

	summary := formatSummary(messages)
	if err := api.SendMessage(team.AccessToken, team.ChannelID, summary); err != nil {
		log.Printf("PostSummaryForTeam: failed to post summary to Slack for team %s: %v", team.TeamID, err)
	}
}

func formatSummary(messages []db.UserMessage) string {
	userMap := make(map[string][]string)

	for _, msg := range messages {
		userMap[msg.UserID] = append(userMap[msg.UserID], msg.Message)
	}

	var summary strings.Builder
	summary.WriteString("Team Daily Standup Summary:\n")

	for userID, updates := range userMap {
		summary.WriteString(fmt.Sprintf("\n• <@%s>\n", userID))
		for _, u := range updates {
			summary.WriteString(fmt.Sprintf("   - %s\n", u))
		}
	}

	return summary.String()
}
