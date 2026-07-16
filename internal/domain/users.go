package domain

const (
	UserRole      = "user"
	AdminRole     = "admin"
	ActivatedRole = "activated"
)

type User struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Surname          string           `json:"surname"`
	Phone            string           `json:"phone"`
	Email            string           `json:"email"`
	Roles            []string         `json:"roles,omitempty"`
	Password         string           `json:"password"`
	Activated        bool             `json:"activated"`
	VerificationCode VerificationCode `json:"verification_code"`
}
