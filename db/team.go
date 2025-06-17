package db

import (
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

	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "team_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_token", "bot_user_id", "updated_at"}),
	}).Create(&team).Error
}

func GetTeamConfig(teamID string) (*TeamConfig, error) {
	var team TeamConfig
	err := DB.Where("team_id = ?", teamID).First(&team).Error
	return &team, err
}

func UpdateChannelID(teamID, channelID string) error {
    now := time.Now().UTC()
    return DB.Model(&TeamConfig{}).
        Where("team_id = ?", teamID).
        Updates(map[string]any{
            "channel_id": channelID,
            "updated_at": now,
        }).Error
}

func UpdatePostTime(teamID, postTime string) error {
	now := time.Now().UTC()
    return DB.Model(&TeamConfig{}).
        Where("team_id = ?", teamID).
        Updates(map[string]any{
            "post_time": postTime,
            "updated_at": now,
        }).Error
}