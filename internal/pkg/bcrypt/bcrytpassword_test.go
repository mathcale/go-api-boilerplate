package bcrypt

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BcryptPasswordTestSuite struct {
	suite.Suite
	bp Password
}

func (s *BcryptPasswordTestSuite) SetupTest() {
	s.bp = NewPassword()
}

func TestJWTAuth(t *testing.T) {
	suite.Run(t, new(BcryptPasswordTestSuite))
}

func (s *BcryptPasswordTestSuite) TestHash() {
	s.Run("should hash password", func() {
		hash, err := s.bp.Hash("any-password")

		s.NoError(err)
		s.NotNil(hash)
		s.NotEmpty(hash)
	})

	s.Run("should return error when hasing fails", func() {
		hash, err := s.bp.Hash(
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		)

		s.Error(err)
		s.ErrorContains(err, "password hashing failed")
		s.Nil(hash)
	})
}

func (s *BcryptPasswordTestSuite) TestVerify() {
	pass := "any-password"

	s.Run("should verify password", func() {
		hash, err := s.bp.Hash(pass)
		if err != nil {
			s.FailNow(err.Error())
		}

		err = s.bp.Verify(pass, *hash)

		s.NoError(err)
	})

	s.Run("should return error when password and hash mismatch", func() {
		err := s.bp.Verify(pass, "not-a-bcrypt-hash")

		s.Error(err)
	})
}
