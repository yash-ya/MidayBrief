package db

import "time"

type TeamConfig struct {
    ID           uint      `gorm:"primaryKey"`
    TeamID       string    `gorm:"uniqueIndex;not null"`
    TeamName     string
    AccessToken  string    `gorm:"not null"`
    BotUserID    string
    ChannelID    string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type UserMessage struct {
    ID        uint      `gorm:"primaryKey"`
    TeamID    string    `gorm:"index;not null"`
    UserID    string    `gorm:"not null"`
    Text      string    `gorm:"type:text;not null"`
    Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}