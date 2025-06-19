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

	http.Redirect(w, r, redirect, http.StatusFound)
}

func HandleSlackOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	baseURL := os.Getenv("BASE_URL")
	if clientID == "" || clientSecret == "" || baseURL == "" {
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
		log.Printf("OAuth request error: %v", err)
		http.Error(w, "OAuth request failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read OAuth response", http.StatusInternalServerError)
		return
	}
	log.Printf("oauth response - %s", string(body))

	var oauthResp OAuthResponse
	if err := json.Unmarshal(body, &oauthResp); err != nil {
		http.Error(w, "Failed to parse Slack OAuth response", http.StatusInternalServerError)
		return
	}

	if !oauthResp.Ok {
		http.Error(w, fmt.Sprintf("Slack error: %s", oauthResp.Error), http.StatusBadRequest)
		return
	}

	encryptedToken, err := utils.Encrypt(oauthResp.AccessToken)
	if err != nil {
		log.Printf("Access token encryption failed: %s", err)
		encryptedToken = oauthResp.AccessToken
	}

	timezone, err := getUserTimeZone(oauthResp.AccessToken, oauthResp.AuthedUser.ID)
	if err != nil {
		log.Println("Could not fetch team timezone:", err)
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
		log.Printf("Failed to save team config: %v", err)
		http.Error(w, "Failed to save team configuration", http.StatusInternalServerError)
		return
	}
	welcomeMsg := fmt.Sprintf(slackWelcomeMessage, timezone)
	sendDM(oauthResp.Team.ID, oauthResp.AuthedUser.ID, welcomeMsg)

	log.Printf("OAuth successful for team %s (%s)", oauthResp.Team.Name, oauthResp.Team.ID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Slack app installed successfully"))
}
