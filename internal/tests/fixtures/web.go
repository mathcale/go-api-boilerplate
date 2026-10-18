package fixtures

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

const (
	WebEmail           = "jane@example.com"
	WebPassword        = "secret123"
	WebNewPassword     = "new-password123"
	WebCode            = "the-code"
	WebBadCode         = "bad-code"
	WebAccessToken     = "at"
	WebRefreshToken    = "rt"
	WebOldRefreshToken = "old-refresh-token"
	WebBadRefreshToken = "bad-token"
	WebInvalidUserID   = "not-a-uuid"
	WebShortUserID     = "11111111-1111-1111-1111-11111111111"
	WebMalformedBody   = "{"
	WebEmptyBody       = "{}"
)

func WebUserID() uuid.UUID {
	return uuid.New()
}

func WebTokenPair() *auth.TokenPair {
	return &auth.TokenPair{AccessToken: WebAccessToken, RefreshToken: WebRefreshToken}
}

func WebCreatedUser() *user.User {
	return &user.User{
		ID:      uuid.New(),
		Name:    "Jane",
		Surname: "Doe",
		Email:   WebEmail,
		Active:  false,
	}
}

func WebAuthenticatedUser(id uuid.UUID) *user.User {
	return &user.User{ID: id, Name: "Jane", Email: WebEmail}
}

func WebSignInBody() string {
	return webJSON(dto.SignInInput{Email: WebEmail, Password: WebPassword})
}

func WebSignInMissingPasswordBody() string {
	return webJSON(map[string]string{"email": WebEmail})
}

func WebSignUpBody() string {
	return webJSON(dto.SignUpInput{
		Name:     "Jane",
		Surname:  "Doe",
		Email:    WebEmail,
		Password: WebPassword,
	})
}

func WebSignUpMissingFieldsBody() string {
	return webJSON(map[string]string{"name": "Jane"})
}

func WebRefreshTokenBody(token string) string {
	return webJSON(dto.RefreshTokenInput{RefreshToken: token})
}

func WebUserCodeBody(userID, code string) string {
	return webJSON(dto.ConfirmAccountInput{UserID: userID, Code: code})
}

func WebEmailBody() string {
	return webJSON(dto.SetRecoveryCodeInput{Email: WebEmail})
}

func WebUpdatePasswordBody(userID, code string) string {
	return webJSON(dto.UpdatePasswordInput{
		UserID:      userID,
		Code:        code,
		NewPassword: WebNewPassword,
	})
}

func webJSON(v interface{}) string {
	out, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(out)
}

func WebVerifiedAccessToken(userID uuid.UUID, version int) *jwt.Token {
	return &jwt.Token{
		Subject:      userID.String(),
		ExtraClaims:  jwt.ExtraClaims{Roles: WebRoles()},
		TokenVersion: version,
	}
}

func WebAuthState(version int, active bool) *user.AuthState {
	return &user.AuthState{TokenVersion: version, Active: active}
}

func WebRoles() []string {
	return []string{"user", "manager"}
}
