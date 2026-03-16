package requests

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type MagicLinkRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyMagicLinkRequest struct {
	MagicToken string `json:"magic_token" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type VerifyOtpRequest struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
	Otp         string `json:"otp" binding:"required"`
}
