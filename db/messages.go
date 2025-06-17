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

func GetMessagesForTeamToday(teamID string) ([]UserMessage, error) {
    today := time.Now().UTC().Truncate(24 * time.Hour)
    tomorrow := today.Add(24 * time.Hour)

    var messages []UserMessage
    err := DB.Where("team_id = ? AND timestamp >= ? AND timestamp < ?", teamID, today, tomorrow).Find(&messages).Error
    return messages, err
}