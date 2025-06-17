package scheduler

import (
	"MidayBrief/api"
	"MidayBrief/db"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

func StartScheduler() {
	c := cron.New()

	c.AddFunc("@every 1m", func() {
		currentTime := time.Now().UTC().Format("15:04")
		var teams []db.TeamConfig

		if err := db.DB.Where("post_time = ?", currentTime).Find(&teams).Error; err != nil {
			log.Printf("StartScheduler: failed to query teams for time %s: %v", currentTime, err)
			return
		}

		for _, team := range teams {
			go PostSummaryForTeam(team)
		}
	})

	c.Start()
}

func PostSummaryForTeam(team db.TeamConfig) {
	if team.AccessToken == "" || team.ChannelID == "" {
		log.Printf("PostSummaryForTeam: missing credentials for team %s", team.TeamID)
		return
	}

	messages, err := db.GetMessagesForTeamToday(team.TeamID)
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