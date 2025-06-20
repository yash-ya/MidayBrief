package db

import (
	"MidayBrief/utils"
	"fmt"
	"log"
	"time"
)

func SaveUserMessage(teamID, userID, text string) error {
	message := UserMessage{
		TeamID:      teamID,
		UserID:      userID,
		Message:     text,
		Timestamp:   time.Now().UTC(),
		MessageHash: utils.Hash(text),
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
	startOfDayInUTC := startOfDay.UTC()
	endOfDayInUTC := endOfDay.UTC()

	var messages []UserMessage
	err := DB.Where("team_id = ? AND timestamp >= ? AND timestamp < ?", teamID, startOfDayInUTC, endOfDayInUTC).Find(&messages).Error

	if err != nil {
		return nil, fmt.Errorf("GetMessagesForTeamToday: failed to fetch messages for team %s: %w", teamID, err)
	}
	return messages, nil
}

func CleanupMessages(teamID string) error {
	return DB.Where("team_id = ?", teamID).Delete(&UserMessage{}).Error
}

func IsDuplicateMessage(teamID, userID, hash, timezone string) bool {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Invalid timezone for duplicate check: %s. Defaulting to UTC", timezone)
		location = time.UTC
	}

	startOfDay := time.Now().In(location).Truncate(24 * time.Hour)
	startOfDayUTC := startOfDay.UTC()

	var count int64
	DB.Model(&UserMessage{}).
		Where("team_id = ? AND user_id = ? AND message_hash = ? AND timestamp >= ?", teamID, userID, hash, startOfDayUTC).
		Count(&count)

	return count > 0
}
