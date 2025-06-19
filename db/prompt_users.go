package db

import "gorm.io/gorm/clause"

func AddPromptUser(teamID, userID string) error {
	return DB.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&PromptUser{TeamID: teamID, UserID: userID, IsActive: true}).Error
}

func RemovePromptUser(teamID, userID string) error {
	return DB.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&UserMessage{}).Error
}
