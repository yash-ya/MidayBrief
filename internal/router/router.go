package router

import (
	"MidayBrief/internal/slack"
	"net/http"
)

func SetupRoutes() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("✅ MidayBrief is alive"))
	})
	http.HandleFunc("/slack/install", slack.HandleSlackInstall)
	http.HandleFunc("/slack/oauth/callback", slack.HandleSlackOAuthCallback)
	http.HandleFunc("/slack/events", slack.HandleSlackEvents)
}