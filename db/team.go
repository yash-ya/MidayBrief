package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SaveTeamConfig(team TeamConfig) error {
	now := time.Now().UTC()
	team.UpdatedAt = now

	var existing TeamConfig
	result := DB.Where("team_id = ?", team.TeamID).First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		team.CreatedAt = now
	}

	err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_token", "bot_user_id", "updated_at", "admin_user_id"}),
	}).Create(&team).Error

	if err != nil {
		log.Printf("[ERROR] SaveTeamConfig: failed to save team %s: %v\n", team.TeamID, err)
		return fmt.Errorf("SaveTeamConfig: failed to save team %s: %w", team.TeamID, err)
	}

	log.Printf("[INFO] SaveTeamConfig: successfully saved team %s\n", team.TeamID)
	return nil
}

func GetAllTeamConfigs() ([]TeamConfig, error) {
	var teams []TeamConfig
	err := DB.Find(&teams).Error
	if err != nil {
		log.Printf("[ERROR] GetAllTeamConfigs: failed to retrieve teams: %v\n", err)
		return nil, err
	}

	log.Printf("[INFO] GetAllTeamConfigs: fetched %d team configs\n", len(teams))
	return teams, nil
}

func GetTeamConfig(teamID string) (*TeamConfig, error) {
	var team TeamConfig
	err := DB.Where("team_id = ?", teamID).First(&team).Error
	if err != nil {
		log.Printf("[ERROR] GetTeamConfig: failed to retrieve team %s: %v\n", teamID, err)
		return nil, fmt.Errorf("GetTeamConfig: failed to retrieve team %s: %w", teamID, err)
	}

	log.Printf("[INFO] GetTeamConfig: retrieved config for team %s\n", teamID)
	return &team, nil
}

func UpdateChannelID(teamID, channelID string) error {
	now := time.Now().UTC()
	err := DB.Model(&TeamConfig{}).
		Where("team_id = ?", teamID).
		Updates(map[string]any{
			"channel_id": channelID,
			"updated_at": now,
		}).Error

	if err != nil {
		log.Printf("[ERROR] UpdateChannelID: failed for team %s: %v\n", teamID, err)
		return fmt.Errorf("UpdateChannelID: failed for team %s: %w", teamID, err)
	}

	log.Printf("[INFO] UpdateChannelID: updated channel for team %s\n", teamID)
	return nil
}

func UpdatePostTime(teamID, postTime string) error {
	now := time.Now().UTC()
	err := DB.Model(&TeamConfig{}).
		Where("team_id = ?", teamID).
		Updates(map[string]any{
			"post_time":  postTime,
			"updated_at": now,
		}).Error

	if err != nil {
		log.Printf("[ERROR] UpdatePostTime: failed for team %s: %v\n", teamID, err)
		return fmt.Errorf("UpdatePostTime: failed for team %s: %w", teamID, err)
	}

	log.Printf("[INFO] UpdatePostTime: updated post time for team %s\n", teamID)
	return nil
}

func UpdatePromptTime(teamID, promptTime string) error {
	now := time.Now().UTC()
	err := DB.Model(&TeamConfig{}).
		Where("team_id = ?", teamID).
		Updates(map[string]any{
			"prompt_time": promptTime,
			"updated_at":  now,
		}).Error

	if err != nil {
		log.Printf("[ERROR] UpdatePromptTime: failed for team %s: %v\n", teamID, err)
		return fmt.Errorf("UpdatePromptTime: failed for team %s: %w", teamID, err)
	}

	log.Printf("[INFO] UpdatePromptTime: updated prompt time for team %s\n", teamID)
	return nil
}

func UpdateTimezone(teamID, timezone string) error {
	now := time.Now().UTC()
	err := DB.Model(&TeamConfig{}).
		Where("team_id = ?", teamID).
		Updates(map[string]any{
			"timezone":   timezone,
			"updated_at": now,
		}).Error

	if err != nil {
		log.Printf("[ERROR] UpdateTimezone: failed for team %s: %v\n", teamID, err)
		return fmt.Errorf("UpdateTimezone: failed for team %s: %w", teamID, err)
	}

	log.Printf("[INFO] UpdateTimezone: updated timezone for team %s\n", teamID)
	return nil
}
