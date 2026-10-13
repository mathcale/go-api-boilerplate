package mappers

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
)

type MappersRefreshTokenTestSuite struct {
	suite.Suite
}

func TestMappersRefreshToken(t *testing.T) {
	suite.Run(t, new(MappersRefreshTokenTestSuite))
}

func (s *MappersRefreshTokenTestSuite) TestRefreshTokenModelToDomain() {
	s.Run("should map a refresh token model to its domain representation", func() {
		m := fixtures.DomainPkgRefreshTokenModel()

		got := RefreshTokenModelToDomain(m)

		s.Equal(m.ID, got.ID)
		s.Equal(m.UserID, got.UserID)
		s.Equal(m.FamilyID, got.FamilyID)
		s.Equal(m.ExpiresAt, got.ExpiresAt)
		s.Equal(m.UsedAt, got.UsedAt)
		s.Equal(m.RevokedAt, got.RevokedAt)
		s.Equal(m.CreatedAt, got.CreatedAt)
	})
}

func (s *MappersRefreshTokenTestSuite) TestRefreshTokenDomainToModel() {
	s.Run("should map a domain refresh token to its persistence model", func() {
		rt := fixtures.DomainPkgRefreshToken()

		got := RefreshTokenDomainToModel(rt)

		s.Equal(rt.ID, got.ID)
		s.Equal(rt.UserID, got.UserID)
		s.Equal(rt.FamilyID, got.FamilyID)
		s.Equal(rt.ExpiresAt, got.ExpiresAt)
		s.Nil(got.UsedAt)
		s.Nil(got.RevokedAt)
		s.Equal(rt.CreatedAt, got.CreatedAt)
	})
}
