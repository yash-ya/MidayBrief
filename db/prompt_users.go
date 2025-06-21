package db

import (
	"time"

	"gorm.io/gorm/clause"
)

func AddPromptUser(teamID, userID string) error {
	return DB.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&PromptUser{TeamID: teamID, UserID: userID, IsActive: true, CreatedAt: time.Now().UTC()}).Error
}

func RemovePromptUser(teamID, userID string) error {
	return DB.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&PromptUser{}).Error
}

func GetAllPromptUser(teamID string) ([]PromptUser, error) {
	var users []PromptUser
	err := DB.Where("team_id = ?", teamID).Find(&users).Error
	return users, err
}
