package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"
)

const (
	Issuer = "himnario-api"
)

var SigningMethod = jwt.SigningMethodHS256

type Claims struct {
	User *domain.User
	jwt.RegisteredClaims
}

func GenerateToken(user *domain.User) (string, error) {
	now := time.Now()
	expireTime := time.Duration(property.GetJwtProperty().ExpirationTime) * time.Minute
	claims := Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Subject:   user.Email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expireTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(SigningMethod, claims)

	return token.SignedString([]byte(property.GetJwtProperty().SecretKey))
}

func ValidateToken(tokenString string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != SigningMethod {
			return nil, jwt.ErrSignatureInvalid // invalid signing method
		}
		return []byte(property.GetJwtProperty().SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.User, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
