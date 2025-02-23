package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
)

type JWTAuth struct {
	mock.Mock
}

func (m *JWTAuth) IssueAccessToken(id string) (*jwt.Token, error) {
	args := m.Called(id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *JWTAuth) IssueRefreshToken(id string) (*jwt.Token, error) {
	args := m.Called(id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *JWTAuth) VerifyAccessToken(token string) (*jwt.Token, error) {
	args := m.Called(token)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *JWTAuth) VerifyRefreshToken(token string) (*jwt.Token, error) {
	args := m.Called(token)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jwt.Token), args.Error(1)
}
