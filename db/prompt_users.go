package db

import (
	"log"
	"time"

	"gorm.io/gorm/clause"
)

func AddPromptUser(teamID, userID string) error {
	err := DB.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&PromptUser{
			TeamID:    teamID,
			UserID:    userID,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
		}).Error

	if err != nil {
		log.Printf("[ERROR] AddPromptUser: failed to add user %s to team %s: %v\n", userID, teamID, err)
		return err
	}

	log.Printf("[INFO] AddPromptUser: added user %s to team %s\n", userID, teamID)
	return nil
}

func RemovePromptUser(teamID, userID string) error {
	err := DB.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&PromptUser{}).Error
	if err != nil {
		log.Printf("[ERROR] RemovePromptUser: failed to remove user %s from team %s: %v\n", userID, teamID, err)
		return err
	}

	log.Printf("[INFO] RemovePromptUser: removed user %s from team %s\n", userID, teamID)
	return nil
}

func GetAllPromptUser(teamID string) ([]PromptUser, error) {
	var users []PromptUser
	err := DB.Where("team_id = ?", teamID).Find(&users).Error
	if err != nil {
		log.Printf("[ERROR] GetAllPromptUser: failed to fetch users for team %s: %v\n", teamID, err)
		return nil, err
	}

	log.Printf("[INFO] GetAllPromptUser: fetched %d users for team %s\n", len(users), teamID)
	return users, nil
}

func IsPromptUserExist(teamID, userID string) bool {
	var count int64
	DB.Model(&PromptUser{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count)

	log.Printf("[INFO] IsPromptUserExist: %s user already exists for team %s", userID, teamID)
	return count > 0
}
