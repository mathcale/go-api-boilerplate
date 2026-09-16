package jwt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/tests/mocks"
)

func newAuth() jwt.JWTAuth {
	return jwt.NewJWTAuth(
		mocks.NoopLogger{},
		[]byte("access-secret"),
		[]byte("refresh-secret"),
		15,
		60,
		"boilerplate",
		"boilerplate-clients",
	)
}

func TestJWTAuth_AccessTokenRoundTrip(t *testing.T) {
	auth := newAuth()

	issued, err := auth.IssueAccessToken(jwt.IssueTokenParams{
		UserID:      "user-123",
		ExtraClaims: jwt.ExtraClaims{Roles: []string{"admin", "user"}},
	})

	require.NoError(t, err)
	require.NotNil(t, issued.Token)

	verified, err := auth.VerifyAccessToken(*issued.Token)

	require.NoError(t, err)
	assert.Equal(t, "user-123", verified.Subject)
	assert.Equal(t, "boilerplate", verified.Issuer)
	assert.Equal(t, []string{"admin", "user"}, verified.ExtraClaims.Roles)
}

func TestJWTAuth_AccessTokenRejectedByRefreshVerifier(t *testing.T) {
	auth := newAuth()

	issued, err := auth.IssueAccessToken(jwt.IssueTokenParams{UserID: "user-123"})
	require.NoError(t, err)

	// A token signed with the access secret must not validate as a refresh token.
	_, err = auth.VerifyRefreshToken(*issued.Token)
	require.Error(t, err)
}

func TestJWTAuth_RejectsTamperedToken(t *testing.T) {
	auth := newAuth()

	_, err := auth.VerifyAccessToken("not-a-real-token")
	require.Error(t, err)
}

func TestJWTAuth_RefreshTokenCarriesTokenID(t *testing.T) {
	auth := newAuth()

	issued, err := auth.IssueRefreshToken(jwt.IssueTokenParams{
		UserID:  "user-123",
		TokenID: "some-uuid",
	})
	require.NoError(t, err)
	require.NotNil(t, issued.Token)

	verified, err := auth.VerifyRefreshToken(*issued.Token)

	require.NoError(t, err)
	assert.Equal(t, "some-uuid", verified.ID)
}

func TestJWTAuth_AccessTokenHasNoTokenIDByDefault(t *testing.T) {
	auth := newAuth()

	issued, err := auth.IssueAccessToken(jwt.IssueTokenParams{UserID: "user-123"})
	require.NoError(t, err)

	verified, err := auth.VerifyAccessToken(*issued.Token)

	require.NoError(t, err)
	assert.Empty(t, verified.ID)
}
