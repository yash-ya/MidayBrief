package utils

import (
	"context"
	"fmt"
	"log"
	"time"
)

func CanUpdateNow(postTime string, timezone string) bool {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("[ERROR] CanUpdateNow: invalid timezone '%s': %v\n", timezone, err)
		return false
	}

	now := time.Now().In(loc)
	postTimeParsed, err := time.ParseInLocation("15:04", postTime, loc)
	if err != nil {
		log.Printf("[ERROR] CanUpdateNow: invalid post time format '%s': %v\n", postTime, err)
		return false
	}

	postToday := time.Date(now.Year(), now.Month(), now.Day(), postTimeParsed.Hour(), postTimeParsed.Minute(), 0, 0, loc)
	diff := postToday.Sub(now)

	log.Printf("[INFO] CanUpdateNow: current time %s, post time %s, diff: %v\n", now.Format("15:04"), postToday.Format("15:04"), diff)

	return diff > 30*time.Minute
}

func GetPromptProgressMessage(state PromptState) string {
	questions := []string{
		"What did you work on yesterday?",
		"What are you working on today?",
		"Any blockers you're facing?",
	}

	step := state.Step

	if step >= len(questions) {
		return "✅ You've already completed all the questions."
	}

	var last string
	if step > 0 {
		last = fmt.Sprintf("✅ Last answered: _%s_", questions[step-1])
	}
	next := fmt.Sprintf("🔜 Next up: _%s_", questions[step])

	return fmt.Sprintf("📝 You're already in the middle of an update.\nYou've answered %d of %d questions.\n%s\n%s", step, len(questions), last, next)
}

func IsRateLimited(teamID, userID, command string, ttl time.Duration, ctx context.Context) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	key := fmt.Sprintf("ratelimit:%s:%s:%s", teamID, userID, command)

	exists, err := RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %v", err)
	}

	if exists == 1 {
		return true, nil
	}

	err = RedisClient.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return false, fmt.Errorf("failed to set rate limit: %v", err)
	}

	return false, nil
}
