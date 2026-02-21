package security

import "golang.org/x/crypto/bcrypt"

const seed = 12

func CompareHashAndPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateHashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), seed)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
