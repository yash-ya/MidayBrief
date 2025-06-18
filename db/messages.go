package db

import (
	"fmt"
	"log"
	"time"
)

func SaveUserMessage(teamID, userID, text string) error {
	log.Printf("Message saved : %s", text)
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

func GetMessagesForTeamToday(teamID string) ([]UserMessage, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var messages []UserMessage
	err := DB.Where("team_id = ? AND timestamp >= ? AND timestamp < ?", teamID, today, tomorrow).
		Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("GetMessagesForTeamToday: failed to fetch messages for team %s: %w", teamID, err)
	}
	return messages, nil
}

func CleanupMessages(teamID string) error {
	return DB.Where("team_id = ?", teamID).Delete(&UserMessage{}).Error
}