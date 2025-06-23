package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

type PromptState struct {
	Step      int               `json:"step"`
	Responses map[string]string `json:"responses"`
	StartedAt time.Time         `json:"started_at"`
}

func InitRedis() {
	url := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(strings.TrimSpace(url))
	if err != nil {
		log.Fatalf("[ERROR] InitRedis: failed to parse redis url: %v\n", err)
	}
	RedisClient = redis.NewClient(opt)

	if _, err = RedisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("[ERROR] InitRedis: Redis connection failed: %v\n", err)
	}
	log.Println("[INFO] Redis connection established")
}

func GetPromptStateKey(teamID, userID string) string {
	return fmt.Sprintf("prompt_state:%s:%s", teamID, userID)
}

func SetPromptState(teamID, userID string, state PromptState, ctx context.Context) error {
	key := GetPromptStateKey(teamID, userID)
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("[ERROR] SetPromptState: failed to marshal state for user %s: %v\n", userID, err)
		return err
	}
	if err := RedisClient.Set(ctx, key, data, promptTTL*time.Minute).Err(); err != nil {
		log.Printf("[ERROR] SetPromptState: Redis set failed for key %s: %v\n", key, err)
		return err
	}
	return nil
}

func SetPromptExpiry(teamID, userID string, ttl time.Duration, ctx context.Context) error {
	key := fmt.Sprintf("prompt_expiry:%s:%s", teamID, userID)
	if err := RedisClient.Set(ctx, key, "active", ttl).Err(); err != nil {
		log.Printf("[ERROR] SetPromptExpiry: Redis set failed for key %s: %v\n", key, err)
		return err
	}
	log.Printf("[INFO] SetPromptExpiry: Redis set for key %s\n", key)
	return nil
}

func GetPromptState(teamID, userID string, ctx context.Context) (*PromptState, error) {
	key := GetPromptStateKey(teamID, userID)
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		log.Printf("[ERROR] GetPromptState: Redis get failed for key %s: %v\n", key, err)
		return nil, err
	}

	var state PromptState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		log.Printf("[ERROR] GetPromptState: failed to unmarshal value for key %s: %v\n", key, err)
		return nil, err
	}
	return &state, nil
}

func DeletePromptState(teamID, userID string, ctx context.Context) error {
	key := GetPromptStateKey(teamID, userID)
	if err := RedisClient.Del(ctx, key).Err(); err != nil {
		log.Printf("[ERROR] DeletePromptState: Redis delete failed for key %s: %v\n", key, err)
		return err
	}
	return nil
}
