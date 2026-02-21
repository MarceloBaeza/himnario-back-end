package domain

import "time"

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleEditor UserRole = "editor"
	RoleViewer UserRole = "viewer"
)

// Usuario cuando llega para alguna accion
type User struct {
	Email string   `json:"email"`
	Name  string   `json:"name"`
	Role  UserRole `json:"role"`
	Token string   `json:"token,omitempty"`
}

// Usuario para registro
type UserRegistry struct {
	Email     string
	Password  string
	Name      string
	Role      UserRole
	CreatedAt time.Time
}

// Usuario para login
type UserLogin struct {
	Email    string
	Password string
}
