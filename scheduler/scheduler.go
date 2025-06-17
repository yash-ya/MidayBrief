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
        nowUTC := time.Now().UTC().Format("15:04") // current time as HH:MM
        var teams []db.TeamConfig
        if err := db.DB.Where("post_time = ?", nowUTC).Find(&teams).Error; err != nil {
            log.Println("Failed to query team configs for cron:", err)
            return
        }

        for _, team := range teams {
            go PostSummaryForTeam(team)
        }
    })

    c.Start()
}

func PostSummaryForTeam(team db.TeamConfig) {
	messages, err := db.GetMessagesForTeamToday(team.TeamID)
	if err != nil {
		log.Println("Error fetching messages for team:", err)
		return
	}
	if len(messages) == 0 {
		log.Println("No messages for team:", team.TeamID)
		return
	}

	summary := formatSummary(messages)

	err = api.SendMessage(team.AccessToken, team.ChannelID, summary)
	if err != nil {
		log.Println("Failed to post summary to Slack:", err)
	}
}

func formatSummary(messages []db.UserMessage) string {
	userMap := make(map[string][]string)

	for _, msg := range messages {
		userMap[msg.UserID] = append(userMap[msg.UserID], msg.Message)
	}

	var summary strings.Builder
	summary.WriteString("*📢 Team Daily Standup Summary:*\n")

	for userID, updates := range userMap {
		summary.WriteString(fmt.Sprintf("\n• <@%s>\n", userID))
		for _, u := range updates {
			summary.WriteString(fmt.Sprintf("   - %s\n", u))
		}
	}

	return summary.String()
}