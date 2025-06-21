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

func InitRedis() {
	url := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(strings.TrimSpace(url))
	if err != nil {
		log.Fatalf("failed to parse redis url: %v", err)
	}
	RedisClient = redis.NewClient(opt)

	_, err = RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	log.Println("Redis connection established")
}

type PromptState struct {
	Step      int               `json:"step"`
	Responses map[string]string `json:"responses"`
}

func GetPromptStateKey(teamID, userID string) string {
	return fmt.Sprintf("prompt_state:%s:%s", teamID, userID)
}

func SetPromptState(teamID, userID string, state PromptState, ctx context.Context) any {
	key := GetPromptStateKey(teamID, userID)
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, data, 12*time.Hour).Err()
}

func GetPromptState(teamID, userID string, ctx context.Context) (*PromptState, error) {
	key := GetPromptStateKey(teamID, userID)
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var state PromptState
	err = json.Unmarshal([]byte(val), &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func DeletePromptState(teamID, userID string, ctx context.Context) error {
	key := GetPromptStateKey(teamID, userID)
	return RedisClient.Del(ctx, key).Err()
}
