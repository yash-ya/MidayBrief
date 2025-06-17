package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	SlackOAuthAuthorizeURL   = "https://slack.com/oauth/v2/authorize"
	SlackOAuthTokenURL       = "https://slack.com/api/oauth.v2.access"
	SlackOAuthAuthorizeScope = "chat:write,users:read,channels:read,groups:read"
)

func HandleSlackInstall(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("SLACK_CLIENT_ID")
	baseURL := os.Getenv("BASE_URL")

	log.Println("ClientID - ", clientID)
	log.Println("BaseURL - ", baseURL)

	redirect := fmt.Sprintf(
		"%s?client_id=%s&scope=%s&redirect_uri=%s/slack/oauth/callback", 
		SlackOAuthAuthorizeURL, 
		clientID, 
		SlackOAuthAuthorizeScope, 
		baseURL)
	
	http.Redirect(w, r, redirect, http.StatusFound)
}

func HandleSlackOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	redirectURI := os.Getenv("BASE_URL") + "/slack/oauth/callback"

	response, error := http.PostForm(SlackOAuthTokenURL, url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
	})

	if error != nil {
		http.Error(w, "OAuth request failed", http.StatusInternalServerError)
		return
	}
	body, _ := io.ReadAll(response.Body)

	var oauthResponse OAuthResponse
	if err := json.Unmarshal(body, &oauthResponse); err!=nil {
		http.Error(w, "Failed to parse Slack OAuth response", http.StatusInternalServerError)
		return
	}

	if !oauthResponse.Ok {
		http.Error(w, "Slack error: "+oauthResponse.Error, http.StatusBadRequest)
		return
	}

	fmt.Printf("OAuth successful for team %s (%s). Bot token: %s\n", oauthResponse.Team.Name, oauthResponse.Team.ID, oauthResponse.AccessToken)
}