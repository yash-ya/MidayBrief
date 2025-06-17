package slack

type OAuthResponse struct {
	Ok                  bool       `json:"ok"`
	AppID               string     `json:"app_id"`
	AuthedUser          AuthedUser `json:"authed_user"`
	Scope               string     `json:"scope"`
	TokenType           string     `json:"token_type"`
	AccessToken         string     `json:"access_token"`
	BotUserID           string     `json:"bot_user_id"`
	Team                Team       `json:"team"`
	Enterprise          any        `json:"enterprise"`
	IsEnterpriseInstall bool       `json:"is_enterprise_install"`
	Error               string     `json:"error,omitempty"`
}

type AuthedUser struct {
	ID string `json:"id"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type urlVerification struct {
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type SlackEvent struct {
	Type   string         `json:"type"`
	TeamID string         `json:"team_id"`
	Event  SlackEventData `json:"event"`
}

type SlackEventData struct {
	Type        string `json:"type"`
	User        string `json:"user"`
	Text        string `json:"text"`
	Channel     string `json:"channel"`
	ChannelType string `json:"channel_type"`
}