package models

type User struct {
	ID        int    `gorm:"AUTO_INCREMENT" json:"id"`
	UUID      string `json:"uuid"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    int    `json:"status"`
	Password  string `json:"password"`
	LastLogin int64  `json:"last_login"`
	LastIP    string `json:"last_ip"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	DeletedAt int64  `json:"deleted_at"`
}
