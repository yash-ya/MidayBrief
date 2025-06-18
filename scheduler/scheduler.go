package scheduler

import (
	"MidayBrief/api"
	"MidayBrief/db"
	"MidayBrief/utils"
	"fmt"
	"log"
	"strings"
	"time"
)

func StartScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	log.Println("Scheduler started...")

	for range ticker.C {
		runScheduledSummaries()
	}
}

func runScheduledSummaries() {
	teams, err := db.GetAllTeamConfigs()
	if err != nil {
		log.Println("Failed to fetch team configs:", err)
		return
	}

	for _, team := range teams {
		if team.PostTime == "" || team.Timezone == "" || team.ChannelID == "" {
			continue
		}

		loc, err := time.LoadLocation(team.Timezone)
		if err != nil {
			log.Printf("Invalid timezone for team %s: %s\n", team.TeamID, team.Timezone)
			continue
		}

		now := time.Now().In(loc).Format("15:04")
		if now == team.PostTime {
			log.Printf("Triggering summary for team %s at %s (%s)", team.TeamID, now, team.Timezone)
			go postSummaryForTeam(team, loc)
			// cleanupErr := db.CleanupMessages(team.TeamID)
			// if cleanupErr != nil {
			// 	log.Printf("Unable to clean the messages for team %s: %s\n", team.TeamID, cleanupErr)
			// }
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
	accessToken, _ := utils.Decrypt(team.AccessToken)
	if err := api.SendMessage(accessToken, team.ChannelID, summary); err != nil {
		log.Printf("PostSummaryForTeam: failed to post summary to Slack for team %s: %v", team.TeamID, err)
	}
}

func formatSummary(messages []db.UserMessage) string {
	userMap := make(map[string][]string)

	for _, msg := range messages {
		plainMessage, _ := utils.Decrypt(msg.Message)
		userMap[msg.UserID] = append(userMap[msg.UserID], plainMessage)
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