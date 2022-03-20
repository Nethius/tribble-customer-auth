package model

type User struct {
	ID       uint   `json:""`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token";sql:"-"`
}
