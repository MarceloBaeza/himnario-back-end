package request

type NewHymn struct {
	Title   string `json:"title" validate:"required"`
	Content any    `json:"content" validate:"required"`
	User    User   `json:"user" validate:"required"`
}
type User struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}
