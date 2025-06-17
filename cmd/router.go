package main

import (
	"net/http"

	"MidayBrief/api"

	"github.com/go-chi/chi/v5"
)

func SetupRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", api.HandleHealthCheck)

	r.Get("/slack/install", api.HandleSlackInstall)
	r.Get("/slack/oauth/callback", api.HandleSlackOAuthCallback)
	r.Post("/slack/events", api.HandleSlackEvents)

	return r
}
