package api

const (
	slackOAuthAuthorizeURL   = "https://slack.com/oauth/v2/authorize"
	slackOAuthTokenURL       = "https://slack.com/api/oauth.v2.access"
	slackOAuthAuthorizeScope = "chat:write,users:read,channels:read,groups:read"
	slackCallbackEndpoint    = "/slack/oauth/callback"
	slackPostMessagesURL     = "https://slack.com/api/chat.postMessage"
	slackTeamInfoURL         = "https://slack.com/api/team.info"
	slackWelcomeMessage      = "🎉 Thanks for installing MidayBrief!\n\n" +
		"Let's get you set up:\n\n" +
		"1. Set summary channel: `config #channel-name`\n" +
		"2. Set post time: `post time HH:MM` (24-hr format)\n" +
		"3. Your current timezone has been auto-detected and saved.\n" +
		"   If you'd like to change it, use: `timezone Your/Region` (e.g., `timezone Asia/Kolkata`)"
)
