package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
)

type JWTAuth struct {
	mock.Mock
}

func (m *JWTAuth) IssueAccessToken(params jwt.IssueTokenParams) (*jwt.Token, error) {
	args := m.Called(params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *JWTAuth) IssueRefreshToken(params jwt.IssueTokenParams) (*jwt.Token, error) {
	args := m.Called(params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *JWTAuth) VerifyAccessToken(at string) (*jwt.Token, error) {
	args := m.Called(at)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *JWTAuth) VerifyRefreshToken(rt string) (*jwt.Token, error) {
	args := m.Called(rt)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jwt.Token), args.Error(1)
}
