package utils

import (
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
