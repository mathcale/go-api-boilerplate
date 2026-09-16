package auth

import "github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"

const logFieldUserID = "user_id"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func newTokenPair(access, refresh *jwt.Token) TokenPair {
	pair := TokenPair{}

	if access != nil && access.Token != nil {
		pair.AccessToken = *access.Token
	}

	if refresh != nil && refresh.Token != nil {
		pair.RefreshToken = *refresh.Token
	}

	return pair
}
