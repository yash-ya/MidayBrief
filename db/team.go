package db

import "gorm.io/gorm/clause"

func SaveTeamConfig(team TeamConfig) error {
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "team_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_token", "bot_user_is", "updated_at"}),
	}).Create(&team).Error
}