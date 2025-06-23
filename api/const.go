package api

import "time"

const (
	slackOAuthAuthorizeURL      = "https://slack.com/oauth/v2/authorize"
	slackOAuthTokenURL          = "https://slack.com/api/oauth.v2.access"
	slackOAuthAuthorizeScope    = "chat:write,users:read,channels:read,groups:read"
	slackCallbackEndpoint       = "/slack/oauth/callback"
	slackPostMessagesURL        = "https://slack.com/api/chat.postMessage"
	slackUserInfoURL            = "https://slack.com/api/users.info"
	slackUsersListURL           = "https://slack.com/api/users.list"
	promptSessionDuration       = 15 * time.Hour
	slackWelcomeMessageForAdmin = "Hey there, Admin! 👋 Thanks for installing *MidayBrief* — your team's stand-up assistant.\n\n" +
		"I've auto-detected your timezone as *%s*. If that's not right, you can change it anytime with:\n" +
		"`timezone Your/Timezone` (e.g., `timezone Europe/London`)\n\n" +
		"Let's quickly set up your team's daily standups:\n\n" +
		"1️⃣ *Choose a channel to post standups:*\n" +
		"`config #channel-name`\n\n" +
		"2️⃣ *Set a time for the daily summary post (24-hour format):*\n" +
		"`post time HH:MM` → Example: `post time 10:00`\n\n" +
		"3️⃣ *Set a time to prompt your team for updates:*\n" +
		"`prompt time HH:MM` → Example: `prompt time 09:30`\n\n" +
		"4️⃣ *Decide who gets prompted:*\n" +
		"• To include everyone: `add all users`\n" +
		"• To choose specific people: `add user @alice @bob`\n" +
		"• To remove specific people: `remove user @alice @bob`\n\n" +
		"🛠️ You can always tweak these settings later by sending the individual commands above.\n\n" +
		"--- \n\n" +
		"*A heads-up on the Team Member experience:*\n" +
		"Your team members will receive a brief welcome from me, explaining that I'm their daily standup buddy. Each day, I'll ask them 3 quick questions:\n" +
		"1. What they did yesterday\n" +
		"2. What they plan to do today\n" +
		"3. Any blockers in their way\n\n" +
		"At the scheduled post time, I'll compile everyone's responses and share the team summary in the standup channel.\n" +
		"If a team member ever wants to reinitiate their update (e.g., if a session expires), they just need to type: *update*.\n" +
		"Let's keep things short and crisp. Talk soon! 🙂"
	slackUserWelcomeMessage = "Hey there! I'm *MidayBrief*, your daily standup buddy. 👋\n\n" +
		"Each day, I'll ask you 3 quick questions:\n" +
		"1. What you did yesterday\n" +
		"2. What you plan to do today\n" +
		"3. Any blockers in your way\n\n" +
		"At the scheduled post time, I'll compile everyone's responses and share the team summary in the standup channel.\n" +
		"If you ever want to reinitiate your update (e.g., if you were interrupted or your session expired), just type: *update*.\n" +
		"Let's keep things short and crisp. Talk soon! 🙂"
	unrecognizedCommandMessage = "Oops! I couldn't quite figure out what you meant. 🤔 It looks like that command isn't in my playbook.\n\n Try commands like `config #your-channel` or `post time 17:00`. \n\nFor a full list of commands, just type `help`!"
	updatePromptMessage        = "Alright, let's get a fresh update started! 👋\n\nWe'll begin with the first question:\n\n🕐 *What did you work on yesterday?*\nFeel free to share your key highlights or any progress you made."
)
