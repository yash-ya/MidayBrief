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