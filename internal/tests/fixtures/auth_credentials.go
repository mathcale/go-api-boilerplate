package fixtures

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/pkg/jwt"
	"go-api-boilerplate/internal/usecases/auth"
)

const (
	AuthCredsEmail          = "jane@example.com"
	AuthCredsUnknownEmail   = "nobody@example.com"
	AuthCredsPassword       = "secret123"
	AuthCredsWrongPassword  = "wrong"
	AuthCredsStoredHash     = "stored-hash"
	AuthCredsDummyHash      = "dummy-hash"
	AuthCredsHashedPassword = "hashed-password"
	AuthCredsNewPassword    = "new-password"
	AuthCredsNewHash        = "new-hash"
	AuthCredsAccessToken    = "access-token"
	AuthCredsRefreshToken   = "refresh-token"
	AuthCredsCode           = "code"
	AuthCredsMixedCaseEmail = "  Jane@Example.COM "
	AuthCredsShortPassword  = "short"
)

func AuthCredsNewID() uuid.UUID {
	return uuid.New()
}

func AuthCredsBoolPtr(v bool) *bool {
	return &v
}

func AuthCredsStringPtr(v string) *string {
	return &v
}

func AuthCredsActiveUser() *user.User {
	return &user.User{
		ID:       uuid.New(),
		Email:    AuthCredsEmail,
		Password: AuthCredsStoredHash,
		Active:   true,
	}
}

func AuthCredsActiveUserWithVersion(version int) *user.User {
	u := AuthCredsActiveUser()
	u.TokenVersion = version

	return u
}

func AuthCredsInactiveUser() *user.User {
	u := AuthCredsActiveUser()
	u.Active = false

	return u
}

func AuthCredsRecoveryUser() *user.User {
	return &user.User{ID: uuid.New(), Email: AuthCredsEmail}
}

func AuthCredsSignUpInput() auth.SignUpInput {
	return auth.SignUpInput{
		Name:     WebName,
		Surname:  WebSurname,
		Email:    AuthCredsEmail,
		Password: AuthCredsPassword,
	}
}

func AuthCredsSignUpInputWithoutName() auth.SignUpInput {
	in := AuthCredsSignUpInput()
	in.Name = ""

	return in
}

func AuthCredsAccessJWT() *jwt.Token {
	return &jwt.Token{Token: AuthCredsStringPtr(AuthCredsAccessToken)}
}

func AuthCredsRefreshJWT() *jwt.Token {
	expiresAt := time.Now().Add(24 * time.Hour)

	return &jwt.Token{Token: AuthCredsStringPtr(AuthCredsRefreshToken), ExpiresAt: &expiresAt}
}

func AuthCredsRefreshJWTWithoutExpiry() *jwt.Token {
	return &jwt.Token{Token: AuthCredsStringPtr(AuthCredsRefreshToken)}
}

func AuthCredsValidConfirmationCode() *user.ConfirmationCode {
	return &user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Minute)}
}

func AuthCredsExpiredConfirmationCode() *user.ConfirmationCode {
	return &user.ConfirmationCode{ID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}
}

func AuthCredsLongPassword() string {
	return strings.Repeat("a", user.MaxPasswordBytes+1)
}

func AuthCredsSignUpInputWithMixedCaseEmail() auth.SignUpInput {
	in := AuthCredsSignUpInput()
	in.Email = AuthCredsMixedCaseEmail

	return in
}

func AuthCredsSignUpInputWithPassword(password string) auth.SignUpInput {
	in := AuthCredsSignUpInput()
	in.Password = password

	return in
}
