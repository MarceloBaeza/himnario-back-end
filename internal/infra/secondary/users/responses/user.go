package responses

type AuthRow struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	Role         string
}
