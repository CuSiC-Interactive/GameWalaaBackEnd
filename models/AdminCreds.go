package models

type AdminCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}
