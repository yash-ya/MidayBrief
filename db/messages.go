package db

import (
	"MidayBrief/utils"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm/clause"
)

func SaveUserMessage(teamID, userID, text string) error {
	message := UserMessage{
		TeamID:      teamID,
		UserID:      userID,
		Message:     text,
		Timestamp:   time.Now().UTC(),
		MessageHash: utils.Hash(text),
	}
	err := DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "team_id"},
			{Name: "user_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"message",
			"timestamp",
			"message_hash",
		}),
	}).Create(&message).Error

	if err != nil {
		log.Printf("[ERROR] SaveUserMessage: failed to save message for team %s, user %s: %v\n", teamID, userID, err)
		return fmt.Errorf("SaveUserMessage: failed to save message: %w", err)
	}

	log.Printf("[INFO] Saved message for team %s, user %s\n", teamID, userID)
	return nil
}

func UpdateUserMessage(teamID, userID, message string) error {
	now := time.Now().UTC()
	err := DB.Model(&UserMessage{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Updates(map[string]any{
			"message":      message,
			"timestamp":    now,
			"message_hash": utils.Hash(message),
		}).Error

	if err != nil {
		log.Printf("[ERROR] UpdateUserMessage: failed to update user message for user %s and team %s: %v\n", userID, teamID, err)
		return fmt.Errorf("UpdateUserMessage: failed to update the stand up update %s: %w", teamID, err)
	}

	log.Printf("[INFO] UpdateUserMessage: updated user %s standup update for team %s\n", userID, teamID)
	return nil
}

func GetMessagesForTeamToday(teamID string, location *time.Location) ([]UserMessage, error) {
	var messages []UserMessage
	err := DB.Where("team_id = ?", teamID).Find(&messages).Error

	if err != nil {
		log.Printf("[ERROR] GetMessagesForTeamToday: failed to fetch messages for team %s: %v\n", teamID, err)
		return nil, fmt.Errorf("GetMessagesForTeamToday: %w", err)
	}

	log.Printf("[INFO] Fetched %d messages for team %s\n", len(messages), teamID)
	return messages, nil
}

func CleanupMessages(teamID string) error {
	if err := DB.Where("team_id = ?", teamID).Delete(&UserMessage{}).Error; err != nil {
		log.Printf("[ERROR] CleanupMessages: failed for team %s: %v\n", teamID, err)
		return err
	}
	log.Printf("[INFO] Cleaned up messages for team %s\n", teamID)
	return nil
}

func IsUserMessageExist(teamID, userID string) bool {
	var count int64
	DB.Model(&UserMessage{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count)
	log.Printf("[INFO] User message exist check for user %s in team %s: %t\n", userID, teamID, count > 0)
	return count > 0
}

func IsDuplicateMessage(teamID, userID, hash, timezone string) bool {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("[WARN] Invalid timezone '%s' for duplicate check, defaulting to UTC\n", timezone)
		location = time.UTC
	}

	startOfDay := time.Now().In(location).Truncate(24 * time.Hour)
	startOfDayUTC := startOfDay.UTC()

	var count int64
	DB.Model(&UserMessage{}).
		Where("team_id = ? AND user_id = ? AND message_hash = ? AND timestamp >= ?", teamID, userID, hash, startOfDayUTC).
		Count(&count)

	log.Printf("[INFO] Duplicate check for user %s in team %s: %t\n", userID, teamID, count > 0)
	return count > 0
}
