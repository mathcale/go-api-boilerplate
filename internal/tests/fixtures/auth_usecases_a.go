package fixtures

import (
	"time"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
)

const AuthAEmail = "jane@example.com"

func AuthAUser(id uuid.UUID, active bool) *user.User {
	return &user.User{ID: id, Email: AuthAEmail, Active: active}
}

func AuthAUserWithVersion(id uuid.UUID, version int) *user.User {
	u := AuthAUser(id, true)
	u.TokenVersion = version

	return u
}

func AuthAUserWithRoles(id uuid.UUID, roles []string) *user.User {
	u := AuthAUser(id, true)
	u.Roles = roles

	return u
}

func AuthAValidConfirmationCode(id uuid.UUID) *user.ConfirmationCode {
	return &user.ConfirmationCode{ID: id, ExpiresAt: time.Now().Add(time.Minute)}
}

func AuthAExpiredConfirmationCode(id uuid.UUID) *user.ConfirmationCode {
	return &user.ConfirmationCode{ID: id, ExpiresAt: time.Now().Add(-time.Minute)}
}

func AuthAStoredRefreshToken(userID, familyID uuid.UUID) *user.RefreshToken {
	return &user.RefreshToken{
		ID:        familyID,
		UserID:    userID,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

func AuthAUsedRefreshToken(userID, familyID uuid.UUID) *user.RefreshToken {
	rt := AuthAStoredRefreshToken(userID, familyID)
	usedAt := time.Now()
	rt.UsedAt = &usedAt

	return rt
}

func AuthARevokedRefreshToken(userID, familyID uuid.UUID) *user.RefreshToken {
	rt := AuthAStoredRefreshToken(userID, familyID)
	revokedAt := time.Now()
	rt.RevokedAt = &revokedAt

	return rt
}

func AuthAVerifiedJWT(userID uuid.UUID, tokenID string) *jwt.Token {
	return &jwt.Token{Subject: userID.String(), ID: tokenID}
}

func AuthAVerifiedJWTWithVersion(userID uuid.UUID, tokenID string, version int) *jwt.Token {
	return &jwt.Token{Subject: userID.String(), ID: tokenID, TokenVersion: version}
}

func AuthAAccessJWT(token string) *jwt.Token {
	return &jwt.Token{Token: &token}
}

func AuthARefreshJWT(token string, expiresAt *time.Time) *jwt.Token {
	return &jwt.Token{Token: &token, ExpiresAt: expiresAt}
}
