package db

import "time"

type TeamConfig struct {
    ID           uint      `gorm:"primaryKey"`
    TeamID       string    `gorm:"uniqueIndex;not null"`
    AccessToken  string    `gorm:"not null"`
    BotUserID    string
    ChannelID    string
    PostTime     string
    Timezone     string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type UserMessage struct {
    ID        uint      `gorm:"primaryKey"`
    TeamID    string    `gorm:"index;not null"`
    UserID    string    `gorm:"not null"`
    Message   string    `gorm:"not null"`
    Timestamp time.Time 
}