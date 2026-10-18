package fixtures

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

const (
	AuthBEmail          = "jane@example.com"
	AuthBUnknownEmail   = "nobody@example.com"
	AuthBPassword       = "secret123"
	AuthBWrongPassword  = "wrong"
	AuthBStoredHash     = "stored-hash"
	AuthBHashedPassword = "hashed-password"
	AuthBNewPassword    = "new-password"
	AuthBNewHash        = "new-hash"
	AuthBAccessToken    = "access-token"
	AuthBRefreshToken   = "refresh-token"
	AuthBCode           = "code"
	AuthBMixedCaseEmail = "  Jane@Example.COM "
	AuthBShortPassword  = "short"
)

func AuthBNewID() uuid.UUID {
	return uuid.New()
}

func AuthBBoolPtr(v bool) *bool {
	return &v
}

func AuthBStringPtr(v string) *string {
	return &v
}

func AuthBActiveUser() *user.User {
	return &user.User{
		ID:       uuid.New(),
		Email:    AuthBEmail,
		Password: AuthBStoredHash,
		Active:   true,
	}
}

func AuthBActiveUserWithVersion(version int) *user.User {
	u := AuthBActiveUser()
	u.TokenVersion = version

	return u
}

func AuthBInactiveUser() *user.User {
	u := AuthBActiveUser()
	u.Active = false

	return u
}

func AuthBRecoveryUser() *user.User {
	return &user.User{ID: uuid.New(), Email: AuthBEmail}
}

func AuthBSignUpInput() auth.SignUpInput {
	return auth.SignUpInput{
		Name:     "Jane",
		Surname:  "Doe",
		Email:    AuthBEmail,
		Password: AuthBPassword,
	}
}

func AuthBSignUpInputWithoutName() auth.SignUpInput {
	in := AuthBSignUpInput()
	in.Name = ""

	return in
}

func AuthBAccessJWT() *jwt.Token {
	return &jwt.Token{Token: AuthBStringPtr(AuthBAccessToken)}
}

func AuthBRefreshJWT() *jwt.Token {
	expiresAt := time.Now().Add(24 * time.Hour)

	return &jwt.Token{Token: AuthBStringPtr(AuthBRefreshToken), ExpiresAt: &expiresAt}
}

func AuthBRefreshJWTWithoutExpiry() *jwt.Token {
	return &jwt.Token{Token: AuthBStringPtr(AuthBRefreshToken)}
}

func AuthBValidConfirmationCode() *user.ConfirmationCode {
	return &user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Minute)}
}

func AuthBExpiredConfirmationCode() *user.ConfirmationCode {
	return &user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}
}

func AuthBLongPassword() string {
	return strings.Repeat("a", user.MaxPasswordBytes+1)
}

func AuthBSignUpInputWithMixedCaseEmail() auth.SignUpInput {
	in := AuthBSignUpInput()
	in.Email = AuthBMixedCaseEmail

	return in
}

func AuthBSignUpInputWithPassword(password string) auth.SignUpInput {
	in := AuthBSignUpInput()
	in.Password = password

	return in
}
