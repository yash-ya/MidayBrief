package api

const (
	slackOAuthAuthorizeURL   = "https://slack.com/oauth/v2/authorize"
	slackOAuthTokenURL       = "https://slack.com/api/oauth.v2.access"
	slackOAuthAuthorizeScope = "chat:write,users:read,channels:read,groups:read"
	slackCallbackEndpoint    = "/slack/oauth/callback"
	slackPostMessagesURL     = "https://slack.com/api/chat.postMessage"
	slackUserInfoURL         = "https://slack.com/api/users.info"
	slackUsersListURL        = "https://slack.com/api/users.list"
	slackWelcomeMessage      = "Hey there! 👋 Thanks for installing *MidayBrief* — your team's stand-up assistant.\n\n" +
		"I’ve auto-detected your timezone as *%s*. If that’s not right, you can change it anytime with:\n" +
		"`timezone Your/Timezone` (e.g. `timezone Europe/London`)\n\n" +
		"Let’s quickly set things up:\n\n" +
		"1️⃣ Choose a channel to post standups:\n" +
		"`config #channel-name`\n\n" +
		"2️⃣ Set a time for the daily summary post (24-hour format):\n" +
		"`post time HH:MM`  → Example: `post time 10:00`\n\n" +
		"3️⃣ Set a time to prompt your team for updates:\n" +
		"`prompt time HH:MM`  → Example: `prompt time 09:30`\n\n" +
		"4️⃣ Decide who gets prompted:\n" +
		"• To include everyone: `add all users`\n" +
		"• To choose specific people: `add user @alice @bob`\n\n" +
		"• To remove specific people: `remove user @alice @bob`\n\n" +
		"🛠️ You can always tweak these settings later by sending the individual commands above."
	slackUserWelcomeMessage = " Hey there! I’m *MidayBrief*, your daily standup buddy.\n\n" +
		"Each day, I’ll ask you 3 quick questions: \n" +
		"1. What you did yesterday\n" +
		"2. What you plan to do today\n" +
		"3. Any blockers in your way\n\n" +
		"At the scheduled post time, I’ll compile everyone’s responses and share the team summary in the standup channel.\n" +
		"If you ever want to reinitiate your update, just type: *update*\n" +
		"Let’s keep things short and crisp. Talk soon! 🙂\n"
)
