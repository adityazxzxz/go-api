package resources

type OtpResponse struct {
	Error       bool   `json:"error"`
	Message     string `json:"message"`
	ChallengeID string `json:"challenge_id"`
	OtpDebug    string `json:"otp_debug,omitempty"`
}
