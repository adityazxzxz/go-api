package resources

type Response struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResponseLogin struct {
	Error       bool   `json:"error"`
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
