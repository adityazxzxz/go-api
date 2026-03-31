package resources

type ResponseLogin struct {
	Error        bool   `json:"error"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type ResponseRefresh struct {
	Error        bool   `json:"error"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type OtpResponse struct {
	Error       bool   `json:"error"`
	Message     string `json:"message"`
	ChallengeID string `json:"challenge_id"`
	OtpDebug    string `json:"otp_debug,omitempty"`
}

type MagicLinkData struct {
	MagicToken string `json:"magic_token"`
	URL        string `json:"url"`
}

type MagicLinkResponse struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Data    MagicLinkData `json:"data"`
}
