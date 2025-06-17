package api

const (
	slackOAuthAuthorizeURL   = "https://slack.com/oauth/v2/authorize"
	slackOAuthTokenURL       = "https://slack.com/api/oauth.v2.access"
	slackOAuthAuthorizeScope = "chat:write,users:read,channels:read,groups:read"
	slackCallbackEndpoint    = "/slack/oauth/callback"
	slackPostMessagesURL     = "https://slack.com/api/chat.postMessage"
	slackWelcomeMessage      = "🎉 Thanks for installing MidayBrief!\n\n" +
		"Let's get you set up:\n\n" + "1. Set summary channel: `config #channel-name`\n" +
		"2. Set post time: `post time HH:MM` (24-hr UTC)\n" +
		"3. Set your timezone: `timezone Asia/Kolkata`"
)