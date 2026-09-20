package webcontext_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/infra/web/webcontext"
)

type WebContextTestSuite struct {
	suite.Suite
}

func TestWebContext(t *testing.T) {
	suite.Run(t, new(WebContextTestSuite))
}

func (s *WebContextTestSuite) TestUserIDContext() {
	s.Run("should store and retrieve the user id from context", func() {
		userID := uuid.New()
		ctx := webcontext.WithUserID(context.Background(), userID)

		got, ok := webcontext.UserIDFromContext(ctx)

		s.True(ok)
		s.Equal(userID, got)
	})
}

func (s *WebContextTestSuite) TestUserIDFromContext_Missing() {
	s.Run("should return false and a nil uuid when the user id is not in context", func() {
		got, ok := webcontext.UserIDFromContext(context.Background())

		s.False(ok)
		s.Equal(uuid.Nil, got)
	})
}

func (s *WebContextTestSuite) TestRolesContext() {
	s.Run("should store and retrieve roles from context", func() {
		roles := []string{"admin", "user"}
		ctx := webcontext.WithRoles(context.Background(), roles)

		got, ok := webcontext.RolesFromContext(ctx)

		s.True(ok)
		s.Equal(roles, got)
	})
}

func (s *WebContextTestSuite) TestRolesFromContext_Missing() {
	s.Run("should return false and nil when roles are not in context", func() {
		got, ok := webcontext.RolesFromContext(context.Background())

		s.False(ok)
		s.Nil(got)
	})
}

func (s *WebContextTestSuite) TestCorrelationIDContext() {
	s.Run("should store and retrieve the correlation id from context", func() {
		ctx := webcontext.WithCorrelationID(context.Background(), "correlation-123")

		got, ok := webcontext.CorrelationIDFromContext(ctx)

		s.True(ok)
		s.Equal("correlation-123", got)
	})
}

func (s *WebContextTestSuite) TestCorrelationIDFromContext_Missing() {
	s.Run("should return false and an empty string when the correlation id is not in context", func() {
		got, ok := webcontext.CorrelationIDFromContext(context.Background())

		s.False(ok)
		s.Empty(got)
	})
}
