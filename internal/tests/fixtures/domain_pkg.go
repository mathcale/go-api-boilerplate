package fixtures

import (
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/infra/database/models"
	"go-api-boilerplate/internal/pkg/jwt"
	"go-api-boilerplate/internal/tests/mocks"
)

const (
	DomainPkgAccessSecret  = "access-secret"
	DomainPkgRefreshSecret = "refresh-secret"
	DomainPkgIssuer        = "boilerplate"
	DomainPkgAudience      = "boilerplate-clients"
	DomainPkgAvatarURL     = "https://example.com/avatar.png"
	DomainPkgTokenVersion  = 7
)

func DomainPkgCreateUserParams() user.CreateUserParams {
	return user.CreateUserParams{
		Name:     "  Jane  ",
		Surname:  "  Doe  ",
		Email:    "  Jane@Example.com  ",
		Password: "hashed",
		Active:   false,
	}
}

func DomainPkgUser() user.User {
	now := time.Now()
	avatarURL := DomainPkgAvatarURL

	return user.User{
		ID:        uuid.New(),
		Name:      "Jane",
		Surname:   "Doe",
		AvatarURL: &avatarURL,
		Email:     "jane@example.com",
		Password:  "hashed",
		Active:    true,
		Roles:     []string{"admin"},
		CreatedAt: now,
		UpdatedAt: now,

		TokenVersion: 3,
	}
}

func DomainPkgUserModel() models.User {
	now := time.Now()
	avatarURL := DomainPkgAvatarURL

	return models.User{
		ID:        uuid.New(),
		Name:      "Jane",
		Surname:   "Doe",
		AvatarURL: &avatarURL,
		Email:     "jane@example.com",
		Password:  "hashed",
		Active:    true,
		Roles:     pq.StringArray{"admin"},
		CreatedAt: now,
		UpdatedAt: now,

		TokenVersion: 3,
	}
}

func DomainPkgUserAuthStateModel() models.UserAuthState {
	return models.UserAuthState{TokenVersion: 5, Active: true}
}

func DomainPkgConfirmationCode() user.ConfirmationCode {
	now := time.Now()

	return user.ConfirmationCode{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Code:      "the-code",
		Purpose:   user.PurposePasswordRecovery,
		ExpiresAt: now,
		Used:      false,
		CreatedAt: now,
	}
}

func DomainPkgConfirmationCodeModel() models.ConfirmationCode {
	now := time.Now()

	return models.ConfirmationCode{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Code:      "the-code",
		Purpose:   "account_confirmation",
		ExpiresAt: now,
		Used:      true,
		CreatedAt: now,
	}
}

func DomainPkgExpiredConfirmationCode() user.ConfirmationCode {
	return user.ConfirmationCode{ExpiresAt: time.Now().Add(-time.Minute)}
}

func DomainPkgActiveConfirmationCode() user.ConfirmationCode {
	return user.ConfirmationCode{ExpiresAt: time.Now().Add(time.Minute)}
}

func DomainPkgRefreshToken() user.RefreshToken {
	now := time.Now()

	return user.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		FamilyID:  uuid.New(),
		ExpiresAt: now,
		CreatedAt: now,
	}
}

func DomainPkgRefreshTokenModel() models.RefreshToken {
	now := time.Now()
	usedAt := now
	revokedAt := now

	return models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		FamilyID:  uuid.New(),
		ExpiresAt: now,
		UsedAt:    &usedAt,
		RevokedAt: &revokedAt,
		CreatedAt: now,
	}
}

func DomainPkgFamilyRefreshToken() user.RefreshToken {
	return user.NewRefreshTokenFamily(uuid.New(), uuid.New(), time.Now().Add(time.Hour))
}

func DomainPkgExpiredRefreshToken() user.RefreshToken {
	return user.RefreshToken{ExpiresAt: time.Now().Add(-time.Minute)}
}

func DomainPkgActiveRefreshToken() user.RefreshToken {
	return user.RefreshToken{ExpiresAt: time.Now().Add(time.Minute)}
}

func DomainPkgUsedRefreshToken() user.RefreshToken {
	usedAt := time.Now()

	return user.RefreshToken{UsedAt: &usedAt}
}

func DomainPkgRevokedRefreshToken() user.RefreshToken {
	revokedAt := time.Now()

	return user.RefreshToken{RevokedAt: &revokedAt}
}

func DomainPkgJWTAuth() jwt.JWTAuth {
	return jwt.NewJWTAuth(
		mocks.NoopLogger{},
		[]byte(DomainPkgAccessSecret),
		[]byte(DomainPkgRefreshSecret),
		15,
		60,
		DomainPkgIssuer,
		DomainPkgAudience,
	)
}

func DomainPkgClaims() jwtlib.MapClaims {
	return jwtlib.MapClaims{
		"sub": "user-123",
		"iss": DomainPkgIssuer,
		"aud": DomainPkgAudience,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": jwt.TokenTypeAccess,
		"tv":  DomainPkgTokenVersion,
	}
}

func DomainPkgSharedSecretJWTAuth() jwt.JWTAuth {
	return jwt.NewJWTAuth(
		mocks.NoopLogger{},
		[]byte(DomainPkgAccessSecret),
		[]byte(DomainPkgAccessSecret),
		15,
		60,
		DomainPkgIssuer,
		DomainPkgAudience,
	)
}

func DomainPkgSignAccessClaims(claims jwtlib.MapClaims) (string, error) {
	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS512, claims).
		SignedString([]byte(DomainPkgAccessSecret))
}
