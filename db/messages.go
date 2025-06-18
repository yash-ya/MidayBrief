package db

import (
	"fmt"
	"log"
	"time"
)

func SaveUserMessage(teamID, userID, text string) error {
	message := UserMessage{
		TeamID:    teamID,
		UserID:    userID,
		Message:   text,
		Timestamp: time.Now().UTC(),
	}

	if err := DB.Create(&message).Error; err != nil {
		return fmt.Errorf("SaveUserMessage: failed to save message for team %s, user %s: %w", teamID, userID, err)
	}
	return nil
}

func GetMessagesForTeamToday(teamID string, location *time.Location) ([]UserMessage, error) {
	now := time.Now().In(location)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	endOfDay := startOfDay.Add(24 * time.Hour)
	log.Printf("Start Of Day - %s", startOfDay)
	log.Printf("End Of Day - %s", endOfDay)
	startUTC := startOfDay.UTC()
	endUTC := endOfDay.UTC()
	log.Printf("Start Of Day UTC - %s", startUTC)
	log.Printf("End Of Day UTC - %s", endUTC)
	var messages []UserMessage
	err := DB.Where("team_id = ? AND timestamp >= ? AND timestamp < ?", teamID, startUTC, endUTC).
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("GetMessagesForTeamToday: failed to fetch messages for team %s: %w", teamID, err)
	}
	return messages, nil
}

func CleanupMessages(teamID string) error {
	return DB.Where("team_id = ?", teamID).Delete(&UserMessage{}).Error
}