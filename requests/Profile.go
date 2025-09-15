package requests

type Update struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UpdatedAt int64  `json:"updated_at"`
}
