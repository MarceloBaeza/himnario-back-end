package domain

import "time"

type Hymn struct {
	Title     string    `json:"title"`
	Content   any       `json:"content,omitempty"`
	EmailUser string    `json:"email_user,omitempty"`
	Id        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
type Category struct {
	ID   int
	Name string
}

func ValidateRolCreateEditHymns(role UserRole) error {
	switch role {
	case RoleAdmin, RoleEditor:
		return nil
	default:
		return ErrUnauthorized
	}
}
