package bcrypt_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/bcrypt"
)

type PasswordTestSuite struct {
	suite.Suite
}

func TestPassword(t *testing.T) {
	suite.Run(t, new(PasswordTestSuite))
}

func (s *PasswordTestSuite) TestPassword_HashAndVerify() {
	s.Run("should hash a password and successfully verify it", func() {
		pw := bcrypt.NewPassword()

		hash, err := pw.Hash("s3cr3t-password")
		s.Require().NoError(err)
		s.Require().NotNil(hash)
		s.NotEqual("s3cr3t-password", *hash)

		s.Require().NoError(pw.Verify("s3cr3t-password", *hash))
	})
}

func (s *PasswordTestSuite) TestPassword_VerifyRejectsWrongPassword() {
	s.Run("should reject verification with a wrong password", func() {
		pw := bcrypt.NewPassword()

		hash, err := pw.Hash("s3cr3t-password")
		s.Require().NoError(err)

		s.Require().Error(pw.Verify("wrong-password", *hash))
	})
}

func (s *PasswordTestSuite) TestPassword_HashRejectsOverlongPassword() {
	s.Run("should reject hashing a password longer than bcrypt's 72 byte limit", func() {
		pw := bcrypt.NewPassword()

		tooLong := strings.Repeat("a", 73)

		hash, err := pw.Hash(tooLong)

		s.Require().Error(err)
		s.Nil(hash)
	})
}
