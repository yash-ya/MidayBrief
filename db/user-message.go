package db

import "time"

func SaveUserMessage(teamID, userID, text string) error {
	now := time.Now().UTC()
	message := UserMessage {
		TeamID: teamID,
		UserID: userID,
		Message: text,
		Timestamp: now,
	}
	return DB.Create(&message).Error
}