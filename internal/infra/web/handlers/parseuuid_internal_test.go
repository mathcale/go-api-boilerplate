package handlers

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// The exported handlers always validate the "uuid" struct tag before calling
// parseUUID, so its error branch is unreachable through the public HTTP API.
// It is exercised directly here as a defensive-code safety net.
type ParseUUIDTestSuite struct {
	suite.Suite
}

func TestParseUUID(t *testing.T) {
	suite.Run(t, new(ParseUUIDTestSuite))
}

func (s *ParseUUIDTestSuite) TestInvalid() {
	s.Run("should return an error when the string is not a valid uuid", func() {
		_, err := parseUUID("not-a-uuid", "test_origin")

		s.Require().Error(err)
	})
}

func (s *ParseUUIDTestSuite) TestValid() {
	s.Run("should return the parsed uuid when the string is a valid uuid", func() {
		id, err := parseUUID("123e4567-e89b-12d3-a456-426614174000", "test_origin")

		s.Require().NoError(err)
		s.Equal("123e4567-e89b-12d3-a456-426614174000", id.String())
	})
}
