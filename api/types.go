package api

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

type Commands struct {
	Operation []string
}

type MessageResponse struct {
	Ok      bool    `json:"ok"`
	Error   string  `json:"error"`
	Channel string  `json:"channel"`
	Ts      string  `json:"ts"`
	Message Message `json:"message"`
}

type Message struct {
	Text        string       `json:"text"`
	Username    string       `json:"username"`
	BotID       string       `json:"bot_id"`
	Attachments []Attachment `json:"attachments"`
	Type        string       `json:"type"`
	Subtype     string       `json:"subtype"`
	Ts          string       `json:"ts"`
}

type Attachment struct {
	Text     string `json:"text"`
	ID       int64  `json:"id"`
	Fallback string `json:"fallback"`
}
