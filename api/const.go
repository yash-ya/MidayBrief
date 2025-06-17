package api

const (
	slackOAuthAuthorizeURL   = "https://slack.com/oauth/v2/authorize"
	slackOAuthTokenURL       = "https://slack.com/api/oauth.v2.access"
	slackOAuthAuthorizeScope = "chat:write,users:read,channels:read,groups:read"
	slackCallbackEndpoint    = "/slack/oauth/callback"
	slackPostMessages        = "https://slack.com/api/chat.postMessage"
)