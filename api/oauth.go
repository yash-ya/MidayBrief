package api

import (
	"MidayBrief/db"
	"MidayBrief/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func HandleSlackInstall(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("SLACK_CLIENT_ID")
	baseURL := os.Getenv("BASE_URL")
	if clientID == "" || baseURL == "" {
		log.Println("[ERROR] SLACK_CLIENT_ID or BASE_URL not set in environment variables")
		http.Error(w, "Slack Client ID or Base URL not configured", http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf(
		"%s?client_id=%s&scope=%s&redirect_uri=%s/slack/oauth/callback",
		slackOAuthAuthorizeURL,
		clientID,
		slackOAuthAuthorizeScope,
		baseURL,
	)

	log.Println("[INFO] Redirecting to Slack OAuth URL:", redirect)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func HandleSlackOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("[ERROR] Missing authorization code in request")
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	baseURL := os.Getenv("BASE_URL")
	if clientID == "" || clientSecret == "" || baseURL == "" {
		log.Println("[ERROR] SLACK_CLIENT_ID, SLACK_CLIENT_SECRET, or BASE_URL not set")
		http.Error(w, "Missing Slack credentials or base URL", http.StatusInternalServerError)
		return
	}

	redirectURI := baseURL + slackCallbackEndpoint

	resp, err := http.PostForm(slackOAuthTokenURL, url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		log.Printf("[ERROR] OAuth token request failed: %v\n", err)
		http.Error(w, "OAuth request failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read OAuth response body: %v\n", err)
		http.Error(w, "Failed to read OAuth response", http.StatusInternalServerError)
		return
	}

	var oauthResp OAuthResponse
	if err := json.Unmarshal(body, &oauthResp); err != nil {
		log.Printf("[ERROR] Failed to unmarshal OAuth response: %v\n", err)
		http.Error(w, "Failed to parse Slack OAuth response", http.StatusInternalServerError)
		return
	}

	if !oauthResp.Ok {
		log.Printf("[ERROR] Slack OAuth returned error: %s\n", oauthResp.Error)
		http.Error(w, fmt.Sprintf("Slack error: %s", oauthResp.Error), http.StatusBadRequest)
		return
	}

	encryptedToken, err := utils.Encrypt(oauthResp.AccessToken)
	if err != nil {
		log.Printf("[ERROR] Failed to encrypt access token: %v\n", err)
		encryptedToken = oauthResp.AccessToken
	}

	timezone, err := getUserTimeZone(oauthResp.AccessToken, oauthResp.AuthedUser.ID)
	if err != nil {
		log.Printf("[WARN] Could not fetch user timezone: %v. Defaulting to UTC\n", err)
		timezone = "UTC"
	}

	team := db.TeamConfig{
		TeamID:      oauthResp.Team.ID,
		AccessToken: encryptedToken,
		BotUserID:   oauthResp.BotUserID,
		AdminUserID: oauthResp.AuthedUser.ID,
		Timezone:    timezone,
	}

	if err := db.SaveTeamConfig(team); err != nil {
		log.Printf("[ERROR] Failed to save team config for team %s: %v\n", team.TeamID, err)
		http.Error(w, "Failed to save team configuration", http.StatusInternalServerError)
		return
	}

	welcomeMsg := fmt.Sprintf(slackWelcomeMessage, timezone)
	if err := SendMessage(oauthResp.AccessToken, oauthResp.AuthedUser.ID, welcomeMsg); err != nil {
		log.Printf("[ERROR] Failed to send welcome message to admin %s: %v\n", oauthResp.AuthedUser.ID, err)
	}

	log.Printf("[INFO] Slack OAuth installation successful for team '%s' (%s)\n", oauthResp.Team.Name, oauthResp.Team.ID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Slack app installed successfully. You can now return to Slack."))
}
