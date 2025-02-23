package fixtures

import (
	"time"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
)

func NewToken() jwt.Token {
	t := "any-jwt-token"
	now := time.Now()
	oneHourLater := now.Add(1 * time.Hour)

	return jwt.Token{
		Token:     &t,
		Issuer:    "any-iss",
		Subject:   generateUUID().String(),
		Audience:  []string{"any-aud"},
		ExpiresAt: &oneHourLater,
		IssuedAt:  &now,
	}
}
