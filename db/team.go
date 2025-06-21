package db

import (
	"MidayBrief/utils"
	"fmt"
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
		return fmt.Errorf("SaveTeamConfig: failed to save team %s: %w", team.TeamID, err)
	}
	return nil
}

func GetAllTeamConfigs() ([]TeamConfig, error) {
	var teams []TeamConfig
	err := DB.Find(&teams).Error
	if err == nil {
		for _, team := range teams {
			team.AccessToken, _ = utils.Decrypt(team.AccessToken)
		}
	}
	return teams, err
}

func GetTeamConfig(teamID string) (*TeamConfig, error) {
	var team TeamConfig
	err := DB.Where("team_id = ?", teamID).First(&team).Error
	if err != nil {
		return nil, fmt.Errorf("GetTeamConfig: failed to retrieve team %s: %w", teamID, err)
	}
	team.AccessToken, _ = utils.Decrypt(team.AccessToken)
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
		return fmt.Errorf("UpdateChannelID: failed for team %s: %w", teamID, err)
	}
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
		return fmt.Errorf("UpdatePostTime: failed for team %s: %w", teamID, err)
	}
	return nil
}

func UpdatePromptTime(teamID, postTime string) error {
	now := time.Now().UTC()
	err := DB.Model(&TeamConfig{}).
		Where("team_id = ?", teamID).
		Updates(map[string]any{
			"prompt_time": postTime,
			"updated_at":  now,
		}).Error

	if err != nil {
		return fmt.Errorf("UpdatePromptTime: failed for team %s: %w", teamID, err)
	}
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
		return fmt.Errorf("UpdateTimezone: failed for team %s: %w", teamID, err)
	}
	return nil
}
