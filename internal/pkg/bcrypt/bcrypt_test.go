package bcrypt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PasswordTestSuite struct {
	suite.Suite
}

func TestPassword(t *testing.T) {
	suite.Run(t, new(PasswordTestSuite))
}

func (s *PasswordTestSuite) TestHash() {
	s.Run("should hash a password into a different value", func() {
		pw := NewPassword()

		hash, err := pw.Hash("s3cr3t-password")

		s.Require().NoError(err)
		s.Require().NotNil(hash)
		s.NotEqual("s3cr3t-password", *hash)
	})

	s.Run("should reject hashing a password longer than bcrypt's 72 byte limit", func() {
		pw := NewPassword()

		hash, err := pw.Hash(strings.Repeat("a", 73))

		s.Require().Error(err)
		s.Nil(hash)
	})
}

func (s *PasswordTestSuite) TestVerify() {
	s.Run("should verify a password against its hash", func() {
		pw := NewPassword()

		hash, err := pw.Hash("s3cr3t-password")
		s.Require().NoError(err)

		s.Require().NoError(pw.Verify("s3cr3t-password", *hash))
	})

	s.Run("should reject verification with a wrong password", func() {
		pw := NewPassword()

		hash, err := pw.Hash("s3cr3t-password")
		s.Require().NoError(err)

		s.Require().Error(pw.Verify("wrong-password", *hash))
	})
}
