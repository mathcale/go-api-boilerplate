package fixtures

import (
	"time"

	"github.com/google/uuid"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/pkg/jwt"
)

const AuthSessionEmail = "jane@example.com"

func AuthSessionUser(id uuid.UUID, active bool) *user.User {
	return &user.User{ID: id, Email: AuthSessionEmail, Active: active}
}

func AuthSessionUserWithVersion(id uuid.UUID, version int) *user.User {
	u := AuthSessionUser(id, true)
	u.TokenVersion = version

	return u
}

func AuthSessionUserWithRoles(id uuid.UUID, roles []string) *user.User {
	u := AuthSessionUser(id, true)
	u.Roles = roles

	return u
}

func AuthSessionValidConfirmationCode(id uuid.UUID) *user.ConfirmationCode {
	return &user.ConfirmationCode{ID: id, ExpiresAt: time.Now().Add(time.Minute)}
}

func AuthSessionExpiredConfirmationCode(id uuid.UUID) *user.ConfirmationCode {
	return &user.ConfirmationCode{ID: id, ExpiresAt: time.Now().Add(-time.Minute)}
}

func AuthSessionStoredRefreshToken(userID, familyID uuid.UUID) *user.RefreshToken {
	return &user.RefreshToken{
		ID:        familyID,
		UserID:    userID,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

func AuthSessionUsedRefreshToken(userID, familyID uuid.UUID) *user.RefreshToken {
	rt := AuthSessionStoredRefreshToken(userID, familyID)
	usedAt := time.Now()
	rt.UsedAt = &usedAt

	return rt
}

func AuthSessionRevokedRefreshToken(userID, familyID uuid.UUID) *user.RefreshToken {
	rt := AuthSessionStoredRefreshToken(userID, familyID)
	revokedAt := time.Now()
	rt.RevokedAt = &revokedAt

	return rt
}

func AuthSessionVerifiedJWT(userID uuid.UUID, tokenID string) *jwt.Token {
	return &jwt.Token{Subject: userID.String(), ID: tokenID}
}

func AuthSessionVerifiedJWTWithVersion(userID uuid.UUID, tokenID string, version int) *jwt.Token {
	return &jwt.Token{Subject: userID.String(), ID: tokenID, TokenVersion: version}
}

func AuthSessionAccessJWT(token string) *jwt.Token {
	return &jwt.Token{Token: &token}
}

func AuthSessionRefreshJWT(token string, expiresAt *time.Time) *jwt.Token {
	return &jwt.Token{Token: &token, ExpiresAt: expiresAt}
}
